package middlewares

import (
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"moviepilot-go/pkg/jwt"
	"moviepilot-go/pkg/logger"
	"moviepilot-go/pkg/response"
)

// AuthMiddleware JWT 认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过健康检查端点
		if c.FullPath() == "/health" {
			c.Next()
			return
		}

		// 对于 workflow API，允许通过 API Key 认证（用于 CLI 工具）
		if strings.HasPrefix(c.FullPath(), "/api/workflows") {
			apiKey := c.GetHeader("X-API-Key")
			if apiKey != "" {
				// TODO: 验证 API Key，暂时跳过
				c.Next()
				return
			}
		}

		// JWT 认证
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Warn("missing authorization header",
				zap.String("path", c.FullPath()),
				zap.String("client_ip", c.ClientIP()),
			)
			response.Unauthorized(c, "missing authorization header")
			return
		}

		// 检查 Bearer 前缀
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			logger.Warn("invalid authorization header format",
				zap.String("path", c.FullPath()),
				zap.String("client_ip", c.ClientIP()),
			)
			response.Unauthorized(c, "invalid authorization header format")
			return
		}

		// 验证 token
		claims, err := jwt.ParseToken(parts[1])
		if err != nil {
			logger.Warn("invalid token",
				zap.String("path", c.FullPath()),
				zap.String("client_ip", c.ClientIP()),
				zap.Error(err),
			)
			response.Unauthorized(c, "invalid token")
			return
		}

		// 将用户信息存储到上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		logger.Debug("user authenticated",
			zap.Uint("user_id", claims.UserID),
			zap.String("username", claims.Username),
			zap.String("path", c.FullPath()),
		)

		c.Next()
	}
}
