package enums

// NotificationType 消息类型
type NotificationType string

const (
	// 资源下载
	NotificationTypeDownload NotificationType = "资源下载"
	// 整理入库
	NotificationTypeOrganize NotificationType = "整理入库"
	// 订阅
	NotificationTypeSubscribe NotificationType = "订阅"
	// 站点消息
	NotificationTypeSiteMessage NotificationType = "站点"
	// 媒体服务器通知
	NotificationTypeMediaServer NotificationType = "媒体服务器"
	// 处理失败需要人工干预
	NotificationTypeManual NotificationType = "手动处理"
	// 插件消息
	NotificationTypePlugin NotificationType = "插件"
	// 其它消息
	NotificationTypeOther NotificationType = "其它"
)

// ContentType 消息内容类型
type ContentType string

const (
	// 订阅添加成功
	ContentTypeSubscribeAdded ContentType = "subscribeAdded"
	// 订阅完成
	ContentTypeSubscribeComplete ContentType = "subscribeComplete"
	// 入库成功
	ContentTypeOrganizeSuccess ContentType = "organizeSuccess"
	// 下载开始(添加下载任务成功)
	ContentTypeDownloadAdded ContentType = "downloadAdded"
)

// MessageChannel 消息渠道
type MessageChannel string

const (
	MessageChannelWechat       MessageChannel = "微信"
	MessageChannelTelegram     MessageChannel = "Telegram"
	MessageChannelSlack        MessageChannel = "Slack"
	MessageChannelSynologyChat MessageChannel = "SynologyChat"
	MessageChannelVoceChat     MessageChannel = "VoceChat"
	MessageChannelWeb          MessageChannel = "Web"
	MessageChannelWebPush      MessageChannel = "WebPush"
)
