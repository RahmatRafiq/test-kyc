package services

import (
	"fmt"
	"image"
	"math"

	"golang_starter_kit_2025/app/helpers"
)

// HOGDetector implements HOG (Histogram of Oriented Gradients) for face detection
type HOGDetector struct {
	imageProcessor *helpers.ImageProcessor
	cellSize       int         // Size of each cell (typically 8x8)
	blockSize      int         // Size of each block in cells (typically 2x2)
	numBins        int         // Number of orientation bins (typically 9)
	windowSize     image.Point // Detection window size (typically 64x128 for person, we'll use 64x64 for face)
}

// NewHOGDetector creates a new HOG detector
func NewHOGDetector() *HOGDetector {
	return &HOGDetector{
		imageProcessor: &helpers.ImageProcessor{},
		cellSize:       8,                         // 8x8 pixel cells
		blockSize:      2,                         // 2x2 cell blocks
		numBins:        9,                         // 9 orientation bins (20 degrees each)
		windowSize:     image.Point{X: 64, Y: 64}, // 64x64 detection window for faces
	}
}

// ExtractHOGFeatures extracts HOG features from an image
func (hog *HOGDetector) ExtractHOGFeatures(imagePath string) ([]float64, error) {
	// Load and preprocess image
	img, err := hog.imageProcessor.LoadImage(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load image: %v", err)
	}

	// Convert to grayscale
	gray := hog.imageProcessor.ConvertToGrayscale(img)

	// Resize to standard window size
	resized := hog.imageProcessor.ResizeImage(gray, hog.windowSize.X, hog.windowSize.Y)
	grayResized := hog.imageProcessor.ConvertToGrayscale(resized)

	// Calculate gradients
	magnitude, direction := hog.imageProcessor.CalculateGradient(grayResized)

	// Calculate HOG features
	features := hog.calculateHOGFeatures(magnitude, direction)

	return features, nil
}

// calculateHOGFeatures calculates HOG features from gradient magnitude and direction
func (hog *HOGDetector) calculateHOGFeatures(magnitude, direction [][]float64) []float64 {
	height := len(magnitude)
	width := len(magnitude[0])

	// Calculate number of cells
	cellsX := width / hog.cellSize
	cellsY := height / hog.cellSize

	// Create cell histograms
	cellHistograms := make([][][]float64, cellsY)
	for y := 0; y < cellsY; y++ {
		cellHistograms[y] = make([][]float64, cellsX)
		for x := 0; x < cellsX; x++ {
			cellHistograms[y][x] = make([]float64, hog.numBins)
		}
	}

	// Fill cell histograms
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			cellX := x / hog.cellSize
			cellY := y / hog.cellSize

			if cellX >= cellsX || cellY >= cellsY {
				continue
			}

			// Convert direction to bin index (0 to numBins-1)
			angle := direction[y][x]
			if angle < 0 {
				angle += math.Pi
			}

			binSize := math.Pi / float64(hog.numBins)
			binIndex := int(angle / binSize)
			if binIndex >= hog.numBins {
				binIndex = hog.numBins - 1
			}

			// Add magnitude to histogram bin
			cellHistograms[cellY][cellX][binIndex] += magnitude[y][x]
		}
	}

	// Normalize histograms using blocks
	var features []float64
	blocksX := cellsX - hog.blockSize + 1
	blocksY := cellsY - hog.blockSize + 1

	for blockY := 0; blockY < blocksY; blockY++ {
		for blockX := 0; blockX < blocksX; blockX++ {
			// Collect features from all cells in this block
			var blockFeatures []float64

			for dy := 0; dy < hog.blockSize; dy++ {
				for dx := 0; dx < hog.blockSize; dx++ {
					cellY := blockY + dy
					cellX := blockX + dx
					blockFeatures = append(blockFeatures, cellHistograms[cellY][cellX]...)
				}
			}

			// L2 normalize the block
			normalizedBlock := hog.imageProcessor.NormalizeVector(blockFeatures)
			features = append(features, normalizedBlock...)
		}
	}

	return features
}

// DetectFaces detects faces in an image using sliding window
func (hog *HOGDetector) DetectFaces(imagePath string) ([]image.Rectangle, error) {
	img, err := hog.imageProcessor.LoadImage(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load image: %v", err)
	}

	bounds := img.Bounds()
	var faces []image.Rectangle

	// Sliding window detection
	stepSize := 16 // Move window by 16 pixels each step

	for y := 0; y <= bounds.Max.Y-hog.windowSize.Y; y += stepSize {
		for x := 0; x <= bounds.Max.X-hog.windowSize.X; x += stepSize {
			// Extract window
			windowRect := image.Rect(x, y, x+hog.windowSize.X, y+hog.windowSize.Y)
			window := hog.imageProcessor.CropImage(img, windowRect)

			// Check if this window contains a face
			if hog.isFace(window) {
				faces = append(faces, windowRect)
			}
		}
	}

	// Apply non-maximum suppression to remove overlapping detections
	faces = hog.nonMaximumSuppression(faces)

	return faces, nil
}

// isFace determines if a window contains a face using a simple threshold
// In a real implementation, this would use a trained classifier
func (hog *HOGDetector) isFace(window image.Image) bool {
	// Convert to grayscale and calculate some basic features
	gray := hog.imageProcessor.ConvertToGrayscale(window)
	magnitude, _ := hog.imageProcessor.CalculateGradient(gray)

	// Simple heuristic: check if there's enough gradient activity
	// This is a placeholder - in real implementation you'd use trained SVM classifier
	var totalMagnitude float64
	pixelCount := 0

	for y := 0; y < len(magnitude); y++ {
		for x := 0; x < len(magnitude[0]); x++ {
			totalMagnitude += magnitude[y][x]
			pixelCount++
		}
	}

	avgMagnitude := totalMagnitude / float64(pixelCount)

	// Threshold for face detection (this is very basic)
	return avgMagnitude > 10.0 // Adjust threshold as needed
}

// nonMaximumSuppression removes overlapping detections
func (hog *HOGDetector) nonMaximumSuppression(faces []image.Rectangle) []image.Rectangle {
	if len(faces) <= 1 {
		return faces
	}

	var result []image.Rectangle
	used := make([]bool, len(faces))

	for i := 0; i < len(faces); i++ {
		if used[i] {
			continue
		}

		current := faces[i]
		result = append(result, current)
		used[i] = true

		// Mark overlapping faces as used
		for j := i + 1; j < len(faces); j++ {
			if used[j] {
				continue
			}

			if hog.calculateIOU(current, faces[j]) > 0.3 { // 30% overlap threshold
				used[j] = true
			}
		}
	}

	return result
}

// calculateIOU calculates Intersection over Union for two rectangles
func (hog *HOGDetector) calculateIOU(rect1, rect2 image.Rectangle) float64 {
	// Calculate intersection
	intersection := rect1.Intersect(rect2)
	if intersection.Empty() {
		return 0.0
	}

	intersectionArea := float64(intersection.Dx() * intersection.Dy())

	// Calculate union
	area1 := float64(rect1.Dx() * rect1.Dy())
	area2 := float64(rect2.Dx() * rect2.Dy())
	unionArea := area1 + area2 - intersectionArea

	if unionArea == 0 {
		return 0.0
	}

	return intersectionArea / unionArea
}

// CompareFaceFeatures compares two sets of HOG features
func (hog *HOGDetector) CompareFaceFeatures(features1, features2 []float64) float64 {
	if len(features1) != len(features2) {
		return 0.0
	}

	// Calculate cosine similarity
	var dotProduct, norm1, norm2 float64

	for i := 0; i < len(features1); i++ {
		dotProduct += features1[i] * features2[i]
		norm1 += features1[i] * features1[i]
		norm2 += features2[i] * features2[i]
	}

	if norm1 == 0 || norm2 == 0 {
		return 0.0
	}

	similarity := dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))

	// Convert to 0-1 scale (cosine similarity ranges from -1 to 1)
	return (similarity + 1.0) / 2.0
}
