// Package actions 提供动作系统的业务逻辑实现
package actions

import (
	"context"
	"fmt"
	"time"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/repository/interfaces"
	"github.com/yfh-yun/moviepilot-go/internal/service/actions/types"

	"go.uber.org/zap"
)

// DownloadManager 下载管理器
// 负责处理下载任务的创建、管理和状态跟踪
type DownloadManager struct {
	downloadRepo interfaces.DownloadRepository
	mediaRepo    interfaces.MediaRepository
	cache        *WorkflowCache
	logger       *zap.Logger
}

// AddDownloadParams 添加下载参数
type AddDownloadParams struct {
	Downloader string   `json:"downloader" description:"下载器"`
	SavePath   string   `json:"save_path" description:"保存路径"`
	Labels     []string `json:"labels" description:"标签列表"`
	OnlyLack   bool     `json:"only_lack" description:"仅下载缺失的资源"`
	Sites      []int    `json:"sites" description:"指定站点ID列表"`
	Quality    string   `json:"quality" description:"质量要求"`
	Resolution string   `json:"resolution" description:"分辨率要求"`
}

// DownloadTask 下载任务
type DownloadTask struct {
	DownloadID string            `json:"download_id"`
	Downloader string            `json:"downloader"`
	Title      string            `json:"title"`
	Size       int64             `json:"size"`
	SavePath   string            `json:"save_path"`
	Labels     []string          `json:"labels"`
	Status     string            `json:"status"`
	Progress   float64           `json:"progress"`
	Speed      int64             `json:"speed"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
	Metadata   map[string]string `json:"metadata"`
}

// DownloadResult 下载结果
type DownloadResult struct {
	Success        bool          `json:"success"`
	DownloadID     string        `json:"download_id"`
	Message        string        `json:"message"`
	Task           *DownloadTask `json:"task,omitempty"`
	Error          error         `json:"error,omitempty"`
	ProcessingTime time.Duration `json:"processing_time"`
}

// NewDownloadManager 创建下载管理器实例
func NewDownloadManager(
	downloadRepo interfaces.DownloadRepository,
	mediaRepo interfaces.MediaRepository,
	cache *WorkflowCache,
) *DownloadManager {
	return &DownloadManager{
		downloadRepo: downloadRepo,
		mediaRepo:    mediaRepo,
		cache:        cache,
		logger:       logger.Logger,
	}
}

// AddDownload 添加下载任务
// 实现Python版本AddDownloadAction的完整功能
func (dm *DownloadManager) AddDownload(
	ctx context.Context,
	workflowID int64,
	params *AddDownloadParams,
	torrents []*types.TorrentInfo,
) ([]*DownloadResult, error) {
	startTime := time.Now()
	results := make([]*DownloadResult, 0, len(torrents))

	dm.logger.Info("开始添加下载任务",
		zap.Int64("workflow_id", workflowID),
		zap.Int("torrent_count", len(torrents)),
		zap.Strings("sites", dm.intSliceToStringSlice(params.Sites)),
	)

	for _, torrent := range torrents {
		// 检查工作流是否已停止
		if dm.isWorkflowStopped(ctx, workflowID) {
			dm.logger.Info("工作流已停止，终止下载任务添加", zap.Int64("workflow_id", workflowID))
			break
		}

		result := dm.processSingleTorrent(ctx, workflowID, params, torrent)
		results = append(results, result)

		if result.Success {
			dm.logger.Info("下载任务添加成功",
				zap.String("download_id", result.DownloadID),
				zap.String("title", torrent.Title),
			)
		} else {
			dm.logger.Warn("下载任务添加失败",
				zap.String("title", torrent.Title),
				zap.Error(result.Error),
			)
		}
	}

	// 统计结果
	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		}
	}

	dm.logger.Info("下载任务添加完成",
		zap.Int64("workflow_id", workflowID),
		zap.Int("total", len(torrents)),
		zap.Int("success", successCount),
		zap.Duration("processing_time", time.Since(startTime)),
	)

	return results, nil
}

// processSingleTorrent 处理单个种子文件
func (dm *DownloadManager) processSingleTorrent(
	ctx context.Context,
	workflowID int64,
	params *AddDownloadParams,
	torrent *types.TorrentInfo,
) *DownloadResult {
	startTime := time.Now()

	// 生成缓存键
	cacheKey := fmt.Sprintf("%s-%s", torrent.SiteID, torrent.Title)

	// 检查缓存
	if dm.checkCache(ctx, workflowID, cacheKey) {
		return &DownloadResult{
			Success:        false,
			Message:        fmt.Sprintf("%s 已添加过下载，跳过", torrent.Title),
			ProcessingTime: time.Since(startTime),
		}
	}

	// 识别媒体信息
	mediaInfo, err := dm.recognizeMedia(ctx, torrent)
	if err != nil {
		return &DownloadResult{
			Success:        false,
			Message:        fmt.Sprintf("%s 未识别到媒体信息，无法下载", torrent.Title),
			Error:          err,
			ProcessingTime: time.Since(startTime),
		}
	}

	// 检查是否仅下载缺失资源
	if params.OnlyLack {
		exists, err := dm.checkMediaExists(ctx, mediaInfo)
		if err != nil {
			return &DownloadResult{
				Success:        false,
				Message:        "检查媒体存在性失败",
				Error:          err,
				ProcessingTime: time.Since(startTime),
			}
		}

		if exists {
			return &DownloadResult{
				Success:        false,
				Message:        fmt.Sprintf("%s 媒体库中已存在，跳过", torrent.Title),
				ProcessingTime: time.Since(startTime),
			}
		}
	}

	// 创建下载任务
	downloadID, err := dm.createDownloadTask(ctx, params, torrent, mediaInfo)
	if err != nil {
		return &DownloadResult{
			Success:        false,
			Message:        fmt.Sprintf("创建下载任务失败: %v", err),
			Error:          err,
			ProcessingTime: time.Since(startTime),
		}
	}

	// 保存缓存
	if err := dm.saveCache(ctx, workflowID, cacheKey); err != nil {
		dm.logger.Warn("保存缓存失败", zap.Error(err))
	}

	return &DownloadResult{
		Success:        true,
		DownloadID:     downloadID,
		Message:        "下载任务创建成功",
		ProcessingTime: time.Since(startTime),
	}
}

// recognizeMedia 识别媒体信息
func (dm *DownloadManager) recognizeMedia(ctx context.Context, torrent *types.TorrentInfo) (*types.MediaInfo, error) {
	// 这里应该调用媒体识别链
	// 暂时返回基本的媒体信息
	mediaInfo := &types.MediaInfo{
		Title:       torrent.Title,
		Year:        torrent.Year,
		Type:        torrent.Type,
		Season:      torrent.Season,
		Episodes:    torrent.Episodes,
		Resolution:  torrent.Resolution,
		Quality:     torrent.Quality,
		Source:      torrent.SiteName,
		Description: torrent.Description,
	}

	return mediaInfo, nil
}

// checkMediaExists 检查媒体是否已存在
func (dm *DownloadManager) checkMediaExists(ctx context.Context, mediaInfo *types.MediaInfo) (bool, error) {
	// 查询媒体库中是否已存在
	medias, err := dm.mediaRepo.List(ctx, 1, 1000)
	if err != nil {
		return false, err
	}

	for _, media := range medias {
		if dm.isSameMedia(media, mediaInfo) {
			return true, nil
		}
	}

	return false, nil
}

// isSameMedia 判断是否为同一媒体
func (dm *DownloadManager) isSameMedia(existing *types.Media, new *types.MediaInfo) bool {
	// 简单的媒体匹配逻辑
	if existing.Title != new.Title {
		return false
	}

	if existing.Year != new.Year {
		return false
	}

	if existing.Type != new.Type {
		return false
	}

	// 对于电视剧，检查季和集
	if existing.Type == "tv" {
		if existing.Season != new.Season {
			return false
		}

		// 检查集数是否有重叠
		if len(existing.Episodes) > 0 && len(new.Episodes) > 0 {
			for _, ep := range new.Episodes {
				for _, existingEp := range existing.Episodes {
					if ep == existingEp {
						return true
					}
				}
			}
		}
	}

	return true
}

// createDownloadTask 创建下载任务
func (dm *DownloadManager) createDownloadTask(
	ctx context.Context,
	params *AddDownloadParams,
	torrent *types.TorrentInfo,
	mediaInfo *types.MediaInfo,
) (string, error) {
	// 生成下载ID
	downloadID := fmt.Sprintf("dl_%d_%s", time.Now().Unix(), torrent.Hash[:8])

	// 创建下载任务记录
	download := &types.Download{
		ID:         downloadID,
		Title:      torrent.Title,
		URL:        torrent.DownloadURL,
		Hash:       torrent.Hash,
		Size:       torrent.Size,
		Type:       mediaInfo.Type,
		Season:     mediaInfo.Season,
		Episodes:   mediaInfo.Episodes,
		Downloader: params.Downloader,
		SavePath:   params.SavePath,
		Labels:     params.Labels,
		Status:     "pending",
		SiteID:     torrent.SiteID,
		SiteName:   torrent.SiteName,
		MediaID:    mediaInfo.ID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// 保存到数据库
	if err := dm.downloadRepo.Create(ctx, download); err != nil {
		return "", fmt.Errorf("保存下载任务失败: %w", err)
	}

	return downloadID, nil
}

// checkCache 检查缓存
func (dm *DownloadManager) checkCache(ctx context.Context, workflowID int64, key string) bool {
	if dm.cache == nil {
		return false
	}

	cacheKey := fmt.Sprintf("download_cache_%d", workflowID)
	exists, err := dm.cache.Exists(ctx, cacheKey, key)
	if err != nil {
		dm.logger.Warn("检查缓存失败", zap.Error(err))
		return false
	}

	return exists
}

// saveCache 保存缓存
func (dm *DownloadManager) saveCache(ctx context.Context, workflowID int64, key string) error {
	if dm.cache == nil {
		return nil
	}

	cacheKey := fmt.Sprintf("download_cache_%d", workflowID)
	return dm.cache.Set(ctx, cacheKey, key, 24*time.Hour)
}

// isWorkflowStopped 检查工作流是否已停止
func (dm *DownloadManager) isWorkflowStopped(ctx context.Context, workflowID int64) bool {
	// 这里应该检查工作流状态
	// 暂时返回false
	return false
}

// GetDownloadStatus 获取下载状态
func (dm *DownloadManager) GetDownloadStatus(ctx context.Context, downloadID string) (*DownloadTask, error) {
	download, err := dm.downloadRepo.GetByID(ctx, downloadID)
	if err != nil {
		return nil, err
	}

	task := &DownloadTask{
		DownloadID: download.ID,
		Downloader: download.Downloader,
		Title:      download.Title,
		Size:       download.Size,
		SavePath:   download.SavePath,
		Labels:     download.Labels,
		Status:     download.Status,
		Progress:   download.Progress,
		Speed:      download.Speed,
		CreatedAt:  download.CreatedAt,
		UpdatedAt:  download.UpdatedAt,
		Metadata: map[string]string{
			"site_id":   download.SiteID,
			"site_name": download.SiteName,
			"hash":      download.Hash,
		},
	}

	return task, nil
}

// ListDownloads 列出下载任务
func (dm *DownloadManager) ListDownloads(ctx context.Context, page, pageSize int, status string) ([]*DownloadTask, int64, error) {
	downloads, total, err := dm.downloadRepo.List(ctx, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	tasks := make([]*DownloadTask, len(downloads))
	for i, download := range downloads {
		tasks[i] = &DownloadTask{
			DownloadID: download.ID,
			Downloader: download.Downloader,
			Title:      download.Title,
			Size:       download.Size,
			SavePath:   download.SavePath,
			Labels:     download.Labels,
			Status:     download.Status,
			Progress:   download.Progress,
			Speed:      download.Speed,
			CreatedAt:  download.CreatedAt,
			UpdatedAt:  download.UpdatedAt,
			Metadata: map[string]string{
				"site_id":   download.SiteID,
				"site_name": download.SiteName,
				"hash":      download.Hash,
			},
		}
	}

	return tasks, total, nil
}

// CancelDownload 取消下载任务
func (dm *DownloadManager) CancelDownload(ctx context.Context, downloadID string) error {
	// 更新下载状态为已取消
	download, err := dm.downloadRepo.GetByID(ctx, downloadID)
	if err != nil {
		return err
	}

	download.Status = "cancelled"
	download.UpdatedAt = time.Now()

	return dm.downloadRepo.Update(ctx, download)
}

// RetryDownload 重试下载任务
func (dm *DownloadManager) RetryDownload(ctx context.Context, downloadID string) error {
	// 重置下载状态为待处理
	download, err := dm.downloadRepo.GetByID(ctx, downloadID)
	if err != nil {
		return err
	}

	download.Status = "pending"
	download.Progress = 0
	download.Speed = 0
	download.UpdatedAt = time.Now()

	return dm.downloadRepo.Update(ctx, download)
}

// intSliceToStringSlice 将int切片转换为string切片
func (dm *DownloadManager) intSliceToStringSlice(intSlice []int) []string {
	if intSlice == nil {
		return []string{}
	}

	stringSlice := make([]string, len(intSlice))
	for i, v := range intSlice {
		stringSlice[i] = fmt.Sprintf("%d", v)
	}
	return stringSlice
}
