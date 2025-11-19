// Package actions 提供增强的下载管理器实现
package actions

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/repository/interfaces"
	"github.com/yfh-yun/moviepilot-go/internal/service/interfaces"
	"github.com/yfh-yun/moviepilot-go/internal/service/torrent"

	"go.uber.org/zap"
)

// EnhancedDownloadManager 增强下载管理器
// 集成种子处理、下载器管理和高级下载策略
type EnhancedDownloadManager struct {
	downloadRepo     interfaces.DownloadRepository
	mediaRepo        interfaces.MediaRepository
	torrentProcessor *torrent.TorrentProcessor
	downloaders      map[string]interfaces.DownloaderClient
	factory          interfaces.DownloaderFactory
	logger           *zap.Logger

	// 配置
	config *DownloadManagerConfig

	// 状态管理
	activeDownloads map[string]*ActiveDownload
	mutex           sync.RWMutex
}

// DownloadManagerConfig 下载管理器配置
type DownloadManagerConfig struct {
	// 种子处理配置
	EnableMagnetSupport    bool          `json:"enable_magnet_support"`
	EnableComplexURLDecode bool          `json:"enable_complex_url_decode"`
	MaxTorrentSize         int64         `json:"max_torrent_size"` // 最大种子文件大小
	DownloadTimeout        time.Duration `json:"download_timeout"` // 下载超时

	// 下载器配置
	DefaultDownloader string                                  `json:"default_downloader"`
	DownloaderConfigs map[string]*interfaces.DownloaderConfig `json:"downloader_configs"`
	FailoverEnabled   bool                                    `json:"failover_enabled"`
	FailoverTimeout   time.Duration                           `json:"failover_timeout"`

	// 速度限制配置
	GlobalDownloadLimit int64 `json:"global_download_limit"` // 全局下载速度限制 (bytes/s)
	GlobalUploadLimit   int64 `json:"global_upload_limit"`   // 全局上传速度限制 (bytes/s)

	// 队列配置
	MaxQueueSize       int           `json:"max_queue_size"`
	MaxConcurrentTasks int           `json:"max_concurrent_tasks"`
	TaskTimeout        time.Duration `json:"task_timeout"`

	// 重试配置
	MaxRetryCount           int           `json:"max_retry_count"`
	RetryInterval           time.Duration `json:"retry_interval"`
	RetryExponentialBackoff bool          `json:"retry_exponential_backoff"`

	// 高级选项
	EnableAutoCategorize bool          `json:"enable_auto_categorize"`
	EnableSmartSeeding   bool          `json:"enable_smart_seeding"`
	EnableHealthCheck    bool          `json:"enable_health_check"`
	HealthCheckInterval  time.Duration `json:"health_check_interval"`
}

// ActiveDownload 活跃下载
type ActiveDownload struct {
	ID               string
	TorrentInfo      *torrent.TorrentData
	DownloadRequest  *interfaces.AddTorrentRequest
	DownloaderClient interfaces.DownloaderClient
	Progress         *DownloadProgress
	StartTime        time.Time
	LastUpdate       time.Time
	RetryCount       int
	Status           string // pending, downloading, completed, failed, paused
	Error            string
	CancelChan       chan struct{}
}

// DownloadProgress 下载进度
type DownloadProgress struct {
	TotalSize      int64   `json:"total_size"`
	DownloadedSize int64   `json:"downloaded_size"`
	UploadSize     int64   `json:"upload_size"`
	DownloadSpeed  int64   `json:"download_speed"`
	UploadSpeed    int64   `json:"upload_speed"`
	Progress       float64 `json:"progress"`
	ETA            int64   `json:"eta"`
	State          string  `json:"state"`
	ConnectedSeeds int     `json:"connected_seeds"`
	ConnectedPeers int     `json:"connected_peers"`
}

// AddDownloadRequest 增强的下载请求
type AddDownloadRequest struct {
	// 基础参数
	*AddDownloadParams

	// 增强参数
	TorrentURL  string                `json:"torrent_url"`
	TorrentData []byte                `json:"torrent_data,omitempty"`
	MediaInfo   *interfaces.MediaInfo `json:"media_info,omitempty"`
	MetaInfo    *interfaces.MetaInfo  `json:"meta_info,omitempty"`
	SiteInfo    *SiteInfo             `json:"site_info,omitempty"`

	// 高级选项
	Priority   int  `json:"priority"`
	Queued     bool `json:"queued"`
	AutoStart  bool `json:"auto_start"`
	Sequential bool `json:"sequential"`
	FirstLast  bool `json:"first_last"`

	// 限速配置
	DownloadLimit int64 `json:"download_limit"`
	UploadLimit   int64 `json:"upload_limit"`

	// 元数据和标签
	Tags           []string          `json:"tags"`
	CustomMetadata map[string]string `json:"custom_metadata"`

	// 回调函数
	OnProgress func(progress *DownloadProgress) `json:"-"`
	OnComplete func(result *DownloadResult)     `json:"-"`
	OnError    func(error error)                `json:"-"`
}

// SiteInfo 站点信息
type SiteInfo struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Domain     string `json:"domain"`
	Proxy      string `json:"proxy"`
	Cookie     string `json:"cookie"`
	UserAgent  string `json:"user_agent"`
	Downloader string `json:"downloader"`
}

// DownloadResult 下载结果
type DownloadResult struct {
	Success        bool                     `json:"success"`
	DownloadID     string                   `json:"download_id"`
	TorrentHash    string                   `json:"torrent_hash"`
	Title          string                   `json:"title"`
	Size           int64                    `json:"size"`
	SavePath       string                   `json:"save_path"`
	Downloader     string                   `json:"downloader"`
	Status         string                   `json:"status"`
	Progress       float64                  `json:"progress"`
	Speed          int64                    `json:"speed"`
	ETA            int64                    `json:"eta"`
	CreateTime     time.Time                `json:"create_time"`
	CompleteTime   *time.Time               `json:"complete_time,omitempty"`
	Error          string                   `json:"error,omitempty"`
	Files          []interfaces.TorrentFile `json:"files,omitempty"`
	Trackers       []string                 `json:"trackers,omitempty"`
	ProcessingTime time.Duration            `json:"processing_time"`
	Metadata       map[string]string        `json:"metadata,omitempty"`
}

// NewEnhancedDownloadManager 创建增强下载管理器
func NewEnhancedDownloadManager(
	downloadRepo interfaces.DownloadRepository,
	mediaRepo interfaces.MediaRepository,
	torrentProcessor *torrent.TorrentProcessor,
	factory interfaces.DownloaderFactory,
	config *DownloadManagerConfig,
) *EnhancedDownloadManager {
	manager := &EnhancedDownloadManager{
		downloadRepo:     downloadRepo,
		mediaRepo:        mediaRepo,
		torrentProcessor: torrentProcessor,
		downloaders:      make(map[string]interfaces.DownloaderClient),
		factory:          factory,
		logger:           logger.Logger,
		config:           config,
		activeDownloads:  make(map[string]*ActiveDownload),
	}

	// 初始化下载器
	go manager.initializeDownloaders()

	// 启动健康检查
	if config.EnableHealthCheck {
		go manager.startHealthCheck()
	}

	return manager
}

// AddDownload 添加下载任务（增强版）
func (edm *EnhancedDownloadManager) AddDownload(ctx context.Context, req *AddDownloadRequest) (*DownloadResult, error) {
	edm.logger.Info("开始添加下载任务",
		zap.String("title", req.MediaInfo.Title),
		zap.String("downloader", req.Downloader),
		zap.String("url", req.TorrentURL))

	startTime := time.Now()

	// 验证请求
	if err := edm.validateRequest(req); err != nil {
		return &DownloadResult{
			Success:        false,
			Error:          err.Error(),
			ProcessingTime: time.Since(startTime),
		}, nil
	}

	// 处理种子URL/数据
	torrentData, err := edm.processTorrent(ctx, req)
	if err != nil {
		return &DownloadResult{
			Success:        false,
			Error:          fmt.Sprintf("种子处理失败: %v", err),
			ProcessingTime: time.Since(startTime),
		}, nil
	}

	// 选择下载器
	downloaderName := edm.selectDownloader(req.Downloader)
	downloaderClient, err := edm.getDownloaderClient(downloaderName)
	if err != nil {
		return &DownloadResult{
			Success:        false,
			Error:          fmt.Sprintf("获取下载器失败: %v", err),
			ProcessingTime: time.Since(startTime),
		}, nil
	}

	// 构建下载请求
	addReq := edm.buildDownloadRequest(req, torrentData)

	// 添加到下载器
	addResp, err := downloaderClient.AddTorrent(ctx, addReq)
	if err != nil {
		// 尝试故障转移
		if edm.config.FailoverEnabled {
			return edm.handleFailover(ctx, req, err)
		}

		return &DownloadResult{
			Success:        false,
			Error:          fmt.Sprintf("添加下载任务失败: %v", err),
			ProcessingTime: time.Since(startTime),
		}, nil
	}

	if !addResp.Success {
		return &DownloadResult{
			Success:        false,
			Error:          addResp.Message,
			ProcessingTime: time.Since(startTime),
		}, nil
	}

	// 创建活跃下载记录
	activeDownload := &ActiveDownload{
		ID:               addResp.Hash,
		TorrentInfo:      torrentData,
		DownloadRequest:  addReq,
		DownloaderClient: downloaderClient,
		Progress:         &DownloadProgress{State: "queued"},
		StartTime:        time.Now(),
		LastUpdate:       time.Now(),
		Status:           "queued",
		CancelChan:       make(chan struct{}),
	}

	edm.mutex.Lock()
	edm.activeDownloads[addResp.Hash] = activeDownload
	edm.mutex.Unlock()

	// 保存下载记录
	downloadRecord := edm.createDownloadRecord(req, torrentData, addResp.Hash)
	if err := edm.downloadRepo.Create(ctx, downloadRecord); err != nil {
		edm.logger.Warn("保存下载记录失败", zap.Error(err))
	}

	// 如果配置为自动开始，启动下载
	if req.AutoStart {
		go edm.startDownload(ctx, activeDownload)
	}

	edm.logger.Info("下载任务添加成功",
		zap.String("hash", addResp.Hash),
		zap.String("title", req.MediaInfo.Title))

	return &DownloadResult{
		Success:        true,
		DownloadID:     addResp.Hash,
		TorrentHash:    addResp.Hash,
		Title:          req.MediaInfo.Title,
		Downloader:     downloaderName,
		Status:         "queued",
		CreateTime:     time.Now(),
		ProcessingTime: time.Since(startTime),
		Metadata:       req.CustomMetadata,
	}, nil
}

// processTorrent 处理种子URL或数据
func (edm *EnhancedDownloadManager) processTorrent(ctx context.Context, req *AddDownloadRequest) (*torrent.TorrentData, error) {
	// 如果已有种子数据，直接使用
	if len(req.TorrentData) > 0 {
		return edm.torrentProcessor.ProcessTorrentURL(ctx, req.TorrentURL, &torrent.ProcessOptions{
			Title:    req.MediaInfo.Title,
			SiteName: req.SiteInfo.Name,
		})
	}

	// 处理种子URL
	if req.TorrentURL != "" {
		return edm.torrentProcessor.ProcessTorrentURL(ctx, req.TorrentURL, &torrent.ProcessOptions{
			Title:     req.MediaInfo.Title,
			UserAgent: req.SiteInfo.UserAgent,
			Cookie:    req.SiteInfo.Cookie,
			Proxy:     req.SiteInfo.Proxy,
			SiteName:  req.SiteInfo.Name,
		})
	}

	return nil, fmt.Errorf("未提供种子URL或数据")
}

// buildDownloadRequest 构建下载器请求
func (edm *EnhancedDownloadManager) buildDownloadRequest(req *AddDownloadRequest, torrentData *torrent.TorrentData) *interfaces.AddTorrentRequest {
	// 确定保存路径
	savePath := req.SavePath
	if savePath == "" {
		savePath = edm.generateSavePath(req.MediaInfo)
	}

	// 构建标签
	tags := req.Labels
	if len(tags) == 0 {
		tags = edm.generateTags(req.MediaInfo, req.SiteInfo)
	}

	addReq := &interfaces.AddTorrentRequest{
		// 种子数据
		URL:     torrentData.URL,
		RawData: torrentData.Data,

		// 下载配置
		SavePath: savePath,
		Category: edm.generateCategory(req.MediaInfo),
		Tags:     tags,
		Priority: req.Priority,

		// 下载策略
		Sequential: req.Sequential,
		FirstLast:  req.FirstLast,
		Paused:     !req.AutoStart,

		// 速度限制
		DownloadLimit: req.DownloadLimit,
		UploadLimit:   req.UploadLimit,

		// 元数据
		Metadata: req.CustomMetadata,
	}

	// 文件选择（如果有媒体信息）
	if req.MediaInfo != nil && req.MediaInfo.Type == "tv" && len(req.MediaInfo.Episodes) > 0 {
		// 这里应该根据集数选择文件
		// 简化实现，选择所有文件
	}

	return addReq
}

// startDownload 启动下载
func (edm *EnhancedDownloadManager) startDownload(ctx context.Context, activeDownload *ActiveDownload) {
	edm.logger.Info("启动下载", zap.String("hash", activeDownload.ID))

	activeDownload.Status = "downloading"
	activeDownload.StartTime = time.Now()

	// 监控下载进度
	go edm.monitorDownloadProgress(ctx, activeDownload)
}

// monitorDownloadProgress 监控下载进度
func (edm *EnhancedDownloadManager) monitorDownloadProgress(ctx context.Context, activeDownload *ActiveDownload) {
	ticker := time.NewTicker(5 * time.Second) // 每5秒更新一次
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-activeDownload.CancelChan:
			return
		case <-ticker.C:
			// 获取下载状态
			torrentInfo, err := activeDownload.DownloaderClient.GetTorrentInfo(ctx, activeDownload.ID)
			if err != nil {
				edm.logger.Warn("获取下载状态失败",
					zap.String("hash", activeDownload.ID),
					zap.Error(err))
				continue
			}

			// 更新进度
			edm.updateProgress(activeDownload, torrentInfo)

			// 检查下载是否完成
			if torrentInfo.State == "completed" || torrentInfo.Progress >= 1.0 {
				edm.onDownloadComplete(activeDownload, torrentInfo)
				return
			}

			// 检查下载是否失败
			if torrentInfo.State == "error" || torrentInfo.State == "missingFiles" {
				edm.onDownloadError(activeDownload, fmt.Errorf("下载失败，状态: %s", torrentInfo.State))
				return
			}
		}
	}
}

// updateProgress 更新下载进度
func (edm *EnhancedDownloadManager) updateProgress(activeDownload *ActiveDownload, torrentInfo *interfaces.TorrentInfo) {
	activeDownload.Progress = &DownloadProgress{
		TotalSize:      torrentInfo.Size,
		DownloadedSize: torrentInfo.Downloaded,
		UploadSize:     torrentInfo.Uploaded,
		DownloadSpeed:  torrentInfo.DownloadSpeed,
		UploadSpeed:    torrentInfo.UploadSpeed,
		Progress:       torrentInfo.Progress,
		ETA:            torrentInfo.ETA,
		State:          torrentInfo.State,
	}

	activeDownload.LastUpdate = time.Now()

	// 执行进度回调（如果设置）
	// 这里可以调用用户提供的回调函数
}

// onDownloadComplete 下载完成处理
func (edm *EnhancedDownloadManager) onDownloadComplete(activeDownload *ActiveDownload, torrentInfo *interfaces.TorrentInfo) {
	edm.logger.Info("下载完成",
		zap.String("hash", activeDownload.ID),
		zap.String("title", activeDownload.TorrentInfo.Title),
		zap.Duration("duration", time.Since(activeDownload.StartTime)))

	activeDownload.Status = "completed"
	activeDownload.Progress.State = "completed"

	// 更新数据库记录
	activeDownload.Progress.Progress = 1.0
	edm.updateDownloadRecord(activeDownload, torrentInfo)

	// 执行完成回调
	// 这里可以调用用户提供的回调函数
}

// onDownloadError 下载错误处理
func (edm *EnhancedDownloadManager) onDownloadError(activeDownload *ActiveDownload, err error) {
	edm.logger.Error("下载失败",
		zap.String("hash", activeDownload.ID),
		zap.String("title", activeDownload.TorrentInfo.Title),
		zap.Error(err))

	activeDownload.Status = "failed"
	activeDownload.Error = err.Error()

	// 检查是否需要重试
	if activeDownload.RetryCount < edm.config.MaxRetryCount {
		activeDownload.RetryCount++
		edm.logger.Info("准备重试下载",
			zap.String("hash", activeDownload.ID),
			zap.Int("retry_count", activeDownload.RetryCount))

		// 计算重试延迟
		retryDelay := edm.config.RetryInterval
		if edm.config.RetryExponentialBackoff {
			retryDelay = time.Duration(int64(retryDelay) * int64(1<<activeDownload.RetryCount))
		}

		time.AfterFunc(retryDelay, func() {
			go edm.retryDownload(activeDownload)
		})
	}

	// 执行错误回调
	// 这里可以调用用户提供的回调函数
}

// retryDownload 重试下载
func (edm *EnhancedDownloadManager) retryDownload(activeDownload *ActiveDownload) {
	ctx := context.Background()

	// 重新添加到下载器
	_, err := activeDownload.DownloaderClient.AddTorrent(ctx, activeDownload.DownloadRequest)
	if err != nil {
		edm.onDownloadError(activeDownload, fmt.Errorf("重试添加失败: %w", err))
		return
	}

	// 重置状态并重新启动
	activeDownload.Status = "downloading"
	activeDownload.Progress.State = "queued"

	go edm.startDownload(ctx, activeDownload)
}

// GetDownloadStatus 获取下载状态
func (edm *EnhancedDownloadManager) GetDownloadStatus(ctx context.Context, downloadID string) (*DownloadResult, error) {
	edm.mutex.RLock()
	activeDownload, exists := edm.activeDownloads[downloadID]
	edm.mutex.RUnlock()

	if !exists {
		// 从数据库查询
		download, err := edm.downloadRepo.GetByID(ctx, downloadID)
		if err != nil {
			return nil, err
		}

		return edm.convertDownloadToResult(download), nil
	}

	// 获取实时状态
	torrentInfo, err := activeDownload.DownloaderClient.GetTorrentInfo(ctx, downloadID)
	if err != nil {
		return nil, err
	}

	edm.updateProgress(activeDownload, torrentInfo)

	return edm.convertActiveDownloadToResult(activeDownload, torrentInfo), nil
}

// ListDownloads 列出所有下载
func (edm *EnhancedDownloadManager) ListDownloads(ctx context.Context, statusFilter string) ([]*DownloadResult, error) {
	downloads, err := edm.downloadRepo.List(ctx, 1, 1000)
	if err != nil {
		return nil, err
	}

	var results []*DownloadResult
	for _, download := range downloads {
		if statusFilter != "" && download.Status != statusFilter {
			continue
		}

		result := edm.convertDownloadToResult(download)

		// 如果是活跃下载，更新实时状态
		edm.mutex.RLock()
		if activeDownload, exists := edm.activeDownloads[download.ID]; exists {
			torrentInfo, _ := activeDownload.DownloaderClient.GetTorrentInfo(ctx, download.ID)
			if torrentInfo != nil {
				result = edm.convertActiveDownloadToResult(activeDownload, torrentInfo)
			}
		}
		edm.mutex.RUnlock()

		results = append(results, result)
	}

	return results, nil
}

// CancelDownload 取消下载
func (edm *EnhancedDownloadManager) CancelDownload(ctx context.Context, downloadID string) error {
	edm.mutex.Lock()
	defer edm.mutex.Unlock()

	activeDownload, exists := edm.activeDownloads[downloadID]
	if !exists {
		return fmt.Errorf("下载任务不存在: %s", downloadID)
	}

	// 发送取消信号
	close(activeDownload.CancelChan)

	// 从下载器删除
	err := activeDownload.DownloaderClient.RemoveTorrent(ctx, downloadID)
	if err != nil {
		edm.logger.Warn("从下载器删除任务失败", zap.Error(err))
	}

	// 更新状态
	activeDownload.Status = "cancelled"

	// 从活跃列表移除
	delete(edm.activeDownloads, downloadID)

	edm.logger.Info("下载任务已取消", zap.String("download_id", downloadID))
	return nil
}

// PauseDownload 暂停下载
func (edm *EnhancedDownloadManager) PauseDownload(ctx context.Context, downloadID string) error {
	return edm.executeDownloadAction(ctx, downloadID, "pause")
}

// ResumeDownload 恢复下载
func (edm *EnhancedDownloadManager) ResumeDownload(ctx context.Context, downloadID string) error {
	return edm.executeDownloadAction(ctx, downloadID, "resume")
}

// executeDownloadAction 执行下载动作
func (edm *EnhancedDownloadManager) executeDownloadAction(ctx context.Context, downloadID string, action string) error {
	edm.mutex.RLock()
	activeDownload, exists := edm.activeDownloads[downloadID]
	edm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("下载任务不存在: %s", downloadID)
	}

	var err error
	switch action {
	case "pause":
		err = activeDownload.DownloaderClient.PauseTorrent(ctx, downloadID)
		if err == nil {
			activeDownload.Status = "paused"
		}
	case "resume":
		err = activeDownload.DownloaderClient.ResumeTorrent(ctx, downloadID)
		if err == nil {
			activeDownload.Status = "downloading"
		}
	default:
		return fmt.Errorf("不支持的动作: %s", action)
	}

	return err
}

// 辅助方法

// validateRequest 验证请求
func (edm *EnhancedDownloadManager) validateRequest(req *AddDownloadRequest) error {
	if req.TorrentURL == "" && len(req.TorrentData) == 0 {
		return fmt.Errorf("必须提供种子URL或数据")
	}

	if req.MediaInfo == nil || req.MediaInfo.Title == "" {
		return fmt.Errorf("必须提供媒体信息")
	}

	return nil
}

// selectDownloader 选择下载器
func (edm *EnhancedDownloadManager) selectDownloader(preferred string) string {
	if preferred != "" {
		// 检查指定的下载器是否可用
		if _, exists := edm.config.DownloaderConfigs[preferred]; exists {
			return preferred
		}
	}

	// 返回默认下载器
	return edm.config.DefaultDownloader
}

// getDownloaderClient 获取下载器客户端
func (edm *EnhancedDownloadManager) getDownloaderClient(name string) (interfaces.DownloaderClient, error) {
	edm.mutex.RLock()
	client, exists := edm.downloaders[name]
	edm.mutex.RUnlock()

	if exists {
		return client, nil
	}

	// 创建新的客户端
	config, exists := edm.config.DownloaderConfigs[name]
	if !exists {
		return nil, fmt.Errorf("下载器配置不存在: %s", name)
	}

	client, err := edm.factory.CreateClient(config)
	if err != nil {
		return nil, err
	}

	// 启动客户端
	if err := client.Start(); err != nil {
		return nil, fmt.Errorf("启动下载器失败: %w", err)
	}

	edm.mutex.Lock()
	edm.downloaders[name] = client
	edm.mutex.Unlock()

	return client, nil
}

// handleFailover 处理故障转移
func (edm *EnhancedDownloadManager) handleFailover(ctx context.Context, req *AddDownloadRequest, originalError error) (*DownloadResult, error) {
	edm.logger.Info("启动故障转移", zap.Error(originalError))

	// 尝试其他可用的下载器
	for downloaderName := range edm.config.DownloaderConfigs {
		if downloaderName == req.Downloader {
			continue // 跳过失败的下载器
		}

		req.Downloader = downloaderName
		result, err := edm.AddDownload(ctx, req)
		if err == nil && result.Success {
			edm.logger.Info("故障转移成功", zap.String("new_downloader", downloaderName))
			return result, nil
		}
	}

	return &DownloadResult{
		Success: false,
		Error:   fmt.Sprintf("所有下载器都失败，原始错误: %v", originalError),
	}, nil
}

// generateSavePath 生成保存路径
func (edm *EnhancedDownloadManager) generateSavePath(mediaInfo *interfaces.MediaInfo) string {
	// 根据媒体类型生成默认路径
	switch mediaInfo.Type {
	case "movie":
		return "/downloads/movies"
	case "tv":
		return "/downloads/tv"
	case "anime":
		return "/downloads/anime"
	default:
		return "/downloads/others"
	}
}

// generateCategory 生成分类
func (edm *EnhancedDownloadManager) generateCategory(mediaInfo *interfaces.MediaInfo) string {
	if edm.config.EnableAutoCategorize {
		return mediaInfo.Category
	}
	return ""
}

// generateTags 生成标签
func (edm *EnhancedDownloadManager) generateTags(mediaInfo *interfaces.MediaInfo, siteInfo *SiteInfo) []string {
	var tags []string

	// 添加媒体类型标签
	if mediaInfo.Type != "" {
		tags = append(tags, mediaInfo.Type)
	}

	// 添加站点标签
	if siteInfo != nil && siteInfo.Name != "" {
		tags = append(tags, siteInfo.Name)
	}

	// 添加年份标签
	if mediaInfo.Year > 0 {
		tags = append(tags, fmt.Sprintf("%d", mediaInfo.Year))
	}

	return tags
}

// initializeDownloaders 初始化下载器
func (edm *EnhancedDownloadManager) initializeDownloaders() {
	for name, config := range edm.config.DownloaderConfigs {
		if _, err := edm.getDownloaderClient(name); err != nil {
			edm.logger.Error("初始化下载器失败",
				zap.String("name", name),
				zap.Error(err))
		}
	}
}

// startHealthCheck 启动健康检查
func (edm *EnhancedDownloadManager) startHealthCheck() {
	ticker := time.NewTicker(edm.config.HealthCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		edm.mutex.RLock()
		activeDownloads := make([]*ActiveDownload, 0, len(edm.activeDownloads))
		for _, download := range edm.activeDownloads {
			activeDownloads = append(activeDownloads, download)
		}
		edm.mutex.RUnlock()

		for _, download := range activeDownloads {
			if download.Status == "downloading" || download.Status == "paused" {
				// 检查下载器是否健康
				if err := download.DownloaderClient.GetStatus(); err != nil {
					edm.logger.Warn("下载器健康检查失败",
						zap.String("downloader", download.DownloaderClient.(fmt.Stringer).String()),
						zap.String("hash", download.ID),
						zap.Error(err))
				}
			}
		}
	}
}

// 转换方法
func (edm *EnhancedDownloadManager) createDownloadRecord(req *AddDownloadRequest, torrentData *torrent.TorrentData, hash string) *interfaces.Download {
	return &interfaces.Download{
		ID:         hash,
		Title:      req.MediaInfo.Title,
		URL:        req.TorrentURL,
		Hash:       torrentData.Hash,
		Size:       torrentData.Size,
		Type:       req.MediaInfo.Type,
		Season:     req.MediaInfo.Season,
		Episodes:   req.MediaInfo.Episodes,
		Downloader: req.Downloader,
		SavePath:   edm.generateSavePath(req.MediaInfo),
		Labels:     req.Labels,
		Status:     "queued",
		SiteID:     fmt.Sprintf("%d", req.SiteInfo.ID),
		SiteName:   req.SiteInfo.Name,
		MediaID:    req.MediaInfo.ID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

func (edm *EnhancedDownloadManager) updateDownloadRecord(activeDownload *ActiveDownload, torrentInfo *interfaces.TorrentInfo) {
	// 更新数据库记录
	ctx := context.Background()
	download := &interfaces.Download{
		ID:        activeDownload.ID,
		Status:    activeDownload.Status,
		Progress:  activeDownload.Progress.Progress,
		Speed:     activeDownload.Progress.DownloadSpeed,
		UpdatedAt: time.Now(),
	}

	if torrentInfo.State == "completed" {
		download.Status = "completed"
		download.Progress = 1.0
	}

	if err := edm.downloadRepo.Update(ctx, download); err != nil {
		edm.logger.Warn("更新下载记录失败", zap.Error(err))
	}
}

func (edm *EnhancedDownloadManager) convertDownloadToResult(download *interfaces.Download) *DownloadResult {
	return &DownloadResult{
		DownloadID: download.ID,
		Title:      download.Title,
		Size:       download.Size,
		SavePath:   download.SavePath,
		Downloader: download.Downloader,
		Status:     download.Status,
		Progress:   download.Progress,
		Speed:      download.Speed,
		CreateTime: download.CreatedAt,
		Metadata: map[string]string{
			"site_id":   download.SiteID,
			"site_name": download.SiteName,
			"hash":      download.Hash,
		},
	}
}

func (edm *EnhancedDownloadManager) convertActiveDownloadToResult(activeDownload *ActiveDownload, torrentInfo *interfaces.TorrentInfo) *DownloadResult {
	result := &DownloadResult{
		DownloadID: activeDownload.ID,
		Title:      activeDownload.TorrentInfo.Title,
		Size:       torrentInfo.Size,
		SavePath:   activeDownload.DownloadRequest.SavePath,
		Downloader: "qbittorrent", // 暂时硬编码
		Status:     activeDownload.Status,
		Progress:   activeDownload.Progress.Progress,
		Speed:      activeDownload.Progress.DownloadSpeed,
		ETA:        activeDownload.Progress.ETA,
		CreateTime: activeDownload.StartTime,
		Files:      torrentInfo.Files,
		Trackers:   torrentInfo.Trackers,
		Metadata:   activeDownload.DownloadRequest.Metadata,
	}

	if torrentInfo.State == "completed" {
		completedTime := time.Now()
		result.CompleteTime = &completedTime
	}

	if activeDownload.Status == "failed" {
		result.Error = activeDownload.Error
	}

	return result
}
