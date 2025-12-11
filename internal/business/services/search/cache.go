package search

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// CacheService 搜索缓存服务
type CacheService interface {
	// Get 获取缓存
	Get(ctx context.Context, key string) ([]*SearchResult, error)

	// Set 设置缓存
	Set(ctx context.Context, key string, results []*SearchResult, ttl time.Duration) error

	// Delete 删除缓存
	Delete(ctx context.Context, key string) error

	// GenerateKey 生成缓存键
	GenerateKey(query *SearchQuery) string

	// Clear 清空所有搜索缓存
	Clear(ctx context.Context) error

	// GetStats 获取缓存统计
	GetStats(ctx context.Context) (*CacheStats, error)
}

// cacheService 缓存服务实现
type cacheService struct {
	redis  *redis.Client
	prefix string
	logger *zap.Logger
}

// NewCacheService 创建缓存服务
func NewCacheService(redis *redis.Client) CacheService {
	return &cacheService{
		redis:  redis,
		prefix: "search:cache:",
		logger: logger.GetLogger(),
	}
}

// CacheStats 缓存统计
type CacheStats struct {
	TotalKeys   int64   `json:"total_keys"`
	HitRate     float64 `json:"hit_rate"`
	MemoryUsage int64   `json:"memory_usage"` // 字节
}

// Get 获取缓存
func (s *cacheService) Get(ctx context.Context, key string) ([]*SearchResult, error) {
	fullKey := s.prefix + key

	s.logger.Debug("获取搜索缓存", zap.String("key", key))

	data, err := s.redis.Get(ctx, fullKey).Result()
	if err == redis.Nil {
		s.logger.Debug("缓存未命中", zap.String("key", key))
		return nil, nil
	}
	if err != nil {
		s.logger.Error("获取缓存失败", zap.Error(err))
		return nil, err
	}

	var results []*SearchResult
	if err := json.Unmarshal([]byte(data), &results); err != nil {
		s.logger.Error("解析缓存数据失败", zap.Error(err))
		return nil, err
	}

	s.logger.Debug("缓存命中", zap.String("key", key), zap.Int("count", len(results)))
	return results, nil
}

// Set 设置缓存
func (s *cacheService) Set(ctx context.Context, key string, results []*SearchResult, ttl time.Duration) error {
	fullKey := s.prefix + key

	s.logger.Debug("设置搜索缓存",
		zap.String("key", key),
		zap.Int("count", len(results)),
		zap.Duration("ttl", ttl),
	)

	data, err := json.Marshal(results)
	if err != nil {
		s.logger.Error("序列化缓存数据失败", zap.Error(err))
		return err
	}

	if err := s.redis.Set(ctx, fullKey, data, ttl).Err(); err != nil {
		s.logger.Error("设置缓存失败", zap.Error(err))
		return err
	}

	return nil
}

// Delete 删除缓存
func (s *cacheService) Delete(ctx context.Context, key string) error {
	fullKey := s.prefix + key

	s.logger.Debug("删除搜索缓存", zap.String("key", key))

	if err := s.redis.Del(ctx, fullKey).Err(); err != nil {
		s.logger.Error("删除缓存失败", zap.Error(err))
		return err
	}

	return nil
}

// GenerateKey 生成缓存键
func (s *cacheService) GenerateKey(query *SearchQuery) string {
	// 将查询参数序列化为字符串
	keyData := fmt.Sprintf("%s|%s|%s|%s|%d|%d",
		query.Keyword,
		query.Type,
		query.Quality,
		query.Resolution,
		query.MinSize,
		query.MaxSize,
	)

	// 使用 MD5 生成短键
	hash := md5.Sum([]byte(keyData))
	return hex.EncodeToString(hash[:])
}

// Clear 清空所有搜索缓存
func (s *cacheService) Clear(ctx context.Context) error {
	s.logger.Info("清空搜索缓存")

	// 获取所有匹配的键
	pattern := s.prefix + "*"
	iter := s.redis.Scan(ctx, 0, pattern, 0).Iterator()

	keys := make([]string, 0)
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		s.logger.Error("扫描缓存键失败", zap.Error(err))
		return err
	}

	if len(keys) > 0 {
		if err := s.redis.Del(ctx, keys...).Err(); err != nil {
			s.logger.Error("删除缓存失败", zap.Error(err))
			return err
		}
		s.logger.Info("清空缓存完成", zap.Int("count", len(keys)))
	}

	return nil
}

// GetStats 获取缓存统计
func (s *cacheService) GetStats(ctx context.Context) (*CacheStats, error) {
	s.logger.Debug("获取缓存统计")

	stats := &CacheStats{}

	// 统计键数量
	pattern := s.prefix + "*"
	iter := s.redis.Scan(ctx, 0, pattern, 0).Iterator()

	count := int64(0)
	totalSize := int64(0)

	for iter.Next(ctx) {
		count++

		// 获取键的大小
		key := iter.Val()
		size, err := s.redis.MemoryUsage(ctx, key).Result()
		if err == nil {
			totalSize += size
		}
	}

	if err := iter.Err(); err != nil {
		s.logger.Error("扫描缓存键失败", zap.Error(err))
		return nil, err
	}

	stats.TotalKeys = count
	stats.MemoryUsage = totalSize

	// 计算命中率（简化版，实际应该记录命中和未命中次数）
	stats.HitRate = 0.0 // TODO: 实现命中率统计

	return stats, nil
}

// CacheMiddleware 缓存中间件
type CacheMiddleware struct {
	cache  CacheService
	ttl    time.Duration
	logger *zap.Logger
}

// NewCacheMiddleware 创建缓存中间件
func NewCacheMiddleware(cache CacheService, ttl time.Duration) *CacheMiddleware {
	return &CacheMiddleware{
		cache:  cache,
		ttl:    ttl,
		logger: logger.GetLogger(),
	}
}

// Wrap 包装搜索函数
func (m *CacheMiddleware) Wrap(searchFunc func(ctx context.Context, query *SearchQuery) ([]*SearchResult, error)) func(ctx context.Context, query *SearchQuery) ([]*SearchResult, error) {
	return func(ctx context.Context, query *SearchQuery) ([]*SearchResult, error) {
		// 生成缓存键
		cacheKey := m.cache.GenerateKey(query)

		// 尝试从缓存获取
		results, err := m.cache.Get(ctx, cacheKey)
		if err != nil {
			m.logger.Warn("获取缓存失败，继续执行搜索", zap.Error(err))
		} else if results != nil {
			m.logger.Info("使用缓存结果", zap.String("key", cacheKey))
			return results, nil
		}

		// 执行实际搜索
		results, err = searchFunc(ctx, query)
		if err != nil {
			return nil, err
		}

		// 保存到缓存
		if err := m.cache.Set(ctx, cacheKey, results, m.ttl); err != nil {
			m.logger.Warn("保存缓存失败", zap.Error(err))
		}

		return results, nil
	}
}
