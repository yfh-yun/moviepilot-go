package cache

import (
	"context"
	"encoding/json"
	"time"
)

// SimpleCache 简单缓存实现
// 原Python: 对应服务层使用的缓存实现
type SimpleCache struct {
	backend Backend
	region  string
}

// NewSimpleCache 创建简单缓存
func NewSimpleCache(backend Backend, region string) *SimpleCache {
	return &SimpleCache{
		backend: backend,
		region:  region,
	}
}

// GetJSON 获取JSON格式缓存
func (c *SimpleCache) GetJSON(ctx context.Context, key string, dest any) error {
	var data []byte
	hit, err := c.backend.Get(c.region, key, &data)
	if err != nil {
		return err
	}
	if !hit {
		return nil // 未命中，不返回错误
	}
	return json.Unmarshal(data, dest)
}

// SetJSON 设置JSON格式缓存
func (c *SimpleCache) SetJSON(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	ttlSeconds := int64(ttl.Seconds())
	return c.backend.Set(c.region, key, data, ttlSeconds)
}

// Get 获取字符串缓存
func (c *SimpleCache) Get(ctx context.Context, key string) (string, error) {
	var data string
	hit, err := c.backend.Get(c.region, key, &data)
	if err != nil {
		return "", err
	}
	if !hit {
		return "", nil // 未命中，返回空字符串
	}
	return data, nil
}

// Set 设置字符串缓存
func (c *SimpleCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	ttlSeconds := int64(ttl.Seconds())
	return c.backend.Set(c.region, key, value, ttlSeconds)
}

// Delete 删除缓存
func (c *SimpleCache) Delete(ctx context.Context, key string) error {
	return c.backend.Delete(c.region, key)
}

// Clear 清空缓存
func (c *SimpleCache) Clear(ctx context.Context, pattern string) error {
	// 简单实现，清空整个区域
	return c.backend.Clear(c.region)
}
