package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	CPUThreshold  float64  `json:"cpu_threshold"`
	IdleDuration  int      `json:"idle_duration"`
	CheckInterval int      `json:"check_interval"`
	Command       string   `json:"command"`
	Args          []string `json:"args"`
	WorkingDir    string   `json:"working_dir"`
	LogFile       string   `json:"log_file"`
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

	cfg.applyDefaults()

	if err := cfg.Validate(); err != nil {
		return Config{}, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

func (c *Config) applyDefaults() {
	if c.CPUThreshold == 0 {
		c.CPUThreshold = 15.0
	}
	if c.IdleDuration == 0 {
		c.IdleDuration = 120
	}
	if c.CheckInterval == 0 {
		c.CheckInterval = 5
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
