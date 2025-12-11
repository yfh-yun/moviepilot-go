package middleware

import (
	"moviepilot-go/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RequestLoggingMiddleware 统一的 API 请求日志记录中间件
func RequestLoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			logger.Info("HTTP request",
				zap.String("method", param.Method),
				zap.String("path", param.Path),
				zap.Int("status", param.StatusCode),
				zap.Duration("latency", param.Latency),
				zap.String("client_ip", param.ClientIP),
				zap.String("user_agent", param.Request.UserAgent()),
				zap.String("error", param.ErrorMessage),
			)
			return ""
		},
	})
}
