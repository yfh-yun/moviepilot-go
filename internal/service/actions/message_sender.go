// Package actions 提供动作系统的业务逻辑实现
package actions

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/repository/interfaces"
	"github.com/yfh-yun/moviepilot-go/internal/service/actions/types"

	"go.uber.org/zap"
)

// MessageSender 消息发送器
// 负责处理消息的发送、路由和管理，实现Python版本SendMessageAction的完整功能
type MessageSender struct {
	messageRepo interfaces.MessageRepository
	userRepo    interfaces.UserRepository
	cache       *WorkflowCache
	logger      *zap.Logger

	// 消息渠道配置
	channels map[string]MessageChannel
}

// MessageChannel 消息渠道接口
type MessageChannel interface {
	Send(ctx context.Context, message *MessageRequest) error
	GetType() string
	IsEnabled() bool
	GetConfig() map[string]interface{}
}

// MessageRequest 消息请求
type MessageRequest struct {
	Title     string            `json:"title"`
	Content   string            `json:"content"`
	Type      string            `json:"type"`    // info, warning, error, success
	Channel   string            `json:"channel"` // wechat, telegram, email, webhook
	UserIDs   []uint            `json:"user_ids"`
	Username  string            `json:"username"`
	Template  string            `json:"template"`
	Variables map[string]string `json:"variables"`
	Priority  int               `json:"priority"` // 1-10
	Async     bool              `json:"async"`    // 是否异步发送
	Retry     int               `json:"retry"`    // 重试次数
	Timeout   int               `json:"timeout"`  // 超时时间（秒）
}

// SendMessageParams 发送消息参数
type SendMessageParams struct {
	Clients  []string `json:"clients" description:"消息渠道"`
	UserID   string   `json:"userid" description:"用户ID"`
	Title    string   `json:"title" description:"消息标题"`
	Content  string   `json:"content" description:"消息内容"`
	Type     string   `json:"type" description:"消息类型"`
	Template string   `json:"template" description:"消息模板"`
	Async    bool     `json:"async" description:"是否异步发送"`
	Priority int      `json:"priority" description:"优先级"`
}

// SendMessageResult 发送消息结果
type SendMessageResult struct {
	Success        bool          `json:"success"`
	MessageID      int           `json:"message_id"`
	Channel        string        `json:"channel"`
	UserIDs        []uint        `json:"user_ids"`
	Title          string        `json:"title"`
	Content        string        `json:"content"`
	Type           string        `json:"type"`
	Status         string        `json:"status"`
	SentAt         *time.Time    `json:"sent_at,omitempty"`
	Error          error         `json:"error,omitempty"`
	ProcessingTime time.Duration `json:"processing_time"`
	FailedUsers    []uint        `json:"failed_users,omitempty"`
}

// NewMessageSender 创建消息发送器实例
func NewMessageSender(
	messageRepo interfaces.MessageRepository,
	userRepo interfaces.UserRepository,
	cache *WorkflowCache,
) *MessageSender {
	ms := &MessageSender{
		messageRepo: messageRepo,
		userRepo:    userRepo,
		cache:       cache,
		logger:      logger.Logger,
		channels:    make(map[string]MessageChannel),
	}

	// 初始化默认消息渠道
	ms.initDefaultChannels()

	return ms
}

// initDefaultChannels 初始化默认消息渠道
func (ms *MessageSender) initDefaultChannels() {
	// 这里可以初始化各种消息渠道
	// 例如：微信、Telegram、邮件、Webhook等
	// 暂时留空，后续可以添加具体的渠道实现
}

// SendMessage 发送消息
// 实现Python版本SendMessageAction的完整功能
func (ms *MessageSender) SendMessage(
	ctx context.Context,
	workflowID int64,
	params *SendMessageParams,
	messages []*types.Message,
) ([]*SendMessageResult, error) {
	startTime := time.Now()
	results := make([]*SendMessageResult, 0)

	ms.logger.Info("开始发送消息",
		zap.Int64("workflow_id", workflowID),
		zap.Strings("clients", params.Clients),
		zap.String("user_id", params.UserID),
		zap.Int("message_count", len(messages)),
	)

	// 确定目标用户
	userIDs, err := ms.determineTargetUsers(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("确定目标用户失败: %w", err)
	}

	// 确定消息渠道
	channels := ms.determineChannels(params)

	// 处理每条消息
	for _, message := range messages {
		// 检查工作流是否已停止
		if ms.isWorkflowStopped(ctx, workflowID) {
			ms.logger.Info("工作流已停止，终止消息发送", zap.Int64("workflow_id", workflowID))
			break
		}

		// 处理单条消息
		messageResults := ms.processMessage(ctx, workflowID, params, message, userIDs, channels)
		results = append(results, messageResults...)
	}

	// 统计结果
	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		}
	}

	ms.logger.Info("消息发送完成",
		zap.Int64("workflow_id", workflowID),
		zap.Int("total_messages", len(messages)),
		zap.Int("total_sends", len(results)),
		zap.Int("success_count", successCount),
		zap.Duration("processing_time", time.Since(startTime)),
	)

	return results, nil
}

// determineTargetUsers 确定目标用户
func (ms *MessageSender) determineTargetUsers(ctx context.Context, params *SendMessageParams) ([]uint, error) {
	var userIDs []uint

	if params.UserID != "" {
		// 指定了用户ID
		if strings.Contains(params.UserID, ",") {
			// 多个用户ID
			userIDStrs := strings.Split(params.UserID, ",")
			for _, userIDStr := range userIDStrs {
				var userID uint
				if _, err := fmt.Sscanf(strings.TrimSpace(userIDStr), "%d", &userID); err == nil {
					userIDs = append(userIDs, userID)
				}
			}
		} else {
			// 单个用户ID
			var userID uint
			if _, err := fmt.Sscanf(params.UserID, "%d", &userID); err == nil {
				userIDs = append(userIDs, userID)
			}
		}
	} else {
		// 未指定用户，发送给所有活跃用户
		users, err := ms.userRepo.List(ctx, 1, 1000)
		if err != nil {
			return nil, fmt.Errorf("获取用户列表失败: %w", err)
		}

		for _, user := range users {
			if user.IsActive {
				userIDs = append(userIDs, user.ID)
			}
		}
	}

	return userIDs, nil
}

// determineChannels 确定消息渠道
func (ms *MessageSender) determineChannels(params *SendMessageParams) []string {
	var channels []string

	if len(params.Clients) > 0 {
		// 指定了消息渠道
		channels = params.Clients
	} else {
		// 使用所有可用的消息渠道
		for channelType, channel := range ms.channels {
			if channel.IsEnabled() {
				channels = append(channels, channelType)
			}
		}
	}

	return channels
}

// processMessage 处理单条消息
func (ms *MessageSender) processMessage(
	ctx context.Context,
	workflowID int64,
	params *SendMessageParams,
	message *types.Message,
	userIDs []uint,
	channels []string,
) []*SendMessageResult {
	results := make([]*SendMessageResult, 0, len(channels))

	// 准备消息内容
	title := message.Title
	content := message.Content
	messageType := message.Type

	if params.Title != "" {
		title = params.Title
	}
	if params.Content != "" {
		content = params.Content
	}
	if params.Type != "" {
		messageType = params.Type
	}

	// 应用模板
	if params.Template != "" {
		title, content = ms.applyTemplate(params.Template, title, content, map[string]string{
			"progress": fmt.Sprintf("%d%%", 0), // 这里可以从上下文获取进度
		})
	}

	// 为每个渠道发送消息
	for _, channelType := range channels {
		result := ms.sendToChannel(ctx, workflowID, channelType, title, content, messageType, userIDs, params)
		results = append(results, result)
	}

	return results
}

// sendToChannel 发送到指定渠道
func (ms *MessageSender) sendToChannel(
	ctx context.Context,
	workflowID int64,
	channelType string,
	title string,
	content string,
	messageType string,
	userIDs []uint,
	params *SendMessageParams,
) *SendMessageResult {
	startTime := time.Now()

	// 检查渠道是否存在
	channel, exists := ms.channels[channelType]
	if !exists {
		return &SendMessageResult{
			Success:        false,
			Channel:        channelType,
			UserIDs:        userIDs,
			Title:          title,
			Content:        content,
			Type:           messageType,
			Status:         "failed",
			Error:          fmt.Errorf("消息渠道 %s 不存在", channelType),
			ProcessingTime: time.Since(startTime),
		}
	}

	// 创建消息记录
	message := &types.Message{
		Title:     title,
		Content:   content,
		Type:      messageType,
		Channel:   channelType,
		Status:    "pending",
		UserID:    ms.getPrimaryUserID(userIDs),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 保存消息到数据库
	if err := ms.messageRepo.Create(ctx, message); err != nil {
		return &SendMessageResult{
			Success:        false,
			MessageID:      message.ID,
			Channel:        channelType,
			UserIDs:        userIDs,
			Title:          title,
			Content:        content,
			Type:           messageType,
			Status:         "failed",
			Error:          fmt.Errorf("保存消息失败: %w", err),
			ProcessingTime: time.Since(startTime),
		}
	}

	// 准备发送请求
	request := &MessageRequest{
		Title:    title,
		Content:  content,
		Type:     messageType,
		Channel:  channelType,
		UserIDs:  userIDs,
		Template: params.Template,
		Priority: params.Priority,
		Async:    params.Async,
		Retry:    3,  // 默认重试3次
		Timeout:  30, // 默认超时30秒
	}

	// 发送消息
	err := channel.Send(ctx, request)
	if err != nil {
		// 更新消息状态为失败
		message.Status = "failed"
		message.ErrorMsg = err.Error()
		ms.messageRepo.Update(ctx, message)

		return &SendMessageResult{
			Success:        false,
			MessageID:      message.ID,
			Channel:        channelType,
			UserIDs:        userIDs,
			Title:          title,
			Content:        content,
			Type:           messageType,
			Status:         "failed",
			Error:          err,
			ProcessingTime: time.Since(startTime),
		}
	}

	// 更新消息状态为已发送
	now := time.Now()
	message.Status = "sent"
	message.SentAt = &now
	ms.messageRepo.Update(ctx, message)

	return &SendMessageResult{
		Success:        true,
		MessageID:      message.ID,
		Channel:        channelType,
		UserIDs:        userIDs,
		Title:          title,
		Content:        content,
		Type:           messageType,
		Status:         "sent",
		SentAt:         &now,
		ProcessingTime: time.Since(startTime),
	}
}

// applyTemplate 应用消息模板
func (ms *MessageSender) applyTemplate(template string, title string, content string, variables map[string]string) (string, string) {
	// 这里可以实现模板引擎
	// 暂时简单替换变量
	newTitle := title
	newContent := content

	for key, value := range variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		newTitle = strings.ReplaceAll(newTitle, placeholder, value)
		newContent = strings.ReplaceAll(newContent, placeholder, value)
	}

	return newTitle, newContent
}

// getPrimaryUserID 获取主要用户ID
func (ms *MessageSender) getPrimaryUserID(userIDs []uint) uint {
	if len(userIDs) > 0 {
		return userIDs[0]
	}
	return 0
}

// isWorkflowStopped 检查工作流是否已停止
func (ms *MessageSender) isWorkflowStopped(ctx context.Context, workflowID int64) bool {
	// 这里应该检查工作流状态
	// 暂时返回false
	return false
}

// RegisterChannel 注册消息渠道
func (ms *MessageSender) RegisterChannel(channel MessageChannel) {
	ms.channels[channel.GetType()] = channel
	ms.logger.Info("注册消息渠道", zap.String("type", channel.GetType()))
}

// UnregisterChannel 注销消息渠道
func (ms *MessageSender) UnregisterChannel(channelType string) {
	if _, exists := ms.channels[channelType]; exists {
		delete(ms.channels, channelType)
		ms.logger.Info("注销消息渠道", zap.String("type", channelType))
	}
}

// GetChannels 获取所有消息渠道
func (ms *MessageSender) GetChannels() map[string]MessageChannel {
	return ms.channels
}

// GetChannel 获取指定消息渠道
func (ms *MessageSender) GetChannel(channelType string) (MessageChannel, bool) {
	channel, exists := ms.channels[channelType]
	return channel, exists
}

// SendProgressMessage 发送进度消息
func (ms *MessageSender) SendProgressMessage(
	ctx context.Context,
	workflowID int64,
	progress int,
	message string,
	userIDs []uint,
) error {
	params := &SendMessageParams{
		Clients: []string{"system"}, // 使用系统渠道
		Title:   "工作流进度",
		Content: fmt.Sprintf("当前进度：%d%%\n%s", progress, message),
		Type:    "info",
		Async:   true,
	}

	// 创建临时消息
	tempMessage := &types.Message{
		Title:   params.Title,
		Content: params.Content,
		Type:    params.Type,
	}

	results, err := ms.SendMessage(ctx, workflowID, params, []*types.Message{tempMessage})
	if err != nil {
		return err
	}

	// 检查是否全部成功
	for _, result := range results {
		if !result.Success {
			return result.Error
		}
	}

	return nil
}

// SendErrorMessage 发送错误消息
func (ms *MessageSender) SendErrorMessage(
	ctx context.Context,
	workflowID int64,
	errorMsg string,
	userIDs []uint,
) error {
	params := &SendMessageParams{
		Clients:  []string{"system"},
		Title:    "工作流错误",
		Content:  fmt.Sprintf("执行过程中发生错误：%s", errorMsg),
		Type:     "error",
		Priority: 10,    // 高优先级
		Async:    false, // 同步发送确保错误消息及时送达
	}

	tempMessage := &types.Message{
		Title:   params.Title,
		Content: params.Content,
		Type:    params.Type,
	}

	results, err := ms.SendMessage(ctx, workflowID, params, []*types.Message{tempMessage})
	if err != nil {
		return err
	}

	for _, result := range results {
		if !result.Success {
			return result.Error
		}
	}

	return nil
}

// SendSuccessMessage 发送成功消息
func (ms *MessageSender) SendSuccessMessage(
	ctx context.Context,
	workflowID int64,
	message string,
	userIDs []uint,
) error {
	params := &SendMessageParams{
		Clients:  []string{"system"},
		Title:    "工作流完成",
		Content:  fmt.Sprintf("工作流执行成功：%s", message),
		Type:     "success",
		Priority: 5,
		Async:    true,
	}

	tempMessage := &types.Message{
		Title:   params.Title,
		Content: params.Content,
		Type:    params.Type,
	}

	results, err := ms.SendMessage(ctx, workflowID, params, []*types.Message{tempMessage})
	if err != nil {
		return err
	}

	for _, result := range results {
		if !result.Success {
			return result.Error
		}
	}

	return nil
}
