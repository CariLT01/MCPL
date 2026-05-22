package ui

import (
	"math"
	"sync/atomic"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

var targetPercentage atomic.Uint32 = atomic.Uint32{}
var globalPg *canvas.Rectangle
var currentValueTween float32 = 0.0
var k float32 = 0.1

// Track the total loop progression to keep a smooth timeline for the indeterminate state
var indeterminateTime float32 = 0.0

const maxTrackWidth float32 = 630.0
const trackXOffset float32 = 5.0
const trackYPos float32 = 390.0

func ProgressBarSizeAnimationCallback(progress float32) {
	if globalPg == nil {
		return
	}

	targetFloat := math.Float32frombits(targetPercentage.Load())

	// CASE 1: Indeterminate Mode (Target is 0)
	if targetFloat == 0 {
		// Increment a running loop counter slightly each frame.
		// Since progress ticks 0.0 -> 1.0 periodically, we accumulate over time.
		indeterminateTime += 0.03
		if indeterminateTime > 2*math.Pi {
			indeterminateTime -= 2 * math.Pi
		}

		// 1. Give the bar a fixed modern width (e.g., 30% of the maximum track width)
		indeterminateWidth := maxTrackWidth * 0.3
		globalPg.Resize(fyne.NewSize(indeterminateWidth, 5))

		// 2. Map a sine wave (-1 to 1) to a clean 0 to 1 range
		// This creates a smooth "slow down at the edges" modern easing feel
		normalizedSine := (float32(math.Sin(float64(indeterminateTime))) + 1.0) / 2.0

		// 3. Calculate the available space to travel across
		travelDistance := maxTrackWidth - indeterminateWidth
		newX := trackXOffset + (normalizedSine * travelDistance)

		// 4. Update the object's position across the track
		globalPg.Move(fyne.NewPos(newX, trackYPos))
		globalPg.Refresh()

		// Reset the normal tween baseline to 0 so it's ready when loading starts
		currentValueTween = 0.0
		return
	}

	// CASE 2: Determinate Mode (Target > 0)
	// Force the position back to the beginning of the track if it was previously animating
	if globalPg.Position().X != trackXOffset {
		globalPg.Move(fyne.NewPos(trackXOffset, trackYPos))
	}

	// Your standard exponential decay easing
	currentValueTween = targetFloat + (currentValueTween-targetFloat)*k

	globalPg.Resize(fyne.NewSize(maxTrackWidth*currentValueTween, 5))
	globalPg.Refresh()
}

func UpdateProgressBar(app fyne.App, progressBar *canvas.Rectangle, value float32) {
	globalPg = progressBar
	targetPercentage.Store(math.Float32bits(value))
}

func UpdateProgressBarText(app fyne.App, statusText *canvas.Text, value string) {
	app.Driver().DoFromGoroutine(func() {
		statusText.Text = value
		statusText.Refresh()
	}, false)
}
