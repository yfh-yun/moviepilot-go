// Package actions 提供动作系统的业务逻辑实现
package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"

	"go.uber.org/zap"
)

// DownloadStatusMonitor 下载状态监控器
// 提供实时下载状态监控、批量状态查询和工作流集成功能
type DownloadStatusMonitor struct {
	downloadService DownloadStatusService
	cache           *WorkflowCache
	logger          *zap.Logger
	mutex           sync.RWMutex

	// 状态缓存
	statusCache    map[string]*DownloadStatus
	cacheTimestamp map[string]time.Time
	cacheMutex     sync.RWMutex
}

// DownloadStatusService 下载状态服务接口
type DownloadStatusService interface {
	GetDownloadStatus(ctx context.Context, downloadID string) (*DownloadStatus, error)
	ListDownloadStatuses(ctx context.Context, downloadIDs []string) ([]*DownloadStatus, error)
	GetDownloadProgress(ctx context.Context, downloadID string) (*DownloadProgress, error)
}

// NewDownloadStatusMonitor 创建下载状态监控器实例
func NewDownloadStatusMonitor(
	downloadService DownloadStatusService,
	cache *WorkflowCache,
) *DownloadStatusMonitor {
	return &DownloadStatusMonitor{
		downloadService: downloadService,
		cache:           cache,
		logger:          logger.NewLogger("download_status_monitor"),
		statusCache:     make(map[string]*DownloadStatus),
		cacheTimestamp:  make(map[string]time.Time),
	}
}

// FetchDownloadsAction 获取下载任务动作
// 实现Python项目中的fetch_downloads.py功能
type FetchDownloadsAction struct {
	monitor *DownloadStatusMonitor
	logger  *zap.Logger
}

// NewFetchDownloadsAction 创建获取下载任务动作实例
func NewFetchDownloadsAction(monitor *DownloadStatusMonitor) *FetchDownloadsAction {
	return &FetchDownloadsAction{
		monitor: monitor,
		logger:  logger.NewLogger("fetch_downloads_action"),
	}
}

// Execute 执行获取下载任务动作
func (a *FetchDownloadsAction) Execute(
	ctx context.Context,
	workflowID int64,
	params map[string]interface{},
	actionCtx *ActionContext,
) (*ActionContext, error) {
	a.logger.Info("开始获取下载任务状态",
		zap.Int64("workflow_id", workflowID),
		zap.Int("download_count", len(actionCtx.Downloads)))

	allComplete := true
	var completedDownloads []*DownloadTask

	// 批量获取下载状态
	if len(actionCtx.Downloads) > 0 {
		downloadIDs := make([]string, len(actionCtx.Downloads))
		for i, download := range actionCtx.Downloads {
			downloadIDs[i] = download.DownloadID
		}

		statuses, err := a.monitor.downloadService.ListDownloadStatuses(ctx, downloadIDs)
		if err != nil {
			a.logger.Error("批量获取下载状态失败", zap.Error(err))
			return nil, fmt.Errorf("批量获取下载状态失败: %w", err)
		}

		// 更新下载任务状态
		statusMap := make(map[string]*DownloadStatus)
		for _, status := range statuses {
			statusMap[status.DownloadID] = status
		}

		for _, download := range actionCtx.Downloads {
			status, exists := statusMap[download.DownloadID]
			if !exists {
				a.logger.Warn("未找到下载任务状态", zap.String("download_id", download.DownloadID))
				continue
			}

			// 更新下载任务信息
			download.Path = status.Path
			download.Progress = status.Progress
			download.Speed = status.Speed
			download.ETA = status.ETA
			download.Completed = status.Progress >= 100

			if !download.Completed {
				allComplete = false
			} else {
				completedDownloads = append(completedDownloads, download)
			}

			// 缓存状态信息
			cacheKey := fmt.Sprintf("download_status:%s", download.DownloadID)
			a.monitor.cache.SaveCache(workflowID, cacheKey, status, 30*time.Second)

			a.logger.Info("下载任务状态更新",
				zap.String("download_id", download.DownloadID),
				zap.Float64("progress", status.Progress),
				zap.Bool("completed", download.Completed),
				zap.String("path", status.Path))
		}
	}

	// 检查是否有新增的已完成下载
	if len(completedDownloads) > 0 {
		a.logger.Info("检测到新的已完成下载",
			zap.Int("completed_count", len(completedDownloads)),
			zap.Int64("workflow_id", workflowID))

		// 触发下载完成事件（可选）
		if err := a.triggerDownloadCompletedEvent(ctx, completedDownloads, workflowID); err != nil {
			a.logger.Warn("触发下载完成事件失败", zap.Error(err))
		}
	}

	// 如果所有下载都完成，标记工作流可以继续
	if allComplete && len(actionCtx.Downloads) > 0 {
		a.logger.Info("所有下载任务已完成",
			zap.Int64("workflow_id", workflowID),
			zap.Int("total_downloads", len(actionCtx.Downloads)))

		// 可以在这里设置工作流状态或触发下一步动作
	}

	// 更新上下文中的下载状态
	actionCtx.Variables["downloads_all_complete"] = allComplete
	actionCtx.Variables["downloads_completed_count"] = len(completedDownloads)

	return actionCtx, nil
}

// triggerDownloadCompletedEvent 触发下载完成事件
func (a *FetchDownloadsAction) triggerDownloadCompletedEvent(ctx context.Context, downloads []*DownloadTask, workflowID int64) error {
	// 这里可以实现事件发送逻辑
	// 例如发送到消息队列、触发其他工作流等

	eventData := map[string]interface{}{
		"event_type":     "downloads_completed",
		"workflow_id":    workflowID,
		"download_count": len(downloads),
		"downloads":      downloads,
		"timestamp":      time.Now(),
	}

	eventJSON, _ := json.Marshal(eventData)
	a.logger.Info("触发下载完成事件",
		zap.String("event_data", string(eventJSON)))

	return nil
}

// GetDownloadStatusWithCache 带缓存的获取下载状态
func (m *DownloadStatusMonitor) GetDownloadStatusWithCache(ctx context.Context, downloadID string, workflowID int64) (*DownloadStatus, error) {
	m.cacheMutex.RLock()
	cachedStatus, exists := m.statusCache[downloadID]
	cachedTime, timeExists := m.cacheTimestamp[downloadID]
	m.cacheMutex.RUnlock()

	// 检查缓存是否存在且未过期（30秒）
	if exists && timeExists && time.Since(cachedTime) < 30*time.Second {
		return cachedStatus, nil
	}

	// 从服务获取最新状态
	status, err := m.downloadService.GetDownloadStatus(ctx, downloadID)
	if err != nil {
		return nil, err
	}

	// 更新缓存
	m.cacheMutex.Lock()
	m.statusCache[downloadID] = status
	m.cacheTimestamp[downloadID] = time.Now()
	m.cacheMutex.Unlock()

	// 保存到工作流缓存
	cacheKey := fmt.Sprintf("download_status:%s", downloadID)
	m.cache.SaveCache(workflowID, cacheKey, status, 30*time.Second)

	return status, nil
}

// BatchMonitorDownloads 批量监控下载状态
func (m *DownloadStatusMonitor) BatchMonitorDownloads(
	ctx context.Context,
	downloads []*DownloadTask,
	workflowID int64,
	interval time.Duration,
) (<-chan *DownloadStatusUpdate, error) {
	updateChan := make(chan *DownloadStatusUpdate, len(downloads))

	go func() {
		defer close(updateChan)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// 批量获取状态
				downloadIDs := make([]string, len(downloads))
				for i, download := range downloads {
					downloadIDs[i] = download.DownloadID
				}

				statuses, err := m.downloadService.ListDownloadStatuses(ctx, downloadIDs)
				if err != nil {
					m.logger.Error("批量获取下载状态失败", zap.Error(err))
					continue
				}

				// 发送状态更新
				for _, status := range statuses {
					select {
					case updateChan <- &DownloadStatusUpdate{
						DownloadID: status.DownloadID,
						Status:     status,
						Timestamp:  time.Now(),
					}:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return updateChan, nil
}

// WaitForDownloadCompletion 等待下载完成
func (m *DownloadStatusMonitor) WaitForDownloadCompletion(
	ctx context.Context,
	downloads []*DownloadTask,
	workflowID int64,
	timeout time.Duration,
) error {
	if len(downloads) == 0 {
		return nil
	}

	// 创建超时上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 监控状态更新
	updateChan, err := m.BatchMonitorDownloads(timeoutCtx, downloads, workflowID, 5*time.Second)
	if err != nil {
		return err
	}

	completedCount := 0
	totalCount := len(downloads)

	for {
		select {
		case <-timeoutCtx.Done():
			return fmt.Errorf("等待下载完成超时，已完成 %d/%d", completedCount, totalCount)

		case update, ok := <-updateChan:
			if !ok {
				return fmt.Errorf("下载状态监控通道关闭")
			}

			if update.Status.Progress >= 100 {
				completedCount++
				m.logger.Info("下载任务完成",
					zap.String("download_id", update.DownloadID),
					zap.Int("completed", completedCount),
					zap.Int("total", totalCount))

				if completedCount >= totalCount {
					m.logger.Info("所有下载任务已完成", zap.Int64("workflow_id", workflowID))
					return nil
				}
			}
		}
	}
}

// GetMonitoringStatistics 获取监控统计信息
func (m *DownloadStatusMonitor) GetMonitoringStatistics() *MonitoringStatistics {
	m.cacheMutex.RLock()
	defer m.cacheMutex.RUnlock()

	return &MonitoringStatistics{
		CachedDownloads:  len(m.statusCache),
		LastUpdateTime:   time.Now(),
		ActiveMonitoring: true,
	}
}

// 数据结构定义

// DownloadStatus 下载状态
type DownloadStatus struct {
	DownloadID string    `json:"download_id"`
	Status     string    `json:"status"`     // "downloading", "completed", "paused", "failed"
	Progress   float64   `json:"progress"`   // 0-100
	Speed      float64   `json:"speed"`      // 下载速度 bytes/s
	Path       string    `json:"path"`       // 保存路径
	Size       int64     `json:"size"`       // 文件大小
	Downloaded int64     `json:"downloaded"` // 已下载大小
	ETA        int64     `json:"eta"`        // 预计剩余时间（秒）
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}

// DownloadProgress 下载进度
type DownloadProgress struct {
	DownloadID string  `json:"download_id"`
	Progress   float64 `json:"progress"`
	Speed      float64 `json:"speed"`
	ETA        int64   `json:"eta"`
}

// DownloadStatusUpdate 下载状态更新
type DownloadStatusUpdate struct {
	DownloadID string          `json:"download_id"`
	Status     *DownloadStatus `json:"status"`
	Timestamp  time.Time       `json:"timestamp"`
}

// MonitoringStatistics 监控统计信息
type MonitoringStatistics struct {
	CachedDownloads  int       `json:"cached_downloads"`
	LastUpdateTime   time.Time `json:"last_update_time"`
	ActiveMonitoring bool      `json:"active_monitoring"`
}

// ActionContext 工作流上下文
type ActionContext struct {
	WorkflowID int64                  `json:"workflow_id"`
	Medias     []*MediaInfo           `json:"medias"`
	Torrents   []*TorrentInfo         `json:"torrents"`
	Downloads  []*DownloadTask        `json:"downloads"`
	Subscribes []*Subscribe           `json:"subscribes"`
	Messages   []*Message             `json:"messages"`
	Variables  map[string]interface{} `json:"variables"`
	Cache      *WorkflowCache         `json:"cache"`
}

// DownloadTask 下载任务
type DownloadTask struct {
	DownloadID string  `json:"download_id"`
	Downloader string  `json:"downloader"`
	Path       string  `json:"path"`
	Progress   float64 `json:"progress"`
	Speed      float64 `json:"speed"`
	ETA        int64   `json:"eta"`
	Completed  bool    `json:"completed"`
}

// MediaInfo 媒体信息
type MediaInfo struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Year     int    `json:"year"`
	Type     string `json:"type"`
	Season   int    `json:"season"`
	Episode  int    `json:"episode"`
	TMDBID   int    `json:"tmdb_id"`
	DoubanID int    `json:"douban_id"`
}

// TorrentInfo 种子信息
type TorrentInfo struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Size        int64     `json:"size"`
	Hash        string    `json:"hash"`
	Site        string    `json:"site"`
	PublishTime time.Time `json:"publish_time"`
}

// Subscribe 订阅信息
type Subscribe struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Year     int    `json:"year"`
	Type     string `json:"type"`
	Season   int    `json:"season"`
	Status   string `json:"status"`
	TMDBID   int    `json:"tmdb_id"`
	DoubanID int    `json:"douban_id"`
}

// Message 消息信息
type Message struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Type    string `json:"type"`
	UserID  uint   `json:"user_id"`
}
