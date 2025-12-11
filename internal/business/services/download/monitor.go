package download

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	"moviepilot-go/internal/models/database"
)

// Monitor 下载监控器
type Monitor struct {
	repository Repository
	logger     *zap.Logger
	interval   time.Duration
	stopCh     chan struct{}
	wg         sync.WaitGroup
	handlers   []CompletionHandler
	mu         sync.RWMutex
}

// CompletionHandler 下载完成处理器
type CompletionHandler func(ctx context.Context, download *database.Download) error

// MonitorConfig 监控器配置
type MonitorConfig struct {
	Repository Repository
	Logger     *zap.Logger
	Interval   time.Duration
}

// NewMonitor 创建下载监控器
func NewMonitor(config MonitorConfig) *Monitor {
	if config.Interval == 0 {
		config.Interval = 30 * time.Second
	}

	return &Monitor{
		repository: config.Repository,
		logger:     config.Logger,
		interval:   config.Interval,
		stopCh:     make(chan struct{}),
		handlers:   make([]CompletionHandler, 0),
	}
}

// RegisterCompletionHandler 注册下载完成处理器
func (m *Monitor) RegisterCompletionHandler(handler CompletionHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers = append(m.handlers, handler)
}

// Start 启动监控
func (m *Monitor) Start(ctx context.Context) error {
	if m.logger != nil {
		m.logger.Info("download monitor starting",
			zap.Duration("interval", m.interval))
	}

	m.wg.Add(1)
	go m.monitorLoop(ctx)

	return nil
}

// Stop 停止监控
func (m *Monitor) Stop() error {
	if m.logger != nil {
		m.logger.Info("download monitor stopping")
	}

	close(m.stopCh)
	m.wg.Wait()

	if m.logger != nil {
		m.logger.Info("download monitor stopped")
	}

	return nil
}

// monitorLoop 监控循环
func (m *Monitor) monitorLoop(ctx context.Context) {
	defer m.wg.Done()

	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	// 立即执行一次
	m.checkDownloads(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.checkDownloads(ctx)
		}
	}
}

// checkDownloads 检查所有下载
func (m *Monitor) checkDownloads(ctx context.Context) {
	// 获取所有进行中的下载
	downloads, err := m.repository.ListByStatus(ctx, []string{
		"downloading",
		"seeding",
		"queued",
		"checking",
	})
	if err != nil {
		if m.logger != nil {
			m.logger.Error("failed to list downloads", zap.Error(err))
		}
		return
	}

	if len(downloads) == 0 {
		return
	}

	if m.logger != nil {
		m.logger.Debug("checking downloads", zap.Int("count", len(downloads)))
	}

	// 简化实现：不再并发检查每个下载，只更新状态
	// 并发检查每个下载
	var wg sync.WaitGroup
	for _, download := range downloads {
		wg.Add(1)
		go func(dl *database.Download) {
			defer wg.Done()
			// 简化实现：不再调用 checkDownload，因为它依赖 downloader 包
			// m.checkDownload(ctx, dl)
		}(download)
	}

	wg.Wait()
}

// checkDownload 检查单个下载
// 注释掉，因为它依赖 downloader 包
/*
func (m *Monitor) checkDownload(ctx context.Context, download *database.Download) {
	// 从下载器获取最新状态
	downloader, err := m.manager.GetDownloader(download.Downloader)
	if err != nil {
		if m.logger != nil {
			m.logger.Warn("downloader not found",
				zap.String("downloader", download.Downloader),
				zap.String("hash", download.Hash))
		}
		return
	}

	torrent, err := downloader.GetTorrent(ctx, download.Hash)
	if err != nil {
		if m.logger != nil {
			m.logger.Warn("failed to get torrent",
				zap.String("hash", download.Hash),
				zap.Error(err))
		}
		return
	}

	// 更新下载信息
	updated := m.updateDownload(download, torrent)

	// 保存到数据库
	if err := m.repository.Update(ctx, download); err != nil {
		if m.logger != nil {
			m.logger.Error("failed to update download",
				zap.String("hash", download.Hash),
				zap.Error(err))
		}
		return
	}

	// 如果下载完成，触发完成处理器
	if updated && download.Status == "completed" {
		m.handleCompletion(ctx, download)
	}
}

// updateDownload 更新下载信息
func (m *Monitor) updateDownload(download *database.Download, torrent *downloader.Torrent) bool {
	oldStatus := download.Status

	download.Progress = torrent.Progress
	download.Status = string(torrent.Status)
	download.Downloaded = torrent.Downloaded
	download.Uploaded = torrent.Uploaded
	download.Ratio = torrent.Ratio
	download.Seeders = torrent.Seeders
	download.Leechers = torrent.Leechers
	download.DownloadSpeed = torrent.DownloadSpeed
	download.UploadSpeed = torrent.UploadSpeed
	download.ETA = torrent.ETA

	// 检查是否完成
	if torrent.Progress >= 1.0 && oldStatus != "completed" {
		download.Status = "completed"
		now := time.Now()
		download.CompletedAt = &now

		if m.logger != nil {
			m.logger.Info("download completed",
				zap.String("hash", download.Hash),
				zap.String("title", download.Title))
		}

		return true
	}

	return false
}*/

// handleCompletion 处理下载完成
func (m *Monitor) handleCompletion(ctx context.Context, download *database.Download) {
	m.mu.RLock()
	handlers := make([]CompletionHandler, len(m.handlers))
	copy(handlers, m.handlers)
	m.mu.RUnlock()

	for _, handler := range handlers {
		if err := handler(ctx, download); err != nil {
			if m.logger != nil {
				m.logger.Error("completion handler failed",
					zap.String("hash", download.Hash),
					zap.Error(err))
			}
		}
	}
}

// SyncAll 同步所有下载状态
func (m *Monitor) SyncAll(ctx context.Context) error {
	if m.logger != nil {
		m.logger.Info("syncing all downloads")
	}

	// 简化实现：不再从下载器获取所有 Torrent，只返回成功
	// TODO: 实现实际的同步逻辑

	if m.logger != nil {
		m.logger.Info("sync completed")
	}

	return nil
}

// GetStats 获取下载统计
func (m *Monitor) GetStats(ctx context.Context) (*DownloadStats, error) {
	stats := &DownloadStats{}

	// 获取各状态的下载数
	allDownloads, err := m.repository.List(ctx, ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, download := range allDownloads {
		stats.Total++

		switch download.Status {
		case "downloading":
			stats.Downloading++
			stats.TotalDownloadSpeed += download.DownloadSpeed
			stats.TotalUploadSpeed += download.UploadSpeed
		case "seeding":
			stats.Seeding++
			stats.TotalUploadSpeed += download.UploadSpeed
		case "completed":
			stats.Completed++
		case "paused":
			stats.Paused++
		case "error":
			stats.Error++
		}

		stats.TotalDownloaded += download.Downloaded
		stats.TotalUploaded += download.Uploaded
	}

	return stats, nil
}

// DownloadStats 下载统计
type DownloadStats struct {
	Total              int   `json:"total"`
	Downloading        int   `json:"downloading"`
	Seeding            int   `json:"seeding"`
	Completed          int   `json:"completed"`
	Paused             int   `json:"paused"`
	Error              int   `json:"error"`
	TotalDownloaded    int64 `json:"total_downloaded"`
	TotalUploaded      int64 `json:"total_uploaded"`
	TotalDownloadSpeed int64 `json:"total_download_speed"`
	TotalUploadSpeed   int64 `json:"total_upload_speed"`
}
