package notification

import "context"

// NotificationType 通知类型
type NotificationType string

const (
	// NotificationTypeText 文本消息
	NotificationTypeText NotificationType = "text"
	// NotificationTypeImage 图片消息
	NotificationTypeImage NotificationType = "image"
	// NotificationTypeFile 文件消息
	NotificationTypeFile NotificationType = "file"
	// NotificationTypeMarkdown Markdown 消息
	NotificationTypeMarkdown NotificationType = "markdown"
)

// Message 通知消息
type Message struct {
	// Type 消息类型
	Type NotificationType `json:"type"`
	// Title 消息标题（可选）
	Title string `json:"title,omitempty"`
	// Content 消息内容
	Content string `json:"content"`
	// ImageURL 图片URL（仅图片消息）
	ImageURL string `json:"image_url,omitempty"`
	// FileURL 文件URL（仅文件消息）
	FileURL string `json:"file_url,omitempty"`
	// FileName 文件名（仅文件消息）
	FileName string `json:"file_name,omitempty"`
	// Extra 额外参数
	Extra map[string]any `json:"extra,omitempty"`
}

// Client 通知客户端接口
// 所有通知渠道（Telegram、WeChat、邮件等）都需要实现此接口
type Client interface {
	// Name 返回通知渠道名称
	Name() string

	// SendText 发送文本消息
	SendText(ctx context.Context, message string) error

	// SendImage 发送图片消息
	SendImage(ctx context.Context, imageURL string, caption string) error

	// SendFile 发送文件消息
	SendFile(ctx context.Context, fileURL string, filename string) error

	// SendMarkdown 发送 Markdown 消息（可选，部分渠道支持）
	SendMarkdown(ctx context.Context, markdown string) error

	// Send 发送通用消息
	Send(ctx context.Context, msg *Message) error

	// TestConnection 测试连接
	TestConnection(ctx context.Context) error
}

// Factory 通知客户端工厂
type Factory struct {
	clients map[string]Client
}

// NewFactory 创建工厂实例
func NewFactory() *Factory {
	return &Factory{
		clients: make(map[string]Client),
	}
}

// Register 注册通知客户端
func (f *Factory) Register(client Client) {
	f.clients[client.Name()] = client
}

// Get 获取指定名称的通知客户端
func (f *Factory) Get(name string) (Client, bool) {
	client, ok := f.clients[name]
	return client, ok
}

// List 列出所有已注册的通知客户端
func (f *Factory) List() []string {
	names := make([]string, 0, len(f.clients))
	for name := range f.clients {
		names = append(names, name)
	}
	return names
}
