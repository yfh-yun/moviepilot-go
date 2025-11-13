package modules

import (
	"moviepilot-go/pkg/models"
)

// MessageBase 消息基类
type MessageBase struct {
	*ServiceBase
	// 消息通道
	Channel *models.MessageChannel
}

// NewMessageBase 创建一个新的消息基类实�?func NewMessageBase() *MessageBase {
	return &MessageBase{
		ServiceBase: NewServiceBase(),
		Channel:     nil,
	}
}

// GetConfigs 获取已启用的消息通知渠道的配置字�?func (m *MessageBase) GetConfigs() map[string]interface{} {
	// 注意：这里需要调用ServiceConfigHelper.get_notification_configs()
	// 由于这是Python代码，我们需要在Go中实现相应的功能
	configs := make(map[string]interface{}) // 这里应该从配置中获取实际的通知配置
	
	if m.ServiceName == "" {
		return make(map[string]interface{})
	}
	
	// 过滤出启用的配置
	enabledConfigs := make(map[string]interface{})
	for name, config := range configs {
		// 这里需要检查config是否启用以及类型是否匹配
		// 由于缺少具体实现，暂时返回空映射
		_ = name
		_ = config
	}
	
	return enabledConfigs
}

// CheckMessage 检查消息渠道及消息类型，判断是否处理消�?func (m *MessageBase) CheckMessage(message *models.Notification, source *string) bool {
	// 检查消息渠�?	if message.Channel != nil && m.Channel != nil && *message.Channel != *m.Channel {
		return false
	}
	
	// 检查消息来�?	if message.Source != nil && source != nil && *message.Source != *source {
		return false
	}
	
	// 不是定向发送时，检查消息类型开�?	if message.UserID == nil && message.MType != nil {
		conf := m.GetConfig(source)
		if conf != nil {
			// 这里需要类型断言并检查开�?			// 由于缺少具体实现，暂时返回true
			_ = conf
		}
	}
	
	return true
}
