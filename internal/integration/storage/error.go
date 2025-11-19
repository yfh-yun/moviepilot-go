package storage

import "fmt"

// Error 存储错误
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Error 实现error接口
func (e *Error) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("存储错误[%s]: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("存储错误[%s]: %s", e.Code, e.Message)
}

// 预定义错误
var (
	ErrNotImplemented   = &Error{Code: "NOT_IMPLEMENTED", Message: "功能未实现"}
	ErrNotConnected     = &Error{Code: "NOT_CONNECTED", Message: "存储未连接"}
	ErrFileNotFound     = &Error{Code: "FILE_NOT_FOUND", Message: "文件未找到"}
	ErrPermissionDenied = &Error{Code: "PERMISSION_DENIED", Message: "权限不足"}
	ErrQuotaExceeded    = &Error{Code: "QUOTA_EXCEEDED", Message: "存储配额不足"}
	ErrNetworkError     = &Error{Code: "NETWORK_ERROR", Message: "网络错误"}
	ErrInvalidPath      = &Error{Code: "INVALID_PATH", Message: "无效路径"}
)

// NewError 创建新的错误
func NewError(code, message string, details ...string) error {
	err := &Error{
		Code:    code,
		Message: message,
	}

	if len(details) > 0 {
		err.Details = details[0]
	}

	return err
}

// IsNotFoundError 检查是否为"未找到"错误
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	if storageErr, ok := err.(*Error); ok {
		return storageErr.Code == "FILE_NOT_FOUND"
	}

	return false
}

// IsPermissionError 检查是否为权限错误
func IsPermissionError(err error) bool {
	if err == nil {
		return false
	}

	if storageErr, ok := err.(*Error); ok {
		return storageErr.Code == "PERMISSION_DENIED"
	}

	return false
}

// IsQuotaError 检查是否为配额错误
func IsQuotaError(err error) bool {
	if err == nil {
		return false
	}

	if storageErr, ok := err.(*Error); ok {
		return storageErr.Code == "QUOTA_EXCEEDED"
	}

	return false
}

// WrapError 包装错误
func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}

	return &Error{
		Code:    "WRAPPED_ERROR",
		Message: fmt.Sprintf("%s: %v", message, err),
	}
}
