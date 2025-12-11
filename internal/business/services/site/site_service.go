package site

import (
	"context"
	"fmt"
	"time"

	"moviepilot-go/internal/models"
	"moviepilot-go/internal/repositories"
)

// CreateSiteRequest 创建站点请求
type CreateSiteRequest struct {
	Name           string `json:"name" binding:"required"`
	URL            string `json:"url" binding:"required,url"`
	Type           string `json:"type" binding:"required,oneof=pt public rss"`
	Priority       int    `json:"priority"`
	Cookie         string `json:"cookie"`
	UserAgent      string `json:"user_agent"`
	Proxy          string `json:"proxy"`
	CheckinEnabled bool   `json:"checkin_enabled"`
	CheckinCron    string `json:"checkin_cron"`
	CheckinURL     string `json:"checkin_url"`
}

// UpdateSiteRequest 更新站点请求
type UpdateSiteRequest struct {
	Name           *string `json:"name"`
	URL            *string `json:"url"`
	Type           *string `json:"type"`
	Priority       *int    `json:"priority"`
	Enabled        *bool   `json:"enabled"`
	Cookie         *string `json:"cookie"`
	UserAgent      *string `json:"user_agent"`
	Proxy          *string `json:"proxy"`
	CheckinEnabled *bool   `json:"checkin_enabled"`
	CheckinCron    *string `json:"checkin_cron"`
	CheckinURL     *string `json:"checkin_url"`
}

// SiteService 站点服务接口
type SiteService interface {
	// Create 创建站点
	Create(ctx context.Context, userID uint, req *CreateSiteRequest) (*models.Site, error)
	// GetByID 获取站点
	GetByID(ctx context.Context, id uint) (*models.Site, error)
	// GetByDomain 根据域名获取站点
	GetByDomain(ctx context.Context, domain string) (*models.Site, error)
	// Update 更新站点
	Update(ctx context.Context, id uint, req *UpdateSiteRequest) (*models.Site, error)
	// Delete 删除站点
	Delete(ctx context.Context, id uint) error
	// List 获取站点列表
	List(ctx context.Context, userID uint, page, limit int) ([]*models.Site, int64, error)
	// ListByPriority 按优先级获取站点列表
	ListByPriority(ctx context.Context, userID uint) ([]*models.Site, error)
	// GetEnabledSites 获取启用的站点
	GetEnabledSites(ctx context.Context, userID uint) ([]*models.Site, error)
	// UpdatePriority 更新站点优先级
	UpdatePriority(ctx context.Context, id uint, priority int) error
	// UpdatePriorities 批量更新站点优先级
	UpdatePriorities(ctx context.Context, priorities map[uint]int) error
	// Reset 重置所有站点
	Reset(ctx context.Context) error
	// ValidateCookie 验证 Cookie
	ValidateCookie(ctx context.Context, id uint) (bool, error)
	// UpdateStats 更新站点统计
	UpdateStats(ctx context.Context, id uint, upload, download int64) error
	// GetAllSites 获取所有站点
	GetAllSites(ctx context.Context) ([]*models.Site, error)
}

// siteService 站点服务实现
type siteService struct {
	siteRepo repositories.SiteRepository
}

// NewSiteService 创建站点服务
func NewSiteService(siteRepo repositories.SiteRepository) SiteService {
	return &siteService{
		siteRepo: siteRepo,
	}
}

// Create 创建站点
func (s *siteService) Create(ctx context.Context, userID uint, req *CreateSiteRequest) (*models.Site, error) {
	// 创建站点
	site := &models.Site{
		UserID:         userID,
		Name:           req.Name,
		URL:            req.URL,
		Type:           req.Type,
		Priority:       req.Priority,
		Enabled:        true,
		Cookie:         req.Cookie,
		UserAgent:      req.UserAgent,
		Proxy:          req.Proxy,
		CheckinEnabled: req.CheckinEnabled,
		CheckinCron:    req.CheckinCron,
		CheckinURL:     req.CheckinURL,
		Status:         "active",
	}

	// 设置默认值
	if site.Priority == 0 {
		site.Priority = 5
	}
	if site.CheckinCron == "" {
		site.CheckinCron = "0 8 * * *" // 默认每天 8:00
	}

	if err := s.siteRepo.Create(ctx, site); err != nil {
		return nil, fmt.Errorf("创建站点失败: %w", err)
	}

	return site, nil
}

// GetByID 获取站点
func (s *siteService) GetByID(ctx context.Context, id uint) (*models.Site, error) {
	site, err := s.siteRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取站点失败: %w", err)
	}
	return site, nil
}

// Update 更新站点
func (s *siteService) Update(ctx context.Context, id uint, req *UpdateSiteRequest) (*models.Site, error) {
	// 获取站点
	site, err := s.siteRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("站点不存在")
	}

	// 更新字段
	if req.Name != nil {
		site.Name = *req.Name
	}
	if req.URL != nil {
		site.URL = *req.URL
	}
	if req.Type != nil {
		site.Type = *req.Type
	}
	if req.Priority != nil {
		site.Priority = *req.Priority
	}
	if req.Enabled != nil {
		site.Enabled = *req.Enabled
	}
	if req.Cookie != nil {
		site.Cookie = *req.Cookie
	}
	if req.UserAgent != nil {
		site.UserAgent = *req.UserAgent
	}
	if req.Proxy != nil {
		site.Proxy = *req.Proxy
	}
	if req.CheckinEnabled != nil {
		site.CheckinEnabled = *req.CheckinEnabled
	}
	if req.CheckinCron != nil {
		site.CheckinCron = *req.CheckinCron
	}
	if req.CheckinURL != nil {
		site.CheckinURL = *req.CheckinURL
	}

	site.UpdatedAt = time.Now()

	if err := s.siteRepo.Update(ctx, site); err != nil {
		return nil, fmt.Errorf("更新站点失败: %w", err)
	}

	return site, nil
}

// Delete 删除站点
func (s *siteService) Delete(ctx context.Context, id uint) error {
	if err := s.siteRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("删除站点失败: %w", err)
	}
	return nil
}

// List 获取站点列表
func (s *siteService) List(ctx context.Context, userID uint, page, limit int) ([]*models.Site, int64, error) {
	sites, total, err := s.siteRepo.List(ctx, userID, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("获取站点列表失败: %w", err)
	}
	return sites, total, nil
}

// GetEnabledSites 获取启用的站点
func (s *siteService) GetEnabledSites(ctx context.Context, userID uint) ([]*models.Site, error) {
	sites, err := s.siteRepo.GetEnabledSites(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("获取启用站点失败: %w", err)
	}
	return sites, nil
}

// ValidateCookie 验证 Cookie
func (s *siteService) ValidateCookie(ctx context.Context, id uint) (bool, error) {
	// 获取站点
	site, err := s.siteRepo.GetByID(ctx, id)
	if err != nil {
		return false, fmt.Errorf("站点不存在")
	}

	// 这里应该实际访问站点验证 Cookie
	// 简化实现：检查 Cookie 是否为空
	if site.Cookie == "" {
		return false, nil
	}

	// TODO: 实际的 Cookie 验证逻辑
	// 1. 使用 Cookie 访问站点
	// 2. 检查响应状态
	// 3. 更新站点状态

	return true, nil
}

// UpdateStats 更新站点统计
func (s *siteService) UpdateStats(ctx context.Context, id uint, upload, download int64) error {
	// 计算分享率
	var ratio float64
	if download > 0 {
		ratio = float64(upload) / float64(download)
	}

	if err := s.siteRepo.UpdateStats(ctx, id, upload, download, ratio); err != nil {
		return fmt.Errorf("更新站点统计失败: %w", err)
	}

	return nil
}

// GetByDomain 根据域名获取站点
func (s *siteService) GetByDomain(ctx context.Context, domain string) (*models.Site, error) {
	site, err := s.siteRepo.GetByDomain(ctx, domain)
	if err != nil {
		return nil, fmt.Errorf("获取站点失败: %w", err)
	}
	return site, nil
}

// ListByPriority 按优先级获取站点列表
func (s *siteService) ListByPriority(ctx context.Context, userID uint) ([]*models.Site, error) {
	sites, err := s.siteRepo.ListByPriority(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("获取站点列表失败: %w", err)
	}
	return sites, nil
}

// UpdatePriority 更新站点优先级
func (s *siteService) UpdatePriority(ctx context.Context, id uint, priority int) error {
	if err := s.siteRepo.UpdatePriority(ctx, id, priority); err != nil {
		return fmt.Errorf("更新站点优先级失败: %w", err)
	}
	return nil
}

// UpdatePriorities 批量更新站点优先级
func (s *siteService) UpdatePriorities(ctx context.Context, priorities map[uint]int) error {
	if err := s.siteRepo.UpdatePriorities(ctx, priorities); err != nil {
		return fmt.Errorf("批量更新站点优先级失败: %w", err)
	}
	return nil
}

// Reset 重置所有站点
func (s *siteService) Reset(ctx context.Context) error {
	if err := s.siteRepo.Reset(ctx); err != nil {
		return fmt.Errorf("重置站点失败: %w", err)
	}
	return nil
}

// GetAllSites 获取所有站点
func (s *siteService) GetAllSites(ctx context.Context) ([]*models.Site, error) {
	sites, err := s.siteRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取所有站点失败: %w", err)
	}
	return sites, nil
}
