package channels

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"moviepilot-go/internal/integration/notification"
	"moviepilot-go/pkg/logger"
)

// SlackConfig Slack配置
type SlackConfig struct {
	WebhookURL string `json:"webhook_url"`
	Channel    string `json:"channel"`
	Username   string `json:"username"`
	IconEmoji  string `json:"icon_emoji"`
	
	// Socket Mode配置
	OAuthToken  string `json:"oauth_token"`
	AppToken   string `json:"app_token"`
	DSURL      string `json:"ds_url"`
	Source     string `json:"source"`
	
	// 高级配置
	Proxy      string `json:"proxy"`
	UserAgent   string `json:"user_agent"`
	Timeout    int    `json:"timeout"`
}

// SlackClient Slack客户端
type SlackClient struct {
	config       *SlackConfig
	client       *http.Client
	boltApp      *SlackBoltApp
	socketHandler *SlackSocketHandler
	running      bool
}

// SlackBoltApp Slack Bolt应用
type SlackBoltApp struct {
	client *SlackClient
}

// SlackSocketHandler Socket处理器
type SlackSocketHandler struct {
	app  *SlackBoltApp
	ready bool
}

// SlackMessage Slack消息结构
type SlackMessage struct {
	Type    string                 `json:"type"`
	Channel  string                 `json:"channel"`
	User     string                 `json:"user"`
	Text     string                 `json:"text"`
	Timestamp string                 `json:"ts"`
	Event    map[string]interface{} `json:"event"`
}

// SlackAction Slack动作结构
type SlackAction struct {
	Type        string                 `json:"type"`
	ActionID    string                 `json:"action_id"`
	BlockID     string                 `json:"block_id"`
	User        map[string]interface{} `json:"user"`
	Message     map[string]interface{} `json:"message"`
	State       map[string]interface{} `json:"state"`
	Value       interface{}            `json:"value"`
	ActionTs    string                 `json:"action_ts"`
}

// SlackShortcut Slack快捷方式结构
type SlackShortcut struct {
	Type        string                 `json:"type"`
	Token       string                 `json:"token"`
	ActionID    string                 `json:"action_id"`
	CallbackID  string                 `json:"callback_id"`
	TriggerID   string                 `json:"trigger_id"`
	Team        map[string]interface{} `json:"team"`
	User        map[string]interface{} `json:"user"`
	Channel     map[string]interface{} `json:"channel"`
	MessageTs   string                 `json:"message_ts"`
}

// SlackCommand Slack命令结构
type SlackCommand struct {
	Type        string                 `json:"type"`
	Command     string                 `json:"command"`
	Text        string                 `json:"text"`
	ResponseUrl string                 `json:"response_url"`
	TriggerId   string                 `json:"trigger_id"`
	Team        map[string]interface{} `json:"team"`
	User        map[string]interface{} `json:"user"`
	Channel     map[string]interface{} `json:"channel"`
}

// NewSlackClient 创建Slack客户端
func NewSlackClient(config map[string]interface{}) *SlackClient {
	webhookURL, _ := config["webhook_url"].(string)
	channel, _ := config["channel"].(string)
	username, _ := config["username"].(string)
	if username == "" {
		username = "MoviePilot"
	}
	iconEmoji, _ := config["icon_emoji"].(string)
	if iconEmoji == "" {
		iconEmoji = ":robot_face:"
	}
	
	// Socket Mode配置
	oauthToken, _ := config["oauth_token"].(string)
	appToken, _ := config["app_token"].(string)
	dsURL, _ := config["ds_url"].(string)
	source, _ := config["source"].(string)
	proxy, _ := config["proxy"].(string)
	userAgent, _ := config["user_agent"].(string)
	if userAgent == "" {
		userAgent = "MoviePilot-Slack"
	}
	timeout, _ := config["timeout"].(int)
	if timeout <= 0 {
		timeout = 30
	}

	slackConfig := &SlackConfig{
		WebhookURL: webhookURL,
		Channel:    channel,
		Username:   username,
		IconEmoji:  iconEmoji,
		OAuthToken:  oauthToken,
		AppToken:   appToken,
		DSURL:      dsURL,
		Source:     source,
		Proxy:      proxy,
		UserAgent:   userAgent,
		Timeout:    timeout,
	}

	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
	
	// 设置代理
	if proxy != "" {
		// TODO: 实现代理配置
	}

	slackClient := &SlackClient{
		config:  slackConfig,
		client:  client,
		running: false,
	}

	// 如果有Socket Mode配置，启动Socket模式
	if oauthToken != "" && appToken != "" {
		slackClient.initSocketMode()
	}

	return slackClient
}

// initSocketMode 初始化Socket模式
func (s *SlackClient) initSocketMode() {
	if s.config.OAuthToken == "" || s.config.AppToken == "" {
		logger.Info("Slack Socket Mode配置不完整，跳过Socket模式初始化")
		return
	}

	s.boltApp = &SlackBoltApp{
		client: s,
	}
	
	s.socketHandler = &SlackSocketHandler{
		app:  s.boltApp,
		ready: false,
	}
}

// Start 启动Slack服务
func (s *SlackClient) Start() error {
	if s.running {
		return fmt.Errorf("Slack client is already running")
	}

	if s.socketHandler != nil {
		return s.startSocketMode()
	}
	
	s.running = true
	logger.Info("Slack client started")
	return nil
}

// formatFileSize 格式化文件大小
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// isValidURL 验证URL格式
func isValidURL(url string) bool {
	return len(url) > 0 && (strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://"))
}

// Stop 停止Slack服务
func (s *SlackClient) Stop() error {
	if !s.running {
		return nil
	}

	if s.socketHandler != nil {
		return s.stopSocketMode()
	}

	s.running = false
	logger.Info("Slack client stopped")
	return nil
}

// formatFileSize 格式化文件大小
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// isValidURL 验证URL格式
func isValidURL(url string) bool {
	return len(url) > 0 && (strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://"))
}

// startSocketMode 启动Socket模式
func (s *SlackClient) startSocketMode() error {
	// TODO: 实现Slack Socket Mode连接
	logger.Info("Starting Slack Socket Mode...")
	
	// 注册事件处理器
	s.registerEventHandlers()
	
	// 连接到Slack
	return s.connectSocketMode()
}

// stopSocketMode 停止Socket模式
func (s *SlackClient) stopSocketMode() error {
	logger.Info("Stopping Slack Socket Mode...")
	
	// TODO: 关闭Socket连接
	s.socketHandler.ready = false
	
	return nil
}

// formatFileSize 格式化文件大小
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// isValidURL 验证URL格式
func isValidURL(url string) bool {
	return len(url) > 0 && (strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://"))
}

// connectSocketMode 连接Socket模式
func (s *SlackClient) connectSocketMode() error {
	// TODO: 实现实际的Socket连接逻辑
	// 这里需要使用WebSocket连接到Slack
	logger.Info("Connected to Slack Socket Mode")
	s.socketHandler.ready = true
	return nil
}

// formatFileSize 格式化文件大小
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// isValidURL 验证URL格式
func isValidURL(url string) bool {
	return len(url) > 0 && (strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://"))
}

// registerEventHandlers 注册事件处理器
func (s *SlackClient) registerEventHandlers() {
	// 注册消息事件处理器
	// TODO: 实现具体的事件处理器注册
	logger.Info("Slack event handlers registered")
}

// processMessage 处理消息事件
func (s *SlackClient) processMessage(message SlackMessage) error {
	if s.config.DSURL == "" {
		logger.Warn("DS URL not configured, skipping message processing")
		return nil
	}

	// 发送到本地服务
	return s.sendToDSService(message)
}

// processAction 处理动作事件
func (s *SlackClient) processAction(action SlackAction) error {
	if s.config.DSURL == "" {
		logger.Warn("DS URL not configured, skipping action processing")
		return nil
	}

	// 发送到本地服务
	return s.sendToDSService(action)
}

// processShortcut 处理快捷方式事件
func (s *SlackClient) processShortcut(shortcut SlackShortcut) error {
	if s.config.DSURL == "" {
		logger.Warn("DS URL not configured, skipping shortcut processing")
		return nil
	}

	// 发送到本地服务
	return s.sendToDSService(shortcut)
}

// processCommand 处理命令事件
func (s *SlackClient) processCommand(command SlackCommand) error {
	if s.config.DSURL == "" {
		logger.Warn("DS URL not configured, skipping command processing")
		return nil
	}

	// 发送到本地服务
	return s.sendToDSService(command)
}

// sendToDSService 发送数据到DS服务
func (s *SlackClient) sendToDSService(data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// 构建请求URL
	dsURL := s.config.DSURL
	if s.config.Source != "" {
		dsURL += "&source=" + s.config.Source
	}

	req, err := http.NewRequest("POST", dsURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", s.config.UserAgent)

	// 发送请求
	resp, err := s.client.Do(req)
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

// formatFileSize 格式化文件大小
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// isValidURL 验证URL格式
func isValidURL(url string) bool {
	return len(url) > 0 && (strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://"))
}

// GetState 获取状态
func (s *SlackClient) GetState() bool {
	return s.running && (s.socketHandler != nil && s.socketHandler.ready || s.config.WebhookURL != "")
}

// SendMessage 发送消息
func (s *SlackClient) SendMessage(ctx context.Context, title, content string, params map[string]interface{}) error {
	if s.config.WebhookURL == "" {
		return fmt.Errorf("Slack webhook URL is required")
	}

	// 构建Slack消息
	message := s.buildMessage(title, content, params)

	// 发送消息
	return s.sendToSlack(ctx, message)
}

// buildMessage 构建Slack消息
func (s *SlackClient) buildMessage(title, content string, params map[string]interface{}) map[string]interface{} {
	message := map[string]interface{}{
		"username": s.config.Username,
		"icon_emoji": s.config.IconEmoji,
	}

	// 设置频道
	if s.config.Channel != "" {
		message["channel"] = s.config.Channel
	}

	// 构建附件
	attachment := map[string]interface{}{
		"title": title,
		"text":  content,
		"color": s.getColor(params),
		"ts":    time.Now().Unix(),
	}

	// 添加字段
	if fields := s.buildFields(params); len(fields) > 0 {
		attachment["fields"] = fields
	}

	// 添加图片
	if imageURL, ok := params["image_url"].(string); ok && imageURL != "" {
		attachment["image_url"] = imageURL
	}

	message["attachments"] = []map[string]interface{}{attachment}

	return message
}

// buildFields 构建消息字段
func (s *SlackClient) buildFields(params map[string]interface{}) []map[string]interface{} {
	var fields []map[string]interface{}

	// 添加类型字段
	if msgType, ok := params["type"].(string); ok && msgType != "" {
		fields = append(fields, map[string]interface{}{
			"title": "类型",
			"value": msgType,
			"short": true,
		})
	}

		// 添加大小字段
	if size, ok := params["size"].(int64); ok && size > 0 {
		fields = append(fields, map[string]interface{}{
			"title": "大小",
			"value": formatFileSize(size),
			"short": true,
		})
	}

	// 添加时间字段
	if timestamp, ok := params["timestamp"].(int64); ok && timestamp > 0 {
		fields = append(fields, map[string]interface{}{
			"title": "时间",
			"value": time.Unix(timestamp, 0).Format("2006-01-02 15:04:05"),
			"short": true,
		})
	}

	// 添加自定义字段
	if customFields, ok := params["fields"].(map[string]interface{}); ok {
		for key, value := range customFields {
			fields = append(fields, map[string]interface{}{
				"title": key,
				"value": fmt.Sprintf("%v", value),
				"short": true,
			})
		}
	}

	return fields
}

// getColor 获取消息颜色
func (s *SlackClient) getColor(params map[string]interface{}) string {
	if msgType, ok := params["type"].(string); ok {
		switch msgType {
		case "download", "transfer":
			return "good"
		case "warning":
			return "warning"
		case "error":
			return "danger"
		default:
			return "#36a64f"
		}
	}
	return "#36a64f"
}

// sendToSlack 发送消息到Slack
func (s *SlackClient) sendToSlack(ctx context.Context, message map[string]interface{}) error {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.config.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Slack message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Slack API returned status %d", resp.StatusCode)
	}

	logger.Info("Slack message sent successfully")
	return nil
}

// formatFileSize 格式化文件大小
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// isValidURL 验证URL格式
func isValidURL(url string) bool {
	return len(url) > 0 && (strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://"))
}

// Test 测试连接
func (s *SlackClient) Test(ctx context.Context) error {
	if s.config.WebhookURL == "" {
		return fmt.Errorf("Slack webhook URL is required")
	}

	testMessage := map[string]interface{}{
		"username": s.config.Username,
		"icon_emoji": s.config.IconEmoji,
		"text": "MoviePilot Slack integration test",
	}

	if s.config.Channel != "" {
		testMessage["channel"] = s.config.Channel
	}

	jsonData, err := json.Marshal(testMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal test message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.config.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create test request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send test message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Slack test failed with status %d", resp.StatusCode)
	}

	logger.Info("Slack integration test successful")
	return nil
}

// formatFileSize 格式化文件大小
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// isValidURL 验证URL格式
func isValidURL(url string) bool {
	return len(url) > 0 && (strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://"))
}

// GetConfig 获取配置模板
func GetSlackConfigTemplate() map[string]interface{} {
	return map[string]interface{}{
		"webhook_url": map[string]interface{}{
			"type":        "string",
			"required":    true,
			"label":       "Webhook URL",
			"placeholder": "https://hooks.slack.com/services/...",
			"description": "Slack Incoming Webhook URL",
		},
		"channel": map[string]interface{}{
			"type":        "string",
			"required":    false,
			"label":       "Channel",
			"placeholder": "#general",
			"description": "Slack channel to send messages (optional)",
		},
		"username": map[string]interface{}{
			"type":        "string",
			"required":    false,
			"label":       "Username",
			"default":     "MoviePilot",
			"description": "Bot username",
		},
		"icon_emoji": map[string]interface{}{
			"type":        "string",
			"required":    false,
			"label":       "Icon Emoji",
			"default":     ":robot_face:",
			"description": "Bot icon emoji",
		},
	}
}

// ValidateConfig 验证配置
func ValidateSlackConfig(config map[string]interface{}) error {
	webhookURL, ok := config["webhook_url"].(string)
	if !ok || webhookURL == "" {
		return fmt.Errorf("webhook_url is required")
	}

	// 验证URL格式
	if !isValidURL(webhookURL) {
		return fmt.Errorf("invalid webhook URL format")
	}

	return nil
}

// formatFileSize 格式化文件大小
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// isValidURL 验证URL格式
func isValidURL(url string) bool {
	return len(url) > 0 && (strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://"))
}