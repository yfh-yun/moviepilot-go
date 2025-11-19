package utils

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// IsValidPhone checks if a string is a valid phone number
func IsValidPhone(phone string) bool {
	// Remove all non-digit characters
	digits := ""
	for _, r := range phone {
		if unicode.IsDigit(r) {
			digits += string(r)
		}
	}

	// Check phone number length
	if len(digits) < 10 || len(digits) > 15 {
		return false
	}

	// Basic phone number validation
	phoneRegex := regexp.MustCompile(`^\+?[\d\s\-\(\)]+$`)
	return phoneRegex.MatchString(phone)
}

// IsValidURL checks if a string is a valid URL
func IsValidURL(url string) bool {
	if len(url) < 3 || len(url) > 2083 {
		return false
	}

	urlRegex := regexp.MustCompile(`^(https?:\/\/)?([\da-z\.-]+)\.([a-z\.]{2,6})([\/\w \.-]*)*\/?$`)
	return urlRegex.MatchString(url)
}

// IsValidIP checks if a string is a valid IP address
func IsValidIP(ip string) bool {
	ipv4Regex := regexp.MustCompile(`^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)

	// Simple IPv6 check (basic pattern)
	ipv6Regex := regexp.MustCompile(`^([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}$`)

	return ipv4Regex.MatchString(ip) || ipv6Regex.MatchString(ip)
}

// IsValidDomain checks if a string is a valid domain name
func IsValidDomain(domain string) bool {
	if len(domain) < 1 || len(domain) > 253 {
		return false
	}

	domainRegex := regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)
	return domainRegex.MatchString(domain)
}

// IsValidUsername checks if a string is a valid username
func IsValidUsername(username string) bool {
	if len(username) < 3 || len(username) > 20 {
		return false
	}

	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	return usernameRegex.MatchString(username)
}

// IsValidPassword checks if a string is a valid password
func IsValidPassword(password string, minLength int) bool {
	if len(password) < minLength {
		return false
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasDigit && hasSpecial
}

// IsValidUUID checks if a string is a valid UUID
func IsValidUUID(uuid string) bool {
	uuidRegex := regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)
	return uuidRegex.MatchString(uuid)
}

// IsValidHash checks if a string is a valid hash (MD5, SHA1, SHA256, etc.)
func IsValidHash(hash string) bool {
	if len(hash) < 32 || len(hash) > 64 {
		return false
	}

	hashRegex := regexp.MustCompile(`^[a-fA-F0-9]+$`)
	return hashRegex.MatchString(hash)
}

// IsValidNumber checks if a string is a valid number
func IsValidNumber(number string) bool {
	numberRegex := regexp.MustCompile(`^-?\d+(\.\d+)?$`)
	return numberRegex.MatchString(number)
}

// IsValidInteger checks if a string is a valid integer
func IsValidInteger(number string) bool {
	integerRegex := regexp.MustCompile(`^-?\d+$`)
	return integerRegex.MatchString(number)
}

// IsValidDecimal checks if a string is a valid decimal number
func IsValidDecimal(number string) bool {
	decimalRegex := regexp.MustCompile(`^-?\d+\.\d+$`)
	return decimalRegex.MatchString(number)
}

// IsValidPercentage checks if a string is a valid percentage (0-100)
func IsValidPercentage(percentage string) bool {
	if !IsValidNumber(percentage) {
		return false
	}

	val := strings.TrimSpace(percentage)
	if strings.HasSuffix(val, "%") {
		val = strings.TrimSuffix(val, "%")
	}

	num := 0.0
	_, err := fmt.Sscanf(val, "%f", &num)
	if err != nil {
		return false
	}

	return num >= 0 && num <= 100
}

// IsValidHexColor checks if a string is a valid hex color code
func IsValidHexColor(color string) bool {
	colorRegex := regexp.MustCompile(`^#([a-fA-F0-9]{6}|[a-fA-F0-9]{3})$`)
	return colorRegex.MatchString(color)
}

// IsValidJSON checks if a string is valid JSON
func IsValidJSON(jsonStr string) bool {
	var js interface{}
	return json.Unmarshal([]byte(jsonStr), &js) == nil
}

// IsValidBase64 checks if a string is valid Base64
func IsValidBase64(base64Str string) bool {
	base64Regex := regexp.MustCompile(`^[A-Za-z0-9+/]*={0,2}$`)
	if !base64Regex.MatchString(base64Str) {
		return false
	}

	// Check if the length is a multiple of 4
	return len(base64Str)%4 == 0
}

// IsValidMimeType checks if a string is a valid MIME type
func IsValidMimeType(mimeType string) bool {
	mimeRegex := regexp.MustCompile(`^[a-zA-Z0-9]+/[a-zA-Z0-9\-+]+$`)
	return mimeRegex.MatchString(mimeType)
}

// IsValidFileExtension checks if a string is a valid file extension
func IsValidFileExtension(extension string) bool {
	if len(extension) < 1 || len(extension) > 10 {
		return false
	}

	extensionRegex := regexp.MustCompile(`^\.[a-zA-Z0-9]+$`)
	return extensionRegex.MatchString(extension)
}

// IsValidLanguageCode checks if a string is a valid language code (ISO 639-1)
func IsValidLanguageCode(code string) bool {
	languageRegex := regexp.MustCompile(`^[a-z]{2}(-[A-Z]{2})?$`)
	return languageRegex.MatchString(code)
}

// IsValidCountryCode checks if a string is a valid country code (ISO 3166-1 alpha-2)
func IsValidCountryCode(code string) bool {
	countryRegex := regexp.MustCompile(`^[A-Z]{2}$`)
	return countryRegex.MatchString(code)
}

// IsValidTimeZone checks if a string is a valid timezone identifier
func IsValidTimeZone(timezone string) bool {
	timezoneRegex := regexp.MustCompile(`^[A-Za-z_]+/[A-Za-z_]+$`)
	return timezoneRegex.MatchString(timezone)
}

// IsValidCreditCard checks if a string is a valid credit card number
func IsValidCreditCard(cardNumber string) bool {
	// Remove all non-digit characters
	digits := ""
	for _, r := range cardNumber {
		if unicode.IsDigit(r) {
			digits += string(r)
		}
	}

	// Check credit card number length
	if len(digits) < 13 || len(digits) > 19 {
		return false
	}

	// Luhn algorithm check
	return isValidLuhn(digits)
}

// isValidLuhn implements the Luhn algorithm for credit card validation
func isValidLuhn(number string) bool {
	sum := 0
	double := false

	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')

		if double {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		double = !double
	}

	return sum%10 == 0
}

// IsValidISBN checks if a string is a valid ISBN (10 or 13)
func IsValidISBN(isbn string) bool {
	// Remove hyphens and spaces
	cleanISBN := strings.ReplaceAll(strings.ReplaceAll(isbn, "-", ""), " ", "")

	if len(cleanISBN) == 10 {
		return isValidISBN10(cleanISBN)
	} else if len(cleanISBN) == 13 {
		return isValidISBN13(cleanISBN)
	}

	return false
}

// isValidISBN10 validates ISBN-10
func isValidISBN10(isbn string) bool {
	sum := 0
	for i := 0; i < 9; i++ {
		digit := int(isbn[i] - '0')
		if digit < 0 || digit > 9 {
			return false
		}
		sum += digit * (10 - i)
	}

	lastChar := isbn[9]
	if lastChar == 'X' || lastChar == 'x' {
		sum += 10
	} else {
		digit := int(lastChar - '0')
		if digit < 0 || digit > 9 {
			return false
		}
		sum += digit
	}

	return sum%11 == 0
}

// isValidISBN13 validates ISBN-13
func isValidISBN13(isbn string) bool {
	sum := 0
	for i := 0; i < 12; i++ {
		digit := int(isbn[i] - '0')
		if digit < 0 || digit > 9 {
			return false
		}
		if i%2 == 0 {
			sum += digit
		} else {
			sum += digit * 3
		}
	}

	lastDigit := int(isbn[12] - '0')
	if lastDigit < 0 || lastDigit > 9 {
		return false
	}

	checkDigit := (10 - (sum % 10)) % 10
	return lastDigit == checkDigit
}
