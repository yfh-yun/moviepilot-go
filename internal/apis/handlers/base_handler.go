// Package handlers MoviePilot API处理器基础模块
// 提供统一的基础处理器和错误处理机制
package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// BaseHandler 基础API处理器
// 提供统一的错误处理、参数验证和响应格式
type BaseHandler struct {
	logger *zap.Logger
}

// NewBaseHandler 创建基础处理器
func NewBaseHandler() *BaseHandler {
	return &BaseHandler{
		logger: logger.Logger,
	}
}

// 统一的响应结构
type APIResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Code      int         `json:"code,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// 成功响应
func (h *BaseHandler) Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Success:   true,
		Data:      data,
		Timestamp: time.Now().Unix(),
	})
}

// 带消息的成功响应
func (h *BaseHandler) SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
	})
}

// 错误响应
func (h *BaseHandler) Error(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, APIResponse{
		Success:   false,
		Message:   message,
		Code:      code,
		Timestamp: time.Now().Unix(),
	})

	h.logger.Warn("API Error",
		zap.Int("code", code),
		zap.String("message", message),
		zap.String("path", c.FullPath()),
		zap.String("client_ip", c.ClientIP()),
	)
}

// 服务器错误响应
func (h *BaseHandler) ServerError(c *gin.Context, err error) {
	h.logger.Error("Internal Server Error",
		zap.Error(err),
		zap.String("path", c.FullPath()),
		zap.String("client_ip", c.ClientIP()),
	)

	c.JSON(http.StatusInternalServerError, APIResponse{
		Success:   false,
		Message:   "服务器内部错误",
		Code:      response.CodeServerError,
		Timestamp: time.Now().Unix(),
	})
}

// 参数验证错误
func (h *BaseHandler) ValidationError(c *gin.Context, err error) {
	h.logger.Warn("Validation Error",
		zap.Error(err),
		zap.String("path", c.FullPath()),
	)

	c.JSON(http.StatusBadRequest, APIResponse{
		Success:   false,
		Message:   err.Error(),
		Code:      response.CodeInvalidParams,
		Timestamp: time.Now().Unix(),
	})
}

// 认证错误
func (h *BaseHandler) Unauthorized(c *gin.Context, message string) {
	h.logger.Warn("Unauthorized",
		zap.String("message", message),
		zap.String("path", c.FullPath()),
		zap.String("client_ip", c.ClientIP()),
	)

	c.JSON(http.StatusUnauthorized, APIResponse{
		Success:   false,
		Message:   message,
		Code:      response.CodeUnauthorized,
		Timestamp: time.Now().Unix(),
	})
}

// 资源不存在错误
func (h *BaseHandler) NotFound(c *gin.Context, resource string) {
	h.logger.Warn("Resource Not Found",
		zap.String("resource", resource),
		zap.String("path", c.FullPath()),
	)

	c.JSON(http.StatusNotFound, APIResponse{
		Success:   false,
		Message:   resource + "不存在",
		Code:      response.CodeNotFound,
		Timestamp: time.Now().Unix(),
	})
}

// 权限不足错误
func (h *BaseHandler) Forbidden(c *gin.Context, message string) {
	h.logger.Warn("Forbidden",
		zap.String("message", message),
		zap.String("path", c.FullPath()),
		zap.String("client_ip", c.ClientIP()),
	)

	c.JSON(http.StatusForbidden, APIResponse{
		Success:   false,
		Message:   message,
		Code:      response.CodeForbidden,
		Timestamp: time.Now().Unix(),
	})
}

// 获取分页参数
func (h *BaseHandler) GetPaginationParams(c *gin.Context) (page, pageSize int) {
	page = h.getIntParam(c, "page", 1)
	pageSize = h.getIntParam(c, "page_size", 20)

	// 限制分页大小
	if pageSize > 100 {
		pageSize = 100
	}
	if pageSize < 1 {
		pageSize = 1
	}
	if page < 1 {
		page = 1
	}

	return page, pageSize
}

// 获取排序参数
func (h *BaseHandler) GetSortParams(c *gin.Context) (sortBy, sortOrder string) {
	sortBy = c.DefaultQuery("sort_by", "created_at")
	sortOrder = c.DefaultQuery("sort_order", "desc")

	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	return sortBy, sortOrder
}

// 辅助方法：获取整数参数
func (h *BaseHandler) getIntParam(c *gin.Context, key string, defaultValue int) int {
	value := c.DefaultQuery(key, "")
	if value == "" {
		return defaultValue
	}

	var result int
	_, err := fmt.Sscanf(value, "%d", &result)
	if err != nil {
		return defaultValue
	}

	return result
}

// 以下为兼容性方法，保持与原有代码的兼容性

// Login 登录处理器（兼容性实现）
func (h *BaseHandler) Login(c *gin.Context) {
	// 获取表单数据
	username := c.PostForm("username")
	password := c.PostForm("password")
	_ = c.PostForm("otp_password")

	// 验证输入参数
	if username == "" || password == "" {
		h.Error(c, response.CodeInvalidParams, "用户名和密码不能为空")
		return
	}

	h.logger.Info("User login attempt",
		zap.String("username", username),
		zap.String("ip", c.ClientIP()),
	)

	// 临时返回模拟数据
	h.Success(c, gin.H{
		"access_token": "mock_token",
		"token_type":   "bearer",
		"super_user":   true,
		"user_id":      1,
		"user_name":    username,
		"avatar":       "",
		"level":        1,
		"permissions":  gin.H{},
		"widzard":      false,
	})
}

// Logout 登出处理器（兼容性实现）
func (h *BaseHandler) Logout(c *gin.Context) {
	h.logger.Info("User logout",
		zap.String("ip", c.ClientIP()),
	)

	h.SuccessWithMessage(c, "登出成功", nil)
}

// GetSystemInfo 获取系统信息（兼容性实现）
func (h *BaseHandler) GetSystemInfo(c *gin.Context) {
	h.Success(c, gin.H{
		"name":       "MoviePilot",
		"version":    "2.8.1",
		"go_version": "go1.24.4+",
		"build_time": time.Now().Format("2006-01-02 15:04:05"),
		"git_commit": "unknown",
	})
}

// HealthCheck 健康检查（兼容性实现）
func (h *BaseHandler) HealthCheck(c *gin.Context) {
	h.Success(c, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
		"version":   "2.8.1",
	})
}

// GetVersion 获取版本信息（兼容性实现）
func (h *BaseHandler) GetVersion(c *gin.Context) {
	h.Success(c, gin.H{
		"version": "2.8.1",
		"build":   "2023120101",
	})
}
