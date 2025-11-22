package utils

import (
	"bytes"
	"crypto/rand"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// String utilities - enhanced string manipulation functions

// RandomString generates a random string of specified length
func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)

	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return ""
		}
		result[i] = charset[num.Int64()]
	}

	return string(result)
}

// StringToSnakeCase converts a string to snake_case
func StringToSnakeCase(s string) string {
	var result bytes.Buffer

	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result.WriteByte('_')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// StringToCamelCase converts a string to camelCase
func StringToCamelCase(s string) string {
	words := strings.Fields(strings.ReplaceAll(s, "_", " "))
	for i := range words {
		if i == 0 {
			words[i] = strings.ToLower(words[i])
		} else {
			words[i] = TitleCase(words[i])
		}
	}
	return strings.Join(words, "")
}

// PascalCase converts a string to PascalCase
func PascalCase(s string) string {
	words := strings.Fields(strings.ReplaceAll(s, "_", " "))
	for i := range words {
		words[i] = TitleCase(words[i])
	}
	return strings.Join(words, "")
}

// TitleCase converts a string to title case
func TitleCase(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}

// KebabCase converts a string to kebab-case
func KebabCase(s string) string {
	return strings.ReplaceAll(StringToSnakeCase(s), "_", "-")
}

// ReverseString reverses a string
func ReverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// CountWords counts the number of words in a string
func CountWords(s string) int {
	return len(strings.Fields(s))
}

// TruncateString truncates a string to specified length with ellipsis
func TruncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length] + "..."
}

// TruncateWords truncates a string to specified number of words
func TruncateWords(s string, wordCount int) string {
	words := strings.Fields(s)
	if len(words) <= wordCount {
		return s
	}
	return strings.Join(words[:wordCount], " ") + "..."
}

// RemoveExtraSpaces removes extra spaces from a string
func RemoveExtraSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// ExtractNumbers extracts all numbers from a string
func ExtractNumbers(s string) []int {
	re := regexp.MustCompile(`\d+`)
	matches := re.FindAllString(s, -1)

	var numbers []int
	for _, match := range matches {
		if num, err := strconv.Atoi(match); err == nil {
			numbers = append(numbers, num)
		}
	}

	return numbers
}

// ExtractEmails extracts all email addresses from a string
func ExtractEmails(s string) []string {
	re := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	return re.FindAllString(s, -1)
}

// ExtractURLs extracts all URLs from a string
func ExtractURLs(s string) []string {
	re := regexp.MustCompile(`https?://[^\s]+`)
	return re.FindAllString(s, -1)
}

// StringContainsAny checks if a string contains any of the given substrings
func StringContainsAny(s string, substrings []string) bool {
	for _, substr := range substrings {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}

// StringContainsAll checks if a string contains all of the given substrings
func StringContainsAll(s string, substrings []string) bool {
	for _, substr := range substrings {
		if !strings.Contains(s, substr) {
			return false
		}
	}
	return true
}

// ReplaceAll replaces all occurrences of multiple old strings with new strings
func ReplaceAll(s string, oldNew map[string]string) string {
	result := s
	for old, new := range oldNew {
		result = strings.ReplaceAll(result, old, new)
	}
	return result
}

// MaskEmail masks an email address for privacy
func MaskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}

	localPart := parts[0]
	domain := parts[1]

	if len(localPart) <= 2 {
		return "***@" + domain
	}

	maskedLocal := localPart[:2] + strings.Repeat("*", len(localPart)-2)
	return maskedLocal + "@" + domain
}

// MaskPhone masks a phone number for privacy
func MaskPhone(phone string) string {
	if len(phone) <= 4 {
		return strings.Repeat("*", len(phone))
	}

	return strings.Repeat("*", len(phone)-4) + phone[len(phone)-4:]
}

// PadLeft pads a string on the left to specified length
func PadLeft(s string, length int, padChar rune) string {
	if len(s) >= length {
		return s
	}

	padding := length - len(s)
	return strings.Repeat(string(padChar), padding) + s
}

// PadRight pads a string on the right to specified length
func PadRight(s string, length int, padChar rune) string {
	if len(s) >= length {
		return s
	}

	padding := length - len(s)
	return s + strings.Repeat(string(padChar), padding)
}

// PadCenter pads a string on both sides to specified length
func PadCenter(s string, length int, padChar rune) string {
	if len(s) >= length {
		return s
	}

	padding := length - len(s)
	leftPadding := padding / 2
	rightPadding := padding - leftPadding

	return strings.Repeat(string(padChar), leftPadding) + s + strings.Repeat(string(padChar), rightPadding)
}

// RemovePrefix removes a prefix from a string if it exists
func RemovePrefix(s, prefix string) string {
	if strings.HasPrefix(s, prefix) {
		return s[len(prefix):]
	}
	return s
}

// RemoveSuffix removes a suffix from a string if it exists
func RemoveSuffix(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		return s[:len(s)-len(suffix)]
	}
	return s
}

// RemoveSubstrings removes all occurrences of specified substrings
func RemoveSubstrings(s string, substrings []string) string {
	result := s
	for _, substr := range substrings {
		result = strings.ReplaceAll(result, substr, "")
	}
	return result
}

// SplitByLength splits a string into chunks of specified length
func SplitByLength(s string, chunkSize int) []string {
	if chunkSize <= 0 {
		return []string{s}
	}

	var chunks []string
	runes := []rune(s)

	for i := 0; i < len(runes); i += chunkSize {
		end := i + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[i:end]))
	}

	return chunks
}

// JoinStrings joins strings with a separator, ignoring empty strings
func JoinStrings(separator string, strs ...string) string {
	var nonEmpty []string
	for _, s := range strs {
		if s != "" {
			nonEmpty = append(nonEmpty, s)
		}
	}
	return strings.Join(nonEmpty, separator)
}

// CommonPrefix finds the common prefix of two strings
func CommonPrefix(s1, s2 string) string {
	minLen := min(len(s1), len(s2))

	for i := 0; i < minLen; i++ {
		if s1[i] != s2[i] {
			return s1[:i]
		}
	}

	return s1[:minLen]
}

// CommonSuffix finds the common suffix of two strings
func CommonSuffix(s1, s2 string) string {
	r1, r2 := []rune(s1), []rune(s2)
	len1, len2 := len(r1), len(r2)
	minLen := min(len1, len2)

	for i := 1; i <= minLen; i++ {
		if r1[len1-i] != r2[len2-i] {
			return string(r1[len1-i+1:])
		}
	}

	return string(r1[len1-minLen:])
}

// LevenshteinDistance calculates the Levenshtein distance between two strings
func LevenshteinDistance(s1, s2 string) int {
	r1, r2 := []rune(s1), []rune(s2)
	len1, len2 := len(r1), len(r2)

	if len1 == 0 {
		return len2
	}
	if len2 == 0 {
		return len1
	}

	distance := make([][]int, len1+1)
	for i := range distance {
		distance[i] = make([]int, len2+1)
		distance[i][0] = i
	}
	for j := range distance[0] {
		distance[0][j] = j
	}

	for i := 1; i <= len1; i++ {
		for j := 1; j <= len2; j++ {
			cost := 1
			if r1[i-1] == r2[j-1] {
				cost = 0
			}
			distance[i][j] = min3(
				distance[i-1][j]+1,
				distance[i][j-1]+1,
				distance[i-1][j-1]+cost,
			)
		}
	}

	return distance[len1][len2]
}

// JaroWinklerSimilarity calculates Jaro-Winkler similarity between two strings
func JaroWinklerSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}

	len1, len2 := len(s1), len(s2)
	if len1 == 0 || len2 == 0 {
		return 0.0
	}

	// Calculate Jaro distance
	matchDistance := max(len1, len2)/2 - 1

	matches1 := make([]bool, len1)
	matches2 := make([]bool, len2)

	matches := 0
	for i := 0; i < len1; i++ {
		start := max(0, i-matchDistance)
		end := min(len2, i+matchDistance+1)

		for j := start; j < end; j++ {
			if !matches2[j] && s1[i] == s2[j] {
				matches1[i] = true
				matches2[j] = true
				matches++
				break
			}
		}
	}

	if matches == 0 {
		return 0.0
	}

	// Calculate transpositions
	transpositions := 0
	k := 0
	for i := 0; i < len1; i++ {
		if matches1[i] {
			for !matches2[k] {
				k++
			}
			if s1[i] != s2[k] {
				transpositions++
			}
			k++
		}
	}

	jaro := (float64(matches)/float64(len1) +
		float64(matches)/float64(len2) +
		float64(matches-transpositions/2)/float64(matches)) / 3.0

	// Calculate Jaro-Winkler similarity
	prefix := 0
	for i := 0; i < min(min(len1, len2), 4); i++ {
		if s1[i] == s2[i] {
			prefix++
		} else {
			break
		}
	}

	return jaro + float64(prefix)*0.1*(1-jaro)
}

// min helper function for two arguments
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// min3 helper function for three arguments
func min3(a, b, c int) int {
	return min(min(a, b), c)
}

// max helper function
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
