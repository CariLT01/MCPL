package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type glassTheme struct {
	fyne.Theme
}

func (m *glassTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	// This removes the "white border" / background of the PopUp
	case theme.ColorNameBackground:
		return color.Transparent
	// Semi-transparent background for Entry fields
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 255, G: 255, B: 255, A: 40}
	// Semi-transparent background for Buttons
	case theme.ColorNameButton:
		return color.NRGBA{R: 255, G: 255, B: 255, A: 60}
	// Make the cursor/text slightly more visible against the blur
	case theme.ColorNameForeground:
		return color.White
	case theme.ColorNamePlaceHolder:
		return color.NRGBA{R: 255, G: 255, B: 255, A: 150} // Slightly faded white
	}

	return m.Theme.Color(name, variant)
}
