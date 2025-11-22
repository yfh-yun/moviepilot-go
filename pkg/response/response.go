// Package response 统一API响应格式
// 提供标准化的API响应格式和错误处理机制
package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"moviepilot-go/pkg/logger"
)

// ResponseCode 响应状态码
type ResponseCode int

const (
	// CodeSuccess 成功
	CodeSuccess ResponseCode = 0

	// CodeInvalidParams 参数错误
	CodeInvalidParams ResponseCode = 1001

	// CodeUnauthorized 未授权
	CodeUnauthorized ResponseCode = 1002

	// CodeForbidden 禁止访问
	CodeForbidden ResponseCode = 1003

	// CodeNotFound 资源未找到
	CodeNotFound ResponseCode = 1004

	// CodeServerError 服务器错误
	CodeServerError ResponseCode = 5000

	// CodeDatabaseError 数据库错误
	CodeDatabaseError ResponseCode = 5001

	// CodeServiceUnavailable 服务不可用
	CodeServiceUnavailable ResponseCode = 5002

	// CodeRateLimit 请求频率限制
	CodeRateLimit ResponseCode = 4001

	// CodeOTPRequired 需要双因素认证
	CodeOTPRequired ResponseCode = 2001

	// CodeInvalidOTP 双因素认证码错误
	CodeInvalidOTP ResponseCode = 2002

	// CodeUserDisabled 用户已被禁用
	CodeUserDisabled ResponseCode = 2003

	// CodeUserExists 用户已存在
	CodeUserExists ResponseCode = 2004

	// CodeInvalidPassword 密码错误
	CodeInvalidPassword ResponseCode = 2005
)

// Response 统一响应结构
type Response struct {
	Success   bool         `json:"success"`
	Code      ResponseCode `json:"code"`
	Message   string       `json:"message"`
	Data      interface{}  `json:"data,omitempty"`
	Timestamp int64        `json:"timestamp"`
}

// PaginatedResponse 分页响应结构
type PaginatedResponse struct {
	Items      interface{} `json:"items"`
	TotalCount int64       `json:"total_count"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success:   true,
		Code:      CodeSuccess,
		Message:   "操作成功",
		Data:      data,
		Timestamp: time.Now().Unix(),
	})
}

// Error 错误响应
func Error(c *gin.Context, code ResponseCode, message string) {
	var statusCode int

	// 根据错误码确定HTTP状态码
	switch code {
	case CodeInvalidParams:
		statusCode = http.StatusBadRequest
	case CodeUnauthorized:
		statusCode = http.StatusUnauthorized
	case CodeForbidden:
		statusCode = http.StatusForbidden
	case CodeNotFound:
		statusCode = http.StatusNotFound
	case CodeRateLimit:
		statusCode = http.StatusTooManyRequests
	default:
		statusCode = http.StatusInternalServerError
	}

	c.JSON(statusCode, Response{
		Success:   false,
		Code:      code,
		Message:   message,
		Timestamp: time.Now().Unix(),
	})
}

// SuccessWithPagination 带分页的成功响应
func SuccessWithPagination(c *gin.Context, items interface{}, totalCount int64, page, pageSize int) {
	totalPages := 0
	if pageSize > 0 {
		totalPages = int((totalCount + int64(pageSize) - 1) / int64(pageSize))
	}

	data := PaginatedResponse{
		Items:      items,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	Success(c, data)
}

// ErrorWithLog 记录错误日志的错误响应
func ErrorWithLog(c *gin.Context, code ResponseCode, message string, err error) {
	// 记录错误日志
	if err != nil {
		logger.Error("API Error",
			zap.Int("code", int(code)),
			zap.String("message", message),
			zap.String("error", err.Error()),
			zap.String("path", c.Request.URL.Path),
			zap.String("func", "ErrorWithLog"),
		)
	}

	Error(c, code, message)
}

// BadRequest 400错误响应
func BadRequest(c *gin.Context, message string) {
	if message == "" {
		message = "请求参数错误"
	}
	Error(c, CodeInvalidParams, message)
}

// InvalidParams 参数错误响应
func InvalidParams(c *gin.Context, message string) {
	if message == "" {
		message = "参数错误"
	}
	Error(c, CodeInvalidParams, message)
}

// Unauthorized 未授权错误响应
func Unauthorized(c *gin.Context, message string) {
	if message == "" {
		message = "未授权访问"
	}
	Error(c, CodeUnauthorized, message)
}

// NotFound 资源未找到错误响应
func NotFound(c *gin.Context, message string) {
	if message == "" {
		message = "资源未找到"
	}
	Error(c, CodeNotFound, message)
}

// ServerError 服务器错误响应
func ServerError(c *gin.Context, message string) {
	if message == "" {
		message = "服务器内部错误"
	}
	Error(c, CodeServerError, message)
}

// DatabaseError 数据库错误响应
func DatabaseError(c *gin.Context, err error) {
	ErrorWithLog(c, CodeDatabaseError, "数据库操作失败", err)
}

// RateLimit 请求频率限制错误响应
func RateLimit(c *gin.Context) {
	Error(c, CodeRateLimit, "请求过于频繁，请稍后再试")
}

// ValidateError 参数验证错误
func ValidateError(c *gin.Context, err error) {
	InvalidParams(c, err.Error())
}
