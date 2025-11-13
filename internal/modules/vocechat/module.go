package vocechat

import (
	"encoding/json"
	"fmt"
	"strings"
	
	"moviepilot-go/pkg/models"
	"moviepilot-go/pkg/modules"
)

// VoceChatModule VoceChat模块
type VoceChatModule struct {
	*modules.ModuleBase
	*modules.MessageBase
	channel models.MessageChannel
}

// NewVoceChatModule 创建新的VoceChat模块实例
func NewVoceChatModule() *VoceChatModule {
	module := &VoceChatModule{
		ModuleBase:  modules.NewModuleBase(),
		MessageBase: modules.NewMessageBase(),
	}
	
	// 设置模块属�?	module.Name = "VoceChat"
	module.Type = models.ModuleTypeNotification
	module.SubType = models.MessageChannelVoceChat
	module.Priority = 4
	
	return module
}

// InitModule 初始化模�?func (v *VoceChatModule) InitModule() error {
	// 初始化服�?	v.InitService(strings.ToLower("vocechat"), NewVoceChat)
	v.Channel = &models.MessageChannelVoceChat
	return nil
}

// HandleConfigChanged 处理配置变更事件
func (v *VoceChatModule) HandleConfigChanged(eventData *models.ConfigChangeEventData) {
	if eventData == nil {
		return
	}
	
	// 检查是否是通知配置变更
	if eventData.Key != string(models.SystemConfigKeyNotifications) {
		return
	}
	
	fmt.Println("配置变更，重新加载VoceChat模块...")
	v.InitModule()
}

// GetName 获取模块名称
func (v *VoceChatModule) GetName() string {
	return "VoceChat"
}

// GetType 获取模块类型
func (v *VoceChatModule) GetType() models.ModuleType {
	return models.ModuleTypeNotification
}

// GetSubType 获取模块的子类型
func (v *VoceChatModule) GetSubType() models.MessageChannel {
	return models.MessageChannelVoceChat
}

// GetPriority 获取模块优先�?func (v *VoceChatModule) GetPriority() int {
	return 4
}

// Stop 停止模块
func (v *VoceChatModule) Stop() error {
	// 实现停止逻辑
	return nil
}

// Test 测试模块连接�?func (v *VoceChatModule) Test() (bool, string) {
	// 获取实例
	instances := v.GetInstances()
	if len(instances) == 0 {
		return true, "" // 没有配置实例，返回成�?	}
	
	// 检查每个实例的状�?	for name, client := range instances {
		// 类型断言获取VoceChat实例
		if vocechat, ok := client.(*VoceChat); ok {
			state := vocechat.GetState()
			if !state {
				return false, fmt.Sprintf("VoceChat %s 未就�?, name)
			}
		}
	}
	
	return true, ""
}

// InitSetting 初始化设�?func (v *VoceChatModule) InitSetting() (string, interface{}) {
	// 实现初始化设置逻辑
	return "", nil
}

// MessageParser 解析消息内容
func (v *VoceChatModule) MessageParser(source string, body []byte, form map[string]interface{}, 
	args map[string]interface{}) *models.CommingMessage {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("VoceChat消息处理发生错误: %v\n", r)
		}
	}()
	
	// 获取服务配置
	clientConfig := v.GetConfig(&source)
	if clientConfig == nil {
		return nil
	}
	
	// 报文�?	var msgBody map[string]interface{}
	err := json.Unmarshal(body, &msgBody)
	if err != nil {
		fmt.Printf("VoceChat消息解析失败: %v\n", err)
		return nil
	}
	
	// 类型
	msgType := ""
	if detail, ok := msgBody["detail"].(map[string]interface{}); ok {
		if t, ok := detail["type"].(string); ok {
			msgType = t
		}
	}
	
	if msgType != "normal" {
		// 非新消息
		return nil
	}
	
	fmt.Printf("收到VoceChat请求�?v\n", msgBody)
	
	// 文本内容
	var content string
	if detail, ok := msgBody["detail"].(map[string]interface{}); ok {
		if c, ok := detail["content"].(string); ok {
			content = c
		}
	}
	
	// 用户ID
	var userid string
	if target, ok := msgBody["target"].(map[string]interface{}); ok {
		if gid, ok := target["gid"].(float64); ok {
			// 获取配置中的频道ID
			configMap := clientConfig.(map[string]interface{})
			if channelID, ok := configMap["VOCECHAT_CHANNEL_ID"].(string); ok {
				if fmt.Sprintf("%.0f", gid) == channelID {
					// 来自监听频道的消�?					userid = fmt.Sprintf("GID#%.0f", gid)
				}
			}
		}
	}
	
	// 如果不是来自频道，则来自个人
	if userid == "" {
		if fromUID, ok := msgBody["from_uid"].(float64); ok {
			userid = fmt.Sprintf("UID#%.0f", fromUID)
		}
	}
	
	// 处理消息内容
	if content != "" && userid != "" {
		fmt.Printf("收到来自 %s 的VoceChat消息：userid=%s, text=%s\n", source, userid, content)
		return &models.CommingMessage{
			Channel:  models.MessageChannelVoceChat,
			Source:   source,
			UserID:   userid,
			Username: userid,
			Text:     content,
		}
	}
	
	return nil
}

// PostMessage 发送消�?func (v *VoceChatModule) PostMessage(message *models.Notification) {
	configs := v.GetConfigs()
	for name, conf := range configs {
		// 检查消�?		if !v.CheckMessage(message, &name) {
			continue
		}
		
		var userID string
		if message.UserID != nil {
			userID = *message.UserID
		} else if message.Targets != nil {
			if id, ok := (*message.Targets)["vocechat_userid"].(string); ok {
				userID = id
			}
		}
		
		// 获取实例
		instance := v.GetInstance(&name)
		if instance != nil {
			if vocechat, ok := instance.(*VoceChat); ok {
				vocechat.SendMsg(message.Title, message.Text, userID, message.Link)
			}
		}
	}
}

// PostMediasMessage 发送媒体信息选择列表
func (v *VoceChatModule) PostMediasMessage(message *models.Notification, medias []models.MediaInfo) {
	configs := v.GetConfigs()
	for name, conf := range configs {
		// 检查消�?		if !v.CheckMessage(message, &name) {
			continue
		}
		
		// 获取实例
		instance := v.GetInstance(&name)
		if instance != nil {
			if vocechat, ok := instance.(*VoceChat); ok {
				// 先发送标�?				if message.Title != "" {
					vocechat.SendMsg(message.Title, "", message.UserID, message.Link)
				}
				
				// 再发送内�?				vocechat.SendMediasMsg(medias, message.UserID, message.Title, message.Link)
			}
		}
	}
}

// PostTorrentsMessage 发送种子信息选择列表
func (v *VoceChatModule) PostTorrentsMessage(message *models.Notification, torrents []models.Context) {
	configs := v.GetConfigs()
	for name, conf := range configs {
		// 检查消�?		if !v.CheckMessage(message, &name) {
			continue
		}
		
		var userID string
		if message.UserID != nil {
			userID = *message.UserID
		} else if message.Targets != nil {
			if id, ok := (*message.Targets)["vocechat_userid"].(string); ok {
				userID = id
			}
		}
		
		if userID == "" {
			fmt.Println("用户没有指定 VoceChat用户ID，消息无法发�?)
			return
		}
		
		// 获取实例
		instance := v.GetInstance(&name)
		if instance != nil {
			if vocechat, ok := instance.(*VoceChat); ok {
				vocechat.SendTorrentsMsg(torrents, userID, message.Title, message.Link)
			}
		}
	}
}

// RegisterCommands 注册命令
func (v *VoceChatModule) RegisterCommands(commands map[string]map[string]interface{}) {
	// VoceChat模块不需要注册命�?}
