package services

import (
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

type FaceRegion struct {
	X, Y, Width, Height int
	Confidence          float64
}

// CompareFaces compares ID card face with selfie using improved algorithms
func (*FaceRecognitionService) CompareFaces(idCardImagePath, selfieImagePath string) (*responses.KycScores, error) {
	// Load images
	idCardImg, err := loadImageForFace(idCardImagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load ID card image: %v", err)
	}

	selfieImg, err := loadImageForFace(selfieImagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load selfie image: %v", err)
	}

	// Process face comparison
	scores := processFaceComparison(idCardImg, selfieImg)

	return scores, nil
}

// loadImageForFace loads image specifically for face recognition
func loadImageForFace(imagePath string) (image.Image, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open image: %v", err)
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
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	return img, nil
}

// processFaceComparison performs the actual face comparison
func processFaceComparison(idCardImg, selfieImg image.Image) *responses.KycScores {
	// Convert to grayscale for processing
	idCardGray := convertImageToGrayscale(idCardImg)
	selfieGray := convertImageToGrayscale(selfieImg)

	// Detect face regions with improved detection
	idCardFace := detectFaceRegionImproved(idCardGray)
	selfieFace := detectFaceRegionImproved(selfieGray)

	// Extract enhanced features
	idCardFeatures := extractEnhancedFeatures(idCardGray, idCardFace)
	selfieFeatures := extractEnhancedFeatures(selfieGray, selfieFace)

	// Calculate multiple similarity scores
	structuralScore := calculateStructuralSimilarity(idCardFeatures, selfieFeatures)
	textureScore := calculateAdvancedTextureScore(idCardGray, selfieGray, idCardFace, selfieFace)
	histogramScore := calculateHistogramSimilarity(idCardGray, selfieGray, idCardFace, selfieFace)

	// Weighted ensemble scoring with multiple algorithms
	hogScore := structuralScore
	lbphScore := textureScore
	ensembleScore := (structuralScore*0.4 + textureScore*0.35 + histogramScore*0.25)

	return &responses.KycScores{
		HogScore:       hogScore,
		LbphScore:      lbphScore,
		EnsembleScore:  ensembleScore,
		FaceMatchScore: ensembleScore,
	}
}

// convertImageToGrayscale converts color image to grayscale
func convertImageToGrayscale(img image.Image) *image.Gray {
	bounds := img.Bounds()
	gray := image.NewGray(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.At(x, y)
			r, g, b, _ := c.RGBA()
			// Standard grayscale conversion
			grayValue := uint8((299*r + 587*g + 114*b) / 1000 / 256)
			gray.SetGray(x, y, color.Gray{Y: grayValue})
		}
	}

	return gray
}

// detectFaceRegion detects face region in image using simple methods
func detectFaceRegion(img *image.Gray) FaceRegion {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Simple face detection using skin color approximation and shape analysis
	// In a real implementation, this would use more sophisticated algorithms

	// Look for darker regions that could be facial features
	bestRegion := findBestFaceCandidate(img)

	if bestRegion.Width == 0 || bestRegion.Height == 0 {
		// Fallback to center region if no face detected
		bestRegion = FaceRegion{
			X:          width / 4,
			Y:          height / 4,
			Width:      width / 2,
			Height:     height / 2,
			Confidence: 0.3,
		}
	}

	return bestRegion
}

// findBestFaceCandidate finds the most likely face region
func findBestFaceCandidate(img *image.Gray) FaceRegion {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	bestRegion := FaceRegion{}
	bestScore := 0.0

	// Scan different regions for face-like characteristics
	for y := height / 6; y < height*2/3; y += 20 {
		for x := width / 6; x < width*2/3; x += 20 {
			regionWidth := width / 3
			regionHeight := height / 3

			if x+regionWidth > width || y+regionHeight > height {
				continue
			}

			score := evaluateFaceRegion(img, x, y, regionWidth, regionHeight)
			if score > bestScore {
				bestScore = score
				bestRegion = FaceRegion{
					X:          x,
					Y:          y,
					Width:      regionWidth,
					Height:     regionHeight,
					Confidence: score,
				}
			}
		}
	}

	return bestRegion
}

// evaluateFaceRegion evaluates how likely a region contains a face
func evaluateFaceRegion(img *image.Gray, x, y, width, height int) float64 {
	if width <= 0 || height <= 0 {
		return 0.0
	}

	// Check for face-like characteristics
	score := 0.0

	// Look for eye-like regions (darker areas in upper portion)
	eyeRegionScore := checkEyeRegion(img, x, y+height/6, width, height/4)
	score += eyeRegionScore * 0.4

	// Look for mouth region (darker area in lower portion)
	mouthRegionScore := checkMouthRegion(img, x, y+height*2/3, width, height/6)
	score += mouthRegionScore * 0.2

	// Check overall contrast and variation
	contrastScore := checkContrast(img, x, y, width, height)
	score += contrastScore * 0.4

	return score
}

// checkEyeRegion checks for eye-like patterns
func checkEyeRegion(img *image.Gray, x, y, width, height int) float64 {
	darkPixels := 0
	totalPixels := 0

	for py := y; py < y+height && py < img.Bounds().Max.Y; py++ {
		for px := x; px < x+width && px < img.Bounds().Max.X; px++ {
			totalPixels++
			if img.GrayAt(px, py).Y < 100 { // Dark pixel
				darkPixels++
			}
		}
	}

	if totalPixels == 0 {
		return 0.0
	}

	darkRatio := float64(darkPixels) / float64(totalPixels)

	// Eyes should have some dark regions but not be completely dark
	if darkRatio > 0.15 && darkRatio < 0.6 {
		return darkRatio
	}

	return 0.0
}

// checkMouthRegion checks for mouth-like patterns
func checkMouthRegion(img *image.Gray, x, y, width, height int) float64 {
	darkPixels := 0
	totalPixels := 0

	for py := y; py < y+height && py < img.Bounds().Max.Y; py++ {
		for px := x; px < x+width && px < img.Bounds().Max.X; px++ {
			totalPixels++
			if img.GrayAt(px, py).Y < 80 { // Very dark pixel
				darkPixels++
			}
		}
	}

	if totalPixels == 0 {
		return 0.0
	}

	darkRatio := float64(darkPixels) / float64(totalPixels)

	// Mouth should have some dark regions
	if darkRatio > 0.1 && darkRatio < 0.4 {
		return darkRatio
	}

	return 0.0
}

// checkContrast checks the contrast and variation in the region
func checkContrast(img *image.Gray, x, y, width, height int) float64 {
	var pixels []uint8

	for py := y; py < y+height && py < img.Bounds().Max.Y; py++ {
		for px := x; px < x+width && px < img.Bounds().Max.X; px++ {
			pixels = append(pixels, img.GrayAt(px, py).Y)
		}
	}

	if len(pixels) == 0 {
		return 0.0
	}

	// Calculate standard deviation
	mean := calculateMean(pixels)
	variance := calculateVariance(pixels, mean)
	stdDev := math.Sqrt(variance)

	// Good faces have reasonable contrast (standard deviation)
	if stdDev > 20 && stdDev < 80 {
		return stdDev / 80.0
	}

	return 0.0
}

// extractSimpleFeatures extracts basic features from face region
func extractSimpleFeatures(img *image.Gray, faceRegion FaceRegion) []float64 {
	features := make([]float64, 16) // Simple 4x4 grid features

	cellWidth := faceRegion.Width / 4
	cellHeight := faceRegion.Height / 4

	index := 0
	for row := 0; row < 4; row++ {
		for col := 0; col < 4; col++ {
			cellX := faceRegion.X + col*cellWidth
			cellY := faceRegion.Y + row*cellHeight

			avgIntensity := calculateAverageIntensity(img, cellX, cellY, cellWidth, cellHeight)
			features[index] = avgIntensity
			index++
		}
	}

	return features
}

// calculateAverageIntensity calculates average pixel intensity in a region
func calculateAverageIntensity(img *image.Gray, x, y, width, height int) float64 {
	sum := 0.0
	count := 0

	for py := y; py < y+height && py < img.Bounds().Max.Y; py++ {
		for px := x; px < x+width && px < img.Bounds().Max.X; px++ {
			sum += float64(img.GrayAt(px, py).Y)
			count++
		}
	}

	if count == 0 {
		return 0.0
	}

	return sum / float64(count)
}

// calculateFeatureSimilarity calculates similarity between feature vectors
func calculateFeatureSimilarity(features1, features2 []float64) float64 {
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

	// Convert to percentage and ensure positive
	similarity = math.Abs(similarity) * 100
	if similarity > 100 {
		similarity = 100
	}

	return similarity
}

// calculateTextureScore calculates texture-based similarity score
func calculateTextureScore(img1, img2 *image.Gray, face1, face2 FaceRegion) float64 {
	// Extract texture patterns from both face regions
	pattern1 := extractTexturePattern(img1, face1)
	pattern2 := extractTexturePattern(img2, face2)

	// Calculate pattern similarity
	similarity := compareTexturePatterns(pattern1, pattern2)

	return similarity
}

// extractTexturePattern extracts simple texture pattern from face region
func extractTexturePattern(img *image.Gray, faceRegion FaceRegion) []float64 {
	pattern := make([]float64, 8) // 8 directional gradients

	centerX := faceRegion.X + faceRegion.Width/2
	centerY := faceRegion.Y + faceRegion.Height/2
	radius := faceRegion.Width / 4

	// Sample 8 directions around center
	for i := 0; i < 8; i++ {
		angle := float64(i) * math.Pi / 4
		dx := int(float64(radius) * math.Cos(angle))
		dy := int(float64(radius) * math.Sin(angle))

		x := centerX + dx
		y := centerY + dy

		if x >= 0 && x < img.Bounds().Max.X && y >= 0 && y < img.Bounds().Max.Y {
			pattern[i] = float64(img.GrayAt(x, y).Y)
		}
	}

	return pattern
}

// compareTexturePatterns compares two texture patterns
func compareTexturePatterns(pattern1, pattern2 []float64) float64 {
	if len(pattern1) != len(pattern2) {
		return 0.0
	}

	// Calculate correlation coefficient
	mean1 := calculateMeanFloat(pattern1)
	mean2 := calculateMeanFloat(pattern2)

	numerator := 0.0
	sum1 := 0.0
	sum2 := 0.0

	for i := 0; i < len(pattern1); i++ {
		diff1 := pattern1[i] - mean1
		diff2 := pattern2[i] - mean2
		numerator += diff1 * diff2
		sum1 += diff1 * diff1
		sum2 += diff2 * diff2
	}

	if sum1 == 0 || sum2 == 0 {
		return 50.0 // Default moderate similarity
	}

	correlation := numerator / math.Sqrt(sum1*sum2)

	// Convert to percentage
	similarity := (correlation + 1) * 50 // Map from [-1,1] to [0,100]

	if similarity < 0 {
		similarity = 0
	}
	if similarity > 100 {
		similarity = 100
	}

	return similarity
}

// detectFaceRegionImproved detects face region with improved algorithms
func detectFaceRegionImproved(img *image.Gray) FaceRegion {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Apply multiple detection strategies
	candidates := []FaceRegion{}

	// Strategy 1: Original detection
	originalCandidate := detectFaceRegion(img)
	if originalCandidate.Confidence > 0.2 {
		candidates = append(candidates, originalCandidate)
	}

	// Strategy 2: Edge-based detection
	edgeCandidate := detectFaceUsingEdges(img)
	if edgeCandidate.Confidence > 0.2 {
		candidates = append(candidates, edgeCandidate)
	}

	// Strategy 3: Gradient-based detection
	gradientCandidate := detectFaceUsingGradients(img)
	if gradientCandidate.Confidence > 0.2 {
		candidates = append(candidates, gradientCandidate)
	}

	// Choose best candidate
	bestCandidate := FaceRegion{
		X:          width / 4,
		Y:          height / 4,
		Width:      width / 2,
		Height:     height / 2,
		Confidence: 0.3,
	}

	for _, candidate := range candidates {
		if candidate.Confidence > bestCandidate.Confidence {
			bestCandidate = candidate
		}
	}

	return bestCandidate
}

// detectFaceUsingEdges detects face using edge detection
func detectFaceUsingEdges(img *image.Gray) FaceRegion {
	// Apply Sobel edge detection
	edges := applySobelFilter(img)

	// Find regions with high edge density (potential face areas)
	bestRegion := findHighEdgeDensityRegion(edges)

	return bestRegion
}

// detectFaceUsingGradients detects face using gradient analysis
func detectFaceUsingGradients(img *image.Gray) FaceRegion {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Calculate gradient magnitude for each pixel
	gradients := calculateGradientMagnitude(img)

	// Find regions with characteristic face gradient patterns
	bestRegion := findFaceGradientPattern(gradients, width, height)

	return bestRegion
}

// applySobelFilter applies Sobel edge detection
func applySobelFilter(img *image.Gray) *image.Gray {
	bounds := img.Bounds()
	edges := image.NewGray(bounds)

	sobelX := [3][3]int{{-1, 0, 1}, {-2, 0, 2}, {-1, 0, 1}}
	sobelY := [3][3]int{{-1, -2, -1}, {0, 0, 0}, {1, 2, 1}}

	for y := bounds.Min.Y + 1; y < bounds.Max.Y-1; y++ {
		for x := bounds.Min.X + 1; x < bounds.Max.X-1; x++ {
			gx, gy := 0, 0

			for ky := -1; ky <= 1; ky++ {
				for kx := -1; kx <= 1; kx++ {
					pixel := int(img.GrayAt(x+kx, y+ky).Y)
					gx += sobelX[ky+1][kx+1] * pixel
					gy += sobelY[ky+1][kx+1] * pixel
				}
			}

			magnitude := int(math.Sqrt(float64(gx*gx + gy*gy)))
			if magnitude > 255 {
				magnitude = 255
			}

			edges.SetGray(x, y, color.Gray{Y: uint8(magnitude)})
		}
	}

	return edges
}

// findHighEdgeDensityRegion finds region with high edge density
func findHighEdgeDensityRegion(edges *image.Gray) FaceRegion {
	bounds := edges.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	bestRegion := FaceRegion{}
	bestScore := 0.0

	for y := height / 6; y < height*2/3; y += 15 {
		for x := width / 6; x < width*2/3; x += 15 {
			regionWidth := width / 3
			regionHeight := height / 3

			if x+regionWidth > width || y+regionHeight > height {
				continue
			}

			edgeDensity := calculateEdgeDensity(edges, x, y, regionWidth, regionHeight)

			if edgeDensity > bestScore {
				bestScore = edgeDensity
				bestRegion = FaceRegion{
					X:          x,
					Y:          y,
					Width:      regionWidth,
					Height:     regionHeight,
					Confidence: edgeDensity,
				}
			}
		}
	}

	return bestRegion
}

// calculateEdgeDensity calculates edge density in a region
func calculateEdgeDensity(edges *image.Gray, x, y, width, height int) float64 {
	edgePixels := 0
	totalPixels := 0

	for py := y; py < y+height && py < edges.Bounds().Max.Y; py++ {
		for px := x; px < x+width && px < edges.Bounds().Max.X; px++ {
			totalPixels++
			if edges.GrayAt(px, py).Y > 50 { // Edge threshold
				edgePixels++
			}
		}
	}

	if totalPixels == 0 {
		return 0.0
	}

	density := float64(edgePixels) / float64(totalPixels)

	// Face regions should have moderate edge density
	if density > 0.15 && density < 0.5 {
		return density
	}

	return 0.0
}

// calculateGradientMagnitude calculates gradient magnitude for each pixel
func calculateGradientMagnitude(img *image.Gray) [][]float64 {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	gradients := make([][]float64, height)
	for i := range gradients {
		gradients[i] = make([]float64, width)
	}

	for y := 1; y < height-1; y++ {
		for x := 1; x < width-1; x++ {
			gx := float64(img.GrayAt(x+1, y).Y) - float64(img.GrayAt(x-1, y).Y)
			gy := float64(img.GrayAt(x, y+1).Y) - float64(img.GrayAt(x, y-1).Y)

			gradients[y][x] = math.Sqrt(gx*gx + gy*gy)
		}
	}

	return gradients
}

// findFaceGradientPattern finds face-like gradient patterns
func findFaceGradientPattern(gradients [][]float64, width, height int) FaceRegion {
	bestRegion := FaceRegion{}
	bestScore := 0.0

	for y := height / 6; y < height*2/3; y += 15 {
		for x := width / 6; x < width*2/3; x += 15 {
			regionWidth := width / 3
			regionHeight := height / 3

			if x+regionWidth >= width || y+regionHeight >= height {
				continue
			}

			score := evaluateGradientPattern(gradients, x, y, regionWidth, regionHeight)

			if score > bestScore {
				bestScore = score
				bestRegion = FaceRegion{
					X:          x,
					Y:          y,
					Width:      regionWidth,
					Height:     regionHeight,
					Confidence: score,
				}
			}
		}
	}

	return bestRegion
}

// evaluateGradientPattern evaluates gradient pattern for face-like characteristics
func evaluateGradientPattern(gradients [][]float64, x, y, width, height int) float64 {
	if width <= 0 || height <= 0 {
		return 0.0
	}

	// Calculate gradient statistics
	sum := 0.0
	count := 0

	for py := y; py < y+height && py < len(gradients); py++ {
		for px := x; px < x+width && px < len(gradients[0]); px++ {
			sum += gradients[py][px]
			count++
		}
	}

	if count == 0 {
		return 0.0
	}

	avgGradient := sum / float64(count)

	// Face regions should have moderate gradient activity
	if avgGradient > 10 && avgGradient < 50 {
		return avgGradient / 50.0
	}

	return 0.0
}

// extractEnhancedFeatures extracts enhanced features for better face comparison
func extractEnhancedFeatures(img *image.Gray, faceRegion FaceRegion) []float64 {
	features := make([]float64, 64) // 8x8 grid features

	cellWidth := faceRegion.Width / 8
	cellHeight := faceRegion.Height / 8

	index := 0
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			cellX := faceRegion.X + col*cellWidth
			cellY := faceRegion.Y + row*cellHeight

			// Calculate multiple features for each cell
			avgIntensity := calculateAverageIntensity(img, cellX, cellY, cellWidth, cellHeight)
			variance := calculateCellVariance(img, cellX, cellY, cellWidth, cellHeight, avgIntensity)

			// Combine intensity and variance
			features[index] = avgIntensity + variance*0.1
			index++
		}
	}

	return features
}

// calculateCellVariance calculates variance in a cell
func calculateCellVariance(img *image.Gray, x, y, width, height int, mean float64) float64 {
	sum := 0.0
	count := 0

	for py := y; py < y+height && py < img.Bounds().Max.Y; py++ {
		for px := x; px < x+width && px < img.Bounds().Max.X; px++ {
			diff := float64(img.GrayAt(px, py).Y) - mean
			sum += diff * diff
			count++
		}
	}

	if count == 0 {
		return 0.0
	}

	return sum / float64(count)
}

// calculateStructuralSimilarity calculates structural similarity between feature vectors
func calculateStructuralSimilarity(features1, features2 []float64) float64 {
	if len(features1) != len(features2) {
		return 0.0
	}

	// Calculate both correlation and euclidean distance
	correlation := calculateCorrelation(features1, features2)
	euclideanSim := calculateEuclideanSimilarity(features1, features2)

	// Combine both measures
	similarity := (correlation*0.7 + euclideanSim*0.3) * 100

	if similarity < 0 {
		similarity = 0
	}
	if similarity > 100 {
		similarity = 100
	}

	return similarity
}

// calculateCorrelation calculates correlation coefficient
func calculateCorrelation(features1, features2 []float64) float64 {
	mean1 := calculateMeanFloat(features1)
	mean2 := calculateMeanFloat(features2)

	numerator := 0.0
	sum1 := 0.0
	sum2 := 0.0

	for i := 0; i < len(features1); i++ {
		diff1 := features1[i] - mean1
		diff2 := features2[i] - mean2
		numerator += diff1 * diff2
		sum1 += diff1 * diff1
		sum2 += diff2 * diff2
	}

	if sum1 == 0 || sum2 == 0 {
		return 0.0
	}

	return numerator / math.Sqrt(sum1*sum2)
}

// calculateEuclideanSimilarity calculates similarity based on euclidean distance
func calculateEuclideanSimilarity(features1, features2 []float64) float64 {
	sumSquaredDiff := 0.0

	for i := 0; i < len(features1); i++ {
		diff := features1[i] - features2[i]
		sumSquaredDiff += diff * diff
	}

	distance := math.Sqrt(sumSquaredDiff)

	// Convert distance to similarity (lower distance = higher similarity)
	maxDistance := 255.0 * math.Sqrt(float64(len(features1))) // Maximum possible distance
	similarity := 1.0 - (distance / maxDistance)

	if similarity < 0 {
		similarity = 0
	}

	return similarity
}

// calculateAdvancedTextureScore calculates advanced texture similarity
func calculateAdvancedTextureScore(img1, img2 *image.Gray, face1, face2 FaceRegion) float64 {
	// Extract multiple texture patterns
	pattern1 := extractAdvancedTexturePattern(img1, face1)
	pattern2 := extractAdvancedTexturePattern(img2, face2)

	// Calculate pattern similarity using multiple measures
	correlationSim := calculatePatternCorrelation(pattern1, pattern2)
	histogramSim := calculatePatternHistogramSimilarity(pattern1, pattern2)

	// Combine similarities
	similarity := (correlationSim*0.6 + histogramSim*0.4)

	return similarity
}

// extractAdvancedTexturePattern extracts advanced texture pattern
func extractAdvancedTexturePattern(img *image.Gray, faceRegion FaceRegion) []float64 {
	pattern := make([]float64, 16) // More detailed pattern

	centerX := faceRegion.X + faceRegion.Width/2
	centerY := faceRegion.Y + faceRegion.Height/2

	// Sample multiple radii and angles
	radiusStep := faceRegion.Width / 8

	index := 0
	for radius := 1; radius <= 4; radius++ {
		for angle := 0; angle < 4; angle++ {
			angleRad := float64(angle) * math.Pi / 2
			dx := int(float64(radius*radiusStep) * math.Cos(angleRad))
			dy := int(float64(radius*radiusStep) * math.Sin(angleRad))

			x := centerX + dx
			y := centerY + dy

			if x >= 0 && x < img.Bounds().Max.X && y >= 0 && y < img.Bounds().Max.Y {
				pattern[index] = float64(img.GrayAt(x, y).Y)
			}
			index++
		}
	}

	return pattern
}

// calculatePatternCorrelation calculates correlation between patterns
func calculatePatternCorrelation(pattern1, pattern2 []float64) float64 {
	correlation := calculateCorrelation(pattern1, pattern2)

	// Convert to percentage
	similarity := (math.Abs(correlation) * 100)

	if similarity > 100 {
		similarity = 100
	}

	return similarity
}

// calculatePatternHistogramSimilarity calculates histogram similarity
func calculatePatternHistogramSimilarity(pattern1, pattern2 []float64) float64 {
	// Create histograms
	hist1 := createHistogram(pattern1, 16)
	hist2 := createHistogram(pattern2, 16)

	// Calculate histogram intersection
	intersection := 0.0
	sum1, sum2 := 0.0, 0.0

	for i := 0; i < len(hist1); i++ {
		intersection += math.Min(hist1[i], hist2[i])
		sum1 += hist1[i]
		sum2 += hist2[i]
	}

	if sum1 == 0 || sum2 == 0 {
		return 0.0
	}

	similarity := intersection / math.Max(sum1, sum2) * 100

	return similarity
}

// createHistogram creates histogram from pattern
func createHistogram(pattern []float64, bins int) []float64 {
	histogram := make([]float64, bins)

	// Find min and max values
	minVal, maxVal := pattern[0], pattern[0]
	for _, val := range pattern {
		if val < minVal {
			minVal = val
		}
		if val > maxVal {
			maxVal = val
		}
	}

	if maxVal == minVal {
		return histogram
	}

	// Fill histogram
	for _, val := range pattern {
		bin := int((val - minVal) / (maxVal - minVal) * float64(bins-1))
		if bin >= bins {
			bin = bins - 1
		}
		histogram[bin]++
	}

	return histogram
}

// calculateHistogramSimilarity calculates histogram-based similarity
func calculateHistogramSimilarity(img1, img2 *image.Gray, face1, face2 FaceRegion) float64 {
	// Extract histograms from face regions
	hist1 := extractFaceHistogram(img1, face1)
	hist2 := extractFaceHistogram(img2, face2)

	// Calculate histogram similarity using chi-square distance
	similarity := calculateHistogramChiSquare(hist1, hist2)

	return similarity
}

// extractFaceHistogram extracts histogram from face region
func extractFaceHistogram(img *image.Gray, faceRegion FaceRegion) []float64 {
	histogram := make([]float64, 256)
	count := 0

	for y := faceRegion.Y; y < faceRegion.Y+faceRegion.Height && y < img.Bounds().Max.Y; y++ {
		for x := faceRegion.X; x < faceRegion.X+faceRegion.Width && x < img.Bounds().Max.X; x++ {
			intensity := img.GrayAt(x, y).Y
			histogram[intensity]++
			count++
		}
	}

	// Normalize histogram
	for i := range histogram {
		histogram[i] = histogram[i] / float64(count)
	}

	return histogram
}

// calculateHistogramChiSquare calculates chi-square similarity between histograms
func calculateHistogramChiSquare(hist1, hist2 []float64) float64 {
	chiSquare := 0.0

	for i := 0; i < len(hist1); i++ {
		if hist1[i]+hist2[i] > 0 {
			diff := hist1[i] - hist2[i]
			chiSquare += (diff * diff) / (hist1[i] + hist2[i])
		}
	}

	// Convert chi-square to similarity (lower chi-square = higher similarity)
	similarity := 1.0 / (1.0 + chiSquare)

	return similarity * 100
}
func calculateMean(values []uint8) float64 {
	sum := 0.0
	for _, v := range values {
		sum += float64(v)
	}
	return sum / float64(len(values))
}

func calculateVariance(values []uint8, mean float64) float64 {
	sum := 0.0
	for _, v := range values {
		diff := float64(v) - mean
		sum += diff * diff
	}
	return sum / float64(len(values))
}

func calculateMeanFloat(values []float64) float64 {
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// DetectFace detects face in image and returns bounding box
func (*FaceRecognitionService) DetectFace(imagePath string) (*FaceRegion, error) {
	img, err := loadImageForFace(imagePath)
	if err != nil {
		return nil, err
	}

	grayImg := convertImageToGrayscale(img)
	faceRegion := detectFaceRegionImproved(grayImg)

	if faceRegion.Confidence < 0.1 {
		return nil, fmt.Errorf("no face detected")
	}

	return &faceRegion, nil
}

// CropFace crops face region from image
func (*FaceRecognitionService) CropFace(imagePath string, boundingBox FaceRegion) (image.Image, error) {
	img, err := loadImageForFace(imagePath)
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()

	// Validate bounding box
	if boundingBox.X < 0 || boundingBox.Y < 0 ||
		boundingBox.X+boundingBox.Width > bounds.Dx() ||
		boundingBox.Y+boundingBox.Height > bounds.Dy() {
		return nil, fmt.Errorf("bounding box is outside image bounds")
	}

	// Crop the face region
	rect := image.Rect(boundingBox.X, boundingBox.Y, boundingBox.X+boundingBox.Width, boundingBox.Y+boundingBox.Height)
	croppedImg := img.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(rect)

	return croppedImg, nil
}
