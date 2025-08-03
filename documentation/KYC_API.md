# KYC API Documentation

## Overview
Sistem KYC (Know Your Customer) untuk verifikasi identitas menggunakan foto ID card dan selfie dengan matching menggunakan algoritma HOG dan LBPH.

## Features
- ✅ Upload foto ID card (base64)
- ✅ Upload foto selfie (base64) 
- ✅ OCR ekstraksi data dari ID card
- ✅ Face recognition matching (HOG + LBPH)
- ✅ Ensemble scoring untuk akurasi tinggi
- ✅ Status tracking verifikasi

## API Endpoints

### 1. Upload ID Card (Base64)

**POST** `/kyc/upload-id-card`

Upload foto ID card dalam format base64.

**Headers:**
```
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "base64_image": "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQAAAQ...",
  "id_card_type": "ktp",
  "user_id": 1
}
```

**Response:**
```json
{
  "status": "success",
  "item": {
    "kyc_verification_id": 1,
    "reference": "KYC-20250803-001",
    "status": "pending",
    "document_type": "id_card",
    "message": "ID card uploaded successfully. Please upload selfie image to continue."
  },
  "message": "ID card berhasil diupload",
  "reference": "KYC-20250803-001"
}
```

### 2. Upload ID Card (File)

**POST** `/kyc/upload-id-card-file`

Upload foto ID card dalam format file multipart.

**Headers:**
```
Authorization: Bearer <jwt_token>
Content-Type: multipart/form-data
```

**Form Data:**
```
id_card_file: [File] (JPEG/PNG, max 10MB)
id_card_type: ktp
user_id: 1
```

**Response:**
```json
{
  "status": "success",
  "item": {
    "kyc_verification_id": 1,
    "reference": "KYC-20250803-001",
    "status": "pending",
    "document_type": "id_card",
    "message": "ID card file uploaded successfully. Please upload selfie to continue."
  },
  "message": "File ID card berhasil diupload",
  "reference": "KYC-20250803-001"
}
```

### 3. Upload Selfie (Base64)

**POST** `/kyc/upload-selfie`

Upload foto selfie untuk matching dengan ID card dalam format base64.

**Headers:**
```
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "base64_image": "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQAAAQ...",
  "kyc_verification_id": 1
}
```

**Response:**
```json
{
  "status": "success",
  "item": {
    "kyc_verification_id": 1,
    "reference": "KYC-20250803-001",
    "status": "processing",
    "document_type": "selfie",
    "message": "Selfie uploaded successfully. Processing verification..."
  },
  "message": "Selfie berhasil diupload dan sedang diproses",
  "reference": "KYC-20250803-001"
}
```

### 4. Upload Selfie (File)

**POST** `/kyc/upload-selfie-file`

Upload foto selfie untuk matching dengan ID card dalam format file multipart.

**Headers:**
```
Authorization: Bearer <jwt_token>
Content-Type: multipart/form-data
```

**Form Data:**
```
selfie_file: [File] (JPEG/PNG, max 10MB)
kyc_verification_id: 1
```

**Response:**
```json
{
  "status": "success",
  "item": {
    "kyc_verification_id": 1,
    "reference": "KYC-20250803-001",
    "status": "processing",
    "document_type": "selfie",
    "message": "Selfie file uploaded successfully. Processing verification..."
  },
  "message": "File selfie berhasil diupload dan sedang diproses",
  "reference": "KYC-20250803-001"
}
```

### 5. Get KYC Status

**GET** `/kyc/status/{reference}`

Mendapatkan status verifikasi KYC.

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "status": "success",
  "item": {
    "id": 1,
    "reference": "KYC-20250803-001",
    "user_id": 1,
    "status": "verified",
    "id_card_type": "ktp",
    "id_card_number": "3171012345678901",
    "id_card_name": "NAMA SESUAI KTP",
    "ocr_confidence": 85.5,
    "face_match_score": 92.3,
    "liveness_score": 0,
    "final_score": 89.8,
    "processed_at": "2025-08-03T10:30:00Z",
    "created_at": "2025-08-03T10:25:00Z",
    "updated_at": "2025-08-03T10:30:00Z"
  },
  "message": "Status KYC berhasil didapatkan",
  "reference": "KYC-20250803-001"
}
```

### 4. Manual Process Verification

**POST** `/kyc/process/{id}`

Memproses verifikasi KYC secara manual (jika auto-process gagal).

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "status": "success",
  "message": "Verifikasi KYC berhasil diproses",
  "reference": "1"
}
```

## Status Codes

| Status | Description |
|--------|-------------|
| `pending` | Menunggu upload dokumen |
| `processing` | Sedang memproses verifikasi |
| `verified` | Verifikasi berhasil |
| `rejected` | Verifikasi ditolak |

## Scoring Algorithm

### OCR Confidence
- **Range:** 0-100%
- **Factors:** Image quality, text detection, card type
- **Threshold:** Minimum 60% untuk pass

### Face Match Score  
- **Algorithm:** HOG (70%) + LBPH (30%) ensemble
- **Range:** 0-100%
- **Threshold:** Minimum 65% untuk pass

### Final Score
- **Formula:** (OCR × 30%) + (Face Match × 70%)
- **Threshold:** 
  - ≥70% = Verified
  - 50-69% = Pending (manual review)
  - <50% = Rejected

## Error Codes

| Code | Message |
|------|---------|
| KYC-ERROR-1 | Parameter tidak valid |
| KYC-ERROR-2 | Gambar ID card diperlukan |
| KYC-ERROR-3 | Gagal mengupload ID card |
| KYC-ERROR-4 | Parameter tidak valid |
| KYC-ERROR-5 | Gambar selfie diperlukan |
| KYC-ERROR-6 | Gagal mengupload selfie |
| KYC-ERROR-7 | Reference KYC diperlukan |
| KYC-ERROR-8 | KYC verification tidak ditemukan |
| KYC-ERROR-9 | ID KYC verification tidak valid |
| KYC-ERROR-10 | Gagal memproses verifikasi KYC |

## Implementation Notes

### Image Requirements
- **Format:** JPEG, PNG
- **Size:** Minimum 640x480px (recommended 800x600px+)
- **Quality:** High contrast, clear text
- **Base64:** Include data URL prefix

### Face Recognition
- **HOG Features:** Histogram of Oriented Gradients
- **LBPH Features:** Local Binary Patterns Histogram  
- **Cosine Similarity:** For feature comparison
- **One-shot Learning:** Single template matching

### Security
- All endpoints require JWT authentication
- File storage uses secure paths
- Image processing is done server-side
- No client-side ML processing required

## Database Schema

### kyc_verifications
```sql
- id (BIGSERIAL PRIMARY KEY)
- reference (VARCHAR UNIQUE)
- user_id (BIGINT)
- status (VARCHAR) 
- id_card_type (VARCHAR)
- id_card_number (VARCHAR)
- id_card_name (VARCHAR)
- id_card_image_path (VARCHAR)
- selfie_image_path (VARCHAR)
- ocr_confidence (DECIMAL)
- face_match_score (DECIMAL)
- hog_score (DECIMAL)
- lbph_score (DECIMAL)
- ensemble_score (DECIMAL)
- final_score (DECIMAL)
- verification_notes (TEXT)
- processed_at (TIMESTAMP)
- created_at (TIMESTAMP)
- updated_at (TIMESTAMP)
```

### kyc_documents
```sql
- id (BIGSERIAL PRIMARY KEY)
- kyc_verification_id (BIGINT FK)
- document_type (VARCHAR)
- image_path (VARCHAR)
- processing_status (VARCHAR)
- processing_result (TEXT)
- created_at (TIMESTAMP)
- updated_at (TIMESTAMP)
```
