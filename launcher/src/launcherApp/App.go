package launcherApp

import (
	"mc-portable-launcher/src/appInit"
	"mc-portable-launcher/src/data"
	"mc-portable-launcher/src/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
)

type LauncherApp struct {
	InstallationContainer *ui.InstallationContainer
	Background            *canvas.Image
	Application           fyne.App

	Window fyne.Window
}

func NewLauncherApp() *LauncherApp {
	return &LauncherApp{}
}

func (launcher *LauncherApp) Initialize() {
	launcher.Application = app.New()

	drv := launcher.Application.Driver()

	if drv, ok := drv.(desktop.Driver); ok {
		launcher.Window = drv.CreateSplashWindow()
		launcher.Window.SetTitle("MCPL")

		launcher.Background = canvas.NewImageFromResource(data.ResourceLauncherbackgroundPng)
		launcher.Background.FillMode = canvas.ImageFillContain
		launcher.Background.Resize(fyne.NewSize(640, 400))

		// initialize main container
		launcher.InstallationContainer = ui.NewInstallationContainer(launcher.Background)
		launcher.InstallationContainer.Initialize()

		// installation container

		/* ***** MAIN MENU CONTAINER */

		launcher.Window.SetContent(launcher.InstallationContainer.Content)
		launcher.Window.Resize(fyne.NewSize(640, 400))

	}
}

func (launcher *LauncherApp) Run() {
	launcher.InstallationContainer.StartAnimations()
	go appInit.InitializeApp(launcher.Application, launcher.Window, launcher.InstallationContainer.ProgressBarValueRect, launcher.InstallationContainer.StatusText)
	launcher.Window.ShowAndRun()
}
