//go:build !darwin

package main

// HideFromDock is a no-op on non-macOS platforms.
func HideFromDock() {}
