package slack

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	
	"moviepilot-go/internal/modules"
	"moviepilot-go/internal/utils"
	"moviepilot-go/pkg/models"
)

// SlackModule Slack模块
type SlackModule struct {
	modules.ModuleBase
	modules.MessageBase[*Slack]
	channel models.MessageChannel
}

// NewSlackModule 创建Slack模块实例
func NewSlackModule() *SlackModule {
	sm := &SlackModule{
		channel: models.MessageChannelSlack,
	}
	
	// 初始化模块服�?	sm.InitService("slack", &Slack{})
	
	return sm
}

// GetName 获取模块名称
func (sm *SlackModule) GetName() string {
	return "Slack"
}

// GetType 获取模块类型
func (sm *SlackModule) GetType() models.ModuleType {
	return models.ModuleTypeNotification
}

// GetSubtype 获取模块子类�?func (sm *SlackModule) GetSubtype() models.MessageChannel {
	return models.MessageChannelSlack
}

// GetPriority 获取模块优先级，数字越小优先级越高，只有同一接口下优先级才生�?func (sm *SlackModule) GetPriority() int {
	return 3
}

// Stop 停止模块
func (sm *SlackModule) Stop() {
	instances := sm.GetInstances()
	for _, client := range instances {
		if client != nil {
			client.Stop()
		}
	}
}

// Test 测试模块连接�?func (sm *SlackModule) Test() (*bool, string) {
	instances := sm.GetInstances()
	if len(instances) == 0 {
		return nil, ""
	}
	
	for name, client := range instances {
		state := client.GetState()
		if !state {
			return utils.BoolPtr(false), fmt.Sprintf("Slack %s 未就�?, name)
		}
	}
	
	return utils.BoolPtr(true), ""
}

// InitSetting 初始化设�?func (sm *SlackModule) InitSetting() (string, interface{}) {
	return "", nil
}

// MessageParser 解析消息内容，返回字�?func (sm *SlackModule) MessageParser(source string, body, form, args interface{}) *models.CommingMessage {
	// 获取服务配置
	clientConfig := sm.GetConfig(source)
	if clientConfig == nil {
		return nil
	}
	
	bodyStr, ok := body.(string)
	if !ok {
		utils.Log.Debug("无法解析Slack消息�?)
		return nil
	}
	
	var msgJSON map[string]interface{}
	if err := json.Unmarshal([]byte(bodyStr), &msgJSON); err != nil {
		utils.Log.Debugf("解析Slack消息失败�?v", err)
		return nil
	}
	
	if msgJSON != nil {
		msgType, typeExists := msgJSON["type"].(string)
		
		var userid, text, username string
		var isCallback bool
		var callbackData, messageID, chatID *string
		
		if typeExists && msgType == "message" {
			if uid, uidExists := msgJSON["user"].(string); uidExists {
				userid = uid
			}
			if txt, txtExists := msgJSON["text"].(string); txtExists {
				text = txt
			}
			username = userid
		} else if typeExists && msgType == "block_actions" {
			if user, userExists := msgJSON["user"].(map[string]interface{}); userExists {
				if uid, uidExists := user["id"].(string); uidExists {
					userid = uid
				}
				if uname, unameExists := user["name"].(string); unameExists {
					username = uname
				}
			}
			
			if actions, actionsExists := msgJSON["actions"].([]interface{}); actionsExists && len(actions) > 0 {
				if action, actionOk := actions[0].(map[string]interface{}); actionOk {
					if value, valueExists := action["value"].(string); valueExists {
						callbackDataVal := value
						callbackData = &callbackDataVal
						// 使用CALLBACK前缀标识按钮回调
						text = fmt.Sprintf("CALLBACK:%s", callbackDataVal)
						isCallback = true
					}
				}
			}
			
			// 获取原消息信息用于编�?			if message, messageExists := msgJSON["message"].(map[string]interface{}); messageExists {
				// Slack消息的时间戳作为消息ID
				if ts, tsExists := message["ts"].(string); tsExists {
					messageID = &ts
				}
			}
			
			if channel, channelExists := msgJSON["channel"].(map[string]interface{}); channelExists {
				if cid, cidExists := channel["id"].(string); cidExists {
					chatID = &cid
				}
			} else if container, containerExists := msgJSON["container"].(map[string]interface{}); containerExists {
				if cid, cidExists := container["channel_id"].(string); cidExists {
					chatID = &cid
				}
			}
			
			utils.Log.Infof("收到来自 %s 的Slack按钮回调：userid=%s, username=%s, callback_data=%s",
				clientConfig.Name, userid, username, *callbackData)
			
			// 创建包含回调信息的CommingMessage
			return &models.CommingMessage{
				Channel:      models.MessageChannelSlack,
				Source:       clientConfig.Name,
				UserID:       userid,
				Username:     username,
				Text:         text,
				IsCallback:   isCallback,
				CallbackData: callbackData,
				MessageID:    messageID,
				ChatID:       chatID,
			}
		} else if typeExists && msgType == "event_callback" {
			if event, eventExists := msgJSON["event"].(map[string]interface{}); eventExists {
				if uid, uidExists := event["user"].(string); uidExists {
					userid = uid
				}
				
				if txt, txtExists := event["text"].(string); txtExists {
					// 移除用户提及
					re := regexp.MustCompile(`<@[0-9A-Z]+>`)
					text = strings.TrimSpace(re.ReplaceAllString(txt, ""))
				}
				username = ""
			}
		} else if typeExists && msgType == "shortcut" {
			if user, userExists := msgJSON["user"].(map[string]interface{}); userExists {
				if uid, uidExists := user["id"].(string); uidExists {
					userid = uid
				}
				if uname, unameExists := user["username"].(string); unameExists {
					username = uname
				}
			}
			
			if callbackID, callbackIDExists := msgJSON["callback_id"].(string); callbackIDExists {
				text = callbackID
			}
		} else if command, commandExists := msgJSON["command"].(string); commandExists {
			if uid, uidExists := msgJSON["user_id"].(string); uidExists {
				userid = uid
			}
			if uname, unameExists := msgJSON["user_name"].(string); unameExists {
				username = uname
			}
			text = command
		} else {
			return nil
		}
		
		utils.Log.Infof("收到来自 %s 的Slack消息：userid=%s, username=%s, text=%s",
			clientConfig.Name, userid, username, text)
		
		return &models.CommingMessage{
			Channel:  models.MessageChannelSlack,
			Source:   clientConfig.Name,
			UserID:   userid,
			Username: username,
			Text:     text,
		}
	}
	
	return nil
}

// PostMessage 发送消�?func (sm *SlackModule) PostMessage(message *models.Notification) {
	configs := sm.GetConfigs()
	for _, conf := range configs {
		if !sm.CheckMessage(message, conf.Name) {
			continue
		}
		
		userid := message.UserID
		if userid == "" && message.Targets != nil {
			if uid, exists := (*message.Targets)["slack_userid"]; exists {
				userid = uid.(string)
			}
			if userid == "" {
				utils.Log.Warn("用户没有指定 Slack用户ID，消息无法发�?)
				return
			}
		}
		
		client := sm.GetInstance(conf.Name)
		if client != nil {
			client.SendMsg(
				message.Title,
				message.Text,
				message.Image,
				message.Link,
				userid,
				message.Buttons,
				message.OriginalMessageID,
				message.OriginalChatID,
			)
		}
	}
}

// PostMediasMessage 发送媒体信息选择列表
func (sm *SlackModule) PostMediasMessage(message *models.Notification, medias []*models.MediaInfo) {
	configs := sm.GetConfigs()
	for _, conf := range configs {
		if !sm.CheckMessage(message, conf.Name) {
			continue
		}
		
		client := sm.GetInstance(conf.Name)
		if client != nil {
			// 构造媒体信息列�?			mediaList := make([]map[string]interface{}, 0)
			for _, media := range medias {
				mediaItem := map[string]interface{}{
					"poster_image":  media.GetPosterImage(),
					"vote_star":     media.VoteStar,
					"detail_link":   media.DetailLink,
					"title_year":    media.TitleYear,
					"type":          media.Type,
					"overview":      media.GetOverviewString(50),
				}
				mediaList = append(mediaList, mediaItem)
			}
			
			client.SendMediasMsg(
				mediaList,
				message.UserID,
				message.Title,
				message.Buttons,
				message.OriginalMessageID,
				message.OriginalChatID,
			)
		}
	}
}

// PostTorrentsMessage 发送种子信息选择列表
func (sm *SlackModule) PostTorrentsMessage(message *models.Notification, torrents []*models.Context) {
	configs := sm.GetConfigs()
	for _, conf := range configs {
		if !sm.CheckMessage(message, conf.Name) {
			continue
		}
		
		client := sm.GetInstance(conf.Name)
		if client != nil {
			// 构造种子信息列�?			torrentList := make([]map[string]interface{}, 0)
			for _, context := range torrents {
				torrentItem := map[string]interface{}{
					"torrent_info": context.TorrentInfo,
				}
				torrentList = append(torrentList, torrentItem)
			}
			
			client.SendTorrentsMsg(
				torrentList,
				message.UserID,
				message.Title,
				message.Buttons,
				message.OriginalMessageID,
				message.OriginalChatID,
			)
		}
	}
}

// DeleteMessage 删除消息
func (sm *SlackModule) DeleteMessage(channel models.MessageChannel, source string,
	messageID string, chatID *string) bool {
	
	success := false
	configs := sm.GetConfigs()
	for _, conf := range configs {
		if channel != sm.channel {
			break
		}
		if source != conf.Name {
			continue
		}
		client := sm.GetInstance(conf.Name)
		if client != nil {
			result := client.DeleteMsg(messageID, chatID)
			if result != nil && *result {
				success = true
			}
		}
	}
	return success
}
