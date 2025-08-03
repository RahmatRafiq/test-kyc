package responses

import "time"

type KycUploadResponse struct {
	KycVerificationID uint   `json:"kyc_verification_id"`
	Reference         string `json:"reference"`
	Status            string `json:"status"`
	DocumentType      string `json:"document_type"`
	Message           string `json:"message"`
}

type KycVerificationResponse struct {
	ID        uint   `json:"id"`
	Reference string `json:"reference"`
	UserID    uint   `json:"user_id"`
	Status    string `json:"status"`

	// ID Card Information
	IdCardType   string `json:"id_card_type"`
	IdCardNumber string `json:"id_card_number,omitempty"`
	IdCardName   string `json:"id_card_name,omitempty"`

	// Scores (only show if completed)
	OcrConfidence  *float64 `json:"ocr_confidence,omitempty"`
	FaceMatchScore *float64 `json:"face_match_score,omitempty"`
	LivenessScore  *float64 `json:"liveness_score,omitempty"`
	FinalScore     *float64 `json:"final_score,omitempty"`

	// Timestamps
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type KycProcessingResult struct {
	Success           bool           `json:"success"`
	KycVerificationID uint           `json:"kyc_verification_id"`
	Status            string         `json:"status"`
	Scores            *KycScores     `json:"scores,omitempty"`
	ExtractedData     *ExtractedData `json:"extracted_data,omitempty"`
	ErrorDetails      []string       `json:"error_details,omitempty"`
}

type KycScores struct {
	OcrConfidence  float64 `json:"ocr_confidence"`
	FaceMatchScore float64 `json:"face_match_score"`
	HogScore       float64 `json:"hog_score"`
	LbphScore      float64 `json:"lbph_score"`
	EnsembleScore  float64 `json:"ensemble_score"`
	LivenessScore  float64 `json:"liveness_score"`
	FinalScore     float64 `json:"final_score"`
}

type ExtractedData struct {
	IdCardNumber  string `json:"id_card_number"`
	FullName      string `json:"full_name"`
	DateOfBirth   string `json:"date_of_birth,omitempty"`
	PlaceOfBirth  string `json:"place_of_birth,omitempty"`
	Gender        string `json:"gender,omitempty"`
	Address       string `json:"address,omitempty"`
	Religion      string `json:"religion,omitempty"`
	MaritalStatus string `json:"marital_status,omitempty"`
	Occupation    string `json:"occupation,omitempty"`
	Nationality   string `json:"nationality,omitempty"`
	ValidUntil    string `json:"valid_until,omitempty"`
}

type LivenessCheckResult struct {
	HeadNodDetected bool    `json:"head_nod_detected"`
	BackgroundCheck bool    `json:"background_check"`
	FaceStability   bool    `json:"face_stability"`
	OverallScore    float64 `json:"overall_score"`
	Details         string  `json:"details"`
}
