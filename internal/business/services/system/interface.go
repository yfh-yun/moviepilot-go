// Package system 系统服务
package system

import (
	"context"
)

// Service 系统服务接口
type Service interface {
	// GetSystemInfo 获取系统信息
	GetSystemInfo(ctx context.Context) (interface{}, error)
	
	// GetSystemStats 获取系统统计
	GetSystemStats(ctx context.Context) (interface{}, error)
	
	// GetSystemHealth 获取系统健康状态
	GetSystemHealth(ctx context.Context) (interface{}, error)
	
	// GetEvents 获取系统事件
	GetEvents(ctx context.Context, page, limit int) (interface{}, error)
	
	// GetTasks 获取任务列表
	GetTasks(ctx context.Context, page, limit int) (interface{}, error)
	
	// GetTaskDetail 获取任务详情
	GetTaskDetail(ctx context.Context, taskID string) (interface{}, error)
	
	// CancelTask 取消任务
	CancelTask(ctx context.Context, taskID string) error
	
	// RetryTask 重试任务
	RetryTask(ctx context.Context, taskID string) error
	
	// GetSystemConfig 获取系统配置
	GetSystemConfig(ctx context.Context) (interface{}, error)
	
	// UpdateSystemConfig 更新系统配置
	UpdateSystemConfig(ctx context.Context, config interface{}) error
	
	// GetSystemLogs 获取系统日志
	GetSystemLogs(ctx context.Context, page, limit int) (interface{}, error)
	
	// GetSystemVersion 获取系统版本信息
	GetSystemVersion(ctx context.Context) (interface{}, error)
}