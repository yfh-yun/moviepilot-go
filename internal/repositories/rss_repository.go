package repositories

import (
	"context"

	"gorm.io/gorm"

	"moviepilot-go/internal/models/entity"
)

// RSSRepository RSS订阅数据仓库接口
type RSSRepository interface {
	// Create 创建RSS订阅
	Create(ctx context.Context, feed *entity.SiteRSSFeed) error

	// Update 更新RSS订阅
	Update(ctx context.Context, feed *entity.SiteRSSFeed) error

	// Delete 删除RSS订阅
	Delete(ctx context.Context, id uint) error

	// GetByID 根据ID获取RSS订阅
	GetByID(ctx context.Context, id uint) (*entity.SiteRSSFeed, error)

	// GetBySiteID 根据站点ID获取RSS订阅列表
	GetBySiteID(ctx context.Context, siteID uint) ([]*entity.SiteRSSFeed, error)

	// GetAll 获取所有RSS订阅
	GetAll(ctx context.Context) ([]*entity.SiteRSSFeed, error)

	// GetEnabled 获取启用的RSS订阅
	GetEnabled(ctx context.Context) ([]*entity.SiteRSSFeed, error)
}

// rssRepository RSS仓库实现
type rssRepository struct {
	db *gorm.DB
}

// NewRSSRepository 创建RSS仓库
func NewRSSRepository(db *gorm.DB) RSSRepository {
	return &rssRepository{db: db}
}

// Create 创建RSS订阅
func (r *rssRepository) Create(ctx context.Context, feed *entity.SiteRSSFeed) error {
	return r.db.WithContext(ctx).Create(feed).Error
}

// Update 更新RSS订阅
func (r *rssRepository) Update(ctx context.Context, feed *entity.SiteRSSFeed) error {
	return r.db.WithContext(ctx).Save(feed).Error
}

// Delete 删除RSS订阅
func (r *rssRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.SiteRSSFeed{}, id).Error
}

// GetByID 根据ID获取RSS订阅
func (r *rssRepository) GetByID(ctx context.Context, id uint) (*entity.SiteRSSFeed, error) {
	var feed entity.SiteRSSFeed
	err := r.db.WithContext(ctx).First(&feed, id).Error
	if err != nil {
		return nil, err
	}
	return &feed, nil
}

// GetBySiteID 根据站点ID获取RSS订阅列表
func (r *rssRepository) GetBySiteID(ctx context.Context, siteID uint) ([]*entity.SiteRSSFeed, error) {
	var feeds []*entity.SiteRSSFeed
	err := r.db.WithContext(ctx).
		Where("site_id = ?", siteID).
		Order("created_at DESC").
		Find(&feeds).Error
	return feeds, err
}

// GetAll 获取所有RSS订阅
func (r *rssRepository) GetAll(ctx context.Context) ([]*entity.SiteRSSFeed, error) {
	var feeds []*entity.SiteRSSFeed
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Find(&feeds).Error
	return feeds, err
}

// GetEnabled 获取启用的RSS订阅
func (r *rssRepository) GetEnabled(ctx context.Context) ([]*entity.SiteRSSFeed, error) {
	var feeds []*entity.SiteRSSFeed
	err := r.db.WithContext(ctx).
		Where("enabled = ?", true).
		Order("created_at DESC").
		Find(&feeds).Error
	return feeds, err
}
