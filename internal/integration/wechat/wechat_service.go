// Package wechat 微信服务实现
package wechat

import (
	"context"
	"fmt"
	"time"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/integration/wechat"

	"go.uber.org/zap"
)

// WeChatService 微信服务
type WeChatService struct {
	client *wechat.WeChatClient
	logger *zap.Logger
}

// WeChatServiceConfig 微信服务配置
type WeChatServiceConfig struct {
	CorpID   string `json:"corp_id"`    // 企业ID
	AppSecret string `json:"app_secret"`  // 应用密钥
	AppID     string `json:"app_id"`      // 应用ID
	Proxy     string `json:"proxy"`       // 代理地址
}

// NewWeChatService 创建微信服务
func NewWeChatService(config *WeChatServiceConfig) *WeChatService {
	if config == nil || config.CorpID == "" || config.AppSecret == "" || config.AppID == "" {
		logger.Logger.Error("WeChat config is incomplete")
		return nil
	}

	wechatConfig := &wechat.WeChatConfig{
		CorpID:   config.CorpID,
		AppSecret: config.AppSecret,
		AppID:     config.AppID,
		Proxy:     config.Proxy,
	}

	client := wechat.NewWeChatClient(wechatConfig)
	if client == nil {
		logger.Logger.Error("Failed to create WeChat client")
		return nil
	}

	service := &WeChatService{
		client: client,
		logger: logger.Logger,
	}

	// 初始化状态检查
	if !service.client.GetState() {
		logger.Logger.Error("WeChat client initialization failed")
		return nil
	}

	logger.Logger.Info("WeChat service initialized successfully",
		zap.String("corp_id", config.CorpID),
		zap.String("app_id", config.AppID))

	return service
}

// SendMsg 发送微信消息
// 兼容Python版本的send_msg方法
func (ws *WeChatService) SendMsg(ctx context.Context, title, text, image, userID, link string) (bool, error) {
	if ws.client == nil {
		return false, fmt.Errorf("wechat client is nil")
	}

	ws.logger.Info("Sending WeChat message",
		zap.String("title", title),
		zap.String("user_id", userID),
		zap.Bool("has_image", image != ""),
		zap.Bool("has_link", link != ""))

	// 根据是否有图片选择发送方式
	if image != "" {
		// 发送图文消息
		err := ws.client.SendImageMessage(ctx, title, text, image, userID, link)
		if err != nil {
			ws.logger.Error("Failed to send image message",
				zap.String("title", title),
				zap.Error(err))
			return false, err
		}
	} else {
		// 发送文本消息
		err := ws.client.SendTextMessage(ctx, title, text, userID)
		if err != nil {
			ws.logger.Error("Failed to send text message",
				zap.String("title", title),
				zap.Error(err))
			return false, err
		}
	}

	ws.logger.Info("WeChat message sent successfully",
		zap.String("title", title),
		zap.String("user_id", userID))

	return true, nil
}

// CreateMenus 创建微信菜单
// 兼容Python版本的create_menus方法
func (ws *WeChatService) CreateMenus(ctx context.Context, commands map[string]interface{}) error {
	if ws.client == nil {
		return fmt.Errorf("wechat client is nil")
	}

	// 转换命令格式
	wechatCommands := make(map[string]wechat.Command)
	
	for key, cmdData := range commands {
		if cmdMap, ok := cmdData.(map[string]interface{}); ok {
			cmd := wechat.Command{
				Function:    getString(cmdMap, "function"),
				Description: getString(cmdMap, "description"),
				Category:    getString(cmdMap, "category"),
				Data:        getDataMap(cmdMap, "data"),
			}
			wechatCommands[key] = cmd
		}
	}

	// 创建菜单
	err := ws.client.CreateMenus(ctx, wechatCommands)
	if err != nil {
		ws.logger.Error("Failed to create WeChat menus",
			zap.Error(err))
		return fmt.Errorf("create menus failed: %w", err)
	}

	ws.logger.Info("WeChat menus created successfully",
		zap.Int("command_count", len(wechatCommands)))

	return nil
}

// DeleteMenus 删除微信菜单
func (ws *WeChatService) DeleteMenus(ctx context.Context) error {
	if ws.client == nil {
		return fmt.Errorf("wechat client is nil")
	}

	err := ws.client.DeleteMenus(ctx)
	if err != nil {
		ws.logger.Error("Failed to delete WeChat menus",
			zap.Error(err))
		return fmt.Errorf("delete menus failed: %w", err)
	}

	ws.logger.Info("WeChat menus deleted successfully")
	return nil
}

// SendMediaMessages 发送媒体列表消息
func (ws *WeChatService) SendMediaMessages(ctx context.Context, medias []wechat.MediaInfo, userID string) (bool, error) {
	if ws.client == nil {
		return false, fmt.Errorf("wechat client is nil")
	}

	if len(medias) == 0 {
		return false, fmt.Errorf("media list is empty")
	}

	ws.logger.Info("Sending media messages",
		zap.String("user_id", userID),
		zap.Int("media_count", len(medias)))

	err := ws.client.SendMediaMessages(ctx, medias, userID)
	if err != nil {
		ws.logger.Error("Failed to send media messages",
			zap.String("user_id", userID),
			zap.Error(err))
		return false, err
	}

	ws.logger.Info("Media messages sent successfully",
		zap.String("user_id", userID),
		zap.Int("media_count", len(medias)))

	return true, nil
}

// GetState 获取微信服务状态
func (ws *WeChatService) GetState() bool {
	if ws.client == nil {
		return false
	}
	return ws.client.GetState()
}

// HealthCheck 健康检查
func (ws *WeChatService) HealthCheck(ctx context.Context) error {
	if ws.client == nil {
		return fmt.Errorf("wechat client is nil")
	}
	return ws.client.HealthCheck(ctx)
}

// Close 关闭微信服务
func (ws *WeChatService) Close() error {
	if ws.client != nil {
		return ws.client.Close()
	}
	return nil
}

// 辅助函数

// getString 从map中获取字符串值
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// getDataMap 从map中获取数据map
func getDataMap(m map[string]interface{}, key string) map[string]interface{} {
	if val, ok := m[key]; ok {
		if dataMap, ok := val.(map[string]interface{}); ok {
			return dataMap
		}
	}
	return make(map[string]interface{})
}

// CreateDefaultCommands 创建默认命令
func CreateDefaultCommands() map[string]interface{} {
	return map[string]interface{}{
		"/search": map[string]interface{}{
			"function":    "SearchCommand",
			"description": "搜索电影/剧集",
			"category":    "搜索",
			"data":        map[string]interface{}{},
		},
		"/hot": map[string]interface{}{
			"function":    "HotCommand",
			"description": "热门推荐",
			"category":    "推荐",
			"data":        map[string]interface{}{},
		},
		"/download": map[string]interface{}{
			"function":    "DownloadCommand",
			"description": "下载管理",
			"category":    "下载",
			"data":        map[string]interface{}{},
		},
		"/cookiecloud": map[string]interface{}{
			"function":    "CookieCloudCommand",
			"description": "同步站点",
			"category":    "站点",
			"data":        map[string]interface{}{},
		},
		"/stats": map[string]interface{}{
			"function":    "StatsCommand",
			"description": "统计信息",
			"category":    "统计",
			"data":        map[string]interface{}{},
		},
		"/help": map[string]interface{}{
			"function":    "HelpCommand",
			"description": "帮助信息",
			"category":    "帮助",
			"data":        map[string]interface{}{},
		},
	}
}

// CommandProcessor 命令处理器接口
type CommandProcessor interface {
	ProcessCommand(ctx context.Context, command string, userID string) error
}

// SimpleCommandProcessor 简单命令处理器
type SimpleCommandProcessor struct {
	wechatService *WeChatService
	logger        *zap.Logger
}

// NewSimpleCommandProcessor 创建简单命令处理器
func NewSimpleCommandProcessor(wechatService *WeChatService) *SimpleCommandProcessor {
	return &SimpleCommandProcessor{
		wechatService: wechatService,
		logger:        logger.Logger,
	}
}

// ProcessCommand 处理命令
func (scp *SimpleCommandProcessor) ProcessCommand(ctx context.Context, command, userID string) error {
	scp.logger.Info("Processing WeChat command",
		zap.String("command", command),
		zap.String("user_id", userID))

	switch command {
	case "/search":
		return scp.sendSearchHelp(ctx, userID)
	case "/hot":
		return scp.sendHotRecommendations(ctx, userID)
	case "/help":
		return scp.sendHelp(ctx, userID)
	default:
		return scp.sendUnknownCommand(ctx, userID, command)
	}
}

// sendSearchHelp 发送搜索帮助
func (scp *SimpleCommandProcessor) sendSearchHelp(ctx context.Context, userID string) error {
	title := "🔍 搜索帮助"
	text := "请直接输入电影或剧集名称进行搜索\n\n例如：\n- 星际穿越\n- 复仇者联盟"
	
	_, err := scp.wechatService.SendMsg(ctx, title, text, "", userID, "")
	return err
}

// sendHotRecommendations 发送热门推荐
func (scp *SimpleCommandProcessor) sendHotRecommendations(ctx context.Context, userID string) error {
	title := "🔥 热门推荐"
	text := "当前热门内容：\n\n1. 最新电影推荐\n2. 热门剧集推荐\n3. 经典重温\n\n请访问网站查看完整列表。"
	
	_, err := scp.wechatService.SendMsg(ctx, title, text, "", userID, "")
	return err
}

// sendHelp 发送帮助信息
func (scp *SimpleCommandProcessor) sendHelp(ctx context.Context, userID string) error {
	title := "📖 使用帮助"
	text := "欢迎使用 MoviePilot 微信助手！\n\n可用命令：\n🔍 /search - 搜索功能\n🔥 /hot - 热门推荐\n📥 /download - 下载管理\n🌐 /cookiecloud - 站点同步\n📊 /stats - 统计信息\n❓ /help - 帮助信息\n\n更多功能请访问 Web 界面。"
	
	_, err := scp.wechatService.SendMsg(ctx, title, text, "", userID, "")
	return err
}

// sendUnknownCommand 发送未知命令消息
func (scp *SimpleCommandProcessor) sendUnknownCommand(ctx context.Context, userID, command string) error {
	title := "❓ 未知命令"
	text := fmt.Sprintf("未识别的命令：%s\n\n使用 /help 查看可用命令列表。", command)
	
	_, err := scp.wechatService.SendMsg(ctx, title, text, "", userID, "")
	return err
}

// WebhookHandler Webhook处理器
type WebhookHandler struct {
	wechatService     *WeChatService
	commandProcessor CommandProcessor
	logger           *zap.Logger
}

// NewWebhookHandler 创建Webhook处理器
func NewWebhookHandler(wechatService *WeChatService, processor CommandProcessor) *WebhookHandler {
	return &WebhookHandler{
		wechatService:     wechatService,
		commandProcessor: processor,
		logger:           logger.Logger,
	}
}

// HandleWebhook 处理微信Webhook
func (wh *WebhookHandler) HandleWebhook(ctx context.Context, data []byte) error {
	// 解析微信消息数据
	// 这里需要根据微信的消息格式进行解析
	// 示例实现
	
	wh.logger.Info("Received WeChat webhook",
		zap.ByteString("data", data))

	// TODO: 实现具体的消息解析和处理逻辑
	
	return nil
}

// StartWebhookServer 启动Webhook服务器
func (wh *WebhookHandler) StartWebhookServer(ctx context.Context, port int) error {
	// TODO: 实现HTTP服务器来接收微信webhook
	wh.logger.Info("Starting WeChat webhook server",
		zap.Int("port", port))
	
	select {
	case <-ctx.Done():
		wh.logger.Info("Webhook server stopped")
		return ctx.Err()
	}
}

// ScheduleMessage 定时发送消息
func (ws *WeChatService) ScheduleMessage(ctx context.Context, title, text, cronExpr string) error {
	// TODO: 实现定时消息发送功能
	ws.logger.Info("Scheduled message setup",
		zap.String("title", title),
		zap.String("cron", cronExpr))
	
	return nil
}