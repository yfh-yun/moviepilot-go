package schedulers

import (
	"sync"
)

// JobRegistry 任务注册表
type JobRegistry struct {
	jobs map[string]Job
	mu   sync.RWMutex
}

// NewJobRegistry 创建任务注册表
func NewJobRegistry() *JobRegistry {
	return &JobRegistry{
		jobs: make(map[string]Job),
	}
}

// Register 注册任务
func (r *JobRegistry) Register(job Job) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.jobs[job.ID()] = job
}

// Unregister 注销任务
func (r *JobRegistry) Unregister(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.jobs, id)
}

// Get 获取任务
func (r *JobRegistry) Get(id string) (Job, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	job, exists := r.jobs[id]
	return job, exists
}

// List 列出所有任务
func (r *JobRegistry) List() []Job {
	r.mu.RLock()
	defer r.mu.RUnlock()

	jobs := make([]Job, 0, len(r.jobs))
	for _, job := range r.jobs {
		jobs = append(jobs, job)
	}

	return jobs
}

// Clear 清空注册表
func (r *JobRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.jobs = make(map[string]Job)
}
