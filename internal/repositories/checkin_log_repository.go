package repositories

import (
	"context"
	"time"

	"gorm.io/gorm"

	"moviepilot-go/internal/models"
)

// CheckinLogRepository 签到日志数据访问接口
type CheckinLogRepository interface {
	Create(ctx context.Context, log *models.CheckinLog) error
	GetByID(ctx context.Context, id uint) (*models.CheckinLog, error)
	GetBySiteID(ctx context.Context, siteID uint, page, limit int) ([]*models.CheckinLog, int64, error)
	GetLatestBySiteID(ctx context.Context, siteID uint) (*models.CheckinLog, error)
	GetSuccessCount(ctx context.Context, siteID uint, since time.Time) (int64, error)
	GetFailedCount(ctx context.Context, siteID uint, since time.Time) (int64, error)
	Delete(ctx context.Context, id uint) error
	DeleteOldLogs(ctx context.Context, before time.Time) error
}

type checkinLogRepository struct {
	db *gorm.DB
}

func NewCheckinLogRepository(db *gorm.DB) CheckinLogRepository {
	return &checkinLogRepository{db: db}
}

func (r *checkinLogRepository) Create(ctx context.Context, log *models.CheckinLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *checkinLogRepository) GetByID(ctx context.Context, id uint) (*models.CheckinLog, error) {
	var log models.CheckinLog
	err := r.db.WithContext(ctx).First(&log, id).Error
	return &log, err
}

func (r *checkinLogRepository) GetBySiteID(ctx context.Context, siteID uint, page, limit int) ([]*models.CheckinLog, int64, error) {
	var logs []*models.CheckinLog
	var total int64

	query := r.db.WithContext(ctx).Model(&models.CheckinLog{}).Where("site_id = ?", siteID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("checkin_time DESC").Find(&logs).Error
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *checkinLogRepository) GetLatestBySiteID(ctx context.Context, siteID uint) (*models.CheckinLog, error) {
	var log models.CheckinLog
	err := r.db.WithContext(ctx).
		Where("site_id = ?", siteID).
		Order("checkin_time DESC").
		First(&log).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &log, nil
}

func (r *checkinLogRepository) GetSuccessCount(ctx context.Context, siteID uint, since time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.CheckinLog{}).
		Where("site_id = ? AND success = ? AND checkin_time >= ?", siteID, true, since).
		Count(&count).Error
	return count, err
}

func (r *checkinLogRepository) GetFailedCount(ctx context.Context, siteID uint, since time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.CheckinLog{}).
		Where("site_id = ? AND success = ? AND checkin_time >= ?", siteID, false, since).
		Count(&count).Error
	return count, err
}

func (r *checkinLogRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.CheckinLog{}, id).Error
}

func (r *checkinLogRepository) DeleteOldLogs(ctx context.Context, before time.Time) error {
	return r.db.WithContext(ctx).
		Where("checkin_time < ?", before).
		Delete(&models.CheckinLog{}).Error
}
