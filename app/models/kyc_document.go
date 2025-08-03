package models

import (
	"time"

	"gorm.io/gorm"
)

type KycDocument struct {
	ID       uint           `json:"id" gorm:"primaryKey"`
	UserID   uint           `json:"user_id" gorm:"not null"`
	
	// Document Images
	IdCardImagePath string `json:"id_card_image_path"`
	SelfieImagePath string `json:"selfie_image_path"`
	
	// OCR Extracted Data
	NIK           *string `json:"nik"`
	FullName      *string `json:"full_name"`
	PlaceOfBirth  *string `json:"place_of_birth"`
	DateOfBirth   *time.Time `json:"date_of_birth"`
	Gender        *string `json:"gender"`
	Address       *string `json:"address"`
	RtRw          *string `json:"rt_rw"`
	Village       *string `json:"village"`
	District      *string `json:"district"`
	Regency       *string `json:"regency"`
	Province      *string `json:"province"`
	Religion      *string `json:"religion"`
	MaritalStatus *string `json:"marital_status"`
	Occupation    *string `json:"occupation"`
	
	// OCR Results
	OcrConfidence *float64 `json:"ocr_confidence"`
	OcrRawText    *string  `json:"ocr_raw_text"`
	
	// Face Recognition Results
	FaceMatchScore  *float64 `json:"face_match_score"`
	FaceMatchStatus *string  `json:"face_match_status"` // match, no_match, error
	
	// Processing Status
	Status string `json:"status" gorm:"default:pending"` // pending, processing, completed, failed
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
