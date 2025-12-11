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

	// Close 关闭缓存连接
	Close() error
}

// Cache 缓存接口（用于服务层）
// 原Python: 对应服务层使用的缓存接口
type Cache interface {
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
