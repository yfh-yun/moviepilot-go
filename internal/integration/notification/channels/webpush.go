package channels

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/yfh-yun/moviepilot-go/internal/integration/notification"
)

// WebPushConfig Web Push通知配置
type WebPushConfig struct {
	VAPIDPrivateKey string `json:"vapid_private_key"` // VAPID私钥
	VAPIDPublicKey  string `json:"vapid_public_key"`  // VAPID公钥
	Subject         string `json:"subject"`           // 订阅主题
	TTL             int    `json:"ttl"`               // 消息生存时间（秒）
	Endpoint        string `json:"endpoint"`          // 推送端点（可选）
}

// Subscription Web Push订阅信息
type Subscription struct {
	Endpoint string            `json:"endpoint"`
	Keys     map[string]string `json:"keys"`
}

// WebPushSender Web Push通知发送器
type WebPushSender struct {
	config          *WebPushConfig
	httpClient      *http.Client
	vapidPrivateKey *ecdsa.PrivateKey
	vapidPublicKey  *ecdsa.PublicKey
	subscriptions   map[string]*Subscription // 存储订阅信息
}

// NewWebPushSender 创建新的Web Push通知发送器
func NewWebPushSender(config *WebPushConfig) (*WebPushSender, error) {
	sender := &WebPushSender{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		subscriptions: make(map[string]*Subscription),
	}

	// 解析VAPID密钥
	if err := sender.parseVAPIDKeys(); err != nil {
		return nil, fmt.Errorf("failed to parse VAPID keys: %w", err)
	}

	return sender, nil
}

// Name 返回发送器名称
func (w *WebPushSender) Name() string {
	return "webpush"
}

// SupportedLevels 返回支持的通知级别
func (w *WebPushSender) SupportedLevels() []notification.NotificationLevel {
	return []notification.NotificationLevel{
		notification.LevelInfo,
		notification.LevelWarning,
		notification.LevelError,
		notification.LevelSuccess,
	}
}

// parseVAPIDKeys 解析VAPID密钥
func (w *WebPushSender) parseVAPIDKeys() error {
	if w.config.VAPIDPrivateKey != "" && w.config.VAPIDPublicKey != "" {
		// 从配置中加载密钥
		privateKeyBytes, err := base64.RawURLEncoding.DecodeString(w.config.VAPIDPrivateKey)
		if err != nil {
			return fmt.Errorf("invalid VAPID private key: %w", err)
		}

		publicKeyBytes, err := base64.RawURLEncoding.DecodeString(w.config.VAPIDPublicKey)
		if err != nil {
			return fmt.Errorf("invalid VAPID public key: %w", err)
		}

		x, y := elliptic.Unmarshal(elliptic.P256(), publicKeyBytes)
		if x == nil {
			return fmt.Errorf("invalid VAPID public key format")
		}

		w.vapidPublicKey = &ecdsa.PublicKey{
			Curve: elliptic.P256(),
			X:     x,
			Y:     y,
		}

		w.vapidPrivateKey = &ecdsa.PrivateKey{
			PublicKey: *w.vapidPublicKey,
			D:         new(ecdsa.PrivateKey).SetBytes(privateKeyBytes).D,
		}
	} else {
		// 生成新的VAPID密钥对
		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return fmt.Errorf("failed to generate VAPID keys: %w", err)
		}

		w.vapidPrivateKey = privateKey
		w.vapidPublicKey = &privateKey.PublicKey

		// 更新配置中的密钥
		w.config.VAPIDPrivateKey = base64.RawURLEncoding.EncodeToString(privateKey.D.Bytes())
		w.config.VAPIDPublicKey = base64.RawURLEncoding.EncodeToString(
			elliptic.Marshal(elliptic.P256(), privateKey.PublicKey.X, privateKey.PublicKey.Y),
		)
	}

	return nil
}

// Validate 验证消息
func (w *WebPushSender) Validate(message *notification.Message) error {
	if message.Title == "" {
		return fmt.Errorf("web push notification title cannot be empty")
	}

	if len(message.Content) > 500 {
		return fmt.Errorf("web push notification content too long (max 500 characters)")
	}

	return nil
}

// Send 发送Web Push通知
func (w *WebPushSender) Send(ctx context.Context, message *notification.Message) error {
	if err := w.Validate(message); err != nil {
		return err
	}

	// 构建推送消息
	pushMessage := w.buildPushMessage(message)

	// 发送到所有订阅
	var errors []string
	for _, subscription := range w.subscriptions {
		err := w.sendToSubscription(ctx, subscription, pushMessage)
		if err != nil {
			errors = append(errors, fmt.Sprintf("subscription %s: %v", subscription.Endpoint, err))
		}
	}

	// 如果指定了特定端点，也发送到该端点
	if w.config.Endpoint != "" {
		subscription := &Subscription{
			Endpoint: w.config.Endpoint,
			Keys: map[string]string{
				"p256dh": "",
				"auth":   "",
			},
		}
		err := w.sendToSubscription(ctx, subscription, pushMessage)
		if err != nil {
			errors = append(errors, fmt.Sprintf("endpoint %s: %v", w.config.Endpoint, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to send web push notifications: %s", strings.Join(errors, "; "))
	}

	return nil
}

// buildPushMessage 构建推送消息
func (w *WebPushSender) buildPushMessage(message *notification.Message) []byte {
	// 构建推送数据
	pushData := map[string]interface{}{
		"title":     message.Title,
		"body":      message.Content,
		"icon":      message.ImageURL,
		"url":       message.LinkURL,
		"timestamp": time.Now().Unix(),
		"level":     message.Level.String(),
	}

	// 根据级别添加图标和颜色
	switch message.Level {
	case notification.LevelError:
		pushData["badge"] = "/icons/error-badge.png"
		pushData["color"] = "#e74c3c"
	case notification.LevelWarning:
		pushData["badge"] = "/icons/warning-badge.png"
		pushData["color"] = "#f39c12"
	case notification.LevelSuccess:
		pushData["badge"] = "/icons/success-badge.png"
		pushData["color"] = "#27ae60"
	case notification.LevelInfo:
		pushData["badge"] = "/icons/info-badge.png"
		pushData["color"] = "#3498db"
	}

	// 转换为JSON
	jsonData, _ := json.Marshal(pushData)
	return jsonData
}

// sendToSubscription 发送到特定订阅
func (w *WebPushSender) sendToSubscription(ctx context.Context, subscription *Subscription, message []byte) error {
	// 创建Web Push选项
	options := &webpush.Options{
		Subscriber:      w.config.Subject,
		TTL:             w.config.TTL,
		VAPIDPublicKey:  w.config.VAPIDPublicKey,
		VAPIDPrivateKey: w.config.VAPIDPrivateKey,
	}

	// 转换为webpush-go的订阅格式
	webpushSubscription := &webpush.Subscription{
		Endpoint: subscription.Endpoint,
		Keys: webpush.Keys{
			P256dh: subscription.Keys["p256dh"],
			Auth:   subscription.Keys["auth"],
		},
	}

	// 发送推送
	resp, err := webpush.SendNotification(message, webpushSubscription, options)
	if err != nil {
		return fmt.Errorf("failed to send web push: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("web push failed with status: %s", resp.Status)
	}

	return nil
}

// AddSubscription 添加订阅
func (w *WebPushSender) AddSubscription(subscription *Subscription) {
	w.subscriptions[subscription.Endpoint] = subscription
}

// RemoveSubscription 移除订阅
func (w *WebPushSender) RemoveSubscription(endpoint string) {
	delete(w.subscriptions, endpoint)
}

// GetVAPIDKeys 获取VAPID密钥
func (w *WebPushSender) GetVAPIDKeys() (string, string) {
	return w.config.VAPIDPublicKey, w.config.VAPIDPrivateKey
}

// HealthCheck 健康检查
func (w *WebPushSender) HealthCheck(ctx context.Context) error {
	// 检查是否有有效的订阅
	if len(w.subscriptions) == 0 && w.config.Endpoint == "" {
		return fmt.Errorf("no active subscriptions configured")
	}

	// 检查VAPID密钥
	if w.vapidPrivateKey == nil || w.vapidPublicKey == nil {
		return fmt.Errorf("VAPID keys are not properly configured")
	}

	return nil
}

// Close 关闭发送器
func (w *WebPushSender) Close() error {
	w.subscriptions = make(map[string]*Subscription)
	return nil
}

// CreateWebPushChannel 创建Web Push通知渠道
func CreateWebPushChannel(config *WebPushConfig) (*notification.Channel, error) {
	sender, err := NewWebPushSender(config)
	if err != nil {
		return nil, err
	}

	enabled := config.VAPIDPublicKey != "" && config.VAPIDPrivateKey != ""

	return &notification.Channel{
		Name:        "webpush",
		Description: "Web Push通知渠道",
		Enabled:     enabled,
		Sender:      sender,
		Config: map[string]string{
			"vapid_public_key": config.VAPIDPublicKey,
			"subject":          config.Subject,
			"ttl":              fmt.Sprintf("%d", config.TTL),
		},
	}, nil
}
