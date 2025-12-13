package security

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"moviepilot-go/pkg/cache"
)

// Claims JWT 声明
type Claims struct {
	UserID   uint     `json:"user_id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

// JWTManager JWT 管理器接口
type JWTManager interface {
	// GenerateAccessToken 生成访问令牌
	GenerateAccessToken(userID uint, username string, roles []string) (string, error)
	// GenerateRefreshToken 生成刷新令牌
	GenerateRefreshToken(userID uint, username string) (string, error)
	// ValidateToken 验证令牌
	ValidateToken(tokenString string) (*Claims, error)
	// RefreshAccessToken 刷新访问令牌
	RefreshAccessToken(refreshToken string) (string, error)
}

// jwtManager JWT 管理器实现
type jwtManager struct {
	secretKey            string
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
	cacheBackend         cache.CacheBackend // 缓存后端
	cacheTTL             time.Duration      // 缓存过期时间
}

// NewJWTManager 创建 JWT 管理器
func NewJWTManager(secretKey string, accessTokenDuration, refreshTokenDuration time.Duration) JWTManager {
	return &jwtManager{
		secretKey:            secretKey,
		accessTokenDuration:  accessTokenDuration,
		refreshTokenDuration: refreshTokenDuration,
		cacheBackend:         nil, // 默认不使用缓存
		cacheTTL:             0,
	}
}

// NewJWTManagerWithCache 创建带缓存的 JWT 管理器
func NewJWTManagerWithCache(secretKey string, accessTokenDuration, refreshTokenDuration time.Duration, cacheBackend cache.CacheBackend, cacheTTL time.Duration) JWTManager {
	return &jwtManager{
		secretKey:            secretKey,
		accessTokenDuration:  accessTokenDuration,
		refreshTokenDuration: refreshTokenDuration,
		cacheBackend:         cacheBackend,
		cacheTTL:             cacheTTL,
	}
}

// GenerateAccessToken 生成访问令牌
func (m *jwtManager) GenerateAccessToken(userID uint, username string, roles []string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.accessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "moviepilot",
			Subject:   username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secretKey))
}

// GenerateRefreshToken 生成刷新令牌
func (m *jwtManager) GenerateRefreshToken(userID uint, username string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.refreshTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "moviepilot",
			Subject:   username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secretKey))
}

// ValidateToken 验证令牌（带缓存）
func (m *jwtManager) ValidateToken(tokenString string) (*Claims, error) {
	// 缓存键
	cacheKey := fmt.Sprintf("jwt:%s", tokenString)
	cacheRegion := "security"

	// 如果缓存已初始化，尝试从缓存获取
	if m.cacheBackend != nil {
		cachedClaims, hit, err := m.cacheBackend.Get(cacheKey, cacheRegion)
		if err == nil && hit {
			if claims, ok := cachedClaims.(*Claims); ok {
				return claims, nil
			}
		}
	}

	// 缓存未命中，执行实际验证
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("token 解析失败: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("无效的 token")
	}

	// 如果缓存已初始化，将验证结果存入缓存
	if m.cacheBackend != nil {
		err = m.cacheBackend.Set(cacheKey, claims, m.cacheTTL, cacheRegion)
		if err != nil {
			// 缓存设置失败，仅记录日志，不影响验证结果
			fmt.Printf("缓存 JWT 验证结果失败: %v\n", err)
		}
	}

	return claims, nil
}

// RefreshAccessToken 刷新访问令牌
func (m *jwtManager) RefreshAccessToken(refreshToken string) (string, error) {
	// 验证刷新令牌
	claims, err := m.ValidateToken(refreshToken)
	if err != nil {
		return "", fmt.Errorf("刷新令牌无效: %w", err)
	}

	// 生成新的访问令牌
	return m.GenerateAccessToken(claims.UserID, claims.Username, claims.Roles)
}
