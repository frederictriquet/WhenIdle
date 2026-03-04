package main

import (
	"context"
	"testing"
	"time"
)

func TestMonitorIdleDetection(t *testing.T) {
	cfg := Config{
		CPUThreshold:  50.0,
		IdleDuration:  3,
		CheckInterval: 1,
	}

	idleTriggered := false

	onIdle := func() {
		idleTriggered = true
	}

	onBusy := func() {}

	getState := func() TaskState {
		return Stopped
	}

	monitor := NewMonitor(cfg, onIdle, onBusy, getState)

	monitor.checkIdle = func() (bool, string) {
		return true, "CPU at 20.0%"
	}

	monitor.tick(3)
	monitor.tick(3)
	monitor.tick(3)

	if !idleTriggered {
		t.Error("Expected OnIdle to be triggered after 3+ checks")
	}
}

func TestMonitorBusyDetection(t *testing.T) {
	cfg := Config{
		CPUThreshold:  20.0,
		IdleDuration:  2,
		CheckInterval: 1,
	}

	busyTriggered := false

	onIdle := func() {}

	onBusy := func() {
		busyTriggered = true
	}

	getState := func() TaskState {
		return Running
	}

	monitor := NewMonitor(cfg, onIdle, onBusy, getState)

	monitor.checkIdle = func() (bool, string) {
		return false, "CPU at 80.0%"
	}

	monitor.tick(2)

	if !busyTriggered {
		t.Error("Expected OnBusy to trigger when busy and task is running")
	}
}

func TestMonitorContextCancel(t *testing.T) {
	cfg := Config{
		CPUThreshold:  20.0,
		IdleDuration:  10,
		CheckInterval: 1,
	}

	onIdle := func() {}
	onBusy := func() {}
	getState := func() TaskState { return Stopped }

	monitor := NewMonitor(cfg, onIdle, onBusy, getState)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go monitor.Start(ctx)
	time.Sleep(100 * time.Millisecond)

	cancel()
	time.Sleep(100 * time.Millisecond)
}

func TestMonitorTaskLaunchedFlag(t *testing.T) {
	cfg := Config{
		CPUThreshold:  50.0,
		IdleDuration:  1,
		CheckInterval: 1,
	}

	idleCount := 0
	onIdle := func() {
		idleCount++
	}

	onBusy := func() {}
	getState := func() TaskState { return Stopped }

	monitor := NewMonitor(cfg, onIdle, onBusy, getState)

	monitor.checkIdle = func() (bool, string) {
		return true, "CPU at 20.0%"
	}

	monitor.tick(1)
	monitor.tick(1)
	monitor.tick(1)

	if idleCount != 1 {
		t.Errorf("Expected OnIdle called once, got %d", idleCount)
	}

	monitor.tick(1)
	monitor.tick(1)

	if idleCount != 1 {
		t.Errorf("Expected OnIdle still called once, got %d", idleCount)
	}

	monitor.checkIdle = func() (bool, string) {
		return false, "CPU at 80.0%"
	}
	monitor.tick(1)

	monitor.checkIdle = func() (bool, string) {
		return true, "CPU at 20.0%"
	}

	monitor.tick(1)
	monitor.tick(1)
	monitor.tick(1)

	if idleCount != 2 {
		t.Errorf("Expected OnIdle called twice after busy->idle, got %d", idleCount)
	}
}

func TestMonitorUserIdleMode(t *testing.T) {
	cfg := Config{
		IdleMode:      IdleModeUserIdle,
		IdleDuration:  60,
		CheckInterval: 5,
		CPUThreshold:  15,
	}

	idleTriggered := false
	onIdle := func() { idleTriggered = true }
	onBusy := func() {}
	getState := func() TaskState { return Stopped }

	monitor := NewMonitor(cfg, onIdle, onBusy, getState)

	monitor.checkIdle = func() (bool, string) {
		return true, "User idle 90s"
	}

	monitor.tick(1)

	if !idleTriggered {
		t.Error("Expected OnIdle to trigger in user_idle mode")
	}
}

func TestMonitorUserIdleBusy(t *testing.T) {
	cfg := Config{
		IdleMode:      IdleModeUserIdle,
		IdleDuration:  60,
		CheckInterval: 5,
		CPUThreshold:  15,
	}

	busyTriggered := false
	onIdle := func() {}
	onBusy := func() { busyTriggered = true }
	getState := func() TaskState { return Running }

	monitor := NewMonitor(cfg, onIdle, onBusy, getState)

	monitor.checkIdle = func() (bool, string) {
		return false, "User idle 5s"
	}

	monitor.tick(1)

	if !busyTriggered {
		t.Error("Expected OnBusy to trigger when user becomes active")
	}
}

func TestMonitorTickIdleWhenPaused(t *testing.T) {
	cfg := Config{
		CPUThreshold:  50.0,
		IdleDuration:  1,
		CheckInterval: 1,
	}

	idleCount := 0
	onIdle := func() { idleCount++ }
	onBusy := func() {}
	getState := func() TaskState { return Paused }

	monitor := NewMonitor(cfg, onIdle, onBusy, getState)
	monitor.taskLaunched = true // task was already launched once

	monitor.checkIdle = func() (bool, string) {
		return true, "CPU at 10.0%"
	}

	monitor.tick(1)

	if idleCount != 1 {
		t.Errorf("Expected OnIdle called once to resume paused task, got %d", idleCount)
	}
}

func TestMonitorTickRestartOnIdle(t *testing.T) {
	cfg := Config{
		CPUThreshold:  50.0,
		IdleDuration:  1,
		CheckInterval: 1,
		Restart:       true,
	}

	idleCount := 0
	onIdle := func() { idleCount++ }
	onBusy := func() {}
	getState := func() TaskState { return Stopped }

	monitor := NewMonitor(cfg, onIdle, onBusy, getState)
	monitor.taskLaunched = true // task was already launched and completed

	monitor.checkIdle = func() (bool, string) {
		return true, "CPU at 10.0%"
	}

	monitor.tick(1)

	if idleCount != 1 {
		t.Errorf("Expected OnIdle called once to restart task, got %d", idleCount)
	}
}

func TestMonitorTickBusyNotRunning(t *testing.T) {
	cfg := Config{
		CPUThreshold:  50.0,
		IdleDuration:  3,
		CheckInterval: 1,
	}

	busyCalled := false
	onIdle := func() {}
	onBusy := func() { busyCalled = true }
	getState := func() TaskState { return Stopped }

	monitor := NewMonitor(cfg, onIdle, onBusy, getState)
	monitor.idleCount = 2 // was idle, now goes busy

	monitor.checkIdle = func() (bool, string) {
		return false, "CPU at 80.0%"
	}

	monitor.tick(3) // busy but task not Running — logs but no OnBusy call

	if busyCalled {
		t.Error("Expected OnBusy NOT to be called when task is not Running")
	}
	if monitor.idleCount != 0 {
		t.Errorf("Expected idleCount reset to 0, got %d", monitor.idleCount)
	}
}

func TestMonitorCheckCPUIdle(t *testing.T) {
	cfg := Config{
		CPUThreshold:  50.0,
		IdleDuration:  5,
		CheckInterval: 1,
	}

	onIdle := func() {}
	onBusy := func() {}
	getState := func() TaskState { return Stopped }

	monitor := NewMonitor(cfg, onIdle, onBusy, getState)

	// Call the real checkCPUIdle — should not panic and return a valid detail
	idle, detail := monitor.checkCPUIdle()
	_ = idle
	if detail == "" {
		t.Error("Expected non-empty detail from checkCPUIdle")
	}
}

func TestMonitorCheckUserIdle(t *testing.T) {
	cfg := Config{
		IdleMode:      IdleModeUserIdle,
		IdleDuration:  300,
		CheckInterval: 5,
		CPUThreshold:  15,
	}

	onIdle := func() {}
	onBusy := func() {}
	getState := func() TaskState { return Stopped }

	monitor := NewMonitor(cfg, onIdle, onBusy, getState)

	// Call the real checkUserIdle — should not panic
	idle, detail := monitor.checkUserIdle()
	_ = idle
	if detail == "" {
		t.Error("Expected non-empty detail from checkUserIdle")
	}
}
