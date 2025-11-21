package repositories

import (
	"context"
	"errors"
	
	"gorm.io/gorm"
	
	"moviepilot-go/internal/repositories/interfaces"
	"moviepilot-go/internal/models"
)

// UserConfigRepositoryImpl 用户配置仓储实现
type UserConfigRepositoryImpl struct {
	db *gorm.DB
}

// NewUserConfigRepository 创建用户配置仓储实例
func NewUserConfigRepository(db *gorm.DB) interfaces.UserConfigRepository {
	return &UserConfigRepositoryImpl{db: db}
}

// Create 创建用户配置
func (r *UserConfigRepositoryImpl) Create(ctx context.Context, config *models.UserConfig) error {
	return r.db.WithContext(ctx).Create(config).Error
}

// GetByID 根据ID获取用户配置
func (r *UserConfigRepositoryImpl) GetByID(ctx context.Context, id string) (*models.UserConfig, error) {
	var config models.UserConfig
	err := r.db.WithContext(ctx).First(&config, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &config, nil
}

// GetByUserID 根据用户ID获取配置
func (r *UserConfigRepositoryImpl) GetByUserID(ctx context.Context, userID string) (*models.UserConfig, error) {
	var config models.UserConfig
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&config).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &config, nil
}

// Update 更新用户配置
func (r *UserConfigRepositoryImpl) Update(ctx context.Context, config *models.UserConfig) error {
	return r.db.WithContext(ctx).Save(config).Error
}

// Delete 删除用户配置
func (r *UserConfigRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.UserConfig{}, "id = ?", id).Error
}

// List 获取用户配置列表
func (r *UserConfigRepositoryImpl) List(ctx context.Context, params interfaces.ListUserConfigParams) ([]*models.UserConfig, int64, error) {
	var configs []*models.UserConfig
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.UserConfig{})
	
	// 添加过滤条件
	if params.UserID != "" {
		query = query.Where("user_id = ?", params.UserID)
	}
	if params.Key != "" {
		query = query.Where("key = ?", params.Key)
	}
	
	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// 分页查询
	offset := (params.Page - 1) * params.PageSize
	err := query.Offset(offset).Limit(params.PageSize).Find(&configs).Error
	
	return configs, total, err
}