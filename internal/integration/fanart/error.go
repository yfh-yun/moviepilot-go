package fanart

import (
	"errors"
	"fmt"
	"net/http"
)

// Error Fanart API错误
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Error 实现error接口
func (e *Error) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("Fanart API错误 %d: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("Fanart API错误 %d: %s", e.Code, e.Message)
}

// 预定义错误
var (
	ErrInvalidAPIKey     = errors.New("无效的API密钥")
	ErrRateLimitExceeded = errors.New("API调用频率限制")
	ErrNotFound          = errors.New("资源未找到")
	ErrServerError       = errors.New("服务器内部错误")
	ErrNetworkError      = errors.New("网络连接错误")
	ErrInvalidResponse   = errors.New("无效的API响应")
	ErrInvalidParameters = errors.New("无效的参数")
)

// NewError 创建新的错误
func NewError(code int, message string, details ...string) error {
	err := &Error{
		Code:    code,
		Message: message,
	}

	if len(details) > 0 {
		err.Details = details[0]
	}

	return err
}

// WrapError 包装错误
func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}

	if fanartErr, ok := err.(*Error); ok {
		return &Error{
			Code:    fanartErr.Code,
			Message: fmt.Sprintf("%s: %s", message, fanartErr.Message),
			Details: fanartErr.Details,
		}
	}

	return &Error{
		Code:    500,
		Message: fmt.Sprintf("%s: %v", message, err),
	}
}

// HandleHTTPError 处理HTTP错误
func HandleHTTPError(statusCode int) error {
	switch statusCode {
	case http.StatusOK:
		return nil
	case http.StatusUnauthorized:
		return ErrInvalidAPIKey
	case http.StatusForbidden:
		return ErrRateLimitExceeded
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusTooManyRequests:
		return ErrRateLimitExceeded
	case http.StatusInternalServerError:
		return ErrServerError
	case http.StatusBadRequest:
		return ErrInvalidParameters
	default:
		return fmt.Errorf("HTTP错误 %d", statusCode)
	}
}

// IsRetryableError 检查错误是否可重试
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	switch err {
	case ErrRateLimitExceeded, ErrServerError, ErrNetworkError:
		return true
	}

	if fanartErr, ok := err.(*Error); ok {
		return fanartErr.Code >= 500 // 服务器错误可重试
	}

	return false
}

// ShouldCacheError 检查错误是否应该缓存
func ShouldCacheError(err error) bool {
	if err == nil {
		return true
	}

	// 不缓存以下错误
	switch err {
	case ErrInvalidAPIKey, ErrInvalidParameters:
		return false
	}

	// 客户端错误不缓存
	if fanartErr, ok := err.(*Error); ok {
		return fanartErr.Code < 400 || fanartErr.Code >= 500
	}

	return true
}

// GetErrorCode 获取错误代码
func GetErrorCode(err error) int {
	if err == nil {
		return 0
	}

	if fanartErr, ok := err.(*Error); ok {
		return fanartErr.Code
	}

	return 500
}

// IsNotFoundError 检查是否为"未找到"错误
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	return errors.Is(err, ErrNotFound)
}

// IsRateLimitError 检查是否为频率限制错误
func IsRateLimitError(err error) bool {
	if err == nil {
		return false
	}

	return errors.Is(err, ErrRateLimitExceeded)
}

// WaitForRetry 根据错误类型计算重试等待时间
func WaitForRetry(err error, attempt int) time.Duration {
	if IsRateLimitError(err) {
		// 频率限制等待1分钟
		return time.Minute
	}

	// 指数退避策略
	baseDelay := time.Second * 2
	maxDelay := time.Minute * 5

	delay := baseDelay * time.Duration(1<<uint(attempt-1))
	if delay > maxDelay {
		delay = maxDelay
	}

	return delay
}
