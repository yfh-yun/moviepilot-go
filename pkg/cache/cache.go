package cache

import (
	"time"
)

// Default constants
const (
	DefaultCacheRegion = "DEFAULT"
	DefaultCacheSize   = 1024
	DefaultCacheTTL    = 365 * 24 * time.Hour
)

// CacheType represents the type of cache
type CacheType string

const (
	// CacheTypeLRU represents an LRU cache
	CacheTypeLRU CacheType = "lru"
	// CacheTypeTTL represents a TTL cache
	CacheTypeTTL CacheType = "ttl"
)

// CacheBackend defines the interface for cache backends
// It provides methods to interact with cache storage
// Similar to Python's CacheBackend abstract class

type CacheBackend interface {
	// Set stores a value in the cache
	Set(key string, value interface{}, ttl time.Duration, region string, opts ...Option) error

	// Get retrieves a value from the cache
	Get(key string, region string) (interface{}, bool, error)

	// Exists checks if a key exists in the cache
	Exists(key string, region string) (bool, error)

	// Delete removes a key from the cache
	Delete(key string, region string) error

	// Clear removes all keys from the cache, or from a specific region
	Clear(region string) error

	// Items returns all key-value pairs from the cache
	Items(region string) (map[string]interface{}, error)

	// Keys returns all keys from the cache
	Keys(region string) ([]string, error)

	// Values returns all values from the cache
	Values(region string) ([]interface{}, error)

	// Update updates the cache with the given map
	Update(other map[string]interface{}, region string, ttl time.Duration, opts ...Option) error

	// Pop removes and returns the value for the given key
	Pop(key string, region string, defaultValue ...interface{}) (interface{}, error)

	// Popitem removes and returns the last key-value pair
	Popitem(region string) (string, interface{}, error)

	// Setdefault sets the default value for the given key
	Setdefault(key string, defaultValue interface{}, region string, ttl time.Duration, opts ...Option) (interface{}, error)

	// Close closes the cache connection
	Close() error

	// IsRedis checks if the backend is Redis
	IsRedis() bool

	// GetRegion returns the formatted region string
	GetRegion(region string) string
}

// AsyncCacheBackend defines the interface for asynchronous cache backends
// It provides methods to interact with cache storage asynchronously
// Similar to Python's AsyncCacheBackend abstract class

type AsyncCacheBackend interface {
	// Set stores a value in the cache
	Set(key string, value interface{}, ttl time.Duration, region string, opts ...Option) error

	// Get retrieves a value from the cache
	Get(key string, region string) (interface{}, bool, error)

	// Exists checks if a key exists in the cache
	Exists(key string, region string) (bool, error)

	// Delete removes a key from the cache
	Delete(key string, region string) error

	// Clear removes all keys from the cache, or from a specific region
	Clear(region string) error

	// Items returns all key-value pairs from the cache
	Items(region string) (map[string]interface{}, error)

	// Keys returns all keys from the cache
	Keys(region string) ([]string, error)

	// Values returns all values from the cache
	Values(region string) ([]interface{}, error)

	// Update updates the cache with the given map
	Update(other map[string]interface{}, region string, ttl time.Duration, opts ...Option) error

	// Pop removes and returns the value for the given key
	Pop(key string, region string, defaultValue ...interface{}) (interface{}, error)

	// Popitem removes and returns the last key-value pair
	Popitem(region string) (string, interface{}, error)

	// Setdefault sets the default value for the given key
	Setdefault(key string, defaultValue interface{}, region string, ttl time.Duration, opts ...Option) (interface{}, error)

	// Close closes the cache connection
	Close() error

	// IsRedis checks if the backend is Redis
	IsRedis() bool

	// GetRegion returns the formatted region string
	GetRegion(region string) string
}

// Option represents an optional parameter for cache operations
// Similar to Python's **kwargs in cache methods
type Option func(*options)

// options holds optional parameters for cache operations
type options struct {
	MaxSize int
}

// WithMaxSize sets the maximum size for the cache
func WithMaxSize(maxSize int) Option {
	return func(o *options) {
		o.MaxSize = maxSize
	}
}

// getOptions processes the provided options and returns the options struct
func getOptions(opts ...Option) *options {
	options := &options{
		MaxSize: DefaultCacheSize,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// BaseCacheBackend 基础缓存后端实现，提供通用方法
// 类似Python中的CacheBackend基类

type BaseCacheBackend struct{}

// IsRedis 默认返回false，由具体实现覆盖
func (b *BaseCacheBackend) IsRedis() bool {
	return false
}

// GetRegion 返回格式化的区域字符串
func (b *BaseCacheBackend) GetRegion(region string) string {
	if region == "" {
		return "region:default"
	}
	return "region:" + region
}
