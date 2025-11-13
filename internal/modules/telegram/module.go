package telegram

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	
	"moviepilot-go/internal/config"
	"moviepilot-go/internal/modules"
	"moviepilot-go/internal/utils"
	"moviepilot-go/pkg/models"
)

// TelegramModule Telegram模块
type TelegramModule struct {
	modules.ModuleBase
	modules.MessageBase[*Telegram]
	channel models.MessageChannel
}

// NewTelegramModule 创建Telegram模块实例
func NewTelegramModule() *TelegramModule {
	tm := &TelegramModule{
		channel: models.MessageChannelTelegram,
	}
	
	// 初始化模块服�?	tm.InitService("telegram", &Telegram{})
	
	return tm
}

// GetName 获取模块名称
func (tm *TelegramModule) GetName() string {
	return "Telegram"
}

// GetType 获取模块类型
func (tm *TelegramModule) GetType() models.ModuleType {
	return models.ModuleTypeNotification
}

// GetSubtype 获取模块子类�?func (tm *TelegramModule) GetSubtype() models.MessageChannel {
	return models.MessageChannelTelegram
}

// GetPriority 获取模块优先级，数字越小优先级越高，只有同一接口下优先级才生�?func (tm *TelegramModule) GetPriority() int {
	return 0
}

// Stop 停止模块
func (tm *TelegramModule) Stop() {
	instances := tm.GetInstances()
	for _, client := range instances {
		if client != nil {
			client.Stop()
		}
	}
}

// Test 测试模块连接�?func (tm *TelegramModule) Test() (*bool, string) {
	instances := tm.GetInstances()
	if len(instances) == 0 {
		return nil, ""
	}
	
	for name, client := range instances {
		state := client.GetState()
		if !state {
			return utils.BoolPtr(false), fmt.Sprintf("Telegram %s 未就�?, name)
		}
	}
	
	return utils.BoolPtr(true), ""
}

// InitSetting 初始化设�?func (tm *TelegramModule) InitSetting() (string, interface{}) {
	// 根据实际需要实�?	return "", nil
}

// MessageParser 解析消息内容，返回字�?func (tm *TelegramModule) MessageParser(source string, body, form, args interface{}) *models.CommingMessage {
	// 获取服务配置
	clientConfig := tm.GetConfig(source)
	if clientConfig == nil {
		return nil
	}
	
	client := tm.GetInstance(clientConfig.Name)
	
	// 解析消息�?	bodyStr, ok := body.(string)
	if !ok {
		utils.Log.Debug("无法解析Telegram消息�?)
		return nil
	}
	
	var message map[string]interface{}
	if err := json.Unmarshal([]byte(bodyStr), &message); err != nil {
		utils.Log.Debugf("解析Telegram消息失败�?v", err)
		return nil
	}
	
	if message != nil {
		// 处理按钮回调
		if callbackQuery, exists := message["callback_query"]; exists {
			return tm.handleCallbackQuery(message, clientConfig)
		}
		
		// 处理普通消�?		return tm.handleTextMessage(message, clientConfig, client)
	}
	
	return nil
}

// handleCallbackQuery 处理按钮回调查询
func (tm *TelegramModule) handleCallbackQuery(message map[string]interface{}, 
	clientConfig *models.NotificationConf) *models.CommingMessage {
	
	callbackQuery, ok := message["callback_query"].(map[string]interface{})
	if !ok {
		return nil
	}
	
	userInfo, ok := callbackQuery["from"].(map[string]interface{})
	if !ok {
		return nil
	}
	
	callbackData, ok := callbackQuery["data"].(string)
	if !ok {
		return nil
	}
	
	userID := int64(0)
	if userIDVal, exists := userInfo["id"].(float64); exists {
		userID = int64(userIDVal)
	}
	
	username := ""
	if usernameVal, exists := userInfo["username"].(string); exists {
		username = usernameVal
	}
	
	if callbackData != "" && userID != 0 {
		utils.Log.Infof("收到来自 %s 的Telegram按钮回调：userid=%d, username=%s, callback_data=%s",
			clientConfig.Name, userID, username, callbackData)
		
		// 将callback_data作为特殊格式的text返回，以便主程序识别这是按钮回调
		callbackText := fmt.Sprintf("CALLBACK:%s", callbackData)
		
		var messageID *int
		var chatID *string
		
		if msg, msgExists := callbackQuery["message"].(map[string]interface{}); msgExists {
			if msgID, msgIDExists := msg["message_id"].(float64); msgIDExists {
				msgIDInt := int(msgID)
				messageID = &msgIDInt
			}
			
			if chat, chatExists := msg["chat"].(map[string]interface{}); chatExists {
				if chatIDVal, chatIDExists := chat["id"].(float64); chatIDExists {
					chatIDStr := fmt.Sprintf("%.0f", chatIDVal)
					chatID = &chatIDStr
				}
			}
		}
		
		// 创建包含完整回调信息的CommingMessage
		return &models.CommingMessage{
			Channel:       models.MessageChannelTelegram,
			Source:        clientConfig.Name,
			UserID:        userID,
			Username:      username,
			Text:          callbackText,
			IsCallback:    true,
			CallbackData:  callbackData,
			MessageID:     messageID,
			ChatID:        chatID,
			CallbackQuery: callbackQuery,
		}
	}
	
	return nil
}

// handleTextMessage 处理普通文本消�?func (tm *TelegramModule) handleTextMessage(msg map[string]interface{},
	clientConfig *models.NotificationConf, client *Telegram) *models.CommingMessage {
	
	text, textExists := msg["text"].(string)
	if !textExists {
		return nil
	}
	
	from, fromExists := msg["from"].(map[string]interface{})
	if !fromExists {
		return nil
	}
	
	userID := int64(0)
	if userIDVal, userIDExists := from["id"].(float64); userIDExists {
		userID = int64(userIDVal)
	}
	
	username := ""
	if usernameVal, usernameExists := from["username"].(string); usernameExists {
		username = usernameVal
	}
	
	// Extract chat_id to enable correct reply targeting
	var chatID *float64
	if chat, chatExists := msg["chat"].(map[string]interface{}); chatExists {
		if chatIDVal, chatIDExists := chat["id"].(float64); chatIDExists {
			chatID = &chatIDVal
		}
	}
	
	var chatIDStr *string
	if chatID != nil {
		chatIDStrVal := fmt.Sprintf("%.0f", *chatID)
		chatIDStr = &chatIDStrVal
	}
	
	if text != "" && userID != 0 {
		utils.Log.Infof("收到来自 %s 的Telegram消息：userid=%d, username=%s, chat_id=%v, text=%s",
			clientConfig.Name, userID, username, chatID, text)
		
		// Clean bot mentions from text to ensure consistent processing
		var botUsername *string
		if client != nil {
			botUsername = client.botUsername
		}
		cleanedText := tm.cleanBotMention(text, botUsername)
		
		// 检查权�?		adminUsers := ""
		if adminUsersVal, exists := clientConfig.Config["TELEGRAM_ADMINS"]; exists {
			adminUsers = adminUsersVal.(string)
		}
		
		userList := ""
		if userlistVal, exists := clientConfig.Config["TELEGRAM_USERS"]; exists {
			userList = userlistVal.(string)
		}
		
		configChatID := ""
		if chatIDVal, exists := clientConfig.Config["TELEGRAM_CHAT_ID"]; exists {
			configChatID = chatIDVal.(string)
		}
		
		userIDStr := fmt.Sprintf("%d", userID)
		if strings.HasPrefix(cleanedText, "/") {
			if adminUsers != "" &&
				!strings.Contains(adminUsers, userIDStr) &&
				userIDStr != configChatID {
				
				if client != nil {
					client.SendMsg("只有管理员才有权限执行此命令", "", "", userIDStr, "", nil, nil, nil)
				}
				return nil
			}
		} else {
			if userList != "" &&
				!strings.Contains(userList, userIDStr) {
				utils.Log.Infof("用户%d不在用户白名单中，无法使用此机器�?, userID)
				if client != nil {
					client.SendMsg("你不在用户白名单中，无法使用此机器人", "", "", userIDStr, "", nil, nil, nil)
				}
				return nil
			}
		}
		
		return &models.CommingMessage{
			Channel:  models.MessageChannelTelegram,
			Source:   clientConfig.Name,
			UserID:   userID,
			Username: username,
			Text:     cleanedText, // Use cleaned text
			ChatID:   chatIDStr,
		}
	}
	
	return nil
}

// cleanBotMention 清理消息中的@bot部分，确保文本处理一致�?func (tm *TelegramModule) cleanBotMention(text string, botUsername *string) string {
	if text == "" || botUsername == nil {
		return text
	}
	
	// Remove @bot_username from the beginning and any position in text
	cleaned := text
	mentionPattern := fmt.Sprintf("@%s", *botUsername)
	
	// Remove mention at the beginning with optional following space
	if strings.HasPrefix(cleaned, mentionPattern) {
		cleaned = strings.TrimPrefix(cleaned, mentionPattern)
		cleaned = strings.TrimSpace(cleaned)
	}
	
	// Remove mention at any other position
	cleaned = strings.ReplaceAll(cleaned, mentionPattern, "")
	cleaned = strings.TrimSpace(cleaned)
	
	// Clean up multiple spaces
	re := regexp.MustCompile(`\s+`)
	cleaned = re.ReplaceAllString(cleaned, " ")
	cleaned = strings.TrimSpace(cleaned)
	
	return cleaned
}

// PostMessage 发送消�?func (tm *TelegramModule) PostMessage(message *models.Notification) {
	configs := tm.GetConfigs()
	for _, conf := range configs {
		if !tm.CheckMessage(message, conf.Name) {
			continue
		}
		
		userid := message.UserID
		if userid == "" && message.Targets != nil {
			if uid, exists := (*message.Targets)["telegram_userid"]; exists {
				userid = uid.(string)
			}
			if userid == "" {
				utils.Log.Warn("用户没有指定 Telegram用户ID，消息无法发�?)
				return
			}
		}
		
		client := tm.GetInstance(conf.Name)
		if client != nil {
			client.SendMsg(
				message.Title,
				message.Text,
				message.Image,
				userid,
				message.Link,
				message.Buttons,
				message.OriginalMessageID,
				message.OriginalChatID,
			)
		}
	}
}

// PostMediasMessage 发送媒体信息选择列表
func (tm *TelegramModule) PostMediasMessage(message *models.Notification, medias []*models.MediaInfo) {
	configs := tm.GetConfigs()
	for _, conf := range configs {
		if !tm.CheckMessage(message, conf.Name) {
			continue
		}
		
		client := tm.GetInstance(conf.Name)
		if client != nil {
			// TODO: 实现发送媒体列表消息的逻辑
			// 这里需要根据具体需求实�?			utils.Log.Debug("发送媒体列表消息功能待实现")
		}
	}
}

// PostTorrentsMessage 发送种子信息选择列表
func (tm *TelegramModule) PostTorrentsMessage(message *models.Notification, torrents []*models.Context) {
	configs := tm.GetConfigs()
	for _, conf := range configs {
		if !tm.CheckMessage(message, conf.Name) {
			continue
		}
		
		client := tm.GetInstance(conf.Name)
		if client != nil {
			// TODO: 实现发送种子列表消息的逻辑
			// 这里需要根据具体需求实�?			utils.Log.Debug("发送种子列表消息功能待实现")
		}
	}
}

// DeleteMessage 删除消息
func (tm *TelegramModule) DeleteMessage(channel models.MessageChannel, source string,
	messageID int, chatID *int64) bool {
	
	success := false
	configs := tm.GetConfigs()
	for _, conf := range configs {
		if channel != tm.channel {
			break
		}
		if source != conf.Name {
			continue
		}
		client := tm.GetInstance(conf.Name)
		if client != nil {
			result := client.DeleteMsg(messageID, chatID)
			if result != nil && *result {
				success = true
			}
		}
	}
	return success
}

// RegisterCommands 注册命令
func (tm *TelegramModule) RegisterCommands(commands map[string]map[string]string) {
	configs := tm.GetConfigs()
	for _, clientConfig := range configs {
		client := tm.GetInstance(clientConfig.Name)
		if client == nil {
			continue
		}
		
		// 过滤命令
		filteredCommands := make(map[string]map[string]string)
		for cmd, desc := range commands {
			filteredCommands[cmd] = desc
		}
		
		// 如果 filteredCommands 为空，则跳过注册
		if len(filteredCommands) == 0 {
			utils.Log.Debug("Filtered commands are empty, skipping registration.")
			client.DeleteCommands()
			continue
		}
		
		// 注册命令
		client.RegisterCommands(filteredCommands)
	}
}
