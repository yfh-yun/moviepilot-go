package notification

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	
	"moviepilot-go/internal/integration/notification/channels"
)

// NotificationService 通知服务
type NotificationService struct {
	providers map[string]NotificationProvider
	config    *NotificationConfig
	queue     chan *NotificationMessage
	history   []NotificationHistory
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
	logger    *zap.Logger
}

// NewNotificationService 创建通知服务
func NewNotificationService(logger *zap.Logger) *NotificationService {
	ctx, cancel := context.WithCancel(context.Background())
	
	service := &NotificationService{
		providers: make(map[string]NotificationProvider),
		queue:     make(chan *NotificationMessage, 1000),
		history:   make([]NotificationHistory, 0),
		ctx:       ctx,
		cancel:    cancel,
		logger:    logger,
	}
	
	// 启动队列处理
	go service.processQueue()
	
	return service
}

// LoadConfig 加载配置
func (s *NotificationService) LoadConfig(config *NotificationConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.config = config
	
	// 初始化提供商
	for name, providerConfig := range config.Providers {
		if !providerConfig.Enabled {
			continue
		}
		
		provider, err := s.createProvider(providerConfig.Type)
		if err != nil {
			s.logger.Error("创建通知提供商失败", 
				zap.String("provider", name),
				zap.String("type", providerConfig.Type),
				zap.Error(err))
			continue
		}
		
		if err := provider.Initialize(providerConfig.Config); err != nil {
			s.logger.Error("初始化通知提供商失败",
				zap.String("provider", name),
				zap.Error(err))
			continue
		}
		
		s.providers[name] = provider
		s.logger.Info("通知提供商初始化成功",
			zap.String("provider", name),
			zap.String("type", providerConfig.Type))
	}
	
	return nil
}

// Send 发送通知
func (s *NotificationService) Send(ctx context.Context, message *NotificationMessage) error {
	// 设置时间戳
	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}
	
	// 应用通知规则
	rules := s.applyRules(message)
	if len(rules) == 0 {
		s.logger.Debug("通知消息未匹配任何规则", zap.String("message_id", message.ID))
		return nil
	}
	
	// 根据规则选择提供商
	providers := s.selectProviders(rules)
	if len(providers) == 0 {
		return fmt.Errorf("没有可用的通知提供商")
	}
	
	// 发送到选定的提供商
	for _, providerName := range providers {
		provider, exists := s.providers[providerName]
		if !exists {
			s.logger.Warn("通知提供商不存在", zap.String("provider", providerName))
			continue
		}
		
		if err := provider.Send(ctx, message); err != nil {
			s.logger.Error("发送通知失败",
				zap.String("provider", providerName),
				zap.String("message_id", message.ID),
				zap.Error(err))
			
			// 记录失败状态
			s.recordDeliveryStatus(message.ID, providerName, "failed", err.Error())
		} else {
			s.logger.Info("通知发送成功",
				zap.String("provider", providerName),
				zap.String("message_id", message.ID))
			
			// 记录成功状态
			s.recordDeliveryStatus(message.ID, providerName, "sent", "")
		}
	}
	
	return nil
}

// SendAsync 异步发送通知
func (s *NotificationService) SendAsync(message *NotificationMessage) error {
	select {
	case s.queue <- message:
		return nil
	default:
		return fmt.Errorf("通知队列已满")
	}
}

// createProvider 创建通知提供商
func (s *NotificationService) createProvider(providerType string) (NotificationProvider, error) {
	switch providerType {
	case "slack":
		return providers.NewSlackProvider(s.logger), nil
	case "telegram":
		return providers.NewTelegramProvider(s.logger), nil
	case "wechat":
		return providers.NewWeChatProvider(s.logger), nil
	case "synologychat":
		return providers.NewSynologyChatProvider(s.logger), nil
	case "vocechat":
		return providers.NewVoceChatProvider(s.logger), nil
	default:
		return nil, fmt.Errorf("不支持的通知提供商类型: %s", providerType)
	}
}

// applyRules 应用通知规则
func (s *NotificationService) applyRules(message *NotificationMessage) []NotificationRule {
	var matchedRules []NotificationRule
	
	if s.config == nil {
		return matchedRules
	}
	
	for _, rule := range s.config.Rules {
		if !rule.Enabled {
			continue
		}
		
		// 检查规则条件是否匹配
		if s.matchRule(rule, message) {
			matchedRules = append(matchedRules, rule)
		}
	}
	
	return matchedRules
}

// matchRule 检查规则匹配
func (s *NotificationService) matchRule(rule NotificationRule, message *NotificationMessage) bool {
	// 简单的规则匹配逻辑
	// 在实际应用中可以实现更复杂的条件判断
	
	// 检查级别匹配
	if level, exists := rule.Conditions["level"]; exists {
		if level != string(message.Level) {
			return false
		}
	}
	
	// 检查渠道匹配
	if channel, exists := rule.Conditions["channel"]; exists {
		if channel != message.Channel {
			return false
		}
	}
	
	// 检查用户匹配
	if userID, exists := rule.Conditions["user_id"]; exists {
		if userID != message.UserID {
			return false
		}
	}
	
	return true
}

// selectProviders 根据规则选择提供商
func (s *NotificationService) selectProviders(rules []NotificationRule) []string {
	providerSet := make(map[string]bool)
	
	for _, rule := range rules {
		for _, provider := range rule.Providers {
			providerSet[provider] = true
		}
	}
	
	var providers []string
	for provider := range providerSet {
		if _, exists := s.providers[provider]; exists {
			providers = append(providers, provider)
		}
	}
	
	return providers
}

// processQueue 处理队列
func (s *NotificationService) processQueue() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case message := <-s.queue:
			ctx, cancel := context.WithTimeout(s.ctx, 30*time.Second)
			if err := s.Send(ctx, message); err != nil {
				s.logger.Error("异步发送通知失败", zap.Error(err))
			}
			cancel()
		}
	}
}

// recordDeliveryStatus 记录投递状态
func (s *NotificationService) recordDeliveryStatus(messageID, provider, status, error string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	deliveryStatus := &DeliveryStatus{
		MessageID:  messageID,
		Provider:   provider,
		Status:     status,
		Error:      error,
		SentAt:     time.Now(),
		RetryCount: 0,
	}
	
	// 查找或创建历史记录
	for i, history := range s.history {
		if history.Message.ID == messageID {
			s.history[i].Status = append(s.history[i].Status, deliveryStatus)
			s.history[i].UpdatedAt = time.Now()
			return
		}
	}
	
	// 创建新的历史记录
	history := NotificationHistory{
		ID:        messageID,
		Message:   &NotificationMessage{ID: messageID},
		Status:    []*DeliveryStatus{deliveryStatus},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.history = append(s.history, history)
}

// GetProviders 获取所有提供商
func (s *NotificationService) GetProviders() map[string]NotificationProvider {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	result := make(map[string]NotificationProvider)
	for name, provider := range s.providers {
		result[name] = provider
	}
	
	return result
}

// GetHistory 获取通知历史
func (s *NotificationService) GetHistory(limit int) []NotificationHistory {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if limit <= 0 || limit > len(s.history) {
		limit = len(s.history)
	}
	
	// 返回最新的历史记录
	start := len(s.history) - limit
	if start < 0 {
		start = 0
	}
	
	result := make([]NotificationHistory, limit)
	copy(result, s.history[start:])
	
	return result
}

// HealthCheck 健康检查
func (s *NotificationService) HealthCheck(ctx context.Context) map[string]bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	results := make(map[string]bool)
	
	for name, provider := range s.providers {
		results[name] = provider.IsHealthy(ctx)
	}
	
	return results
}

// Close 关闭服务
func (s *NotificationService) Close() error {
	s.cancel()
	
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// 关闭所有提供商
	for name, provider := range s.providers {
		if err := provider.Close(); err != nil {
			s.logger.Error("关闭通知提供商失败",
				zap.String("provider", name),
				zap.Error(err))
		}
	}
	
	return nil
}