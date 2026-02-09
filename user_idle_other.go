//go:build !darwin

package main

// UserIdleSeconds is not supported on non-macOS platforms and always returns 0.
func UserIdleSeconds() float64 {
	return 0
}
