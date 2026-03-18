package main

import (
	"fmt"
	"image/color"
	"mc-portable-launcher/src/appInit"
	"mc-portable-launcher/src/config"
	"mc-portable-launcher/src/data"
	"mc-portable-launcher/src/exp"

	_ "image/png"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
)

func main() {

	exp.CheckToken()

	fmt.Println("Hello")

	launcherApp := app.New()

	drv := launcherApp.Driver()

	if drv, ok := drv.(desktop.Driver); ok {
		launcherWindow := drv.CreateSplashWindow()
		launcherWindow.SetTitle("MCPL")

		background := canvas.NewImageFromResource(data.ResourceLauncherbackgroundPng)
		background.FillMode = canvas.ImageFillContain
		background.Resize(fyne.NewSize(640, 400))

		softwareTitleText := canvas.NewText("MCPL "+config.LAUNCHER_VERSION, color.White)
		softwareTitleText.TextStyle = fyne.TextStyle{Bold: true}
		softwareTitleText.TextSize = 48
		softwareTitleText.Move(fyne.NewPos(10, 20))

		softwareDetails := canvas.NewText("Java Edition • Game Launcher", color.White)
		softwareDetails.TextSize = 18
		softwareDetails.Move(fyne.NewPos(12, 85))

		softwareVersion := canvas.NewText("Version: "+config.VERSION, color.NRGBA{R: 255, G: 255, B: 255, A: 128})
		softwareVersion.TextSize = 16
		softwareVersion.Move(fyne.NewPos(12, 230))
		licenseLabel := canvas.NewText("Personal Copy — Avoid Sharing", color.White)
		licenseLabel.TextStyle = fyne.TextStyle{Bold: false}
		licenseLabel.TextSize = 16
		licenseLabel.Move(fyne.NewPos(12, 250))

		multiplayerNote := canvas.NewText(config.LAUNCHER_CHANGELOG, color.NRGBA{R: 255, G: 255, B: 255, A: 80})
		multiplayerNote.TextSize = 12
		multiplayerNote.Move(fyne.NewPos(12, 300))

		funFactText := canvas.NewText("Fun fact: MCPL and its tooling has more than 10000 lines of code", color.NRGBA{R: 255, G: 255, B: 255, A: 80})
		funFactText.TextSize = 12
		funFactText.Move(fyne.NewPos(12, 320))

		statusText := canvas.NewText("Launching...", color.White)
		statusText.TextSize = 12
		statusText.Move(fyne.NewPos(5, 373))

		progressBarBackgroundRect := canvas.NewRectangle(color.NRGBA{R: 255, G: 255, B: 255, A: 25})
		progressBarBackgroundRect.CornerRadius = 9999999
		progressBarBackgroundRect.Resize(fyne.NewSize(630, 5))
		progressBarBackgroundRect.Move(fyne.NewPos(5, 390))

		progressBarValueRect := canvas.NewRectangle(color.White)
		progressBarValueRect.Resize(fyne.NewSize(0, 5))
		progressBarValueRect.CornerRadius = 9999999
		progressBarValueRect.Move(fyne.NewPos(5, 390))

		go appInit.InitializeApp(launcherApp, launcherWindow, progressBarValueRect, statusText)

		content := container.NewWithoutLayout(background, softwareTitleText, softwareDetails, softwareVersion, licenseLabel, multiplayerNote, funFactText, statusText, progressBarBackgroundRect, progressBarValueRect)

		launcherWindow.SetContent(content)
		launcherWindow.Resize(fyne.NewSize(640, 400))
		launcherWindow.ShowAndRun()
	}

}
