package middlewares

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"moviepilot-go/pkg/logger"
)

// RequestIDMiddleware 在请求上下文中注入 request_id，便于日志追踪。
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()
		}

		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, logger.ContextKeyRequestID, requestID)
		c.Request = c.Request.WithContext(ctx)
		c.Writer.Header().Set("X-Request-ID", requestID)
		c.Set(string(logger.ContextKeyRequestID), requestID)

		c.Next()
	}
}
