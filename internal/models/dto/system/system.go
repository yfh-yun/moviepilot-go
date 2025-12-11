package system

import (
	"time"
)

// ServiceInfo 封装服务相关信息的数据类
type ServiceInfo struct {
	// 名称
	Name string
	// 实例
	Instance any
	// 模块
	Module any
	// 类型
	Type string
	// 配置
	Config any
}

// MediaServerConf 媒体服务器配置
type MediaServerConf struct {
	// 名称
	Name string `json:"name,omitempty"`
	// 类型 emby/jellyfin/plex
	Type string `json:"type,omitempty"`
	// 配置
	Config map[string]any `json:"config,omitempty"`
	// 是否启用
	Enabled bool `json:"enabled,omitempty"`
	// 同步媒体库列表
	SyncLibraries []string `json:"sync_libraries,omitempty"`
}

// DownloaderConf 下载器配置
type DownloaderConf struct {
	// 名称
	Name string `json:"name,omitempty"`
	// 类型 qbittorrent/transmission
	Type string `json:"type,omitempty"`
	// 是否默认
	Default bool `json:"default,omitempty"`
	// 配置
	Config map[string]any `json:"config,omitempty"`
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
	Config map[string]any `json:"config,omitempty"`
	// 场景开关
	Switchs []string `json:"switchs,omitempty"`
	// 是否启用
	Enabled bool `json:"enabled,omitempty"`
}

// NotificationSwitchConf 通知场景开关配置
type NotificationSwitchConf struct {
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
	Config map[string]any `json:"config,omitempty"`
}

// TransferDirectoryConf 文件整理目录配置
type TransferDirectoryConf struct {
	// 名称
	Name string `json:"name,omitempty"`
	// 优先级
	Priority int `json:"priority,omitempty"`
	// 存储
	Storage string `json:"storage,omitempty"`
	// 下载目录
	DownloadPath string `json:"download_path,omitempty"`
	// 适用媒体类型
	MediaType string `json:"media_type,omitempty"`
	// 适用媒体类别
	MediaCategory string `json:"media_category,omitempty"`
	// 下载类型子目录
	DownloadTypeFolder bool `json:"download_type_folder,omitempty"`
	// 下载类别子目录
	DownloadCategoryFolder bool `json:"download_category_folder,omitempty"`
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
	// 媒体库目录存储
	LibraryStorage string `json:"library_storage,omitempty"`
	// 智能重命名
	Renaming bool `json:"renaming,omitempty"`
	// 刮削
	Scraping bool `json:"scraping,omitempty"`
	// 是否发送通知
	Notify bool `json:"notify,omitempty"`
	// 媒体库类型子目录
	LibraryTypeFolder bool `json:"library_type_folder,omitempty"`
	// 媒体库类别子目录
	LibraryCategoryFolder bool `json:"library_category_folder,omitempty"`
}

// ProgressInfo 进度信息
type ProgressInfo struct {
	// 进度类型
	Type string `json:"type"`
	// 进度百分比
	Percent int `json:"percent"`
	// 进度状态
	Status string `json:"status"`
	// 进度消息
	Message string `json:"message"`
	// 进度详情
	Details map[string]any `json:"details,omitempty"`
	// 开始时间
	StartTime *time.Time `json:"start_time,omitempty"`
	// 预计完成时间
	EstimatedTime *time.Time `json:"estimated_time,omitempty"`
}

// ModuleInfo 模块信息
type ModuleInfo struct {
	// 模块ID
	ID string `json:"id"`
	// 模块名称
	Name string `json:"name"`
	// 模块版本
	Version string `json:"version,omitempty"`
	// 模块状态
	Status string `json:"status,omitempty"`
	// 模块描述
	Description string `json:"description,omitempty"`
	// 模块配置
	Config map[string]any `json:"config,omitempty"`
}

// RuleTestResult 规则测试结果
type RuleTestResult struct {
	// 优先级
	Priority int `json:"priority"`
	// 是否匹配
	Match bool `json:"match"`
	// 规则组名称
	RuleGroupName string `json:"rule_group_name"`
	// 测试标题
	Title string `json:"title"`
	// 副标题
	Subtitle string `json:"subtitle,omitempty"`
	// 匹配的规则
	Rules []string `json:"rules,omitempty"`
	// 不匹配的原因
	Reason string `json:"reason,omitempty"`
	// 媒体信息
	MediaInfo map[string]any `json:"media_info,omitempty"`
}

// SystemVersion 系统版本信息
type SystemVersion struct {
	// 当前版本
	Current string `json:"current"`
	// 最新版本
	Latest string `json:"latest"`
	// 版本列表
	Releases []map[string]any `json:"releases"`
	// 更新日志
	Changelog string `json:"changelog,omitempty"`
	// 是否有更新
	HasUpdate bool `json:"has_update"`
}

// NetworkTestResult 网络测试结果
type NetworkTestResult struct {
	// URL
	URL string `json:"url"`
	// 是否使用代理
	Proxy bool `json:"proxy"`
	// 响应时间(毫秒)
	ResponseTime int64 `json:"response_time"`
	// 状态码
	StatusCode int `json:"status_code,omitempty"`
	// 是否成功
	Success bool `json:"success"`
	// 错误信息
	Error string `json:"error,omitempty"`
	// 响应大小(字节)
	ResponseSize int64 `json:"response_size,omitempty"`
	// 响应头
	Headers map[string]string `json:"headers,omitempty"`
}

// SchedulerInfo 定时任务信息
type SchedulerInfo struct {
	// 任务ID
	ID string `json:"id"`
	// 任务名称
	Name string `json:"name"`
	// 任务描述
	Description string `json:"description,omitempty"`
	// 任务状态
	Status string `json:"status"` // running, stopped, paused
	// 上次执行时间
	LastRun *time.Time `json:"last_run,omitempty"`
	// 下次执行时间
	NextRun *time.Time `json:"next_run,omitempty"`
	// 执行间隔
	Interval string `json:"interval,omitempty"`
	// Cron表达式
	Cron string `json:"cron,omitempty"`
	// 执行次数
	RunCount int `json:"run_count"`
	// 成功次数
	SuccessCount int `json:"success_count"`
	// 失败次数
	FailureCount int `json:"failure_count"`
}

// SystemInfo 系统信息
type SystemInfo struct {
	// 系统版本
	Version string `json:"version"`
	// 构建时间
	BuildTime string `json:"build_time"`
	// Go版本
	GoVersion string `json:"go_version"`
	// 平台
	Platform string `json:"platform"`
	// 环境
	Environment map[string]any `json:"environment"`
	// 认证版本
	AuthVersion string `json:"auth_version,omitempty"`
	// 索引器版本
	IndexerVersion string `json:"indexer_version,omitempty"`
	// 前端版本
	FrontendVersion string `json:"frontend_version,omitempty"`
	// 系统资源
	SystemResources map[string]any `json:"system_resources,omitempty"`
}

// SystemMessage 系统消息
type SystemMessage struct {
	// 消息ID
	ID string `json:"id,omitempty"`
	// 消息类型
	Type string `json:"type"`
	// 消息级别
	Level string `json:"level"` // info, warning, error
	// 消息标题
	Title string `json:"title,omitempty"`
	// 消息内容
	Content string `json:"content"`
	// 消息来源
	Source string `json:"source,omitempty"`
	// 消息时间
	Timestamp time.Time `json:"timestamp"`
	// 附加数据
	Extra map[string]any `json:"extra,omitempty"`
	// 是否已读
	Read bool `json:"read,omitempty"`
	// 消息角色
	Role string `json:"role,omitempty"`
}

// ImageCacheRequest 图片缓存请求
type ImageCacheRequest struct {
	// 图片URL
	URL string `json:"url"`
	// If-None-Match
	IfNoneMatch string `json:"if_none_match,omitempty"`
	// 是否使用代理
	Proxy bool `json:"proxy,omitempty"`
}

// ImageCacheResponse 图片缓存响应
type ImageCacheResponse struct {
	// 图片数据
	Data []byte `json:"data"`
	// MIME类型
	ContentType string `json:"content_type"`
	// ETag
	ETag string `json:"etag"`
	// 缓存控制
	CacheControl string `json:"cache_control"`
	// 响应码
	StatusCode int `json:"status_code"`
	// 是否来自缓存
	FromCache bool `json:"from_cache"`
}
