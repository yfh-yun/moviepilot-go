package repositories

import (
	"context"
	"errors"
	
	"gorm.io/gorm"
	
	"moviepilot-go/internal/repositories/interfaces"
	"moviepilot-go/internal/models"
)

// SiteStatisticRepositoryImpl 站点统计仓储实现
type SiteStatisticRepositoryImpl struct {
	db *gorm.DB
}

// NewSiteStatisticRepository 创建站点统计仓储实例
func NewSiteStatisticRepository(db *gorm.DB) interfaces.SiteStatisticRepository {
	return &SiteStatisticRepositoryImpl{db: db}
}

// Create 创建站点统计
func (r *SiteStatisticRepositoryImpl) Create(ctx context.Context, statistic *models.SiteStatistic) error {
	return r.db.WithContext(ctx).Create(statistic).Error
}

// GetByID 根据ID获取站点统计
func (r *SiteStatisticRepositoryImpl) GetByID(ctx context.Context, id string) (*models.SiteStatistic, error) {
	var statistic models.SiteStatistic
	err := r.db.WithContext(ctx).First(&statistic, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &statistic, nil
}

// GetBySiteID 根据站点ID获取统计
func (r *SiteStatisticRepositoryImpl) GetBySiteID(ctx context.Context, siteID string) (*models.SiteStatistic, error) {
	var statistic models.SiteStatistic
	err := r.db.WithContext(ctx).Where("site_id = ?", siteID).First(&statistic).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &statistic, nil
}

// Update 更新站点统计
func (r *SiteStatisticRepositoryImpl) Update(ctx context.Context, statistic *models.SiteStatistic) error {
	return r.db.WithContext(ctx).Save(statistic).Error
}

// Delete 删除站点统计
func (r *SiteStatisticRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.SiteStatistic{}, "id = ?", id).Error
}

// List 获取站点统计列表
func (r *SiteStatisticRepositoryImpl) List(ctx context.Context, params interfaces.ListSiteStatisticParams) ([]*models.SiteStatistic, int64, error) {
	var statistics []*models.SiteStatistic
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.SiteStatistic{})
	
	// 添加过滤条件
	if params.SiteID != "" {
		query = query.Where("site_id = ?", params.SiteID)
	}
	if params.DateFrom != "" {
		query = query.Where("date >= ?", params.DateFrom)
	}
	if params.DateTo != "" {
		query = query.Where("date <= ?", params.DateTo)
	}
	
	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// 分页查询
	offset := (params.Page - 1) * params.PageSize
	err := query.Offset(offset).Limit(params.PageSize).Order("date DESC").Find(&statistics).Error
	
	return statistics, total, err
}

// UpdateStatistics 更新统计数据
func (r *SiteStatisticRepositoryImpl) UpdateStatistics(ctx context.Context, siteID string, increment map[string]int64) error {
	updates := make(map[string]interface{})
	for key, value := range increment {
		updates[key] = gorm.Expr(key + " + ?", value)
	}
	
	return r.db.WithContext(ctx).Model(&models.SiteStatistic{}).Where("site_id = ?", siteID).Updates(updates).Error
}