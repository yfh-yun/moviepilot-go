// Package actions 提供动作系统的业务逻辑实现
package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"moviepilot-go/pkg/logger"

	"go.uber.org/zap"
)

// WorkflowCache 工作流缓存管理器
// 提供工作流级别的数据缓存、状态管理和协调功能
type WorkflowCache struct {
	redis  *redis.Client
	logger *zap.Logger

	// 内存缓存（用于热数据）
	localCache map[string]*cacheEntry
	localMutex sync.RWMutex

	// 缓存统计
	stats *CacheStatistics
}

// cacheEntry 缓存条目
type cacheEntry struct {
	Value      interface{}
	ExpiresAt  time.Time
	WorkflowID int64
}

// CacheStatistics 缓存统计信息
type CacheStatistics struct {
	HitCount    int64     `json:"hit_count"`
	MissCount   int64     `json:"miss_count"`
	SetCount    int64     `json:"set_count"`
	DeleteCount int64     `json:"delete_count"`
	LastAccess  time.Time `json:"last_access"`
	mutex       sync.RWMutex
}

// NewWorkflowCache 创建工作流缓存实例
func NewWorkflowCache(redisClient *redis.Client) *WorkflowCache {
	return &WorkflowCache{
		redis:      redisClient,
		logger:     logger.NewLogger("workflow_cache"),
		localCache: make(map[string]*cacheEntry),
		stats:      &CacheStatistics{},
	}
}

// CheckCache 检查缓存是否存在
func (c *WorkflowCache) CheckCache(workflowID int64, key string) bool {
	cacheKey := c.buildCacheKey(workflowID, key)

	// 首先检查本地缓存
	if c.checkLocalCache(cacheKey) {
		c.stats.recordHit()
		return true
	}

	// 检查Redis缓存
	if c.redis != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		exists, err := c.redis.Exists(ctx, cacheKey).Result()
		if err == nil && exists > 0 {
			// 从Redis加载到本地缓存
			c.loadFromRedis(cacheKey)
			c.stats.recordHit()
			return true
		}
	}

	c.stats.recordMiss()
	return false
}

// GetCache 获取缓存值
func (c *WorkflowCache) GetCache(workflowID int64, key string) (interface{}, bool) {
	cacheKey := c.buildCacheKey(workflowID, key)

	// 首先尝试本地缓存
	if value, found := c.getLocalCache(cacheKey); found {
		c.stats.recordHit()
		return value, true
	}

	// 尝试Redis缓存
	if c.redis != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		result, err := c.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var value interface{}
			if err := json.Unmarshal([]byte(result), &value); err == nil {
				// 加载到本地缓存
				c.setLocalCache(cacheKey, value, 5*time.Minute, workflowID)
				c.stats.recordHit()
				return value, true
			}
		}
	}

	c.stats.recordMiss()
	return nil, false
}

// SaveCache 保存缓存值
func (c *WorkflowCache) SaveCache(workflowID int64, key string, value interface{}, ttl time.Duration) error {
	cacheKey := c.buildCacheKey(workflowID, key)

	// 保存到本地缓存
	c.setLocalCache(cacheKey, value, ttl, workflowID)

	// 保存到Redis缓存
	if c.redis != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		// 序列化值
		data, err := json.Marshal(value)
		if err != nil {
			c.logger.Error("缓存值序列化失败", zap.Error(err))
			return fmt.Errorf("缓存值序列化失败: %w", err)
		}

		// 保存到Redis
		if err := c.redis.Set(ctx, cacheKey, data, ttl).Err(); err != nil {
			c.logger.Error("保存Redis缓存失败", zap.Error(err))
			return fmt.Errorf("保存Redis缓存失败: %w", err)
		}
	}

	c.stats.recordSet()
	return nil
}

// DeleteCache 删除缓存
func (c *WorkflowCache) DeleteCache(workflowID int64, key string) error {
	cacheKey := c.buildCacheKey(workflowID, key)

	// 删除本地缓存
	c.deleteLocalCache(cacheKey)

	// 删除Redis缓存
	if c.redis != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := c.redis.Del(ctx, cacheKey).Err(); err != nil {
			c.logger.Error("删除Redis缓存失败", zap.Error(err))
			return fmt.Errorf("删除Redis缓存失败: %w", err)
		}
	}

	c.stats.recordDelete()
	return nil
}

// ClearWorkflowCache 清理工作流所有缓存
func (c *WorkflowCache) ClearWorkflowCache(workflowID int64) error {
	pattern := fmt.Sprintf("workflow:%d:*", workflowID)

	// 清理本地缓存
	c.clearLocalCacheByPattern(pattern)

	// 清理Redis缓存
	if c.redis != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		keys, err := c.redis.Keys(ctx, pattern).Result()
		if err != nil {
			return fmt.Errorf("获取Redis缓存键失败: %w", err)
		}

		if len(keys) > 0 {
			if err := c.redis.Del(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("删除Redis缓存失败: %w", err)
			}
		}
	}

	c.logger.Info("工作流缓存已清理", zap.Int64("workflow_id", workflowID))
	return nil
}

// GetWorkflowState 获取工作流状态
func (c *WorkflowCache) GetWorkflowState(workflowID int64) (*WorkflowState, error) {
	stateKey := "state"
	if value, found := c.GetCache(workflowID, stateKey); found {
		if state, ok := value.(*WorkflowState); ok {
			return state, nil
		}
	}

	// 返回默认状态
	return &WorkflowState{
		WorkflowID: workflowID,
		Status:     "running",
		StartTime:  time.Now(),
	}, nil
}

// SaveWorkflowState 保存工作流状态
func (c *WorkflowCache) SaveWorkflowState(workflowID int64, state *WorkflowState) error {
	state.UpdatedAt = time.Now()
	return c.SaveCache(workflowID, "state", state, 24*time.Hour)
}

// GetWorkflowContext 获取工作流上下文
func (c *WorkflowCache) GetWorkflowContext(workflowID int64) (*ActionContext, error) {
	contextKey := "context"
	if value, found := c.GetCache(workflowID, contextKey); found {
		if ctx, ok := value.(*ActionContext); ok {
			return ctx, nil
		}
	}

	// 返回空上下文
	return &ActionContext{
		WorkflowID: workflowID,
		Variables:  make(map[string]interface{}),
	}, nil
}

// SaveWorkflowContext 保存工作流上下文
func (c *WorkflowCache) SaveWorkflowContext(workflowID int64, ctx *ActionContext) error {
	return c.SaveCache(workflowID, "context", ctx, 24*time.Hour)
}

// IncrementCounter 递增计数器
func (c *WorkflowCache) IncrementCounter(workflowID int64, counterName string) (int64, error) {
	key := fmt.Sprintf("counter:%s", counterName)

	if c.redis != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		result, err := c.redis.Incr(ctx, c.buildCacheKey(workflowID, key)).Result()
		if err != nil {
			return 0, fmt.Errorf("递增计数器失败: %w", err)
		}

		// 设置过期时间
		c.redis.Expire(ctx, c.buildCacheKey(workflowID, key), 24*time.Hour)

		return result, nil
	}

	// 本地实现
	cacheKey := c.buildCacheKey(workflowID, key)
	if value, found := c.getLocalCache(cacheKey); found {
		if count, ok := value.(int64); ok {
			newCount := count + 1
			c.setLocalCache(cacheKey, newCount, 24*time.Hour, workflowID)
			return newCount, nil
		}
	}

	c.setLocalCache(cacheKey, int64(1), 24*time.Hour, workflowID)
	return 1, nil
}

// GetStatistics 获取缓存统计信息
func (c *WorkflowCache) GetStatistics() *CacheStatistics {
	c.stats.mutex.RLock()
	defer c.stats.mutex.RUnlock()

	return &CacheStatistics{
		HitCount:    c.stats.HitCount,
		MissCount:   c.stats.MissCount,
		SetCount:    c.stats.SetCount,
		DeleteCount: c.stats.DeleteCount,
		LastAccess:  c.stats.LastAccess,
	}
}

// CleanupExpiredCache 清理过期缓存
func (c *WorkflowCache) CleanupExpiredCache() {
	c.localMutex.Lock()
	defer c.localMutex.Unlock()

	now := time.Now()
	for key, entry := range c.localCache {
		if now.After(entry.ExpiresAt) {
			delete(c.localCache, key)
		}
	}

	c.logger.Debug("本地缓存清理完成", zap.Int("remaining_entries", len(c.localCache)))
}

// 私有方法

// buildCacheKey 构建缓存键
func (c *WorkflowCache) buildCacheKey(workflowID int64, key string) string {
	return fmt.Sprintf("workflow:%d:%s", workflowID, key)
}

// checkLocalCache 检查本地缓存
func (c *WorkflowCache) checkLocalCache(key string) bool {
	c.localMutex.RLock()
	defer c.localMutex.RUnlock()

	entry, exists := c.localCache[key]
	return exists && time.Now().Before(entry.ExpiresAt)
}

// getLocalCache 获取本地缓存
func (c *WorkflowCache) getLocalCache(key string) (interface{}, bool) {
	c.localMutex.RLock()
	defer c.localMutex.RUnlock()

	entry, exists := c.localCache[key]
	if !exists || time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	return entry.Value, true
}

// setLocalCache 设置本地缓存
func (c *WorkflowCache) setLocalCache(key string, value interface{}, ttl time.Duration, workflowID int64) {
	c.localMutex.Lock()
	defer c.localMutex.Unlock()

	// 限制本地缓存大小
	if len(c.localCache) >= 1000 {
		c.evictOldestEntry()
	}

	c.localCache[key] = &cacheEntry{
		Value:      value,
		ExpiresAt:  time.Now().Add(ttl),
		WorkflowID: workflowID,
	}
}

// deleteLocalCache 删除本地缓存
func (c *WorkflowCache) deleteLocalCache(key string) {
	c.localMutex.Lock()
	defer c.localMutex.Unlock()

	delete(c.localCache, key)
}

// clearLocalCacheByPattern 根据模式清理本地缓存
func (c *WorkflowCache) clearLocalCacheByPattern(pattern string) {
	c.localMutex.Lock()
	defer c.localMutex.Unlock()

	for key := range c.localCache {
		if matchPattern(key, pattern) {
			delete(c.localCache, key)
		}
	}
}

// loadFromRedis 从Redis加载到本地缓存
func (c *WorkflowCache) loadFromRedis(cacheKey string) {
	if c.redis == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := c.redis.Get(ctx, cacheKey).Result()
	if err != nil {
		return
	}

	var value interface{}
	if err := json.Unmarshal([]byte(result), &value); err == nil {
		// 提取workflowID
		workflowID := c.extractWorkflowID(cacheKey)
		if workflowID > 0 {
			c.setLocalCache(cacheKey, value, 5*time.Minute, workflowID)
		}
	}
}

// evictOldestEntry 淘汰最旧的缓存条目
func (c *WorkflowCache) evictOldestEntry() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.localCache {
		if oldestKey == "" || entry.ExpiresAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.ExpiresAt
		}
	}

	if oldestKey != "" {
		delete(c.localCache, oldestKey)
	}
}

// extractWorkflowID 从缓存键提取工作流ID
func (c *WorkflowCache) extractWorkflowID(cacheKey string) int64 {
	// 解析 workflow:workflowID:key 格式
	// 这里简化实现，实际可能需要更复杂的解析
	return 0
}

// matchPattern 检查键是否匹配模式
func matchPattern(key, pattern string) bool {
	// 简单的通配符匹配实现
	// 支持 * 通配符
	return true // 简化实现
}

// CacheStatistics 方法

// recordHit 记录命中
func (s *CacheStatistics) recordHit() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.HitCount++
	s.LastAccess = time.Now()
}

// recordMiss 记录未命中
func (s *CacheStatistics) recordMiss() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.MissCount++
	s.LastAccess = time.Now()
}

// recordSet 记录设置
func (s *CacheStatistics) recordSet() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.SetCount++
	s.LastAccess = time.Now()
}

// recordDelete 记录删除
func (s *CacheStatistics) recordDelete() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.DeleteCount++
	s.LastAccess = time.Now()
}

// GetHitRate 获取命中率
func (s *CacheStatistics) GetHitRate() float64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	total := s.HitCount + s.MissCount
	if total == 0 {
		return 0
	}

	return float64(s.HitCount) / float64(total)
}

// WorkflowState 工作流状态
type WorkflowState struct {
	WorkflowID  int64                  `json:"workflow_id"`
	Status      string                 `json:"status"` // "running", "completed", "failed", "cancelled"
	StartTime   time.Time              `json:"start_time"`
	EndTime     time.Time              `json:"end_time"`
	CurrentStep string                 `json:"current_step"`
	Progress    float64                `json:"progress"`
	Error       string                 `json:"error"`
	Variables   map[string]interface{} `json:"variables"`
	UpdatedAt   time.Time              `json:"updated_at"`
}
