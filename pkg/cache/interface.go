package cache

import (
	"context"
	"time"
)

// Backend 缓存后端接口
// 原Python: CacheBackend in app/core/cache.py
type Backend interface {
	// Set 设置缓存
	// region: 缓存区域
	// key: 缓存键
	// value: 缓存值
	// ttlSeconds: 过期时间（秒），0表示永不过期
	Set(region, key string, value any, ttlSeconds int64) error

	// Get 获取缓存
	// region: 缓存区域
	// key: 缓存键
	// dest: 目标对象指针，用于接收缓存值
	// 返回值: (是否命中, 错误)
	Get(region, key string, dest any) (bool, error)

	// Exists 检查缓存是否存在
	Exists(region, key string) (bool, error)

	// Delete 删除缓存
	Delete(region, key string) error

	// Clear 清空缓存
	// region为空时清空所有区域
	Clear(region string) error

	// Items 获取指定区域的所有缓存项
	Items(region string) (map[string]any, error)

	// Keys 获取指定区域的所有缓存键
	Keys(region string) ([]string, error)

	// Values 获取指定区域的所有缓存值
	Values(region string) ([]any, error)

	// Update 更新缓存，类似dict.update()
	Update(region string, other map[string]any, ttlSeconds int64) error

	// Pop 弹出缓存项，类似dict.pop()
	// 如果key不存在且提供了default，则返回default
	// 如果key不存在且未提供default，则返回nil
	Pop(region, key string, defaultValue ...any) (any, error)

	// Popitem 弹出最后一个缓存项，类似dict.popitem()
	// 如果缓存为空，返回error
	Popitem(region string) (string, any, error)

	// Setdefault 设置默认值，类似dict.setdefault()
	// 如果key不存在，则设置并返回default
	// 如果key存在，则返回现有值
	Setdefault(region, key string, defaultValue any, ttlSeconds int64) (any, error)

	// Close 关闭缓存连接
	Close() error

	// GetItem 获取缓存项，类似dict[key]
	GetItem(region, key string, dest any) (any, error)

	// SetItem 设置缓存项，类似dict[key] = value
	SetItem(region, key string, value any) error

	// DeleteItem 删除缓存项，类似del dict[key]
	DeleteItem(region, key string) error

	// Contains 检查键是否存在，类似key in dict
	Contains(region, key string) (bool, error)

	// Iter 返回缓存的迭代器，类似iter(dict)
	Iter(region string) ([]string, error)

	// Len 返回缓存项的数量，类似len(dict)
	Len(region string) (int, error)

	// GetRegion 获取缓存区域，原Python: get_region()
	GetRegion(region string) string

	// IsRedis 判断当前缓存后端是否为Redis，原Python: is_redis()
	IsRedis() bool
}

// AsyncBackend 异步缓存后端接口
// 原Python: AsyncCacheBackend in app/core/cache.py
type AsyncBackend interface {
	// AsyncSet 异步设置缓存
	AsyncSet(region, key string, value any, ttlSeconds int64) chan error

	// AsyncGet 异步获取缓存
	// 返回值: (是否命中, 错误, 缓存值)
	AsyncGet(region, key string, dest any) chan [3]any

	// AsyncExists 异步检查缓存是否存在
	AsyncExists(region, key string) chan [2]any

	// AsyncDelete 异步删除缓存
	AsyncDelete(region, key string) chan error

	// AsyncClear 异步清空缓存
	// region为空时清空所有区域
	AsyncClear(region string) chan error

	// AsyncItems 异步获取指定区域的所有缓存项
	AsyncItems(region string) chan [2]any

	// AsyncKeys 异步获取指定区域的所有缓存键
	AsyncKeys(region string) chan [2]any

	// AsyncValues 异步获取指定区域的所有缓存值
	AsyncValues(region string) chan [2]any

	// AsyncUpdate 异步更新缓存，类似dict.update()
	AsyncUpdate(region string, other map[string]any, ttlSeconds int64) chan error

	// AsyncPop 异步弹出缓存项，类似dict.pop()
	// 返回值: (缓存值, 错误)
	AsyncPop(region, key string, defaultValue ...any) chan [2]any

	// AsyncPopitem 异步弹出最后一个缓存项，类似dict.popitem()
	// 返回值: (缓存键, 缓存值, 错误)
	AsyncPopitem(region string) chan [3]any

	// AsyncSetdefault 异步设置默认值，类似dict.setdefault()
	// 返回值: (缓存值, 错误)
	AsyncSetdefault(region, key string, defaultValue any, ttlSeconds int64) chan [2]any

	// AsyncClose 异步关闭缓存连接
	AsyncClose() chan error
}

// ServiceCache 服务层缓存接口
// 原Python: 对应服务层使用的缓存接口
type ServiceCache interface {
	// GetJSON 获取JSON格式缓存
	GetJSON(ctx context.Context, key string, dest any) error

	// SetJSON 设置JSON格式缓存
	SetJSON(ctx context.Context, key string, value any, ttl time.Duration) error

	// Get 获取字符串缓存
	Get(ctx context.Context, key string) (string, error)

	// Set 设置字符串缓存
	Set(ctx context.Context, key string, value any, ttl time.Duration) error

	// Delete 删除缓存
	Delete(ctx context.Context, key string) error

	// Clear 清空缓存
	Clear(ctx context.Context, pattern string) error
}

// Type 缓存类型
// 原Python: 对应不同的CacheBackend实现
type Type string

const (
	BackendMemory Type = "memory" // 内存缓存
	BackendRedis  Type = "redis"  // Redis缓存
	BackendFile   Type = "file"   // 文件缓存
)

// Config 缓存配置
// 原Python: 对应Cache工厂的配置参数
type Config struct {
	Type        Type          // 缓存类型
	DefaultTTL  time.Duration // 默认过期时间
	MaxSize     int           // 最大缓存数量（仅内存缓存）
	FileBaseDir string        // 文件缓存基础目录
	RedisURL    string        // Redis连接URL
}
