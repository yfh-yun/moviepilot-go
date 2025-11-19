package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// JobStatus 表示任务状态
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
)

// JobType 表示任务类型
type JobType string

const (
	JobTypeSubscribe JobType = "subscribe"
	JobTypeDownload  JobType = "download"
	JobTypeTransfer  JobType = "transfer"
	JobTypeScan      JobType = "scan"
	JobTypeCleanup   JobType = "cleanup"
	JobTypeBackup    JobType = "backup"
	JobTypePlugin    JobType = "plugin"
)

// JobResult 表示任务执行结果
type JobResult struct {
	ID        string         `json:"id"`
	Status    JobStatus      `json:"status"`
	StartTime time.Time      `json:"start_time"`
	EndTime   *time.Time     `json:"end_time,omitempty"`
	Duration  *time.Duration `json:"duration,omitempty"`
	Error     *string        `json:"error,omitempty"`
	Result    interface{}    `json:"result,omitempty"`
}

// JobConfig 表示任务配置
type JobConfig struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Type        JobType       `json:"type"`
	Schedule    string        `json:"schedule"` // cron表达式或interval
	Enabled     bool          `json:"enabled"`
	MaxRetries  int           `json:"max_retries"`
	Timeout     time.Duration `json:"timeout"`
	Concurrent  bool          `json:"concurrent"`
	Description string        `json:"description"`
	Params      interface{}   `json:"params"`
}

// JobHandler 表示任务处理器
type JobHandler func(ctx context.Context, config *JobConfig) (*JobResult, error)

// JobScheduler 任务调度器
type JobScheduler struct {
	logger    *zap.Logger
	scheduler gocron.Scheduler
	jobs      map[string]*JobConfig
	handlers  map[JobType]JobHandler
	results   map[string]*JobResult
	running   map[string]bool
}

// NewJobScheduler 创建新的任务调度器
func NewJobScheduler(logger *zap.Logger) (*JobScheduler, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduler: %w", err)
	}

	return &JobScheduler{
		logger:    logger,
		scheduler: s,
		jobs:      make(map[string]*JobConfig),
		handlers:  make(map[JobType]JobHandler),
		results:   make(map[string]*JobResult),
		running:   make(map[string]bool),
	}, nil
}

// RegisterHandler 注册任务处理器
func (s *JobScheduler) RegisterHandler(jobType JobType, handler JobHandler) {
	s.handlers[jobType] = handler
	s.logger.Info("注册任务处理器", zap.String("type", string(jobType)))
}

// AddJob 添加任务
func (s *JobScheduler) AddJob(config *JobConfig) error {
	if config.ID == "" {
		config.ID = uuid.New().String()
	}

	if _, exists := s.jobs[config.ID]; exists {
		return fmt.Errorf("job with ID %s already exists", config.ID)
	}

	s.jobs[config.ID] = config

	if !config.Enabled {
		s.logger.Info("任务已添加但未启用", zap.String("id", config.ID), zap.String("name", config.Name))
		return nil
	}

	return s.scheduleJob(config)
}

// scheduleJob 调度任务
func (s *JobScheduler) scheduleJob(config *JobConfig) error {
	jobFunc := func() {
		s.executeJob(config)
	}

	_, err := s.scheduler.NewJob(
		gocron.CronJob(config.Schedule, false),
		gocron.NewTask(jobFunc),
	)
	if err != nil {
		return fmt.Errorf("failed to schedule job %s: %w", config.ID, err)
	}

	s.logger.Info("任务已调度",
		zap.String("id", config.ID),
		zap.String("name", config.Name),
		zap.String("schedule", config.Schedule))
	return nil
}

// executeJob 执行任务
func (s *JobScheduler) executeJob(config *JobConfig) {
	if s.running[config.ID] && !config.Concurrent {
		s.logger.Warn("任务正在运行，跳过本次执行", zap.String("id", config.ID))
		return
	}

	s.running[config.ID] = true
	defer func() {
		s.running[config.ID] = false
	}()

	result := &JobResult{
		ID:        uuid.New().String(),
		Status:    JobStatusRunning,
		StartTime: time.Now(),
	}

	s.results[result.ID] = result

	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	handler, exists := s.handlers[config.Type]
	if !exists {
		errorMsg := fmt.Sprintf("no handler registered for job type %s", config.Type)
		s.logger.Error("任务执行失败", zap.String("id", config.ID), zap.String("error", errorMsg))
		s.updateJobResult(result, JobStatusFailed, &errorMsg, nil)
		return
	}

	jobResult, err := handler(ctx, config)
	if err != nil {
		errorMsg := err.Error()
		s.logger.Error("任务执行失败", zap.String("id", config.ID), zap.Error(err))
		s.updateJobResult(result, JobStatusFailed, &errorMsg, nil)
		return
	}

	s.updateJobResult(result, JobStatusCompleted, nil, jobResult)
	s.logger.Info("任务执行完成", zap.String("id", config.ID), zap.String("name", config.Name))
}

// updateJobResult 更新任务结果
func (s *JobScheduler) updateJobResult(result *JobResult, status JobStatus, errorMsg *string, jobResult interface{}) {
	endTime := time.Now()
	duration := endTime.Sub(result.StartTime)

	result.Status = status
	result.EndTime = &endTime
	result.Duration = &duration
	result.Error = errorMsg
	result.Result = jobResult

	// 保存到数据库或其他持久化存储
	s.logger.Debug("任务结果更新",
		zap.String("id", result.ID),
		zap.String("status", string(status)),
		zap.Duration("duration", duration))
}

// GetJobStatus 获取任务状态
func (s *JobScheduler) GetJobStatus(jobID string) (*JobResult, error) {
	result, exists := s.results[jobID]
	if !exists {
		return nil, fmt.Errorf("job result not found: %s", jobID)
	}
	return result, nil
}

// ListJobs 列出所有任务
func (s *JobScheduler) ListJobs() []*JobConfig {
	jobs := make([]*JobConfig, 0, len(s.jobs))
	for _, job := range s.jobs {
		jobs = append(jobs, job)
	}
	return jobs
}

// EnableJob 启用任务
func (s *JobScheduler) EnableJob(jobID string) error {
	job, exists := s.jobs[jobID]
	if !exists {
		return fmt.Errorf("job not found: %s", jobID)
	}

	job.Enabled = true
	return s.scheduleJob(job)
}

// DisableJob 禁用任务
func (s *JobScheduler) DisableJob(jobID string) error {
	job, exists := s.jobs[jobID]
	if !exists {
		return fmt.Errorf("job not found: %s", jobID)
	}

	job.Enabled = false
	// 从调度器中移除任务（需要实现）
	s.logger.Info("任务已禁用", zap.String("id", jobID), zap.String("name", job.Name))
	return nil
}

// Start 启动调度器
func (s *JobScheduler) Start() {
	s.scheduler.Start()
	s.logger.Info("任务调度器已启动")
}

// Stop 停止调度器
func (s *JobScheduler) Stop() {
	if err := s.scheduler.Shutdown(); err != nil {
		s.logger.Error("调度器停止失败", zap.Error(err))
	} else {
		s.logger.Info("任务调度器已停止")
	}
}

// InitializeDefaultJobs 初始化默认任务
func (s *JobScheduler) InitializeDefaultJobs() error {
	defaultJobs := []*JobConfig{
		{
			ID:          "subscribe-scan",
			Name:        "订阅扫描",
			Type:        JobTypeSubscribe,
			Schedule:    "*/10 * * * *", // 每10分钟
			Enabled:     true,
			MaxRetries:  3,
			Timeout:     5 * time.Minute,
			Concurrent:  false,
			Description: "扫描订阅状态并处理新的媒体",
		},
		{
			ID:          "download-monitor",
			Name:        "下载监控",
			Type:        JobTypeDownload,
			Schedule:    "*/1 * * * *", // 每1分钟
			Enabled:     true,
			MaxRetries:  5,
			Timeout:     2 * time.Minute,
			Concurrent:  true,
			Description: "监控下载任务状态和进度",
		},
		{
			ID:          "file-scan",
			Name:        "文件扫描",
			Type:        JobTypeScan,
			Schedule:    "0 */2 * * *", // 每2小时
			Enabled:     true,
			MaxRetries:  2,
			Timeout:     10 * time.Minute,
			Concurrent:  false,
			Description: "扫描媒体文件目录并更新媒体库",
		},
		{
			ID:          "cleanup-history",
			Name:        "历史清理",
			Type:        JobTypeCleanup,
			Schedule:    "0 2 * * *", // 每天凌晨2点
			Enabled:     true,
			MaxRetries:  1,
			Timeout:     30 * time.Minute,
			Concurrent:  false,
			Description: "清理过期的历史记录和临时文件",
		},
	}

	for _, job := range defaultJobs {
		if err := s.AddJob(job); err != nil {
			return fmt.Errorf("failed to add default job %s: %w", job.ID, err)
		}
	}

	s.logger.Info("默认任务初始化完成", zap.Int("count", len(defaultJobs)))
	return nil
}
