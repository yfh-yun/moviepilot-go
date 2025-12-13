package cache

import (
	"fmt"
	"os"
	"time"
)

// NewBackend 创建缓存后端
// 原Python: 对应Cache/FileCache/AsyncCache等工厂函数
func NewBackend(config Config) (CacheBackend, error) {
	var backend CacheBackend
	var err error

	// 根据配置创建不同类型的缓存后端
	switch config.Type {
	case BackendMemory:
		// 创建内存缓存
		backend = NewMemoryBackend(config)

	case BackendRedis:
		// 创建Redis缓存
		backend, err = NewRedisBackend(config)
		if err != nil {
			return nil, fmt.Errorf("创建Redis缓存失败: %w", err)
		}

	case BackendFile:
		// 创建文件缓存
		backend, err = NewFileBackend(config)
		if err != nil {
			return nil, fmt.Errorf("创建文件缓存失败: %w", err)
		}

	default:
		// 默认使用内存缓存
		backend = NewMemoryBackend(config)
	}

	return backend, nil
}

// FileCache 获取文件缓存后端实例（Redis或文件系统）
// 原Python: FileCache in app/core/cache.py
// base: 文件缓存基础目录
// ttl: 过期时间（秒），仅在Redis环境中有效
func FileCache(base string, ttl int64) CacheBackend {
	// 从环境变量获取配置，实际项目中应替换为从全局配置获取
	// TODO: 替换为实际的全局配置模块
	cacheBackendType := os.Getenv("CACHE_BACKEND_TYPE")

	// 根据配置选择缓存类型
	cacheType := BackendFile
	if cacheBackendType == "redis" {
		cacheType = BackendRedis
	}

	// 创建缓存配置
	config := Config{
		Type:        cacheType,
		DefaultTTL:  time.Duration(ttl) * time.Second,
		MaxSize:     DefaultCacheSize,
		FileBaseDir: base,
		RedisURL:    os.Getenv("REDIS_URL"),
	}

	// 根据配置创建缓存后端
	backend, err := NewBackend(config)
	if err != nil {
		// 创建失败，使用内存缓存作为降级方案
		backend = NewMemoryBackend(config)
	}

	return backend
}

// AsyncFileCache 获取异步文件缓存后端实例（Redis或文件系统）
// 原Python: AsyncFileCache in app/core/cache.py
// base: 文件缓存基础目录
// ttl: 过期时间（秒），仅在Redis环境中有效
func AsyncFileCache(base string, ttl int64) AsyncCacheBackend {
	// 从环境变量获取配置，实际项目中应替换为从全局配置获取
	// TODO: 替换为实际的全局配置模块
	cacheBackendType := os.Getenv("CACHE_BACKEND_TYPE")

	// 根据配置选择缓存类型
	cacheType := BackendFile
	if cacheBackendType == "redis" {
		cacheType = BackendRedis
	}

	// 创建缓存配置
	config := Config{
		Type:        cacheType,
		DefaultTTL:  time.Duration(ttl) * time.Second,
		MaxSize:     DefaultCacheSize,
		FileBaseDir: base,
		RedisURL:    os.Getenv("REDIS_URL"),
	}

	// 根据配置创建异步缓存后端
	if cacheType == BackendRedis {
		// Redis缓存同时支持同步和异步
		backend, err := NewRedisBackend(config)
		if err != nil {
			// 创建失败，使用异步内存缓存作为降级方案
			return NewAsyncMemoryBackend(config)
		}
		return backend
	}

	// 文件缓存或内存缓存
	return NewAsyncMemoryBackend(config)
}

// Cache 获取缓存后端实例（内存或Redis）
// 原Python: Cache in app/core/cache.py
// cacheType: 缓存类型，ttl为0时使用LRU缓存
// maxsize: 缓存最大大小
// ttl: 过期时间（秒）
func Cache(cacheType string, maxsize int, ttl int64) CacheBackend {
	// 从环境变量获取配置，实际项目中应替换为从全局配置获取
	// TODO: 替换为实际的全局配置模块
	cacheBackendType := os.Getenv("CACHE_BACKEND_TYPE")

	// 根据配置选择缓存类型
	backendType := BackendMemory
	if cacheBackendType == "redis" {
		backendType = BackendRedis
	}

	// 创建缓存配置
	config := Config{
		Type:        backendType,
		DefaultTTL:  time.Duration(ttl) * time.Second,
		MaxSize:     maxsize,
		FileBaseDir: "",
		RedisURL:    os.Getenv("REDIS_URL"),
	}

	// 根据配置创建缓存后端
	backend, err := NewBackend(config)
	if err != nil {
		// 创建失败，使用内存缓存作为降级方案
		backend = NewMemoryBackend(config)
	}

	return backend
}

// AsyncCache 获取异步缓存后端实例（内存或Redis）
// 原Python: AsyncCache in app/core/cache.py
// cacheType: 缓存类型，ttl为0时使用LRU缓存
// maxsize: 缓存最大大小
// ttl: 过期时间（秒）
func AsyncCache(cacheType string, maxsize int, ttl int64) AsyncCacheBackend {
	// 从环境变量获取配置，实际项目中应替换为从全局配置获取
	// TODO: 替换为实际的全局配置模块
	cacheBackendType := os.Getenv("CACHE_BACKEND_TYPE")

	// 根据配置选择缓存类型
	backendType := BackendMemory
	if cacheBackendType == "redis" {
		backendType = BackendRedis
	}

	// 创建缓存配置
	config := Config{
		Type:        backendType,
		DefaultTTL:  time.Duration(ttl) * time.Second,
		MaxSize:     maxsize,
		FileBaseDir: "",
		RedisURL:    os.Getenv("REDIS_URL"),
	}

	// 根据配置创建异步缓存后端
	if backendType == BackendRedis {
		// Redis缓存同时支持同步和异步
		backend, err := NewRedisBackend(config)
		if err != nil {
			// 创建失败，使用异步内存缓存作为降级方案
			return NewAsyncMemoryBackend(config)
		}
		return backend
	}

	// 内存缓存
	return NewAsyncMemoryBackend(config)
}

