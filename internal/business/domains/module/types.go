package module

// ModuleType 模块类型枚举
type ModuleType string

const (
	ModuleTypeDownloader  ModuleType = "downloader"
	ModuleTypeMediaServer ModuleType = "mediaserver"
	ModuleTypeMessage     ModuleType = "message"
	ModuleTypeStorage     ModuleType = "storage"
	ModuleTypeOther       ModuleType = "other"
)

// DownloaderType 下载器类型枚举
type DownloaderType string

const (
	DownloaderTypeQBittorrent  DownloaderType = "qbittorrent"
	DownloaderTypeTransmission DownloaderType = "transmission"
	DownloaderTypeDeluge       DownloaderType = "deluge"
)

// MediaServerType 媒体服务器类型枚举
type MediaServerType string

const (
	MediaServerTypeEmby     MediaServerType = "emby"
	MediaServerTypePlex     MediaServerType = "plex"
	MediaServerTypeJellyfin MediaServerType = "jellyfin"
)

// MessageChannel 消息通道类型枚举
type MessageChannel string

const (
	MessageChannelTelegram MessageChannel = "telegram"
	MessageChannelSlack    MessageChannel = "slack"
	MessageChannelWeChat   MessageChannel = "wechat"
	MessageChannelDiscord  MessageChannel = "discord"
)

// StorageSchema 存储模式类型枚举
type StorageSchema string

const (
	StorageSchemaRclone   StorageSchema = "rclone"
	StorageSchemaLocal    StorageSchema = "local"
	StorageSchemaAliPan   StorageSchema = "alipan"
	StorageSchemaOneDrive StorageSchema = "onedrive"
)

// OtherModulesType 其他模块类型枚举
type OtherModulesType string

const (
	OtherModulesTypeRecommendation OtherModulesType = "recommendation"
	OtherModulesTypeNotification   OtherModulesType = "notification"
)

// Module 通用模块接口
type Module interface {
	// ID 获取模块ID
	ID() string

	// Type 获取模块类型
	Type() ModuleType

	// SubType 获取模块子类型
	SubType() string

	// Init 初始化模块
	Init(cfg any) error

	// Stop 停止模块
	Stop() error

	// Test 测试模块连接
	Test() (bool, string)

	// SettingInfo 获取模块配置开关信息
	SettingInfo() (*Setting, bool)
}

// Setting 模块配置开关信息
type Setting struct {
	Key   string // 对应配置中的字段名
	Value any    // True 或某个子类型值
}
