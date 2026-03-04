package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

func UpdateProgressBar(app fyne.App, progressBar *canvas.Rectangle, value float32) {
	// We send the work to the Main UI Thread
	app.Driver().DoFromGoroutine(func() {
		progressBar.Resize(fyne.NewSize(630*value, 5))
		progressBar.Refresh()
	}, false)
}

func UpdateProgressBarText(app fyne.App, statusText *canvas.Text, value string) {
	app.Driver().DoFromGoroutine(func() {
		statusText.Text = value
		statusText.Refresh()
	}, false)
}
