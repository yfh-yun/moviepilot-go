package actions

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"moviepilot-go/internal/repositories"
	"moviepilot-go/pkg/logger"
	"moviepilot-go/pkg/errors"
)

var log = logger.NewLogger("clean_torrents")

// CleanTorrentsService 种子清理服务接口
type CleanTorrentsService interface {
	// CleanTorrents 清理种子
	CleanTorrents(ctx context.Context, req CleanTorrentRequest) (*CleanTorrentResponse, error)
	
	// GetCleanTask 获取清理任务状态
	GetCleanTask(ctx context.Context, taskID string) (*CleanTorrentTask, error)
	
	// ListCleanTasks 列出清理任务
	ListCleanTasks(ctx context.Context, limit, offset int) ([]CleanTorrentTask, error)
	
	// CancelCleanTask 取消清理任务
	CancelCleanTask(ctx context.Context, taskID string) error
	
	// GetCleanStats 获取清理统计信息
	GetCleanStats(ctx context.Context) (*CleanTorrentStats, error)
}

// CleanTorrentsServiceImpl 种子清理服务实现
type CleanTorrentsServiceImpl struct {
	taskRepository   repositories.TaskRepository
	downloaderService DownloaderService
	taskMap         sync.Map // 存储进行中的任务
	mutex           sync.RWMutex
}

// DownloaderService 下载器服务接口
type DownloaderService interface {
	// GetTorrents 获取种子列表
	GetTorrents(ctx context.Context, downloaders []string, states []TorrentState) ([]TorrentInfo, error)
	
	// RemoveTorrent 删除种子
	RemoveTorrent(ctx context.Context, downloader, torrentID string, deleteData bool) error
	
	// MarkTorrentForRemoval 标记种子待删除
	MarkTorrentForRemoval(ctx context.Context, downloader, torrentID string) error
}

// NewCleanTorrentsService 创建种子清理服务实例
func NewCleanTorrentsService(taskRepository repositories.TaskRepository, downloaderService DownloaderService) CleanTorrentsService {
	return &CleanTorrentsServiceImpl{
		taskRepository:   taskRepository,
		downloaderService: downloaderService,
	}
}

// CleanTorrents 清理种子
func (s *CleanTorrentsServiceImpl) CleanTorrents(ctx context.Context, req CleanTorrentRequest) (*CleanTorrentResponse, error) {
	// 验证请求参数
	if err := req.Validate(); err != nil {
		log.Error("Invalid clean request", "error", err.Error())
		return nil, err
	}

	// 生成任务ID
	taskID := uuid.New().String()
	
	// 创建任务
	task := &CleanTorrentTask{
		TaskID:      taskID,
		Request:     req,
		Status:      CleanStatusPending,
		StartTime:   time.Now(),
		Progress:    0,
	}

	// 保存任务到内存中
	s.taskMap.Store(taskID, task)

	// 异步执行清理任务
	go s.executeCleanTask(ctx, task)

	return &CleanTorrentResponse{
		TaskID:  taskID,
		Status:  CleanStatusPending,
		Message: "Clean task started",
	}, nil
}

// executeCleanTask 执行清理任务
func (s *CleanTorrentsServiceImpl) executeCleanTask(ctx context.Context, task *CleanTorrentTask) {
	defer func() {
		if r := recover(); r != nil {
			task.Status = CleanStatusFailed
			task.Error = fmt.Sprintf("Panic recovered: %v", r)
			endTime := time.Now()
			task.EndTime = &endTime
			log.Error("Clean task panicked", "task_id", task.TaskID, "error", task.Error)
		}
		// 从内存中移除任务（可选，根据业务需求决定是否保留）
		// s.taskMap.Delete(task.TaskID)
	}()

	task.Status = CleanStatusRunning
	log.Info("Starting clean task", "task_id", task.TaskID)

	// 确定要查询的种子状态
	var states []TorrentState
	if task.Request.IncludeCompleted {
		states = append(states, TorrentStateCompleted)
	}
	if task.Request.IncludeDownloading {
		states = append(states, TorrentStateDownloading)
	}
	if task.Request.IncludePaused {
		states = append(states, TorrentStatePaused)
	}

	// 获取种子列表
	torrents, err := s.downloaderService.GetTorrents(ctx, task.Request.Downloaders, states)
	if err != nil {
		task.Status = CleanStatusFailed
		task.Error = fmt.Sprintf("Failed to get torrents: %v", err)
		endTime := time.Now()
		task.EndTime = &endTime
		log.Error("Failed to get torrents", "task_id", task.TaskID, "error", err.Error())
		return
	}

	task.ProcessedCount = len(torrents)
	log.Info("Retrieved torrents for cleaning", "task_id", task.TaskID, "count", len(torrents))

	// 过滤并清理种子
	for i, torrent := range torrents {
		// 检查是否应该清理该种子
		shouldClean, reason := s.shouldCleanTorrent(torrent, task.Request)
		if shouldClean {
			var cleanErr error
			if task.Request.OnlyMark {
				cleanErr = s.downloaderService.MarkTorrentForRemoval(ctx, torrent.Downloader, torrent.ID)
			} else {
				// 默认不删除数据，只删除种子
				cleanErr = s.downloaderService.RemoveTorrent(ctx, torrent.Downloader, torrent.ID, false)
			}

			if cleanErr != nil {
				task.FailedCount++
				log.Error("Failed to clean torrent", "task_id", task.TaskID, "torrent_id", torrent.ID, "error", cleanErr.Error())
			} else {
				task.CleanedCount++
				log.Info("Torrent cleaned", "task_id", task.TaskID, "torrent_id", torrent.ID, "name", torrent.Name, "reason", reason)
			}
		}

		// 更新进度
		task.Progress = (i + 1) * 100 / len(torrents)

		// 检查任务是否被取消
		if ctx.Err() != nil {
			task.Status = CleanStatusCancelled
			endTime := time.Now()
			task.EndTime = &endTime
			log.Info("Clean task cancelled", "task_id", task.TaskID)
			return
		}
	}

	// 任务完成
	task.Status = CleanStatusCompleted
	endTime := time.Now()
	task.EndTime = &endTime
	log.Info("Clean task completed", "task_id", task.TaskID, "cleaned_count", task.CleanedCount, "failed_count", task.FailedCount)

	// 保存任务到数据库
	// 这里可以调用taskRepository保存任务记录
}

// shouldCleanTorrent 判断是否应该清理种子
func (s *CleanTorrentsServiceImpl) shouldCleanTorrent(torrent TorrentInfo, req CleanTorrentRequest) (bool, string) {
	// 检查排除标签
	for _, tag := range req.ExcludeTags {
		for _, torrentTag := range torrent.Tags {
			if torrentTag == tag {
				return false, "excluded by tag"
			}
		}
	}

	// 检查排除Tracker
	for _, tracker := range req.ExcludeTrackers {
		for _, torrentTracker := range torrent.Trackers {
			if torrentTracker == tracker {
				return false, "excluded by tracker"
			}
		}
	}

	// 根据策略判断是否清理
	switch req.Strategy {
	case CleanStrategyByTime:
		if torrent.SeedingTime > req.TimeThreshold {
			return true, fmt.Sprintf("seeding time (%d hours) exceeds threshold (%d hours)", torrent.SeedingTime, req.TimeThreshold)
		}
	case CleanStrategyByRatio:
		if torrent.Ratio > req.RatioThreshold {
			return true, fmt.Sprintf("ratio (%.2f) exceeds threshold (%.2f)", torrent.Ratio, req.RatioThreshold)
		}
	case CleanStrategyBySeeder:
		if torrent.SeederCount > req.SeederThreshold {
			return true, fmt.Sprintf("seeder count (%d) exceeds threshold (%d)", torrent.SeederCount, req.SeederThreshold)
		}
	case CleanStrategyByStorage:
		// 这里简化处理，实际应该结合存储监控服务
		// 检查种子大小是否超过存储阈值的一部分
		storageThresholdBytes := int64(req.StorageThreshold) * 1024 * 1024 * 1024
		if torrent.Size > storageThresholdBytes/10 { // 简化判断，实际应该更复杂
			return true, fmt.Sprintf("size (%.2f GB) is large and storage needs cleaning", float64(torrent.Size)/(1024*1024*1024))
		}
	}

	return false, ""
}

// GetCleanTask 获取清理任务状态
func (s *CleanTorrentsServiceImpl) GetCleanTask(ctx context.Context, taskID string) (*CleanTorrentTask, error) {
	if taskID == "" {
		return nil, errors.NewValidationError("task_id is required")
	}

	// 先从内存中查找
	if task, found := s.taskMap.Load(taskID); found {
		return task.(*CleanTorrentTask), nil
	}

	// 再从数据库中查找
	// 这里应该调用taskRepository查询历史任务
	return nil, errors.NewNotFoundError("clean task not found")
}

// ListCleanTasks 列出清理任务
func (s *CleanTorrentsServiceImpl) ListCleanTasks(ctx context.Context, limit, offset int) ([]CleanTorrentTask, error) {
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	// 从数据库中查询任务列表
	// 这里应该调用taskRepository查询历史任务
	// 同时从内存中获取进行中的任务并合并

	// 简化实现，返回空列表
	return []CleanTorrentTask{}, nil
}

// CancelCleanTask 取消清理任务
func (s *CleanTorrentsServiceImpl) CancelCleanTask(ctx context.Context, taskID string) error {
	if taskID == "" {
		return errors.NewValidationError("task_id is required")
	}

	// 从内存中查找任务并取消
	if task, found := s.taskMap.Load(taskID); found {
		cleanTask := task.(*CleanTorrentTask)
		if cleanTask.Status == CleanStatusRunning || cleanTask.Status == CleanStatusPending {
			// 这里应该通过context取消任务
			// 由于我们使用的是goroutine，实际应该使用更复杂的任务取消机制
			log.Info("Clean task cancellation requested", "task_id", taskID)
			return nil
		}
		return errors.NewInvalidOperationError("cannot cancel task that is not running or pending")
	}

	return errors.NewNotFoundError("clean task not found")
}

// GetCleanStats 获取清理统计信息
func (s *CleanTorrentsServiceImpl) GetCleanStats(ctx context.Context) (*CleanTorrentStats, error) {
	// 从数据库中统计任务信息
	// 这里应该调用taskRepository进行统计

	// 简化实现
	runningCount := 0
	s.taskMap.Range(func(_, value interface{}) bool {
		task := value.(*CleanTorrentTask)
		if task.Status == CleanStatusRunning {
			runningCount++
		}
		return true
	})

	return &CleanTorrentStats{
		TotalTasks:      0, // 需要从数据库统计
		RunningTasks:    runningCount,
		CompletedTasks:  0, // 需要从数据库统计
		FailedTasks:     0, // 需要从数据库统计
		TotalCleaned:    0, // 需要从数据库统计
		AverageCleanTime: 0, // 需要从数据库统计
	}, nil
}
