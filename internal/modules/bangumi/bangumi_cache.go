package bangumi

import (
	"sync"
	"time"
	
	"moviepilot-go/internal/core/config"
	"moviepilot-go/internal/utils/cache"
)

// BangumiCache Bangumi缓存数据
type BangumiCache struct {
	cache    *cache.TTLCache
	maxsize  int
	ttl      int
	region   string
	mutex    sync.RWMutex
}

// NewBangumiCache 创建BangumiCache实例
func NewBangumiCache() *BangumiCache {
	bc := &BangumiCache{
		maxsize: config.Config.CONF.Bangumi,
		ttl:     config.Config.CONF.Meta,
		region:  "__bangumi_cache__",
	}
	
	// 初始化缓�?	bc.cache = cache.NewTTLCache(bc.region, bc.maxsize, bc.ttl)
	
	return bc
}

// Get 从缓存中获取数据
func (bc *BangumiCache) Get(key string) (interface{}, bool) {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()
	
	return bc.cache.Get(key)
}

// Set 将数据存入缓�?func (bc *BangumiCache) Set(key string, value interface{}) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()
	
	bc.cache.Set(key, value)
}

// Clear 清空缓存
func (bc *BangumiCache) Clear() {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()
	
	bc.cache.Clear()
}
