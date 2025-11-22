package middlewares

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
	"moviepilot-go/pkg/response"
)

// RecoveryMiddleware 捕获 panic 并输出统一 JSON。
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic recovered",
					zap.Any("error", r),
					zap.String("path", c.FullPath()),
				)
				if !c.Writer.Written() {
					response.Error(c, response.CodeServerError, "服务器内部错误")
				} else {
					c.AbortWithStatusJSON(http.StatusInternalServerError, response.Response{
						Success:   false,
						Code:      response.CodeServerError,
						Message:   "服务器内部错误",
						Timestamp: time.Now().Unix(),
					})
				}
			}
		}()
		c.Next()
	}
}
