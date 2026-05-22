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

	const glassRadius float32 = 16.0
	const blurRadius int = 12        // Adjust for a wider/softer shadow spread
	const shadowIntensity uint8 = 35 // Scale up or down to darken/lighten

	// 1. Generate the true blurred shadow texture
	shadowImgRaw := GenerateBlurredShadow(int(dialogSize.Width), int(dialogSize.Height), glassRadius, blurRadius, shadowIntensity)
	glassShadow := canvas.NewImageFromImage(shadowImgRaw)
	glassShadow.FillMode = canvas.ImageFillStretch

	// 2. Define the offset padding needed to prevent shadow clipping
	shadowPad := float32(blurRadius * 2)

	// 3. Build your glass pane stack (Without the shadow inside it)
	glassTint := canvas.NewRectangle(color.NRGBA{R: 0, G: 0, B: 0, A: 15}) // Bumped up slightly to pop from shadow
	glassTint.CornerRadius = glassRadius

	glassBorder := canvas.NewRectangle(color.Transparent)
	glassBorder.StrokeColor = color.NRGBA{R: 255, G: 255, B: 255, A: 60}
	glassBorder.StrokeWidth = 1.0
	glassBorder.CornerRadius = glassRadius

	glassContainer := container.NewStack(
		CreateBlurredBackground(blurredBg, dialogPos, dialogSize, canvasSize),
		glassTint,
		glassBorder,
		container.NewPadded(content),
	)

	themedContent := container.NewThemeOverride(glassContainer, &glassTheme{Theme: theme.DefaultTheme()})
	themedContent.Resize(dialogSize)

	// Position the glass pane strictly inside the center of your overlay frame
	themedContent.Move(fyne.NewPos(shadowPad, shadowPad))

	// 4. Position and size the shadow background to extend cleanly beyond the dialog borders
	glassShadow.Resize(fyne.NewSize(dialogSize.Width+(shadowPad*2), dialogSize.Height+(shadowPad*2)))
	glassShadow.Move(fyne.NewPos(0, 0))

	// 5. Wrap them together in a canvas that accounts for the combined size
	dialogWrapper := container.NewWithoutLayout(glassShadow, themedContent)
	dialogWrapper.Resize(fyne.NewSize(dialogSize.Width+(shadowPad*2), dialogSize.Height+(shadowPad*2)))

	// Shift the wrapper coordinate backward by the padding size so the dialog hits your exact original target coordinates
	dialogWrapper.Move(fyne.NewPos(dialogPos.X-shadowPad, dialogPos.Y-shadowPad))

	customOverlay := container.NewWithoutLayout(dialogWrapper)
	window.Canvas().Overlays().Add(customOverlay)

	// 3. Update the Cancel button to close the overlay instead of exiting the app (optional, but good practice)
	// You'll need to update the button definition in your `content` VBox to do this:
	// widget.NewButton("Cancel", func() {
	//     window.Canvas().Overlays().Remove(customOverlay)
	// })

	// 4. Add it directly to the window overlays, bypassing widget.PopUp entirely
	window.Canvas().Overlays().Add(customOverlay)
}
