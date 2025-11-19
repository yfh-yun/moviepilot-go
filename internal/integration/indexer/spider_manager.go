// Package indexer 站点爬虫管理器
package indexer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/yfh-yun/moviepilot-go/internal/integration/indexer/spiders"
	"github.com/yfh-yun/moviepilot-go/pkg/logger"

	"go.uber.org/zap"
)

// SpiderManager 站点爬虫管理器
type SpiderManager struct {
	spiders map[string]SiteSpider
	mutex   sync.RWMutex
	logger  *zap.Logger
}

// NewSpiderManager 创建爬虫管理器
func NewSpiderManager() *SpiderManager {
	return &SpiderManager{
		spiders: make(map[string]SiteSpider),
		logger:  logger.Logger,
	}
}

// RegisterSpider 注册爬虫
func (sm *SpiderManager) RegisterSpider(spider SiteSpider) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.spiders[spider.GetDomain()] = spider
	sm.logger.Info("Spider registered",
		zap.String("domain", spider.GetDomain()),
		zap.String("name", spider.GetName()))
}

// GetSpider 获取爬虫
func (sm *SpiderManager) GetSpider(domain string) SiteSpider {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	return sm.spiders[domain]
}

// GetAllSpiders 获取所有爬虫
func (sm *SpiderManager) GetAllSpiders() map[string]SiteSpider {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	result := make(map[string]SiteSpider)
	for domain, spider := range sm.spiders {
		result[domain] = spider
	}
	return result
}

// RemoveSpider 移除爬虫
func (sm *SpiderManager) RemoveSpider(domain string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	delete(sm.spiders, domain)
	sm.logger.Info("Spider removed", zap.String("domain", domain))
}

// SearchAllSites 在所有站点搜索
func (sm *SpiderManager) SearchAllSites(ctx context.Context, keyword, mediaType string) ([]*TorrentInfo, error) {
	var allTorrents []*TorrentInfo
	var wg sync.WaitGroup
	resultsChan := make(chan []*TorrentInfo, len(sm.spiders))
	errorsChan := make(chan error, len(sm.spiders))

	sm.mutex.RLock()
	spiderCount := len(sm.spiders)
	spiderMap := make(map[string]SiteSpider)
	for domain, spider := range sm.spiders {
		spiderMap[domain] = spider
	}
	sm.mutex.RUnlock()

	// 在所有站点并行搜索
	for domain, spider := range spiderMap {
		wg.Add(1)
		go func(d string, s SiteSpider) {
			defer wg.Done()

			sm.logger.Debug("Searching in site",
				zap.String("domain", d),
				zap.String("keyword", keyword),
				zap.String("media_type", mediaType))

			// 搜索超时控制
			searchCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			torrents, err := s.Search(searchCtx, keyword, mediaType)
			if err != nil {
				sm.logger.Warn("Search failed in site",
					zap.String("domain", d),
					zap.Error(err))
				errorsChan <- fmt.Errorf("site %s: %w", d, err)
				return
			}

			sm.logger.Info("Search completed in site",
				zap.String("domain", d),
				zap.Int("torrent_count", len(torrents)))

			resultsChan <- torrents
		}(domain, spider)
	}

	// 等待所有搜索完成
	go func() {
		wg.Wait()
		close(resultsChan)
		close(errorsChan)
	}()

	// 收集结果
	var errors []error
	for i := 0; i < spiderCount; i++ {
		select {
		case torrents := <-resultsChan:
			allTorrents = append(allTorrents, torrents...)
		case err := <-errorsChan:
			if err != nil {
				errors = append(errors, err)
			}
		}
	}

	// 如果所有站点都失败，返回错误
	if len(errors) == spiderCount && spiderCount > 0 {
		return nil, fmt.Errorf("all sites failed to search, errors: %v", errors)
	}

	return allTorrents, nil
}

// SearchInSpecificSites 在指定站点搜索
func (sm *SpiderManager) SearchInSpecificSites(ctx context.Context, domains []string, keyword, mediaType string) ([]*TorrentInfo, error) {
	var allTorrents []*TorrentInfo
	var wg sync.WaitGroup
	resultsChan := make(chan []*TorrentInfo, len(domains))
	errorsChan := make(chan error, len(domains)

	for _, domain := range domains {
		spider := sm.GetSpider(domain)
		if spider == nil {
			sm.logger.Warn("Spider not found for domain", zap.String("domain", domain))
			errorsChan <- fmt.Errorf("spider not found for domain %s", domain)
			continue
		}

		wg.Add(1)
		go func(d string, s SiteSpider) {
			defer wg.Done()

			searchCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			torrents, err := s.Search(searchCtx, keyword, mediaType)
			if err != nil {
				sm.logger.Warn("Search failed in site",
					zap.String("domain", d),
					zap.Error(err))
				errorsChan <- fmt.Errorf("site %s: %w", d, err)
				return
			}

			resultsChan <- torrents
		}(domain, spider)
	}

	// 等待所有搜索完成
	go func() {
		wg.Wait()
		close(resultsChan)
		close(errorsChan)
	}()

	// 收集结果
	for i := 0; i < len(domains); i++ {
		select {
		case torrents := <-resultsChan:
			allTorrents = append(allTorrents, torrents...)
		case err := <-errorsChan:
			sm.logger.Debug("Search error in site", zap.Error(err))
		}
	}

	return allTorrents, nil
}

// GetTorrentDetailsFromAllSites 从所有站点获取种子详情
func (sm *SpiderManager) GetTorrentDetailsFromAllSites(ctx context.Context, id string) ([]*TorrentDetail, error) {
	var details []*TorrentDetail
	var wg sync.WaitGroup
	resultsChan := make(chan *TorrentDetail, len(sm.spiders))
	errorsChan := make(chan error, len(sm.spiders))

	sm.mutex.RLock()
	spiderCount := len(sm.spiders)
	spiderMap := make(map[string]SiteSpider)
	for domain, spider := range sm.spiders {
		spiderMap[domain] = spider
	}
	sm.mutex.RUnlock()

	for domain, spider := range spiderMap {
		wg.Add(1)
		go func(d string, s SiteSpider) {
			defer wg.Done()

			detailCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
			defer cancel()

			detail, err := s.GetTorrentDetail(detailCtx, id)
			if err != nil {
				sm.logger.Debug("Get torrent detail failed in site",
					zap.String("domain", d),
					zap.Error(err))
				errorsChan <- fmt.Errorf("site %s: %w", d, err)
				return
			}

			resultsChan <- detail
		}(domain, spider)
	}

	// 等待所有请求完成
	go func() {
		wg.Wait()
		close(resultsChan)
		close(errorsChan)
	}()

	// 收集结果
	for i := 0; i < spiderCount; i++ {
		select {
		case detail := <-resultsChan:
			if detail != nil {
				details = append(details, detail)
			}
		case <-errorsChan:
			// 忽略错误，继续处理其他站点
		}
	}

	return details, nil
}

// GetAllUserInfos 获取所有站点的用户信息
func (sm *SpiderManager) GetAllUserInfos(ctx context.Context) map[string]*UserInfo {
	var userInfos = make(map[string]*UserInfo)
	var wg sync.WaitGroup
	resultsChan := make(chan *UserInfoResult, len(sm.spiders))

	type UserInfoResult struct {
		Domain   string
		UserInfo *UserInfo
		Error    error
	}

	sm.mutex.RLock()
	spiderMap := make(map[string]SiteSpider)
	for domain, spider := range sm.spiders {
		spiderMap[domain] = spider
	}
	sm.mutex.RUnlock()

	for domain, spider := range spiderMap {
		wg.Add(1)
		go func(d string, s SiteSpider) {
			defer wg.Done()

			infoCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
			defer cancel()

			userInfo, err := s.GetUserInfo(infoCtx)
			resultsChan <- &UserInfoResult{
				Domain:   d,
				UserInfo: userInfo,
				Error:    err,
			}
		}(domain, spider)
	}

	// 等待所有请求完成
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// 收集结果
	for i := 0; i < len(spiderMap); i++ {
		result := <-resultsChan
		if result.Error != nil {
			sm.logger.Debug("Get user info failed in site",
				zap.String("domain", result.Domain),
				zap.Error(result.Error))
		} else if result.UserInfo != nil {
			userInfos[result.Domain] = result.UserInfo
		}
	}

	return userInfos
}

// CreateSpider 创建爬虫实例
func (sm *SpiderManager) CreateSpider(siteType, domain string) (SiteSpider, error) {
	switch siteType {
	case "nexusphp":
		return spiders.NewNexusPHPSpider(siteType, domain), nil
	case "gazelle":
		return spiders.NewGazelleSpider(siteType, domain), nil
	default:
		// 尝试使用通用基础爬虫
		return spiders.NewBaseSpider(siteType, domain), nil
	}
}

// InitializeDefaultSpiders 初始化默认爬虫
func (sm *SpiderManager) InitializeDefaultSpiders() {
	// 初始化一些常见站点的爬虫
	commonSites := map[string]string{
		"hdchina":       "nexusphp",
		"hddolby":       "nexusphp",
		"pterclub":       "nexusphp",
		"ourbits":       "nexusphp",
		"hdsky":         "nexusphp",
		"red":           "gazelle",
		"ops":           "gazelle",
		"anidex":        "base",
		"ygg":           "base",
	}

	for domain, siteType := range commonSites {
		spider, err := sm.CreateSpider(siteType, domain)
		if err != nil {
			sm.logger.Error("Failed to create spider",
				zap.String("domain", domain),
				zap.String("type", siteType),
				zap.Error(err))
			continue
		}
		sm.RegisterSpider(spider)
	}

	sm.logger.Info("Default spiders initialized",
		zap.Int("count", len(commonSites)))
}

// HealthCheck 健康检查
func (sm *SpiderManager) HealthCheck(ctx context.Context) map[string]bool {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	results := make(map[string]bool)
	var wg sync.WaitGroup

	for domain, spider := range sm.spiders {
		wg.Add(1)
		go func(d string, s SiteSpider) {
			defer wg.Done()

			healthCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			// 通过检查用户信息来判断站点是否可访问
			_, err := s.GetUserInfo(healthCtx)
			results[d] = err == nil

			if err != nil {
				sm.logger.Debug("Site health check failed",
					zap.String("domain", d),
					zap.Error(err))
			}
		}(domain, spider)
	}

	wg.Wait()
	return results
}

// GetStatistics 获取统计信息
func (sm *SpiderManager) GetStatistics() map[string]interface{} {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	stats := map[string]interface{}{
		"total_spiders":  len(sm.spiders),
		"spider_types":   make(map[string]int),
		"spider_domains": make([]string, 0, len(sm.spiders)),
	}

	typeStats := make(map[string]int)
	domains := make([]string, 0, len(sm.spiders))

	for domain, spider := range sm.spiders {
		domains = append(domains, domain)
		
		// 通过反射或其他方式确定爬虫类型
		switch spider.(type) {
		case *spiders.NexusPHPSpider:
			typeStats["nexusphp"]++
		case *spiders.GazelleSpider:
			typeStats["gazelle"]++
		default:
			typeStats["base"]++
		}
	}

	stats["spider_types"] = typeStats
	stats["spider_domains"] = domains

	return stats
}

// UpdateSpiderSettings 更新爬虫设置
func (sm *SpiderManager) UpdateSpiderSettings(domain string, settings map[string]string) error {
	sm.mutex.RLock()
	spider := sm.spiders[domain]
	sm.mutex.RUnlock()

	if spider == nil {
		return fmt.Errorf("spider not found for domain: %s", domain)
	}

	// 这里可以根据需要实现设置更新逻辑
	// 例如更新请求头、代理设置等
	sm.logger.Info("Spider settings updated",
		zap.String("domain", domain),
		zap.Any("settings", settings))

	return nil
}

// ClearAllSpiders 清除所有爬虫
func (sm *SpiderManager) ClearAllSpiders() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.spiders = make(map[string]SiteSpider)
	sm.logger.Info("All spiders cleared")
}

// ReloadSpiders 重新加载爬虫
func (sm *SpiderManager) ReloadSpiders() {
	sm.ClearAllSpiders()
	sm.InitializeDefaultSpiders()
	sm.logger.Info("Spiders reloaded")
}