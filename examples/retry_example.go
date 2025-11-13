// Package main 提供重试功能的详细示�?package main

import (
	"fmt"
	"reflect"
	"time"
	
	"moviepilot-go/internal/utils"
)

// NetworkError 模拟网络错误
type NetworkError struct {
	message string
}

func (e *NetworkError) Error() string {
	return e.message
}

// DatabaseError 模拟数据库错�?type DatabaseError struct {
	message string
}

func (e *DatabaseError) Error() string {
	return e.message
}

// TimeoutError 模拟超时错误，实现立即异常接�?type TimeoutError struct {
	message string
}

func (e *TimeoutError) Error() string {
	return e.message
}

func (e *TimeoutError) IsImmediate() bool {
	return true
}

func main() {
	fmt.Println("=== 重试功能完整示例 ===")
	
	// 示例1: 只对特定类型错误重试
	fmt.Println("\n1. 只对NetworkError重试:")
	networkErrCount := 0
	networkFunc := func() error {
		networkErrCount++
		if networkErrCount <= 2 {
			return &NetworkError{message: fmt.Sprintf("网络错误 %d", networkErrCount)}
		}
		return nil
	}
	
	err := utils.Retry(networkFunc, reflect.TypeOf(&NetworkError{}), 5, 500*time.Millisecond, 2, func(msg string) {
		fmt.Printf("   [日志] %s\n", msg)
	})
	
	if err == nil {
		fmt.Println("   [结果] 函数执行成功")
	}
	
	// 示例2: 混合错误类型，只重试特定类型
	fmt.Println("\n2. 混合错误类型，只对DatabaseError重试:")
	mixedErrCount := 0
	mixedFunc := func() error {
		mixedErrCount++
		switch mixedErrCount {
		case 1:
			return &NetworkError{message: "网络错误"}
		case 2:
			return &DatabaseError{message: "数据库错�?}
		default:
			return nil
		}
	}
	
	err = utils.Retry(mixedFunc, reflect.TypeOf(&DatabaseError{}), 5, 500*time.Millisecond, 2, func(msg string) {
		fmt.Printf("   [日志] %s\n", msg)
	})
	
	if err != nil {
		fmt.Printf("   [结果] 函数执行失败: %v\n", err)
	} else {
		fmt.Println("   [结果] 函数执行成功")
	}
	
	// 示例3: 立即异常不会重试
	fmt.Println("\n3. 立即异常不会重试:")
	timeoutCount := 0
	timeoutFunc := func() error {
		timeoutCount++
		return &TimeoutError{message: fmt.Sprintf("超时错误 %d", timeoutCount)}
	}
	
	err = utils.Retry(timeoutFunc, nil, 5, 500*time.Millisecond, 2, func(msg string) {
		fmt.Printf("   [日志] %s\n", msg)
	})
	
	if err != nil {
		fmt.Printf("   [结果] 函数立即失败: %v\n", err)
	}
	
	// 示例4: 指数退避重�?	fmt.Println("\n4. 指数退避重�?")
	expCount := 0
	expFunc := func() error {
		expCount++
		if expCount <= 3 {
			return &NetworkError{message: fmt.Sprintf("网络错误 %d", expCount)}
		}
		return nil
	}
	
	err = utils.ExpBackoffRetry(expFunc, reflect.TypeOf(&NetworkError{}), 5, 200*time.Millisecond, 2, time.Second, func(msg string) {
		fmt.Printf("   [日志] %s\n", msg)
	})
	
	if err == nil {
		fmt.Println("   [结果] 函数执行成功")
	}
}
