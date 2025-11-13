package models

// CommingMessage 外来消息
type CommingMessage struct {
	// 用户ID
	UserID interface{} `json:"userid,omitempty"`
	// 用户名称
	Username string `json:"username,omitempty"`
	// 消息渠道
	Channel string `json:"channel,omitempty"`
	// 来源（渠道名称）
	Source string `json:"source,omitempty"`
	// 消息��?	Text string `json:"text,omitempty"`
	// 时间
	Date string `json:"date,omitempty"`
	// 消息方向
	Action int `json:"action,omitempty"`
	// 是否为回调消��?	IsCallback bool `json:"is_callback,omitempty"`
	// 回调数据
	CallbackData string `json:"callback_data,omitempty"`
	// 消息ID（用于回调时定位原消息）
	MessageID interface{} `json:"message_id,omitempty"`
	// 聊天ID（用于回调时定位聊天��?	ChatID string `json:"chat_id,omitempty"`
	// 完整的回调查询信息（原始数据��?	CallbackQuery map[string]interface{} `json:"callback_query,omitempty"`
}

// Notification 消息
type Notification struct {
	// 消息渠道
	Channel string `json:"channel,omitempty"`
	// 消息来源
	Source string `json:"source,omitempty"`
	// 消息类型
	MType string `json:"mtype,omitempty"`
	// 内容类型
	CType *string `json:"ctype,omitempty"`
	// 标题
	Title *string `json:"title,omitempty"`
	// 文本内容
	Text *string `json:"text,omitempty"`
	// 图片
	Image *string `json:"image,omitempty"`
	// 链接
	Link *string `json:"link,omitempty"`
	// 用户ID
	UserID interface{} `json:"userid,omitempty"`
	// 用户名称
	Username *string `json:"username,omitempty"`
	// 时间
	Date *string `json:"date,omitempty"`
	// 消息方向
	Action int `json:"action,omitempty"`
	// 消息目标用户ID字典，未指定用户ID时使�?
	Targets map[string]interface{} `json:"targets,omitempty"`
	// 按钮列表，格式：[[{"text": "按钮文本", "callback_data": "回调数据", "url": "链接"}]]
	Buttons [][]map[string]interface{} `json:"buttons,omitempty"`
	// 原消息ID，用于编辑消�?
	OriginalMessageID interface{} `json:"original_message_id,omitempty"`
	// 原消息的聊天ID，用于编辑消�?
	OriginalChatID *string `json:"original_chat_id,omitempty"`
}

// UpdateFromMap 从map更新Notification字段
func (n *Notification) UpdateFromMap(data map[string]interface{}) {
	if channel, ok := data["channel"].(string); ok {
		n.Channel = channel
	}
	if source, ok := data["source"].(string); ok {
		n.Source = source
	}
	if mtype, ok := data["mtype"].(string); ok {
		n.MType = mtype
	}
	if ctype, ok := data["ctype"].(string); ok {
		n.CType = &ctype
	}
	if title, ok := data["title"].(string); ok {
		n.Title = &title
	}
	if text, ok := data["text"].(string); ok {
		n.Text = &text
	}
	if image, ok := data["image"].(string); ok {
		n.Image = &image
	}
	if link, ok := data["link"].(string); ok {
		n.Link = &link
	}
	if userid, ok := data["userid"].(interface{}); ok {
		n.UserID = userid
	}
	if username, ok := data["username"].(string); ok {
		n.Username = &username
	}
	if date, ok := data["date"].(string); ok {
		n.Date = &date
	}
	if action, ok := data["action"].(int); ok {
		n.Action = action
	}
	if targets, ok := data["targets"].(map[string]interface{}); ok {
		n.Targets = targets
	}
	if buttons, ok := data["buttons"].([][]map[string]interface{}); ok {
		n.Buttons = buttons
	}
	if originalMessageID, ok := data["original_message_id"].(interface{}); ok {
		n.OriginalMessageID = originalMessageID
	}
	if originalChatID, ok := data["original_chat_id"].(string); ok {
		n.OriginalChatID = &originalChatID
	}
}

// NotificationSwitch 消息开��?type NotificationSwitch struct {
	// 消息类型
	MType string `json:"mtype,omitempty"`
	// 微信开��?	Wechat bool `json:"wechat,omitempty"`
	// TG开��?	Telegram bool `json:"telegram,omitempty"`
	// Slack开��?	Slack bool `json:"slack,omitempty"`
	// SynologyChat开��?	SynologyChat bool `json:"synologychat,omitempty"`
	// VoceChat开��?	VoceChat bool `json:"vocechat,omitempty"`
	// WebPush开��?	WebPush bool `json:"webpush,omitempty"`
}

// Subscription 客户端消息订��?type Subscription struct {
	Endpoint string                 `json:"endpoint,omitempty"`
	Keys     map[string]interface{} `json:"keys,omitempty"`
}

// SubscriptionMessage 客户端订阅消息体
type SubscriptionMessage struct {
	Title string                 `json:"title,omitempty"`
	Body  string                 `json:"body,omitempty"`
	Icon  string                 `json:"icon,omitempty"`
	URL   string                 `json:"url,omitempty"`
	Data  map[string]interface{} `json:"data,omitempty"`
}

// ChannelCapability 渠道能力枚举
type ChannelCapability string

const (
	// 支持内联按钮
	InlineButtons ChannelCapability = "inline_buttons"
	// 支持菜单命令
	MenuCommands ChannelCapability = "menu_commands"
	// 支持消息编辑
	MessageEditing ChannelCapability = "message_editing"
	// 支持消息删除
	MessageDeletion ChannelCapability = "message_deletion"
	// 支持回调查询
	CallbackQueries ChannelCapability = "callback_queries"
	// 支持富文��?	RichText ChannelCapability = "rich_text"
	// 支持图片
	Images ChannelCapability = "images"
	// 支持链接
	Links ChannelCapability = "links"
	// 支持文件发��?	FileSending ChannelCapability = "file_sending"
)

// ChannelCapabilities 渠道能力配置
type ChannelCapabilities struct {
	Channel              string             `json:"channel"`
	Capabilities         []ChannelCapability `json:"capabilities"`
	MaxButtonsPerRow     int                `json:"max_buttons_per_row"`
	MaxButtonRows        int                `json:"max_button_rows"`
	MaxButtonTextLength  int                `json:"max_button_text_length"`
	FallbackEnabled      bool               `json:"fallback_enabled"`
}

// MessageChannel 消息渠道枚举
const (
	MessageChannelWechat       = "微信"
	MessageChannelTelegram     = "Telegram"
	MessageChannelSlack        = "Slack"
	MessageChannelSynologyChat = "SynologyChat"
	MessageChannelVoceChat     = "VoceChat"
	MessageChannelWeb          = "Web"
	MessageChannelWebPush      = "WebPush"
)

// ChannelCapabilityManager 渠道能力管理��?type ChannelCapabilityManager struct {
	Capabilities map[string]*ChannelCapabilities
}

// NewChannelCapabilityManager 创建一个新��?ChannelCapabilityManager 实例
func NewChannelCapabilityManager() *ChannelCapabilityManager {
	manager := &ChannelCapabilityManager{
		Capabilities: make(map[string]*ChannelCapabilities),
	}

	// 初始化各渠道的能力配��?	manager.Capabilities[MessageChannelTelegram] = &ChannelCapabilities{
		Channel: MessageChannelTelegram,
		Capabilities: []ChannelCapability{
			InlineButtons,
			MenuCommands,
			MessageEditing,
			MessageDeletion,
			CallbackQueries,
			RichText,
			Images,
			Links,
			FileSending,
		},
		MaxButtonsPerRow:    4,
		MaxButtonRows:       10,
		MaxButtonTextLength: 30,
		FallbackEnabled:     false,
	}

	manager.Capabilities[MessageChannelWechat] = &ChannelCapabilities{
		Channel: MessageChannelWechat,
		Capabilities: []ChannelCapability{
			Images,
			Links,
			MenuCommands,
		},
		MaxButtonsPerRow:    5,
		MaxButtonRows:       10,
		MaxButtonTextLength: 30,
		FallbackEnabled:     true,
	}

	manager.Capabilities[MessageChannelSlack] = &ChannelCapabilities{
		Channel: MessageChannelSlack,
		Capabilities: []ChannelCapability{
			InlineButtons,
			MessageEditing,
			MessageDeletion,
			CallbackQueries,
			RichText,
			Images,
			Links,
			MenuCommands,
		},
		MaxButtonsPerRow:    3,
		MaxButtonRows:       8,
		MaxButtonTextLength: 25,
		FallbackEnabled:     true,
	}

	manager.Capabilities[MessageChannelSynologyChat] = &ChannelCapabilities{
		Channel: MessageChannelSynologyChat,
		Capabilities: []ChannelCapability{
			RichText,
			Images,
			Links,
		},
		MaxButtonsPerRow:    5,
		MaxButtonRows:       10,
		MaxButtonTextLength: 30,
		FallbackEnabled:     true,
	}

	manager.Capabilities[MessageChannelVoceChat] = &ChannelCapabilities{
		Channel: MessageChannelVoceChat,
		Capabilities: []ChannelCapability{
			RichText,
			Images,
			Links,
		},
		MaxButtonsPerRow:    5,
		MaxButtonRows:       10,
		MaxButtonTextLength: 30,
		FallbackEnabled:     true,
	}

	manager.Capabilities[MessageChannelWebPush] = &ChannelCapabilities{
		Channel: MessageChannelWebPush,
		Capabilities: []ChannelCapability{
			Links,
		},
		MaxButtonsPerRow:    5,
		MaxButtonRows:       10,
		MaxButtonTextLength: 30,
		FallbackEnabled:     true,
	}

	manager.Capabilities[MessageChannelWeb] = &ChannelCapabilities{
		Channel: MessageChannelWeb,
		Capabilities: []ChannelCapability{
			RichText,
			Images,
			Links,
		},
		MaxButtonsPerRow:    5,
		MaxButtonRows:       10,
		MaxButtonTextLength: 30,
		FallbackEnabled:     true,
	}

	return manager
}

// GetCapabilities 获取渠道能力
func (cm *ChannelCapabilityManager) GetCapabilities(channel string) *ChannelCapabilities {
	return cm.Capabilities[channel]
}

// SupportsCapability 检查渠道是否支持某项能��?func (cm *ChannelCapabilityManager) SupportsCapability(channel string, capability ChannelCapability) bool {
	channelCaps := cm.GetCapabilities(channel)
	if channelCaps == nil {
		return false
	}

	for _, cap := range channelCaps.Capabilities {
		if cap == capability {
			return true
		}
	}
	return false
}

// SupportsButtons 检查渠道是否支持按��?func (cm *ChannelCapabilityManager) SupportsButtons(channel string) bool {
	return cm.SupportsCapability(channel, InlineButtons)
}

// SupportsCallbacks 检查渠道是否支持回��?func (cm *ChannelCapabilityManager) SupportsCallbacks(channel string) bool {
	return cm.SupportsCapability(channel, CallbackQueries)
}

// SupportsEditing 检查渠道是否支持消息编��?func (cm *ChannelCapabilityManager) SupportsEditing(channel string) bool {
	return cm.SupportsCapability(channel, MessageEditing)
}

// SupportsDeletion 检查渠道是否支持消息删��?func (cm *ChannelCapabilityManager) SupportsDeletion(channel string) bool {
	return cm.SupportsCapability(channel, MessageDeletion)
}

// GetMaxButtonsPerRow 获取每行最大按钮数
func (cm *ChannelCapabilityManager) GetMaxButtonsPerRow(channel string) int {
	channelCaps := cm.GetCapabilities(channel)
	if channelCaps == nil {
		return 2
	}
	return channelCaps.MaxButtonsPerRow
}

// GetMaxButtonRows 获取最大按钮行��?func (cm *ChannelCapabilityManager) GetMaxButtonRows(channel string) int {
	channelCaps := cm.GetCapabilities(channel)
	if channelCaps == nil {
		return 5
	}
	return channelCaps.MaxButtonRows
}

// GetMaxButtonTextLength 获取按钮文本最大长��?func (cm *ChannelCapabilityManager) GetMaxButtonTextLength(channel string) int {
	channelCaps := cm.GetCapabilities(channel)
	if channelCaps == nil {
		return 20
	}
	return channelCaps.MaxButtonTextLength
}

// ShouldUseFallback 是否应该使用降级策略
func (cm *ChannelCapabilityManager) ShouldUseFallback(channel string) bool {
	channelCaps := cm.GetCapabilities(channel)
	if channelCaps == nil {
		return true
	}
	return channelCaps.FallbackEnabled
}

// NewCommingMessage 创建一个新��?CommingMessage 实例
func NewCommingMessage() *CommingMessage {
	return &CommingMessage{
		CallbackQuery: make(map[string]interface{}),
	}
}

// NewNotification 创建一个新��?Notification 实例
func NewNotification() *Notification {
	return &Notification{
		Targets:  make(map[string]interface{}),
		Buttons:  make([][]map[string]interface{}, 0),
	}
}

// NewNotificationWithCType 创建一个带有CType的Notification实例
func NewNotificationWithCType(ctype string) *Notification {
	return &Notification{
		CType:    &ctype,
		Targets:  make(map[string]interface{}),
		Buttons:  make([][]map[string]interface{}, 0),
	}
}

// NewNotificationWithTitleAndText 创建一个带有标题和文本的Notification实例
func NewNotificationWithTitleAndText(title, text string) *Notification {
	return &Notification{
		Title:    &title,
		Text:     &text,
		Targets:  make(map[string]interface{}),
		Buttons:  make([][]map[string]interface{}, 0),
	}
}

// NewNotificationSwitch 创建一个新��?NotificationSwitch 实例
func NewNotificationSwitch() *NotificationSwitch {
	return &NotificationSwitch{}
}

// NewSubscription 创建一个新��?Subscription 实例
func NewSubscription() *Subscription {
	return &Subscription{
		Keys: make(map[string]interface{}),
	}
}

// NewSubscriptionMessage 创建一个新��?SubscriptionMessage 实例
func NewSubscriptionMessage() *SubscriptionMessage {
	return &SubscriptionMessage{
		Data: make(map[string]interface{}),
	}
}
