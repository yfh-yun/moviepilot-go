// Package parser 索引器Parser包
package parser

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// ParserManager 解析器管理器
type ParserManager struct {
	parsers map[string]Parser
	configs map[string]*SiteConfig
	mutex   sync.RWMutex
}

// SiteConfig 站点配置
type SiteConfig struct {
	Name      string `json:"name"`      // 站点名称
	Domain    string `json:"domain"`    // 站点域名
	Charset   string `json:"charset"`   // 字符编码
	UserAgent string `json:"user_agent"` // User-Agent
	Timeout   int    `json:"timeout"`   // 超时时间（秒）
	Enabled   bool   `json:"enabled"`   // 是否启用
}

// NewParserManager 创建解析器管理器
func NewParserManager() *ParserManager {
	return &ParserManager{
		parsers: make(map[string]Parser),
		configs: make(map[string]*SiteConfig),
	}
}

// RegisterParser 注册解析器
func (pm *ParserManager) RegisterParser(siteID string, parser Parser) error {
	if parser == nil {
		return fmt.Errorf("解析器不能为空")
	}

	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pm.parsers[siteID] = parser

	// 从解析器获取站点信息
	siteInfo := parser.GetSiteInfo()
	if siteInfo != nil {
		config := &SiteConfig{
			Name:    siteInfo.Name,
			Domain:  siteInfo.Domain,
			Charset: siteInfo.Charset,
			Enabled: true,
		}
		pm.configs[siteID] = config
	}

	return nil
}

// GetParser 获取解析器
func (pm *ParserManager) GetParser(siteID string) (Parser, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	parser, exists := pm.parsers[siteID]
	if !exists {
		return nil, fmt.Errorf("解析器不存在: %s", siteID)
	}

	return parser, nil
}

// DetectParser 检测URL对应的解析器
func (pm *ParserManager) DetectParser(url string) (string, Parser, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	for siteID, parser := range pm.parsers {
		if parser.IsSupported(url) {
			config := pm.configs[siteID]
			if config != nil && config.Enabled {
				return siteID, parser, nil
			}
			return "", nil, fmt.Errorf("解析器已禁用: %s", siteID)
		}
	}

	return "", nil, fmt.Errorf("未找到支持该URL的解析器: %s", url)
}

// ListParsers 列出所有解析器
func (pm *ParserManager) ListParsers() map[string]*SiteConfig {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	result := make(map[string]*SiteConfig)
	for siteID, config := range pm.configs {
		configCopy := *config
		result[siteID] = &configCopy
	}

	return result
}

// EnableParser 启用解析器
func (pm *ParserManager) EnableParser(siteID string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	config, exists := pm.configs[siteID]
	if !exists {
		return fmt.Errorf("解析器配置不存在: %s", siteID)
	}

	config.Enabled = true
	return nil
}

// DisableParser 禁用解析器
func (pm *ParserManager) DisableParser(siteID string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	config, exists := pm.configs[siteID]
	if !exists {
		return fmt.Errorf("解析器配置不存在: %s", siteID)
	}

	config.Enabled = false
	return nil
}

// UpdateConfig 更新站点配置
func (pm *ParserManager) UpdateConfig(siteID string, config *SiteConfig) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if config == nil {
		return fmt.Errorf("配置不能为空")
	}

	// 验证配置
	if config.Name == "" {
		return fmt.Errorf("站点名称不能为空")
	}

	if config.Domain == "" {
		return fmt.Errorf("站点域名不能为空")
	}

	// 更新配置
	pm.configs[siteID] = config

	// 如果解析器存在，更新站点信息
	if parser, exists := pm.parsers[siteID]; exists {
		siteInfo := parser.GetSiteInfo()
		if siteInfo != nil {
			siteInfo.Name = config.Name
			siteInfo.Domain = config.Domain
			siteInfo.Charset = config.Charset
		}
	}

	return nil
}

// ParseTorrentPage 解析种子列表页
func (pm *ParserManager) ParseTorrentPage(ctx context.Context, siteID, htmlContent string, page int) ([]*TorrentInfo, error) {
	parser, err := pm.GetParser(siteID)
	if err != nil {
		return nil, fmt.Errorf("获取解析器失败: %w", err)
	}

	return parser.ParseTorrentPage(ctx, htmlContent, page)
}

// ParseTorrentDetail 解析种子详情页
func (pm *ParserManager) ParseTorrentDetail(ctx context.Context, siteID, htmlContent string) (*TorrentDetail, error) {
	parser, err := pm.GetParser(siteID)
	if err != nil {
		return nil, fmt.Errorf("获取解析器失败: %w", err)
	}

	return parser.ParseTorrentDetail(ctx, htmlContent)
}

// ParseSearchResult 解析搜索结果页
func (pm *ParserManager) ParseSearchResult(ctx context.Context, siteID, htmlContent, keyword string, page int) ([]*TorrentInfo, error) {
	parser, err := pm.GetParser(siteID)
	if err != nil {
		return nil, fmt.Errorf("获取解析器失败: %w", err)
	}

	return parser.ParseSearchResult(ctx, htmlContent, keyword, page)
}

// ParseUserPage 解析用户页
func (pm *ParserManager) ParseUserPage(ctx context.Context, siteID, htmlContent, userID string) (*UserInfo, error) {
	parser, err := pm.GetParser(siteID)
	if err != nil {
		return nil, fmt.Errorf("获取解析器失败: %w", err)
	}

	return parser.ParseUserPage(ctx, htmlContent, userID)
}

// BatchParseTorrentPage 批量解析种子列表页
func (pm *ParserManager) BatchParseTorrentPage(ctx context.Context, requests map[string]ParseRequest) (map[string]*ParseResult) {
	results := make(map[string]*ParseResult)
	var wg sync.WaitGroup
	var mutex sync.Mutex

	for siteID, request := range requests {
		wg.Add(1)
		go func(id string, req ParseRequest) {
			defer wg.Done()

			result := &ParseResult{
				SiteID: id,
				Success: false,
			}

			// 解析种子列表
			torrents, err := pm.ParseTorrentPage(ctx, id, req.HTMLContent, req.Page)
			if err != nil {
				result.Error = err.Error()
			} else {
				result.Torrents = torrents
				result.Success = true
			}

			mutex.Lock()
			results[id] = result
			mutex.Unlock()
		}(siteID, request)
	}

	wg.Wait()
	return results
}

// BatchParseTorrentDetail 批量解析种子详情页
func (pm *ParserManager) BatchParseTorrentDetail(ctx context.Context, requests map[string]ParseRequest) (map[string]*ParseResult) {
	results := make(map[string]*ParseResult)
	var wg sync.WaitGroup
	var mutex sync.Mutex

	for siteID, request := range requests {
		wg.Add(1)
		go func(id string, req ParseRequest) {
			defer wg.Done()

			result := &ParseResult{
				SiteID: id,
				Success: false,
			}

			// 解析种子详情
			detail, err := pm.ParseTorrentDetail(ctx, id, req.HTMLContent)
			if err != nil {
				result.Error = err.Error()
			} else {
				result.Detail = detail
				result.Success = true
			}

			mutex.Lock()
			results[id] = result
			mutex.Unlock()
		}(siteID, request)
	}

	wg.Wait()
	return results
}

// GetCategories 获取所有站点分类
func (pm *ParserManager) GetCategories() map[string][]string {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	categories := make(map[string][]string)

	for siteID, parser := range pm.parsers {
		siteInfo := parser.GetSiteInfo()
		if siteInfo != nil {
			categories[siteID] = siteInfo.Categories
		}
	}

	return categories
}

// GetFeatures 获取所有站点特性
func (pm *ParserManager) GetFeatures() map[string][]string {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	features := make(map[string][]string)

	for siteID, parser := range pm.parsers {
		siteInfo := parser.GetSiteInfo()
		if siteInfo != nil {
			features[siteID] = siteInfo.Features
		}
	}

	return features
}

// SearchByCategory 按分类搜索解析器
func (pm *ParserManager) SearchByCategory(category string) []string {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	var result []string

	for siteID, parser := range pm.parsers {
		config := pm.configs[siteID]
		if config == nil || !config.Enabled {
			continue
		}

		siteInfo := parser.GetSiteInfo()
		if siteInfo != nil {
			for _, cat := range siteInfo.Categories {
				if strings.EqualFold(cat, category) {
					result = append(result, siteID)
					break
				}
			}
		}
	}

	return result
}

// SearchByFeature 按特性搜索解析器
func (pm *ParserManager) SearchByFeature(feature string) []string {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	var result []string

	for siteID, parser := range pm.parsers {
		config := pm.configs[siteID]
		if config == nil || !config.Enabled {
			continue
		}

		siteInfo := parser.GetSiteInfo()
		if siteInfo != nil {
			for _, feat := range siteInfo.Features {
				if strings.EqualFold(feat, feature) {
					result = append(result, siteID)
					break
				}
			}
		}
	}

	return result
}

// ValidateConfig 验证配置
func (pm *ParserManager) ValidateConfig(config *SiteConfig) error {
	if config == nil {
		return fmt.Errorf("配置不能为空")
	}

	if config.Name == "" {
		return fmt.Errorf("站点名称不能为空")
	}

	if config.Domain == "" {
		return fmt.Errorf("站点域名不能为空")
	}

	// 验证域名格式
	if !strings.Contains(config.Domain, ".") {
		return fmt.Errorf("域名格式无效: %s", config.Domain)
	}

	// 验证字符编码
	if config.Charset != "" && config.Charset != "UTF-8" && config.Charset != "GBK" && config.Charset != "GB2312" {
		return fmt.Errorf("不支持的字符编码: %s", config.Charset)
	}

	// 验证超时时间
	if config.Timeout < 1 || config.Timeout > 300 {
		return fmt.Errorf("超时时间必须在1-300秒之间")
	}

	return nil
}

// GetStatistics 获取统计信息
func (pm *ParserManager) GetStatistics() *Statistics {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	stats := &Statistics{
		TotalParsers:   len(pm.parsers),
		EnabledParsers: 0,
		DisabledParsers: 0,
		ParsersByDomain: make(map[string]int),
		Categories:     make(map[string]int),
		Features:       make(map[string]int),
	}

	for siteID, config := range pm.configs {
		if config.Enabled {
			stats.EnabledParsers++
		} else {
			stats.DisabledParsers++
		}

		stats.ParsersByDomain[config.Domain]++

		// 统计分类和特性
		if parser, exists := pm.parsers[siteID]; exists {
			siteInfo := parser.GetSiteInfo()
			if siteInfo != nil {
				for _, category := range siteInfo.Categories {
					stats.Categories[category]++
				}

				for _, feature := range siteInfo.Features {
					stats.Features[feature]++
				}
			}
		}
	}

	return stats
}

// ParseRequest 解析请求
type ParseRequest struct {
	SiteID      string `json:"site_id"`       // 站点ID
	HTMLContent  string `json:"html_content"`   // HTML内容
	Page         int    `json:"page"`          // 页码
	Keyword      string `json:"keyword"`       // 关键词
	UserID       string `json:"user_id"`       // 用户ID
}

// ParseResult 解析结果
type ParseResult struct {
	SiteID   string         `json:"site_id"`   // 站点ID
	Success   bool           `json:"success"`   // 是否成功
	Error     string         `json:"error"`     // 错误信息
	Torrents  []*TorrentInfo `json:"torrents"`  // 种子列表
	Detail    *TorrentDetail `json:"detail"`    // 种子详情
	UserInfo  *UserInfo      `json:"user_info"` // 用户信息
	ParseTime string         `json:"parse_time"` // 解析时间
}

// Statistics 统计信息
type Statistics struct {
	TotalParsers    int            `json:"total_parsers"`    // 总解析器数
	EnabledParsers int            `json:"enabled_parsers"` // 启用解析器数
	DisabledParsers int           `json:"disabled_parsers"` // 禁用解析器数
	ParsersByDomain map[string]int `json:"parsers_by_domain"` // 按域名分组
	Categories       map[string]int `json:"categories"`        // 分类统计
	Features        map[string]int `json:"features"`         // 特性统计
}

// Close 关闭管理器
func (pm *ParserManager) Close() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// 清理资源
	pm.parsers = make(map[string]Parser)
	pm.configs = make(map[string]*SiteConfig)
}