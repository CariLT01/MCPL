package setup

import (
	"fmt"
	"mc-portable-launcher/src/config"
	"mc-portable-launcher/src/data"
	"mc-portable-launcher/src/diff"
	"mc-portable-launcher/src/other"
	"mc-portable-launcher/src/ui"
	"mc-portable-launcher/src/utils"
	"mc-portable-launcher/src/ux"
	"os"
	"path/filepath"

	_ "image/png"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

func Setup(app fyne.App, progressBar *canvas.Rectangle, statusText *canvas.Text) {

	fmt.Println("read installation hashes")
	ui.UpdateProgressBarText(app, statusText, "Retrieving instance installation hash map...")
	installationHashes := diff.ReadInstallationHashes()
	fmt.Println("read current hashes")
	ui.UpdateProgressBarText(app, statusText, "Computing current installation hashes...")
	currentHashes := diff.ComputeCurrentHashes(app, statusText)
	fmt.Println("compute diff")
	ui.UpdateProgressBarText(app, statusText, "Performing differential integrity analysis...")
	needsDelete, needsAdd, upToDate := diff.ComputeDiffMaps(currentHashes, installationHashes)

	fmt.Println("merge")

	ui.UpdateProgressBarText(app, statusText, "Retrieving persistent installation files...")
	unskippableFiles := diff.ReadUnskippableFiles()

	exePath, _ := os.Executable()
	baseDir := filepath.Dir(exePath)
	gameDir := filepath.Join(baseDir, config.VERSION_DIR_STRING)

	utils.DeleteFiles(needsDelete, gameDir)

	modsDir := filepath.Join(gameDir, "mods")
	launcherBatch := filepath.Join(gameDir, "launch.bat")
	launcherExe := filepath.Join(gameDir, "launcher.exe")
	libsDir := filepath.Join(gameDir, "libraries")

	fileLock := utils.TryLock(gameDir)

	// Delete mods directory
	if utils.FileExists(modsDir) {
		os.RemoveAll(modsDir)
	}
	if utils.FileExists(libsDir) {
		// os.RemoveAll(libsDir)
	}

	if utils.FileExists(launcherBatch) {
		os.Remove(launcherBatch)
	}

	if utils.FileExists(launcherExe) {
		os.Remove(launcherExe)
	}

	if len(needsAdd) > 0 {
		fmt.Println("Missing assets: needs add")

		fileName := "bin/static.7z"
		fileNameTmp := "static.7z.tmp.7z"

		err := ExtractFiles(data.StaticZipFile, fileName, fileNameTmp, app, progressBar, statusText)
		if err != nil {
			ux.ShowErrorLog(err, "unpack files static")
		}

		tmpPath := filepath.Join(baseDir, fileNameTmp)

		err = Unzip7zWithProgress(app, tmpPath, gameDir, progressBar, statusText, needsAdd, len(upToDate))
		if err != nil {
			ux.ShowErrorLog(err, "extract files")
		}
		err = os.Remove(fileNameTmp)
		if err != nil {
			ux.ShowErrorLog(err, "remove tmp.7z")
		}
	}

	// dynamic data unpack

	fileNameDyn := "bin/dynamic.7z"
	fileNameDynTmp := "dynamic.7z.tmp.7z"

	err := ExtractFiles(data.DyamicZipFile, fileNameDyn, fileNameDynTmp, app, progressBar, statusText)
	if err != nil {
		ux.ShowErrorLog(err, "unpack files dynamic")
	}

	tmpPath := filepath.Join(baseDir, fileNameDynTmp)

	err = Unzip7zWithProgress(app, tmpPath, gameDir, progressBar, statusText, unskippableFiles, len(upToDate))
	if err != nil {
		ux.ShowErrorLog(err, "extract files")
	}
	err = os.Remove(fileNameDynTmp)
	if err != nil {
		ux.ShowErrorLog(err, "remove dyn.7z")
	}

	fileLock.Unlock()
	other.MarkInstallationComplete(gameDir)
}
