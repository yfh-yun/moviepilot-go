package channels

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"moviepilot-go/internal/integration/notification"
	"moviepilot-go/pkg/logger"
	"go.uber.org/zap"
)

// TelegramConfig Telegram通知配置
type TelegramConfig struct {
	BotToken    string `json:"bot_token"`     // Bot Token
	ChatID      string `json:"chat_id"`       // 聊天ID或频道ID
	ParseMode    string `json:"parse_mode"`    // 解析模式，支持Markdown、HTML
	APIURL      string `json:"api_url"`       // Bot API地址
	DSURL        string `json:"ds_url"`        // 回调地址
	Source       string `json:"source"`       // 渠道标识
	Proxy        string `json:"proxy"`         // 代理设置
	UserAgent    string `json:"user_agent"`    // User-Agent
	Timeout      int    `json:"timeout"`       // 超时时间
}

// TelegramSender Telegram通知发送器
type TelegramSender struct {
	config         *TelegramConfig
	httpClient     *http.Client
	bot            *TelegramBot
	running        bool
	eventHandlers  map[string]func(interface{}) error
	userChatMapping map[string]string // user_id -> chat_id
	botUsername    string
}

// TelegramBot Telegram Bot结构
type TelegramBot struct {
	token     string
	apiURL    string
	userAgent  string
}

// NewTelegramSender 创建新的Telegram通知发送器
func NewTelegramSender(config *TelegramConfig) *TelegramSender {
	// 设置默认值
	if config.ParseMode == "" {
		config.ParseMode = "Markdown"
	}
	if config.UserAgent == "" {
		config.UserAgent = "MoviePilot-Telegram"
	}
	if config.Timeout <= 0 {
		config.Timeout = 30
	}

	// 设置API URL
	apiURL := config.APIURL
	if apiURL == "" {
		apiURL = "https://api.telegram.org"
	}

	sender := &TelegramSender{
		config:         config,
		httpClient: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
		bot: &TelegramBot{
			token:    config.BotToken,
			apiURL:   apiURL,
			userAgent: config.UserAgent,
		},
		running:        false,
		eventHandlers:  make(map[string]func(interface{}) error),
		userChatMapping: make(map[string]string),
	}

	// 如果有回调URL配置，初始化Bot
	if config.DSURL != "" && config.BotToken != "" {
		sender.initBot()
	}

	return sender
}

// Name 返回发送器名称
func (t *TelegramSender) Name() string {
	return "telegram"
}

// initBot 初始化Bot
func (t *TelegramSender) initBot() {
	if t.config.BotToken == "" {
		logger.Error("Telegram bot token is required")
		return
	}

	// 注册命令处理器
	t.registerCommandHandlers()
	
	// 注册回调查询处理器
	t.registerCallbackHandlers()
	
	// 获取Bot信息
	t.getBotInfo()
}

// registerCommandHandlers 注册命令处理器
func (t *TelegramSender) registerCommandHandlers() {
	// 注册 /start 和 /help 命令
	t.eventHandlers["command:start"] = t.handleStartCommand
	t.eventHandlers["command:help"] = t.handleHelpCommand
}

// registerCallbackHandlers 注册回调查询处理器
func (t *TelegramSender) registerCallbackHandlers() {
	// 注册通用回调处理器
	t.eventHandlers["callback:general"] = t.handleCallbackQuery
}

// getBotInfo 获取Bot信息
func (t *TelegramSender) getBotInfo() {
	if t.bot == nil || t.bot.token == "" {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", t.bot.apiURL+"/getMe", nil)
	if err != nil {
		logger.Error("Failed to create getMe request", zap.Error(err))
		return
	}

	req.Header.Set("User-Agent", t.bot.userAgent)

	resp, err := t.httpClient.Do(req)
	if err != nil {
		logger.Error("Failed to get bot info", zap.Error(err))
		return
	}
	defer resp.Body.Close()

	var botInfo struct {
		OK     bool   `json:"ok"`
		Result struct {
			ID       int64  `json:"id"`
			Username string `json:"username"`
		} `json:"result"`
		Description string `json:"description"`
		ErrorCode int    `json:"error_code"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&botInfo); err != nil {
		logger.Error("Failed to decode bot info", zap.Error(err))
		return
	}

	if !botInfo.OK {
		logger.Error("Bot is not valid", zap.String("error", botInfo.Description))
		return
	}

	t.botUsername = botInfo.Result.Username
	logger.Info("Telegram bot initialized", 
		zap.String("username", "@"+t.botUsername))
}

// handleStartCommand 处理/start命令
func (t *TelegramSender) handleStartCommand(data interface{}) error {
	// TODO: 实现start命令处理逻辑
	logger.Info("Received /start command")
	return t.sendWelcomeMessage(data)
}

// handleHelpCommand 处理/help命令
func (t *TelegramSender) handleHelpCommand(data interface{}) error {
	// TODO: 实现help命令处理逻辑
	logger.Info("Received /help command")
	return t.sendHelpMessage(data)
}

// handleCallbackQuery 处理回调查询
func (t *TelegramSender) handleCallbackQuery(data interface{}) error {
	// 解析回调数据
	callbackData, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid callback data format")
	}

	queryID, _ := callbackData["id"].(string)
	userInfo, _ := callbackData["from"].(map[string]interface{})
	userID := fmt.Sprintf("%.0f", userInfo["id"])

	logger.Info("Received callback query", 
		zap.String("callback_data", queryID),
		zap.String("user_id", userID))

	// 发送回调数据到主程序处理
	return t.sendToDSService(callbackData)
}

// sendWelcomeMessage 发送欢迎消息
func (t *TelegramSender) sendWelcomeMessage(data interface{}) error {
	// TODO: 实现发送欢迎消息
	return nil
}

// sendHelpMessage 发送帮助消息
func (t *TelegramSender) sendHelpMessage(data interface{}) error {
	// TODO: 实现发送帮助消息
	return nil
}

// updateUserChatMapping 更新用户-聊天映射
func (t *TelegramSender) updateUserChatMapping(userID, chatID string) {
	t.userChatMapping[userID] = chatID
	logger.Debug("Updated user-chat mapping", 
		zap.String("user_id", userID),
		zap.String("chat_id", chatID))
}

// shouldProcessMessage 判断是否应该处理消息
func (t *TelegramSender) shouldProcessMessage(message interface{}) bool {
	// TODO: 实现消息过滤逻辑
	return true
}

// SupportedLevels 返回支持的通知级别
func (t *TelegramSender) SupportedLevels() []notification.NotificationLevel {
	return []notification.NotificationLevel{
		notification.LevelInfo,
		notification.LevelWarning,
		notification.LevelError,
		notification.LevelSuccess,
	}
}

// Validate 验证消息
func (t *TelegramSender) Validate(message *notification.Message) error {
	if message.Title == "" && message.Content == "" {
		return fmt.Errorf("message title and content cannot both be empty")
	}

	// Telegram消息长度限制
	if len(message.Content) > 4096 {
		return fmt.Errorf("message content too long (max 4096 characters)")
	}

	return nil
}

// Send 发送Telegram通知
func (t *TelegramSender) Send(ctx context.Context, message *notification.Message) error {
	if t.config.BotToken == "" || t.config.ChatID == "" {
		return fmt.Errorf("telegram bot token or chat ID is not configured")
	}

	// 构建消息文本
	text := t.buildMessageText(message)

	// 构建API URL
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.config.BotToken)

	// 构建请求参数
	params := url.Values{}
	params.Set("chat_id", t.config.ChatID)
	params.Set("text", text)

	if t.config.ParseMode != "" {
		params.Set("parse_mode", t.config.ParseMode)
	}

	// 如果有链接，禁用预览
	if message.LinkURL != "" {
		params.Set("disable_web_page_preview", "true")
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	var apiResp struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
		ErrorCode   int    `json:"error_code"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !apiResp.OK {
		return fmt.Errorf("telegram API error: %s (code: %d)", apiResp.Description, apiResp.ErrorCode)
	}

	return nil
}

// buildMessageText 构建消息文本
func (t *TelegramSender) buildMessageText(message *notification.Message) string {
	var text strings.Builder

	// 添加级别标识和标题
	switch message.Level {
	case notification.LevelError:
		text.WriteString("🚨 *")
		text.WriteString(message.Title)
		text.WriteString("*\n\n")
	case notification.LevelWarning:
		text.WriteString("⚠️ *")
		text.WriteString(message.Title)
		text.WriteString("*\n\n")
	case notification.LevelSuccess:
		text.WriteString("✅ *")
		text.WriteString(message.Title)
		text.WriteString("*\n\n")
	case notification.LevelInfo:
		text.WriteString("ℹ️ *")
		text.WriteString(message.Title)
		text.WriteString("*\n\n")
	default:
		if message.Title != "" {
			text.WriteString("*")
			text.WriteString(message.Title)
			text.WriteString("*\n\n")
		}
	}

	// 添加内容
	text.WriteString(message.Content)

	// 添加链接（如果有）
	if message.LinkURL != "" {
		text.WriteString("\n\n[查看详情](")
		text.WriteString(message.LinkURL)
		text.WriteString(")")
	}

	// 添加时间戳
	text.WriteString("\n\n_")
	text.WriteString(time.Now().Format("2006-01-02 15:04:05"))
	text.WriteString("_")

	return text.String()
}

// HealthCheck 健康检查
func (t *TelegramSender) HealthCheck(ctx context.Context) error {
	// 测试获取Bot信息
	if t.config.BotToken == "" {
		return fmt.Errorf("telegram bot token is not configured")
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/getMe", t.config.BotToken)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get bot info: %w", err)
	}
	defer resp.Body.Close()

	var apiResp struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !apiResp.OK {
		return fmt.Errorf("telegram bot is not valid: %s", apiResp.Description)
	}

	return nil
}

// Close 关闭发送器
func (t *TelegramSender) Close() error {
	return nil
}

// sendToDSService 发送数据到DS服务
func (t *TelegramSender) sendToDSService(data interface{}) error {
	if t.config.DSURL == "" {
		logger.Warn("DS URL not configured, skipping DS service call")
		return nil
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// 构建请求URL
	dsURL := t.config.DSURL
	if t.config.Source != "" {
		dsURL += "&source=" + t.config.Source
	}

	req, err := http.NewRequest("POST", dsURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", t.config.UserAgent)

	// 发送请求
	resp, err := t.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send to DS service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("DS service returned status %d", resp.StatusCode)
	}

	logger.Debug("Data sent to DS service successfully")
	return nil
}

// Start 启动Telegram Bot
func (t *TelegramSender) Start() error {
	if t.running {
		return fmt.Errorf("Telegram sender is already running")
	}

	if t.config.BotToken == "" {
		return fmt.Errorf("bot token is required")
	}

	if t.bot != nil {
		// 启动轮询服务
		go t.startPolling()
	}

	t.running = true
	logger.Info("Telegram sender started")
	return nil
}

// Stop 停止Telegram Bot
func (t *TelegramSender) Stop() error {
	if !t.running {
		return nil
	}

	t.running = false
	logger.Info("Telegram sender stopped")
	return nil
}

// startPolling 开始轮询消息
func (t *TelegramSender) startPolling() {
	for t.running {
		// TODO: 实现消息轮询逻辑
		time.Sleep(1 * time.Second)
	}
}

// GetState 获取状态
func (t *TelegramSender) GetState() bool {
	return t.running && t.config.BotToken != ""
}

// CreateTelegramChannel 创建Telegram通知渠道
func CreateTelegramChannel(config *TelegramConfig) *notification.Channel {
	enabled := config.BotToken != "" && config.ChatID != ""
	
	return &notification.Channel{
		Name:        "telegram",
		Description: "Telegram通知渠道",
		Enabled:     enabled,
		Sender:      NewTelegramSender(config),
		Config: map[string]string{
			"bot_token":  config.BotToken,
			"chat_id":    config.ChatID,
			"parse_mode": config.ParseMode,
			"api_url":    config.APIURL,
			"ds_url":     config.DSURL,
			"source":     config.Source,
			"user_agent": config.UserAgent,
		},
	}
}
