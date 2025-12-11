package repositories

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"moviepilot-go/internal/models/database"
	"moviepilot-go/internal/repositories/interfaces"
)

// SiteRepositoryImpl 站点仓储实现
type SiteRepositoryImpl struct {
	db *gorm.DB
}

// NewSiteRepository 创建站点仓储实例
func NewSiteRepository(db *gorm.DB) interfaces.SiteRepository {
	return &SiteRepositoryImpl{db: db}
}

// Create 创建站点
func (r *SiteRepositoryImpl) Create(ctx context.Context, site *database.Site) error {
	return r.db.WithContext(ctx).Create(site).Error
}

// Update 更新站点
func (r *SiteRepositoryImpl) Update(ctx context.Context, site *database.Site) error {
	return r.db.WithContext(ctx).Save(site).Error
}

// Delete 删除站点
func (r *SiteRepositoryImpl) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&database.Site{}, id).Error
}

// GetByID 根据ID获取站点
func (r *SiteRepositoryImpl) GetByID(ctx context.Context, id uint) (*database.Site, error) {
	var site database.Site
	err := r.db.WithContext(ctx).First(&site, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &site, nil
}

// GetByDomain 根据域名获取站点
func (r *SiteRepositoryImpl) GetByDomain(ctx context.Context, domain string) (*database.Site, error) {
	var site database.Site
	err := r.db.WithContext(ctx).Where("domain = ?", domain).First(&site).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &site, nil
}

// List 获取站点列表
func (r *SiteRepositoryImpl) List(ctx context.Context, opts interfaces.ListOptions) ([]*database.Site, int64, error) {
	var sites []*database.Site
	var total int64

	query := r.db.WithContext(ctx).Model(&database.Site{})

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页
	offset := opts.GetOffset()
	limit := opts.GetLimit()
	query = query.Offset(offset).Limit(limit)

	// 应用排序
	orderBy := "pri DESC, created_at DESC"
	if opts.OrderBy != "" {
		order := opts.Order
		if order == "" {
			order = "DESC"
		}
		orderBy = fmt.Sprintf("%s %s", opts.OrderBy, order)
	}
	query = query.Order(orderBy)

	// 执行查询
	err := query.Find(&sites).Error
	return sites, total, err
}

// ListActive 获取启用的站点列表
func (r *SiteRepositoryImpl) ListActive(ctx context.Context) ([]*database.Site, error) {
	var sites []*database.Site
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Order("pri DESC").Find(&sites).Error
	return sites, err
}

// UpdateStatus 更新站点状态
func (r *SiteRepositoryImpl) UpdateStatus(ctx context.Context, id uint, isActive bool) error {
	return r.db.WithContext(ctx).Model(&database.Site{}).Where("id = ?", id).Update("is_active", isActive).Error
}

// IncrementFailCount 增加失败次数
func (r *SiteRepositoryImpl) IncrementFailCount(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&database.Site{}).Where("id = ?", id).UpdateColumn("fail_count", gorm.Expr("fail_count + ?", 1)).Error
}

// ResetFailCount 重置失败次数
func (r *SiteRepositoryImpl) ResetFailCount(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&database.Site{}).Where("id = ?", id).UpdateColumn("fail_count", 0).Error
}

// UpdateStatistics 更新站点统计
func (r *SiteRepositoryImpl) UpdateStatistics(ctx context.Context, siteName string, success bool, seconds int) error {
	// TODO: 实现站点统计更新逻辑
	// 1. 查询站点
	// 2. 更新统计信息
	// 3. 保存到数据库
	return nil
}
