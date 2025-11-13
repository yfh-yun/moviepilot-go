// Package main 提供工具函数使用示例
package main

import (
	"errors"
	"fmt"
	"reflect"
	"time"
	
	"moviepilot-go/internal/utils"
)

// CustomError 自定义错误类�?type CustomError struct {
	message string
}

func (e *CustomError) Error() string {
	return e.message
}

func main() {
	// 示例1: 使用重试功能
	fmt.Println("=== 重试功能示例 ===")
	
	// 定义一个可能失败的函数
	attempt := 0
	failingFunc := func() error {
		attempt++
		if attempt < 3 {
			return &CustomError{message: "模拟自定义错�?}
		}
		fmt.Println("函数执行成功")
		return nil
	}
	
	// 使用重试机制执行函数，只对CustomError类型进行重试
	err := utils.Retry(failingFunc, reflect.TypeOf(&CustomError{}), 5, time.Second, 2, func(msg string) {
		fmt.Println("日志:", msg)
	})
	
	if err != nil {
		fmt.Printf("重试后仍然失�? %v\n", err)
	}
	
	// 示例2: 使用执行时间记录功能
	fmt.Println("\n=== 执行时间记录示例 ===")
	
	// 定义一个耗时的函�?	timedFunc := func() (interface{}, error) {
		time.Sleep(2 * time.Second)
		return "任务完成", nil
	}
	
	result, err := utils.LogExecutionTime(timedFunc, "耗时任务", func(msg string) {
		fmt.Println("执行时间日志:", msg)
	})
	
	if err != nil {
		fmt.Printf("执行出错: %v\n", err)
	} else {
		fmt.Printf("执行结果: %v\n", result)
	}
	
	// 示例3: 使用立即异常（不会被重试�?	fmt.Println("\n=== 立即异常示例 ===")
	
	immediateErrFunc := func() error {
		return utils.NewImmediateException("这是立即异常，不会被重试")
	}
	
	err = utils.Retry(immediateErrFunc, nil, 3, time.Second, 2, func(msg string) {
		fmt.Println("日志:", msg)
	})
	
	if err != nil {
		fmt.Printf("立即异常: %v\n", err)
	}
	
	// 示例4: 异步函数执行时间记录
	fmt.Println("\n=== 异步函数执行时间记录示例 ===")
	
	asyncTimedFunc := func() (interface{}, error) {
		time.Sleep(1 * time.Second)
		return "异步任务完成", nil
	}
	
	result, err = utils.LogAsyncExecutionTime(asyncTimedFunc, "异步耗时任务", func(msg string) {
		fmt.Println("异步执行时间日志:", msg)
	})
	
	if err != nil {
		fmt.Printf("异步执行出错: %v\n", err)
	} else {
		fmt.Printf("异步执行结果: %v\n", result)
	}
}
