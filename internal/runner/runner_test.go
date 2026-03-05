package runner

import (
	"testing"
	"time"

	"whenidle/internal/config"
)

func TestTaskRunnerStateMachine(t *testing.T) {
	cfg := config.Config{
		Command:    "/bin/echo",
		Args:       []string{"test"},
		WorkingDir: "/tmp",
	}

	runner := NewTaskRunner(cfg)

	if state := runner.State(); state != Stopped {
		t.Errorf("Expected initial state Stopped, got %s", state)
	}

	runner.OnIdle()
	if state := runner.State(); state != Running {
		t.Errorf("Expected state Running after OnIdle, got %s", state)
	}

	// WaitDone is deterministic — no time.Sleep race
	runner.WaitDone()

	if state := runner.State(); state != Stopped {
		t.Errorf("Expected state Stopped after task completion, got %s", state)
	}
}

func TestTaskRunnerPausableTask(t *testing.T) {
	cfg := config.Config{
		Command:    "/bin/sleep",
		Args:       []string{"2"},
		WorkingDir: "/tmp",
	}

	runner := NewTaskRunner(cfg)

	runner.OnIdle()
	time.Sleep(100 * time.Millisecond)

	if state := runner.State(); state != Running {
		t.Errorf("Expected Running, got %s", state)
	}

	runner.OnBusy()
	time.Sleep(100 * time.Millisecond)

	if state := runner.State(); state != Paused {
		t.Errorf("Expected Paused, got %s", state)
	}

	runner.OnIdle()
	time.Sleep(100 * time.Millisecond)

	if state := runner.State(); state != Running {
		t.Errorf("Expected Running after OnIdle, got %s", state)
	}

	runner.Stop()
	time.Sleep(100 * time.Millisecond)

	if state := runner.State(); state != Stopped {
		t.Errorf("Expected Stopped after Stop, got %s", state)
	}
}

func TestTaskRunnerStopWithoutStart(t *testing.T) {
	cfg := config.Config{
		Command:    "/bin/echo",
		Args:       []string{"test"},
		WorkingDir: "/tmp",
	}

	runner := NewTaskRunner(cfg)
	runner.Stop()

	if state := runner.State(); state != Stopped {
		t.Errorf("Expected Stopped, got %s", state)
	}
}

func TestTaskRunnerStopFromPaused(t *testing.T) {
	cfg := config.Config{
		Command:    "/bin/sleep",
		Args:       []string{"5"},
		WorkingDir: "/tmp",
	}

	runner := NewTaskRunner(cfg)
	runner.OnIdle()
	time.Sleep(100 * time.Millisecond)

	runner.OnBusy()
	time.Sleep(50 * time.Millisecond)

	if state := runner.State(); state != Paused {
		t.Fatalf("Expected Paused before stop, got %s", state)
	}

	// Stop from Paused: should resume then SIGTERM
	runner.Stop()
	runner.WaitDone()

	if state := runner.State(); state != Stopped {
		t.Errorf("Expected Stopped after Stop from Paused, got %s", state)
	}
}

func TestTaskRunnerOnIdleWhenRunning(t *testing.T) {
	cfg := config.Config{
		Command:    "/bin/sleep",
		Args:       []string{"2"},
		WorkingDir: "/tmp",
	}

	runner := NewTaskRunner(cfg)
	runner.OnIdle()
	time.Sleep(50 * time.Millisecond)

	// Second OnIdle while already Running should be a no-op
	runner.OnIdle()

	if state := runner.State(); state != Running {
		t.Errorf("Expected Running, got %s", state)
	}

	runner.Stop()
}

func TestTaskRunnerOnBusyWhenStopped(t *testing.T) {
	cfg := config.Config{
		Command:    "/bin/echo",
		Args:       []string{"test"},
		WorkingDir: "/tmp",
	}

	runner := NewTaskRunner(cfg)
	// OnBusy on a Stopped runner should be a no-op
	runner.OnBusy()

	if state := runner.State(); state != Stopped {
		t.Errorf("Expected Stopped, got %s", state)
	}
}

func TestTaskRunnerStartFailure(t *testing.T) {
	cfg := config.Config{
		Command:    "/nonexistent/command/that/does/not/exist",
		WorkingDir: "/tmp",
	}

	runner := NewTaskRunner(cfg)
	runner.OnIdle()

	// Start failed: state stays Stopped
	if state := runner.State(); state != Stopped {
		t.Errorf("Expected Stopped after start failure, got %s", state)
	}
}

func TestTaskRunnerWaitDone(t *testing.T) {
	cfg := config.Config{
		Command:    "/bin/echo",
		Args:       []string{"hello"},
		WorkingDir: "/tmp",
	}

	runner := NewTaskRunner(cfg)

	// WaitDone before start should return immediately
	runner.WaitDone()

	runner.OnIdle()

	// WaitDone should block until echo completes
	done := make(chan struct{})
	go func() {
		runner.WaitDone()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Error("WaitDone did not return within 3 seconds")
	}
}

func TestTaskStateString(t *testing.T) {
	cases := []struct {
		state TaskState
		want  string
	}{
		{Stopped, "stopped"},
		{Running, "running"},
		{Paused, "paused"},
		{TaskState(99), "unknown"},
	}

	for _, tc := range cases {
		got := tc.state.String()
		if got != tc.want {
			t.Errorf("TaskState(%d).String() = %q, want %q", tc.state, got, tc.want)
		}
	}
}
