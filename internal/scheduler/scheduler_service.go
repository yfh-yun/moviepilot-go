package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/yfh-yun/moviepilot-go/internal/repository"
	"github.com/yfh-yun/moviepilot-go/internal/service"

	"go.uber.org/zap"
)

// SchedulerService 调度器服务
type SchedulerService struct {
	logger           *zap.Logger
	jobScheduler     *JobScheduler
	workflowEngine   *WorkflowEngine
	workflowRepo     repository.WorkflowRepository
	subscribeService service.SubscribeService
	downloadService  service.DownloadService
	messageService   service.MessageService
	pluginService    service.PluginService
}

// NewSchedulerService 创建新的调度器服务
func NewSchedulerService(
	logger *zap.Logger,
	workflowRepo repository.WorkflowRepository,
	subscribeService service.SubscribeService,
	downloadService service.DownloadService,
	messageService service.MessageService,
	pluginService service.PluginService,
) (*SchedulerService, error) {

	jobScheduler, err := NewJobScheduler(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create job scheduler: %w", err)
	}

	workflowEngine := NewWorkflowEngine(logger)

	service := &SchedulerService{
		logger:           logger,
		jobScheduler:     jobScheduler,
		workflowEngine:   workflowEngine,
		workflowRepo:     workflowRepo,
		subscribeService: subscribeService,
		downloadService:  downloadService,
		messageService:   messageService,
		pluginService:    pluginService,
	}

	// 初始化默认处理器
	service.initializeHandlers()

	return service, nil
}

// initializeHandlers 初始化处理器
func (s *SchedulerService) initializeHandlers() {
	// 注册任务处理器
	s.jobScheduler.RegisterHandler(JobTypeSubscribe, s.handleSubscribeJob)
	s.jobScheduler.RegisterHandler(JobTypeDownload, s.handleDownloadJob)
	s.jobScheduler.RegisterHandler(JobTypeScan, s.handleScanJob)
	s.jobScheduler.RegisterHandler(JobTypeCleanup, s.handleCleanupJob)
	s.jobScheduler.RegisterHandler(JobTypeBackup, s.handleBackupJob)
	s.jobScheduler.RegisterHandler(JobTypePlugin, s.handlePluginJob)

	// 初始化工作流处理器
	s.workflowEngine.InitializeDefaultHandlers()

	s.logger.Info("调度器服务处理器初始化完成")
}

// Start 启动调度器服务
func (s *SchedulerService) Start() error {
	// 初始化默认任务
	if err := s.jobScheduler.InitializeDefaultJobs(); err != nil {
		return fmt.Errorf("failed to initialize default jobs: %w", err)
	}

	// 加载活跃的工作流
	if err := s.loadActiveWorkflows(); err != nil {
		return fmt.Errorf("failed to load active workflows: %w", err)
	}

	// 启动调度器
	s.jobScheduler.Start()

	s.logger.Info("调度器服务已启动")
	return nil
}

// Stop 停止调度器服务
func (s *SchedulerService) Stop() {
	s.jobScheduler.Stop()
	s.logger.Info("调度器服务已停止")
}

// loadActiveWorkflows 加载活跃的工作流
func (s *SchedulerService) loadActiveWorkflows() error {
	ctx := context.Background()

	workflows, err := s.workflowRepo.FindByStatus(ctx, "active")
	if err != nil {
		return fmt.Errorf("failed to find active workflows: %w", err)
	}

	s.logger.Info("加载活跃工作流", zap.Int("count", len(workflows)))

	// 这里可以添加工作流到调度器的逻辑
	// 根据工作流的触发器类型调度工作流

	return nil
}

// Job handlers

// handleSubscribeJob 处理订阅任务
func (s *SchedulerService) handleSubscribeJob(ctx context.Context, config *JobConfig) (*JobResult, error) {
	s.logger.Info("执行订阅扫描任务",
		zap.String("job_id", config.ID),
		zap.String("job_name", config.Name))

	// 调用订阅服务进行扫描
	// 实际实现需要根据配置参数调用相应的订阅服务方法

	result := &JobResult{
		Status: JobStatusCompleted,
		Result: map[string]interface{}{
			"scanned_count":   0,
			"processed_count": 0,
			"errors":          0,
		},
	}

	return result, nil
}

// handleDownloadJob 处理下载任务
func (s *SchedulerService) handleDownloadJob(ctx context.Context, config *JobConfig) (*JobResult, error) {
	s.logger.Info("执行下载监控任务",
		zap.String("job_id", config.ID),
		zap.String("job_name", config.Name))

	// 调用下载服务进行监控
	// 实际实现需要根据配置参数调用相应的下载服务方法

	result := &JobResult{
		Status: JobStatusCompleted,
		Result: map[string]interface{}{
			"monitored_count": 0,
			"completed_count": 0,
			"failed_count":    0,
		},
	}

	return result, nil
}

// handleScanJob 处理扫描任务
func (s *SchedulerService) handleScanJob(ctx context.Context, config *JobConfig) (*JobResult, error) {
	s.logger.Info("执行文件扫描任务",
		zap.String("job_id", config.ID),
		zap.String("job_name", config.Name))

	// 实际实现需要调用文件扫描服务

	result := &JobResult{
		Status: JobStatusCompleted,
		Result: map[string]interface{}{
			"scanned_files": 0,
			"new_media":     0,
			"errors":        0,
		},
	}

	return result, nil
}

// handleCleanupJob 处理清理任务
func (s *SchedulerService) handleCleanupJob(ctx context.Context, config *JobConfig) (*JobResult, error) {
	s.logger.Info("执行历史清理任务",
		zap.String("job_id", config.ID),
		zap.String("job_name", config.Name))

	// 实际实现需要调用清理服务

	result := &JobResult{
		Status: JobStatusCompleted,
		Result: map[string]interface{}{
			"cleaned_records": 0,
			"freed_space":     "0MB",
			"errors":          0,
		},
	}

	return result, nil
}

// handleBackupJob 处理备份任务
func (s *SchedulerService) handleBackupJob(ctx context.Context, config *JobConfig) (*JobResult, error) {
	s.logger.Info("执行备份任务",
		zap.String("job_id", config.ID),
		zap.String("job_name", config.Name))

	// 实际实现需要调用备份服务

	result := &JobResult{
		Status: JobStatusCompleted,
		Result: map[string]interface{}{
			"backup_size": "0MB",
			"backup_time": time.Now().Format(time.RFC3339),
			"errors":      0,
		},
	}

	return result, nil
}

// handlePluginJob 处理插件任务
func (s *SchedulerService) handlePluginJob(ctx context.Context, config *JobConfig) (*JobResult, error) {
	s.logger.Info("执行插件任务",
		zap.String("job_id", config.ID),
		zap.String("job_name", config.Name))

	// 实际实现需要调用插件服务

	result := &JobResult{
		Status: JobStatusCompleted,
		Result: map[string]interface{}{
			"executed_plugins": 0,
			"success_count":    0,
			"failed_count":     0,
		},
	}

	return result, nil
}

// Workflow methods

// ExecuteWorkflow 执行工作流
func (s *SchedulerService) ExecuteWorkflow(ctx context.Context, workflowID string, triggerData map[string]interface{}) (*WorkflowInstance, error) {
	workflow, err := s.workflowRepo.FindByID(ctx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to find workflow: %w", err)
	}

	return s.workflowEngine.ExecuteWorkflow(ctx, workflow, triggerData)
}

// GetWorkflowInstance 获取工作流实例
func (s *SchedulerService) GetWorkflowInstance(instanceID string) (*WorkflowInstance, error) {
	return s.workflowEngine.GetWorkflowInstance(instanceID)
}

// ListJobStatus 列出任务状态
func (s *SchedulerService) ListJobStatus() []*JobConfig {
	return s.jobScheduler.ListJobs()
}

// ListWorkflowInstances 列出工作流实例
func (s *SchedulerService) ListWorkflowInstances() []*WorkflowInstance {
	return s.workflowEngine.ListWorkflowInstances()
}

// AddJob 添加任务
func (s *SchedulerService) AddJob(config *JobConfig) error {
	return s.jobScheduler.AddJob(config)
}

// EnableJob 启用任务
func (s *SchedulerService) EnableJob(jobID string) error {
	return s.jobScheduler.EnableJob(jobID)
}

// DisableJob 禁用任务
func (s *SchedulerService) DisableJob(jobID string) error {
	return s.jobScheduler.DisableJob(jobID)
}
