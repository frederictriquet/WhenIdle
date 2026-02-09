package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	configPath := flag.String("config", "", "Path to configuration file (JSON)")
	guiMode := flag.Bool("gui", false, "Run with system tray GUI")
	flag.Parse()

	// GUI mode
	if *guiMode {
		RunGUI(*configPath) // Blocks on main thread
		return
	}

	// CLI mode (original behavior)
	if *configPath == "" {
		fmt.Fprintln(os.Stderr, "Usage: whenidle --config <path-to-config.json>")
		fmt.Fprintln(os.Stderr, "   or: whenidle --gui [--config <path>]")
		os.Exit(1)
	}

	cfg, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("[FATAL] %v", err)
	}

	// Setup log file if configured
	if cfg.LogFile != "" {
		f, err := os.OpenFile(cfg.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("[FATAL] Cannot open log file: %v", err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	log.Println("[INFO] WhenIdle starting")
	log.Printf("[INFO] Config: threshold=%.1f%%, idle_duration=%ds, check_interval=%ds",
		cfg.CPUThreshold, cfg.IdleDuration, cfg.CheckInterval)
	log.Printf("[INFO] Task: %s %v (workdir: %s)", cfg.Command, cfg.Args, cfg.WorkingDir)

	runner := NewTaskRunner(cfg)
	monitor := NewMonitor(cfg, runner.OnIdle, runner.OnBusy, runner.State)

	// Start monitoring in background
	ctx, cancel := context.WithCancel(context.Background())
	go monitor.Start(ctx)

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	sig := <-sigCh
	log.Printf("[INFO] Received %v, shutting down...", sig)

	// Stop monitoring
	cancel()

	// Stop the task if running
	runner.Stop()
	runner.WaitDone()

	log.Println("[INFO] WhenIdle stopped")
}
