package utils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !errors.Is(err, os.ErrNotExist)
}

func DeleteFiles(paths []string, gameDir string) {

	for _, path := range paths {
		if FileExists(path) {
			fmt.Println("Deleting: " + path)
			os.Remove(filepath.Join(gameDir, path))
		}
	}
}
