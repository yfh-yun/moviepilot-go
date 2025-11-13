package models

// MediaType 媒体类型
type MediaType string

const (
	MediaTypeMovie      MediaType = "电影"
	MediaTypeTV         MediaType = "电视�?
	MediaTypeCollection MediaType = "系列"
	MediaTypeUnknown    MediaType = "未知"
)

// SortType 排序类型枚举
type SortType string

const (
	SortTypeTime   SortType = "time"   // 按时间排�?	SortTypeCount  SortType = "count"  // 按人数排�?	SortTypeRating SortType = "rating" // 按评分排�?)

// TorrentStatus 种子状�?type TorrentStatus string

const (
	TorrentStatusTransfer    TorrentStatus = "可转�?
	TorrentStatusDownloading TorrentStatus = "下载�?
)

// EventType 异步广播事件
type EventType string

const (
	// 插件需要重�?	EventTypePluginReload EventType = "plugin.reload"
	// 触发插件动作
	EventTypePluginAction EventType = "plugin.action"
	// 插件触发事件
	EventTypePluginTriggered EventType = "plugin.triggered"
	// 执行命令
	EventTypeCommandExcute EventType = "command.excute"
	// 站点已删�?	EventTypeSiteDeleted EventType = "site.deleted"
	// 站点已更�?	EventTypeSiteUpdated EventType = "site.updated"
	// 站点已刷�?	EventTypeSiteRefreshed EventType = "site.refreshed"
	// 转移完成
	EventTypeTransferComplete EventType = "transfer.complete"
	// 下载已添�?	EventTypeDownloadAdded EventType = "download.added"
	// 删除历史记录
	EventTypeHistoryDeleted EventType = "history.deleted"
	// 删除下载源文�?	EventTypeDownloadFileDeleted EventType = "downloadfile.deleted"
	// 删除下载任务
	EventTypeDownloadDeleted EventType = "download.deleted"
	// 收到用户外来消息
	EventTypeUserMessage EventType = "user.message"
	// 收到Webhook消息
	EventTypeWebhookMessage EventType = "webhook.message"
	// 发送消息通知
	EventTypeNoticeMessage EventType = "notice.message"
	// 订阅已添�?	EventTypeSubscribeAdded EventType = "subscribe.added"
	// 订阅已调�?	EventTypeSubscribeModified EventType = "subscribe.modified"
	// 订阅已删�?	EventTypeSubscribeDeleted EventType = "subscribe.deleted"
	// 订阅已完�?	EventTypeSubscribeComplete EventType = "subscribe.complete"
	// 系统错误
	EventTypeSystemError EventType = "system.error"
	// 刮削元数�?	EventTypeMetadataScrape EventType = "metadata.scrape"
	// 模块需要重�?	EventTypeModuleReload EventType = "module.reload"
	// 配置项更�?	EventTypeConfigChanged EventType = "config.updated"
	// 消息交互动作
	EventTypeMessageAction EventType = "message.action"
	// 执行工作�?	EventTypeWorkflowExecute EventType = "workflow.execute"
)

// EventType中文名称翻译字典
var EventTypeNames = map[EventType]string{
	EventTypePluginReload:        "插件重载",
	EventTypePluginAction:        "触发插件动作",
	EventTypePluginTriggered:     "触发插件事件",
	EventTypeCommandExcute:       "执行命令",
	EventTypeSiteDeleted:         "站点已删�?,
	EventTypeSiteUpdated:         "站点已更�?,
	EventTypeSiteRefreshed:       "站点已刷�?,
	EventTypeTransferComplete:    "整理完成",
	EventTypeDownloadAdded:       "添加下载",
	EventTypeHistoryDeleted:      "删除历史记录",
	EventTypeDownloadFileDeleted: "删除下载源文�?,
	EventTypeDownloadDeleted:     "删除下载任务",
	EventTypeUserMessage:         "收到用户消息",
	EventTypeWebhookMessage:      "收到Webhook消息",
	EventTypeNoticeMessage:       "发送消息通知",
	EventTypeSubscribeAdded:      "添加订阅",
	EventTypeSubscribeModified:   "订阅已调�?,
	EventTypeSubscribeDeleted:    "订阅已删�?,
	EventTypeSubscribeComplete:   "订阅已完�?,
	EventTypeSystemError:         "系统错误",
	EventTypeMetadataScrape:      "刮削元数�?,
	EventTypeModuleReload:        "模块重载",
	EventTypeConfigChanged:       "配置项更�?,
	EventTypeMessageAction:       "消息交互动作",
	EventTypeWorkflowExecute:     "执行工作�?,
}

// ChainEventType 同步链式事件
type ChainEventType string

const (
	// 名称识别
	ChainEventTypeNameRecognize ChainEventType = "name.recognize"
	// 认证验证
	ChainEventTypeAuthVerification ChainEventType = "auth.verification"
	// 认证拦截
	ChainEventTypeAuthIntercept ChainEventType = "auth.intercept"
	// 命令注册
	ChainEventTypeCommandRegister ChainEventType = "command.register"
	// 整理重命�?	ChainEventTypeTransferRename ChainEventType = "transfer.rename"
	// 整理拦截
	ChainEventTypeTransferIntercept ChainEventType = "transfer.intercept"
	// 资源选择
	ChainEventTypeResourceSelection ChainEventType = "resource.selection"
	// 资源下载
	ChainEventTypeResourceDownload ChainEventType = "resource.download"
	// 探索数据�?	ChainEventTypeDiscoverSource ChainEventType = "discover.source"
	// 媒体识别转换
	ChainEventTypeMediaRecognizeConvert ChainEventType = "media.recognize.convert"
	// 推荐数据�?	ChainEventTypeRecommendSource ChainEventType = "recommend.source"
	// 工作流执�?	ChainEventTypeWorkflowExecution ChainEventType = "workflow.execution"
	// 存储操作选择
	ChainEventTypeStorageOperSelection ChainEventType = "storage.operation"
)

// SystemConfigKey 系统配置Key字典
type SystemConfigKey string

const (
	// 下载器配�?	SystemConfigKeyDownloaders SystemConfigKey = "Downloaders"
	// 媒体服务器配�?	SystemConfigKeyMediaServers SystemConfigKey = "MediaServers"
	// 消息通知配置
	SystemConfigKeyNotifications SystemConfigKey = "Notifications"
	// 通知场景开关设�?	SystemConfigKeyNotificationSwitchs SystemConfigKey = "NotificationSwitchs"
	// 目录配置
	SystemConfigKeyDirectories SystemConfigKey = "Directories"
	// 存储配置
	SystemConfigKeyStorages SystemConfigKey = "Storages"
	// 搜索站点范围
	SystemConfigKeyIndexerSites SystemConfigKey = "IndexerSites"
	// 订阅站点范围
	SystemConfigKeyRssSites SystemConfigKey = "RssSites"
	// 自定义制作组/字幕�?	SystemConfigKeyCustomReleaseGroups SystemConfigKey = "CustomReleaseGroups"
	// 自定义占位符
	SystemConfigKeyCustomization SystemConfigKey = "Customization"
	// 自定义识别词
	SystemConfigKeyCustomIdentifiers SystemConfigKey = "CustomIdentifiers"
	// 转移屏蔽�?	SystemConfigKeyTransferExcludeWords SystemConfigKey = "TransferExcludeWords"
	// 种子优先级规�?	SystemConfigKeyTorrentsPriority SystemConfigKey = "TorrentsPriority"
	// 用户自定义规�?	SystemConfigKeyCustomFilterRules SystemConfigKey = "CustomFilterRules"
	// 用户规则�?	SystemConfigKeyUserFilterRuleGroups SystemConfigKey = "UserFilterRuleGroups"
	// 搜索默认过滤规则�?	SystemConfigKeySearchFilterRuleGroups SystemConfigKey = "SearchFilterRuleGroups"
	// 订阅默认过滤规则�?	SystemConfigKeySubscribeFilterRuleGroups SystemConfigKey = "SubscribeFilterRuleGroups"
	// 订阅默认参数
	SystemConfigKeySubscribeDefaultParams SystemConfigKey = "SubscribeDefaultParams"
	// 洗版默认过滤规则�?	SystemConfigKeyBestVersionFilterRuleGroups SystemConfigKey = "BestVersionFilterRuleGroups"
	// 订阅统计
	SystemConfigKeySubscribeReport SystemConfigKey = "SubscribeReport"
	// 用户自定义CSS
	SystemConfigKeyUserCustomCSS SystemConfigKey = "UserCustomCSS"
	// 用户已安装的插件
	SystemConfigKeyUserInstalledPlugins SystemConfigKey = "UserInstalledPlugins"
	// 插件文件夹分组配�?	SystemConfigKeyPluginFolders SystemConfigKey = "PluginFolders"
	// 默认电影订阅规则
	SystemConfigKeyDefaultMovieSubscribeConfig SystemConfigKey = "DefaultMovieSubscribeConfig"
	// 默认电视剧订阅规�?	SystemConfigKeyDefaultTvSubscribeConfig SystemConfigKey = "DefaultTvSubscribeConfig"
	// 用户站点认证参数
	SystemConfigKeyUserSiteAuthParams SystemConfigKey = "UserSiteAuthParams"
	// Follow订阅分享�?	SystemConfigKeyFollowSubscribers SystemConfigKey = "FollowSubscribers"
	// 通知发送时�?	SystemConfigKeyNotificationSendTime SystemConfigKey = "NotificationSendTime"
	// 通知消息格式模板
	SystemConfigKeyNotificationTemplates SystemConfigKey = "NotificationTemplates"
	// 刮削开关设�?	SystemConfigKeyScrapingSwitchs SystemConfigKey = "ScrapingSwitchs"
	// 插件安装统计
	SystemConfigKeyPluginInstallReport SystemConfigKey = "PluginInstallReport"
	// 配置向导状�?	SystemConfigKeySetupWizardState SystemConfigKey = "SetupWizardState"
)

// ProgressKey 处理进度Key字典
type ProgressKey string

const (
	// 搜索
	ProgressKeySearch ProgressKey = "search"
	// 整理
	ProgressKeyFileTransfer ProgressKey = "filetransfer"
	// 批量重命�?	ProgressKeyBatchRename ProgressKey = "batchrename"
)

// MediaImageType 媒体图片类型
type MediaImageType string

const (
	MediaImageTypePoster   MediaImageType = "poster_path"
	MediaImageTypeBackdrop MediaImageType = "backdrop_path"
)

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
	NotificationTypeMediaServer NotificationType = "媒体服务�?
	// 处理失败需要人工干�?	NotificationTypeManual NotificationType = "手动处理"
	// 插件消息
	NotificationTypePlugin NotificationType = "插件"
	// 其它消息
	NotificationTypeOther NotificationType = "其它"
)

// ContentType 消息内容类型
// 操作状态的通知消息类型标识
type ContentType string

const (
	// 订阅添加成功
	ContentTypeSubscribeAdded ContentType = "subscribeAdded"
	// 订阅完成
	ContentTypeSubscribeComplete ContentType = "subscribeComplete"
	// 入库成功
	ContentTypeOrganizeSuccess ContentType = "organizeSuccess"
	// 下载开�?添加下载任务成功)
	ContentTypeDownloadAdded ContentType = "downloadAdded"
)

// MessageChannel 消息渠道
type MessageChannel string

const (
	// 微信
	MessageChannelWechat MessageChannel = "微信"
	// Telegram
	MessageChannelTelegram MessageChannel = "Telegram"
	// Slack
	MessageChannelSlack MessageChannel = "Slack"
	// SynologyChat
	MessageChannelSynologyChat MessageChannel = "SynologyChat"
	// VoceChat
	MessageChannelVoceChat MessageChannel = "VoceChat"
	// Web
	MessageChannelWeb MessageChannel = "Web"
	// WebPush
	MessageChannelWebPush MessageChannel = "WebPush"
)

// DownloaderType 下载器类�?type DownloaderType string

const (
	// Qbittorrent
	DownloaderTypeQbittorrent DownloaderType = "Qbittorrent"
	// Transmission
	DownloaderTypeTransmission DownloaderType = "Transmission"
)

// MediaServerType 媒体服务器类�?type MediaServerType string

const (
	// Emby
	MediaServerTypeEmby MediaServerType = "Emby"
	// Jellyfin
	MediaServerTypeJellyfin MediaServerType = "Jellyfin"
	// Plex
	MediaServerTypePlex MediaServerType = "Plex"
	// 飞牛影视
	MediaServerTypeTrimeMedia MediaServerType = "TrimeMedia"
)

// MediaRecognizeType 识别器类�?type MediaRecognizeType string

const (
	// 豆瓣
	MediaRecognizeTypeDouban MediaRecognizeType = "豆瓣"
	// TMDB
	MediaRecognizeTypeTMDB MediaRecognizeType = "TheMovieDb"
	// TVDB
	MediaRecognizeTypeTVDB MediaRecognizeType = "TheTvDb"
	// bangumi
	MediaRecognizeTypeBangumi MediaRecognizeType = "Bangumi"
)

// UserConfigKey 用户配置Key字典
type UserConfigKey string

const (
	// 监控面板
	UserConfigKeyDashboard UserConfigKey = "Dashboard"
)

// StorageSchema 支持的存储类�?type StorageSchema string

const (
	// 存储类型
	StorageSchemaLocal   StorageSchema = "local"
	StorageSchemaAlipan  StorageSchema = "alipan"
	StorageSchemaU115    StorageSchema = "u115"
	StorageSchemaRclone  StorageSchema = "rclone"
	StorageSchemaAlist   StorageSchema = "alist"
	StorageSchemaSMB     StorageSchema = "smb"
)

// ModuleType 模块类型
type ModuleType string

const (
	// 下载�?	ModuleTypeDownloader ModuleType = "downloader"
	// 媒体服务�?	ModuleTypeMediaServer ModuleType = "mediaserver"
	// 消息服务
	ModuleTypeNotification ModuleType = "notification"
	// 媒体识别
	ModuleTypeMediaRecognize ModuleType = "mediarecognize"
	// 站点索引
	ModuleTypeIndexer ModuleType = "indexer"
	// 其它
	ModuleTypeOther ModuleType = "other"
)

// OtherModulesType 其他杂项模块类型
type OtherModulesType string

const (
	// 字幕
	OtherModulesTypeSubtitle OtherModulesType = "站点字幕"
	// Fanart
	OtherModulesTypeFanart OtherModulesType = "Fanart"
	// 文件整理
	OtherModulesTypeFileManager OtherModulesType = "文件整理"
	// 过滤�?	OtherModulesTypeFilter OtherModulesType = "过滤�?
	// 站点索引
	OtherModulesTypeIndexer OtherModulesType = "站点索引"
	// PostgreSQL
	OtherModulesTypePostgreSQL OtherModulesType = "PostgreSQL"
	// Redis
	OtherModulesTypeRedis OtherModulesType = "Redis"
)
