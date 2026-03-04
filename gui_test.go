package main

import (
	"reflect"
	"testing"

	"fyne.io/fyne/v2/test"
)

// newTestGUI creates a GUI instance backed by a headless Fyne test app.
func newTestGUI(t *testing.T) *GUI {
	t.Helper()
	a := test.NewApp()
	t.Cleanup(func() { a.Quit() })
	return &GUI{
		app:    a,
		logBuf: NewLogBuffer(100),
		config: Config{
			Command:       "/bin/echo",
			Args:          []string{"test"},
			WorkingDir:    "/tmp",
			CPUThreshold:  50.0,
			IdleDuration:  9999, // prevent monitor from firing during tests
			CheckInterval: 5,
			IdleMode:      IdleModeCPU,
		},
		configPath: t.TempDir() + "/config.json",
	}
}

// ── splitArgs ────────────────────────────────────────────────────────────────

func TestSplitArgsSimple(t *testing.T) {
	got := splitArgs("foo bar baz")
	want := []string{"foo", "bar", "baz"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("splitArgs(\"foo bar baz\") = %v, want %v", got, want)
	}
}

func TestSplitArgsQuoted(t *testing.T) {
	got := splitArgs(`foo "bar baz" qux`)
	want := []string{"foo", "bar baz", "qux"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("splitArgs with quotes = %v, want %v", got, want)
	}
}

func TestSplitArgsMakeArgs(t *testing.T) {
	got := splitArgs(`etl-analyze-laws ARGS="--limit 1"`)
	want := []string{"etl-analyze-laws", "ARGS=--limit 1"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("splitArgs with ARGS= = %v, want %v", got, want)
	}
}

func TestSplitArgsEmpty(t *testing.T) {
	got := splitArgs("")
	if got != nil {
		t.Errorf("splitArgs(\"\") = %v, want nil", got)
	}

	got = splitArgs("   ")
	if got != nil {
		t.Errorf("splitArgs(\"   \") = %v, want nil", got)
	}
}

func TestSplitArgsSingleArg(t *testing.T) {
	got := splitArgs("hello")
	want := []string{"hello"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("splitArgs(\"hello\") = %v, want %v", got, want)
	}
}

// ── GUI lifecycle ────────────────────────────────────────────────────────────

func TestGUISetupTray(t *testing.T) {
	g := newTestGUI(t)
	g.setupTray()
	// Test driver doesn't implement desktop.App → g.desk stays nil, no panic
	if g.desk != nil {
		t.Error("Expected g.desk to be nil with headless test app")
	}
}

func TestGUIUpdateTrayIconNilDesk(t *testing.T) {
	g := newTestGUI(t)
	// g.desk is nil → should return early without panic
	g.updateTrayIcon()
}

func TestGUIStartMonitoring(t *testing.T) {
	g := newTestGUI(t)

	if g.enabled {
		t.Fatal("Expected enabled=false initially")
	}

	g.startMonitoring()
	defer g.stopMonitoring()

	if !g.enabled {
		t.Error("Expected enabled=true after startMonitoring")
	}
	if g.runner == nil {
		t.Error("Expected runner to be non-nil")
	}
	if g.monitor == nil {
		t.Error("Expected monitor to be non-nil")
	}
}

func TestGUIStopMonitoring(t *testing.T) {
	g := newTestGUI(t)

	g.startMonitoring()
	g.stopMonitoring()

	if g.enabled {
		t.Error("Expected enabled=false after stopMonitoring")
	}
}

func TestGUIStopMonitoringWhenNotEnabled(t *testing.T) {
	g := newTestGUI(t)
	// Must not panic when called without a prior startMonitoring
	g.stopMonitoring()
	if g.enabled {
		t.Error("Expected enabled=false")
	}
}

func TestGUIToggleEnabled(t *testing.T) {
	g := newTestGUI(t)

	g.toggleEnabled() // enable
	if !g.enabled {
		t.Error("Expected enabled=true after first toggle")
	}

	g.toggleEnabled() // disable
	if g.enabled {
		t.Error("Expected enabled=false after second toggle")
	}
}

func TestGUIShowLogsWindow(t *testing.T) {
	g := newTestGUI(t)
	g.showLogsWindow()

	if g.logsWindow == nil {
		t.Fatal("Expected logsWindow to be set")
	}

	// Second call should reuse the existing window
	first := g.logsWindow
	g.showLogsWindow()
	if g.logsWindow != first {
		t.Error("Expected same window on second showLogsWindow call")
	}
}

func TestGUIShowConfigWindow(t *testing.T) {
	g := newTestGUI(t)
	g.showConfigWindow()

	if g.configWindow == nil {
		t.Fatal("Expected configWindow to be set")
	}

	// Second call should reuse the existing window
	first := g.configWindow
	g.showConfigWindow()
	if g.configWindow != first {
		t.Error("Expected same window on second showConfigWindow call")
	}
}
