// Package actions 提供动作系统的接口定义
package actions

import (
	"context"

	"moviepilot-go/internal/business/services/actions/types"
)

// Action 定义所有动作的通用接口
type Action interface {
	// GetActionID 获取动作唯一标识
	GetActionID() string
	// Name 获取动作名称
	Name() string
	// Description 获取动作描述
	Description() string
	// Data 获取动作元数据
	Data() map[string]interface{}
	// Execute 执行动作
	Execute(ctx context.Context, workflowID int64, params map[string]interface{}, actionContext *types.ActionContext) (*types.ActionContext, error)
	// SetError 设置错误信息
	SetError(errorMsg string)
	// GetError 获取错误信息
	GetError() string
}

// ActionManager 动作管理器接口
type ActionManager interface {
	// RegisterAction 注册动作
	RegisterAction(action Action)
	// UnregisterAction 注销动作
	UnregisterAction(actionID string)
	// GetAction 获取动作实例
	GetAction(actionID string) (Action, error)
	// ListActions 列出所有动作
	ListActions() []Action
	// ExecuteAction 执行指定动作
	ExecuteAction(ctx context.Context, workflowID int64, actionID string, params map[string]interface{}, actionContext *types.ActionContext) (*types.ActionContext, error)
}

// DownloadManager 下载管理器接口
type DownloadManager interface {
	// AddDownload 添加下载任务
	AddDownload(ctx context.Context, workflowID int64, params *AddDownloadParams, torrents []*types.Torrent) ([]*AddDownloadResult, error)
	// FetchDownloads 获取下载列表
	FetchDownloads(ctx context.Context, params map[string]interface{}) ([]*types.Download, error)
	// GetDownloadStatus 获取下载状态
	GetDownloadStatus(ctx context.Context, downloadID string) (*types.DownloadStatus, error)
	// CancelDownload 取消下载
	CancelDownload(ctx context.Context, downloadID string) error
}

// MediaFetcher 媒体信息获取器接口
type MediaFetcher interface {
	// FetchMediaByID 根据ID获取媒体信息
	FetchMediaByID(ctx context.Context, mediaType string, mediaID string) (*types.Media, error)
	// SearchMedia 搜索媒体
	SearchMedia(ctx context.Context, keyword string, mediaType string, page int, limit int) ([]*types.Media, int64, error)
	// FetchMediaDetails 获取媒体详情
	FetchMediaDetails(ctx context.Context, mediaType string, mediaID string, language string) (*types.MediaDetails, error)
	// FetchMediaCredits 获取媒体演职人员信息
	FetchMediaCredits(ctx context.Context, mediaType string, mediaID string) (*types.MediaCredits, error)
}

// SubscribeManager 订阅管理器接口
type SubscribeManager interface {
	// AddSubscribe 添加订阅
	AddSubscribe(ctx context.Context, params *AddSubscribeParams) (*types.Subscribe, error)
	// GetSubscribe 获取订阅信息
	GetSubscribe(ctx context.Context, subscribeID string) (*types.Subscribe, error)
	// UpdateSubscribe 更新订阅
	UpdateSubscribe(ctx context.Context, subscribeID string, params map[string]interface{}) (*types.Subscribe, error)
	// DeleteSubscribe 删除订阅
	DeleteSubscribe(ctx context.Context, subscribeID string) error
	// ListSubscribes 列出订阅
	ListSubscribes(ctx context.Context, params map[string]interface{}) ([]*types.Subscribe, int64, error)
}

// RSSFetcher RSS获取器接口
type RSSFetcher interface {
	// FetchRSS 获取RSS内容
	FetchRSS(ctx context.Context, url string) ([]*types.RSSItem, error)
	// ParseRSS 解析RSS内容
	ParseRSS(content []byte) ([]*types.RSSItem, error)
}

// TorrentFetcher 种子获取器接口
type TorrentFetcher interface {
	// SearchTorrents 搜索种子
	SearchTorrents(ctx context.Context, keyword string, params map[string]interface{}) ([]*types.Torrent, error)
	// GetTorrentDetail 获取种子详情
	GetTorrentDetail(ctx context.Context, torrentID string, source string) (*types.TorrentDetail, error)
}

// MediaFilter 媒体过滤器接口
type MediaFilter interface {
	// FilterMedias 根据规则过滤媒体
	FilterMedias(ctx context.Context, medias []*types.Media, rules map[string]interface{}) ([]*types.Media, error)
	// EvaluateRule 评估单个过滤规则
	EvaluateRule(media *types.Media, rule map[string]interface{}) (bool, error)
}

// TorrentFilter 种子过滤器接口
type TorrentFilter interface {
	// FilterTorrents 根据规则过滤种子
	FilterTorrents(ctx context.Context, torrents []*types.Torrent, rules map[string]interface{}) ([]*types.Torrent, error)
	// EvaluateRule 评估单个过滤规则
	EvaluateRule(torrent *types.Torrent, rule map[string]interface{}) (bool, error)
}

// PluginInvoker 插件调用器接口
type PluginInvoker interface {
	// InvokePlugin 调用插件
	InvokePlugin(ctx context.Context, pluginType string, pluginID string, method string, params map[string]interface{}) (interface{}, error)
	// ListAvailablePlugins 列出可用插件
	ListAvailablePlugins(ctx context.Context, pluginType string) ([]*types.PluginInfo, error)
}

// FileScanner 文件扫描器接口
type FileScanner interface {
	// ScanFiles 扫描文件
	ScanFiles(ctx context.Context, paths []string, recursive bool, filter map[string]interface{}) ([]*types.FileInfo, error)
	// ScanDirectory 扫描目录
	ScanDirectory(ctx context.Context, path string, recursive bool) ([]*types.FileInfo, error)
}

// FileScraper 文件刮削器接口
type FileScraper interface {
	// ScrapeFile 刮削文件元数据
	ScrapeFile(ctx context.Context, filePath string) (*types.FileMetadata, error)
	// ParseNFO 解析NFO文件
	ParseNFO(ctx context.Context, filePath string) (*types.NFOMetadata, error)
}

// EventSender 事件发送器接口
type EventSender interface {
	// SendEvent 发送事件
	SendEvent(ctx context.Context, eventType string, eventData map[string]interface{}) error
	// SendSystemEvent 发送系统事件
	SendSystemEvent(ctx context.Context, eventType string, message string, level string) error
}

// MessageSender 消息发送器接口
type MessageSender interface {
	// SendMessage 发送消息
	SendMessage(ctx context.Context, params *SendMessageParams) error
	// SendNotification 发送通知
	SendNotification(ctx context.Context, title string, content string, level string, receivers []string) error
}
