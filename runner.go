package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
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
	env    []string // environment for spawned processes
}

func NewTaskRunner(config Config) *TaskRunner {
	env := resolveLoginEnv()
	return &TaskRunner{
		config: config,
		state:  Stopped,
		env:    env,
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

	if r.env != nil {
		r.cmd.Env = r.env
	}

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
	if err := r.signalGroup(syscall.SIGSTOP); err != nil {
		log.Printf("[ERROR] Failed to pause task: %v", err)
		return
	}

	r.state = Paused
	log.Printf("[INFO] Task paused (PID %d)", r.cmd.Process.Pid)
}

func (r *TaskRunner) resume() {
	if err := r.signalGroup(syscall.SIGCONT); err != nil {
		log.Printf("[ERROR] Failed to resume task: %v", err)
		return
	}

	r.state = Running
	log.Printf("[INFO] Task resumed (PID %d)", r.cmd.Process.Pid)
}

const stopTimeout = 5 * time.Second

// signalGroup sends a signal to the entire process group of the running task.
// Returns an error if the process group cannot be determined or the signal fails.
// Must be called with mu locked.
func (r *TaskRunner) signalGroup(sig syscall.Signal) error {
	if r.cmd == nil || r.cmd.Process == nil {
		return fmt.Errorf("no running process")
	}

	pgid, err := syscall.Getpgid(r.cmd.Process.Pid)
	if err != nil {
		return fmt.Errorf("failed to get process group: %w", err)
	}

	if err := syscall.Kill(-pgid, sig); err != nil {
		return fmt.Errorf("failed to send %v to process group: %w", sig, err)
	}

	return nil
}

// Stop terminates the task gracefully. If paused, resumes first then sends SIGTERM.
// If the process doesn't exit within 5 seconds, sends SIGKILL.
func (r *TaskRunner) Stop() {
	r.mu.Lock()

	if r.state == Stopped || r.cmd == nil || r.cmd.Process == nil {
		r.mu.Unlock()
		return
	}

	// If paused, resume first so the process can handle SIGTERM gracefully.
	// A SIGTERM sent to a SIGSTOP'd process is queued but not delivered until SIGCONT.
	if r.state == Paused {
		_ = r.signalGroup(syscall.SIGCONT)
	}

	log.Printf("[INFO] Stopping task (PID %d)...", r.cmd.Process.Pid)
	if err := r.signalGroup(syscall.SIGTERM); err != nil {
		log.Printf("[ERROR] Failed to stop task: %v", err)
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
			log.Printf("[WARN] Task did not exit after %v, sending SIGKILL", stopTimeout)
			_ = r.signalGroup(syscall.SIGKILL)
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

// resolveLoginEnv returns the current process environment with PATH replaced
// by the user's login shell PATH. This is necessary when running as a macOS
// Launch Agent, which provides only a minimal PATH (/usr/bin:/bin:/usr/sbin:/sbin).
func resolveLoginEnv() []string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/zsh"
	}

	out, err := exec.Command(shell, "-l", "-c", "echo $PATH").Output()
	if err != nil {
		log.Printf("[WARN] Cannot resolve login shell PATH via %s: %v", shell, err)
		return nil
	}

	loginPATH := strings.TrimSpace(string(out))
	if loginPATH == "" {
		return nil
	}

	env := os.Environ()
	for i, e := range env {
		if strings.HasPrefix(e, "PATH=") {
			env[i] = "PATH=" + loginPATH
			log.Printf("[INFO] Resolved login shell PATH: %s", loginPATH)
			return env
		}
	}

	return append(env, "PATH="+loginPATH)
}
