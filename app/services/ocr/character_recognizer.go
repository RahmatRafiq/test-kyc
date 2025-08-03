package ocr

import (
	"image"
	"regexp"
	"strings"
)

// CharacterRecognizer handles character recognition
type CharacterRecognizer struct{}

// ExtractTextFromRegion extracts text from a specific image region
func (cr *CharacterRecognizer) ExtractTextFromRegion(img *image.Gray, region TextRegion) string {
	// Create sub-image for the region
	rect := image.Rect(region.X, region.Y, region.X+region.Width, region.Y+region.Height)
	subImg := img.SubImage(rect).(*image.Gray)

	// Use simple pattern recognition for text extraction
	return cr.recognizeTextPattern(subImg)
}

// ExtractTextFromLine extracts text from a detected text line
func (cr *CharacterRecognizer) ExtractTextFromLine(img *image.Gray, line TextRegion) string {
	// Create sub-image for the line
	rect := image.Rect(line.X, line.Y, line.X+line.Width, line.Y+line.Height)
	subImg := img.SubImage(rect).(*image.Gray)

	// Use improved character recognition
	return cr.recognizeTextInLine(subImg)
}

// recognizeTextPattern performs basic pattern recognition on image region
func (cr *CharacterRecognizer) recognizeTextPattern(img *image.Gray) string {
	bounds := img.Bounds()
	result := ""

	// Scan for character-like patterns
	charWidth := 12
	for x := bounds.Min.X; x < bounds.Max.X-charWidth; x += charWidth {
		char := cr.recognizeCharacterPattern(img, x, bounds.Min.Y, charWidth, bounds.Dy())
		if char != "" {
			result += char
		}
	}

	return strings.TrimSpace(result)
}

// recognizeTextInLine performs character recognition on a text line
func (cr *CharacterRecognizer) recognizeTextInLine(img *image.Gray) string {
	result := ""

	// Find individual characters by vertical projection
	detector := &TextDetector{}
	characters := detector.SegmentCharacters(img)

	for _, charRegion := range characters {
		char := cr.recognizeCharacterAdvanced(img, charRegion)
		if char != "" {
			result += char
		}
	}

	return result
}

// recognizeCharacterPattern recognizes individual characters using simple heuristics
func (cr *CharacterRecognizer) recognizeCharacterPattern(img *image.Gray, startX, startY, width, height int) string {
	blackPixels := 0
	totalPixels := 0

	// Count pixels in the character region
	for y := startY; y < startY+height && y < img.Bounds().Max.Y; y++ {
		for x := startX; x < startX+width && x < img.Bounds().Max.X; x++ {
			pixel := img.GrayAt(x, y)
			if pixel.Y < 128 { // Dark pixel
				blackPixels++
			}
			totalPixels++
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
			return "8" // Dense characters like 8, B, etc.
		} else if density > 0.25 {
			return "A" // Medium density
		} else {
			return "1" // Light characters like 1, I, etc.
		}
	}

	return ""
}

// recognizeCharacterAdvanced performs advanced character recognition
func (cr *CharacterRecognizer) recognizeCharacterAdvanced(img *image.Gray, charRegion TextRegion) string {
	// Extract character features
	features := cr.extractCharacterFeatures(img, charRegion)

	// Use template matching against known patterns
	return cr.matchCharacterTemplate(features, charRegion)
}

// extractCharacterFeatures extracts features from a character region
func (cr *CharacterRecognizer) extractCharacterFeatures(img *image.Gray, region TextRegion) map[string]float64 {
	features := make(map[string]float64)

	// Calculate various features
	features["density"] = cr.calculateCharacterDensity(img, region)
	features["aspect_ratio"] = float64(region.Height) / float64(region.Width)
	features["top_heavy"] = cr.calculateTopHeaviness(img, region)
	features["symmetry"] = cr.calculateHorizontalSymmetry(img, region)
	features["vertical_lines"] = cr.countVerticalLines(img, region)
	features["horizontal_lines"] = cr.countHorizontalLines(img, region)

	return features
}

// Helper functions for character feature extraction
func (cr *CharacterRecognizer) calculateCharacterDensity(img *image.Gray, region TextRegion) float64 {
	blackPixels := 0
	totalPixels := 0

	for y := region.Y; y < region.Y+region.Height && y < img.Bounds().Max.Y; y++ {
		for x := region.X; x < region.X+region.Width && x < img.Bounds().Max.X; x++ {
			pixel := img.GrayAt(x, y)
			if pixel.Y < 128 {
				blackPixels++
			}
			totalPixels++
		}
	}

	if totalPixels == 0 {
		return 0.0
	}

	return float64(blackPixels) / float64(totalPixels)
}

func (cr *CharacterRecognizer) calculateTopHeaviness(img *image.Gray, region TextRegion) float64 {
	topHalf := 0
	bottomHalf := 0
	midY := region.Y + region.Height/2

	for y := region.Y; y < region.Y+region.Height && y < img.Bounds().Max.Y; y++ {
		for x := region.X; x < region.X+region.Width && x < img.Bounds().Max.X; x++ {
			pixel := img.GrayAt(x, y)
			if pixel.Y < 128 {
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

func (cr *CharacterRecognizer) calculateHorizontalSymmetry(img *image.Gray, region TextRegion) float64 {
	// Simple symmetry calculation
	return 0.5 // Placeholder - would need more sophisticated implementation
}

func (cr *CharacterRecognizer) countVerticalLines(img *image.Gray, region TextRegion) float64 {
	// Count vertical line-like features
	return 0.0 // Placeholder
}

func (cr *CharacterRecognizer) countHorizontalLines(img *image.Gray, region TextRegion) float64 {
	// Count horizontal line-like features
	return 0.0 // Placeholder
}

// matchCharacterTemplate matches features against known character templates
func (cr *CharacterRecognizer) matchCharacterTemplate(features map[string]float64, region TextRegion) string {
	density := features["density"]
	aspectRatio := features["aspect_ratio"]
	topHeavy := features["top_heavy"]

	// Prioritize numbers for NIK recognition (KTP contains 16-digit NIK)
	if density > 0.15 && density < 0.7 {
		// Number recognition based on density and shape
		if density > 0.5 { // Dense characters
			if aspectRatio > 1.5 {
				return "8" // Tall and dense
			} else {
				return "6" // Square and dense
			}
		} else if density > 0.35 { // Medium density
			if aspectRatio > 1.8 {
				if topHeavy > 0.6 {
					return "1" // Very tall and top-heavy
				} else {
					return "7" // Tall with bottom weight
				}
			} else if aspectRatio < 0.8 {
				return "0" // Round/square
			} else {
				if topHeavy > 0.6 {
					return "4" // Top-heavy medium char
				} else {
					return "5" // Bottom-heavy medium char
				}
			}
		} else if density > 0.2 { // Light density
			if aspectRatio > 1.8 {
				return "1" // Very thin
			} else if aspectRatio > 1.4 {
				if topHeavy > 0.55 {
					return "2" // Top-heavy
				} else {
					return "3" // Bottom-heavy
				}
			} else {
				return "9" // Medium proportions, light
			}
		} else { // Very light density
			if aspectRatio > 1.5 {
				return "1" // Thin line
			} else {
				return "0" // Light circle
			}
		}
	}

	// Letter recognition for names (Indonesian names)
	if density > 0.1 && density < 0.5 {
		if aspectRatio > 2.0 {
			return "I" // Very tall and thin
		} else if aspectRatio > 1.5 {
			if topHeavy > 0.6 {
				return "T" // Top-heavy tall
			} else {
				return "L" // Bottom-heavy tall
			}
		} else if aspectRatio < 0.7 {
			if density > 0.3 {
				return "O" // Dense round
			} else {
				return "C" // Light round
			}
		} else { // Normal proportions
			if topHeavy > 0.6 {
				if density > 0.3 {
					return "A" // Top-heavy, medium density
				} else {
					return "Y" // Top-heavy, light
				}
			} else if topHeavy < 0.4 {
				if density > 0.3 {
					return "M" // Bottom-heavy, dense
				} else {
					return "U" // Bottom-heavy, light
				}
			} else {
				if density > 0.3 {
					return "H" // Balanced, dense
				} else {
					return "N" // Balanced, light
				}
			}
		}
	}

	// Return empty if no clear match
	return ""
}

// ImproveTextRecognition applies post-processing to improve recognition results
func (cr *CharacterRecognizer) ImproveTextRecognition(text string, context string) string {
	// Remove obvious errors and apply contextual corrections
	result := text
	
	// Remove non-printable characters
	result = regexp.MustCompile(`[^\x20-\x7E]`).ReplaceAllString(result, "")
	
	// Clean up common OCR errors
	result = strings.ReplaceAll(result, "0", "O") // In names, 0 is usually O
	result = strings.ReplaceAll(result, "5", "S") // 5 might be S in names
	
	// Context-specific improvements
	switch strings.ToLower(context) {
	case "nik", "id_number":
		// NIK should only contain numbers
		result = regexp.MustCompile(`[^0-9]`).ReplaceAllString(result, "")
		// NIK should be 16 digits
		if len(result) > 16 {
			result = result[:16]
		}
	case "name":
		// Names should only contain letters and spaces
		result = regexp.MustCompile(`[^a-zA-Z\s]`).ReplaceAllString(result, "")
		// Clean up multiple spaces
		result = regexp.MustCompile(`\s+`).ReplaceAllString(result, " ")
		result = strings.TrimSpace(result)
		result = strings.ToUpper(result)
	}
	
	return result
}
