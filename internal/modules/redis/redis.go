package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"moviepilot-go/internal/config"
	"moviepilot-go/internal/utils"
)

// RedisHelper Redis连接和操作助手类
type RedisHelper struct {
	redisURL string
	client   *redis.Client
}

// NewRedisHelper 创建Redis助手实例
func NewRedisHelper() *RedisHelper {
	return &RedisHelper{
		redisURL: config.Config.CACHE_BACKEND_URL,
	}
}

// connect 建立Redis连接
func (r *RedisHelper) connect() error {
	if r.client == nil {
		r.client = redis.NewClient(&redis.Options{
			Addr:         r.redisURL,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		})

		// 测试连接，确保Redis可用
		ctx := context.Background()
		if err := r.client.Ping(ctx).Err(); err != nil {
			utils.Log.Errorf("Failed to connect to Redis: %v", err)
			r.client = nil
			return fmt.Errorf("Redis connection failed: %v", err)
		}

		utils.Log.Infof("Successfully connected to Redis: %s", r.redisURL)
		r.setMemoryLimit()
	}

	return nil
}

// setMemoryLimit 动态设置Redis最大内存和内存淘汰策略
func (r *RedisHelper) setMemoryLimit() {
	// TODO: 实现内存限制设置逻辑
	// 这需要根据项目的具体配置来实�?}

// Test 测试Redis连接�?func (r *RedisHelper) Test() bool {
	err := r.connect()
	if err != nil {
		utils.Log.Errorf("Redis connection test failed: %v", err)
		return false
	}
	return true
}

// Close 关闭Redis客户端的连接�?func (r *RedisHelper) Close() {
	if r.client != nil {
		r.client.Close()
		r.client = nil
		utils.Log.Debug("Redis connection closed")
	}
}
