// Package utils 提供通用工具函数
package utils

import (
	"fmt"
	"reflect"
	"time"
)

// Retry 重试装饰�?func Retry(maxRetries int, exceptions []reflect.Type, defaultReturn interface{}, callback func() (interface{}, error)) interface{} {
	for i := 0; i < maxRetries; i++ {
		result, err := callback()
		if err == nil {
			return result
		}

		// 检查异常类型是否在指定列表�?		shouldRetry := false
		errType := reflect.TypeOf(err)
		for _, exceptionType := range exceptions {
			if errType == exceptionType {
				shouldRetry = true
				break
			}
		}

		// 如果是立即异常或不需要重试的异常，则直接返回默认�?		if !shouldRetry {
			return defaultReturn
		}

		// 等待一段时间后重试
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	return defaultReturn
}

// RetryWithResult 重试装饰器（带回调函数处理结果）
func RetryWithResult(maxRetries int, exceptions []reflect.Type, defaultReturn interface{}, callback func() (interface{}, error), resultCallback func(interface{}) interface{}) interface{} {
	for i := 0; i < maxRetries; i++ {
		result, err := callback()
		if err == nil {
			return resultCallback(result)
		}

		// 检查异常类型是否在指定列表�?		shouldRetry := false
		errType := reflect.TypeOf(err)
		for _, exceptionType := range exceptions {
			if errType == exceptionType {
				shouldRetry = true
				break
			}
		}

		// 如果是立即异常或不需要重试的异常，则直接返回默认�?		if !shouldRetry {
			return defaultReturn
		}

		// 等待一段时间后重试
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	return defaultReturn
}

// ExecuteWithTimeout 执行带超时的操作
func ExecuteWithTimeout(timeout time.Duration, callback func() (interface{}, error)) (interface{}, error) {
	resultChan := make(chan interface{}, 1)
	errChan := make(chan error, 1)

	go func() {
		result, err := callback()
		if err != nil {
			errChan <- err
		} else {
			resultChan <- result
		}
	}()

	select {
	case result := <-resultChan:
		return result, nil
	case err := <-errChan:
		return nil, err
	case <-time.After(timeout):
		return nil, fmt.Errorf("操作超时")
	}
}
