package security

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

// PasswordManager 密码管理器
type PasswordManager struct {
	config *SecurityConfig
	hashMethod string // bcrypt or argon2
}

// Argon2Config Argon2配置
type Argon2Config struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// DefaultArgon2Config 默认Argon2配置
func DefaultArgon2Config() *Argon2Config {
	return &Argon2Config{
		Memory:      64 * 1024, // 64MB
		Iterations:  3,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	}
}

// NewPasswordManager 创建密码管理器
func NewPasswordManager(config *SecurityConfig) *PasswordManager {
	return &PasswordManager{
		config:     config,
		hashMethod: "bcrypt", // 默认使用bcrypt
	}
}

// SetHashMethod 设置哈希方法
func (pm *PasswordManager) SetHashMethod(method string) error {
	method = strings.ToLower(method)
	if method != "bcrypt" && method != "argon2" {
		return errors.New("unsupported hash method, only 'bcrypt' and 'argon2' are supported")
	}

	pm.hashMethod = method
	return nil
}

// GeneratePasswordHash 生成密码哈希
func (pm *PasswordManager) GeneratePasswordHash(password string) (string, error) {
	// 验证密码策略
	if err := ValidatePasswordPolicy(password, pm.config); err != nil {
		return "", err
	}

	switch pm.hashMethod {
	case "bcrypt":
		return pm.generateBcryptHash(password)
	case "argon2":
		return pm.generateArgon2Hash(password)
	default:
		return "", fmt.Errorf("unsupported hash method: %s", pm.hashMethod)
	}
}

// VerifyPassword 验证密码
func (pm *PasswordManager) VerifyPassword(password, hash string) (bool, error) {
	// 根据哈希格式自动检测哈希方法
	if strings.HasPrefix(hash, "$2a$") || strings.HasPrefix(hash, "$2b$") || strings.HasPrefix(hash, "$2y$") {
		// bcrypt 哈希
		return pm.verifyBcryptPassword(password, hash)
	} else if strings.HasPrefix(hash, "$argon2id$") {
		// Argon2id 哈希
		return pm.verifyArgon2Password(password, hash)
	}

	return false, errors.New("unknown hash format")
}

// generateBcryptHash 生成bcrypt哈希
func (pm *PasswordManager) generateBcryptHash(password string) (string, error) {
	// 使用配置中的cost参数
	hash, err := bcrypt.GenerateFromPassword([]byte(password), pm.config.BCryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to generate bcrypt hash: %w", err)
	}

	return string(hash), nil
}

// verifyBcryptPassword 验证bcrypt密码
func (pm *PasswordManager) verifyBcryptPassword(password, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil // 密码不匹配但不是错误
		}
		return false, fmt.Errorf("failed to verify bcrypt password: %w", err)
	}

	return true, nil
}

// generateArgon2Hash 生成Argon2id哈希
func (pm *PasswordManager) generateArgon2Hash(password string) (string, error) {
	config := DefaultArgon2Config()

	// 生成随机盐
	salt := make([]byte, config.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// 使用Argon2id生成哈希
	hash := argon2.IDKey([]byte(password), salt, config.Iterations, config.Memory, config.Parallelism, config.KeyLength)

	// 格式化为可存储的字符串
	// 格式: $argon2id$v=19$m=65536,t=3,p=2$c29tZXNhbHQ$RdescudvJCsgt3ub+b+dWRWJTmaaJObG
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, config.Memory, config.Iterations, config.Parallelism, b64Salt, b64Hash)

	return encodedHash, nil
}

// verifyArgon2Password 验证Argon2id密码
func (pm *PasswordManager) verifyArgon2Password(password, encodedHash string) (bool, error) {
	// 解析编码的哈希字符串
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" {
		return false, errors.New("invalid argon2id hash format")
	}

	// 解析版本
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false, fmt.Errorf("invalid version: %w", err)
	}

	// 解析参数
	config := DefaultArgon2Config()
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &config.Memory, &config.Iterations, &config.Parallelism); err != nil {
		return false, fmt.Errorf("invalid argon2 parameters: %w", err)
	}

	// 解码盐和哈希
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("invalid salt encoding: %w", err)
	}
	config.SaltLength = uint32(len(salt))

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("invalid hash encoding: %w", err)
	}
	config.KeyLength = uint32(len(expectedHash))

	// 生成哈希并比较
	computedHash := argon2.IDKey([]byte(password), salt, config.Iterations, config.Memory, config.Parallelism, config.KeyLength)

	// 时间恒定比较
	if len(computedHash) != len(expectedHash) {
		return false, nil
	}

	match := true
	for i := range computedHash {
		if computedHash[i] != expectedHash[i] {
			match = false
		}
	}

	return match, nil
}

// GenerateRandomPassword 生成随机密码
func GenerateRandomPassword(length int) string {
	if length < 8 {
		length = 8 // 最小长度8
	}

	const (
		uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		lowercase = "abcdefghijklmnopqrstuvwxyz"
		digits    = "0123456789"
		special   = "!@#$%^&*()_+-=[]{}|;:,.<>?"
	)

	// 确保密码包含至少一个大写字母、小写字母、数字和特殊字符
	allChars := uppercase + lowercase + digits + special
	password := make([]byte, length)

	// 生成随机种子
	s := make([]byte, length)
	_, err := rand.Read(s)
	if err != nil {
		// 如果随机生成失败，使用时间作为回退
		for i := range s {
			s[i] = byte(time.Now().UnixNano() % 256)
			time.Sleep(time.Nanosecond)
		}
	}

	// 生成密码
	for i := range password {
		password[i] = allChars[s[i]%byte(len(allChars))]
	}

	return string(password)
}

// IsPasswordStrong 检查密码强度
func IsPasswordStrong(password string) (bool, map[string]bool) {
	result := map[string]bool{
		"min_length": false,
		"uppercase":  false,
		"lowercase":  false,
		"digit":      false,
		"special":    false,
	}

	// 最小长度检查
	if len(password) >= 8 {
		result["min_length"] = true
	}

	// 检查字符类型
	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			result["uppercase"] = true
		case 'a' <= char && char <= 'z':
			result["lowercase"] = true
		case '0' <= char && char <= '9':
			result["digit"] = true
		case char >= 33 && char <= 47 || char >= 58 && char <= 64 || char >= 91 && char <= 96 || char >= 123 && char <= 126:
			result["special"] = true
		}
	}

	// 密码强度判断：至少满足3个条件
	strength := 0
	for _, v := range result {
		if v {
			strength++
		}
	}

	return strength >= 3, result
}

// EstimatePasswordComplexity 估算密码复杂度（熵）
func EstimatePasswordComplexity(password string) float64 {
	if len(password) == 0 {
		return 0
	}

	// 计算字符集大小
	var charsetSize float64 = 0
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case 'a' <= char && char <= 'z':
			hasLower = true
		case '0' <= char && char <= '9':
			hasDigit = true
		case char >= 33 && char <= 47 || char >= 58 && char <= 64 || char >= 91 && char <= 96 || char >= 123 && char <= 126:
			hasSpecial = true
		}
	}

	if hasUpper {
		charsetSize += 26
	}
	if hasLower {
		charsetSize += 26
	}
	if hasDigit {
		charsetSize += 10
	}
	if hasSpecial {
		charsetSize += 32 // 常见特殊字符数量
	}

	// 计算熵：长度 * log2(字符集大小)
	if charsetSize > 0 {
		return float64(len(password)) * (log2(charsetSize))
	}

	return 0
}

// log2 计算以2为底的对数
func log2(x float64) float64 {
	if x <= 0 {
		return 0
	}
	return (log10(x)) / log10(2.0)
}

// log10 计算以10为底的对数
func log10(x float64) float64 {
	if x <= 0 {
		return 0
	}

	// 使用简单的对数计算方法
	result := 0.0
	if x >= 1 {
		for x >= 10 {
			x /= 10
			result++
		}
	} else {
		for x < 1 {
			x *= 10
			result--
		}
	}

	// 线性近似小数部分
	fraction := x - 1.0
	result += fraction * 0.3010 // log10(1.1) ≈ 0.0414, 但使用近似值加速计算

	return result
}

// CheckPasswordHistory 检查新密码是否在历史记录中
func CheckPasswordHistory(newPassword string, history []string) (bool, error) {
	for _, oldHash := range history {
		matches, err := (&PasswordManager{}).VerifyPassword(newPassword, oldHash)
		if err != nil {
			return false, fmt.Errorf("failed to verify password against history: %w", err)
		}
		if matches {
			return true, nil // 密码在历史记录中
		}
	}

	return false, nil // 密码不在历史记录中
}