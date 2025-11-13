package synologychat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"moviepilot-go/internal/config"
	"moviepilot-go/internal/utils"
)

// SynologyChat Synology Chat模块
type SynologyChat struct {
	webhookURL string
	token      string
	domain     string
	client     *http.Client
	mutex      sync.Mutex
}

// NewSynologyChat 创建SynologyChat实例
func NewSynologyChat(webhookURL, token string, options map[string]interface{}) *SynologyChat {
	if webhookURL == "" || token == "" {
		utils.Log.Error("SynologyChat配置不完整！")
		return nil
	}

	sc := &SynologyChat{
		webhookURL: webhookURL,
		token:      token,
		client:     &http.Client{Timeout: 30 * time.Second},
	}

	// 设置域名
	sc.domain = utils.StringUtils.GetBaseURL(webhookURL)

	// 设置代理
	if config.Config.PROXY != "" {
		proxyURL, err := url.Parse(config.Config.PROXY)
		if err == nil {
			sc.client.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			}
		}
	}

	return sc
}

// CheckToken 检查token
func (sc *SynologyChat) CheckToken(token string) bool {
	return token == sc.token
}

// GetState 获取状�?func (sc *SynologyChat) GetState() bool {
	if sc.webhookURL == "" || sc.token == "" {
		return false
	}
	users := sc.getBotUsers()
	return len(users) > 0
}

// SendMsg 发送消�?func (sc *SynologyChat) SendMsg(title, text, image, userid, link string) *bool {
	if title == "" && text == "" {
		utils.Log.Error("标题和内容不能同时为�?)
		result := false
		return &result
	}
	
	if sc.webhookURL == "" || sc.token == "" {
		result := false
		return &result
	}
	
	defer func() {
		if r := recover(); r != nil {
			utils.Log.Errorf("SynologyChat发送消息错误：%v", r)
		}
	}()
	
	// 拼装消息内容
	titles := strings.Split(title, "\n")
	if len(titles) > 1 {
		title = titles[0]
		if text == "" {
			text = strings.Join(titles[1:], "\n")
		} else {
			text = fmt.Sprintf("%s\n%s", strings.Join(titles[1:], "\n"), text)
		}
	}
	
	var caption string
	if text != "" {
		caption = fmt.Sprintf("*%s*\n%s", title, strings.ReplaceAll(text, "\n\n", "\n"))
	} else {
		caption = title
	}
	
	if link != "" {
		caption = fmt.Sprintf("%s\n[查看详情](%s)", caption, link)
	}
	
	// 获取用户ID列表
	var userids []int
	if userid != "" {
		id, err := strconv.Atoi(userid)
		if err == nil {
			userids = append(userids, id)
		}
	} else {
		userids = sc.getBotUsers()
		if len(userids) == 0 {
			utils.Log.Error("SynologyChat机器人没有对任何用户可见")
			result := false
			return &result
		}
	}
	
	payloadData := map[string]interface{}{
		"text": strings.ReplaceAll(url.QueryEscape(caption), "+", "%20"),
	}
	
	if image != "" {
		payloadData["file_url"] = strings.ReplaceAll(url.QueryEscape(image), "+", "%20")
	}
	
	payloadData["user_ids"] = userids
	
	return sc.sendRequest(payloadData)
}

// SendMediasMsg 发送媒体列表消�?func (sc *SynologyChat) SendMediasMsg(medias []map[string]interface{}, userid, title string) *bool {
	if len(medias) == 0 {
		result := false
		return &result
	}
	
	if sc.webhookURL == "" || sc.token == "" {
		result := false
		return &result
	}
	
	defer func() {
		if r := recover(); r != nil {
			utils.Log.Errorf("SynologyChat发送消息错误：%v", r)
		}
	}()
	
	if title == "" {
		result := false
		return &result
	}
	
	index, image, caption := 1, "", fmt.Sprintf("*%s*", title)
	for _, media := range medias {
		if image == "" {
			if img, ok := media["image"].(string); ok {
				image = img
			}
		}
		
		voteAverage := 0.0
		if vote, ok := media["vote_average"].(float64); ok {
			voteAverage = vote
		}
		
		titleYear := ""
		if ty, ok := media["title_year"].(string); ok {
			titleYear = ty
		}
		
		detailLink := ""
		if link, ok := media["detail_link"].(string); ok {
			detailLink = link
		}
		
		mediaType := ""
		if mType, ok := media["type"].(string); ok {
			mediaType = mType
		}
		
		if voteAverage > 0 {
			caption = fmt.Sprintf("%s\n%d. <%s|%s>\n_类型�?s，评分：%.1f_", 
				caption, index, detailLink, titleYear, mediaType, voteAverage)
		} else {
			caption = fmt.Sprintf("%s\n%d. <%s|%s>\n_类型�?s_", 
				caption, index, detailLink, titleYear, mediaType)
		}
		index++
	}
	
	// 获取用户ID列表
	var userids []int
	if userid != "" {
		id, err := strconv.Atoi(userid)
		if err == nil {
			userids = append(userids, id)
		}
	} else {
		userids = sc.getBotUsers()
	}
	
	payloadData := map[string]interface{}{
		"text": strings.ReplaceAll(url.QueryEscape(caption), "+", "%20"),
		"user_ids": userids,
	}
	
	return sc.sendRequest(payloadData)
}

// SendTorrentsMsg 发送种子列表消�?func (sc *SynologyChat) SendTorrentsMsg(torrents []map[string]interface{}, userid, title, link string) *bool {
	if sc.webhookURL == "" || sc.token == "" {
		return nil
	}
	
	if len(torrents) == 0 {
		result := false
		return &result
	}
	
	defer func() {
		if r := recover(); r != nil {
			utils.Log.Errorf("SynologyChat发送消息错误：%v", r)
		}
	}()
	
	index, caption := 1, fmt.Sprintf("*%s*", title)
	for _, context := range torrents {
		torrent, ok := context["torrent_info"].(map[string]interface{})
		if !ok {
			continue
		}
		
		siteName := ""
		if name, nameOk := torrent["site_name"].(string); nameOk {
			siteName = name
		}
		
		pageURL := ""
		if url, urlOk := torrent["page_url"].(string); urlOk {
			pageURL = url
		}
		
		titleStr := ""
		if t, tOk := torrent["title"].(string); tOk {
			titleStr = t
		}
		
		description := ""
		if desc, descOk := torrent["description"].(string); descOk {
			description = desc
		}
		
		size := int64(0)
		if s, sOk := torrent["size"].(int64); sOk {
			size = s
		}
		
		volumeFactor := ""
		if vf, vfOk := torrent["volume_factor"].(string); vfOk {
			volumeFactor = vf
		}
		
		seeders := 0
		if sd, sdOk := torrent["seeders"].(int); sdOk {
			seeders = sd
		}
		
		meta := utils.MetaInfo.NewMetaInfo(titleStr, description)
		seasonEpisode := meta.SeasonEpisode
		resourceTerm := meta.ResourceTerm
		videoTerm := meta.VideoTerm
		releaseGroup := meta.ReleaseGroup
		
		torrentTitle := fmt.Sprintf("%s %s %s %s", seasonEpisode, resourceTerm, videoTerm, releaseGroup)
		torrentTitle = strings.Join(strings.Fields(torrentTitle), " ")
		
		seederStr := fmt.Sprintf("%d�?, seeders)
		
		caption = fmt.Sprintf("%s\n%d.�?s�?%s|%s> %s %s %s\n_%s_", 
			caption, index, siteName, pageURL, torrentTitle, 
			utils.StringUtils.StrFileSize(size), volumeFactor, seederStr, description)
		index++
	}
	
	if link != "" {
		caption = fmt.Sprintf("%s\n[查看详情](%s)", caption, link)
	}
	
	// 获取用户ID列表
	var userids []int
	if userid != "" {
		id, err := strconv.Atoi(userid)
		if err == nil {
			userids = append(userids, id)
		}
	} else {
		userids = sc.getBotUsers()
	}
	
	payloadData := map[string]interface{}{
		"text": strings.ReplaceAll(url.QueryEscape(caption), "+", "%20"),
		"user_ids": userids,
	}
	
	return sc.sendRequest(payloadData)
}

// getBotUsers 查询机器人可见的用户列表
func (sc *SynologyChat) getBotUsers() []int {
	if sc.domain == "" || sc.token == "" {
		return []int{}
	}
	
	reqURL := fmt.Sprintf("%s/webapi/entry.cgi?api=SYNO.Chat.External&method=user_list&version=2&token=%s",
		sc.domain, sc.token)
	
	resp, err := sc.client.Get(reqURL)
	if err != nil {
		utils.Log.Errorf("获取SynologyChat用户列表失败�?v", err)
		return []int{}
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		utils.Log.Errorf("获取SynologyChat用户列表失败，状态码�?d", resp.StatusCode)
		return []int{}
	}
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		utils.Log.Errorf("解析SynologyChat用户列表响应失败�?v", err)
		return []int{}
	}
	
	data, dataOk := result["data"].(map[string]interface{})
	if !dataOk {
		utils.Log.Error("SynologyChat用户列表响应格式错误")
		return []int{}
	}
	
	users, usersOk := data["users"].([]interface{})
	if !usersOk {
		utils.Log.Error("SynologyChat用户列表响应格式错误")
		return []int{}
	}
	
	var userids []int
	for _, userItem := range users {
		if user, userOk := userItem.(map[string]interface{}); userOk {
			deleted, deletedOk := user["deleted"].(bool)
			if deletedOk && !deleted {
				if userID, idOk := user["user_id"].(float64); idOk {
					userids = append(userids, int(userID))
				}
			}
		}
	}
	
	return userids
}

// sendRequest 发送消息请�?func (sc *SynologyChat) sendRequest(payloadData map[string]interface{}) *bool {
	payloadBytes, err := json.Marshal(payloadData)
	if err != nil {
		utils.Log.Errorf("序列化SynologyChat消息失败�?v", err)
		result := false
		return &result
	}
	
	payload := fmt.Sprintf("payload=%s", url.QueryEscape(string(payloadBytes)))
	
	resp, err := sc.client.Post(sc.webhookURL, "application/x-www-form-urlencoded", 
		bytes.NewBufferString(payload))
	if err != nil {
		utils.Log.Errorf("发送SynologyChat消息请求失败�?v", err)
		result := false
		return &result
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		utils.Log.Errorf("SynologyChat请求失败，状态码�?d，原因：%s", resp.StatusCode, resp.Status)
		result := false
		return &result
	}
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		utils.Log.Errorf("解析SynologyChat响应失败�?v", err)
		resultVal := false
		return &resultVal
	}
	
	if result != nil {
		if errorData, errorOk := result["error"].(map[string]interface{}); errorOk {
			errno, errnoOk := errorData["code"].(float64)
			errmsg, errmsgOk := errorData["errors"].(string)
			if errnoOk && errno != 0 {
				utils.Log.Errorf("SynologyChat返回错误�?v-%s", errno, errmsg)
				resultVal := false
				return &resultVal
			}
		}
		resultVal := true
		return &resultVal
	} else {
		utils.Log.Error("SynologyChat返回空响�?)
		resultVal := false
		return &resultVal
	}
}
