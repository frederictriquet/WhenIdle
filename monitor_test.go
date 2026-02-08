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

	// Mock getState always returns Stopped (so OnIdle can trigger)
	getState := func() TaskState {
		return Stopped
	}

	monitor := NewCPUMonitor(cfg, onIdle, onBusy, getState)

	// Mock getCPU to return low CPU
	monitor.getCPU = func() (float64, error) {
		return 20.0, nil // Below 50% threshold
	}

	// Simulate ticks until idle is triggered
	monitor.tick(3) // idleCount = 1
	monitor.tick(3) // idleCount = 2
	monitor.tick(3) // idleCount = 3, should trigger idle
	monitor.tick(3) // idleCount = 4

	if !idleTriggered {
		t.Error("Expected OnIdle to be triggered after 3+ checks at low CPU")
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

	// Mock getState returns Running (task is running)
	getState := func() TaskState {
		return Running
	}

	monitor := NewCPUMonitor(cfg, onIdle, onBusy, getState)

	// Mock: start with high CPU, then return high CPU
	monitor.getCPU = func() (float64, error) {
		return 80.0, nil // Well above 20% threshold
	}

	// First tick: high CPU with task Running
	monitor.tick(2)

	if !busyTriggered {
		t.Error("Expected OnBusy to trigger when CPU is high and task is running")
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

	monitor := NewCPUMonitor(cfg, onIdle, onBusy, getState)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run for a moment
	go monitor.Start(ctx)
	time.Sleep(100 * time.Millisecond)

	// Cancel context
	cancel()
	time.Sleep(100 * time.Millisecond)

	// Monitor should have stopped without panicking
	// (we can't easily verify it stopped, but if it panics the test fails)
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

	monitor := NewCPUMonitor(cfg, onIdle, onBusy, getState)

	// Mock low CPU
	monitor.getCPU = func() (float64, error) {
		return 20.0, nil
	}

	// Three low-CPU ticks (enough to trigger idle)
	monitor.tick(1)
	monitor.tick(1)
	monitor.tick(1)

	// OnIdle should be called once
	if idleCount != 1 {
		t.Errorf("Expected OnIdle called once, got %d", idleCount)
	}

	// Two more low-CPU ticks should NOT call OnIdle again
	monitor.tick(1)
	monitor.tick(1)

	if idleCount != 1 {
		t.Errorf("Expected OnIdle still called once, got %d", idleCount)
	}

	// Simulate a busy spike
	monitor.getCPU = func() (float64, error) {
		return 80.0, nil
	}
	monitor.tick(1)

	// Reset to low CPU
	monitor.getCPU = func() (float64, error) {
		return 20.0, nil
	}

	// Return to idle
	monitor.tick(1)
	monitor.tick(1)
	monitor.tick(1)

	// OnIdle should be called again (second time)
	if idleCount != 2 {
		t.Errorf("Expected OnIdle called twice after busy->idle, got %d", idleCount)
	}
}
