package main

import (
	"fmt"
	"time"

	"moviepilot-go/internal/utils"
)

func main() {
	fmt.Println("=== 限流器示�?===")

	// 示例1: 指数退避限流器
	exampleExponentialBackoff()

	// 示例2: 时间窗口限流�?	exampleWindowRateLimiter()

	// 示例3: 组合限流�?	exampleCompositeRateLimiter()
}

// 模拟一个需要限流的函数
func someFunction(id int) (string, error) {
	return fmt.Sprintf("执行函数 %d 成功", id), nil
}

// 模拟一个触发限流异常的函数
func failingFunction(id int) (string, error) {
	return "", &utils.LimitException{Message: "模拟触发限流异常"}
}

func exampleExponentialBackoff() {
	fmt.Println("\n--- 指数退避限流器示例 ---")
	
	// 创建指数退避限流器 (基础等待时间1秒，最大等待时�?�?
	limiter := utils.NewExponentialBackoffRateLimiter(1.0, 5.0, 2.0, "指数退避测�?, true)
	handler := utils.NewRateLimitHandler(limiter, false)

	// 连续调用5次函�?	for i := 1; i <= 5; i++ {
		fmt.Printf("�?%d 次调�? ", i)
		
		// 检查是否可以调�?		canCall, message := limiter.CanCall()
		if !canCall {
			fmt.Printf("被限�? %s\n", message)
		} else {
			// 模拟函数调用
			result, err := handler.Handle(someFunction, i)
			if err != nil {
				fmt.Printf("调用失败: %v\n", err)
			} else if result == nil {
				fmt.Println("调用被跳�?)
			} else {
				fmt.Printf("调用成功: %s\n", result.(string))
			}
		}
		
		time.Sleep(500 * time.Millisecond)
	}

	// 模拟触发限流异常
	fmt.Println("\n模拟触发限流异常:")
	_, _ = handler.Handle(failingFunction, 999)
	
	// 再次尝试调用
	fmt.Println("触发异常后再次尝试调�?")
	canCall, message := limiter.CanCall()
	if !canCall {
		fmt.Printf("被限�? %s\n", message)
		time.Sleep(1 * time.Second) // 等待一段时�?	}
	
	// 重置限流�?	limiter.Reset()
	fmt.Println("已重置限流器")
	
	canCall, _ = limiter.CanCall()
	if canCall {
		fmt.Println("重置后可以正常调�?)
	}
}

func exampleWindowRateLimiter() {
	fmt.Println("\n--- 时间窗口限流器示�?---")
	
	// 创建时间窗口限流�?(5秒内最多调�?�?
	limiter := utils.NewWindowRateLimiter(3, 5.0, "时间窗口测试", true)
	handler := utils.NewRateLimitHandler(limiter, true)

	// 快速连续调�?次函�?	for i := 1; i <= 5; i++ {
		fmt.Printf("�?%d 次调�? ", i)
		
		result, err := handler.Handle(someFunction, i)
		if err != nil {
			fmt.Printf("调用失败: %v\n", err)
		} else if result == nil {
			fmt.Println("调用被跳�?)
		} else {
			fmt.Printf("调用成功: %s\n", result.(string))
		}
		
		time.Sleep(100 * time.Millisecond)
	}

	// 等待窗口期过后再调用
	fmt.Println("\n等待窗口期过�?..")
	time.Sleep(6 * time.Second)
	
	result, err := handler.Handle(someFunction, 6)
	if err != nil {
		fmt.Printf("调用失败: %v\n", err)
	} else if result == nil {
		fmt.Println("调用被跳�?)
	} else {
		fmt.Printf("窗口期过后调用成�? %s\n", result.(string))
	}
}

func exampleCompositeRateLimiter() {
	fmt.Println("\n--- 组合限流器示�?---")
	
	// 创建两个不同的限流器
	exponentialLimiter := utils.NewExponentialBackoffRateLimiter(1.0, 3.0, 2.0, "指数退�?, true)
	windowLimiter := utils.NewWindowRateLimiter(2, 5.0, "时间窗口", true)
	
	// 创建组合限流�?	compositeLimiter := utils.NewCompositeRateLimiter([]utils.BaseRateLimiter{exponentialLimiter, windowLimiter}, "组合测试", true)
	handler := utils.NewRateLimitHandler(compositeLimiter, false)

	// 连续调用几次函数
	for i := 1; i <= 4; i++ {
		fmt.Printf("�?%d 次调�? ", i)
		
		result, err := handler.Handle(someFunction, i)
		if err != nil {
			fmt.Printf("调用失败: %v\n", err)
		} else if result == nil {
			fmt.Println("调用被跳�?)
		} else {
			fmt.Printf("调用成功: %s\n", result.(string))
		}
		
		time.Sleep(500 * time.Millisecond)
	}
}
