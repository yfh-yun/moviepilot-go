package actions

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"moviepilot-go/pkg/logger"
	"moviepilot-go/pkg/plugin"
)

// TorrentManager 种子管理器接口
type TorrentManager interface {
	// FetchTorrents 获取种子列表
	FetchTorrents(ctx context.Context, params *FetchTorrentsParams) (*TorrentResponse, error)
	// GetTorrentByID 根据ID获取种子信息
	GetTorrentByID(ctx context.Context, id, downloaderID string) (*TorrentItem, error)
	// GetTorrentStats 获取种子统计信息
	GetTorrentStats(ctx context.Context, downloaderID string) (*TorrentStats, error)
	// GetTorrentFiles 获取种子文件列表
	GetTorrentFiles(ctx context.Context, id, downloaderID string) (*TorrentFilesResponse, error)
	// GetTorrentTrackers 获取种子追踪器信息
	GetTorrentTrackers(ctx context.Context, id, downloaderID string) (*TorrentTrackersResponse, error)
	// GetCategories 获取分类列表
	GetCategories(ctx context.Context, downloaderID string) (*CategoryResponse, error)
	// GetDownloaderStatus 获取下载器状态
	GetDownloaderStatus(ctx context.Context, downloaderID string) (*TorrentManagerStatus, error)
	// GetAvailableDownloaders 获取可用的下载器列表
	GetAvailableDownloaders(ctx context.Context) ([]*TorrentManagerStatus, error)
}

// torrentManager 种子管理器实现
type torrentManager struct {
	pluginManager plugin.PluginManager
	logger        logger.Logger
}

// NewTorrentManager 创建种子管理器实例
func NewTorrentManager(pluginManager plugin.PluginManager, logger logger.Logger) TorrentManager {
	return &torrentManager{
		pluginManager: pluginManager,
		logger:        logger,
	}
}

// FetchTorrents 获取种子列表
func (m *torrentManager) FetchTorrents(ctx context.Context, params *FetchTorrentsParams) (*TorrentResponse, error) {
	startTime := time.Now()
	log := m.logger.WithContext(ctx)

	// 验证参数
	if params == nil {
		params = &FetchTorrentsParams{}
	}

	// 设置默认值
	if params.Limit <= 0 || params.Limit > 1000 {
		params.Limit = 50
	}
	if params.Offset < 0 {
		params.Offset = 0
	}
	if params.SortBy == "" {
		params.SortBy = TorrentSortFieldAdded
	}
	if params.SortOrder == "" {
		params.SortOrder = SortOrderDesc
	}

	// 记录查询参数
	log.Debug("Fetching torrents with parameters", 
		"status", params.Status,
		"category", params.Category,
		"search", params.Search,
		"tags", params.Tags,
		"downloader_type", params.DownloaderType,
		"downloader_id", params.DownloaderID,
		"sort_by", params.SortBy,
		"sort_order", params.SortOrder,
		"limit", params.Limit,
		"offset", params.Offset,
	)

	// 获取所有下载器插件
	plugins, err := m.getDownloaderPlugins(ctx, params.DownloaderType, params.DownloaderID)
	if err != nil {
		log.Error("Failed to get downloader plugins", "error", err.Error())
		return nil, fmt.Errorf("获取下载器插件失败: %w", err)
	}

	if len(plugins) == 0 {
		log.Warn("No downloader plugins found")
		return &TorrentResponse{
			Success: true,
			Torrents: []*TorrentItem{},
			Total: 0,
			Filtered: 0,
			Page: (params.Offset / params.Limit) + 1,
			PageSize: params.Limit,
			TotalPages: 0,
			ProcessingTime: time.Since(startTime),
		}, nil
	}

	// 从所有下载器获取种子
	var allTorrents []*TorrentItem
	for _, p := range plugins {
		torrents, err := m.fetchTorrentsFromPlugin(ctx, p)
		if err != nil {
			log.Error("Failed to fetch torrents from plugin", 
				"plugin_id", p.ID(), 
				"plugin_name", p.Name(), 
				"error", err.Error())
			continue // 继续尝试其他下载器
		}
		allTorrents = append(allTorrents, torrents...)
	}

	// 过滤种子
	filteredTorrents := m.filterTorrents(allTorrents, params)
	log.Debug("Torrent filtering results", 
		"total", len(allTorrents), 
		"filtered", len(filteredTorrents))

	// 排序种子
	sortedTorrents := m.sortTorrents(filteredTorrents, params.SortBy, params.SortOrder)

	// 分页
	total := len(sortedTorrents)
	pagedTorrents := m.paginateTorrents(sortedTorrents, params.Limit, params.Offset)
	page := (params.Offset / params.Limit) + 1
	totalPages := (total + params.Limit - 1) / params.Limit

	// 返回结果
	result := &TorrentResponse{
		Success: true,
		Torrents: pagedTorrents,
		Total: len(allTorrents),
		Filtered: total,
		Page: page,
		PageSize: params.Limit,
		TotalPages: totalPages,
		ProcessingTime: time.Since(startTime),
	}

	log.Info("Successfully fetched torrents", 
		"total", result.Total,
		"filtered", result.Filtered,
		"returned", len(result.Torrents),
		"processing_time", result.ProcessingTime)

	return result, nil
}

// GetTorrentByID 根据ID获取种子信息
func (m *torrentManager) GetTorrentByID(ctx context.Context, id, downloaderID string) (*TorrentItem, error) {
	log := m.logger.WithContext(ctx)
	log.Debug("Getting torrent by ID", "torrent_id", id, "downloader_id", downloaderID)

	// 获取所有下载器插件
	plugins, err := m.getDownloaderPlugins(ctx, "", downloaderID)
	if err != nil {
		log.Error("Failed to get downloader plugins", "error", err.Error())
		return nil, fmt.Errorf("获取下载器插件失败: %w", err)
	}

	// 遍历所有下载器查找种子
	for _, p := range plugins {
		// 如果指定了downloaderID，则只查询匹配的下载器
		if downloaderID != "" && p.ID() != downloaderID {
			continue
		}

		// 从插件获取种子
		torrents, err := m.fetchTorrentsFromPlugin(ctx, p)
		if err != nil {
			log.Error("Failed to fetch torrents from plugin", 
				"plugin_id", p.ID(), 
				"error", err.Error())
			continue
		}

		// 查找匹配的种子
		for _, torrent := range torrents {
			if torrent.ID == id || torrent.Hash == id {
				log.Info("Found torrent by ID", "torrent_id", id, "plugin_id", p.ID())
				return torrent, nil
			}
		}
	}

	log.Warn("Torrent not found", "torrent_id", id, "downloader_id", downloaderID)
	return nil, fmt.Errorf("未找到ID为 %s 的种子", id)
}

// GetTorrentStats 获取种子统计信息
func (m *torrentManager) GetTorrentStats(ctx context.Context, downloaderID string) (*TorrentStats, error) {
	log := m.logger.WithContext(ctx)
	log.Debug("Getting torrent statistics", "downloader_id", downloaderID)

	// 获取所有下载器插件
	plugins, err := m.getDownloaderPlugins(ctx, "", downloaderID)
	if err != nil {
		log.Error("Failed to get downloader plugins", "error", err.Error())
		return nil, fmt.Errorf("获取下载器插件失败: %w", err)
	}

	stats := &TorrentStats{
		Active: 0,
		Downloading: 0,
		Seeding: 0,
		Paused: 0,
		Completed: 0,
		Error: 0,
		Total: 0,
		TotalSize: 0,
		TotalCompletedSize: 0,
		TotalUploaded: 0,
		TotalDownloaded: 0,
		TotalRatio: 0,
		CurrentDownloadSpeed: 0,
		CurrentUploadSpeed: 0,
		LastUpdated: time.Now(),
	}

	var totalRatioSum float64
	var ratioCount int

	// 从所有下载器收集统计信息
	for _, p := range plugins {
		torrents, err := m.fetchTorrentsFromPlugin(ctx, p)
		if err != nil {
			log.Error("Failed to fetch torrents from plugin", 
				"plugin_id", p.ID(), 
				"error", err.Error())
			continue
		}

		for _, torrent := range torrents {
			stats.Total++
			stats.TotalSize += torrent.Size
			stats.TotalCompletedSize += torrent.Completed
			stats.TotalUploaded += torrent.Uploaded
			stats.TotalDownloaded += torrent.Downloaded

			// 计算总分享率
			if torrent.Ratio > 0 {
				totalRatioSum += torrent.Ratio
				ratioCount++
			}

			// 根据状态计数
			switch torrent.Status {
			case TorrentStatusDownloading:
				stats.Downloading++
				stats.Active++
				stats.CurrentDownloadSpeed += torrent.DownloadSpeed
				stats.CurrentUploadSpeed += torrent.UploadSpeed
			case TorrentStatusSeeding:
				stats.Seeding++
				stats.Active++
				stats.CurrentUploadSpeed += torrent.UploadSpeed
			case TorrentStatusPaused:
				stats.Paused++
			case TorrentStatusCompleted:
				stats.Completed++
			case TorrentStatusError:
				stats.Error++
			case TorrentStatusChecking, TorrentStatusQueued:
				stats.Active++
			}
		}
	}

	// 计算平均分享率
	if ratioCount > 0 {
		stats.TotalRatio = totalRatioSum / float64(ratioCount)
	}

	log.Info("Got torrent statistics", 
		"total", stats.Total,
		"active", stats.Active,
		"downloading", stats.Downloading,
		"seeding", stats.Seeding)

	return stats, nil
}

// GetTorrentFiles 获取种子文件列表
func (m *torrentManager) GetTorrentFiles(ctx context.Context, id, downloaderID string) (*TorrentFilesResponse, error) {
	log := m.logger.WithContext(ctx)
	log.Debug("Getting torrent files", "torrent_id", id, "downloader_id", downloaderID)

	// 获取种子信息
	torrent, err := m.GetTorrentByID(ctx, id, downloaderID)
	if err != nil {
		return nil, err
	}

	// 返回文件列表
	response := &TorrentFilesResponse{
		Success: true,
		Files: torrent.Files,
		Total: len(torrent.Files),
		TorrentID: torrent.ID,
		TorrentHash: torrent.Hash,
	}

	log.Info("Got torrent files", 
		"torrent_id", id,
		"file_count", len(torrent.Files))

	return response, nil
}

// GetTorrentTrackers 获取种子追踪器信息
func (m *torrentManager) GetTorrentTrackers(ctx context.Context, id, downloaderID string) (*TorrentTrackersResponse, error) {
	log := m.logger.WithContext(ctx)
	log.Debug("Getting torrent trackers", "torrent_id", id, "downloader_id", downloaderID)

	// 获取种子信息
	torrent, err := m.GetTorrentByID(ctx, id, downloaderID)
	if err != nil {
		return nil, err
	}

	// 返回追踪器列表
	response := &TorrentTrackersResponse{
		Success: true,
		Trackers: torrent.Trackers,
		Total: len(torrent.Trackers),
		TorrentID: torrent.ID,
		TorrentHash: torrent.Hash,
	}

	log.Info("Got torrent trackers", 
		"torrent_id", id,
		"tracker_count", len(torrent.Trackers))

	return response, nil
}

// GetCategories 获取分类列表
func (m *torrentManager) GetCategories(ctx context.Context, downloaderID string) (*CategoryResponse, error) {
	log := m.logger.WithContext(ctx)
	log.Debug("Getting torrent categories", "downloader_id", downloaderID)

	// 获取所有下载器插件
	plugins, err := m.getDownloaderPlugins(ctx, "", downloaderID)
	if err != nil {
		log.Error("Failed to get downloader plugins", "error", err.Error())
		return nil, fmt.Errorf("获取下载器插件失败: %w", err)
	}

	// 统计分类信息
	categories := make(map[string]*CategoryInfo)

	for _, p := range plugins {
		torrents, err := m.fetchTorrentsFromPlugin(ctx, p)
		if err != nil {
			log.Error("Failed to fetch torrents from plugin", 
				"plugin_id", p.ID(), 
				"error", err.Error())
			continue
		}

		for _, torrent := range torrents {
			if torrent.Category == "" {
				continue
			}

			key := fmt.Sprintf("%s:%s", p.ID(), torrent.Category)
			if _, exists := categories[key]; !exists {
				categories[key] = &CategoryInfo{
					Name: torrent.Category,
					SavePath: torrent.SavePath,
					Count: 0,
					DownloaderID: p.ID(),
				}
			}
			categories[key].Count++
		}
	}

	// 转换为列表
	var categoryList []*CategoryInfo
	for _, cat := range categories {
		categoryList = append(categoryList, cat)
	}

	// 排序
	sort.Slice(categoryList, func(i, j int) bool {
		return categoryList[i].Name < categoryList[j].Name
	})

	// 返回结果
	response := &CategoryResponse{
		Success: true,
		Categories: categoryList,
		Total: len(categoryList),
	}

	log.Info("Got torrent categories", "count", len(categoryList))

	return response, nil
}

// GetDownloaderStatus 获取下载器状态
func (m *torrentManager) GetDownloaderStatus(ctx context.Context, downloaderID string) (*TorrentManagerStatus, error) {
	log := m.logger.WithContext(ctx)
	log.Debug("Getting downloader status", "downloader_id", downloaderID)

	// 获取下载器插件
	plugins, err := m.getDownloaderPlugins(ctx, "", downloaderID)
	if err != nil {
		log.Error("Failed to get downloader plugins", "error", err.Error())
		return nil, fmt.Errorf("获取下载器插件失败: %w", err)
	}

	if len(plugins) == 0 {
		return nil, fmt.Errorf("未找到下载器: %s", downloaderID)
	}

	// 获取第一个匹配的下载器状态
	plugin := plugins[0]
	status, err := m.getPluginStatus(ctx, plugin)
	if err != nil {
		log.Error("Failed to get plugin status", "plugin_id", plugin.ID(), "error", err.Error())
		return nil, fmt.Errorf("获取下载器状态失败: %w", err)
	}

	log.Info("Got downloader status", "downloader_id", downloaderID, "connected", status.Connected)

	return status, nil
}

// GetAvailableDownloaders 获取可用的下载器列表
func (m *torrentManager) GetAvailableDownloaders(ctx context.Context) ([]*TorrentManagerStatus, error) {
	log := m.logger.WithContext(ctx)
	log.Debug("Getting available downloaders")

	// 获取所有下载器插件
	plugins, err := m.getDownloaderPlugins(ctx, "", "")
	if err != nil {
		log.Error("Failed to get downloader plugins", "error", err.Error())
		return nil, fmt.Errorf("获取下载器插件失败: %w", err)
	}

	// 获取所有下载器状态
	var statusList []*TorrentManagerStatus
	for _, p := range plugins {
		status, err := m.getPluginStatus(ctx, p)
		if err != nil {
			log.Error("Failed to get plugin status", "plugin_id", p.ID(), "error", err.Error())
			continue
		}
		statusList = append(statusList, status)
	}

	log.Info("Got available downloaders", "count", len(statusList))

	return statusList, nil
}

// 辅助方法：获取下载器插件
func (m *torrentManager) getDownloaderPlugins(ctx context.Context, pluginType, pluginID string) ([]plugin.Plugin, error) {
	log := m.logger.WithContext(ctx)

	// 插件类型固定为downloader
	if pluginType == "" {
		pluginType = "downloader"
	}

	// 获取插件列表
	plugins, err := m.pluginManager.GetPluginsByType(pluginType)
	if err != nil {
		log.Error("Failed to get plugins by type", "type", pluginType, "error", err.Error())
		return nil, err
	}

	// 如果指定了插件ID，则过滤
	if pluginID != "" {
		var filtered []plugin.Plugin
		for _, p := range plugins {
			if p.ID() == pluginID {
				filtered = append(filtered, p)
				break
			}
		}
		plugins = filtered
	}

	return plugins, nil
}

// 辅助方法：从插件获取种子
func (m *torrentManager) fetchTorrentsFromPlugin(ctx context.Context, p plugin.Plugin) ([]*TorrentItem, error) {
	log := m.logger.WithContext(ctx)
	log.Debug("Fetching torrents from plugin", "plugin_id", p.ID(), "plugin_name", p.Name())

	// 通过插件接口获取种子数据
	// 这里需要根据实际的插件接口实现
	// 暂时返回模拟数据
	// TODO: 实现真实的插件调用逻辑
	
	return []*TorrentItem{}, nil
}

// 辅助方法：获取插件状态
func (m *torrentManager) getPluginStatus(ctx context.Context, p plugin.Plugin) (*TorrentManagerStatus, error) {
	log := m.logger.WithContext(ctx)
	log.Debug("Getting plugin status", "plugin_id", p.ID(), "plugin_name", p.Name())

	// TODO: 实现真实的插件状态获取逻辑
	return &TorrentManagerStatus{
		DownloaderType: DownloaderType(p.ID()),
		DownloaderID: p.ID(),
		Connected: true,
		Version: "1.0.0",
		Status: "online",
		Message: "Connected and working properly",
		Stats: &TorrentStats{},
		LastConnectionTime: time.Now(),
	}, nil
}

// 辅助方法：过滤种子
func (m *torrentManager) filterTorrents(torrents []*TorrentItem, params *FetchTorrentsParams) []*TorrentItem {
	var filtered []*TorrentItem

	for _, torrent := range torrents {
		// 状态过滤
		if params.Status != "" && string(torrent.Status) != params.Status {
			continue
		}

		// 分类过滤
		if params.Category != "" && torrent.Category != params.Category {
			continue
		}

		// 搜索过滤
		if params.Search != "" {
			searchLower := strings.ToLower(params.Search)
			nameLower := strings.ToLower(torrent.Name)
			if !strings.Contains(nameLower, searchLower) {
				// 尝试使用正则表达式匹配
				matched, err := regexp.MatchString(params.Search, torrent.Name)
				if err != nil || !matched {
					continue
				}
			}
		}

		// 标签过滤
		if len(params.Tags) > 0 {
			tagMatch := false
			torrentTags := make(map[string]bool)
			for _, tag := range torrent.Tags {
				torrentTags[strings.ToLower(tag)] = true
			}

			for _, tag := range params.Tags {
				if torrentTags[strings.ToLower(tag)] {
					tagMatch = true
					break
				}
			}

			if !tagMatch {
				continue
			}
		}

		// 活动状态过滤
		if params.OnlyActive {
			if torrent.Status != TorrentStatusDownloading && 
			   torrent.Status != TorrentStatusSeeding && 
			   torrent.Status != TorrentStatusChecking && 
			   torrent.Status != TorrentStatusQueued {
				continue
			}
		}

		// 完成状态过滤
		if params.OnlyCompleted && torrent.Status != TorrentStatusCompleted {
			continue
		}

		// 下载中过滤
		if params.OnlyDownloading && torrent.Status != TorrentStatusDownloading {
			continue
		}

		// 做种中过滤
		if params.OnlySeeding && torrent.Status != TorrentStatusSeeding {
			continue
		}

		// 已暂停过滤
		if params.OnlyPaused && torrent.Status != TorrentStatusPaused {
			continue
		}

		filtered = append(filtered, torrent)
	}

	return filtered
}

// 辅助方法：排序种子
func (m *torrentManager) sortTorrents(torrents []*TorrentItem, sortBy TorrentSortField, sortOrder SortOrder) []*TorrentItem {
	// 创建副本以避免修改原始数据
	sorted := make([]*TorrentItem, len(torrents))
	copy(sorted, torrents)

	// 排序
	sort.Slice(sorted, func(i, j int) bool {
		iTorrent := sorted[i]
		jTorrent := sorted[j]

		var less bool

		switch sortBy {
		case TorrentSortFieldName:
			less = iTorrent.Name < jTorrent.Name
		case TorrentSortFieldSize:
			less = iTorrent.Size < jTorrent.Size
		case TorrentSortFieldProgress:
			less = iTorrent.Progress < jTorrent.Progress
		case TorrentSortFieldAdded:
			less = iTorrent.AddedAt.Before(jTorrent.AddedAt)
		case TorrentSortFieldCompleted:
			less = iTorrent.CompletedAt.Before(jTorrent.CompletedAt)
		case TorrentSortFieldRatio:
			less = iTorrent.Ratio < jTorrent.Ratio
		case TorrentSortFieldSeeds:
			less = iTorrent.Seeds < jTorrent.Seeds
		case TorrentSortFieldPeers:
			less = iTorrent.Peers < jTorrent.Peers
		case TorrentSortFieldStatus:
			// 按状态优先级排序
			statusPriority := map[TorrentStatus]int{
				TorrentStatusError:      0,
				TorrentStatusDownloading: 1,
				TorrentStatusSeeding:    2,
				TorrentStatusChecking:   3,
				TorrentStatusQueued:     4,
				TorrentStatusCompleted:  5,
				TorrentStatusPaused:     6,
				TorrentStatusUnknown:    7,
			}
			less = statusPriority[iTorrent.Status] < statusPriority[jTorrent.Status]
		default:
			// 默认按添加时间排序
			less = iTorrent.AddedAt.Before(jTorrent.AddedAt)
		}

		// 根据排序顺序反转结果
		if sortOrder == SortOrderDesc {
			less = !less
		}

		return less
	})

	return sorted
}

// 辅助方法：分页种子
func (m *torrentManager) paginateTorrents(torrents []*TorrentItem, limit, offset int) []*TorrentItem {
	start := offset
	end := offset + limit

	if start >= len(torrents) {
		return []*TorrentItem{}
	}

	if end > len(torrents) {
		end = len(torrents)
	}

	return torrents[start:end]
}
