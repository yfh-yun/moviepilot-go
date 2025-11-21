// Package jwt JWT认证系统
// 提供Token生成、验证、刷新功能
package jwt

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

// Claims JWT声明结构
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// TokenPair Token对（访问Token和刷新Token）
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // 访问Token过期时间（秒）
	TokenType    string `json:"token_type"` // Token类型，固定为"Bearer"
}

var (
	// ErrInvalidToken Token无效
	ErrInvalidToken = errors.New("invalid token")
	// ErrExpiredToken Token已过期
	ErrExpiredToken = errors.New("token expired")
	// ErrInvalidSignature 签名无效
	ErrInvalidSignature = errors.New("invalid signature")
)

// generateJTI 生成JWT ID (JTI)
func generateJTI() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// 如果随机数生成失败，使用时间戳作为fallback
		return fmt.Sprintf("jti-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// GenerateToken 生成访问Token和刷新Token
// userID: 用户ID
// username: 用户名
// role: 用户角色
// 返回: TokenPair包含访问Token和刷新Token
func GenerateToken(userID uint, username, role string) (*TokenPair, error) {
	// 输入验证
	if userID == 0 {
		return nil, fmt.Errorf("user ID cannot be zero")
	}
	if username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}
	if role == "" {
		return nil, fmt.Errorf("role cannot be empty")
	}

	secret := getJWTSecret()

	// 访问Token过期时间（从配置读取，默认24小时）
	accessExpireMinutes := viper.GetInt("jwt.expire_minutes")
	if accessExpireMinutes == 0 {
		accessExpireMinutes = 1440 // 默认24小时
	}
	accessExpire := time.Now().Add(time.Duration(accessExpireMinutes) * time.Minute)

	// 刷新Token过期时间（默认7天）
	refreshExpireDays := viper.GetInt("jwt.refresh_expire_days")
	if refreshExpireDays == 0 {
		refreshExpireDays = 7
	}
	refreshExpire := time.Now().Add(time.Duration(refreshExpireDays) * 24 * time.Hour)

	// 生成访问Token
	now := time.Now()
	accessClaims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        generateJTI(),
			ExpiresAt: jwt.NewNumericDate(accessExpire),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "moviepilot-go",
			Subject:   username,
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(secret))
	if err != nil {
		return nil, fmt.Errorf("generate access token failed: %w", err)
	}

	// 生成刷新Token
	refreshClaims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        generateJTI(),
			ExpiresAt: jwt.NewNumericDate(refreshExpire),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "moviepilot-go",
			Subject:   username,
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(secret))
	if err != nil {
		return nil, fmt.Errorf("generate refresh token failed: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    int64(accessExpireMinutes * 60), // 转换为秒
		TokenType:    "Bearer",
	}, nil
}

// ParseToken 解析Token并验证
// tokenString: Token字符串
// 返回: Claims解析后的声明，error错误信息
func ParseToken(tokenString string) (*Claims, error) {
	secret := getJWTSecret()

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		// 检查是否是过期错误
		if err.Error() == "token is expired" {
			return nil, ErrExpiredToken
		}
		// 检查是否是签名错误
		if err.Error() == "signature is invalid" {
			return nil, ErrInvalidSignature
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// RefreshToken 刷新访问Token
// refreshTokenString: 刷新Token字符串
// 返回: TokenPair新的Token对，error错误信息
func RefreshToken(refreshTokenString string) (*TokenPair, error) {
	// 解析刷新Token
	claims, err := ParseToken(refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("parse refresh token failed: %w", err)
	}

	// 生成新的Token对
	return GenerateToken(claims.UserID, claims.Username, claims.Role)
}

// ValidateToken 验证Token是否有效
// tokenString: Token字符串
// 返回: bool是否有效，error错误信息
func ValidateToken(tokenString string) (bool, error) {
	_, err := ParseToken(tokenString)
	if err != nil {
		return false, err
	}
	return true, nil
}

// parseTokenSafe 安全解析Token的辅助函数
func parseTokenSafe(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, ErrInvalidToken
	}
	return ParseToken(tokenString)
}

// GetUserIDFromToken 从Token中获取用户ID
// tokenString: Token字符串
// 返回: uint用户ID，error错误信息
func GetUserIDFromToken(tokenString string) (uint, error) {
	claims, err := parseTokenSafe(tokenString)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}

// GetUsernameFromToken 从Token中获取用户名
// tokenString: Token字符串
// 返回: string用户名，error错误信息
func GetUsernameFromToken(tokenString string) (string, error) {
	claims, err := parseTokenSafe(tokenString)
	if err != nil {
		return "", err
	}
	return claims.Username, nil
}

// GetRoleFromToken 从Token中获取角色
// tokenString: Token字符串
// 返回: string角色，error错误信息
func GetRoleFromToken(tokenString string) (string, error) {
	claims, err := parseTokenSafe(tokenString)
	if err != nil {
		return "", err
	}
	return claims.Role, nil
}

// getJWTSecret 获取JWT密钥
func getJWTSecret() string {
	secret := viper.GetString("jwt.secret")
	if secret == "" {
		// 在开发环境使用默认密钥，生产环境必须配置密钥
		if viper.GetString("app.env") == "production" {
			panic("JWT secret must be configured in production environment")
		}
		secret = "moviepilot-dev-secret-key"
	}
	return secret
}
