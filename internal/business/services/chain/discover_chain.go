package chain

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/yfh-yun/moviepilot-go/pkg/cache"
	"github.com/yfh-yun/moviepilot-go/internal/models"
)

// DiscoverChain 发现处理链
// 负责发现热门内容、最新内容、推荐内容等发现功能

type DiscoveryChain struct {
	logger           *zap.Logger
	mediaRepository  MediaRepository
	userBehaviorRepo UserBehaviorRepository
	contentIndexer   ContentIndexer
	recommender      Recommender
	cache            *cache.RedisCache
}

// NewDiscoveryChain 创建发现处理链实例
func NewDiscoveryChain(
	logger *zap.Logger,
	mediaRepo MediaRepository,
	userBehaviorRepo UserBehaviorRepository,
	contentIndexer ContentIndexer,
	recommender Recommender,
	cache *cache.RedisCache,
) *DiscoveryChain {
	return &DiscoveryChain{
		logger:           logger.Named("discovery_chain"),
		mediaRepository:  mediaRepo,
		userBehaviorRepo: userBehaviorRepo,
		contentIndexer:   contentIndexer,
		recommender:      recommender,
		cache:            cache,
	}
}

// DiscoverPopular 发现热门内容
func (d *DiscoveryChain) DiscoverPopular(ctx context.Context, userID string, limit int) ([]*MediaItem, error) {
	d.logger.Info("开始发现热门内容",
		zap.String("user_id", userID),
		zap.Int("limit", limit))

	// 检查缓存
	cacheKey := fmt.Sprintf("discover:popular:%s:%d", userID, limit)
	if cached, err := d.getCachedResults(ctx, cacheKey); err == nil && cached != nil {
		d.logger.Debug("从缓存获取热门内容", zap.Int("count", len(cached)))
		return cached, nil
	}

	// 获取热门媒体内容
	popularContents, err := d.contentIndexer.GetPopularContents(ctx, limit)
	if err != nil {
		d.logger.Error("获取热门内容失败", zap.Error(err))
		return nil, fmt.Errorf("获取热门内容失败: %w", err)
	}

	// 基于用户行为过滤和排序
	filteredContents, err := d.filterByUserBehavior(ctx, userID, popularContents)
	if err != nil {
		d.logger.Warn("用户行为过滤失败，返回原始内容", zap.Error(err))
		filteredContents = popularContents
	}

	// 缓存结果
	if err := d.cacheResults(ctx, cacheKey, filteredContents, 30*time.Minute); err != nil {
		d.logger.Warn("缓存热门内容失败", zap.Error(err))
	}

	d.logger.Info("热门内容发现完成",
		zap.Int("original_count", len(popularContents)),
		zap.Int("filtered_count", len(filteredContents)))

	return filteredContents, nil
}

// DiscoverLatest 发现最新内容
func (d *DiscoveryChain) DiscoverLatest(ctx context.Context, userID string, limit int, contentType string) ([]*MediaItem, error) {
	d.logger.Info("开始发现最新内容",
		zap.String("user_id", userID),
		zap.Int("limit", limit),
		zap.String("content_type", contentType))

	// 检查缓存
	cacheKey := fmt.Sprintf("discover:latest:%s:%s:%d", userID, contentType, limit)
	if cached, err := d.getCachedResults(ctx, cacheKey); err == nil && cached != nil {
		d.logger.Debug("从缓存获取最新内容", zap.Int("count", len(cached)))
		return cached, nil
	}

	// 获取最新媒体内容
	latestContents, err := d.contentIndexer.GetLatestContents(ctx, limit, contentType)
	if err != nil {
		d.logger.Error("获取最新内容失败", zap.Error(err))
		return nil, fmt.Errorf("获取最新内容失败: %w", err)
	}

	// 基于用户偏好过滤
	filteredContents, err := d.filterByUserPreference(ctx, userID, latestContents)
	if err != nil {
		d.logger.Warn("用户偏好过滤失败，返回原始内容", zap.Error(err))
		filteredContents = latestContents
	}

	// 缓存结果
	if err := d.cacheResults(ctx, cacheKey, filteredContents, 15*time.Minute); err != nil {
		d.logger.Warn("缓存最新内容失败", zap.Error(err))
	}

	d.logger.Info("最新内容发现完成",
		zap.Int("original_count", len(latestContents)),
		zap.Int("filtered_count", len(filteredContents)))

	return filteredContents, nil
}

// DiscoverRecommended 发现推荐内容
func (d *DiscoveryChain) DiscoverRecommended(ctx context.Context, userID string, limit int) ([]*MediaItem, error) {
	d.logger.Info("开始发现推荐内容",
		zap.String("user_id", userID),
		zap.Int("limit", limit))

	// 检查缓存
	cacheKey := fmt.Sprintf("discover:recommended:%s:%d", userID, limit)
	if cached, err := d.getCachedResults(ctx, cacheKey); err == nil && cached != nil {
		d.logger.Debug("从缓存获取推荐内容", zap.Int("count", len(cached)))
		return cached, nil
	}

	// 获取推荐内容
	recommendedContents, err := d.recommender.GetRecommendations(ctx, userID, limit)
	if err != nil {
		d.logger.Error("获取推荐内容失败", zap.Error(err))
		return nil, fmt.Errorf("获取推荐内容失败: %w", err)
	}

	// 过滤已观看内容
	filteredContents, err := d.filterWatchedContents(ctx, userID, recommendedContents)
	if err != nil {
		d.logger.Warn("过滤已观看内容失败", zap.Error(err))
		filteredContents = recommendedContents
	}

	// 缓存结果
	if err := d.cacheResults(ctx, cacheKey, filteredContents, 1*time.Hour); err != nil {
		d.logger.Warn("缓存推荐内容失败", zap.Error(err))
	}

	d.logger.Info("推荐内容发现完成",
		zap.Int("original_count", len(recommendedContents)),
		zap.Int("filtered_count", len(filteredContents)))

	return filteredContents, nil
}

// DiscoverPersonalized 个性化发现
func (d *DiscoveryChain) DiscoverPersonalized(ctx context.Context, userID string, limit int) ([]*MediaItem, error) {
	d.logger.Info("开始个性化发现",
		zap.String("user_id", userID),
		zap.Int("limit", limit))

	// 检查缓存
	cacheKey := fmt.Sprintf("discover:personalized:%s:%d", userID, limit)
	if cached, err := d.getCachedResults(ctx, cacheKey); err == nil && cached != nil {
		d.logger.Debug("从缓存获取个性化内容", zap.Int("count", len(cached)))
		return cached, nil
	}

	// 获取用户行为数据
	userBehavior, err := d.userBehaviorRepo.GetUserBehavior(ctx, userID)
	if err != nil {
		d.logger.Error("获取用户行为数据失败", zap.Error(err))
		return nil, fmt.Errorf("获取用户行为数据失败: %w", err)
	}

	// 基于用户行为生成个性化推荐
	personalizedContents, err := d.generatePersonalizedContent(ctx, userBehavior, limit)
	if err != nil {
		d.logger.Error("生成个性化内容失败", zap.Error(err))
		return nil, fmt.Errorf("生成个性化内容失败: %w", err)
	}

	// 缓存结果
	if err := d.cacheResults(ctx, cacheKey, personalizedContents, 30*time.Minute); err != nil {
		d.logger.Warn("缓存个性化内容失败", zap.Error(err))
	}

	d.logger.Info("个性化发现完成",
		zap.Int("generated_count", len(personalizedContents)))

	return personalizedContents, nil
}

// DiscoverSimilar 发现相似内容
func (d *DiscoveryChain) DiscoverSimilar(ctx context.Context, mediaID string, limit int) ([]*MediaItem, error) {
	d.logger.Info("开始发现相似内容",
		zap.String("media_id", mediaID),
		zap.Int("limit", limit))

	// 检查缓存
	cacheKey := fmt.Sprintf("discover:similar:%s:%d", mediaID, limit)
	if cached, err := d.getCachedResults(ctx, cacheKey); err == nil && cached != nil {
		d.logger.Debug("从缓存获取相似内容", zap.Int("count", len(cached)))
		return cached, nil
	}

	// 获取目标媒体信息
	targetMedia, err := d.mediaRepository.GetMediaByID(ctx, mediaID)
	if err != nil {
		d.logger.Error("获取目标媒体信息失败", zap.Error(err))
		return nil, fmt.Errorf("获取目标媒体信息失败: %w", err)
	}

	// 获取相似媒体
	similarContents, err := d.contentIndexer.GetSimilarContents(ctx, targetMedia, limit)
	if err != nil {
		d.logger.Error("获取相似内容失败", zap.Error(err))
		return nil, fmt.Errorf("获取相似内容失败: %w", err)
	}

	// 缓存结果
	if err := d.cacheResults(ctx, cacheKey, similarContents, 1*time.Hour); err != nil {
		d.logger.Warn("缓存相似内容失败", zap.Error(err))
	}

	d.logger.Info("相似内容发现完成",
		zap.Int("similar_count", len(similarContents)))

	return similarContents, nil
}

// DiscoverByCategory 按分类发现
func (d *DiscoveryChain) DiscoverByCategory(ctx context.Context, category string, userID string, limit int) ([]*MediaItem, error) {
	d.logger.Info("按分类发现内容",
		zap.String("category", category),
		zap.String("user_id", userID),
		zap.Int("limit", limit))

	// 检查缓存
	cacheKey := fmt.Sprintf("discover:category:%s:%s:%d", category, userID, limit)
	if cached, err := d.getCachedResults(ctx, cacheKey); err == nil && cached != nil {
		d.logger.Debug("从缓存获取分类内容", zap.Int("count", len(cached)))
		return cached, nil
	}

	// 获取分类内容
	categoryContents, err := d.contentIndexer.GetContentsByCategory(ctx, category, limit)
	if err != nil {
		d.logger.Error("获取分类内容失败", zap.Error(err))
		return nil, fmt.Errorf("获取分类内容失败: %w", err)
	}

	// 基于用户偏好过滤
	filteredContents, err := d.filterByUserPreference(ctx, userID, categoryContents)
	if err != nil {
		d.logger.Warn("用户偏好过滤失败，返回原始内容", zap.Error(err))
		filteredContents = categoryContents
	}

	// 缓存结果
	if err := d.cacheResults(ctx, cacheKey, filteredContents, 1*time.Hour); err != nil {
		d.logger.Warn("缓存分类内容失败", zap.Error(err))
	}

	d.logger.Info("分类内容发现完成",
		zap.Int("original_count", len(categoryContents)),
		zap.Int("filtered_count", len(filteredContents)))

	return filteredContents, nil
}

// 辅助方法

// filterByUserBehavior 基于用户行为过滤内容
func (d *DiscoveryChain) filterByUserBehavior(ctx context.Context, userID string, contents []*MediaItem) ([]*MediaItem, error) {
	userBehavior, err := d.userBehaviorRepo.GetUserBehavior(ctx, userID)
	if err != nil {
		return nil, err
	}

	filtered := make([]*MediaItem, 0)
	for _, content := range contents {
		// 排除已观看内容
		if d.isContentWatched(userBehavior, content.ID) {
			continue
		}

		// 基于用户偏好评分
		score := d.calculateContentScore(userBehavior, content)
		if score > 0.1 { // 阈值过滤
			filtered = append(filtered, content)
		}
	}

	return filtered, nil
}

// filterByUserPreference 基于用户偏好过滤内容
func (d *DiscoveryChain) filterByUserPreference(ctx context.Context, userID string, contents []*MediaItem) ([]*MediaItem, error) {
	userBehavior, err := d.userBehaviorRepo.GetUserBehavior(ctx, userID)
	if err != nil {
		return nil, err
	}

	filtered := make([]*MediaItem, 0)
	for _, content := range contents {
		// 基于用户偏好评分
		score := d.calculatePreferenceScore(userBehavior, content)
		if score > 0.3 { // 偏好阈值
			filtered = append(filtered, content)
		}
	}

	return filtered, nil
}

// filterWatchedContents 过滤已观看内容
func (d *DiscoveryChain) filterWatchedContents(ctx context.Context, userID string, contents []*MediaItem) ([]*MediaItem, error) {
	userBehavior, err := d.userBehaviorRepo.GetUserBehavior(ctx, userID)
	if err != nil {
		return nil, err
	}

	filtered := make([]*MediaItem, 0)
	for _, content := range contents {
		if !d.isContentWatched(userBehavior, content.ID) {
			filtered = append(filtered, content)
		}
	}

	return filtered, nil
}

// generatePersonalizedContent 生成个性化内容
func (d *DiscoveryChain) generatePersonalizedContent(ctx context.Context, userBehavior *models.UserBehavior, limit int) ([]*models.MediaItem, error) {
	// 基于用户行为生成个性化推荐
	// 实现协同过滤、内容推荐等算法
	return d.recommender.GetPersonalizedRecommendations(ctx, userBehavior, limit)
}

// 缓存相关方法

func (d *DiscoveryChain) getCachedResults(ctx context.Context, key string) ([]*models.MediaItem, error) {
	return d.cache.GetCachedResults(ctx, key)
}

func (d *DiscoveryChain) cacheResults(ctx context.Context, key string, results []*models.MediaItem, ttl time.Duration) error {
	return d.cache.CacheResults(ctx, key, results, ttl)
}

// 评分计算相关方法

func (d *DiscoveryChain) calculateContentScore(userBehavior *models.UserBehavior, content *models.MediaItem) float64 {
	// 基于用户行为计算内容评分
	// 实现评分算法
	return 0.5
}

func (d *DiscoveryChain) calculatePreferenceScore(userBehavior *models.UserBehavior, content *models.MediaItem) float64 {
	// 基于用户偏好计算内容评分
	// 实现偏好评分算法
	return 0.5
}

func (d *DiscoveryChain) isContentWatched(userBehavior *models.UserBehavior, contentID string) bool {
	// 检查内容是否已被用户观看
	return false
}

// 依赖接口定义

type MediaRepository interface {
	GetMediaByID(ctx context.Context, mediaID string) (*models.MediaItem, error)
}

type UserBehaviorRepository interface {
	GetUserBehavior(ctx context.Context, userID string) (*models.UserBehavior, error)
}

type ContentIndexer interface {
	GetPopularContents(ctx context.Context, limit int) ([]*models.MediaItem, error)
	GetLatestContents(ctx context.Context, limit int, contentType string) ([]*models.MediaItem, error)
	GetSimilarContents(ctx context.Context, target *models.MediaItem, limit int) ([]*models.MediaItem, error)
	GetContentsByCategory(ctx context.Context, category string, limit int) ([]*models.MediaItem, error)
}

type Recommender interface {
	GetRecommendations(ctx context.Context, userID string, limit int) ([]*models.MediaItem, error)
	GetPersonalizedRecommendations(ctx context.Context, userBehavior *models.UserBehavior, limit int) ([]*models.MediaItem, error)
}

// 注意: MediaItem 和 UserBehavior 已移至 models 包
