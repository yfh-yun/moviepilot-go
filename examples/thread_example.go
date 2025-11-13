package main

import (
	"fmt"
	"time"

	"moviepilot-go/internal/helper"
)

func main() {
	fmt.Println("Thread Helper Example")
	
	// 创建线程帮助类实�?	threadHelper := helper.NewThreadHelper()
	
	if threadHelper == nil {
		fmt.Println("Failed to create ThreadHelper")
		return
	}
	
	fmt.Printf("ThreadHelper created successfully with %d workers\n", threadHelper.maxWorkers)
	
	// 测试提交任务
	fmt.Println("\n=== 提交单个任务 ===")
	taskFunc := func() (interface{}, error) {
		time.Sleep(500 * time.Millisecond) // 模拟耗时操作
		return "Hello from task!", nil
	}
	
	resultChan := threadHelper.Submit(taskFunc)
	
	// 获取任务结果
	result := <-resultChan
	if result.Error != nil {
		fmt.Printf("Task failed: %v\n", result.Error)
	} else {
		fmt.Printf("Task result: %v\n", result.Value)
	}
	
	// 测试提交多个任务
	fmt.Println("\n=== 提交多个任务 ===")
	tasks := make([]chan helper.TaskResult, 0)
	
	for i := 0; i < 5; i++ {
		taskID := i
		taskFunc := func() (interface{}, error) {
			// 模拟不同耗时的操�?			time.Sleep(time.Duration((taskID+1)*200) * time.Millisecond)
			return fmt.Sprintf("Result from task %d", taskID), nil
		}
		
		resultChan := threadHelper.Submit(taskFunc)
		tasks = append(tasks, resultChan)
	}
	
	// 收集所有任务结�?	for i, resultChan := range tasks {
		result := <-resultChan
		if result.Error != nil {
			fmt.Printf("Task %d failed: %v\n", i, result.Error)
		} else {
			fmt.Printf("Task %d result: %v\n", i, result.Value)
		}
	}
	
	// 测试提交无返回值的任务
	fmt.Println("\n=== 提交无返回值的任务 ===")
	var counter int
	taskFuncNoReturn := func() {
		time.Sleep(300 * time.Millisecond)
		counter++
		fmt.Printf("Task executed, counter: %d\n", counter)
	}
	
	threadHelper.SubmitFunc(taskFuncNoReturn)
	threadHelper.SubmitFunc(taskFuncNoReturn)
	threadHelper.SubmitFunc(taskFuncNoReturn)
	
	// 等待任务执行完成
	time.Sleep(1 * time.Second)
	fmt.Printf("Final counter value: %d\n", counter)
	
	// 测试并发执行任务
	fmt.Println("\n=== 并发执行任务 ===")
	concurrentTasks := make([]chan helper.TaskResult, 0)
	
	// 提交10个并发任�?	for i := 0; i < 10; i++ {
		taskID := i
		taskFunc := func() (interface{}, error) {
			// 模拟随机耗时操作
			time.Sleep(time.Duration(taskID*50) * time.Millisecond)
			return fmt.Sprintf("Concurrent task %d completed", taskID), nil
		}
		
		resultChan := threadHelper.Submit(taskFunc)
		concurrentTasks = append(concurrentTasks, resultChan)
	}
	
	// 收集并发任务结果
	for i, resultChan := range concurrentTasks {
		result := <-resultChan
		if result.Error != nil {
			fmt.Printf("Concurrent task %d failed: %v\n", i, result.Error)
		} else {
			fmt.Printf("Concurrent task %d result: %v\n", i, result.Value)
		}
	}
	
	// 关闭线程�?	fmt.Println("\n=== 关闭线程�?===")
	threadHelper.Shutdown()
	fmt.Println("Thread pool shutdown completed")
	
	fmt.Println("\nExample completed")
}
