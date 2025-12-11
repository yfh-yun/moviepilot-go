package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisBackend Redis缓存后端
// 原Python: RedisBackend in app/core/cache.py
type RedisBackend struct {
	client *redis.Client // Redis客户端
	config Config        // 配置
	ctx    context.Context
}

// NewRedisBackend 创建Redis缓存后端
func NewRedisBackend(config Config) (*RedisBackend, error) {
	// 解析Redis URL
	opts, err := redis.ParseURL(config.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("解析Redis URL失败: %w", err)
	}

	// 创建Redis客户端
	client := redis.NewClient(opts)

	// 测试连接
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("连接Redis失败: %w", err)
	}

	return &RedisBackend{
		client: client,
		config: config,
		ctx:    ctx,
	}, nil
}

// buildKey 构建Redis键
// 原Python: 对应key的格式化规则
func (r *RedisBackend) buildKey(region, key string) string {
	return fmt.Sprintf("region:%s:%s", region, key)
}

// Set 设置缓存
func (r *RedisBackend) Set(region, key string, value any, ttlSeconds int64) error {
	// 序列化值
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}

	// 构建Redis键
	redisKey := r.buildKey(region, key)

	// 计算TTL：优先使用单次调用传入的 ttlSeconds，其次使用配置中的 DefaultTTL
	var ttl time.Duration
	if ttlSeconds > 0 {
		ttl = time.Duration(ttlSeconds) * time.Second
	} else {
		ttl = r.config.DefaultTTL
	}

	// 设置缓存（ttl<=0 表示永不过期）
	return r.client.Set(r.ctx, redisKey, bytes, ttl).Err()
}

// Get 获取缓存
func (r *RedisBackend) Get(region, key string, dest any) (bool, error) {
	// 构建Redis键
	redisKey := r.buildKey(region, key)

	// 获取缓存
	bytes, err := r.client.Get(r.ctx, redisKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			// 键不存在
			return false, nil
		}
		return false, err
	}

	// 反序列化值
	if err := json.Unmarshal(bytes, dest); err != nil {
		return false, err
	}

	return true, nil
}

// Exists 检查缓存是否存在
func (r *RedisBackend) Exists(region, key string) (bool, error) {
	// 构建Redis键
	redisKey := r.buildKey(region, key)

	// 检查是否存在
	result, err := r.client.Exists(r.ctx, redisKey).Result()
	if err != nil {
		return false, err
	}

	return result > 0, nil
}

// Delete 删除缓存
func (r *RedisBackend) Delete(region, key string) error {
	// 构建Redis键
	redisKey := r.buildKey(region, key)

	// 删除缓存
	return r.client.Del(r.ctx, redisKey).Err()
}

// Clear 清空缓存
func (r *RedisBackend) Clear(region string) error {
	if region == "" {
		// 清空所有缓存（使用通配符）
		keys, err := r.client.Keys(r.ctx, "region:*:*").Result()
		if err != nil {
			return err
		}

		if len(keys) > 0 {
			return r.client.Del(r.ctx, keys...).Err()
		}
		return nil
	}

	// 清空指定区域
	keys, err := r.client.Keys(r.ctx, fmt.Sprintf("region:%s:*", region)).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return r.client.Del(r.ctx, keys...).Err()
	}

	return nil
}

// Close 关闭Redis连接
func (r *RedisBackend) Close() error {
	return r.client.Close()
}
