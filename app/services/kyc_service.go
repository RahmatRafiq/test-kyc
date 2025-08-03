package services

import (
	"fmt"

	"golang_starter_kit_2025/app/models"
	"golang_starter_kit_2025/app/responses"
	"golang_starter_kit_2025/facades"
)

type KycService struct {
	ocrService  *OcrService
	faceService *FaceRecognitionService
}

func NewKycService() *KycService {
	return &KycService{
		ocrService:  NewOcrService(),
		faceService: NewFaceRecognitionService(),
	}
}

// InitializeKycSession creates a new KYC session for authenticated user
func (s *KycService) InitializeKycSession(userID uint) (*models.KycDocument, error) {
	// Check if user already has a KYC document
	var existingKyc models.KycDocument
	if err := facades.DB.Where("user_id = ?", userID).First(&existingKyc).Error; err == nil {
		// User already has KYC, return existing one
		return &existingKyc, nil
	}

	// Create new KYC document
	kycDoc := &models.KycDocument{
		UserID: userID,
		Status: "pending",
	}

	if err := facades.DB.Create(kycDoc).Error; err != nil {
		return nil, fmt.Errorf("failed to create KYC session: %v", err)
	}

	return kycDoc, nil
}

// ProcessIDCard processes uploaded ID card image
func (s *KycService) ProcessIDCard(userID uint, imagePath string) (*models.KycDocument, error) {
	// Find existing KYC document or create new one
	var kycDoc models.KycDocument
	if err := facades.DB.Where("user_id = ?", userID).First(&kycDoc).Error; err != nil {
		// If not found, create new KYC document
		kycDoc = models.KycDocument{
			UserID: userID,
			Status: "pending",
		}
		if err := facades.DB.Create(&kycDoc).Error; err != nil {
			return nil, fmt.Errorf("failed to create KYC document: %v", err)
		}
	}

	// Update with image path
	kycDoc.IdCardImagePath = imagePath
	kycDoc.Status = "processing"

	// Extract data using OCR
	extractedData, err := s.ocrService.ExtractIDCardData(imagePath)
	if err != nil {
		kycDoc.Status = "failed"
		facades.DB.Save(&kycDoc)
		return nil, fmt.Errorf("OCR extraction failed: %v", err)
	}

	// Update KYC document with extracted data
	s.updateKycDocumentWithOCRData(&kycDoc, extractedData)

	// Save to database
	if err := facades.DB.Save(&kycDoc).Error; err != nil {
		return nil, fmt.Errorf("failed to save OCR results: %v", err)
	}

	return &kycDoc, nil
}

// ProcessSelfie processes uploaded selfie image and performs face matching
func (s *KycService) ProcessSelfie(userID uint, imagePath string) (*models.KycDocument, error) {
	// Find existing KYC document
	var kycDoc models.KycDocument
	if err := facades.DB.Where("user_id = ?", userID).First(&kycDoc).Error; err != nil {
		return nil, fmt.Errorf("please upload ID card first")
	}

	// Check if ID card was processed first
	if kycDoc.IdCardImagePath == "" {
		return nil, fmt.Errorf("ID card must be processed first")
	}

	// Update with selfie image path
	kycDoc.SelfieImagePath = imagePath
	kycDoc.Status = "processing"

	// Perform face matching
	faceResult, err := s.faceService.CompareFaces(kycDoc.IdCardImagePath, imagePath)
	if err != nil {
		kycDoc.Status = "failed"
		facades.DB.Save(&kycDoc)
		return nil, fmt.Errorf("face matching failed: %v", err)
	}

	// Update with face matching results
	kycDoc.FaceMatchScore = &faceResult.Score
	kycDoc.FaceMatchStatus = &faceResult.Status

	// Determine final status
	if faceResult.Match {
		kycDoc.Status = "completed"
	} else {
		kycDoc.Status = "failed"
	}

	// Save to database
	if err := facades.DB.Save(&kycDoc).Error; err != nil {
		return nil, fmt.Errorf("failed to save face matching results: %v", err)
	}

	return &kycDoc, nil
}

// GetKycResult retrieves KYC processing result
func (s *KycService) GetKycResult(userID uint) (*responses.KycResponse, error) {
	var kycDoc models.KycDocument
	if err := facades.DB.Where("user_id = ?", userID).First(&kycDoc).Error; err != nil {
		return nil, fmt.Errorf("no KYC data found for user")
	}

	response := &responses.KycResponse{
		ID:        kycDoc.ID,
		UserID:    kycDoc.UserID,
		Status:    kycDoc.Status,
		CreatedAt: kycDoc.CreatedAt,
		UpdatedAt: kycDoc.UpdatedAt,
	}

	// Add extracted data if available
	if kycDoc.NIK != nil {
		response.ExtractedData = &responses.ExtractedDataResponse{
			NIK:           kycDoc.NIK,
			FullName:      kycDoc.FullName,
			PlaceOfBirth:  kycDoc.PlaceOfBirth,
			DateOfBirth:   kycDoc.DateOfBirth,
			Gender:        kycDoc.Gender,
			Address:       kycDoc.Address,
			RtRw:          kycDoc.RtRw,
			Village:       kycDoc.Village,
			District:      kycDoc.District,
			Regency:       kycDoc.Regency,
			Province:      kycDoc.Province,
			Religion:      kycDoc.Religion,
			MaritalStatus: kycDoc.MaritalStatus,
			Occupation:    kycDoc.Occupation,
			Confidence:    kycDoc.OcrConfidence,
			RawText:       kycDoc.OcrRawText,
		}
	}

	// Add face match result if available
	if kycDoc.FaceMatchScore != nil {
		response.FaceMatchResult = &responses.FaceMatchResponse{
			Score:  kycDoc.FaceMatchScore,
			Status: kycDoc.FaceMatchStatus,
			Match:  kycDoc.FaceMatchStatus != nil && *kycDoc.FaceMatchStatus == "match",
		}
	}

	return response, nil
}

// GetKycStatus gets current status of KYC session
func (s *KycService) GetKycStatus(userID uint) (*responses.KycUploadResponse, error) {
	var kycDoc models.KycDocument
	if err := facades.DB.Where("user_id = ?", userID).First(&kycDoc).Error; err != nil {
		// Return default status if no KYC document exists yet
		return &responses.KycUploadResponse{
			UserID:         userID,
			Status:         "pending",
			IdCardUploaded: false,
			SelfieUploaded: false,
			NextStep:       "upload_id_card",
			Message:        "Please upload your ID card to start KYC verification",
		}, nil
	}

	response := &responses.KycUploadResponse{
		UserID:         kycDoc.UserID,
		Status:         kycDoc.Status,
		IdCardUploaded: kycDoc.IdCardImagePath != "",
		SelfieUploaded: kycDoc.SelfieImagePath != "",
	}

	// Determine next step and message
	if !response.IdCardUploaded {
		response.NextStep = "upload_id_card"
		response.Message = "Please upload your ID card"
	} else if !response.SelfieUploaded {
		response.NextStep = "upload_selfie"
		response.Message = "Please upload your selfie"
	} else if kycDoc.Status == "processing" {
		response.NextStep = "wait"
		response.Message = "Processing your documents..."
	} else if kycDoc.Status == "completed" {
		response.NextStep = "completed"
		response.Message = "KYC verification completed successfully"
	} else if kycDoc.Status == "failed" {
		response.NextStep = "retry"
		response.Message = "KYC verification failed. Please try again."
	}

	return response, nil
}

// updateKycDocumentWithOCRData updates KYC document with OCR extracted data
func (s *KycService) updateKycDocumentWithOCRData(kycDoc *models.KycDocument, data *ExtractedData) {
	kycDoc.NIK = data.NIK
	kycDoc.FullName = data.FullName
	kycDoc.PlaceOfBirth = data.PlaceOfBirth
	kycDoc.DateOfBirth = data.DateOfBirth
	kycDoc.Gender = data.Gender
	kycDoc.Address = data.Address
	kycDoc.RtRw = data.RtRw
	kycDoc.Village = data.Village
	kycDoc.District = data.District
	kycDoc.Regency = data.Regency
	kycDoc.Province = data.Province
	kycDoc.Religion = data.Religion
	kycDoc.MaritalStatus = data.MaritalStatus
	kycDoc.Occupation = data.Occupation
	kycDoc.OcrConfidence = &data.Confidence
	kycDoc.OcrRawText = &data.RawText
}
