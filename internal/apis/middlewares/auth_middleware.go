package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"moviepilot-go/pkg/security"
)

// AuthMiddleware JWT 认证中间件
func AuthMiddleware(jwtManager security.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Header 获取 Token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "未提供认证令牌",
			})
			c.Abort()
			return
		}

		// 检查 Bearer 前缀
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "认证令牌格式错误",
			})
			c.Abort()
			return
		}

		token := parts[1]

		// 验证 Token
		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "认证令牌无效",
			})
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("roles", claims.Roles)

		c.Next()
	}
}

// GetUserID 从上下文获取用户ID
func GetUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	id, ok := userID.(uint)
	return id, ok
}

// GetUsername 从上下文获取用户名
func GetUsername(c *gin.Context) (string, bool) {
	username, exists := c.Get("username")
	if !exists {
		return "", false
	}
	name, ok := username.(string)
	return name, ok
}

// GetRoles 从上下文获取角色列表
func GetRoles(c *gin.Context) ([]string, bool) {
	roles, exists := c.Get("roles")
	if !exists {
		return nil, false
	}
	roleList, ok := roles.([]string)
	return roleList, ok
}
