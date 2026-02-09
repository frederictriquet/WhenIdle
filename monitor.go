package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
)

// Monitor watches for system idleness and triggers task start/pause accordingly.
//
// It supports two modes via pluggable idle detection:
//   - CPU-based: Idleness detected when CPU usage < threshold for N consecutive checks
//   - User-idle: Idleness detected when no keyboard/mouse events for N seconds
//
// The idle detection strategy is injected via checkIdle function, making the Monitor
// testable and extensible without interface overhead.
type Monitor struct {
	config       Config
	idleCount    int
	taskLaunched bool // true once we triggered a task in this idle period
	onIdle       func()
	onBusy       func()
	getState     func() TaskState
	checkIdle    func() (idle bool, detail string) // pluggable idle detection strategy
}

func NewMonitor(config Config, onIdle func(), onBusy func(), getState func() TaskState) *Monitor {
	m := &Monitor{
		config:   config,
		onIdle:   onIdle,
		onBusy:   onBusy,
		getState: getState,
	}

	switch config.IdleMode {
	case IdleModeUserIdle:
		m.checkIdle = m.checkUserIdle
	default:
		m.checkIdle = m.checkCPUIdle
	}

	return m
}

func (m *Monitor) Start(ctx context.Context) {
	checksNeeded := m.config.IdleDuration / m.config.CheckInterval
	if checksNeeded < 1 {
		checksNeeded = 1
	}

	ticker := time.NewTicker(time.Duration(m.config.CheckInterval) * time.Second)
	defer ticker.Stop()

	if m.config.IdleMode == IdleModeUserIdle {
		checksNeeded = 1 // UserIdleSeconds already measures duration
		log.Printf("[INFO] Monitor started: user idle mode, idle after %ds, polling every %ds",
			m.config.IdleDuration, m.config.CheckInterval)
	} else {
		// Prime the CPU measurement so subsequent calls with interval=0 return meaningful values
		_, _ = cpu.Percent(0, false)
		log.Printf("[INFO] Monitor started: CPU mode, threshold=%.1f%%, idle after %ds (%d checks), polling every %ds",
			m.config.CPUThreshold, m.config.IdleDuration, checksNeeded, m.config.CheckInterval)
	}

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

// tick processes a single idle check and decides whether to trigger idle/busy events.
func (m *Monitor) tick(checksNeeded int) {
	idle, detail := m.checkIdle()
	state := m.getState()

	if idle {
		m.idleCount++

		if state != Running {
			log.Printf("[INFO] %s - idle for %d/%d checks (task: %s)",
				detail, m.idleCount, checksNeeded, state)
		}

		if m.idleCount >= checksNeeded && !m.taskLaunched {
			m.taskLaunched = true
			log.Println("[INFO] System is idle - triggering task")
			m.onIdle()
		} else if m.idleCount >= checksNeeded && m.taskLaunched && state == Stopped && m.config.Restart {
			log.Println("[INFO] Task finished, restarting")
			m.onIdle()
		} else if m.idleCount >= checksNeeded && state == Paused {
			log.Println("[INFO] System is idle again - resuming task")
			m.onIdle()
		}
	} else {
		if state == Running {
			log.Printf("[INFO] %s - busy, pausing task", detail)
			m.onBusy()
		} else if state != Stopped || m.idleCount > 0 {
			log.Printf("[INFO] %s - busy", detail)
		}
		m.idleCount = 0
		m.taskLaunched = false
	}
}

// checkCPUIdle returns true if CPU usage is below the configured threshold.
func (m *Monitor) checkCPUIdle() (bool, string) {
	percentages, err := cpu.Percent(0, false)
	if err != nil {
		log.Printf("[ERROR] CPU check failed: %v", err)
		return false, "CPU error"
	}
	if len(percentages) == 0 {
		return true, "CPU at 0.0%"
	}
	usage := percentages[0]
	return usage < m.config.CPUThreshold, fmt.Sprintf("CPU at %.1f%%", usage)
}

// checkUserIdle returns true if no keyboard/mouse input for longer than IdleDuration.
//
// This method uses UserIdleSeconds() which queries macOS Quartz Event Services
// to get the time elapsed since the last input event. Unlike CPU-based detection,
// this provides an instantaneous measurement rather than requiring N consecutive checks.
//
// Returns:
//   - idle: true if seconds since last input >= IdleDuration
//   - detail: Human-readable string (e.g., "User idle 120s")
func (m *Monitor) checkUserIdle() (bool, string) {
	seconds := UserIdleSeconds()
	idle := seconds >= float64(m.config.IdleDuration)
	return idle, fmt.Sprintf("User idle %.0fs", seconds)
}
