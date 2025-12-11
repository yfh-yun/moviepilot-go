package enums

// ProgressKey 处理进度Key字典
type ProgressKey string

const (
	// 搜索
	ProgressKeySearch ProgressKey = "search"
	// 整理
	ProgressKeyFileTransfer ProgressKey = "filetransfer"
	// 批量重命名
	ProgressKeyBatchRename ProgressKey = "batchrename"
)

// SortType 排序类型枚举
type SortType string

const (
	SortTypeTime   SortType = "time"   // 按时间排序
	SortTypeCount  SortType = "count"  // 按人数排序
	SortTypeRating SortType = "rating" // 按评分排序
)

// TorrentStatus 种子状态
type TorrentStatus string

const (
	TorrentStatusTransfer    TorrentStatus = "可转移"
	TorrentStatusDownloading TorrentStatus = "下载中"
)
