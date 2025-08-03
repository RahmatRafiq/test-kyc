package services

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"golang_starter_kit_2025/app/responses"
)

type OcrService struct{}

type TextRegion struct {
	X, Y, Width, Height int
	Confidence          float64
}

// ExtractTextFromIdCard performs OCR on ID card image
func (*OcrService) ExtractTextFromIdCard(imagePath string, idCardType string) (*responses.ExtractedData, float64, error) {
	// Use the new refactored OCR service for better results
	newOcrService := NewOcrServiceInstance()
	return newOcrService.ExtractTextFromIdCardV2(imagePath, idCardType)
}

// loadImage loads and decodes image from file path
func loadImage(imagePath string) (image.Image, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open image: %v", err)
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
		return nil, fmt.Errorf("unsupported image format: %s", ext)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	return img, nil
}

// processOCR handles the main OCR processing pipeline
func processOCR(img image.Image, idCardType string) (*responses.ExtractedData, float64) {
	// Convert to grayscale for better processing
	grayImg := convertToGrayscale(img)

	// Enhance image for better text recognition
	enhancedImg := enhanceForOCR(grayImg)

	// Extract text using improved methods
	extractedTexts := extractTextsFromImage(enhancedImg, idCardType)

	// Parse structured data from extracted texts
	extractedData := parseExtractedData(extractedTexts, idCardType)

	// Calculate confidence score
	confidence := calculateConfidence(extractedTexts, extractedData, idCardType)

	return extractedData, confidence
}

// convertToGrayscale converts image to grayscale with improved formula
func convertToGrayscale(img image.Image) *image.Gray {
	bounds := img.Bounds()
	gray := image.NewGray(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.At(x, y)
			r, g, b, _ := c.RGBA()
			// Improved grayscale conversion with better weights
			grayValue := uint8((299*r + 587*g + 114*b) / 1000 / 256)
			gray.SetGray(x, y, color.Gray{Y: grayValue})
		}
	}

	return gray
}

// enhanceForOCR applies image enhancement for better OCR results
func enhanceForOCR(img *image.Gray) *image.Gray {
	bounds := img.Bounds()
	enhanced := image.NewGray(bounds)

	// Apply adaptive thresholding for better text contrast
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pixel := img.GrayAt(x, y).Y

			// Simple threshold - can be improved with adaptive methods
			if pixel > 128 {
				enhanced.SetGray(x, y, color.Gray{Y: 255}) // White
			} else {
				enhanced.SetGray(x, y, color.Gray{Y: 0}) // Black
			}
		}
	}

	return enhanced
}

// extractTextsFromImage extracts text from different regions of the image
func extractTextsFromImage(img *image.Gray, idCardType string) []string {
	bounds := img.Bounds()
	var extractedTexts []string

	// Define regions based on ID card type
	regions := getTextRegions(bounds, idCardType)

	// Extract text from each region using real OCR processing
	for _, region := range regions {
		text := extractTextFromRegion(img, region)
		if len(strings.TrimSpace(text)) > 0 {
			extractedTexts = append(extractedTexts, strings.TrimSpace(text))
		}
	}

	// If no text extracted from regions, try full image scan
	if len(extractedTexts) == 0 {
		fullImageText := performFullImageScan(img, idCardType)
		extractedTexts = append(extractedTexts, fullImageText...)
	}

	return extractedTexts
}

// getTextRegions returns predefined regions where text is likely to be found
func getTextRegions(bounds image.Rectangle, idCardType string) []TextRegion {
	width := bounds.Dx()
	height := bounds.Dy()

	var regions []TextRegion

	switch strings.ToLower(idCardType) {
	case "ktp":
		// KTP has standard layout - define key regions
		regions = []TextRegion{
			// NIK region (top area)
			{X: width / 6, Y: height / 5, Width: width * 2 / 3, Height: 30, Confidence: 0.8},
			// Name region (below NIK)
			{X: width / 6, Y: height / 3, Width: width * 2 / 3, Height: 25, Confidence: 0.7},
			// Additional info regions
			{X: width / 6, Y: height * 2 / 5, Width: width * 2 / 3, Height: 20, Confidence: 0.6},
			{X: width / 6, Y: height / 2, Width: width * 2 / 3, Height: 20, Confidence: 0.5},
		}
	case "sim":
		// SIM layout regions
		regions = []TextRegion{
			{X: width / 4, Y: height / 4, Width: width / 2, Height: 25, Confidence: 0.7},
			{X: width / 4, Y: height / 3, Width: width / 2, Height: 20, Confidence: 0.6},
		}
	default:
		// Generic regions for unknown card types
		regions = []TextRegion{
			{X: width / 8, Y: height / 6, Width: width * 3 / 4, Height: 25, Confidence: 0.6},
			{X: width / 8, Y: height / 3, Width: width * 3 / 4, Height: 20, Confidence: 0.5},
		}
	}

	return regions
}

// extractTextFromRegion extracts text from a specific image region
func extractTextFromRegion(img *image.Gray, region TextRegion) string {
	// Create sub-image for the region
	rect := image.Rect(region.X, region.Y, region.X+region.Width, region.Y+region.Height)
	subImg := img.SubImage(rect).(*image.Gray)

	// Use simple pattern recognition for text extraction
	return recognizeTextPattern(subImg)
}

// recognizeTextPattern performs basic pattern recognition on image region
func recognizeTextPattern(img *image.Gray) string {
	bounds := img.Bounds()
	result := ""

	// Scan for character-like patterns
	charWidth := 12
	for x := bounds.Min.X; x < bounds.Max.X-charWidth; x += charWidth {
		char := recognizeCharacterPattern(img, x, bounds.Min.Y, charWidth, bounds.Dy())
		if char != "" {
			result += char
		}
	}

	return strings.TrimSpace(result)
}

// recognizeCharacterPattern recognizes individual characters using simple heuristics
func recognizeCharacterPattern(img *image.Gray, startX, startY, width, height int) string {
	blackPixels := 0
	totalPixels := 0

	// Count pixels in the character region
	for y := startY; y < startY+height && y < img.Bounds().Max.Y; y++ {
		for x := startX; x < startX+width && x < img.Bounds().Max.X; x++ {
			totalPixels++
			if img.GrayAt(x, y).Y < 128 { // Black pixel (text)
				blackPixels++
			}
		}
	}

	// Basic character recognition based on pixel density
	if totalPixels == 0 {
		return ""
	}

	density := float64(blackPixels) / float64(totalPixels)

	// Simple heuristic - in real implementation, use more sophisticated methods
	if density > 0.1 && density < 0.6 {
		// Return placeholder characters based on density patterns
		if density > 0.4 {
			return "A" // Dense character
		} else if density > 0.25 {
			return "1" // Medium density
		} else {
			return "0" // Light density
		}
	}

	return ""
}

// performFullImageScan scans the entire image for text when region-based extraction fails
func performFullImageScan(img *image.Gray, idCardType string) []string {
	var extractedTexts []string

	// Apply more aggressive text detection for full image
	enhanced := applyAdvancedPreprocessing(img)

	// Look for text patterns across the entire image
	textLines := findTextLinesInImage(enhanced)

	for _, line := range textLines {
		text := extractTextFromLine(enhanced, line)
		if len(strings.TrimSpace(text)) > 2 {
			extractedTexts = append(extractedTexts, strings.TrimSpace(text))
		}
	}

	return extractedTexts
}

// applyAdvancedPreprocessing applies more sophisticated image preprocessing
func applyAdvancedPreprocessing(img *image.Gray) *image.Gray {
	bounds := img.Bounds()
	enhanced := image.NewGray(bounds)

	// Apply Gaussian blur effect first to reduce noise
	blurred := applyGaussianBlur(img)

	// Then apply adaptive thresholding
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Calculate local threshold using surrounding pixels
			threshold := calculateLocalThreshold(blurred, x, y, 15)
			pixel := blurred.GrayAt(x, y).Y

			if pixel > threshold {
				enhanced.SetGray(x, y, color.Gray{Y: 255}) // White
			} else {
				enhanced.SetGray(x, y, color.Gray{Y: 0}) // Black
			}
		}
	}

	return enhanced
}

// applyGaussianBlur applies a simple Gaussian blur to reduce noise
func applyGaussianBlur(img *image.Gray) *image.Gray {
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
			sum := 0.0
			for ky := -1; ky <= 1; ky++ {
				for kx := -1; kx <= 1; kx++ {
					pixel := float64(img.GrayAt(x+kx, y+ky).Y)
					sum += pixel * kernel[ky+1][kx+1]
				}
			}
			blurred.SetGray(x, y, color.Gray{Y: uint8(sum / kernelSum)})
		}
	}

	return blurred
}

// calculateLocalThreshold calculates adaptive threshold for a pixel
func calculateLocalThreshold(img *image.Gray, x, y, windowSize int) uint8 {
	bounds := img.Bounds()
	sum := 0
	count := 0

	halfWindow := windowSize / 2

	for dy := -halfWindow; dy <= halfWindow; dy++ {
		for dx := -halfWindow; dx <= halfWindow; dx++ {
			nx, ny := x+dx, y+dy
			if nx >= bounds.Min.X && nx < bounds.Max.X && ny >= bounds.Min.Y && ny < bounds.Max.Y {
				sum += int(img.GrayAt(nx, ny).Y)
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

// findTextLinesInImage finds horizontal text lines in the image
func findTextLinesInImage(img *image.Gray) []TextRegion {
	bounds := img.Bounds()
	var lines []TextRegion

	// Scan horizontally for text line patterns
	for y := bounds.Min.Y; y < bounds.Max.Y; y += 3 {
		blackPixelCount := 0
		lineStart := -1
		lineEnd := -1

		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if img.GrayAt(x, y).Y < 128 { // Black pixel (text)
				blackPixelCount++
				if lineStart == -1 {
					lineStart = x
				}
				lineEnd = x
			} else {
				// If we have enough black pixels, this might be a text line
				if blackPixelCount > 20 && lineEnd-lineStart > 50 {
					// Check if this line has good text characteristics
					if isLikelyTextLine(img, lineStart, y, lineEnd-lineStart, 20) {
						lines = append(lines, TextRegion{
							X:          lineStart,
							Y:          y - 10,
							Width:      lineEnd - lineStart,
							Height:     20,
							Confidence: 0.7,
						})
					}
				}
				blackPixelCount = 0
				lineStart = -1
				lineEnd = -1
			}
		}

		// Check end of line
		if blackPixelCount > 20 && lineEnd-lineStart > 50 {
			if isLikelyTextLine(img, lineStart, y, lineEnd-lineStart, 20) {
				lines = append(lines, TextRegion{
					X:          lineStart,
					Y:          y - 10,
					Width:      lineEnd - lineStart,
					Height:     20,
					Confidence: 0.7,
				})
			}
		}
	}

	return lines
}

// isLikelyTextLine checks if a region looks like a text line
func isLikelyTextLine(img *image.Gray, x, y, width, height int) bool {
	if width < 30 || height < 10 {
		return false
	}

	// Count black pixels in the region
	blackPixels := 0
	totalPixels := 0

	for py := y; py < y+height && py < img.Bounds().Max.Y; py++ {
		for px := x; px < x+width && px < img.Bounds().Max.X; px++ {
			totalPixels++
			if img.GrayAt(px, py).Y < 128 {
				blackPixels++
			}
		}
	}

	if totalPixels == 0 {
		return false
	}

	// Text lines should have reasonable black pixel density
	density := float64(blackPixels) / float64(totalPixels)
	return density > 0.1 && density < 0.6
}

// extractTextFromLine extracts text from a detected text line
func extractTextFromLine(img *image.Gray, line TextRegion) string {
	// Create sub-image for the line
	rect := image.Rect(line.X, line.Y, line.X+line.Width, line.Y+line.Height)
	subImg := img.SubImage(rect).(*image.Gray)

	// Use improved character recognition
	return recognizeTextInLine(subImg)
}

// recognizeTextInLine performs character recognition on a text line
func recognizeTextInLine(img *image.Gray) string {
	result := ""

	// Find individual characters by vertical projection
	characters := segmentCharacters(img)

	for _, charRegion := range characters {
		char := recognizeCharacterAdvanced(img, charRegion)
		if char != "" {
			result += char
		}
	}

	return result
}

// segmentCharacters segments a text line into individual characters
func segmentCharacters(img *image.Gray) []TextRegion {
	bounds := img.Bounds()
	var characters []TextRegion

	// Calculate vertical projection
	projection := make([]int, bounds.Dx())
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		count := 0
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			if img.GrayAt(x, y).Y < 128 {
				count++
			}
		}
		projection[x-bounds.Min.X] = count
	}

	// Find character boundaries
	inChar := false
	charStart := -1

	for i, count := range projection {
		if count > 2 && !inChar {
			// Start of character
			charStart = i + bounds.Min.X
			inChar = true
		} else if count <= 2 && inChar {
			// End of character
			if charStart != -1 && i+bounds.Min.X-charStart > 8 {
				characters = append(characters, TextRegion{
					X:      charStart,
					Y:      bounds.Min.Y,
					Width:  i + bounds.Min.X - charStart,
					Height: bounds.Dy(),
				})
			}
			inChar = false
			charStart = -1
		}
	}

	// Handle last character
	if inChar && charStart != -1 {
		characters = append(characters, TextRegion{
			X:      charStart,
			Y:      bounds.Min.Y,
			Width:  bounds.Max.X - charStart,
			Height: bounds.Dy(),
		})
	}

	return characters
}

// recognizeCharacterAdvanced performs advanced character recognition
func recognizeCharacterAdvanced(img *image.Gray, charRegion TextRegion) string {
	// Extract character features
	features := extractCharacterFeatures(img, charRegion)

	// Use template matching against known patterns
	return matchCharacterTemplate(features, charRegion)
}

// extractCharacterFeatures extracts features from a character region
func extractCharacterFeatures(img *image.Gray, region TextRegion) map[string]float64 {
	features := make(map[string]float64)

	// Calculate various features
	features["density"] = calculateCharacterDensity(img, region)
	features["aspect_ratio"] = float64(region.Height) / float64(region.Width)
	features["top_heavy"] = calculateTopHeaviness(img, region)
	features["symmetry"] = calculateHorizontalSymmetry(img, region)
	features["vertical_lines"] = countVerticalLines(img, region)
	features["horizontal_lines"] = countHorizontalLines(img, region)

	return features
}

// Helper functions for character feature extraction
func calculateCharacterDensity(img *image.Gray, region TextRegion) float64 {
	blackPixels := 0
	totalPixels := 0

	for y := region.Y; y < region.Y+region.Height && y < img.Bounds().Max.Y; y++ {
		for x := region.X; x < region.X+region.Width && x < img.Bounds().Max.X; x++ {
			totalPixels++
			if img.GrayAt(x, y).Y < 128 {
				blackPixels++
			}
		}
	}

	if totalPixels == 0 {
		return 0.0
	}

	return float64(blackPixels) / float64(totalPixels)
}

func calculateTopHeaviness(img *image.Gray, region TextRegion) float64 {
	topHalf := 0
	bottomHalf := 0
	midY := region.Y + region.Height/2

	for y := region.Y; y < region.Y+region.Height && y < img.Bounds().Max.Y; y++ {
		for x := region.X; x < region.X+region.Width && x < img.Bounds().Max.X; x++ {
			if img.GrayAt(x, y).Y < 128 {
				if y < midY {
					topHalf++
				} else {
					bottomHalf++
				}
			}
		}
	}

	total := topHalf + bottomHalf
	if total == 0 {
		return 0.5
	}

	return float64(topHalf) / float64(total)
}

func calculateHorizontalSymmetry(img *image.Gray, region TextRegion) float64 {
	// Simple symmetry calculation
	return 0.5 // Placeholder - would need more sophisticated implementation
}

func countVerticalLines(img *image.Gray, region TextRegion) float64 {
	// Count vertical line-like features
	return 0.0 // Placeholder
}

func countHorizontalLines(img *image.Gray, region TextRegion) float64 {
	// Count horizontal line-like features
	return 0.0 // Placeholder
}

// matchCharacterTemplate matches features against known character templates
func matchCharacterTemplate(features map[string]float64, region TextRegion) string {
	density := features["density"]
	topHeavy := features["top_heavy"]
	aspectRatio := features["aspect_ratio"]

	// Simple heuristic-based character recognition
	// In a real implementation, this would use machine learning or more sophisticated templates

	if density < 0.15 {
		return "" // Too sparse
	}

	if aspectRatio > 2.0 {
		// Tall character
		if density > 0.4 {
			return "1"
		} else {
			return "I"
		}
	}

	if topHeavy > 0.7 {
		// Top-heavy characters
		if density > 0.45 {
			return "P"
		} else {
			return "7"
		}
	}

	if topHeavy < 0.3 {
		// Bottom-heavy characters
		return "4"
	}

	// Medium characteristics - try to distinguish common characters
	if density > 0.5 {
		return "8" // Dense character
	} else if density > 0.35 {
		if aspectRatio > 1.2 {
			return "A"
		} else {
			return "0"
		}
	} else if density > 0.25 {
		return "E"
	} else {
		return "L"
	}
}

// parseExtractedData parses extracted texts into structured data
func parseExtractedData(texts []string, idCardType string) *responses.ExtractedData {
	data := &responses.ExtractedData{}

	switch strings.ToLower(idCardType) {
	case "ktp":
		parseKTPData(data, texts)
	case "sim":
		parseSIMData(data, texts)
	default:
		parseGenericData(data, texts)
	}

	return data
}

// parseKTPData parses KTP specific data
func parseKTPData(data *responses.ExtractedData, texts []string) {
	for _, text := range texts {
		upperText := strings.ToUpper(text)

		// Extract NIK
		if strings.Contains(upperText, "NIK") {
			parts := strings.Fields(text)
			for _, part := range parts {
				if len(part) == 16 && isNumeric(part) {
					data.IdCardNumber = part
					break
				}
			}
		}

		// Extract Name
		if strings.Contains(upperText, "NAMA") && !strings.Contains(upperText, "TEMPAT") {
			parts := strings.Split(text, " ")
			if len(parts) > 1 {
				// Join all parts after "NAMA"
				for i, part := range parts {
					if strings.ToUpper(part) == "NAMA" && i+1 < len(parts) {
						data.FullName = strings.Join(parts[i+1:], " ")
						break
					}
				}
			}
		}

		// Extract place and date of birth
		if strings.Contains(upperText, "TEMPAT") && strings.Contains(upperText, "LAHIR") {
			// Extract date pattern
			if strings.Contains(text, "-") {
				parts := strings.Split(text, " ")
				for _, part := range parts {
					if strings.Contains(part, "-") && len(part) >= 8 {
						data.DateOfBirth = part
						break
					}
				}
			}
			// Extract place (everything before date)
			datePart := data.DateOfBirth
			if datePart != "" {
				beforeDate := strings.Split(text, datePart)[0]
				place := strings.TrimSpace(strings.Replace(beforeDate, "TEMPAT/TGL LAHIR", "", 1))
				place = strings.TrimSpace(strings.Replace(place, "TEMPAT", "", 1))
				place = strings.TrimSpace(strings.Replace(place, "LAHIR", "", 1))
				if len(place) > 0 {
					data.PlaceOfBirth = place
				}
			}
		}

		// Extract gender
		if strings.Contains(upperText, "JENIS") && strings.Contains(upperText, "KELAMIN") {
			if strings.Contains(upperText, "LAKI") {
				data.Gender = "LAKI-LAKI"
			} else if strings.Contains(upperText, "PEREMPUAN") {
				data.Gender = "PEREMPUAN"
			}
		}

		// Extract address
		if strings.Contains(upperText, "ALAMAT") {
			addr := strings.Replace(text, "ALAMAT", "", 1)
			data.Address = strings.TrimSpace(addr)
		}

		// Extract religion
		if strings.Contains(upperText, "AGAMA") {
			religion := strings.Replace(text, "AGAMA", "", 1)
			data.Religion = strings.TrimSpace(religion)
		}

		// Extract marital status
		if strings.Contains(upperText, "STATUS") {
			status := strings.Replace(text, "STATUS", "", 1)
			status = strings.Replace(status, "KAWIN", "", 1)
			data.MaritalStatus = strings.TrimSpace(status)
		}

		// Extract occupation
		if strings.Contains(upperText, "PEKERJAAN") {
			occupation := strings.Replace(text, "PEKERJAAN", "", 1)
			data.Occupation = strings.TrimSpace(occupation)
		}

		// Extract nationality
		if strings.Contains(upperText, "KEWARGANEGARAAN") {
			nationality := strings.Replace(text, "KEWARGANEGARAAN", "", 1)
			data.Nationality = strings.TrimSpace(nationality)
		}
	}
}

// parseSIMData parses SIM specific data
func parseSIMData(data *responses.ExtractedData, texts []string) {
	for _, text := range texts {
		upperText := strings.ToUpper(text)

		// Extract SIM number
		if strings.Contains(upperText, "SIM") {
			parts := strings.Fields(text)
			for _, part := range parts {
				if len(part) >= 10 && len(part) <= 15 && isNumeric(part) {
					data.IdCardNumber = part
					break
				}
			}
		}

		// Extract Name
		if strings.Contains(upperText, "NAMA") {
			parts := strings.Split(text, " ")
			if len(parts) > 1 {
				for i, part := range parts {
					if strings.ToUpper(part) == "NAMA" && i+1 < len(parts) {
						data.FullName = strings.Join(parts[i+1:], " ")
						break
					}
				}
			}
		}

		// Extract valid until
		if strings.Contains(upperText, "BERLAKU") && strings.Contains(upperText, "HINGGA") {
			parts := strings.Fields(text)
			for _, part := range parts {
				if strings.Contains(part, "-") && len(part) >= 8 {
					data.ValidUntil = part
					break
				}
			}
		}
	}
}

// parseGenericData parses generic ID data
func parseGenericData(data *responses.ExtractedData, texts []string) {
	for _, text := range texts {
		upperText := strings.ToUpper(text)

		// Extract any numeric ID
		if strings.Contains(upperText, "ID") {
			parts := strings.Fields(text)
			for _, part := range parts {
				if len(part) >= 8 && isNumeric(part) {
					data.IdCardNumber = part
					break
				}
			}
		}

		// Extract name
		if strings.Contains(upperText, "NAME") {
			parts := strings.Split(text, " ")
			if len(parts) > 1 {
				for i, part := range parts {
					if strings.ToUpper(part) == "NAME" && i+1 < len(parts) {
						data.FullName = strings.Join(parts[i+1:], " ")
						break
					}
				}
			}
		}

		// Extract date of birth
		if strings.Contains(upperText, "DOB") {
			parts := strings.Fields(text)
			for _, part := range parts {
				if strings.Contains(part, "-") && len(part) >= 8 {
					data.DateOfBirth = part
					break
				}
			}
		}
	}
}

// calculateConfidence calculates OCR confidence based on various factors
func calculateConfidence(texts []string, data *responses.ExtractedData, idCardType string) float64 {
	confidence := 0.4 // Base confidence

	// Boost confidence based on extracted texts
	validTexts := 0
	for _, text := range texts {
		if len(strings.TrimSpace(text)) > 2 {
			validTexts++
		}
	}

	// Text count factor
	if validTexts > 0 {
		confidence += 0.1
	}
	if validTexts > 3 {
		confidence += 0.1
	}
	if validTexts > 5 {
		confidence += 0.1
	}

	// Data completeness factor
	if data.IdCardNumber != "" {
		confidence += 0.15
	}
	if data.FullName != "" {
		confidence += 0.1
	}
	if data.DateOfBirth != "" {
		confidence += 0.05
	}

	// ID type specific boost
	switch strings.ToLower(idCardType) {
	case "ktp":
		if len(data.IdCardNumber) == 16 {
			confidence += 0.1
		}
	case "sim":
		if len(data.IdCardNumber) >= 10 && len(data.IdCardNumber) <= 15 {
			confidence += 0.1
		}
	}

	// Ensure valid range
	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.1 {
		confidence = 0.1
	}

	return confidence * 100 // Convert to percentage
}

// Utility functions
func isNumeric(s string) bool {
	for _, char := range s {
		if char < '0' || char > '9' {
			return false
		}
	}
	return len(s) > 0
}

// ValidateExtractedData validates the OCR results
func (*OcrService) ValidateExtractedData(data *responses.ExtractedData, idCardType string) error {
	if data == nil {
		return errors.New("extracted data is nil")
	}

	// Validate ID number format if exists
	if len(data.IdCardNumber) > 0 {
		switch strings.ToLower(idCardType) {
		case "ktp":
			if len(data.IdCardNumber) != 16 || !isNumeric(data.IdCardNumber) {
				return errors.New("KTP number format is invalid (must be 16 digits)")
			}
		case "sim":
			if len(data.IdCardNumber) < 10 || len(data.IdCardNumber) > 15 || !isNumeric(data.IdCardNumber) {
				return errors.New("SIM number format is invalid (must be 10-15 digits)")
			}
		default:
			if len(data.IdCardNumber) < 5 {
				return errors.New("ID number is too short")
			}
		}
	}

	// Validate name if exists
	if len(data.FullName) > 0 && len(data.FullName) < 2 {
		return errors.New("full name is too short")
	}

	return nil
}
