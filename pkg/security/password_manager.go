package security

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"time"

	"golang.org/x/crypto/bcrypt"

	"moviepilot-go/pkg/cache"
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
	cacheBackend     cache.CacheBackend // 缓存后端
	cacheTTL         time.Duration      // 缓存过期时间
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
		cacheBackend:     nil, // 默认不使用缓存
		cacheTTL:         0,
	}
}

// NewPasswordManagerWithCache 创建带缓存的密码管理器
func NewPasswordManagerWithCache(config PasswordConfig, cacheBackend cache.CacheBackend, cacheTTL time.Duration) PasswordManager {
	return &passwordManager{
		minLength:        config.MinLength,
		requireUppercase: config.RequireUppercase,
		requireLowercase: config.RequireLowercase,
		requireNumber:    config.RequireNumber,
		requireSpecial:   config.RequireSpecial,
		cacheBackend:     cacheBackend,
		cacheTTL:         cacheTTL,
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

// ValidatePasswordStrength 验证密码强度（带缓存）
func (m *passwordManager) ValidatePasswordStrength(password string) error {
	// 缓存键：使用SHA256哈希避免存储明文密码
	passwordHash := fmt.Sprintf("%x", sha256.Sum256([]byte(password)))
	cacheKey := fmt.Sprintf("password_strength:%s", passwordHash)
	cacheRegion := "security"

	// 如果缓存已初始化，尝试从缓存获取
	if m.cacheBackend != nil {
		cachedResult, hit, err := m.cacheBackend.Get(cacheKey, cacheRegion)
		if err == nil && hit {
			if errResult, ok := cachedResult.(error); ok {
				return errResult
			}
			// 缓存结果为nil，说明密码强度足够
			return nil
		}
	}

	// 缓存未命中，执行实际验证
	var errResult error

	// 检查长度
	if len(password) < m.minLength {
		errResult = fmt.Errorf("密码长度至少为 %d 位", m.minLength)
	}

	// 检查大写字母
	if errResult == nil && m.requireUppercase {
		matched, _ := regexp.MatchString(`[A-Z]`, password)
		if !matched {
			errResult = fmt.Errorf("密码必须包含至少一个大写字母")
		}
	}

	// 检查小写字母
	if errResult == nil && m.requireLowercase {
		matched, _ := regexp.MatchString(`[a-z]`, password)
		if !matched {
			errResult = fmt.Errorf("密码必须包含至少一个小写字母")
		}
	}

	// 检查数字
	if errResult == nil && m.requireNumber {
		matched, _ := regexp.MatchString(`[0-9]`, password)
		if !matched {
			errResult = fmt.Errorf("密码必须包含至少一个数字")
		}
	}

	// 检查特殊字符
	if errResult == nil && m.requireSpecial {
		matched, _ := regexp.MatchString(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`, password)
		if !matched {
			errResult = fmt.Errorf("密码必须包含至少一个特殊字符")
		}
	}

	// 如果缓存已初始化，将验证结果存入缓存
	if m.cacheBackend != nil {
		// 存储错误结果或nil（表示密码强度足够）
		err := m.cacheBackend.Set(cacheKey, errResult, m.cacheTTL, cacheRegion)
		if err != nil {
			// 缓存设置失败，仅记录日志，不影响验证结果
			fmt.Printf("缓存密码强度验证结果失败: %v\n", err)
		}
	}

	return errResult
}
