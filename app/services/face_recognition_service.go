package services

import (
	"errors"
	"fmt"
	"image"
	"os"
)

type FaceRecognitionService struct {
	hogDetector *HOGDetector
}

type FaceMatchResult struct {
	Score  float64
	Status string // match, no_match, error
	Match  bool
}

func NewFaceRecognitionService() *FaceRecognitionService {
	return &FaceRecognitionService{
		hogDetector: NewHOGDetector(),
	}
}

// CompareFaces compares two face images using HOG features
func (f *FaceRecognitionService) CompareFaces(idCardImagePath, selfieImagePath string) (*FaceMatchResult, error) {
	// Validate input images
	if err := f.ValidateImage(idCardImagePath); err != nil {
		return &FaceMatchResult{Score: 0, Status: "error", Match: false},
			fmt.Errorf("ID card image validation failed: %v", err)
	}

	if err := f.ValidateImage(selfieImagePath); err != nil {
		return &FaceMatchResult{Score: 0, Status: "error", Match: false},
			fmt.Errorf("selfie image validation failed: %v", err)
	}

	// Extract HOG features from both images
	fmt.Println("Extracting features from ID card image...")
	idCardFeatures, err := f.hogDetector.ExtractHOGFeatures(idCardImagePath)
	if err != nil {
		return &FaceMatchResult{Score: 0, Status: "error", Match: false},
			fmt.Errorf("failed to extract features from ID card: %v", err)
	}

	fmt.Println("Extracting features from selfie image...")
	selfieFeatures, err := f.hogDetector.ExtractHOGFeatures(selfieImagePath)
	if err != nil {
		return &FaceMatchResult{Score: 0, Status: "error", Match: false},
			fmt.Errorf("failed to extract features from selfie: %v", err)
	}

	// Compare features using HOG detector
	confidence := f.hogDetector.CompareFaceFeatures(idCardFeatures, selfieFeatures)

	// Determine match status based on confidence threshold
	matchThreshold := 0.75 // 75% similarity threshold
	lowThreshold := 0.5    // 50% minimum for no_match vs error

	result := &FaceMatchResult{
		Score: confidence,
		Match: confidence >= matchThreshold,
	}

	if confidence >= matchThreshold {
		result.Status = "match"
	} else if confidence >= lowThreshold {
		result.Status = "no_match"
	} else {
		result.Status = "error"
	}

	fmt.Printf("Face recognition result: Match=%v, Confidence=%.2f, Status=%s\n",
		result.Match, result.Score, result.Status)

	return result, nil
}

func (f *FaceRecognitionService) ExtractFaceFeatures(imagePath string) ([]float64, error) {
	if err := f.ValidateImage(imagePath); err != nil {
		return nil, fmt.Errorf("image validation failed: %v", err)
	}

	fmt.Printf("Extracting HOG features from: %s\n", imagePath)
	features, err := f.hogDetector.ExtractHOGFeatures(imagePath)
	if err != nil {
		return nil, fmt.Errorf("feature extraction failed: %v", err)
	}

	fmt.Printf("Extracted %d features\n", len(features))
	return features, nil
}

func (f *FaceRecognitionService) DetectFaces(imagePath string) (int, error) {
	if err := f.ValidateImage(imagePath); err != nil {
		return 0, fmt.Errorf("image validation failed: %v", err)
	}

	fmt.Printf("Detecting faces in: %s\n", imagePath)
	faces, err := f.hogDetector.DetectFaces(imagePath)
	if err != nil {
		return 0, fmt.Errorf("face detection failed: %v", err)
	}

	fmt.Printf("Detected %d faces\n", len(faces))
	return len(faces), nil
}

func (f *FaceRecognitionService) GetFaceRectangles(imagePath string) ([]image.Rectangle, error) {
	if err := f.ValidateImage(imagePath); err != nil {
		return nil, fmt.Errorf("image validation failed: %v", err)
	}

	return f.hogDetector.DetectFaces(imagePath)
}

func (f *FaceRecognitionService) ValidateImage(imagePath string) error {
	if imagePath == "" {
		return errors.New("image path cannot be empty")
	}

	// Check if file exists
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return fmt.Errorf("image file does not exist: %s", imagePath)
	}

	// Try to decode the image to validate format
	_, err := f.hogDetector.imageProcessor.LoadImage(imagePath)
	if err != nil {
		return fmt.Errorf("invalid image format or corrupted file: %v", err)
	}

	return nil
}

// ValidateImageQuality checks if the image is suitable for face recognition
func (f *FaceRecognitionService) ValidateImageQuality(imagePath string) (bool, string) {
	if err := f.ValidateImage(imagePath); err != nil {
		return false, fmt.Sprintf("Image validation failed: %v", err)
	}

	// Detect faces to validate quality
	faceCount, err := f.DetectFaces(imagePath)
	if err != nil {
		return false, fmt.Sprintf("Face detection failed: %v", err)
	}

	if faceCount == 0 {
		return false, "No faces detected in the image"
	}

	if faceCount > 1 {
		return false, "Multiple faces detected, please use image with single face"
	}

	return true, "Image quality is acceptable for face recognition"
}
