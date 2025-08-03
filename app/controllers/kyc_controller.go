package controllers

import (
	"net/http"
	"strconv"

	"golang_starter_kit_2025/app/helpers"
	"golang_starter_kit_2025/app/requests"
	"golang_starter_kit_2025/app/responses"
	"golang_starter_kit_2025/app/services"

	"github.com/gin-gonic/gin"
)

type KycController struct {
	service *services.KycService
}

func NewKycController(service *services.KycService) *KycController {
	return &KycController{service: service}
}

// @Summary		Upload ID Card
// @Description	API untuk mengupload foto ID card (KTP/SIM/etc) dalam format base64
// @Tags			KYC
// @Accept			json
// @Produce		json
// @Param			body	body		requests.KycUploadIdCardRequest	true	"ID Card Upload Data"
// @Success		200		{object}	helpers.ResponseParams[responses.KycUploadResponse]
// @Router			/kyc/upload-id-card [post]
func (c *KycController) UploadIdCard(ctx *gin.Context) {
	var request requests.KycUploadIdCardRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		helpers.ResponseError(ctx, &helpers.ResponseParams[any]{
			Errors:    map[string]string{"error": err.Error()},
			Message:   "Parameter tidak valid",
			Reference: "KYC-ERROR-1",
		}, http.StatusBadRequest)
		return
	}

	// Validate base64 image
	if len(request.Base64Image) == 0 {
		helpers.ResponseError(ctx, &helpers.ResponseParams[any]{
			Errors:    map[string]string{"base64_image": "base64_image is required"},
			Message:   "Gambar ID card diperlukan",
			Reference: "KYC-ERROR-2",
		}, http.StatusBadRequest)
		return
	}

	// Process upload
	result, err := c.service.UploadIdCard(request)
	if err != nil {
		helpers.ResponseError(ctx, &helpers.ResponseParams[any]{
			Errors:    map[string]string{"error": err.Error()},
			Message:   "Gagal mengupload ID card",
			Reference: "KYC-ERROR-3",
		}, http.StatusInternalServerError)
		return
	}

	helpers.ResponseSuccess(ctx, &helpers.ResponseParams[responses.KycUploadResponse]{
		Item:      result,
		Message:   "ID card berhasil diupload",
		Reference: result.Reference,
	}, http.StatusOK)
}

// @Summary		Upload ID Card File
// @Description	API untuk mengupload foto ID card (KTP/SIM/etc) dalam format file multipart
// @Tags			KYC
// @Accept			multipart/form-data
// @Produce		json
// @Param			id_card_file	formData	file	true	"ID Card Image File"
// @Param			id_card_type	formData	string	true	"ID Card Type (ktp, sim, passport, etc)"
// @Param			user_id			formData	int		true	"User ID"
// @Success		200				{object}	helpers.ResponseParams[responses.KycUploadResponse]
// @Router			/kyc/upload-id-card-file [post]
func (c *KycController) UploadIdCardFile(ctx *gin.Context) {
	// Get form data
	idCardType := ctx.PostForm("id_card_type")
	userIdStr := ctx.PostForm("user_id")

	// Validate form data
	if idCardType == "" {
		helpers.ResponseError(ctx, &helpers.ResponseParams[any]{
			Errors:    map[string]string{"id_card_type": "id_card_type is required"},
			Message:   "Tipe ID card diperlukan",
			Reference: "KYC-ERROR-11",
		}, http.StatusBadRequest)
		return
	}

	if userIdStr == "" {
		helpers.ResponseError(ctx, &helpers.ResponseParams[any]{
			Errors:    map[string]string{"user_id": "user_id is required"},
			Message:   "User ID diperlukan",
			Reference: "KYC-ERROR-12",
		}, http.StatusBadRequest)
		return
	}

	userId, err := strconv.ParseUint(userIdStr, 10, 32)
	if err != nil {
		helpers.ResponseError(ctx, &helpers.ResponseParams[any]{
			Errors:    map[string]string{"user_id": "invalid user_id format"},
			Message:   "Format User ID tidak valid",
			Reference: "KYC-ERROR-13",
		}, http.StatusBadRequest)
		return
	}

	// Get uploaded file
	file, err := ctx.FormFile("id_card_file")
	if err != nil {
		helpers.ResponseError(ctx, &helpers.ResponseParams[any]{
			Errors:    map[string]string{"id_card_file": "id_card_file is required"},
			Message:   "File ID card diperlukan",
			Reference: "KYC-ERROR-14",
		}, http.StatusBadRequest)
		return
	}

	// Process file upload
	result, err := c.service.UploadIdCardFile(file, idCardType, uint(userId))
	if err != nil {
		helpers.ResponseError(ctx, &helpers.ResponseParams[any]{
			Errors:    map[string]string{"error": err.Error()},
			Message:   "Gagal mengupload file ID card",
			Reference: "KYC-ERROR-15",
		}, http.StatusInternalServerError)
		return
	}

	helpers.ResponseSuccess(ctx, &helpers.ResponseParams[responses.KycUploadResponse]{
		Item:      result,
		Message:   "File ID card berhasil diupload",
		Reference: result.Reference,
	}, http.StatusOK)
}

// @Summary		Upload Selfie
// @Description	API untuk mengupload foto selfie dalam format base64 untuk matching dengan ID card
// @Tags			KYC
// @Accept			json
// @Produce		json
// @Param			body	body		requests.KycUploadSelfieRequest	true	"Selfie Upload Data"
// @Success		200		{object}	helpers.ResponseParams[responses.KycUploadResponse]
// @Router			/kyc/upload-selfie [post]
func (c *KycController) UploadSelfie(ctx *gin.Context) {
	var request requests.KycUploadSelfieRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		helpers.ResponseError(ctx, &helpers.ResponseParams[any]{
			Errors:    map[string]string{"error": err.Error()},
			Message:   "Parameter tidak valid",
			Reference: "KYC-ERROR-4",
		}, http.StatusBadRequest)
		return
	}

	// Validate base64 image
	if len(request.Base64Image) == 0 {
		helpers.ResponseError(ctx, &helpers.ResponseParams[any]{
			Errors:    map[string]string{"base64_image": "base64_image is required"},
			Message:   "Gambar selfie diperlukan",
			Reference: "KYC-ERROR-5",
		}, http.StatusBadRequest)
		return
	}

	// Process upload
	result, err := c.service.UploadSelfie(request)
	if err != nil {
		helpers.ResponseError(ctx, &helpers.ResponseParams[any]{
			Errors:    map[string]string{"error": err.Error()},
			Message:   "Gagal mengupload selfie",
			Reference: "KYC-ERROR-6",
		}, http.StatusInternalServerError)
		return
	}

	helpers.ResponseSuccess(ctx, &helpers.ResponseParams[responses.KycUploadResponse]{
		Item:      result,
		Message:   "Selfie berhasil diupload dan sedang diproses",
		Reference: result.Reference,
	}, http.StatusOK)
}

// @Summary		Upload Selfie File
// @Description	API untuk mengupload foto selfie dalam format file multipart untuk matching dengan ID card
// @Tags			KYC
// @Accept			multipart/form-data
// @Produce		json
// @Param			selfie_file			formData	file	true	"Selfie Image File"
// @Param			kyc_verification_id	formData	int		true	"KYC Verification ID"
// @Success		200					{object}	helpers.ResponseParams[responses.KycUploadResponse]
// @Router			/kyc/upload-selfie-file [post]
func (c *KycController) UploadSelfieFile(ctx *gin.Context) {
	// Get form data
	kycVerificationIdStr := ctx.PostForm("kyc_verification_id")

	// Validate form data
	if kycVerificationIdStr == "" {
		helpers.ResponseError(ctx, &helpers.ResponseParams[any]{
			Errors:    map[string]string{"kyc_verification_id": "kyc_verification_id is required"},
			Message:   "KYC Verification ID diperlukan",
			Reference: "KYC-ERROR-16",
		}, http.StatusBadRequest)
		return
	}

	kycVerificationId, err := strconv.ParseUint(kycVerificationIdStr, 10, 32)
	if err != nil {
		helpers.ResponseError(ctx, &helpers.ResponseParams[any]{
			Errors:    map[string]string{"kyc_verification_id": "invalid kyc_verification_id format"},
			Message:   "Format KYC Verification ID tidak valid",
			Reference: "KYC-ERROR-17",
		}, http.StatusBadRequest)
		return
	}

	// Get uploaded file
	file, err := ctx.FormFile("selfie_file")
	if err != nil {
		helpers.ResponseError(ctx, &helpers.ResponseParams[any]{
			Errors:    map[string]string{"selfie_file": "selfie_file is required"},
			Message:   "File selfie diperlukan",
			Reference: "KYC-ERROR-18",
		}, http.StatusBadRequest)
		return
	}

	// Process file upload
	result, err := c.service.UploadSelfieFile(file, uint(kycVerificationId))
	if err != nil {
		helpers.ResponseError(ctx, &helpers.ResponseParams[any]{
			Errors:    map[string]string{"error": err.Error()},
			Message:   "Gagal mengupload file selfie",
			Reference: "KYC-ERROR-19",
		}, http.StatusInternalServerError)
		return
	}

	helpers.ResponseSuccess(ctx, &helpers.ResponseParams[responses.KycUploadResponse]{
		Item:      result,
		Message:   "File selfie berhasil diupload dan sedang diproses",
		Reference: result.Reference,
	}, http.StatusOK)
}

// @Summary		Get KYC Status
// @Description	API untuk mendapatkan status verifikasi KYC berdasarkan reference
// @Tags			KYC
// @Accept			json
// @Produce		json
// @Param			reference	path		string	true	"KYC Reference"
// @Success		200			{object}	helpers.ResponseParams[responses.KycVerificationResponse]
// @Router			/kyc/status/{reference} [get]
func (c *KycController) GetStatus(ctx *gin.Context) {
	reference := ctx.Param("reference")
	if reference == "" {
		helpers.ResponseError(ctx, &helpers.ResponseParams[any]{
			Errors:    map[string]string{"reference": "reference is required"},
			Message:   "Reference KYC diperlukan",
			Reference: "KYC-ERROR-7",
		}, http.StatusBadRequest)
		return
	}

	// Get KYC status
	result, err := c.service.GetKycStatus(reference)
	if err != nil {
		helpers.ResponseError(ctx, &helpers.ResponseParams[any]{
			Errors:    map[string]string{"error": err.Error()},
			Message:   "KYC verification tidak ditemukan",
			Reference: "KYC-ERROR-8",
		}, http.StatusNotFound)
		return
	}

	helpers.ResponseSuccess(ctx, &helpers.ResponseParams[responses.KycVerificationResponse]{
		Item:      result,
		Message:   "Status KYC berhasil didapatkan",
		Reference: reference,
	}, http.StatusOK)
}

// @Summary		Process KYC Verification
// @Description	API untuk memproses verifikasi KYC secara manual (jika auto-process gagal)
// @Tags			KYC
// @Accept			json
// @Produce		json
// @Param			id	path		string	true	"KYC Verification ID"
// @Success		200	{object}	helpers.ResponseParams[any]
// @Router			/kyc/process/{id} [post]
func (c *KycController) ProcessVerification(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		helpers.ResponseError(ctx, &helpers.ResponseParams[any]{
			Errors:    map[string]string{"id": "invalid ID format"},
			Message:   "ID KYC verification tidak valid",
			Reference: "KYC-ERROR-9",
		}, http.StatusBadRequest)
		return
	}

	// Process verification
	err = c.service.ProcessKycVerification(uint(id))
	if err != nil {
		helpers.ResponseError(ctx, &helpers.ResponseParams[any]{
			Errors:    map[string]string{"error": err.Error()},
			Message:   "Gagal memproses verifikasi KYC",
			Reference: "KYC-ERROR-10",
		}, http.StatusInternalServerError)
		return
	}

	helpers.ResponseSuccess(ctx, &helpers.ResponseParams[any]{
		Message:   "Verifikasi KYC berhasil diproses",
		Reference: idStr,
	}, http.StatusOK)
}
