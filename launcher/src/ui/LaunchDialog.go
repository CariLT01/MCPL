package ui

import (
	"image/color"
	"mc-portable-launcher/src/launch"
	"os"

	_ "image/png"

	"image"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func AskAndLaunch(window fyne.Window, blurredBg image.Image) {
	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Username")

	dialogSize := fyne.NewSize(400, 220)
	dialogPos := fyne.NewPos(
		(window.Canvas().Size().Width-dialogSize.Width)/2,
		(window.Canvas().Size().Height-dialogSize.Height)/2,
	)

	// Inner content
	content := container.NewVBox(
		widget.NewLabelWithStyle("Username:", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		usernameEntry,
		layout.NewSpacer(),
		container.NewHBox(
			layout.NewSpacer(),
			widget.NewButton("Play", func() { launch.LaunchGame(usernameEntry.Text) }),
			widget.NewButton("Cancel", func() { os.Exit(0) }),
			layout.NewSpacer(),
		),
	)

	// THE "GLASS" STACK
	// We use a MaxLayout to ensure the background fills the entire PopUp area
	canvasSize := window.Canvas().Size()

	// THE FIX: If the UI thread hasn't finished the initial layout pass,
	// canvasSize will be 0x0. Fallback to the splash screen default size.
	if canvasSize.Width <= 0 || canvasSize.Height <= 0 {
		canvasSize = fyne.NewSize(640, 400)
	}

	// THE "GLASS" STACK
	glassContainer := container.NewStack(
		// Added canvasSize here ---------------------------------------v
		CreateBlurredBackground(blurredBg, dialogPos, dialogSize, canvasSize),
		canvas.NewRectangle(color.NRGBA{R: 255, G: 255, B: 255, A: 25}),
		container.NewPadded(content),
	)

	// APPLY CUSTOM THEME: This is the magic part
	// Wrap the glassContainer in our custom theme override
	themedContent := container.NewThemeOverride(glassContainer, &glassTheme{Theme: theme.DefaultTheme()})

	// 1. Manually set the size and position on your themed content
	themedContent.Resize(dialogSize)
	themedContent.Move(dialogPos)

	// 2. Wrap it in a container without a layout so it respects your exact X/Y coordinates
	customOverlay := container.NewWithoutLayout(themedContent)

	// 3. Update the Cancel button to close the overlay instead of exiting the app (optional, but good practice)
	// You'll need to update the button definition in your `content` VBox to do this:
	// widget.NewButton("Cancel", func() {
	//     window.Canvas().Overlays().Remove(customOverlay)
	// })

	// 4. Add it directly to the window overlays, bypassing widget.PopUp entirely
	window.Canvas().Overlays().Add(customOverlay)
}
