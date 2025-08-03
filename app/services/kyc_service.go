package services

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang_starter_kit_2025/app/helpers"
	"golang_starter_kit_2025/app/models"
	"golang_starter_kit_2025/app/requests"
	"golang_starter_kit_2025/app/responses"
	"golang_starter_kit_2025/facades"
)

type KycService struct {
	ocrService             *OcrService
	faceRecognitionService *FaceRecognitionService
	livenessService        *LivenessService
}

func NewKycService() *KycService {
	return &KycService{
		ocrService:             &OcrService{},
		faceRecognitionService: &FaceRecognitionService{},
		livenessService:        &LivenessService{},
	}
}

// UploadIdCard handles ID card upload and initial processing
func (s *KycService) UploadIdCard(request requests.KycUploadIdCardRequest) (*responses.KycUploadResponse, error) {
	// Generate reference for this KYC verification
	reference := helpers.GenerateReference("KYC")

	// Create KYC verification record
	kycVerification := models.KycVerification{
		Reference:  reference,
		UserID:     request.UserID,
		Status:     "pending",
		IdCardType: request.IdCardType,
	}

	// Save to database
	if err := facades.DB.Create(&kycVerification).Error; err != nil {
		return nil, fmt.Errorf("failed to create KYC verification: %v", err)
	}

	// Generate filename and save image
	filename := fmt.Sprintf("idcard_%d_%s.jpg", kycVerification.ID, reference)
	idCardPath := fmt.Sprintf("kyc/%d", kycVerification.UserID)

	// Store base64 image using existing helper
	if err := helpers.StoreBase64File(request.Base64Image, idCardPath, filename); err != nil {
		return nil, fmt.Errorf("failed to store ID card image: %v", err)
	}

	// Update KYC record with image path
	fullImagePath := helpers.StoragePath() + "/" + idCardPath + "/" + filename
	kycVerification.IdCardImagePath = fullImagePath

	if err := facades.DB.Save(&kycVerification).Error; err != nil {
		return nil, fmt.Errorf("failed to update KYC verification: %v", err)
	}

	// Create document record
	document := models.KycDocument{
		KycVerificationID: kycVerification.ID,
		DocumentType:      "id_card",
		ImagePath:         fullImagePath,
		ProcessingStatus:  "pending",
	}

	if err := facades.DB.Create(&document).Error; err != nil {
		return nil, fmt.Errorf("failed to create document record: %v", err)
	}

	response := &responses.KycUploadResponse{
		KycVerificationID: kycVerification.ID,
		Reference:         reference,
		Status:            "pending",
		DocumentType:      "id_card",
		Message:           "ID card uploaded successfully. Please upload selfie image to continue.",
	}

	return response, nil
}

// UploadSelfie handles selfie upload
func (s *KycService) UploadSelfie(request requests.KycUploadSelfieRequest) (*responses.KycUploadResponse, error) {
	// Get existing KYC verification
	var kycVerification models.KycVerification
	if err := facades.DB.First(&kycVerification, request.KycVerificationID).Error; err != nil {
		return nil, fmt.Errorf("KYC verification not found: %v", err)
	}

	// Check if ID card is already uploaded
	if kycVerification.IdCardImagePath == "" {
		return nil, fmt.Errorf("please upload ID card first")
	}

	// Generate filename and save selfie image
	filename := fmt.Sprintf("selfie_%d_%s.jpg", kycVerification.ID, kycVerification.Reference)
	selfiePath := fmt.Sprintf("kyc/%d", kycVerification.UserID)

	// Store base64 image using existing helper
	if err := helpers.StoreBase64File(request.Base64Image, selfiePath, filename); err != nil {
		return nil, fmt.Errorf("failed to store selfie image: %v", err)
	}

	// Update KYC record with selfie path
	fullImagePath := helpers.StoragePath() + "/" + selfiePath + "/" + filename
	kycVerification.SelfieImagePath = fullImagePath
	kycVerification.Status = "processing"

	if err := facades.DB.Save(&kycVerification).Error; err != nil {
		return nil, fmt.Errorf("failed to update KYC verification: %v", err)
	}

	// Create document record
	document := models.KycDocument{
		KycVerificationID: kycVerification.ID,
		DocumentType:      "selfie",
		ImagePath:         fullImagePath,
		ProcessingStatus:  "pending",
	}

	if err := facades.DB.Create(&document).Error; err != nil {
		return nil, fmt.Errorf("failed to create document record: %v", err)
	}

	// Start processing automatically
	go s.ProcessKycVerification(kycVerification.ID)

	response := &responses.KycUploadResponse{
		KycVerificationID: kycVerification.ID,
		Reference:         kycVerification.Reference,
		Status:            "processing",
		DocumentType:      "selfie",
		Message:           "Selfie uploaded successfully. Processing verification...",
	}

	return response, nil
}

// UploadIdCardFile handles ID card file upload
func (s *KycService) UploadIdCardFile(file *multipart.FileHeader, idCardType string, userID uint) (*responses.KycUploadResponse, error) {
	// Validate file type
	if !isValidImageFile(file.Filename) {
		return nil, fmt.Errorf("invalid file type. Only JPEG, JPG, PNG files are allowed")
	}

	// Validate file size (max 10MB)
	if file.Size > 10*1024*1024 {
		return nil, fmt.Errorf("file size too large. Maximum 10MB allowed")
	}

	// Generate reference for this KYC verification
	reference := helpers.GenerateReference("KYC")

	// Create KYC verification record
	kycVerification := models.KycVerification{
		Reference:  reference,
		UserID:     userID,
		Status:     "pending",
		IdCardType: idCardType,
	}

	// Save to database
	if err := facades.DB.Create(&kycVerification).Error; err != nil {
		return nil, fmt.Errorf("failed to create KYC verification: %v", err)
	}

	// Generate filename and path
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("idcard_%d_%s%s", kycVerification.ID, reference, ext)
	idCardPath := fmt.Sprintf("kyc/%d", kycVerification.UserID)

	// Save file using custom file helper
	fullImagePath, err := saveUploadedFile(file, idCardPath, filename)
	if err != nil {
		return nil, fmt.Errorf("failed to save ID card file: %v", err)
	}

	// Update KYC record with image path
	kycVerification.IdCardImagePath = fullImagePath

	if err := facades.DB.Save(&kycVerification).Error; err != nil {
		return nil, fmt.Errorf("failed to update KYC verification: %v", err)
	}

	// Create document record
	document := models.KycDocument{
		KycVerificationID: kycVerification.ID,
		DocumentType:      "id_card",
		ImagePath:         fullImagePath,
		ProcessingStatus:  "pending",
	}

	if err := facades.DB.Create(&document).Error; err != nil {
		return nil, fmt.Errorf("failed to create document record: %v", err)
	}

	response := &responses.KycUploadResponse{
		KycVerificationID: kycVerification.ID,
		Reference:         reference,
		Status:            "pending",
		DocumentType:      "id_card",
		Message:           "ID card file uploaded successfully. Please upload selfie to continue.",
	}

	return response, nil
}

// UploadSelfieFile handles selfie file upload
func (s *KycService) UploadSelfieFile(file *multipart.FileHeader, kycVerificationID uint) (*responses.KycUploadResponse, error) {
	// Validate file type
	if !isValidImageFile(file.Filename) {
		return nil, fmt.Errorf("invalid file type. Only JPEG, JPG, PNG files are allowed")
	}

	// Validate file size (max 10MB)
	if file.Size > 10*1024*1024 {
		return nil, fmt.Errorf("file size too large. Maximum 10MB allowed")
	}

	// Get existing KYC verification
	var kycVerification models.KycVerification
	if err := facades.DB.First(&kycVerification, kycVerificationID).Error; err != nil {
		return nil, fmt.Errorf("KYC verification not found: %v", err)
	}

	// Check if ID card is already uploaded
	if kycVerification.IdCardImagePath == "" {
		return nil, fmt.Errorf("please upload ID card first")
	}

	// Generate filename and path
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("selfie_%d_%s%s", kycVerification.ID, kycVerification.Reference, ext)
	selfiePath := fmt.Sprintf("kyc/%d", kycVerification.UserID)

	// Save file using custom file helper
	fullImagePath, err := saveUploadedFile(file, selfiePath, filename)
	if err != nil {
		return nil, fmt.Errorf("failed to save selfie file: %v", err)
	}

	// Update KYC record with selfie path
	kycVerification.SelfieImagePath = fullImagePath
	kycVerification.Status = "processing"

	if err := facades.DB.Save(&kycVerification).Error; err != nil {
		return nil, fmt.Errorf("failed to update KYC verification: %v", err)
	}

	// Create document record
	document := models.KycDocument{
		KycVerificationID: kycVerification.ID,
		DocumentType:      "selfie",
		ImagePath:         fullImagePath,
		ProcessingStatus:  "pending",
	}

	if err := facades.DB.Create(&document).Error; err != nil {
		return nil, fmt.Errorf("failed to create document record: %v", err)
	}

	// Start processing automatically
	go s.ProcessKycVerification(kycVerification.ID)

	response := &responses.KycUploadResponse{
		KycVerificationID: kycVerification.ID,
		Reference:         kycVerification.Reference,
		Status:            "processing",
		DocumentType:      "selfie",
		Message:           "Selfie file uploaded successfully. Processing verification...",
	}

	return response, nil
}

// ProcessKycVerification processes the complete KYC verification
func (s *KycService) ProcessKycVerification(kycVerificationID uint) error {
	// Get KYC verification
	var kycVerification models.KycVerification
	if err := facades.DB.First(&kycVerification, kycVerificationID).Error; err != nil {
		return fmt.Errorf("KYC verification not found: %v", err)
	}

	// Check if both images are uploaded
	if kycVerification.IdCardImagePath == "" || kycVerification.SelfieImagePath == "" {
		return fmt.Errorf("both ID card and selfie images are required")
	}

	// Update status to processing
	kycVerification.Status = "processing"
	facades.DB.Save(&kycVerification)

	// Step 1: Perform OCR on ID card
	extractedData, ocrConfidence, err := s.ocrService.ExtractTextFromIdCard(
		kycVerification.IdCardImagePath,
		kycVerification.IdCardType,
	)
	if err != nil {
		s.updateKycStatus(kycVerificationID, "rejected", fmt.Sprintf("OCR failed: %v", err))
		return err
	}

	// Validate extracted data
	if err := s.ocrService.ValidateExtractedData(extractedData, kycVerification.IdCardType); err != nil {
		s.updateKycStatus(kycVerificationID, "rejected", fmt.Sprintf("OCR validation failed: %v", err))
		return err
	}

	// Step 2: Perform face recognition
	faceScores, err := s.faceRecognitionService.CompareFaces(
		kycVerification.IdCardImagePath,
		kycVerification.SelfieImagePath,
	)
	if err != nil {
		s.updateKycStatus(kycVerificationID, "rejected", fmt.Sprintf("Face recognition failed: %v", err))
		return err
	}

	// Step 3: Calculate final score (tanpa liveness dulu)
	finalScore := s.calculateFinalScore(ocrConfidence, faceScores.EnsembleScore, 0) // liveness = 0 dulu

	// Step 4: Determine verification result
	status := s.determineVerificationStatus(finalScore, ocrConfidence, faceScores.EnsembleScore)

	// Step 5: Update database with results
	extractedDataJSON, _ := json.Marshal(extractedData)

	kycVerification.IdCardNumber = extractedData.IdCardNumber
	kycVerification.IdCardName = extractedData.FullName
	kycVerification.OcrConfidence = ocrConfidence
	kycVerification.OcrExtractedData = string(extractedDataJSON)
	kycVerification.FaceMatchScore = faceScores.FaceMatchScore
	kycVerification.HogScore = faceScores.HogScore
	kycVerification.LbphScore = faceScores.LbphScore
	kycVerification.EnsembleScore = faceScores.EnsembleScore
	kycVerification.LivenessScore = 0 // Skip liveness for now
	kycVerification.FinalScore = finalScore
	kycVerification.Status = status
	kycVerification.VerificationNotes = s.generateVerificationNotes(ocrConfidence, faceScores.EnsembleScore, 0)

	now := time.Now()
	kycVerification.ProcessedAt = &now

	if err := facades.DB.Save(&kycVerification).Error; err != nil {
		return fmt.Errorf("failed to update KYC verification: %v", err)
	}

	return nil
}

// GetKycStatus returns the current status of KYC verification
func (s *KycService) GetKycStatus(reference string) (*responses.KycVerificationResponse, error) {
	var kycVerification models.KycVerification
	if err := facades.DB.Where("reference = ?", reference).First(&kycVerification).Error; err != nil {
		return nil, fmt.Errorf("KYC verification not found")
	}

	response := &responses.KycVerificationResponse{
		ID:           kycVerification.ID,
		Reference:    kycVerification.Reference,
		UserID:       kycVerification.UserID,
		Status:       kycVerification.Status,
		IdCardType:   kycVerification.IdCardType,
		IdCardNumber: kycVerification.IdCardNumber,
		IdCardName:   kycVerification.IdCardName,
		CreatedAt:    kycVerification.CreatedAt,
		UpdatedAt:    kycVerification.UpdatedAt,
		ProcessedAt:  kycVerification.ProcessedAt,
	}

	// Add scores if verification is completed
	if kycVerification.Status == "verified" || kycVerification.Status == "rejected" {
		response.OcrConfidence = &kycVerification.OcrConfidence
		response.FaceMatchScore = &kycVerification.FaceMatchScore
		response.LivenessScore = &kycVerification.LivenessScore
		response.FinalScore = &kycVerification.FinalScore
	}

	return response, nil
}

// Helper methods

func (s *KycService) updateKycStatus(kycVerificationID uint, status, notes string) {
	var kycVerification models.KycVerification
	if err := facades.DB.First(&kycVerification, kycVerificationID).Error; err != nil {
		return
	}

	kycVerification.Status = status
	kycVerification.VerificationNotes = notes
	now := time.Now()
	kycVerification.ProcessedAt = &now

	facades.DB.Save(&kycVerification)
}

func (s *KycService) calculateFinalScore(ocrConfidence, faceMatchScore, livenessScore float64) float64 {
	// Weighted scoring without liveness
	ocrWeight := 0.3
	faceWeight := 0.7
	// livenessWeight := 0.2  // Skip for now

	finalScore := (ocrConfidence * ocrWeight) + (faceMatchScore * faceWeight)

	if finalScore > 100.0 {
		finalScore = 100.0
	}
	if finalScore < 0.0 {
		finalScore = 0.0
	}

	return finalScore
}

func (s *KycService) determineVerificationStatus(finalScore, ocrConfidence, faceMatchScore float64) string {
	// Thresholds for verification
	minFinalScore := 70.0
	minOcrConfidence := 60.0
	minFaceMatchScore := 65.0

	if finalScore >= minFinalScore &&
		ocrConfidence >= minOcrConfidence &&
		faceMatchScore >= minFaceMatchScore {
		return "verified"
	} else if finalScore >= 50.0 && ocrConfidence >= 40.0 && faceMatchScore >= 40.0 {
		return "pending" // Needs manual review
	} else {
		return "rejected"
	}
}

func (s *KycService) generateVerificationNotes(ocrConfidence, faceMatchScore, livenessScore float64) string {
	notes := fmt.Sprintf("OCR Confidence: %.2f%%, Face Match: %.2f%%",
		ocrConfidence, faceMatchScore)

	if ocrConfidence < 60.0 {
		notes += " | Low OCR confidence"
	}
	if faceMatchScore < 65.0 {
		notes += " | Low face match score"
	}

	return notes
}

// Helper functions for file upload

// isValidImageFile checks if the uploaded file is a valid image
func isValidImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validExtensions := []string{".jpg", ".jpeg", ".png"}

	for _, validExt := range validExtensions {
		if ext == validExt {
			return true
		}
	}
	return false
}

// saveUploadedFile saves the uploaded file to storage directory
func saveUploadedFile(file *multipart.FileHeader, path, filename string) (string, error) {
	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %v", err)
	}
	defer src.Close()

	// Ensure storage path exists
	fullPath := helpers.StoragePath() + "/" + path
	if err := os.MkdirAll(fullPath, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create storage directory: %v", err)
	}

	// Create destination file
	fullImagePath := fullPath + "/" + filename
	dst, err := os.Create(fullImagePath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %v", err)
	}
	defer dst.Close()

	// Copy file contents
	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("failed to copy file: %v", err)
	}

	return fullImagePath, nil
}
