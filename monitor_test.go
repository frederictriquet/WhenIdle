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

	// Mock checkIdle to return idle
	monitor.checkIdle = func() (bool, string) {
		return true, "CPU at 20.0%"
	}

	// Simulate ticks until idle is triggered
	monitor.tick(3) // idleCount = 1
	monitor.tick(3) // idleCount = 2
	monitor.tick(3) // idleCount = 3, should trigger idle

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

	// Mock checkIdle to return busy
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

	// Mock idle
	monitor.checkIdle = func() (bool, string) {
		return true, "CPU at 20.0%"
	}

	// Three idle ticks
	monitor.tick(1)
	monitor.tick(1)
	monitor.tick(1)

	if idleCount != 1 {
		t.Errorf("Expected OnIdle called once, got %d", idleCount)
	}

	// Two more idle ticks should NOT call OnIdle again
	monitor.tick(1)
	monitor.tick(1)

	if idleCount != 1 {
		t.Errorf("Expected OnIdle still called once, got %d", idleCount)
	}

	// Simulate a busy spike
	monitor.checkIdle = func() (bool, string) {
		return false, "CPU at 80.0%"
	}
	monitor.tick(1)

	// Reset to idle
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
		CPUThreshold:  15, // should be ignored
	}

	idleTriggered := false
	onIdle := func() { idleTriggered = true }
	onBusy := func() {}
	getState := func() TaskState { return Stopped }

	monitor := NewMonitor(cfg, onIdle, onBusy, getState)

	// Mock: user idle for 90 seconds (above 60s threshold)
	monitor.checkIdle = func() (bool, string) {
		return true, "User idle 90s"
	}

	// In user_idle mode, checksNeeded=1, so one tick suffices
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

	// Mock: user active (only 5s idle, below 60s threshold)
	monitor.checkIdle = func() (bool, string) {
		return false, "User idle 5s"
	}

	monitor.tick(1)

	if !busyTriggered {
		t.Error("Expected OnBusy to trigger when user becomes active")
	}
}
