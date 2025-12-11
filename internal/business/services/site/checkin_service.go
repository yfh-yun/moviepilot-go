package site

import (
	"context"
	"fmt"
	"time"

	"moviepilot-go/internal/models"
	"moviepilot-go/internal/repositories"
)

// CheckinService 签到服务接口
type CheckinService interface {
	// Checkin 执行签到
	Checkin(ctx context.Context, siteID uint) (*models.CheckinLog, error)
	// CheckinAll 签到所有启用的站点
	CheckinAll(ctx context.Context) error
	// GetCheckinLogs 获取签到日志
	GetCheckinLogs(ctx context.Context, siteID uint, page, limit int) ([]*models.CheckinLog, int64, error)
}

// checkinService 签到服务实现
type checkinService struct {
	siteRepo       repositories.SiteRepository
	checkinLogRepo repositories.CheckinLogRepository
}

// NewCheckinService 创建签到服务
func NewCheckinService(
	siteRepo repositories.SiteRepository,
	checkinLogRepo repositories.CheckinLogRepository,
) CheckinService {
	return &checkinService{
		siteRepo:       siteRepo,
		checkinLogRepo: checkinLogRepo,
	}
}

// Checkin 执行签到
func (s *checkinService) Checkin(ctx context.Context, siteID uint) (*models.CheckinLog, error) {
	// 获取站点
	site, err := s.siteRepo.GetByID(ctx, siteID)
	if err != nil {
		return nil, fmt.Errorf("站点不存在")
	}

	// 检查站点是否启用
	if !site.Enabled {
		return nil, fmt.Errorf("站点已禁用")
	}

	// 检查签到是否启用
	if !site.CheckinEnabled {
		return nil, fmt.Errorf("签到未启用")
	}

	// TODO: 实际的签到逻辑
	// 1. 使用 Cookie 访问签到URL
	// 2. 解析签到结果
	// 3. 更新站点信息

	// 创建签到日志
	log := &models.CheckinLog{
		SiteID:       siteID,
		Success:      true,
		Message:      "签到成功",
		Bonus:        10,
		ContinueDays: 1,
		CheckinTime:  time.Now(),
	}

	if err := s.checkinLogRepo.Create(ctx, log); err != nil {
		return nil, fmt.Errorf("创建签到日志失败: %w", err)
	}

	// 更新站点最后签到时间
	now := time.Now()
	site.LastCheckin = &now
	site.UpdatedAt = now

	if err := s.siteRepo.Update(ctx, site); err != nil {
		return nil, fmt.Errorf("更新站点失败: %w", err)
	}

	return log, nil
}

// CheckinAll 签到所有启用的站点
func (s *checkinService) CheckinAll(ctx context.Context) error {
	// 获取所有启用签到的站点
	sites, err := s.siteRepo.GetCheckinEnabledSites(ctx)
	if err != nil {
		return fmt.Errorf("获取站点列表失败: %w", err)
	}

	// 签到每个站点
	for _, site := range sites {
		if _, err := s.Checkin(ctx, site.ID); err != nil {
			// 记录错误但继续处理其他站点
			continue
		}
	}

	return nil
}

// GetCheckinLogs 获取签到日志
func (s *checkinService) GetCheckinLogs(ctx context.Context, siteID uint, page, limit int) ([]*models.CheckinLog, int64, error) {
	logs, total, err := s.checkinLogRepo.GetBySiteID(ctx, siteID, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("获取签到日志失败: %w", err)
	}
	return logs, total, nil
}
