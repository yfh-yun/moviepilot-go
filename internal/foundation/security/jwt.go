package security

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTManager JWT管理器
type JWTManager struct {
	secretKey          []byte
	expiration         time.Duration
	refreshExpiration  time.Duration
	algorithm          string
}

// NewJWTManager 创建新的JWT管理器
func NewJWTManager(config *SecurityConfig) *JWTManager {
	return &JWTManager{
		secretKey:         []byte(config.JWTSecretKey),
		expiration:        config.JWTExpiration,
		refreshExpiration: config.JWTRefreshExpiration,
		algorithm:         config.JWTAlgorithm,
	}
}

// GenerateTokenPair 生成令牌对
func (j *JWTManager) GenerateTokenPair(user *User) (*TokenPair, error) {
	if user == nil {
		return nil, errors.New("user cannot be nil")
	}

	// 生成访问令牌
	accessToken, accessExpires, err := j.generateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// 生成刷新令牌
	refreshToken, refreshExpires, err := j.generateRefreshToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		ExpiresIn:        int64(time.Until(accessExpires).Seconds()),
		RefreshExpiresIn: int64(time.Until(refreshExpires).Seconds()),
		TokenType:        "Bearer",
		CreatedAt:        time.Now(),
	}, nil
}

// generateAccessToken 生成访问令牌
func (j *JWTManager) generateAccessToken(user *User) (string, time.Time, error) {
	expirationTime := time.Now().Add(j.expiration)
	jti, err := j.generateJTI()
	if err != nil {
		return "", time.Time{}, err
	}

	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		RoleIDs:  user.RoleIDs,
		Exp:      expirationTime.Unix(),
		Iat:      time.Now().Unix(),
		Iss:      "moviepilot",
		Sub:      user.ID,
		Jti:      jti,
		Type:     TokenTypeAccess,
	}

	token, err := j.createToken(claims)
	if err != nil {
		return "", time.Time{}, err
	}

	return token, expirationTime, nil
}

// generateRefreshToken 生成刷新令牌
func (j *JWTManager) generateRefreshToken(user *User) (string, time.Time, error) {
	expirationTime := time.Now().Add(j.refreshExpiration)
	jti, err := j.generateJTI()
	if err != nil {
		return "", time.Time{}, err
	}

	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		RoleIDs:  user.RoleIDs,
		Exp:      expirationTime.Unix(),
		Iat:      time.Now().Unix(),
		Iss:      "moviepilot",
		Sub:      user.ID,
		Jti:      jti,
		Type:     TokenTypeRefresh,
	}

	token, err := j.createToken(claims)
	if err != nil {
		return "", time.Time{}, err
	}

	return token, expirationTime, nil
}

// createToken 创建JWT令牌
func (j *JWTManager) createToken(claims *Claims) (string, error) {
	// 根据算法选择签名方法
	var signingMethod jwt.SigningMethod
	switch j.algorithm {
	case "HS256":
		signingMethod = jwt.SigningMethodHS256
	case "HS384":
		signingMethod = jwt.SigningMethodHS384
	case "HS512":
		signingMethod = jwt.SigningMethodHS512
	default:
		return "", fmt.Errorf("unsupported algorithm: %s", j.algorithm)
	}

	token := jwt.NewWithClaims(signingMethod, claims)
	tokenString, err := token.SignedString(j.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken 验证令牌
func (j *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	// 解析令牌
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return j.secretKey, nil
	})

	if err != nil {
		// 检查是否是过期错误
		var validationError *jwt.ValidationError
		if errors.As(err, &validationError) && validationError.Errors&jwt.ValidationErrorExpired != 0 {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	// 验证令牌有效性
	if !token.Valid {
		return nil, ErrInvalidToken
	}

	// 提取声明
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// RefreshToken 使用刷新令牌获取新的访问令牌
func (j *JWTManager) RefreshToken(refreshToken string) (*TokenPair, error) {
	// 验证刷新令牌
	claims, err := j.ValidateToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// 确保是刷新令牌
	if claims.Type != TokenTypeRefresh {
		return nil, errors.New("not a refresh token")
	}

	// 这里应该根据claims.UserID从数据库获取用户信息
	// 为了简化，我们创建一个基本的用户对象
	// 在实际应用中，应该从数据库获取完整的用户信息
	user := &User{
		ID:       claims.UserID,
		Username: claims.Username,
		RoleIDs:  claims.RoleIDs,
	}

	// 生成新的令牌对
	return j.GenerateTokenPair(user)
}

// generateJTI 生成JWT ID
func (j *JWTManager) generateJTI() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return base64.URLEncoding.EncodeToString(b), nil
}

// ExtractTokenFromHeader 从Authorization头提取令牌
func ExtractTokenFromHeader(authHeader string) (string, error) {
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return "", errors.New("invalid authorization header format")
	}

	return authHeader[7:], nil
}

// ValidateTokenType 验证令牌类型
func ValidateTokenType(claims *Claims, expectedType string) error {
	if claims.Type != expectedType {
		return fmt.Errorf("invalid token type, expected %s got %s", expectedType, claims.Type)
	}
	return nil
}

// IsTokenExpired 检查令牌是否即将过期（在指定的阈值内）
func IsTokenExpired(claims *Claims, threshold time.Duration) bool {
	expirationTime := time.Unix(claims.Exp, 0)
	return time.Until(expirationTime) < threshold
}

// CalculateTokenRemainingTime 计算令牌剩余有效时间
func CalculateTokenRemainingTime(claims *Claims) time.Duration {
	expirationTime := time.Unix(claims.Exp, 0)
	remaining := time.Until(expirationTime)
	if remaining < 0 {
		return 0
	}
	return remaining
}