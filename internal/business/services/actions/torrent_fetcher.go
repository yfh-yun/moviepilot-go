// Package actions 种子获取相关业务逻辑
package actions

import (
	"context"
	"fmt"
	"time"

	"moviepilot-go/pkg/logger"
	"moviepilot-go/internal/repositories"
	"moviepilot-go/internal/business/services"
)

// FetchTorrentsAction 种子获取动作
type FetchTorrentsAction struct {
	torrentRepo repository.TorrentRepository
	userRepo    repository.UserRepository
	pluginMgr   service.PluginManager
	logger      logger.Logger
}

// NewFetchTorrentsAction 创建种子获取动作实例
func NewFetchTorrentsAction(
	torrentRepo repository.TorrentRepository,
	userRepo repository.UserRepository,
	pluginMgr service.PluginManager,
	logger logger.Logger,
) *FetchTorrentsAction {
	return &FetchTorrentsAction{
		torrentRepo: torrentRepo,
		userRepo:    userRepo,
		pluginMgr:   pluginMgr,
		logger:      logger,
	}
}

// Execute 执行种子获取动作
func (a *FetchTorrentsAction) Execute(ctx context.Context, req *FetchTorrentsRequest) (*FetchTorrentsResponse, error) {
	a.logger.Info("开始执行种子获取动作",
		logger.String("site", req.Site),
		logger.String("keyword", req.Keyword),
		logger.String("user_id", req.UserID),
	)

	// 1. 验证用户权限
	if err := a.validateUserPermission(ctx, req.UserID); err != nil {
		a.logger.Error("用户权限验证失败", logger.Error(err))
		return nil, fmt.Errorf("用户权限验证失败: %w", err)
	}

	// 2. 获取站点配置
	siteConfig, err := a.getSiteConfig(req.Site)
	if err != nil {
		a.logger.Error("获取站点配置失败", 
			logger.String("site", req.Site),
			logger.Error(err))
		return nil, fmt.Errorf("获取站点配置失败: %w", err)
	}

	// 3. 执行种子搜索
	torrents, err := a.searchTorrents(ctx, siteConfig, req)
	if err != nil {
		a.logger.Error("种子搜索失败", logger.Error(err))
		return nil, fmt.Errorf("种子搜索失败: %w", err)
	}

	// 4. 保存搜索结果
	if err := a.saveSearchResults(ctx, torrents, req); err != nil {
		a.logger.Error("保存搜索结果失败", logger.Error(err))
		return nil, fmt.Errorf("保存搜索结果失败: %w", err)
	}

	// 5. 返回结果
	response := &FetchTorrentsResponse{
		Success:    true,
		Torrents:   torrents,
		Total:      len(torrents),
		Message:    "种子获取完成",
		Site:       req.Site,
		Keyword:    req.Keyword,
		SearchedAt: time.Now(),
	}

	a.logger.Info("种子获取动作执行完成",
		logger.String("site", req.Site),
		logger.Int("total", response.Total),
	)

	return response, nil
}

// FetchTorrentsRequest 种子获取请求
type FetchTorrentsRequest struct {
	UserID  string `json:"user_id" validate:"required"`
	Site    string `json:"site" validate:"required"`
	Keyword string `json:"keyword"`
	Category string `json:"category"`
	SizeMin int64  `json:"size_min"`
	SizeMax int64  `json:"size_max"`
	PerPage int    `json:"per_page"`
	Page    int    `json:"page"`
}

// FetchTorrentsResponse 种子获取响应
type FetchTorrentsResponse struct {
	Success    bool                 `json:"success"`
	Torrents   []repository.Torrent `json:"torrents"`
	Total      int                  `json:"total"`
	Message    string               `json:"message"`
	Site       string               `json:"site"`
	Keyword    string               `json:"keyword"`
	SearchedAt time.Time            `json:"searched_at"`
}

// validateUserPermission 验证用户权限
func (a *FetchTorrentsAction) validateUserPermission(ctx context.Context, userID string) error {
	user, err := a.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("获取用户信息失败: %w", err)
	}

	if !user.IsActive() {
		return fmt.Errorf("用户已被禁用")
	}

	if !user.HasPermission("torrent_search") {
		return fmt.Errorf("用户无搜索种子权限")
	}

	return nil
}

// getSiteConfig 获取站点配置
func (a *FetchTorrentsAction) getSiteConfig(site string) (*SiteConfig, error) {
	// 从插件或配置中获取站点信息
	config, err := a.pluginMgr.GetSiteConfig(site)
	if err != nil {
		return nil, err
	}

	return &SiteConfig{
		Name:      config.Name,
		BaseURL:   config.BaseURL,
		SearchURL: config.SearchURL,
		Headers:   config.Headers,
		Cookies:   config.Cookies,
		Enabled:   config.Enabled,
	}, nil
}

// SiteConfig 站点配置
type SiteConfig struct {
	Name      string            `json:"name"`
	BaseURL   string            `json:"base_url"`
	SearchURL string            `json:"search_url"`
	Headers   map[string]string `json:"headers"`
	Cookies   map[string]string `json:"cookies"`
	Enabled   bool              `json:"enabled"`
}

// searchTorrents 搜索种子
func (a *FetchTorrentsAction) searchTorrents(ctx context.Context, siteConfig *SiteConfig, req *FetchTorrentsRequest) ([]repository.Torrent, error) {
	// 调用插件进行种子搜索
	searchParams := map[string]interface{}{
		"keyword":  req.Keyword,
		"category": req.Category,
		"size_min": req.SizeMin,
		"size_max": req.SizeMax,
		"per_page": req.PerPage,
		"page":     req.Page,
	}

	result, err := a.pluginMgr.CallPlugin(ctx, "site", "search_torrents", map[string]interface{}{
		"site":   req.Site,
		"config": siteConfig,
		"params": searchParams,
	})

	if err != nil {
		return nil, fmt.Errorf("插件调用失败: %w", err)
	}

	// 转换搜索结果
	torrents, ok := result["torrents"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("插件返回格式错误")
	}

	var resultTorrents []repository.Torrent
	for _, torrent := range torrents {
		torrentMap, ok := torrent.(map[string]interface{})
		if !ok {
			continue
		}

		torrent := repository.Torrent{
			ID:          getString(torrentMap, "id"),
			Title:       getString(torrentMap, "title"),
			URL:         getString(torrentMap, "url"),
			DownloadURL: getString(torrentMap, "download_url"),
			Size:        getInt64(torrentMap, "size"),
			Category:    getString(torrentMap, "category"),
			Site:        req.Site,
			UploadAt:    time.Now(),
			CreatedAt:   time.Now(),
		}

		resultTorrents = append(resultTorrents, torrent)
	}

	return resultTorrents, nil
}

// saveSearchResults 保存搜索结果
func (a *FetchTorrentsAction) saveSearchResults(ctx context.Context, torrents []repository.Torrent, req *FetchTorrentsRequest) error {
	for _, torrent := range torrents {
		// 检查是否已存在
		existing, err := a.torrentRepo.GetBySiteAndID(ctx, torrent.Site, torrent.ID)
		if err == nil && existing != nil {
			// 更新现有记录
			torrent.ID = existing.ID
			if err := a.torrentRepo.Update(ctx, &torrent); err != nil {
				a.logger.Error("更新种子记录失败", 
					logger.String("torrent_id", torrent.ID),
					logger.Error(err))
			}
		} else {
			// 创建新记录
			if err := a.torrentRepo.Create(ctx, &torrent); err != nil {
				a.logger.Error("创建种子记录失败", 
					logger.String("torrent_id", torrent.ID),
					logger.Error(err))
			}
		}
	}

	return nil
}

// GetString 工具函数：从map中获取string值
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// GetInt64 工具函数：从map中获取int64值
func getInt64(m map[string]interface{}, key string) int64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int64:
			return v
		case float64:
			return int64(v)
		case int:
			return int64(v)
		}
	}
	return 0
}