package tvdb

import (
	"sync"
	"time"
)

// Cache TVDB数据缓存
type Cache struct {
	seriesCache   map[int]*Series
	episodesCache map[int][]Episode
	searchCache   map[string][]Series
	mu            sync.RWMutex
	defaultTTL    time.Duration
	maxEntries    int
}

// cacheEntry 缓存条目
type cacheEntry struct {
	data      interface{}
	expiresAt time.Time
}

// NewCache 创建缓存实例
func NewCache(defaultTTL time.Duration, maxEntries int) *Cache {
	return &Cache{
		seriesCache:   make(map[int]*Series),
		episodesCache: make(map[int][]Episode),
		searchCache:   make(map[string][]Series),
		defaultTTL:    defaultTTL,
		maxEntries:    maxEntries,
	}
}

// GetSeries 获取剧集缓存
func (c *Cache) GetSeries(seriesID int) (*Series, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if entry, exists := c.seriesCache[seriesID]; exists {
		// 检查是否过期
		if time.Now().Before(entry.expiresAt) {
			return entry.data.(*Series), true
		}
		// 过期则删除
		delete(c.seriesCache, seriesID)
	}

	return nil, false
}

// SetSeries 设置剧集缓存
func (c *Cache) SetSeries(seriesID int, series *Series) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查缓存大小，如果超过限制则清理过期条目
	if len(c.seriesCache) >= c.maxEntries {
		c.cleanupExpired()
	}

	c.seriesCache[seriesID] = &cacheEntry{
		data:      series,
		expiresAt: time.Now().Add(c.defaultTTL),
	}
}

// GetEpisodes 获取剧集列表缓存
func (c *Cache) GetEpisodes(seriesID int) ([]Episode, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if entry, exists := c.episodesCache[seriesID]; exists {
		// 检查是否过期
		if time.Now().Before(entry.expiresAt) {
			return entry.data.([]Episode), true
		}
		// 过期则删除
		delete(c.episodesCache, seriesID)
	}

	return nil, false
}

// SetEpisodes 设置剧集列表缓存
func (c *Cache) SetEpisodes(seriesID int, episodes []Episode) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查缓存大小，如果超过限制则清理过期条目
	if len(c.episodesCache) >= c.maxEntries {
		c.cleanupExpired()
	}

	c.episodesCache[seriesID] = &cacheEntry{
		data:      episodes,
		expiresAt: time.Now().Add(c.defaultTTL),
	}
}

// GetSearch 获取搜索缓存
func (c *Cache) GetSearch(query string) ([]Series, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if entry, exists := c.searchCache[query]; exists {
		// 检查是否过期
		if time.Now().Before(entry.expiresAt) {
			return entry.data.([]Series), true
		}
		// 过期则删除
		delete(c.searchCache, query)
	}

	return nil, false
}

// SetSearch 设置搜索缓存
func (c *Cache) SetSearch(query string, series []Series) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查缓存大小，如果超过限制则清理过期条目
	if len(c.searchCache) >= c.maxEntries {
		c.cleanupExpired()
	}

	c.searchCache[query] = &cacheEntry{
		data:      series,
		expiresAt: time.Now().Add(c.defaultTTL),
	}
}

// cleanupExpired 清理过期缓存
func (c *Cache) cleanupExpired() {
	now := time.Now()

	// 清理剧集缓存
	for key, entry := range c.seriesCache {
		if now.After(entry.expiresAt) {
			delete(c.seriesCache, key)
		}
	}

	// 清理剧集列表缓存
	for key, entry := range c.episodesCache {
		if now.After(entry.expiresAt) {
			delete(c.episodesCache, key)
		}
	}

	// 清理搜索缓存
	for key, entry := range c.searchCache {
		if now.After(entry.expiresAt) {
			delete(c.searchCache, key)
		}
	}
}

// Clear 清空缓存
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.seriesCache = make(map[int]*Series)
	c.episodesCache = make(map[int][]Episode)
	c.searchCache = make(map[string][]Series)
}

// Size 获取缓存大小
func (c *Cache) Size() (int, int, int) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.seriesCache), len(c.episodesCache), len(c.searchCache)
}

// RemoveSeries 移除特定剧集缓存
func (c *Cache) RemoveSeries(seriesID int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.seriesCache, seriesID)
	delete(c.episodesCache, seriesID)
}

// RemoveSearch 移除特定搜索缓存
func (c *Cache) RemoveSearch(query string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.searchCache, query)
}

// GetExpiredCount 获取过期条目数量
func (c *Cache) GetExpiredCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	now := time.Now()
	count := 0

	for _, entry := range c.seriesCache {
		if now.After(entry.expiresAt) {
			count++
		}
	}

	for _, entry := range c.episodesCache {
		if now.After(entry.expiresAt) {
			count++
		}
	}

	for _, entry := range c.searchCache {
		if now.After(entry.expiresAt) {
			count++
		}
	}

	return count
}
