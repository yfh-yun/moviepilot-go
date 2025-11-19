// Package wechat 微信企业版客户端实现
package wechat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/pkg/httpclient"

	"go.uber.org/zap"
)

// WeChatClient 微信企业版客户端
type WeChatClient struct {
	corpID     string
	appSecret   string
	appID       string
	proxy       string
	baseURL     string
	httpClient  *httpclient.Client
	logger      *zap.Logger

	// Token相关
	accessToken     string
	accessTokenTTL  time.Time
	tokenMutex      sync.RWMutex
}

// WeChatConfig 微信配置
type WeChatConfig struct {
	CorpID     string `json:"corp_id"`      // 企业ID
	AppSecret   string `json:"app_secret"`    // 应用密钥
	AppID       string `json:"app_id"`        // 应用ID
	Proxy       string `json:"proxy"`         // 代理地址
}

// AccessTokenResponse 访问令牌响应
type AccessTokenResponse struct {
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// MessageResponse 消息响应
type MessageResponse struct {
	ErrCode      int    `json:"errcode"`
	ErrMsg       string `json:"errmsg"`
	InvalidUser  string `json:"invaliduser"`
	InvalidParty string `json:"invalidparty"`
	InvalidTag   string `json:"invalidtag"`
}

// MenuButton 菜单按钮
type MenuButton struct {
	Type      string        `json:"type,omitempty"`
	Name      string        `json:"name"`
	Key       string        `json:"key,omitempty"`
	URL       string        `json:"url,omitempty"`
	SubButton []MenuButton  `json:"sub_button,omitempty"`
}

// MenuRequest 菜单请求
type MenuRequest struct {
	Button []MenuButton `json:"button"`
}

// Command 命令定义
type Command struct {
	Function    string                 `json:"function"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Data        map[string]interface{} `json:"data"`
}

// NewWeChatClient 创建微信客户端
func NewWeChatClient(config *WeChatConfig) *WeChatClient {
	if config == nil {
		return nil
	}

	baseURL := "https://qyapi.weixin.qq.com"
	if config.Proxy != "" {
		baseURL = config.Proxy
	}

	client := &WeChatClient{
		corpID:    config.CorpID,
		appSecret:  config.AppSecret,
		appID:      config.AppID,
		proxy:      config.Proxy,
		baseURL:    baseURL,
		httpClient: httpclient.NewClient(httpclient.Options{
			Timeout: 30 * time.Second,
			Headers: map[string]string{
				"Content-Type": "application/json",
				"User-Agent":   "MoviePilot-WeChat/1.0",
			},
		}),
		logger: logger.Logger,
	}

	return client
}

// GetState 获取微信客户端状态
func (w *WeChatClient) GetState() bool {
	token, err := w.getAccessToken(context.Background())
	return token != "" && err == nil
}

// getAccessToken 获取访问令牌（带缓存）
func (w *WeChatClient) getAccessToken(ctx context.Context) (string, error) {
	w.tokenMutex.RLock()
	
	// 检查缓存的token是否有效
	if w.accessToken != "" && time.Now().Before(w.accessTokenTTL) {
		token := w.accessToken
		w.tokenMutex.RUnlock()
		return token, nil
	}
	
	w.tokenMutex.RUnlock()

	// 需要获取新token
	w.tokenMutex.Lock()
	defer w.tokenMutex.Unlock()

	// 双重检查，防止重复获取
	if w.accessToken != "" && time.Now().Before(w.accessTokenTTL) {
		return w.accessToken, nil
	}

	return w.refreshAccessToken(ctx)
}

// refreshAccessToken 刷新访问令牌
func (w *WeChatClient) refreshAccessToken(ctx context.Context) (string, error) {
	if w.corpID == "" || w.appSecret == "" {
		return "", fmt.Errorf("corp_id or app_secret is empty")
	}

	tokenURL := fmt.Sprintf("%s/cgi-bin/gettoken?corpid=%s&corpsecret=%s",
		w.baseURL, w.corpID, w.appSecret)

	req, err := http.NewRequestWithContext(ctx, "GET", tokenURL, nil)
	if err != nil {
		return "", fmt.Errorf("create token request failed: %w", err)
	}

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("get access token failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("get access token failed with status: %d", resp.StatusCode)
	}

	var tokenResp AccessTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("decode token response failed: %w", err)
	}

	if tokenResp.ErrCode != 0 {
		return "", fmt.Errorf("get access token failed: %s (code: %d)", tokenResp.ErrMsg, tokenResp.ErrCode)
	}

	// 缓存token
	w.accessToken = tokenResp.AccessToken
	w.accessTokenTTL = time.Now().Add(time.Duration(tokenResp.ExpiresIn-300) * time.Second) // 提前5分钟过期

	w.logger.Info("Access token refreshed",
		zap.String("expires_in", fmt.Sprintf("%ds", tokenResp.ExpiresIn)))

	return w.accessToken, nil
}

// SendMessage 发送消息
func (w *WeChatClient) SendMessage(ctx context.Context, msgType string, content interface{}, userID string) error {
	token, err := w.getAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("get access token failed: %w", err)
	}

	if userID == "" {
		userID = "@all"
	}

	message := map[string]interface{}{
		"touser":  userID,
		"msgtype": msgType,
		"agentid": w.appID,
	}

	// 根据消息类型设置内容
	switch msgType {
	case "text":
		message["text"] = map[string]string{
			"content": fmt.Sprintf("%v", content),
		}
	case "news":
		message["news"] = content
	default:
		return fmt.Errorf("unsupported message type: %s", msgType)
	}

	return w.postRequest(ctx, "/cgi-bin/message/send", message)
}

// SendTextMessage 发送文本消息
func (w *WeChatClient) SendTextMessage(ctx context.Context, title, text, userID string) error {
	content := fmt.Sprintf("%s\n\n%s", title, text)
	return w.SendMessage(ctx, "text", content, userID)
}

// SendImageMessage 发送图文消息
func (w *WeChatClient) SendImageMessage(ctx context.Context, title, description, imageURL, link, userID string) error {
	if userID == "" {
		userID = "@all"
	}

	content := map[string]interface{}{
		"articles": []map[string]interface{}{
			{
				"title":       title,
				"description": description,
				"picurl":      imageURL,
				"url":         link,
			},
		},
	}

	return w.SendMessage(ctx, "news", content, userID)
}

// SendMediaMessages 发送媒体列表消息
func (w *WeChatClient) SendMediaMessages(ctx context.Context, medias []MediaInfo, userID string) error {
	if userID == "" {
		userID = "@all"
	}

	var articles []map[string]interface{}
	for i, media := range medias {
		var title string
		if media.VoteAverage > 0 {
			title = fmt.Sprintf("%d. %s\n类型：%s，评分：%.1f", 
				i+1, media.TitleYear, media.Type, media.VoteAverage)
		} else {
			title = fmt.Sprintf("%d. %s\n类型：%s", 
				i+1, media.TitleYear, media.Type)
		}

		// 第一条使用详细图片，其他使用海报
		picURL := media.GetMessageImage()
		if i > 0 {
			picURL = media.GetPosterImage()
		}

		articles = append(articles, map[string]interface{}{
			"title":       title,
			"description": "",
			"picurl":      picURL,
			"url":         media.DetailLink,
		})
	}

	content := map[string]interface{}{
		"articles": articles,
	}

	return w.SendMessage(ctx, "news", content, userID)
}

// MediaInfo 媒体信息
type MediaInfo struct {
	Title       string  `json:"title"`
	Year        int     `json:"year"`
	Type        string  `json:"type"`
	VoteAverage float64 `json:"vote_average"`
	PosterImage string  `json:"poster_image"`
	MessageImage string `json:"message_image"`
	DetailLink  string  `json:"detail_link"`
}

// TitleYear 标题年份
func (m *MediaInfo) TitleYear() string {
	if m.Year > 0 {
		return fmt.Sprintf("%s (%d)", m.Title, m.Year)
	}
	return m.Title
}

// GetMessageImage 获取消息图片
func (m *MediaInfo) GetMessageImage() string {
	if m.MessageImage != "" {
		return m.MessageImage
	}
	return m.PosterImage
}

// GetPosterImage 获取海报图片
func (m *MediaInfo) GetPosterImage() string {
	return m.PosterImage
}

// CreateMenus 创建微信菜单
func (w *WeChatClient) CreateMenus(ctx context.Context, commands map[string]Command) error {
	// 按分类分组命令
	categoryMap := make(map[string]map[string]Command)
	for key, cmd := range commands {
		category := cmd.Category
		if category == "" {
			continue
		}

		if categoryMap[category] == nil {
			categoryMap[category] = make(map[string]Command)
		}
		categoryMap[category][key] = cmd
	}

	// 构建菜单按钮
	var buttons []MenuButton
	for category, cmds := range categoryMap {
		// 构建子菜单
		var subButtons []MenuButton
		count := 0
		for key, cmd := range cmds {
			if count >= 5 { // 子菜单最多5个
				break
			}

			subButtons = append(subButtons, MenuButton{
				Type: "click",
				Name: cmd.Description,
				Key:  key,
			})
			count++
		}

		if len(subButtons) > 0 {
			buttons = append(buttons, MenuButton{
				Name:      category,
				SubButton: subButtons,
			})
		}

		// 一级菜单最多3个
		if len(buttons) >= 3 {
			break
		}
	}

	menuReq := MenuRequest{
		Button: buttons,
	}

	token, err := w.getAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("get access token failed: %w", err)
	}

	menuURL := fmt.Sprintf("%s/cgi-bin/menu/create?access_token=%s&agentid=%s",
		w.baseURL, token, w.appID)

	reqData, err := json.Marshal(menuReq)
	if err != nil {
		return fmt.Errorf("marshal menu request failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", menuURL, bytes.NewBuffer(reqData))
	if err != nil {
		return fmt.Errorf("create menu request failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("create menu request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("create menu failed with status: %d", resp.StatusCode)
	}

	var menuResp MessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&menuResp); err != nil {
		return fmt.Errorf("decode menu response failed: %w", err)
	}

	if menuResp.ErrCode != 0 {
		return fmt.Errorf("create menu failed: %s (code: %d)", menuResp.ErrMsg, menuResp.ErrCode)
	}

	w.logger.Info("Menu created successfully",
		zap.Int("button_count", len(menuResp.Button)),
		zap.Strings("categories", w.getCategoryNames(commands)))

	return nil
}

// DeleteMenus 删除微信菜单
func (w *WeChatClient) DeleteMenus(ctx context.Context) error {
	token, err := w.getAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("get access token failed: %w", err)
	}

	deleteURL := fmt.Sprintf("%s/cgi-bin/menu/delete?access_token=%s&agentid=%s",
		w.baseURL, token, w.appID)

	req, err := http.NewRequestWithContext(ctx, "GET", deleteURL, nil)
	if err != nil {
		return fmt.Errorf("create delete menu request failed: %w", err)
	}

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("delete menu request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("delete menu failed with status: %d", resp.StatusCode)
	}

	var deleteResp MessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&deleteResp); err != nil {
		return fmt.Errorf("decode delete menu response failed: %w", err)
	}

	if deleteResp.ErrCode != 0 {
		return fmt.Errorf("delete menu failed: %s (code: %d)", deleteResp.ErrMsg, deleteResp.ErrCode)
	}

	w.logger.Info("Menu deleted successfully")
	return nil
}

// postRequest 发送POST请求
func (w *WeChatClient) postRequest(ctx context.Context, path string, data interface{}) error {
	token, err := w.getAccessToken(ctx)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s%s?access_token=%s", w.baseURL, path, token)

	reqData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal request data failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqData))
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	var response MessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("decode response failed: %w", err)
	}

	if response.ErrCode == 42001 {
		// Token过期，强制刷新
		w.tokenMutex.Lock()
		w.accessToken = ""
		w.accessTokenTTL = time.Time{}
		w.tokenMutex.Unlock()

		newToken, err := w.getAccessToken(ctx)
		if err != nil {
			return fmt.Errorf("refresh token failed: %w", err)
		}

		// 重试请求
		return w.postRequest(ctx, path, data)
	}

	if response.ErrCode != 0 {
		return fmt.Errorf("wechat api error: %s (code: %d)", response.ErrMsg, response.ErrCode)
	}

	return nil
}

// getCategoryNames 获取分类名称列表
func (w *WeChatClient) getCategoryNames(commands map[string]Command) []string {
	categorySet := make(map[string]bool)
	for _, cmd := range commands {
		if cmd.Category != "" {
			categorySet[cmd.Category] = true
		}
	}

	var categories []string
	for category := range categorySet {
		categories = append(categories, category)
	}

	return categories
}

// HealthCheck 健康检查
func (w *WeChatClient) HealthCheck(ctx context.Context) error {
	_, err := w.getAccessToken(ctx)
	return err
}

// Close 关闭客户端
func (w *WeChatClient) Close() error {
	// 清理资源
	w.tokenMutex.Lock()
	w.accessToken = ""
	w.accessTokenTTL = time.Time{}
	w.tokenMutex.Unlock()

	return nil
}