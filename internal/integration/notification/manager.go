package notification

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Manager 管理所有通知渠道
type Manager struct {
	channels map[string]*Channel
	mu       sync.RWMutex
}

// NewManager 创建新的通知管理器
func NewManager() *Manager {
	return &Manager{
		channels: make(map[string]*Channel),
	}
}

// RegisterChannel 注册通知渠道
func (m *Manager) RegisterChannel(channel *Channel) error {
	if channel == nil {
		return fmt.Errorf("channel cannot be nil")
	}

	if channel.Name == "" {
		return fmt.Errorf("channel name cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.channels[channel.Name]; exists {
		return fmt.Errorf("channel %s already exists", channel.Name)
	}

	channel.CreatedAt = time.Now()
	channel.UpdatedAt = time.Now()
	m.channels[channel.Name] = channel

	return nil
}

// UnregisterChannel 注销通知渠道
func (m *Manager) UnregisterChannel(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	channel, exists := m.channels[name]
	if !exists {
		return fmt.Errorf("channel %s not found", name)
	}

	// 关闭发送器
	if channel.Sender != nil {
		channel.Sender.Close()
	}

	delete(m.channels, name)
	return nil
}

// EnableChannel 启用通知渠道
func (m *Manager) EnableChannel(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	channel, exists := m.channels[name]
	if !exists {
		return fmt.Errorf("channel %s not found", name)
	}

	channel.Enabled = true
	channel.UpdatedAt = time.Now()

	return nil
}

// DisableChannel 禁用通知渠道
func (m *Manager) DisableChannel(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	channel, exists := m.channels[name]
	if !exists {
		return fmt.Errorf("channel %s not found", name)
	}

	channel.Enabled = false
	channel.UpdatedAt = time.Now()

	return nil
}

// GetChannel 获取指定渠道
func (m *Manager) GetChannel(name string) (*Channel, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	channel, exists := m.channels[name]
	if !exists {
		return nil, fmt.Errorf("channel %s not found", name)
	}

	return channel, nil
}

// ListChannels 列出所有渠道
func (m *Manager) ListChannels() []*Channel {
	m.mu.RLock()
	defer m.mu.RUnlock()

	channels := make([]*Channel, 0, len(m.channels))
	for _, channel := range m.channels {
		channels = append(channels, channel)
	}

	return channels
}

// ListEnabledChannels 列出所有启用的渠道
func (m *Manager) ListEnabledChannels() []*Channel {
	m.mu.RLock()
	defer m.mu.RUnlock()

	enabledChannels := make([]*Channel, 0)
	for _, channel := range m.channels {
		if channel.Enabled {
			enabledChannels = append(enabledChannels, channel)
		}
	}

	return enabledChannels
}

// Send 发送通知到指定渠道
func (m *Manager) Send(ctx context.Context, message *Message) (*NotificationResult, error) {
	if err := message.Validate(); err != nil {
		return nil, err
	}

	channel, err := m.GetChannel(message.Channel)
	if err != nil {
		return nil, err
	}

	if !channel.Enabled {
		return &NotificationResult{
			Channel: message.Channel,
			Success: false,
			Error:   ErrChannelDisabled.Error(),
			SentAt:  time.Now(),
		}, nil
	}

	if channel.Sender == nil {
		return &NotificationResult{
			Channel: message.Channel,
			Success: false,
			Error:   "sender not configured",
			SentAt:  time.Now(),
		}, nil
	}

	// 验证消息级别是否支持
	supportedLevels := channel.Sender.SupportedLevels()
	levelSupported := false
	for _, level := range supportedLevels {
		if level == message.Level {
			levelSupported = true
			break
		}
	}

	if !levelSupported {
		return &NotificationResult{
			Channel: message.Channel,
			Success: false,
			Error:   ErrUnsupportedLevel.Error(),
			SentAt:  time.Now(),
		}, nil
	}

	// 验证消息格式
	if err := channel.Sender.Validate(message); err != nil {
		return &NotificationResult{
			Channel: message.Channel,
			Success: false,
			Error:   err.Error(),
			SentAt:  time.Now(),
		}, nil
	}

	// 健康检查
	if err := channel.Sender.HealthCheck(ctx); err != nil {
		return &NotificationResult{
			Channel: message.Channel,
			Success: false,
			Error:   ErrChannelUnhealthy.Error(),
			SentAt:  time.Now(),
		}, nil
	}

	// 发送消息
	err = channel.Sender.Send(ctx, message)
	result := &NotificationResult{
		Channel: message.Channel,
		SentAt:  time.Now(),
	}

	if err != nil {
		result.Success = false
		result.Error = err.Error()
	} else {
		result.Success = true
		result.MessageID = message.ID
	}

	return result, nil
}

// SendToMultiple 发送消息到多个渠道
func (m *Manager) SendToMultiple(ctx context.Context, message *Message, channels []string) ([]*NotificationResult, error) {
	results := make([]*NotificationResult, 0, len(channels))
	errors := make([]error, 0)

	for _, channelName := range channels {
		// 创建消息副本，避免竞态条件
		channelMessage := &Message{
			ID:         message.ID,
			Title:      message.Title,
			Content:    message.Content,
			Level:      message.Level,
			Priority:   message.Priority,
			Channel:    channelName,
			Category:   message.Category,
			Tags:       append([]string{}, message.Tags...),
			ImageURL:   message.ImageURL,
			LinkURL:    message.LinkURL,
			Timestamp:  message.Timestamp,
			TTL:        message.TTL,
			CustomData: message.CustomData,
		}

		result, err := m.Send(ctx, channelMessage)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to send to %s: %w", channelName, err))
		} else {
			results = append(results, result)
		}
	}

	if len(errors) > 0 {
		return results, fmt.Errorf("multiple errors occurred: %v", errors)
	}

	return results, nil
}

// Broadcast 广播消息到所有启用的渠道
func (m *Manager) Broadcast(ctx context.Context, message *Message) ([]*NotificationResult, error) {
	enabledChannels := m.ListEnabledChannels()
	channelNames := make([]string, 0, len(enabledChannels))

	for _, channel := range enabledChannels {
		channelNames = append(channelNames, channel.Name)
	}

	return m.SendToMultiple(ctx, message, channelNames)
}

// HealthCheck 检查所有渠道的健康状态
func (m *Manager) HealthCheck(ctx context.Context) map[string]error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	healthResults := make(map[string]error)

	for name, channel := range m.channels {
		if !channel.Enabled {
			healthResults[name] = nil // 禁用渠道不检查健康状态
			continue
		}

		if channel.Sender == nil {
			healthResults[name] = fmt.Errorf("sender not configured")
			continue
		}

		if err := channel.Sender.HealthCheck(ctx); err != nil {
			healthResults[name] = err
		} else {
			healthResults[name] = nil
		}
	}

	return healthResults
}

// Close 关闭所有通知渠道
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errors []error

	for name, channel := range m.channels {
		if channel.Sender != nil {
			if err := channel.Sender.Close(); err != nil {
				errors = append(errors, fmt.Errorf("failed to close %s: %w", name, err))
			}
		}
	}

	m.channels = make(map[string]*Channel)

	if len(errors) > 0 {
		return fmt.Errorf("errors occurred while closing channels: %v", errors)
	}

	return nil
}

// Statistics 获取通知统计信息
type Statistics struct {
	TotalChannels   int `json:"total_channels"`
	EnabledChannels int `json:
"enabled_channels"`
	HealthyChannels   int `json:"healthy_channels"`
	UnhealthyChannels int `json:"unhealthy_channels"`
}

// GetStatistics 获取通知系统统计信息
func (m *Manager) GetStatistics(ctx context.Context) (*Statistics, error) {
	healthResults := m.HealthCheck(ctx)

	stats := &Statistics{
		TotalChannels: len(m.channels),
	}

	for _, health := range healthResults {
		if health == nil {
			stats.HealthyChannels++
		} else {
			stats.UnhealthyChannels++
		}
	}

	stats.EnabledChannels = len(m.ListEnabledChannels())

	return stats, nil
}
