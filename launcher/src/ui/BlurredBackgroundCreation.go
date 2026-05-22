package ui

import (
	"image/draw"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"image"
	"image/color"

	_ "image/png"
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

	// 4. Draw the crop into a NEW image to reset the (0,0) origin
	resetImg := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(resetImg, resetImg.Bounds(), blurredBg, image.Pt(x, y), draw.Src)

	// 5. NEW: Round the image corners to match your UI
	// Map your Fyne UI radius (e.g., 12.0) to actual pixel units using scaleX
	pixelRadius := int(12.0 * scaleX)
	roundedImg := applyCornerRadius(resetImg, pixelRadius)

	img := canvas.NewImageFromImage(roundedImg)
	img.FillMode = canvas.ImageFillStretch

	return img
}

// Helper function to dynamically clip the image corners to a specific pixel radius
func applyCornerRadius(src *image.RGBA, radius int) image.Image {
	if radius <= 0 {
		return src
	}

	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	dst := image.NewRGBA(bounds)

	// Copy the original image into our new destination image
	draw.Draw(dst, bounds, src, bounds.Min, draw.Src)

	// Inline function to check if a pixel coordinate is outside the circle sector
	isOutsideCorner := func(cx, cy, px, py int) bool {
		dx := px - cx
		dy := py - cy
		// Pythagorean theorem to see if the pixel is further out than the radius
		return (dx*dx + dy*dy) > radius*radius
	}

	// Loop through and clear out the 4 corner zones to alpha 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			// Top-Left corner
			if x < radius && y < radius {
				if isOutsideCorner(radius, radius, x, y) {
					dst.Set(x, y, color.Transparent)
				}
			} // Top-Right corner
			if x >= w-radius && y < radius {
				if isOutsideCorner(w-radius-1, radius, x, y) {
					dst.Set(x, y, color.Transparent)
				}
			} // Bottom-Left corner
			if x < radius && y >= h-radius {
				if isOutsideCorner(radius, h-radius-1, x, y) {
					dst.Set(x, y, color.Transparent)
				}
			} // Bottom-Right corner
			if x >= w-radius && y >= h-radius {
				if isOutsideCorner(w-radius-1, h-radius-1, x, y) {
					dst.Set(x, y, color.Transparent)
				}
			}
		}
	}

	return dst
}

func GenerateBlurredShadow(width, height int, radius float32, blurRadius int, intensity uint8) image.Image {
	// Add padding around the image so the blur doesn't clip at the edges
	pad := blurRadius * 2
	imgW := width + (pad * 2)
	imgH := height + (pad * 2)

	img := image.NewRGBA(image.Rect(0, 0, imgW, imgH))

	// Draw a solid black rounded rect in the center (where the dialog sits)
	rFloat := float64(radius)
	for y := pad; y < imgH-pad; y++ {
		for x := pad; x < imgW-pad; x++ {
			// Basic rounded corner check
			dx := float64(x - pad)
			dy := float64(y - pad)
			wF := float64(width)
			hF := float64(height)

			inCorner := false
			if dx < rFloat && dy < rFloat && (rFloat-dx)*(rFloat-dx)+(rFloat-dy)*(rFloat-dy) > rFloat*rFloat {
				inCorner = true
			}
			if dx > wF-rFloat && dy < rFloat && (dx-(wF-rFloat))*(dx-(wF-rFloat))+(rFloat-dy)*(rFloat-dy) > rFloat*rFloat {
				inCorner = true
			}
			if dx < rFloat && dy > hF-rFloat && (rFloat-dx)*(rFloat-dx)+(dy-(hF-rFloat))*(dy-(hF-rFloat)) > rFloat*rFloat {
				inCorner = true
			}
			if dx > wF-rFloat && dy > hF-rFloat && (dx-(wF-rFloat))*(dx-(wF-rFloat))+(dy-(hF-rFloat))*(dy-(hF-rFloat)) > rFloat*rFloat {
				inCorner = true
			}

			if !inCorner {
				img.Set(x, y, color.RGBA{0, 0, 0, intensity})
			}
		}
	}

	// Two-pass box blur to simulate Gaussian blur
	blurred := image.NewRGBA(img.Bounds())

	// Horizontal pass
	for y := 0; y < imgH; y++ {
		for x := 0; x < imgW; x++ {
			var alphaSum int
			var count int
			for k := -blurRadius; k <= blurRadius; k++ {
				nx := x + k
				if nx >= 0 && nx < imgW {
					_, _, _, a := img.At(nx, y).RGBA()
					alphaSum += int(a >> 8)
					count++
				}
			}
			blurred.Set(x, y, color.RGBA{0, 0, 0, uint8(alphaSum / count)})
		}
	}

	// Vertical pass
	final := image.NewRGBA(img.Bounds())
	for x := 0; x < imgW; x++ {
		for y := 0; y < imgH; y++ {
			var alphaSum int
			var count int
			for k := -blurRadius; k <= blurRadius; k++ {
				ny := y + k
				if ny >= 0 && ny < imgH {
					_, _, _, a := blurred.At(x, ny).RGBA()
					alphaSum += int(a >> 8)
					count++
				}
			}
			final.Set(x, y, color.RGBA{0, 0, 0, uint8(alphaSum / count)})
		}
	}

	return final
}
