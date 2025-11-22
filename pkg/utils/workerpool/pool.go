package workerpool

import (
	"context"
	"sync"
)

// Pool Goroutine 池
type Pool struct {
	workers   int
	taskQueue chan func()
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
}

// New 创建新的 Goroutine 池
func New(workers int) *Pool {
	if workers <= 0 {
		workers = 10 // 默认 10 个 worker
	}

	ctx, cancel := context.WithCancel(context.Background())

	pool := &Pool{
		workers:   workers,
		taskQueue: make(chan func(), workers*2), // 缓冲队列
		ctx:       ctx,
		cancel:    cancel,
	}

	// 启动 workers
	for i := 0; i < workers; i++ {
		pool.wg.Add(1)
		go pool.worker()
	}

	return pool
}

// worker 工作协程
func (p *Pool) worker() {
	defer p.wg.Done()

	for {
		select {
		case task, ok := <-p.taskQueue:
			if !ok {
				return
			}
			task()
		case <-p.ctx.Done():
			return
		}
	}
}

// Submit 提交任务
func (p *Pool) Submit(task func()) {
	select {
	case p.taskQueue <- task:
	case <-p.ctx.Done():
		return
	}
}

// Wait 等待所有任务完成
func (p *Pool) Wait() {
	close(p.taskQueue)
	p.wg.Wait()
}

// Stop 停止池
func (p *Pool) Stop() {
	p.cancel()
	p.wg.Wait()
}

// Size 返回池大小
func (p *Pool) Size() int {
	return p.workers
}
