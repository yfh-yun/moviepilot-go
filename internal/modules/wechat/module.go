package wechat

import (
	"fmt"
	"strings"
	
	"moviepilot-go/pkg/models"
	"moviepilot-go/pkg/modules"
)

// WechatModule 微信模块
type WechatModule struct {
	*modules.ModuleBase
	*modules.MessageBase
	channel models.MessageChannel
}

// NewWechatModule 创建新的微信模块实例
func NewWechatModule() *WechatModule {
	module := &WechatModule{
		ModuleBase:  modules.NewModuleBase(),
		MessageBase: modules.NewMessageBase(),
	}
	
	// 设置模块属�?	module.Name = "微信"
	module.Type = models.ModuleTypeNotification
	module.SubType = models.MessageChannelWechat
	module.Priority = 1
	
	return module
}

// InitModule 初始化模�?func (w *WechatModule) InitModule() error {
	// 初始化服�?	// 注意：这里需要根据实际情况调整实�?	// super().init_service(service_name=WeChat.__name__.lower(), service_type=WeChat)
	
	w.Channel = &models.MessageChannelWechat
	return nil
}

// HandleConfigChanged 处理配置变更事件
func (w *WechatModule) HandleConfigChanged(eventData *models.ConfigChangeEventData) {
	if eventData == nil {
		return
	}
	
	// 检查是否是通知配置变更
	if eventData.Key != string(models.SystemConfigKeyNotifications) {
		return
	}
	
	fmt.Println("配置变更，重新加载Wechat模块...")
	w.InitModule()
}

// GetName 获取模块名称
func (w *WechatModule) GetName() string {
	return "微信"
}

// GetType 获取模块类型
func (w *WechatModule) GetType() models.ModuleType {
	return models.ModuleTypeNotification
}

// GetSubType 获取模块的子类型
func (w *WechatModule) GetSubType() interface{} {
	return models.MessageChannelWechat
}

// GetPriority 获取模块优先�?func (w *WechatModule) GetPriority() int {
	return 1
}

// Stop 停止模块
func (w *WechatModule) Stop() error {
	// 实现停止逻辑
	return nil
}

// Test 测试模块连接�?func (w *WechatModule) Test() (bool, string) {
	// 获取实例
	instances := w.GetInstances()
	if len(instances) == 0 {
		return true, "" // 没有配置实例，返回成�?	}
	
	// 检查每个实例的状�?	for name, client := range instances {
		// 类型断言获取WeChat实例
		if wechat, ok := client.(*WeChat); ok {
			state := wechat.GetState()
			if !state {
				return false, fmt.Sprintf("企业微信 %s 未就�?, name)
			}
		}
	}
	
	return true, ""
}

// InitSetting 初始化设�?func (w *WechatModule) InitSetting() (string, interface{}) {
	// 实现初始化设置逻辑
	return "", nil
}

// MessageParser 解析消息内容
func (w *WechatModule) MessageParser(source string, body []byte, form map[string]interface{}, 
	args map[string]interface{}) *models.CommingMessage {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("微信消息处理发生错误: %v\n", r)
		}
	}()
	
	// 获取服务配置
	clientConfig := w.GetConfig(&source)
	if clientConfig == nil {
		return nil
	}
	
	// 获取实例
	// client := w.GetInstance(&clientConfig.Name)
	
	// URL参数
	sVerifyMsgSig, ok1 := args["msg_signature"].(string)
	sVerifyTimeStamp, ok2 := args["timestamp"].(string)
	sVerifyNonce, ok3 := args["nonce"].(string)
	
	if !ok1 || !ok2 || !ok3 {
		fmt.Printf("微信请求参数错误: %v\n", args)
		return nil
	}
	
	// 获取配置参数
	configMap := clientConfig.(map[string]interface{})
	wechatToken, _ := configMap["WECHAT_TOKEN"].(string)
	wechatEncodingAESKey, _ := configMap["WECHAT_ENCODING_AESKEY"].(string)
	wechatCorpID, _ := configMap["WECHAT_CORPID"].(string)
	
	// 解密模块
	wxcpt, err := NewWXBizMsgCrypt(wechatToken, wechatEncodingAESKey, wechatCorpID)
	if err != nil {
		fmt.Printf("创建微信解密模块失败: %v\n", err)
		return nil
	}
	
	// 报文数据
	if len(body) == 0 {
		fmt.Println("微信请求数据为空")
		return nil
	}
	
	fmt.Printf("收到微信请求: %s\n", string(body))
	
	ret, sMsg := wxcpt.DecryptMsg(string(body), sVerifyMsgSig, sVerifyTimeStamp, sVerifyNonce)
	if ret != WXBizMsgCryptOK {
		fmt.Printf("解密微信消息失败 DecryptMsg ret = %d\n", ret)
		return nil
	}
	
	// 解析XML报文
	// 注意：这里需要实现XML解析逻辑
	
	// 消息类型
	// msgType := DomUtils.tag_value(root_node, "MsgType")
	// Event event事件只有click才有�?enter_agent无效
	// event := DomUtils.tag_value(root_node, "Event")
	// 用户ID
	// userID := DomUtils.tag_value(root_node, "FromUserName")
	
	// 没的消息类型和用户ID的消息不�?	// if msgType == "" || userID == "" {
	//     fmt.Println("解析不到消息类型和用户ID")
	//     return nil
	// }
	
	// 解析消息内容
	// var content string
	// if msgType == "event" && event == "click" {
	//     // 校验用户有权限执行交互命�?	//     if wechatAdmins, ok := configMap["WECHAT_ADMINS"].(string); ok && wechatAdmins != "" {
	//         adminList := strings.Split(wechatAdmins, ",")
	//         hasPermission := false
	//         for _, admin := range adminList {
	//             if userID == strings.TrimSpace(admin) {
	//                 hasPermission = true
	//                 break
	//             }
	//         }
	//         
	//         if !hasPermission {
	//             // client.send_msg(title="用户无权限执行菜单命�?, userid=user_id)
	//             return nil
	//         }
	//     }
	//     
	//     // 根据EventKey执行命令
	//     content = DomUtils.tag_value(root_node, "EventKey")
	//     fmt.Printf("收到来自 %s 的微信事�? userid=%s, event=%s\n", clientConfig.Name, userID, content)
	// } else if msgType == "text" {
	//     // 文本消息
	//     content = DomUtils.tag_value(root_node, "Content", default="")
	//     fmt.Printf("收到来自 %s 的微信消�? userid=%s, text=%s\n", clientConfig.Name, userID, content)
	// } else {
	//     return nil
	// }
	
	// if content != "" {
	//     // 处理消息内容
	//     return &models.CommingMessage{
	//         Channel:  models.MessageChannelWechat,
	//         Source:   clientConfig.Name,
	//         UserID:   userID,
	//         Username: userID,
	//         Text:     content,
	//     }
	// }
	
	return nil
}

// PostMessage 发送消�?func (w *WechatModule) PostMessage(message *models.Notification) {
	configs := w.GetConfigs()
	for name, conf := range configs {
		// 检查消�?		if !w.CheckMessage(message, &name) {
			continue
		}
		
		var userID string
		if message.UserID != nil {
			userID = *message.UserID
		} else if message.Targets != nil {
			if id, ok := (*message.Targets)["wechat_userid"].(string); ok {
				userID = id
			}
		}
		
		if userID == "" {
			fmt.Println("用户没有指定 微信用户ID，消息无法发�?)
			return
		}
		
		// 获取实例
		instance := w.GetInstance(&name)
		if instance != nil {
			if wechat, ok := instance.(*WeChat); ok {
				wechat.SendMsg(message.Title, message.Text, message.Image, userID, message.Link)
			}
		}
	}
}

// PostMediasMessage 发送媒体信息选择列表
func (w *WechatModule) PostMediasMessage(message *models.Notification, medias []models.MediaInfo) {
	configs := w.GetConfigs()
	for name, conf := range configs {
		// 检查消�?		if !w.CheckMessage(message, &name) {
			continue
		}
		
		// 获取实例
		instance := w.GetInstance(&name)
		if instance != nil {
			if wechat, ok := instance.(*WeChat); ok {
				// 先发送标�?				if message.Title != "" {
					wechat.SendMsg(message.Title, "", "", message.UserID, message.Link)
				}
				
				// 再发送内�?				wechat.SendMediasMsg(medias, message.UserID)
			}
		}
	}
}

// PostTorrentsMessage 发送种子信息选择列表
func (w *WechatModule) PostTorrentsMessage(message *models.Notification, torrents []models.Context) {
	configs := w.GetConfigs()
	for name, conf := range configs {
		// 检查消�?		if !w.CheckMessage(message, &name) {
			continue
		}
		
		// 获取实例
		instance := w.GetInstance(&name)
		if instance != nil {
			if wechat, ok := instance.(*WeChat); ok {
				wechat.SendTorrentsMsg(torrents, message.UserID, message.Title, message.Link)
			}
		}
	}
}

// RegisterCommands 注册命令
func (w *WechatModule) RegisterCommands(commands map[string]map[string]interface{}) {
	configs := w.GetConfigs()
	for name, clientConfig := range configs {
		// 类型断言获取配置
		configMap, ok := clientConfig.(map[string]interface{})
		if !ok {
			continue
		}
		
		// 如果没有配置消息解密相关参数，则也没有必要进行菜单初始化
		encodingAESKey, ok1 := configMap["WECHAT_ENCODING_AESKEY"].(string)
		token, ok2 := configMap["WECHAT_TOKEN"].(string)
		
		if !ok1 || !ok2 || encodingAESKey == "" || token == "" {
			fmt.Printf("%s 缺少消息解密参数，跳过后续菜单初始化\n", name)
			continue
		}
		
		// 获取实例
		instance := w.GetInstance(&name)
		if instance == nil {
			continue
		}
		
		// 检查是否是WeChat实例
		if wechat, ok := instance.(*WeChat); ok {
			// 过滤命令，只保留有category的命�?			filteredCommands := make(map[string]map[string]interface{})
			for key, value := range commands {
				if category, ok := value["category"].(string); ok && category != "" {
					filteredCommands[key] = value
				}
			}
			
			// 如果过滤后的命令为空，则跳过注册
			if len(filteredCommands) == 0 {
				fmt.Println("Filtered commands are empty, skipping registration.")
				wechat.DeleteMenus()
				continue
			}
			
			// 创建菜单
			wechat.CreateMenus(filteredCommands)
		}
	}
}
