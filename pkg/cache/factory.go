package cache

import (
	"fmt"
)

// NewBackend 创建缓存后端
// 原Python: 对应Cache/FileCache/AsyncCache等工厂函数
type NewBackendFunc func(config Config) (Backend, error)

// NewBackend 创建缓存后端
func NewBackend(config Config) (Backend, error) {
	var backend Backend
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
