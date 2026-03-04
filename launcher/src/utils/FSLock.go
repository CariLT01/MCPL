package utils

import (
	"mc-portable-launcher/src/ux"
	"os"
	"path/filepath"

	"github.com/gofrs/flock"
)

func TryLock(path string) *flock.Flock {

	lockPath := filepath.Join(path, "LOCK.lock")
	fileLock := flock.New(lockPath)

	os.MkdirAll(path, 0755)

	locked, err := fileLock.TryLock()

	if err != nil {
		ux.ShowError("Locking failed", "Failed to lock the directory. Is another instance of the installer running?")
		os.Exit(1)
	}

	if !locked {
		ux.ShowError("More than one instance detected", "Unable to start. Another instance is running.")
		os.Exit(1)
	}

	return fileLock
}
