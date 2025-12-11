package common

import "errors"

// ImmediateException 用于立即抛出异常而不重试的特殊异常类
// 当不希望使用重试机制时，可以返回此错误
type ImmediateException struct {
	Message string
}

func (e *ImmediateException) Error() string {
	return e.Message
}

// NewImmediateException 创建立即异常
func NewImmediateException(message string) error {
	return &ImmediateException{Message: message}
}

// IsImmediateException 判断是否为立即异常
func IsImmediateException(err error) bool {
	var ie *ImmediateException
	return errors.As(err, &ie)
}

// LimitException 用于表示本地限流器或外部触发的限流异常的基类
// 该异常类可用于本地限流逻辑或外部限流处理
type LimitException struct {
	ImmediateException
}

// NewLimitException 创建限流异常
func NewLimitException(message string) error {
	return &LimitException{
		ImmediateException: ImmediateException{Message: message},
	}
}

// IsLimitException 判断是否为限流异常
func IsLimitException(err error) bool {
	var le *LimitException
	return errors.As(err, &le)
}

// APIRateLimitException 用于表示API速率限制的异常类
// 当API调用触发速率限制时，可以返回此错误以立即终止操作并报告错误
type APIRateLimitException struct {
	LimitException
}

// NewAPIRateLimitException 创建API速率限制异常
func NewAPIRateLimitException(message string) error {
	return &APIRateLimitException{
		LimitException: LimitException{
			ImmediateException: ImmediateException{Message: message},
		},
	}
}

// IsAPIRateLimitException 判断是否为API速率限制异常
func IsAPIRateLimitException(err error) bool {
	var ale *APIRateLimitException
	return errors.As(err, &ale)
}

// RateLimitExceededException 用于表示本地限流器触发的异常类
// 当函数调用频率超过限流器的限制时，可以返回此错误以停止当前操作并告知调用者限流情况
// 这个异常通常用于本地限流逻辑（例如 RateLimiter），当系统检测到函数调用频率过高时，触发限流并返回该异常
type RateLimitExceededException struct {
	LimitException
}

// NewRateLimitExceededException 创建本地限流异常
func NewRateLimitExceededException(message string) error {
	return &RateLimitExceededException{
		LimitException: LimitException{
			ImmediateException: ImmediateException{Message: message},
		},
	}
}

// IsRateLimitExceededException 判断是否为本地限流异常
func IsRateLimitExceededException(err error) bool {
	var rle *RateLimitExceededException
	return errors.As(err, &rle)
}
