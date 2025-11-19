package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisHelper Redis辅助工具
type RedisHelper struct {
	client *redis.Client
}

// NewRedisHelper 创建Redis辅助工具实例
func NewRedisHelper(addr, password string, db int) *RedisHelper {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisHelper{
		client: rdb,
	}
}

// NewRedisHelperWithClient 使用现有客户端创建Redis辅助工具
func NewRedisHelperWithClient(client *redis.Client) *RedisHelper {
	return &RedisHelper{
		client: client,
	}
}

// GetClient 获取Redis客户端
func (r *RedisHelper) GetClient() *redis.Client {
	return r.client
}

// Ping 检查Redis连接
func (r *RedisHelper) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Set 设置键值对
func (r *RedisHelper) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// Get 获取值
func (r *RedisHelper) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// GetBytes 获取字节数组值
func (r *RedisHelper) GetBytes(ctx context.Context, key string) ([]byte, error) {
	return r.client.Get(ctx, key).Bytes()
}

// SetJSON 设置JSON值
func (r *RedisHelper) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("JSON序列化失败: %w", err)
	}
	return r.client.Set(ctx, key, data, expiration).Err()
}

// GetJSON 获取JSON值
func (r *RedisHelper) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), dest)
}

// Del 删除键
func (r *RedisHelper) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在
func (r *RedisHelper) Exists(ctx context.Context, keys ...string) (int64, error) {
	return r.client.Exists(ctx, keys...).Result()
}

// Expire 设置过期时间
func (r *RedisHelper) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

// TTL 获取剩余过期时间
func (r *RedisHelper) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, key).Result()
}

// Persist 移除过期时间
func (r *RedisHelper) Persist(ctx context.Context, key string) error {
	return r.client.Persist(ctx, key).Err()
}

// Incr 递增
func (r *RedisHelper) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

// IncrBy 按指定值递增
func (r *RedisHelper) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return r.client.IncrBy(ctx, key, value).Result()
}

// Decr 递减
func (r *RedisHelper) Decr(ctx context.Context, key string) (int64, error) {
	return r.client.Decr(ctx, key).Result()
}

// DecrBy 按指定值递减
func (r *RedisHelper) DecrBy(ctx context.Context, key string, value int64) (int64, error) {
	return r.client.DecrBy(ctx, key, value).Result()
}

// HSet 设置哈希字段
func (r *RedisHelper) HSet(ctx context.Context, key string, values ...interface{}) error {
	return r.client.HSet(ctx, key, values...).Err()
}

// HGet 获取哈希字段值
func (r *RedisHelper) HGet(ctx context.Context, key, field string) (string, error) {
	return r.client.HGet(ctx, key, field).Result()
}

// HGetAll 获取所有哈希字段
func (r *RedisHelper) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.client.HGetAll(ctx, key).Result()
}

// HDel 删除哈希字段
func (r *RedisHelper) HDel(ctx context.Context, key string, fields ...string) error {
	return r.client.HDel(ctx, key, fields...).Err()
}

// HExists 检查哈希字段是否存在
func (r *RedisHelper) HExists(ctx context.Context, key, field string) (bool, error) {
	return r.client.HExists(ctx, key, field).Result()
}

// HKeys 获取所有哈希字段名
func (r *RedisHelper) HKeys(ctx context.Context, key string) ([]string, error) {
	return r.client.HKeys(ctx, key).Result()
}

// HVals 获取所有哈希字段值
func (r *RedisHelper) HVals(ctx context.Context, key string) ([]string, error) {
	return r.client.HVals(ctx, key).Result()
}

// HLen 获取哈希字段数量
func (r *RedisHelper) HLen(ctx context.Context, key string) (int64, error) {
	return r.client.HLen(ctx, key).Result()
}

// HIncrBy 递增哈希字段值
func (r *RedisHelper) HIncrBy(ctx context.Context, key, field string, incr int64) (int64, error) {
	return r.client.HIncrBy(ctx, key, field, incr).Result()
}

// HIncrByFloat 递增哈希字段浮点值
func (r *RedisHelper) HIncrByFloat(ctx context.Context, key, field string, incr float64) (float64, error) {
	return r.client.HIncrByFloat(ctx, key, field, incr).Result()
}

// LPush 从左侧推入列表
func (r *RedisHelper) LPush(ctx context.Context, key string, values ...interface{}) error {
	return r.client.LPush(ctx, key, values...).Err()
}

// RPush 从右侧推入列表
func (r *RedisHelper) RPush(ctx context.Context, key string, values ...interface{}) error {
	return r.client.RPush(ctx, key, values...).Err()
}

// LPop 从左侧弹出列表元素
func (r *RedisHelper) LPop(ctx context.Context, key string) (string, error) {
	return r.client.LPop(ctx, key).Result()
}

// RPop 从右侧弹出列表元素
func (r *RedisHelper) RPop(ctx context.Context, key string) (string, error) {
	return r.client.RPop(ctx, key).Result()
}

// LLen 获取列表长度
func (r *RedisHelper) LLen(ctx context.Context, key string) (int64, error) {
	return r.client.LLen(ctx, key).Result()
}

// LIndex 获取列表指定索引元素
func (r *RedisHelper) LIndex(ctx context.Context, key string, index int64) (string, error) {
	return r.client.LIndex(ctx, key, index).Result()
}

// LRange 获取列表范围内的元素
func (r *RedisHelper) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.client.LRange(ctx, key, start, stop).Result()
}

// LTrim 修剪列表
func (r *RedisHelper) LTrim(ctx context.Context, key string, start, stop int64) error {
	return r.client.LTrim(ctx, key, start, stop).Err()
}

// LRem 移除列表元素
func (r *RedisHelper) LRem(ctx context.Context, key string, count int64, value interface{}) error {
	return r.client.LRem(ctx, key, count, value).Err()
}

// SAdd 添加集合成员
func (r *RedisHelper) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return r.client.SAdd(ctx, key, members...).Err()
}

// SMembers 获取所有集合成员
func (r *RedisHelper) SMembers(ctx context.Context, key string) ([]string, error) {
	return r.client.SMembers(ctx, key).Result()
}

// SCard 获取集合成员数量
func (r *RedisHelper) SCard(ctx context.Context, key string) (int64, error) {
	return r.client.SCard(ctx, key).Result()
}

// SIsMember 检查是否为集合成员
func (r *RedisHelper) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return r.client.SIsMember(ctx, key, member).Result()
}

// SRem 移除集合成员
func (r *RedisHelper) SRem(ctx context.Context, key string, members ...interface{}) error {
	return r.client.SRem(ctx, key, members...).Err()
}

// SPop 随机弹出集合成员
func (r *RedisHelper) SPop(ctx context.Context, key string) (string, error) {
	return r.client.SPop(ctx, key).Result()
}

// SRandMember 随机获取集合成员
func (r *RedisHelper) SRandMember(ctx context.Context, key string, count int64) ([]string, error) {
	return r.client.SRandMemberN(ctx, key, count).Result()
}

// SInter 获取多个集合的交集
func (r *RedisHelper) SInter(ctx context.Context, keys ...string) ([]string, error) {
	return r.client.SInter(ctx, keys...).Result()
}

// SUnion 获取多个集合的并集
func (r *RedisHelper) SUnion(ctx context.Context, keys ...string) ([]string, error) {
	return r.client.SUnion(ctx, keys...).Result()
}

// SDiff 获取多个集合的差集
func (r *RedisHelper) SDiff(ctx context.Context, keys ...string) ([]string, error) {
	return r.client.SDiff(ctx, keys...).Result()
}

// ZAdd 添加有序集合成员
func (r *RedisHelper) ZAdd(ctx context.Context, key string, members ...*redis.Z) error {
	return r.client.ZAdd(ctx, key, members...).Err()
}

// ZRange 获取有序集合范围内的成员
func (r *RedisHelper) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.client.ZRange(ctx, key, start, stop).Result()
}

// ZRangeWithScores 获取有序集合范围内的成员及分数
func (r *RedisHelper) ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error) {
	return r.client.ZRangeWithScores(ctx, key, start, stop).Result()
}

// ZScore 获取有序集合成员分数
func (r *RedisHelper) ZScore(ctx context.Context, key, member string) (float64, error) {
	return r.client.ZScore(ctx, key, member).Result()
}

// ZRank 获取有序集合成员排名
func (r *RedisHelper) ZRank(ctx context.Context, key, member string) (int64, error) {
	return r.client.ZRank(ctx, key, member).Result()
}

// ZRevRank 获取有序集合成员倒序排名
func (r *RedisHelper) ZRevRank(ctx context.Context, key, member string) (int64, error) {
	return r.client.ZRevRank(ctx, key, member).Result()
}

// ZCard 获取有序集合成员数量
func (r *RedisHelper) ZCard(ctx context.Context, key string) (int64, error) {
	return r.client.ZCard(ctx, key).Result()
}

// ZCount 获取有序集合指定分数范围内的成员数量
func (r *RedisHelper) ZCount(ctx context.Context, key, min, max string) (int64, error) {
	return r.client.ZCount(ctx, key, min, max).Result()
}

// ZRem 移除有序集合成员
func (r *RedisHelper) ZRem(ctx context.Context, key string, members ...interface{}) error {
	return r.client.ZRem(ctx, key, members...).Err()
}

// ZIncrBy 递增有序集合成员分数
func (r *RedisHelper) ZIncrBy(ctx context.Context, key string, incr float64, member string) (float64, error) {
	return r.client.ZIncrBy(ctx, key, incr, member).Result()
}

// Keys 获取匹配模式的所有键
func (r *RedisHelper) Keys(ctx context.Context, pattern string) ([]string, error) {
	return r.client.Keys(ctx, pattern).Result()
}

// Scan 扫描键
func (r *RedisHelper) Scan(ctx context.Context, cursor uint64, match string, count int64) (keys []string, nextCursor uint64, err error) {
	return r.client.Scan(ctx, cursor, match, count).Result()
}

// FlushDB 清空当前数据库
func (r *RedisHelper) FlushDB(ctx context.Context) error {
	return r.client.FlushDB(ctx).Err()
}

// FlushAll 清空所有数据库
func (r *RedisHelper) FlushAll(ctx context.Context) error {
	return r.client.FlushAll(ctx).Err()
}

// Info 获取Redis信息
func (r *RedisHelper) Info(ctx context.Context, sections ...string) (string, error) {
	return r.client.Info(ctx, sections...).Result()
}

// ConfigGet 获取配置
func (r *RedisHelper) ConfigGet(ctx context.Context, parameter string) (map[string]string, error) {
	return r.client.ConfigGet(ctx, parameter).Result()
}

// ConfigSet 设置配置
func (r *RedisHelper) ConfigSet(ctx context.Context, parameter, value string) error {
	return r.client.ConfigSet(ctx, parameter, value).Err()
}

// Dbsize 获取当前数据库键数量
func (r *RedisHelper) Dbsize(ctx context.Context) (int64, error) {
	return r.client.DBSize(ctx).Result()
}

// LastSave 获取最后保存时间
func (r *RedisHelper) LastSave(ctx context.Context) (int64, error) {
	return r.client.LastSave(ctx).Result()
}

// BgSave 后台保存
func (r *RedisHelper) BgSave(ctx context.Context) error {
	return r.client.BgSave(ctx).Err()
}

// Save 前台保存
func (r *RedisHelper) Save(ctx context.Context) error {
	return r.client.Save(ctx).Err()
}

// Close 关闭连接
func (r *RedisHelper) Close() error {
	return r.client.Close()
}

// Pipeline 创建管道
func (r *RedisHelper) Pipeline() redis.Pipeliner {
	return r.client.Pipeline()
}

// TxPipeline 创建事务管道
func (r *RedisHelper) TxPipeline() redis.Pipeliner {
	return r.client.TxPipeline()
}

// Publish 发布消息
func (r *RedisHelper) Publish(ctx context.Context, channel string, message interface{}) error {
	return r.client.Publish(ctx, channel, message).Err()
}

// Subscribe 订阅频道
func (r *RedisHelper) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return r.client.Subscribe(ctx, channels...)
}

// PSubscribe 订阅模式
func (r *RedisHelper) PSubscribe(ctx context.Context, patterns ...string) *redis.PubSub {
	return r.client.PSubscribe(ctx, patterns...)
}

// Eval 执行Lua脚本
func (r *RedisHelper) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	return r.client.Eval(ctx, script, keys, args...).Result()
}

// EvalSha 执行Lua脚本（通过SHA）
func (r *RedisHelper) EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) (interface{}, error) {
	return r.client.EvalSha(ctx, sha1, keys, args...).Result()
}

// ScriptLoad 加载Lua脚本
func (r *RedisHelper) ScriptLoad(ctx context.Context, script string) (string, error) {
	return r.client.ScriptLoad(ctx, script).Result()
}

// ScriptExists 检查脚本是否存在
func (r *RedisHelper) ScriptExists(ctx context.Context, sha1 ...string) ([]bool, error) {
	return r.client.ScriptExists(ctx, sha1...).Result()
}

// ScriptFlush 清空脚本缓存
func (r *RedisHelper) ScriptFlush(ctx context.Context) error {
	return r.client.ScriptFlush(ctx).Err()
}

// BitOp 位操作
func (r *RedisHelper) BitOp(ctx context.Context, op string, destKey string, keys ...string) (int64, error) {
	return r.client.BitOp(ctx, op, destKey, keys...).Result()
}

// BitCount 位计数
func (r *RedisHelper) BitCount(ctx context.Context, key string, start, end int64) (int64, error) {
	return r.client.BitCount(ctx, key, &redis.BitCount{Start: start, End: end}).Result()
}

// BitPos 位位置
func (r *RedisHelper) BitPos(ctx context.Context, key string, bit int64, start, end int64) (int64, error) {
	return r.client.BitPos(ctx, key, bit, start, end).Result()
}

// SetBit 设置位
func (r *RedisHelper) SetBit(ctx context.Context, key string, offset int64, value int) (int64, error) {
	return r.client.SetBit(ctx, key, offset, value).Result()
}

// GetBit 获取位
func (r *RedisHelper) GetBit(ctx context.Context, key string, offset int64) (int64, error) {
	return r.client.GetBit(ctx, key, offset).Result()
}

// GeoAdd 添加地理位置
func (r *RedisHelper) GeoAdd(ctx context.Context, key string, location *redis.GeoLocation) error {
	return r.client.GeoAdd(ctx, key, location).Err()
}

// GeoDist 计算地理位置距离
func (r *RedisHelper) GeoDist(ctx context.Context, key string, member1, member2, unit string) (float64, error) {
	return r.client.GeoDist(ctx, key, member1, member2, unit).Result()
}

// GeoPos 获取地理位置坐标
func (r *RedisHelper) GeoPos(ctx context.Context, key string, members ...string) ([]*redis.GeoPos, error) {
	return r.client.GeoPos(ctx, key, members...).Result()
}

// GeoRadius 根据坐标查询地理位置
func (r *RedisHelper) GeoRadius(ctx context.Context, key string, longitude, latitude float64, query *redis.GeoRadiusQuery) ([]redis.GeoLocation, error) {
	return r.client.GeoRadius(ctx, key, longitude, latitude, query).Result()
}

// HyperLogLogAdd 添加到HyperLogLog
func (r *RedisHelper) HyperLogLogAdd(ctx context.Context, key string, values ...interface{}) error {
	return r.client.PFAdd(ctx, key, values...).Err()
}

// HyperLogLogCount 获取HyperLogLog基数
func (r *RedisHelper) HyperLogLogCount(ctx context.Context, keys ...string) (int64, error) {
	return r.client.PFCount(ctx, keys...).Result()
}

// HyperLogLogMerge 合并HyperLogLog
func (r *RedisHelper) HyperLogLogMerge(ctx context.Context, dest string, keys ...string) error {
	return r.client.PFMerge(ctx, dest, keys...).Err()
}

// XAdd 添加流消息
func (r *RedisHelper) XAdd(ctx context.Context, stream string, values map[string]interface{}) (string, error) {
	return r.client.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		Values: values,
	}).Result()
}

// XRead 读取流消息
func (r *RedisHelper) XRead(ctx context.Context, streams ...string) (map[string]map[string]map[string]string, error) {
	xStreams := make([]redis.XStream, len(streams)/2)
	for i := 0; i < len(streams); i += 2 {
		xStreams[i/2] = redis.XStream{
			Stream: streams[i],
			Messages: []redis.XMessage{
				{ID: streams[i+1]},
			},
		}
	}
	result, err := r.client.XRead(ctx, &redis.XReadArgs{
		Streams: xStreams,
	}).Result()
	if err != nil {
		return nil, err
	}

	output := make(map[string]map[string]map[string]string)
	for _, stream := range result {
		output[stream.Stream] = make(map[string]map[string]string)
		for _, msg := range stream.Messages {
			output[stream.Stream][msg.ID] = msg.Values
		}
	}
	return output, nil
}

// XLen 获取流长度
func (r *RedisHelper) XLen(ctx context.Context, stream string) (int64, error) {
	return r.client.XLen(ctx, stream).Result()
}

// XDel 删除流消息
func (r *RedisHelper) XDel(ctx context.Context, stream string, ids ...string) (int64, error) {
	return r.client.XDel(ctx, stream, ids...).Result()
}

// XTrim 修剪流
func (r *RedisHelper) XTrim(ctx context.Context, stream string, maxLen int64) (int64, error) {
	return r.client.XTrim(ctx, stream, maxLen).Result()
}

// GetInt 获取整数值
func (r *RedisHelper) GetInt(ctx context.Context, key string) (int, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(val)
}

// GetInt64 获取64位整数值
func (r *RedisHelper) GetInt64(ctx context.Context, key string) (int64, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(val, 10, 64)
}

// GetFloat64 获取浮点数值
func (r *RedisHelper) GetFloat64(ctx context.Context, key string) (float64, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(val, 64)
}

// GetBool 获取布尔值
func (r *RedisHelper) GetBool(ctx context.Context, key string) (bool, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return strconv.ParseBool(val)
}

// SetInt 设置整数值
func (r *RedisHelper) SetInt(ctx context.Context, key string, value int, expiration time.Duration) error {
	return r.client.Set(ctx, key, strconv.Itoa(value), expiration).Err()
}

// SetInt64 设置64位整数值
func (r *RedisHelper) SetInt64(ctx context.Context, key string, value int64, expiration time.Duration) error {
	return r.client.Set(ctx, key, strconv.FormatInt(value, 10), expiration).Err()
}

// SetFloat64 设置浮点数值
func (r *RedisHelper) SetFloat64(ctx context.Context, key string, value float64, expiration time.Duration) error {
	return r.client.Set(ctx, key, strconv.FormatFloat(value, 'f', -1, 64), expiration).Err()
}

// SetBool 设置布尔值
func (r *RedisHelper) SetBool(ctx context.Context, key string, value bool, expiration time.Duration) error {
	return r.client.Set(ctx, key, strconv.FormatBool(value), expiration).Err()
}