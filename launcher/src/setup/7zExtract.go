package setup

import (
	"fmt"
	"io"
	"math"
	"mc-portable-launcher/src/ui"
	"mc-portable-launcher/src/utils"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"github.com/bodgit/sevenzip"
)

func Unzip7zWithProgress(app fyne.App, src7z string, destDir string, progressBar *canvas.Rectangle, statusText *canvas.Text, needToExtract []string, upToDate int) error {
	ui.UpdateProgressBarText(app, statusText, "Opening 7z archive... (due to solid compression, this might take a minute or two)")
	ui.UpdateProgressBar(app, progressBar, 0)

	var needToExtractSet map[string]struct{} = utils.SliceToMap(needToExtract)

	processed := 0
	enumerated := 0

	r, err := sevenzip.OpenReader(src7z)
	if err != nil {
		return err
	}
	defer r.Close()

	var totalBytes int64
	for _, f := range r.File {
		if !f.FileInfo().IsDir() {
			totalBytes += int64(f.UncompressedSize)
		}
	}

	var processedBytes int64
	timeStart := time.Now()
	destDirClean := filepath.Clean(destDir)

	// Use a Mutex for thread-safe map access and UI timing
	var mu sync.Mutex
	lastBarUpdate := time.Now()
	createdDirs := make(map[string]bool)

	// Pre-create all directories (Synchronous is fine here)
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			path := filepath.Join(destDirClean, f.Name)

			os.MkdirAll(path, 0755)
			createdDirs[path] = true
		}
	}

	workerCount := runtime.NumCPU()
	if workerCount > 1 {
		workerCount = 1
	}
	sem := make(chan struct{}, workerCount)
	var wg sync.WaitGroup

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}

		relative := strings.ReplaceAll(f.Name, "zzzzassets", "assets")
		enumerated++

		if processed >= len(needToExtract) {
			// we have extracted everything
			fmt.Println("statistics: enumerated: " + strconv.Itoa(enumerated) + " processed: " + strconv.Itoa(processed))
			break
		}

		if _, exists := needToExtractSet[filepath.Clean(relative)]; !exists {
			fmt.Println("Skipping: " + relative + " processed: " + strconv.Itoa(processed) + "/" + strconv.Itoa(len(needToExtract)))
			continue
		} else {

			processed++
			fmt.Println("Extracting: " + relative + "processed: " + strconv.Itoa(processed) + "/" + strconv.Itoa(len(needToExtract)))
		}

		wg.Add(1)
		sem <- struct{}{}

		go func(f *sevenzip.File) {
			defer wg.Done()
			defer func() { <-sem }()

			// FIX 1: Each goroutine MUST have its own buffer
			copyBuf := make([]byte, 1024*1024) // 32KB is usually plenty for I/O

			path := filepath.Join(destDirClean, filepath.Clean(f.Name))

			// Prevent ZipSlip
			if !strings.HasPrefix(path, destDirClean+string(os.PathSeparator)) {
				return
			}

			parentDir := filepath.Dir(path)

			// FIX 2: Thread-safe directory checking
			mu.Lock()
			if !createdDirs[parentDir] {
				os.MkdirAll(parentDir, 0755)
				createdDirs[parentDir] = true
			}
			mu.Unlock()

			rc, err := f.Open()
			if err != nil {
				return
			}
			defer rc.Close()

			dstFile, err := os.Create(strings.ReplaceAll(path, "zzzzassets", "assets"))
			if err != nil {
				return
			}
			defer dstFile.Close()

			for {
				n, readErr := rc.Read(copyBuf)
				if n > 0 {
					dstFile.Write(copyBuf[:n])
					newProcessed := atomic.AddInt64(&processedBytes, int64(n))

					// FIX 3: Thread-safe UI throttling
					mu.Lock()
					if time.Since(lastBarUpdate) > 100*time.Millisecond {
						progress := float32(newProcessed) / float32(totalBytes)
						ui.UpdateProgressBar(app, progressBar, progress)

						elapsed := time.Since(timeStart).Seconds()
						// Avoid division by zero
						if progress > 0 {
							remaining := (1 - progress) * float32(elapsed) / progress
							formattedTime := utils.FormatSeconds(int(math.Round(float64(remaining))))
							ui.UpdateProgressBarText(app, statusText, fmt.Sprintf("Extracting (%s): %s",
								formattedTime, utils.TruncateMiddle(f.Name, 32)))
						}
						lastBarUpdate = time.Now()
					}
					mu.Unlock()
				}
				if readErr == io.EOF {
					break
				}
				if readErr != nil {
					break
				}
			}
		}(f)
	}

	wg.Wait()
	ui.UpdateProgressBarText(app, statusText, "Extraction complete!")
	ui.UpdateProgressBar(app, progressBar, 1.0)
	return nil
}
