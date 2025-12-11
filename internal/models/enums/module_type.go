package enums

// ModuleType 模块类型
type ModuleType string

const (
	// 下载器
	ModuleTypeDownloader ModuleType = "downloader"
	// 媒体服务器
	ModuleTypeMediaServer ModuleType = "mediaserver"
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
	// 过滤器
	OtherModulesTypeFilter OtherModulesType = "过滤器"
	// 站点索引
	OtherModulesTypeIndexer OtherModulesType = "站点索引"
	// PostgreSQL
	OtherModulesTypePostgreSQL OtherModulesType = "PostgreSQL"
	// Redis
	OtherModulesTypeRedis OtherModulesType = "Redis"
)

// DownloaderType 下载器类型
type DownloaderType string

const (
	// Qbittorrent
	DownloaderTypeQbittorrent DownloaderType = "Qbittorrent"
	// Transmission
	DownloaderTypeTransmission DownloaderType = "Transmission"
)

// MediaServerType 媒体服务器类型
type MediaServerType string

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
