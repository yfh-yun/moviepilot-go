// Package spider 索引器Spider包
package spider

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// SpiderManager Spider管理器
type SpiderManager struct {
	spiders map[string]Spider
	configs map[string]*SiteConfig
	mutex   sync.RWMutex
}

// SiteConfig 站点配置
type SiteConfig struct {
	Name         string            `json:"name"`         // 站点名称
	URL          string            `json:"url"`          // 站点URL
	Type         string            `json:"type"`         // Spider类型
	Enabled      bool              `json:"enabled"`      // 是否启用
	Username     string            `json:"username"`     // 用户名
	Password     string            `json:"password"`     // 密码
	Cookie       string            `json:"cookie"`       // Cookie
	UserAgent    string            `json:"user_agent"`   // User-Agent
	Timeout      time.Duration     `json:"timeout"`      // 超时时间
	RetryCount   int               `json:"retry_count"`   // 重试次数
	RateLimit    time.Duration     `json:"rate_limit"`   // 请求频率限制
	Headers      map[string]string `json:"headers"`      // 自定义请求头
	Proxy        string            `json:"proxy"`        // 代理设置
	CustomParams map[string]string `json:"custom_params"` // 自定义参数
}

// NewSpiderManager 创建Spider管理器
func NewSpiderManager() *SpiderManager {
	return &SpiderManager{
		spiders: make(map[string]Spider),
		configs: make(map[string]*SiteConfig),
	}
}

// RegisterSpider 注册Spider
func (sm *SpiderManager) RegisterSpider(spiderType string, creator func(config *SiteConfig) (Spider, error)) {
	// 这里可以注册Spider创建函数
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	logger.Info("注册Spider类型", zap.String("type", spiderType))
}

// AddSite 添加站点
func (sm *SpiderManager) AddSite(siteID string, config *SiteConfig) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if config == nil {
		return fmt.Errorf("站点配置不能为空")
	}

	// 验证配置
	if err := sm.validateSiteConfig(config); err != nil {
		return fmt.Errorf("站点配置验证失败: %w", err)
	}

	// 创建Spider实例
	spider, err := sm.createSpider(config)
	if err != nil {
		return fmt.Errorf("创建Spider失败: %w", err)
	}

	sm.configs[siteID] = config
	sm.spiders[siteID] = spider

	logger.Info("添加站点",
		zap.String("site_id", siteID),
		zap.String("name", config.Name),
		zap.String("url", config.URL),
		zap.String("type", config.Type))

	return nil
}

// RemoveSite 移除站点
func (sm *SpiderManager) RemoveSite(siteID string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if _, exists := sm.configs[siteID]; !exists {
		return fmt.Errorf("站点不存在: %s", siteID)
	}

	delete(sm.configs, siteID)
	delete(sm.spiders, siteID)

	logger.Info("移除站点", zap.String("site_id", siteID))
	return nil
}

// GetSpider 获取Spider实例
func (sm *SpiderManager) GetSpider(siteID string) (Spider, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	spider, exists := sm.spiders[siteID]
	if !exists {
		return nil, fmt.Errorf("Spider不存在: %s", siteID)
	}

	return spider, nil
}

// GetConfig 获取站点配置
func (sm *SpiderManager) GetConfig(siteID string) (*SiteConfig, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	config, exists := sm.configs[siteID]
	if !exists {
		return nil, fmt.Errorf("站点配置不存在: %s", siteID)
	}

	// 返回配置的副本，防止外部修改
	configCopy := *config
	return &configCopy, nil
}

// ListSites 列出所有站点
func (sm *SpiderManager) ListSites() map[string]*SiteConfig {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	result := make(map[string]*SiteConfig)
	for siteID, config := range sm.configs {
		configCopy := *config
		result[siteID] = &configCopy
	}

	return result
}

// EnableSite 启用站点
func (sm *SpiderManager) EnableSite(siteID string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	config, exists := sm.configs[siteID]
	if !exists {
		return fmt.Errorf("站点不存在: %s", siteID)
	}

	config.Enabled = true
	logger.Info("启用站点", zap.String("site_id", siteID))
	return nil
}

// DisableSite 禁用站点
func (sm *SpiderManager) DisableSite(siteID string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	config, exists := sm.configs[siteID]
	if !exists {
		return fmt.Errorf("站点不存在: %s", siteID)
	}

	config.Enabled = false
	logger.Info("禁用站点", zap.String("site_id", siteID))
	return nil
}

// UpdateSite 更新站点配置
func (sm *SpiderManager) UpdateSite(siteID string, config *SiteConfig) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if _, exists := sm.configs[siteID]; !exists {
		return fmt.Errorf("站点不存在: %s", siteID)
	}

	// 验证新配置
	if err := sm.validateSiteConfig(config); err != nil {
		return fmt.Errorf("站点配置验证失败: %w", err)
	}

	// 创建新的Spider实例
	spider, err := sm.createSpider(config)
	if err != nil {
		return fmt.Errorf("创建Spider失败: %w", err)
	}

	sm.configs[siteID] = config
	sm.spiders[siteID] = spider

	logger.Info("更新站点配置", zap.String("site_id", siteID))
	return nil
}

// Search 在所有启用的站点上搜索
func (sm *SpiderManager) Search(ctx context.Context, keyword string, filters *SearchFilters) (map[string][]*TorrentItem, error) {
	sites := sm.ListSites()
	results := make(map[string][]*TorrentItem)
	var wg sync.WaitGroup
	var mutex sync.Mutex

	logger.Info("开始多站点搜索",
		zap.String("keyword", keyword),
		zap.Int("sites", len(sites)))

	for siteID, config := range sites {
		if !config.Enabled {
			continue
		}

		wg.Add(1)
		go func(id string, cfg *SiteConfig) {
			defer wg.Done()

			// 搜索前添加超时保护
			searchCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
			defer cancel()

			items, err := sm.searchSite(searchCtx, id, keyword, filters)
			if err != nil {
				logger.Warn("站点搜索失败",
					zap.String("site_id", id),
					zap.String("name", cfg.Name),
					zap.Error(err))
				return
			}

			mutex.Lock()
			results[id] = items
			mutex.Unlock()

			logger.Debug("站点搜索完成",
				zap.String("site_id", id),
				zap.String("name", cfg.Name),
				zap.Int("count", len(items)))
		}(siteID, config)
	}

	wg.Wait()

	totalResults := 0
	for siteID, items := range results {
		totalResults += len(items)
		logger.Debug("站点搜索结果",
			zap.String("site_id", siteID),
			zap.Int("count", len(items)))
	}

	logger.Info("多站点搜索完成",
		zap.String("keyword", keyword),
		zap.Int("total_results", totalResults),
		zap.Int("successful_sites", len(results)))

	return results, nil
}

// searchSite 在单个站点搜索
func (sm *SpiderManager) searchSite(ctx context.Context, siteID string, keyword string, filters *SearchFilters) ([]*TorrentItem, error) {
	spider, err := sm.GetSpider(siteID)
	if err != nil {
		return nil, err
	}

	config, err := sm.GetConfig(siteID)
	if err != nil {
		return nil, err
	}

	if !config.Enabled {
		return nil, fmt.Errorf("站点已禁用: %s", siteID)
	}

	// 添加站点特定的过滤器
	siteFilters := sm.applySiteFilters(filters, config)

	return spider.Search(ctx, keyword, siteFilters)
}

// applySiteFilters 应用站点特定过滤器
func (sm *SpiderManager) applySiteFilters(filters *SearchFilters, config *SiteConfig) *SearchFilters {
	if filters == nil {
		return nil
	}

	// 创建过滤器副本
	siteFilters := *filters
	siteFilters.Custom = make(map[string]string)

	// 复制原有自定义参数
	for k, v := range filters.Custom {
		siteFilters.Custom[k] = v
	}

	// 添加站点特定参数
	for k, v := range config.CustomParams {
		siteFilters.Custom[k] = v
	}

	return &siteFilters
}

// GetSupportedURLs 获取支持的URL列表
func (sm *SpiderManager) GetSupportedURLs() []string {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	var urls []string
	for _, config := range sm.configs {
		if config.Enabled {
			urls = append(urls, config.URL)
		}
	}

	return urls
}

// DetectSpider 检测URL对应的Spider
func (sm *SpiderManager) DetectSpider(url string) (string, Spider, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	for siteID, spider := range sm.spiders {
		if spider.IsSupported(url) {
			config := sm.configs[siteID]
			if config.Enabled {
				return siteID, spider, nil
			}
			return "", nil, fmt.Errorf("站点已禁用: %s", siteID)
		}
	}

	return "", nil, fmt.Errorf("未找到支持该URL的Spider: %s", url)
}

// HealthCheck 健康检查
func (sm *SpiderManager) HealthCheck(ctx context.Context) map[string]error {
	sites := sm.ListSites()
	results := make(map[string]error)

	logger.Info("开始站点健康检查", zap.Int("sites", len(sites)))

	var wg sync.WaitGroup
	var mutex sync.Mutex

	for siteID, config := range sites {
		if !config.Enabled {
			continue
		}

		wg.Add(1)
		go func(id string) {
			defer wg.Done()

			// 设置健康检查超时
			healthCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			err := sm.checkSiteHealth(healthCtx, id)
			
			mutex.Lock()
			results[id] = err
			mutex.Unlock()

			if err != nil {
				logger.Warn("站点健康检查失败",
					zap.String("site_id", id),
					zap.String("name", config.Name),
					zap.Error(err))
			} else {
				logger.Debug("站点健康检查成功",
					zap.String("site_id", id),
					zap.String("name", config.Name))
			}
		}(siteID)
	}

	wg.Wait()

	healthyCount := 0
	unhealthyCount := 0
	for siteID, err := range results {
		if err == nil {
			healthyCount++
		} else {
			unhealthyCount++
		}
		logger.Debug("健康检查结果",
			zap.String("site_id", siteID),
			zap.Bool("healthy", err == nil),
			zap.Error(err))
	}

	logger.Info("站点健康检查完成",
		zap.Int("healthy", healthyCount),
		zap.Int("unhealthy", unhealthyCount))

	return results
}

// checkSiteHealth 检查单个站点健康状态
func (sm *SpiderManager) checkSiteHealth(ctx context.Context, siteID string) error {
	spider, err := sm.GetSpider(siteID)
	if err != nil {
		return err
	}

	// 尝试获取用户信息来验证连接
	_, err = spider.GetUserInfo(ctx)
	if err != nil {
		return fmt.Errorf("连接测试失败: %w", err)
	}

	return nil
}

// GetStatistics 获取统计信息
func (sm *SpiderManager) GetStatistics() *Statistics {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	stats := &Statistics{
		TotalSites:     len(sm.configs),
		EnabledSites:   0,
		DisabledSites:  0,
		SitesByType:    make(map[string]int),
		LastUpdated:    time.Now(),
	}

	for _, config := range sm.configs {
		if config.Enabled {
			stats.EnabledSites++
		} else {
			stats.DisabledSites++
		}

		stats.SitesByType[config.Type]++
	}

	return stats
}

// Statistics 统计信息
type Statistics struct {
	TotalSites    int            `json:"total_sites"`    // 总站点数
	EnabledSites  int            `json:"enabled_sites"`  // 启用站点数
	DisabledSites int            `json:"disabled_sites"` // 禁用站点数
	SitesByType   map[string]int `json:"sites_by_type"` // 按类型分组的站点数
	LastUpdated   time.Time      `json:"last_updated"`   // 最后更新时间
}

// createSpider 创建Spider实例
func (sm *SpiderManager) createSpider(config *SiteConfig) (Spider, error) {
	switch config.Type {
	case "nexusphp":
		return NewNexusPHP(config.URL), nil
	case "gazelle":
		// TODO: 实现Gazelle Spider
		return nil, fmt.Errorf("Gazelle Spider尚未实现")
	case "discuz":
		// TODO: 实现Discuz Spider
		return nil, fmt.Errorf("Discuz Spider尚未实现")
	case "unit3d":
		// TODO: 实现Unit3D Spider
		return nil, fmt.Errorf("Unit3D Spider尚未实现")
	default:
		return nil, fmt.Errorf("不支持的Spider类型: %s", config.Type)
	}
}

// validateSiteConfig 验证站点配置
func (sm *SpiderManager) validateSiteConfig(config *SiteConfig) error {
	if config.Name == "" {
		return fmt.Errorf("站点名称不能为空")
	}

	if config.URL == "" {
		return fmt.Errorf("站点URL不能为空")
	}

	if config.Type == "" {
		return fmt.Errorf("Spider类型不能为空")
	}

	// 验证URL格式
	if !isURL(config.URL) {
		return fmt.Errorf("站点URL格式无效: %s", config.URL)
	}

	return nil
}

// isURL 检查是否为有效URL
func isURL(url string) bool {
	return len(url) > 7 && 
		   (url[:7] == "http://" || url[:8] == "https://")
}

// Close 关闭管理器
func (sm *SpiderManager) Close() error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// 清理资源
	logger.Info("Spider管理器已关闭")

	return nil
}