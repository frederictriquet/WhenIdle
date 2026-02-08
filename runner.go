package main

import (
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

type TaskState int

const (
	Stopped TaskState = iota
	Running
	Paused
)

func (s TaskState) String() string {
	switch s {
	case Stopped:
		return "stopped"
	case Running:
		return "running"
	case Paused:
		return "paused"
	default:
		return "unknown"
	}
}

type TaskRunner struct {
	config Config
	cmd    *exec.Cmd
	state  TaskState
	mu     sync.Mutex
	done   chan struct{}
}

func NewTaskRunner(config Config) *TaskRunner {
	return &TaskRunner{
		config: config,
		state:  Stopped,
	}
}

func (r *TaskRunner) State() TaskState {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.state
}

// OnIdle is called when the system becomes idle.
// If stopped, starts the task. If paused, resumes it.
func (r *TaskRunner) OnIdle() {
	r.mu.Lock()
	defer r.mu.Unlock()

	switch r.state {
	case Stopped:
		r.start()
	case Paused:
		r.resume()
	case Running:
		// Already running, nothing to do
	}
}

// OnBusy is called when the system becomes busy.
// Pauses the running task with SIGSTOP.
func (r *TaskRunner) OnBusy() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.state == Running {
		r.pause()
	}
}

// start launches the configured task as a new process.
// Must be called with mu locked.
func (r *TaskRunner) start() {
	r.cmd = exec.Command(r.config.Command, r.config.Args...)

	if r.config.WorkingDir != "" {
		r.cmd.Dir = r.config.WorkingDir
	}

	// Set process group so we can signal the entire group (including child processes).
	// This is critical: without Setpgid, signals would only affect the direct child,
	// not any subprocesses it spawns.
	r.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	r.cmd.Stdout = os.Stdout
	r.cmd.Stderr = os.Stderr

	if err := r.cmd.Start(); err != nil {
		log.Printf("[ERROR] Failed to start task: %v", err)
		return
	}

	r.state = Running
	r.done = make(chan struct{})
	log.Printf("[INFO] Task started (PID %d): %s %v", r.cmd.Process.Pid, r.config.Command, r.config.Args)

	// Watch for process exit in background
	go r.wait()
}

func (r *TaskRunner) wait() {
	err := r.cmd.Wait()

	r.mu.Lock()
	defer r.mu.Unlock()

	if err != nil {
		log.Printf("[INFO] Task exited with error: %v", err)
	} else {
		log.Println("[INFO] Task completed successfully")
	}

	r.state = Stopped
	r.cmd = nil
	close(r.done)
}

func (r *TaskRunner) pause() {
	if r.cmd == nil || r.cmd.Process == nil {
		return
	}

	// Send SIGSTOP to the process group
	pgid, err := syscall.Getpgid(r.cmd.Process.Pid)
	if err != nil {
		log.Printf("[ERROR] Failed to get process group: %v", err)
		return
	}

	if err := syscall.Kill(-pgid, syscall.SIGSTOP); err != nil {
		log.Printf("[ERROR] Failed to pause task (SIGSTOP): %v", err)
		return
	}

	r.state = Paused
	log.Printf("[INFO] Task paused (PID %d)", r.cmd.Process.Pid)
}

func (r *TaskRunner) resume() {
	if r.cmd == nil || r.cmd.Process == nil {
		return
	}

	pgid, err := syscall.Getpgid(r.cmd.Process.Pid)
	if err != nil {
		log.Printf("[ERROR] Failed to get process group: %v", err)
		return
	}

	if err := syscall.Kill(-pgid, syscall.SIGCONT); err != nil {
		log.Printf("[ERROR] Failed to resume task (SIGCONT): %v", err)
		return
	}

	r.state = Running
	log.Printf("[INFO] Task resumed (PID %d)", r.cmd.Process.Pid)
}

const stopTimeout = 5 * time.Second

// Stop terminates the task gracefully. If paused, resumes first then sends SIGTERM.
// If the process doesn't exit within 5 seconds, sends SIGKILL.
func (r *TaskRunner) Stop() {
	r.mu.Lock()

	if r.state == Stopped || r.cmd == nil || r.cmd.Process == nil {
		r.mu.Unlock()
		return
	}

	pgid, err := syscall.Getpgid(r.cmd.Process.Pid)
	if err != nil {
		log.Printf("[ERROR] Failed to get process group for stop: %v", err)
		r.mu.Unlock()
		return
	}

	// If paused, resume first so the process can handle SIGTERM gracefully.
	// A SIGTERM sent to a SIGSTOP'd process is queued but not delivered until SIGCONT.
	if r.state == Paused {
		_ = syscall.Kill(-pgid, syscall.SIGCONT)
	}

	log.Printf("[INFO] Stopping task (PID %d)...", r.cmd.Process.Pid)
	if err := syscall.Kill(-pgid, syscall.SIGTERM); err != nil {
		log.Printf("[ERROR] Failed to send SIGTERM: %v", err)
	}

	done := r.done
	r.mu.Unlock()

	// Wait for process to exit, with timeout.
	// Lock must be released during the wait to avoid blocking the wait() goroutine
	// which needs to acquire the lock to update state.
	if done != nil {
		select {
		case <-done:
			return
		case <-time.After(stopTimeout):
			// Process didn't exit gracefully within timeout, force kill
			r.mu.Lock()
			if r.cmd != nil && r.cmd.Process != nil {
				log.Printf("[WARN] Task did not exit after %v, sending SIGKILL", stopTimeout)
				pgid, err := syscall.Getpgid(r.cmd.Process.Pid)
				if err == nil {
					_ = syscall.Kill(-pgid, syscall.SIGKILL)
				}
			}
			r.mu.Unlock()
		}
	}
}

// WaitDone blocks until the current task exits. Returns immediately if no task is running.
func (r *TaskRunner) WaitDone() {
	r.mu.Lock()
	done := r.done
	r.mu.Unlock()

	if done != nil {
		<-done
	}
}
