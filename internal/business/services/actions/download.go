// Package actions 提供下载相关的业务逻辑实现
package actions

import (
	"context"
	"fmt"
	"strings"
	"time"

	"moviepilot-go/internal/business/services/actions/types"
	"moviepilot-go/internal/repositories/interfaces"
	"moviepilot-go/pkg/logger"
)

// AddDownloadParams 下载参数结构
type AddDownloadParams struct {
	Downloader string   `json:"downloader"`
	SavePath   string   `json:"save_path"`
	Labels     []string `json:"labels"`
	OnlyLack   bool     `json:"only_lack"`
	Sites      []int    `json:"sites"`
	Quality    string   `json:"quality"`
	Resolution string   `json:"resolution"`
}

// AddDownloadResult 下载结果结构
type AddDownloadResult struct {
	Success    bool   `json:"success"`
	DownloadID string `json:"download_id"`
	Message    string `json:"message"`
	Torrent    *types.Torrent `json:"torrent,omitempty"`
}

// download.go 实现下载管理器的核心功能
type DownloadManagerImpl struct {
	downloadRepo interfaces.DownloadRepository
	mediaRepo    interfaces.MediaRepository
	cache        *WorkflowCache
	logger       logger.Logger
}

// NewDownloadManager 创建下载管理器实例
func NewDownloadManager(
	downloadRepo interfaces.DownloadRepository,
	mediaRepo interfaces.MediaRepository,
	cache *WorkflowCache,
) *DownloadManagerImpl {
	return &DownloadManagerImpl{
		downloadRepo: downloadRepo,
		mediaRepo:    mediaRepo,
		cache:        cache,
		logger:       logger.NewLogger("download_manager"),
	}
}

// AddDownload 添加下载任务
func (m *DownloadManagerImpl) AddDownload(
	ctx context.Context,
	workflowID int64,
	params *AddDownloadParams,
	torrents []*types.Torrent,
) ([]*AddDownloadResult, error) {
	m.logger.Debug("开始添加下载任务", "workflow_id", workflowID, "torrents_count", len(torrents))

	results := make([]*AddDownloadResult, 0, len(torrents))

	// 遍历所有种子进行处理
	for _, torrent := range torrents {
		result := &AddDownloadResult{
			Torrent: torrent,
		}

		// 检查是否只下载缺失内容
		if params.OnlyLack {
			isLack, err := m.checkIsLack(ctx, torrent)
			if err != nil {
				m.logger.Error("检查缺失内容失败", "error", err, "torrent_name", torrent.Title)
				result.Success = false
				result.Message = fmt.Sprintf("检查缺失失败: %v", err)
				results = append(results, result)
				continue
			}

			if !isLack {
				m.logger.Info("跳过已存在的内容", "torrent_name", torrent.Title)
				result.Success = false
				result.Message = "内容已存在，跳过下载"
				results = append(results, result)
				continue
			}
		}

		// 添加下载任务
		downloadID, err := m.addTorrentToDownloader(ctx, params, torrent)
		if err != nil {
			m.logger.Error("添加到下载器失败", "error", err, "torrent_name", torrent.Title)
			result.Success = false
			result.Message = fmt.Sprintf("添加失败: %v", err)
			results = append(results, result)
			continue
		}

		// 保存下载记录
		downloadRecord := &types.Download{
			ID:         downloadID,
			Title:      torrent.Title,
			URL:        torrent.URL,
			Downloader: params.Downloader,
			SavePath:   params.SavePath,
			Status:     "pending",
			Labels:     params.Labels,
			WorkflowID: workflowID,
			CreatedAt:  m.currentTime(),
		}

		if err := m.downloadRepo.Create(ctx, downloadRecord); err != nil {
			m.logger.Error("保存下载记录失败", "error", err, "download_id", downloadID)
			// 不影响下载结果，只记录错误
		}

		result.Success = true
		result.DownloadID = downloadID
		result.Message = "下载任务添加成功"
		results = append(results, result)
		m.logger.Info("下载任务添加成功", "download_id", downloadID, "torrent_name", torrent.Title)
	}

	m.logger.Debug("下载任务添加完成", "total", len(torrents), "success", m.getSuccessCount(results))
	return results, nil
}

// checkIsLack 检查内容是否缺失
func (m *DownloadManagerImpl) checkIsLack(ctx context.Context, torrent *types.Torrent) (bool, error) {
	// 实现检查逻辑：根据种子名称搜索媒体库，判断是否已存在
	// 这里是简化实现，实际应该根据种子信息匹配媒体库中的内容
	keyword := torrent.Title
	// 尝试从标题中提取关键词
	if strings.Contains(keyword, " ") {
		parts := strings.Split(keyword, " ")
		if len(parts) > 0 {
			keyword = parts[0]
		}
	}

	medias, err := m.mediaRepo.Search(ctx, keyword, 1, 10)
	if err != nil {
		return true, err // 默认当作缺失处理
	}

	// 如果找不到匹配的媒体，认为内容缺失
	return len(medias) == 0, nil
}

// addTorrentToDownloader 添加种子到下载器
func (m *DownloadManagerImpl) addTorrentToDownloader(ctx context.Context, params *AddDownloadParams, torrent *types.Torrent) (string, error) {
	// 根据不同的下载器类型处理
	switch strings.ToLower(params.Downloader) {
	case "qbittorrent":
		return m.addToQBitTorrent(ctx, params, torrent)
	case "transmission":
		return m.addToTransmission(ctx, params, torrent)
	default:
		return "", fmt.Errorf("不支持的下载器类型: %s", params.Downloader)
	}
}

// addToQBitTorrent 添加到qBittorrent
func (m *DownloadManagerImpl) addToQBitTorrent(ctx context.Context, params *AddDownloadParams, torrent *types.Torrent) (string, error) {
	// 这里应该调用qBittorrent的API添加下载
	// 暂时返回模拟的下载ID
	downloadID := fmt.Sprintf("qb_%s_%d", m.generateUniqueID(), m.currentTimeUnix())
	return downloadID, nil
}

// addToTransmission 添加到Transmission
func (m *DownloadManagerImpl) addToTransmission(ctx context.Context, params *AddDownloadParams, torrent *types.Torrent) (string, error) {
	// 这里应该调用Transmission的API添加下载
	// 暂时返回模拟的下载ID
	downloadID := fmt.Sprintf("tr_%s_%d", m.generateUniqueID(), m.currentTimeUnix())
	return downloadID, nil
}

// getSuccessCount 获取成功的数量
func (m *DownloadManagerImpl) getSuccessCount(results []*AddDownloadResult) int {
	count := 0
	for _, result := range results {
		if result.Success {
			count++
		}
	}
	return count
}

// generateUniqueID 生成唯一ID
func (m *DownloadManagerImpl) generateUniqueID() string {
	// 简化实现，实际应该使用UUID或更安全的方式
	return fmt.Sprintf("%d", m.currentTimeUnix())
}

// currentTime 获取当前时间字符串
func (m *DownloadManagerImpl) currentTime() string {
	return fmt.Sprintf("%d", m.currentTimeUnix())
}

// currentTimeUnix 获取当前时间戳
func (m *DownloadManagerImpl) currentTimeUnix() int64 {
	return time.Now().Unix()
}
