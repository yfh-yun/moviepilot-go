package helper

import (
	"context"
	"runtime"
	"sync"
)

// ThreadHelper 线程池管�?type ThreadHelper struct {
	// 使用goroutine和channel实现线程池功�?	maxWorkers int
	taskQueue  chan Task
	workers    []*Worker
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// Task 任务结构
type Task struct {
	Function func() (interface{}, error)
	Result   chan TaskResult
}

// TaskResult 任务执行结果
type TaskResult struct {
	Value interface{}
	Error error
}

// Worker 工作协程
type Worker struct {
	ID      int
	TaskCh  chan Task
	QuitCh  chan bool
	Context context.Context
}

// NewThreadHelper 创建ThreadHelper实例（单例模式）
var (
	threadHelperInstance *ThreadHelper
	threadHelperOnce     sync.Once
)

// NewThreadHelper 创建ThreadHelper单例实例
func NewThreadHelper() *ThreadHelper {
	threadHelperOnce.Do(func() {
		// 根据CPU核心数计算worker数量，类似Python版本中的 multiprocessing.cpu_count() * 2 + 1
		numWorkers := runtime.NumCPU()*2 + 1
		
		ctx, cancel := context.WithCancel(context.Background())
		
		threadHelperInstance = &ThreadHelper{
			maxWorkers: numWorkers,
			taskQueue:  make(chan Task, 1000), // 缓冲任务队列
			ctx:        ctx,
			cancel:     cancel,
		}
		
		// 启动工作协程
		threadHelperInstance.startWorkers()
	})
	
	return threadHelperInstance
}

// startWorkers 启动工作协程
func (t *ThreadHelper) startWorkers() {
	t.workers = make([]*Worker, t.maxWorkers)
	
	for i := 0; i < t.maxWorkers; i++ {
		worker := &Worker{
			ID:      i,
			TaskCh:  make(chan Task, 100), // 每个工作协程的任务缓冲队�?			QuitCh:  make(chan bool),
			Context: t.ctx,
		}
		
		t.workers[i] = worker
		t.wg.Add(1)
		
		// 启动工作协程
		go worker.Start(&t.wg)
	}
	
	// 启动任务分发�?	go t.dispatchTasks()
}

// dispatchTasks 任务分发�?func (t *ThreadHelper) dispatchTasks() {
	for {
		select {
		case task := <-t.taskQueue:
			// 简单的轮询分发任务
			// 在实际应用中，可以使用更复杂的负载均衡策�?			for _, worker := range t.workers {
				select {
				case worker.TaskCh <- task:
					goto dispatched
				default:
					// 如果当前工作协程忙碌，尝试下一�?					continue
				}
			}
			
			// 如果所有工作协程都忙碌，阻塞等�?			if len(t.workers) > 0 {
				worker := t.workers[0] // 简化处理，实际应用中可以轮�?				worker.TaskCh <- task
			}
			
		dispatched:
		case <-t.ctx.Done():
			return
		}
	}
}

// Start 启动工作协程
func (w *Worker) Start(wg *sync.WaitGroup) {
	defer wg.Done()
	
	for {
		select {
		case task := <-w.TaskCh:
			// 执行任务
			value, err := task.Function()
			
			// 返回结果
			task.Result <- TaskResult{
				Value: value,
				Error: err,
			}
		case <-w.QuitCh:
			return
		case <-w.Context.Done():
			return
		}
	}
}

// Submit 提交任务
func (t *ThreadHelper) Submit(function func() (interface{}, error)) chan TaskResult {
	/*
	 * 提交任务
	 * :param function: 函数
	 * :return: 结果通道
	 */
	resultChan := make(chan TaskResult, 1)
	
	task := Task{
		Function: function,
		Result:   resultChan,
	}
	
	// 将任务放入队�?	t.taskQueue <- task
	
	return resultChan
}

// SubmitFunc 提交任务（简化版本，只接受无返回值的函数�?func (t *ThreadHelper) SubmitFunc(function func()) {
	/*
	 * 提交任务（简化版本）
	 * :param function: 函数
	 */
	taskFunc := func() (interface{}, error) {
		function()
		return nil, nil
	}
	
	resultChan := make(chan TaskResult, 1)
	
	task := Task{
		Function: taskFunc,
		Result:   resultChan,
	}
	
	// 将任务放入队�?	t.taskQueue <- task
}

// Shutdown 关闭线程�?func (t *ThreadHelper) Shutdown() {
	/*
	 * 关闭线程�?	 */
	// 取消上下�?	t.cancel()
	
	// 等待所有工作协程结�?	t.wg.Wait()
	
	// 关闭任务队列
	close(t.taskQueue)
	
	// 关闭所有工作协�?	for _, worker := range t.workers {
		close(worker.QuitCh)
		close(worker.TaskCh)
	}
}
