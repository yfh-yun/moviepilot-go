package schedulers

import (
	"context"
	"strconv"
	"sync"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// Scheduler 调度器接口
type Scheduler interface {
	Start() error
	Stop() error
	AddJob(job Job, trigger Trigger) error
	RemoveJob(id string) error
	PauseJob(id string) error
	ResumeJob(id string) error
	RunJob(id string) error // 立即执行
	ListJobs() []JobInfo
}

// scheduler 调度器实现
type scheduler struct {
	cron   *cron.Cron
	jobs   map[string]*jobEntry
	mu     sync.RWMutex
	logger *zap.Logger
}

// jobEntry 任务条目
type jobEntry struct {
	job     Job
	entryID cron.EntryID
	paused  bool
	trigger Trigger // 保存trigger用于恢复任务
}

// NewScheduler 创建调度器实例
func NewScheduler() Scheduler {
	return &scheduler{
		cron:   cron.New(cron.WithSeconds()),
		jobs:   make(map[string]*jobEntry),
		logger: logger.GetLogger(),
	}
}

// Start 启动调度器
func (s *scheduler) Start() error {
	s.cron.Start()
	s.logger.Info("scheduler started")
	return nil
}

// Stop 停止调度器
func (s *scheduler) Stop() error {
	ctx := s.cron.Stop()
	<-ctx.Done()
	s.logger.Info("scheduler stopped")
	return nil
}

// AddJob 添加任务
func (s *scheduler) AddJob(job Job, trigger Trigger) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 将Trigger转换为cron spec
	spec := s.triggerToCronSpec(trigger)

	entryID, err := s.cron.AddFunc(spec, func() {
		s.runJob(job)
	})
	if err != nil {
		return err
	}

	s.jobs[job.ID()] = &jobEntry{
		job:     job,
		entryID: entryID,
		paused:  false,
		trigger: trigger,
	}

	s.logger.Info("job added",
		zap.String("id", job.ID()),
		zap.String("name", job.Name()),
		zap.String("spec", spec))
	return nil
}

// RemoveJob 删除任务
func (s *scheduler) RemoveJob(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.jobs[id]
	if !exists {
		return nil
	}

	s.cron.Remove(entry.entryID)
	delete(s.jobs, id)

	s.logger.Info("job removed", zap.String("id", id))
	return nil
}

// PauseJob 暂停任务
func (s *scheduler) PauseJob(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.jobs[id]
	if !exists {
		return nil
	}

	// robfig/cron/v3 不支持 Pause，通过删除任务来实现暂停
	s.cron.Remove(entry.entryID)
	entry.paused = true

	s.logger.Info("job paused", zap.String("id", id))
	return nil
}

// ResumeJob 恢复任务
func (s *scheduler) ResumeJob(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.jobs[id]
	if !exists {
		return nil
	}

	if !entry.paused {
		return nil
	}

	// 重新添加任务
	spec := s.triggerToCronSpec(entry.trigger)
	entryID, err := s.cron.AddFunc(spec, func() {
		s.runJob(entry.job)
	})
	if err != nil {
		return err
	}

	entry.entryID = entryID
	entry.paused = false

	s.logger.Info("job resumed", zap.String("id", id))
	return nil
}

// RunJob 立即执行任务
func (s *scheduler) RunJob(id string) error {
	s.mu.RLock()
	entry, exists := s.jobs[id]
	s.mu.RUnlock()

	if !exists {
		return nil
	}

	s.runJob(entry.job)
	return nil
}

// ListJobs 列出所有任务
func (s *scheduler) ListJobs() []JobInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	infos := make([]JobInfo, 0, len(s.jobs))
	for _, entry := range s.jobs {
		infos = append(infos, JobInfo{
			ID:     entry.job.ID(),
			Name:   entry.job.Name(),
			Status: s.getJobStatus(entry),
		})
	}

	return infos
}

// runJob 执行任务
func (s *scheduler) runJob(job Job) {
	ctx := context.Background()
	s.logger.Info("job started", zap.String("id", job.ID()))

	if err := job.Run(ctx); err != nil {
		s.logger.Error("job failed",
			zap.String("id", job.ID()),
			zap.Error(err))
	} else {
		s.logger.Info("job completed", zap.String("id", job.ID()))
	}
}

// triggerToCronSpec 将Trigger转换为Cron表达式
func (s *scheduler) triggerToCronSpec(trigger Trigger) string {
	switch t := trigger.(type) {
	case *IntervalTrigger:
		// 将间隔转换为Cron表达式
		// 简化处理：这里只处理小时和分钟
		minutes := int(t.Interval.Minutes())
		if minutes < 60 {
			return "0 */" + strconv.Itoa(minutes) + " * * * *" // 每N分钟
		}
		hours := int(t.Interval.Hours())
		return "0 0 */" + strconv.Itoa(hours) + " * * *" // 每N小时
	case *CronTrigger:
		return t.Spec
	default:
		return "0 0 * * * *" // 默认每天凌晨
	}
}

// getJobStatus 获取任务状态
func (s *scheduler) getJobStatus(entry *jobEntry) string {
	if entry.paused {
		return "paused"
	}
	return "running"
}
