package chain

import (
	"context"
	"errors"
	"net/http"

	"moviepilot-go/pkg/logger"
	"moviepilot-go/internal/models"
	"moviepilot-go/internal/repositories"
	"moviepilot-go/internal/business/services"
	"moviepilot-go/pkg/cache"
)

// WebhookChain Webhook处理链
type WebhookChain struct {
	cache          *cache.Cache
	logger         *logger.Logger
	webhookRepo    *repository.WebhookRepository
	webhookService *service.WebhookService
}

// NewWebhookChain 创建Webhook处理链实例
func NewWebhookChain(cache *cache.Cache, logger *logger.Logger, webhookRepo *repository.WebhookRepository) *WebhookChain {
	return &WebhookChain{
		cache:          cache,
		logger:         logger,
		webhookRepo:    webhookRepo,
		webhookService: service.NewWebhookService(webhookRepo, logger),
	}
}

// SendWebhook 发送Webhook
func (c *WebhookChain) SendWebhook(ctx context.Context, webhookID int64, data interface{}, eventType string) error {
	c.logger.Info("发送Webhook", "webhookID", webhookID, "eventType", eventType)

	// 查询Webhook配置
	webhook, err := c.webhookRepo.GetWebhookByID(ctx, webhookID)
	if err != nil {
		c.logger.Error("查询Webhook配置失败", "error", err)
		return err
	}

	if webhook == nil {
		return errors.New("Webhook配置不存在")
	}

	// 检查Webhook状态
	if !webhook.Enabled {
		c.logger.Warn("Webhook未启用，跳过发送", "webhookID", webhookID)
		return nil
	}

	// 检查事件类型是否匹配
	if !c.isEventTypeMatched(webhook, eventType) {
		c.logger.Debug("事件类型不匹配，跳过发送", "webhookID", webhookID, "eventType", eventType)
		return nil
	}

	// 发送Webhook
	err = c.webhookService.SendWebhook(ctx, webhook, data, eventType)
	if err != nil {
		c.logger.Error("发送Webhook失败", "error", err)
		return err
	}

	c.logger.Info("Webhook发送成功", "webhookID", webhookID)
	return nil
}

// isEventTypeMatched 检查事件类型是否匹配
func (c *WebhookChain) isEventTypeMatched(webhook *model.Webhook, eventType string) bool {
	// 如果未指定事件类型，则全部匹配
	if webhook.EventTypes == "" || webhook.EventTypes == "*" {
		return true
	}

	// 检查事件类型是否在配置中
	// 这里可以实现更复杂的事件类型匹配逻辑
	return webhook.EventTypes == eventType
}

// ProcessWebhookEvent 处理Webhook事件
func (c *WebhookChain) ProcessWebhookEvent(ctx context.Context, event model.WebhookEvent) error {
	c.logger.Info("处理Webhook事件", "eventType", event.EventType)

	// 获取所有启用的Webhook
	webhooks, err := c.webhookRepo.GetEnabledWebhooks(ctx)
	if err != nil {
		c.logger.Error("获取启用的Webhook列表失败", "error", err)
		return err
	}

	// 发送到所有匹配的Webhook
	for _, webhook := range webhooks {
		if c.isEventTypeMatched(webhook, event.EventType) {
			err := c.SendWebhook(ctx, webhook.ID, event.Data, event.EventType)
			if err != nil {
				c.logger.Error("发送Webhook失败", "webhookID", webhook.ID, "error", err)
				// 继续发送其他Webhook
			}
		}
	}

	return nil
}

// CreateWebhook 创建Webhook配置
func (c *WebhookChain) CreateWebhook(ctx context.Context, webhookData model.WebhookCreateData) (*model.Webhook, error) {
	c.logger.Info("创建Webhook配置", "name", webhookData.Name)

	webhook, err := c.webhookService.CreateWebhook(ctx, webhookData)
	if err != nil {
		c.logger.Error("创建Webhook配置失败", "error", err)
		return nil, err
	}

	c.logger.Info("创建Webhook配置成功", "webhookID", webhook.ID)
	return webhook, nil
}

// UpdateWebhook 更新Webhook配置
func (c *WebhookChain) UpdateWebhook(ctx context.Context, webhookID int64, updateData model.WebhookUpdateData) (*model.Webhook, error) {
	c.logger.Info("更新Webhook配置", "webhookID", webhookID)

	webhook, err := c.webhookService.UpdateWebhook(ctx, webhookID, updateData)
	if err != nil {
		c.logger.Error("更新Webhook配置失败", "error", err)
		return nil, err
	}

	c.logger.Info("更新Webhook配置成功", "webhookID", webhookID)
	return webhook, nil
}

// DeleteWebhook 删除Webhook配置
func (c *WebhookChain) DeleteWebhook(ctx context.Context, webhookID int64) error {
	c.logger.Info("删除Webhook配置", "webhookID", webhookID)

	err := c.webhookService.DeleteWebhook(ctx, webhookID)
	if err != nil {
		c.logger.Error("删除Webhook配置失败", "error", err)
		return err
	}

	c.logger.Info("删除Webhook配置成功", "webhookID", webhookID)
	return nil
}

// GetWebhookList 获取Webhook配置列表
func (c *WebhookChain) GetWebhookList(ctx context.Context, page, pageSize int) ([]*model.Webhook, int64, error) {
	c.logger.Info("获取Webhook配置列表", "page", page, "pageSize", pageSize)

	webhooks, total, err := c.webhookRepo.GetWebhookList(ctx, page, pageSize)
	if err != nil {
		c.logger.Error("获取Webhook配置列表失败", "error", err)
		return nil, 0, err
	}

	c.logger.Info("获取Webhook配置列表成功", "count", len(webhooks))
	return webhooks, total, nil
}

// EnableWebhook 启用Webhook
func (c *WebhookChain) EnableWebhook(ctx context.Context, webhookID int64) error {
	c.logger.Info("启用Webhook", "webhookID", webhookID)

	err := c.webhookService.EnableWebhook(ctx, webhookID)
	if err != nil {
		c.logger.Error("启用Webhook失败", "error", err)
		return err
	}

	c.logger.Info("启用Webhook成功", "webhookID", webhookID)
	return nil
}

// DisableWebhook 禁用Webhook
func (c *WebhookChain) DisableWebhook(ctx context.Context, webhookID int64) error {
	c.logger.Info("禁用Webhook", "webhookID", webhookID)

	err := c.webhookService.DisableWebhook(ctx, webhookID)
	if err != nil {
		c.logger.Error("禁用Webhook失败", "error", err)
		return err
	}

	c.logger.Info("禁用Webhook成功", "webhookID", webhookID)
	return nil
}

// TestWebhook 测试Webhook
func (c *WebhookChain) TestWebhook(ctx context.Context, webhookID int64, testData interface{}) error {
	c.logger.Info("测试Webhook", "webhookID", webhookID)

	err := c.webhookService.TestWebhook(ctx, webhookID, testData)
	if err != nil {
		c.logger.Error("测试Webhook失败", "error", err)
		return err
	}

	c.logger.Info("测试Webhook成功", "webhookID", webhookID)
	return nil
}

// GetWebhookStats 获取Webhook统计信息
func (c *WebhookChain) GetWebhookStats(ctx context.Context) (*model.WebhookStats, error) {
	c.logger.Info("获取Webhook统计信息")

	stats, err := c.webhookRepo.GetWebhookStats(ctx)
	if err != nil {
		c.logger.Error("获取Webhook统计信息失败", "error", err)
		return nil, err
	}

	return stats, nil
}

// ProcessWebhookResponse 处理Webhook响应
func (c *WebhookChain) ProcessWebhookResponse(ctx context.Context, response *http.Response) error {
	c.logger.Info("处理Webhook响应", "statusCode", response.StatusCode)

	// 记录响应日志
	if response.StatusCode >= 200 && response.StatusCode < 300 {
		c.logger.Info("Webhook响应成功", "statusCode", response.StatusCode)
	} else {
		c.logger.Warn("Webhook响应失败", "statusCode", response.StatusCode)
	}

	return nil
}
