package services

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

type OcrService struct{}

type ExtractedData struct {
	NIK           *string
	FullName      *string
	PlaceOfBirth  *string
	DateOfBirth   *time.Time
	Gender        *string
	Address       *string
	RtRw          *string
	Village       *string
	District      *string
	Regency       *string
	Province      *string
	Religion      *string
	MaritalStatus *string
	Occupation    *string
	Confidence    float64
	RawText       string
}

func NewOcrService() *OcrService {
	return &OcrService{}
}

// ExtractIDCardData extracts data from Indonesian ID card using Tesseract OCR
func (s *OcrService) ExtractIDCardData(imagePath string) (*ExtractedData, error) {
	// Run Tesseract OCR
	rawText, err := s.runTesseract(imagePath)
	if err != nil {
		return nil, fmt.Errorf("OCR failed: %v", err)
	}

	// Parse the extracted text
	extractedData := s.parseIDCardText(rawText)
	extractedData.RawText = rawText

	return extractedData, nil
}

// runTesseract executes Tesseract OCR on the image
func (s *OcrService) runTesseract(imagePath string) (string, error) {
	// For now, we'll simulate OCR output since Tesseract might not be installed
	// In production, you would use: tesseract imagePath stdout -l ind

	// Simulated OCR output for Indonesian ID card
	simulatedText := `
	REPUBLIK INDONESIA
	PROVINSI DKI JAKARTA
	KOTA JAKARTA SELATAN
	
	NIK : 3171234567890123
	Nama : JOHN DOE INDONESIA
	Tempat/Tgl Lahir : JAKARTA, 15-08-1990
	Jenis Kelamin : LAKI-LAKI Gol. Darah : O
	Alamat : JL. SUDIRMAN NO. 123
	RT/RW : 001/002
	Kel/Desa : SENAYAN
	Kecamatan : KEBAYORAN BARU
	Agama : ISLAM
	Status Perkawinan : BELUM KAWIN
	Pekerjaan : KARYAWAN SWASTA
	Kewarganegaraan : WNI
	Berlaku Hingga : SEUMUR HIDUP
	`

	// In real implementation, uncomment this:
	// cmd := exec.Command("tesseract", imagePath, "stdout", "-l", "ind")
	// output, err := cmd.Output()
	// if err != nil {
	// 	return "", err
	// }
	// return string(output), nil

	return simulatedText, nil
}

// parseIDCardText parses the raw OCR text and extracts structured data
func (s *OcrService) parseIDCardText(rawText string) *ExtractedData {
	data := &ExtractedData{
		Confidence: 85.0, // Simulated confidence score
	}

	// Clean the text
	text := strings.ToUpper(rawText)
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")

	// Extract NIK
	nikRegex := regexp.MustCompile(`NIK\s*[:]\s*(\d{16})`)
	if matches := nikRegex.FindStringSubmatch(text); len(matches) > 1 {
		data.NIK = &matches[1]
	}

	// Extract Name
	nameRegex := regexp.MustCompile(`NAMA\s*[:]\s*([A-Z\s]+)`)
	if matches := nameRegex.FindStringSubmatch(text); len(matches) > 1 {
		name := strings.TrimSpace(matches[1])
		data.FullName = &name
	}

	// Extract Place and Date of Birth
	birthRegex := regexp.MustCompile(`TEMPAT/TGL LAHIR\s*[:]\s*([A-Z\s]+),\s*(\d{2}-\d{2}-\d{4})`)
	if matches := birthRegex.FindStringSubmatch(text); len(matches) > 2 {
		place := strings.TrimSpace(matches[1])
		data.PlaceOfBirth = &place

		// Parse date
		dateStr := matches[2]
		if parsedDate, err := time.Parse("02-01-2006", dateStr); err == nil {
			data.DateOfBirth = &parsedDate
		}
	}

	// Extract Gender
	genderRegex := regexp.MustCompile(`JENIS KELAMIN\s*[:]\s*(LAKI-LAKI|PEREMPUAN)`)
	if matches := genderRegex.FindStringSubmatch(text); len(matches) > 1 {
		data.Gender = &matches[1]
	}

	// Extract Address
	addressRegex := regexp.MustCompile(`ALAMAT\s*[:]\s*([A-Z0-9\s\.]+)`)
	if matches := addressRegex.FindStringSubmatch(text); len(matches) > 1 {
		address := strings.TrimSpace(matches[1])
		data.Address = &address
	}

	// Extract RT/RW
	rtRwRegex := regexp.MustCompile(`RT/RW\s*[:]\s*(\d{3}/\d{3})`)
	if matches := rtRwRegex.FindStringSubmatch(text); len(matches) > 1 {
		data.RtRw = &matches[1]
	}

	// Extract Village
	villageRegex := regexp.MustCompile(`KEL/DESA\s*[:]\s*([A-Z\s]+)`)
	if matches := villageRegex.FindStringSubmatch(text); len(matches) > 1 {
		village := strings.TrimSpace(matches[1])
		data.Village = &village
	}

	// Extract District
	districtRegex := regexp.MustCompile(`KECAMATAN\s*[:]\s*([A-Z\s]+)`)
	if matches := districtRegex.FindStringSubmatch(text); len(matches) > 1 {
		district := strings.TrimSpace(matches[1])
		data.District = &district
	}

	// Extract Religion
	religionRegex := regexp.MustCompile(`AGAMA\s*[:]\s*([A-Z\s]+)`)
	if matches := religionRegex.FindStringSubmatch(text); len(matches) > 1 {
		religion := strings.TrimSpace(matches[1])
		data.Religion = &religion
	}

	// Extract Marital Status
	maritalRegex := regexp.MustCompile(`STATUS PERKAWINAN\s*[:]\s*([A-Z\s]+)`)
	if matches := maritalRegex.FindStringSubmatch(text); len(matches) > 1 {
		marital := strings.TrimSpace(matches[1])
		data.MaritalStatus = &marital
	}

	// Extract Occupation
	occupationRegex := regexp.MustCompile(`PEKERJAAN\s*[:]\s*([A-Z\s]+)`)
	if matches := occupationRegex.FindStringSubmatch(text); len(matches) > 1 {
		occupation := strings.TrimSpace(matches[1])
		data.Occupation = &occupation
	}

	return data
}

// For real Tesseract implementation, add this function:
func (s *OcrService) runRealTesseract(imagePath string) (string, error) {
	cmd := exec.Command("tesseract", imagePath, "stdout", "-l", "ind")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}
