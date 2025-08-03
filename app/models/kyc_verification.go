package models

import (
	"time"

	"gorm.io/gorm"
)

// KycVerification represents the main KYC verification record
type KycVerification struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	Reference string `gorm:"type:varchar(100);uniqueIndex" json:"reference"`
	UserID    uint   `json:"user_id"`
	Status    string `gorm:"type:varchar(50);default:'pending'" json:"status"` // pending, processing, verified, rejected

	// ID Card Information
	IdCardType      string `gorm:"type:varchar(50)" json:"id_card_type"` // ktp, sim, etc
	IdCardNumber    string `gorm:"type:varchar(100)" json:"id_card_number"`
	IdCardName      string `gorm:"type:varchar(255)" json:"id_card_name"`
	IdCardImagePath string `gorm:"type:varchar(500)" json:"id_card_image_path"`

	// Selfie Information
	SelfieImagePath string `gorm:"type:varchar(500)" json:"selfie_image_path"`

	// OCR Results
	OcrConfidence    float64 `gorm:"type:decimal(5,2)" json:"ocr_confidence"`
	OcrExtractedData string  `gorm:"type:text" json:"ocr_extracted_data"` // JSON string

	// Face Recognition Results
	FaceMatchScore float64 `gorm:"type:decimal(5,2)" json:"face_match_score"`
	HogScore       float64 `gorm:"type:decimal(5,2)" json:"hog_score"`
	LbphScore      float64 `gorm:"type:decimal(5,2)" json:"lbph_score"`
	EnsembleScore  float64 `gorm:"type:decimal(5,2)" json:"ensemble_score"`

	// Liveness Check
	LivenessScore  float64 `gorm:"type:decimal(5,2)" json:"liveness_score"`
	LivenessChecks string  `gorm:"type:text" json:"liveness_checks"` // JSON string for detailed checks

	// Final Scores
	FinalScore        float64 `gorm:"type:decimal(5,2)" json:"final_score"`
	VerificationNotes string  `gorm:"type:text" json:"verification_notes"`

	// Timestamps
	ProcessedAt *time.Time     `json:"processed_at"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" swaggerignore:"true"`

	// Relations
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// KycDocument represents individual documents/images in KYC process
type KycDocument struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	KycVerificationID uint           `json:"kyc_verification_id"`
	DocumentType      string         `gorm:"type:varchar(50)" json:"document_type"` // id_card, selfie, additional
	ImagePath         string         `gorm:"type:varchar(500)" json:"image_path"`
	ProcessingStatus  string         `gorm:"type:varchar(50);default:'pending'" json:"processing_status"`
	ProcessingResult  string         `gorm:"type:text" json:"processing_result"` // JSON string
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"deleted_at" swaggerignore:"true"`

	// Relations
	KycVerification KycVerification `gorm:"foreignKey:KycVerificationID" json:"kyc_verification,omitempty"`
}
