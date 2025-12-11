package security

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// ErrTokenNotFound 令牌未找到
	ErrTokenNotFound = errors.New("token not found")
	// ErrTokenInvalid 令牌无效
	ErrTokenInvalid = errors.New("token is invalid")
	// ErrTokenExpired 令牌已过期
	ErrTokenExpired = errors.New("token has expired")
	// ErrTokenPurposeMismatch 令牌用途不匹配
	ErrTokenPurposeMismatch = errors.New("token purpose mismatch")
)

// TokenPurpose 令牌用途
type TokenPurpose string

const (
	// PurposeAuth 认证令牌
	PurposeAuth TokenPurpose = "authentication"
	// PurposeResource 资源访问令牌
	PurposeResource TokenPurpose = "resource"
)

// TokenClaims JWT声明
type TokenClaims struct {
	UserID    string       `json:"sub"`
	Username  string       `json:"username"`
	SuperUser bool         `json:"super_user"`
	Level     int          `json:"level"`
	Purpose   TokenPurpose `json:"purpose"`
	jwt.RegisteredClaims
}

// CreateAccessToken 创建访问令牌
func CreateAccessToken(secretKey string, userID, username string, superUser bool, level int, purpose TokenPurpose, expiresDelta *time.Duration) (string, error) {
	var ttl time.Duration

	if expiresDelta != nil {
		if expiresDelta.Seconds() <= 0 {
			return "", fmt.Errorf("expiration must be positive")
		}
		ttl = *expiresDelta
	} else {
		// 默认过期时间：认证令牌30分钟，资源令牌1小时
		if purpose == PurposeResource {
			ttl = 1 * time.Hour
		} else {
			ttl = 30 * time.Minute
		}
	}

	now := time.Now().UTC()
	claims := &TokenClaims{
		UserID:    userID,
		Username:  username,
		SuperUser: superUser,
		Level:     level,
		Purpose:   purpose,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// VerifyToken 验证令牌
func VerifyToken(secretKey, tokenStr string, purpose TokenPurpose) (*TokenClaims, error) {
	if tokenStr == "" {
		return nil, ErrTokenNotFound
	}

	token, err := jwt.ParseWithClaims(tokenStr, &TokenClaims{}, func(t *jwt.Token) (any, error) {
		// 验证签名算法
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}

	if claims.Purpose != purpose {
		return nil, ErrTokenPurposeMismatch
	}

	return claims, nil
}

// RefreshToken 刷新令牌
func RefreshToken(secretKey, oldTokenStr string, purpose TokenPurpose) (string, error) {
	claims, err := VerifyToken(secretKey, oldTokenStr, purpose)
	if err != nil {
		return "", err
	}

	// 创建新令牌，使用相同的用户信息
	return CreateAccessToken(secretKey, claims.UserID, claims.Username, claims.SuperUser, claims.Level, purpose, nil)
}

// ExtractTokenFromBearer 从Bearer头中提取令牌
func ExtractTokenFromBearer(authHeader string) (string, error) {
	if authHeader == "" {
		return "", ErrTokenNotFound
	}

	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return "", ErrTokenInvalid
	}

	return authHeader[7:], nil
}
