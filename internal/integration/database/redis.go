package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"

	"github.com/redis/go-redis/v9"
)

// RedisClient Redis客户端
type RedisClient struct {
	client *redis.Client
	config *RedisConfig
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host         string        `json:"host"`
	Port         int           `json:"port"`
	Password     string        `json:"password"`
	Database     int           `json:"database"`
	PoolSize     int           `json:"pool_size"`
	MinIdleConns int           `json:"min_idle_conns"`
	MaxRetries   int           `json:"max_retries"`
	DialTimeout  time.Duration `json:"dial_timeout"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	PoolTimeout  time.Duration `json:"pool_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
}

// NewRedisClient 创建Redis客户端
func NewRedisClient(config *RedisConfig) (*RedisClient, error) {
	if config == nil {
		return nil, fmt.Errorf("Redis config is required")
	}

	// 设置默认值
	if config.Port == 0 {
		config.Port = 6379
	}
	if config.PoolSize == 0 {
		config.PoolSize = 20
	}
	if config.MinIdleConns == 0 {
		config.MinIdleConns = 5
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.DialTimeout == 0 {
		config.DialTimeout = time.Second * 5
	}
	if config.ReadTimeout == 0 {
		config.ReadTimeout = time.Second * 3
	}
	if config.WriteTimeout == 0 {
		config.WriteTimeout = time.Second * 3
	}
	if config.PoolTimeout == 0 {
		config.PoolTimeout = time.Second * 4
	}
	if config.IdleTimeout == 0 {
		config.IdleTimeout = time.Minute * 5
	}

	// 创建Redis客户端
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password:     config.Password,
		DB:           config.Database,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		MaxRetries:   config.MaxRetries,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		PoolTimeout:  config.PoolTimeout,
		IdleTimeout:  config.IdleTimeout,
	})

	redisClient := &RedisClient{
		client: client,
		config: config,
	}

	// 测试连接
	if err := redisClient.Test(context.Background()); err != nil {
		client.Close()
		return nil, fmt.Errorf("Redis connection test failed: %w", err)
	}

	logger.Info("Redis client connected successfully to %s:%d/%d", config.Host, config.Port, config.Database)
	return redisClient, nil
}

// Test 测试连接
func (c *RedisClient) Test(ctx context.Context) error {
	if c.client == nil {
		return fmt.Errorf("Redis client is not initialized")
	}

	// 执行ping测试
	result, err := c.client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("Redis connection test failed: %w", err)
	}

	if result != "PONG" {
		return fmt.Errorf("Redis ping returned unexpected result: %s", result)
	}

	logger.Info("Redis connection test successful")
	return nil
}

// Close 关闭连接
func (c *RedisClient) Close() error {
	if c.client != nil {
		err := c.client.Close()
		if err != nil {
			return fmt.Errorf("failed to close Redis client: %w", err)
		}
		logger.Info("Redis client closed")
	}
	return nil
}

// GetClient 获取Redis客户端
func (c *RedisClient) GetClient() *redis.Client {
	return c.client
}

// GetStats 获取连接统计信息
func (c *RedisClient) GetStats() *redis.PoolStats {
	if c.client == nil {
		return nil
	}
	return c.client.PoolStats()
}

// Set 设置键值对
func (c *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if c.client == nil {
		return fmt.Errorf("Redis client is not initialized")
	}

	err := c.client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}

	return nil
}

// Get 获取值
func (c *RedisClient) Get(ctx context.Context, key string) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("Redis client is not initialized")
	}

	result, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("key %s does not exist", key)
		}
		return "", fmt.Errorf("failed to get key %s: %w", key, err)
	}

	return result, nil
}

// GetJSON 获取JSON值
func (c *RedisClient) GetJSON(ctx context.Context, key string, dest interface{}) error {
	if c.client == nil {
		return fmt.Errorf("Redis client is not initialized")
	}

	result, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key %s does not exist", key)
		}
		return fmt.Errorf("failed to get key %s: %w", key, err)
	}

	err = json.Unmarshal([]byte(result), dest)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON for key %s: %w", key, err)
	}

	return nil
}

// SetJSON 设置JSON值
func (c *RedisClient) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if c.client == nil {
		return fmt.Errorf("Redis client is not initialized")
	}

	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON for key %s: %w", key, err)
	}

	err = c.client.Set(ctx, key, jsonData, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set JSON key %s: %w", key, err)
	}

	return nil
}

// Delete 删除键
func (c *RedisClient) Delete(ctx context.Context, keys ...string) (int64, error) {
	if c.client == nil {
		return 0, fmt.Errorf("Redis client is not initialized")
	}

	if len(keys) == 0 {
		return 0, nil
	}

	result, err := c.client.Del(ctx, keys...).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to delete keys %v: %w", keys, err)
	}

	return result, nil
}

// Exists 检查键是否存在
func (c *RedisClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	if c.client == nil {
		return 0, fmt.Errorf("Redis client is not initialized")
	}

	if len(keys) == 0 {
		return 0, nil
	}

	result, err := c.client.Exists(ctx, keys...).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to check existence of keys %v: %w", keys, err)
	}

	return result, nil
}

// Expire 设置过期时间
func (c *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	if c.client == nil {
		return false, fmt.Errorf("Redis client is not initialized")
	}

	result, err := c.client.Expire(ctx, key, expiration).Result()
	if err != nil {
		return false, fmt.Errorf("failed to set expiration for key %s: %w", key, err)
	}

	return result, nil
}

// TTL 获取剩余过期时间
func (c *RedisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	if c.client == nil {
		return 0, fmt.Errorf("Redis client is not initialized")
	}

	result, err := c.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL for key %s: %w", key, err)
	}

	return result, nil
}

// Incr 递增
func (c *RedisClient) Incr(ctx context.Context, key string) (int64, error) {
	if c.client == nil {
		return 0, fmt.Errorf("Redis client is not initialized")
	}

	result, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment key %s: %w", key, err)
	}

	return result, nil
}

// Decr 递减
func (c *RedisClient) Decr(ctx context.Context, key string) (int64, error) {
	if c.client == nil {
		return 0, fmt.Errorf("Redis client is not initialized")
	}

	result, err := c.client.Decr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to decrement key %s: %w", key, err)
	}

	return result, nil
}

// HSet 设置哈希字段
func (c *RedisClient) HSet(ctx context.Context, key string, values ...interface{}) (int64, error) {
	if c.client == nil {
		return 0, fmt.Errorf("Redis client is not initialized")
	}

	result, err := c.client.HSet(ctx, key, values...).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to set hash fields for key %s: %w", key, err)
	}

	return result, nil
}

// HGet 获取哈希字段
func (c *RedisClient) HGet(ctx context.Context, key, field string) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("Redis client is not initialized")
	}

	result, err := c.client.HGet(ctx, key, field).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("field %s does not exist in hash %s", field, key)
		}
		return "", fmt.Errorf("failed to get hash field %s from key %s: %w", field, key, err)
	}

	return result, nil
}

// HGetAll 获取所有哈希字段
func (c *RedisClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	if c.client == nil {
		return nil, fmt.Errorf("Redis client is not initialized")
	}

	result, err := c.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get all hash fields from key %s: %w", key, err)
	}

	return result, nil
}

// HDel 删除哈希字段
func (c *RedisClient) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	if c.client == nil {
		return 0, fmt.Errorf("Redis client is not initialized")
	}

	result, err := c.client.HDel(ctx, key, fields...).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to delete hash fields %v from key %s: %w", fields, key, err)
	}

	return result, nil
}

// LPush 列表左推
func (c *RedisClient) LPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	if c.client == nil {
		return 0, fmt.Errorf("Redis client is not initialized")
	}

	result, err := c.client.LPush(ctx, key, values...).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to lpush to list %s: %w", key, err)
	}

	return result, nil
}

// RPush 列表右推
func (c *RedisClient) RPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	if c.client == nil {
		return 0, fmt.Errorf("Redis client is not initialized")
	}

	result, err := c.client.RPush(ctx, key, values...).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to rpush to list %s: %w", key, err)
	}

	return result, nil
}

// LPop 列表左弹
func (c *RedisClient) LPop(ctx context.Context, key string) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("Redis client is not initialized")
	}

	result, err := c.client.LPop(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("list %s is empty", key)
		}
		return "", fmt.Errorf("failed to lpop from list %s: %w", key, err)
	}

	return result, nil
}

// RPop 列表右弹
func (c *RedisClient) RPop(ctx context.Context, key string) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("Redis client is not initialized")
	}

	result, err := c.client.RPop(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("list %s is empty", key)
		}
		return "", fmt.Errorf("failed to rpop from list %s: %w", key, err)
	}

	return result, nil
}

// LRange 获取列表范围
func (c *RedisClient) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	if c.client == nil {
		return nil, fmt.Errorf("Redis client is not initialized")
	}

	result, err := c.client.LRange(ctx, key, start, stop).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get range from list %s: %w", key, err)
	}

	return result, nil
}

// LLen 获取列表长度
func (c *RedisClient) LLen(ctx context.Context, key string) (int64, error) {
	if c.client == nil {
		return 0, fmt.Errorf("Redis client is not initialized")
	}

	result, err := c.client.LLen(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get length of list %s: %w", key, err)
	}

	return result, nil
}

// SAdd 集合添加
func (c *RedisClient) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	if c.client == nil {
		return 0, fmt.Errorf("Redis client is not initialized")
	}

	result, err := c.client.SAdd(ctx, key, members...).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to add members to set %s: %w", key, err)
	}

	return result, nil
}

// SMembers 获取集合成员
func (c *RedisClient) SMembers(ctx context.Context, key string) ([]string, error) {
	if c.client == nil {
		return nil, fmt.Errorf("Redis client is not initialized")
	}

	result, err := c.client.SMembers(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get members from set %s: %w", key, err)
	}

	return result, nil
}

// SRem 集合删除
func (c *RedisClient) SRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	if c.client == nil {
		return 0, fmt.Errorf("Redis client is not initialized")
	}

	result, err := c.client.SRem(ctx, key, members...).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to remove members from set %s: %w", key, err)
	}

	return result, nil
}

// SCard 获取集合大小
func (c *RedisClient) SCard(ctx context.Context, key string) (int64, error) {
	if c.client == nil {
		return 0, fmt.Errorf("Redis client is not initialized")
	}

	result, err := c.client.SCard(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get size of set %s: %w", key, err)
	}

	return result, nil
}

// Keys 获取匹配的键
func (c *RedisClient) Keys(ctx context.Context, pattern string) ([]string, error) {
	if c.client == nil {
		return nil, fmt.Errorf("Redis client is not initialized")
	}

	result, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get keys matching pattern %s: %w", pattern, err)
	}

	return result, nil
}

// FlushDB 清空当前数据库
func (c *RedisClient) FlushDB(ctx context.Context) error {
	if c.client == nil {
		return fmt.Errorf("Redis client is not initialized")
	}

	err := c.client.FlushDB(ctx).Err()
	if err != nil {
		return fmt.Errorf("failed to flush database: %w", err)
	}

	logger.Info("Redis database flushed successfully")
	return nil
}

// Info 获取Redis信息
func (c *RedisClient) Info(ctx context.Context, section ...string) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("Redis client is not initialized")
	}

	result, err := c.client.Info(ctx, section...).Result()
	if err != nil {
		return "", fmt.Errorf("failed to get Redis info: %w", err)
	}

	return result, nil
}
