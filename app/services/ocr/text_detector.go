package ocr

import (
	"image"
	"strings"
)

// TextRegion represents a region containing text
type TextRegion struct {
	X, Y, Width, Height int
	Confidence          float64
}

// TextDetector handles text detection in images
type TextDetector struct{}

// FindTextLinesInImage finds horizontal text lines in the image
func (*TextDetector) FindTextLinesInImage(img *image.Gray) []TextRegion {
	bounds := img.Bounds()
	var lines []TextRegion

	// Scan horizontally for text line patterns
	for y := bounds.Min.Y; y < bounds.Max.Y; y += 3 {
		lineStart := -1
		lineWidth := 0

		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pixel := img.GrayAt(x, y)
			if pixel.Y < 128 { // Dark pixel (potential text)
				if lineStart == -1 {
					lineStart = x
				}
				lineWidth = x - lineStart + 1
			} else if lineStart != -1 && lineWidth > 20 {
				// End of potential text line
				height := 20 // Default text height
				if isLikelyTextLine(img, lineStart, y-5, lineWidth, height) {
					lines = append(lines, TextRegion{
						X:          lineStart,
						Y:          y - 5,
						Width:      lineWidth,
						Height:     height,
						Confidence: 0.8,
					})
				}
				lineStart = -1
				lineWidth = 0
			}
		}

		// Check end of line
		if lineStart != -1 && lineWidth > 20 {
			height := 20
			if isLikelyTextLine(img, lineStart, y-5, lineWidth, height) {
				lines = append(lines, TextRegion{
					X:          lineStart,
					Y:          y - 5,
					Width:      lineWidth,
					Height:     height,
					Confidence: 0.8,
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
			pixel := img.GrayAt(px, py)
			if pixel.Y < 128 {
				blackPixels++
			}
			totalPixels++
		}
	}

	if totalPixels == 0 {
		return false
	}

	// Text lines should have reasonable black pixel density
	density := float64(blackPixels) / float64(totalPixels)
	return density > 0.1 && density < 0.6
}

// GetTextRegions returns predefined regions where text is likely to be found
func (*TextDetector) GetTextRegions(bounds image.Rectangle, idCardType string) []TextRegion {
	width := bounds.Dx()
	height := bounds.Dy()

	var regions []TextRegion

	switch strings.ToLower(idCardType) {
	case "ktp", "e-ktp":
		// KTP specific regions - more precise positioning for Indonesian e-KTP layout
		regions = append(regions, 
			// NIK region (top section, usually around 20-25% from top)
			TextRegion{X: width/8, Y: height/5, Width: width*6/8, Height: height/18, Confidence: 0.95},
			// Alternative NIK position (sometimes higher)
			TextRegion{X: width/8, Y: height/8, Width: width*6/8, Height: height/18, Confidence: 0.9},
			
			// Name region (usually around 30-40% from top)  
			TextRegion{X: width/8, Y: height/3, Width: width*6/8, Height: height/15, Confidence: 0.95},
			// Alternative name position
			TextRegion{X: width/8, Y: height*2/7, Width: width*6/8, Height: height/15, Confidence: 0.9},
			
			// Place/Date of birth (around 45-50% from top)
			TextRegion{X: width/8, Y: height*2/5, Width: width*6/8, Height: height/15, Confidence: 0.85},
			
			// Gender line (around 55% from top)
			TextRegion{X: width/8, Y: height*11/20, Width: width*3/8, Height: height/20, Confidence: 0.8},
			
			// Address region (middle-bottom area, 60-75% from top)
			TextRegion{X: width/8, Y: height*3/5, Width: width*5/8, Height: height/8, Confidence: 0.75},
			
			// RT/RW line (around 75% from top)
			TextRegion{X: width/8, Y: height*3/4, Width: width*4/8, Height: height/20, Confidence: 0.7},
			
			// Religion/Status lines (bottom section)
			TextRegion{X: width/8, Y: height*4/5, Width: width*3/8, Height: height/20, Confidence: 0.7},
			TextRegion{X: width*5/8, Y: height*4/5, Width: width*2/8, Height: height/20, Confidence: 0.7},
		)
	case "sim":
		// SIM specific regions
		regions = append(regions,
			// SIM number area
			TextRegion{X: width/8, Y: height/6, Width: width*3/4, Height: height/12, Confidence: 0.9},
			// Name area
			TextRegion{X: width/8, Y: height/3, Width: width*3/4, Height: height/12, Confidence: 0.9},
			// Additional info
			TextRegion{X: width/8, Y: height/2, Width: width*3/4, Height: height/12, Confidence: 0.8},
		)
	default:
		// Generic regions for unknown card types - scan more areas
		for i := 0; i < 5; i++ {
			yPos := height/6 + (i * height/6)
			regions = append(regions,
				TextRegion{X: width/10, Y: yPos, Width: width*4/5, Height: height/10, Confidence: 0.7},
			)
		}
	}

	return regions
}

// SegmentCharacters segments a text line into individual characters
func (*TextDetector) SegmentCharacters(img *image.Gray) []TextRegion {
	bounds := img.Bounds()
	var characters []TextRegion

	// Calculate vertical projection
	projection := make([]int, bounds.Dx())
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		count := 0
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			pixel := img.GrayAt(x, y)
			if pixel.Y < 128 { // Dark pixel
				count++
			}
		}
		projection[x-bounds.Min.X] = count
	}

	// Find character boundaries
	inChar := false
	charStart := -1
	minCharWidth := 3
	minGap := 2

	for i, count := range projection {
		if count > 0 && !inChar {
			// Start of character
			charStart = i + bounds.Min.X
			inChar = true
		} else if count == 0 && inChar {
			// End of character
			charWidth := i + bounds.Min.X - charStart
			if charWidth >= minCharWidth {
				characters = append(characters, TextRegion{
					X:          charStart,
					Y:          bounds.Min.Y,
					Width:      charWidth,
					Height:     bounds.Dy(),
					Confidence: 0.8,
				})
			}
			inChar = false

			// Skip small gaps
			gapCount := 0
			for j := i + 1; j < len(projection) && projection[j] == 0; j++ {
				gapCount++
				if gapCount >= minGap {
					break
				}
			}
			if gapCount < minGap {
				i += gapCount
			}
		}
	}

	// Handle last character
	if inChar && charStart != -1 {
		charWidth := bounds.Max.X - charStart
		if charWidth >= minCharWidth {
			characters = append(characters, TextRegion{
				X:          charStart,
				Y:          bounds.Min.Y,
				Width:      charWidth,
				Height:     bounds.Dy(),
				Confidence: 0.8,
			})
		}
	}

	return characters
}
