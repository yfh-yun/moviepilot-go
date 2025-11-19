package tvdb

import (
	"fmt"
	"net/http"
)

// TVDBError TVDB API错误类型
type TVDBError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Method  string `json:"method"`
	URL     string `json:"url"`
}

// Error 实现error接口
func (e *TVDBError) Error() string {
	return fmt.Sprintf("TVDB API错误 [%d]: %s - %s %s", e.Code, e.Message, e.Method, e.URL)
}

// 预定义错误
var (
	// HTTP错误
	ErrUnauthorized    = &TVDBError{Code: 401, Message: "认证失败，请检查API密钥和PIN"}
	ErrForbidden       = &TVDBError{Code: 403, Message: "权限不足"}
	ErrNotFound        = &TVDBError{Code: 404, Message: "资源不存在"}
	ErrTooManyRequests = &TVDBError{Code: 429, Message: "请求过于频繁，请稍后重试"}
	ErrInternalServer  = &TVDBError{Code: 500, Message: "TVDB服务器内部错误"}

	// 业务错误
	ErrInvalidAPIKey   = &TVDBError{Code: 1001, Message: "无效的API密钥"}
	ErrInvalidPIN      = &TVDBError{Code: 1002, Message: "无效的PIN码"}
	ErrTokenExpired    = &TVDBError{Code: 1003, Message: "令牌已过期"}
	ErrTokenInvalid    = &TVDBError{Code: 1004, Message: "无效的令牌"}
	ErrSeriesNotFound  = &TVDBError{Code: 1005, Message: "剧集不存在"}
	ErrEpisodeNotFound = &TVDBError{Code: 1006, Message: "剧集不存在"}
	ErrSearchFailed    = &TVDBError{Code: 1007, Message: "搜索失败"}

	// 网络错误
	ErrNetworkTimeout = &TVDBError{Code: 2001, Message: "网络请求超时"}
	ErrNetworkError   = &TVDBError{Code: 2002, Message: "网络连接错误"}

	// 解析错误
	ErrJSONParse       = &TVDBError{Code: 3001, Message: "JSON解析失败"}
	ErrInvalidResponse = &TVDBError{Code: 3002, Message: "无效的响应格式"}
)

// HandleHTTPError 处理HTTP状态码错误
func HandleHTTPError(statusCode int) error {
	switch statusCode {
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusForbidden:
		return ErrForbidden
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusTooManyRequests:
		return ErrTooManyRequests
	case http.StatusInternalServerError:
		return ErrInternalServer
	default:
		return &TVDBError{
			Code:    statusCode,
			Message: fmt.Sprintf("HTTP错误 %d", statusCode),
		}
	}
}

// IsNotFoundError 检查是否为"未找到"错误
func IsNotFoundError(err error) bool {
	tvdbErr, ok := err.(*TVDBError)
	if !ok {
		return false
	}
	return tvdbErr.Code == 404 || tvdbErr.Code == 1005 || tvdbErr.Code == 1006
}

// IsAuthError 检查是否为认证错误
func IsAuthError(err error) bool {
	tvdbErr, ok := err.(*TVDBError)
	if !ok {
		return false
	}
	return tvdbErr.Code == 401 || tvdbErr.Code == 1001 || tvdbErr.Code == 1002 ||
		tvdbErr.Code == 1003 || tvdbErr.Code == 1004
}

// IsRetryableError 检查是否为可重试错误
func IsRetryableError(err error) bool {
	tvdbErr, ok := err.(*TVDBError)
	if !ok {
		// 非TVDB错误默认不可重试
		return false
	}

	// 网络错误和服务器错误可以重试
	switch tvdbErr.Code {
	case 429, 500, 2001, 2002:
		return true
	default:
		return false
	}
}

// NewError 创建新的TVDB错误
func NewError(code int, message, method, url string) error {
	return &TVDBError{
		Code:    code,
		Message: message,
		Method:  method,
		URL:     url,
	}
}

// WrapError 包装错误并添加上下文
func WrapError(err error, method, url string) error {
	if tvdbErr, ok := err.(*TVDBError); ok {
		tvdbErr.Method = method
		tvdbErr.URL = url
		return tvdbErr
	}

	return &TVDBError{
		Code:    9999, // 未知错误
		Message: err.Error(),
		Method:  method,
		URL:     url,
	}
}
