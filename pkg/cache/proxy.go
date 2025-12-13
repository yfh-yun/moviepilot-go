package cache

import (
	"time"
)

// CacheProxy 缓存代理类
// 原Python: CacheProxy in app/core/cache.py
type CacheProxy struct {
	backend CacheBackend // 底层缓存后端
	region  string       // 缓存区域
}

// NewCacheProxy 创建缓存代理实例
func NewCacheProxy(backend CacheBackend, region string) *CacheProxy {
	return &CacheProxy{
		backend: backend,
		region:  region,
	}
}

// IsRedis 检查当前缓存后端是否为Redis
func (p *CacheProxy) IsRedis() bool {
	return p.backend.IsRedis()
}

// Get 获取缓存值
func (p *CacheProxy) Get(key string) (interface{}, bool, error) {
	return p.backend.Get(key, p.region)
}

// Set 设置缓存值
func (p *CacheProxy) Set(key string, value interface{}, ttl time.Duration, opts ...Option) error {
	return p.backend.Set(key, value, ttl, p.region, opts...)
}

// Delete 删除缓存值
func (p *CacheProxy) Delete(key string) error {
	return p.backend.Delete(key, p.region)
}

// Exists 检查缓存键是否存在
func (p *CacheProxy) Exists(key string) (bool, error) {
	return p.backend.Exists(key, p.region)
}

// Clear 清除缓存
func (p *CacheProxy) Clear() error {
	return p.backend.Clear(p.region)
}

// Items 获取所有缓存项
func (p *CacheProxy) Items() (map[string]interface{}, error) {
	return p.backend.Items(p.region)
}

// Keys 获取所有缓存键
func (p *CacheProxy) Keys() ([]string, error) {
	return p.backend.Keys(p.region)
}

// Values 获取所有缓存值
func (p *CacheProxy) Values() ([]interface{}, error) {
	return p.backend.Values(p.region)
}

// Update 更新缓存
func (p *CacheProxy) Update(other map[string]interface{}, ttl time.Duration, opts ...Option) error {
	return p.backend.Update(other, p.region, ttl, opts...)
}

// Pop 弹出缓存项
func (p *CacheProxy) Pop(key string, defaultValue ...interface{}) (interface{}, error) {
	return p.backend.Pop(key, p.region, defaultValue...)
}

// Popitem 弹出最后一个缓存项
func (p *CacheProxy) Popitem() (string, interface{}, error) {
	return p.backend.Popitem(p.region)
}

// Setdefault 设置默认值
func (p *CacheProxy) Setdefault(key string, defaultValue interface{}, ttl time.Duration, opts ...Option) (interface{}, error) {
	return p.backend.Setdefault(key, defaultValue, p.region, ttl, opts...)
}

// Close 关闭缓存连接
func (p *CacheProxy) Close() error {
	return p.backend.Close()
}

// TTLCache 基于TTL的缓存类
// 原Python: TTLCache in app/core/cache.py
type TTLCache struct {
	*CacheProxy
}

// NewTTLCache 创建TTL缓存实例
// maxsize: 缓存的最大条目数
// ttl: 缓存的存活时间
// region: 缓存的区，为None时使用默认区
func NewTTLCache(region string, maxsize int, ttl int64) *TTLCache {
	// 创建缓存配置
	config := Config{
		Type:        BackendMemory,
		DefaultTTL:  time.Duration(ttl),
		MaxSize:     maxsize,
		FileBaseDir: "",
		RedisURL:    "",
	}

	// 创建缓存后端
	backend := NewMemoryBackend(config)

	// 创建并返回TTLCache实例
	return &TTLCache{
		CacheProxy: NewCacheProxy(backend, region),
	}
}

// LRUCache 基于LRU的缓存类
// 原Python: LRUCache in app/core/cache.py
type LRUCache struct {
	*CacheProxy
}

// NewLRUCache 创建LRU缓存实例
// region: 缓存的区，为None时使用默认区
// maxsize: 缓存的最大条目数
func NewLRUCache(region string, maxsize int) *LRUCache {
	// 创建缓存配置
	config := Config{
		Type:        BackendMemory,
		DefaultTTL:  0, // LRU缓存无TTL
		MaxSize:     maxsize,
		FileBaseDir: "",
		RedisURL:    "",
	}

	// 创建缓存后端
	backend := NewMemoryBackend(config)

	// 创建并返回LRUCache实例
	return &LRUCache{
		CacheProxy: NewCacheProxy(backend, region),
	}
}

