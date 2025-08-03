package helpers

import (
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
)

type ImageProcessor struct{}

func (ip *ImageProcessor) LoadImage(imagePath string) (image.Image, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(imagePath))
	switch ext {
	case ".jpg", ".jpeg":
		return jpeg.Decode(file)
	case ".png":
		return png.Decode(file)
	default:
		img, _, err := image.Decode(file)
		return img, err
	}
}

func (ip *ImageProcessor) ConvertToGrayscale(src image.Image) *image.Gray {
	bounds := src.Bounds()
	gray := image.NewGray(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			originalColor := src.At(x, y)
			grayColor := color.GrayModel.Convert(originalColor).(color.Gray)
			gray.Set(x, y, grayColor)
		}
	}

	return gray
}

func (ip *ImageProcessor) ResizeImage(src image.Image, width, height int) *image.RGBA {
	srcBounds := src.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()

	dst := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			srcX := int(float64(x) * float64(srcW) / float64(width))
			srcY := int(float64(y) * float64(srcH) / float64(height))

			if srcX >= srcW {
				srcX = srcW - 1
			}
			if srcY >= srcH {
				srcY = srcH - 1
			}

			dst.Set(x, y, src.At(srcX+srcBounds.Min.X, srcY+srcBounds.Min.Y))
		}
	}

	return dst
}

func (ip *ImageProcessor) CropImage(src image.Image, rect image.Rectangle) *image.RGBA {
	cropped := image.NewRGBA(rect)
	draw.Draw(cropped, rect, src, rect.Min, draw.Src)
	return cropped
}

func (ip *ImageProcessor) CalculateGradient(gray *image.Gray) ([][]float64, [][]float64) {
	bounds := gray.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	magnitude := make([][]float64, height)
	direction := make([][]float64, height)

	for y := 0; y < height; y++ {
		magnitude[y] = make([]float64, width)
		direction[y] = make([]float64, width)
	}

	sobelX := [][]int{
		{-1, 0, 1},
		{-2, 0, 2},
		{-1, 0, 1},
	}
	sobelY := [][]int{
		{-1, -2, -1},
		{0, 0, 0},
		{1, 2, 1},
	}

	for y := 1; y < height-1; y++ {
		for x := 1; x < width-1; x++ {
			var gx, gy float64

			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					pixel := gray.GrayAt(x+dx, y+dy).Y
					pixelValue := float64(pixel)

					gx += pixelValue * float64(sobelX[dy+1][dx+1])
					gy += pixelValue * float64(sobelY[dy+1][dx+1])
				}
			}

			magnitude[y][x] = math.Sqrt(gx*gx + gy*gy)
			direction[y][x] = math.Atan2(gy, gx)
		}
	}

	return magnitude, direction
}

func (ip *ImageProcessor) NormalizeVector(vector []float64) []float64 {
	var sum float64
	for _, v := range vector {
		sum += v * v
	}

	length := math.Sqrt(sum)
	if length == 0 {
		return vector
	}

	normalized := make([]float64, len(vector))
	for i, v := range vector {
		normalized[i] = v / length
	}

	return normalized
}
