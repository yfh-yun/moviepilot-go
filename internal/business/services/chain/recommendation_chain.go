package chain

import (
	"context"
	"fmt"
	"time"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/models"
	"github.com/yfh-yun/moviepilot-go/internal/repositories"

	"go.uber.org/zap"
)

// RecommendationChain 推荐处理链
// 负责基于用户行为、偏好、热门度等生成个性化推荐

type RecommendationChain struct {
	userRepo     repository.UserRepository
	mediaRepo    repository.MediaRepository
	behaviorRepo repository.BehaviorRepository
	logger       *zap.Logger
}

// NewRecommendationChain 创建推荐处理链实例
func NewRecommendationChain(
	userRepo repository.UserRepository,
	mediaRepo repository.MediaRepository,
	behaviorRepo repository.BehaviorRepository,
) *RecommendationChain {
	return &RecommendationChain{
		userRepo:     userRepo,
		mediaRepo:    mediaRepo,
		behaviorRepo: behaviorRepo,
		logger:       logger.GetLogger().With(zap.String("module", "chain.recommendation")),
	}
}

// Execute 执行推荐处理链
func (rc *RecommendationChain) Execute(ctx context.Context, req *RecommendationRequest) (*RecommendationResponse, error) {
	rc.logger.Info("开始执行推荐处理链", zap.String("user_id", req.UserID))

	// 验证请求参数
	if err := rc.validateRequest(req); err != nil {
		return nil, fmt.Errorf("推荐请求验证失败: %w", err)
	}

	// 1. 获取用户偏好信息
	preferences, err := rc.getUserPreferences(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("获取用户偏好失败: %w", err)
	}

	// 2. 获取用户历史行为
	behaviors, err := rc.getUserBehavior(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("获取用户行为失败: %w", err)
	}

	// 3. 获取热门内容
	trending, err := rc.getTrendingContent(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("获取热门内容失败: %w", err)
	}

	// 4. 执行推荐算法
	recommendations, err := rc.calculateRecommendations(ctx, preferences, behaviors, trending)
	if err != nil {
		return nil, fmt.Errorf("推荐计算失败: %w", err)
	}

	// 5. 过滤和排序
	filteredRecommendations := rc.filterAndSort(recommendations, req)

	response := &RecommendationResponse{
		UserID:          req.UserID,
		Recommendations: filteredRecommendations,
		GeneratedAt:     time.Now(),
		AlgorithmUsed:   "hybrid-recommendation",
	}

	rc.logger.Info("推荐处理链执行完成",
		zap.String("user_id", req.UserID),
		zap.Int("recommendation_count", len(filteredRecommendations)))

	return response, nil
}

// getUserPreferences 获取用户偏好信息
func (rc *RecommendationChain) getUserPreferences(ctx context.Context, userID string) (*UserPreferences, error) {
	// 获取用户配置的偏好设置
	user, err := rc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	preferences := &UserPreferences{
		UserID:         userID,
		FavoriteGenres: rc.extractFavoriteGenres(user),
		FavoriteActors: rc.extractFavoriteActors(user),
		PreferredTypes: rc.extractPreferredTypes(user),
		QualityLevel:   rc.determineQualityPreference(user),
		UpdateAt:       time.Now(),
	}

	return preferences, nil
}

// getUserBehavior 获取用户行为数据
func (rc *RecommendationChain) getUserBehavior(ctx context.Context, userID string) (*UserBehavior, error) {
	recentBehaviors, err := rc.behaviorRepo.GetRecentBehavior(ctx, userID, 30) // 最近30天
	if err != nil {
		return nil, err
	}

	behavior := &UserBehavior{
		UserID:        userID,
		WatchHistory:  rc.extractWatchHistory(recentBehaviors),
		SearchHistory: rc.extractSearchHistory(recentBehaviors),
		RatingHistory: rc.extractRatingHistory(recentBehaviors),
		Interactions:  rc.countInteractions(recentBehaviors),
	}

	return behavior, nil
}

// getTrendingContent 获取热门内容
func (rc *RecommendationChain) getTrendingContent(ctx context.Context, req *RecommendationRequest) (*TrendingContent, error) {
	trending, err := rc.mediaRepo.GetTrendingMedia(ctx, &repository.TrendingQuery{
		Limit:     50,
		TimeRange: "7d", // 7天内
		MediaType: req.MediaType,
	})
	if err != nil {
		return nil, err
	}

	return &TrendingContent{
		MediaList:  trending,
		UpdatedAt:  time.Now(),
		TrendScore: rc.calculateTrendScore(trending),
	}, nil
}

// calculateRecommendations 计算推荐内容
func (rc *RecommendationChain) calculateRecommendations(
	ctx context.Context,
	preferences *UserPreferences,
	behavior *UserBehavior,
	trending *TrendingContent,
) ([]*Recommendation, error) {
	var recommendations []*Recommendation

	// 协同过滤推荐
	collaborativeRecs, err := rc.collaborativeFiltering(ctx, preferences, behavior)
	if err != nil {
		rc.logger.Warn("协同过滤推荐失败", zap.Error(err))
	}
	recommendations = append(recommendations, collaborativeRecs...)

	// 内容基于推荐
	contentRecs, err := rc.contentBasedFiltering(ctx, preferences, behavior)
	if err != nil {
		rc.logger.Warn("内容基于推荐失败", zap.Error(err))
	}
	recommendations = append(recommendations, contentRecs...)

	// 热门推荐
	trendingRecs := rc.trendingBasedRecommendation(trending, preferences)
	recommendations = append(recommendations, trendingRecs...)

	// 混合推荐策略
	hybridRecs := rc.hybridRecommendation(recommendations, behavior, trending)

	return hybridRecs, nil
}

// collaborativeFiltering 协同过滤推荐
func (rc *RecommendationChain) collaborativeFiltering(
	ctx context.Context,
	preferences *UserPreferences,
	behavior *UserBehavior,
) ([]*Recommendation, error) {
	// 实现基于用户的协同过滤算法
	similarUsers, err := rc.findSimilarUsers(ctx, preferences, behavior)
	if err != nil {
		return nil, err
	}

	return rc.generateCollaborativeRecommendations(ctx, similarUsers, preferences), nil
}

// contentBasedFiltering 内容基于推荐
func (rc *RecommendationChain) contentBasedFiltering(
	ctx context.Context,
	preferences *UserPreferences,
	behavior *UserBehavior,
) ([]*Recommendation, error) {
	// 基于内容特征的推荐算法
	return rc.generateContentBasedRecommendations(ctx, preferences, behavior), nil
}

// trendingBasedRecommendation 热门推荐
func (rc *RecommendationChain) trendingBasedRecommendation(
	trending *TrendingContent,
	preferences *UserPreferences,
) []*Recommendation {
	var recommendations []*Recommendation

	for _, media := range trending.MediaList {
		if rc.isMediaRelevant(media, preferences) {
			recommendation := &Recommendation{
				MediaID:   media.ID,
				MediaType: media.Type,
				Title:     media.Title,
				Score:     rc.calculateTrendingScore(media, trending),
				Reason:    "热门推荐",
				Algorithm: "trending",
				CreatedAt: time.Now(),
			}
			recommendations = append(recommendations, recommendation)
		}
	}

	return recommendations
}

// hybridRecommendation 混合推荐策略
func (rc *RecommendationChain) hybridRecommendation(
	recommendations []*Recommendation,
	behavior *UserBehavior,
	trending *TrendingContent,
) []*Recommendation {
	// 去重处理
	deduplicated := rc.deduplicateRecommendations(recommendations)

	// 多样性优化
	diverseRecs := rc.ensureDiversity(deduplicated, behavior)

	// 新鲜度优化
	freshRecs := rc.ensureFreshness(diverseRecs, trending)

	// 个性化排序
	sortedRecs := rc.personalizedSorting(freshRecs, behavior)

	return sortedRecs
}

// filterAndSort 过滤和排序推荐结果
func (rc *RecommendationChain) filterAndSort(
	recommendations []*Recommendation,
	req *RecommendationRequest,
) []*Recommendation {
	// 过滤已观看内容
	filtered := rc.filterWatchedContent(recommendations, req.UserID)

	// 应用用户过滤器
	filtered = rc.applyUserFilters(filtered, req.Filters)

	// 按分数排序
	sorted := rc.sortByScore(filtered)

	// 限制返回数量
	if len(sorted) > req.MaxResults {
		sorted = sorted[:req.MaxResults]
	}

	return sorted
}

// validateRequest 验证推荐请求
func (rc *RecommendationChain) validateRequest(req *RecommendationRequest) error {
	if req.UserID == "" {
		return fmt.Errorf("用户ID不能为空")
	}

	if req.MaxResults <= 0 || req.MaxResults > 100 {
		req.MaxResults = 20 // 默认值
	}

	return nil
}

// 辅助方法实现...

// RecommendationRequest 推荐请求
type RecommendationRequest struct {
	UserID     string                 `json:"user_id" validate:"required"`
	MediaType  string                 `json:"media_type"` // movie, tv, anime, etc.
	MaxResults int                    `json:"max_results"`
	Filters    map[string]interface{} `json:"filters"`
	Context    map[string]interface{} `json:"context"`
}

// RecommendationResponse 推荐响应
type RecommendationResponse struct {
	UserID          string            `json:"user_id"`
	Recommendations []*Recommendation `json:"recommendations"`
	GeneratedAt     time.Time         `json:"generated_at"`
	AlgorithmUsed   string            `json:"algorithm_used"`
	TotalCount      int               `json:"total_count"`
}

// Recommendation 推荐项
type Recommendation struct {
	MediaID   string    `json:"media_id"`
	MediaType string    `json:"media_type"`
	Title     string    `json:"title"`
	Score     float64   `json:"score"`     // 推荐分数
	Reason    string    `json:"reason"`    // 推荐理由
	Algorithm string    `json:"algorithm"` // 使用的算法
	CreatedAt time.Time `json:"created_at"`
}

// UserPreferences 用户偏好
type UserPreferences struct {
	UserID         string            `json:"user_id"`
	FavoriteGenres []string          `json:"favorite_genres"`
	FavoriteActors []string          `json:"favorite_actors"`
	PreferredTypes []string          `json:"preferred_types"`
	QualityLevel   string            `json:"quality_level"` // high, medium, low
	UpdateAt       time.Time         `json:"update_at"`
	CustomRules    map[string]string `json:"custom_rules"`
}

// UserBehavior 用户行为
type UserBehavior struct {
	UserID        string                     `json:"user_id"`
	WatchHistory  []*model.WatchHistoryItem  `json:"watch_history"`
	SearchHistory []*model.SearchHistoryItem `json:"search_history"`
	RatingHistory []*model.RatingHistoryItem `json:"rating_history"`
	Interactions  int                        `json:"interactions"`
}

// TrendingContent 热门内容
type TrendingContent struct {
	MediaList  []*model.Media `json:"media_list"`
	UpdatedAt  time.Time      `json:"updated_at"`
	TrendScore float64        `json:"trend_score"`
}
