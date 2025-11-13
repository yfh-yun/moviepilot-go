package helper

import (
	"fmt"
	"testing"
	"time"
)

func TestNewThreadHelper(t *testing.T) {
	// 测试创建ThreadHelper实例
	helper := NewThreadHelper()
	if helper == nil {
		t.Error("Failed to create ThreadHelper instance")
	}
	
	// 验证最大工作协程数是否正确设置
	if helper.maxWorkers <= 0 {
		t.Error("maxWorkers should be greater than 0")
	}
	
	// 验证任务队列是否正确创建
	if helper.taskQueue == nil {
		t.Error("taskQueue should not be nil")
	}
}

func TestSubmit(t *testing.T) {
	// 测试提交任务
	helper := NewThreadHelper()
	
	// 创建一个简单的任务函数
	taskFunc := func() (interface{}, error) {
		time.Sleep(100 * time.Millisecond) // 模拟一些工�?		return "task result", nil
	}
	
	// 提交任务
	resultChan := helper.Submit(taskFunc)
	
	// 等待结果
	select {
	case result := <-resultChan:
		if result.Error != nil {
			t.Errorf("Task execution failed: %v", result.Error)
		}
		if result.Value != "task result" {
			t.Errorf("Task result mismatch: expected 'task result', got %v", result.Value)
		}
	case <-time.After(5 * time.Second):
		t.Error("Task execution timeout")
	}
}

func TestSubmitFunc(t *testing.T) {
	// 测试提交无返回值的任务
	helper := NewThreadHelper()
	
	// 创建一个简单的任务函数
	var executed bool
	taskFunc := func() {
		time.Sleep(100 * time.Millisecond) // 模拟一些工�?		executed = true
	}
	
	// 提交任务
	helper.SubmitFunc(taskFunc)
	
	// 等待任务执行完成
	time.Sleep(200 * time.Millisecond)
	
	// 验证任务是否执行
	if !executed {
		t.Error("Task was not executed")
	}
}

func TestMultipleTasks(t *testing.T) {
	// 测试提交多个任务
	helper := NewThreadHelper()
	
	// 提交多个任务
	results := make([]chan TaskResult, 0)
	for i := 0; i < 5; i++ {
		taskID := i
		taskFunc := func() (interface{}, error) {
			time.Sleep(time.Duration(taskID*100) * time.Millisecond) // 模拟不同耗时的工�?			return fmt.Sprintf("task %d result", taskID), nil
		}
		
		resultChan := helper.Submit(taskFunc)
		results = append(results, resultChan)
	}
	
	// 收集所有结�?	for i, resultChan := range results {
		select {
		case result := <-resultChan:
			if result.Error != nil {
				t.Errorf("Task %d execution failed: %v", i, result.Error)
			}
			expected := fmt.Sprintf("task %d result", i)
			if result.Value != expected {
				t.Errorf("Task %d result mismatch: expected %s, got %v", i, expected, result.Value)
			}
		case <-time.After(5 * time.Second):
			t.Errorf("Task %d execution timeout", i)
		}
	}
}

func TestShutdown(t *testing.T) {
	// 测试关闭线程�?	helper := NewThreadHelper()
	
	// 提交一些任�?	for i := 0; i < 3; i++ {
		taskFunc := func() (interface{}, error) {
			time.Sleep(100 * time.Millisecond)
			return "result", nil
		}
		
		helper.Submit(taskFunc)
	}
	
	// 关闭线程�?	helper.Shutdown()
	
	// 验证线程池已关闭
	// 注意：由于Shutdown是异步的，这里只是简单验证方法能被调�?	t.Log("Thread pool shutdown completed")
}

func TestWorkerStart(t *testing.T) {
	// 测试工作协程启动
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	worker := &Worker{
		ID:      1,
		TaskCh:  make(chan Task, 10),
		QuitCh:  make(chan bool),
		Context: ctx,
	}
	
	var wg sync.WaitGroup
	wg.Add(1)
	
	// 启动工作协程
	go worker.Start(&wg)
	
	// 提交一个简单任�?	taskFunc := func() (interface{}, error) {
		return "worker task result", nil
	}
	
	resultChan := make(chan TaskResult, 1)
	task := Task{
		Function: taskFunc,
		Result:   resultChan,
	}
	
	worker.TaskCh <- task
	
	// 等待结果
	select {
	case result := <-resultChan:
		if result.Error != nil {
			t.Errorf("Worker task execution failed: %v", result.Error)
		}
		if result.Value != "worker task result" {
			t.Errorf("Worker task result mismatch: expected 'worker task result', got %v", result.Value)
		}
	case <-time.After(5 * time.Second):
		t.Error("Worker task execution timeout")
	}
	
	// 关闭工作协程
	close(worker.QuitCh)
	wg.Wait()
}
