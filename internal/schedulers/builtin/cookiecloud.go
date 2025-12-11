package builtin

import (
	"context"

	"go.uber.org/zap"

	"moviepilot-go/internal/business/services/site"
	"moviepilot-go/pkg/logger"
)

// CookieCloudJob CookieCloud同步任务
type CookieCloudJob struct {
	siteService site.CookieService
	logger      *zap.Logger
}

// NewCookieCloudJob 创建CookieCloud同步任务
func NewCookieCloudJob(siteService site.CookieService) *CookieCloudJob {
	return &CookieCloudJob{
		siteService: siteService,
		logger:      logger.GetLogger(),
	}
}

// ID 返回任务ID
func (j *CookieCloudJob) ID() string {
	return "cookiecloud"
}

// Name 返回任务名称
func (j *CookieCloudJob) Name() string {
	return "同步CookieCloud站点"
}

// Run 执行任务
func (j *CookieCloudJob) Run(ctx context.Context) error {
	j.logger.Info("开始执行CookieCloud同步任务")

	// 调用siteService.SyncCookieCloud(ctx)同步CookieCloud站点
	if err := j.siteService.SyncCookieCloud(ctx); err != nil {
		j.logger.Error("CookieCloud同步任务执行失败", zap.Error(err))
		return err
	}

	j.logger.Info("CookieCloud同步任务执行成功")
	return nil
}
