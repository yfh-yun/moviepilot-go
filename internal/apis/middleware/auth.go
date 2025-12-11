package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"moviepilot-go/pkg/logger"
	"moviepilot-go/pkg/security"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthMiddleware JWT 认证中间件
func AuthMiddleware() gin.HandlerFunc {
	log := logger.GetLogger()

	return func(c *gin.Context) {
		// 获取 Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Warn("Missing authorization header",
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method))

			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "未提供认证信息",
			})
			c.Abort()
			return
		}

		// 检查 Bearer 前缀
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.Warn("Invalid authorization header format",
				zap.String("header", authHeader))

			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "认证信息格式错误",
			})
			c.Abort()
			return
		}

		token := parts[1]

		// 解析 Token
		// 使用默认密钥和超时时间创建 JWT 管理器
		jwtManager := security.NewJWTManager("default_secret_key", 24*time.Hour, 7*24*time.Hour)
		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			log.Warn("Invalid token",
				zap.Error(err),
				zap.String("path", c.Request.URL.Path))

			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "无效的认证信息",
			})
			c.Abort()
			return
		}

		// 检查是否为管理员角色
		isAdmin := false
		for _, role := range claims.Roles {
			if role == "admin" {
				isAdmin = true
				break
			}
		}

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("is_admin", isAdmin)

		log.Debug("User authenticated",
			zap.String("user_id", fmt.Sprintf("%d", claims.UserID)),
			zap.String("username", claims.Username),
			zap.Bool("is_admin", isAdmin))

		c.Next()
	}
}

// OptionalAuthMiddleware 可选认证中间件（不强制要求认证）
func OptionalAuthMiddleware() gin.HandlerFunc {
	log := logger.GetLogger()

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		token := parts[1]
		// 使用默认密钥和超时时间创建 JWT 管理器
		jwtManager := security.NewJWTManager("default_secret_key", 24*time.Hour, 7*24*time.Hour)
		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			log.Debug("Optional auth failed", zap.Error(err))
			c.Next()
			return
		}

		// 检查是否为管理员角色
		isAdmin := false
		for _, role := range claims.Roles {
			if role == "admin" {
				isAdmin = true
				break
			}
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("is_admin", isAdmin)

		c.Next()
	}
}

// AdminMiddleware 管理员权限中间件（需要先经过 AuthMiddleware）
func AdminMiddleware() gin.HandlerFunc {
	log := logger.GetLogger()

	return func(c *gin.Context) {
		isAdmin, exists := c.Get("is_admin")
		if !exists || !isAdmin.(bool) {
			log.Warn("Admin access denied",
				zap.String("path", c.Request.URL.Path),
				zap.Any("user_id", c.GetString("user_id")))

			c.JSON(http.StatusForbidden, gin.H{
				"error": "需要管理员权限",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetUserID 从上下文获取用户 ID
func GetUserID(c *gin.Context) string {
	userID, exists := c.Get("user_id")
	if !exists {
		return ""
	}
	return userID.(string)
}

// GetUsername 从上下文获取用户名
func GetUsername(c *gin.Context) string {
	username, exists := c.Get("username")
	if !exists {
		return ""
	}
	return username.(string)
}

// IsAdmin 检查是否是管理员
func IsAdmin(c *gin.Context) bool {
	isAdmin, exists := c.Get("is_admin")
	if !exists {
		return false
	}
	return isAdmin.(bool)
}
