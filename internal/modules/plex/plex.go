package plex

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"moviepilot-go/internal/utils"
	"moviepilot-go/pkg/models"
)

// Plex Plex媒体服务器连接和操作�?type Plex struct {
	Host          string
	Token         string
	PlayHost      string
	SyncLibraries []string
	client        *http.Client
}

// NewPlex 创建Plex实例
func NewPlex(host, token, playHost string, syncLibraries []string) *Plex {
	if host == "" || token == "" {
		utils.Log.Error("Plex服务器配置不完整�?)
		return nil
	}

	// 标准化URL
	host = utils.StandardizeBaseURL(host)
	playHost = utils.StandardizeBaseURL(playHost)

	return &Plex{
		Host:          host,
		Token:         token,
		PlayHost:      playHost,
		SyncLibraries: syncLibraries,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// IsInactive 判断是否需要重�?func (p *Plex) IsInactive() bool {
	if p.Host == "" || p.Token == "" {
		return false
	}
	// 简单检查，实际应该检查连接状�?	return false
}

// Reconnect 重连
func (p *Plex) Reconnect() {
	utils.Log.Info("重新连接Plex服务�?..")
	// 重新连接逻辑，这里简化处�?}

// Authenticate 用户认证
func (p *Plex) Authenticate(username, password string) (*string, *string) {
	if username == "" || password == "" {
		return nil, nil
	}

	// Plex认证逻辑，这里简化处�?	// 实际应该调用Plex API进行认证
	utils.Log.Debugf("尝试认证用户: %s", username)
	return &p.Token, &username
}

// GetLibrarys 获取媒体服务器所有媒体库列表
func (p *Plex) GetLibrarys(hidden bool) []models.MediaServerLibrary {
	if p.Host == "" {
		return []models.MediaServerLibrary{}
	}

	// 这里应该调用Plex API获取媒体库列�?	// 简化实现，返回空列�?	utils.Log.Debug("获取Plex媒体库列�?)
	return []models.MediaServerLibrary{}
}

// GetMediasCount 获得电影、电视剧、动漫媒体数�?func (p *Plex) GetMediasCount() *models.Statistic {
	if p.Host == "" {
		return &models.Statistic{}
	}

	// 这里应该调用Plex API获取媒体统计信息
	// 简化实现，返回空统�?	utils.Log.Debug("获取Plex媒体统计信息")
	return &models.Statistic{}
}

// GetMovies 根据标题和年份，检查电影是否在Plex中存�?func (p *Plex) GetMovies(title, originalTitle, year string, tmdbID *int) []models.MediaServerItem {
	if p.Host == "" {
		return []models.MediaServerItem{}
	}

	// 这里应该调用Plex API搜索电影
	// 简化实现，返回空列�?	utils.Log.Debugf("在Plex中搜索电�? %s (%s)", title, year)
	return []models.MediaServerItem{}
}

// GetTVEpisodes 根据标题、年份、季查询电视剧所有集信息
func (p *Plex) GetTVEpisodes(itemID, title, originalTitle, year string, season *int) (string, map[int][]int) {
	if p.Host == "" {
		return "", map[int][]int{}
	}

	// 这里应该调用Plex API搜索电视剧集信息
	// 简化实现，返回空结�?	utils.Log.Debugf("在Plex中搜索电视剧: %s (%s)", title, year)
	return "", map[int][]int{}
}

// GetRemoteImageByID 根据ItemId从Plex查询图片地址
func (p *Plex) GetRemoteImageByID(itemID, imageType string, depth int, plexURL bool) *string {
	if p.Host == "" || itemID == "" {
		return nil
	}

	// 这里应该调用Plex API获取图片URL
	// 简化实现，返回nil
	utils.Log.Debugf("获取Plex项目图片: %s", itemID)
	return nil
}

// GetItemInfo 获取单个项目详情
func (p *Plex) GetItemInfo(itemID string) *models.MediaServerItem {
	if p.Host == "" || itemID == "" {
		return nil
	}

	// 这里应该调用Plex API获取项目详情
	// 简化实现，返回nil
	utils.Log.Debugf("获取Plex项目详情: %s", itemID)
	return nil
}

// GetItems 获取媒体服务器项目列�?func (p *Plex) GetItems(parent string, startIndex, limit int) []models.MediaServerItem {
	if p.Host == "" || parent == "" {
		return []models.MediaServerItem{}
	}

	// 这里应该调用Plex API获取项目列表
	// 简化实现，返回空列�?	utils.Log.Debugf("获取Plex项目列表: %s", parent)
	return []models.MediaServerItem{}
}

// GetWebhookMessage 解析Plex报文
func (p *Plex) GetWebhookMessage(form url.Values) *models.WebhookEventInfo {
	if len(form) == 0 {
		return nil
	}

	payload := form.Get("payload")
	if payload == "" {
		return nil
	}

	var message map[string]interface{}
	err := json.Unmarshal([]byte(payload), &message)
	if err != nil {
		utils.Log.Debugf("解析Plex webhook出错�?v", err)
		return nil
	}

	eventType, ok := message["event"].(string)
	if !ok || eventType == "" {
		return nil
	}

	utils.Log.Debugf("接收到Plex webhook�?v", message)
	eventItem := &models.WebhookEventInfo{
		Event:   &eventType,
		Channel: utils.StringPtr("plex"),
	}

	// 解析Metadata
	if metadata, ok := message["Metadata"].(map[string]interface{}); ok {
		if mediaType, ok := metadata["type"].(string); ok && mediaType == "episode" {
			eventItem.ItemType = utils.StringPtr("TV")
			
			grandparentTitle := ""
			if val, ok := metadata["grandparentTitle"].(string); ok {
				grandparentTitle = val
			}
			
			parentIndex := 0
			if val, ok := metadata["parentIndex"].(float64); ok {
				parentIndex = int(val)
			}
			
			index := 0
			if val, ok := metadata["index"].(float64); ok {
				index = int(val)
			}
			
			title := ""
			if val, ok := metadata["title"].(string); ok {
				title = val
			}
			
			itemName := fmt.Sprintf("%s S%dE%d %s", grandparentTitle, parentIndex, index, title)
			eventItem.ItemName = &itemName
			
			if key, ok := metadata["key"].(string); ok {
				eventItem.ItemID = &key
			}
			
			eventItem.SeasonID = utils.StringPtr(strconv.Itoa(parentIndex))
			eventItem.EpisodeID = utils.StringPtr(strconv.Itoa(index))
			
			if summary, ok := metadata["summary"].(string); ok {
				var overview string
				if len(summary) > 100 {
					overview = summary[:100] + "..."
				} else {
					overview = summary
				}
				eventItem.Overview = &overview
			}
		} else {
			itemType := "MOV"
			if mediaType == "show" {
				itemType = "SHOW"
			}
			eventItem.ItemType = &itemType
			
			title := ""
			if val, ok := metadata["title"].(string); ok {
				title = val
			}
			
			year := ""
			if val, ok := metadata["year"].(float64); ok {
				year = fmt.Sprintf("(%d)", int(val))
			}
			
			itemName := fmt.Sprintf("%s %s", title, year)
			eventItem.ItemName = &itemName
			
			if key, ok := metadata["key"].(string); ok {
				eventItem.ItemID = &key
			}
			
			if summary, ok := metadata["summary"].(string); ok {
				var overview string
				if len(summary) > 100 {
					overview = summary[:100] + "..."
				} else {
					overview = summary
				}
				eventItem.Overview = &overview
			}
		}
	}

	// 解析Player信息
	if player, ok := message["Player"].(map[string]interface{}); ok {
		if publicAddress, ok := player["publicAddress"].(string); ok {
			eventItem.IP = &publicAddress
		}
		
		if title, ok := player["title"].(string); ok {
			eventItem.Client = &title
		}
		
		// 这里给个�?防止拼消息的时候出现None
		eventItem.DeviceName = utils.StringPtr(" ")
	}

	// 解析Account信息
	if account, ok := message["Account"].(map[string]interface{}); ok {
		if title, ok := account["title"].(string); ok {
			eventItem.UserName = &title
		}
	}

	// 获取消息图片
	if eventItem.ItemID != nil {
		// 根据返回的item_id去调用媒体服务器获取
		imageURL := p.GetRemoteImageByID(*eventItem.ItemID, "Backdrop", 0, false)
		eventItem.ImageURL = imageURL
	}

	// 转换message为json_object
	jsonBytes, _ := json.Marshal(message)
	jsonObject := make(map[string]interface{})
	json.Unmarshal(jsonBytes, &jsonObject)
	eventItem.JSONObject = jsonObject

	return eventItem
}

// GetPlayURL 拼装媒体播放链接
func (p *Plex) GetPlayURL(itemID string) string {
	if p.Host == "" || itemID == "" {
		return ""
	}
	
	// 这里应该获取Plex服务器的machineIdentifier
	machineID := "machine_id_placeholder"
	
	playHost := p.PlayHost
	if playHost == "" {
		playHost = p.Host
	}
	
	return fmt.Sprintf("%sweb/index.html#!/server/%s/details?key=%s&X-Plex-Token=%s", 
		playHost, machineID, itemID, p.Token)
}

// GetResume 获取继续观看的媒�?func (p *Plex) GetResume(num int) []models.MediaServerPlayItem {
	if p.Host == "" {
		return []models.MediaServerPlayItem{}
	}

	// 这里应该调用Plex API获取继续观看列表
	// 简化实现，返回空列�?	utils.Log.Debug("获取Plex继续观看列表")
	return []models.MediaServerPlayItem{}
}

// GetLatest 获取最近添加媒�?func (p *Plex) GetLatest(num int) []models.MediaServerPlayItem {
	if p.Host == "" {
		return []models.MediaServerPlayItem{}
	}

	// 这里应该调用Plex API获取最近添加列�?	// 简化实现，返回空列�?	utils.Log.Debug("获取Plex最近添加列�?)
	return []models.MediaServerPlayItem{}
}

// Close 关闭连接
func (p *Plex) Close() {
	// 关闭连接逻辑
	utils.Log.Debug("关闭Plex连接")
}
