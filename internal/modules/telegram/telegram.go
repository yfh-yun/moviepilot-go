package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"moviepilot-go/internal/config"
	"moviepilot-go/internal/utils"
)

// RetryException 重试异常
type RetryException struct {
	Message string
}

func (e *RetryException) Error() string {
	return e.Message
}

// Telegram Telegram模块
type Telegram struct {
	telegramToken  string
	telegramChatID string
	dsURL          string
	botUsername    *string
	client         *http.Client
	
	// 存储回调处理�?	callbackHandlers map[string]func(map[string]interface{})
	
	// userid -> chat_id 映射用于回复定位
	userChatMapping map[string]string
	mappingMutex    sync.RWMutex
	
	// 消息处理函数
	messageHandler   func(map[string]interface{})
	callbackHandler  func(map[string]interface{})
	
	// 停止信号
	stopChan chan struct{}
}

// NewTelegram 创建Telegram实例
func NewTelegram(telegramToken, telegramChatID string, options map[string]interface{}) *Telegram {
	if telegramToken == "" || telegramChatID == "" {
		utils.Log.Error("Telegram配置不完整！")
		return nil
	}
	
	// 创建Telegram实例
	tg := &Telegram{
		telegramToken:    telegramToken,
		telegramChatID:   telegramChatID,
		dsURL:            fmt.Sprintf("http://127.0.0.1:%d/api/v1/message?token=%s", config.Config.PORT, config.Config.API_TOKEN),
		callbackHandlers: make(map[string]func(map[string]interface{})),
		userChatMapping:  make(map[string]string),
		client:           &http.Client{Timeout: 30 * time.Second},
		stopChan:         make(chan struct{}),
	}
	
	// 标记渠道来源
	if name, ok := options["name"].(string); ok && name != "" {
		tg.dsURL = fmt.Sprintf("%s&source=%s", tg.dsURL, name)
	}
	
	// 设置代理
	if config.Config.PROXY != "" {
		proxyURL, err := url.Parse(config.Config.PROXY)
		if err == nil {
			tg.client.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			}
		}
	}
	
	// 获取并存储bot用户名用于@检�?	tg.getBotUsername()
	
	// 启动消息轮询
	go tg.startPolling()
	
	utils.Log.Info("Telegram消息接收服务启动")
	return tg
}

// getBotUsername 获取Bot用户�?func (t *Telegram) getBotUsername() {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/getMe", t.telegramToken)
	
	resp, err := t.client.Get(apiURL)
	if err != nil {
		utils.Log.Errorf("获取bot信息失败: %v", err)
		return
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.Log.Errorf("读取响应失败: %v", err)
		return
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		utils.Log.Errorf("解析响应失败: %v", err)
		return
	}
	
	if ok, okExists := result["ok"].(bool); !ok || !okExists {
		utils.Log.Error("获取bot信息失败")
		return
	}
	
	if result, resultExists := result["result"].(map[string]interface{}); resultExists {
		if username, usernameExists := result["username"].(string); usernameExists {
			t.botUsername = &username
			utils.Log.Infof("Telegram bot用户�? @%s", username)
			return
		}
	}
	
	utils.Log.Error("解析bot用户名失�?)
}

// updateuserChatMapping 更新用户与聊天的映射关系
func (t *Telegram) updateuserChatMapping(userid, chatID int64) {
	t.mappingMutex.Lock()
	defer t.mappingMutex.Unlock()
	
	if userid != 0 && chatID != 0 {
		t.userChatMapping[strconv.FormatInt(userid, 10)] = strconv.FormatInt(chatID, 10)
	}
}

// getuserChatID 获取用户对应的聊天ID
func (t *Telegram) getuserChatID(userid string) *string {
	t.mappingMutex.RLock()
	defer t.mappingMutex.RUnlock()
	
	if chatID, exists := t.userChatMapping[userid]; exists {
		return &chatID
	}
	return nil
}

// shouldProcessMessage 判断是否应该处理这条消息
func (t *Telegram) shouldProcessMessage(message map[string]interface{}) bool {
	chat, chatExists := message["chat"].(map[string]interface{})
	if !chatExists {
		utils.Log.Debug("无法获取聊天信息")
		return true
	}
	
	chatType, typeExists := chat["type"].(string)
	if !typeExists {
		utils.Log.Debug("无法获取聊天类型")
		return true
	}
	
	// 私聊消息总是处理
	if chatType == "private" {
		from, fromExists := message["from"].(map[string]interface{})
		if fromExists {
			userid, useridExists := from["id"].(float64)
			if useridExists {
				utils.Log.Debugf("处理私聊消息：用�?%.0f", userid)
			}
		}
		return true
	}
	
	// 群聊中的命令消息总是处理（以/开头）
	if text, textExists := message["text"].(string); textExists && strings.HasPrefix(text, "/") {
		utils.Log.Debugf("处理群聊命令消息�?s...", text[:min(20, len(text))])
		return true
	}
	
	// 群聊中检查是否@了机器人
	if chatType == "group" || chatType == "supergroup" {
		if t.botUsername == nil {
			// 如果没有获取到bot用户名，为了安全起见处理所有消�?			utils.Log.Debug("未获取到bot用户名，处理所有群聊消�?)
			return true
		}
		
		// 检查消息文本中是否包含@bot_username
		if text, textExists := message["text"].(string); textExists {
			mention := fmt.Sprintf("@%s", *t.botUsername)
			if strings.Contains(text, mention) {
				utils.Log.Debugf("检测到%s，处理群聊消�?, mention)
				return true
			}
		}
		
		// 检查消息实体中是否有提及bot
		if entities, entitiesExists := message["entities"].([]interface{}); entitiesExists {
			for _, entityItem := range entities {
				if entity, entityOk := entityItem.(map[string]interface{}); entityOk {
					entityType, typeExists := entity["type"].(string)
					if typeExists && entityType == "mention" {
						if text, textExists := message["text"].(string); textExists {
							offset, offsetExists := entity["offset"].(float64)
							length, lengthExists := entity["length"].(float64)
							if offsetExists && lengthExists {
								if int(offset)+int(length) <= len(text) {
									mentionText := text[int(offset):int(offset)+int(length)]
									mention := fmt.Sprintf("@%s", *t.botUsername)
									if mentionText == mention {
										utils.Log.Debugf("通过实体检测到%s，处理群聊消�?, mention)
										return true
									}
								}
							}
						}
					}
				}
			}
		}
		
		// 群聊中没有@机器人，不处�?		text := ""
		if textVal, textExists := message["text"].(string); textExists {
			text = textVal
		}
		utils.Log.Debugf("群聊消息未@机器人，跳过处理�?s...", text[:min(30, len(text))])
		return false
	}
	
	// 其他类型的聊天默认处�?	utils.Log.Debugf("处理其他类型聊天消息�?s", chatType)
	return true
}

// GetState 获取状�?func (t *Telegram) GetState() bool {
	return t.client != nil
}

// SendMsg 发送Telegram消息
func (t *Telegram) SendMsg(title, text, image, userid, link string, 
	buttons [][]map[string]string, 
	originalMessageID *int, originalChatID *string) *bool {
	
	if t.telegramToken == "" || t.telegramChatID == "" {
		return nil
	}
	
	if title == "" && text == "" {
		utils.Log.Warn("标题和内容不能同时为�?)
		result := false
		return &result
	}
	
	defer func() {
		if r := recover(); r != nil {
			utils.Log.Errorf("发送消息失败：%v", r)
		}
	}()
	
	var caption string
	if text != "" {
		// 对text进行Markdown特殊字符转义
		escapedText := strings.ReplaceAll(text, "_", "\\_")
		escapedText = strings.ReplaceAll(escapedText, "`", "\\`")
		caption = fmt.Sprintf("*%s*\n%s", title, escapedText)
	} else {
		caption = fmt.Sprintf("*%s*", title)
	}
	
	if link != "" {
		caption = fmt.Sprintf("%s\n[查看详情](%s)", caption, link)
	}
	
	// 确定目标chat_id
	chatID := t.determineTargetChatID(userid, originalChatID)
	
	// 创建按钮键盘
	var replyMarkup *map[string]interface{}
	if buttons != nil {
		replyMarkup = t.createInlineKeyboard(buttons)
	}
	
	// 判断是编辑消息还是发送新消息
	if originalMessageID != nil && originalChatID != nil {
		// 编辑消息
		return t.editMessage(*originalChatID, *originalMessageID, caption, buttons, image)
	} else {
		// 发送新消息
		return t.sendRequest(chatID, image, caption, replyMarkup)
	}
}

// determineTargetChatID 确定目标聊天ID
func (t *Telegram) determineTargetChatID(userid string, originalChatID *string) string {
	// 1. 优先使用原消息的聊天ID (编辑消息场景)
	if originalChatID != nil {
		return *originalChatID
	}
	
	// 2. 如果有userid，尝试从映射中获取用户的聊天ID
	if userid != "" {
		mappedChatID := t.getuserChatID(userid)
		if mappedChatID != nil {
			return *mappedChatID
		}
		// 如果映射中没有，回退到使用userid作为聊天ID (私聊场景)
		return userid
	}
	
	// 3. 最后使用默认聊天ID
	return t.telegramChatID
}

// createInlineKeyboard 创建内联键盘
func (t *Telegram) createInlineKeyboard(buttons [][]map[string]string) *map[string]interface{} {
	var keyboard [][]map[string]string
	for _, row := range buttons {
		var buttonRow []map[string]string
		for _, button := range row {
			buttonRow = append(buttonRow, button)
		}
		keyboard = append(keyboard, buttonRow)
	}
	
	result := map[string]interface{}{
		"inline_keyboard": keyboard,
	}
	
	return &result
}

// editMessage 编辑已发送的消息
func (t *Telegram) editMessage(chatID string, messageID int, text string, 
	buttons [][]map[string]string, image string) *bool {
	
	if t.client == nil {
		return nil
	}
	
	defer func() {
		if r := recover(); r != nil {
			utils.Log.Errorf("编辑消息失败�?v", r)
		}
	}()
	
	// 创建按钮键盘
	var replyMarkup *map[string]interface{}
	if buttons != nil {
		replyMarkup = t.createInlineKeyboard(buttons)
	}
	
	apiURL := ""
	var postData map[string]interface{}
	
	if image != "" {
		// 如果有图片，使用editMessageMedia
		apiURL = fmt.Sprintf("https://api.telegram.org/bot%s/editMessageMedia", t.telegramToken)
		media := map[string]interface{}{
			"type":    "photo",
			"media":   image,
			"caption": text,
			"parse_mode": "Markdown",
		}
		
		postData = map[string]interface{}{
			"chat_id":    chatID,
			"message_id": messageID,
			"media":      media,
		}
	} else {
		// 如果没有图片，使用editMessageText
		apiURL = fmt.Sprintf("https://api.telegram.org/bot%s/editMessageText", t.telegramToken)
		postData = map[string]interface{}{
			"chat_id":    chatID,
			"message_id": messageID,
			"text":       text,
			"parse_mode": "Markdown",
		}
	}
	
	// 添加reply_markup
	if replyMarkup != nil {
		postData["reply_markup"] = replyMarkup
	}
	
	// 发送请�?	jsonData, err := json.Marshal(postData)
	if err != nil {
		utils.Log.Errorf("序列化数据失败：%v", err)
		result := false
		return &result
	}
	
	resp, err := t.client.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		utils.Log.Errorf("发送请求失败：%v", err)
		result := false
		return &result
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.Log.Errorf("读取响应失败�?v", err)
		result := false
		return &result
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		utils.Log.Errorf("解析响应失败�?v", err)
		resultVal := false
		return &resultVal
	}
	
	if ok, okExists := result["ok"].(bool); okExists && ok {
		resultVal := true
		return &resultVal
	}
	
	utils.Log.Errorf("编辑消息失败�?s", string(body))
	resultVal := false
	return &resultVal
}

// sendRequest 向Telegram发送报�?func (t *Telegram) sendRequest(userid, image, caption string, replyMarkup *map[string]interface{}) *bool {
	if image != "" {
		res, err := utils.RequestUtils.GetRes(image, nil, nil, 30)
		if err != nil || res == nil {
			utils.Log.Error("获取图片失败")
			panic(&RetryException{Message: "获取图片失败"})
		}
		
		if res.Body != nil {
			// 使用随机标识构建图片文件的完整路径，并写入图片内容到文件
			tempDir := filepath.Join(config.Config.TEMP_PATH, "telegram")
			if _, err := os.Stat(tempDir); os.IsNotExist(err) {
				os.MkdirAll(tempDir, 0755)
			}
			
			imageFile := filepath.Join(tempDir, uuid.New().String())
			if err := os.WriteFile(imageFile, res.Body, 0644); err != nil {
				utils.Log.Errorf("写入图片文件失败�?v", err)
				panic(&RetryException{Message: "写入图片文件失败"})
			}
			
			// 发送图片到Telegram
			ret := t.sendPhoto(userid, imageFile, caption, replyMarkup)
			if !ret {
				panic(&RetryException{Message: "发送图片消息失�?})
			}
			return &ret
		}
	}
	
	// �?096分段循环发送消�?	ret := false
	if len(caption) > 4095 {
		for i := 0; i < len(caption); i += 4095 {
			end := i + 4095
			if end > len(caption) {
				end = len(caption)
			}
			
			text := caption[i:end]
			var markup *map[string]interface{}
			if i == 0 {
				markup = replyMarkup
			}
			
			ret = t.sendMessage(userid, text, markup)
		}
	} else {
		ret = t.sendMessage(userid, caption, replyMarkup)
	}
	
	if !ret {
		panic(&RetryException{Message: "发送文本消息失�?})
	}
	
	return &ret
}

// sendPhoto 发送图片消�?func (t *Telegram) sendPhoto(chatID, imageFile, caption string, replyMarkup *map[string]interface{}) bool {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendPhoto", t.telegramToken)
	
	// 创建multipart表单
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	
	// 添加chat_id字段
	writer.WriteField("chat_id", chatID)
	
	// 添加caption字段
	writer.WriteField("caption", caption)
	
	// 添加parse_mode字段
	writer.WriteField("parse_mode", "Markdown")
	
	// 添加photo文件
	file, err := os.Open(imageFile)
	if err != nil {
		utils.Log.Errorf("打开图片文件失败�?v", err)
		return false
	}
	defer file.Close()
	
	part, err := writer.CreateFormFile("photo", filepath.Base(imageFile))
	if err != nil {
		utils.Log.Errorf("创建表单文件失败�?v", err)
		return false
	}
	
	_, err = io.Copy(part, file)
	if err != nil {
		utils.Log.Errorf("复制文件内容失败�?v", err)
		return false
	}
	
	// 添加reply_markup
	if replyMarkup != nil {
		replyMarkupJSON, err := json.Marshal(replyMarkup)
		if err == nil {
			writer.WriteField("reply_markup", string(replyMarkupJSON))
		}
	}
	
	writer.Close()
	
	// 发送请�?	req, err := http.NewRequest("POST", apiURL, &buf)
	if err != nil {
		utils.Log.Errorf("创建请求失败�?v", err)
		return false
	}
	
	req.Header.Set("Content-Type", writer.FormDataContentType())
	
	resp, err := t.client.Do(req)
	if err != nil {
		utils.Log.Errorf("发送请求失败：%v", err)
		return false
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.Log.Errorf("读取响应失败�?v", err)
		return false
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		utils.Log.Errorf("解析响应失败�?v", err)
		return false
	}
	
	if ok, okExists := result["ok"].(bool); okExists && ok {
		return true
	}
	
	utils.Log.Errorf("发送图片消息失败：%s", string(body))
	return false
}

// sendMessage 发送文本消�?func (t *Telegram) sendMessage(chatID, text string, replyMarkup *map[string]interface{}) bool {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.telegramToken)
	
	postData := map[string]interface{}{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "Markdown",
	}
	
	// 添加reply_markup
	if replyMarkup != nil {
		postData["reply_markup"] = replyMarkup
	}
	
	// 发送请�?	jsonData, err := json.Marshal(postData)
	if err != nil {
		utils.Log.Errorf("序列化数据失败：%v", err)
		return false
	}
	
	resp, err := t.client.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		utils.Log.Errorf("发送请求失败：%v", err)
		return false
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.Log.Errorf("读取响应失败�?v", err)
		return false
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		utils.Log.Errorf("解析响应失败�?v", err)
		return false
	}
	
	if ok, okExists := result["ok"].(bool); okExists && ok {
		return true
	}
	
	utils.Log.Errorf("发送文本消息失败：%s", string(body))
	return false
}

// AnswerCallbackQuery 回应回调查询
func (t *Telegram) AnswerCallbackQuery(callbackQueryID, text string, showAlert bool) *bool {
	if t.client == nil {
		return nil
	}
	
	defer func() {
		if r := recover(); r != nil {
			utils.Log.Errorf("回应回调查询失败�?v", r)
		}
	}()
	
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/answerCallbackQuery", t.telegramToken)
	
	postData := map[string]interface{}{
		"callback_query_id": callbackQueryID,
	}
	
	if text != "" {
		postData["text"] = text
	}
	
	if showAlert {
		postData["show_alert"] = showAlert
	}
	
	// 发送请�?	jsonData, err := json.Marshal(postData)
	if err != nil {
		utils.Log.Errorf("序列化数据失败：%v", err)
		result := false
		return &result
	}
	
	resp, err := t.client.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		utils.Log.Errorf("发送请求失败：%v", err)
		result := false
		return &result
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.Log.Errorf("读取响应失败�?v", err)
		result := false
		return &result
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		utils.Log.Errorf("解析响应失败�?v", err)
		resultVal := false
		return &resultVal
	}
	
	if ok, okExists := result["ok"].(bool); okExists && ok {
		resultVal := true
		return &resultVal
	}
	
	utils.Log.Errorf("回应回调查询失败�?s", string(body))
	resultVal := false
	return &resultVal
}

// DeleteMsg 删除Telegram消息
func (t *Telegram) DeleteMsg(messageID int, chatID *int64) *bool {
	if t.telegramToken == "" || t.telegramChatID == "" {
		return nil
	}
	
	defer func() {
		if r := recover(); r != nil {
			utils.Log.Errorf("删除Telegram消息异常�?v", r)
		}
	}()
	
	// 确定要删除消息的聊天ID
	targetChatID := t.telegramChatID
	if chatID != nil {
		targetChatID = strconv.FormatInt(*chatID, 10)
	}
	
	// 删除消息
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/deleteMessage", t.telegramToken)
	
	postData := map[string]interface{}{
		"chat_id":    targetChatID,
		"message_id": messageID,
	}
	
	// 发送请�?	jsonData, err := json.Marshal(postData)
	if err != nil {
		utils.Log.Errorf("序列化数据失败：%v", err)
		result := false
		return &result
	}
	
	resp, err := t.client.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		utils.Log.Errorf("发送请求失败：%v", err)
		result := false
		return &result
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.Log.Errorf("读取响应失败�?v", err)
		result := false
		return &result
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		utils.Log.Errorf("解析响应失败�?v", err)
		resultVal := false
		return &resultVal
	}
	
	if ok, okExists := result["ok"].(bool); okExists && ok {
		utils.Log.Infof("成功删除Telegram消息: chat_id=%s, message_id=%d", targetChatID, messageID)
		resultVal := true
		return &resultVal
	}
	
	utils.Log.Errorf("删除Telegram消息失败: chat_id=%s, message_id=%d, response=%s", targetChatID, messageID, string(body))
	resultVal := false
	return &resultVal
}

// RegisterCommands 注册菜单命令
func (t *Telegram) RegisterCommands(commands map[string]map[string]string) {
	if t.client == nil {
		return
	}
	
	// 清理菜单命令
	t.DeleteCommands()
	
	// 设置bot命令
	if commands != nil && len(commands) > 0 {
		var botCommands []map[string]string
		for cmd, desc := range commands {
			if description, descExists := desc["description"]; descExists {
				botCommands = append(botCommands, map[string]string{
					"command":     strings.TrimPrefix(cmd, "/"),
					"description": description,
				})
			}
		}
		
		apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/setMyCommands", t.telegramToken)
		
		postData := map[string]interface{}{
			"commands": botCommands,
		}
		
		// 发送请�?		jsonData, err := json.Marshal(postData)
		if err != nil {
			utils.Log.Errorf("序列化数据失败：%v", err)
			return
		}
		
		resp, err := t.client.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			utils.Log.Errorf("发送请求失败：%v", err)
			return
		}
		defer resp.Body.Close()
		
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			utils.Log.Errorf("读取响应失败�?v", err)
			return
		}
		
		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			utils.Log.Errorf("解析响应失败�?v", err)
			return
		}
		
		if ok, okExists := result["ok"].(bool); !okExists || !ok {
			utils.Log.Errorf("设置菜单命令失败�?s", string(body))
		}
	}
}

// DeleteCommands 清理菜单命令
func (t *Telegram) DeleteCommands() {
	if t.client == nil {
		return
	}
	
	// 清理菜单命令
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/deleteMyCommands", t.telegramToken)
	
	resp, err := t.client.Post(apiURL, "application/json", nil)
	if err != nil {
		utils.Log.Errorf("发送请求失败：%v", err)
		return
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.Log.Errorf("读取响应失败�?v", err)
		return
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		utils.Log.Errorf("解析响应失败�?v", err)
		return
	}
	
	if ok, okExists := result["ok"].(bool); !okExists || !ok {
		utils.Log.Errorf("清理菜单命令失败�?s", string(body))
	}
}

// startPolling 启动消息轮询
func (t *Telegram) startPolling() {
	offset := 0
	
	for {
		select {
		case <-t.stopChan:
			utils.Log.Info("Telegram消息接收服务已停�?)
			return
		default:
			// 轮询消息
			t.pollMessages(&offset)
			
			// 等待一段时间再继续轮询
			time.Sleep(1 * time.Second)
		}
	}
}

// pollMessages 轮询消息
func (t *Telegram) pollMessages(offset *int) {
	defer func() {
		if r := recover(); r != nil {
			utils.Log.Errorf("轮询消息失败�?v", r)
		}
	}()
	
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates", t.telegramToken)
	
	params := url.Values{}
	params.Add("offset", strconv.Itoa(*offset))
	params.Add("timeout", "30")
	params.Add("allowed_updates", `["message","callback_query"]`)
	
	fullURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())
	
	resp, err := t.client.Get(fullURL)
	if err != nil {
		utils.Log.Errorf("获取更新失败�?v", err)
		return
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.Log.Errorf("读取响应失败�?v", err)
		return
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		utils.Log.Errorf("解析响应失败�?v", err)
		return
	}
	
	if ok, okExists := result["ok"].(bool); !okExists || !ok {
		utils.Log.Errorf("获取更新失败�?s", string(body))
		return
	}
	
	if updates, updatesExists := result["result"].([]interface{}); updatesExists {
		for _, updateItem := range updates {
			if update, updateOk := updateItem.(map[string]interface{}); updateOk {
				updateID, updateIDExists := update["update_id"].(float64)
				if updateIDExists {
					*offset = int(updateID) + 1
				}
				
				// 处理按钮回调
				if callbackQuery, callbackQueryExists := update["callback_query"].(map[string]interface{}); callbackQueryExists {
					t.handleCallbackQuery(callbackQuery)
					continue
				}
				
				// 处理普通消�?				if message, messageExists := update["message"].(map[string]interface{}); messageExists {
					t.handleMessage(message)
				}
			}
		}
	}
}

// handleCallbackQuery 处理按钮回调查询
func (t *Telegram) handleCallbackQuery(callbackQuery map[string]interface{}) {
	defer func() {
		if r := recover(); r != nil {
			utils.Log.Errorf("处理按钮回调失败�?v", r)
		}
	}()
	
	from, fromExists := callbackQuery["from"].(map[string]interface{})
	if !fromExists {
		utils.Log.Error("无法获取回调来源信息")
		return
	}
	
	userid, useridExists := from["id"].(float64)
	if !useridExists {
		utils.Log.Error("无法获取回调用户ID")
		return
	}
	
	message, messageExists := callbackQuery["message"].(map[string]interface{})
	if !messageExists {
		utils.Log.Error("无法获取回调消息信息")
		return
	}
	
	chat, chatExists := message["chat"].(map[string]interface{})
	if !chatExists {
		utils.Log.Error("无法获取回调聊天信息")
		return
	}
	
	chatID, chatIDExists := chat["id"].(float64)
	if !chatIDExists {
		utils.Log.Error("无法获取回调聊天ID")
		return
	}
	
	// 更新用户-chat映射
	t.updateuserChatMapping(int64(userid), int64(chatID))
	
	// 解析回调数据
	callbackData, dataExists := callbackQuery["data"].(string)
	if !dataExists {
		utils.Log.Error("无法获取回调数据")
		return
	}
	
	utils.Log.Infof("收到来自的Telegram按钮回调：userid=%.0f, callback_data=%s", userid, callbackData)
	
	// 先确认回调，避免用户看到loading状�?	callbackQueryID, idExists := callbackQuery["id"].(string)
	if idExists {
		t.AnswerCallbackQuery(callbackQueryID, "", false)
	}
	
	// 发送回调数据给主程序处�?	callbackJSON := map[string]interface{}{
		"callback_query": callbackQuery,
	}
	
	// 发送给主程序处�?	jsonData, err := json.Marshal(callbackJSON)
	if err != nil {
		utils.Log.Errorf("序列化回调数据失败：%v", err)
		return
	}
	
	// 发送POST请求
	resp, err := utils.RequestUtils.PostRes(t.dsURL, jsonData, map[string]string{
		"Content-Type": "application/json",
	}, nil, 15)
	
	if err != nil || resp == nil {
		utils.Log.Errorf("发送回调数据失败：%v", err)
		return
	}
}

// handleMessage 处理普通消�?func (t *Telegram) handleMessage(message map[string]interface{}) {
	defer func() {
		if r := recover(); r != nil {
			utils.Log.Errorf("处理消息失败�?v", r)
		}
	}()
	
	from, fromExists := message["from"].(map[string]interface{})
	if !fromExists {
		utils.Log.Debug("无法获取消息来源信息")
		return
	}
	
	userid, useridExists := from["id"].(float64)
	if !useridExists {
		utils.Log.Debug("无法获取用户ID")
		return
	}
	
	chat, chatExists := message["chat"].(map[string]interface{})
	if !chatExists {
		utils.Log.Debug("无法获取聊天信息")
		return
	}
	
	chatID, chatIDExists := chat["id"].(float64)
	if !chatIDExists {
		utils.Log.Debug("无法获取聊天ID")
		return
	}
	
	// 更新用户-chat映射
	t.updateuserChatMapping(int64(userid), int64(chatID))
	
	// 检查是否应该处理这条消�?	if !t.shouldProcessMessage(message) {
		return
	}
	
	text, textExists := message["text"].(string)
	username := ""
	if usernameVal, usernameExists := from["username"].(string); usernameExists {
		username = usernameVal
	}
	
	if textExists && userid != 0 {
		utils.Log.Infof("收到来自的Telegram消息：userid=%.0f, username=%s, chat_id=%.0f, text=%s", 
			userid, username, chatID, text)
		
		// 发送给主程序处�?		jsonData, err := json.Marshal(message)
		if err != nil {
			utils.Log.Errorf("序列化消息数据失败：%v", err)
			return
		}
		
		// 发送POST请求
		resp, err := utils.RequestUtils.PostRes(t.dsURL, jsonData, map[string]string{
			"Content-Type": "application/json",
		}, nil, 15)
		
		if err != nil || resp == nil {
			utils.Log.Errorf("发送消息数据失败：%v", err)
			return
		}
	}
}

// Stop 停止Telegram消息接收服务
func (t *Telegram) Stop() {
	close(t.stopChan)
	utils.Log.Info("Telegram消息接收服务停止信号已发�?)
}

// min 返回两个整数中的较小�?func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
