package download

import (
	"context"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"moviepilot-go/pkg/logger"
)

// AnalyticsService 下载分析服务
type AnalyticsService interface {
	// GetOverallStats 获取总体统计
	GetOverallStats(ctx context.Context) (*OverallStats, error)

	// GetDailyStats 获取每日统计
	GetDailyStats(ctx context.Context, days int) ([]*DailyStats, error)

	// GetCategoryStats 获取分类统计
	GetCategoryStats(ctx context.Context) ([]*CategoryStats, error)

	// GetSpeedStats 获取速度统计
	GetSpeedStats(ctx context.Context, hours int) (*SpeedStats, error)
}

// analyticsService 分析服务实现
type analyticsService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewAnalyticsService 创建分析服务
func NewAnalyticsService(db *gorm.DB) AnalyticsService {
	return &analyticsService{
		db:     db,
		logger: logger.GetLogger(),
	}
}

// OverallStats 总体统计
type OverallStats struct {
	TotalTasks       int64   `json:"total_tasks"`
	CompletedTasks   int64   `json:"completed_tasks"`
	FailedTasks      int64   `json:"failed_tasks"`
	ActiveTasks      int64   `json:"active_tasks"`
	TotalDownloaded  int64   `json:"total_downloaded"`   // 字节
	TotalUploaded    int64   `json:"total_uploaded"`     // 字节
	AvgDownloadSpeed float64 `json:"avg_download_speed"` // 字节/秒
	AvgUploadSpeed   float64 `json:"avg_upload_speed"`   // 字节/秒
	SuccessRate      float64 `json:"success_rate"`       // 百分比
}

// DailyStats 每日统计
type DailyStats struct {
	Date             string  `json:"date"`
	NewTasks         int64   `json:"new_tasks"`
	CompletedTasks   int64   `json:"completed_tasks"`
	FailedTasks      int64   `json:"failed_tasks"`
	Downloaded       int64   `json:"downloaded"`
	Uploaded         int64   `json:"uploaded"`
	AvgDownloadSpeed float64 `json:"avg_download_speed"`
}

// CategoryStats 分类统计
type CategoryStats struct {
	Category       string  `json:"category"`
	TaskCount      int64   `json:"task_count"`
	TotalSize      int64   `json:"total_size"`
	CompletedCount int64   `json:"completed_count"`
	SuccessRate    float64 `json:"success_rate"`
}

// SpeedStats 速度统计
type SpeedStats struct {
	CurrentDownloadSpeed int64   `json:"current_download_speed"`
	CurrentUploadSpeed   int64   `json:"current_upload_speed"`
	PeakDownloadSpeed    int64   `json:"peak_download_speed"`
	PeakUploadSpeed      int64   `json:"peak_upload_speed"`
	AvgDownloadSpeed     float64 `json:"avg_download_speed"`
	AvgUploadSpeed       float64 `json:"avg_upload_speed"`
}

// DownloadHistory 下载历史（简化模型）
type DownloadHistory struct {
	ID           uint   `gorm:"primaryKey"`
	Hash         string `gorm:"size:100;uniqueIndex"`
	Title        string `gorm:"size:500"`
	Size         int64
	Status       string `gorm:"size:20"`
	Downloaded   int64
	Uploaded     int64
	DownloadRate int64
	UploadRate   int64
	Category     string    `gorm:"size:100"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	CompletedAt  *time.Time
}

// TableName 表名
func (DownloadHistory) TableName() string {
	return "download_history"
}

// GetOverallStats 获取总体统计
func (s *analyticsService) GetOverallStats(ctx context.Context) (*OverallStats, error) {
	s.logger.Info("获取下载总体统计")

	stats := &OverallStats{}

	// 总任务数
	if err := s.db.WithContext(ctx).
		Model(&DownloadHistory{}).
		Count(&stats.TotalTasks).Error; err != nil {
		return nil, err
	}

	// 完成任务数
	if err := s.db.WithContext(ctx).
		Model(&DownloadHistory{}).
		Where("status = ?", "completed").
		Count(&stats.CompletedTasks).Error; err != nil {
		return nil, err
	}

	// 失败任务数
	if err := s.db.WithContext(ctx).
		Model(&DownloadHistory{}).
		Where("status = ?", "failed").
		Count(&stats.FailedTasks).Error; err != nil {
		return nil, err
	}

	// 活跃任务数
	if err := s.db.WithContext(ctx).
		Model(&DownloadHistory{}).
		Where("status IN ?", []string{"downloading", "queued", "pending"}).
		Count(&stats.ActiveTasks).Error; err != nil {
		return nil, err
	}

	// 总下载量和上传量
	var sums struct {
		TotalDownloaded int64
		TotalUploaded   int64
	}
	if err := s.db.WithContext(ctx).
		Model(&DownloadHistory{}).
		Select("SUM(downloaded) as total_downloaded, SUM(uploaded) as total_uploaded").
		Scan(&sums).Error; err == nil {
		stats.TotalDownloaded = sums.TotalDownloaded
		stats.TotalUploaded = sums.TotalUploaded
	}

	// 平均速度
	var avgSpeeds struct {
		AvgDownloadSpeed float64
		AvgUploadSpeed   float64
	}
	if err := s.db.WithContext(ctx).
		Model(&DownloadHistory{}).
		Select("AVG(download_rate) as avg_download_speed, AVG(upload_rate) as avg_upload_speed").
		Where("status = ?", "downloading").
		Scan(&avgSpeeds).Error; err == nil {
		stats.AvgDownloadSpeed = avgSpeeds.AvgDownloadSpeed
		stats.AvgUploadSpeed = avgSpeeds.AvgUploadSpeed
	}

	// 成功率
	if stats.TotalTasks > 0 {
		stats.SuccessRate = float64(stats.CompletedTasks) / float64(stats.TotalTasks) * 100
	}

	return stats, nil
}

// GetDailyStats 获取每日统计
func (s *analyticsService) GetDailyStats(ctx context.Context, days int) ([]*DailyStats, error) {
	s.logger.Info("获取每日统计", zap.Int("days", days))

	startDate := time.Now().AddDate(0, 0, -days)
	dailyStats := make([]*DailyStats, 0)

	for i := 0; i < days; i++ {
		date := startDate.AddDate(0, 0, i)
		dateStr := date.Format("2006-01-02")
		nextDate := date.AddDate(0, 0, 1)

		stats := &DailyStats{
			Date: dateStr,
		}

		// 新增任务数
		s.db.WithContext(ctx).
			Model(&DownloadHistory{}).
			Where("created_at >= ? AND created_at < ?", date, nextDate).
			Count(&stats.NewTasks)

		// 完成任务数
		s.db.WithContext(ctx).
			Model(&DownloadHistory{}).
			Where("status = ? AND completed_at >= ? AND completed_at < ?", "completed", date, nextDate).
			Count(&stats.CompletedTasks)

		// 失败任务数
		s.db.WithContext(ctx).
			Model(&DownloadHistory{}).
			Where("status = ? AND created_at >= ? AND created_at < ?", "failed", date, nextDate).
			Count(&stats.FailedTasks)

		// 下载量和上传量
		var sums struct {
			Downloaded int64
			Uploaded   int64
		}
		s.db.WithContext(ctx).
			Model(&DownloadHistory{}).
			Select("SUM(downloaded) as downloaded, SUM(uploaded) as uploaded").
			Where("created_at >= ? AND created_at < ?", date, nextDate).
			Scan(&sums)
		stats.Downloaded = sums.Downloaded
		stats.Uploaded = sums.Uploaded

		dailyStats = append(dailyStats, stats)
	}

	return dailyStats, nil
}

// GetCategoryStats 获取分类统计
func (s *analyticsService) GetCategoryStats(ctx context.Context) ([]*CategoryStats, error) {
	s.logger.Info("获取分类统计")

	var results []*CategoryStats

	query := `
		SELECT 
			category,
			COUNT(*) as task_count,
			SUM(size) as total_size,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_count
		FROM download_history
		WHERE category != ''
		GROUP BY category
		ORDER BY task_count DESC
	`

	if err := s.db.WithContext(ctx).Raw(query).Scan(&results).Error; err != nil {
		s.logger.Error("获取分类统计失败", zap.Error(err))
		return nil, err
	}

	// 计算成功率
	for _, stat := range results {
		if stat.TaskCount > 0 {
			stat.SuccessRate = float64(stat.CompletedCount) / float64(stat.TaskCount) * 100
		}
	}

	return results, nil
}

// GetSpeedStats 获取速度统计
func (s *analyticsService) GetSpeedStats(ctx context.Context, hours int) (*SpeedStats, error) {
	s.logger.Info("获取速度统计", zap.Int("hours", hours))

	stats := &SpeedStats{}

	startTime := time.Now().Add(-time.Duration(hours) * time.Hour)

	// 当前速度（活跃任务的平均速度）
	var currentSpeeds struct {
		DownloadSpeed int64
		UploadSpeed   int64
	}
	if err := s.db.WithContext(ctx).
		Model(&DownloadHistory{}).
		Select("AVG(download_rate) as download_speed, AVG(upload_rate) as upload_speed").
		Where("status = ?", "downloading").
		Scan(&currentSpeeds).Error; err == nil {
		stats.CurrentDownloadSpeed = currentSpeeds.DownloadSpeed
		stats.CurrentUploadSpeed = currentSpeeds.UploadSpeed
	}

	// 峰值速度
	var peakSpeeds struct {
		PeakDownload int64
		PeakUpload   int64
	}
	if err := s.db.WithContext(ctx).
		Model(&DownloadHistory{}).
		Select("MAX(download_rate) as peak_download, MAX(upload_rate) as peak_upload").
		Where("created_at >= ?", startTime).
		Scan(&peakSpeeds).Error; err == nil {
		stats.PeakDownloadSpeed = peakSpeeds.PeakDownload
		stats.PeakUploadSpeed = peakSpeeds.PeakUpload
	}

	// 平均速度
	var avgSpeeds struct {
		AvgDownload float64
		AvgUpload   float64
	}
	if err := s.db.WithContext(ctx).
		Model(&DownloadHistory{}).
		Select("AVG(download_rate) as avg_download, AVG(upload_rate) as avg_upload").
		Where("created_at >= ?", startTime).
		Scan(&avgSpeeds).Error; err == nil {
		stats.AvgDownloadSpeed = avgSpeeds.AvgDownload
		stats.AvgUploadSpeed = avgSpeeds.AvgUpload
	}

	return stats, nil
}
