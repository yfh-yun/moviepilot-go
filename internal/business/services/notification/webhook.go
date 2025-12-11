package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// WebhookNotifier Webhook 通知器
type WebhookNotifier struct {
	webhookURL string
	method     string
	headers    map[string]string
	client     *http.Client
	logger     *zap.Logger
}

// NewWebhookNotifier 创建 Webhook 通知器
func NewWebhookNotifier(webhookURL string) *WebhookNotifier {
	return &WebhookNotifier{
		webhookURL: webhookURL,
		method:     "POST",
		headers:    make(map[string]string),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger.GetLogger(),
	}
}

// SetMethod 设置 HTTP 方法
func (w *WebhookNotifier) SetMethod(method string) {
	w.method = method
}

// SetHeader 设置请求头
func (w *WebhookNotifier) SetHeader(key, value string) {
	w.headers[key] = value
}

// Name 获取名称
func (w *WebhookNotifier) Name() string {
	return "Webhook"
}

// Send 发送通知
func (w *WebhookNotifier) Send(ctx context.Context, notification *Notification) error {
	w.logger.Info("发送 Webhook 通知",
		zap.String("title", notification.Title),
		zap.String("url", w.webhookURL),
	)

	// 构造请求体
	payload := w.buildPayload(notification)

	jsonData, err := json.Marshal(payload)
	if err != nil {
		w.logger.Error("序列化 Webhook 数据失败", zap.Error(err))
		return err
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, w.method, w.webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		w.logger.Error("创建 Webhook 请求失败", zap.Error(err))
		return err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	for key, value := range w.headers {
		req.Header.Set(key, value)
	}

	// 发送请求
	resp, err := w.client.Do(req)
	if err != nil {
		w.logger.Error("发送 Webhook 请求失败", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Webhook 返回错误状态码: %d", resp.StatusCode)
	}

	w.logger.Info("Webhook 通知发送成功")
	return nil
}

// Test 测试连接
func (w *WebhookNotifier) Test(ctx context.Context) error {
	w.logger.Info("测试 Webhook 连接")

	testNotification := &Notification{
		Title:   "测试通知",
		Content: "这是一条测试消息",
		Type:    TypeInfo,
	}

	return w.Send(ctx, testNotification)
}

// buildPayload 构造请求体
func (w *WebhookNotifier) buildPayload(notification *Notification) map[string]any {
	payload := map[string]any{
		"title":     notification.Title,
		"content":   notification.Content,
		"type":      notification.Type,
		"priority":  notification.Priority,
		"timestamp": time.Now().Unix(),
	}

	// 合并额外数据
	if notification.Data != nil {
		for key, value := range notification.Data {
			payload[key] = value
		}
	}

	return payload
}
