package actions

import (
	"time"
)

// TorrentStatus 种子状态枚举
type TorrentStatus string

const (
	// TorrentStatusDownloading 下载中
	TorrentStatusDownloading TorrentStatus = "downloading"
	// TorrentStatusSeeding 做种中
	TorrentStatusSeeding TorrentStatus = "seeding"
	// TorrentStatusPaused 已暂停
	TorrentStatusPaused TorrentStatus = "paused"
	// TorrentStatusCompleted 已完成
	TorrentStatusCompleted TorrentStatus = "completed"
	// TorrentStatusError 错误
	TorrentStatusError TorrentStatus = "error"
	// TorrentStatusChecking 校验中
	TorrentStatusChecking TorrentStatus = "checking"
	// TorrentStatusQueued 队列中
	TorrentStatusQueued TorrentStatus = "queued"
	// TorrentStatusUnknown 未知状态
	TorrentStatusUnknown TorrentStatus = "unknown"
)

// TorrentType 种子类型枚举
type TorrentType string

const (
	// TorrentTypeMovie 电影
	TorrentTypeMovie TorrentType = "movie"
	// TorrentTypeSeries 剧集
	TorrentTypeSeries TorrentType = "series"
	// TorrentTypeAnimation 动画
	TorrentTypeAnimation TorrentType = "animation"
	// TorrentTypeDocumentary 纪录片
	TorrentTypeDocumentary TorrentType = "documentary"
	// TorrentTypeMusic 音乐
	TorrentTypeMusic TorrentType = "music"
	// TorrentTypeApplication 应用
	TorrentTypeApplication TorrentType = "application"
	// TorrentTypeGame 游戏
	TorrentTypeGame TorrentType = "game"
	// TorrentTypeOther 其他
	TorrentTypeOther TorrentType = "other"
)

// TorrentPriority 种子优先级枚举
type TorrentPriority string

const (
	// TorrentPriorityHigh 高优先级
	TorrentPriorityHigh TorrentPriority = "high"
	// TorrentPriorityNormal 普通优先级
	TorrentPriorityNormal TorrentPriority = "normal"
	// TorrentPriorityLow 低优先级
	TorrentPriorityLow TorrentPriority = "low"
)

// TorrentSortField 种子排序字段枚举
type TorrentSortField string

const (
	// TorrentSortFieldName 按名称排序
	TorrentSortFieldName TorrentSortField = "name"
	// TorrentSortFieldSize 按大小排序
	TorrentSortFieldSize TorrentSortField = "size"
	// TorrentSortFieldProgress 按进度排序
	TorrentSortFieldProgress TorrentSortField = "progress"
	// TorrentSortFieldAdded 按添加时间排序
	TorrentSortFieldAdded TorrentSortField = "added"
	// TorrentSortFieldCompleted 按完成时间排序
	TorrentSortFieldCompleted TorrentSortField = "completed"
	// TorrentSortFieldRatio 按分享率排序
	TorrentSortFieldRatio TorrentSortField = "ratio"
	// TorrentSortFieldSeeds 按做种数排序
	TorrentSortFieldSeeds TorrentSortField = "seeds"
	// TorrentSortFieldPeers 按下载数排序
	TorrentSortFieldPeers TorrentSortField = "peers"
	// TorrentSortFieldStatus 按状态排序
	TorrentSortFieldStatus TorrentSortField = "status"
)

// SortOrder 排序顺序枚举
type SortOrder string

const (
	// SortOrderAsc 升序
	SortOrderAsc SortOrder = "asc"
	// SortOrderDesc 降序
	SortOrderDesc SortOrder = "desc"
)

// DownloaderType 下载器类型枚举
type DownloaderType string

const (
	// DownloaderTypeQBitTorrent qBittorrent下载器
	DownloaderTypeQBitTorrent DownloaderType = "qbittorrent"
	// DownloaderTypeDeluge Deluge下载器
	DownloaderTypeDeluge DownloaderType = "deluge"
	// DownloaderTypeTransmission Transmission下载器
	DownloaderTypeTransmission DownloaderType = "transmission"
	// DownloaderTypeAria2 Aria2下载器
	DownloaderTypeAria2 DownloaderType = "aria2"
	// DownloaderTypeuTorrent uTorrent下载器
	DownloaderTypeuTorrent DownloaderType = "utorrent"
)

// TorrentItem 种子项结构
type TorrentItem struct {
	// ID 种子唯一标识
	ID string `json:"id"`
	// Hash 种子哈希值
	Hash string `json:"hash"`
	// Name 种子名称
	Name string `json:"name"`
	// Status 种子状态
	Status TorrentStatus `json:"status"`
	// Type 种子类型
	Type TorrentType `json:"type"`
	// Priority 优先级
	Priority TorrentPriority `json:"priority"`
	// Size 总大小（字节）
	Size int64 `json:"size"`
	// Completed 已完成大小（字节）
	Completed int64 `json:"completed"`
	// Progress 下载进度（0-100）
	Progress float64 `json:"progress"`
	// Uploaded 已上传大小（字节）
	Uploaded int64 `json:"uploaded"`
	// Downloaded 已下载大小（字节）
	Downloaded int64 `json:"downloaded"`
	// Ratio 分享率
	Ratio float64 `json:"ratio"`
	// ETA 预计完成时间（秒）
	ETA int64 `json:"eta"`
	// DownloadSpeed 下载速度（字节/秒）
	DownloadSpeed int64 `json:"download_speed"`
	// UploadSpeed 上传速度（字节/秒）
	UploadSpeed int64 `json:"upload_speed"`
	// Seeds 连接的做种数
	Seeds int `json:"seeds"`
	// Peers 连接的下载数
	Peers int `json:"peers"`
	// Availability 可用性
	Availability float64 `json:"availability"`
	// Category 分类
	Category string `json:"category"`
	// Tags 标签
	Tags []string `json:"tags"`
	// SavePath 保存路径
	SavePath string `json:"save_path"`
	// AddedAt 添加时间
	AddedAt time.Time `json:"added_at"`
	// CompletedAt 完成时间
	CompletedAt time.Time `json:"completed_at"`
	// LastActivityAt 最后活动时间
	LastActivityAt time.Time `json:"last_activity_at"`
	// Trackers 追踪器列表
	Trackers []*TrackerStatus `json:"trackers"`
	// Files 文件列表
	Files []*TorrentFile `json:"files"`
	// DownloaderType 下载器类型
	DownloaderType DownloaderType `json:"downloader_type"`
	// DownloaderID 下载器ID
	DownloaderID string `json:"downloader_id"`
	// Metadata 额外元数据
	Metadata map[string]interface{} `json:"metadata"`
}

// TorrentFile 种子文件结构
type TorrentFile struct {
	// ID 文件唯一标识
	ID int `json:"id"`
	// Name 文件名
	Name string `json:"name"`
	// Path 文件路径
	Path string `json:"path"`
	// Size 文件大小（字节）
	Size int64 `json:"size"`
	// Completed 已完成大小（字节）
	Completed int64 `json:"completed"`
	// Progress 文件下载进度（0-100）
	Progress float64 `json:"progress"`
	// Priority 文件优先级
	Priority int `json:"priority"`
	// IsSelected 是否被选中
	IsSelected bool `json:"is_selected"`
	// Index 文件索引
	Index int `json:"index"`
	// ContentType 文件类型
	ContentType string `json:"content_type"`
}

// TrackerStatus 追踪器状态结构
type TrackerStatus struct {
	// URL 追踪器URL
	URL string `json:"url"`
	// Status 追踪器状态
	Status string `json:"status"`
	// Tier 追踪器层级
	Tier int `json:"tier"`
	// Seeds 做种数
	Seeds int `json:"seeds"`
	// Peers 下载数
	Peers int `json:"peers"`
	// Leeches 吸血数
	Leeches int `json:"leeches"`
	// LastAnnounceTime 最后通告时间
	LastAnnounceTime time.Time `json:"last_announce_time"`
	// NextAnnounceTime 下次通告时间
	NextAnnounceTime time.Time `json:"next_announce_time"`
	// LastAnnounceResult 最后通告结果
	LastAnnounceResult string `json:"last_announce_result"`
	// Message 状态消息
	Message string `json:"message"`
}

// FetchTorrentsParams 种子获取参数
type FetchTorrentsParams struct {
	// Status 状态过滤
	Status string `json:"status" form:"status"`
	// Category 分类过滤
	Category string `json:"category" form:"category"`
	// Search 搜索关键词
	Search string `json:"search" form:"search"`
	// Tags 标签过滤
	Tags []string `json:"tags" form:"tags"`
	// DownloaderType 下载器类型
	DownloaderType string `json:"downloader_type" form:"downloader_type"`
	// DownloaderID 下载器ID
	DownloaderID string `json:"downloader_id" form:"downloader_id"`
	// SortBy 排序字段
	SortBy TorrentSortField `json:"sort_by" form:"sort_by"`
	// SortOrder 排序顺序
	SortOrder SortOrder `json:"sort_order" form:"sort_order"`
	// Limit 返回数量限制
	Limit int `json:"limit" form:"limit"`
	// Offset 偏移量
	Offset int `json:"offset" form:"offset"`
	// IncludeFiles 是否包含文件列表
	IncludeFiles bool `json:"include_files" form:"include_files"`
	// IncludeTrackers 是否包含追踪器信息
	IncludeTrackers bool `json:"include_trackers" form:"include_trackers"`
	// OnlyActive 是否只显示活动的种子
	OnlyActive bool `json:"only_active" form:"only_active"`
	// OnlyCompleted 是否只显示已完成的种子
	OnlyCompleted bool `json:"only_completed" form:"only_completed"`
	// OnlyDownloading 是否只显示下载中的种子
	OnlyDownloading bool `json:"only_downloading" form:"only_downloading"`
	// OnlySeeding 是否只显示做种中的种子
	OnlySeeding bool `json:"only_seeding" form:"only_seeding"`
	// OnlyPaused 是否只显示已暂停的种子
	OnlyPaused bool `json:"only_paused" form:"only_paused"`
}

// TorrentResponse 种子响应结构
type TorrentResponse struct {
	// Success 是否成功
	Success bool `json:"success"`
	// Torrents 种子列表
	Torrents []*TorrentItem `json:"torrents"`
	// Total 总数
	Total int `json:"total"`
	// Filtered 过滤后数量
	Filtered int `json:"filtered"`
	// Page 当前页码
	Page int `json:"page"`
	// PageSize 每页大小
	PageSize int `json:"page_size"`
	// TotalPages 总页数
	TotalPages int `json:"total_pages"`
	// ProcessingTime 处理时间
	ProcessingTime time.Duration `json:"processing_time"`
	// Error 错误信息
	Error *TorrentError `json:"error,omitempty"`
}

// TorrentError 种子错误结构
type TorrentError struct {
	// Code 错误代码
	Code string `json:"code"`
	// Message 错误消息
	Message string `json:"message"`
	// Details 详细信息
	Details string `json:"details,omitempty"`
	// Timestamp 错误时间戳
	Timestamp time.Time `json:"timestamp"`
}

// TorrentStats 种子统计信息
type TorrentStats struct {
	// Active 活动种子数
	Active int `json:"active"`
	// Downloading 下载中种子数
	Downloading int `json:"downloading"`
	// Seeding 做种中种子数
	Seeding int `json:"seeding"`
	// Paused 已暂停种子数
	Paused int `json:"paused"`
	// Completed 已完成种子数
	Completed int `json:"completed"`
	// Error 错误种子数
	Error int `json:"error"`
	// Total 总种子数
	Total int `json:"total"`
	// TotalSize 总大小（字节）
	TotalSize int64 `json:"total_size"`
	// TotalCompletedSize 总已完成大小（字节）
	TotalCompletedSize int64 `json:"total_completed_size"`
	// TotalUploaded 总上传量（字节）
	TotalUploaded int64 `json:"total_uploaded"`
	// TotalDownloaded 总下载量（字节）
	TotalDownloaded int64 `json:"total_downloaded"`
	// TotalRatio 总分享率
	TotalRatio float64 `json:"total_ratio"`
	// CurrentDownloadSpeed 当前下载速度（字节/秒）
	CurrentDownloadSpeed int64 `json:"current_download_speed"`
	// CurrentUploadSpeed 当前上传速度（字节/秒）
	CurrentUploadSpeed int64 `json:"current_upload_speed"`
	// LastUpdated 最后更新时间
	LastUpdated time.Time `json:"last_updated"`
}

// TorrentManagerStatus 下载管理器状态
type TorrentManagerStatus struct {
	// DownloaderType 下载器类型
	DownloaderType DownloaderType `json:"downloader_type"`
	// DownloaderID 下载器ID
	DownloaderID string `json:"downloader_id"`
	// Connected 是否连接
	Connected bool `json:"connected"`
	// Version 下载器版本
	Version string `json:"version"`
	// Status 状态
	Status string `json:"status"`
	// Message 状态消息
	Message string `json:"message"`
	// Stats 统计信息
	Stats *TorrentStats `json:"stats"`
	// LastConnectionTime 最后连接时间
	LastConnectionTime time.Time `json:"last_connection_time"`
}

// TorrentFilesResponse 种子文件响应
type TorrentFilesResponse struct {
	// Success 是否成功
	Success bool `json:"success"`
	// Files 文件列表
	Files []*TorrentFile `json:"files"`
	// Total 总数
	Total int `json:"total"`
	// TorrentID 种子ID
	TorrentID string `json:"torrent_id"`
	// TorrentHash 种子哈希
	TorrentHash string `json:"torrent_hash"`
	// Error 错误信息
	Error *TorrentError `json:"error,omitempty"`
}

// TorrentTrackersResponse 种子追踪器响应
type TorrentTrackersResponse struct {
	// Success 是否成功
	Success bool `json:"success"`
	// Trackers 追踪器列表
	Trackers []*TrackerStatus `json:"trackers"`
	// Total 总数
	Total int `json:"total"`
	// TorrentID 种子ID
	TorrentID string `json:"torrent_id"`
	// TorrentHash 种子哈希
	TorrentHash string `json:"torrent_hash"`
	// Error 错误信息
	Error *TorrentError `json:"error,omitempty"`
}

// TorrentOperationRequest 种子操作请求
type TorrentOperationRequest struct {
	// IDs 种子ID列表
	IDs []string `json:"ids"`
	// Hashes 种子哈希列表
	Hashes []string `json:"hashes"`
	// DownloaderID 下载器ID
	DownloaderID string `json:"downloader_id"`
}

// TorrentOperationResponse 种子操作响应
type TorrentOperationResponse struct {
	// Success 是否成功
	Success bool `json:"success"`
	// AffectedCount 影响的种子数量
	AffectedCount int `json:"affected_count"`
	// Message 操作消息
	Message string `json:"message"`
	// Error 错误信息
	Error *TorrentError `json:"error,omitempty"`
}

// CategoryInfo 分类信息
type CategoryInfo struct {
	// Name 分类名称
	Name string `json:"name"`
	// SavePath 保存路径
	SavePath string `json:"save_path"`
	// Count 种子数量
	Count int `json:"count"`
	// DownloaderID 下载器ID
	DownloaderID string `json:"downloader_id"`
}

// CategoryResponse 分类响应
type CategoryResponse struct {
	// Success 是否成功
	Success bool `json:"success"`
	// Categories 分类列表
	Categories []*CategoryInfo `json:"categories"`
	// Total 总数
	Total int `json:"total"`
	// Error 错误信息
	Error *TorrentError `json:"error,omitempty"`
}
