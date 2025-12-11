package security

import (
	"crypto/rand"
	"math/big"
)

// GenerateRandomString 生成随机字符串
func GenerateRandomString(length int) (string, error) {
	if length <= 0 {
		length = 16
	}

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)

	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[num.Int64()]
	}

	return string(result), nil
}

// IsValidTokenFormat 检查令牌格式是否有效
func IsValidTokenFormat(token string) bool {
	if len(token) < 8 || len(token) > 256 {
		return false
	}

	// 只允许字母、数字、下划线和连字符
	for _, c := range token {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-') {
			return false
		}
	}

	return true
}

// IsValidPasswordFormat 检查密码格式是否有效
func IsValidPasswordFormat(password string) bool {
	if len(password) < 8 {
		return false
	}

	// 至少包含一个字母和一个数字
	hasLetter := false
	hasDigit := false

	for _, c := range password {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			hasLetter = true
		} else if c >= '0' && c <= '9' {
			hasDigit = true
		}
	}

	return hasLetter && hasDigit
}
