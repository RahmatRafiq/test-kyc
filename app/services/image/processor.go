package image

import (
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
)

// ImageProcessor handles basic image operations
type ImageProcessor struct{}

// LoadImage loads and decodes image from file path
func (*ImageProcessor) LoadImage(imagePath string) (image.Image, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Decode image based on extension
	var img image.Image
	ext := strings.ToLower(filepath.Ext(imagePath))
	switch ext {
	case ".jpg", ".jpeg":
		img, err = jpeg.Decode(file)
	case ".png":
		img, err = png.Decode(file)
	default:
		return nil, errors.New("unsupported image format")
	}

	if err != nil {
		return nil, err
	}

	return img, nil
}

// ConvertToGrayscale converts image to grayscale with improved formula
func (*ImageProcessor) ConvertToGrayscale(img image.Image) *image.Gray {
	bounds := img.Bounds()
	gray := image.NewGray(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			// Use luminance formula for better grayscale conversion
			grayValue := uint8((299*r + 587*g + 114*b + 500) / 1000 >> 8)
			gray.Set(x, y, color.Gray{Y: grayValue})
		}
	}

	return gray
}

// ApplyGaussianBlur applies a simple Gaussian blur to reduce noise
func (*ImageProcessor) ApplyGaussianBlur(img *image.Gray) *image.Gray {
	bounds := img.Bounds()
	blurred := image.NewGray(bounds)

	// Simple 3x3 Gaussian kernel
	kernel := [3][3]float64{
		{1, 2, 1},
		{2, 4, 2},
		{1, 2, 1},
	}
	kernelSum := 16.0

	for y := bounds.Min.Y + 1; y < bounds.Max.Y-1; y++ {
		for x := bounds.Min.X + 1; x < bounds.Max.X-1; x++ {
			var sum float64
			for ky := -1; ky <= 1; ky++ {
				for kx := -1; kx <= 1; kx++ {
					pixel := img.GrayAt(x+kx, y+ky)
					sum += float64(pixel.Y) * kernel[ky+1][kx+1]
				}
			}
			blurred.Set(x, y, color.Gray{Y: uint8(sum / kernelSum)})
		}
	}

	return blurred
}

// CalculateLocalThreshold calculates adaptive threshold for a pixel
func (*ImageProcessor) CalculateLocalThreshold(img *image.Gray, x, y, windowSize int) uint8 {
	bounds := img.Bounds()
	sum := 0
	count := 0

	halfWindow := windowSize / 2

	for dy := -halfWindow; dy <= halfWindow; dy++ {
		for dx := -halfWindow; dx <= halfWindow; dx++ {
			px, py := x+dx, y+dy
			if px >= bounds.Min.X && px < bounds.Max.X && py >= bounds.Min.Y && py < bounds.Max.Y {
				sum += int(img.GrayAt(px, py).Y)
				count++
			}
		}
	}

	if count == 0 {
		return 128
	}

	mean := sum / count
	return uint8(mean - 10) // Slight bias towards darker threshold
}

// EnhanceForOCR applies image enhancement for better OCR results
func (p *ImageProcessor) EnhanceForOCR(img *image.Gray) *image.Gray {
	bounds := img.Bounds()
	enhanced := image.NewGray(bounds)

	// Apply adaptive thresholding for better text contrast
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pixel := img.GrayAt(x, y)
			threshold := p.CalculateLocalThreshold(img, x, y, 15)

			if pixel.Y > threshold {
				enhanced.Set(x, y, color.Gray{Y: 255}) // White
			} else {
				enhanced.Set(x, y, color.Gray{Y: 0}) // Black
			}
		}
	}

	return enhanced
}

// ApplyAdvancedPreprocessing applies more sophisticated image preprocessing
func (p *ImageProcessor) ApplyAdvancedPreprocessing(img *image.Gray) *image.Gray {
	// Apply Gaussian blur effect first to reduce noise
	blurred := p.ApplyGaussianBlur(img)

	// Then apply adaptive thresholding
	return p.EnhanceForOCR(blurred)
}
