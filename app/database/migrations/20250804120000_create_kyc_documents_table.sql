-- +++ UP Migration
CREATE TABLE kyc_documents (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    
    -- Document Images
    id_card_image_path VARCHAR(500) NULL,
    selfie_image_path VARCHAR(500) NULL,
    
    -- OCR Extracted Data
    nik VARCHAR(16) NULL,
    full_name VARCHAR(255) NULL,
    place_of_birth VARCHAR(100) NULL,
    date_of_birth DATE NULL,
    gender ENUM('LAKI-LAKI', 'PEREMPUAN') NULL,
    address TEXT NULL,
    rt_rw VARCHAR(20) NULL,
    village VARCHAR(100) NULL,
    district VARCHAR(100) NULL,
    regency VARCHAR(100) NULL,
    province VARCHAR(100) NULL,
    religion VARCHAR(50) NULL,
    marital_status VARCHAR(50) NULL,
    occupation VARCHAR(100) NULL,
    
    -- OCR Results
    ocr_confidence DECIMAL(5,2) NULL,
    ocr_raw_text TEXT NULL,
    
    -- Face Recognition Results
    face_match_score DECIMAL(5,2) NULL,
    face_match_status ENUM('match', 'no_match', 'error') NULL,
    
    -- Processing Status
    status ENUM('pending', 'processing', 'completed', 'failed') DEFAULT 'pending',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    
    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    INDEX idx_nik (nik),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- --- DOWN Migration
DROP TABLE IF EXISTS kyc_documents;
