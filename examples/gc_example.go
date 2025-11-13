// Package main 提供内存回收工具使用示例
package main

import (
	"fmt"
	"strings"
	
	"moviepilot-go/internal/utils"
)

func main() {
	fmt.Println("=== 内存回收工具使用示例 ===")
	
	gcUtils := &utils.GCUtils{}
	
	// 示例1: 获取当前内存使用情况
	fmt.Println("\n1. 获取当前内存使用情况:")
	memoryUsage := gcUtils.GetMemoryUsage()
	fmt.Printf("当前内存使用: %.2f MB\n", memoryUsage)
	
	// 示例2: 使用MemoryGC装饰�?	fmt.Println("\n2. 使用MemoryGC装饰�?")
	
	// 创建一个消耗内存的函数
	memoryIntensiveFunc := func(args ...interface{}) (interface{}, error) {
		// 创建大量字符串来模拟内存使用
		var data []string
		for i := 0; i < 10000; i++ {
			data = append(data, strings.Repeat("测试数据", 100))
		}
		return len(data), nil
	}
	
	// 应用内存回收装饰�?	decoratedFunc := gcUtils.MemoryGC(true, true)(memoryIntensiveFunc)
	result, err := decoratedFunc()
	if err != nil {
		fmt.Printf("函数执行出错: %v\n", err)
	} else {
		fmt.Printf("函数执行结果: %v\n", result)
	}
	
	// 示例3: 使用AutoGC装饰�?	fmt.Println("\n3. 使用AutoGC装饰�?")
	
	autoGCFunc := func(args ...interface{}) (interface{}, error) {
		// 创建一些数�?		data := make([][]byte, 1000)
		for i := 0; i < 1000; i++ {
			data[i] = make([]byte, 1024) // 1KB数据
		}
		return len(data), nil
	}
	
	autoDecoratedFunc := gcUtils.AutoGC()(autoGCFunc)
	result2, err := autoDecoratedFunc()
	if err != nil {
		fmt.Printf("AutoGC函数执行出错: %v\n", err)
	} else {
		fmt.Printf("AutoGC函数执行结果: %v\n", result2)
	}
	
	// 示例4: 使用MemoryMonitor装饰�?	fmt.Println("\n4. 使用MemoryMonitor装饰�?")
	
	monitorFunc := func(args ...interface{}) (interface{}, error) {
		// 创建一些数�?		data := make(map[int][]byte)
		for i := 0; i < 500; i++ {
			data[i] = make([]byte, 2048) // 2KB数据
		}
		return len(data), nil
	}
	
	// 设置较低的阈值来触发监控
	monitorDecoratedFunc := gcUtils.MemoryWatch(1.0)(monitorFunc) // 1MB阈�?	result3, err := monitorDecoratedFunc()
	if err != nil {
		fmt.Printf("MemoryMonitor函数执行出错: %v\n", err)
	} else {
		fmt.Printf("MemoryMonitor函数执行结果: %v\n", result3)
	}
	
	// 示例5: 再次检查内存使用情�?	fmt.Println("\n5. 再次检查内存使用情�?")
	memoryUsageAfter := gcUtils.GetMemoryUsage()
	fmt.Printf("操作后内存使�? %.2f MB\n", memoryUsageAfter)
	fmt.Printf("内存变化: %.2f MB\n", memoryUsageAfter-memoryUsage)
}
