package utils

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ThreadHelper 线程池管理助手
type ThreadHelper struct {
	pool     *WorkerPool
	maxWorkers int
	mutex    sync.RWMutex
}

// WorkerPool 工作池
type WorkerPool struct {
	workers    int
	jobQueue   chan Job
	workerPool chan chan Job
	quit       chan bool
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
}

// Job 任务接口
type Job interface {
	Execute() error
}

// TaskFunc 函数任务
type TaskFunc struct {
	fn func() error
}

// Execute 执行函数任务
func (t *TaskFunc) Execute() error {
	return t.fn()
}

// NewThreadHelper 创建线程助手实例
func NewThreadHelper(maxWorkers int) *ThreadHelper {
	if maxWorkers <= 0 {
		maxWorkers = 10 // 默认10个工作线程
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &ThreadHelper{
		maxWorkers: maxWorkers,
		pool: &WorkerPool{
			workers:    maxWorkers,
			jobQueue:   make(chan Job, maxWorkers*2),
			workerPool: make(chan chan Job, maxWorkers),
			quit:       make(chan bool),
			ctx:        ctx,
			cancel:     cancel,
		},
	}
}

// Start 启动线程池
func (th *ThreadHelper) Start() {
	th.mutex.Lock()
	defer th.mutex.Unlock()

	for i := 0; i < th.maxWorkers; i++ {
		worker := NewWorker(th.pool.workerPool, th.pool.quit)
		worker.Start()
	}

	go th.pool.dispatch()
}

// Stop 停止线程池
func (th *ThreadHelper) Stop() {
	th.mutex.Lock()
	defer th.mutex.Unlock()

	if th.pool != nil {
		th.pool.cancel()
		close(th.pool.quit)
		th.pool.wg.Wait()
	}
}

// Submit 提交任务
func (th *ThreadHelper) Submit(job Job) error {
	th.mutex.RLock()
	defer th.mutex.RUnlock()

	if th.pool == nil {
		return fmt.Errorf("thread pool is not initialized")
	}

	select {
	case th.pool.jobQueue <- job:
		return nil
	case <-th.pool.ctx.Done():
		return fmt.Errorf("thread pool is shutting down")
	default:
		return fmt.Errorf("job queue is full")
	}
}

// SubmitFunc 提交函数任务
func (th *ThreadHelper) SubmitFunc(fn func() error) error {
	task := &TaskFunc{fn: fn}
	return th.Submit(task)
}

// SubmitWithTimeout 提交任务并设置超时
func (th *ThreadHelper) SubmitWithTimeout(job Job, timeout time.Duration) error {
	th.mutex.RLock()
	defer th.mutex.RUnlock()

	if th.pool == nil {
		return fmt.Errorf("thread pool is not initialized")
	}

	select {
	case th.pool.jobQueue <- job:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("submit job timeout after %v", timeout)
	case <-th.pool.ctx.Done():
		return fmt.Errorf("thread pool is shutting down")
	}
}

// GetQueueLength 获取队列长度
func (th *ThreadHelper) GetQueueLength() int {
	th.mutex.RLock()
	defer th.mutex.RUnlock()

	if th.pool == nil {
		return 0
	}

	return len(th.pool.jobQueue)
}

// GetWorkerCount 获取工作线程数量
func (th *ThreadHelper) GetWorkerCount() int {
	th.mutex.RLock()
	defer th.mutex.RUnlock()

	return th.maxWorkers
}

// SetWorkerCount 设置工作线程数量
func (th *ThreadHelper) SetWorkerCount(count int) error {
	th.mutex.Lock()
	defer th.mutex.Unlock()

	if count <= 0 {
		return fmt.Errorf("worker count must be positive")
	}

	if count == th.maxWorkers {
		return nil
	}

	// 停止当前线程池
	if th.pool != nil {
		th.pool.cancel()
		close(th.pool.quit)
		th.pool.wg.Wait()
	}

	// 创建新的线程池
	th.maxWorkers = count
	ctx, cancel := context.WithCancel(context.Background())
	th.pool = &WorkerPool{
		workers:    count,
		jobQueue:   make(chan Job, count*2),
		workerPool: make(chan chan Job, count),
		quit:       make(chan bool),
		ctx:        ctx,
		cancel:     cancel,
	}

	// 启动新线程池
	for i := 0; i < count; i++ {
		worker := NewWorker(th.pool.workerPool, th.pool.quit)
		worker.Start()
	}

	go th.pool.dispatch()

	return nil
}

// dispatch 分发任务
func (wp *WorkerPool) dispatch() {
	for {
		select {
		case job := <-wp.jobQueue:
			go func() {
				jobChannel := <-wp.workerPool
				jobChannel <- job
			}()
		case <-wp.quit:
			return
		case <-wp.ctx.Done():
			return
		}
	}
}

// Worker 工作线程
type Worker struct {
	workerPool chan chan Job
	jobChannel chan Job
	quit       chan bool
}

// NewWorker 创建工作线程
func NewWorker(workerPool chan chan Job, quit chan bool) *Worker {
	return &Worker{
		workerPool: workerPool,
		jobChannel: make(chan Job),
		quit:       quit,
	}
}

// Start 启动工作线程
func (w *Worker) Start() {
	go func() {
		for {
			// 将当前工作线程注册到工作池
			w.workerPool <- w.jobChannel

			select {
			case job := <-w.jobChannel:
				// 执行任务
				if err := job.Execute(); err != nil {
					// 记录错误日志（这里简化处理）
					fmt.Printf("Job execution error: %v\n", err)
				}
			case <-w.quit:
				return
			}
		}
	}()
}

// Stop 停止工作线程
func (w *Worker) Stop() {
	close(w.quit)
}

// AsyncTask 异步任务结果
type AsyncTask struct {
	Result interface{}
	Error  error
	Done   chan struct{}
}

// NewAsyncTask 创建异步任务
func NewAsyncTask() *AsyncTask {
	return &AsyncTask{
		Done: make(chan struct{}),
	}
}

// Complete 完成任务
func (at *AsyncTask) Complete(result interface{}, err error) {
	at.Result = result
	at.Error = err
	close(at.Done)
}

// Wait 等待任务完成
func (at *AsyncTask) Wait() (interface{}, error) {
	<-at.Done
	return at.Result, at.Error
}

// WaitWithTimeout 带超时的等待
func (at *AsyncTask) WaitWithTimeout(timeout time.Duration) (interface{}, error) {
	select {
	case <-at.Done:
		return at.Result, at.Error
	case <-time.After(timeout):
		return nil, fmt.Errorf("task timeout after %v", timeout)
	}
}