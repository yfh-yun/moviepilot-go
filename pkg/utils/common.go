package utils

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// String utilities

// IsEmpty checks if a string is empty after trimming spaces
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// IsNotEmpty checks if a string is not empty after trimming spaces
func IsNotEmpty(s string) bool {
	return !IsEmpty(s)
}

// ContainsString checks if a slice contains a specific string
func ContainsString(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

// RemoveDuplicates removes duplicate strings from a slice
func RemoveDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// StringToInt converts a string to int with default value on error
func StringToInt(s string, defaultValue int) int {
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	return defaultValue
}

// StringToInt64 converts a string to int64 with default value on error
func StringToInt64(s string, defaultValue int64) int64 {
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i
	}
	return defaultValue
}

// StringToFloat64 converts a string to float64 with default value on error
func StringToFloat64(s string, defaultValue float64) float64 {
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return defaultValue
}

// StringToBool converts a string to bool with default value on error
func StringToBool(s string, defaultValue bool) bool {
	if b, err := strconv.ParseBool(s); err == nil {
		return b
	}
	return defaultValue
}

// File utilities

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// IsDirectory checks if a path is a directory
func IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// GetFileSize gets the size of a file
func GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// GetFileExtension gets the extension of a file
func GetFileExtension(path string) string {
	return strings.ToLower(filepath.Ext(path))
}

// GetFileNameWithoutExtension gets the filename without extension
func GetFileNameWithoutExtension(path string) string {
	filename := filepath.Base(path)
	return strings.TrimSuffix(filename, filepath.Ext(filename))
}

// CreateDirectory creates a directory recursively
func CreateDirectory(path string) error {
	return os.MkdirAll(path, 0755)
}

// Time utilities

// CurrentTimestamp returns current timestamp in seconds
func CurrentTimestamp() int64 {
	return time.Now().Unix()
}

// CurrentTimestampMillis returns current timestamp in milliseconds
func CurrentTimestampMillis() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// FormatDate formats a time to string
func FormatDate(t time.Time, format string) string {
	return t.Format(format)
}

// ParseDate parses a string to time
func ParseDate(dateStr, format string) (time.Time, error) {
	return time.Parse(format, dateStr)
}

// IsValidDate checks if a date string is valid according to the format
func IsValidDate(dateStr, format string) bool {
	_, err := time.Parse(format, dateStr)
	return err == nil
}

// Math utilities

// Min returns the minimum of two integers
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Max returns the maximum of two integers
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// MinInt64 returns the minimum of two int64
func MinInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// MaxInt64 returns the maximum of two int64
func MaxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// MinFloat64 returns the minimum of two float64
func MinFloat64(a, b float64) float64 {
	return math.Min(a, b)
}

// MaxFloat64 returns the maximum of two float64
func MaxFloat64(a, b float64) float64 {
	return math.Max(a, b)
}

// Clamp clamps a value between min and max
func Clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// ClampFloat64 clamps a float64 value between min and max
func ClampFloat64(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// Network utilities

// IsValidIP checks if a string is a valid IP address
func IsValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// IsValidPort checks if a port number is valid
func IsValidPort(port int) bool {
	return port > 0 && port <= 65535
}

// IsValidURL checks if a string is a valid URL
func IsValidURL(url string) bool {
	_, err := url.Parse(url)
	return err == nil
}

// GetLocalIP returns the local IP address
func GetLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", fmt.Errorf("cannot get local IP address")
}

// Crypto utilities

// GenerateRandomString generates a random string of specified length
func GenerateRandomString(length int) (string, error) {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)

	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		result[i] = chars[num.Int64()]
	}

	return string(result), nil
}

// MD5Hash generates MD5 hash of a string
func MD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

// SHA256Hash generates SHA256 hash of a string
func SHA256Hash(text string) string {
	hash := sha256.Sum256([]byte(text))
	return hex.EncodeToString(hash[:])
}

// HashFile generates MD5 hash of a file
func HashFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// Validation utilities

// IsValidEmail checks if a string is a valid email address
func IsValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// IsValidPhone checks if a string is a valid phone number
func IsValidPhone(phone string) bool {
	phoneRegex := regexp.MustCompile(`^[+]?[\d\s-()]{10,}$`)
	return phoneRegex.MatchString(phone)
}

// IsValidUsername checks if a string is a valid username
func IsValidUsername(username string) bool {
	if len(username) < 3 || len(username) > 20 {
		return false
	}
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	return usernameRegex.MatchString(username)
}

// IsStrongPassword checks if a password is strong enough
func IsStrongPassword(password string) bool {
	if len(password) < 8 {
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

// String manipulation utilities

// Truncate truncates a string to the specified length
func Truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length] + "..."
}

// Capitalize capitalizes the first letter of a string
func Capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// SnakeCase converts a string to snake_case
func SnakeCase(s string) string {
	var result []rune

	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}

	return string(result)
}

// CamelCase converts a string to camelCase
func CamelCase(s string) string {
	words := strings.Fields(strings.ReplaceAll(s, "_", " "))
	for i := range words {
		if i == 0 {
			words[i] = strings.ToLower(words[i])
		} else {
			words[i] = Capitalize(strings.ToLower(words[i]))
		}
	}
	return strings.Join(words, "")
}

// Pagination utilities

// CalculateOffset calculates the offset for pagination
func CalculateOffset(page, pageSize int) int {
	if page < 1 {
		page = 1
	}
	return (page - 1) * pageSize
}

// CalculateTotalPages calculates the total number of pages
func CalculateTotalPages(totalItems, pageSize int) int {
	if totalItems <= 0 || pageSize <= 0 {
		return 0
	}
	return int(math.Ceil(float64(totalItems) / float64(pageSize)))
}

// Error handling utilities

// WrapError wraps an error with additional context
func WrapError(err error, context string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", context, err)
}

// ErrorContains checks if an error contains a specific substring
func ErrorContains(err error, substr string) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), substr)
}

// Data conversion utilities

// InterfaceToString converts an interface to string
func InterfaceToString(i interface{}) string {
	if i == nil {
		return ""
	}

	switch v := i.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// StringToInterface converts a string to appropriate type
func StringToInterface(s string) interface{} {
	// Try to convert to int
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}

	// Try to convert to float64
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}

	// Try to convert to bool
	if b, err := strconv.ParseBool(s); err == nil {
		return b
	}

	// Return as string
	return s
}

// Miscellaneous utilities

// Pointer returns a pointer to the given value
func Pointer[T any](v T) *T {
	return &v
}

// Dereference returns the dereferenced value or default if nil
func Dereference[T any](ptr *T, defaultValue T) T {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

// SafeDereference safely dereferences a pointer
func SafeDereference[T any](ptr *T) (T, bool) {
	if ptr != nil {
		return *ptr, true
	}
	var zero T
	return zero, false
}
