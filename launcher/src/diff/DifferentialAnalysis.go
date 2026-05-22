package diff

import (
	"encoding/hex"
	"fmt"
	"io"
	"mc-portable-launcher/src/config"
	"mc-portable-launcher/src/data"
	"mc-portable-launcher/src/ui"
	"mc-portable-launcher/src/utils"
	"mc-portable-launcher/src/ux"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"io/fs"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"github.com/cespare/xxhash/v2"
)

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hasher := xxhash.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%016x", hasher.Sum64()), nil
}

func WalkDirectoryAndHash(root string, gameDirRoot string, hashMap map[string]string, app fyne.App, statusText *canvas.Text) error {
	type result struct {
		path string
		hash string
	}

	absPath, err := filepath.Abs(root)

	if err != nil {
		return err
	}

	if !utils.FileExists(absPath) {
		return nil
	}

	paths := make(chan string, 100)
	results := make(chan result, 100)
	var wg sync.WaitGroup

	var processed atomic.Int64

	processed.Store(0)

	done := make(chan struct{})

	// 1. Start Workers
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range paths {
				h, err := hashFile(path) // Helper function to hash a single file
				if err != nil {
					ux.ShowErrorLog(err, "failed to hash file")
					continue
				}
				processed.Add(1)
				rel, _ := filepath.Rel(gameDirRoot, path)
				results <- result{rel, h}
			}
		}()
	}

	// start status text update
	go func() {
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				count := processed.Load()
				ui.UpdateProgressBarText(app, statusText, fmt.Sprintf("Enumerating local objects in: %d files", count))
			case <-done:
				return
			}
		}
	}()

	// 2. Start Walker
	go func() {
		filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if !d.IsDir() {
				paths <- path
			}
			return nil
		})
		close(paths)
	}()

	// 3. Closer & Collector
	go func() {
		wg.Wait()
		close(results)
		close(done)
	}()

	for res := range results {
		hashMap[res.path] = res.hash
	}

	return nil
}

func ComputeDiffMaps(oldHashes map[string]string, newHashes map[string]string) ([]string, []string, []string) {
	var toDelete, toAdd, unchanged []string

	for path, hash := range oldHashes {
		if newHash, ok := newHashes[path]; ok {
			if newHash != hash {
				fmt.Println("needs update: " + path)
				toAdd = append(toAdd, path) // changed file needs update
			} else {
				unchanged = append(unchanged, path) // hash is identical
			}
		} else {
			toDelete = append(toDelete, path)
		}
	}

	for path, _ := range newHashes {
		if _, ok := oldHashes[path]; !ok {
			fmt.Println("needs add: " + path)
			toAdd = append(toAdd, path)
		}
	}

	fmt.Printf("statistics: delete: %d add: %d up-to-date: %d \n", len(toDelete), len(toAdd), len(unchanged))

	return toDelete, toAdd, unchanged
}

func ComputeCurrentHashes(app fyne.App, statusText *canvas.Text) map[string]string {
	hashMap := make(map[string]string)

	exePath, _ := os.Executable()
	gameDir := filepath.Join(filepath.Dir(exePath), config.VERSION_DIR_STRING)

	err := WalkDirectoryAndHash(filepath.Join(gameDir, "assets"), gameDir, hashMap, app, statusText)
	err = WalkDirectoryAndHash(filepath.Join(gameDir, "java"), gameDir, hashMap, app, statusText)
	err = WalkDirectoryAndHash(filepath.Join(gameDir, "libraries"), gameDir, hashMap, app, statusText)

	if err != nil {
		fmt.Println(err)
	}

	return hashMap
}

func ReadInstallationHashes(binaryData []byte) map[string]string {
	hashMap := make(map[string]string)

	readerPosition := 0

	for {
		// read the first byte, contains the size of the path
		size := binaryData[readerPosition]
		readerPosition++

		// now we should expect N bytes following

		// first byte is identifier (A: asset, F: file)

		identifier := binaryData[readerPosition]
		readerPosition++

		// read the path, which is of size N - 1

		path := binaryData[readerPosition : readerPosition+int(size)-1]
		readerPosition += int(size) - 1

		// 8 bytes after for XXHash checksum

		hash := hex.EncodeToString(binaryData[readerPosition : readerPosition+8])
		readerPosition += 8

		if string(identifier) == "A" {
			// found an A identifier

			// we can build the original path

			hexPath := hex.EncodeToString(path)
			// build assets path
			assetsPath := filepath.Clean("assets/objects/" + hexPath[:2] + "/" + hexPath)
			hashMap[assetsPath] = string(hash)
			fmt.Printf("assets path: %s hash: %s \n", assetsPath, hash)

		} else if string(identifier) == "F" {
			// found a F identifier

			cleanedPath := filepath.Clean(string(path))
			hashMap[cleanedPath] = string(hash)
			fmt.Printf("file path: %s hash: %s \n", cleanedPath, hash)
		}

		// at the end check
		if readerPosition >= len(binaryData) {
			break
		}

	}

	return hashMap
}

func ReadUnskippableFiles() []string {
	fmt.Println("Read unskippable binary")
	unskippableFiles := make([]string, 0, 32)

	readerPosition := 0

	for {
		size := data.UnskippableBinary[readerPosition]
		readerPosition++

		path := data.UnskippableBinary[readerPosition : readerPosition+int(size)]
		readerPosition += int(size)

		unskippableFiles = append(unskippableFiles, filepath.Clean(string(path)))

		// fmt.Println("Unskippable: " + string(path))

		if readerPosition >= len(data.UnskippableBinary) {
			break
		}

	}

	fmt.Println("finish unskippable binary read")

	return unskippableFiles

}
