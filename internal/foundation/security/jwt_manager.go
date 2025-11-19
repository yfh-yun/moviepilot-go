// Package security 安全认证模块
package security

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// JWTManager JWT管理器
type JWTManager struct {
	secretKey     string
	tokenDuration time.Duration
}

// UserClaims 用户声明
type UserClaims struct {
	UserID     uint   `json:"user_id"`
	Username   string `json:"username"`
	IsSuperUser bool  `json:"is_super_user"`
	Level      int    `json:"level"`
	Purpose    string `json:"purpose"`
	jwt.RegisteredClaims
}

// NewJWTManager 创建JWT管理器
func NewJWTManager(secretKey string, tokenDuration time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:     secretKey,
		tokenDuration: tokenDuration,
	}
}

// GenerateToken 生成JWT Token
func (manager *JWTManager) GenerateToken(userID uint, username string, isSuperUser bool, level int, purpose string) (string, error) {
	now := time.Now()
	claims := UserClaims{
		UserID:      userID,
		Username:    username,
		IsSuperUser: isSuperUser,
		Level:       level,
		Purpose:     purpose,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(manager.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "moviepilot",
			Subject:   fmt.Sprintf("%d", userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(manager.secretKey))
}

// VerifyToken 验证JWT Token
func (manager *JWTManager) VerifyToken(tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(manager.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// RefreshToken 刷新Token
func (manager *JWTManager) RefreshToken(tokenString string) (string, error) {
	claims, err := manager.VerifyToken(tokenString)
	if err != nil {
		return "", err
	}

	// 检查Token是否即将过期（30分钟内）
	if time.Until(claims.ExpiresAt.Time) > 30*time.Minute {
		return "", fmt.Errorf("token is not eligible for refresh")
	}

	return manager.GenerateToken(claims.UserID, claims.Username, claims.IsSuperUser, claims.Level, claims.Purpose)
}

// PasswordManager 密码管理器
type PasswordManager struct {
	cost int
}

// NewPasswordManager 创建密码管理器
func NewPasswordManager() *PasswordManager {
	return &PasswordManager{
		cost: bcrypt.DefaultCost,
	}
}

// HashPassword 哈希密码
func (pm *PasswordManager) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), pm.cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// VerifyPassword 验证密码
func (pm *PasswordManager) VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateRandomString 生成随机字符串
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random string: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// APIKey API密钥管理
type APIKey struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Key       string    `json:"key"`
	UserID    uint      `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	IsActive  bool      `json:"is_active"`
	Scopes    []string  `json:"scopes"`
}

// APIKeyManager API密钥管理器
type APIKeyManager struct {
	keys map[string]*APIKey
}

// NewAPIKeyManager 创建API密钥管理器
func NewAPIKeyManager() *APIKeyManager {
	return &APIKeyManager{
		keys: make(map[string]*APIKey),
	}
}

// GenerateAPIKey 生成API密钥
func (akm *APIKeyManager) GenerateAPIKey(name string, userID uint, scopes []string, expiresAt *time.Time) (*APIKey, error) {
	keyID, err := GenerateRandomString(16)
	if err != nil {
		return nil, err
	}

	keyValue, err := GenerateRandomString(32)
	if err != nil {
		return nil, err
	}

	apiKey := &APIKey{
		ID:        keyID,
		Name:      name,
		Key:       keyValue,
		UserID:    userID,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
		IsActive:  true,
		Scopes:    scopes,
	}

	akm.keys[keyValue] = apiKey
	return apiKey, nil
}

// ValidateAPIKey 验证API密钥
func (akm *APIKeyManager) ValidateAPIKey(key string) (*APIKey, error) {
	apiKey, exists := akm.keys[key]
	if !exists {
		return nil, fmt.Errorf("invalid API key")
	}

	if !apiKey.IsActive {
		return nil, fmt.Errorf("API key is inactive")
	}

	if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
		return nil, fmt.Errorf("API key has expired")
	}

	return apiKey, nil
}

// RevokeAPIKey 撤销API密钥
func (akm *APIKeyManager) RevokeAPIKey(key string) error {
	apiKey, exists := akm.keys[key]
	if !exists {
		return fmt.Errorf("API key not found")
	}

	apiKey.IsActive = false
	return nil
}