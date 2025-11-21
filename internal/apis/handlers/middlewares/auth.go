// Package middleware 提供HTTP中间件
package middlewares

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// JWTClaims JWT声明
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// AuthMiddleware 认证中间件
type AuthMiddleware struct {
	secretKey []byte
	logger    *zap.Logger
}

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware(secretKey string, logger *zap.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		secretKey: []byte(secretKey),
		logger:    logger,
	}
}

// JWTAuthMiddleware JWT认证中间件
func (am *AuthMiddleware) JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"code":    1002,
				"error":   "缺少认证token",
			})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"code":    1002,
				"error":   "token格式错误",
			})
			c.Abort()
			return
		}

		// 提取token
		tokenString := authHeader[7:]
		claims := &JWTClaims{}

		// 解析token
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// 验证签名算法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return am.secretKey, nil
		})

		if err != nil {
			am.logger.Warn("JWT token解析失败", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"code":    1002,
				"error":   "token无效",
			})
			c.Abort()
			return
		}

		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"code":    1002,
				"error":   "token已过期",
			})
			c.Abort()
			return
		}

		// 检查token是否过期
		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"code":    1002,
				"error":   "token已过期",
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// RequireRole 角色验证中间件
func (am *AuthMiddleware) RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"code":    1002,
				"error":   "未找到角色信息",
			})
			c.Abort()
			return
		}

		if role != requiredRole && role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"code":    1003,
				"error":   "权限不足",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin 需要管理员权限
func (am *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return am.RequireRole("admin")
}

// GenerateToken 生成JWT token
func (am *AuthMiddleware) GenerateToken(userID, username, role string, expiresIn time.Duration) (string, error) {
	claims := JWTClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(am.secretKey)
}

// RefreshToken 刷新token
func (am *AuthMiddleware) RefreshToken(tokenString string) (string, error) {
	claims := &JWTClaims{}

	// 解析token（不验证过期时间）
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return am.secretKey, nil
	}, jwt.WithoutClaimsValidation())

	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", jwt.ErrSignatureInvalid
	}

	// 生成新token
	return am.GenerateToken(claims.UserID, claims.Username, claims.Role, 24*time.Hour)
}