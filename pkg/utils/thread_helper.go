package utils

import (
	"sync"

	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// ThreadHelper 线程池管理帮助类
type ThreadHelper struct {
	logger     *zap.Logger
	maxWorkers int
	workers    chan struct{}
	wg         sync.WaitGroup
	mutex      sync.Mutex
	closed     bool
}

// NewThreadHelper 创建线程池管理实例
func NewThreadHelper(maxWorkers int) *ThreadHelper {
	return &ThreadHelper{
		logger:     logger.GetLogger(),
		maxWorkers: maxWorkers,
		workers:    make(chan struct{}, maxWorkers),
		closed:     false,
	}
}

// Submit 提交任务到线程池
func (h *ThreadHelper) Submit(task func()) {
	h.mutex.Lock()
	if h.closed {
		h.mutex.Unlock()
		h.logger.Error("线程池已关闭，无法提交新任务")
		return
	}
	h.wg.Add(1)
	h.mutex.Unlock()

	// 等待可用的工作者
	h.workers <- struct{}{}

	// 启动goroutine执行任务
	go func() {
		defer func() {
			// 释放工作者
			<-h.workers
			// 减少等待组计数
			h.wg.Done()
			// 捕获panic
			if r := recover(); r != nil {
				h.logger.Error("任务执行panic", zap.Any("panic", r))
			}
		}()

		// 执行任务
		task()
	}()
}

// SubmitWithResult 提交任务到线程池，并返回结果
func (h *ThreadHelper) SubmitWithResult(task func() any) <-chan any {
	result := make(chan any, 1)

	h.Submit(func() {
		defer close(result)
		result <- task()
	})

	return result
}

// Wait 等待所有任务完成
func (h *ThreadHelper) Wait() {
	h.wg.Wait()
}

// Shutdown 关闭线程池
func (h *ThreadHelper) Shutdown() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.closed {
		return
	}

	h.logger.Info("正在关闭线程池...")

	// 标记线程池已关闭
	h.closed = true

	// 等待所有任务完成
	h.mutex.Unlock()
	h.wg.Wait()
	h.mutex.Lock()

	// 关闭工作者通道
	close(h.workers)

	h.logger.Info("线程池已关闭")
}

// GetMaxWorkers 获取最大工作者数量
func (h *ThreadHelper) GetMaxWorkers() int {
	return h.maxWorkers
}

// GetActiveWorkers 获取当前活跃的工作者数量
func (h *ThreadHelper) GetActiveWorkers() int {
	return len(h.workers)
}

// IsClosed 检查线程池是否已关闭
func (h *ThreadHelper) IsClosed() bool {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	return h.closed
}
