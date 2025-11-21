// Package actions 提供下载管理的数据结构定义
package actions

import (
	"time"
)

// DownloadStatus 下载状态枚举
type DownloadStatus string

const (
	// DownloadStatusPending 等待中
	DownloadStatusPending DownloadStatus = "pending"
	// DownloadStatusDownloading 下载中
	DownloadStatusDownloading DownloadStatus = "downloading"
	// DownloadStatusPaused 暂停中
	DownloadStatusPaused DownloadStatus = "paused"
	// DownloadStatusCompleted 已完成
	DownloadStatusCompleted DownloadStatus = "completed"
	// DownloadStatusError 下载错误
	DownloadStatusError DownloadStatus = "error"
	// DownloadStatusSeeding 做种中
	DownloadStatusSeeding DownloadStatus = "seeding"
	// DownloadStatusStalled 停滞
	DownloadStatusStalled DownloadStatus = "stalled"
	// DownloadStatusChecking 校验中
	DownloadStatusChecking DownloadStatus = "checking"
)

// DownloadType 下载类型枚举
type DownloadType string

const (
	// DownloadTypeTorrent 种子下载
	DownloadTypeTorrent DownloadType = "torrent"
	// DownloadTypeMagnet 磁力链接下载
	DownloadTypeMagnet DownloadType = "magnet"
	// DownloadTypeURL HTTP下载
	DownloadTypeURL DownloadType = "url"
	// DownloadTypeNZB NZB下载
	DownloadTypeNZB DownloadType = "nzb"
)

// DownloadItem 下载项数据结构
type DownloadItem struct {
	// 下载ID
	ID string `json:"id"`
	// 下载器ID
	DownloaderID string `json:"downloader_id"`
	// 下载类型
	Type DownloadType `json:"type"`
	// 标题
	Title string `json:"title"`
	// 保存路径
	SavePath string `json:"save_path"`
	// 状态
	Status DownloadStatus `json:"status"`
	// 错误信息
	ErrorMessage string `json:"error_message,omitempty"`
	// 总大小 (字节)
	TotalSize int64 `json:"total_size"`
	// 已下载大小 (字节)
	DownloadedSize int64 `json:"downloaded_size"`
	// 上传大小 (字节)
	UploadedSize int64 `json:"uploaded_size"`
	// 下载速度 (字节/秒)
	DownloadSpeed int64 `json:"download_speed"`
	// 上传速度 (字节/秒)
	UploadSpeed int64 `json:"upload_speed"`
	// 完成百分比
	Progress float64 `json:"progress"`
	// 剩余时间估计 (秒)
	ETA int `json:"eta"`
	// 分享率
	Ratio float64 `json:"ratio"`
	// 创建时间
	CreatedAt time.Time `json:"created_at"`
	// 完成时间
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	// 更新时间
	UpdatedAt time.Time `json:"updated_at"`
	// 标签
	Tags []string `json:"tags"`
	// 种子信息
	TorrentInfo *TorrentInfo `json:"torrent_info,omitempty"`
	// 媒体信息
	MediaInfo *MediaInfo `json:"media_info,omitempty"`
	// 来源信息
	SourceInfo map[string]interface{} `json:"source_info,omitempty"`
}

// TorrentInfo 种子信息
type TorrentInfo struct {
	// 种子名称
	Name string `json:"name"`
	// 磁力链接
	MagnetURI string `json:"magnet_uri,omitempty"`
	// 哈希值
	Hash string `json:"hash"`
	// 发布者
	Publisher string `json:"publisher,omitempty"`
	// 发布站点
	Site string `json:"site,omitempty"`
	// 种子文件路径
	TorrentFilePath string `json:"torrent_file_path,omitempty"`
	// 做种人数
	SeederCount int `json:"seeder_count"`
	// 下载人数
	LeecherCount int `json:"leecher_count"`
	// 完成人数
	CompletedCount int `json:"completed_count"`
	// 包含文件数
	FileCount int `json:"file_count"`
	// 备注
	Notes string `json:"notes,omitempty"`
}

// MediaInfo 媒体信息
type MediaInfo struct {
	// 媒体ID
	ID string `json:"id,omitempty"`
	// 媒体类型 (movie, series, music, etc)
	MediaType string `json:"media_type,omitempty"`
	// 标题
	Title string `json:"title,omitempty"`
	// 原始标题
	OriginalTitle string `json:"original_title,omitempty"`
	// 年份
	Year int `json:"year,omitempty"`
	// 季数 (电视剧)
	Season int `json:"season,omitempty"`
	// 集数 (电视剧)
	Episode int `json:"episode,omitempty"`
	// 分辨率
	Resolution string `json:"resolution,omitempty"`
	// 视频编码
	VideoCodec string `json:"video_codec,omitempty"`
	// 音频编码
	AudioCodec string `json:"audio_codec,omitempty"`
	// 格式
	Format string `json:"format,omitempty"`
	// 质量
	Quality string `json:"quality,omitempty"`
	// 组名
	Group string `json:"group,omitempty"`
	// TMDB ID
	TmdbID string `json:"tmdb_id,omitempty"`
	// IMDB ID
	ImdbID string `json:"imdb_id,omitempty"`
	// TVDB ID
	TvdbID string `json:"tvdb_id,omitempty"`
}

// FetchDownloadsParams 获取下载列表参数
type FetchDownloadsParams struct {
	// 下载器ID过滤
	DownloaderID string `json:"downloader_id,omitempty"`
	// 状态过滤
	Statuses []DownloadStatus `json:"statuses,omitempty"`
	// 类型过滤
	Types []DownloadType `json:"types,omitempty"`
	// 关键词过滤
	Keywords []string `json:"keywords,omitempty"`
	// 标签过滤
	Tags []string `json:"tags,omitempty"`
	// 是否包含详细信息
	IncludeDetails bool `json:"include_details"`
	// 是否包含媒体信息
	IncludeMediaInfo bool `json:"include_media_info"`
	// 是否只返回未完成的下载
	OnlyActive bool `json:"only_active"`
	// 排序字段
	OrderBy string `json:"order_by"`
	// 排序方向
	OrderDir string `json:"order_dir"`
	// 分页限制
	Limit int `json:"limit"`
	// 分页偏移
	Offset int `json:"offset"`
}

// FetchDownloadsResult 获取下载列表结果
type FetchDownloadsResult struct {
	// 总数
	Total int64 `json:"total"`
	// 下载项列表
	Items []*DownloadItem `json:"items"`
	// 请求参数
	Params FetchDownloadsParams `json:"params"`
}

// GetDownloadStatusParams 获取下载状态参数
type GetDownloadStatusParams struct {
	// 下载ID
	ID string `json:"id" binding:"required"`
	// 下载器ID
	DownloaderID string `json:"downloader_id,omitempty"`
	// 是否包含详细信息
	IncludeDetails bool `json:"include_details"`
	// 是否包含媒体信息
	IncludeMediaInfo bool `json:"include_media_info"`
}

// GetDownloadStatusResult 获取下载状态结果
type GetDownloadStatusResult struct {
	// 下载项
	Item *DownloadItem `json:"item,omitempty"`
	// 是否找到
	Found bool `json:"found"`
}

// DownloadStats 下载统计信息
type DownloadStats struct {
	// 总下载数
	TotalCount int `json:"total_count"`
	// 各状态数量
	StatusCounts map[DownloadStatus]int `json:"status_counts"`
	// 各类型数量
	TypeCounts map[DownloadType]int `json:"type_counts"`
	// 总大小 (字节)
	TotalSize int64 `json:"total_size"`
	// 已下载大小 (字节)
	TotalDownloadedSize int64 `json:"total_downloaded_size"`
	// 当前下载速度 (字节/秒)
	CurrentDownloadSpeed int64 `json:"current_download_speed"`
	// 当前上传速度 (字节/秒)
	CurrentUploadSpeed int64 `json:"current_upload_speed"`
	// 今日下载量 (字节)
	TodayDownloaded int64 `json:"today_downloaded"`
	// 今日上传量 (字节)
	TodayUploaded int64 `json:"today_uploaded"`
	// 总下载量 (字节)
	AllTimeDownloaded int64 `json:"all_time_downloaded"`
	// 总上传量 (字节)
	AllTimeUploaded int64 `json:"all_time_uploaded"`
	// 平均分享率
	AverageRatio float64 `json:"average_ratio"`
}

// DownloadHistory 下载历史
type DownloadHistory struct {
	// 历史记录ID
	ID string `json:"id"`
	// 下载项ID
	DownloadID string `json:"download_id"`
	// 标题
	Title string `json:"title"`
	// 状态
	Status DownloadStatus `json:"status"`
	// 大小 (字节)
	Size int64 `json:"size"`
	// 下载耗时 (秒)
	DownloadTime int `json:"download_time"`
	// 分享率
	Ratio float64 `json:"ratio"`
	// 创建时间
	CreatedAt time.Time `json:"created_at"`
	// 完成时间
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	// 删除时间
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// DownloadFilter 下载过滤器
type DownloadFilter struct {
	// 下载器ID
	DownloaderID string
	// 状态列表
	Statuses []DownloadStatus
	// 类型列表
	Types []DownloadType
	// 关键词列表
	Keywords []string
	// 标签列表
	Tags []string
	// 开始时间
	StartDate *time.Time
	// 结束时间
	EndDate *time.Time
	// 最小大小
	MinSize int64
	// 最大大小
	MaxSize int64
	// 最小分享率
	MinRatio float64
	// 是否只显示活跃下载
	OnlyActive bool
	// 是否只显示已完成
	OnlyCompleted bool
	// 排序字段
	OrderBy string
	// 排序方向
	OrderDir string
	// 分页限制
	Limit int
	// 分页偏移
	Offset int
}

// DownloadSummary 下载摘要信息
type DownloadSummary struct {
	// 下载项ID
	ID string `json:"id"`
	// 标题
	Title string `json:"title"`
	// 状态
	Status DownloadStatus `json:"status"`
	// 进度
	Progress float64 `json:"progress"`
	// 下载速度
	DownloadSpeed int64 `json:"download_speed"`
	// 剩余时间
	ETA int `json:"eta"`
	// 更新时间
	UpdatedAt time.Time `json:"updated_at"`
}

// BatchDownloadParams 批量操作参数
type BatchDownloadParams struct {
	// 下载ID列表
	IDs []string `json:"ids" binding:"required"`
	// 下载器ID
	DownloaderID string `json:"downloader_id,omitempty"`
}

// BatchDownloadResult 批量操作结果
type BatchDownloadResult struct {
	// 成功数量
	SuccessCount int `json:"success_count"`
	// 失败数量
	FailedCount int `json:"failed_count"`
	// 失败详情
	FailedItems []BatchFailedItem `json:"failed_items,omitempty"`
}

// BatchFailedItem 批量操作失败项
type BatchFailedItem struct {
	// 下载ID
	ID string `json:"id"`
	// 错误信息
	Error string `json:"error"`
}
