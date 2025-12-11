package repositories

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"moviepilot-go/internal/models/database"
	"moviepilot-go/internal/repositories/interfaces"
)

// SiteUserDataRepositoryImpl 站点用户数据仓储实现
type SiteUserDataRepositoryImpl struct {
	db *gorm.DB
}

// NewSiteUserDataRepository 创建站点用户数据仓储实例
func NewSiteUserDataRepository(db *gorm.DB) interfaces.SiteUserDataRepository {
	return &SiteUserDataRepositoryImpl{db: db}
}

// Create 创建站点用户数据
func (r *SiteUserDataRepositoryImpl) Create(ctx context.Context, userData *database.SiteUserData) error {
	return r.db.WithContext(ctx).Create(userData).Error
}

// GetByID 根据ID获取站点用户数据
func (r *SiteUserDataRepositoryImpl) GetByID(ctx context.Context, id string) (*database.SiteUserData, error) {
	var userData database.SiteUserData
	err := r.db.WithContext(ctx).First(&userData, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &userData, nil
}

// GetByUserID 根据用户ID获取站点用户数据
func (r *SiteUserDataRepositoryImpl) GetByUserID(ctx context.Context, userID string) ([]*database.SiteUserData, error) {
	var userDataList []*database.SiteUserData
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&userDataList).Error
	return userDataList, err
}

// GetBySiteID 根据站点ID获取用户数据
func (r *SiteUserDataRepositoryImpl) GetBySiteID(ctx context.Context, siteID string) ([]*database.SiteUserData, error) {
	var userDataList []*database.SiteUserData
	err := r.db.WithContext(ctx).Where("site_id = ?", siteID).Find(&userDataList).Error
	return userDataList, err
}

// GetByUserAndSite 根据用户ID和站点ID获取数据
func (r *SiteUserDataRepositoryImpl) GetByUserAndSite(ctx context.Context, userID, siteID string) (*database.SiteUserData, error) {
	var userData database.SiteUserData
	err := r.db.WithContext(ctx).Where("user_id = ? AND site_id = ?", userID, siteID).First(&userData).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &userData, nil
}

// Update 更新站点用户数据
func (r *SiteUserDataRepositoryImpl) Update(ctx context.Context, userData *database.SiteUserData) error {
	return r.db.WithContext(ctx).Save(userData).Error
}

// Delete 删除站点用户数据
func (r *SiteUserDataRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&database.SiteUserData{}, "id = ?", id).Error
}

// List 获取站点用户数据列表
func (r *SiteUserDataRepositoryImpl) List(ctx context.Context, params interfaces.ListSiteUserDataParams) ([]*database.SiteUserData, int64, error) {
	var userDataList []*database.SiteUserData
	var total int64

	query := r.db.WithContext(ctx).Model(&database.SiteUserData{})

	// 添加过滤条件
	if params.UserID != "" {
		query = query.Where("user_id = ?", params.UserID)
	}
	if params.SiteID != "" {
		query = query.Where("site_id = ?", params.SiteID)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (params.Page - 1) * params.PageSize
	err := query.Offset(offset).Limit(params.PageSize).Find(&userDataList).Error

	return userDataList, total, err
}
