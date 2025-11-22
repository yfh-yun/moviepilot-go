package utils

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"moviepilot-go/pkg/errors"
	"moviepilot-go/pkg/logger"
)

// RedisHelper Redis辅助工具
type RedisHelper struct {
	client *redis.Client
}

// NewRedisHelper 创建Redis辅助工具实例
func NewRedisHelper(addr, password string, db int) *RedisHelper {
	logger.Debug("Creating Redis helper instance",
		zap.String("addr", addr),
		zap.Int("db", db),
		zap.String("func", "NewRedisHelper"))

	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		PoolSize:     10,              // 连接池大小
		MinIdleConns: 3,               // 最小空闲连接数
		MaxRetries:   3,               // 最大重试次数
		DialTimeout:  5 * time.Second, // 连接超时
		ReadTimeout:  3 * time.Second, // 读取超时
		WriteTimeout: 3 * time.Second, // 写入超时
		PoolTimeout:  4 * time.Second, // 连接池超时
	})

	return &RedisHelper{
		client: rdb,
	}
}

// NewRedisHelperWithClient 使用现有客户端创建Redis辅助工具
func NewRedisHelperWithClient(client *redis.Client) *RedisHelper {
	logger.Debug("Creating Redis helper with existing client",
		zap.String("func", "NewRedisHelperWithClient"))

	return &RedisHelper{
		client: client,
	}
}

// GetClient 获取Redis客户端
func (r *RedisHelper) GetClient() *redis.Client {
	logger.Debug("Getting Redis client", zap.String("func", "GetClient"))
	return r.client
}

// Ping 检查Redis连接
func (r *RedisHelper) Ping(ctx context.Context) error {
	logger.Debug("Pinging Redis", zap.String("func", "Ping"))

	err := r.client.Ping(ctx).Err()
	if err != nil {
		logger.Error("Redis ping failed",
			zap.String("error", err.Error()),
			zap.String("func", "Ping"))
		return errors.NewAppError(500, "Redis connection failed", err.Error())
	}

	logger.Info("Redis ping successful", zap.String("func", "Ping"))
	return nil
}

// Set 设置键值对
func (r *RedisHelper) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if key == "" {
		err := errors.NewAppError(400, "Key cannot be empty", "")
		logger.Error("Invalid key for Redis Set",
			zap.String("error", err.Error()),
			zap.String("func", "Set"))
		return err
	}

	logger.Debug("Setting Redis key",
		zap.String("key", key),
		zap.Duration("expiration", expiration),
		zap.String("func", "Set"))

	err := r.client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		logger.Error("Failed to set Redis key",
			zap.String("key", key),
			zap.String("error", err.Error()),
			zap.String("func", "Set"))
		return errors.NewAppError(500, "Failed to set Redis key", err.Error())
	}

	return nil
}

// Get 获取值
func (r *RedisHelper) Get(ctx context.Context, key string) (string, error) {
	if key == "" {
		err := errors.NewAppError(400, "Key cannot be empty", "")
		logger.Error("Invalid key for Redis Get",
			zap.String("error", err.Error()),
			zap.String("func", "Get"))
		return "", err
	}

	logger.Debug("Getting Redis key",
		zap.String("key", key),
		zap.String("func", "Get"))

	result, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			logger.Debug("Redis key not found",
				zap.String("key", key),
				zap.String("func", "Get"))
			return "", errors.NewAppError(404, "Key not found", key)
		}
		logger.Error("Failed to get Redis key",
			zap.String("key", key),
			zap.String("error", err.Error()),
			zap.String("func", "Get"))
		return "", errors.NewAppError(500, "Failed to get Redis key", err.Error())
	}

	return result, nil
}

// GetBytes 获取字节数组值
func (r *RedisHelper) GetBytes(ctx context.Context, key string) ([]byte, error) {
	if key == "" {
		err := errors.NewAppError(400, "Key cannot be empty", "")
		logger.Error("Invalid key for Redis GetBytes",
			zap.String("error", err.Error()),
			zap.String("func", "GetBytes"))
		return nil, err
	}

	logger.Debug("Getting Redis key as bytes",
		zap.String("key", key),
		zap.String("func", "GetBytes"))

	result, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			logger.Debug("Redis key not found",
				zap.String("key", key),
				zap.String("func", "GetBytes"))
			return nil, errors.NewAppError(404, "Key not found", key)
		}
		logger.Error("Failed to get Redis key as bytes",
			zap.String("key", key),
			zap.String("error", err.Error()),
			zap.String("func", "GetBytes"))
		return nil, errors.NewAppError(500, "Failed to get Redis key", err.Error())
	}

	return result, nil
}

// SetJSON 设置JSON值
func (r *RedisHelper) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if key == "" {
		err := errors.NewAppError(400, "Key cannot be empty", "")
		logger.Error("Invalid key for Redis SetJSON",
			zap.String("error", err.Error()),
			zap.String("func", "SetJSON"))
		return err
	}

	logger.Debug("Setting Redis key as JSON",
		zap.String("key", key),
		zap.Duration("expiration", expiration),
		zap.String("func", "SetJSON"))

	data, err := json.Marshal(value)
	if err != nil {
		logger.Error("JSON serialization failed",
			zap.String("key", key),
			zap.String("error", err.Error()),
			zap.String("func", "SetJSON"))
		return errors.NewAppError(500, "JSON serialization failed", err.Error())
	}

	err = r.client.Set(ctx, key, data, expiration).Err()
	if err != nil {
		logger.Error("Failed to set Redis key as JSON",
			zap.String("key", key),
			zap.String("error", err.Error()),
			zap.String("func", "SetJSON"))
		return errors.NewAppError(500, "Failed to set Redis key", err.Error())
	}

	return nil
}

// GetJSON 获取JSON值
func (r *RedisHelper) GetJSON(ctx context.Context, key string, dest interface{}) error {
	if key == "" {
		err := errors.NewAppError(400, "Key cannot be empty", "")
		logger.Error("Invalid key for Redis GetJSON",
			zap.String("error", err.Error()),
			zap.String("func", "GetJSON"))
		return err
	}

	if dest == nil {
		err := errors.NewAppError(400, "Destination cannot be nil", "")
		logger.Error("Invalid destination for Redis GetJSON",
			zap.String("error", err.Error()),
			zap.String("func", "GetJSON"))
		return err
	}

	logger.Debug("Getting Redis key as JSON",
		zap.String("key", key),
		zap.String("func", "GetJSON"))

	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			logger.Debug("Redis key not found",
				zap.String("key", key),
				zap.String("func", "GetJSON"))
			return errors.NewAppError(404, "Key not found", key)
		}
		logger.Error("Failed to get Redis key for JSON",
			zap.String("key", key),
			zap.String("error", err.Error()),
			zap.String("func", "GetJSON"))
		return errors.NewAppError(500, "Failed to get Redis key", err.Error())
	}

	err = json.Unmarshal([]byte(data), dest)
	if err != nil {
		logger.Error("JSON deserialization failed",
			zap.String("key", key),
			zap.String("error", err.Error()),
			zap.String("func", "GetJSON"))
		return errors.NewAppError(500, "JSON deserialization failed", err.Error())
	}

	return nil
}

// Del 删除键
func (r *RedisHelper) Del(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		err := errors.NewAppError(400, "Keys cannot be empty", "")
		logger.Error("Invalid keys for Redis Del",
			zap.String("error", err.Error()),
			zap.String("func", "Del"))
		return err
	}

	logger.Debug("Deleting Redis keys",
		zap.Strings("keys", keys),
		zap.String("func", "Del"))

	err := r.client.Del(ctx, keys...).Err()
	if err != nil {
		logger.Error("Failed to delete Redis keys",
			zap.Strings("keys", keys),
			zap.String("error", err.Error()),
			zap.String("func", "Del"))
		return errors.NewAppError(500, "Failed to delete Redis keys", err.Error())
	}

	return nil
}

// Exists 检查键是否存在
func (r *RedisHelper) Exists(ctx context.Context, keys ...string) (int64, error) {
	if len(keys) == 0 {
		err := errors.NewAppError(400, "Keys cannot be empty", "")
		logger.Error("Invalid keys for Redis Exists",
			zap.String("error", err.Error()),
			zap.String("func", "Exists"))
		return 0, err
	}

	logger.Debug("Checking Redis keys existence",
		zap.Strings("keys", keys),
		zap.String("func", "Exists"))

	count, err := r.client.Exists(ctx, keys...).Result()
	if err != nil {
		logger.Error("Failed to check Redis keys existence",
			zap.Strings("keys", keys),
			zap.String("error", err.Error()),
			zap.String("func", "Exists"))
		return 0, errors.NewAppError(500, "Failed to check key existence", err.Error())
	}

	return count, nil
}

// Expire 设置过期时间
func (r *RedisHelper) Expire(ctx context.Context, key string, expiration time.Duration) error {
	if key == "" {
		err := errors.NewAppError(400, "Key cannot be empty", "")
		logger.Error("Invalid key for Redis Expire",
			zap.String("error", err.Error()),
			zap.String("func", "Expire"))
		return err
	}

	logger.Debug("Setting Redis key expiration",
		zap.String("key", key),
		zap.Duration("expiration", expiration),
		zap.String("func", "Expire"))

	err := r.client.Expire(ctx, key, expiration).Err()
	if err != nil {
		logger.Error("Failed to set Redis key expiration",
			zap.String("key", key),
			zap.Duration("expiration", expiration),
			zap.String("error", err.Error()),
			zap.String("func", "Expire"))
		return errors.NewAppError(500, "Failed to set key expiration", err.Error())
	}

	return nil
}

// TTL 获取剩余过期时间
func (r *RedisHelper) TTL(ctx context.Context, key string) (time.Duration, error) {
	if key == "" {
		err := errors.NewAppError(400, "Key cannot be empty", "")
		logger.Error("Invalid key for Redis TTL",
			zap.String("error", err.Error()),
			zap.String("func", "TTL"))
		return 0, err
	}

	logger.Debug("Getting Redis key TTL",
		zap.String("key", key),
		zap.String("func", "TTL"))

	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		logger.Error("Failed to get Redis key TTL",
			zap.String("key", key),
			zap.String("error", err.Error()),
			zap.String("func", "TTL"))
		return 0, errors.NewAppError(500, "Failed to get key TTL", err.Error())
	}

	return ttl, nil
}

// Persist 移除过期时间
func (r *RedisHelper) Persist(ctx context.Context, key string) error {
	if key == "" {
		err := errors.NewAppError(400, "Key cannot be empty", "")
		logger.Error("Invalid key for Redis Persist",
			zap.String("error", err.Error()),
			zap.String("func", "Persist"))
		return err
	}

	logger.Debug("Removing Redis key expiration",
		zap.String("key", key),
		zap.String("func", "Persist"))

	err := r.client.Persist(ctx, key).Err()
	if err != nil {
		logger.Error("Failed to remove Redis key expiration",
			zap.String("key", key),
			zap.String("error", err.Error()),
			zap.String("func", "Persist"))
		return errors.NewAppError(500, "Failed to remove key expiration", err.Error())
	}

	return nil
}

// Incr 递增
func (r *RedisHelper) Incr(ctx context.Context, key string) (int64, error) {
	if key == "" {
		err := errors.NewAppError(400, "Key cannot be empty", "")
		logger.Error("Invalid key for Redis Incr",
			zap.String("error", err.Error()),
			zap.String("func", "Incr"))
		return 0, err
	}

	logger.Debug("Incrementing Redis key",
		zap.String("key", key),
		zap.String("func", "Incr"))

	result, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		logger.Error("Failed to increment Redis key",
			zap.String("key", key),
			zap.String("error", err.Error()),
			zap.String("func", "Incr"))
		return 0, errors.NewAppError(500, "Failed to increment key", err.Error())
	}

	return result, nil
}

// IncrBy 按指定值递增
func (r *RedisHelper) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	if key == "" {
		err := errors.NewAppError(400, "Key cannot be empty", "")
		logger.Error("Invalid key for Redis IncrBy",
			zap.String("error", err.Error()),
			zap.String("func", "IncrBy"))
		return 0, err
	}

	logger.Debug("Incrementing Redis key by value",
		zap.String("key", key),
		zap.Int64("value", value),
		zap.String("func", "IncrBy"))

	result, err := r.client.IncrBy(ctx, key, value).Result()
	if err != nil {
		logger.Error("Failed to increment Redis key by value",
			zap.String("key", key),
			zap.Int64("value", value),
			zap.String("error", err.Error()),
			zap.String("func", "IncrBy"))
		return 0, errors.NewAppError(500, "Failed to increment key by value", err.Error())
	}

	return result, nil
}

// Decr 递减
func (r *RedisHelper) Decr(ctx context.Context, key string) (int64, error) {
	if key == "" {
		err := errors.NewAppError(400, "Key cannot be empty", "")
		logger.Error("Invalid key for Redis Decr",
			zap.String("error", err.Error()),
			zap.String("func", "Decr"))
		return 0, err
	}

	logger.Debug("Decrementing Redis key",
		zap.String("key", key),
		zap.String("func", "Decr"))

	result, err := r.client.Decr(ctx, key).Result()
	if err != nil {
		logger.Error("Failed to decrement Redis key",
			zap.String("key", key),
			zap.String("error", err.Error()),
			zap.String("func", "Decr"))
		return 0, errors.NewAppError(500, "Failed to decrement key", err.Error())
	}

	return result, nil
}

// DecrBy 按指定值递减
func (r *RedisHelper) DecrBy(ctx context.Context, key string, value int64) (int64, error) {
	if key == "" {
		err := errors.NewAppError(400, "Key cannot be empty", "")
		logger.Error("Invalid key for Redis DecrBy",
			zap.String("error", err.Error()),
			zap.String("func", "DecrBy"))
		return 0, err
	}

	logger.Debug("Decrementing Redis key by value",
		zap.String("key", key),
		zap.Int64("value", value),
		zap.String("func", "DecrBy"))

	result, err := r.client.DecrBy(ctx, key, value).Result()
	if err != nil {
		logger.Error("Failed to decrement Redis key by value",
			zap.String("key", key),
			zap.Int64("value", value),
			zap.String("error", err.Error()),
			zap.String("func", "DecrBy"))
		return 0, errors.NewAppError(500, "Failed to decrement key by value", err.Error())
	}

	return result, nil
}

// HSet 设置哈希字段
func (r *RedisHelper) HSet(ctx context.Context, key string, values ...interface{}) error {
	if key == "" {
		err := errors.NewAppError(400, "Key cannot be empty", "")
		logger.Error("Invalid key for Redis HSet",
			zap.String("error", err.Error()),
			zap.String("func", "HSet"))
		return err
	}

	if len(values) == 0 {
		err := errors.NewAppError(400, "Values cannot be empty", "")
		logger.Error("Invalid values for Redis HSet",
			zap.String("error", err.Error()),
			zap.String("func", "HSet"))
		return err
	}

	logger.Debug("Setting Redis hash fields",
		zap.String("key", key),
		zap.Int("value_count", len(values)),
		zap.String("func", "HSet"))

	err := r.client.HSet(ctx, key, values...).Err()
	if err != nil {
		logger.Error("Failed to set Redis hash fields",
			zap.String("key", key),
			zap.String("error", err.Error()),
			zap.String("func", "HSet"))
		return errors.NewAppError(500, "Failed to set hash fields", err.Error())
	}

	return nil
}

// HGet 获取哈希字段值
func (r *RedisHelper) HGet(ctx context.Context, key, field string) (string, error) {
	if key == "" || field == "" {
		err := errors.NewAppError(400, "Key and field cannot be empty", "")
		logger.Error("Invalid key or field for Redis HGet",
			zap.String("error", err.Error()),
			zap.String("func", "HGet"))
		return "", err
	}

	logger.Debug("Getting Redis hash field",
		zap.String("key", key),
		zap.String("field", field),
		zap.String("func", "HGet"))

	result, err := r.client.HGet(ctx, key, field).Result()
	if err != nil {
		if err == redis.Nil {
			logger.Debug("Redis hash field not found",
				zap.String("key", key),
				zap.String("field", field),
				zap.String("func", "HGet"))
			return "", errors.NewAppError(404, "Hash field not found", field)
		}
		logger.Error("Failed to get Redis hash field",
			zap.String("key", key),
			zap.String("field", field),
			zap.String("error", err.Error()),
			zap.String("func", "HGet"))
		return "", errors.NewAppError(500, "Failed to get hash field", err.Error())
	}

	return result, nil
}

// HGetAll 获取所有哈希字段
func (r *RedisHelper) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	if key == "" {
		err := errors.NewAppError(400, "Key cannot be empty", "")
		logger.Error("Invalid key for Redis HGetAll",
			zap.String("error", err.Error()),
			zap.String("func", "HGetAll"))
		return nil, err
	}

	logger.Debug("Getting all Redis hash fields",
		zap.String("key", key),
		zap.String("func", "HGetAll"))

	result, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		logger.Error("Failed to get all Redis hash fields",
			zap.String("key", key),
			zap.String("error", err.Error()),
			zap.String("func", "HGetAll"))
		return nil, errors.NewAppError(500, "Failed to get all hash fields", err.Error())
	}

	return result, nil
}

// HDel 删除哈希字段
func (r *RedisHelper) HDel(ctx context.Context, key string, fields ...string) error {
	if key == "" {
		err := errors.NewAppError(400, "Key cannot be empty", "")
		logger.Error("Invalid key for Redis HDel",
			zap.String("error", err.Error()),
			zap.String("func", "HDel"))
		return err
	}

	if len(fields) == 0 {
		err := errors.NewAppError(400, "Fields cannot be empty", "")
		logger.Error("Invalid fields for Redis HDel",
			zap.String("error", err.Error()),
			zap.String("func", "HDel"))
		return err
	}

	logger.Debug("Deleting Redis hash fields",
		zap.String("key", key),
		zap.Strings("fields", fields),
		zap.String("func", "HDel"))

	err := r.client.HDel(ctx, key, fields...).Err()
	if err != nil {
		logger.Error("Failed to delete Redis hash fields",
			zap.String("key", key),
			zap.Strings("fields", fields),
			zap.String("error", err.Error()),
			zap.String("func", "HDel"))
		return errors.NewAppError(500, "Failed to delete hash fields", err.Error())
	}

	return nil
}

// HExists 检查哈希字段是否存在
func (r *RedisHelper) HExists(ctx context.Context, key, field string) (bool, error) {
	if key == "" || field == "" {
		err := errors.NewAppError(400, "Key and field cannot be empty", "")
		logger.Error("Invalid key or field for Redis HExists",
			zap.String("error", err.Error()),
			zap.String("func", "HExists"))
		return false, err
	}

	logger.Debug("Checking Redis hash field existence",
		zap.String("key", key),
		zap.String("field", field),
		zap.String("func", "HExists"))

	exists, err := r.client.HExists(ctx, key, field).Result()
	if err != nil {
		logger.Error("Failed to check Redis hash field existence",
			zap.String("key", key),
			zap.String("field", field),
			zap.String("error", err.Error()),
			zap.String("func", "HExists"))
		return false, errors.NewAppError(500, "Failed to check hash field existence", err.Error())
	}

	return exists, nil
}

// HKeys 获取所有哈希字段名
func (r *RedisHelper) HKeys(ctx context.Context, key string) ([]string, error) {
	if key == "" {
		err := errors.NewAppError(400, "Key cannot be empty", "")
		logger.Error("Invalid key for Redis HKeys",
			zap.String("error", err.Error()),
			zap.String("func", "HKeys"))
		return nil, err
	}

	logger.Debug("Getting Redis hash field names",
		zap.String("key", key),
		zap.String("func", "HKeys"))

	keys, err := r.client.HKeys(ctx, key).Result()
	if err != nil {
		logger.Error("Failed to get Redis hash field names",
			zap.String("key", key),
			zap.String("error", err.Error()),
			zap.String("func", "HKeys"))
		return nil, errors.NewAppError(500, "Failed to get hash field names", err.Error())
	}

	return keys, nil
}

// HVals 获取所有哈希字段值
func (r *RedisHelper) HVals(ctx context.Context, key string) ([]string, error) {
	if key == "" {
		err := errors.NewAppError(400, "Key cannot be empty", "")
		logger.Error("Invalid key for Redis HVals",
			zap.String("error", err.Error()),
			zap.String("func", "HVals"))
		return nil, err
	}

	logger.Debug("Getting Redis hash field values",
		zap.String("key", key),
		zap.String("func", "HVals"))

	values, err := r.client.HVals(ctx, key).Result()
	if err != nil {
		logger.Error("Failed to get Redis hash field values",
			zap.String("key", key),
			zap.String("error", err.Error()),
			zap.String("func", "HVals"))
		return nil, errors.NewAppError(500, "Failed to get hash field values", err.Error())
	}

	return values, nil
}

// HLen 获取哈希字段数量
func (r *RedisHelper) HLen(ctx context.Context, key string) (int64, error) {
	if key == "" {
		err := errors.NewAppError(400, "Key cannot be empty", "")
		logger.Error("Invalid key for Redis HLen",
			zap.String("error", err.Error()),
			zap.String("func", "HLen"))
		return 0, err
	}

	logger.Debug("Getting Redis hash field count",
		zap.String("key", key),
		zap.String("func", "HLen"))

	count, err := r.client.HLen(ctx, key).Result()
	if err != nil {
		logger.Error("Failed to get Redis hash field count",
			zap.String("key", key),
			zap.String("error", err.Error()),
			zap.String("func", "HLen"))
		return 0, errors.NewAppError(500, "Failed to get hash field count", err.Error())
	}

	return count, nil
}

// HIncrBy 递增哈希字段值
func (r *RedisHelper) HIncrBy(ctx context.Context, key, field string, incr int64) (int64, error) {
	if key == "" || field == "" {
		err := errors.NewAppError(400, "Key and field cannot be empty", "")
		logger.Error("Invalid key or field for Redis HIncrBy",
			zap.String("error", err.Error()),
			zap.String("func", "HIncrBy"))
		return 0, err
	}

	logger.Debug("Incrementing Redis hash field",
		zap.String("key", key),
		zap.String("field", field),
		zap.Int64("increment", incr),
		zap.String("func", "HIncrBy"))

	result, err := r.client.HIncrBy(ctx, key, field, incr).Result()
	if err != nil {
		logger.Error("Failed to increment Redis hash field",
			zap.String("key", key),
			zap.String("field", field),
			zap.Int64("increment", incr),
			zap.String("error", err.Error()),
			zap.String("func", "HIncrBy"))
		return 0, errors.NewAppError(500, "Failed to increment hash field", err.Error())
	}

	return result, nil
}

// HIncrByFloat 递增哈希字段浮点值
func (r *RedisHelper) HIncrByFloat(ctx context.Context, key, field string, incr float64) (float64, error) {
	if key == "" || field == "" {
		err := errors.NewAppError(400, "Key and field cannot be empty", "")
		logger.Error("Invalid key or field for Redis HIncrByFloat",
			zap.String("error", err.Error()),
			zap.String("func", "HIncrByFloat"))
		return 0, err
	}

	logger.Debug("Incrementing Redis hash field by float",
		zap.String("key", key),
		zap.String("field", field),
		zap.Float64("increment", incr),
		zap.String("func", "HIncrByFloat"))

	result, err := r.client.HIncrByFloat(ctx, key, field, incr).Result()
	if err != nil {
		logger.Error("Failed to increment Redis hash field by float",
			zap.String("key", key),
			zap.String("field", field),
			zap.Float64("increment", incr),
			zap.String("error", err.Error()),
			zap.String("func", "HIncrByFloat"))
		return 0, errors.NewAppError(500, "Failed to increment hash field by float", err.Error())
	}

	return result, nil
}

// LPush 从左侧推入列表
func (r *RedisHelper) LPush(ctx context.Context, key string, values ...interface{}) error {
	if key == "" {
		err := errors.NewAppError(400, "Key cannot be empty", "")
		logger.Error("Invalid key for Redis LPush",
			zap.String("error", err.Error()),
			zap.String("func", "LPush"))
		return err
	}

	if len(values) == 0 {
		err := errors.NewAppError(400, "Values cannot be empty", "")
		logger.Error("Invalid values for Redis LPush",
			zap.String("error", err.Error()),
			zap.String("func", "LPush"))
		return err
	}

	logger.Debug("Pushing to Redis list from left",
		zap.String("key", key),
		zap.Int("value_count", len(values)),
		zap.String("func", "LPush"))

	err := r.client.LPush(ctx, key, values...).Err()
	if err != nil {
		logger.Error("Failed to push to Redis list from left",
			zap.String("key", key),
			zap.String("error", err.Error()),
			zap.String("func", "LPush"))
		return errors.NewAppError(500, "Failed to push to list", err.Error())
	}

	return nil
}

// RPush 从右侧推入列表
func (r *RedisHelper) RPush(ctx context.Context, key string, values ...interface{}) error {
	if key == "" {
		err := errors.NewAppError(400, "Key cannot be empty", "")
		logger.Error("Invalid key for Redis RPush",
			zap.String("error", err.Error()),
			zap.String("func", "RPush"))
		return err
	}

	if len(values) == 0 {
		err := errors.NewAppError(400, "Values cannot be empty", "")
		logger.Error("Invalid values for Redis RPush",
			zap.String("error", err.Error()),
			zap.String("func", "RPush"))
		return err
	}

	logger.Debug("Pushing to Redis list from right",
		zap.String("key", key),
		zap.Int("value_count", len(values)),
		zap.String("func", "RPush"))

	err := r.client.RPush(ctx, key, values...).Err()
	if err != nil {
		logger.Error("Failed to push to Redis list from right",
			zap.String("key", key),
			zap.String("error", err.Error()),
			zap.String("func", "RPush"))
		return errors.NewAppError(500, "Failed to push to list", err.Error())
	}

	return nil
}

// LPop 从左侧弹出列表元素
func (r *RedisHelper) LPop(ctx context.Context, key string) (string, error) {
	if key == "" {
		err := errors.NewAppError(400, "Key cannot be empty", "")
		logger.Error("Invalid key for Redis LPop",
			zap.String("error", err.Error()),
			zap.String("func", "LPop"))
		return "", err
	}

	logger.Debug("Popping from Redis list from left",
		zap.String("key", key),
		zap.String("func", "LPop"))

	result, err := r.client.LPop(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			logger.Debug("Redis list is empty",
				zap.String("key", key),
				zap.String("func", "LPop"))
			return "", errors.NewAppError(404, "List is empty", key)
		}
		logger.Error("Failed to pop from Redis list from left",
			zap.String("key", key),
			zap.String("error", err.Error()),
			zap.String("func", "LPop"))
		return "", errors.NewAppError(500, "Failed to pop from list", err.Error())
	}

	return result, nil
}

// RPop 从右侧弹出列表元素
func (r *RedisHelper) RPop(ctx context.Context, key string) (string, error) {
	if key == "" {
		err := errors.NewAppError(400, "Key cannot be empty", "")
		logger.Error("Invalid key for Redis RPop",
			zap.String("error", err.Error()),
			zap.String("func", "RPop"))
		return "", err
	}

	logger.Debug("Popping from Redis list from right",
		zap.String("key", key),
		zap.String("func", "RPop"))

	result, err := r.client.RPop(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			logger.Debug("Redis list is empty",
				zap.String("key", key),
				zap.String("func", "RPop"))
			return "", errors.NewAppError(404, "List is empty", key)
		}
		logger.Error("Failed to pop from Redis list from right",
			zap.String("key", key),
			zap.String("error", err.Error()),
			zap.String("func", "RPop"))
		return "", errors.NewAppError(500, "Failed to pop from list", err.Error())
	}

	return result, nil
}

// LLen 获取列表长度
func (r *RedisHelper) LLen(ctx context.Context, key string) (int64, error) {
	if key == "" {
		err := errors.NewAppError(400, "Key cannot be empty", "")
		logger.Error("Invalid key for Redis LLen",
			zap.String("error", err.Error()),
			zap.String("func", "LLen"))
		return 0, err
	}

	logger.Debug("Getting Redis list length",
		zap.String("key", key),
		zap.String("func", "LLen"))

	length, err := r.client.LLen(ctx, key).Result()
	if err != nil {
		logger.Error("Failed to get Redis list length",
			zap.String("key", key),
			zap.String("error", err.Error()),
			zap.String("func", "LLen"))
		return 0, errors.NewAppError(500, "Failed to get list length", err.Error())
	}

	return length, nil
}

// AllowRequest 令牌桶算法实现限流
// key: Redis键名
// limit: 每秒允许的请求数（令牌生成速率）
// burst: 突发容量（桶的最大容量）
// window: 时间窗口
func (r *RedisHelper) AllowRequest(key string, limit float64, burst int, window time.Duration) (bool, error) {
	ctx := context.Background()

	if key == "" {
		err := errors.NewAppError(400, "Key cannot be empty", "")
		logger.Error("Invalid key for Redis AllowRequest",
			zap.String("error", err.Error()),
			zap.String("func", "AllowRequest"))
		return false, err
	}

	if limit <= 0 || burst <= 0 {
		err := errors.NewAppError(400, "Limit and burst must be positive", "")
		logger.Error("Invalid limit or burst for Redis AllowRequest",
			zap.Float64("limit", limit),
			zap.Int("burst", burst),
			zap.String("func", "AllowRequest"))
		return false, err
	}

	logger.Debug("Checking rate limit with token bucket",
		zap.String("key", key),
		zap.Float64("limit", limit),
		zap.Int("burst", burst),
		zap.Duration("window", window),
		zap.String("func", "AllowRequest"))

	now := time.Now().UnixNano()

	// Lua 脚本实现令牌桶算法
	// KEYS[1]: 令牌桶键名
	// ARGV[1]: 当前时间戳（纳秒）
	// ARGV[2]: 令牌生成速率（每秒）
	// ARGV[3]: 桶容量
	// ARGV[4]: 时间窗口（纳秒）
	luaScript := `
		local key = KEYS[1]
		local now = tonumber(ARGV[1])
		local rate = tonumber(ARGV[2])
		local capacity = tonumber(ARGV[3])
		local window = tonumber(ARGV[4])
		
		-- 获取当前令牌数和上次更新时间
		local bucket = redis.call('HMGET', key, 'tokens', 'last_time')
		local tokens = tonumber(bucket[1])
		local last_time = tonumber(bucket[2])
		
		-- 如果是第一次访问，初始化令牌桶
		if tokens == nil then
			tokens = capacity
			last_time = now
		end
		
		-- 计算时间差（秒）
		local delta = (now - last_time) / 1e9
		
		-- 根据时间差补充令牌
		tokens = math.min(capacity, tokens + delta * rate)
		
		-- 尝试消费一个令牌
		local allowed = 0
		if tokens >= 1 then
			tokens = tokens - 1
			allowed = 1
		end
		
		-- 更新令牌桶状态
		redis.call('HMSET', key, 'tokens', tokens, 'last_time', now)
		redis.call('EXPIRE', key, math.ceil(window / 1e9))
		
		return allowed
	`

	// 执行 Lua 脚本
	result, err := r.client.Eval(ctx, luaScript, []string{key}, now, limit, burst, window.Nanoseconds()).Result()
	if err != nil {
		logger.Error("Failed to execute rate limit script",
			zap.String("key", key),
			zap.String("error", err.Error()),
			zap.String("func", "AllowRequest"))
		return false, errors.NewAppError(500, "Failed to check rate limit", err.Error())
	}

	allowed := result.(int64) == 1

	if !allowed {
		logger.Debug("Rate limit exceeded",
			zap.String("key", key),
			zap.Float64("limit", limit),
			zap.Int("burst", burst),
			zap.String("func", "AllowRequest"))
	}

	return allowed, nil
}
