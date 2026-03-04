//go:build darwin

package main

import "testing"

func TestHideFromDock(t *testing.T) {
	// Should not panic — CGo call to NSApp hide-from-dock API
	HideFromDock()
}
