package common

import (
	"errors"
	"fmt"
)

// 错误码定义
const (
	// 通用错误码
	ErrCodeSuccess       = 0
	ErrCodeBadRequest    = 400
	ErrCodeUnauthorized  = 401
	ErrCodeForbidden     = 403
	ErrCodeNotFound      = 404
	ErrCodeInternalError = 500

	// 业务错误码
	ErrCodeValidationFailed = 1001
	ErrCodeDatabaseError    = 1002
	ErrCodeServiceError     = 1003
	ErrCodeNotImplemented   = 1004
)

// APIError API错误结构
type APIError struct {
	Code    int
	Message string
	Err     error
}

// Error 实现error接口
func (e *APIError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap 支持errors.Unwrap
func (e *APIError) Unwrap() error {
	return e.Err
}

// NewAPIError 创建API错误
func NewAPIError(code int, message string, err error) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// 预定义错误
var (
	ErrBadRequest     = errors.New("bad request")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrForbidden      = errors.New("forbidden")
	ErrNotFound       = errors.New("not found")
	ErrInternalError  = errors.New("internal server error")
	ErrValidation     = errors.New("validation failed")
	ErrDatabase       = errors.New("database error")
	ErrService        = errors.New("service error")
	ErrNotImplemented = errors.New("not implemented")
)

// WrapError 包装错误
func WrapError(code int, message string, err error) *APIError {
	return NewAPIError(code, message, err)
}
