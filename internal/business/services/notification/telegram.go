package notification

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// TelegramNotifier Telegram 通知器
type TelegramNotifier struct {
	botToken string
	chatID   string
	client   *http.Client
	logger   *zap.Logger
}

// NewTelegramNotifier 创建 Telegram 通知器
func NewTelegramNotifier(botToken, chatID string) *TelegramNotifier {
	return &TelegramNotifier{
		botToken: botToken,
		chatID:   chatID,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger.GetLogger(),
	}
}

// Name 获取名称
func (t *TelegramNotifier) Name() string {
	return "Telegram"
}

// Send 发送通知
func (t *TelegramNotifier) Send(ctx context.Context, notification *Notification) error {
	t.logger.Info("发送 Telegram 通知",
		zap.String("title", notification.Title),
		zap.String("chat_id", t.chatID),
	)

	// 构造消息
	message := t.formatMessage(notification)

	// 发送请求
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.botToken)

	params := url.Values{}
	params.Set("chat_id", t.chatID)
	params.Set("text", message)
	params.Set("parse_mode", "HTML")

	resp, err := t.client.PostForm(apiURL, params)
	if err != nil {
		t.logger.Error("发送 Telegram 通知失败", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Telegram API 返回错误状态码: %d", resp.StatusCode)
	}

	t.logger.Info("Telegram 通知发送成功")
	return nil
}

// Test 测试连接
func (t *TelegramNotifier) Test(ctx context.Context) error {
	t.logger.Info("测试 Telegram 连接")

	testNotification := &Notification{
		Title:   "测试通知",
		Content: "这是一条测试消息",
		Type:    TypeInfo,
	}

	return t.Send(ctx, testNotification)
}

// formatMessage 格式化消息
func (t *TelegramNotifier) formatMessage(notification *Notification) string {
	// 根据类型选择图标
	icon := t.getIcon(notification.Type)

	message := fmt.Sprintf("<b>%s %s</b>\n\n%s",
		icon,
		notification.Title,
		notification.Content,
	)

	// 添加额外数据
	if len(notification.Data) > 0 {
		message += "\n\n<i>详细信息:</i>"
		for key, value := range notification.Data {
			message += fmt.Sprintf("\n• %s: %v", key, value)
		}
	}

	return message
}

// getIcon 获取类型图标
func (t *TelegramNotifier) getIcon(notifType NotificationType) string {
	icons := map[NotificationType]string{
		TypeInfo:      "ℹ️",
		TypeSuccess:   "✅",
		TypeWarning:   "⚠️",
		TypeError:     "❌",
		TypeDownload:  "⬇️",
		TypeSubscribe: "📺",
	}

	if icon, ok := icons[notifType]; ok {
		return icon
	}
	return "📢"
}
