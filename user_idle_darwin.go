//go:build darwin

package main

/*
#cgo LDFLAGS: -framework CoreGraphics
#include <CoreGraphics/CoreGraphics.h>

// userIdleSeconds returns the number of seconds since the last keyboard or mouse event.
// This is a thin CGo wrapper around CGEventSourceSecondsSinceLastEventType.
double userIdleSeconds() {
	return CGEventSourceSecondsSinceLastEventType(
		kCGEventSourceStateCombinedSessionState,
		kCGAnyInputEventType
	);
}
*/
import "C"

// UserIdleSeconds returns the number of seconds since the last keyboard or mouse event.
//
// Uses macOS Quartz Event Services (CoreGraphics framework) to query the system
// event source for the time elapsed since any input event (keyboard, mouse, trackpad).
//
// No accessibility permissions are required for this API.
//
// Returns:
//   - The number of seconds (as float64) since the last input event
//   - Returns 0.0 immediately after any keyboard or mouse activity
//
// Example:
//
//	seconds := UserIdleSeconds()
//	if seconds > 300 {
//	    fmt.Println("User has been idle for more than 5 minutes")
//	}
func UserIdleSeconds() float64 {
	return float64(C.userIdleSeconds())
}
