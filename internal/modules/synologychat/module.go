package synologychat

import (
	"fmt"
	"net/url"
	"strconv"
	
	"moviepilot-go/internal/modules"
	"moviepilot-go/internal/utils"
	"moviepilot-go/pkg/models"
)

// SynologyChatModule Synology Chat模块
type SynologyChatModule struct {
	modules.ModuleBase
	modules.MessageBase[*SynologyChat]
	channel models.MessageChannel
}

// NewSynologyChatModule 创建SynologyChat模块实例
func NewSynologyChatModule() *SynologyChatModule {
	scm := &SynologyChatModule{
		channel: models.MessageChannelSynologyChat,
	}
	
	// 初始化模块服�?	scm.InitService("synologychat", &SynologyChat{})
	
	return scm
}

// GetName 获取模块名称
func (scm *SynologyChatModule) GetName() string {
	return "Synology Chat"
}

// GetType 获取模块类型
func (scm *SynologyChatModule) GetType() models.ModuleType {
	return models.ModuleTypeNotification
}

// GetSubtype 获取模块子类�?func (scm *SynologyChatModule) GetSubtype() models.MessageChannel {
	return models.MessageChannelSynologyChat
}

// GetPriority 获取模块优先级，数字越小优先级越高，只有同一接口下优先级才生�?func (scm *SynologyChatModule) GetPriority() int {
	return 5
}

// Stop 停止模块
func (scm *SynologyChatModule) Stop() {
	// Synology Chat模块不需要特殊停止处�?}

// Test 测试模块连接�?func (scm *SynologyChatModule) Test() (*bool, string) {
	instances := scm.GetInstances()
	if len(instances) == 0 {
		return nil, ""
	}
	
	for name, client := range instances {
		state := client.GetState()
		if !state {
			return utils.BoolPtr(false), fmt.Sprintf("Synology Chat %s 未就�?, name)
		}
	}
	
	return utils.BoolPtr(true), ""
}

// InitSetting 初始化设�?func (scm *SynologyChatModule) InitSetting() (string, interface{}) {
	// 根据实际需要实�?	return "", nil
}

// MessageParser 解析消息内容，返回字�?func (scm *SynologyChatModule) MessageParser(source string, body, form, args interface{}) *models.CommingMessage {
	defer func() {
		if r := recover(); r != nil {
			utils.Log.Debugf("解析SynologyChat消息失败�?v", r)
		}
	}()
	
	// 获取服务配置
	clientConfig := scm.GetConfig(source)
	if clientConfig == nil {
		return nil
	}
	
	client := scm.GetInstance(clientConfig.Name)
	if client == nil {
		return nil
	}
	
	// 解析消息
	formMap, ok := form.(map[string]interface{})
	if !ok || formMap == nil {
		return nil
	}
	
	// 校验token
	token, tokenOk := formMap["token"].(string)
	if !tokenOk || !client.CheckToken(token) {
		return nil
	}
	
	// 文本
	text, textOk := formMap["text"].(string)
	
	// 用户ID
	userIDStr, userIDOk := formMap["user_id"].(string)
	var userID int64
	if userIDOk {
		if id, err := strconv.ParseInt(userIDStr, 10, 64); err == nil {
			userID = id
		}
	}
	
	// 获取用户�?	username, usernameOk := formMap["username"].(string)
	
	if textOk && text != "" && userID != 0 {
		utils.Log.Infof("收到来自 %s 的SynologyChat消息：userid=%d, username=%s, text=%s",
			clientConfig.Name, userID, username, text)
		
		return &models.CommingMessage{
			Channel:  models.MessageChannelSynologyChat,
			Source:   clientConfig.Name,
			UserID:   userID,
			Username: username,
			Text:     text,
		}
	}
	
	return nil
}

// PostMessage 发送消�?func (scm *SynologyChatModule) PostMessage(message *models.Notification) {
	configs := scm.GetConfigs()
	for _, conf := range configs {
		if !scm.CheckMessage(message, conf.Name) {
			continue
		}
		
		userid := message.UserID
		if userid == "" && message.Targets != nil {
			if uid, exists := (*message.Targets)["synologychat_userid"]; exists {
				userid = uid.(string)
			}
			if userid == "" {
				utils.Log.Warn("用户没有指定 SynologyChat用户ID，消息无法发�?)
				return
			}
		}
		
		client := scm.GetInstance(conf.Name)
		if client != nil {
			client.SendMsg(
				message.Title,
				message.Text,
				message.Image,
				userid,
				message.Link,
			)
		}
	}
}

// PostMediasMessage 发送媒体信息选择列表
func (scm *SynologyChatModule) PostMediasMessage(message *models.Notification, medias []*models.MediaInfo) {
	configs := scm.GetConfigs()
	for _, conf := range configs {
		if !scm.CheckMessage(message, conf.Name) {
			continue
		}
		
		client := scm.GetInstance(conf.Name)
		if client != nil {
			// 构造媒体信息列�?			mediaList := make([]map[string]interface{}, 0)
			for _, media := range medias {
				mediaItem := map[string]interface{}{
					"image":        media.GetMessageImage(),
					"vote_average": media.VoteAverage,
					"title_year":   media.TitleYear,
					"detail_link":  media.DetailLink,
					"type":         media.Type,
				}
				mediaList = append(mediaList, mediaItem)
			}
			
			client.SendMediasMsg(mediaList, message.UserID, message.Title)
		}
	}
}

// PostTorrentsMessage 发送种子信息选择列表
func (scm *SynologyChatModule) PostTorrentsMessage(message *models.Notification, torrents []*models.Context) {
	configs := scm.GetConfigs()
	for _, conf := range configs {
		if !scm.CheckMessage(message, conf.Name) {
			continue
		}
		
		client := scm.GetInstance(conf.Name)
		if client != nil {
			// 构造种子信息列�?			torrentList := make([]map[string]interface{}, 0)
			for _, context := range torrents {
				torrentItem := map[string]interface{}{
					"torrent_info": context.TorrentInfo,
				}
				torrentList = append(torrentList, torrentItem)
			}
			
			client.SendTorrentsMsg(torrentList, message.UserID, message.Title, message.Link)
		}
	}
}
