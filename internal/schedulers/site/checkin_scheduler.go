package site

import (
	"context"
	"fmt"

	"github.com/robfig/cron/v3"

	"moviepilot-go/internal/business/services/site"
)

// CheckinScheduler 签到调度器
type CheckinScheduler struct {
	cron           *cron.Cron
	checkinService site.CheckinService
	cronSpec       string
}

// NewCheckinScheduler 创建签到调度器
func NewCheckinScheduler(checkinService site.CheckinService, cronSpec string) *CheckinScheduler {
	return &CheckinScheduler{
		cron:           cron.New(),
		checkinService: checkinService,
		cronSpec:       cronSpec,
	}
}

// Start 启动调度器
func (s *CheckinScheduler) Start() error {
	// 添加定时任务
	_, err := s.cron.AddFunc(s.cronSpec, func() {
		ctx := context.Background()
		if err := s.checkinService.CheckinAll(ctx); err != nil {
			// 记录错误日志
			fmt.Printf("签到失败: %v\n", err)
		}
	})

	if err != nil {
		return fmt.Errorf("添加定时任务失败: %w", err)
	}

	// 启动调度器
	s.cron.Start()
	return nil
}

// Stop 停止调度器
func (s *CheckinScheduler) Stop() {
	s.cron.Stop()
}
