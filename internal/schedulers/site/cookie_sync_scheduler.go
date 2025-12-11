package site

import (
	"context"
	"fmt"

	"github.com/robfig/cron/v3"

	"moviepilot-go/internal/business/services/site"
)

// CookieSyncScheduler Cookie 同步调度器
type CookieSyncScheduler struct {
	cron          *cron.Cron
	cookieService site.CookieService
	cronSpec      string
}

// NewCookieSyncScheduler 创建 Cookie 同步调度器
func NewCookieSyncScheduler(cookieService site.CookieService, cronSpec string) *CookieSyncScheduler {
	return &CookieSyncScheduler{
		cron:          cron.New(),
		cookieService: cookieService,
		cronSpec:      cronSpec,
	}
}

// Start 启动调度器
func (s *CookieSyncScheduler) Start() error {
	// 添加定时任务
	_, err := s.cron.AddFunc(s.cronSpec, func() {
		ctx := context.Background()
		if err := s.cookieService.SyncAllCookies(ctx); err != nil {
			// 记录错误日志
			fmt.Printf("Cookie 同步失败: %v\n", err)
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
func (s *CookieSyncScheduler) Stop() {
	s.cron.Stop()
}
