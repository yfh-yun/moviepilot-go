package services

import (
	"context"
	"moviepilot-go/internal/business/domains"
	"moviepilot-go/internal/business/policies"
	"moviepilot-go/pkg/logger"
	"time"
)

// DomainService 领域服务接口
type DomainService interface {
	ValidateUser(user *domains.User) error
	ValidateMedia(media *domains.Media) error
	ValidateSubscribe(subscribe *domains.Subscribe) error
	ValidateTransfer(transfer *domains.Transfer) error
}

// DomainServiceImpl 领域服务实现
type DomainServiceImpl struct {
	policyManager *policies.PolicyManager
}

// NewDomainService 创建领域服务
func NewDomainService(policyManager *policies.PolicyManager) DomainService {
	return &DomainServiceImpl{
		policyManager: policyManager,
	}
}

// ValidateUser 验证用户
func (s *DomainServiceImpl) ValidateUser(user *domains.User) error {
	logger.Debug("验证用户", "user_id", user.ID, "username", user.Username)
	
	if user.Username == "" {
		return errors.New("用户名不能为空")
	}
	
	if user.Email == "" {
		return errors.New("邮箱不能为空")
	}
	
	// 可以添加更多验证逻辑
	
	logger.Info("用户验证通过", "user_id", user.ID)
	return nil
}

// ValidateMedia 验证媒体
func (s *DomainServiceImpl) ValidateMedia(media *domains.Media) error {
	logger.Debug("验证媒体", "media_id", media.ID, "title", media.Title)
	
	if media.Title == "" {
		return errors.New("媒体标题不能为空")
	}
	
	if media.MediaType == "" {
		return errors.New("媒体类型不能为空")
	}
	
	if media.MediaType != "movie" && media.MediaType != "tv" {
		return errors.New("无效的媒体类型")
	}
	
	logger.Info("媒体验证通过", "media_id", media.ID)
	return nil
}

// ValidateSubscribe 验证订阅
func (s *DomainServiceImpl) ValidateSubscribe(subscribe *domains.Subscribe) error {
	logger.Debug("验证订阅", "subscribe_id", subscribe.ID, "name", subscribe.Name)
	
	if subscribe.Name == "" {
		return errors.New("订阅名称不能为空")
	}
	
	if subscribe.UserID == 0 {
		return errors.New("用户ID不能为空")
	}
	
	if subscribe.MediaID == 0 {
		return errors.New("媒体ID不能为空")
	}
	
	// 使用订阅策略验证
	if err := s.policyManager.GetSubscriptionPolicy().ValidateSubscription(subscribe, []domains.Subscribe{}); err != nil {
		return err
	}
	
	logger.Info("订阅验证通过", "subscribe_id", subscribe.ID)
	return nil
}

// ValidateTransfer 验证转移
func (s *DomainServiceImpl) ValidateTransfer(transfer *domains.Transfer) error {
	logger.Debug("验证转移", "transfer_id", transfer.ID, "source", transfer.SourcePath)
	
	if transfer.SourcePath == "" {
		return errors.New("源路径不能为空")
	}
	
	if transfer.TargetPath == "" {
		return errors.New("目标路径不能为空")
	}
	
	if transfer.UserID == 0 {
		return errors.New("用户ID不能为空")
	}
	
	logger.Info("转移验证通过", "transfer_id", transfer.ID)
	return nil
}

// EventService 事件服务接口
type EventService interface {
	PublishEvent(ctx context.Context, event *domains.Event) error
	SubscribeToEvents(ctx context.Context, eventType string, handler EventHandler) error
	UnsubscribeFromEvents(ctx context.Context, eventType string, handler EventHandler) error
}

// EventHandler 事件处理器
type EventHandler func(ctx context.Context, event *domains.Event) error

// Event 事件领域实体
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]string      `json:"metadata"`
}

// EventServiceImpl 事件服务实现
type EventServiceImpl struct {
	handlers map[string][]EventHandler
}

// NewEventService 创建事件服务
func NewEventService() EventService {
	return &EventServiceImpl{
		handlers: make(map[string][]EventHandler),
	}
}

// PublishEvent 发布事件
func (s *EventServiceImpl) PublishEvent(ctx context.Context, event *domains.Event) error {
	logger.Info("发布事件", "event_type", event.Type, "event_id", event.ID)
	
	if handlers, exists := s.handlers[event.Type]; exists {
		for _, handler := range handlers {
			if err := handler(ctx, event); err != nil {
				logger.Error("事件处理失败", "error", err.Error(), "event_id", event.ID)
				// 继续处理其他处理器，不中断
			}
		}
	}
	
	return nil
}

// SubscribeToEvents 订阅事件
func (s *EventServiceImpl) SubscribeToEvents(ctx context.Context, eventType string, handler EventHandler) error {
	logger.Debug("订阅事件", "event_type", eventType)
	
	if s.handlers[eventType] == nil {
		s.handlers[eventType] = make([]EventHandler, 0)
	}
	
	s.handlers[eventType] = append(s.handlers[eventType], handler)
	return nil
}

// UnsubscribeFromEvents 取消订阅事件
func (s *EventServiceImpl) UnsubscribeFromEvents(ctx context.Context, eventType string, handler EventHandler) error {
	logger.Debug("取消订阅事件", "event_type", eventType)
	
	if handlers, exists := s.handlers[eventType]; exists {
		for i, h := range handlers {
			// 简单的引用比较，实际可能需要更复杂的逻辑
			if &h == &handler {
				s.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
				break
			}
		}
	}
	
	return nil
}