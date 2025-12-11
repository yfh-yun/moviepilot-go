package cache

import (
	"encoding/json"
	"sync"
	"time"
)

// entry 内存缓存条目
// 原Python: 对应内存缓存中的单个缓存项
type entry struct {
	Value    []byte    `json:"value"`     // 缓存值（序列化后）
	ExpireAt time.Time `json:"expire_at"` // 过期时间
}

// MemoryBackend 内存缓存后端
// 原Python: MemoryBackend in app/core/cache.py
type MemoryBackend struct {
	cache  map[string]map[string]*entry // region -> key -> entry
	mutex  sync.RWMutex                 // 读写锁
	config Config                       // 配置
}

// NewMemoryBackend 创建内存缓存后端
func NewMemoryBackend(config Config) *MemoryBackend {
	return &MemoryBackend{
		cache:  make(map[string]map[string]*entry),
		config: config,
	}
}

// Set 设置缓存
func (m *MemoryBackend) Set(region, key string, value any, ttlSeconds int64) error {
	// 序列化值
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}

	// 计算过期时间
	expireAt := time.Time{}
	if ttlSeconds > 0 {
		expireAt = time.Now().Add(time.Duration(ttlSeconds) * time.Second)
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 确保区域存在
	if _, exists := m.cache[region]; !exists {
		m.cache[region] = make(map[string]*entry)
	}

	// 设置缓存
	m.cache[region][key] = &entry{
		Value:    bytes,
		ExpireAt: expireAt,
	}

	return nil
}

// Get 获取缓存
func (m *MemoryBackend) Get(region, key string, dest any) (bool, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// 检查区域是否存在
	regionMap, exists := m.cache[region]
	if !exists {
		return false, nil
	}

	// 检查键是否存在
	e, exists := regionMap[key]
	if !exists {
		return false, nil
	}

	// 检查是否过期
	if !e.ExpireAt.IsZero() && time.Now().After(e.ExpireAt) {
		// 过期，异步删除
		go func() {
			_ = m.Delete(region, key)
		}()
		return false, nil
	}

	// 反序列化值
	if err := json.Unmarshal(e.Value, dest); err != nil {
		return false, err
	}

	return true, nil
}

// Exists 检查缓存是否存在
func (m *MemoryBackend) Exists(region, key string) (bool, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// 检查区域是否存在
	regionMap, exists := m.cache[region]
	if !exists {
		return false, nil
	}

	// 检查键是否存在
	e, exists := regionMap[key]
	if !exists {
		return false, nil
	}

	// 检查是否过期
	if !e.ExpireAt.IsZero() && time.Now().After(e.ExpireAt) {
		// 过期，异步删除
		go func() {
			_ = m.Delete(region, key)
		}()
		return false, nil
	}

	return true, nil
}

// Delete 删除缓存
func (m *MemoryBackend) Delete(region, key string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 检查区域是否存在
	regionMap, exists := m.cache[region]
	if !exists {
		return nil
	}

	// 删除键
	delete(regionMap, key)

	// 如果区域为空，删除区域
	if len(regionMap) == 0 {
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
		m.cache = make(map[string]map[string]*entry)
	} else {
		// 清空指定区域
		delete(m.cache, region)
	}

	return nil
}

// Close 关闭缓存连接（内存缓存无需关闭）
func (m *MemoryBackend) Close() error {
	return nil
}
