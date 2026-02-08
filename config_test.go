package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigValid(t *testing.T) {
	// Create a temporary config file
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

	// Minimal config without optional fields
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

func TestValidateInvalidThreshold(t *testing.T) {
	cfg := Config{
		CPUThreshold: 150.0, // Invalid: > 100
		IdleDuration: 60,
		CheckInterval: 5,
		Command: "/bin/echo",
		WorkingDir: "/tmp",
	}

	if err := cfg.Validate(); err == nil {
		t.Error("Expected validation error for CPUThreshold > 100")
	}
}

func TestValidateInvalidDuration(t *testing.T) {
	cfg := Config{
		CPUThreshold: 20.0,
		IdleDuration: -10, // Invalid: negative
		CheckInterval: 5,
		Command: "/bin/echo",
		WorkingDir: "/tmp",
	}

	if err := cfg.Validate(); err == nil {
		t.Error("Expected validation error for negative IdleDuration")
	}
}

func TestValidateNoCommand(t *testing.T) {
	cfg := Config{
		CPUThreshold: 20.0,
		IdleDuration: 60,
		CheckInterval: 5,
		Command: "", // Invalid: required
		WorkingDir: "/tmp",
	}

	if err := cfg.Validate(); err == nil {
		t.Error("Expected validation error for empty Command")
	}
}

func TestValidateInvalidWorkingDir(t *testing.T) {
	cfg := Config{
		CPUThreshold: 20.0,
		IdleDuration: 60,
		CheckInterval: 5,
		Command: "/bin/echo",
		WorkingDir: "/nonexistent/path/that/should/not/exist",
	}

	if err := cfg.Validate(); err == nil {
		t.Error("Expected validation error for non-existent WorkingDir")
	}
}
