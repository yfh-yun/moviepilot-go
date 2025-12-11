package dto

import (
	"path/filepath"

	. "moviepilot-go/internal/models/types"
)

// Event 事件模型
type Event struct {
	// 事件类型
	EventType string `json:"event_type"`
	// 事件数据
	EventData map[string]any `json:"event_data,omitempty"`
	// 事件优先级
	Priority int `json:"priority,omitempty"`
}

// BaseEventData 事件数据的基类
type BaseEventData struct{}

// ConfigChangeEventData ConfigChange 事件的数据模型
type ConfigChangeEventData struct {
	BaseEventData
	// 配置项的键
	Key string `json:"key"`
	// 配置项的新值
	Value any `json:"value,omitempty"`
	// 配置项的变更类型，如 'add', 'update', 'delete'
	ChangeType string `json:"change_type,omitempty"`
}

// ChainEventData 链式事件数据的基类
type ChainEventData struct {
	BaseEventData
}

// AuthCredentials AuthVerification 事件的数据模型
type AuthCredentials struct {
	ChainEventData
	// 输入参数
	Username  string `json:"username,omitempty"`
	Password  string `json:"password,omitempty"`
	MFACode   string `json:"mfa_code,omitempty"`
	Code      string `json:"code,omitempty"`
	GrantType string `json:"grant_type"`

	// 输出参数
	Token   string `json:"token,omitempty"`
	Channel string `json:"channel,omitempty"`
	Service string `json:"service,omitempty"`
}

// AuthInterceptCredentials AuthIntercept 事件的数据模型
type AuthInterceptCredentials struct {
	ChainEventData
	// 输入参数
	Username string `json:"username"`
	Channel  string `json:"channel"`
	Service  string `json:"service"`
	Status   string `json:"status"`
	Token    string `json:"token,omitempty"`

	// 输出参数
	Source string `json:"source,omitempty"`
	Cancel bool   `json:"cancel,omitempty"`
}

// CommandRegisterEventData CommandRegister 事件的数据模型
type CommandRegisterEventData struct {
	ChainEventData
	// 输入参数
	Commands map[string]map[string]any `json:"commands"`
	Origin   string                    `json:"origin"`
	Service  string                    `json:"service,omitempty"`

	// 输出参数
	Cancel bool   `json:"cancel,omitempty"`
	Source string `json:"source,omitempty"`
}

// TransferRenameEventData TransferRename 事件的数据模型
type TransferRenameEventData struct {
	ChainEventData
	// 输入参数
	TemplateString string         `json:"template_string"`
	RenameDict     map[string]any `json:"rename_dict"`
	Path           string         `json:"path,omitempty"`
	RenderStr      string         `json:"render_str"`

	// 输出参数
	Updated    bool   `json:"updated,omitempty"`
	UpdatedStr string `json:"updated_str,omitempty"`
	Source     string `json:"source,omitempty"`
}

// ResourceSelectionEventData ResourceSelection 事件的数据模型
type ResourceSelectionEventData struct {
	BaseEventData
	// 输入参数
	Contexts   []*Context `json:"contexts,omitempty"`
	Downloader string     `json:"downloader,omitempty"`
	Origin     string     `json:"origin,omitempty"`

	// 输出参数
	Updated         bool       `json:"updated,omitempty"`
	UpdatedContexts []*Context `json:"updated_contexts,omitempty"`
	Source          string     `json:"source,omitempty"`
}

// ResourceDownloadEventData ResourceDownload 事件的数据模型
type ResourceDownloadEventData struct {
	ChainEventData
	// 输入参数
	Context    *Context       `json:"context,omitempty"`
	Episodes   []int          `json:"episodes,omitempty"`
	Channel    MessageChannel `json:"channel,omitempty"`
	Origin     string         `json:"origin,omitempty"`
	Downloader string         `json:"downloader,omitempty"`
	Options    map[string]any `json:"options,omitempty"`

	// 输出参数
	Cancel bool   `json:"cancel,omitempty"`
	Source string `json:"source,omitempty"`
	Reason string `json:"reason,omitempty"`
}

// TransferInterceptEventData TransferIntercept 事件的数据模型
type TransferInterceptEventData struct {
	ChainEventData
	// 输入参数
	FileItem      *FileItem      `json:"fileitem"`
	MediaInfo     *MediaInfo     `json:"mediainfo"`
	TargetStorage string         `json:"target_storage"`
	TargetPath    string         `json:"target_path"`
	TransferType  string         `json:"transfer_type"`
	Options       map[string]any `json:"options,omitempty"`

	// 输出参数
	Cancel bool   `json:"cancel,omitempty"`
	Source string `json:"source,omitempty"`
	Reason string `json:"reason,omitempty"`
}

// DiscoverMediaSource 探索媒体数据源的基类
type DiscoverMediaSource struct {
	// 数据源名称
	Name string `json:"name"`
	// 媒体ID的前缀，不含:
	MediaIDPrefix string `json:"mediaid_prefix"`
	// 媒体数据源API地址
	APIPath string `json:"api_path"`
	// 过滤参数
	FilterParams map[string]any `json:"filter_params,omitempty"`
	// 过滤参数UI配置
	FilterUI []map[string]any `json:"filter_ui,omitempty"`
	// UI依赖关系字典
	Depends map[string][]any `json:"depends,omitempty"`
}

// DiscoverSourceEventData DiscoverSource 事件的数据模型
type DiscoverSourceEventData struct {
	ChainEventData
	// 输出参数
	ExtraSources []*DiscoverMediaSource `json:"extra_sources,omitempty"`
}

// RecommendMediaSource 推荐媒体数据源的基类
type RecommendMediaSource struct {
	// 数据源名称
	Name string `json:"name"`
	// 媒体数据源API地址
	APIPath string `json:"api_path"`
	// 类型
	Type string `json:"type"`
}

// RecommendSourceEventData RecommendSource 事件的数据模型
type RecommendSourceEventData struct {
	ChainEventData
	// 输出参数
	ExtraSources []*RecommendMediaSource `json:"extra_sources,omitempty"`
}

// MediaRecognizeConvertEventData MediaRecognizeConvert 事件的数据模型
type MediaRecognizeConvertEventData struct {
	ChainEventData
	// 输入参数
	MediaID     string `json:"mediaid"`
	ConvertType string `json:"convert_type"`

	// 输出参数
	MediaDict map[string]any `json:"media_dict,omitempty"`
}

// StorageOperSelectionEventData StorageOperSelect 事件的数据模型
type StorageOperSelectionEventData struct {
	ChainEventData
	// 输入参数
	Storage string `json:"storage,omitempty"`

	// 输出参数
	StorageOper any `json:"storage_oper,omitempty"` // 存储操作对象
}

// WorkflowExecutionEventData 工作流执行事件数据
type WorkflowExecutionEventData struct {
	ChainEventData
	// 工作流ID
	WorkflowID int `json:"workflow_id"`
	// 工作流名称
	WorkflowName string `json:"workflow_name"`
	// 触发类型
	TriggerType string `json:"trigger_type"`
	// 执行上下文
	Context map[string]any `json:"context,omitempty"`
}

// NameRecognizeEventData 名称识别事件数据
type NameRecognizeEventData struct {
	ChainEventData
	// 输入参数
	Title string `json:"title"`

	// 输出参数
	MetaInfo *MetaInfo `json:"meta_info,omitempty"`
}

// FileItemExt FileItem扩展方法
type FileItemExt struct {
	*FileItem
}

// Dir 返回文件所在目录
func (f *FileItemExt) Dir() string {
	return filepath.Dir(f.Path)
}

// Base 返回文件名
func (f *FileItemExt) Base() string {
	return filepath.Base(f.Path)
}
