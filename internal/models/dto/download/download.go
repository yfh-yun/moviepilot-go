package download

// DownloadTask 下载任务
type DownloadTask struct {
	// 任务ID
	DownloadID string `json:"download_id,omitempty"`
	// 下载器
	Downloader string `json:"downloader,omitempty"`
	// 下载路径
	Path string `json:"path,omitempty"`
	// 是否完成
	Completed bool `json:"completed,omitempty"`
}
