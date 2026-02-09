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
	desk       desktop.App
	runner     *TaskRunner
	monitor    *Monitor
	logBuf     *LogBuffer
	config     Config
	configPath string

	enabled    bool
	enableItem *fyne.MenuItem
	cancelMon  context.CancelFunc
	monCtx     context.Context

	logsWindow   fyne.Window
	configWindow fyne.Window
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
	fyneApp := app.New()

	gui := &GUI{
		app:        fyneApp,
		logBuf:     logBuf,
		config:     cfg,
		configPath: configPath,
		enabled:    false,
	}

	gui.setupTray()

	// Hide from Dock after Fyne has fully started
	gui.app.Lifecycle().SetOnStarted(func() {
		HideFromDock()
	})

	// Stop monitoring when app quits (e.g. via Fyne's built-in "Quit" menu item)
	gui.app.Lifecycle().SetOnStopped(func() {
		gui.stopMonitoring()
		log.Println("[INFO] WhenIdle GUI stopped")
	})

	gui.app.Run() // Blocks here (main thread event loop, no window needed)
}

// setupTray configures the system tray menu.
func (g *GUI) setupTray() {
	desk, ok := g.app.(desktop.App)
	if !ok {
		return
	}
	g.desk = desk

	g.enableItem = fyne.NewMenuItem("Enable Monitoring", func() {
		g.toggleEnabled()
	})

	menu := fyne.NewMenu("WhenIdle",
		g.enableItem,
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Configure Task...", func() {
			g.showConfigWindow()
		}),
		fyne.NewMenuItem("View Logs...", func() {
			g.showLogsWindow()
		}),
	)
	desk.SetSystemTrayMenu(menu)
	g.updateTrayIcon()
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
	g.updateTrayIcon()
}

// startMonitoring creates and starts the monitor and runner.
func (g *GUI) startMonitoring() {
	g.runner = NewTaskRunner(g.config)

	// Wrap callbacks to update tray icon when task state changes
	onIdle := func() {
		g.runner.OnIdle()
		fyne.Do(func() { g.updateTrayIcon() })
	}
	onBusy := func() {
		g.runner.OnBusy()
		fyne.Do(func() { g.updateTrayIcon() })
	}
	g.monitor = NewMonitor(g.config, onIdle, onBusy, g.runner.State)

	g.monCtx, g.cancelMon = context.WithCancel(context.Background())
	go g.monitor.Start(g.monCtx)

	g.enabled = true
}

// stopMonitoring stops the monitor and runner if running.
func (g *GUI) stopMonitoring() {
	if !g.enabled {
		return
	}

	if g.cancelMon != nil {
		g.cancelMon()
	}

	if g.runner != nil {
		g.runner.Stop()
		g.runner.WaitDone()
	}

	g.enabled = false
}

// updateTrayIcon updates the system tray icon and menu label based on current state.
func (g *GUI) updateTrayIcon() {
	if g.desk == nil {
		return
	}

	if !g.enabled {
		g.desk.SetSystemTrayIcon(iconDisabled)
		g.enableItem.Label = "Enable Monitoring"
		g.enableItem.Checked = false
	} else {
		// Check task state for more granular icon
		state := Stopped
		if g.runner != nil {
			state = g.runner.State()
		}
		switch state {
		case Running:
			g.desk.SetSystemTrayIcon(iconRunning)
		case Paused:
			g.desk.SetSystemTrayIcon(iconPaused)
		default:
			g.desk.SetSystemTrayIcon(iconIdle)
		}
		g.enableItem.Label = "Disable Monitoring"
		g.enableItem.Checked = true
	}
}

// showConfigWindow opens a dialog with the configuration form, or brings an existing one to front.
func (g *GUI) showConfigWindow() {
	if g.configWindow != nil {
		g.configWindow.Show()
		g.configWindow.RequestFocus()
		return
	}

	w := g.app.NewWindow("Configure Task")
	w.Resize(fyne.NewSize(500, 400))
	g.configWindow = w

	// Create form fields
	commandEntry := widget.NewEntry()
	commandEntry.SetText(g.config.Command)

	argsEntry := widget.NewEntry()
	argsEntry.SetText(strings.Join(g.config.Args, " "))

	workdirEntry := widget.NewEntry()
	workdirEntry.SetText(g.config.WorkingDir)

	idleModeRadio := widget.NewRadioGroup([]string{"CPU Activity", "User Activity"}, nil)
	if g.config.IdleMode == IdleModeUserIdle {
		idleModeRadio.SetSelected("User Activity")
	} else {
		idleModeRadio.SetSelected("CPU Activity")
	}
	idleModeRadio.Horizontal = true

	thresholdEntry := widget.NewEntry()
	thresholdEntry.SetText(fmt.Sprintf("%.1f", g.config.CPUThreshold))

	idleDurationEntry := widget.NewEntry()
	idleDurationEntry.SetText(fmt.Sprintf("%d", g.config.IdleDuration))

	checkIntervalEntry := widget.NewEntry()
	checkIntervalEntry.SetText(fmt.Sprintf("%d", g.config.CheckInterval))

	restartCheck := widget.NewCheck("", nil)
	restartCheck.Checked = g.config.Restart

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Idle Mode", Widget: idleModeRadio, HintText: "What triggers the task"},
			{Text: "Command", Widget: commandEntry, HintText: "Path to executable (e.g. /bin/echo)"},
			{Text: "Arguments", Widget: argsEntry, HintText: "Space-separated arguments"},
			{Text: "Working Directory", Widget: workdirEntry, HintText: "Leave empty for current dir"},
			{Text: "CPU Threshold (%)", Widget: thresholdEntry, HintText: "Only for CPU mode"},
			{Text: "Idle Duration (s)", Widget: idleDurationEntry, HintText: "Seconds, default 120"},
			{Text: "Check Interval (s)", Widget: checkIntervalEntry, HintText: "Seconds, default 5"},
			{Text: "Restart when done", Widget: restartCheck, HintText: "Re-launch task after completion"},
		},
		OnSubmit: func() {
			// Parse numeric fields
			threshold, err := strconv.ParseFloat(thresholdEntry.Text, 64)
			if err != nil {
				dialog.ShowError(fmt.Errorf("CPU threshold: %w", err), w)
				return
			}

			idleDur, err := strconv.Atoi(idleDurationEntry.Text)
			if err != nil {
				dialog.ShowError(fmt.Errorf("idle duration: %w", err), w)
				return
			}

			checkInt, err := strconv.Atoi(checkIntervalEntry.Text)
			if err != nil {
				dialog.ShowError(fmt.Errorf("check interval: %w", err), w)
				return
			}

			idleMode := IdleModeCPU
			if idleModeRadio.Selected == "User Activity" {
				idleMode = IdleModeUserIdle
			}

			newCfg := Config{
				Command:       commandEntry.Text,
				Args:          splitArgs(argsEntry.Text),
				WorkingDir:    workdirEntry.Text,
				CPUThreshold:  threshold,
				IdleDuration:  idleDur,
				CheckInterval: checkInt,
				Restart:       restartCheck.Checked,
				IdleMode:      idleMode,
			}

			// Single source of truth for validation
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
			g.configWindow = nil
			w.Close()
		},
		OnCancel: func() {
			g.configWindow = nil
			w.Close()
		},
		SubmitText: "Save",
		CancelText: "Cancel",
	}

	w.SetCloseIntercept(func() {
		g.configWindow = nil
		w.Close()
	})

	w.SetContent(container.NewVBox(
		widget.NewLabel("Configure Task Settings"),
		form,
	))
	w.Show()
}

// showLogsWindow opens a window displaying live logs, or brings an existing one to front.
func (g *GUI) showLogsWindow() {
	if g.logsWindow != nil {
		g.logsWindow.Show()
		g.logsWindow.RequestFocus()
		return
	}

	w := g.app.NewWindow("Logs")
	w.Resize(fyne.NewSize(700, 500))
	g.logsWindow = w

	// RichText is read-only by nature and displays with normal text colors
	// (unlike a disabled MultiLineEntry which greys out text).
	logText := widget.NewRichTextWithText(strings.Join(g.logBuf.Lines(), "\n"))
	logText.Wrapping = fyne.TextWrapWord

	scrolled := container.NewVScroll(logText)

	// Register onChange to update logs live.
	// Update text segments directly (not ParseMarkdown, which merges single newlines).
	g.logBuf.SetOnChange(func() {
		fyne.Do(func() {
			logText.Segments = []widget.RichTextSegment{
				&widget.TextSegment{Text: strings.Join(g.logBuf.Lines(), "\n")},
			}
			logText.Refresh()
			scrolled.ScrollToBottom()
		})
	})

	// Unregister onChange when window closes
	w.SetCloseIntercept(func() {
		g.logBuf.SetOnChange(nil)
		g.logsWindow = nil
		w.Close()
	})

	w.SetContent(container.NewBorder(
		widget.NewLabel("Live Logs (last 500 lines)"),
		nil, nil, nil,
		scrolled,
	))
	w.Show()
}

// splitArgs splits a string into arguments, respecting double-quoted strings.
// e.g., `foo "bar baz" qux` → ["foo", "bar baz", "qux"]
// e.g., `ARGS="--limit 1"` → [`ARGS=--limit 1`]
func splitArgs(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	var args []string
	var current strings.Builder
	inQuotes := false

	for i := 0; i < len(s); i++ {
		ch := s[i]
		switch {
		case ch == '"':
			inQuotes = !inQuotes
		case ch == ' ' && !inQuotes:
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteByte(ch)
		}
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}
