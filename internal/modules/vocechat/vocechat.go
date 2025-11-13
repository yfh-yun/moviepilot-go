package vocechat

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"moviepilot-go/internal/utils"
	"moviepilot-go/pkg/models"
)

// VoceChat VoceChat客户�?type VoceChat struct {
	// host
	host string
	// apikey
	apikey string
	// 频道ID
	channelID string
	
	// 请求对象
	client *http.Client
}

// NewVoceChat 创建新的VoceChat实例
func NewVoceChat(host, apikey, channelID string) *VoceChat {
	if host == "" || apikey == "" || channelID == "" {
		// 记录错误日志
		fmt.Println("VoceChat配置不完整！")
		return nil
	}
	
	v := &VoceChat{
		host:      host,
		apikey:    apikey,
		channelID: channelID,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	
	// 确保host格式正确
	if !strings.HasSuffix(v.host, "/") {
		v.host += "/"
	}
	if !strings.HasPrefix(v.host, "http") {
		v.host = "http://" + v.host
	}
	
	return v
}

// GetState 获取状�?func (v *VoceChat) GetState() bool {
	return v.getGroups() != nil
}

// getGroups 获取频道列表
func (v *VoceChat) getGroups() interface{} {
	if v.client == nil {
		return nil
	}
	
	req, err := http.NewRequest("GET", v.host+"api/bot", nil)
	if err != nil {
		return nil
	}
	
	req.Header.Set("content-type", "text/markdown")
	req.Header.Set("x-api-key", v.apikey)
	req.Header.Set("accept", "application/json; charset=utf-8")
	
	resp, err := v.client.Do(req)
	if err != nil || resp == nil {
		return nil
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == 200 {
		// 简化实现，实际应该解析返回的JSON
		return make(map[string]interface{})
	}
	
	return nil
}

// SendMsg 微信消息发送入口，支持文本、图片、链接跳转、指定发送对�?func (v *VoceChat) SendMsg(title, text, userID, link string) bool {
	if v.client == nil {
		return false
	}

	if title == "" && text == "" {
		fmt.Println("标题和内容不能同时为�?)
		return false
	}

	var caption string
	if text != "" {
		caption = fmt.Sprintf("**%s**\n%s", title, text)
	} else {
		caption = fmt.Sprintf("**%s**", title)
	}

	if link != "" {
		caption = fmt.Sprintf("%s\n[查看详情](%s)", caption, link)
	}

	var chatID string
	if userID != "" {
		chatID = userID
	} else {
		chatID = fmt.Sprintf("GID#%s", v.channelID)
	}

	return v.sendRequest(chatID, caption)
}

// SendMediasMsg 发送列表类消息
func (v *VoceChat) SendMediasMsg(medias []models.MediaInfo, userID, title, link string) bool {
	if v.client == nil {
		return false
	}

	index, caption := 1, fmt.Sprintf("**%s**", title)
	for _, media := range medias {
		if media.VoteAverage > 0 {
			caption = fmt.Sprintf("%s\n%d. [%s](%s)\n_类型�?s，评分：%.1f_", caption, index,
				media.TitleYear, media.DetailLink, media.Type, media.VoteAverage)
		} else {
			caption = fmt.Sprintf("%s\n%d. [%s](%s)\n_类型�?s_", caption, index,
				media.TitleYear, media.DetailLink, media.Type)
		}
		index++
	}

	if link != "" {
		caption = fmt.Sprintf("%s\n[查看详情](%s)", caption, link)
	}

	var chatID string
	if userID != "" {
		chatID = userID
	} else {
		chatID = fmt.Sprintf("GID#%s", v.channelID)
	}

	return v.sendRequest(chatID, caption)
}

// SendTorrentsMsg 发送列表消�?func (v *VoceChat) SendTorrentsMsg(torrents []models.Context, userID, title, link string) bool {
	if v.client == nil {
		return false
	}

	if len(torrents) == 0 {
		return false
	}

	index, caption := 1, fmt.Sprintf("**%s**", title)
	for _, context := range torrents {
		torrent := context.TorrentInfo
		// 注意：这里需要根据实际情况构造meta信息
		metaInfo := "" // 简化处理，实际应该解析种子信息
		
		torrentTitle := fmt.Sprintf("%s %s %s %s", metaInfo, 
			torrent.ResourceTerm, torrent.VideoTerm, torrent.ReleaseGroup)
		torrentTitle = strings.Join(strings.Fields(torrentTitle), " ")
		
		free := torrent.VolumeFactor
		seeder := fmt.Sprintf("%d�?, torrent.Seeders)
		
		caption = fmt.Sprintf("%s\n%d.�?s】[%s](%s) %s %s %s", caption, index, 
			torrent.SiteName, torrentTitle, torrent.PageUrl, 
			utils.StrFileSize(torrent.Size), free, seeder)
		index++
	}

	if link != "" {
		caption = fmt.Sprintf("%s\n[查看详情](%s)", caption, link)
	}

	var chatID string
	if userID != "" {
		chatID = userID
	} else {
		chatID = fmt.Sprintf("GID#%s", v.channelID)
	}

	return v.sendRequest(chatID, caption)
}

// sendRequest 向VoceChat发送报�?func (v *VoceChat) sendRequest(userid, caption string) bool {
	var action, idstr string
	
	if strings.HasPrefix(userid, "GID#") {
		action = "send_to_group"
		idstr = userid[4:] // 去掉"GID#"前缀
	} else {
		action = "send_to_user"
		idstr = userid[4:] // 去掉"UID#"前缀，如果有的话
		if idstr == userid {
			// 如果没有前缀，则使用原始ID
			idstr = userid
		}
	}

	url := fmt.Sprintf("%sapi/bot/%s/%s", v.host, action, idstr)
	
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(caption))
	if err != nil {
		fmt.Printf("创建请求失败: %v\n", err)
		return false
	}
	
	req.Header.Set("content-type", "text/markdown")
	req.Header.Set("x-api-key", v.apikey)
	req.Header.Set("accept", "application/json; charset=utf-8")
	
	resp, err := v.client.Do(req)
	if err != nil {
		fmt.Printf("VoceChat发送消息失败，连接失败: %v\n", err)
		return false
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == 200 {
		return true
	}
	
	fmt.Printf("VoceChat发送消息失败，错误码：%d\n", resp.StatusCode)
	return false
}
