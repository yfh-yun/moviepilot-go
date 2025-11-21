package chain

import (
	"context"
	"fmt"

	"moviepilot-go/internal/models"

	"go.uber.org/zap"
)

// PostMessage 发送消息
func (mc *MessageChain) PostMessage(ctx context.Context, notification *models.Notification, options ...MessageOption) error {
	mc.logger.Info("开始发送消息",
		zap.String("title", notification.Title),
		zap.String("type", notification.Type))

	// 处理消息选项
	opts := &MessageOptions{}
	for _, opt := range options {
		opt(opts)
	}

	// 渲染消息内容
	renderedMessage, err := mc.renderMessage(notification, opts)
	if err != nil {
		mc.logger.Error("渲染消息失败", "error", err)
		return fmt.Errorf("渲染消息失败: %w", err)
	}

	// 保存消息到数据库
	if opts.saveToDB {
		err = mc.saveMessageToDB(ctx, notification, renderedMessage)
		if err != nil {
			mc.logger.Warn("保存消息到数据库失败", "error", err)
		}
	}

	// 发送消息到各个渠道
	err = mc.sendMessageToChannels(ctx, notification, renderedMessage, opts)
	if err != nil {
		mc.logger.Error("发送消息到渠道失败", "error", err)
		return fmt.Errorf("发送消息到渠道失败: %w", err)
	}

	mc.logger.Info("消息发送完成", "title", notification.Title)
	return nil
}

// PostMediaList 发送媒体列表
func (mc *MessageChain) PostMediaList(ctx context.Context, notification *models.Notification, mediaList []*models.MediaInfo, options ...MessageOption) error {
	mc.logger.Info("开始发送媒体列表",
		zap.String("title", notification.Title),
		zap.Int("count", len(mediaList)))

	// 使用媒体列表模板渲染消息
	renderedMessage, err := mc.templateHelper.RenderMediaList(notification, mediaList)
	if err != nil {
		mc.logger.Error("渲染媒体列表消息失败", "error", err)
		return fmt.Errorf("渲染媒体列表消息失败: %w", err)
	}

	// 处理消息选项
	opts := &MessageOptions{}
	for _, opt := range options {
		opt(opts)
	}

	// 保存消息到数据库
	if opts.saveToDB {
		err = mc.saveMediaListToDB(ctx, notification, mediaList, renderedMessage)
		if err != nil {
			mc.logger.Warn("保存媒体列表到数据库失败", "error", err)
		}
	}

	// 发送消息到各个渠道
	err = mc.sendMessageToChannels(ctx, notification, renderedMessage, opts)
	if err != nil {
		mc.logger.Error("发送媒体列表到渠道失败", "error", err)
		return fmt.Errorf("发送媒体列表到渠道失败: %w", err)
	}

	mc.logger.Info("媒体列表发送完成", "count", len(mediaList))
	return nil
}

// PostTorrentList 发送种子列表
func (mc *MessageChain) PostTorrentList(ctx context.Context, notification *models.Notification, torrentList []*models.TorrentInfo, options ...MessageOption) error {
	mc.logger.Info("开始发送种子列表",
		zap.String("title", notification.Title),
		zap.Int("count", len(torrentList)))

	// 使用种子列表模板渲染消息
	renderedMessage, err := mc.templateHelper.RenderTorrentList(notification, torrentList)
	if err != nil {
		mc.logger.Error("渲染种子列表消息失败", "error", err)
		return fmt.Errorf("渲染种子列表消息失败: %w", err)
	}

	// 处理消息选项
	opts := &MessageOptions{}
	for _, opt := range options {
		opt(opts)
	}

	// 保存消息到数据库
	if opts.saveToDB {
		err = mc.saveTorrentListToDB(ctx, notification, torrentList, renderedMessage)
		if err != nil {
			mc.logger.Warn("保存种子列表到数据库失败", "error", err)
		}
	}

	// 发送消息到各个渠道
	err = mc.sendMessageToChannels(ctx, notification, renderedMessage, opts)
	if err != nil {
		mc.logger.Error("发送种子列表到渠道失败", "error", err)
		return fmt.Errorf("发送种子列表到渠道失败: %w", err)
	}

	mc.logger.Info("种子列表发送完成", "count", len(torrentList))
	return nil
}

// DeleteMessage 删除消息
func (mc *MessageChain) DeleteMessage(ctx context.Context, messageID string, channel string) error {
	mc.logger.Info("开始删除消息",
		zap.String("message_id", messageID),
		zap.String("channel", channel))

	// 从数据库删除消息
	err := mc.messageRepo.DeleteMessage(ctx, messageID)
	if err != nil {
		mc.logger.Error("从数据库删除消息失败", "error", err)
		return fmt.Errorf("从数据库删除消息失败: %w", err)
	}

	// 从各个渠道删除消息
	err = mc.deleteMessageFromChannels(ctx, messageID, channel)
	if err != nil {
		mc.logger.Error("从渠道删除消息失败", "error", err)
		return fmt.Errorf("从渠道删除消息失败: %w", err)
	}

	mc.logger.Info("消息删除完成", "message_id", messageID)
	return nil
}

// renderMessage 渲染消息
func (mc *MessageChain) renderMessage(notification *models.Notification, opts *MessageOptions) (string, error) {
	switch notification.Type {
	case "media_notification":
		if notification.MediaInfo != nil {
			return mc.templateHelper.RenderMediaNotification(notification, notification.MediaInfo)
		}
	case "torrent_notification":
		if notification.TorrentInfo != nil {
			return mc.templateHelper.RenderTorrentNotification(notification, notification.TorrentInfo)
		}
	case "transfer_notification":
		if notification.TransferInfo != nil {
			return mc.templateHelper.RenderTransferNotification(notification, notification.TransferInfo)
		}
	default:
		// 通用消息渲染
		return mc.renderGenericMessage(notification, opts)
	}

	return notification.Content, nil
}

// renderGenericMessage 渲染通用消息
func (mc *MessageChain) renderGenericMessage(notification *models.Notification, opts *MessageOptions) (string, error) {
	if opts.useHTML {
		return mc.templateHelper.RenderHTML("generic_notification", notification)
	}
	return mc.templateHelper.Render("generic_notification", notification)
}

// saveMessageToDB 保存消息到数据库
func (mc *MessageChain) saveMessageToDB(ctx context.Context, notification *models.Notification, renderedMessage string) error {
	// 构建消息记录
	messageRecord := &models.MessageRecord{
		Title:       notification.Title,
		Content:     renderedMessage,
		Type:        notification.Type,
		UserID:      notification.UserID,
		Username:    notification.Username,
		Channels:    notification.Channels,
		Priority:    notification.Priority,
		CreatedAt:   notification.CreatedAt,
	}

	return mc.messageRepo.CreateMessage(ctx, messageRecord)
}

// saveMediaListToDB 保存媒体列表到数据库
func (mc *MessageChain) saveMediaListToDB(ctx context.Context, notification *models.Notification, mediaList []*models.MediaInfo, renderedMessage string) error {
	// 构建消息记录
	messageRecord := &models.MessageRecord{
		Title:       notification.Title,
		Content:     renderedMessage,
		Type:        "media_list",
		UserID:      notification.UserID,
		Username:    notification.Username,
		Channels:    notification.Channels,
		Priority:    notification.Priority,
		CreatedAt:   notification.CreatedAt,
		MediaCount:  len(mediaList),
	}

	return mc.messageRepo.CreateMessage(ctx, messageRecord)
}

// saveTorrentListToDB 保存种子列表到数据库
func (mc *MessageChain) saveTorrentListToDB(ctx context.Context, notification *models.Notification, torrentList []*models.TorrentInfo, renderedMessage string) error {
	// 构建消息记录
	messageRecord := &models.MessageRecord{
		Title:         notification.Title,
		Content:       renderedMessage,
		Type:          "torrent_list",
		UserID:        notification.UserID,
		Username:      notification.Username,
		Channels:      notification.Channels,
		Priority:      notification.Priority,
		CreatedAt:     notification.CreatedAt,
		TorrentCount:  len(torrentList),
	}

	return mc.messageRepo.CreateMessage(ctx, messageRecord)
}

// sendMessageToChannels 发送消息到各个渠道
func (mc *MessageChain) sendMessageToChannels(ctx context.Context, notification *models.Notification, renderedMessage string, opts *MessageOptions) error {
	// 获取目标用户的消息渠道
	userMessageChannels, err := mc.getUserMessageChannels(ctx, notification.UserID)
	if err != nil {
		mc.logger.Warn("获取用户消息渠道失败", "user_id", notification.UserID, "error", err)
		return err
	}

	// 如果指定了特定渠道，则只使用指定渠道
	if len(notification.Channels) > 0 {
		userMessageChannels = notification.Channels
	}

	// 遍历所有渠道发送消息
	for _, channel := range userMessageChannels {
		err := mc.sendMessageToChannel(ctx, channel, notification, renderedMessage, opts)
		if err != nil {
			mc.logger.Error("发送消息到渠道失败",
				zap.String("channel", channel),
				zap.Error(err))
			// 继续发送到其他渠道，不因为一个渠道失败而停止
			continue
		}
	}

	return nil
}

// sendMessageToChannel 发送消息到指定渠道
func (mc *MessageChain) sendMessageToChannel(ctx context.Context, channel string, notification *models.Notification, renderedMessage string, opts *MessageOptions) error {
	// 调用插件系统发送消息到指定渠道
	return mc.pluginMgr.SendMessage(ctx, &SendMessageRequest{
		Channel:    channel,
		Title:      notification.Title,
		Content:    renderedMessage,
		UserID:     notification.UserID,
		Username:   notification.Username,
		Priority:   notification.Priority,
		Immediate:  opts.immediate,
	})
}

// deleteMessageFromChannels 从各个渠道删除消息
func (mc *MessageChain) deleteMessageFromChannels(ctx context.Context, messageID string, channel string) error {
	// 调用插件系统从指定渠道删除消息
	return mc.pluginMgr.DeleteMessage(ctx, &DeleteMessageRequest{
		MessageID: messageID,
		Channel:   channel,
	})
}

// getUserMessageChannels 获取用户的消息渠道
func (mc *MessageChain) getUserMessageChannels(ctx context.Context, userID string) ([]string, error) {
	// 从用户设置中获取消息渠道
	userSettings, err := mc.userRepo.GetUserSettings(ctx, userID)
	if err != nil {
		return nil, err
	}

	if userSettings != nil && len(userSettings.MessageChannels) > 0 {
		return userSettings.MessageChannels, nil
	}

	// 返回默认渠道
	return []string{"web"}, nil
}

// MessageOptions 消息选项
type MessageOptions struct {
	saveToDB   bool
	immediate  bool
	useHTML    bool
	maxLength  int
}

// MessageOption 消息选项函数
type MessageOption func(*MessageOptions)

// WithSaveToDB 设置保存到数据库
func WithSaveToDB(save bool) MessageOption {
	return func(opts *MessageOptions) {
		opts.saveToDB = save
	}
}

// WithImmediate 设置立即发送
func WithImmediate(immediate bool) MessageOption {
	return func(opts *MessageOptions) {
		opts.immediate = immediate
	}
}

// WithHTML 设置使用HTML格式
func WithHTML(useHTML bool) MessageOption {
	return func(opts *MessageOptions) {
		opts.useHTML = useHTML
	}
}

// WithMaxLength 设置最大长度
func WithMaxLength(maxLength int) MessageOption {
	return func(opts *MessageOptions) {
		opts.maxLength = maxLength
	}
}

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	Channel   string `json:"channel"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Priority  int    `json:"priority"`
	Immediate bool   `json:"immediate"`
}

// DeleteMessageRequest 删除消息请求
type DeleteMessageRequest struct {
	MessageID string `json:"message_id"`
	Channel   string `json:"channel"`
}