package builtin

import (
	"context"
)

// MediaServerSyncJob 媒体服务器同步任务
type MediaServerSyncJob struct {
	// mediaServerService mediaserver.Service
	// 注意：实际实现中需要注入mediaServerService，这里简化处理
}

// NewMediaServerSyncJob 创建媒体服务器同步任务
func NewMediaServerSyncJob() *MediaServerSyncJob {
	return &MediaServerSyncJob{}
}

// ID 返回任务ID
func (j *MediaServerSyncJob) ID() string {
	return "mediaserver_sync"
}

// Name 返回任务名称
func (j *MediaServerSyncJob) Name() string {
	return "同步媒体服务器"
}

// Run 执行任务
func (j *MediaServerSyncJob) Run(ctx context.Context) error {
	// 实际实现中需要调用mediaServerService.Sync(ctx)
	// 这里简化处理，返回nil表示成功
	return nil
}
