package models

// ImmediateError 用于立即返回错误而不重试的特殊错误类型�?// 当不希望使用重试机制时，可以返回此错误�?type ImmediateError struct {
	Message string
}

func (e *ImmediateError) Error() string {
	return e.Message
}

// LimitError 用于表示本地限流器或外部触发的限流错误的基类�?// 该错误类型可用于本地限流逻辑或外部限流处理�?type LimitError struct {
	Message string
}

func (e *LimitError) Error() string {
	return e.Message
}

// APIRateLimitError 用于表示API速率限制的错误类型�?// 当API调用触发速率限制时，可以返回此错误以立即终止操作并报告错误�?type APIRateLimitError struct {
	Message string
}

func (e *APIRateLimitError) Error() string {
	return e.Message
}

// RateLimitExceededError 用于表示本地限流器触发的错误类型�?// 当函数调用频率超过限流器的限制时，可以返回此错误以停止当前操作并告知调用者限流情况�?// 这个错误通常用于本地限流逻辑，当系统检测到函数调用频率过高时，触发限流并返回该错误�?type RateLimitExceededError struct {
	Message string
}

func (e *RateLimitExceededError) Error() string {
	return e.Message
}

// NewImmediateError 创建一个新�?ImmediateError 实例
func NewImmediateError(message string) *ImmediateError {
	return &ImmediateError{Message: message}
}

// NewLimitError 创建一个新�?LimitError 实例
func NewLimitError(message string) *LimitError {
	return &LimitError{Message: message}
}

// NewAPIRateLimitError 创建一个新�?APIRateLimitError 实例
func NewAPIRateLimitError(message string) *APIRateLimitError {
	return &APIRateLimitError{Message: message}
}

// NewRateLimitExceededError 创建一个新�?RateLimitExceededError 实例
func NewRateLimitExceededError(message string) *RateLimitExceededError {
	return &RateLimitExceededError{Message: message}
}

// IsImmediateError 检查错误是否为 ImmediateError 类型
func IsImmediateError(err error) bool {
	_, ok := err.(*ImmediateError)
	return ok
}

// IsLimitError 检查错误是否为 LimitError 类型
func IsLimitError(err error) bool {
	_, ok := err.(*LimitError)
	return ok
}

// IsAPIRateLimitError 检查错误是否为 APIRateLimitError 类型
func IsAPIRateLimitError(err error) bool {
	_, ok := err.(*APIRateLimitError)
	return ok
}

// IsRateLimitExceededError 检查错误是否为 RateLimitExceededError 类型
func IsRateLimitExceededError(err error) bool {
	_, ok := err.(*RateLimitExceededError)
	return ok
}

// 预定义的错误实例
var (
	ErrImmediate = errors.New("immediate error")
	ErrLimit     = errors.New("limit error")
	ErrAPIRateLimit = errors.New("API rate limit exceeded")
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
)
