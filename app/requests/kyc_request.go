package requests

type KycUploadIdCardRequest struct {
	Base64Image string `json:"base64_image" binding:"required" validate:"required"`
	IdCardType  string `json:"id_card_type" binding:"required" validate:"required"` // ktp, sim, passport, etc
	UserID      uint   `json:"user_id" binding:"required" validate:"required"`
}

type KycUploadSelfieRequest struct {
	Base64Image       string   `json:"base64_image" binding:"required" validate:"required"`
	KycVerificationID uint     `json:"kyc_verification_id" binding:"required" validate:"required"`
	LivenessFrames    []string `json:"liveness_frames,omitempty"` // Optional: sequence of frames for liveness check
}

type KycProcessRequest struct {
	KycVerificationID uint `json:"kyc_verification_id" binding:"required" validate:"required"`
}

type KycStatusRequest struct {
	Reference string `json:"reference,omitempty"`
	UserID    uint   `json:"user_id,omitempty"`
}
