package search

import (
	"context"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"moviepilot-go/pkg/logger"
)

// HistoryService 搜索历史服务
type HistoryService interface {
	// RecordSearch 记录搜索
	RecordSearch(ctx context.Context, record *SearchRecord) error

	// GetSearchHistory 获取搜索历史
	GetSearchHistory(ctx context.Context, userID string, limit int) ([]*SearchRecord, error)

	// GetPopularSearches 获取热门搜索
	GetPopularSearches(ctx context.Context, limit int, days int) ([]*PopularSearch, error)

	// GetSearchStats 获取搜索统计
	GetSearchStats(ctx context.Context, userID string) (*SearchStats, error)

	// CleanOldRecords 清理旧记录
	CleanOldRecords(ctx context.Context, days int) error
}

// historyService 历史服务实现
type historyService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewHistoryService 创建历史服务
func NewHistoryService(db *gorm.DB) HistoryService {
	return &historyService{
		db:     db,
		logger: logger.GetLogger(),
	}
}

// SearchRecord 搜索记录
type SearchRecord struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      string    `gorm:"type:varchar(100);index" json:"user_id"`
	Keyword     string    `gorm:"type:varchar(500);not null;index" json:"keyword"`
	Type        string    `gorm:"type:varchar(20)" json:"type"` // movie, tv
	Quality     string    `gorm:"type:varchar(50)" json:"quality"`
	Resolution  string    `gorm:"type:varchar(50)" json:"resolution"`
	ResultCount int       `json:"result_count"`
	Duration    int64     `json:"duration"` // 搜索耗时（毫秒）
	Success     bool      `gorm:"not null;default:true" json:"success"`
	ErrorMsg    string    `gorm:"type:varchar(500)" json:"error_msg"`
	CreatedAt   time.Time `gorm:"autoCreateTime;index" json:"created_at"`
}

// TableName 表名
func (SearchRecord) TableName() string {
	return "search_history"
}

// PopularSearch 热门搜索
type PopularSearch struct {
	Keyword     string    `json:"keyword"`
	SearchCount int64     `json:"search_count"`
	LastSearch  time.Time `json:"last_search"`
}

// SearchStats 搜索统计
type SearchStats struct {
	UserID         string     `json:"user_id"`
	TotalSearches  int64      `json:"total_searches"`
	UniqueKeywords int64      `json:"unique_keywords"`
	AvgDuration    float64    `json:"avg_duration"`
	SuccessRate    float64    `json:"success_rate"`
	LastSearchTime *time.Time `json:"last_search_time"`
	TopKeywords    []string   `json:"top_keywords"`
}

// RecordSearch 记录搜索
func (s *historyService) RecordSearch(ctx context.Context, record *SearchRecord) error {
	s.logger.Info("记录搜索",
		zap.String("user_id", record.UserID),
		zap.String("keyword", record.Keyword),
		zap.Int("result_count", record.ResultCount),
	)

	if err := s.db.WithContext(ctx).Create(record).Error; err != nil {
		s.logger.Error("记录搜索失败", zap.Error(err))
		return err
	}

	return nil
}

// GetSearchHistory 获取搜索历史
func (s *historyService) GetSearchHistory(ctx context.Context, userID string, limit int) ([]*SearchRecord, error) {
	s.logger.Info("获取搜索历史",
		zap.String("user_id", userID),
		zap.Int("limit", limit),
	)

	var records []*SearchRecord

	query := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&records).Error; err != nil {
		s.logger.Error("获取搜索历史失败", zap.Error(err))
		return nil, err
	}

	return records, nil
}

// GetPopularSearches 获取热门搜索
func (s *historyService) GetPopularSearches(ctx context.Context, limit int, days int) ([]*PopularSearch, error) {
	s.logger.Info("获取热门搜索",
		zap.Int("limit", limit),
		zap.Int("days", days),
	)

	var results []*PopularSearch

	startDate := time.Now().AddDate(0, 0, -days)

	query := `
		SELECT 
			keyword,
			COUNT(*) as search_count,
			MAX(created_at) as last_search
		FROM search_history
		WHERE created_at >= ? AND success = true
		GROUP BY keyword
		ORDER BY search_count DESC
		LIMIT ?
	`

	if err := s.db.WithContext(ctx).Raw(query, startDate, limit).Scan(&results).Error; err != nil {
		s.logger.Error("获取热门搜索失败", zap.Error(err))
		return nil, err
	}

	return results, nil
}

// GetSearchStats 获取搜索统计
func (s *historyService) GetSearchStats(ctx context.Context, userID string) (*SearchStats, error) {
	s.logger.Info("获取搜索统计", zap.String("user_id", userID))

	stats := &SearchStats{
		UserID: userID,
	}

	// 总搜索数
	if err := s.db.WithContext(ctx).
		Model(&SearchRecord{}).
		Where("user_id = ?", userID).
		Count(&stats.TotalSearches).Error; err != nil {
		return nil, err
	}

	// 唯一关键词数
	if err := s.db.WithContext(ctx).
		Model(&SearchRecord{}).
		Where("user_id = ?", userID).
		Distinct("keyword").
		Count(&stats.UniqueKeywords).Error; err != nil {
		return nil, err
	}

	// 平均耗时
	var avgDuration struct {
		Avg float64
	}
	if err := s.db.WithContext(ctx).
		Model(&SearchRecord{}).
		Select("AVG(duration) as avg").
		Where("user_id = ? AND success = ?", userID, true).
		Scan(&avgDuration).Error; err == nil {
		stats.AvgDuration = avgDuration.Avg
	}

	// 成功率
	var successCount int64
	if err := s.db.WithContext(ctx).
		Model(&SearchRecord{}).
		Where("user_id = ? AND success = ?", userID, true).
		Count(&successCount).Error; err == nil {
		if stats.TotalSearches > 0 {
			stats.SuccessRate = float64(successCount) / float64(stats.TotalSearches) * 100
		}
	}

	// 最后搜索时间
	var lastRecord SearchRecord
	if err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		First(&lastRecord).Error; err == nil {
		stats.LastSearchTime = &lastRecord.CreatedAt
	}

	// 热门关键词（Top 5）
	var topKeywords []struct {
		Keyword string
		Count   int64
	}
	if err := s.db.WithContext(ctx).
		Model(&SearchRecord{}).
		Select("keyword, COUNT(*) as count").
		Where("user_id = ?", userID).
		Group("keyword").
		Order("count DESC").
		Limit(5).
		Scan(&topKeywords).Error; err == nil {
		stats.TopKeywords = make([]string, 0, len(topKeywords))
		for _, kw := range topKeywords {
			stats.TopKeywords = append(stats.TopKeywords, kw.Keyword)
		}
	}

	return stats, nil
}

// CleanOldRecords 清理旧记录
func (s *historyService) CleanOldRecords(ctx context.Context, days int) error {
	s.logger.Info("清理旧搜索记录", zap.Int("days", days))

	cutoffTime := time.Now().AddDate(0, 0, -days)

	result := s.db.WithContext(ctx).
		Where("created_at < ?", cutoffTime).
		Delete(&SearchRecord{})

	if result.Error != nil {
		s.logger.Error("清理旧记录失败", zap.Error(result.Error))
		return result.Error
	}

	s.logger.Info("清理完成", zap.Int64("deleted", result.RowsAffected))
	return nil
}
