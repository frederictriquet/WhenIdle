//go:build !darwin

package monitor

// UserIdleSeconds is not supported on non-macOS platforms and always returns 0.
func UserIdleSeconds() float64 {
	return 0
}
