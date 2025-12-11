package workerpool

import (
	"sync"
)

// Pool 工作池

type Pool struct {
	workers   int
	jobs      chan func()
	waitGroup sync.WaitGroup
}

// New 创建工作池
func New(workers int) *Pool {
	if workers <= 0 {
		workers = 4
	}

	pool := &Pool{
		workers: workers,
		jobs:    make(chan func(), workers*2),
	}

	// 启动工作协程
	for i := 0; i < workers; i++ {
		go pool.worker()
	}

	return pool
}

// worker 工作协程
func (p *Pool) worker() {
	for job := range p.jobs {
		job()
		p.waitGroup.Done()
	}
}

// Submit 提交任务
func (p *Pool) Submit(job func()) {
	p.waitGroup.Add(1)
	p.jobs <- job
}

// Wait 等待所有任务完成
func (p *Pool) Wait() {
	p.waitGroup.Wait()
}

// Close 关闭工作池
func (p *Pool) Close() {
	close(p.jobs)
}
