package main

import (
	"testing"
	"time"
)

func TestTaskRunnerStateMachine(t *testing.T) {
	cfg := Config{
		Command:    "/bin/echo",
		Args:       []string{"test"},
		WorkingDir: "/tmp",
	}

	runner := NewTaskRunner(cfg)

	// Initially stopped
	if state := runner.State(); state != Stopped {
		t.Errorf("Expected initial state Stopped, got %s", state)
	}

	// Start task
	runner.OnIdle()
	if state := runner.State(); state != Running {
		t.Errorf("Expected state Running after OnIdle, got %s", state)
	}

	// Wait for completion (echo is quick)
	time.Sleep(500 * time.Millisecond)

	// After task completes, should be back to Stopped
	if state := runner.State(); state != Stopped {
		t.Errorf("Expected state Stopped after task completion, got %s", state)
	}
}

func TestTaskRunnerPausableTask(t *testing.T) {
	cfg := Config{
		Command:    "/bin/sleep",
		Args:       []string{"2"},
		WorkingDir: "/tmp",
	}

	runner := NewTaskRunner(cfg)

	// Start task
	runner.OnIdle()
	time.Sleep(100 * time.Millisecond) // Let it start

	if state := runner.State(); state != Running {
		t.Errorf("Expected Running, got %s", state)
	}

	// Pause task
	runner.OnBusy()
	time.Sleep(100 * time.Millisecond)

	if state := runner.State(); state != Paused {
		t.Errorf("Expected Paused, got %s", state)
	}

	// Resume task
	runner.OnIdle()
	time.Sleep(100 * time.Millisecond)

	if state := runner.State(); state != Running {
		t.Errorf("Expected Running after OnIdle, got %s", state)
	}

	// Stop task
	runner.Stop()
	time.Sleep(100 * time.Millisecond)

	if state := runner.State(); state != Stopped {
		t.Errorf("Expected Stopped after Stop, got %s", state)
	}
}

func TestTaskRunnerStopWithoutStart(t *testing.T) {
	cfg := Config{
		Command:    "/bin/echo",
		Args:       []string{"test"},
		WorkingDir: "/tmp",
	}

	runner := NewTaskRunner(cfg)

	// Stop without starting should not panic
	runner.Stop()

	if state := runner.State(); state != Stopped {
		t.Errorf("Expected Stopped, got %s", state)
	}
}
