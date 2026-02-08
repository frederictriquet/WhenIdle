package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

const defaultGUIConfigPath = ".config/whenidle/config.json"

// GUI holds the Fyne application state and references to daemon components.
type GUI struct {
	app        fyne.App
	runner     *TaskRunner
	monitor    *CPUMonitor
	logBuf     *LogBuffer
	config     Config
	configPath string

	enabled   bool
	cancelMon context.CancelFunc
	monCtx    context.Context
}

// RunGUI starts the Fyne application with system tray.
// Blocks on the main thread (required by macOS).
func RunGUI(configPath string) {
	// Use default config path if not provided
	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("[FATAL] Cannot determine home directory: %v", err)
		}
		configPath = filepath.Join(homeDir, defaultGUIConfigPath)
	}

	// Ensure config directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		log.Fatalf("[FATAL] Cannot create config directory: %v", err)
	}

	// Load or create default config
	cfg, err := LoadConfig(configPath)
	if err != nil {
		// If config doesn't exist, create a default one
		cfg = Config{
			Command: "/bin/echo",
			Args:    []string{"WhenIdle task"},
		}
		cfg.applyDefaults()

		if err := SaveConfig(configPath, cfg); err != nil {
			log.Printf("[WARN] Cannot save default config: %v", err)
		}
	}

	// Create log buffer
	logBuf := NewLogBuffer(500)

	// Setup logging to both stdout and buffer
	multiWriter := io.MultiWriter(os.Stdout, logBuf)
	log.SetOutput(multiWriter)

	log.Println("[INFO] WhenIdle GUI starting")

	// Create GUI instance
	gui := &GUI{
		app:        app.New(),
		logBuf:     logBuf,
		config:     cfg,
		configPath: configPath,
		enabled:    false,
	}

	gui.setupTray()

	// Create a hidden window (required for ShowAndRun)
	w := gui.app.NewWindow("WhenIdle")
	w.Resize(fyne.NewSize(1, 1))
	w.SetCloseIntercept(func() {
		w.Hide() // Hide instead of quit
	})

	w.ShowAndRun() // Blocks here (main thread event loop)
}

// setupTray configures the system tray menu.
func (g *GUI) setupTray() {
	if desk, ok := g.app.(desktop.App); ok {
		menu := fyne.NewMenu("WhenIdle",
			fyne.NewMenuItem("Enable Monitoring", func() {
				g.toggleEnabled()
			}),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Configure Task...", func() {
				g.showConfigWindow()
			}),
			fyne.NewMenuItem("View Logs...", func() {
				g.showLogsWindow()
			}),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Quit", func() {
				g.stopMonitoring()
				g.app.Quit()
			}),
		)
		desk.SetSystemTrayMenu(menu)
	}
}

// toggleEnabled switches between enabled and disabled states.
func (g *GUI) toggleEnabled() {
	if g.enabled {
		g.stopMonitoring()
		log.Println("[INFO] Monitoring disabled via GUI")
	} else {
		g.startMonitoring()
		log.Println("[INFO] Monitoring enabled via GUI")
	}
}

// startMonitoring creates and starts the monitor and runner.
func (g *GUI) startMonitoring() {
	g.runner = NewTaskRunner(g.config)
	g.monitor = NewCPUMonitor(g.config, g.runner.OnIdle, g.runner.OnBusy, g.runner.State)

	g.monCtx, g.cancelMon = context.WithCancel(context.Background())
	go g.monitor.Start(g.monCtx)

	g.enabled = true
}

// stopMonitoring stops the monitor and runner if running.
func (g *GUI) stopMonitoring() {
	if !g.enabled {
		return
	}

	// Cancel monitor context
	if g.cancelMon != nil {
		g.cancelMon()
	}

	// Stop runner
	if g.runner != nil {
		g.runner.Stop()
		g.runner.WaitDone()
	}

	g.enabled = false
}

// showConfigWindow opens a dialog with the configuration form.
func (g *GUI) showConfigWindow() {
	w := g.app.NewWindow("Configure Task")
	w.Resize(fyne.NewSize(500, 400))

	// Create form fields
	commandEntry := widget.NewEntry()
	commandEntry.SetText(g.config.Command)

	argsEntry := widget.NewEntry()
	argsEntry.SetText(strings.Join(g.config.Args, " "))

	workdirEntry := widget.NewEntry()
	workdirEntry.SetText(g.config.WorkingDir)

	thresholdEntry := widget.NewEntry()
	thresholdEntry.SetText(fmt.Sprintf("%.1f", g.config.CPUThreshold))

	idleDurationEntry := widget.NewEntry()
	idleDurationEntry.SetText(fmt.Sprintf("%d", g.config.IdleDuration))

	checkIntervalEntry := widget.NewEntry()
	checkIntervalEntry.SetText(fmt.Sprintf("%d", g.config.CheckInterval))

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Command", Widget: commandEntry, HintText: "Path to executable (e.g. /bin/echo)"},
			{Text: "Arguments", Widget: argsEntry, HintText: "Space-separated arguments"},
			{Text: "Working Directory", Widget: workdirEntry, HintText: "Leave empty for current dir"},
			{Text: "CPU Threshold (%)", Widget: thresholdEntry, HintText: "0-100, default 15.0"},
			{Text: "Idle Duration (s)", Widget: idleDurationEntry, HintText: "Seconds, default 120"},
			{Text: "Check Interval (s)", Widget: checkIntervalEntry, HintText: "Seconds, default 5"},
		},
		OnSubmit: func() {
			// Parse and validate
			newCfg := Config{
				Command:    commandEntry.Text,
				Args:       splitArgs(argsEntry.Text),
				WorkingDir: workdirEntry.Text,
			}

			// Parse numeric fields
			threshold, err := strconv.ParseFloat(thresholdEntry.Text, 64)
			if err != nil || threshold <= 0 || threshold > 100 {
				dialog.ShowError(fmt.Errorf("invalid CPU threshold: must be 0-100"), w)
				return
			}
			newCfg.CPUThreshold = threshold

			idleDur, err := strconv.Atoi(idleDurationEntry.Text)
			if err != nil || idleDur <= 0 {
				dialog.ShowError(fmt.Errorf("invalid idle duration: must be positive integer"), w)
				return
			}
			newCfg.IdleDuration = idleDur

			checkInt, err := strconv.Atoi(checkIntervalEntry.Text)
			if err != nil || checkInt <= 0 {
				dialog.ShowError(fmt.Errorf("invalid check interval: must be positive integer"), w)
				return
			}
			newCfg.CheckInterval = checkInt

			// Validate
			if err := newCfg.Validate(); err != nil {
				dialog.ShowError(err, w)
				return
			}

			// Save config
			if err := SaveConfig(g.configPath, newCfg); err != nil {
				dialog.ShowError(fmt.Errorf("failed to save config: %w", err), w)
				return
			}

			// Update GUI config
			g.config = newCfg

			// If monitoring is enabled, restart with new config
			if g.enabled {
				g.stopMonitoring()
				g.startMonitoring()
			}

			log.Printf("[INFO] Configuration updated: command=%s", newCfg.Command)
			w.Close()
		},
		OnCancel: func() {
			w.Close()
		},
		SubmitText: "Save",
		CancelText: "Cancel",
	}

	w.SetContent(container.NewVBox(
		widget.NewLabel("Configure Task Settings"),
		form,
	))
	w.Show()
}

// showLogsWindow opens a window displaying live logs.
func (g *GUI) showLogsWindow() {
	w := g.app.NewWindow("Logs")
	w.Resize(fyne.NewSize(700, 500))

	// Create multi-line entry (read-only)
	logText := widget.NewMultiLineEntry()
	logText.Disable()
	logText.Wrapping = fyne.TextWrapWord

	// Populate with current logs
	lines := g.logBuf.Lines()
	logText.SetText(strings.Join(lines, "\n"))

	scrolled := container.NewScroll(logText)

	// Register onChange to update logs live
	g.logBuf.SetOnChange(func() {
		// Must use fyne.Do since this is called from a goroutine
		fyne.Do(func() {
			lines := g.logBuf.Lines()
			logText.SetText(strings.Join(lines, "\n"))
			// Scroll to bottom
			scrolled.ScrollToBottom()
		})
	})

	// Unregister onChange when window closes
	w.SetCloseIntercept(func() {
		g.logBuf.SetOnChange(nil)
		w.Close()
	})

	w.SetContent(container.NewBorder(
		widget.NewLabel("Live Logs (last 500 lines)"),
		nil, nil, nil,
		scrolled,
	))
	w.Show()
}

// splitArgs splits a string into arguments, respecting quoted strings.
// Simple implementation: splits on spaces, doesn't handle complex quoting.
func splitArgs(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return strings.Fields(s)
}
