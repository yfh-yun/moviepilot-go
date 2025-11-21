package repositories

import (
	"context"
	"errors"
	
	"gorm.io/gorm"
	
	"moviepilot-go/internal/repositories/interfaces"
	"moviepilot-go/internal/models"
)

// SubscribeHistoryRepositoryImpl 订阅历史仓储实现
type SubscribeHistoryRepositoryImpl struct {
	db *gorm.DB
}

// NewSubscribeHistoryRepository 创建订阅历史仓储实例
func NewSubscribeHistoryRepository(db *gorm.DB) interfaces.SubscribeHistoryRepository {
	return &SubscribeHistoryRepositoryImpl{db: db}
}

// Create 创建订阅历史
func (r *SubscribeHistoryRepositoryImpl) Create(ctx context.Context, history *models.SubscribeHistory) error {
	return r.db.WithContext(ctx).Create(history).Error
}

// GetByID 根据ID获取订阅历史
func (r *SubscribeHistoryRepositoryImpl) GetByID(ctx context.Context, id string) (*models.SubscribeHistory, error) {
	var history models.SubscribeHistory
	err := r.db.WithContext(ctx).First(&history, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &history, nil
}

// GetBySubscribeID 根据订阅ID获取历史
func (r *SubscribeHistoryRepositoryImpl) GetBySubscribeID(ctx context.Context, subscribeID string, params interfaces.ListSubscribeHistoryParams) ([]*models.SubscribeHistory, int64, error) {
	var histories []*models.SubscribeHistory
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.SubscribeHistory{}).Where("subscribe_id = ?", subscribeID)
	
	// 添加过滤条件
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.DateFrom != "" {
		query = query.Where("created_at >= ?", params.DateFrom)
	}
	if params.DateTo != "" {
		query = query.Where("created_at <= ?", params.DateTo)
	}
	
	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// 分页查询
	offset := (params.Page - 1) * params.PageSize
	err := query.Offset(offset).Limit(params.PageSize).Order("created_at DESC").Find(&histories).Error
	
	return histories, total, err
}

// Update 更新订阅历史
func (r *SubscribeHistoryRepositoryImpl) Update(ctx context.Context, history *models.SubscribeHistory) error {
	return r.db.WithContext(ctx).Save(history).Error
}

// Delete 删除订阅历史
func (r *SubscribeHistoryRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.SubscribeHistory{}, "id = ?", id).Error
}

// List 获取订阅历史列表
func (r *SubscribeHistoryRepositoryImpl) List(ctx context.Context, params interfaces.ListSubscribeHistoryParams) ([]*models.SubscribeHistory, int64, error) {
	var histories []*models.SubscribeHistory
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.SubscribeHistory{})
	
	// 添加过滤条件
	if params.SubscribeID != "" {
		query = query.Where("subscribe_id = ?", params.SubscribeID)
	}
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.DateFrom != "" {
		query = query.Where("created_at >= ?", params.DateFrom)
	}
	if params.DateTo != "" {
		query = query.Where("created_at <= ?", params.DateTo)
	}
	
	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// 分页查询
	offset := (params.Page - 1) * params.PageSize
	err := query.Offset(offset).Limit(params.PageSize).Order("created_at DESC").Find(&histories).Error
	
	return histories, total, err
}

// GetLatestBySubscribeID 获取订阅的最新历史
func (r *SubscribeHistoryRepositoryImpl) GetLatestBySubscribeID(ctx context.Context, subscribeID string) (*models.SubscribeHistory, error) {
	var history models.SubscribeHistory
	err := r.db.WithContext(ctx).Where("subscribe_id = ?", subscribeID).Order("created_at DESC").First(&history).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &history, nil
}