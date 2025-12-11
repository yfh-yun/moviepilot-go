package repositories

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"moviepilot-go/internal/models"
)

// SiteRepository 站点数据访问接口
type SiteRepository interface {
	Create(ctx context.Context, site *models.Site) error
	GetByID(ctx context.Context, id uint) (*models.Site, error)
	GetByDomain(ctx context.Context, domain string) (*models.Site, error)
	Update(ctx context.Context, site *models.Site) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, userID uint, page, limit int) ([]*models.Site, int64, error)
	ListByPriority(ctx context.Context, userID uint) ([]*models.Site, error)
	GetByUserID(ctx context.Context, userID uint) ([]*models.Site, error)
	GetEnabledSites(ctx context.Context, userID uint) ([]*models.Site, error)
	GetCheckinEnabledSites(ctx context.Context) ([]*models.Site, error)
	UpdateStatus(ctx context.Context, id uint, status, errorMsg string) error
	UpdateStats(ctx context.Context, id uint, upload, download int64, ratio float64) error
	UpdatePriority(ctx context.Context, id uint, priority int) error
	UpdatePriorities(ctx context.Context, priorities map[uint]int) error
	Reset(ctx context.Context) error
	GetAll(ctx context.Context) ([]*models.Site, error)
}

type siteRepository struct {
	db *gorm.DB
}

func NewSiteRepository(db *gorm.DB) SiteRepository {
	return &siteRepository{db: db}
}

func (r *siteRepository) Create(ctx context.Context, site *models.Site) error {
	return r.db.WithContext(ctx).Create(site).Error
}

func (r *siteRepository) GetByID(ctx context.Context, id uint) (*models.Site, error) {
	var site models.Site
	err := r.db.WithContext(ctx).First(&site, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("站点不存在: id=%d", id)
		}
		return nil, err
	}
	return &site, nil
}

func (r *siteRepository) Update(ctx context.Context, site *models.Site) error {
	return r.db.WithContext(ctx).Save(site).Error
}

func (r *siteRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Site{}, id).Error
}

func (r *siteRepository) List(ctx context.Context, userID uint, page, limit int) ([]*models.Site, int64, error) {
	var sites []*models.Site
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Site{})
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&sites).Error
	if err != nil {
		return nil, 0, err
	}

	return sites, total, nil
}

func (r *siteRepository) GetByUserID(ctx context.Context, userID uint) ([]*models.Site, error) {
	var sites []*models.Site
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&sites).Error
	return sites, err
}

func (r *siteRepository) GetEnabledSites(ctx context.Context, userID uint) ([]*models.Site, error) {
	var sites []*models.Site
	query := r.db.WithContext(ctx).Where("enabled = ? AND status = ?", true, "active")
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	err := query.Find(&sites).Error
	return sites, err
}

func (r *siteRepository) GetCheckinEnabledSites(ctx context.Context) ([]*models.Site, error) {
	var sites []*models.Site
	err := r.db.WithContext(ctx).
		Where("enabled = ? AND checkin_enabled = ?", true, true).
		Find(&sites).Error
	return sites, err
}

func (r *siteRepository) UpdateStatus(ctx context.Context, id uint, status, errorMsg string) error {
	return r.db.WithContext(ctx).
		Model(&models.Site{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":        status,
			"error_message": errorMsg,
		}).Error
}

func (r *siteRepository) UpdateStats(ctx context.Context, id uint, upload, download int64, ratio float64) error {
	return r.db.WithContext(ctx).
		Model(&models.Site{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"upload":   upload,
			"download": download,
			"ratio":    ratio,
		}).Error
}

// GetByDomain 根据域名获取站点
func (r *siteRepository) GetByDomain(ctx context.Context, domain string) (*models.Site, error) {
	var site models.Site
	err := r.db.WithContext(ctx).Where("url LIKE ?", "%"+domain+"%").First(&site).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("站点不存在: domain=%s", domain)
		}
		return nil, err
	}
	return &site, nil
}

// ListByPriority 按优先级排序获取站点列表
func (r *siteRepository) ListByPriority(ctx context.Context, userID uint) ([]*models.Site, error) {
	var sites []*models.Site
	query := r.db.WithContext(ctx).Order("priority DESC")
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	err := query.Find(&sites).Error
	return sites, err
}

// UpdatePriority 更新站点优先级
func (r *siteRepository) UpdatePriority(ctx context.Context, id uint, priority int) error {
	return r.db.WithContext(ctx).
		Model(&models.Site{}).
		Where("id = ?", id).
		Update("priority", priority).Error
}

// UpdatePriorities 批量更新站点优先级
func (r *siteRepository) UpdatePriorities(ctx context.Context, priorities map[uint]int) error {
	for id, priority := range priorities {
		if err := r.UpdatePriority(ctx, id, priority); err != nil {
			return err
		}
	}
	return nil
}

// Reset 重置所有站点数据
func (r *siteRepository) Reset(ctx context.Context) error {
	// 删除所有站点
	if err := r.db.WithContext(ctx).Delete(&models.Site{}).Error; err != nil {
		return err
	}
	// 删除站点相关数据（Cookie、签到记录、统计等）
	if err := r.db.WithContext(ctx).Delete(&models.SiteCookie{}).Error; err != nil {
		return err
	}
	if err := r.db.WithContext(ctx).Delete(&models.CheckinLog{}).Error; err != nil {
		return err
	}
	if err := r.db.WithContext(ctx).Delete(&models.SiteStats{}).Error; err != nil {
		return err
	}
	return nil
}

// GetAll 获取所有站点
func (r *siteRepository) GetAll(ctx context.Context) ([]*models.Site, error) {
	var sites []*models.Site
	err := r.db.WithContext(ctx).Find(&sites).Error
	return sites, err
}
