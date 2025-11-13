// Package utils 提供通用工具函数
package utils

import (
	"fmt"
	"time"
)

// ExecutableFunc 可执行函数类�?type ExecutableFunc func() (interface{}, error)

// AsyncExecutableFunc 可执行的异步函数类型
type AsyncExecutableFunc func() (interface{}, error)

// LogExecutionTime 记录函数执行时间
// fn: 要执行的函数
// name: 函数名称（用于日志记录）
// logger: 日志记录函数
func LogExecutionTime(
	fn ExecutableFunc,
	name string,
	logger func(string),
) (interface{}, error) {
	startTime := time.Now()
	result, err := fn()
	endTime := time.Now()
	
	executionTime := endTime.Sub(startTime)
	msg := fmt.Sprintf("%s execution time: %.2f seconds", name, executionTime.Seconds())
	
	if logger != nil {
		logger(msg)
	} else {
		fmt.Println(msg)
	}
	
	return result, err
}

// LogAsyncExecutionTime 记录异步函数执行时间
// fn: 要执行的异步函数
// name: 函数名称（用于日志记录）
// logger: 日志记录函数
func LogAsyncExecutionTime(
	fn AsyncExecutableFunc,
	name string,
	logger func(string),
) (interface{}, error) {
	startTime := time.Now()
	result, err := fn()
	endTime := time.Now()
	
	executionTime := endTime.Sub(startTime)
	msg := fmt.Sprintf("%s execution time: %.2f seconds", name, executionTime.Seconds())
	
	if logger != nil {
		logger(msg)
	} else {
		fmt.Println(msg)
	}
	
	return result, err
}

// TimedFunc 包装函数以记录执行时�?// name: 函数名称（用于日志记录）
// logger: 日志记录函数
func TimedFunc(name string, logger func(string)) func(ExecutableFunc) ExecutableFunc {
	return func(fn ExecutableFunc) ExecutableFunc {
		return func() (interface{}, error) {
			return LogExecutionTime(fn, name, logger)
		}
	}
}

// TimedAsyncFunc 包装异步函数以记录执行时�?// name: 函数名称（用于日志记录）
// logger: 日志记录函数
func TimedAsyncFunc(name string, logger func(string)) func(AsyncExecutableFunc) AsyncExecutableFunc {
	return func(fn AsyncExecutableFunc) AsyncExecutableFunc {
		return func() (interface{}, error) {
			return LogAsyncExecutionTime(fn, name, logger)
		}
	}
}
