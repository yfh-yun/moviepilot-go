package download

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// QueueService 下载队列服务
type QueueService interface {
	// Add 添加到队列
	Add(ctx context.Context, task *DownloadTask) error

	// Remove 从队列移除
	Remove(ctx context.Context, taskID string) error

	// Get 获取任务
	Get(ctx context.Context, taskID string) (*DownloadTask, error)

	// List 列出队列中的任务
	List(ctx context.Context, status string) ([]*DownloadTask, error)

	// UpdateStatus 更新任务状态
	UpdateStatus(ctx context.Context, taskID string, status TaskStatus) error

	// GetNext 获取下一个待下载任务
	GetNext(ctx context.Context) (*DownloadTask, error)

	// GetQueueStats 获取队列统计
	GetQueueStats(ctx context.Context) (*QueueStats, error)
}

// queueService 队列服务实现
type queueService struct {
	tasks  map[string]*DownloadTask
	mutex  sync.RWMutex
	logger *zap.Logger
}

// NewQueueService 创建队列服务
func NewQueueService() QueueService {
	return &queueService{
		tasks:  make(map[string]*DownloadTask),
		logger: logger.GetLogger(),
	}
}

// DownloadTask 下载任务
type DownloadTask struct {
	ID           string     `json:"id"`
	Hash         string     `json:"hash"`
	Title        string     `json:"title"`
	Size         int64      `json:"size"`
	Status       TaskStatus `json:"status"`
	Priority     int        `json:"priority"`      // 优先级：1-10，数字越大优先级越高
	Progress     float64    `json:"progress"`      // 进度：0-100
	DownloadRate int64      `json:"download_rate"` // 下载速度（字节/秒）
	UploadRate   int64      `json:"upload_rate"`   // 上传速度（字节/秒）
	Downloaded   int64      `json:"downloaded"`    // 已下载大小
	Uploaded     int64      `json:"uploaded"`      // 已上传大小
	Seeders      int        `json:"seeders"`
	Leechers     int        `json:"leechers"`
	SavePath     string     `json:"save_path"`
	Category     string     `json:"category"`
	Tags         []string   `json:"tags"`
	AddedAt      time.Time  `json:"added_at"`
	StartedAt    *time.Time `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at"`
	ErrorMsg     string     `json:"error_msg"`
}

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending     TaskStatus = "pending"     // 等待中
	TaskStatusQueued      TaskStatus = "queued"      // 队列中
	TaskStatusDownloading TaskStatus = "downloading" // 下载中
	TaskStatusPaused      TaskStatus = "paused"      // 已暂停
	TaskStatusCompleted   TaskStatus = "completed"   // 已完成
	TaskStatusFailed      TaskStatus = "failed"      // 失败
	TaskStatusSeeding     TaskStatus = "seeding"     // 做种中
)

// QueueStats 队列统计
type QueueStats struct {
	Total        int   `json:"total"`
	Pending      int   `json:"pending"`
	Queued       int   `json:"queued"`
	Downloading  int   `json:"downloading"`
	Paused       int   `json:"paused"`
	Completed    int   `json:"completed"`
	Failed       int   `json:"failed"`
	Seeding      int   `json:"seeding"`
	TotalSize    int64 `json:"total_size"`
	DownloadRate int64 `json:"download_rate"`
	UploadRate   int64 `json:"upload_rate"`
}

// Add 添加到队列
func (s *queueService) Add(ctx context.Context, task *DownloadTask) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.logger.Info("添加下载任务到队列",
		zap.String("id", task.ID),
		zap.String("title", task.Title),
		zap.Int("priority", task.Priority),
	)

	if task.AddedAt.IsZero() {
		task.AddedAt = time.Now()
	}

	if task.Status == "" {
		task.Status = TaskStatusQueued
	}

	s.tasks[task.ID] = task

	return nil
}

// Remove 从队列移除
func (s *queueService) Remove(ctx context.Context, taskID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.logger.Info("从队列移除任务", zap.String("id", taskID))

	delete(s.tasks, taskID)

	return nil
}

// Get 获取任务
func (s *queueService) Get(ctx context.Context, taskID string) (*DownloadTask, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	task, ok := s.tasks[taskID]
	if !ok {
		return nil, nil
	}

	return task, nil
}

// List 列出队列中的任务
func (s *queueService) List(ctx context.Context, status string) ([]*DownloadTask, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	tasks := make([]*DownloadTask, 0)

	for _, task := range s.tasks {
		if status == "" || string(task.Status) == status {
			tasks = append(tasks, task)
		}
	}

	return tasks, nil
}

// UpdateStatus 更新任务状态
func (s *queueService) UpdateStatus(ctx context.Context, taskID string, status TaskStatus) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	task, ok := s.tasks[taskID]
	if !ok {
		return nil
	}

	s.logger.Info("更新任务状态",
		zap.String("id", taskID),
		zap.String("old_status", string(task.Status)),
		zap.String("new_status", string(status)),
	)

	task.Status = status

	if status == TaskStatusDownloading && task.StartedAt == nil {
		now := time.Now()
		task.StartedAt = &now
	}

	if status == TaskStatusCompleted && task.CompletedAt == nil {
		now := time.Now()
		task.CompletedAt = &now
		task.Progress = 100.0
	}

	return nil
}

// GetNext 获取下一个待下载任务
func (s *queueService) GetNext(ctx context.Context) (*DownloadTask, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var nextTask *DownloadTask
	maxPriority := -1

	for _, task := range s.tasks {
		if task.Status == TaskStatusQueued || task.Status == TaskStatusPending {
			if task.Priority > maxPriority {
				maxPriority = task.Priority
				nextTask = task
			}
		}
	}

	return nextTask, nil
}

// GetQueueStats 获取队列统计
func (s *queueService) GetQueueStats(ctx context.Context) (*QueueStats, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	stats := &QueueStats{}

	for _, task := range s.tasks {
		stats.Total++
		stats.TotalSize += task.Size

		switch task.Status {
		case TaskStatusPending:
			stats.Pending++
		case TaskStatusQueued:
			stats.Queued++
		case TaskStatusDownloading:
			stats.Downloading++
			stats.DownloadRate += task.DownloadRate
			stats.UploadRate += task.UploadRate
		case TaskStatusPaused:
			stats.Paused++
		case TaskStatusCompleted:
			stats.Completed++
		case TaskStatusFailed:
			stats.Failed++
		case TaskStatusSeeding:
			stats.Seeding++
			stats.UploadRate += task.UploadRate
		}
	}

	return stats, nil
}
