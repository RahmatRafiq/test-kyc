package responses

import "time"

type KycResponse struct {
	ID              uint                   `json:"id"`
	UserID          uint                   `json:"user_id"`
	Status          string                 `json:"status"`
	ExtractedData   *ExtractedDataResponse `json:"extracted_data,omitempty"`
	FaceMatchResult *FaceMatchResponse     `json:"face_match_result,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

type ExtractedDataResponse struct {
	NIK           *string    `json:"nik"`
	FullName      *string    `json:"full_name"`
	PlaceOfBirth  *string    `json:"place_of_birth"`
	DateOfBirth   *time.Time `json:"date_of_birth"`
	Gender        *string    `json:"gender"`
	Address       *string    `json:"address"`
	RtRw          *string    `json:"rt_rw"`
	Village       *string    `json:"village"`
	District      *string    `json:"district"`
	Regency       *string    `json:"regency"`
	Province      *string    `json:"province"`
	Religion      *string    `json:"religion"`
	MaritalStatus *string    `json:"marital_status"`
	Occupation    *string    `json:"occupation"`
	Confidence    *float64   `json:"confidence"`
	RawText       *string    `json:"raw_text,omitempty"`
}

type FaceMatchResponse struct {
	Score  *float64 `json:"score"`
	Status *string  `json:"status"`
	Match  bool     `json:"match"`
}

type KycUploadResponse struct {
	UserID         uint   `json:"user_id"`
	Status         string `json:"status"`
	Message        string `json:"message"`
	NextStep       string `json:"next_step"`
	IdCardUploaded bool   `json:"id_card_uploaded"`
	SelfieUploaded bool   `json:"selfie_uploaded"`
}
