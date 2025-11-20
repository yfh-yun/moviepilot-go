// Package webhook Webhook服务
package webhook

import (
	"context"
)

// Service Webhook服务接口
type Service interface {
	// ProcessWebhook 处理webhook请求
	ProcessWebhook(ctx context.Context, eventType string, payload interface{}) error
	
	// GetWebhooks 获取webhook列表
	GetWebhooks(ctx context.Context) ([]interface{}, error)
	
	// GetWebhookByID 根据ID获取webhook
	GetWebhookByID(ctx context.Context, id string) (interface{}, error)
	
	// CreateWebhook 创建webhook
	CreateWebhook(ctx context.Context, webhook interface{}) error
	
	// UpdateWebhook 更新webhook
	UpdateWebhook(ctx context.Context, id string, webhook interface{}) error
	
	// DeleteWebhook 删除webhook
	DeleteWebhook(ctx context.Context, id string) error
	
	// TestWebhook 测试webhook
	TestWebhook(ctx context.Context, id string) error
	
	// GetWebhookLogs 获取webhook日志
	GetWebhookLogs(ctx context.Context, id string, page, limit int) ([]interface{}, error)
	
	// GetWebhookStats 获取webhook统计
	GetWebhookStats(ctx context.Context, id string) (interface{}, error)
}