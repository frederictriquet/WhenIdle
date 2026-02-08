package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>

void hideFromDock() {
	dispatch_async(dispatch_get_main_queue(), ^{
		[NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];
	});
}
*/
import "C"

// HideFromDock sets the macOS activation policy to NSApplicationActivationPolicyAccessory.
// This prevents the app from appearing in the Dock and Cmd-Tab switcher, making it
// a background-only (menu bar) application.
//
// Must be called after the Fyne app is created, typically via SetOnStarted lifecycle hook,
// to ensure NSApp is initialized. Uses dispatch_async to guarantee main thread execution.
func HideFromDock() {
	C.hideFromDock()
}
