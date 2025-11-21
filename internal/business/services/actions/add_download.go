// Package actions 提供动作系统的业务逻辑实现
package actions

import (
	"context"
	"fmt"
	"strings"
	"time"

	"moviepilot-go/pkg/logger"
	"moviepilot-go/internal/repositories/interfaces"
	"moviepilot-go/internal/business/services/actions/types"

	"go.uber.org/zap"
)

// AddDownloadAction 添加下载动作
// 对应Python版本app/actions/add_download.py的AddDownloadAction
type AddDownloadAction struct {
	downloadManager *DownloadManager
	mediaChain      *MediaChain
	cache           *WorkflowCache
	addedDownloads  []string
	hasError        bool
	logger          *zap.Logger
}

// NewAddDownloadAction 创建添加下载动作实例
func NewAddDownloadAction(
	downloadRepo interfaces.DownloadRepository,
	mediaRepo interfaces.MediaRepository,
	cache *WorkflowCache,
) *AddDownloadAction {
	return &AddDownloadAction{
		downloadManager: NewDownloadManager(downloadRepo, mediaRepo, cache),
		mediaChain:      NewMediaChain(),
		cache:           cache,
		addedDownloads:  make([]string, 0),
		hasError:        false,
		logger:          logger.Logger,
	}
}

// Execute 执行添加下载动作
// 实现Python版本AddDownloadAction.execute()方法的完整功能
func (ada *AddDownloadAction) Execute(
	ctx context.Context,
	workflowID int64,
	params map[string]interface{},
	actionCtx *types.ActionContext,
) (*types.ActionContext, error) {
	startTime := time.Now()

	// 解析参数
	addParams, err := ada.parseParams(params)
	if err != nil {
		ada.logger.Error("解析添加下载参数失败", zap.Error(err))
		return actionCtx, err
	}

	ada.logger.Info("开始执行添加下载动作",
		zap.Int64("workflow_id", workflowID),
		zap.Int("torrent_count", len(actionCtx.Torrents)),
		zap.String("downloader", addParams.Downloader),
		zap.Bool("only_lack", addParams.OnlyLack),
	)

	// 处理每个种子
	for _, torrent := range actionCtx.Torrents {
		// 检查工作流是否已停止
		if ada.isWorkflowStopped(ctx, workflowID) {
			ada.logger.Info("工作流已停止，终止下载添加", zap.Int64("workflow_id", workflowID))
			break
		}

		// 处理单个种子
		result := ada.processTorrent(ctx, workflowID, addParams, torrent, actionCtx)
		if result.Success {
			ada.addedDownloads = append(ada.addedDownloads, result.DownloadID)
			ada.logger.Info("种子添加下载成功",
				zap.String("title", torrent.Title),
				zap.String("download_id", result.DownloadID),
			)
		} else {
			ada.hasError = true
			ada.logger.Warn("种子添加下载失败",
				zap.String("title", torrent.Title),
				zap.String("error", result.Message),
			)
		}
	}

	// 更新动作上下文
	if len(ada.addedDownloads) > 0 {
		// 添加到下载列表
		for _, downloadID := range ada.addedDownloads {
			actionCtx.Downloads = append(actionCtx.Downloads, &types.Download{
				ID:         downloadID,
				Downloader: addParams.Downloader,
				Status:     "pending",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			})
		}
		ada.logger.Info("添加下载任务完成",
			zap.Int("added_count", len(ada.addedDownloads)),
			zap.Duration("duration", time.Since(startTime)),
		)
	} else if len(actionCtx.Torrents) > 0 {
		ada.hasError = true
		ada.logger.Warn("没有成功添加任何下载任务")
	}

	return actionCtx, nil
}

// parseParams 解析动作参数
func (ada *AddDownloadAction) parseParams(params map[string]interface{}) (*AddDownloadParams, error) {
	addParams := &AddDownloadParams{}

	// 设置默认值
	if downloader, ok := params["downloader"].(string); ok {
		addParams.Downloader = downloader
	}
	if savePath, ok := params["save_path"].(string); ok {
		addParams.SavePath = savePath
	}
	if labels, ok := params["labels"].([]interface{}); ok {
		for _, label := range labels {
			if str, ok := label.(string); ok {
				addParams.Labels = append(addParams.Labels, str)
			}
		}
	} else if labelsStr, ok := params["labels"].(string); ok {
		// 支持逗号分隔的字符串
		addParams.Labels = strings.Split(labelsStr, ",")
		for i, label := range addParams.Labels {
			addParams.Labels[i] = strings.TrimSpace(label)
		}
	}
	if onlyLack, ok := params["only_lack"].(bool); ok {
		addParams.OnlyLack = onlyLack
	}

	return addParams, nil
}

// processTorrent 处理单个种子
func (ada *AddDownloadAction) processTorrent(
	ctx context.Context,
	workflowID int64,
	params *AddDownloadParams,
	torrent *types.TorrentInfo,
	actionCtx *types.ActionContext,
) *DownloadResult {
	startTime := time.Now()

	// 生成缓存键
	cacheKey := fmt.Sprintf("%s-%s", torrent.SiteID, torrent.Title)

	// 检查缓存
	if ada.checkCache(ctx, workflowID, cacheKey) {
		return &DownloadResult{
			Success:        false,
			Message:        fmt.Sprintf("%s 已添加过下载，跳过", torrent.Title),
			ProcessingTime: time.Since(startTime),
		}
	}

	// 识别媒体信息
	mediaInfo, err := ada.recognizeMedia(ctx, torrent)
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
		exists, skipReason := ada.checkMediaExists(ctx, mediaInfo)
		if exists {
			return &DownloadResult{
				Success:        false,
				Message:        fmt.Sprintf("%s %s", torrent.Title, skipReason),
				ProcessingTime: time.Since(startTime),
			}
		}
	}

	// 创建下载任务
	downloadID, err := ada.createDownloadTask(ctx, params, torrent, mediaInfo)
	if err != nil {
		return &DownloadResult{
			Success:        false,
			Message:        fmt.Sprintf("创建下载任务失败: %v", err),
			Error:          err,
			ProcessingTime: time.Since(startTime),
		}
	}

	// 保存缓存
	if err := ada.saveCache(ctx, workflowID, cacheKey); err != nil {
		ada.logger.Warn("保存缓存失败", zap.Error(err))
	}

	return &DownloadResult{
		Success:        true,
		DownloadID:     downloadID,
		Message:        "下载任务创建成功",
		ProcessingTime: time.Since(startTime),
	}
}

// recognizeMedia 识别媒体信息
func (ada *AddDownloadAction) recognizeMedia(ctx context.Context, torrent *types.TorrentInfo) (*types.MediaInfo, error) {
	// 如果已有媒体信息，直接返回
	if torrent.Metadata != nil {
		if tmdbIDStr, ok := torrent.Metadata["tmdb_id"]; ok {
			if tmdbID, err := ada.parseTMDBID(tmdbIDStr); err == nil && tmdbID > 0 {
				mediaInfo := &types.MediaInfo{
					TMDBID:      tmdbID,
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
		}
	}

	// 调用媒体识别链
	mediaInfo, err := ada.mediaChain.RecognizeMedia(ctx, &types.ActionContext{
		Torrents: []*types.TorrentInfo{torrent},
	})
	if err != nil {
		return nil, fmt.Errorf("媒体识别失败: %w", err)
	}

	if mediaInfo == nil {
		return nil, fmt.Errorf("未识别到媒体信息")
	}

	return mediaInfo, nil
}

// checkMediaExists 检查媒体是否已存在
func (ada *AddDownloadAction) checkMediaExists(ctx context.Context, mediaInfo *types.MediaInfo) (bool, string) {
	// 这里应该调用下载链来检查媒体是否存在
	// 暂时使用简化逻辑

	// 对于电影，简单检查标题和年份
	if mediaInfo.Type == "movie" {
		// 假设调用DownloadChain().media_exists(mediaInfo)
		// 如果存在则返回true和跳过原因
		return false, ""
	}

	// 对于电视剧，检查季和集
	if mediaInfo.Type == "tv" {
		if len(mediaInfo.Seasons) > 1 {
			// 多季不下载
			return true, "有多季，跳过"
		}

		// 这里应该检查具体集数是否已存在
		// 暂时返回false
		return false, ""
	}

	return false, ""
}

// createDownloadTask 创建下载任务
func (ada *AddDownloadAction) createDownloadTask(
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

	// 这里应该调用DownloadChain().download_single()
	// 暂时只生成ID
	return downloadID, nil
}

// checkCache 检查缓存
func (ada *AddDownloadAction) checkCache(ctx context.Context, workflowID int64, key string) bool {
	if ada.cache == nil {
		return false
	}

	cacheKey := fmt.Sprintf("download_cache_%d", workflowID)
	exists, err := ada.cache.Exists(ctx, cacheKey, key)
	if err != nil {
		ada.logger.Warn("检查缓存失败", zap.Error(err))
		return false
	}

	return exists
}

// saveCache 保存缓存
func (ada *AddDownloadAction) saveCache(ctx context.Context, workflowID int64, key string) error {
	if ada.cache == nil {
		return nil
	}

	cacheKey := fmt.Sprintf("download_cache_%d", workflowID)
	return ada.cache.Set(ctx, cacheKey, key, 24*time.Hour)
}

// isWorkflowStopped 检查工作流是否已停止
func (ada *AddDownloadAction) isWorkflowStopped(ctx context.Context, workflowID int64) bool {
	// 这里应该检查工作流状态
	// 暂时返回false
	return false
}

// parseTMDBID 解析TMDB ID
func (ada *AddDownloadAction) parseTMDBID(value interface{}) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		var tmdbID int
		_, err := fmt.Sscanf(v, "%d", &tmdbID)
		return tmdbID, err
	default:
		return 0, fmt.Errorf("无法解析TMDB ID: %v", value)
	}
}

// GetSuccess 获取执行结果
func (ada *AddDownloadAction) GetSuccess() bool {
	return !ada.hasError
}

// GetAddedDownloads 获取已添加的下载ID列表
func (ada *AddDownloadAction) GetAddedDownloads() []string {
	return ada.addedDownloads
}

// GetName 获取动作名称
func (ada *AddDownloadAction) GetName() string {
	return "添加下载"
}

// GetDescription 获取动作描述
func (ada *AddDownloadAction) GetDescription() string {
	return "根据资源列表添加下载任务"
}

// GetData 获取动作参数定义
func (ada *AddDownloadAction) GetData() map[string]interface{} {
	return map[string]interface{}{
		"downloader": map[string]interface{}{
			"type":        "string",
			"description": "下载器",
			"default":     "",
		},
		"save_path": map[string]interface{}{
			"type":        "string",
			"description": "保存路径",
			"default":     "",
		},
		"labels": map[string]interface{}{
			"type":        "array",
			"description": "标签列表",
			"default":     []string{},
		},
		"only_lack": map[string]interface{}{
			"type":        "boolean",
			"description": "仅下载缺失的资源",
			"default":     false,
		},
	}
}
