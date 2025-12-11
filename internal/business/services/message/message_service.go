package message

import (
	"context"
	"fmt"
	"time"

	"moviepilot-go/internal/business/services/base"
	"moviepilot-go/internal/models/database"
	"moviepilot-go/internal/models/dto"
	"moviepilot-go/internal/models/types"
	"moviepilot-go/internal/repositories/interfaces"
	"moviepilot-go/pkg/logger"

	"go.uber.org/zap"
)

// MessageService 消息服务
// 原MessageChain，负责外来消息处理
type MessageService struct {
	*base.ServiceBase
	repo interfaces.MessageRepository
}

// NewMessageService 创建MessageService实例
func NewMessageService(repo interfaces.MessageRepository) *MessageService {
	return &MessageService{
		ServiceBase: base.NewServiceBase(),
		repo:        repo,
	}
}

// Initialize 初始化服务
func (s *MessageService) Initialize() error {
	logger.Info("Initializing MessageService")
	return nil
}

// Name 获取服务名称
func (s *MessageService) Name() string {
	return "MessageService"
}

// Close 关闭服务
func (s *MessageService) Close() error {
	logger.Info("Closing MessageService")
	return nil
}

// HandleMessage 处理外来消息
func (s *MessageService) HandleMessage(ctx context.Context, msg *dto.CommingMessage) error {
	logger.Debug("Handling incoming message",
		zap.String("channel", string(msg.Channel)),
		zap.Any("user_id", msg.UserID),
		zap.String("text", msg.Text))

	// 1. 解析消息
	command, args, err := s.ParseCommand(ctx, msg.Text)
	if err != nil {
		logger.Error("Failed to parse command", zap.Error(err))
		return err
	}

	// 2. 识别命令并执行
	// TODO: 实现具体的命令执行逻辑
	logger.Info("Command parsed", zap.String("command", command), zap.Strings("args", args))

	// 3. 保存消息到数据库
	userIDStr := fmt.Sprintf("%v", msg.UserID)
	message := &database.Message{
		Channel:  string(msg.Channel),
		Source:   msg.Source,
		Type:     "command",
		Title:    "Command",
		Text:     msg.Text,
		UserID:   userIDStr,
		Username: msg.Username,
		Action:   msg.Action,
		RegTime:  time.Now().Format("2006-01-02 15:04:05"),
	}

	if err := s.repo.Create(ctx, message); err != nil {
		logger.Error("Failed to save message", zap.Error(err))
		return err
	}

	return nil
}

// ParseCommand 解析命令
func (s *MessageService) ParseCommand(ctx context.Context, text string) (string, []string, error) {
	logger.Debug("Parsing command", zap.String("text", text))

	// TODO: 实现更复杂的命令解析逻辑
	// 简单实现：按空格分割
	parts := []string{text}
	if len(parts) == 0 {
		return "", nil, fmt.Errorf("empty command")
	}

	command := parts[0]
	args := []string{}
	if len(parts) > 1 {
		args = parts[1:]
	}

	return command, args, nil
}

// SendNotification 发送通知
func (s *MessageService) SendNotification(ctx context.Context, notification *dto.Notification) error {
	logger.Info("Sending notification",
		zap.String("title", notification.Title))

	// 1. 保存通知到数据库
	userIDStr := fmt.Sprintf("%v", notification.UserID)
	message := &database.Message{
		Channel:  "system",
		Source:   notification.Source,
		Type:     string(notification.MType),
		Title:    notification.Title,
		Text:     notification.Title, // 使用Title作为Text内容
		UserID:   userIDStr,
		Username: "system",
		Action:   1,
		RegTime:  time.Now().Format("2006-01-02 15:04:05"),
		IsRead:   false,
	}

	if err := s.repo.Create(ctx, message); err != nil {
		logger.Error("Failed to save notification", zap.Error(err))
		return err
	}

	// TODO: 2. 调用通知渠道发送通知
	// 这里可以集成邮件、Webhook等通知渠道

	return nil
}

// SendMessage 发送消息
func (s *MessageService) SendMessage(ctx context.Context, channel types.MessageChannel, userID any, message string) error {
	logger.Info("Sending message",
		zap.String("channel", string(channel)),
		zap.Any("user_id", userID),
		zap.String("message", message))

	// 保存消息到数据库
	msg := &database.Message{
		Channel:  string(channel),
		Source:   "system",
		Type:     "message",
		Title:    "Message",
		Text:     message,
		UserID:   fmt.Sprintf("%v", userID),
		Username: "system",
		Action:   1,
		RegTime:  time.Now().Format("2006-01-02 15:04:05"),
		IsRead:   false,
	}

	if err := s.repo.Create(ctx, msg); err != nil {
		logger.Error("Failed to save message", zap.Error(err))
		return err
	}

	// TODO: 调用具体的消息发送渠道

	return nil
}
