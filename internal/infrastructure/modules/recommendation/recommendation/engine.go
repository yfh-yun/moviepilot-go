package recommendation

import (
	"context"
	"math"

	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// Engine 推荐引擎接口
type Engine interface {
	// GetRecommendations 获取推荐列表
	GetRecommendations(ctx context.Context, userID int, limit int) ([]Recommendation, error)

	// GetSimilarItems 获取相似项目
	GetSimilarItems(ctx context.Context, itemID int, limit int) ([]Recommendation, error)

	// RecordFeedback 记录反馈
	RecordFeedback(ctx context.Context, userID int, itemID int, rating float64) error
}

// Recommendation 推荐结果
type Recommendation struct {
	ItemID int     `json:"item_id"`
	Score  float64 `json:"score"`
	Reason string  `json:"reason"`
}

// engine 推荐引擎实现
type engine struct {
	logger *zap.Logger
}

// NewEngine 创建推荐引擎
func NewEngine() Engine {
	return &engine{
		logger: logger.GetLogger(),
	}
}

// GetRecommendations 获取推荐列表
func (e *engine) GetRecommendations(ctx context.Context, userID int, limit int) ([]Recommendation, error) {
	e.logger.Info("获取推荐列表", zap.Int("userID", userID), zap.Int("limit", limit))

	// 简化实现：返回示例推荐
	recommendations := []Recommendation{
		{ItemID: 1, Score: 0.95, Reason: "基于您的观看历史"},
		{ItemID: 2, Score: 0.88, Reason: "相似用户喜欢"},
		{ItemID: 3, Score: 0.82, Reason: "热门推荐"},
	}

	if len(recommendations) > limit {
		recommendations = recommendations[:limit]
	}

	return recommendations, nil
}

// GetSimilarItems 获取相似项目
func (e *engine) GetSimilarItems(ctx context.Context, itemID int, limit int) ([]Recommendation, error) {
	e.logger.Info("获取相似项目", zap.Int("itemID", itemID), zap.Int("limit", limit))

	// 简化实现：返回示例相似项目
	similar := []Recommendation{
		{ItemID: 10, Score: 0.92, Reason: "相同类型"},
		{ItemID: 11, Score: 0.85, Reason: "相同导演"},
		{ItemID: 12, Score: 0.78, Reason: "相同演员"},
	}

	if len(similar) > limit {
		similar = similar[:limit]
	}

	return similar, nil
}

// RecordFeedback 记录反馈
func (e *engine) RecordFeedback(ctx context.Context, userID int, itemID int, rating float64) error {
	e.logger.Info("记录推荐反馈",
		zap.Int("userID", userID),
		zap.Int("itemID", itemID),
		zap.Float64("rating", rating))

	// 简化实现：仅记录日志
	return nil
}

// calculateCosineSimilarity 计算余弦相似度
func calculateCosineSimilarity(vec1, vec2 []float64) float64 {
	if len(vec1) != len(vec2) {
		return 0
	}

	var dotProduct, norm1, norm2 float64
	for i := range vec1 {
		dotProduct += vec1[i] * vec2[i]
		norm1 += vec1[i] * vec1[i]
		norm2 += vec2[i] * vec2[i]
	}

	if norm1 == 0 || norm2 == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))
}
