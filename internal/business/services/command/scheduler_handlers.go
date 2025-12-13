package command

import (
	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// CookieCloudSchedulerHandler CookieCloud同步调度器命令
type CookieCloudSchedulerHandler struct {
	logger *zap.Logger
}

// NewCookieCloudSchedulerHandler 创建CookieCloud同步调度器命令
func NewCookieCloudSchedulerHandler() *CookieCloudSchedulerHandler {
	return &CookieCloudSchedulerHandler{
		logger: logger.GetLogger(),
	}
}

// ID 调度器任务ID
func (h *CookieCloudSchedulerHandler) ID() string {
	return "cookiecloud"
}

// Name 命令名称
func (h *CookieCloudSchedulerHandler) Name() string {
	return "cookiecloud"
}

// Description 命令描述
func (h *CookieCloudSchedulerHandler) Description() string {
	return "同步站点"
}

// Category 命令分类
func (h *CookieCloudSchedulerHandler) Category() string {
	return "站点"
}

// MediaServerSyncSchedulerHandler 媒体服务器同步调度器命令
type MediaServerSyncSchedulerHandler struct {
	logger *zap.Logger
}

// NewMediaServerSyncSchedulerHandler 创建媒体服务器同步调度器命令
func NewMediaServerSyncSchedulerHandler() *MediaServerSyncSchedulerHandler {
	return &MediaServerSyncSchedulerHandler{
		logger: logger.GetLogger(),
	}
}

// ID 调度器任务ID
func (h *MediaServerSyncSchedulerHandler) ID() string {
	return "mediaserver_sync"
}

// Name 命令名称
func (h *MediaServerSyncSchedulerHandler) Name() string {
	return "mediaserver_sync"
}

// Description 命令描述
func (h *MediaServerSyncSchedulerHandler) Description() string {
	return "同步媒体服务器"
}

// Category 命令分类
func (h *MediaServerSyncSchedulerHandler) Category() string {
	return "管理"
}

// SubscribeRefreshSchedulerHandler 订阅刷新调度器命令
type SubscribeRefreshSchedulerHandler struct {
	logger *zap.Logger
}

// NewSubscribeRefreshSchedulerHandler 创建订阅刷新调度器命令
func NewSubscribeRefreshSchedulerHandler() *SubscribeRefreshSchedulerHandler {
	return &SubscribeRefreshSchedulerHandler{
		logger: logger.GetLogger(),
	}
}

// ID 调度器任务ID
func (h *SubscribeRefreshSchedulerHandler) ID() string {
	return "subscribe_refresh"
}

// Name 命令名称
func (h *SubscribeRefreshSchedulerHandler) Name() string {
	return "subscribe_refresh"
}

// Description 命令描述
func (h *SubscribeRefreshSchedulerHandler) Description() string {
	return "刷新订阅"
}

// Category 命令分类
func (h *SubscribeRefreshSchedulerHandler) Category() string {
	return "订阅"
}

// SubscribeSearchSchedulerHandler 订阅搜索调度器命令
type SubscribeSearchSchedulerHandler struct {
	logger *zap.Logger
}

// NewSubscribeSearchSchedulerHandler 创建订阅搜索调度器命令
func NewSubscribeSearchSchedulerHandler() *SubscribeSearchSchedulerHandler {
	return &SubscribeSearchSchedulerHandler{
		logger: logger.GetLogger(),
	}
}

// ID 调度器任务ID
func (h *SubscribeSearchSchedulerHandler) ID() string {
	return "subscribe_search"
}

// Name 命令名称
func (h *SubscribeSearchSchedulerHandler) Name() string {
	return "subscribe_search"
}

// Description 命令描述
func (h *SubscribeSearchSchedulerHandler) Description() string {
	return "搜索订阅"
}

// Category 命令分类
func (h *SubscribeSearchSchedulerHandler) Category() string {
	return "订阅"
}

// SubscribeTmdbSchedulerHandler 订阅元数据更新调度器命令
type SubscribeTmdbSchedulerHandler struct {
	logger *zap.Logger
}

// NewSubscribeTmdbSchedulerHandler 创建订阅元数据更新调度器命令
func NewSubscribeTmdbSchedulerHandler() *SubscribeTmdbSchedulerHandler {
	return &SubscribeTmdbSchedulerHandler{
		logger: logger.GetLogger(),
	}
}

// ID 调度器任务ID
func (h *SubscribeTmdbSchedulerHandler) ID() string {
	return "subscribe_tmdb"
}

// Name 命令名称
func (h *SubscribeTmdbSchedulerHandler) Name() string {
	return "subscribe_tmdb"
}

// Description 命令描述
func (h *SubscribeTmdbSchedulerHandler) Description() string {
	return "订阅元数据更新"
}

// Category 命令分类
func (h *SubscribeTmdbSchedulerHandler) Category() string {
	return "订阅"
}

// TransferSchedulerHandler 下载文件整理调度器命令
type TransferSchedulerHandler struct {
	logger *zap.Logger
}

// NewTransferSchedulerHandler 创建下载文件整理调度器命令
func NewTransferSchedulerHandler() *TransferSchedulerHandler {
	return &TransferSchedulerHandler{
		logger: logger.GetLogger(),
	}
}

// ID 调度器任务ID
func (h *TransferSchedulerHandler) ID() string {
	return "transfer"
}

// Name 命令名称
func (h *TransferSchedulerHandler) Name() string {
	return "transfer"
}

// Description 命令描述
func (h *TransferSchedulerHandler) Description() string {
	return "下载文件整理"
}

// Category 命令分类
func (h *TransferSchedulerHandler) Category() string {
	return "管理"
}
