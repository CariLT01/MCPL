package appInit

import (
	"mc-portable-launcher/src/config"
	"mc-portable-launcher/src/data"
	"mc-portable-launcher/src/other"
	"mc-portable-launcher/src/setup"
	"mc-portable-launcher/src/ui"
	"mc-portable-launcher/src/utils"
	"mc-portable-launcher/src/ux"
	"os"
	"path/filepath"
	"time"

	_ "image/png"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

func InitializeApp(app fyne.App, window fyne.Window, progressBar *canvas.Rectangle, statusText *canvas.Text) {
	time.Sleep(200 * time.Millisecond)

	img, err := utils.ResourceToImage(data.ResourceLauncherbackgroundblurredPng)
	if err != nil {
		ux.ShowErrorLog(err, "decode blurred background")
	}

	exePath, _ := os.Executable()
	baseDir := filepath.Dir(exePath)
	gameDir := filepath.Join(baseDir, config.VERSION_DIR_STRING)

	brokenInstallation := false
	installationFound := false

	if utils.FileExists(gameDir) {

		lock := utils.TryLock(gameDir)
		lock.Unlock()

		installationFound = true
		markerPath := filepath.Join(gameDir, "INSTALLATION_COMPLETE")
		if utils.FileExists(markerPath) {
			if other.OpenMarkerFileAndCheckVersion(markerPath) == true {
				brokenInstallation = false
			} else {
				ux.ShowError("MCPL -- Version mismatch",
					`
This installation's launcher or game version does not match this one's version. It will be reinstalled to match this launcher's version. Your saves and configuration will be kept intact.

As a temporary solution to installation integrity, the contents of the following folders will be replaced:

  - Mods

Backup any important files if needed before pressing OK or closing this dialog.
If you wish to cancel, close the launcher window. Otherwise, the installation will proceed.
			`)
				brokenInstallation = true
			}
		} else {
			brokenInstallation = true
			ux.ShowError("MCPL -- Incomplete Installation",
				`
This installation is incomplete. It will be reinstalled. Your saves and configuration will be kept intact.

As a temporary solution to installation integrity, the contents of the following folders will be replaced:

  - Mods

Backup any important files if needed before pressing OK or closing this dialog.
If you wish to cancel, close the launcher window. Otherwise, the installation will proceed.
			`)
		}
	} else {
		installationFound = false
		brokenInstallation = false
	}

	lock := utils.TryLock(gameDir)
	lock.Unlock()

	if brokenInstallation == false && installationFound == true {

		ui.UpdateProgressBar(app, progressBar, 1.0)
		ui.AskAndLaunch(window, img)
	} else {
		ui.UpdateProgressBar(app, progressBar, 0.0)
		setup.Setup(app, progressBar, statusText)
		ui.UpdateProgressBar(app, progressBar, 1.0)
		ui.AskAndLaunch(window, img)
	}
}
