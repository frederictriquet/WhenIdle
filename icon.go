package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"

	"fyne.io/fyne/v2"
)

// iconSize is the recommended size for macOS menu bar icons (22x22 pixels).
// Using standard size ensures proper display on Retina and non-Retina displays.
const iconSize = 22

// Tray icon states: a simple filled circle with different colors.
var (
	iconDisabled = generateCircleIcon("disabled", color.NRGBA{R: 140, G: 140, B: 140, A: 255}) // gray
	iconIdle     = generateCircleIcon("idle", color.NRGBA{R: 80, G: 180, B: 80, A: 255})       // green
	iconRunning  = generateCircleIcon("running", color.NRGBA{R: 50, G: 140, B: 240, A: 255})   // blue
	iconPaused   = generateCircleIcon("paused", color.NRGBA{R: 240, G: 180, B: 40, A: 255})    // orange
)

// generateCircleIcon creates a simple filled circle PNG icon as a Fyne resource.
func generateCircleIcon(name string, c color.Color) fyne.Resource {
	img := image.NewNRGBA(image.Rect(0, 0, iconSize, iconSize))

	cx, cy := iconSize/2, iconSize/2
	radius := iconSize/2 - 2

	for y := range iconSize {
		for x := range iconSize {
			dx := x - cx
			dy := y - cy
			if dx*dx+dy*dy <= radius*radius {
				img.Set(x, y, c)
			}
		}
	}

	var buf bytes.Buffer
	_ = png.Encode(&buf, img)

	return fyne.NewStaticResource("tray-"+name+".png", buf.Bytes())
}
