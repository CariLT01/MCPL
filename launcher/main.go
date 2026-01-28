package main

import (
	"bufio"
	"encoding/base64"
	"errors"
	"fmt"
	"image/color"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"unicode/utf8"

	"crypto/ed25519"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"github.com/bodgit/sevenzip"
	"github.com/gofrs/flock"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/sys/windows"
)

var VERSION = "1.21.11 • Fabric 0.18.3"
var VERSION_DIR_STRING = "12111_FA0183"
var LAUNCHER_VERSION = "L1.1"

// REMEMBER TO CHANGE FOR TOKENS
var PUBLIC_KEY = "7MIyc6g3LVbRU1mvqy+qZKqn3DT7cerlu9jAMJg17/M="

func checkToken() {
	pubByes, err := base64.StdEncoding.DecodeString(PUBLIC_KEY)
	if err != nil {
		os.Exit(1)
	}
	pubKey := ed25519.PublicKey(pubByes)
	token, err := jwt.Parse(ISSUED_TOKEN, func(t *jwt.Token) (interface{}, error) {
		// Ensure algorithm is EdDSA
		if t.Method.Alg() != jwt.SigningMethodEdDSA.Alg() {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return pubKey, nil
	})
	if err != nil || !token.Valid {
		os.Exit(1) // silently fail
	}
}

func TruncateMiddle(s string, maxLen int) string {
	// 1. Check if truncation is even necessary
	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}

	// 2. Handle edge case: maxLen is too small for "..."
	if maxLen <= 3 {
		return "..."
	}

	runes := []rune(s)
	// Calculate how many characters to keep on each side
	// (maxLen - 3 for the dots)
	sideLen := (maxLen - 3) / 2

	start := string(runes[:sideLen])
	end := string(runes[len(runes)-sideLen:])

	return start + "..." + end
}

func updateProgressBar(app fyne.App, progressBar *canvas.Rectangle, value float32) {
	// We send the work to the Main UI Thread
	app.Driver().DoFromGoroutine(func() {
		progressBar.Resize(fyne.NewSize(630*value, 5))
		progressBar.Refresh()
	}, false)
}

func updateProgressBarText(app fyne.App, statusText *canvas.Text, value string) {
	app.Driver().DoFromGoroutine(func() {
		statusText.Text = value
		statusText.Refresh()
	}, false)
}

func extractFiles(app fyne.App, progressBar *canvas.Rectangle, statusText *canvas.Text) error {
	updateProgressBarText(app, statusText, "Unpacking...")

	src, err := installationZipFile.Open("data.7z")

	if err != nil {
		return err
	}

	defer src.Close()

	exePath, _ := os.Executable()
	installPath := filepath.Join(filepath.Dir(exePath), "tmp.7z")

	dst, err := os.Create(installPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	info, _ := src.Stat()
	totalSize := info.Size()

	buf := make([]byte, 32*1024*1024)
	var written int64

	for {
		n, err := src.Read(buf)
		if n > 0 {
			dst.Write(buf[:n])
			written += int64(n)
			updateProgressBar(app, progressBar, float32(written)/float32(totalSize))
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	updateProgressBarText(app, statusText, "Unpacked")

	return nil
}

func formatSeconds(seconds int) string {
	minutes := seconds / 60
	secs := seconds % 60
	return fmt.Sprintf("%02d:%02d", minutes, secs)
}

func unzip7zWithProgress(app fyne.App, src7z string, destDir string, progressBar *canvas.Rectangle, statusText *canvas.Text) error {
	// 1. Initial UI State
	updateProgressBarText(app, statusText, "Opening 7z archive...")
	updateProgressBar(app, progressBar, 0)

	r, err := sevenzip.OpenReader(src7z)
	if err != nil {
		return err
	}
	defer r.Close()

	// 2. Calculate total size for progress percentage
	var totalBytes int64
	for _, f := range r.File {
		totalBytes += int64(f.UncompressedSize)
	}

	var processedBytes int64

	// FIX: Separate timers for Bar and Text to prevent "resetting" each other
	lastBarUpdate := time.Now()

	timeStart := time.Now()

	createdDirs := make(map[string]bool)
	destDirClean := filepath.Clean(destDir)
	copyBuf := make([]byte, 1024*1024) // 1MB buffer

	for _, f := range r.File {
		path := filepath.Join(destDir, f.Name)

		// Security: ZipSlip check
		if !strings.HasPrefix(filepath.Clean(path), destDirClean+string(os.PathSeparator)) {
			continue
		}

		// Optimization: Directory Handling
		parentDir := filepath.Dir(path)
		if !createdDirs[parentDir] {
			os.MkdirAll(parentDir, 0755)
			createdDirs[parentDir] = true
		}

		if f.FileInfo().IsDir() {
			continue
		}

		// FIX: Update text immediately when a new file starts
		// This ensures the filename changes even if the file is tiny.

		// Extraction logic
		rc, err := f.Open()
		if err != nil {
			return err
		}

		fDest, err := os.Create(path)
		if err != nil {
			rc.Close()
			return err
		}

		// Optimization: Buffered Writer
		bufferedWriter := bufio.NewWriterSize(fDest, 1024*1024)

		// Manual copy loop
		for {
			n, readErr := rc.Read(copyBuf)
			if n > 0 {
				if _, writeErr := bufferedWriter.Write(copyBuf[:n]); writeErr != nil {
					fDest.Close()
					rc.Close()
					return writeErr
				}
				processedBytes += int64(n)

				// Throttled Progress Bar Update (every 50ms)
				// We don't throttle the text here because we update it once per file above
				if time.Since(lastBarUpdate) > 50*time.Millisecond {
					progress := float32(processedBytes) / float32(totalBytes)
					updateProgressBar(app, progressBar, progress)
					lastBarUpdate = time.Now()

					elapsed := time.Since(timeStart).Seconds()

					estimatedTimeRemaining := (1 - progress) * float32(elapsed) / progress
					formattedTimeRemaining := formatSeconds(int(math.Round(float64(estimatedTimeRemaining))))

					updateProgressBarText(app, statusText, "Extracting "+" ("+formattedTimeRemaining+" remaining): "+TruncateMiddle(f.Name, 32))
				}
			}
			if readErr == io.EOF {
				break
			}
			if readErr != nil {
				fDest.Close()
				rc.Close()
				return readErr
			}
		}

		bufferedWriter.Flush()
		fDest.Close()
		rc.Close()
	}

	// Final UI Update
	updateProgressBarText(app, statusText, "Extraction complete!")
	updateProgressBar(app, progressBar, 1.0)

	return nil
}

func launchGame(username string) {
	exePath, _ := os.Executable()
	gameDir := filepath.Join(filepath.Dir(exePath), VERSION_DIR_STRING)
	wrapperPath := filepath.Join(gameDir, "launcher.exe")

	cmd := exec.Command(wrapperPath, username, ISSUED_TOKEN)
	cmd.Dir = gameDir

	// KEY FIX: Manually set the I/O streams.
	// Even if the GUI doesn't have a console, setting these tells Windows
	// to connect the new console's buffers to the process streams.
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	cmd.SysProcAttr = &windows.SysProcAttr{
		CreationFlags: windows.CREATE_NEW_CONSOLE,
	}

	err := cmd.Start()
	if err != nil {
		fmt.Println("Failed to launch helper:", err)
		return
	}

	// Hand-off complete.
	time.Sleep(200 * time.Millisecond)
	os.Exit(0)
}

func askAndLaunch(window fyne.Window) {
	// 1. Create the input field
	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Username")

	// 2. Create the Dialog
	inputDialog := dialog.NewForm("Username (Offline): ", "Confirm", "Cancel", []*widget.FormItem{
		{Text: "Username", Widget: usernameEntry},
	}, func(confirm bool) {
		if confirm {
			// User clicked "Confirm"
			launchGame(usernameEntry.Text)
		} else {
			os.Exit(0)
		}
	}, window)

	// 3. Show it
	inputDialog.Resize(fyne.NewSize(400, 250))
	inputDialog.Show()
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !errors.Is(err, os.ErrNotExist)
}

func setup(app fyne.App, progressBar *canvas.Rectangle, statusText *canvas.Text) {

	exePath, _ := os.Executable()
	baseDir := filepath.Dir(exePath)
	gameDir := filepath.Join(baseDir, VERSION_DIR_STRING)

	modsDir := filepath.Join(gameDir, "mods")
	launcherBatch := filepath.Join(gameDir, "launch.bat")
	launcherExe := filepath.Join(gameDir, "launcher.exe")

	fileLock := tryLock(gameDir)

	// Delete mods directory
	if fileExists(modsDir) {
		os.RemoveAll(modsDir)
	}

	if fileExists(launcherBatch) {
		os.Remove(launcherBatch)
	}

	if fileExists(launcherExe) {
		os.Remove(launcherExe)
	}

	extractFiles(app, progressBar, statusText)

	tmpPath := filepath.Join(baseDir, "tmp.7z")

	unzip7zWithProgress(app, tmpPath, gameDir, progressBar, statusText)
	os.Remove("tmp.7z")
	fileLock.Unlock()
	markInstallationComplete(gameDir)
}

func showError(title string, message string) {
	titleText, _ := windows.UTF16FromString(title)
	messageText, _ := windows.UTF16FromString(message)

	windows.MessageBox(0, &messageText[0], &titleText[0], 0)
}

func getInstallationHash() string {
	return LAUNCHER_VERSION + "-" + VERSION_DIR_STRING
}

func markInstallationComplete(path string) {
	markerPath := filepath.Join(path, "INSTALLATION_COMPLETE")

	f, err := os.Create(markerPath)
	if err != nil {
		return
	}

	defer f.Close()

	content := getInstallationHash()

	_, err = f.WriteString(content)
	if err != nil {
		return
	}
}

func openMarkerFileAndCheckVersion(path string) bool {
	content, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	if string(content) == getInstallationHash() {
		return true
	}
	return false
}

func tryLock(path string) *flock.Flock {

	lockPath := filepath.Join(path, "LOCK.lock")
	fileLock := flock.New(lockPath)

	os.MkdirAll(path, 0755)

	locked, err := fileLock.TryLock()

	if err != nil {
		showError("Locking failed", "Failed to lock the directory. Is another instance of the installer running?")
		os.Exit(1)
	}

	if !locked {
		showError("More than one instance detected", "Another instance of Minecraft Portable is installing in this directory. Please close this instance and wait for the other one to finish.")
		os.Exit(1)
	}

	return fileLock
}

func initializeApp(app fyne.App, window fyne.Window, progressBar *canvas.Rectangle, statusText *canvas.Text) {
	time.Sleep(200 * time.Millisecond)

	exePath, _ := os.Executable()
	baseDir := filepath.Dir(exePath)
	gameDir := filepath.Join(baseDir, VERSION_DIR_STRING)

	brokenInstallation := false
	installationFound := false

	if fileExists(gameDir) {

		lock := tryLock(gameDir)
		lock.Unlock()

		installationFound = true
		markerPath := filepath.Join(gameDir, "INSTALLATION_COMPLETE")
		if fileExists(markerPath) {
			if openMarkerFileAndCheckVersion(markerPath) == true {
				brokenInstallation = false
			} else {
				showError("Version mismatch", "This installation is either outdated or too new for this launcher (most likely outdated). It will be reinstalled to match this launcher's version. Your saves and configuration will not be deleted. However, the MODS folder will be deleted and replaced, so please backup any additional mods you installed.")
				brokenInstallation = true
			}
		} else {
			brokenInstallation = true
			showError("Incomplete Installation", "This installation is incomplete. It will be reinstalled. Your saves and configuration will not be deleted. However, the MODS folder will be deleted and replaced, so please backup any additional mods you installed.")
		}
	} else {
		installationFound = false
		brokenInstallation = false
	}

	lock := tryLock(gameDir)
	lock.Unlock()

	if brokenInstallation == false && installationFound == true {

		askAndLaunch(window)
	} else {
		setup(app, progressBar, statusText)
		askAndLaunch(window)
	}
}

func main() {

	checkToken()

	fmt.Println("Hello")

	launcherApp := app.New()

	drv := launcherApp.Driver()

	if drv, ok := drv.(desktop.Driver); ok {
		launcherWindow := drv.CreateSplashWindow()
		launcherWindow.SetTitle("Minecraft Portable")

		background := canvas.NewImageFromResource(resourceLauncherbackgroundPng)
		background.FillMode = canvas.ImageFillContain
		background.Resize(fyne.NewSize(640, 400))

		softwareTitleText := canvas.NewText("Minecraft Portable", color.Black)
		softwareTitleText.TextStyle = fyne.TextStyle{Bold: true}
		softwareTitleText.TextSize = 48
		softwareTitleText.Move(fyne.NewPos(10, 20))

		softwareDetails := canvas.NewText("Java Edition • Game Launcher", color.Black)
		softwareDetails.TextSize = 18
		softwareDetails.Move(fyne.NewPos(12, 85))

		softwareVersion := canvas.NewText("Version: "+VERSION, color.NRGBA{R: 0, G: 0, B: 0, A: 128})
		softwareVersion.TextSize = 16
		softwareVersion.Move(fyne.NewPos(12, 230))
		licenseLabel := canvas.NewText("Personal Copy — Avoid Sharing", color.Black)
		licenseLabel.TextStyle = fyne.TextStyle{Bold: false}
		licenseLabel.TextSize = 16
		licenseLabel.Move(fyne.NewPos(12, 250))

		multiplayerNote := canvas.NewText("New: synchronize worlds with WorldSync", color.NRGBA{R: 0, G: 0, B: 0, A: 80})
		multiplayerNote.TextSize = 12
		multiplayerNote.Move(fyne.NewPos(12, 300))

		statusText := canvas.NewText("Launching...", color.Black)
		statusText.TextSize = 12
		statusText.Move(fyne.NewPos(5, 373))

		progressBarBackgroundRect := canvas.NewRectangle(color.NRGBA{R: 0, G: 0, B: 0, A: 25})
		progressBarBackgroundRect.CornerRadius = 9999999
		progressBarBackgroundRect.Resize(fyne.NewSize(630, 5))
		progressBarBackgroundRect.Move(fyne.NewPos(5, 390))

		progressBarValueRect := canvas.NewRectangle(color.Black)
		progressBarValueRect.Resize(fyne.NewSize(0, 5))
		progressBarValueRect.CornerRadius = 9999999
		progressBarValueRect.Move(fyne.NewPos(5, 390))

		go initializeApp(launcherApp, launcherWindow, progressBarValueRect, statusText)

		content := container.NewWithoutLayout(background, softwareTitleText, softwareDetails, softwareVersion, licenseLabel, multiplayerNote, statusText, progressBarBackgroundRect, progressBarValueRect)

		launcherWindow.SetContent(content)
		launcherWindow.Resize(fyne.NewSize(640, 400))
		launcherWindow.ShowAndRun()
	}

}
