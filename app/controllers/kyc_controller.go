package controllers

import (
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"golang_starter_kit_2025/app/helpers"
	"golang_starter_kit_2025/app/responses"
	"golang_starter_kit_2025/app/services"

	"github.com/gin-gonic/gin"
)

type KycController struct {
	kycService  *services.KycService
	fileService *services.FileService
}

func NewKycController() *KycController {
	return &KycController{
		kycService:  services.NewKycService(),
		fileService: &services.FileService{},
	}
}

// @Summary		Initialize KYC Session
// @Description	Initialize a new KYC verification session for authenticated user
// @Tags			KYC
// @Accept			json
// @Produce		json
// @Security		Bearer
// @Success		200		{object}	helpers.ResponseParams[responses.KycUploadResponse]
// @Failure		400		{object}	helpers.ResponseParams[any]
// @Router			/kyc/initialize [post]
func (c *KycController) InitializeSession(ctx *gin.Context) {
	// Get user ID from JWT token (set by auth middleware)
	userID, exists := ctx.Get("user_id")
	if !exists {
		helpers.ResponseError(ctx, &helpers.ResponseParams[any]{
			Errors:    map[string]string{"auth": "User not authenticated"},
			Message:   "Authentication required",
			Reference: "ERROR-KYC-1",
		}, http.StatusUnauthorized)
		return
	}

	// Initialize KYC session
	kycDoc, err := c.kycService.InitializeKycSession(userID.(uint))
	if err != nil {
		helpers.ResponseError(ctx, &helpers.ResponseParams[any]{
			Errors:    map[string]string{"error": err.Error()},
			Message:   "Failed to initialize KYC session",
			Reference: "ERROR-KYC-2",
		}, http.StatusInternalServerError)
		return
	}

	// Get status
	status, err := c.kycService.GetKycStatus(kycDoc.UserID)
	if err != nil {
		helpers.ResponseError(ctx, &helpers.ResponseParams[any]{
			Errors:    map[string]string{"error": err.Error()},
			Message:   "Failed to get KYC status",
			Reference: "ERROR-KYC-3",
		}, http.StatusInternalServerError)
		return
	}

	helpers.ResponseSuccess(ctx, &helpers.ResponseParams[responses.KycUploadResponse]{
		Item:    status,
		Message: "KYC session initialized successfully",
	}, http.StatusOK)
}

// @Summary		Upload ID Card
// @Description	Upload ID card image for OCR processing
// @Tags			KYC
// @Accept			multipart/form-data
// @Produce		json
// @Security		Bearer
// @Param			id_card		formData	file	true	"ID Card Image"
// @Success		200			{object}	helpers.ResponseParams[responses.KycUploadResponse]
// @Failure		400			{object}	helpers.ResponseParams[any]
// @Router			/kyc/upload-id-card [post]
func (c *KycController) UploadIDCard(ctx *gin.Context) {
	// Get user ID from JWT token
	userID, exists := ctx.Get("user_id")
	if !exists {
		helpers.ResponseError(ctx, &helpers.ResponseParams[any]{
			Errors:    map[string]string{"auth": "User not authenticated"},
			Message:   "Authentication required",
			Reference: "ERROR-KYC-4",
		}, http.StatusUnauthorized)
		return
	}

	// Check file
	file, err := ctx.FormFile("id_card")
	if err != nil {
		helpers.ResponseError(ctx, &helpers.ResponseParams[any]{
			Errors:    map[string]string{"id_card": "ID card image is required"},
			Message:   "ID card image is required",
			Reference: "ERROR-KYC-6",
		}, http.StatusBadRequest)
		return
	}

	// Validate file type
	if !c.isValidImageFile(file.Filename) {
		helpers.ResponseError(ctx, &helpers.ResponseParams[any]{
			Errors:    map[string]string{"id_card": "Invalid file type. Only JPG, JPEG, PNG are allowed"},
			Message:   "Invalid file type",
			Reference: "ERROR-KYC-7",
		}, http.StatusBadRequest)
		return
	}

	// Upload file
	userIDStr := strconv.FormatUint(uint64(userID.(uint)), 10)
	fileName, err := c.fileService.UploadFile(ctx, "id_card", "kyc/"+userIDStr)
	if err != nil {
		helpers.ResponseError(ctx, &helpers.ResponseParams[any]{
			Errors:    map[string]string{"error": err.Error()},
			Message:   "Failed to upload file",
			Reference: "ERROR-KYC-8",
		}, http.StatusInternalServerError)
		return
	}

	// Process ID card
	imagePath := helpers.StoragePath() + "/kyc/" + userIDStr + "/" + *fileName
	kycDoc, err := c.kycService.ProcessIDCard(userID.(uint), imagePath)
	if err != nil {
		helpers.ResponseError(ctx, &helpers.ResponseParams[any]{
			Errors:    map[string]string{"error": err.Error()},
			Message:   "Failed to process ID card",
			Reference: "ERROR-KYC-9",
		}, http.StatusInternalServerError)
		return
	}

	// Get updated status
	status, err := c.kycService.GetKycStatus(kycDoc.UserID)
	if err != nil {
		helpers.ResponseError(ctx, &helpers.ResponseParams[any]{
			Errors:    map[string]string{"error": err.Error()},
			Message:   "Failed to get KYC status",
			Reference: "ERROR-KYC-10",
		}, http.StatusInternalServerError)
		return
	}

	helpers.ResponseSuccess(ctx, &helpers.ResponseParams[responses.KycUploadResponse]{
		Item:    status,
		Message: "ID card uploaded and processed successfully",
	}, http.StatusOK)
}

// @Summary		Upload Selfie
// @Description	Upload selfie image for face recognition
// @Tags			KYC
// @Accept			multipart/form-data
// @Produce		json
// @Security		Bearer
// @Param			selfie		formData	file	true	"Selfie Image"
// @Success		200			{object}	helpers.ResponseParams[responses.KycUploadResponse]
// @Failure		400			{object}	helpers.ResponseParams[any]
// @Router			/kyc/upload-selfie [post]
func (c *KycController) UploadSelfie(ctx *gin.Context) {
	// Get user ID from JWT token
	userID, exists := ctx.Get("user_id")
	if !exists {
		helpers.ResponseError(ctx, &helpers.ResponseParams[any]{
			Errors:    map[string]string{"auth": "User not authenticated"},
			Message:   "Authentication required",
			Reference: "ERROR-KYC-11",
		}, http.StatusUnauthorized)
		return
	}

	// Check file
	file, err := ctx.FormFile("selfie")
	if err != nil {
		helpers.ResponseError(ctx, &helpers.ResponseParams[any]{
			Errors:    map[string]string{"selfie": "Selfie image is required"},
			Message:   "Selfie image is required",
			Reference: "ERROR-KYC-13",
		}, http.StatusBadRequest)
		return
	}

	// Validate file type
	if !c.isValidImageFile(file.Filename) {
		helpers.ResponseError(ctx, &helpers.ResponseParams[any]{
			Errors:    map[string]string{"selfie": "Invalid file type. Only JPG, JPEG, PNG are allowed"},
			Message:   "Invalid file type",
			Reference: "ERROR-KYC-14",
		}, http.StatusBadRequest)
		return
	}

	// Upload file
	userIDStr := strconv.FormatUint(uint64(userID.(uint)), 10)
	fileName, err := c.fileService.UploadFile(ctx, "selfie", "kyc/"+userIDStr)
	if err != nil {
		helpers.ResponseError(ctx, &helpers.ResponseParams[any]{
			Errors:    map[string]string{"error": err.Error()},
			Message:   "Failed to upload file",
			Reference: "ERROR-KYC-15",
		}, http.StatusInternalServerError)
		return
	}

	// Process selfie
	imagePath := helpers.StoragePath() + "/kyc/" + userIDStr + "/" + *fileName
	kycDoc, err := c.kycService.ProcessSelfie(userID.(uint), imagePath)
	if err != nil {
		helpers.ResponseError(ctx, &helpers.ResponseParams[any]{
			Errors:    map[string]string{"error": err.Error()},
			Message:   "Failed to process selfie",
			Reference: "ERROR-KYC-16",
		}, http.StatusInternalServerError)
		return
	}

	// Get updated status
	status, err := c.kycService.GetKycStatus(kycDoc.UserID)
	if err != nil {
		helpers.ResponseError(ctx, &helpers.ResponseParams[any]{
			Errors:    map[string]string{"error": err.Error()},
			Message:   "Failed to get KYC status",
			Reference: "ERROR-KYC-17",
		}, http.StatusInternalServerError)
		return
	}

	helpers.ResponseSuccess(ctx, &helpers.ResponseParams[responses.KycUploadResponse]{
		Item:    status,
		Message: "Selfie uploaded and processed successfully",
	}, http.StatusOK)
}

// @Summary		Get KYC Result
// @Description	Get KYC verification result for authenticated user
// @Tags			KYC
// @Accept			json
// @Produce		json
// @Security		Bearer
// @Success		200			{object}	helpers.ResponseParams[responses.KycResponse]
// @Failure		404			{object}	helpers.ResponseParams[any]
// @Router			/kyc/result [get]
func (c *KycController) GetResult(ctx *gin.Context) {
	// Get user ID from JWT token
	userID, exists := ctx.Get("user_id")
	if !exists {
		helpers.ResponseError(ctx, &helpers.ResponseParams[any]{
			Errors:    map[string]string{"auth": "User not authenticated"},
			Message:   "Authentication required",
			Reference: "ERROR-KYC-18",
		}, http.StatusUnauthorized)
		return
	}

	result, err := c.kycService.GetKycResult(userID.(uint))
	if err != nil {
		helpers.ResponseError(ctx, &helpers.ResponseParams[any]{
			Errors:    map[string]string{"error": err.Error()},
			Message:   "Failed to get KYC result",
			Reference: "ERROR-KYC-19",
		}, http.StatusNotFound)
		return
	}

	helpers.ResponseSuccess(ctx, &helpers.ResponseParams[responses.KycResponse]{
		Item:    result,
		Message: "KYC result retrieved successfully",
	}, http.StatusOK)
}

// @Summary		Get KYC Status
// @Description	Get current KYC session status for authenticated user
// @Tags			KYC
// @Accept			json
// @Produce		json
// @Security		Bearer
// @Success		200			{object}	helpers.ResponseParams[responses.KycUploadResponse]
// @Failure		404			{object}	helpers.ResponseParams[any]
// @Router			/kyc/status [get]
func (c *KycController) GetStatus(ctx *gin.Context) {
	// Get user ID from JWT token
	userID, exists := ctx.Get("user_id")
	if !exists {
		helpers.ResponseError(ctx, &helpers.ResponseParams[any]{
			Errors:    map[string]string{"auth": "User not authenticated"},
			Message:   "Authentication required",
			Reference: "ERROR-KYC-20",
		}, http.StatusUnauthorized)
		return
	}

	status, err := c.kycService.GetKycStatus(userID.(uint))
	if err != nil {
		helpers.ResponseError(ctx, &helpers.ResponseParams[any]{
			Errors:    map[string]string{"error": err.Error()},
			Message:   "Failed to get KYC status",
			Reference: "ERROR-KYC-21",
		}, http.StatusNotFound)
		return
	}

	helpers.ResponseSuccess(ctx, &helpers.ResponseParams[responses.KycUploadResponse]{
		Item:    status,
		Message: "KYC status retrieved successfully",
	}, http.StatusOK)
}

// Helper function to validate image file types
func (c *KycController) isValidImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validExts := []string{".jpg", ".jpeg", ".png"}

	for _, validExt := range validExts {
		if ext == validExt {
			return true
		}
	}
	return false
}
