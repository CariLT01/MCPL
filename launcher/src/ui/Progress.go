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
var k float32 = 0.95

// Tracks continuous time for the indeterminate loop
var indeterminateTime float32 = 0.0

const maxTrackWidth float32 = 630.0
const trackXOffset float32 = 5.0
const trackYPos float32 = 390.0

// Ideal width during the middle of the track
const minBarWidth float32 = 32.0

// preciseMaterialEase approximates standard cubic-bezier(0.4, 0.0, 0.2, 1.0)
func preciseMaterialEase(t float32) float32 {
	if t <= 0 {
		return 0
	}
	if t >= 1 {
		return 1
	}
	return t * t * (2.5 - 1.5*t)
}

func ProgressBarSizeAnimationCallback(progress float32) {
	if globalPg == nil {
		return
	}

	targetFloat := math.Float32frombits(targetPercentage.Load())

	// CASE 1: Symmetric Contained Indeterminate Mode
	if targetFloat == 0 {
		// Pacing velocity. Adjust this to change the loop duration.
		indeterminateTime += 0.012
		cycleProgress := float32(math.Mod(float64(indeterminateTime), 1.0))

		// 1. Calculate the Leading Edge (Head)
		headFactor := preciseMaterialEase(cycleProgress * 1.5)
		headX := trackXOffset + (headFactor * maxTrackWidth)

		// 2. Calculate the Trailing Edge (Tail)
		tailFactor := preciseMaterialEase((cycleProgress - 0.15) * 1.4)
		tailX := trackXOffset + (tailFactor * maxTrackWidth)

		// 3. Dynamic Scale Window (Symmetric Entry and Exit)
		// We dynamically restrict the max allowed minimum width depending on where
		// we are in the cycle so it can't snap open on the left or stick on the right.
		allowedMinWidth := minBarWidth

		if cycleProgress < 0.25 {
			// Entrance Phase: Scale up from 0 to minBarWidth over the first 25%
			entranceFactor := cycleProgress / 0.25
			allowedMinWidth = minBarWidth * entranceFactor
		} else if cycleProgress > 0.75 {
			// Exit Phase: Scale down from minBarWidth to 0 over the last 25%
			exitFactor := (1.0 - cycleProgress) / 0.25
			allowedMinWidth = minBarWidth * exitFactor
		}

		// 4. Enforce Boundaries
		if tailX < trackXOffset {
			tailX = trackXOffset
		}
		if headX > trackXOffset+maxTrackWidth {
			headX = trackXOffset + maxTrackWidth
		}

		currentWidth := headX - tailX

		// Clamp the width to our dynamic limits and adjust position anchors
		if currentWidth < allowedMinWidth {
			currentWidth = allowedMinWidth
			if headX >= trackXOffset+maxTrackWidth {
				// Pin to right edge when collapsing outward
				tailX = (trackXOffset + maxTrackWidth) - currentWidth
			} else {
				// Pin to left edge / tail position when scaling inward
				headX = tailX + currentWidth
			}
		}

		// --- GEOMETRY UPDATE ---
		globalPg.Resize(fyne.NewSize(currentWidth, 5))
		globalPg.Move(fyne.NewPos(tailX, trackYPos))
		globalPg.Refresh()

		currentValueTween = 0.0
		return
	}

	// CASE 2: Determinate Mode (Target > 0)
	if globalPg.Position().X != trackXOffset {
		globalPg.Move(fyne.NewPos(trackXOffset, trackYPos))
	}

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
