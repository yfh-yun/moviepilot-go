package base

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"moviepilot-go/internal/apis/common"
)

// BaseHandler 所有handler的基类
// 提供公共的响应方法和日志功能
type BaseHandler struct {
	logger *zap.Logger
}

// NewBaseHandler 创建BaseHandler实例
func NewBaseHandler(logger *zap.Logger) *BaseHandler {
	return &BaseHandler{
		logger: logger,
	}
}

// Success 成功响应
func (h *BaseHandler) Success(c *gin.Context, data any) {
	common.Success(c, data)
}

// SuccessWithMessage 带消息的成功响应
func (h *BaseHandler) SuccessWithMessage(c *gin.Context, message string, data any) {
	common.SuccessWithMessage(c, message, data)
}

// SuccessPage 分页成功响应
func (h *BaseHandler) SuccessPage(c *gin.Context, data any, total int64, page, size int) {
	common.SuccessPage(c, data, total, page, size)
}

// Error 错误响应
func (h *BaseHandler) Error(c *gin.Context, code int, message string) {
	h.logger.Error(message, zap.Int("code", code))
	common.Error(c, code, message)
}

// BadRequest 400错误
func (h *BaseHandler) BadRequest(c *gin.Context, message string) {
	h.logger.Warn("Bad request", zap.String("message", message))
	common.BadRequest(c, message)
}

// Unauthorized 401错误
func (h *BaseHandler) Unauthorized(c *gin.Context, message string) {
	h.logger.Warn("Unauthorized", zap.String("message", message))
	common.Unauthorized(c, message)
}

// Forbidden 403错误
func (h *BaseHandler) Forbidden(c *gin.Context, message string) {
	h.logger.Warn("Forbidden", zap.String("message", message))
	common.Forbidden(c, message)
}

// NotFound 404错误
func (h *BaseHandler) NotFound(c *gin.Context, message string) {
	h.logger.Warn("Not found", zap.String("message", message))
	common.NotFound(c, message)
}

// InternalServerError 500错误
func (h *BaseHandler) InternalServerError(c *gin.Context, message string) {
	h.logger.Error("Internal server error", zap.String("message", message))
	common.InternalServerError(c, message)
}

// ValidateRequest 验证请求参数
func (h *BaseHandler) ValidateRequest(c *gin.Context, req any) error {
	return common.ValidateRequest(c, req)
}

// ValidateQuery 验证查询参数
func (h *BaseHandler) ValidateQuery(c *gin.Context, req any) error {
	return common.ValidateQuery(c, req)
}

// ValidateURI 验证URI参数
func (h *BaseHandler) ValidateURI(c *gin.Context, req any) error {
	return common.ValidateURI(c, req)
}

// GetLogger 获取logger
func (h *BaseHandler) GetLogger() *zap.Logger {
	return h.logger
}
