package message

import "moviepilot-go/internal/models/enums"

// CommingMessage 外来消息
type CommingMessage struct {
	// 用户ID
	UserID any `json:"userid,omitempty"` // string or int
	// 用户名称
	Username string `json:"username,omitempty"`
	// 消息渠道
	Channel enums.MessageChannel `json:"channel,omitempty"`
	// 来源（渠道名称）
	Source string `json:"source,omitempty"`
	// 消息体
	Text string `json:"text,omitempty"`
	// 时间
	Date string `json:"date,omitempty"`
	// 消息方向
	Action int `json:"action,omitempty"`
	// 是否为回调消息
	IsCallback bool `json:"is_callback,omitempty"`
	// 回调数据
	CallbackData string `json:"callback_data,omitempty"`
	// 消息ID（用于回调时定位原消息）
	MessageID any `json:"message_id,omitempty"` // string or int
	// 聊天ID（用于回调时定位聊天）
	ChatID string `json:"chat_id,omitempty"`
	// 完整的回调查询信息（原始数据）
	CallbackQuery map[string]any `json:"callback_query,omitempty"`
}

// ToDict 转换为字典
func (m *CommingMessage) ToDict() map[string]any {
	result := make(map[string]any)
	result["userid"] = m.UserID
	result["username"] = m.Username
	result["channel"] = m.Channel
	result["source"] = m.Source
	result["text"] = m.Text
	result["date"] = m.Date
	result["action"] = m.Action
	result["is_callback"] = m.IsCallback
	result["callback_data"] = m.CallbackData
	result["message_id"] = m.MessageID
	result["chat_id"] = m.ChatID
	result["callback_query"] = m.CallbackQuery
	return result
}

// Notification 消息
type Notification struct {
	// 消息渠道
	Channel enums.MessageChannel `json:"channel,omitempty"`
	// 消息来源
	Source string `json:"source,omitempty"`
	// 消息类型
	MType enums.NotificationType `json:"mtype,omitempty"`
	// 内容类型
	CType enums.ContentType `json:"ctype,omitempty"`
	// 标题
	Title string `json:"title,omitempty"`
	// 文本内容
	Text string `json:"text,omitempty"`
	// 图片
	Image string `json:"image,omitempty"`
	// 链接
	Link string `json:"link,omitempty"`
	// 用户ID
	UserID any `json:"userid,omitempty"` // string or int
	// 用户名称
	Username string `json:"username,omitempty"`
	// 时间
	Date string `json:"date,omitempty"`
	// 消息方向
	Action int `json:"action,omitempty"`
	// 消息目标用户ID字典，未指定用户ID时使用
	Targets map[string]any `json:"targets,omitempty"`
	// 按钮列表，格式：[ [{"text": "按钮文本", "callback_data": "回调数据", "url": "链接"}] ]
	Buttons [][]map[string]string `json:"buttons,omitempty"`
	// 原消息ID，用于编辑消息
	OriginalMessageID any `json:"original_message_id,omitempty"` // string or int
	// 原消息的聊天ID，用于编辑消息
	OriginalChatID string `json:"original_chat_id,omitempty"`
}

// ToDict 转换为字典
func (n *Notification) ToDict() map[string]any {
	result := make(map[string]any)
	result["channel"] = n.Channel
	result["source"] = n.Source
	result["mtype"] = n.MType
	result["ctype"] = n.CType
	result["title"] = n.Title
	result["text"] = n.Text
	result["image"] = n.Image
	result["link"] = n.Link
	result["userid"] = n.UserID
	result["username"] = n.Username
	result["date"] = n.Date
	result["action"] = n.Action
	result["targets"] = n.Targets
	result["buttons"] = n.Buttons
	result["original_message_id"] = n.OriginalMessageID
	result["original_chat_id"] = n.OriginalChatID
	return result
}

// NotificationSwitch 消息开关
type NotificationSwitch struct {
	// 消息类型
	MType string `json:"mtype,omitempty"`
	// 微信开关
	Wechat bool `json:"wechat,omitempty"`
	// TG开关
	Telegram bool `json:"telegram,omitempty"`
	// Slack开关
	Slack bool `json:"slack,omitempty"`
	// SynologyChat开关
	SynologyChat bool `json:"synologychat,omitempty"`
	// VoceChat开关
	VoceChat bool `json:"vocechat,omitempty"`
	// WebPush开关
	WebPush bool `json:"webpush,omitempty"`
}

// Subscription 客户端消息订阅
type Subscription struct {
	Endpoint string         `json:"endpoint,omitempty"`
	Keys     map[string]any `json:"keys,omitempty"`
}

// SubscriptionMessage 客户端订阅消息体
type SubscriptionMessage struct {
	Title string         `json:"title,omitempty"`
	Body  string         `json:"body,omitempty"`
	Icon  string         `json:"icon,omitempty"`
	URL   string         `json:"url,omitempty"`
	Data  map[string]any `json:"data,omitempty"`
}

// ChannelCapability 渠道能力枚举
type ChannelCapability string

const (
	// 支持内联按钮
	ChannelCapabilityInlineButtons ChannelCapability = "inline_buttons"
	// 支持菜单命令
	ChannelCapabilityMenuCommands ChannelCapability = "menu_commands"
	// 支持消息编辑
	ChannelCapabilityMessageEditing ChannelCapability = "message_editing"
	// 支持消息删除
	ChannelCapabilityMessageDeletion ChannelCapability = "message_deletion"
	// 支持回调查询
	ChannelCapabilityCallbackQueries ChannelCapability = "callback_queries"
	// 支持富文本
	ChannelCapabilityRichText ChannelCapability = "rich_text"
	// 支持图片
	ChannelCapabilityImages ChannelCapability = "images"
	// 支持链接
	ChannelCapabilityLinks ChannelCapability = "links"
	// 支持文件发送
	ChannelCapabilityFileSending ChannelCapability = "file_sending"
)

// ChannelCapabilities 渠道能力配置
type ChannelCapabilities struct {
	Channel             enums.MessageChannel
	Capabilities        map[ChannelCapability]bool
	MaxButtonsPerRow    int
	MaxButtonRows       int
	MaxButtonTextLength int
	FallbackEnabled     bool
}

// ChannelCapabilityManager 渠道能力管理器
type ChannelCapabilityManager struct {
	capabilities map[enums.MessageChannel]*ChannelCapabilities
}

// NewChannelCapabilityManager 创建渠道能力管理器
func NewChannelCapabilityManager() *ChannelCapabilityManager {
	manager := &ChannelCapabilityManager{
		capabilities: make(map[enums.MessageChannel]*ChannelCapabilities),
	}

	// Telegram配置
	manager.capabilities[enums.MessageChannelTelegram] = &ChannelCapabilities{
		Channel: enums.MessageChannelTelegram,
		Capabilities: map[ChannelCapability]bool{
			ChannelCapabilityInlineButtons:   true,
			ChannelCapabilityMenuCommands:    true,
			ChannelCapabilityMessageEditing:  true,
			ChannelCapabilityMessageDeletion: true,
			ChannelCapabilityCallbackQueries: true,
			ChannelCapabilityRichText:        true,
			ChannelCapabilityImages:          true,
			ChannelCapabilityLinks:           true,
			ChannelCapabilityFileSending:     true,
		},
		MaxButtonsPerRow:    4,
		MaxButtonRows:       10,
		MaxButtonTextLength: 30,
		FallbackEnabled:     true,
	}

	// 微信配置
	manager.capabilities[enums.MessageChannelWechat] = &ChannelCapabilities{
		Channel: enums.MessageChannelWechat,
		Capabilities: map[ChannelCapability]bool{
			ChannelCapabilityImages:       true,
			ChannelCapabilityLinks:        true,
			ChannelCapabilityMenuCommands: true,
		},
		FallbackEnabled: true,
	}

	// Slack配置
	manager.capabilities[enums.MessageChannelSlack] = &ChannelCapabilities{
		Channel: enums.MessageChannelSlack,
		Capabilities: map[ChannelCapability]bool{
			ChannelCapabilityInlineButtons:   true,
			ChannelCapabilityMessageEditing:  true,
			ChannelCapabilityMessageDeletion: true,
			ChannelCapabilityCallbackQueries: true,
			ChannelCapabilityRichText:        true,
			ChannelCapabilityImages:          true,
			ChannelCapabilityLinks:           true,
			ChannelCapabilityMenuCommands:    true,
		},
		MaxButtonsPerRow:    3,
		MaxButtonRows:       8,
		MaxButtonTextLength: 25,
		FallbackEnabled:     true,
	}

	// 其他渠道配置...

	return manager
}

// GetCapabilities 获取渠道能力
func (m *ChannelCapabilityManager) GetCapabilities(channel enums.MessageChannel) *ChannelCapabilities {
	return m.capabilities[channel]
}

// SupportsCapability 检查渠道是否支持某项能力
func (m *ChannelCapabilityManager) SupportsCapability(channel enums.MessageChannel, capability ChannelCapability) bool {
	caps := m.GetCapabilities(channel)
	if caps == nil {
		return false
	}
	return caps.Capabilities[capability]
}

// SupportsButtons 检查渠道是否支持按钮
func (m *ChannelCapabilityManager) SupportsButtons(channel enums.MessageChannel) bool {
	return m.SupportsCapability(channel, ChannelCapabilityInlineButtons)
}

// SupportsCallbacks 检查渠道是否支持回调
func (m *ChannelCapabilityManager) SupportsCallbacks(channel enums.MessageChannel) bool {
	return m.SupportsCapability(channel, ChannelCapabilityCallbackQueries)
}

// SupportsEditing 检查渠道是否支持消息编辑
func (m *ChannelCapabilityManager) SupportsEditing(channel enums.MessageChannel) bool {
	return m.SupportsCapability(channel, ChannelCapabilityMessageEditing)
}

// SupportsDeletion 检查渠道是否支持消息删除
func (m *ChannelCapabilityManager) SupportsDeletion(channel enums.MessageChannel) bool {
	return m.SupportsCapability(channel, ChannelCapabilityMessageDeletion)
}

// GetMaxButtonsPerRow 获取每行最大按钮数
func (m *ChannelCapabilityManager) GetMaxButtonsPerRow(channel enums.MessageChannel) int {
	caps := m.GetCapabilities(channel)
	if caps == nil {
		return 2
	}
	return caps.MaxButtonsPerRow
}

// GetMaxButtonRows 获取最大按钮行数
func (m *ChannelCapabilityManager) GetMaxButtonRows(channel enums.MessageChannel) int {
	caps := m.GetCapabilities(channel)
	if caps == nil {
		return 5
	}
	return caps.MaxButtonRows
}

// GetMaxButtonTextLength 获取按钮文本最大长度
func (m *ChannelCapabilityManager) GetMaxButtonTextLength(channel enums.MessageChannel) int {
	caps := m.GetCapabilities(channel)
	if caps == nil {
		return 20
	}
	return caps.MaxButtonTextLength
}

// ShouldUseFallback 是否应该使用降级策略
func (m *ChannelCapabilityManager) ShouldUseFallback(channel enums.MessageChannel) bool {
	caps := m.GetCapabilities(channel)
	if caps == nil {
		return true
	}
	return caps.FallbackEnabled
}
