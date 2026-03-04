package setup

import "mc-portable-launcher/src/ui"

import (
	"embed"
	"io"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

func ExtractFiles(file embed.FS, fileName string, savePath string, app fyne.App, progressBar *canvas.Rectangle, statusText *canvas.Text) error {
	ui.UpdateProgressBarText(app, statusText, "Unpacking...")

	src, err := file.Open(fileName)

	if err != nil {
		return err
	}

	defer src.Close()

	exePath, _ := os.Executable()
	installPath := filepath.Join(filepath.Dir(exePath), savePath)

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
			ui.UpdateProgressBar(app, progressBar, float32(written)/float32(totalSize))
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	ui.UpdateProgressBarText(app, statusText, "Unpacked")

	return nil
}
