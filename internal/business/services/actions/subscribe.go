// Package actions 提供订阅管理相关的功能实现
package actions

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"moviepilot-go/internal/business/services/actions/types"
	"moviepilot-go/internal/repositories/interfaces"
	"moviepilot-go/pkg/logger"
)

// SubscribeManager 订阅管理器接口
type SubscribeManager interface {
	// 添加订阅
	AddSubscribe(ctx context.Context, params *AddSubscribeParams) (*SubscribeResult, error)
	// 获取订阅
	GetSubscribe(ctx context.Context, subscribeID string) (*SubscribeInfo, error)
	// 更新订阅
	UpdateSubscribe(ctx context.Context, subscribeID string, params *UpdateSubscribeParams) (*SubscribeResult, error)
	// 删除订阅
	DeleteSubscribe(ctx context.Context, subscribeID string) error
	// 暂停订阅
	PauseSubscribe(ctx context.Context, subscribeID string) error
	// 恢复订阅
	ResumeSubscribe(ctx context.Context, subscribeID string) error
	// 获取订阅列表
	ListSubscribes(ctx context.Context, filter *SubscribeFilter) ([]*SubscribeInfo, int64, error)
	// 更新订阅项
	UpdateSubscribeItems(ctx context.Context, subscribeID string) ([]*SubscribeItem, error)
	// 获取订阅项
	GetSubscribeItems(ctx context.Context, filter *SubscribeItemFilter) ([]*SubscribeItem, int64, error)
}

// SubscribeManagerImpl 订阅管理器实现
type SubscribeManagerImpl struct {
	logger            logger.Logger
	subscribeRepo     interfaces.SubscribeRepository
	downloadManager   DownloadManager
	rssFetcher        types.RSSFetcher
	torrentFetcher    types.TorrentFetcher
	mediaFetcher      types.MediaFetcher
}

// NewSubscribeManager 创建订阅管理器实例
func NewSubscribeManager(
	subscribeRepo interfaces.SubscribeRepository,
	downloadManager DownloadManager,
	rssFetcher types.RSSFetcher,
	torrentFetcher types.TorrentFetcher,
	mediaFetcher types.MediaFetcher,
) *SubscribeManagerImpl {
	return &SubscribeManagerImpl{
		logger:          logger.NewLogger("subscribe_manager"),
		subscribeRepo:   subscribeRepo,
		downloadManager: downloadManager,
		rssFetcher:      rssFetcher,
		torrentFetcher:  torrentFetcher,
		mediaFetcher:    mediaFetcher,
	}
}

// AddSubscribe 添加订阅
func (m *SubscribeManagerImpl) AddSubscribe(ctx context.Context, params *AddSubscribeParams) (*SubscribeResult, error) {
	m.logger.Debug("添加订阅", "name", params.Name, "type", params.Config.Type)

	// 验证参数
	if err := m.validateSubscribeConfig(&params.Config); err != nil {
		m.logger.Error("订阅配置验证失败", "error", err.Error())
		return nil, fmt.Errorf("订阅配置验证失败: %w", err)
	}

	// 生成订阅ID
	subscribeID := uuid.New().String()

	// 设置默认值
	now := time.Now()
	params.Config.CreatedAt = now
	if params.Config.UpdateInterval <= 0 {
		params.Config.UpdateInterval = 60 // 默认60分钟
	}

	// 计算下次更新时间
	nextUpdate := now.Add(time.Duration(params.Config.UpdateInterval) * time.Minute)
	params.Config.NextUpdate = &nextUpdate

	// 如果未指定状态，设置为活跃
	if params.Config.Status == "" {
		params.Config.Status = SubscribeStatusActive
	}

	// 创建订阅信息
	subscribeInfo := &SubscribeInfo{
		ID:          subscribeID,
		Name:        params.Name,
		Type:        params.Config.Type,
		Status:      params.Config.Status,
		Description: params.Description,
		Config:      &params.Config,
		Stats: &SubscribeStats{
			TotalItems:     0,
			DownloadedItems: 0,
			FailedItems:    0,
			LastSuccess:    now,
			LastUpdate:     now,
			AverageSize:    0,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	// 保存到数据库
	if err := m.subscribeRepo.Create(ctx, subscribeInfo); err != nil {
		m.logger.Error("保存订阅失败", "error", err.Error())
		return nil, fmt.Errorf("保存订阅失败: %w", err)
	}

	// 如果是活跃状态，立即尝试更新
	if params.Config.Status == SubscribeStatusActive {
		go m.UpdateSubscribeItems(ctx, subscribeID)
	}

	result := &SubscribeResult{
		SubscribeID:    subscribeID,
		Success:        true,
		Message:        "订阅添加成功",
		SubscribeConfig: &params.Config,
		CreatedAt:      now,
	}

	m.logger.Info("订阅添加成功", "subscribe_id", subscribeID, "name", params.Name)
	return result, nil
}

// GetSubscribe 获取订阅
func (m *SubscribeManagerImpl) GetSubscribe(ctx context.Context, subscribeID string) (*SubscribeInfo, error) {
	m.logger.Debug("获取订阅信息", "subscribe_id", subscribeID)

	if subscribeID == "" {
		return nil, errors.New("订阅ID不能为空")
	}

	subscribe, err := m.subscribeRepo.GetByID(ctx, subscribeID)
	if err != nil {
		if errors.Is(err, interfaces.ErrNotFound) {
			return nil, fmt.Errorf("订阅不存在: %s", subscribeID)
		}
		m.logger.Error("获取订阅失败", "error", err.Error(), "subscribe_id", subscribeID)
		return nil, fmt.Errorf("获取订阅失败: %w", err)
	}

	return subscribe, nil
}

// UpdateSubscribe 更新订阅
func (m *SubscribeManagerImpl) UpdateSubscribe(ctx context.Context, subscribeID string, params *UpdateSubscribeParams) (*SubscribeResult, error) {
	m.logger.Debug("更新订阅", "subscribe_id", subscribeID)

	// 获取现有订阅
	subscribe, err := m.GetSubscribe(ctx, subscribeID)
	if err != nil {
		return nil, err
	}

	// 更新字段
	if params.Name != "" {
		subscribe.Name = params.Name
	}
	if params.Description != "" {
		subscribe.Description = params.Description
	}
	if params.Status != nil {
		subscribe.Status = *params.Status
		subscribe.Config.Status = *params.Status
	}

	// 更新配置
	if params.Config.Type != "" {
		subscribe.Type = params.Config.Type
		subscribe.Config.Type = params.Config.Type
	}
	if params.Config.UpdateInterval > 0 {
		subscribe.Config.UpdateInterval = params.Config.UpdateInterval
		// 重新计算下次更新时间
		nextUpdate := time.Now().Add(time.Duration(params.Config.UpdateInterval) * time.Minute)
		subscribe.Config.NextUpdate = &nextUpdate
	}

	// 更新其他配置字段
	if params.Config.AutoDownload {
		subscribe.Config.AutoDownload = params.Config.AutoDownload
	}
	if params.Config.SavePath != "" {
		subscribe.Config.SavePath = params.Config.SavePath
	}
	if params.Config.Downloader != "" {
		subscribe.Config.Downloader = params.Config.Downloader
	}
	if len(params.Config.Labels) > 0 {
		subscribe.Config.Labels = params.Config.Labels
	}
	if len(params.Config.Categories) > 0 {
		subscribe.Config.Categories = params.Config.Categories
	}
	if len(params.Config.ExcludeWords) > 0 {
		subscribe.Config.ExcludeWords = params.Config.ExcludeWords
	}

	// 更新时间
	subscribe.UpdatedAt = time.Now()

	// 验证更新后的配置
	if err := m.validateSubscribeConfig(subscribe.Config); err != nil {
		m.logger.Error("更新后的订阅配置验证失败", "error", err.Error())
		return nil, fmt.Errorf("订阅配置验证失败: %w", err)
	}

	// 保存更新
	if err := m.subscribeRepo.Update(ctx, subscribe); err != nil {
		m.logger.Error("更新订阅失败", "error", err.Error(), "subscribe_id", subscribeID)
		return nil, fmt.Errorf("更新订阅失败: %w", err)
	}

	result := &SubscribeResult{
		SubscribeID:    subscribeID,
		Success:        true,
		Message:        "订阅更新成功",
		SubscribeConfig: subscribe.Config,
		CreatedAt:      subscribe.CreatedAt,
	}

	m.logger.Info("订阅更新成功", "subscribe_id", subscribeID, "name", subscribe.Name)
	return result, nil
}

// DeleteSubscribe 删除订阅
func (m *SubscribeManagerImpl) DeleteSubscribe(ctx context.Context, subscribeID string) error {
	m.logger.Debug("删除订阅", "subscribe_id", subscribeID)

	// 检查订阅是否存在
	_, err := m.GetSubscribe(ctx, subscribeID)
	if err != nil {
		return err
	}

	// 删除订阅
	if err := m.subscribeRepo.Delete(ctx, subscribeID); err != nil {
		m.logger.Error("删除订阅失败", "error", err.Error(), "subscribe_id", subscribeID)
		return fmt.Errorf("删除订阅失败: %w", err)
	}

	// 删除相关的订阅项
	if err := m.subscribeRepo.DeleteItems(ctx, subscribeID); err != nil {
		m.logger.Warn("删除订阅项失败", "error", err.Error(), "subscribe_id", subscribeID)
		// 不返回错误，因为主订阅已删除成功
	}

	m.logger.Info("订阅删除成功", "subscribe_id", subscribeID)
	return nil
}

// PauseSubscribe 暂停订阅
func (m *SubscribeManagerImpl) PauseSubscribe(ctx context.Context, subscribeID string) error {
	m.logger.Debug("暂停订阅", "subscribe_id", subscribeID)

	status := SubscribeStatusPaused
	params := &UpdateSubscribeParams{
		Status: &status,
	}

	_, err := m.UpdateSubscribe(ctx, subscribeID, params)
	if err != nil {
		return err
	}

	m.logger.Info("订阅暂停成功", "subscribe_id", subscribeID)
	return nil
}

// ResumeSubscribe 恢复订阅
func (m *SubscribeManagerImpl) ResumeSubscribe(ctx context.Context, subscribeID string) error {
	m.logger.Debug("恢复订阅", "subscribe_id", subscribeID)

	status := SubscribeStatusActive
	params := &UpdateSubscribeParams{
		Status: &status,
	}

	result, err := m.UpdateSubscribe(ctx, subscribeID, params)
	if err != nil {
		return err
	}

	// 恢复后立即更新订阅项
	go m.UpdateSubscribeItems(ctx, subscribeID)

	m.logger.Info("订阅恢复成功", "subscribe_id", subscribeID)
	return nil
}

// ListSubscribes 获取订阅列表
func (m *SubscribeManagerImpl) ListSubscribes(ctx context.Context, filter *SubscribeFilter) ([]*SubscribeInfo, int64, error) {
	m.logger.Debug("获取订阅列表", "filter", filter)

	// 设置默认值
	if filter == nil {
		filter = &SubscribeFilter{}
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100 // 最大限制
	}

	subscribes, total, err := m.subscribeRepo.List(ctx, filter)
	if err != nil {
		m.logger.Error("获取订阅列表失败", "error", err.Error())
		return nil, 0, fmt.Errorf("获取订阅列表失败: %w", err)
	}

	return subscribes, total, nil
}

// UpdateSubscribeItems 更新订阅项
func (m *SubscribeManagerImpl) UpdateSubscribeItems(ctx context.Context, subscribeID string) ([]*SubscribeItem, error) {
	m.logger.Debug("更新订阅项", "subscribe_id", subscribeID)

	// 获取订阅信息
	subscribe, err := m.GetSubscribe(ctx, subscribeID)
	if err != nil {
		return nil, err
	}

	// 检查订阅状态
	if subscribe.Status != SubscribeStatusActive {
		return nil, fmt.Errorf("订阅状态不允许更新: %s", subscribe.Status)
	}

	var items []*SubscribeItem
	var fetchErr error

	// 根据订阅类型获取订阅项
	switch subscribe.Type {
	case SubscribeTypeRSS:
		items, fetchErr = m.fetchRSSItems(ctx, subscribe)
	case SubscribeTypeTorrent:
		items, fetchErr = m.fetchTorrentItems(ctx, subscribe)
	case SubscribeTypeMedia:
		items, fetchErr = m.fetchMediaItems(ctx, subscribe)
	case SubscribeTypeKeyword:
		items, fetchErr = m.fetchKeywordItems(ctx, subscribe)
	default:
		return nil, fmt.Errorf("不支持的订阅类型: %s", subscribe.Type)
	}

	if fetchErr != nil {
		// 更新错误信息
		subscribe.Config.ErrorCount++
		subscribe.Config.LastError = fetchErr.Error()
		subscribe.Config.Status = SubscribeStatusError
		m.subscribeRepo.Update(ctx, subscribe)
		m.logger.Error("获取订阅项失败", "error", fetchErr.Error(), "subscribe_id", subscribeID)
		return nil, fmt.Errorf("获取订阅项失败: %w", fetchErr)
	}

	// 重置错误状态
	subscribe.Config.ErrorCount = 0
	subscribe.Config.LastError = ""
	subscribe.Config.Status = SubscribeStatusActive

	// 过滤订阅项
	filteredItems := m.filterSubscribeItems(items, subscribe.Config)

	// 保存订阅项并处理自动下载
	newItems := make([]*SubscribeItem, 0)
	now := time.Now()
	for _, item := range filteredItems {
		item.SubscribeID = subscribeID
		item.CreatedAt = now
		item.UpdatedAt = now

		// 检查是否已存在
		existing, _ := m.subscribeRepo.GetItemByHash(ctx, item.Hash)
		if existing == nil {
			// 保存新项
			if err := m.subscribeRepo.CreateItem(ctx, item); err != nil {
				m.logger.Error("保存订阅项失败", "error", err.Error(), "title", item.Title)
				continue
			}
			newItems = append(newItems, item)

			// 自动下载
			if subscribe.Config.AutoDownload {
				go m.downloadSubscribeItem(ctx, item, subscribe.Config)
			}
		}
	}

	// 更新订阅统计
	subscribe.Stats.TotalItems += len(newItems)
	subscribe.Stats.LastUpdate = now
	subscribe.Config.LastUpdate = &now

	// 计算下次更新时间
	nextUpdate := now.Add(time.Duration(subscribe.Config.UpdateInterval) * time.Minute)
	subscribe.Config.NextUpdate = &nextUpdate

	// 保存更新
	m.subscribeRepo.Update(ctx, subscribe)

	m.logger.Info("订阅项更新成功", "subscribe_id", subscribeID, "new_items", len(newItems), "total_items", subscribe.Stats.TotalItems)
	return newItems, nil
}

// GetSubscribeItems 获取订阅项
func (m *SubscribeManagerImpl) GetSubscribeItems(ctx context.Context, filter *SubscribeItemFilter) ([]*SubscribeItem, int64, error) {
	m.logger.Debug("获取订阅项", "filter", filter)

	// 设置默认值
	if filter == nil {
		filter = &SubscribeItemFilter{}
	}
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Limit > 200 {
		filter.Limit = 200 // 最大限制
	}
	if filter.OrderBy == "" {
		filter.OrderBy = "publish_date"
	}
	if filter.OrderDir == "" {
		filter.OrderDir = "desc"
	}

	items, total, err := m.subscribeRepo.ListItems(ctx, filter)
	if err != nil {
		m.logger.Error("获取订阅项失败", "error", err.Error())
		return nil, 0, fmt.Errorf("获取订阅项失败: %w", err)
	}

	return items, total, nil
}

// 私有方法

// validateSubscribeConfig 验证订阅配置
func (m *SubscribeManagerImpl) validateSubscribeConfig(config *SubscribeConfig) error {
	// 验证订阅类型
	if config.Type == "" {
		return errors.New("订阅类型不能为空")
	}

	// 根据类型验证必填字段
	switch config.Type {
	case SubscribeTypeRSS:
		if config.URL == "" {
			return errors.New("RSS订阅必须提供URL")
		}
	case SubscribeTypeTorrent:
		if config.TorrentHash == "" {
			return errors.New("种子订阅必须提供种子哈希")
		}
	case SubscribeTypeMedia:
		if config.MediaID == "" {
			return errors.New("媒体订阅必须提供媒体ID")
		}
	case SubscribeTypeKeyword:
		if len(config.Keywords) == 0 {
			return errors.New("关键词订阅必须提供关键词")
		}
	case SubscribeTypeUser:
		if config.UserID == "" {
			return errors.New("用户订阅必须提供用户ID")
		}
	default:
		return fmt.Errorf("不支持的订阅类型: %s", config.Type)
	}

	// 验证更新间隔
	if config.UpdateInterval < 5 {
		return errors.New("更新间隔不能小于5分钟")
	}

	// 验证下载配置
	if config.AutoDownload {
		if config.Downloader == "" {
			return errors.New("自动下载必须指定下载器")
		}
		if config.SavePath == "" {
			return errors.New("自动下载必须指定保存路径")
		}
	}

	return nil
}

// fetchRSSItems 获取RSS订阅项
func (m *SubscribeManagerImpl) fetchRSSItems(ctx context.Context, subscribe *SubscribeInfo) ([]*SubscribeItem, error) {
	if m.rssFetcher == nil {
		return nil, errors.New("RSS获取器未初始化")
	}

	items, err := m.rssFetcher.Fetch(ctx, subscribe.Config.URL)
	if err != nil {
		return nil, fmt.Errorf("获取RSS失败: %w", err)
	}

	// 转换为订阅项
	subscribeItems := make([]*SubscribeItem, 0, len(items))
	for _, item := range items {
		subscribeItems = append(subscribeItems, &SubscribeItem{
			ID:          uuid.New().String(),
			Title:       item.Title,
			Description: item.Description,
			URL:         item.Link,
			TorrentURL:  item.Enclosure,
			Magnet:      item.Magnet,
			Hash:        item.Hash,
			Size:        item.Size,
			Categories:  item.Categories,
			PublishDate: item.PubDate,
			Metadata:    item.Metadata,
		})
	}

	return subscribeItems, nil
}

// fetchTorrentItems 获取种子订阅项
func (m *SubscribeManagerImpl) fetchTorrentItems(ctx context.Context, subscribe *SubscribeInfo) ([]*SubscribeItem, error) {
	if m.torrentFetcher == nil {
		return nil, errors.New("种子获取器未初始化")
	}

	items, err := m.torrentFetcher.GetTorrentInfo(ctx, subscribe.Config.TorrentHash)
	if err != nil {
		return nil, fmt.Errorf("获取种子信息失败: %w", err)
	}

	// 转换为订阅项
	return []*SubscribeItem{
		{
			ID:          uuid.New().String(),
			Title:       items.Title,
			Description: items.Description,
			Hash:        items.Hash,
			Size:        items.Size,
			Categories:  items.Categories,
			PublishDate: time.Now(),
			Metadata:    items.Metadata,
		},
	}, nil
}

// fetchMediaItems 获取媒体订阅项
func (m *SubscribeManagerImpl) fetchMediaItems(ctx context.Context, subscribe *SubscribeInfo) ([]*SubscribeItem, error) {
	if m.mediaFetcher == nil {
		return nil, errors.New("媒体获取器未初始化")
	}

	items, err := m.mediaFetcher.GetMediaUpdates(ctx, subscribe.Config.MediaID)
	if err != nil {
		return nil, fmt.Errorf("获取媒体更新失败: %w", err)
	}

	// 转换为订阅项
	subscribeItems := make([]*SubscribeItem, 0, len(items))
	for _, item := range items {
		// 将媒体ID存储在元数据中
		if item.Metadata == nil {
			item.Metadata = make(map[string]interface{})
		}
		item.Metadata["media_id"] = item.ID

		subscribeItems = append(subscribeItems, &SubscribeItem{
			ID:          uuid.New().String(),
			Title:       item.Title,
			Description: item.Description,
			Hash:        item.Hash,
			Size:        item.Size,
			Categories:  item.Categories,
			PublishDate: item.ReleaseDate,
			Metadata:    item.Metadata,
		})
	}

	return subscribeItems, nil
}

// fetchKeywordItems 获取关键词订阅项
func (m *SubscribeManagerImpl) fetchKeywordItems(ctx context.Context, subscribe *SubscribeInfo) ([]*SubscribeItem, error) {
	if m.torrentFetcher == nil {
		return nil, errors.New("种子获取器未初始化")
	}

	items := make([]*SubscribeItem, 0)

	// 对每个关键词进行搜索
	for _, keyword := range subscribe.Config.Keywords {
		searchItems, err := m.torrentFetcher.Search(ctx, keyword, subscribe.Config.Categories, 50)
		if err != nil {
			m.logger.Warn("关键词搜索失败", "keyword", keyword, "error", err.Error())
			continue
		}

		// 转换并添加到结果
		for _, item := range searchItems {
			items = append(items, &SubscribeItem{
				ID:          uuid.New().String(),
				Title:       item.Title,
				Description: item.Description,
				TorrentURL:  item.TorrentURL,
				Magnet:      item.Magnet,
				Hash:        item.Hash,
				Size:        item.Size,
				Categories:  item.Categories,
				PublishDate: item.PublishDate,
				Metadata:    item.Metadata,
			})
		}
	}

	return items, nil
}

// filterSubscribeItems 过滤订阅项
func (m *SubscribeManagerImpl) filterSubscribeItems(items []*SubscribeItem, config *SubscribeConfig) []*SubscribeItem {
	filtered := make([]*SubscribeItem, 0)

	for _, item := range items {
		// 大小过滤
		if config.MinSize > 0 && item.Size < config.MinSize {
			continue
		}
		if config.MaxSize > 0 && item.Size > config.MaxSize {
			continue
		}

		// 分类过滤
		if len(config.Categories) > 0 {
			match := false
			for _, category := range item.Categories {
				for _, allowed := range config.Categories {
					if strings.EqualFold(category, allowed) {
						match = true
						break
					}
				}
				if match {
					break
				}
			}
			if !match {
				continue
			}
		}

		// 排除关键词过滤
		if len(config.ExcludeWords) > 0 {
			exclude := false
			lowerTitle := strings.ToLower(item.Title)
			for _, word := range config.ExcludeWords {
				if strings.Contains(lowerTitle, strings.ToLower(word)) {
					exclude = true
					break
				}
			}
			if exclude {
				continue
			}
		}

		// 质量过滤
		if config.Quality != "" && !strings.Contains(strings.ToLower(item.Title), strings.ToLower(config.Quality)) {
			continue
		}

		// 分辨率过滤
		if config.Resolution != "" && !strings.Contains(strings.ToLower(item.Title), strings.ToLower(config.Resolution)) {
			continue
		}

		filtered = append(filtered, item)
	}

	return filtered
}

// downloadSubscribeItem 下载订阅项
func (m *SubscribeManagerImpl) downloadSubscribeItem(ctx context.Context, item *SubscribeItem, config *SubscribeConfig) {
	m.logger.Debug("自动下载订阅项", "title", item.Title, "subscribe_id", item.SubscribeID)

	// 创建下载参数
	downloadParams := &AddDownloadParams{
		Downloader: config.Downloader,
		SavePath:   config.SavePath,
		Labels:     config.Labels,
		Quality:    config.Quality,
		Resolution: config.Resolution,
	}

	// 选择下载URL
	var downloadURL string
	if item.Magnet != "" {
		downloadURL = item.Magnet
	} else if item.TorrentURL != "" {
		downloadURL = item.TorrentURL
	} else {
		m.logger.Warn("订阅项没有可下载的URL", "title", item.Title)
		return
	}

	// 添加到下载
	result, err := m.downloadManager.AddDownload(ctx, downloadURL, downloadParams)
	if err != nil {
		m.logger.Error("订阅项下载失败", "error", err.Error(), "title", item.Title)
		return
	}

	// 更新订阅项状态
	item.DownloadID = result.DownloadID
	item.Downloaded = true
	item.UpdatedAt = time.Now()

	// 保存更新
	if err := m.subscribeRepo.UpdateItem(ctx, item); err != nil {
		m.logger.Error("更新订阅项下载状态失败", "error", err.Error(), "title", item.Title)
	}

	m.logger.Info("订阅项自动下载成功", "title", item.Title, "download_id", result.DownloadID)
}
