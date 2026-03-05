package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// IdleMode specifies the method used to detect system idleness.
type IdleMode string

const (
	// IdleModeCPU detects idleness based on CPU usage falling below a threshold.
	// The system is considered idle when CPU % < cpu_threshold for idle_duration seconds.
	IdleModeCPU IdleMode = "cpu"

	// IdleModeUserIdle detects idleness based on the absence of keyboard/mouse events.
	// The system is considered idle when no input events occur for idle_duration seconds.
	// This mode ignores CPU usage and uses macOS Quartz Event Services.
	IdleModeUserIdle IdleMode = "user_idle"
)

const (
	defaultCPUThreshold  = 15.0 // percent
	defaultIdleDuration  = 120  // seconds
	defaultCheckInterval = 5    // seconds
)

type Config struct {
	CPUThreshold  float64  `json:"cpu_threshold"`
	IdleDuration  int      `json:"idle_duration"`
	CheckInterval int      `json:"check_interval"`
	Command       string   `json:"command"`
	Args          []string `json:"args"`
	WorkingDir    string   `json:"working_dir"`
	LogFile       string   `json:"log_file"`
	Restart       bool     `json:"restart"`
	IdleMode      IdleMode `json:"idle_mode"`
}

func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("cannot read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("cannot parse config file: %w", err)
	}

	cfg.ApplyDefaults()

	if err := cfg.Validate(); err != nil {
		return Config{}, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

func (c *Config) ApplyDefaults() {
	if c.CPUThreshold == 0 {
		c.CPUThreshold = defaultCPUThreshold
	}
	if c.IdleDuration == 0 {
		c.IdleDuration = defaultIdleDuration
	}
	if c.CheckInterval == 0 {
		c.CheckInterval = defaultCheckInterval
	}
	if c.IdleMode == "" {
		c.IdleMode = IdleModeCPU
	}
}

func (c Config) Validate() error {
	if c.CPUThreshold <= 0 || c.CPUThreshold > 100 {
		return fmt.Errorf("cpu_threshold must be between 0 and 100, got %.1f", c.CPUThreshold)
	}
	if c.IdleDuration <= 0 {
		return fmt.Errorf("idle_duration must be positive, got %d", c.IdleDuration)
	}
	if c.CheckInterval <= 0 {
		return fmt.Errorf("check_interval must be positive, got %d", c.CheckInterval)
	}
	if c.IdleMode != IdleModeCPU && c.IdleMode != IdleModeUserIdle {
		return fmt.Errorf("idle_mode must be %q or %q, got %q", IdleModeCPU, IdleModeUserIdle, c.IdleMode)
	}
	if c.Command == "" {
		return fmt.Errorf("command is required")
	}
	if c.WorkingDir != "" {
		info, err := os.Stat(c.WorkingDir)
		if err != nil {
			return fmt.Errorf("working_dir %q does not exist: %w", c.WorkingDir, err)
		}
		if !info.IsDir() {
			return fmt.Errorf("working_dir %q is not a directory", c.WorkingDir)
		}
	}
	return nil
}

// SaveConfig writes the config to the specified JSON file with pretty-printing.
func SaveConfig(path string, cfg Config) error {
	data, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return fmt.Errorf("cannot marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("cannot write config file: %w", err)
	}

	return nil
}
