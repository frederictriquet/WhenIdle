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

// HideFromDock sets the macOS activation policy to "accessory",
// which hides the app from the Dock and the Cmd-Tab switcher.
// Uses dispatch_async to ensure it runs on the main thread.
func HideFromDock() {
	C.hideFromDock()
}
