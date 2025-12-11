package repositories

import (
	"context"
	"time"

	"gorm.io/gorm"

	"moviepilot-go/internal/models/entity"
)

// CheckinRepository 签到数据仓库接口
type CheckinRepository interface {
	// Create 创建签到记录
	Create(ctx context.Context, record *entity.SiteCheckinHistory) error

	// GetHistory 获取签到历史
	GetHistory(ctx context.Context, siteID uint, startDate time.Time) ([]*entity.SiteCheckinHistory, error)

	// GetStats 获取签到统计
	GetStats(ctx context.Context, siteID uint) (*CheckinStats, error)

	// GetLastCheckin 获取最后一次签到
	GetLastCheckin(ctx context.Context, siteID uint) (*entity.SiteCheckinHistory, error)
}

// CheckinStats 签到统计
type CheckinStats struct {
	TotalCheckins   int64
	SuccessCheckins int64
	FailedCheckins  int64
	TotalBonus      float64
	TotalUpload     int64
	TotalDownload   int64
	LastCheckinTime *time.Time
	SuccessRate     float64
}

// checkinRepository 签到仓库实现
type checkinRepository struct {
	db *gorm.DB
}

// NewCheckinRepository 创建签到仓库
func NewCheckinRepository(db *gorm.DB) CheckinRepository {
	return &checkinRepository{db: db}
}

// Create 创建签到记录
func (r *checkinRepository) Create(ctx context.Context, record *entity.SiteCheckinHistory) error {
	return r.db.WithContext(ctx).Create(record).Error
}

// GetHistory 获取签到历史
func (r *checkinRepository) GetHistory(ctx context.Context, siteID uint, startDate time.Time) ([]*entity.SiteCheckinHistory, error) {
	var records []*entity.SiteCheckinHistory
	err := r.db.WithContext(ctx).
		Where("site_id = ? AND created_at >= ?", siteID, startDate).
		Order("created_at DESC").
		Find(&records).Error
	return records, err
}

// GetStats 获取签到统计
func (r *checkinRepository) GetStats(ctx context.Context, siteID uint) (*CheckinStats, error) {
	stats := &CheckinStats{}

	// 总签到数
	if err := r.db.WithContext(ctx).
		Model(&entity.SiteCheckinHistory{}).
		Where("site_id = ?", siteID).
		Count(&stats.TotalCheckins).Error; err != nil {
		return nil, err
	}

	// 成功签到数
	if err := r.db.WithContext(ctx).
		Model(&entity.SiteCheckinHistory{}).
		Where("site_id = ? AND success = ?", siteID, true).
		Count(&stats.SuccessCheckins).Error; err != nil {
		return nil, err
	}

	stats.FailedCheckins = stats.TotalCheckins - stats.SuccessCheckins

	// 成功率
	if stats.TotalCheckins > 0 {
		stats.SuccessRate = float64(stats.SuccessCheckins) / float64(stats.TotalCheckins) * 100
	}

	// 总积分/流量
	var sums struct {
		TotalBonus    float64
		TotalUpload   int64
		TotalDownload int64
	}
	if err := r.db.WithContext(ctx).
		Model(&entity.SiteCheckinHistory{}).
		Select("SUM(bonus) as total_bonus, SUM(upload) as total_upload, SUM(download) as total_download").
		Where("site_id = ? AND success = ?", siteID, true).
		Scan(&sums).Error; err == nil {
		stats.TotalBonus = sums.TotalBonus
		stats.TotalUpload = sums.TotalUpload
		stats.TotalDownload = sums.TotalDownload
	}

	// 最后签到时间
	var lastRecord entity.SiteCheckinHistory
	if err := r.db.WithContext(ctx).
		Where("site_id = ?", siteID).
		Order("created_at DESC").
		First(&lastRecord).Error; err == nil {
		stats.LastCheckinTime = &lastRecord.CreatedAt
	}

	return stats, nil
}

// GetLastCheckin 获取最后一次签到
func (r *checkinRepository) GetLastCheckin(ctx context.Context, siteID uint) (*entity.SiteCheckinHistory, error) {
	var record entity.SiteCheckinHistory
	err := r.db.WithContext(ctx).
		Where("site_id = ?", siteID).
		Order("created_at DESC").
		First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}
