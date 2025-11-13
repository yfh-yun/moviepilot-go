package models

// TransferTorrent 转移种子信息
type TransferTorrent struct {
	// 下载�?	Downloader string `json:"downloader,omitempty"`
	// 标题
	Title string `json:"title,omitempty"`
	// 路径
	Path string `json:"path,omitempty"`
	// Hash
	Hash string `json:"hash,omitempty"`
	// 大小
	Size int64 `json:"size,omitempty"`
	// 标签
	Tags string `json:"tags,omitempty"`
	// 进度
	Progress float64 `json:"progress,omitempty"`
	// 状�?	State string `json:"state,omitempty"`
}

// DownloadingTorrent 下载中种子信�?type DownloadingTorrent struct {
	// 下载�?	Downloader string `json:"downloader,omitempty"`
	// Hash
	Hash string `json:"hash,omitempty"`
	// 标题
	Title string `json:"title,omitempty"`
	// 名称
	Name string `json:"name,omitempty"`
	// 年份
	Year string `json:"year,omitempty"`
	// 季集
	SeasonEpisode string `json:"season_episode,omitempty"`
	// 进度
	Progress float64 `json:"progress,omitempty"`
	// 大小
	Size int64 `json:"size,omitempty"`
	// 状�?	State string `json:"state,omitempty"`
	// 下载速度
	Dlspeed string `json:"dlspeed,omitempty"`
	// 上传速度
	Upspeed string `json:"upspeed,omitempty"`
	// 剩余时间
	LeftTime string `json:"left_time,omitempty"`
}

// DownloaderInfo 下载器信�?type DownloaderInfo struct {
	// 下载速度
	DownloadSpeed float64 `json:"download_speed,omitempty"`
	// 上传速度
	UploadSpeed float64 `json:"upload_speed,omitempty"`
	// 下载大小
	DownloadSize float64 `json:"download_size,omitempty"`
	// 上传大小
	UploadSize float64 `json:"upload_size,omitempty"`
}

// TorrentStatus 种子状�?type TorrentStatus string

const (
	// 转移
	TorrentStatusTransfer TorrentStatus = "transfer"
	// 下载�?	TorrentStatusDownloading TorrentStatus = "downloading"
	// 已完�?	TorrentStatusCompleted TorrentStatus = "completed"
	// 已停�?	TorrentStatusStopped TorrentStatus = "stopped"
	// 错误
	TorrentStatusErrored TorrentStatus = "errored"
)

// NewTransferTorrent 创建一个新�?TransferTorrent 实例
func NewTransferTorrent() *TransferTorrent {
	return &TransferTorrent{}
}

// NewDownloadingTorrent 创建一个新�?DownloadingTorrent 实例
func NewDownloadingTorrent() *DownloadingTorrent {
	return &DownloadingTorrent{}
}

// NewDownloaderInfo 创建一个新�?DownloaderInfo 实例
func NewDownloaderInfo() *DownloaderInfo {
	return &DownloaderInfo{}
}
