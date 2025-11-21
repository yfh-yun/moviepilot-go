// Package download 下载服务实现
package download

import (
	"context"
	"errors"
	"fmt"
	"time"

	"moviepilot-go/internal/business/services"
	"moviepilot-go/pkg/logger"

	"go.uber.org/zap"
)

// 确保Service实现了service.DownloadService接口
var _ service.DownloadService = (*Service)(nil)

// ListDownloads 获取下载任务列表
func (s *Service) ListDownloads(ctx context.Context, params service.ListDownloadsParams) ([]*service.DownloadTask, int64, error) {
	// 获取所有种子
	torrents, err := s.GetTorrents("", params.Status)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get torrents: %w", err)
	}

	// 过滤类型
	if params.Type != "" {
		var filtered []*TorrentStatus
		for _, torrent := range torrents {
			if torrent.MediaInfo != nil && torrent.MediaInfo.Type == params.Type {
				filtered = append(filtered, torrent)
			}
		}
		torrents = filtered
	}

	// 转换为DownloadTask
	var tasks []*service.DownloadTask
	for _, torrent := range torrents {
		task := &service.DownloadTask{
			ID:         torrent.Hash,
			Title:      torrent.Name,
			Type:       "torrent",
			Status:     torrent.State,
			Progress:   torrent.Progress,
			FileSize:   torrent.Size,
			Downloaded: torrent.Downloaded,
			Speed:      torrent.DownloadSpeed,
			ETA:        calculateETA(torrent.Size, torrent.Downloaded, torrent.DownloadSpeed),
			CreatedAt:  time.Now(), // TODO: 从实际数据获取
			UpdatedAt:  time.Now(),
		}
		tasks = append(tasks, task)
	}

	// 分页处理
	total := int64(len(tasks))
	if params.Page > 0 && params.Limit > 0 {
		start := (params.Page - 1) * params.Limit
		end := start + params.Limit
		if start >= len(tasks) {
			return []*service.DownloadTask{}, total, nil
		}
		if end > len(tasks) {
			end = len(tasks)
		}
		tasks = tasks[start:end]
	}

	return tasks, total, nil
}

// GetDownloadDetail 获取下载任务详情
func (s *Service) GetDownloadDetail(ctx context.Context, taskID string) (*service.DownloadTask, error) {
	// 查找种子
	torrents, err := s.GetTorrents("", "")
	if err != nil {
		return nil, fmt.Errorf("failed to get torrents: %w", err)
	}

	for _, torrent := range torrents {
		if torrent.Hash == taskID {
			return &service.DownloadTask{
				ID:         torrent.Hash,
				Title:      torrent.Name,
				Type:       "torrent",
				Status:     torrent.State,
				Progress:   torrent.Progress,
				FileSize:   torrent.Size,
				Downloaded: torrent.Downloaded,
				Speed:      torrent.DownloadSpeed,
				ETA:        calculateETA(torrent.Size, torrent.Downloaded, torrent.DownloadSpeed),
				CreatedAt:  time.Now(), // TODO: 从实际数据获取
				UpdatedAt:  time.Now(),
			}, nil
		}
	}

	return nil, service.ErrDownloadNotFound
}

// CreateDownload 创建下载任务
func (s *Service) CreateDownload(ctx context.Context, params service.CreateDownloadParams) (*service.DownloadTask, error) {
	// 创建添加种子请求
	req := &AddTorrentRequest{
		TorrentURL:     params.URL,
		DownloadDir:    params.SavePath,
		Category:       params.Type,
		DownloaderName: "default", // TODO: 从配置获取默认下载器
		IsPaused:       false,
	}

	// 添加种子
	hash, err := s.AddTorrent(req)
	if err != nil {
		return nil, fmt.Errorf("failed to add torrent: %w", err)
	}

	// 等待一段时间让种子信息更新
	time.Sleep(2 * time.Second)

	// 获取创建的任务详情
	task, err := s.GetDownloadDetail(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get created task: %w", err)
	}

	return task, nil
}

// DeleteDownload 删除下载任务
func (s *Service) DeleteDownload(ctx context.Context, taskID string) error {
	// 查找种子所属的下载器
	torrents, err := s.GetTorrents("", "")
	if err != nil {
		return fmt.Errorf("failed to get torrents: %w", err)
	}

	for _, torrent := range torrents {
		if torrent.Hash == taskID {
			return s.DeleteTorrent(taskID, torrent.Downloader, true)
		}
	}

	return service.ErrDownloadNotFound
}

// PauseDownload 暂停下载任务
func (s *Service) PauseDownload(ctx context.Context, taskID string) error {
	// 查找种子所属的下载器
	torrents, err := s.GetTorrents("", "")
	if err != nil {
		return fmt.Errorf("failed to get torrents: %w", err)
	}

	for _, torrent := range torrents {
		if torrent.Hash == taskID {
			return s.StopTorrent(taskID, torrent.Downloader)
		}
	}

	return service.ErrDownloadNotFound
}

// ResumeDownload 恢复下载任务
func (s *Service) ResumeDownload(ctx context.Context, taskID string) error {
	// 查找种子所属的下载器
	torrents, err := s.GetTorrents("", "")
	if err != nil {
		return fmt.Errorf("failed to get torrents: %w", err)
	}

	for _, torrent := range torrents {
		if torrent.Hash == taskID {
			return s.StartTorrent(taskID, torrent.Downloader)
		}
	}

	return service.ErrDownloadNotFound
}

// GetDownloadStats 获取下载统计信息
func (s *Service) GetDownloadStats(ctx context.Context) (*service.DownloadStats, error) {
	torrents, err := s.GetTorrents("", "")
	if err != nil {
		return nil, fmt.Errorf("failed to get torrents: %w", err)
	}

	stats := &service.DownloadStats{
		TotalTasks:      int64(len(torrents)),
		ActiveTasks:     0,
		CompletedTasks:  0,
		FailedTasks:     0,
		TotalDownloaded: 0,
		TotalSpeed:      0,
	}

	for _, torrent := range torrents {
		stats.TotalDownloaded += torrent.Downloaded
		stats.TotalSpeed += float64(torrent.DownloadSpeed)

		switch torrent.State {
		case "downloading", "uploading":
			stats.ActiveTasks++
		case "completed", "seeding":
			stats.CompletedTasks++
		case "error", "stopped":
			stats.FailedTasks++
		}
	}

	return stats, nil
}

// GetDownloadSpeed 获取下载速度
func (s *Service) GetDownloadSpeed(ctx context.Context) (*service.DownloadSpeed, error) {
	torrents, err := s.GetTorrents("", "")
	if err != nil {
		return nil, fmt.Errorf("failed to get torrents: %w", err)
	}

	speed := &service.DownloadSpeed{
		CurrentSpeed: 0,
		AverageSpeed: 0,
		PeakSpeed:    0,
	}

	var totalSpeed int64
	for _, torrent := range torrents {
		totalSpeed += torrent.DownloadSpeed
		if torrent.DownloadSpeed > speed.PeakSpeed {
			speed.PeakSpeed = torrent.DownloadSpeed
		}
	}

	speed.CurrentSpeed = totalSpeed
	speed.AverageSpeed = float64(totalSpeed) / float64(len(torrents))

	return speed, nil
}

// ClearCompletedDownloads 清理已完成的下载任务
func (s *Service) ClearCompletedDownloads(ctx context.Context) error {
	torrents, err := s.GetTorrents("", "completed")
	if err != nil {
		return fmt.Errorf("failed to get completed torrents: %w", err)
	}

	for _, torrent := range torrents {
		if err := s.DeleteTorrent(torrent.Hash, torrent.Downloader, false); err != nil {
			// 记录错误但继续处理其他任务
			logger.Error("Failed to delete completed torrent", 
				zap.String("hash", torrent.Hash), 
				zap.String("downloader", torrent.Downloader), 
				zap.Error(err))
		}
	}

	return nil
}

// BatchDeleteDownloads 批量删除下载任务
func (s *Service) BatchDeleteDownloads(ctx context.Context, taskIDs []string) error {
	if len(taskIDs) == 0 {
		return errors.New("task IDs cannot be empty")
	}

	torrents, err := s.GetTorrents("", "")
	if err != nil {
		return fmt.Errorf("failed to get torrents: %w", err)
	}

	// 创建hash到下载器的映射
	torrentMap := make(map[string]string)
	for _, torrent := range torrents {
		torrentMap[torrent.Hash] = torrent.Downloader
	}

	// 批量删除
	for _, taskID := range taskIDs {
		downloader, exists := torrentMap[taskID]
		if !exists {
			logger.Warn("Torrent not found for batch delete", zap.String("hash", taskID))
			continue
		}

		if err := s.DeleteTorrent(taskID, downloader, true); err != nil {
			logger.Error("Failed to delete torrent in batch", 
				zap.String("hash", taskID), 
				zap.String("downloader", downloader), 
				zap.Error(err))
		}
	}

	return nil
}

// calculateETA 计算预计完成时间
func calculateETA(fileSize, downloaded, speed int64) string {
	if speed <= 0 {
		return "∞"
	}

	remaining := fileSize - downloaded
	if remaining <= 0 {
		return "已完成"
	}

	seconds := remaining / speed
	if seconds < 60 {
		return fmt.Sprintf("%d秒", seconds)
	}

	minutes := seconds / 60
	if minutes < 60 {
		return fmt.Sprintf("%d分钟", minutes)
	}

	hours := minutes / 60
	if hours < 24 {
		return fmt.Sprintf("%d小时", hours)
	}

	days := hours / 24
	return fmt.Sprintf("%d天", days)
}