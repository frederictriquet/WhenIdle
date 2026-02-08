package main

import (
	"context"
	"log"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
)

type CPUMonitor struct {
	config       Config
	idleCount    int
	taskLaunched bool // true once we triggered a task in this idle period
	onIdle       func()
	onBusy       func()
	getState     func() TaskState
	getCPU       func() (float64, error) // Mockable CPU getter for testing
}

func NewCPUMonitor(config Config, onIdle func(), onBusy func(), getState func() TaskState) *CPUMonitor {
	m := &CPUMonitor{
		config:   config,
		onIdle:   onIdle,
		onBusy:   onBusy,
		getState: getState,
	}
	// Default: use real CPU check
	m.getCPU = m.checkCPU
	return m
}

func (m *CPUMonitor) Start(ctx context.Context) {
	checksNeeded := m.config.IdleDuration / m.config.CheckInterval
	if checksNeeded < 1 {
		checksNeeded = 1
	}

	ticker := time.NewTicker(time.Duration(m.config.CheckInterval) * time.Second)
	defer ticker.Stop()

	log.Printf("[INFO] Monitor started: threshold=%.1f%%, idle after %ds (%d checks), polling every %ds",
		m.config.CPUThreshold, m.config.IdleDuration, checksNeeded, m.config.CheckInterval)

	// Prime the CPU measurement so subsequent calls with interval=0 return meaningful values
	_, _ = cpu.Percent(0, false)

	for {
		select {
		case <-ctx.Done():
			log.Println("[INFO] Monitor stopped")
			return
		case <-ticker.C:
			m.tick(checksNeeded)
		}
	}
}

// tick processes a single CPU measurement and decides whether to trigger idle/busy events.
// The idle detection logic uses a counter to avoid false positives from brief CPU spikes.
// A task is launched only once per idle period (taskLaunched flag) unless the CPU becomes
// busy again, which resets the flag.
func (m *CPUMonitor) tick(checksNeeded int) {
	usage, err := m.getCPU()
	if err != nil {
		log.Printf("[ERROR] CPU check failed: %v", err)
		return
	}

	state := m.getState()

	if usage < m.config.CPUThreshold {
		m.idleCount++
		log.Printf("[INFO] CPU at %.1f%% - idle for %d/%d checks (task: %s)",
			usage, m.idleCount, checksNeeded, state)

		if m.idleCount >= checksNeeded && !m.taskLaunched {
			// First time reaching idle threshold: start or resume the task
			m.taskLaunched = true
			log.Println("[INFO] System is idle - triggering task")
			m.onIdle()
		} else if m.idleCount >= checksNeeded && state == Paused {
			// Task was paused by a brief busy spike but we're idle again
			log.Println("[INFO] System is idle again - resuming task")
			m.onIdle()
		}
	} else {
		if state == Running {
			log.Printf("[INFO] CPU at %.1f%% - system busy, pausing task", usage)
			m.onBusy()
		} else {
			log.Printf("[INFO] CPU at %.1f%% - system busy", usage)
		}
		m.idleCount = 0
		m.taskLaunched = false // Reset: next idle period will trigger a new task
	}
}

// checkCPU returns the current system-wide CPU usage percentage.
// Uses cpu.Percent with interval=0 for non-blocking measurement (compares with last call).
// This requires a "priming" call in Start() to initialize the baseline.
func (m *CPUMonitor) checkCPU() (float64, error) {
	// interval=0: compares CPU times since the last call (non-blocking)
	percentages, err := cpu.Percent(0, false)
	if err != nil {
		return 0, err
	}
	if len(percentages) == 0 {
		return 0, nil
	}
	return percentages[0], nil
}
