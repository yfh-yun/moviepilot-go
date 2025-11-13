package models

import (
	"time"
)

// Event 事件模型
type Event struct {
	// 事件类型
	EventType string `json:"event_type"`
	
	// 事件数据
	EventData map[string]interface{} `json:"event_data,omitempty"`
	
	// 事件优先�?	Priority int `json:"priority,omitempty"`
}

// BaseEventData 事件数据的基类，所有具体事件数据类应继承自此类
type BaseEventData struct {
	// 可以根据需要添加通用字段
}

// ConfigChangeEventData ConfigChange 事件的数据模�?type ConfigChangeEventData struct {
	BaseEventData
	
	// 配置项的�?	Key string `json:"key"`
	
	// 配置项的新�?	Value interface{} `json:"value,omitempty"`
	
	// 配置项的变更类型，如 'add', 'update', 'delete'
	ChangeType string `json:"change_type,omitempty"`
}

// ChainEventData 链式事件数据的基类，所有具体事件数据类应继承自此类
type ChainEventData struct {
	BaseEventData
	// 可以根据需要添加通用字段
}

// AuthCredentials AuthVerification 事件的数据模�?type AuthCredentials struct {
	ChainEventData
	
	// 输入参数
	// 用户名，适用�?"password" grant_type
	Username string `json:"username,omitempty"`
	
	// 用户密码，适用�?"password" grant_type
	Password string `json:"password,omitempty"`
	
	// 一次性密码，目前仅适用�?"password" 认证类型
	MfaCode string `json:"mfa_code,omitempty"`
	
	// 授权码，适用�?"authorization_code" grant_type
	Code string `json:"code,omitempty"`
	
	// 认证类型，如 "password", "authorization_code", "client_credentials"
	GrantType string `json:"grant_type"`
	
	// 权限范围，如 ["read", "write"]
	// Scope []string `json:"scope,omitempty"`
	
	// 输出参数
	// grant_type �?authorization_code 时，输出参数包括 username、token、channel、service
	// 认证令牌
	Token string `json:"token,omitempty"`
	
	// 认证渠道
	Channel string `json:"channel,omitempty"`
	
	// 服务名称
	Service string `json:"service,omitempty"`
}

// AuthInterceptCredentials AuthIntercept 事件的数据模�?type AuthInterceptCredentials struct {
	ChainEventData
	
	// 输入参数
	// 用户�?	Username string `json:"username,omitempty"`
	
	// 认证渠道
	Channel string `json:"channel"`
	
	// 服务名称
	Service string `json:"service"`
	
	// 认证状态，"triggered" �?"completed" 两个状�?	Status string `json:"status"`
	
	// 认证令牌
	Token string `json:"token,omitempty"`
	
	// 输出参数
	// 拦截源，默认值为 "未知拦截�?
	Source string `json:"source,omitempty"`
	
	// 是否取消认证，默认值为 False
	Cancel bool `json:"cancel,omitempty"`
}

// CommandRegisterEventData CommandRegister 事件的数据模�?type CommandRegisterEventData struct {
	ChainEventData
	
	// 输入参数
	// 菜单命令
	Commands map[string]map[string]interface{} `json:"commands"`
	
	// 事件源，可以�?Chain 或具体的模块名称
	Origin string `json:"origin"`
	
	// 服务名称
	Service string `json:"service,omitempty"`
	
	// 输出参数
	// 是否取消注册，默认值为 False
	Cancel bool `json:"cancel,omitempty"`
	
	// 拦截源，默认值为 "未知拦截�?
	Source string `json:"source,omitempty"`
}

// TransferRenameEventData TransferRename 事件的数据模�?type TransferRenameEventData struct {
	ChainEventData
	
	// 输入参数
	// Jinja2 模板字符�?	TemplateString string `json:"template_string"`
	
	// 渲染上下�?	RenameDict map[string]interface{} `json:"rename_dict"`
	
	// 当前文件的目标路�?	Path string `json:"path,omitempty"`
	
	// 渲染生成的字符串
	RenderStr string `json:"render_str"`
	
	// 输出参数
	// 是否已更新，默认值为 False
	Updated bool `json:"updated,omitempty"`
	
	// 更新后的字符�?	UpdatedStr string `json:"updated_str,omitempty"`
	
	// 拦截源，默认值为 "未知拦截�?
	Source string `json:"source,omitempty"`
}

// ResourceSelectionEventData ResourceSelection 事件的数据模�?type ResourceSelectionEventData struct {
	BaseEventData
	
	// 输入参数
	// 待选择的资源上下文列表
	Contexts interface{} `json:"contexts,omitempty"`
	
	// 下载�?	Downloader string `json:"downloader,omitempty"`
	
	// 来源
	Origin string `json:"origin,omitempty"`
	
	// 输出参数
	// 是否已更新，默认值为 False
	Updated bool `json:"updated,omitempty"`
	
	// 已更新的资源上下文列表，默认值为 None
	UpdatedContexts []interface{} `json:"updated_contexts,omitempty"`
	
	// 拦截源，默认值为 "未知拦截�?
	Source string `json:"source,omitempty"`
}

// ResourceDownloadEventData ResourceDownload 事件的数据模�?type ResourceDownloadEventData struct {
	ChainEventData
	
	// 输入参数
	// 当前资源上下�?	Context interface{} `json:"context,omitempty"`
	
	// 需要下载的集数
	Episodes []int `json:"episodes,omitempty"`
	
	// 通知渠道
	Channel string `json:"channel,omitempty"`
	
	// 来源（消息通知、Subscribe、Manual等）
	Origin string `json:"origin,omitempty"`
	
	// 下载�?	Downloader string `json:"downloader,omitempty"`
	
	// 其他参数
	Options map[string]interface{} `json:"options,omitempty"`
	
	// 输出参数
	// 是否取消下载，默认值为 False
	Cancel bool `json:"cancel,omitempty"`
	
	// 拦截源，默认值为 "未知拦截�?
	Source string `json:"source,omitempty"`
	
	// 拦截原因，描述拦截的具体原因
	Reason string `json:"reason,omitempty"`
}

// TransferInterceptEventData TransferIntercept 事件的数据模�?type TransferInterceptEventData struct {
	ChainEventData
	
	// 输入参数
	// 源文�?	FileItem FileItem `json:"fileitem"`
	
	// 媒体信息
	MediaInfo interface{} `json:"mediainfo"`
	
	// 目标存储
	TargetStorage string `json:"target_storage"`
	
	// 目标路径
	TargetPath string `json:"target_path"`
	
	// 整理方式（copy、move、link、softlink等）
	TransferType string `json:"transfer_type"`
	
	// 其他参数
	Options map[string]interface{} `json:"options,omitempty"`
	
	// 输出参数
	// 是否取消整理，默认值为 False
	Cancel bool `json:"cancel,omitempty"`
	
	// 拦截源，默认值为 "未知拦截�?
	Source string `json:"source,omitempty"`
	
	// 拦截原因，描述拦截的具体原因
	Reason string `json:"reason,omitempty"`
}

// DiscoverMediaSource 探索媒体数据源的基类
type DiscoverMediaSource struct {
	// 数据源名�?	Name string `json:"name"`
	
	// 媒体ID的前缀，不�?
	MediaidPrefix string `json:"mediaid_prefix"`
	
	// 媒体数据源API地址
	APIPath string `json:"api_path"`
	
	// 过滤参数
	FilterParams map[string]interface{} `json:"filter_params,omitempty"`
	
	// 过滤参数UI配置
	FilterUI []map[string]interface{} `json:"filter_ui,omitempty"`
	
	// UI依赖关系字典
	Depends map[string][]string `json:"depends,omitempty"`
}

// DiscoverSourceEventData DiscoverSource 事件的数据模�?type DiscoverSourceEventData struct {
	ChainEventData
	
	// 输出参数
	// 额外媒体数据�?	ExtraSources []DiscoverMediaSource `json:"extra_sources,omitempty"`
}

// RecommendMediaSource 推荐媒体数据源的基类
type RecommendMediaSource struct {
	// 数据源名�?	Name string `json:"name"`
	
	// 媒体数据源API地址
	APIPath string `json:"api_path"`
	
	// 类型
	Type string `json:"type"`
}

// RecommendSourceEventData RecommendSource 事件的数据模�?type RecommendSourceEventData struct {
	ChainEventData
	
	// 输出参数
	// 额外媒体数据�?	ExtraSources []RecommendMediaSource `json:"extra_sources,omitempty"`
}

// MediaRecognizeConvertEventData MediaRecognizeConvert 事件的数据模�?type MediaRecognizeConvertEventData struct {
	ChainEventData
	
	// 输入参数
	// 媒体ID，格式为`前缀:ID值`，如 tmdb:12345、douban:1234567
	Mediaid string `json:"mediaid"`
	
	// 转换类型 仅支持：themoviedb/douban，需要转换为对应的媒体数据并返回
	ConvertType string `json:"convert_type"`
	
	// 输出参数
	// TheMovieDb/豆瓣的媒体数�?	MediaDict map[string]interface{} `json:"media_dict,omitempty"`
}

// StorageOperSelectionEventData StorageOperSelect 事件的数据模�?type StorageOperSelectionEventData struct {
	ChainEventData
	
	// 输入参数
	// 存储类型
	Storage string `json:"storage,omitempty"`
	
	// 输出参数
	// 存储操作对象
	StorageOper interface{} `json:"storage_oper,omitempty"`
}

// MessageChannel 消息渠道枚举
const (
	MessageChannelWechat        = "微信"
	MessageChannelTelegram      = "Telegram"
	MessageChannelSlack         = "Slack"
	MessageChannelSynologyChat  = "SynologyChat"
	MessageChannelVoceChat      = "VoceChat"
	MessageChannelWeb           = "Web"
	MessageChannelWebPush       = "WebPush"
)
