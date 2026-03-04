package launch

import (
	"fmt"
	"mc-portable-launcher/src/config"
	"mc-portable-launcher/src/data"
	"mc-portable-launcher/src/ux"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"golang.org/x/sys/windows"
)

func LaunchGame(username string) {
	exePath, _ := os.Executable()
	gameDir := filepath.Join(filepath.Dir(exePath), config.VERSION_DIR_STRING)
	wrapperPath := filepath.Join(gameDir, "launcher.exe")

	cmd := exec.Command(wrapperPath, username, data.ISSUED_TOKEN)
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
		ux.ShowErrorLog(err, "launch instance")
		return
	}

	// Hand-off complete.
	time.Sleep(200 * time.Millisecond)
	os.Exit(0)
}
