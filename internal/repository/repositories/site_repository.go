package repositories

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/yfh-yun/moviepilot-go/internal/database"
	"github.com/yfh-yun/moviepilot-go/internal/repository/interfaces"
	"github.com/yfh-yun/moviepilot-go/internal/model"
)

// SiteRepository 站点仓储实现
type SiteRepository struct{}

func NewSiteRepository() interfaces.SiteRepository {
	return &model.SiteRepository{}
}

func (r *SiteRepository) Create(site *model.Site) error {
	return database.DB.Create(site).Error
}

func (r *SiteRepository) GetByID(id uint) (*model.Site, error) {
	var site model.Site
	err := database.DB.First(&site, id).Error
	if err != nil {
		return nil, err
	}
	return &site, nil
}

func (r *SiteRepository) GetByName(name string) (*model.Site, error) {
	var site model.Site
	err := database.DB.Where("name = ?", name).First(&site).Error
	if err != nil {
		return nil, err
	}
	return &site, nil
}

func (r *SiteRepository) GetByDomain(domain string) (*model.Site, error) {
	var site model.Site
	err := database.DB.Where("domain = ?", domain).First(&site).Error
	if err != nil {
		return nil, err
	}
	return &site, nil
}

func (r *SiteRepository) GetActive() ([]*model.Site, error) {
	var sites []*model.Site
	err := database.DB.Where("is_active = ?", true).Find(&sites).Error
	return sites, err
}

func (r *SiteRepository) GetRSSSites() ([]*model.Site, error) {
	var sites []*model.Site
	err := database.DB.Where("is_rss = ? AND is_active = ?", true, true).Find(&sites).Error
	return sites, err
}

func (r *SiteRepository) GetSearchSites() ([]*model.Site, error) {
	var sites []*model.Site
	err := database.DB.Where("is_active = ?", true).Find(&sites).Error
	return sites, err
}

func (r *SiteRepository) Update(site *model.Site) error {
	return database.DB.Save(site).Error
}

func (r *SiteRepository) Delete(id uint) error {
	return database.DB.Delete(&model.Site{}, id).Error
}

func (r *SiteRepository) List(offset, limit int) ([]*model.Site, int64, error) {
	var sites []*model.Site
	var total int64

	err := database.DB.Model(&model.Site{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = database.DB.Offset(offset).Limit(limit).Order("priority DESC, created_at DESC").Find(&sites).Error
	return sites, total, err
}

func (r *SiteRepository) Search(keyword string, offset, limit int) ([]*model.Site, int64, error) {
	var sites []*model.Site
	var total int64

	query := database.DB.Model(&model.Site{})
	if keyword != "" {
		query = query.Where("name LIKE ? OR domain LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Offset(offset).Limit(limit).Order("priority DESC, created_at DESC").Find(&sites).Error
	return sites, total, err
}

func (r *SiteRepository) Count() (int64, error) {
	var total int64
	err := database.DB.Model(&model.Site{}).Count(&total).Error
	return total, err
}

// ==================== 异步查询方法 ====================

func (r *SiteRepository) GetActiveAsync(ctx context.Context) ([]*model.Site, error) {
	type result struct {
		sites []*model.Site
		err   error
	}

	ch := make(chan result, 1)

	go func() {
		sites, err := r.GetActive()
		ch <- result{sites: sites, err: err}
	}()

	select {
	case res := <-ch:
		return res.sites, res.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (r *SiteRepository) GetByDomainAsync(ctx context.Context, domain string) (*model.Site, error) {
	type result struct {
		site *model.Site
		err  error
	}

	ch := make(chan result, 1)

	go func() {
		site, err := r.GetByDomain(domain)
		ch <- result{site: site, err: err}
	}()

	select {
	case res := <-ch:
		return res.site, res.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (r *SiteRepository) ListAsync(ctx context.Context, offset, limit int) ([]*model.Site, int64, error) {
	type result struct {
		sites []*model.Site
		total int64
		err   error
	}

	ch := make(chan result, 1)

	go func() {
		sites, total, err := r.List(offset, limit)
		ch <- result{sites: sites, total: total, err: err}
	}()

	select {
	case res := <-ch:
		return res.sites, res.total, res.err
	case <-ctx.Done():
		return nil, 0, ctx.Err()
	}
}

// ==================== 高级查询方法 ====================

func (r *SiteRepository) ListOrderByPriority() ([]*model.Site, error) {
	var sites []*model.Site
	err := database.DB.Order("priority DESC, created_at ASC").Find(&sites).Error
	return sites, err
}

func (r *SiteRepository) ListOrderByPriorityAsync(ctx context.Context) ([]*model.Site, error) {
	type result struct {
		sites []*model.Site
		err   error
	}

	ch := make(chan result, 1)

	go func() {
		sites, err := r.ListOrderByPriority()
		ch <- result{sites: sites, err: err}
	}()

	select {
	case res := <-ch:
		return res.sites, res.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (r *SiteRepository) GetDomainsByIDs(ids []uint) ([]string, error) {
	if len(ids) == 0 {
		return []string{}, nil
	}

	var domains []string
	err := database.DB.Model(&model.Site{}).Where("id IN ?", ids).Pluck("domain", &domains).Error
	return domains, err
}

func (r *SiteRepository) Exists(domain string) (bool, error) {
	var count int64
	err := database.DB.Model(&model.Site{}).Where("domain = ?", domain).Count(&count).Error
	return count > 0, err
}

// ==================== 批量操作 ====================

func (r *SiteRepository) BatchCreate(sites []*model.Site) error {
	if len(sites) == 0 {
		return nil
	}

	batchSize := 50
	return database.DB.Transaction(func(tx *gorm.DB) error {
		for i := 0; i < len(sites); i += batchSize {
			end := i + batchSize
			if end > len(sites) {
				end = len(sites)
			}

			batch := sites[i:end]
			if err := tx.CreateInBatches(batch, batchSize).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *SiteRepository) BatchUpdate(sites []*model.Site) error {
	if len(sites) == 0 {
		return nil
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		for _, site := range sites {
			if err := tx.Save(site).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *SiteRepository) BatchDelete(ids []uint) error {
	if len(ids) == 0 {
		return nil
	}

	result := database.DB.Where("id IN ?", ids).Delete(&model.Site{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no sites found to delete")
	}
	return nil
}

// ==================== 统计方法 ====================

func (r *SiteRepository) CountByStatus(isActive bool) (int64, error) {
	var count int64
	err := database.DB.Model(&model.Site{}).Where("is_active = ?", isActive).Count(&count).Error
	return count, err
}

func (r *SiteRepository) GetFailCountThreshold(threshold int) ([]*model.Site, error) {
	var sites []*model.Site
	err := database.DB.Where("fail_count >= ?", threshold).Find(&sites).Error
	return sites, err
}

// ==================== Cookie和认证方法 ====================

func (r *SiteRepository) UpdateCookie(domain, cookie string) error {
	result := database.DB.Model(&model.Site{}).Where("domain = ?", domain).Update("cookie", cookie)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("site with domain %s not found", domain)
	}
	return nil
}

func (r *SiteRepository) UpdateFailCount(domain string, increment bool) error {
	var operation string
	if increment {
		operation = "fail_count + 1"
	} else {
		operation = "GREATEST(fail_count - 1, 0)"
	}

	result := database.DB.Model(&model.Site{}).Where("domain = ?", domain).Update("fail_count", gorm.Expr(operation))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("site with domain %s not found", domain)
	}
	return nil
}

func (r *SiteRepository) ResetFailCount(domain string) error {
	result := database.DB.Model(&model.Site{}).Where("domain = ?", domain).Update("fail_count", 0)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("site with domain %s not found", domain)
	}
	return nil
}
