package webpush

import (
	"encoding/json"
	"fmt"
	"strings"

	"moviepilot-go/internal/config"
	"moviepilot-go/pkg/models"
	"moviepilot-go/pkg/modules"
)

// WebPushModule WebPush模块
type WebPushModule struct {
	*modules.ModuleBase
	*modules.MessageBase
	client *WebPushClient
}

// NewWebPushModule 创建新的WebPush模块实例
func NewWebPushModule() *WebPushModule {
	// 初始化VAPID配置
	vapidConfig := VAPIDConfig{
		PrivateKey: config.Settings.VAPID.PrivateKey,
		Subject:    config.Settings.VAPID.Subject,
		PublicKey:  config.Settings.VAPID.PublicKey,
	}
	
	module := &WebPushModule{
		ModuleBase:  modules.NewModuleBase(),
		MessageBase: modules.NewMessageBase(),
		client:      NewWebPushClient(vapidConfig),
	}

	// 设置模块属�?	module.Name = "WebPush"
	module.Type = models.ModuleTypeNotification
	module.SubType = models.MessageChannelWebPush
	module.Priority = 6

	return module
}

// InitModule 初始化模�?func (w *WebPushModule) InitModule() error {
	// 初始化服�?	w.InitService(strings.ToLower(w.GetName()))
	w.Channel = &models.MessageChannelWebPush
	return nil
}

// HandleConfigChanged 处理配置变更事件
func (w *WebPushModule) HandleConfigChanged(eventData *models.ConfigChangeEventData) {
	if eventData == nil {
		return
	}

	// 检查是否是通知配置变更
	if eventData.Key != string(models.SystemConfigKeyNotifications) {
		return
	}

	fmt.Println("配置变更，重新加载WebPush模块...")
	w.InitModule()
}

// GetName 获取模块名称
func (w *WebPushModule) GetName() string {
	return "WebPush"
}

// GetType 获取模块类型
func (w *WebPushModule) GetType() models.ModuleType {
	return models.ModuleTypeNotification
}

// GetSubType 获取模块的子类型
func (w *WebPushModule) GetSubType() models.MessageChannel {
	return models.MessageChannelWebPush
}

// GetPriority 获取模块优先�?func (w *WebPushModule) GetPriority() int {
	return 6
}

// Stop 停止模块
func (w *WebPushModule) Stop() error {
	// 实现停止逻辑
	return nil
}

// Test 测试模块连接�?func (w *WebPushModule) Test() (bool, string) {
	return true, ""
}

// InitSetting 初始化设�?func (w *WebPushModule) InitSetting() (string, interface{}) {
	// 实现初始化设置逻辑
	return "", nil
}

// PostMessage 发送消�?func (w *WebPushModule) PostMessage(message *models.Notification) {
	configs := w.GetConfigs()
	for name, conf := range configs {
		// 检查消�?		if !w.CheckMessage(message, &name) {
			continue
		}

		// 类型断言获取配置
		configMap, ok := conf.(map[string]interface{})
		if !ok {
			continue
		}

		// 获取WebPush用户配置
		webpushUsers := ""
		if users, exists := configMap["WEBPUSH_USERNAME"]; exists {
			if userStr, ok := users.(string); ok {
				webpushUsers = userStr
			}
		}

		if webpushUsers != "" {
			// 设定了接收用户时，非该用户的消息不接�?			if message.Username == "" || !contains(webpushUsers, message.Username) {
				continue
			}
		}

		// 检查标题和内容是否同时为空
		if message.Title == "" && message.Text == "" {
			fmt.Println("标题和内容不能同时为�?)
			return
		}

		// 构造消息内�?		var caption, content string
		if message.Title != "" {
			caption = message.Title
			content = message.Text
		} else {
			caption = message.Text
			content = ""
		}

		// 构造推送载�?		payload := map[string]string{
			"title": caption,
			"body":  content,
			"url":   message.Link,
		}

		// 如果链接为空，设置默认�?		if payload["url"] == "" {
			payload["url"] = "/?shotcut=message"
		}

		// 获取订阅信息
		subscriptions := config.GlobalVars.GetSubscriptions()
		for _, sub := range subscriptions {
			fmt.Printf("�?%v 发送WebPush�?s %s\n", sub, caption, content)
			
			// 将subscription转换为SubscriptionInfo结构�?			var subscriptionInfo SubscriptionInfo
			if subMap, ok := sub.(map[string]interface{}); ok {
				if endpoint, ok := subMap["endpoint"].(string); ok {
					subscriptionInfo.Endpoint = endpoint
				}
				if keys, ok := subMap["keys"].(map[string]interface{}); ok {
					if p256dh, ok := keys["p256dh"].(string); ok {
						subscriptionInfo.Keys.P256dh = p256dh
					}
					if auth, ok := keys["auth"].(string); ok {
						subscriptionInfo.Keys.Auth = auth
					}
				}
			}

			// 发送WebPush消息
			err := w.client.Send(subscriptionInfo, payload)
			if err != nil {
				fmt.Printf("WebPush发送失�? %v\n", err)
			}
		}
	}
}

// contains 检查用户是否在用户列表�?func contains(userList, user string) bool {
	users := strings.Split(userList, ",")
	for _, u := range users {
		if strings.TrimSpace(u) == user {
			return true
		}
	}
	return false
}
