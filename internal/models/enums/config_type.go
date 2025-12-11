package enums

// SystemConfigKey 系统配置Key字典
type SystemConfigKey string

const (
	// 下载器配置
	SystemConfigKeyDownloaders SystemConfigKey = "Downloaders"
	// 媒体服务器配置
	SystemConfigKeyMediaServers SystemConfigKey = "MediaServers"
	// 消息通知配置
	SystemConfigKeyNotifications SystemConfigKey = "Notifications"
	// 通知场景开关设置
	SystemConfigKeyNotificationSwitchs SystemConfigKey = "NotificationSwitchs"
	// 目录配置
	SystemConfigKeyDirectories SystemConfigKey = "Directories"
	// 存储配置
	SystemConfigKeyStorages SystemConfigKey = "Storages"
	// 搜索站点范围
	SystemConfigKeyIndexerSites SystemConfigKey = "IndexerSites"
	// 订阅站点范围
	SystemConfigKeyRssSites SystemConfigKey = "RssSites"
	// 自定义制作组/字幕组
	SystemConfigKeyCustomReleaseGroups SystemConfigKey = "CustomReleaseGroups"
	// 自定义占位符
	SystemConfigKeyCustomization SystemConfigKey = "Customization"
	// 自定义识别词
	SystemConfigKeyCustomIdentifiers SystemConfigKey = "CustomIdentifiers"
	// 转移屏蔽词
	SystemConfigKeyTransferExcludeWords SystemConfigKey = "TransferExcludeWords"
	// 种子优先级规则
	SystemConfigKeyTorrentsPriority SystemConfigKey = "TorrentsPriority"
	// 用户自定义规则
	SystemConfigKeyCustomFilterRules SystemConfigKey = "CustomFilterRules"
	// 用户规则组
	SystemConfigKeyUserFilterRuleGroups SystemConfigKey = "UserFilterRuleGroups"
	// 搜索默认过滤规则组
	SystemConfigKeySearchFilterRuleGroups SystemConfigKey = "SearchFilterRuleGroups"
	// 订阅默认过滤规则组
	SystemConfigKeySubscribeFilterRuleGroups SystemConfigKey = "SubscribeFilterRuleGroups"
	// 订阅默认参数
	SystemConfigKeySubscribeDefaultParams SystemConfigKey = "SubscribeDefaultParams"
	// 洗版默认过滤规则组
	SystemConfigKeyBestVersionFilterRuleGroups SystemConfigKey = "BestVersionFilterRuleGroups"
	// 订阅统计
	SystemConfigKeySubscribeReport SystemConfigKey = "SubscribeReport"
	// 用户自定义CSS
	SystemConfigKeyUserCustomCSS SystemConfigKey = "UserCustomCSS"
	// 用户已安装的插件
	SystemConfigKeyUserInstalledPlugins SystemConfigKey = "UserInstalledPlugins"
	// 插件文件夹分组配置
	SystemConfigKeyPluginFolders SystemConfigKey = "PluginFolders"
	// 默认电影订阅规则
	SystemConfigKeyDefaultMovieSubscribeConfig SystemConfigKey = "DefaultMovieSubscribeConfig"
	// 默认电视剧订阅规则
	SystemConfigKeyDefaultTvSubscribeConfig SystemConfigKey = "DefaultTvSubscribeConfig"
	// 用户站点认证参数
	SystemConfigKeyUserSiteAuthParams SystemConfigKey = "UserSiteAuthParams"
	// Follow订阅分享者
	SystemConfigKeyFollowSubscribers SystemConfigKey = "FollowSubscribers"
	// 通知发送时间
	SystemConfigKeyNotificationSendTime SystemConfigKey = "NotificationSendTime"
	// 通知消息格式模板
	SystemConfigKeyNotificationTemplates SystemConfigKey = "NotificationTemplates"
	// 刮削开关设置
	SystemConfigKeyScrapingSwitchs SystemConfigKey = "ScrapingSwitchs"
	// 插件安装统计
	SystemConfigKeyPluginInstallReport SystemConfigKey = "PluginInstallReport"
	// 配置向导状态
	SystemConfigKeySetupWizardState SystemConfigKey = "SetupWizardState"
)

// UserConfigKey 用户配置Key字典
type UserConfigKey string

const (
	// 监控面板
	UserConfigKeyDashboard UserConfigKey = "Dashboard"
)
