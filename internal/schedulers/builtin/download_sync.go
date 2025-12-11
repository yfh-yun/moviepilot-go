package builtin

import (
	"context"

	"go.uber.org/zap"

	"moviepilot-go/internal/business/services/download"
	"moviepilot-go/pkg/logger"
)

// DownloadSyncJob 下载状态同步任务
type DownloadSyncJob struct {
	downloadService *download.DownloadService
	logger          *zap.Logger
}

// NewDownloadSyncJob 创建下载状态同步任务
func NewDownloadSyncJob(downloadService *download.DownloadService) *DownloadSyncJob {
	return &DownloadSyncJob{
		downloadService: downloadService,
		logger:          logger.GetLogger(),
	}
}

// ID 返回任务ID
func (j *DownloadSyncJob) ID() string {
	return "download_sync"
}

// Name 返回任务名称
func (j *DownloadSyncJob) Name() string {
	return "同步下载任务状态"
}

// Run 执行任务
func (j *DownloadSyncJob) Run(ctx context.Context) error {
	j.logger.Info("开始执行下载状态同步任务")

	// 调用downloadService同步所有下载任务的状态
	// TODO: 实现具体的同步逻辑
	// if err := j.downloadService.SyncAllTasks(ctx); err != nil {
	// 	j.logger.Error("下载状态同步任务执行失败", zap.Error(err))
	// 	return err
	// }

	j.logger.Info("下载状态同步任务执行成功")
	return nil
}
