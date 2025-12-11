package actions

// Services 定义动作系统所需的服务集合
type Services struct {
	DownloadService     DownloadService     // 下载服务
	EventService        EventService        // 事件服务
	PluginService       PluginService       // 插件服务
	RSSService          RSSService          // RSS服务
	MediaService        MediaService        // 媒体服务
	NotificationService NotificationService // 通知服务
	SubscribeService    SubscribeService    // 订阅服务
	RecommendService    RecommendService    // 推荐服务
	SearchService       SearchService       // 搜索服务
}

// DownloadService 定义下载服务接口
type DownloadService interface {
	// AddDownload 添加下载任务
	AddDownload(ctx any, params any) (string, error)

	// GetDownloads 获取下载列表
	GetDownloads(ctx any, params any) ([]any, error)

	// GetDownload 获取单个下载详情
	GetDownload(ctx any, downloadID string) (any, error)

	// ListTorrents 获取种子列表（根据哈希值）
	ListTorrents(ctx any, hashs []string) ([]any, error)

	// PauseDownload 暂停下载
	PauseDownload(ctx any, downloadID string) error

	// ResumeDownload 恢复下载
	ResumeDownload(ctx any, downloadID string) error

	// CancelDownload 取消下载
	CancelDownload(ctx any, downloadID string) error

	// DeleteDownload 删除下载
	DeleteDownload(ctx any, downloadID string, deleteFiles bool) error
}

// EventService 定义事件服务接口
type EventService interface {
	// PublishEvent 发布事件
	PublishEvent(ctx any, event any) (string, error)

	// GetEvent 获取单个事件
	GetEvent(ctx any, eventID string) (any, error)

	// GetEvents 获取事件列表
	GetEvents(ctx any, params any) ([]any, error)

	// SubscribeEvent 订阅事件
	SubscribeEvent(ctx any, eventName string, handler any) error

	// UnsubscribeEvent 取消订阅事件
	UnsubscribeEvent(ctx any, eventName string, handlerID string) error
}

// PluginService 定义插件服务接口
type PluginService interface {
	// GetPlugins 获取插件列表
	GetPlugins(ctx any, params any) ([]any, error)

	// GetPlugin 获取单个插件信息
	GetPlugin(ctx any, pluginID string) (any, error)

	// InvokePlugin 调用插件方法
	InvokePlugin(ctx any, pluginID string, method string, params any) (any, error)

	// EnablePlugin 启用插件
	EnablePlugin(ctx any, pluginID string) error

	// DisablePlugin 禁用插件
	DisablePlugin(ctx any, pluginID string) error

	// InstallPlugin 安装插件
	InstallPlugin(ctx any, pluginURL string) (string, error)

	// UninstallPlugin 卸载插件
	UninstallPlugin(ctx any, pluginID string) error
}

// RSSService 定义RSS服务接口
type RSSService interface {
	// FetchRSS 抓取单个RSS源
	FetchRSS(ctx any, url string) (any, error)

	// FetchRSSBatch 批量抓取RSS源
	FetchRSSBatch(ctx any, urls []string) ([]any, error)

	// ParseRSSContent 解析RSS内容
	ParseRSSContent(ctx any, content []byte) (any, error)

	// SubscribeRSS 订阅RSS源
	SubscribeRSS(ctx any, url string, options any) (string, error)

	// ParseRSS 解析RSS源
	ParseRSS(ctx any, url string, options map[string]any) ([]map[string]any, error)
}

// MediaService 定义媒体服务接口
type MediaService interface {
	// GetMedias 获取媒体列表
	GetMedias(ctx any, params any) ([]any, error)

	// GetMediaDetails 获取媒体详情
	GetMediaDetails(ctx any, mediaID string, params any) (any, error)

	// SearchMedias 搜索媒体
	SearchMedias(ctx any, params any) ([]any, error)

	// GetMediaLibraries 获取媒体库列表
	GetMediaLibraries(ctx any, serverName string) ([]any, error)
}

// NotificationService 定义通知服务接口
type NotificationService interface {
	// SendNotification 发送通知
	SendNotification(ctx any, notification any) (string, error)

	// GetChannels 获取支持的通知渠道列表
	GetChannels(ctx any) ([]any, error)

	// ValidateChannel 验证通知渠道配置
	ValidateChannel(ctx any, channel string, config any) error
}

// SubscribeService 定义订阅服务接口
type SubscribeService interface {
	// Exists 检查订阅是否已存在
	Exists(ctx any, mediainfo map[string]any, meta map[string]any) (bool, error)

	// CreateSubscribe 创建订阅
	CreateSubscribe(ctx any, subscribe map[string]any) (string, string, error)

	// GetSubscribe 获取单个订阅
	GetSubscribe(ctx any, subscribeID string) (map[string]any, error)

	// GetSubscribes 获取订阅列表
	GetSubscribes(ctx any, params any) ([]any, error)

	// UpdateSubscribe 更新订阅
	UpdateSubscribe(ctx any, subscribeID string, updates any) error

	// DeleteSubscribe 删除订阅
	DeleteSubscribe(ctx any, subscribeID string) error

	// ActivateSubscribe 激活订阅
	ActivateSubscribe(ctx any, subscribeID string) error

	// PauseSubscribe 暂停订阅
	PauseSubscribe(ctx any, subscribeID string) error
}

// RecommendService 定义推荐服务接口
type RecommendService interface {
	// GetMediasFromSource 从指定数据源获取媒体数据
	GetMediasFromSource(ctx any, apiPath string) ([]map[string]any, error)

	// GetRecommendSources 获取推荐数据源列表
	GetRecommendSources(ctx any) ([]map[string]any, error)
}

// SearchService 定义搜索服务接口
type SearchService interface {
	// SearchByTitle 按关键字搜索
	SearchByTitle(ctx any, title string, sites []int) ([]map[string]any, error)

	// SearchByID 按ID搜索
	SearchByID(ctx any, params map[string]any) ([]map[string]any, error)

	// RecognizeMedia 识别媒体信息
	RecognizeMedia(ctx any, meta map[string]any) (map[string]any, error)
}

// NewServices 创建新的服务集合实例
func NewServices() *Services {
	return &Services{}
}
