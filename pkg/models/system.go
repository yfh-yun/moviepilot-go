package models

// ServiceInfo 封装服务相关信息的数据类
type ServiceInfo struct {
	// 名称
	Name string `json:"name,omitempty"`
	// 实例
	Instance interface{} `json:"instance,omitempty"`
	// 模块
	Module interface{} `json:"module,omitempty"`
	// 类型
	Type string `json:"type,omitempty"`
	// 配置
	Config interface{} `json:"config,omitempty"`
}

// MediaServerConf 媒体服务器配�?type MediaServerConf struct {
	// 名称
	Name string `json:"name,omitempty"`
	// 类型 emby/jellyfin/plex
	Type string `json:"type,omitempty"`
	// 配置
	Config map[string]interface{} `json:"config,omitempty"`
	// 是否启用
	Enabled bool `json:"enabled,omitempty"`
	// 同步媒体体库列表
	SyncLibraries []interface{} `json:"sync_libraries,omitempty"`
}

// DownloaderConf 下载器配�?type DownloaderConf struct {
	// 名称
	Name string `json:"name,omitempty"`
	// 类型 qbittorrent/transmission
	Type string `json:"type,omitempty"`
	// 是否默认
	Default bool `json:"default,omitempty"`
	// 配置
	Config map[string]interface{} `json:"config,omitempty"`
	// 是否启用
	Enabled bool `json:"enabled,omitempty"`
}

// NotificationConf 通知配置
type NotificationConf struct {
	// 名称
	Name string `json:"name,omitempty"`
	// 类型 telegram/wechat/vocechat/synologychat/slack/webpush
	Type string `json:"type,omitempty"`
	// 配置
	Config map[string]interface{} `json:"config,omitempty"`
	// 场景开�?	Switchs []interface{} `json:"switchs,omitempty"`
	// 是否启用
	Enabled bool `json:"enabled,omitempty"`
}

// NotificationSwitchConf 通知场景开关配�?type NotificationSwitchConf struct {
	// 场景名称
	Type string `json:"type,omitempty"`
	// 通知范围 all/user/admin
	Action string `json:"action,omitempty"`
}

// StorageConf 存储配置
type StorageConf struct {
	// 类型 local/alipan/u115/rclone/alist
	Type string `json:"type,omitempty"`
	// 名称
	Name string `json:"name,omitempty"`
	// 配置
	Config map[string]interface{} `json:"config,omitempty"`
}

// TransferDirectoryConf 文件整理目录配置
type TransferDirectoryConf struct {
	// 名称
	Name string `json:"name,omitempty"`
	// 优先�?	Priority int `json:"priority,omitempty"`
	// 存储
	Storage string `json:"storage,omitempty"`
	// 下载目录
	DownloadPath string `json:"download_path,omitempty"`
	// 适用媒体类型
	MediaType string `json:"media_type,omitempty"`
	// 适用媒体类别
	MediaCategory string `json:"media_category,omitempty"`
	// 下载类型子目�?	DownloadTypeFolder bool `json:"download_type_folder,omitempty"`
	// 下载类别子目�?	DownloadCategoryFolder bool `json:"download_category_folder,omitempty"`
	// 监控方式 downloader/monitor，None为不监控
	MonitorType string `json:"monitor_type,omitempty"`
	// 监控模式 fast / compatibility
	MonitorMode string `json:"monitor_mode,omitempty"`
	// 整理方式 move/copy/link/softlink
	TransferType string `json:"transfer_type,omitempty"`
	// 文件覆盖模式 always/size/never/latest
	OverwriteMode string `json:"overwrite_mode,omitempty"`
	// 整理到媒体库目录
	LibraryPath string `json:"library_path,omitempty"`
	// 媒体库目录存�?	LibraryStorage string `json:"library_storage,omitempty"`
	// 智能重命�?	Renaming bool `json:"renaming,omitempty"`
	// 刮削
	Scraping bool `json:"scraping,omitempty"`
	// 是否发送通知
	Notify bool `json:"notify,omitempty"`
	// 媒体库类型子目录
	LibraryTypeFolder bool `json:"library_type_folder,omitempty"`
	// 媒体库类别子目录
	LibraryCategoryFolder bool `json:"library_category_folder,omitempty"`
}

// NewServiceInfo 创建一个新�?ServiceInfo 实例
func NewServiceInfo() *ServiceInfo {
	return &ServiceInfo{}
}

// NewMediaServerConf 创建一个新�?MediaServerConf 实例
func NewMediaServerConf() *MediaServerConf {
	return &MediaServerConf{
		Config:        make(map[string]interface{}),
		Enabled:       false,
		SyncLibraries: make([]interface{}, 0),
	}
}

// NewDownloaderConf 创建一个新�?DownloaderConf 实例
func NewDownloaderConf() *DownloaderConf {
	return &DownloaderConf{
		Default: false,
		Config:  make(map[string]interface{}),
		Enabled: false,
	}
}

// NewNotificationConf 创建一个新�?NotificationConf 实例
func NewNotificationConf() *NotificationConf {
	return &NotificationConf{
		Config:  make(map[string]interface{}),
		Switchs: make([]interface{}, 0),
		Enabled: false,
	}
}

// NewNotificationSwitchConf 创建一个新�?NotificationSwitchConf 实例
func NewNotificationSwitchConf() *NotificationSwitchConf {
	return &NotificationSwitchConf{
		Action: "all",
	}
}

// NewStorageConf 创建一个新�?StorageConf 实例
func NewStorageConf() *StorageConf {
	return &StorageConf{
		Config: make(map[string]interface{}),
	}
}

// NewTransferDirectoryConf 创建一个新�?TransferDirectoryConf 实例
func NewTransferDirectoryConf() *TransferDirectoryConf {
	return &TransferDirectoryConf{
		Priority:               0,
		DownloadTypeFolder:     false,
		DownloadCategoryFolder: false,
		MonitorMode:            "fast",
		Notify:                 true,
		LibraryTypeFolder:      false,
		LibraryCategoryFolder:  false,
	}
}
