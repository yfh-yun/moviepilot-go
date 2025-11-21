// Package actions 提供下载列表获取功能的实现
package actions

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"moviepilot-go/internal/business/services/actions/types"
	"moviepilot-go/internal/infrastructure/config"
	"moviepilot-go/internal/repositories/interfaces"
	"moviepilot-go/pkg/logger"
	"moviepilot-go/pkg/plugin"
)

// DownloadFetcher 下载列表获取器
type DownloadFetcher struct {
	logger              logger.Logger
	config              *config.Config
	pluginManager       plugin.PluginManager
	downloadRepo        interfaces.DownloadRepository
	mediaService        types.MediaService
	statisticsCollector *DownloadStatisticsCollector
}

// NewDownloadFetcher 创建下载列表获取器实例
func NewDownloadFetcher(
	config *config.Config,
	pluginManager plugin.PluginManager,
	downloadRepo interfaces.DownloadRepository,
	mediaService types.MediaService,
) *DownloadFetcher {
	return &DownloadFetcher{
		logger:              logger.NewLogger("download_fetcher"),
		config:              config,
		pluginManager:       pluginManager,
		downloadRepo:        downloadRepo,
		mediaService:        mediaService,
		statisticsCollector: NewDownloadStatisticsCollector(),
	}
}

// FetchDownloads 获取下载列表
func (f *DownloadFetcher) FetchDownloads(ctx context.Context, params *FetchDownloadsParams) (*FetchDownloadsResult, error) {
	f.logger.Debug("开始获取下载列表", "params", params)
	
	// 验证参数
	if err := f.validateFetchParams(params); err != nil {
		f.logger.Error("参数验证失败", "error", err.Error())
		return nil, fmt.Errorf("参数验证失败: %w", err)
	}

	// 设置默认值
	f.setDefaultParams(params)

	var allDownloads []*DownloadItem
	var total int64

	// 从所有下载器获取下载列表
	if params.DownloaderID == "" {
		// 获取所有启用的下载器
		downloaders, err := f.getEnabledDownloaders()
		if err != nil {
			f.logger.Error("获取下载器列表失败", "error", err.Error())
			return nil, fmt.Errorf("获取下载器列表失败: %w", err)
		}

		// 从每个下载器获取下载列表
		for _, downloaderID := range downloaders {
			downloads, err := f.fetchFromDownloader(ctx, downloaderID, params)
			if err != nil {
				f.logger.Warn("从下载器获取下载列表失败", "downloader_id", downloaderID, "error", err.Error())
				continue
			}

			allDownloads = append(allDownloads, downloads...)
		}

		// 应用过滤和排序
		filteredDownloads, err := f.filterAndSortDownloads(allDownloads, params)
		if err != nil {
			f.logger.Error("过滤和排序下载列表失败", "error", err.Error())
			return nil, fmt.Errorf("处理下载列表失败: %w", err)
		}

		total = int64(len(filteredDownloads))
		allDownloads = f.applyPagination(filteredDownloads, params.Limit, params.Offset)
	} else {
		// 从指定下载器获取下载列表
		downloads, err := f.fetchFromDownloader(ctx, params.DownloaderID, params)
		if err != nil {
			f.logger.Error("从指定下载器获取下载列表失败", "downloader_id", params.DownloaderID, "error", err.Error())
			return nil, fmt.Errorf("获取下载列表失败: %w", err)
		}

		// 应用过滤和排序
		filteredDownloads, err := f.filterAndSortDownloads(downloads, params)
		if err != nil {
			f.logger.Error("过滤和排序下载列表失败", "error", err.Error())
			return nil, fmt.Errorf("处理下载列表失败: %w", err)
		}

		total = int64(len(filteredDownloads))
		allDownloads = f.applyPagination(filteredDownloads, params.Limit, params.Offset)
	}

	// 如果需要，获取媒体信息
	if params.IncludeMediaInfo {
		if err := f.enrichMediaInfo(ctx, allDownloads); err != nil {
			f.logger.Warn("丰富媒体信息失败", "error", err.Error())
		}
	}

	// 更新统计信息
	f.statisticsCollector.UpdateDownloadStats(allDownloads)

	result := &FetchDownloadsResult{
		Total:  total,
		Items:  allDownloads,
		Params: *params,
	}

	f.logger.Debug("获取下载列表完成", "total", total, "returned", len(allDownloads))
	return result, nil
}

// GetDownloadStatus 获取单个下载的状态
func (f *DownloadFetcher) GetDownloadStatus(ctx context.Context, params *GetDownloadStatusParams) (*GetDownloadStatusResult, error) {
	f.logger.Debug("获取下载状态", "download_id", params.ID, "downloader_id", params.DownloaderID)

	// 验证参数
	if params.ID == "" {
		return nil, errors.New("下载ID不能为空")
	}

	// 从指定下载器或所有下载器中查找
	downloaders := []string{}
	if params.DownloaderID != "" {
		downloaders = append(downloaders, params.DownloaderID)
	} else {
		// 获取所有下载器
		enabledDownloaders, err := f.getEnabledDownloaders()
		if err != nil {
			f.logger.Error("获取下载器列表失败", "error", err.Error())
			return nil, fmt.Errorf("获取下载器列表失败: %w", err)
		}
		downloaders = enabledDownloaders
	}

	// 在所有下载器中查找下载项
	for _, downloaderID := range downloaders {
		item, err := f.getDownloadFromDownloader(ctx, downloaderID, params.ID)
		if err != nil {
			f.logger.Debug("从下载器获取下载项失败", "downloader_id", downloaderID, "error", err.Error())
			continue
		}

		if item != nil {
			// 如果需要，获取媒体信息
			if params.IncludeMediaInfo {
				if err := f.enrichMediaInfoForItem(ctx, item); err != nil {
					f.logger.Warn("丰富媒体信息失败", "error", err.Error())
				}
			}

			return &GetDownloadStatusResult{
				Item:  item,
				Found: true,
			}, nil
		}
	}

	f.logger.Info("下载项未找到", "download_id", params.ID)
	return &GetDownloadStatusResult{Found: false}, nil
}

// GetDownloadStats 获取下载统计信息
func (f *DownloadFetcher) GetDownloadStats(ctx context.Context) (*DownloadStats, error) {
	f.logger.Debug("获取下载统计信息")

	// 获取所有下载器的统计信息
	stats := &DownloadStats{
		StatusCounts: make(map[DownloadStatus]int),
		TypeCounts:   make(map[DownloadType]int),
	}

	// 从所有下载器获取统计
	downloaders, err := f.getEnabledDownloaders()
	if err != nil {
		f.logger.Error("获取下载器列表失败", "error", err.Error())
		return nil, fmt.Errorf("获取下载器列表失败: %w", err)
	}

	for _, downloaderID := range downloaders {
		downloaderStats, err := f.getDownloaderStats(ctx, downloaderID)
		if err != nil {
			f.logger.Warn("从下载器获取统计失败", "downloader_id", downloaderID, "error", err.Error())
			continue
		}

		// 合并统计信息
		stats.TotalCount += downloaderStats.TotalCount
		stats.TotalSize += downloaderStats.TotalSize
		stats.TotalDownloadedSize += downloaderStats.TotalDownloadedSize
		stats.CurrentDownloadSpeed += downloaderStats.CurrentDownloadSpeed
		stats.CurrentUploadSpeed += downloaderStats.CurrentUploadSpeed
		stats.TodayDownloaded += downloaderStats.TodayDownloaded
		stats.TodayUploaded += downloaderStats.TodayUploaded
		stats.AllTimeDownloaded += downloaderStats.AllTimeDownloaded
		stats.AllTimeUploaded += downloaderStats.AllTimeUploaded

		// 合并状态统计
		for status, count := range downloaderStats.StatusCounts {
			stats.StatusCounts[status] += count
		}

		// 合并类型统计
		for downloadType, count := range downloaderStats.TypeCounts {
			stats.TypeCounts[downloadType] += count
		}
	}

	// 计算平均分享率
	if stats.TotalCount > 0 {
		// 这里可以从数据库或缓存中获取平均分享率
		avgRatio, err := f.downloadRepo.GetAverageRatio(ctx)
		if err == nil {
			stats.AverageRatio = avgRatio
		} else {
			stats.AverageRatio = 0
		}
	}

	return stats, nil
}

// 从指定下载器获取下载列表
func (f *DownloadFetcher) fetchFromDownloader(ctx context.Context, downloaderID string, params *FetchDownloadsParams) ([]*DownloadItem, error) {
	f.logger.Debug("从下载器获取下载列表", "downloader_id", downloaderID)

	// 获取下载器插件
	downloader, err := f.pluginManager.GetPluginByID(downloaderID)
	if err != nil {
		return nil, fmt.Errorf("获取下载器插件失败: %w", err)
	}

	// 调用下载器的API获取下载列表
	// 这里简化处理，实际应通过gRPC或其他方式调用下载器插件
	result, err := downloader.CallMethod(ctx, "GetDownloads", params)
	if err != nil {
		return nil, fmt.Errorf("调用下载器API失败: %w", err)
	}

	// 转换结果
	downloads, ok := result.([]*DownloadItem)
	if !ok {
		return nil, errors.New("下载器返回的数据格式不正确")
	}

	// 设置下载器ID
	for _, download := range downloads {
		download.DownloaderID = downloaderID
	}

	return downloads, nil
}

// 从指定下载器获取单个下载项
func (f *DownloadFetcher) getDownloadFromDownloader(ctx context.Context, downloaderID, downloadID string) (*DownloadItem, error) {
	params := &GetDownloadStatusParams{
		ID:           downloadID,
		DownloaderID: downloaderID,
	}

	// 调用下载器API
	downloader, err := f.pluginManager.GetPluginByID(downloaderID)
	if err != nil {
		return nil, err
	}

	result, err := downloader.CallMethod(ctx, "GetDownloadStatus", params)
	if err != nil {
		return nil, err
	}

	statusResult, ok := result.(*GetDownloadStatusResult)
	if !ok || !statusResult.Found {
		return nil, nil
	}

	return statusResult.Item, nil
}

// 获取下载器统计信息
func (f *DownloadFetcher) getDownloaderStats(ctx context.Context, downloaderID string) (*DownloadStats, error) {
	downloader, err := f.pluginManager.GetPluginByID(downloaderID)
	if err != nil {
		return nil, err
	}

	result, err := downloader.CallMethod(ctx, "GetDownloadStats", nil)
	if err != nil {
		return nil, err
	}

	stats, ok := result.(*DownloadStats)
	if !ok {
		return nil, errors.New("下载器返回的统计数据格式不正确")
	}

	return stats, nil
}

// 过滤和排序下载列表
func (f *DownloadFetcher) filterAndSortDownloads(downloads []*DownloadItem, params *FetchDownloadsParams) ([]*DownloadItem, error) {
	filtered := []*DownloadItem{}

	// 应用过滤条件
	for _, download := range downloads {
		// 状态过滤
		if len(params.Statuses) > 0 {
			statusMatch := false
			for _, status := range params.Statuses {
				if download.Status == status {
					statusMatch = true
					break
				}
			}
			if !statusMatch {
				continue
			}
		}

		// 类型过滤
		if len(params.Types) > 0 {
			typeMatch := false
			for _, downloadType := range params.Types {
				if download.Type == downloadType {
					typeMatch = true
					break
				}
			}
			if !typeMatch {
				continue
			}
		}

		// 关键词过滤
		if len(params.Keywords) > 0 {
			keywordMatch := false
			for _, keyword := range params.Keywords {
				if strings.Contains(strings.ToLower(download.Title), strings.ToLower(keyword)) {
					keywordMatch = true
					break
				}
			}
			if !keywordMatch {
				continue
			}
		}

		// 标签过滤
		if len(params.Tags) > 0 {
			tagMatch := false
			downloadTags := make(map[string]bool)
			for _, tag := range download.Tags {
				downloadTags[tag] = true
			}

			for _, tag := range params.Tags {
				if downloadTags[tag] {
					tagMatch = true
					break
				}
			}

			if !tagMatch {
				continue
			}
		}

		// 只显示活跃下载
		if params.OnlyActive {
			if download.Status == DownloadStatusCompleted || download.Status == DownloadStatusError {
				continue
			}
		}

		filtered = append(filtered, download)
	}

	// 应用排序
	// 这里简化处理，实际应根据params.OrderBy和params.OrderDir进行排序
	// 可以使用sort.Slice进行排序

	return filtered, nil
}

// 应用分页
func (f *DownloadFetcher) applyPagination(downloads []*DownloadItem, limit, offset int) []*DownloadItem {
	if offset >= len(downloads) {
		return []*DownloadItem{}
	}

	end := offset + limit
	if end > len(downloads) {
		end = len(downloads)
	}

	return downloads[offset:end]
}

// 丰富媒体信息
func (f *DownloadFetcher) enrichMediaInfo(ctx context.Context, downloads []*DownloadItem) error {
	for _, download := range downloads {
		if err := f.enrichMediaInfoForItem(ctx, download); err != nil {
			f.logger.Warn("丰富媒体信息失败", "download_id", download.ID, "error", err.Error())
		}
	}
	return nil
}

// 丰富单个下载项的媒体信息
func (f *DownloadFetcher) enrichMediaInfoForItem(ctx context.Context, download *DownloadItem) error {
	// 如果已有媒体信息，跳过
	if download.MediaInfo != nil && download.MediaInfo.ID != "" {
		return nil
	}

	// 从数据库获取或通过标题解析媒体信息
	mediaInfo, err := f.mediaService.GetMediaInfoFromTitle(ctx, download.Title)
	if err != nil {
		return err
	}

	download.MediaInfo = mediaInfo
	return nil
}

// 获取所有启用的下载器
func (f *DownloadFetcher) getEnabledDownloaders() ([]string, error) {
	// 从配置或插件管理器获取启用的下载器
	downloaders, err := f.pluginManager.GetPluginsByType("downloader")
	if err != nil {
		return nil, err
	}

	enabledDownloaders := []string{}
	for _, downloader := range downloaders {
		if downloader.IsEnabled() {
			enabledDownloaders = append(enabledDownloaders, downloader.GetID())
		}
	}

	return enabledDownloaders, nil
}

// 验证获取参数
func (f *DownloadFetcher) validateFetchParams(params *FetchDownloadsParams) error {
	if params == nil {
		return errors.New("参数不能为空")
	}

	// 验证下载状态
	validStatuses := map[DownloadStatus]bool{
		DownloadStatusPending:    true,
		DownloadStatusDownloading: true,
		DownloadStatusPaused:     true,
		DownloadStatusCompleted:  true,
		DownloadStatusError:      true,
		DownloadStatusSeeding:    true,
		DownloadStatusStalled:    true,
		DownloadStatusChecking:   true,
	}

	for _, status := range params.Statuses {
		if !validStatuses[status] {
			return fmt.Errorf("无效的下载状态: %s", status)
		}
	}

	// 验证下载类型
	validTypes := map[DownloadType]bool{
		DownloadTypeTorrent: true,
		DownloadTypeMagnet:  true,
		DownloadTypeURL:     true,
		DownloadTypeNZB:     true,
	}

	for _, downloadType := range params.Types {
		if !validTypes[downloadType] {
			return fmt.Errorf("无效的下载类型: %s", downloadType)
		}
	}

	return nil
}

// 设置默认参数
func (f *DownloadFetcher) setDefaultParams(params *FetchDownloadsParams) {
	if params.Limit <= 0 {
		params.Limit = 50
	}

	if params.Limit > 500 {
		params.Limit = 500 // 最大500条
	}

	if params.Offset < 0 {
		params.Offset = 0
	}

	if params.OrderBy == "" {
		params.OrderBy = "updated_at"
	}

	if params.OrderDir == "" {
		params.OrderDir = "desc"
	}
}

// DownloadStatisticsCollector 下载统计收集器
type DownloadStatisticsCollector struct {
	lastUpdate time.Time
	stats      *DownloadStats
}

// NewDownloadStatisticsCollector 创建下载统计收集器
func NewDownloadStatisticsCollector() *DownloadStatisticsCollector {
	return &DownloadStatisticsCollector{
		lastUpdate: time.Now(),
		stats: &DownloadStats{
			StatusCounts: make(map[DownloadStatus]int),
			TypeCounts:   make(map[DownloadType]int),
		},
	}
}

// UpdateDownloadStats 更新下载统计
func (c *DownloadStatisticsCollector) UpdateDownloadStats(downloads []*DownloadItem) {
	// 重置统计
	c.stats = &DownloadStats{
		StatusCounts: make(map[DownloadStatus]int),
		TypeCounts:   make(map[DownloadType]int),
	}

	// 统计下载项
	c.stats.TotalCount = len(downloads)
	for _, download := range downloads {
		c.stats.StatusCounts[download.Status]++
		c.stats.TypeCounts[download.Type]++
		c.stats.TotalSize += download.TotalSize
		c.stats.TotalDownloadedSize += download.DownloadedSize
		c.stats.CurrentDownloadSpeed += download.DownloadSpeed
		c.stats.CurrentUploadSpeed += download.UploadSpeed
	}

	c.lastUpdate = time.Now()
}
