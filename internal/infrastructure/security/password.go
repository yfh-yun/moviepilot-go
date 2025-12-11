package security

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrPasswordTooShort 密码太短
	ErrPasswordTooShort = errors.New("password too short")
)

// HashPassword 生成密码哈希
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// VerifyPassword 验证密码
func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// HashPasswordWithCost 使用指定成本生成密码哈希
func HashPasswordWithCost(password string, cost int) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// GenerateRandomPassword 生成随机密码
func GenerateRandomPassword(length int) (string, error) {
	if length < 8 {
		length = 8
	}
	return GenerateRandomString(length)
}
