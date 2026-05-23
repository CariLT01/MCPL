package ui

import (
	"image/color"
	"mc-portable-launcher/src/config"
	"time"

	_ "image/png"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

type InstallationContainer struct {
	SoftwareTitleText *canvas.Text
	SoftwareDetails   *canvas.Text
	SoftwareVersion   *canvas.Text
	LicenseLabel      *canvas.Text
	MultiplayerNote   *canvas.Text
	FunFactText       *canvas.Text
	StatusText        *canvas.Text

	Background *canvas.Image

	// progress
	ProgressBarBackgroundRect *canvas.Rectangle
	ProgressBarValueRect      *canvas.Rectangle

	Content *fyne.Container

	// animations
	ProgressBarAnimation   *fyne.Animation
	OtherElementsAnimation *fyne.Animation
	TitleAnimation         *fyne.Animation
	DetailsAnimation       *fyne.Animation
	FrameAnimation         *fyne.Animation
}

func NewInstallationContainer(background *canvas.Image) *InstallationContainer {
	return &InstallationContainer{
		Background: background,
	}
}

func (c *InstallationContainer) InitializeContainer() {
	c.SoftwareTitleText = canvas.NewText("MCPL "+config.LAUNCHER_VERSION, color.White)
	c.SoftwareTitleText.TextStyle = fyne.TextStyle{Bold: true}
	c.SoftwareTitleText.TextSize = 48
	c.SoftwareTitleText.Move(fyne.NewPos(10, 20))
	c.SoftwareTitleText.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 0} // animate: opacity = 1 for 1 second
	c.SoftwareDetails = canvas.NewText("Java Edition • Game Launcher", color.White)
	c.SoftwareDetails.TextSize = 18
	c.SoftwareDetails.Move(fyne.NewPos(12, 85))
	c.SoftwareDetails.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 0}
	c.SoftwareVersion = canvas.NewText("Version: "+config.VERSION, color.NRGBA{R: 255, G: 255, B: 255, A: 0}) // animate to 128
	c.SoftwareVersion.TextSize = 16
	c.SoftwareVersion.Move(fyne.NewPos(12, 230))
	c.LicenseLabel = canvas.NewText("Personal Copy — Avoid Sharing", color.White)
	c.LicenseLabel.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 0} // animate to 255
	c.LicenseLabel.TextStyle = fyne.TextStyle{Bold: false}
	c.LicenseLabel.TextSize = 16
	c.LicenseLabel.Move(fyne.NewPos(12, 250))
	c.MultiplayerNote = canvas.NewText(config.LAUNCHER_CHANGELOG, color.NRGBA{R: 255, G: 255, B: 255, A: 0}) // animate to 80
	c.MultiplayerNote.TextSize = 12
	c.MultiplayerNote.Move(fyne.NewPos(12, 300))
	c.FunFactText = canvas.NewText("Fun fact: MCPL and its tooling has more than 10000 lines of code", color.NRGBA{R: 255, G: 255, B: 255, A: 0}) // animate to 80
	c.FunFactText.TextSize = 12
	c.FunFactText.Move(fyne.NewPos(12, 320))
	c.StatusText = canvas.NewText("Launching...", color.White)
	c.StatusText.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 0} // animate to 255
	c.StatusText.TextSize = 12
	c.StatusText.Move(fyne.NewPos(5, 373))
	c.ProgressBarBackgroundRect = canvas.NewRectangle(color.NRGBA{R: 255, G: 255, B: 255, A: 0}) // animate to 25
	c.ProgressBarBackgroundRect.CornerRadius = 9999999
	c.ProgressBarBackgroundRect.Resize(fyne.NewSize(630, 5))
	c.ProgressBarBackgroundRect.Move(fyne.NewPos(5, 390))
	c.ProgressBarValueRect = canvas.NewRectangle(color.White)
	c.ProgressBarValueRect.FillColor = color.NRGBA{R: 255, G: 255, B: 255, A: 0} // animate to 255
	c.ProgressBarValueRect.Resize(fyne.NewSize(0, 5))
	c.ProgressBarValueRect.CornerRadius = 9999999
	c.ProgressBarValueRect.Move(fyne.NewPos(5, 390))
	c.Content = container.NewWithoutLayout(c.Background, c.SoftwareTitleText, c.SoftwareDetails, c.SoftwareVersion, c.LicenseLabel, c.MultiplayerNote, c.FunFactText, c.StatusText, c.ProgressBarBackgroundRect, c.ProgressBarValueRect)
}

func (c *InstallationContainer) InitializeAnimations() {
	// progress bar animation
	c.FrameAnimation = fyne.NewAnimation(time.Second, ProgressBarSizeAnimationCallback)
	c.FrameAnimation.RepeatCount = fyne.AnimationRepeatForever
	c.FrameAnimation.Start()

	c.OtherElementsAnimation = fyne.NewAnimation(500*time.Millisecond, func(f float32) {
		c.SoftwareVersion.Color = color.NRGBA{R: 255, G: 255, B: 255, A: uint8(f * 128)}
		c.SoftwareVersion.Refresh()

		c.MultiplayerNote.Color = color.NRGBA{R: 255, G: 255, B: 255, A: uint8(f * 80)}
		c.MultiplayerNote.Refresh()

		c.FunFactText.Color = color.NRGBA{R: 255, G: 255, B: 255, A: uint8(f * 80)}
		c.FunFactText.Refresh()

		c.LicenseLabel.Color = color.NRGBA{R: 255, G: 255, B: 255, A: uint8(f * 255)}
		c.LicenseLabel.Refresh()
	})

	// version animation
	c.DetailsAnimation = fyne.NewAnimation(500*time.Millisecond, func(f float32) {
		c.SoftwareDetails.Color = color.NRGBA{R: 255, G: 255, B: 255, A: uint8(f * 255)}
		c.SoftwareDetails.Refresh()

		if f >= 1 {
			c.OtherElementsAnimation.Start()
		}
	})

	// progress animation
	c.ProgressBarAnimation = fyne.NewAnimation(500*time.Millisecond, func(f float32) {
		c.ProgressBarValueRect.FillColor = color.NRGBA{R: 255, G: 255, B: 255, A: uint8(f * 255)}
		c.ProgressBarBackgroundRect.FillColor = color.NRGBA{R: 255, G: 255, B: 255, A: uint8(f * 25)}
		c.StatusText.Color = color.NRGBA{R: 255, G: 255, B: 255, A: uint8(f * 255)}
	})

	// title animation
	c.TitleAnimation = fyne.NewAnimation(500*time.Millisecond, func(f float32) {
		c.SoftwareTitleText.Color = color.NRGBA{R: 255, G: 255, B: 255, A: uint8(f * 255)}
		c.SoftwareTitleText.Refresh()

		if f >= 0.5 {
			c.DetailsAnimation.Start()
			c.ProgressBarAnimation.Start()
		}
	})
}

func (c *InstallationContainer) Initialize() {
	c.InitializeContainer()
	c.InitializeAnimations()
}

func (c *InstallationContainer) StartAnimations() {
	c.TitleAnimation.Start()
}
