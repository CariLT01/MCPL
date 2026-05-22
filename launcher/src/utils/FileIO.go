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

		realFilePath := filepath.Join(gameDir, path)

		if FileExists(realFilePath) {
			// fmt.Println("Deleting: " + realFilePath)
			os.Remove(realFilePath)
		} else {
			fmt.Println("File does not exist and cannot be deleted: " + realFilePath)
		}
	}
}
