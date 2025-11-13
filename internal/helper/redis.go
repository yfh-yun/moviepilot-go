package helper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	// 类型缓存集合，针对非容器简单类�?	complexSerializableTypes = make(map[string]bool)
	simpleSerializableTypes = make(map[string]bool)
	typesMutex              sync.RWMutex

	// 默认连接参数
	socketTimeout       = 30 * time.Second
	socketConnectTimeout = 5 * time.Second
	healthCheckInterval = 60 * time.Second

	// RedisHelper单例
	redisHelperInstance *RedisHelper
	redisHelperOnce     sync.Once

	// AsyncRedisHelper单例
	asyncRedisHelperInstance *AsyncRedisHelper
	asyncRedisHelperOnce     sync.Once
)

// serialize 将值序列化为二进制数据，根据序列化方式标识格式
func serialize(value interface{}) ([]byte, error) {
	vt := fmt.Sprintf("%T", value)

	// 针对非容器类型使用缓存策�?	if !isContainerType(value) {
		typesMutex.RLock()
		// 如果已知需要复杂序列化
		if complexSerializableTypes[vt] {
			typesMutex.RUnlock()
			data, err := json.Marshal(value)
			if err != nil {
				return nil, err
			}
			return append([]byte("JSON\x00"), data...), nil
		}
		// 如果已知可以简单序列化
		if simpleSerializableTypes[vt] {
			typesMutex.RUnlock()
			data, err := json.Marshal(value)
			if err != nil {
				return nil, err
			}
			return append([]byte("JSON\x00"), data...), nil
		}
		typesMutex.RUnlock()

		// 对于未知的非容器类型，尝试简单序列化，如抛出异常，再使用复杂序列�?		data, err := json.Marshal(value)
		if err != nil {
			typesMutex.Lock()
			complexSerializableTypes[vt] = true
			typesMutex.Unlock()
			// 使用JSON作为默认序列化方式，Go中json.Marshal比pickle更通用
			return nil, err
		} else {
			typesMutex.Lock()
			simpleSerializableTypes[vt] = true
			typesMutex.Unlock()
		}
		return append([]byte("JSON\x00"), data...), nil
	} else {
		// 针对容器类型，每次尝试简单序列化，不使用缓存
		data, err := json.Marshal(value)
		if err != nil {
			// 使用JSON作为默认序列化方�?			return nil, err
		}
		return append([]byte("JSON\x00"), data...), nil
	}
}

// isContainerType 判断是否为容器类�?func isContainerType(value interface{}) bool {
	switch value.(type) {
	case []interface{}, map[string]interface{}, []string, map[string]string:
		return true
	default:
		return false
	}
}

// deserialize 将二进制数据反序列化为原始值，根据格式标识区分序列化方�?func deserialize(data []byte) (interface{}, error) {
	parts := bytes.SplitN(data, []byte("\x00"), 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid serialized data format")
	}

	formatMarker := parts[0]
	valueData := parts[1]

	if bytes.Equal(formatMarker, []byte("JSON")) {
		var result interface{}
		err := json.Unmarshal(valueData, &result)
		if err != nil {
			return nil, err
		}
		return result, nil
	} else {
		return nil, fmt.Errorf("unknown serialization format: %s", formatMarker)
	}
}

// RedisHelper Redis连接和操作助手类，单例模�?type RedisHelper struct {
	redisURL string
	client   *redis.Client
	mutex    sync.Mutex
}

// GetRedisHelper 获取RedisHelper单例实例
func GetRedisHelper() *RedisHelper {
	redisHelperOnce.Do(func() {
		redisHelperInstance = &RedisHelper{
			redisURL: getCacheBackendURL(),
		}
	})
	return redisHelperInstance
}

// _connect 建立Redis连接
func (r *RedisHelper) _connect() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.client != nil {
		return nil
	}

	opt, err := redis.ParseURL(r.redisURL)
	if err != nil {
		return fmt.Errorf("failed to parse redis url: %v", err)
	}

	opt.DialTimeout = socketConnectTimeout
	opt.ReadTimeout = socketTimeout
	opt.WriteTimeout = socketTimeout
	opt.PoolSize = 10 * runtime.GOMAXPROCS(0)
	opt.MinIdleConns = 5
	opt.MaxConnAge = 5 * time.Minute
	opt.PoolTimeout = 5 * time.Second
	opt.IdleTimeout = 5 * time.Minute
	opt.IdleCheckFrequency = healthCheckInterval

	r.client = redis.NewClient(opt)

	// 测试连接，确保Redis可用
	ctx, cancel := context.WithTimeout(context.Background(), socketTimeout)
	defer cancel()

	err = r.client.Ping(ctx).Err()
	if err != nil {
		r.client = nil
		return fmt.Errorf("failed to ping redis: %v", err)
	}

	// 设置内存限制
	r.setMemoryLimit("allkeys-lru")

	return nil
}

// setMemoryLimit 动态设置Redis最大内存和内存淘汰策略
func (r *RedisHelper) setMemoryLimit(policy string) {
	// 这里应该从配置中获取相关参数
	// 在Go版本中，我们需要一个配置管理器来获取这些�?	maxmemory := getCacheRedisMaxmemory()
	if maxmemory == "" {
		if isBigMemoryMode() {
			maxmemory = "1024mb"
		} else {
			maxmemory = "256mb"
		}
	}

	ctx := context.Background()
	err := r.client.ConfigSet(ctx, "maxmemory", maxmemory).Err()
	if err != nil {
		// 日志记录错误，但在Go版本中暂时省�?	}

	err = r.client.ConfigSet(ctx, "maxmemory-policy", policy).Err()
	if err != nil {
		// 日志记录错误，但在Go版本中暂时省�?	}
}

// __getRegion 获取缓存的区
func (r *RedisHelper) __getRegion(region string) string {
	if region == "" {
		region = "DEFAULT"
	}
	return fmt.Sprintf("region:%s", url.QueryEscape(region))
}

// __makeRedisKey 获取缓存Key
func (r *RedisHelper) __makeRedisKey(region string, key string) string {
	// 使用region作为缓存键的一部分
	regionKey := r.__getRegion(region)
	return fmt.Sprintf("%s:key:%s", regionKey, url.QueryEscape(key))
}

// __getOriginalKey 从Redis键中提取原始key
func (r *RedisHelper) __getOriginalKey(redisKey string) string {
	parts := strings.Split(redisKey, ":key:")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return redisKey
}

// Set 设置缓存
func (r *RedisHelper) Set(key string, value interface{}, ttl int, region string) error {
	if region == "" {
		region = "DEFAULT"
	}

	err := r._connect()
	if err != nil {
		return err
	}

	redisKey := r.__makeRedisKey(region, key)
	// 对值进行序列化
	serializedValue, err := serialize(value)
	if err != nil {
		return fmt.Errorf("failed to serialize value: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), socketTimeout)
	defer cancel()

	var expiration time.Duration
	if ttl > 0 {
		expiration = time.Duration(ttl) * time.Second
	}

	err = r.client.Set(ctx, redisKey, serializedValue, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set key %s in region %s: %v", key, region, err)
	}

	return nil
}

// Exists 判断缓存键是否存�?func (r *RedisHelper) Exists(key string, region string) bool {
	if region == "" {
		region = "DEFAULT"
	}

	err := r._connect()
	if err != nil {
		return false
	}

	redisKey := r.__makeRedisKey(region, key)

	ctx, cancel := context.WithTimeout(context.Background(), socketTimeout)
	defer cancel()

	exists, err := r.client.Exists(ctx, redisKey).Result()
	if err != nil {
		return false
	}

	return exists == 1
}

// Get 获取缓存的�?func (r *RedisHelper) Get(key string, region string) (interface{}, error) {
	if region == "" {
		region = "DEFAULT"
	}

	err := r._connect()
	if err != nil {
		return nil, err
	}

	redisKey := r.__makeRedisKey(region, key)

	ctx, cancel := context.WithTimeout(context.Background(), socketTimeout)
	defer cancel()

	value, err := r.client.Get(ctx, redisKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Key不存�?		}
		return nil, fmt.Errorf("failed to get key %s in region %s: %v", key, region, err)
	}

	deserializedValue, err := deserialize([]byte(value))
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize value: %v", err)
	}

	return deserializedValue, nil
}

// Delete 删除缓存
func (r *RedisHelper) Delete(key string, region string) error {
	if region == "" {
		region = "DEFAULT"
	}

	err := r._connect()
	if err != nil {
		return err
	}

	redisKey := r.__makeRedisKey(region, key)

	ctx, cancel := context.WithTimeout(context.Background(), socketTimeout)
	defer cancel()

	err = r.client.Del(ctx, redisKey).Err()
	if err != nil {
		return fmt.Errorf("failed to delete key %s in region %s: %v", key, region, err)
	}

	return nil
}

// Clear 清除指定区域的缓存或全部缓存
func (r *RedisHelper) Clear(region string) error {
	err := r._connect()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), socketTimeout)
	defer cancel()

	if region != "" {
		cacheRegion := r.__getRegion(region)
		redisKey := fmt.Sprintf("%s:key:*", cacheRegion)

		iter := r.client.Scan(ctx, 0, redisKey, 0).Iterator()
		pipe := r.client.TxPipeline()
		count := 0

		for iter.Next(ctx) {
			pipe.Del(ctx, iter.Val())
			count++

			// 分批执行，避免管道过�?			if count%1000 == 0 {
				_, err := pipe.Exec(ctx)
				if err != nil {
					return fmt.Errorf("failed to clear cache for region %s: %v", region, err)
				}
				pipe = r.client.TxPipeline()
			}
		}

		if count%1000 != 0 {
			_, err := pipe.Exec(ctx)
			if err != nil {
				return fmt.Errorf("failed to clear cache for region %s: %v", region, err)
			}
		}

		if err := iter.Err(); err != nil {
			return fmt.Errorf("scan error: %v", err)
		}
	} else {
		err := r.client.FlushDB(ctx).Err()
		if err != nil {
			return fmt.Errorf("failed to clear all cache: %v", err)
		}
	}

	return nil
}

// Test 测试Redis连接�?func (r *RedisHelper) Test() bool {
	err := r._connect()
	if err != nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), socketTimeout)
	defer cancel()

	return r.client.Ping(ctx).Err() == nil
}

// Close 关闭Redis客户端的连接�?func (r *RedisHelper) Close() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.client != nil {
		err := r.client.Close()
		r.client = nil
		return err
	}

	return nil
}

// AsyncRedisHelper 异步Redis连接和操作助手类，单例模�?type AsyncRedisHelper struct {
	redisURL string
	client   *redis.Client
	mutex    sync.Mutex
}

// GetAsyncRedisHelper 获取AsyncRedisHelper单例实例
func GetAsyncRedisHelper() *AsyncRedisHelper {
	asyncRedisHelperOnce.Do(func() {
		asyncRedisHelperInstance = &AsyncRedisHelper{
			redisURL: getCacheBackendURL(),
		}
	})
	return asyncRedisHelperInstance
}

// _connect 建立异步Redis连接
func (a *AsyncRedisHelper) _connect() error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.client != nil {
		return nil
	}

	opt, err := redis.ParseURL(a.redisURL)
	if err != nil {
		return fmt.Errorf("failed to parse redis url: %v", err)
	}

	opt.DialTimeout = socketConnectTimeout
	opt.ReadTimeout = socketTimeout
	opt.WriteTimeout = socketTimeout
	opt.PoolSize = 10 * runtime.GOMAXPROCS(0)
	opt.MinIdleConns = 5
	opt.MaxConnAge = 5 * time.Minute
	opt.PoolTimeout = 5 * time.Second
	opt.IdleTimeout = 5 * time.Minute
	opt.IdleCheckFrequency = healthCheckInterval

	a.client = redis.NewClient(opt)

	// 测试连接，确保Redis可用
	ctx, cancel := context.WithTimeout(context.Background(), socketTimeout)
	defer cancel()

	err = a.client.Ping(ctx).Err()
	if err != nil {
		a.client = nil
		return fmt.Errorf("failed to ping redis: %v", err)
	}

	// 设置内存限制
	a.setMemoryLimit("allkeys-lru")

	return nil
}

// setMemoryLimit 动态设置Redis最大内存和内存淘汰策略
func (a *AsyncRedisHelper) setMemoryLimit(policy string) {
	maxmemory := getCacheRedisMaxmemory()
	if maxmemory == "" {
		if isBigMemoryMode() {
			maxmemory = "1024mb"
		} else {
			maxmemory = "256mb"
		}
	}

	ctx := context.Background()
	err := a.client.ConfigSet(ctx, "maxmemory", maxmemory).Err()
	if err != nil {
		// 日志记录错误，但在Go版本中暂时省�?	}

	err = a.client.ConfigSet(ctx, "maxmemory-policy", policy).Err()
	if err != nil {
		// 日志记录错误，但在Go版本中暂时省�?	}
}

// __getRegion 获取缓存的区
func (a *AsyncRedisHelper) __getRegion(region string) string {
	if region == "" {
		region = "DEFAULT"
	}
	return fmt.Sprintf("region:%s", url.QueryEscape(region))
}

// __makeRedisKey 获取缓存Key
func (a *AsyncRedisHelper) __makeRedisKey(region string, key string) string {
	// 使用region作为缓存键的一部分
	regionKey := a.__getRegion(region)
	return fmt.Sprintf("%s:key:%s", regionKey, url.QueryEscape(key))
}

// __getOriginalKey 从Redis键中提取原始key
func (a *AsyncRedisHelper) __getOriginalKey(redisKey string) string {
	parts := strings.Split(redisKey, ":key:")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return redisKey
}

// Set 异步设置缓存
func (a *AsyncRedisHelper) Set(key string, value interface{}, ttl int, region string) error {
	if region == "" {
		region = "DEFAULT"
	}

	err := a._connect()
	if err != nil {
		return err
	}

	redisKey := a.__makeRedisKey(region, key)
	// 对值进行序列化
	serializedValue, err := serialize(value)
	if err != nil {
		return fmt.Errorf("failed to serialize value: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), socketTimeout)
	defer cancel()

	var expiration time.Duration
	if ttl > 0 {
		expiration = time.Duration(ttl) * time.Second
	}

	err = a.client.Set(ctx, redisKey, serializedValue, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set key %s in region %s: %v", key, region, err)
	}

	return nil
}

// Exists 异步判断缓存键是否存�?func (a *AsyncRedisHelper) Exists(key string, region string) (bool, error) {
	if region == "" {
		region = "DEFAULT"
	}

	err := a._connect()
	if err != nil {
		return false, err
	}

	redisKey := a.__makeRedisKey(region, key)

	ctx, cancel := context.WithTimeout(context.Background(), socketTimeout)
	defer cancel()

	exists, err := a.client.Exists(ctx, redisKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check existence of key %s in region %s: %v", key, region, err)
	}

	return exists == 1, nil
}

// Get 异步获取缓存的�?func (a *AsyncRedisHelper) Get(key string, region string) (interface{}, error) {
	if region == "" {
		region = "DEFAULT"
	}

	err := a._connect()
	if err != nil {
		return nil, err
	}

	redisKey := a.__makeRedisKey(region, key)

	ctx, cancel := context.WithTimeout(context.Background(), socketTimeout)
	defer cancel()

	value, err := a.client.Get(ctx, redisKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Key不存�?		}
		return nil, fmt.Errorf("failed to get key %s in region %s: %v", key, region, err)
	}

	deserializedValue, err := deserialize([]byte(value))
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize value: %v", err)
	}

	return deserializedValue, nil
}

// Delete 异步删除缓存
func (a *AsyncRedisHelper) Delete(key string, region string) error {
	if region == "" {
		region = "DEFAULT"
	}

	err := a._connect()
	if err != nil {
		return err
	}

	redisKey := a.__makeRedisKey(region, key)

	ctx, cancel := context.WithTimeout(context.Background(), socketTimeout)
	defer cancel()

	err = a.client.Del(ctx, redisKey).Err()
	if err != nil {
		return fmt.Errorf("failed to delete key %s in region %s: %v", key, region, err)
	}

	return nil
}

// Clear 异步清除指定区域的缓存或全部缓存
func (a *AsyncRedisHelper) Clear(region string) error {
	err := a._connect()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), socketTimeout)
	defer cancel()

	if region != "" {
		cacheRegion := a.__getRegion(region)
		redisKey := fmt.Sprintf("%s:key:*", cacheRegion)

		iter := a.client.Scan(ctx, 0, redisKey, 0).Iterator()
		pipe := a.client.TxPipeline()
		count := 0

		for iter.Next(ctx) {
			pipe.Del(ctx, iter.Val())
			count++

			// 分批执行，避免管道过�?			if count%1000 == 0 {
				_, err := pipe.Exec(ctx)
				if err != nil {
					return fmt.Errorf("failed to clear cache for region %s: %v", region, err)
				}
				pipe = a.client.TxPipeline()
			}
		}

		if count%1000 != 0 {
			_, err := pipe.Exec(ctx)
			if err != nil {
				return fmt.Errorf("failed to clear cache for region %s: %v", region, err)
			}
		}

		if err := iter.Err(); err != nil {
			return fmt.Errorf("scan error: %v", err)
		}
	} else {
		err := a.client.FlushDB(ctx).Err()
		if err != nil {
			return fmt.Errorf("failed to clear all cache: %v", err)
		}
	}

	return nil
}

// Test 异步测试Redis连接�?func (a *AsyncRedisHelper) Test() bool {
	err := a._connect()
	if err != nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), socketTimeout)
	defer cancel()

	return a.client.Ping(ctx).Err() == nil
}

// Close 关闭异步Redis客户端的连接�?func (a *AsyncRedisHelper) Close() error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.client != nil {
		err := a.client.Close()
		a.client = nil
		return err
	}

	return nil
}

// 模拟配置获取函数，实际应从配置模块获�?func getCacheBackendURL() string {
	// 这里应该从配置中获取Redis URL
	// 暂时返回默认�?	return "redis://localhost:6379/0"
}

func getCacheRedisMaxmemory() string {
	// 这里应该从配置中获取最大内存设�?	// 暂时返回空字符串表示使用默认�?	return ""
}

func isBigMemoryMode() bool {
	// 这里应该从配置中获取大内存模式设�?	// 暂时返回false
	return false
}
