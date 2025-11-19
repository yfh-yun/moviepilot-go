package notification

import (
	"context"
	"fmt"
	"time"

	"github.com/yfh-yun/moviepilot-go/internal/integration/notification"
	"github.com/yfh-yun/moviepilot-go/internal/repository/interfaces"
)

// Service 通知服务
type Service struct {
	manager     *notification.Manager
	messageRepo interfaces.MessageRepository
}

// NewService 创建新的通知服务
func NewService(manager *notification.Manager, messageRepo interfaces.MessageRepository) *Service {
	return &Service{
		manager:     manager,
		messageRepo: messageRepo,
	}
}

// SendNotificationRequest 发送通知请求
type SendNotificationRequest struct {
	Title      string                            `json:"title" validate:"required"`
	Content    string                            `json:"content" validate:"required"`
	Level      notification.NotificationLevel    `json:"level" validate:"required,oneof=info warning error success"`
	Priority   notification.NotificationPriority `json:"priority" validate:"required,min=0,max=3"`
	Channel    string                            `json:"channel,omitempty"`
	Channels   []string                          `json:"channels,omitempty"`
	Category   string                            `json:"category,omitempty"`
	Tags       []string                          `json:"tags,omitempty"`
	ImageURL   string                            `json:"image_url,omitempty"`
	LinkURL    string                            `json:"link_url,omitempty"`
	TTL        time.Duration                     `json:"ttl,omitempty"`
	CustomData map[string]interface{}            `json:"custom_data,omitempty"`
}

// SendNotificationResponse 发送通知响应
type SendNotificationResponse struct {
	MessageID string                             `json:"message_id"`
	Results   []*notification.NotificationResult `json:"results"`
	SentAt    time.Time                          `json:"sent_at"`
}

// SendNotification 发送通知
func (s *Service) SendNotification(ctx context.Context, req *SendNotificationRequest) (*SendNotificationResponse, error) {
	// 构建消息
	messageBuilder := notification.NewMessageBuilder()
	message := messageBuilder.
		WithTitle(req.Title).
		WithContent(req.Content).
		WithLevel(req.Level).
		WithPriority(req.Priority).
		WithCategory(req.Category).
		WithTags(req.Tags...).
		WithImageURL(req.ImageURL).
		WithLinkURL(req.LinkURL).
		WithTTL(req.TTL).
		Build()

	// 为自定义数据添加所有键值对
	for key, value := range req.CustomData {
		messageBuilder.WithCustomData(key, value)
	}

	var results []*notification.NotificationResult
	var err error

	// 确定发送目标渠道
	if req.Channel != "" {
		// 发送到单个渠道
		message.Channel = req.Channel
		result, sendErr := s.manager.Send(ctx, message)
		if sendErr != nil {
			return nil, fmt.Errorf("failed to send to channel %s: %w", req.Channel, sendErr)
		}
		results = append(results, result)
	} else if len(req.Channels) > 0 {
		// 发送到多个指定渠道
		message.Channel = req.Channels[0] // 临时设置一个渠道
		results, err = s.manager.SendToMultiple(ctx, message, req.Channels)
		if err != nil {
			return nil, fmt.Errorf("failed to send to multiple channels: %w", err)
		}
	} else {
		// 广播到所有启用的渠道
		message.Channel = "broadcast" // 临时设置
		results, err = s.manager.Broadcast(ctx, message)
		if err != nil {
			return nil, fmt.Errorf("failed to broadcast: %w", err)
		}
	}

	// 保存消息到数据库
	dbMessage := &repository.Message{
		Title:      req.Title,
		Content:    req.Content,
		Level:      string(req.Level),
		Priority:   int(req.Priority),
		Category:   req.Category,
		Tags:       req.Tags,
		ImageURL:   req.ImageURL,
		LinkURL:    req.LinkURL,
		CustomData: req.CustomData,
		UserID:     getUserIDFromContext(ctx), // 从上下文中获取用户ID
		CreatedAt:  time.Now(),
	}

	if err := s.messageRepo.Create(ctx, dbMessage); err != nil {
		// 发送成功但保存失败，记录错误但不影响响应
		fmt.Printf("Failed to save message to database: %v", err)
	}

	return &SendNotificationResponse{
		MessageID: message.ID,
		Results:   results,
		SentAt:    time.Now(),
	}, nil
}

// GetChannelsRequest 获取渠道请求
type GetChannelsRequest struct {
	EnabledOnly bool `json:"enabled_only"`
}

// GetChannelsResponse 获取渠道响应
type GetChannelsResponse struct {
	Channels []*notification.Channel `json:"channels"`
}

// GetChannels 获取通知渠道列表
func (s *Service) GetChannels(ctx context.Context, req *GetChannelsRequest) (*GetChannelsResponse, error) {
	var channels []*notification.Channel

	if req.EnabledOnly {
		channels = s.manager.ListEnabledChannels()
	} else {
		channels = s.manager.ListChannels()
	}

	return &GetChannelsResponse{
		Channels: channels,
	}, nil
}

// GetChannelStatusRequest 获取渠道状态请求
type GetChannelStatusRequest struct {
	Name string `json:"name" validate:"required"`
}

// GetChannelStatusResponse 获取渠道状态响应
type GetChannelStatusResponse struct {
	Channel *notification.Channel            `json:"channel"`
	Status  *notification.NotificationResult `json:"status"`
}

// GetChannelStatus 获取渠道状态
func (s *Service) GetChannelStatus(ctx context.Context, req *GetChannelStatusRequest) (*GetChannelStatusResponse, error) {
	channel, err := s.manager.GetChannel(req.Name)
	if err != nil {
		return nil, fmt.Errorf("channel %s not found: %w", req.Name, err)
	}

	// 测试发送一条测试消息来检查状态
	testMessage := notification.NewMessageBuilder().
		WithTitle("测试消息").
		WithContent("这是一条测试消息，用于检查渠道状态。").
		WithLevel(notification.LevelInfo).
		WithPriority(notification.PriorityLow).
		WithChannel(req.Name).
		Build()

	status, err := s.manager.Send(ctx, testMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to send test message: %w", err)
	}

	return &GetChannelStatusResponse{
		Channel: channel,
		Status:  status,
	}, nil
}

// HealthCheckRequest 健康检查请求
type HealthCheckRequest struct {
	IncludeChannels bool `json:"include_channels"`
}

// HealthCheckResponse 健康检查响应
type HealthCheckResponse struct {
	OverallStatus string                   `json:"overall_status"`
	Statistics    *notification.Statistics `json:"statistics"`
	ChannelHealth map[string]error         `json:"channel_health,omitempty"`
}

// HealthCheck 健康检查
func (s *Service) HealthCheck(ctx context.Context, req *HealthCheckRequest) (*HealthCheckResponse, error) {
	// 获取统计信息
	stats, err := s.manager.GetStatistics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}

	response := &HealthCheckResponse{
		Statistics: stats,
	}

	// 确定整体状态
	if stats.HealthyChannels == stats.TotalChannels {
		response.OverallStatus = "healthy"
	} else if stats.HealthyChannels >= stats.TotalChannels/2 {
		response.OverallStatus = "degraded"
	} else {
		response.OverallStatus = "unhealthy"
	}

	// 如果请求包含渠道健康状态
	if req.IncludeChannels {
		channelHealth := s.manager.HealthCheck(ctx)
		response.ChannelHealth = channelHealth
	}

	return response, nil
}

// GetMessagesRequest 获取消息请求
type GetMessagesRequest struct {
	Page     int       `json:"page" validate:"min=1"`
	PageSize int       `json:"page_size" validate:"min=1,max=100"`
	Level    string    `json:"level,omitempty"`
	Category string    `json:"category,omitempty"`
	StartAt  time.Time `json:"start_at,omitempty"`
	EndAt    time.Time `json:"end_at,omitempty"`
}

// GetMessagesResponse 获取消息响应
type GetMessagesResponse struct {
	Messages []*repository.Message `json:"messages"`
	Total    int64                 `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"page_size"`
}

// GetMessages 获取消息历史
func (s *Service) GetMessages(ctx context.Context, req *GetMessagesRequest) (*GetMessagesResponse, error) {
	filter := &repository.MessageFilter{
		Page:     req.Page,
		PageSize: req.PageSize,
		Level:    req.Level,
		Category: req.Category,
		StartAt:  req.StartAt,
		EndAt:    req.EndAt,
		UserID:   getUserIDFromContext(ctx),
	}

	messages, total, err := s.messageRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	return &GetMessagesResponse{
		Messages: messages,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// DeleteMessageRequest 删除消息请求
type DeleteMessageRequest struct {
	MessageID string `json:"message_id" validate:"required"`
}

// DeleteMessage 删除消息
func (s *Service) DeleteMessage(ctx context.Context, req *DeleteMessageRequest) error {
	return s.messageRepo.Delete(ctx, req.MessageID)
}

// getUserIDFromContext 从上下文中获取用户ID
func getUserIDFromContext(ctx context.Context) string {
	// 这里应该从JWT令牌或其他认证方式中获取用户ID
	// 暂时返回空字符串
	return ""
}

// RegisterChannelRequest 注册渠道请求
type RegisterChannelRequest struct {
	Name        string            `json:"name" validate:"required"`
	Description string            `json:"description"`
	Enabled     bool              `json:"enabled"`
	Config      map[string]string `json:"config"`
	SenderType  string            `json:"sender_type" validate:"required,oneof=wechat telegram dingding feishu email webpush"`
}

// RegisterChannel 注册通知渠道
func (s *Service) RegisterChannel(ctx context.Context, req *RegisterChannelRequest) error {
	// 根据发送器类型创建相应的发送器
	var sender notification.Sender

	switch req.SenderType {
	case "wechat":
		// 创建微信发送器
		config := &notification.WeChatConfig{
			CorpID:  req.Config["corp_id"],
			AgentID: req.Config["agent_id"],
			Secret:  req.Config["secret"],
			MsgType: req.Config["msg_type"],
		}
		sender = notification.NewWeChatSender(config)
	case "telegram":
		// 创建Telegram发送器
		// TODO: 实现Telegram发送器
		return fmt.Errorf("telegram sender not implemented yet")
	case "dingding":
		// 创建钉钉发送器
		// TODO: 实现钉钉发送器
		return fmt.Errorf("dingding sender not implemented yet")
	case "feishu":
		// 创建飞书发送器
		// TODO: 实现飞书发送器
		return fmt.Errorf("feishu sender not implemented yet")
	case "email":
		// 创建邮件发送器
		// TODO: 实现邮件发送器
		return fmt.Errorf("email sender not implemented yet")
	case "webpush":
		// 创建Web Push发送器
		// TODO: 实现Web Push发送器
		return fmt.Errorf("webpush sender not implemented yet")
	default:
		return fmt.Errorf("unsupported sender type: %s", req.SenderType)
	}

	// 创建渠道
	channel := &notification.Channel{
		Name:        req.Name,
		Description: req.Description,
		Enabled:     req.Enabled,
		Sender:      sender,
		Config:      req.Config,
	}

	return s.manager.RegisterChannel(channel)
}

// EnableChannelRequest 启用渠道请求
type EnableChannelRequest struct {
	Name string `json:"name" validate:"required"`
}

// EnableChannel 启用通知渠道
func (s *Service) EnableChannel(ctx context.Context, req *EnableChannelRequest) error {
	return s.manager.EnableChannel(req.Name)
}

// DisableChannelRequest 禁用渠道请求
type DisableChannelRequest struct {
	Name string `json:"name" validate:"required"`
}

// DisableChannel 禁用通知渠道
func (s *Service) DisableChannel(ctx context.Context, req *DisableChannelRequest) error {
	return s.manager.DisableChannel(req.Name)
}

// UnregisterChannelRequest 注销渠道请求
type UnregisterChannelRequest struct {
	Name string `json:"name" validate:"required"`
}

// UnregisterChannel 注销通知渠道
func (s *Service) UnregisterChannel(ctx context.Context, req *UnregisterChannelRequest) error {
	return s.manager.UnregisterChannel(req.Name)
}
