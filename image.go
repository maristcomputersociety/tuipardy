package main

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/BourgeoisBear/rasterm"
)

// handle image rendering in terminal
type ImageRenderer struct{}

func NewImageRenderer() *ImageRenderer {
	return &ImageRenderer{}
}

func (ir *ImageRenderer) RenderImageToString(imagePath string, maxWidth, maxHeight int) (string, error) {
	if imagePath == "" {
		return "", nil
	}

	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return "", fmt.Errorf("image file not found: %s", imagePath)
	}

	file, err := os.Open(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to open image file: %w", err)
	}
	defer file.Close()

	img, err := decodeImage(file, imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	var buf bytes.Buffer
	opts := rasterm.KittyImgOpts{}

	if maxWidth > 0 {
		opts.DstCols = uint32(maxWidth)
	}
	if maxHeight > 0 {
		opts.DstRows = uint32(maxHeight)
	}

	err = rasterm.KittyWriteImage(&buf, img, opts)

	if err != nil {
		return "", fmt.Errorf("failed to encode image for terminal: %w", err)
	}

	return buf.String(), nil
}

// GetImageBounds returns the original image dimensions in pixels
func (ir *ImageRenderer) GetImageBounds(imagePath string) (width, height int, err error) {
	if imagePath == "" {
		return 0, 0, fmt.Errorf("empty image path")
	}

	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return 0, 0, fmt.Errorf("image file not found: %s", imagePath)
	}

	file, err := os.Open(imagePath)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to open image file: %w", err)
	}
	defer file.Close()

	img, err := decodeImage(file, imagePath)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to decode image: %w", err)
	}

	bounds := img.Bounds()
	return bounds.Dx(), bounds.Dy(), nil
}

func decodeImage(file *os.File, imagePath string) (image.Image, error) {
	ext := strings.ToLower(filepath.Ext(imagePath))

	switch ext {
	case ".png":
		return png.Decode(file)
	case ".jpg", ".jpeg":
		return jpeg.Decode(file)
	case ".gif":
		return gif.Decode(file)
	default:
		img, _, err := image.Decode(file)
		return img, err
	}
}

func IsImageSupported() bool {
	return rasterm.IsKittyCapable()
}
