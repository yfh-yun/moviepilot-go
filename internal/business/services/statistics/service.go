package statistics

import (
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"moviepilot-go/pkg/logger"
)

// Service 统计服务接口
type Service interface {
	// GetOverallStats 获取总体统计
	GetOverallStats(ctx context.Context) (*OverallStats, error)

	// GetSubscribeStats 获取订阅统计
	GetSubscribeStats(ctx context.Context) (*SubscribeStats, error)

	// GetDownloadStats 获取下载统计
	GetDownloadStats(ctx context.Context) (*DownloadStats, error)

	// GetSearchStats 获取搜索统计
	GetSearchStats(ctx context.Context) (*SearchStats, error)

	// GetTrendData 获取趋势数据
	GetTrendData(ctx context.Context, days int) (*TrendData, error)
}

// service 统计服务实现
type service struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewService 创建统计服务
func NewService(db *gorm.DB) Service {
	return &service{
		db:     db,
		logger: logger.GetLogger(),
	}
}

// OverallStats 总体统计
type OverallStats struct {
	TotalSubscribes  int64 `json:"total_subscribes"`
	ActiveSubscribes int64 `json:"active_subscribes"`
	TotalDownloads   int64 `json:"total_downloads"`
	TotalSearches    int64 `json:"total_searches"`
	TotalStorage     int64 `json:"total_storage"`
	TotalFiles       int64 `json:"total_files"`
}

// SubscribeStats 订阅统计
type SubscribeStats struct {
	Total  int64 `json:"total"`
	Active int64 `json:"active"`
	Paused int64 `json:"paused"`
}

// DownloadStats 下载统计
type DownloadStats struct {
	Total   int64 `json:"total"`
	Success int64 `json:"success"`
	Failed  int64 `json:"failed"`
	Pending int64 `json:"pending"`
}

// SearchStats 搜索统计
type SearchStats struct {
	TotalSearches int            `json:"total_searches"`
	SiteStats     map[string]int `json:"site_stats"`
}

// TrendData 趋势数据
type TrendData struct {
	Labels    []string `json:"labels"`    // 日期标签
	Downloads []int    `json:"downloads"` // 下载数量
	Searches  []int    `json:"searches"`  // 搜索数量
}

// GetOverallStats 获取总体统计
func (s *service) GetOverallStats(ctx context.Context) (*OverallStats, error) {
	s.logger.Info("获取总体统计")

	stats := &OverallStats{}

	// 订阅统计
	s.db.Model(&struct{}{}).Table("subscribes").Count(&stats.TotalSubscribes)
	s.db.Model(&struct{}{}).Table("subscribes").Where("status = ?", "active").Count(&stats.ActiveSubscribes)

	// 下载统计
	s.db.Model(&struct{}{}).Table("download_histories").Count(&stats.TotalDownloads)

	// 搜索统计（简化）
	stats.TotalSearches = 0

	// 存储统计（简化）
	stats.TotalStorage = 0
	stats.TotalFiles = 0

	return stats, nil
}

// GetSubscribeStats 获取订阅统计
func (s *service) GetSubscribeStats(ctx context.Context) (*SubscribeStats, error) {
	s.logger.Info("获取订阅统计")

	stats := &SubscribeStats{}

	s.db.Model(&struct{}{}).Table("subscribes").Count(&stats.Total)
	s.db.Model(&struct{}{}).Table("subscribes").Where("status = ?", "active").Count(&stats.Active)
	s.db.Model(&struct{}{}).Table("subscribes").Where("status = ?", "paused").Count(&stats.Paused)

	return stats, nil
}

// GetDownloadStats 获取下载统计
func (s *service) GetDownloadStats(ctx context.Context) (*DownloadStats, error) {
	s.logger.Info("获取下载统计")

	stats := &DownloadStats{}

	s.db.Model(&struct{}{}).Table("download_histories").Count(&stats.Total)
	s.db.Model(&struct{}{}).Table("download_histories").Where("status = ?", "completed").Count(&stats.Success)
	s.db.Model(&struct{}{}).Table("download_histories").Where("status = ?", "failed").Count(&stats.Failed)
	s.db.Model(&struct{}{}).Table("download_histories").Where("status = ?", "pending").Count(&stats.Pending)

	return stats, nil
}

// GetSearchStats 获取搜索统计
func (s *service) GetSearchStats(ctx context.Context) (*SearchStats, error) {
	s.logger.Info("获取搜索统计")

	stats := &SearchStats{
		SiteStats: make(map[string]int),
	}

	// 简化实现
	stats.TotalSearches = 0

	return stats, nil
}

// GetTrendData 获取趋势数据
func (s *service) GetTrendData(ctx context.Context, days int) (*TrendData, error) {
	s.logger.Info("获取趋势数据", zap.Int("days", days))

	trend := &TrendData{
		Labels:    make([]string, days),
		Downloads: make([]int, days),
		Searches:  make([]int, days),
	}

	// 简化实现：返回示例数据
	for i := 0; i < days; i++ {
		trend.Labels[i] = "Day " + string(rune(i+1))
		trend.Downloads[i] = 10 + i*2
		trend.Searches[i] = 5 + i
	}

	return trend, nil
}
