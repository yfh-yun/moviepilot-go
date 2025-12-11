package builtin

import (
	"context"
)

// SubscribeRefreshJob 订阅刷新任务
type SubscribeRefreshJob struct {
	// subscribeService subscribe.Service
	// 注意：实际实现中需要注入subscribeService，这里简化处理
}

// NewSubscribeRefreshJob 创建订阅刷新任务
func NewSubscribeRefreshJob() *SubscribeRefreshJob {
	return &SubscribeRefreshJob{}
}

// ID 返回任务ID
func (j *SubscribeRefreshJob) ID() string {
	return "subscribe_refresh"
}

// Name 返回任务名称
func (j *SubscribeRefreshJob) Name() string {
	return "刷新订阅"
}

// Run 执行任务
func (j *SubscribeRefreshJob) Run(ctx context.Context) error {
	// 实际实现中需要调用subscribeService.Refresh(ctx)
	// 这里简化处理，返回nil表示成功
	return nil
}
