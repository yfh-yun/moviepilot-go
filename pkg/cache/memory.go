package cache

import (
	"sync"
	"time"
)

// cacheItem 内存缓存项
type cacheItem struct {
	value      interface{}
	expiration time.Time
	addedAt    time.Time
}

// MemoryBackend 内存缓存后端
// 原Python: MemoryBackend in app/core/cache.py
type MemoryBackend struct {
	BaseCacheBackend
	cacheType  CacheType
	maxsize    int
	defaultTTL time.Duration
	cache      map[string]map[string]*cacheItem
	mutex      sync.RWMutex
	stopChan   chan struct{}
}

// NewMemoryBackend 创建内存缓存后端
func NewMemoryBackend(config Config) *MemoryBackend {
	cacheType := CacheTypeLRU
	if config.DefaultTTL > 0 {
		cacheType = CacheTypeTTL
	}

	mb := &MemoryBackend{
		cacheType:  cacheType,
		maxsize:    config.MaxSize,
		defaultTTL: config.DefaultTTL,
		cache:      make(map[string]map[string]*cacheItem),
		stopChan:   make(chan struct{}),
	}

	// 启动过期清理协程（仅TTL缓存需要）
	if cacheType == CacheTypeTTL {
		go mb.cleanupExpired()
	}

	return mb
}

// cleanupExpired 清理过期缓存项
func (m *MemoryBackend) cleanupExpired() {
	ticker := time.NewTicker(5 * time.Minute) // 每5分钟清理一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.mutex.Lock()
			now := time.Now()

			// 遍历所有区域和缓存项
			for region, items := range m.cache {
				for key, item := range items {
					if !item.expiration.IsZero() && now.After(item.expiration) {
						// 过期，删除
						delete(items, key)
					}
				}

				// 如果区域为空，删除该区域
				if len(items) == 0 {
					delete(m.cache, region)
				}
			}
			m.mutex.Unlock()

		case <-m.stopChan:
			return
		}
	}
}

// Set 设置缓存
func (m *MemoryBackend) Set(key string, value interface{}, ttl time.Duration, region string, opts ...Option) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 处理选项
	optsObj := getOptions(opts...)
	maxsize := optsObj.MaxSize
	if maxsize == 0 {
		maxsize = m.maxsize
	}

	// 确保区域存在
	if _, ok := m.cache[region]; !ok {
		m.cache[region] = make(map[string]*cacheItem)
	}

	// 计算过期时间
	var expiration time.Time
	if ttl > 0 {
		expiration = time.Now().Add(ttl)
	} else if m.defaultTTL > 0 {
		expiration = time.Now().Add(m.defaultTTL)
	}

	// 检查LRU缓存大小限制
	if m.cacheType == CacheTypeLRU && maxsize > 0 {
		items := m.cache[region]
		if len(items) >= maxsize {
			// 找到最旧的项并删除
			var oldestKey string
			var oldestTime time.Time
			for k, item := range items {
				if oldestTime.IsZero() || item.addedAt.Before(oldestTime) {
					oldestKey = k
					oldestTime = item.addedAt
				}
			}
			delete(items, oldestKey)
		}
	}

	// 设置缓存项
	m.cache[region][key] = &cacheItem{
		value:      value,
		expiration: expiration,
		addedAt:    time.Now(),
	}

	return nil
}

// Get 获取缓存
func (m *MemoryBackend) Get(key string, region string) (interface{}, bool, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// 检查区域是否存在
	regionItems, ok := m.cache[region]
	if !ok {
		return nil, false, nil
	}

	// 检查缓存项是否存在
	item, ok := regionItems[key]
	if !ok {
		return nil, false, nil
	}

	// 检查是否过期
	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		return nil, false, nil
	}

	return item.value, true, nil
}

// Exists 检查缓存是否存在
func (m *MemoryBackend) Exists(key string, region string) (bool, error) {
	_, hit, err := m.Get(key, region)
	return hit, err
}

// Delete 删除缓存
func (m *MemoryBackend) Delete(key string, region string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 检查区域是否存在
	regionItems, ok := m.cache[region]
	if !ok {
		return nil
	}

	// 删除缓存项
	delete(regionItems, key)

	// 如果区域为空，删除该区域
	if len(regionItems) == 0 {
		delete(m.cache, region)
	}

	return nil
}

// Clear 清空缓存
func (m *MemoryBackend) Clear(region string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if region == "" {
		// 清空所有区域
		m.cache = make(map[string]map[string]*cacheItem)
	} else {
		// 清空指定区域
		delete(m.cache, region)
	}

	return nil
}

// Items 获取指定区域的所有缓存项
func (m *MemoryBackend) Items(region string) (map[string]interface{}, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	result := make(map[string]interface{})
	now := time.Now()

	// 检查区域是否存在
	regionItems, ok := m.cache[region]
	if !ok {
		return result, nil
	}

	// 遍历缓存项
	for key, item := range regionItems {
		// 跳过过期项
		if !item.expiration.IsZero() && now.After(item.expiration) {
			continue
		}
		result[key] = item.value
	}

	return result, nil
}

// Update 更新缓存
func (m *MemoryBackend) Update(other map[string]interface{}, region string, ttl time.Duration, opts ...Option) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 确保区域存在
	if _, ok := m.cache[region]; !ok {
		m.cache[region] = make(map[string]*cacheItem)
	}

	// 计算过期时间
	var expiration time.Time
	if ttl > 0 {
		expiration = time.Now().Add(ttl)
	} else if m.defaultTTL > 0 {
		expiration = time.Now().Add(m.defaultTTL)
	}

	// 批量设置缓存项
	for key, value := range other {
		m.cache[region][key] = &cacheItem{
			value:      value,
			expiration: expiration,
			addedAt:    time.Now(),
		}
	}

	return nil
}

// Pop 弹出缓存项
func (m *MemoryBackend) Pop(key string, region string, defaultValue ...interface{}) (interface{}, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 检查区域是否存在
	regionItems, ok := m.cache[region]
	if !ok {
		if len(defaultValue) > 0 {
			return defaultValue[0], nil
		}
		return nil, nil
	}

	// 检查缓存项是否存在
	item, ok := regionItems[key]
	if !ok {
		if len(defaultValue) > 0 {
			return defaultValue[0], nil
		}
		return nil, nil
	}

	// 删除缓存项
	delete(regionItems, key)

	// 如果区域为空，删除该区域
	if len(regionItems) == 0 {
		delete(m.cache, region)
	}

	// 检查是否过期
	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		if len(defaultValue) > 0 {
			return defaultValue[0], nil
		}
		return nil, nil
	}

	return item.value, nil
}

// Popitem 弹出最后一个缓存项
func (m *MemoryBackend) Popitem(region string) (string, interface{}, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 检查区域是否存在
	regionItems, ok := m.cache[region]
	if !ok || len(regionItems) == 0 {
		return "", nil, nil
	}

	// 遍历找到最后一个键
	var lastKey string
	var lastItem *cacheItem
	for key, item := range regionItems {
		lastKey = key
		lastItem = item
	}

	// 删除缓存项
	delete(regionItems, lastKey)

	// 如果区域为空，删除该区域
	if len(regionItems) == 0 {
		delete(m.cache, region)
	}

	// 检查是否过期
	if !lastItem.expiration.IsZero() && time.Now().After(lastItem.expiration) {
		return "", nil, nil
	}

	return lastKey, lastItem.value, nil
}

// Setdefault 设置默认值
func (m *MemoryBackend) Setdefault(key string, defaultValue interface{}, region string, ttl time.Duration, opts ...Option) (interface{}, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 确保区域存在
	if _, ok := m.cache[region]; !ok {
		m.cache[region] = make(map[string]*cacheItem)
	}

	// 检查缓存项是否存在
	item, ok := m.cache[region][key]
	now := time.Now()
	if ok && (item.expiration.IsZero() || now.Before(item.expiration)) {
		// 缓存项存在且未过期
		return item.value, nil
	}

	// 计算过期时间
	var expiration time.Time
	if ttl > 0 {
		expiration = now.Add(ttl)
	} else if m.defaultTTL > 0 {
		expiration = now.Add(m.defaultTTL)
	}

	// 设置默认值
	newItem := &cacheItem{
		value:      defaultValue,
		expiration: expiration,
		addedAt:    time.Now(),
	}
	m.cache[region][key] = newItem

	return defaultValue, nil
}

// Close 关闭缓存连接
func (m *MemoryBackend) Close() error {
	// 发送停止信号给清理协程
	close(m.stopChan)
	return nil
}

// Keys 返回所有缓存键
func (m *MemoryBackend) Keys(region string) ([]string, error) {
	items, err := m.Items(region)
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(items))
	for k := range items {
		keys = append(keys, k)
	}
	return keys, nil
}

// Values 返回所有缓存值
func (m *MemoryBackend) Values(region string) ([]interface{}, error) {
	items, err := m.Items(region)
	if err != nil {
		return nil, err
	}
	values := make([]interface{}, 0, len(items))
	for _, v := range items {
		values = append(values, v)
	}
	return values, nil
}

// AsyncMemoryBackend 异步内存缓存后端
// 原Python: AsyncMemoryBackend in app/core/cache.py
type AsyncMemoryBackend struct {
	*MemoryBackend
}

// NewAsyncMemoryBackend 创建异步内存缓存后端
func NewAsyncMemoryBackend(config Config) *AsyncMemoryBackend {
	return &AsyncMemoryBackend{
		MemoryBackend: NewMemoryBackend(config),
	}
}

// Keys 返回所有缓存键（异步实现）
func (m *AsyncMemoryBackend) Keys(region string) ([]string, error) {
	return m.MemoryBackend.Keys(region)
}

// Values 返回所有缓存值（异步实现）
func (m *AsyncMemoryBackend) Values(region string) ([]interface{}, error) {
	return m.MemoryBackend.Values(region)
}

// Update 更新缓存（异步实现）
func (m *AsyncMemoryBackend) Update(other map[string]interface{}, region string, ttl time.Duration, opts ...Option) error {
	return m.MemoryBackend.Update(other, region, ttl, opts...)
}
