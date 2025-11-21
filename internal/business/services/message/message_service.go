package message

import (
	"context"
	"errors"
	"time"

	"moviepilot-go/pkg/logger"
	"moviepilot-go/internal/repositories/interfaces"
	"moviepilot-go/internal/models"
	"moviepilot-go/internal/business/services"
)

var (
	// ErrMessageNotFound 消息不存在
	ErrMessageNotFound = errors.New("消息不存在")
	// ErrMessageAlreadyRead 消息已读
	ErrMessageAlreadyRead = errors.New("消息已读")
)

// MessageService 消息服务实现
type MessageService struct {
	messageRepo interfaces.MessageRepository
	userRepo    interfaces.UserRepository
	logger      *logger.Logger
}

// NewMessageService 创建消息服务
func NewMessageService(
	messageRepo interfaces.MessageRepository,
	userRepo interfaces.UserRepository,
	log *logger.Logger,
) service.MessageService {
	return &MessageService{
		messageRepo: messageRepo,
		userRepo:    userRepo,
		logger:      log,
	}
}

// SendMessage 发送消息
func (s *MessageService) SendMessage(ctx context.Context, title, content string, messageType string, userIDs []uint) error {
	s.logger.Info("发送消息", "title", title, "type", messageType, "userCount", len(userIDs))

	// 如果是广播消息(userIDs为空)，发送给所有用户
	if len(userIDs) == 0 {
		users, err := s.userRepo.List(0, 0) // 获取所有用户
		if err != nil {
			s.logger.Error("获取用户列表失败", "error", err)
			return err
		}

		for _, user := range users {
			userIDs = append(userIDs, user.ID)
		}
	}

	// 为每个用户创建消息记录
	for _, userID := range userIDs {
		message := &models.Message{
			Title:     title,
			Content:   content,
			Type:      messageType,
			UserID:    userID,
			IsRead:    false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := s.messageRepo.Create(message); err != nil {
			s.logger.Error("创建消息记录失败", "userID", userID, "error", err)
			// 继续发送给其他用户，不中断
			continue
		}
	}

	s.logger.Info("消息发送完成", "title", title, "recipientCount", len(userIDs))
	return nil
}

// GetMessages 获取用户消息列表
func (s *MessageService) GetMessages(ctx context.Context, userID uint, page, size int) ([]*models.Message, int64, error) {
	if size <= 0 {
		size = 20 // 默认每页20条
	}
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * size

	// 使用List方法并筛选用户消息
	allMessages, total, err := s.messageRepo.List(offset, size)
	if err != nil {
		s.logger.Error("获取消息列表失败", "userID", userID, "error", err)
		return nil, 0, err
	}

	// 筛选属于当前用户的消息
	var userMessages []*models.Message
	for _, message := range allMessages {
		if message.UserID == uint(userID) {
			userMessages = append(userMessages, message)
		}
	}

	s.logger.Debug("获取消息列表成功", "userID", userID, "count", len(userMessages), "total", total)
	return userMessages, total, nil
}

// MarkAsRead 标记消息为已读
func (s *MessageService) MarkAsRead(ctx context.Context, messageID, userID uint) error {
	message, err := s.messageRepo.GetByID(messageID)
	if err != nil {
		s.logger.Error("获取消息失败", "messageID", messageID, "error", err)
		return err
	}

	if message == nil {
		return ErrMessageNotFound
	}

	// 检查消息是否属于该用户
	if message.UserID != userID {
		s.logger.Warn("用户无权访问该消息", "userID", userID, "messageUserID", message.UserID)
		return ErrMessageNotFound
	}

	// 检查消息是否已读
	if message.IsRead {
		return ErrMessageAlreadyRead
	}

	// 更新为已读状态
	message.IsRead = true
	message.UpdatedAt = time.Now()

	if err := s.messageRepo.Update(message); err != nil {
		s.logger.Error("更新消息状态失败", "messageID", messageID, "error", err)
		return err
	}

	s.logger.Info("消息标记为已读", "messageID", messageID, "userID", userID)
	return nil
}

// MarkAllAsRead 标记所有消息为已读
func (s *MessageService) MarkAllAsRead(ctx context.Context, userID uint) error {
	s.logger.Info("标记所有消息为已读", "userID", userID)

	// 获取用户的所有未读消息
	userIDPtr := int(userID)
	unreadMessages, err := s.messageRepo.GetUnread(&userIDPtr)
	if err != nil {
		s.logger.Error("获取未读消息失败", "userID", userID, "error", err)
		return err
	}

	// 批量更新为已读状态
	for _, message := range unreadMessages {
		message.IsRead = true
		message.UpdatedAt = time.Now()

		if err := s.messageRepo.Update(message); err != nil {
			s.logger.Error("更新消息状态失败", "messageID", message.ID, "error", err)
			// 继续处理其他消息
			continue
		}
	}

	s.logger.Info("所有消息标记为已读完成", "userID", userID, "processedCount", len(unreadMessages))
	return nil
}

// DeleteMessage 删除消息
func (s *MessageService) DeleteMessage(ctx context.Context, messageID, userID uint) error {
	message, err := s.messageRepo.GetByID(messageID)
	if err != nil {
		s.logger.Error("获取消息失败", "messageID", messageID, "error", err)
		return err
	}

	if message == nil {
		return ErrMessageNotFound
	}

	// 检查消息是否属于该用户
	if message.UserID != userID {
		s.logger.Warn("用户无权删除该消息", "userID", userID, "messageUserID", message.UserID)
		return ErrMessageNotFound
	}

	if err := s.messageRepo.Delete(messageID); err != nil {
		s.logger.Error("删除消息失败", "messageID", messageID, "error", err)
		return err
	}

	s.logger.Info("消息删除成功", "messageID", messageID, "userID", userID)
	return nil
}

// GetUnreadCount 获取未读消息数量
func (s *MessageService) GetUnreadCount(ctx context.Context, userID uint) (int64, error) {
	userIDPtr := int(userID)
	count, err := s.messageRepo.CountUnread(&userIDPtr)
	if err != nil {
		s.logger.Error("获取未读消息数量失败", "userID", userID, "error", err)
		return 0, err
	}

	return count, nil
}
