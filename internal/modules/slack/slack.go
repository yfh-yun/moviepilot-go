package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"moviepilot-go/internal/config"
	"moviepilot-go/internal/utils"
)

// Slack Slack模块
type Slack struct {
	oauthToken  string
	appToken    string
	channel     string
	client      *http.Client
	dsURL       string
	service     *SocketModeHandler
	mutex       sync.Mutex
}

// SocketModeHandler Socket模式处理�?type SocketModeHandler struct {
	// 这里可以添加实际的Socket模式处理逻辑
	// 由于Go的slack库与Python的slack-bolt库有所不同，这里做了简化处�?	running bool
}

// NewSlack 创建Slack实例
func NewSlack(oauthToken, appToken, channel string, options map[string]interface{}) *Slack {
	if oauthToken == "" || appToken == "" {
		utils.Log.Error("Slack 配置不完整！")
		return nil
	}

	slack := &Slack{
		oauthToken: oauthToken,
		appToken:   appToken,
		channel:    channel,
		client:     &http.Client{Timeout: 30 * time.Second},
		dsURL:      fmt.Sprintf("http://127.0.0.1:%d/api/v1/message?token=%s", config.Config.PORT, config.Config.API_TOKEN),
	}

	// 标记消息来源
	if name, ok := options["name"].(string); ok && name != "" {
		slack.dsURL = fmt.Sprintf("%s&source=%s", slack.dsURL, name)
	}

	// 设置代理
	if config.Config.PROXY != "" {
		proxyURL, err := url.Parse(config.Config.PROXY)
		if err == nil {
			slack.client.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			}
		}
	}

	// 启动服务（简化处理）
	slack.service = &SocketModeHandler{running: true}
	utils.Log.Info("Slack消息接收服务启动")

	return slack
}

// Stop 停止Slack服务
func (s *Slack) Stop() {
	if s.service != nil {
		s.service.running = false
		utils.Log.Info("Slack消息接收服务已停�?)
	}
}

// GetState 获取状�?func (s *Slack) GetState() bool {
	return s.client != nil
}

// SendMsg 发送Slack消息
func (s *Slack) SendMsg(title, text, image, link, userid string, buttons [][]map[string]string,
	originalMessageID, originalChatID *string) (*bool, string) {
	
	if s.client == nil {
		return utils.BoolPtr(false), "消息客户端未就绪"
	}
	
	if title == "" && text == "" {
		return utils.BoolPtr(false), "标题和内容不能同时为�?
	}
	
	defer func() {
		if r := recover(); r != nil {
			utils.Log.Errorf("Slack消息发送失�? %v", r)
		}
	}()
	
	var channel string
	if userid != "" {
		channel = userid
	} else {
		// 消息广播
		channel = s.findPublicChannel()
	}
	
	// 消息文本
	messageText := ""
	// 结构�?	blocks := make([]map[string]interface{}, 0)
	
	if image == "" {
		messageText = fmt.Sprintf("%s\n%s", title, text)
	} else {
		// 消息图片
		if image != "" {
			// 拼装消息内容
			blocks = append(blocks, map[string]interface{}{
				"type": "section",
				"text": map[string]string{
					"type": "mrkdwn",
					"text": fmt.Sprintf("*%s*\n%s", title, text),
				},
				"accessory": map[string]string{
					"type":      "image",
					"image_url": image,
					"alt_text":  title,
				},
			})
		}
		
		// 自定义按�?		if buttons != nil {
			for _, buttonRow := range buttons {
				elements := make([]map[string]interface{}, 0)
				for _, button := range buttonRow {
					if _, exists := button["url"]; exists {
						// URL按钮
						elements = append(elements, map[string]interface{}{
							"type": "button",
							"text": map[string]string{
								"type":  "plain_text",
								"text":  button["text"],
								"emoji": "true",
							},
							"url":       button["url"],
							"action_id": fmt.Sprintf("actionId-url-%s-%d", button["text"], len(elements)),
						})
					} else {
						// 回调按钮
						elements = append(elements, map[string]interface{}{
							"type": "button",
							"text": map[string]string{
								"type":  "plain_text",
								"text":  button["text"],
								"emoji": "true",
							},
							"value":     button["callback_data"],
							"action_id": fmt.Sprintf("actionId-%s", button["callback_data"]),
						})
					}
				}
				if len(elements) > 0 {
					blocks = append(blocks, map[string]interface{}{
						"type":     "actions",
						"elements": elements,
					})
				}
			}
		} else if link != "" {
			// 默认链接按钮
			blocks = append(blocks, map[string]interface{}{
				"type": "actions",
				"elements": []map[string]interface{}{
					{
						"type": "button",
						"text": map[string]string{
							"type":  "plain_text",
							"text":  "查看详情",
							"emoji": "true",
						},
						"value":     "click_me_url",
						"url":       link,
						"action_id": "actionId-url",
					},
				},
			})
		}
	}
	
	// 限制消息文本长度
	if len(messageText) > 1000 {
		messageText = messageText[:1000]
	}
	
	var result bool
	var errMsg string
	
	// 判断是编辑消息还是发送新消息
	if originalMessageID != nil && originalChatID != nil {
		// 编辑消息
		result, errMsg = s.chatUpdate(*originalChatID, *originalMessageID, messageText, blocks)
	} else {
		// 发送新消息
		result, errMsg = s.chatPostMessage(channel, messageText, blocks)
	}
	
	return &result, errMsg
}

// SendMediasMsg 发送媒体列表消�?func (s *Slack) SendMediasMsg(medias []map[string]interface{}, userid, title string, buttons [][]map[string]string,
	originalMessageID, originalChatID *string) *bool {
	
	if s.client == nil {
		result := false
		return &result
	}
	
	if len(medias) == 0 {
		result := false
		return &result
	}
	
	defer func() {
		if r := recover(); r != nil {
			utils.Log.Errorf("Slack消息发送失�? %v", r)
		}
	}()
	
	var channel string
	if userid != "" {
		channel = userid
	} else {
		// 消息广播
		channel = s.findPublicChannel()
	}
	
	// 消息主体
	titleSection := map[string]interface{}{
		"type": "section",
		"text": map[string]string{
			"type": "mrkdwn",
			"text": fmt.Sprintf("*%s*", title),
		},
	}
	
	blocks := []map[string]interface{}{titleSection}
	
	// 列表
	if len(medias) > 0 {
		blocks = append(blocks, map[string]interface{}{
			"type": "divider",
		})
		
		index := 1
		
		// 如果有自定义按钮，先添加所有媒体项，然后添加统一的按�?		if buttons != nil {
			// 添加媒体列表（不带单独的选择按钮�?			for _, media := range medias {
				posterImage := ""
				if img, ok := media["poster_image"].(string); ok {
					posterImage = img
				}
				
				voteStar := ""
				if vs, ok := media["vote_star"].(string); ok {
					voteStar = vs
				}
				
				detailLink := ""
				if link, ok := media["detail_link"].(string); ok {
					detailLink = link
				}
				
				titleYear := ""
				if ty, ok := media["title_year"].(string); ok {
					titleYear = ty
				}
				
				mediaType := ""
				if mt, ok := media["type"].(string); ok {
					mediaType = mt
				}
				
				overview := ""
				if ov, ok := media["overview"].(string); ok {
					overview = ov
				}
				
				if posterImage != "" {
					var text string
					if voteStar != "" {
						text = fmt.Sprintf("%d. *<%s|%s>*\n类型�?s\n%s\n%s",
							index, detailLink, titleYear, mediaType, voteStar, overview)
					} else {
						text = fmt.Sprintf("%d. *<%s|%s>*\n类型�?s\n%s",
							index, detailLink, titleYear, mediaType, overview)
					}
					
					blocks = append(blocks, map[string]interface{}{
						"type": "section",
						"text": map[string]string{
							"type": "mrkdwn",
							"text": text,
						},
						"accessory": map[string]string{
							"type":      "image",
							"image_url": posterImage,
							"alt_text":  titleYear,
						},
					})
					index++
				}
			}
			
			// 添加统一的自定义按钮（在所有媒体项之后�?			for _, buttonRow := range buttons {
				elements := make([]map[string]interface{}, 0)
				for _, button := range buttonRow {
					if _, exists := button["url"]; exists {
						elements = append(elements, map[string]interface{}{
							"type": "button",
							"text": map[string]string{
								"type":  "plain_text",
								"text":  button["text"],
								"emoji": "true",
							},
							"url":       button["url"],
							"action_id": fmt.Sprintf("actionId-url-%s-%d", button["text"], len(elements)),
						})
					} else {
						elements = append(elements, map[string]interface{}{
							"type": "button",
							"text": map[string]string{
								"type":  "plain_text",
								"text":  button["text"],
								"emoji": "true",
							},
							"value":     button["callback_data"],
							"action_id": fmt.Sprintf("actionId-%s", button["callback_data"]),
						})
					}
				}
				if len(elements) > 0 {
					blocks = append(blocks, map[string]interface{}{
						"type":     "actions",
						"elements": elements,
					})
				}
			}
		} else {
			// 使用默认的每个媒体项单独按钮
			for _, media := range medias {
				posterImage := ""
				if img, ok := media["poster_image"].(string); ok {
					posterImage = img
				}
				
				voteStar := ""
				if vs, ok := media["vote_star"].(string); ok {
					voteStar = vs
				}
				
				detailLink := ""
				if link, ok := media["detail_link"].(string); ok {
					detailLink = link
				}
				
				titleYear := ""
				if ty, ok := media["title_year"].(string); ok {
					titleYear = ty
				}
				
				mediaType := ""
				if mt, ok := media["type"].(string); ok {
					mediaType = mt
				}
				
				overview := ""
				if ov, ok := media["overview"].(string); ok {
					overview = ov
				}
				
				if posterImage != "" {
					var text string
					if voteStar != "" {
						text = fmt.Sprintf("%d. *<%s|%s>*\n类型�?s\n%s\n%s",
							index, detailLink, titleYear, mediaType, voteStar, overview)
					} else {
						text = fmt.Sprintf("%d. *<%s|%s>*\n类型�?s\n%s",
							index, detailLink, titleYear, mediaType, overview)
					}
					
					blocks = append(blocks, map[string]interface{}{
						"type": "section",
						"text": map[string]string{
							"type": "mrkdwn",
							"text": text,
						},
						"accessory": map[string]string{
							"type":      "image",
							"image_url": posterImage,
							"alt_text":  titleYear,
						},
					})
					
					// 使用默认选择按钮
					blocks = append(blocks, map[string]interface{}{
						"type": "actions",
						"elements": []map[string]interface{}{
							{
								"type": "button",
								"text": map[string]string{
									"type":  "plain_text",
									"text":  "选择",
									"emoji": "true",
								},
								"value":     fmt.Sprintf("%d", index),
								"action_id": fmt.Sprintf("actionId-%d", index),
							},
						},
					})
					index++
				}
			}
		}
	}
	
	var result bool
	var errMsg string
	
	// 判断是编辑消息还是发送新消息
	if originalMessageID != nil && originalChatID != nil {
		// 编辑消息
		result, errMsg = s.chatUpdate(*originalChatID, *originalMessageID, title, blocks)
	} else {
		// 发送新消息
		result, errMsg = s.chatPostMessage(channel, title, blocks)
	}
	
	if errMsg != "" {
		utils.Log.Errorf("Slack消息发送失�? %s", errMsg)
		resultVal := false
		return &resultVal
	}
	
	resultVal := result
	return &resultVal
}

// SendTorrentsMsg 发送种子列表消�?func (s *Slack) SendTorrentsMsg(torrents []map[string]interface{}, userid, title string, buttons [][]map[string]string,
	originalMessageID, originalChatID *string) *bool {
	
	if s.client == nil {
		return nil
	}
	
	defer func() {
		if r := recover(); r != nil {
			utils.Log.Errorf("Slack消息发送失�? %v", r)
		}
	}()
	
	var channel string
	if userid != "" {
		channel = userid
	} else {
		// 消息广播
		channel = s.findPublicChannel()
	}
	
	// 消息主体
	titleSection := map[string]interface{}{
		"type": "section",
		"text": map[string]string{
			"type": "mrkdwn",
			"text": fmt.Sprintf("*%s*", title),
		},
	}
	
	blocks := []map[string]interface{}{titleSection, {
		"type": "divider",
	}}
	
	// 列表
	index := 1
	
	// 如果有自定义按钮，先添加种子列表，然后添加统一的按�?	if buttons != nil {
		// 添加种子列表（不带单独的选择按钮�?		for _, context := range torrents {
			torrentInfo, ok := context["torrent_info"].(map[string]interface{})
			if !ok {
				continue
			}
			
			siteName := ""
			if sn, ok := torrentInfo["site_name"].(string); ok {
				siteName = sn
			}
			
			titleStr := ""
			if t, ok := torrentInfo["title"].(string); ok {
				titleStr = t
			}
			
			description := ""
			if desc, ok := torrentInfo["description"].(string); ok {
				description = desc
			}
			
			pageURL := ""
			if url, ok := torrentInfo["page_url"].(string); ok {
				pageURL = url
			}
			
			size := int64(0)
			if sz, ok := torrentInfo["size"].(int64); ok {
				size = sz
			}
			
			volumeFactor := ""
			if vf, ok := torrentInfo["volume_factor"].(string); ok {
				volumeFactor = vf
			}
			
			seeders := 0
			if sd, ok := torrentInfo["seeders"].(int); ok {
				seeders = sd
			}
			
			meta := utils.MetaInfo.NewMetaInfo(titleStr, description)
			seasonEpisode := meta.SeasonEpisode
			resourceTerm := meta.ResourceTerm
			videoTerm := meta.VideoTerm
			releaseGroup := meta.ReleaseGroup
			
			titleText := fmt.Sprintf("%s %s %s %s", seasonEpisode, resourceTerm, videoTerm, releaseGroup)
			titleText = regexp.MustCompile(`\s+`).ReplaceAllString(titleText, " ")
			titleText = strings.TrimSpace(titleText)
			
			seederStr := fmt.Sprintf("%d�?, seeders)
			
			text := fmt.Sprintf("%d. �?s�?%s|%s> %s %s %s\n%s",
				index, siteName, pageURL, titleText,
				utils.StringUtils.StrFileSize(size), volumeFactor, seederStr, description)
			
			blocks = append(blocks, map[string]interface{}{
				"type": "section",
				"text": map[string]string{
					"type": "mrkdwn",
					"text": text,
				},
			})
			index++
		}
		
		// 添加统一的自定义按钮
		for _, buttonRow := range buttons {
			elements := make([]map[string]interface{}, 0)
			for _, button := range buttonRow {
				if _, exists := button["url"]; exists {
					elements = append(elements, map[string]interface{}{
						"type": "button",
						"text": map[string]string{
							"type":  "plain_text",
							"text":  button["text"],
							"emoji": "true",
						},
						"url":       button["url"],
						"action_id": fmt.Sprintf("actionId-url-%s-%d", button["text"], len(elements)),
					})
				} else {
					elements = append(elements, map[string]interface{}{
						"type": "button",
						"text": map[string]string{
							"type":  "plain_text",
							"text":  button["text"],
							"emoji": "true",
						},
						"value":     button["callback_data"],
						"action_id": fmt.Sprintf("actionId-%s", button["callback_data"]),
					})
				}
			}
			if len(elements) > 0 {
				blocks = append(blocks, map[string]interface{}{
					"type":     "actions",
					"elements": elements,
				})
			}
		}
	} else {
		// 使用默认的每个种子单独按�?		for _, context := range torrents {
			torrentInfo, ok := context["torrent_info"].(map[string]interface{})
			if !ok {
				continue
			}
			
			siteName := ""
			if sn, ok := torrentInfo["site_name"].(string); ok {
				siteName = sn
			}
			
			titleStr := ""
			if t, ok := torrentInfo["title"].(string); ok {
				titleStr = t
			}
			
			description := ""
			if desc, ok := torrentInfo["description"].(string); ok {
				description = desc
			}
			
			pageURL := ""
			if url, ok := torrentInfo["page_url"].(string); ok {
				pageURL = url
			}
			
			size := int64(0)
			if sz, ok := torrentInfo["size"].(int64); ok {
				size = sz
			}
			
			volumeFactor := ""
			if vf, ok := torrentInfo["volume_factor"].(string); ok {
				volumeFactor = vf
			}
			
			seeders := 0
			if sd, ok := torrentInfo["seeders"].(int); ok {
				seeders = sd
			}
			
			meta := utils.MetaInfo.NewMetaInfo(titleStr, description)
			seasonEpisode := meta.SeasonEpisode
			resourceTerm := meta.ResourceTerm
			videoTerm := meta.VideoTerm
			releaseGroup := meta.ReleaseGroup
			
			titleText := fmt.Sprintf("%s %s %s %s", seasonEpisode, resourceTerm, videoTerm, releaseGroup)
			titleText = regexp.MustCompile(`\s+`).ReplaceAllString(titleText, " ")
			titleText = strings.TrimSpace(titleText)
			
			seederStr := fmt.Sprintf("%d�?, seeders)
			
			text := fmt.Sprintf("%d. �?s�?%s|%s> %s %s %s\n%s",
				index, siteName, pageURL, titleText,
				utils.StringUtils.StrFileSize(size), volumeFactor, seederStr, description)
			
			blocks = append(blocks, map[string]interface{}{
				"type": "section",
				"text": map[string]string{
					"type": "mrkdwn",
					"text": text,
				},
			})
			
			blocks = append(blocks, map[string]interface{}{
				"type": "actions",
				"elements": []map[string]interface{}{
					{
						"type": "button",
						"text": map[string]string{
							"type":  "plain_text",
							"text":  "选择",
							"emoji": "true",
						},
						"value":     fmt.Sprintf("%d", index),
						"action_id": fmt.Sprintf("actionId-%d", index),
					},
				},
			})
			index++
		}
	}
	
	var result bool
	var errMsg string
	
	// 判断是编辑消息还是发送新消息
	if originalMessageID != nil && originalChatID != nil {
		// 编辑消息
		result, errMsg = s.chatUpdate(*originalChatID, *originalMessageID, title, blocks)
	} else {
		// 发送新消息
		result, errMsg = s.chatPostMessage(channel, title, blocks)
	}
	
	if errMsg != "" {
		utils.Log.Errorf("Slack消息发送失�? %s", errMsg)
		resultVal := false
		return &resultVal
	}
	
	resultVal := result
	return &resultVal
}

// DeleteMsg 删除Slack消息
func (s *Slack) DeleteMsg(messageID string, chatID *string) *bool {
	if s.client == nil {
		return nil
	}
	
	defer func() {
		if r := recover(); r != nil {
			utils.Log.Errorf("删除Slack消息异常: %v", r)
		}
	}()
	
	// 确定要删除消息的频道ID
	var targetChannel string
	if chatID != nil {
		targetChannel = *chatID
	} else {
		targetChannel = s.findPublicChannel()
	}
	
	if targetChannel == "" {
		utils.Log.Error("无法确定要删除消息的Slack频道")
		result := false
		return &result
	}
	
	// 删除消息
	apiURL := "https://slack.com/api/chat.delete"
	
	postData := map[string]string{
		"channel": targetChannel,
		"ts":      messageID,
	}
	
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", s.oauthToken),
		"Content-Type":  "application/json",
	}
	
	jsonData, err := json.Marshal(postData)
	if err != nil {
		utils.Log.Errorf("序列化删除消息数据失�? %v", err)
		result := false
		return &result
	}
	
	resp, err := s.client.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		utils.Log.Errorf("发送删除消息请求失�? %v", err)
		result := false
		return &result
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		utils.Log.Errorf("删除Slack消息失败，状态码�?d", resp.StatusCode)
		result := false
		return &result
	}
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		utils.Log.Errorf("解析删除消息响应失败: %v", err)
		resultVal := false
		return &resultVal
	}
	
	if ok, okExists := result["ok"].(bool); okExists && ok {
		utils.Log.Infof("成功删除Slack消息: channel=%s, ts=%s", targetChannel, messageID)
		resultVal := true
		return &resultVal
	} else {
		errorMsg := ""
		if err, errExists := result["error"].(string); errExists {
			errorMsg = err
		}
		utils.Log.Errorf("删除Slack消息失败: %s", errorMsg)
		resultVal := false
		return &resultVal
	}
}

// findPublicChannel 查找公共频道
func (s *Slack) findPublicChannel() string {
	if s.client == nil {
		return ""
	}
	
	conversationID := ""
	
	// 这里简化处理，实际应该调用Slack API获取频道列表
	// 由于Go中需要使用slack库来实现完整功能，这里做了简化处�?	if s.channel != "" {
		conversationID = s.channel
	} else {
		conversationID = "全体"
	}
	
	return conversationID
}

// chatPostMessage 发送消�?func (s *Slack) chatPostMessage(channel, text string, blocks []map[string]interface{}) (bool, string) {
	apiURL := "https://slack.com/api/chat.postMessage"
	
	postData := map[string]interface{}{
		"channel": channel,
		"text":    text,
		"mrkdwn":  true,
	}
	
	if len(blocks) > 0 {
		postData["blocks"] = blocks
	}
	
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", s.oauthToken),
		"Content-Type":  "application/json",
	}
	
	jsonData, err := json.Marshal(postData)
	if err != nil {
		return false, fmt.Sprintf("序列化消息数据失�? %v", err)
	}
	
	resp, err := s.client.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return false, fmt.Sprintf("发送消息请求失�? %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return false, fmt.Sprintf("发送消息失败，状态码�?d", resp.StatusCode)
	}
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Sprintf("解析消息响应失败: %v", err)
	}
	
	if ok, okExists := result["ok"].(bool); okExists && ok {
		return true, ""
	} else {
		errorMsg := ""
		if err, errExists := result["error"].(string); errExists {
			errorMsg = err
		}
		return false, errorMsg
	}
}

// chatUpdate 编辑消息
func (s *Slack) chatUpdate(channel, ts, text string, blocks []map[string]interface{}) (bool, string) {
	apiURL := "https://slack.com/api/chat.update"
	
	postData := map[string]interface{}{
		"channel": channel,
		"ts":      ts,
		"text":    text,
	}
	
	if len(blocks) > 0 {
		postData["blocks"] = blocks
	}
	
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", s.oauthToken),
		"Content-Type":  "application/json",
	}
	
	jsonData, err := json.Marshal(postData)
	if err != nil {
		return false, fmt.Sprintf("序列化编辑消息数据失�? %v", err)
	}
	
	resp, err := s.client.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return false, fmt.Sprintf("发送编辑消息请求失�? %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return false, fmt.Sprintf("编辑消息失败，状态码�?d", resp.StatusCode)
	}
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Sprintf("解析编辑消息响应失败: %v", err)
	}
	
	if ok, okExists := result["ok"].(bool); okExists && ok {
		return true, ""
	} else {
		errorMsg := ""
		if err, errExists := result["error"].(string); errExists {
			errorMsg = err
		}
		return false, errorMsg
	}
}
