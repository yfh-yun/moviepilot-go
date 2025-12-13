package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisBackend Redis缓存后端
type RedisBackend struct {
	BaseCacheBackend
	client *redis.Client
	ttl    time.Duration
}

// NewRedisBackend 创建Redis缓存后端
func NewRedisBackend(config Config) (CacheBackend, error) {
	if config.RedisURL == "" {
		config.RedisURL = "redis://localhost:6379/0"
	}

	opt, err := redis.ParseURL(config.RedisURL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)

	// 测试连接
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisBackend{
		client: client,
		ttl:    config.DefaultTTL,
	}, nil
}

// getFormattedKey 获取格式化后的Redis键名
func (r *RedisBackend) getFormattedKey(region, key string) string {
	return r.GetRegion(region) + ":" + key
}

// Set 设置缓存
func (r *RedisBackend) Set(key string, value interface{}, ttl time.Duration, region string, opts ...Option) error {
	ctx := context.Background()
	formattedKey := r.getFormattedKey(region, key)

	// 序列化值
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	// 计算过期时间
	var expiration time.Duration
	if ttl > 0 {
		expiration = ttl
	} else {
		expiration = r.ttl
	}

	if expiration > 0 {
		return r.client.Set(ctx, formattedKey, data, expiration).Err()
	}

	// 永不过期
	return r.client.Set(ctx, formattedKey, data, 0).Err()
}

// Get 获取缓存
func (r *RedisBackend) Get(key string, region string) (interface{}, bool, error) {
	ctx := context.Background()
	formattedKey := r.getFormattedKey(region, key)

	// 从Redis获取值
	data, err := r.client.Get(ctx, formattedKey).Result()
	if err != nil {
		if err == redis.Nil {
			// 键不存在
			return nil, false, nil
		}
		return nil, false, err
	}

	// 反序列化值
	var value interface{}
	if err := json.Unmarshal([]byte(data), &value); err != nil {
		return nil, false, err
	}

	return value, true, nil
}

// Exists 检查缓存是否存在
func (r *RedisBackend) Exists(key string, region string) (bool, error) {
	ctx := context.Background()
	formattedKey := r.getFormattedKey(region, key)

	count, err := r.client.Exists(ctx, formattedKey).Result()
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// Delete 删除缓存
func (r *RedisBackend) Delete(key string, region string) error {
	ctx := context.Background()
	formattedKey := r.getFormattedKey(region, key)

	return r.client.Del(ctx, formattedKey).Err()
}

// Clear 清空缓存
func (r *RedisBackend) Clear(region string) error {
	ctx := context.Background()
	var pattern string

	if region == "" {
		// 清空所有区域
		pattern = "region:*"
	} else {
		// 清空指定区域
		pattern = r.GetRegion(region) + ":*"
	}

	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return r.client.Del(ctx, keys...).Err()
	}

	return nil
}

// Items 获取指定区域的所有缓存项
func (r *RedisBackend) Items(region string) (map[string]interface{}, error) {
	ctx := context.Background()
	pattern := r.getFormattedKey(region, "*")

	// 获取所有匹配的键
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	regionPrefix := r.GetRegion(region) + ":"

	for _, fullKey := range keys {
		// 提取原始键名
		originalKey := fullKey[len(regionPrefix):]

		// 获取值
		data, err := r.client.Get(ctx, fullKey).Result()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			return nil, err
		}

		// 反序列化值
		var value interface{}
		if err := json.Unmarshal([]byte(data), &value); err != nil {
			return nil, err
		}

		result[originalKey] = value
	}

	return result, nil
}

// Keys 返回所有缓存键
func (r *RedisBackend) Keys(region string) ([]string, error) {
	ctx := context.Background()
	pattern := r.getFormattedKey(region, "*")

	// 获取所有匹配的键
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	result := make([]string, 0, len(keys))
	regionPrefix := r.GetRegion(region) + ":"

	for _, fullKey := range keys {
		// 提取原始键名
		originalKey := fullKey[len(regionPrefix):]
		result = append(result, originalKey)
	}

	return result, nil
}

// Values 返回所有缓存值
func (r *RedisBackend) Values(region string) ([]interface{}, error) {
	ctx := context.Background()
	pattern := r.getFormattedKey(region, "*")

	// 获取所有匹配的键
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	result := make([]interface{}, 0, len(keys))

	for _, fullKey := range keys {
		// 获取值
		data, err := r.client.Get(ctx, fullKey).Result()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			return nil, err
		}

		// 反序列化值
		var value interface{}
		if err := json.Unmarshal([]byte(data), &value); err != nil {
			return nil, err
		}

		result = append(result, value)
	}

	return result, nil
}

// Update 更新缓存
func (r *RedisBackend) Update(other map[string]interface{}, region string, ttl time.Duration, opts ...Option) error {
	for key, value := range other {
		if err := r.Set(key, value, ttl, region, opts...); err != nil {
			return err
		}
	}

	return nil
}

// Pop 弹出缓存项
func (r *RedisBackend) Pop(key string, region string, defaultValue ...interface{}) (interface{}, error) {
	ctx := context.Background()
	formattedKey := r.getFormattedKey(region, key)

	// 获取并删除值
	data, err := r.client.GetDel(ctx, formattedKey).Result()
	if err != nil {
		if err == redis.Nil {
			// 键不存在，返回默认值
			if len(defaultValue) > 0 {
				return defaultValue[0], nil
			}
			return nil, nil
		}
		return nil, err
	}

	// 反序列化值
	var value interface{}
	if err := json.Unmarshal([]byte(data), &value); err != nil {
		return nil, err
	}

	return value, nil
}

// Popitem 弹出最后一个缓存项
func (r *RedisBackend) Popitem(region string) (string, interface{}, error) {
	// 获取所有键
	keys, err := r.Keys(region)
	if err != nil {
		return "", nil, err
	}

	if len(keys) == 0 {
		return "", nil, nil
	}

	// 获取最后一个键
	key := keys[len(keys)-1]
	value, err := r.Pop(key, region)
	if err != nil {
		return "", nil, err
	}

	return key, value, nil
}

// Setdefault 设置默认值
func (r *RedisBackend) Setdefault(key string, defaultValue interface{}, region string, ttl time.Duration, opts ...Option) (interface{}, error) {
	value, hit, err := r.Get(key, region)
	if err != nil {
		return nil, err
	}

	if hit {
		return value, nil
	}

	// 设置默认值
	if err := r.Set(key, defaultValue, ttl, region, opts...); err != nil {
		return nil, err
	}

	return defaultValue, nil
}

// Close 关闭缓存连接
func (r *RedisBackend) Close() error {
	return r.client.Close()
}

// IsRedis 判断当前缓存后端是否为Redis
func (r *RedisBackend) IsRedis() bool {
	return true
}

// AsyncRedisBackend 异步Redis缓存后端
type AsyncRedisBackend struct {
	*RedisBackend
}

// NewAsyncRedisBackend 创建异步Redis缓存后端
func NewAsyncRedisBackend(config Config) (*AsyncRedisBackend, error) {
	backend, err := NewRedisBackend(config)
	if err != nil {
		return nil, err
	}

	return &AsyncRedisBackend{
		RedisBackend: backend.(*RedisBackend),
	}, nil
}

