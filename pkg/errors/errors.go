package errors

import (
	"fmt"
	"net/http"
)

// AppError 应用错误类型
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Error 实现error接口
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Details)
	}
	return e.Message
}

// StatusCode 返回HTTP状态码
func (e *AppError) StatusCode() int {
	return e.Code
}

// NewAppError 创建新的应用错误
func NewAppError(code int, message, details string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// 预定义常用错误
var (
	ErrBadRequest     = NewAppError(http.StatusBadRequest, "Bad Request", "")
	ErrUnauthorized   = NewAppError(http.StatusUnauthorized, "Unauthorized", "")
	ErrForbidden      = NewAppError(http.StatusForbidden, "Forbidden", "")
	ErrNotFound       = NewAppError(http.StatusNotFound, "Not Found", "")
	ErrConflict       = NewAppError(http.StatusConflict, "Conflict", "")
	ErrInternalServer = NewAppError(http.StatusInternalServerError, "Internal Server Error", "")
)

// WrapError 包装错误
func WrapError(err error, message string) *AppError {
	if err == nil {
		return nil
	}
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	return NewAppError(http.StatusInternalServerError, message, err.Error())
}
