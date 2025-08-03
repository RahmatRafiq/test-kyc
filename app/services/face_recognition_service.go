package services

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"

	"golang_starter_kit_2025/app/responses"
)

type FaceRecognitionService struct{}

// FaceDescriptor represents face features
type FaceDescriptor struct {
	HogFeatures  []float64
	LbphFeatures []float64
	BoundingBox  Rectangle
}

type Rectangle struct {
	X, Y, Width, Height int
}

// CompareFaces compares ID card face with selfie using HOG and LBPH
func (*FaceRecognitionService) CompareFaces(idCardImagePath, selfieImagePath string) (*responses.KycScores, error) {
	// Load and preprocess images
	idCardImg, err := loadAndPreprocessImage(idCardImagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load ID card image: %v", err)
	}

	selfieImg, err := loadAndPreprocessImage(selfieImagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load selfie image: %v", err)
	}

	// Detect and extract faces
	idCardFace, err := detectAndExtractFace(idCardImg)
	if err != nil {
		return nil, fmt.Errorf("failed to detect face in ID card: %v", err)
	}

	selfieFace, err := detectAndExtractFace(selfieImg)
	if err != nil {
		return nil, fmt.Errorf("failed to detect face in selfie: %v", err)
	}

	// Extract features using HOG
	idCardHogFeatures := extractHOGFeatures(idCardFace)
	selfieHogFeatures := extractHOGFeatures(selfieFace)

	// Extract features using LBPH
	idCardLbphFeatures := extractLBPHFeatures(idCardFace)
	selfieLbphFeatures := extractLBPHFeatures(selfieFace)

	// Calculate similarity scores
	hogScore := calculateSimilarity(idCardHogFeatures, selfieHogFeatures)
	lbphScore := calculateSimilarity(idCardLbphFeatures, selfieLbphFeatures)

	// Ensemble scoring (weighted combination)
	ensembleScore := calculateEnsembleScore(hogScore, lbphScore)

	scores := &responses.KycScores{
		HogScore:       hogScore,
		LbphScore:      lbphScore,
		EnsembleScore:  ensembleScore,
		FaceMatchScore: ensembleScore, // Use ensemble as main face match score
	}

	return scores, nil
}

// DetectFace detects face in image and returns bounding box
func (*FaceRecognitionService) DetectFace(imagePath string) (*Rectangle, error) {
	img, err := loadAndPreprocessImage(imagePath)
	if err != nil {
		return nil, err
	}

	// Simple face detection using basic algorithms
	// In production, you might want to use more sophisticated methods
	face := detectFaceRegion(img)
	if face == nil {
		return nil, errors.New("no face detected")
	}

	return face, nil
}

// CropFace crops face region from image
func (*FaceRecognitionService) CropFace(imagePath string, boundingBox Rectangle) (image.Image, error) {
	img, err := loadAndPreprocessImage(imagePath)
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()

	// Validate bounding box
	if boundingBox.X < 0 || boundingBox.Y < 0 ||
		boundingBox.X+boundingBox.Width > bounds.Dx() ||
		boundingBox.Y+boundingBox.Height > bounds.Dy() {
		return nil, errors.New("invalid bounding box")
	}

	// Crop the face region
	croppedImg := cropImage(img, boundingBox)
	return croppedImg, nil
}

// Helper functions

func loadAndPreprocessImage(imagePath string) (image.Image, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var img image.Image
	ext := strings.ToLower(filepath.Ext(imagePath))
	switch ext {
	case ".jpg", ".jpeg":
		img, err = jpeg.Decode(file)
	case ".png":
		img, err = png.Decode(file)
	default:
		return nil, fmt.Errorf("unsupported image format: %s", ext)
	}

	if err != nil {
		return nil, err
	}

	// Convert to grayscale for better processing
	return convertToGrayscale(img), nil
}

func convertToGrayscale(img image.Image) image.Image {
	bounds := img.Bounds()
	grayImg := image.NewGray(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			originalColor := img.At(x, y)
			grayColor := color.GrayModel.Convert(originalColor)
			grayImg.Set(x, y, grayColor)
		}
	}

	return grayImg
}

func detectAndExtractFace(img image.Image) (image.Image, error) {
	faceRegion := detectFaceRegion(img)
	if faceRegion == nil {
		return nil, errors.New("no face detected")
	}

	return cropImage(img, *faceRegion), nil
}

func detectFaceRegion(img image.Image) *Rectangle {
	bounds := img.Bounds()

	// Simple face detection algorithm (placeholder)
	// In production, implement proper face detection using:
	// - Haar cascades
	// - HOG + SVM
	// - Deep learning models (if allowed)

	// For now, assume face is in the center region
	width := bounds.Dx()
	height := bounds.Dy()

	// Estimate face region (center 60% of image)
	faceWidth := int(float64(width) * 0.6)
	faceHeight := int(float64(height) * 0.6)
	faceX := (width - faceWidth) / 2
	faceY := (height - faceHeight) / 2

	return &Rectangle{
		X:      faceX,
		Y:      faceY,
		Width:  faceWidth,
		Height: faceHeight,
	}
}

func cropImage(img image.Image, rect Rectangle) image.Image {
	bounds := image.Rect(rect.X, rect.Y, rect.X+rect.Width, rect.Y+rect.Height)
	croppedImg := image.NewRGBA(bounds)

	for y := rect.Y; y < rect.Y+rect.Height; y++ {
		for x := rect.X; x < rect.X+rect.Width; x++ {
			croppedImg.Set(x, y, img.At(x, y))
		}
	}

	return croppedImg
}

// HOG (Histogram of Oriented Gradients) feature extraction
func extractHOGFeatures(img image.Image) []float64 {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Calculate gradients
	gradients := make([][]float64, height)
	orientations := make([][]float64, height)

	for y := 0; y < height; y++ {
		gradients[y] = make([]float64, width)
		orientations[y] = make([]float64, width)

		for x := 0; x < width; x++ {
			gx, gy := calculateGradient(img, x, y)
			gradients[y][x] = math.Sqrt(gx*gx + gy*gy)
			orientations[y][x] = math.Atan2(gy, gx)
		}
	}

	// Create HOG descriptor
	cellSize := 8
	blockSize := 2
	numBins := 9

	features := []float64{}

	for y := 0; y < height-blockSize*cellSize; y += cellSize {
		for x := 0; x < width-blockSize*cellSize; x += cellSize {
			blockFeatures := extractBlockFeatures(gradients, orientations, x, y, cellSize, blockSize, numBins)
			features = append(features, blockFeatures...)
		}
	}

	return features
}

func calculateGradient(img image.Image, x, y int) (float64, float64) {
	bounds := img.Bounds()

	// Handle edge cases
	x1 := x - 1
	x2 := x + 1
	y1 := y - 1
	y2 := y + 1

	if x1 < bounds.Min.X {
		x1 = bounds.Min.X
	}
	if x2 >= bounds.Max.X {
		x2 = bounds.Max.X - 1
	}
	if y1 < bounds.Min.Y {
		y1 = bounds.Min.Y
	}
	if y2 >= bounds.Max.Y {
		y2 = bounds.Max.Y - 1
	}

	// Get pixel intensities
	left := getPixelIntensity(img, x1, y)
	right := getPixelIntensity(img, x2, y)
	top := getPixelIntensity(img, x, y1)
	bottom := getPixelIntensity(img, x, y2)

	gx := right - left
	gy := bottom - top

	return gx, gy
}

func getPixelIntensity(img image.Image, x, y int) float64 {
	c := img.At(x, y)
	r, g, b, _ := c.RGBA()
	// Convert to grayscale using standard formula
	gray := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
	return gray / 65535.0 // RGBA values are 16-bit
}

func extractBlockFeatures(gradients, orientations [][]float64, startX, startY, cellSize, blockSize, numBins int) []float64 {
	features := []float64{}

	for by := 0; by < blockSize; by++ {
		for bx := 0; bx < blockSize; bx++ {
			cellFeatures := extractCellFeatures(gradients, orientations,
				startX+bx*cellSize, startY+by*cellSize, cellSize, numBins)
			features = append(features, cellFeatures...)
		}
	}

	// Normalize block features
	return normalizeFeatures(features)
}

func extractCellFeatures(gradients, orientations [][]float64, startX, startY, cellSize, numBins int) []float64 {
	histogram := make([]float64, numBins)
	binSize := math.Pi / float64(numBins)

	for y := startY; y < startY+cellSize && y < len(gradients); y++ {
		for x := startX; x < startX+cellSize && x < len(gradients[y]); x++ {
			magnitude := gradients[y][x]
			orientation := orientations[y][x]

			// Convert to positive orientation
			if orientation < 0 {
				orientation += math.Pi
			}

			// Find bin
			bin := int(orientation / binSize)
			if bin >= numBins {
				bin = numBins - 1
			}

			histogram[bin] += magnitude
		}
	}

	return histogram
}

// LBPH (Local Binary Patterns Histogram) feature extraction
func extractLBPHFeatures(img image.Image) []float64 {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Calculate LBP values
	lbpImage := make([][]int, height)
	for y := 0; y < height; y++ {
		lbpImage[y] = make([]int, width)
		for x := 0; x < width; x++ {
			if x > 0 && x < width-1 && y > 0 && y < height-1 {
				lbpImage[y][x] = calculateLBP(img, x, y)
			}
		}
	}

	// Create histogram
	histogram := make([]float64, 256) // LBP values range from 0-255

	for y := 1; y < height-1; y++ {
		for x := 1; x < width-1; x++ {
			histogram[lbpImage[y][x]]++
		}
	}

	// Normalize histogram
	total := float64((width - 2) * (height - 2))
	for i := range histogram {
		histogram[i] /= total
	}

	return histogram
}

func calculateLBP(img image.Image, x, y int) int {
	center := getPixelIntensity(img, x, y)

	// 8-neighbor LBP
	neighbors := []struct{ dx, dy int }{
		{-1, -1}, {0, -1}, {1, -1},
		{1, 0}, {1, 1}, {0, 1},
		{-1, 1}, {-1, 0},
	}

	lbpValue := 0
	for i, neighbor := range neighbors {
		neighborIntensity := getPixelIntensity(img, x+neighbor.dx, y+neighbor.dy)
		if neighborIntensity >= center {
			lbpValue |= (1 << i)
		}
	}

	return lbpValue
}

// Similarity calculation
func calculateSimilarity(features1, features2 []float64) float64 {
	if len(features1) != len(features2) {
		return 0.0
	}

	// Calculate cosine similarity
	dotProduct := 0.0
	norm1 := 0.0
	norm2 := 0.0

	for i := 0; i < len(features1); i++ {
		dotProduct += features1[i] * features2[i]
		norm1 += features1[i] * features1[i]
		norm2 += features2[i] * features2[i]
	}

	if norm1 == 0 || norm2 == 0 {
		return 0.0
	}

	similarity := dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))

	// Convert to percentage (0-100)
	if similarity < 0 {
		similarity = 0
	}

	return similarity * 100.0
}

// Ensemble scoring
func calculateEnsembleScore(hogScore, lbphScore float64) float64 {
	// Weighted average (HOG typically more reliable for face recognition)
	hogWeight := 0.7
	lbphWeight := 0.3

	ensembleScore := (hogScore * hogWeight) + (lbphScore * lbphWeight)

	// Ensure score is within valid range
	if ensembleScore > 100.0 {
		ensembleScore = 100.0
	}
	if ensembleScore < 0.0 {
		ensembleScore = 0.0
	}

	return ensembleScore
}

// Feature normalization
func normalizeFeatures(features []float64) []float64 {
	// L2 normalization
	sum := 0.0
	for _, feature := range features {
		sum += feature * feature
	}

	if sum == 0 {
		return features
	}

	norm := math.Sqrt(sum)
	normalized := make([]float64, len(features))
	for i, feature := range features {
		normalized[i] = feature / norm
	}

	return normalized
}
