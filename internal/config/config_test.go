package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigValid(t *testing.T) {
	tmpdir := t.TempDir()
	configPath := filepath.Join(tmpdir, "config.json")

	configContent := `{
		"cpu_threshold": 20.5,
		"idle_duration": 60,
		"check_interval": 5,
		"command": "/bin/echo",
		"args": ["test"],
		"working_dir": "/tmp"
	}`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write temp config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.CPUThreshold != 20.5 {
		t.Errorf("Expected CPUThreshold 20.5, got %.1f", cfg.CPUThreshold)
	}
	if cfg.IdleDuration != 60 {
		t.Errorf("Expected IdleDuration 60, got %d", cfg.IdleDuration)
	}
	if cfg.Command != "/bin/echo" {
		t.Errorf("Expected Command '/bin/echo', got %q", cfg.Command)
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	tmpdir := t.TempDir()
	configPath := filepath.Join(tmpdir, "config.json")

	configContent := `{
		"command": "/bin/sh",
		"working_dir": "/tmp"
	}`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write temp config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.CPUThreshold != 15.0 {
		t.Errorf("Expected default CPUThreshold 15.0, got %.1f", cfg.CPUThreshold)
	}
	if cfg.IdleDuration != 120 {
		t.Errorf("Expected default IdleDuration 120, got %d", cfg.IdleDuration)
	}
	if cfg.CheckInterval != 5 {
		t.Errorf("Expected default CheckInterval 5, got %d", cfg.CheckInterval)
	}
}

func TestLoadConfigFileNotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/to/config.json")
	if err == nil {
		t.Error("Expected error for non-existent config file")
	}
}

func TestLoadConfigInvalidJSON(t *testing.T) {
	tmpdir := t.TempDir()
	configPath := filepath.Join(tmpdir, "config.json")

	if err := os.WriteFile(configPath, []byte(`{invalid json`), 0644); err != nil {
		t.Fatalf("Failed to write temp config: %v", err)
	}

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestValidateInvalidThreshold(t *testing.T) {
	cfg := Config{
		CPUThreshold:  150.0,
		IdleDuration:  60,
		CheckInterval: 5,
		Command:       "/bin/echo",
		WorkingDir:    "/tmp",
	}

	if err := cfg.Validate(); err == nil {
		t.Error("Expected validation error for CPUThreshold > 100")
	}
}

func TestValidateZeroThreshold(t *testing.T) {
	cfg := Config{
		CPUThreshold:  0,
		IdleDuration:  60,
		CheckInterval: 5,
		Command:       "/bin/echo",
		IdleMode:      IdleModeCPU,
	}

	if err := cfg.Validate(); err == nil {
		t.Error("Expected validation error for CPUThreshold = 0")
	}
}

func TestValidateInvalidDuration(t *testing.T) {
	cfg := Config{
		CPUThreshold:  20.0,
		IdleDuration:  -10,
		CheckInterval: 5,
		Command:       "/bin/echo",
		WorkingDir:    "/tmp",
	}

	if err := cfg.Validate(); err == nil {
		t.Error("Expected validation error for negative IdleDuration")
	}
}

func TestValidateInvalidCheckInterval(t *testing.T) {
	cfg := Config{
		CPUThreshold:  20.0,
		IdleDuration:  60,
		CheckInterval: 0,
		Command:       "/bin/echo",
		WorkingDir:    "/tmp",
		IdleMode:      IdleModeCPU,
	}

	if err := cfg.Validate(); err == nil {
		t.Error("Expected validation error for CheckInterval = 0")
	}
}

func TestValidateNoCommand(t *testing.T) {
	cfg := Config{
		CPUThreshold:  20.0,
		IdleDuration:  60,
		CheckInterval: 5,
		IdleMode:      IdleModeCPU,
		Command:       "",
		WorkingDir:    "/tmp",
	}

	if err := cfg.Validate(); err == nil {
		t.Error("Expected validation error for empty Command")
	}
}

func TestValidateInvalidWorkingDir(t *testing.T) {
	cfg := Config{
		CPUThreshold:  20.0,
		IdleDuration:  60,
		CheckInterval: 5,
		IdleMode:      IdleModeCPU,
		Command:       "/bin/echo",
		WorkingDir:    "/nonexistent/path/that/should/not/exist",
	}

	if err := cfg.Validate(); err == nil {
		t.Error("Expected validation error for non-existent WorkingDir")
	}
}

func TestValidateWorkingDirIsFile(t *testing.T) {
	tmpdir := t.TempDir()
	filePath := filepath.Join(tmpdir, "not_a_dir.txt")

	if err := os.WriteFile(filePath, []byte("x"), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	cfg := Config{
		CPUThreshold:  20.0,
		IdleDuration:  60,
		CheckInterval: 5,
		Command:       "/bin/echo",
		WorkingDir:    filePath,
		IdleMode:      IdleModeCPU,
	}

	if err := cfg.Validate(); err == nil {
		t.Error("Expected validation error when WorkingDir is a file, not a directory")
	}
}

func TestSaveConfig(t *testing.T) {
	tmpdir := t.TempDir()
	configPath := filepath.Join(tmpdir, "config.json")

	cfg := Config{
		CPUThreshold:  25.5,
		IdleDuration:  90,
		CheckInterval: 10,
		Command:       "/bin/sleep",
		Args:          []string{"60"},
		WorkingDir:    "/tmp",
		LogFile:       "/tmp/test.log",
		IdleMode:      IdleModeUserIdle,
		Restart:       true,
	}

	if err := SaveConfig(configPath, cfg); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("Config file was not created: %v", err)
	}

	loaded, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if loaded.CPUThreshold != cfg.CPUThreshold {
		t.Errorf("CPUThreshold mismatch: expected %.1f, got %.1f", cfg.CPUThreshold, loaded.CPUThreshold)
	}
	if loaded.IdleDuration != cfg.IdleDuration {
		t.Errorf("IdleDuration mismatch: expected %d, got %d", cfg.IdleDuration, loaded.IdleDuration)
	}
	if loaded.Command != cfg.Command {
		t.Errorf("Command mismatch: expected %q, got %q", cfg.Command, loaded.Command)
	}
	if loaded.IdleMode != cfg.IdleMode {
		t.Errorf("IdleMode mismatch: expected %q, got %q", cfg.IdleMode, loaded.IdleMode)
	}
	if loaded.Restart != cfg.Restart {
		t.Errorf("Restart mismatch: expected %v, got %v", cfg.Restart, loaded.Restart)
	}
}

func TestSaveConfigInvalidPath(t *testing.T) {
	cfg := Config{
		CPUThreshold:  20.0,
		IdleDuration:  60,
		CheckInterval: 5,
		Command:       "/bin/echo",
	}

	err := SaveConfig("/nonexistent/dir/config.json", cfg)
	if err == nil {
		t.Error("Expected error when saving to non-existent directory")
	}
}

func TestIdleModeDefault(t *testing.T) {
	tmpdir := t.TempDir()
	configPath := filepath.Join(tmpdir, "config.json")

	configContent := `{
		"command": "/bin/echo",
		"working_dir": "/tmp"
	}`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write temp config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.IdleMode != IdleModeCPU {
		t.Errorf("Expected default IdleMode %q, got %q", IdleModeCPU, cfg.IdleMode)
	}
}

func TestIdleModeUserIdle(t *testing.T) {
	tmpdir := t.TempDir()
	configPath := filepath.Join(tmpdir, "config.json")

	configContent := `{
		"command": "/bin/echo",
		"working_dir": "/tmp",
		"idle_mode": "user_idle"
	}`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write temp config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.IdleMode != IdleModeUserIdle {
		t.Errorf("Expected IdleMode %q, got %q", IdleModeUserIdle, cfg.IdleMode)
	}
}

func TestIdleModeInvalid(t *testing.T) {
	tmpdir := t.TempDir()
	configPath := filepath.Join(tmpdir, "config.json")

	configContent := `{
		"command": "/bin/echo",
		"working_dir": "/tmp",
		"idle_mode": "invalid_mode"
	}`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write temp config: %v", err)
	}

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Error("Expected validation error for invalid idle_mode")
	}
}
