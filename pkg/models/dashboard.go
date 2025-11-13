package models

// Statistic 统计信息
type Statistic struct {
	// 电影数量
	MovieCount int `json:"movie_count,omitempty"`
	
	// 电视剧数�?	TvCount int `json:"tv_count,omitempty"`
	
	// 集数�?	EpisodeCount int `json:"episode_count,omitempty"`
	
	// 用户数量
	UserCount int `json:"user_count,omitempty"`
}

// Storage 存储信息
type Storage struct {
	// 总存储空�?	TotalStorage float64 `json:"total_storage,omitempty"`
	
	// 已使用空�?	UsedStorage float64 `json:"used_storage,omitempty"`
}

// ProcessInfo 进程信息
type ProcessInfo struct {
	// 进程ID
	Pid int `json:"pid,omitempty"`
	
	// 进程名称
	Name string `json:"name,omitempty"`
	
	// 进程状�?	Status string `json:"status,omitempty"`
	
	// 进程占用CPU
	Cpu float64 `json:"cpu,omitempty"`
	
	// 进程占用内存 MB
	Memory float64 `json:"memory,omitempty"`
	
	// 进程创建时间
	CreateTime float64 `json:"create_time,omitempty"`
	
	// 进程运行时间 �?	RunTime float64 `json:"run_time,omitempty"`
}

// DownloaderInfo 下载器信�?type DownloaderInfo struct {
	// 下载速度
	DownloadSpeed float64 `json:"download_speed,omitempty"`
	
	// 上传速度
	UploadSpeed float64 `json:"upload_speed,omitempty"`
	
	// 下载�?	DownloadSize float64 `json:"download_size,omitempty"`
	
	// 上传�?	UploadSize float64 `json:"upload_size,omitempty"`
	
	// 剩余空间
	FreeSpace float64 `json:"free_space,omitempty"`
}

// ScheduleInfo 定时任务信息
type ScheduleInfo struct {
	// ID
	ID string `json:"id,omitempty"`
	
	// 名称
	Name string `json:"name,omitempty"`
	
	// 提供�?	Provider string `json:"provider,omitempty"`
	
	// 状�?	Status string `json:"status,omitempty"`
	
	// 下次执行时间
	NextRun string `json:"next_run,omitempty"`
}
