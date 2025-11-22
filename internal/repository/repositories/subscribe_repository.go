package repositories

import (
	"context"
	"errors"
	
	"gorm.io/gorm"
	
	"moviepilot-go/internal/repository/interfaces"
	"moviepilot-go/internal/models"
)

// SubscribeRepositoryImpl 订阅仓储实现
type SubscribeRepositoryImpl struct {
	db *gorm.DB
}

// NewSubscribeRepository 创建订阅仓储实例
func NewSubscribeRepository(db *gorm.DB) interfaces.SubscribeRepository {
	return &SubscribeRepositoryImpl{db: db}
}

// Create 创建订阅
func (r *SubscribeRepositoryImpl) Create(ctx context.Context, subscribe *models.Subscribe) error {
	return r.db.WithContext(ctx).Create(subscribe).Error
}

// GetByID 根据ID获取订阅
func (r *SubscribeRepositoryImpl) GetByID(ctx context.Context, id string) (*models.Subscribe, error) {
	var subscribe models.Subscribe
	err := r.db.WithContext(ctx).First(&subscribe, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &subscribe, nil
}

// Update 更新订阅
func (r *SubscribeRepositoryImpl) Update(ctx context.Context, subscribe *models.Subscribe) error {
	return r.db.WithContext(ctx).Save(subscribe).Error
}

// Delete 删除订阅
func (r *SubscribeRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Subscribe{}, "id = ?", id).Error
}

// List 获取订阅列表
func (r *SubscribeRepositoryImpl) List(ctx context.Context, params interfaces.ListSubscribeParams) ([]*models.Subscribe, int64, error) {
	var subscribes []*models.Subscribe
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.Subscribe{})
	
	// 添加过滤条件
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.Type != "" {
		query = query.Where("type = ?", params.Type)
	}
	if params.UserID != "" {
		query = query.Where("user_id = ?", params.UserID)
	}
	
	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// 分页查询
	offset := (params.Page - 1) * params.PageSize
	err := query.Offset(offset).Limit(params.PageSize).Find(&subscribes).Error
	
	return subscribes, total, err
}

// GetByUserID 根据用户ID获取订阅列表
func (r *SubscribeRepositoryImpl) GetByUserID(ctx context.Context, userID string) ([]*models.Subscribe, error) {
	var subscribes []*models.Subscribe
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&subscribes).Error
	return subscribes, err
}

// GetActiveSubscriptions 获取活跃订阅
func (r *SubscribeRepositoryImpl) GetActiveSubscriptions(ctx context.Context) ([]*models.Subscribe, error) {
	var subscribes []*models.Subscribe
	err := r.db.WithContext(ctx).Where("status = ?", "active").Find(&subscribes).Error
	return subscribes, err
}