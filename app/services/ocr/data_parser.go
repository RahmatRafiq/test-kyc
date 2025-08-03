package ocr

import (
	"regexp"
	"strconv"
	"strings"

	"golang_starter_kit_2025/app/responses"
)

// DataParser handles parsing of extracted text into structured data
type DataParser struct{}

// ParseExtractedData parses extracted texts into structured data
func (*DataParser) ParseExtractedData(texts []string, idCardType string) *responses.ExtractedData {
	data := &responses.ExtractedData{}

	switch strings.ToLower(idCardType) {
	case "ktp", "e-ktp":
		parseKTPData(data, texts)
	case "sim":
		parseSIMData(data, texts)
	default:
		parseGenericData(data, texts)
	}

	return data
}

// parseKTPData parses KTP specific data
func parseKTPData(data *responses.ExtractedData, texts []string) {
	for _, text := range texts {
		cleanText := strings.TrimSpace(text)
		if cleanText == "" {
			continue
		}

		// Try to extract NIK (16 digit number)
		if data.IdCardNumber == "" {
			if nik := extractNIK(cleanText); nik != "" {
				data.IdCardNumber = nik
				continue
			}
		}

		// Try to extract name (after "NAMA" or similar keywords)
		if data.FullName == "" {
			if name := extractName(cleanText); name != "" {
				data.FullName = name
				continue
			}
		}

		// Try to extract place and date of birth
		if data.PlaceOfBirth == "" || data.DateOfBirth == "" {
			place, date := extractPlaceAndDateOfBirth(cleanText)
			if place != "" && data.PlaceOfBirth == "" {
				data.PlaceOfBirth = place
			}
			if date != "" && data.DateOfBirth == "" {
				data.DateOfBirth = date
			}
		}

		// Try to extract address
		if data.Address == "" {
			if address := extractAddress(cleanText); address != "" {
				data.Address = address
				continue
			}
		}

		// Try to extract gender
		if data.Gender == "" {
			if gender := extractGender(cleanText); gender != "" {
				data.Gender = gender
				continue
			}
		}

		// Try to extract religion
		if data.Religion == "" {
			if religion := extractReligion(cleanText); religion != "" {
				data.Religion = religion
				continue
			}
		}

		// Try to extract marital status
		if data.MaritalStatus == "" {
			if marital := extractMaritalStatus(cleanText); marital != "" {
				data.MaritalStatus = marital
				continue
			}
		}

		// Try to extract occupation
		if data.Occupation == "" {
			if occupation := extractOccupation(cleanText); occupation != "" {
				data.Occupation = occupation
				continue
			}
		}
	}
}

// extractNIK extracts NIK (16 digit ID number) from text
func extractNIK(text string) string {
	// Clean the text first
	cleanText := strings.ReplaceAll(text, " ", "")
	cleanText = strings.ReplaceAll(cleanText, "-", "")
	cleanText = strings.ReplaceAll(cleanText, ".", "")
	
	// Look for exactly 16 consecutive digits
	re := regexp.MustCompile(`\b\d{16}\b`)
	matches := re.FindAllString(cleanText, -1)
	
	for _, match := range matches {
		// Validate NIK format (basic validation)
		if isValidNIKFormat(match) {
			return match
		}
	}
	
	// Try to find NIK pattern in original text with separators
	re = regexp.MustCompile(`\b\d{2}[\s\-\.]?\d{2}[\s\-\.]?\d{2}[\s\-\.]?\d{6}[\s\-\.]?\d{4}\b`)
	matches = re.FindAllString(text, -1)
	
	for _, match := range matches {
		// Remove spaces and separators
		nik := regexp.MustCompile(`[\s\-\.]`).ReplaceAllString(match, "")
		if len(nik) == 16 && isValidNIKFormat(nik) {
			return nik
		}
	}
	
	// Try to extract from text that contains "NIK" keyword
	if strings.Contains(strings.ToUpper(text), "NIK") {
		parts := strings.Split(strings.ToUpper(text), "NIK")
		if len(parts) > 1 {
			afterNIK := strings.TrimSpace(parts[1])
			afterNIK = strings.TrimLeft(afterNIK, ":- ")
			
			// Extract first 16 digits found
			digitOnly := regexp.MustCompile(`\d+`).FindAllString(afterNIK, -1)
			for _, digits := range digitOnly {
				if len(digits) == 16 && isValidNIKFormat(digits) {
					return digits
				}
			}
		}
	}
	
	return ""
}

// isValidNIKFormat performs basic NIK format validation
func isValidNIKFormat(nik string) bool {
	if len(nik) != 16 {
		return false
	}
	
	// Check if all characters are digits
	for _, char := range nik {
		if char < '0' || char > '9' {
			return false
		}
	}
	
	// Basic format validation
	province, _ := strconv.Atoi(nik[:2])
	regency, _ := strconv.Atoi(nik[2:4])
	district, _ := strconv.Atoi(nik[4:6])
	
	// Basic range checks
	if province < 11 || province > 91 {
		return false
	}
	if regency < 1 || regency > 99 {
		return false
	}
	if district < 1 || district > 99 {
		return false
	}
	
	return true
}

// extractName extracts name from text
func extractName(text string) string {
	text = strings.ToUpper(text)
	text = strings.TrimSpace(text)
	
	// Look for name after keywords
	keywords := []string{"NAMA", "NAME", "NAMA LENGKAP", "NM"}
	
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			// Find text after the keyword
			parts := strings.Split(text, keyword)
			if len(parts) > 1 {
				name := strings.TrimSpace(parts[1])
				// Remove colons and common separators
				name = strings.TrimLeft(name, ":- ")
				// Take only the first line if multiple lines
				lines := strings.Split(name, "\n")
				if len(lines) > 0 {
					name = strings.TrimSpace(lines[0])
					// Remove any remaining digits at the end (sometimes dates get mixed)
					name = regexp.MustCompile(`\d.*$`).ReplaceAllString(name, "")
					name = strings.TrimSpace(name)
					// Validate name (should contain only letters and spaces)
					if isValidName(name) && len(name) > 2 {
						return name
					}
				}
			}
		}
	}
	
	// If no keyword found, check if the entire text looks like a name
	// Skip if it contains digits (likely not a name)
	if !strings.ContainsAny(text, "0123456789") {
		if isValidName(text) && len(text) > 2 && len(text) < 50 {
			words := strings.Fields(text)
			if len(words) >= 2 && len(words) <= 5 { // Reasonable name length
				// Additional check: skip if it contains common non-name keywords
				skipKeywords := []string{"PROVINSI", "KABUPATEN", "KOTA", "TEMPAT", "LAHIR", "JENIS", "KELAMIN", "ALAMAT", "AGAMA", "STATUS", "PEKERJAAN", "RT", "RW", "KEL", "KEC"}
				hasSkipKeyword := false
				for _, skipKeyword := range skipKeywords {
					if strings.Contains(text, skipKeyword) {
						hasSkipKeyword = true
						break
					}
				}
				if !hasSkipKeyword {
					return text
				}
			}
		}
	}
	
	return ""
}

// isValidName checks if text looks like a valid name
func isValidName(text string) bool {
	// Name should only contain letters, spaces, and common name characters
	re := regexp.MustCompile(`^[A-Z\s\.\']+$`)
	return re.MatchString(text)
}

// extractPlaceAndDateOfBirth extracts place and date of birth
func extractPlaceAndDateOfBirth(text string) (string, string) {
	text = strings.ToUpper(text)
	
	// Look for birth info patterns
	keywords := []string{"TEMPAT/TGL LAHIR", "TEMPAT TGL LAHIR", "TTL", "LAHIR"}
	
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			parts := strings.Split(text, keyword)
			if len(parts) > 1 {
				birthInfo := strings.TrimSpace(parts[1])
				birthInfo = strings.TrimLeft(birthInfo, ":- ")
				
				// Try to parse place and date
				return parsePlaceAndDate(birthInfo)
			}
		}
	}
	
	return "", ""
}

// parsePlaceAndDate parses place and date from birth info text
func parsePlaceAndDate(text string) (string, string) {
	// Look for date patterns (DD-MM-YYYY, DD/MM/YYYY, etc.)
	dateRe := regexp.MustCompile(`(\d{1,2}[\-/]\d{1,2}[\-/]\d{4})`)
	dateMatch := dateRe.FindString(text)
	
	if dateMatch != "" {
		// Extract place (text before the date)
		parts := strings.Split(text, dateMatch)
		if len(parts) > 0 {
			place := strings.TrimSpace(parts[0])
			place = strings.TrimRight(place, ",- ")
			if len(place) > 0 {
				return place, dateMatch
			}
		}
	}
	
	return "", ""
}

// extractAddress extracts address information
func extractAddress(text string) string {
	text = strings.ToUpper(text)
	
	keywords := []string{"ALAMAT", "ADDRESS", "JALAN", "JL", "RT", "RW"}
	
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			// This looks like address text
			if len(text) > 10 && len(text) < 200 {
				return strings.TrimSpace(text)
			}
		}
	}
	
	return ""
}

// extractGender extracts gender information
func extractGender(text string) string {
	text = strings.ToUpper(text)
	
	if strings.Contains(text, "LAKI") || strings.Contains(text, "MALE") {
		return "LAKI-LAKI"
	}
	if strings.Contains(text, "PEREMPUAN") || strings.Contains(text, "FEMALE") || strings.Contains(text, "WANITA") {
		return "PEREMPUAN"
	}
	
	return ""
}

// extractReligion extracts religion information
func extractReligion(text string) string {
	text = strings.ToUpper(text)
	
	religions := []string{"ISLAM", "KRISTEN", "KATOLIK", "HINDU", "BUDDHA", "KHONGHUCU"}
	
	for _, religion := range religions {
		if strings.Contains(text, religion) {
			return religion
		}
	}
	
	return ""
}

// extractMaritalStatus extracts marital status
func extractMaritalStatus(text string) string {
	text = strings.ToUpper(text)
	
	if strings.Contains(text, "KAWIN") || strings.Contains(text, "MARRIED") {
		return "KAWIN"
	}
	if strings.Contains(text, "BELUM") || strings.Contains(text, "SINGLE") {
		return "BELUM KAWIN"
	}
	if strings.Contains(text, "CERAI") || strings.Contains(text, "DIVORCED") {
		return "CERAI"
	}
	
	return ""
}

// extractOccupation extracts occupation information
func extractOccupation(text string) string {
	text = strings.ToUpper(text)
	
	keywords := []string{"PEKERJAAN", "OCCUPATION", "KERJA"}
	
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			parts := strings.Split(text, keyword)
			if len(parts) > 1 {
				occupation := strings.TrimSpace(parts[1])
				occupation = strings.TrimLeft(occupation, ":- ")
				if len(occupation) > 0 && len(occupation) < 50 {
					return occupation
				}
			}
		}
	}
	
	return ""
}

// parseSIMData parses SIM specific data
func parseSIMData(data *responses.ExtractedData, texts []string) {
	for _, text := range texts {
		cleanText := strings.TrimSpace(text)
		if cleanText == "" {
			continue
		}

		// SIM number extraction (usually starts with specific pattern)
		if data.IdCardNumber == "" {
			if simNumber := extractSIMNumber(cleanText); simNumber != "" {
				data.IdCardNumber = simNumber
			}
		}

		// Name extraction
		if data.FullName == "" {
			if name := extractName(cleanText); name != "" {
				data.FullName = name
			}
		}
	}
}

// extractSIMNumber extracts SIM number from text
func extractSIMNumber(text string) string {
	// SIM numbers typically have specific patterns
	re := regexp.MustCompile(`\b\d{12,15}\b`)
	matches := re.FindAllString(text, -1)
	
	for _, match := range matches {
		return match // Return first valid looking number
	}
	
	return ""
}

// parseGenericData parses generic ID data
func parseGenericData(data *responses.ExtractedData, texts []string) {
	for _, text := range texts {
		cleanText := strings.TrimSpace(text)
		if cleanText == "" {
			continue
		}

		// Try to find any number that could be an ID
		if data.IdCardNumber == "" {
			re := regexp.MustCompile(`\b\d{10,16}\b`)
			matches := re.FindAllString(cleanText, -1)
			if len(matches) > 0 {
				data.IdCardNumber = matches[0]
			}
		}

		// Try to find name-like text
		if data.FullName == "" {
			if name := extractName(cleanText); name != "" {
				data.FullName = name
			}
		}
	}
}
