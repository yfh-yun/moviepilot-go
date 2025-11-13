// Package utils 提供通用工具函数
package utils

import "errors"

// ImmediateException 立即异常处理
type ImmediateException struct {
	Message string
}

// NewImmediateException 创建一个新�?ImmediateException 实例
func NewImmediateException(message string) *ImmediateException {
	return &ImmediateException{Message: message}
}

// Error 实现 error 接口
func (e *ImmediateException) Error() string {
	return e.Message
}

// IsImmediate 检查是否为立即异常
func (e *ImmediateException) IsImmediate() bool {
	return true
}

// IsImmediateError 检查错误是否为立即异常
func IsImmediateError(err error) bool {
	if _, ok := err.(*ImmediateException); ok {
		return true
	}
	return false
}

// ToImmediateException 将普通错误转换为立即异常
func ToImmediateException(message string) error {
	return errors.New(message)
}
