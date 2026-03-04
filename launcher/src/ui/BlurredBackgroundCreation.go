package ui

import (
	"image/draw"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"image"
)

func CreateBlurredBackground(blurredBg image.Image, pos fyne.Position, size fyne.Size, canvasSize fyne.Size) fyne.CanvasObject {
	imgBounds := blurredBg.Bounds()
	imgW, imgH := float32(imgBounds.Dx()), float32(imgBounds.Dy())

	// 1. Calculate ratios (Units to Pixels)
	scaleX := imgW / canvasSize.Width
	scaleY := imgH / canvasSize.Height

	// 2. Map coordinates
	x := int(pos.X * scaleX)
	y := int(pos.Y * scaleY)
	w := int(size.Width * scaleX)
	h := int(size.Height * scaleY)

	// 3. Clamp bounds
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x+w > int(imgW) {
		w = int(imgW) - x
	}
	if y+h > int(imgH) {
		h = int(imgH) - y
	}

	// 4. THE FIX: Draw the crop into a NEW image to reset the (0,0) origin
	// This prevents the "bottom-right shift" issue
	resetImg := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(resetImg, resetImg.Bounds(), blurredBg, image.Pt(x, y), draw.Src)

	img := canvas.NewImageFromImage(resetImg)
	img.FillMode = canvas.ImageFillStretch

	return img
}
