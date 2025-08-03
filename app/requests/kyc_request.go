package requests

// KycUploadRequest untuk upload dokumen KYC
// User ID akan diambil dari JWT token yang sudah ter-authenticate
type KycUploadRequest struct {
	// Tidak perlu field lain karena user_id akan diambil dari middleware auth
}
