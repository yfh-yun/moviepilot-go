// Package security 安全认证中间件
package security

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMode 认证模式
type AuthMode string

const (
	AuthModeJWT      AuthMode = "jwt"
	AuthModeAPIKey   AuthMode = "apikey"
	AuthModeResource AuthMode = "resource"
)

// AuthConfig 认证配置
type AuthConfig struct {
	Mode         AuthMode
	SkipPaths    []string
	RequiredScopes []string
}

// DefaultAuthConfig 默认认证配置
func DefaultAuthConfig() *AuthConfig {
	return &AuthConfig{
		Mode:         AuthModeJWT,
		SkipPaths:    []string{"/health", "/login", "/register"},
		RequiredScopes: []string{"read", "write"},
	}
}

// AuthMiddleware 认证中间件
func AuthMiddleware(jwtManager *JWTManager, apiKeyManager *APIKeyManager, config *AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查跳过路径
		for _, path := range config.SkipPaths {
			if c.Request.URL.Path == path {
				c.Next()
				return
			}
		}

		var user *UserClaims
		var err error

		// 根据认证模式进行验证
		switch config.Mode {
		case AuthModeJWT:
			user, err = extractJWTFromRequest(c, jwtManager)
		case AuthModeAPIKey:
			user, err = extractAPIKeyFromRequest(c, apiKeyManager)
		case AuthModeResource:
			user, err = extractResourceTokenFromRequest(c, jwtManager)
		default:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unsupported auth mode"})
			c.Abort()
			return
		}

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// 检查权限范围
		if !hasRequiredScopes(user, config.RequiredScopes) {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文
		c.Set("user", user)
		c.Next()
	}
}

// extractJWTFromRequest 从请求中提取JWT
func extractJWTFromRequest(c *gin.Context, jwtManager *JWTManager) (*UserClaims, error) {
	// 从Authorization header获取
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			return jwtManager.VerifyToken(parts[1])
		}
	}

	// 从查询参数获取
	token := c.Query("token")
	if token != "" {
		return jwtManager.VerifyToken(token)
	}

	return nil, fmt.Errorf("no token provided")
}

// extractAPIKeyFromRequest 从请求中提取API密钥
func extractAPIKeyFromRequest(c *gin.Context, apiKeyManager *APIKeyManager) (*UserClaims, error) {
	// 从X-API-Key header获取
	apiKey := c.GetHeader("X-API-Key")
	if apiKey == "" {
		// 从查询参数获取
		apiKey = c.Query("apikey")
	}

	if apiKey == "" {
		return nil, fmt.Errorf("no API key provided")
	}

	keyInfo, err := apiKeyManager.ValidateAPIKey(apiKey)
	if err != nil {
		return nil, err
	}

	// 这里需要根据API Key获取用户信息
	// 简化实现，实际应该从数据库查询
	return &UserClaims{
		UserID:      keyInfo.UserID,
		Username:    "api_user",
		IsSuperUser: false,
		Level:       1,
		Purpose:     "api_access",
	}, nil
}

// extractResourceTokenFromRequest 从请求中提取资源Token
func extractResourceTokenFromRequest(c *gin.Context, jwtManager *JWTManager) (*UserClaims, error) {
	// 从Cookie获取
	token, err := c.Cookie("moviepilot")
	if err != nil {
		return nil, fmt.Errorf("no resource token provided")
	}

	claims, err := jwtManager.VerifyToken(token)
	if err != nil {
		return nil, err
	}

	// 检查Token用途
	if claims.Purpose != "resource_access" {
		return nil, fmt.Errorf("invalid token purpose")
	}

	return claims, nil
}

// hasRequiredScopes 检查用户是否具有所需权限
func hasRequiredScopes(user *UserClaims, requiredScopes []string) bool {
	// 超级用户拥有所有权限
	if user.IsSuperUser {
		return true
	}

	// 这里应该根据用户的权限范围进行检查
	// 简化实现，实际应该从数据库查询用户权限
	return user.Level >= 1
}

// RequireSuperUser 超级用户权限中间件
func RequireSuperUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			c.Abort()
			return
		}

		claims, ok := user.(*UserClaims)
		if !ok || !claims.IsSuperUser {
			c.JSON(http.StatusForbidden, gin.H{"error": "super user required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireLevel 需要特定用户级别的中间件
func RequireLevel(minLevel int) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			c.Abort()
			return
		}

		claims, ok := user.(*UserClaims)
		if !ok || claims.Level < minLevel {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient user level"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetCurrentUser 获取当前用户
func GetCurrentUser(c *gin.Context) (*UserClaims, bool) {
	user, exists := c.Get("user")
	if !exists {
		return nil, false
	}

	claims, ok := user.(*UserClaims)
	return claims, ok
}

// GenerateResourceToken 生成资源访问Token
func GenerateResourceToken(jwtManager *JWTManager, userID uint, username string, isSuperUser bool, level int) (string, error) {
	return jwtManager.GenerateToken(userID, username, isSuperUser, level, "resource_access")
}