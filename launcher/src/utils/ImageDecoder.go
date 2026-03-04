package utils

import (
	"bytes"

	_ "image/png"

	"fyne.io/fyne/v2"
	"image"
)

func ResourceToImage(res *fyne.StaticResource) (image.Image, error) {
	// StaticResource embeds []byte in res.StaticContent
	reader := bytes.NewReader(res.StaticContent)

	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, err
	}

	return img, nil
}
