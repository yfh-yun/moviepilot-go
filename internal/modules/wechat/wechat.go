package wechat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
	
	"moviepilot-go/internal/utils"
	"moviepilot-go/pkg/models"
)

// RetryException 重试异常
type RetryException struct {
	message string
}

func (e *RetryException) Error() string {
	return e.message
}

// WeChat 企业微信客户�?type WeChat struct {
	// 企业微信Token
	accessToken string
	// 企业微信Token过期时间
	expiresIn int64
	// 企业微信Token获取时间
	accessTokenTime time.Time
	// 企业微信CorpID
	corpID string
	// 企业微信AppSecret
	appSecret string
	// 企业微信AppID
	appID string
	// 代理
	proxy string
	
	// 企业微信发送消息URL
	sendMsgURL string
	// 企业微信获取TokenURL
	tokenURL string
	// 企业微信创建菜单URL
	createMenuURL string
	// 企业微信删除菜单URL
	deleteMenuURL string
	
	// 互斥�?	mutex sync.Mutex
}

// NewWeChat 创建新的WeChat实例
func NewWeChat(corpID, appSecret, appID, proxy string) *WeChat {
	if corpID == "" || appSecret == "" || appID == "" {
		// 记录错误日志
		fmt.Println("企业微信配置不完整！")
		return nil
	}
	
	w := &WeChat{
		corpID:    corpID,
		appSecret: appSecret,
		appID:     appID,
		proxy:     proxy,
	}
	
	// 设置默认代理
	if w.proxy == "" {
		w.proxy = "https://qyapi.weixin.qq.com"
	}
	
	// 构造URL
	w.sendMsgURL = fmt.Sprintf("%s/cgi-bin/message/send?access_token=%%s", w.proxy)
	w.tokenURL = fmt.Sprintf("%s/cgi-bin/gettoken?corpid=%s&corpsecret=%s", w.proxy, w.corpID, w.appSecret)
	w.createMenuURL = fmt.Sprintf("%s/cgi-bin/menu/create?access_token=%%s&agentid=%s", w.proxy, w.appID)
	w.deleteMenuURL = fmt.Sprintf("%s/cgi-bin/menu/delete?access_token=%%s&agentid=%s", w.proxy, w.appID)
	
	// 获取access token
	w.getAccessToken()
	
	return w
}

// GetState 获取状�?func (w *WeChat) GetState() bool {
	return w.getAccessToken() != ""
}

// getAccessToken 获取微信Token
func (w *WeChat) getAccessToken() string {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	
	tokenFlag := true
	if w.accessToken == "" {
		tokenFlag = false
	} else {
		if time.Since(w.accessTokenTime).Seconds() >= float64(w.expiresIn) {
			tokenFlag = false
		}
	}
	
	if !tokenFlag {
		if w.corpID == "" || w.appSecret == "" {
			return ""
		}
		
		res, err := http.Get(w.tokenURL)
		if err != nil {
			fmt.Printf("获取微信access_token失败: %v\n", err)
			return ""
		}
		defer res.Body.Close()
		
		if res.StatusCode != 200 {
			fmt.Printf("获取微信access_token失败，错误码�?d，错误原因：%s\n", res.StatusCode, res.Status)
			return ""
		}
		
		body, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Printf("读取响应失败: %v\n", err)
			return ""
		}
		
		var retJSON struct {
			ErrCode     int    `json:"errcode"`
			ErrMsg      string `json:"errmsg"`
			AccessToken string `json:"access_token"`
			ExpiresIn   int64  `json:"expires_in"`
		}
		
		err = json.Unmarshal(body, &retJSON)
		if err != nil {
			fmt.Printf("解析JSON失败: %v\n", err)
			return ""
		}
		
		if retJSON.ErrCode == 0 {
			w.accessToken = retJSON.AccessToken
			w.expiresIn = retJSON.ExpiresIn
			w.accessTokenTime = time.Now()
		} else {
			fmt.Printf("获取微信access_token失败，错误信息：%s\n", retJSON.ErrMsg)
			return ""
		}
	}
	
	return w.accessToken
}

// splitContent 将内容分块为不超�?maxBytes 字节的块
func (w *WeChat) splitContent(content string, maxBytes int) []string {
	contentChunks := make([]string, 0)
	currentChunk := make([]byte, 0)
	
	for _, line := range strings.Split(content, "\n") {
		encodedLine := []byte(line + "\n")
		lineLength := len(encodedLine)
		
		if lineLength > maxBytes {
			// 在处理长行之前，先将 currentChunk 添加�?contentChunks
			if len(currentChunk) > 0 {
				contentChunks = append(contentChunks, string(currentChunk))
				currentChunk = make([]byte, 0)
			}
			
			// 处理长行，拆分为多个不超�?maxBytes 的块
			start := 0
			for start < lineLength {
				end := start + maxBytes
				if end >= lineLength {
					end = lineLength
				} else {
					// 调整以避免拆分多字节字符
					for end > start && (encodedLine[end] & 0xC0) == 0x80 {
						end--
					}
					if end == start {
						// 单个字符超过�?maxBytes，强制包含整个字�?						end = start + 1
						for end < lineLength && (encodedLine[end] & 0xC0) == 0x80 {
							end++
						}
					}
				}
				truncatedLine := string(encodedLine[start:end])
				contentChunks = append(contentChunks, strings.TrimSpace(truncatedLine))
				start = end
			}
			continue // 继续处理下一�?		}
		
		// 检查添加当前行后是否会超过 maxBytes
		if len(currentChunk) + lineLength > maxBytes {
			// �?currentChunk 添加�?contentChunks
			contentChunks = append(contentChunks, string(currentChunk))
			currentChunk = make([]byte, 0)
		}
		
		// 将当前行添加�?currentChunk
		currentChunk = append(currentChunk, encodedLine...)
	}
	
	// 处理剩余�?currentChunk
	if len(currentChunk) > 0 {
		contentChunks = append(contentChunks, string(currentChunk))
	}
	
	return contentChunks
}

// sendMessage 发送文本消�?func (w *WeChat) sendMessage(title, text, userID, link string) bool {
	if title == "" {
		fmt.Println("消息标题不能为空")
		return false
	}
	
	var content string
	if text != "" {
		formattedText := strings.ReplaceAll(text, "\n\n", "\n")
		content = fmt.Sprintf("%s\n%s", title, formattedText)
	} else {
		content = title
	}
	
	if link != "" {
		content = fmt.Sprintf("%s\n点击查看�?s", content, link)
	}
	
	if userID == "" {
		userID = "@all"
	}
	
	// 分块处理逻辑
	contentChunks := w.splitContent(content, 2048)
	
	// 逐块发送消�?	for _, chunk := range contentChunks {
		reqJSON := map[string]interface{}{
			"touser":  userID,
			"msgtype": "text",
			"agentid": w.appID,
			"text": map[string]string{
				"content": chunk,
			},
			"safe":                    0,
			"enable_id_trans":         0,
			"enable_duplicate_check":  0,
		}
		
		// 如果是超长消息，有一个发送失败就全部失败
		if !w.postRequest(w.sendMsgURL, reqJSON) {
			return false
		}
	}
	
	return true
}

// sendImageMessage 发送图文消�?func (w *WeChat) sendImageMessage(title, text, imageURL, userID, link string) bool {
	if text != "" {
		text = strings.ReplaceAll(text, "\n\n", "\n")
	}
	
	if userID == "" {
		userID = "@all"
	}
	
	reqJSON := map[string]interface{}{
		"touser":  userID,
		"msgtype": "news",
		"agentid": w.appID,
		"news": map[string]interface{}{
			"articles": []map[string]string{
				{
					"title":       title,
					"description": text,
					"picurl":      imageURL,
					"url":         link,
				},
			},
		},
	}
	
	return w.postRequest(w.sendMsgURL, reqJSON)
}

// SendMsg 微信消息发送入口，支持文本、图片、链接跳转、指定发送对�?func (w *WeChat) SendMsg(title, text, image, userID, link string) bool {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("发送消息失�? %v\n", r)
		}
	}()
	
	if w.getAccessToken() == "" {
		fmt.Println("获取微信access_token失败，请检查参数配�?)
		return false
	}
	
	var retCode bool
	if image != "" {
		retCode = w.sendImageMessage(title, text, image, userID, link)
	} else {
		retCode = w.sendMessage(title, text, userID, link)
	}
	
	return retCode
}

// SendMediasMsg 发送列表类消息
func (w *WeChat) SendMediasMsg(medias []models.MediaInfo, userID string) bool {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("发送消息失�? %v\n", r)
		}
	}()
	
	if w.getAccessToken() == "" {
		fmt.Println("获取微信access_token失败，请检查参数配�?)
		return false
	}
	
	if userID == "" {
		userID = "@all"
	}
	
	articles := make([]map[string]string, 0)
	index := 1
	
	for _, media := range medias {
		var title string
		if media.VoteAverage > 0 {
			title = fmt.Sprintf("%d. %s\n类型�?s，评分：%.1f", index, media.TitleYear, media.Type, media.VoteAverage)
		} else {
			title = fmt.Sprintf("%d. %s\n类型�?s", index, media.TitleYear, media.Type)
		}
		
		article := map[string]string{
			"title":       title,
			"description": "",
			"picurl":      "", // 这里需要根据实际情况获取图片URL
			"url":         media.DetailLink,
		}
		
		articles = append(articles, article)
		index++
	}
	
	reqJSON := map[string]interface{}{
		"touser":  userID,
		"msgtype": "news",
		"agentid": w.appID,
		"news": map[string]interface{}{
			"articles": articles,
		},
	}
	
	return w.postRequest(w.sendMsgURL, reqJSON)
}

// SendTorrentsMsg 发送列表消�?func (w *WeChat) SendTorrentsMsg(torrents []models.Context, userID, title, link string) bool {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("发送消息失�? %v\n", r)
		}
	}()
	
	if w.getAccessToken() == "" {
		fmt.Println("获取微信access_token失败，请检查参数配�?)
		return false
	}
	
	// 先发送标�?	if title != "" {
		w.sendMessage(title, "", userID, link)
	}
	
	// 发送列�?	if userID == "" {
		userID = "@all"
	}
	
	articles := make([]map[string]string, 0)
	index := 1
	
	for _, context := range torrents {
		torrent := context.TorrentInfo
		// 这里需要根据实际情况构造meta信息
		torrentTitle := fmt.Sprintf("%d.�?s�?%s %s %s %s %s %s�?,
			index, torrent.SiteName, "", "", "", "", utils.StrFileSize(torrent.Size), torrent.VolumeFactor, torrent.Seeders)
		torrentTitle = strings.Join(strings.Fields(torrentTitle), " ")
		
		article := map[string]string{
			"title":       torrentTitle,
			"description": "", // 根据索引设置描述
			"picurl":      "", // 根据索引设置图片URL
			"url":         torrent.PageUrl,
		}
		
		articles = append(articles, article)
		index++
	}
	
	reqJSON := map[string]interface{}{
		"touser":  userID,
		"msgtype": "news",
		"agentid": w.appID,
		"news": map[string]interface{}{
			"articles": articles,
		},
	}
	
	return w.postRequest(w.sendMsgURL, reqJSON)
}

// postRequest 向微信发送请�?func (w *WeChat) postRequest(url string, reqJSON map[string]interface{}) bool {
	// 替换access_token
	url = fmt.Sprintf(url, w.getAccessToken())
	
	// 序列化JSON
	jsonData, err := json.Marshal(reqJSON)
	if err != nil {
		fmt.Printf("序列化JSON失败: %v\n", err)
		return false
	}
	
	// 发送POST请求
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("发送请求失�? %v\n", err)
		return false
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		fmt.Printf("发送请求失败，错误码：%d，错误原因：%s\n", resp.StatusCode, resp.Status)
		return false
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应失败: %v\n", err)
		return false
	}
	
	var retJSON struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	
	err = json.Unmarshal(body, &retJSON)
	if err != nil {
		fmt.Printf("解析JSON失败: %v\n", err)
		return false
	}
	
	if retJSON.ErrCode == 0 {
		return true
	} else {
		if retJSON.ErrCode == 42001 {
			// access_token已过期，重新获取
			w.mutex.Lock()
			w.accessToken = ""
			w.mutex.Unlock()
			fmt.Printf("access_token已过期，尝试重新获取access_token, errcode: %d, errmsg: %s\n", retJSON.ErrCode, retJSON.ErrMsg)
			return false
		} else {
			fmt.Printf("发送请求失败，错误信息�?s\n", retJSON.ErrMsg)
			return false
		}
	}
}

// CreateMenus 自动注册微信菜单
func (w *WeChat) CreateMenus(commands map[string]map[string]interface{}) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("创建菜单失败: %v\n", r)
		}
	}()
	
	// 请求URL
	reqURL := fmt.Sprintf(w.createMenuURL, w.getAccessToken())
	
	// 对commands按category分组
	categoryDict := make(map[string]map[string]map[string]interface{})
	for key, value := range commands {
		if category, ok := value["category"].(string); ok && category != "" {
			if _, exists := categoryDict[category]; !exists {
				categoryDict[category] = make(map[string]map[string]interface{})
			}
			categoryDict[category][key] = value
		}
	}
	
	// 一级菜�?	buttons := make([]map[string]interface{}, 0)
	for category, menu := range categoryDict {
		// 二级菜单
		subButtons := make([]map[string]string, 0)
		for key, value := range menu {
			if description, ok := value["description"].(string); ok {
				subButton := map[string]string{
					"type": "click",
					"name": description,
					"key":  key,
				}
				subButtons = append(subButtons, subButton)
				
				// 限制最�?个子菜单
				if len(subButtons) >= 5 {
					break
				}
			}
		}
		
		button := map[string]interface{}{
			"name":       category,
			"sub_button": subButtons,
		}
		buttons = append(buttons, button)
		
		// 限制最�?个一级菜�?		if len(buttons) >= 3 {
			break
		}
	}
	
	if len(buttons) > 0 {
		// 发送请�?		reqJSON := map[string]interface{}{
			"button": buttons,
		}
		w.postRequest(reqURL, reqJSON)
	}
}

// DeleteMenus 删除微信菜单
func (w *WeChat) DeleteMenus() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("删除菜单失败: %v\n", r)
		}
	}()
	
	// 请求URL
	reqURL := fmt.Sprintf(w.deleteMenuURL, w.getAccessToken())
	
	// 发送请�?	client := &http.Client{}
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		fmt.Printf("创建请求失败: %v\n", err)
		return
	}
	
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("发送请求失�? %v\n", err)
		return
	}
	defer resp.Body.Close()
}
