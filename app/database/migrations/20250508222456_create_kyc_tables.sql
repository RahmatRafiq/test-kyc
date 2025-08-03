-- Create KYC verifications table
CREATE TABLE IF NOT EXISTS kyc_verifications (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    reference VARCHAR(100) UNIQUE NOT NULL,
    user_id BIGINT NOT NULL,
    status VARCHAR(50) DEFAULT 'pending' NOT NULL,
    
    -- ID Card Information
    id_card_type VARCHAR(50) NOT NULL,
    id_card_number VARCHAR(100),
    id_card_name VARCHAR(255),
    id_card_image_path VARCHAR(500),
    
    -- Selfie Information
    selfie_image_path VARCHAR(500),
    
    -- OCR Results
    ocr_confidence DECIMAL(5,2) DEFAULT 0,
    ocr_extracted_data TEXT,
    
    -- Face Recognition Results
    face_match_score DECIMAL(5,2) DEFAULT 0,
    hog_score DECIMAL(5,2) DEFAULT 0,
    lbph_score DECIMAL(5,2) DEFAULT 0,
    ensemble_score DECIMAL(5,2) DEFAULT 0,
    
    -- Liveness Check
    liveness_score DECIMAL(5,2) DEFAULT 0,
    liveness_checks TEXT,
    
    -- Final Scores
    final_score DECIMAL(5,2) DEFAULT 0,
    verification_notes TEXT,
    
    -- Timestamps
    processed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    
    -- Indexes
    KEY idx_kyc_verifications_reference (reference),
    KEY idx_kyc_verifications_user_id (user_id),
    KEY idx_kyc_verifications_status (status),
    KEY idx_kyc_verifications_created_at (created_at)
);

-- Create KYC documents table
CREATE TABLE IF NOT EXISTS kyc_documents (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    kyc_verification_id BIGINT NOT NULL,
    document_type VARCHAR(50) NOT NULL,
    image_path VARCHAR(500) NOT NULL,
    processing_status VARCHAR(50) DEFAULT 'pending' NOT NULL,
    processing_result TEXT,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    
    -- Foreign Keys
    CONSTRAINT fk_kyc_documents_verification_id FOREIGN KEY (kyc_verification_id) REFERENCES kyc_verifications(id) ON DELETE CASCADE,
    
    -- Indexes
    KEY idx_kyc_documents_verification_id (kyc_verification_id),
    KEY idx_kyc_documents_type (document_type),
    KEY idx_kyc_documents_status (processing_status)
);
