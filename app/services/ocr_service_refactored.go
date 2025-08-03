package services

import (
	"errors"
	"fmt"
	"image"
	"strings"

	"golang_starter_kit_2025/app/responses"
	imgservice "golang_starter_kit_2025/app/services/image"
	"golang_starter_kit_2025/app/services/ocr"
)

type NewOcrService struct {
	imageProcessor       *imgservice.ImageProcessor
	textDetector        *ocr.TextDetector
	characterRecognizer *ocr.CharacterRecognizer
	dataParser          *ocr.DataParser
}

// NewOcrServiceInstance creates a new OCR service instance
func NewOcrServiceInstance() *NewOcrService {
	return &NewOcrService{
		imageProcessor:       &imgservice.ImageProcessor{},
		textDetector:        &ocr.TextDetector{},
		characterRecognizer: &ocr.CharacterRecognizer{},
		dataParser:          &ocr.DataParser{},
	}
}

// ExtractTextFromIdCardV2 performs OCR on ID card image (new version)
func (s *NewOcrService) ExtractTextFromIdCardV2(imagePath string, idCardType string) (*responses.ExtractedData, float64, error) {
	fmt.Printf("🔍 Starting OCR V2 for image: %s\n", imagePath)
	fmt.Printf("📄 ID Card Type: %s\n", idCardType)
	
	// Load and validate image
	img, err := s.imageProcessor.LoadImage(imagePath)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to load image: %v", err)
	}
	
	fmt.Printf("✅ Image loaded successfully, dimensions: %dx%d\n", img.Bounds().Dx(), img.Bounds().Dy())

	// Process OCR with improved algorithm
	extractedData, confidence := s.processOCR(img, idCardType)
	
	fmt.Printf("🎯 OCR completed - Confidence: %.2f%%\n", confidence)
	fmt.Printf("📊 Extracted NIK: '%s'\n", extractedData.IdCardNumber)
	fmt.Printf("👤 Extracted Name: '%s'\n", extractedData.FullName)

	return extractedData, confidence, nil
}

// processOCR handles the main OCR processing pipeline
func (s *NewOcrService) processOCR(img image.Image, idCardType string) (*responses.ExtractedData, float64) {
	// Convert to grayscale for better processing
	grayImg := s.imageProcessor.ConvertToGrayscale(img)

	// Enhance image for better text recognition
	enhancedImg := s.imageProcessor.EnhanceForOCR(grayImg)

	// Extract text using improved methods
	extractedTexts := s.extractTextsFromImage(enhancedImg, idCardType)

	// Parse structured data from extracted texts
	extractedData := s.dataParser.ParseExtractedData(extractedTexts, idCardType)

	// Calculate confidence score
	confidence := s.calculateConfidence(extractedTexts, extractedData, idCardType)

	return extractedData, confidence
}

// extractTextsFromImage extracts text from different regions of the image
func (s *NewOcrService) extractTextsFromImage(img *image.Gray, idCardType string) []string {
	bounds := img.Bounds()
	var extractedTexts []string

	fmt.Printf("🔎 Scanning image regions for text...\n")

	// Define regions based on ID card type
	regions := s.textDetector.GetTextRegions(bounds, idCardType)
	fmt.Printf("📍 Found %d predefined regions to scan\n", len(regions))

	// Extract text from each region using real OCR processing
	for i, region := range regions {
		text := s.characterRecognizer.ExtractTextFromRegion(img, region)
		
		if strings.TrimSpace(text) != "" {
			fmt.Printf("📝 Region %d (%d,%d %dx%d): '%s'\n", i+1, region.X, region.Y, region.Width, region.Height, text)
			// Apply context-specific improvements
			context := s.determineTextContext(region, idCardType)
			improvedText := s.characterRecognizer.ImproveTextRecognition(text, context)
			fmt.Printf("🔧 Improved text (context: %s): '%s'\n", context, improvedText)
			extractedTexts = append(extractedTexts, improvedText)
		}
	}

	// If no text extracted from regions, try full image scan
	if len(extractedTexts) == 0 {
		fmt.Printf("⚠️  No text found in predefined regions, trying full image scan...\n")
		fullScanTexts := s.performFullImageScan(img, idCardType)
		extractedTexts = append(extractedTexts, fullScanTexts...)
	}

	fmt.Printf("✅ Total extracted texts: %d\n", len(extractedTexts))
	return extractedTexts
}

// performFullImageScan scans the entire image for text when region-based extraction fails
func (s *NewOcrService) performFullImageScan(img *image.Gray, idCardType string) []string {
	var extractedTexts []string

	// Apply more aggressive text detection for full image
	enhanced := s.imageProcessor.ApplyAdvancedPreprocessing(img)

	// Look for text patterns across the entire image
	textLines := s.textDetector.FindTextLinesInImage(enhanced)

	for _, line := range textLines {
		text := s.characterRecognizer.ExtractTextFromLine(enhanced, line)
		
		if strings.TrimSpace(text) != "" {
			extractedTexts = append(extractedTexts, text)
		}
	}

	return extractedTexts
}

// determineTextContext determines the context of text based on region position
func (s *NewOcrService) determineTextContext(region ocr.TextRegion, idCardType string) string {
	// This is a simplified version - in practice, you'd analyze region position
	// to determine if it's likely to contain NIK, name, address, etc.
	
	switch strings.ToLower(idCardType) {
	case "ktp", "e-ktp":
		// For KTP, top regions are usually NIK, middle regions are name
		if region.Y < 100 {
			return "nik"
		} else if region.Y < 200 {
			return "name"
		} else {
			return "address"
		}
	default:
		return "general"
	}
}

// calculateConfidence calculates OCR confidence based on various factors
func (s *NewOcrService) calculateConfidence(texts []string, data *responses.ExtractedData, idCardType string) float64 {
	baseConfidence := 40.0 // Base confidence

	// Boost confidence based on extracted texts
	if len(texts) > 0 {
		baseConfidence += float64(len(texts)) * 5.0
	}

	// Text count factor
	if len(texts) >= 3 {
		baseConfidence += 10.0
	}

	// Data completeness factor
	if data.IdCardNumber != "" {
		baseConfidence += 20.0
	}
	if data.FullName != "" {
		baseConfidence += 15.0
	}

	// ID type specific boost
	switch strings.ToLower(idCardType) {
	case "ktp", "e-ktp":
		if data.IdCardNumber != "" && len(data.IdCardNumber) == 16 {
			baseConfidence += 15.0 // Valid NIK format
		}
	}

	// Ensure valid range
	if baseConfidence > 100.0 {
		baseConfidence = 100.0
	}
	
	return baseConfidence // Convert to percentage
}

// ValidateExtractedDataV2 validates the OCR results
func (s *NewOcrService) ValidateExtractedDataV2(data *responses.ExtractedData, idCardType string) error {
	if data == nil {
		return errors.New("extracted data is nil")
	}

	// Validate ID number format if exists
	if data.IdCardNumber != "" {
		switch strings.ToLower(idCardType) {
		case "ktp", "e-ktp":
			if len(data.IdCardNumber) != 16 {
				return errors.New("invalid NIK length")
			}
		}
	}

	// Validate name if exists
	if data.FullName != "" {
		if len(data.FullName) < 2 {
			return errors.New("name too short")
		}
	}

	return nil
}
