package fanart

import (
	"sync"
	"time"
)

// Cache Fanart图片缓存
type Cache struct {
	movieImages map[int]*MovieImages
	tvImages    map[int]*TVImages
	mu          sync.RWMutex
	defaultTTL  time.Duration
	maxEntries  int
}

// cacheEntry 缓存条目
type cacheEntry struct {
	data      interface{}
	expiresAt time.Time
}

// NewCache 创建缓存实例
func NewCache(defaultTTL time.Duration, maxEntries int) *Cache {
	return &Cache{
		movieImages: make(map[int]*MovieImages),
		tvImages:    make(map[int]*TVImages),
		defaultTTL:  defaultTTL,
		maxEntries:  maxEntries,
	}
}

// GetMovieImages 获取电影图片缓存
func (c *Cache) GetMovieImages(tmdbID int) (*MovieImages, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if images, exists := c.movieImages[tmdbID]; exists {
		// 检查是否过期
		if time.Now().Before(images.expiresAt) {
			return images.data.(*MovieImages), true
		}
		// 过期则删除
		delete(c.movieImages, tmdbID)
	}

	return nil, false
}

// SetMovieImages 设置电影图片缓存
func (c *Cache) SetMovieImages(tmdbID int, images *MovieImages) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查缓存大小，如果超过限制则清理过期条目
	if len(c.movieImages) >= c.maxEntries {
		c.cleanupExpired()
	}

	c.movieImages[tmdbID] = &cacheEntry{
		data:      images,
		expiresAt: time.Now().Add(c.defaultTTL),
	}
}

// GetTVImages 获取电视剧图片缓存
func (c *Cache) GetTVImages(tvdbID int) (*TVImages, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if images, exists := c.tvImages[tvdbID]; exists {
		// 检查是否过期
		if time.Now().Before(images.expiresAt) {
			return images.data.(*TVImages), true
		}
		// 过期则删除
		delete(c.tvImages, tvdbID)
	}

	return nil, false
}

// SetTVImages 设置电视剧图片缓存
func (c *Cache) SetTVImages(tvdbID int, images *TVImages) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查缓存大小，如果超过限制则清理过期条目
	if len(c.tvImages) >= c.maxEntries {
		c.cleanupExpired()
	}

	c.tvImages[tvdbID] = &cacheEntry{
		data:      images,
		expiresAt: time.Now().Add(c.defaultTTL),
	}
}

// cleanupExpired 清理过期缓存
func (c *Cache) cleanupExpired() {
	now := time.Now()

	// 清理电影图片缓存
	for key, entry := range c.movieImages {
		if now.After(entry.expiresAt) {
			delete(c.movieImages, key)
		}
	}

	// 清理电视剧图片缓存
	for key, entry := range c.tvImages {
		if now.After(entry.expiresAt) {
			delete(c.tvImages, key)
		}
	}
}

// Clear 清空缓存
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.movieImages = make(map[int]*MovieImages)
	c.tvImages = make(map[int]*TVImages)
}

// Size 获取缓存大小
func (c *Cache) Size() (int, int) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.movieImages), len(c.tvImages)
}
