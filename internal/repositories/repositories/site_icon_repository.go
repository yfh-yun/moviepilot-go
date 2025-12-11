package repositories

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"moviepilot-go/internal/models/database"
	"moviepilot-go/internal/repositories/interfaces"
)

// SiteIconRepositoryImpl 站点图标仓储实现
type SiteIconRepositoryImpl struct {
	db *gorm.DB
}

// NewSiteIconRepository 创建站点图标仓储实例
func NewSiteIconRepository(db *gorm.DB) interfaces.SiteIconRepository {
	return &SiteIconRepositoryImpl{db: db}
}

// Create 创建站点图标
func (r *SiteIconRepositoryImpl) Create(ctx context.Context, icon *database.SiteIcon) error {
	return r.db.WithContext(ctx).Create(icon).Error
}

// GetByID 根据ID获取站点图标
func (r *SiteIconRepositoryImpl) GetByID(ctx context.Context, id string) (*database.SiteIcon, error) {
	var icon database.SiteIcon
	err := r.db.WithContext(ctx).First(&icon, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &icon, nil
}

// GetByDomain 根据域名获取站点图标
func (r *SiteIconRepositoryImpl) GetByDomain(ctx context.Context, domain string) (*database.SiteIcon, error) {
	var icon database.SiteIcon
	err := r.db.WithContext(ctx).Where("domain = ?", domain).First(&icon).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &icon, nil
}

// Update 更新站点图标
func (r *SiteIconRepositoryImpl) Update(ctx context.Context, icon *database.SiteIcon) error {
	return r.db.WithContext(ctx).Save(icon).Error
}

// Delete 删除站点图标
func (r *SiteIconRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&database.SiteIcon{}, "id = ?", id).Error
}

// List 获取站点图标列表
func (r *SiteIconRepositoryImpl) List(ctx context.Context, params interfaces.ListSiteIconParams) ([]*database.SiteIcon, int64, error) {
	var icons []*database.SiteIcon
	var total int64

	query := r.db.WithContext(ctx).Model(&database.SiteIcon{})

	// 添加过滤条件
	if params.Domain != "" {
		query = query.Where("domain LIKE ?", "%"+params.Domain+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (params.Page - 1) * params.PageSize
	err := query.Offset(offset).Limit(params.PageSize).Find(&icons).Error

	return icons, total, err
}
