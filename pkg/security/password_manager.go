package security

import (
	"fmt"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

// PasswordManager 密码管理器接口
type PasswordManager interface {
	// HashPassword 加密密码
	HashPassword(password string) (string, error)
	// VerifyPassword 验证密码
	VerifyPassword(hashedPassword, password string) error
	// ValidatePasswordStrength 验证密码强度
	ValidatePasswordStrength(password string) error
}

// passwordManager 密码管理器实现
type passwordManager struct {
	minLength        int
	requireUppercase bool
	requireLowercase bool
	requireNumber    bool
	requireSpecial   bool
}

// PasswordConfig 密码配置
type PasswordConfig struct {
	MinLength        int
	RequireUppercase bool
	RequireLowercase bool
	RequireNumber    bool
	RequireSpecial   bool
}

// DefaultPasswordConfig 默认密码配置
var DefaultPasswordConfig = PasswordConfig{
	MinLength:        8,
	RequireUppercase: true,
	RequireLowercase: true,
	RequireNumber:    true,
	RequireSpecial:   false,
}

// NewPasswordManager 创建密码管理器
func NewPasswordManager(config PasswordConfig) PasswordManager {
	return &passwordManager{
		minLength:        config.MinLength,
		requireUppercase: config.RequireUppercase,
		requireLowercase: config.RequireLowercase,
		requireNumber:    config.RequireNumber,
		requireSpecial:   config.RequireSpecial,
	}
}

// HashPassword 加密密码
func (m *passwordManager) HashPassword(password string) (string, error) {
	// 验证密码强度
	if err := m.ValidatePasswordStrength(password); err != nil {
		return "", err
	}

	// 使用 bcrypt 加密（cost=10）
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("密码加密失败: %w", err)
	}

	return string(hashedBytes), nil
}

// VerifyPassword 验证密码
func (m *passwordManager) VerifyPassword(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return fmt.Errorf("密码错误")
		}
		return fmt.Errorf("密码验证失败: %w", err)
	}
	return nil
}

// ValidatePasswordStrength 验证密码强度
func (m *passwordManager) ValidatePasswordStrength(password string) error {
	// 检查长度
	if len(password) < m.minLength {
		return fmt.Errorf("密码长度至少为 %d 位", m.minLength)
	}

	// 检查大写字母
	if m.requireUppercase {
		matched, _ := regexp.MatchString(`[A-Z]`, password)
		if !matched {
			return fmt.Errorf("密码必须包含至少一个大写字母")
		}
	}

	// 检查小写字母
	if m.requireLowercase {
		matched, _ := regexp.MatchString(`[a-z]`, password)
		if !matched {
			return fmt.Errorf("密码必须包含至少一个小写字母")
		}
	}

	// 检查数字
	if m.requireNumber {
		matched, _ := regexp.MatchString(`[0-9]`, password)
		if !matched {
			return fmt.Errorf("密码必须包含至少一个数字")
		}
	}

	// 检查特殊字符
	if m.requireSpecial {
		matched, _ := regexp.MatchString(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`, password)
		if !matched {
			return fmt.Errorf("密码必须包含至少一个特殊字符")
		}
	}

	return nil
}
