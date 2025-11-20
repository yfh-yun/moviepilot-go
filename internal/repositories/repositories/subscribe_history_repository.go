package repositories

import (
	"errors"
	"fmt"
	"github.com/yfh-yun/moviepilot-go/internal/repositories/interfaces"
	"github.com/yfh-yun/moviepilot-go/internal/models"
	"time"

	"gorm.io/gorm"
)

// subscribeHistoryRepository 订阅历史仓储实现
type subscribeHistoryRepository struct {
	db *gorm.DB
}

// NewSubscribeHistoryRepository 创建订阅历史仓储
func NewSubscribeHistoryRepository(db *gorm.DB) interfaces.SubscribeHistoryRepository {
	return &subscribeHistoryRepository{db: db}
}

// Create 创建订阅历史
func (r *subscribeHistoryRepository) Create(history *model.SubscribeHistory) error {
	if history == nil {
		return errors.New("subscribe history cannot be nil")
	}
	return r.db.Create(history).Error
}

// BatchCreate 批量创建订阅历史
func (r *subscribeHistoryRepository) BatchCreate(histories []*model.SubscribeHistory) error {
	if len(histories) == 0 {
		return nil
	}
	return r.db.CreateInBatches(histories, 100).Error
}

// GetByID 根据ID获取订阅历史
func (r *subscribeHistoryRepository) GetByID(id uint) (*model.SubscribeHistory, error) {
	var history model.SubscribeHistory
	err := r.db.First(&history, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &history, nil
}

// GetBySubscribeID 根据订阅ID获取历史
func (r *subscribeHistoryRepository) GetBySubscribeID(subscribeID uint) ([]*model.SubscribeHistory, error) {
	var histories []*model.SubscribeHistory
	err := r.db.Where("subscribe_id = ?", subscribeID).Order("created_at DESC").Find(&histories).Error
	return histories, err
}

// GetByDownloadHash 根据下载Hash获取历史
func (r *subscribeHistoryRepository) GetByDownloadHash(downloadHash string) ([]*model.SubscribeHistory, error) {
	var histories []*model.SubscribeHistory
	err := r.db.Where("download_hash = ?", downloadHash).Order("created_at DESC").Find(&histories).Error
	return histories, err
}

// GetByDateRange 根据日期范围获取历史
func (r *subscribeHistoryRepository) GetByDateRange(startDate, endDate time.Time) ([]*model.SubscribeHistory, error) {
	var histories []*model.SubscribeHistory
	err := r.db.Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Order("created_at DESC").Find(&histories).Error
	return histories, err
}

// GetByStatus 根据状态获取历史
func (r *subscribeHistoryRepository) GetByStatus(status string) ([]*model.SubscribeHistory, error) {
	var histories []*model.SubscribeHistory
	err := r.db.Where("status = ?", status).Order("created_at DESC").Find(&histories).Error
	return histories, err
}

// GetByType 根据类型获取历史
func (r *subscribeHistoryRepository) GetByType(subscribeType string) ([]*model.SubscribeHistory, error) {
	var histories []*model.SubscribeHistory
	err := r.db.Where("type = ?", subscribeType).Order("created_at DESC").Find(&histories).Error
	return histories, err
}

// Update 更新订阅历史
func (r *subscribeHistoryRepository) Update(history *model.SubscribeHistory) error {
	if history == nil {
		return errors.New("subscribe history cannot be nil")
	}
	return r.db.Save(history).Error
}

// UpdateStatus 更新订阅历史状态
func (r *subscribeHistoryRepository) UpdateStatus(id uint, status string) error {
	result := r.db.Model(&model.SubscribeHistory{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("failed to update subscribe history status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("subscribe history with id %d not found", id)
	}
	return nil
}

// Delete 删除订阅历史
func (r *subscribeHistoryRepository) Delete(id uint) error {
	return r.db.Delete(&model.SubscribeHistory{}, id).Error
}

// DeleteBySubscribeID 根据订阅ID删除历史
func (r *subscribeHistoryRepository) DeleteBySubscribeID(subscribeID uint) error {
	return r.db.Where("subscribe_id = ?", subscribeID).Delete(&model.SubscribeHistory{}).Error
}

// DeleteByDateRange 根据日期范围删除历史
func (r *subscribeHistoryRepository) DeleteByDateRange(startDate, endDate time.Time) error {
	return r.db.Where("created_at BETWEEN ? AND ?", startDate, endDate).Delete(&model.SubscribeHistory{}).Error
}

// List 分页获取订阅历史列表
func (r *subscribeHistoryRepository) List(offset, limit int) ([]*model.SubscribeHistory, int64, error) {
	var histories []*model.SubscribeHistory
	var total int64

	err := r.db.Model(&model.SubscribeHistory{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&histories).Error
	return histories, total, err
}

// Count 统计订阅历史数量
func (r *subscribeHistoryRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.SubscribeHistory{}).Count(&count).Error
	return count, err
}

// CountBySubscribeID 根据订阅ID统计数量
func (r *subscribeHistoryRepository) CountBySubscribeID(subscribeID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.SubscribeHistory{}).Where("subscribe_id = ?", subscribeID).Count(&count).Error
	return count, err
}

// GetLatest 获取最新的订阅历史
func (r *subscribeHistoryRepository) GetLatest(subscribeID uint, limit int) ([]*model.SubscribeHistory, error) {
	var histories []*model.SubscribeHistory
	err := r.db.Where("subscribe_id = ?", subscribeID).
		Order("created_at DESC").
		Limit(limit).
		Find(&histories).Error
	return histories, err
}