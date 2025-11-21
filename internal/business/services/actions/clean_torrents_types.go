package actions

import (
	"time"

	"moviepilot-go/pkg/errors"
)

// CleanStatus 清理状态枚举
type CleanStatus string

const (
	CleanStatusPending   CleanStatus = "pending"
	CleanStatusRunning   CleanStatus = "running"
	CleanStatusCompleted CleanStatus = "completed"
	CleanStatusFailed    CleanStatus = "failed"
	CleanStatusCancelled CleanStatus = "cancelled"
)

// CleanStrategy 清理策略枚举
type CleanStrategy string

const (
	CleanStrategyNone      CleanStrategy = "none"
	CleanStrategyByTime    CleanStrategy = "by_time"
	CleanStrategyByRatio   CleanStrategy = "by_ratio"
	CleanStrategyBySeeder  CleanStrategy = "by_seeder"
	CleanStrategyByStorage CleanStrategy = "by_storage"
)

// TorrentState 种子状态枚举
type TorrentState string

const (
	TorrentStateDownloading TorrentState = "downloading"
	TorrentStateSeeding     TorrentState = "seeding"
	TorrentStatePaused      TorrentState = "paused"
	TorrentStateCompleted   TorrentState = "completed"
	TorrentStateError       TorrentState = "error"
)

// CleanTorrentRequest 清理种子请求结构体
type CleanTorrentRequest struct {
	// 清理策略
	Strategy CleanStrategy `json:"strategy" binding:"required,oneof=none by_time by_ratio by_seeder by_storage"`
	
	// 时间阈值（小时），当策略为by_time时必填
	TimeThreshold int `json:"time_threshold,omitempty" binding:"omitempty,min=1"`
	
	// 做种比率阈值，当策略为by_ratio时必填
	RatioThreshold float64 `json:"ratio_threshold,omitempty" binding:"omitempty,min=0"`
	
	// 做种人数阈值，当策略为by_seeder时必填
	SeederThreshold int `json:"seeder_threshold,omitempty" binding:"omitempty,min=0"`
	
	// 存储阈值（GB），当策略为by_storage时必填
	StorageThreshold int `json:"storage_threshold,omitempty" binding:"omitempty,min=1"`
	
	// 是否只标记不删除
	OnlyMark bool `json:"only_mark"`
	
	// 是否包含已下载完成的种子
	IncludeCompleted bool `json:"include_completed"`
	
	// 是否包含正在下载的种子
	IncludeDownloading bool `json:"include_downloading"`
	
	// 是否包含已暂停的种子
	IncludePaused bool `json:"include_paused"`
	
	// 排除特定标签的种子
	ExcludeTags []string `json:"exclude_tags,omitempty"`
	
	// 排除特定Tracker的种子
	ExcludeTrackers []string `json:"exclude_trackers,omitempty"`
	
	// 下载器名称列表
	Downloaders []string `json:"downloaders,omitempty"`
}

// CleanTorrentResponse 清理种子响应结构体
type CleanTorrentResponse struct {
	// 任务ID
	TaskID string `json:"task_id"`
	
	// 清理状态
	Status CleanStatus `json:"status"`
	
	// 消息
	Message string `json:"message,omitempty"`
}

// CleanTorrentResult 单个种子清理结果
type CleanTorrentResult struct {
	// 种子ID
	TorrentID string `json:"torrent_id"`
	
	// 种子名称
	Name string `json:"name"`
	
	// 下载器名称
	Downloader string `json:"downloader"`
	
	// 清理状态
	Cleaned bool `json:"cleaned"`
	
	// 清理原因
	Reason string `json:"reason,omitempty"`
	
	// 错误信息
	Error string `json:"error,omitempty"`
}

// CleanTorrentTask 清理任务信息
type CleanTorrentTask struct {
	// 任务ID
	TaskID string `json:"task_id"`
	
	// 清理请求参数
	Request CleanTorrentRequest `json:"request"`
	
	// 清理状态
	Status CleanStatus `json:"status"`
	
	// 开始时间
	StartTime time.Time `json:"start_time"`
	
	// 结束时间
	EndTime *time.Time `json:"end_time,omitempty"`
	
	// 已处理的种子数量
	ProcessedCount int `json:"processed_count"`
	
	// 已清理的种子数量
	CleanedCount int `json:"cleaned_count"`
	
	// 失败的种子数量
	FailedCount int `json:"failed_count"`
	
	// 错误信息
	Error string `json:"error,omitempty"`
	
	// 进度百分比
	Progress int `json:"progress"`
}

// CleanTorrentStats 清理统计信息
type CleanTorrentStats struct {
	// 总任务数
	TotalTasks int `json:"total_tasks"`
	
	// 进行中的任务数
	RunningTasks int `json:"running_tasks"`
	
	// 已完成的任务数
	CompletedTasks int `json:"completed_tasks"`
	
	// 失败的任务数
	FailedTasks int `json:"failed_tasks"`
	
	// 总计清理的种子数
	TotalCleaned int `json:"total_cleaned"`
	
	// 平均清理时间
	AverageCleanTime float64 `json:"average_clean_time"`
}

// TorrentInfo 种子信息结构体
type TorrentInfo struct {
	// 种子ID
	ID string `json:"id"`
	
	// 种子名称
	Name string `json:"name"`
	
	// 下载器名称
	Downloader string `json:"downloader"`
	
	// 种子状态
	State TorrentState `json:"state"`
	
	// 下载路径
	SavePath string `json:"save_path"`
	
	// 做种比率
	Ratio float64 `json:"ratio"`
	
	// 做种时间（小时）
	SeedingTime int `json:"seeding_time"`
	
	// 做种人数
	SeederCount int `json:"seeder_count"`
	
	// 大小（字节）
	Size int64 `json:"size"`
	
	// 标签列表
	Tags []string `json:"tags"`
	
	// Tracker列表
	Trackers []string `json:"trackers"`
	
	// 完成时间
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// Validate 验证清理请求参数
func (r *CleanTorrentRequest) Validate() error {
	if r.Strategy == CleanStrategyNone {
		return errors.NewValidationError("strategy cannot be none")
	}

	// 根据策略验证相应的阈值参数
	switch r.Strategy {
	case CleanStrategyByTime:
		if r.TimeThreshold <= 0 {
			return errors.NewValidationError("time_threshold must be greater than 0 when strategy is by_time")
		}
	case CleanStrategyByRatio:
		if r.RatioThreshold < 0 {
			return errors.NewValidationError("ratio_threshold must be greater than or equal to 0 when strategy is by_ratio")
		}
	case CleanStrategyBySeeder:
		if r.SeederThreshold < 0 {
			return errors.NewValidationError("seeder_threshold must be greater than or equal to 0 when strategy is by_seeder")
		}
	case CleanStrategyByStorage:
		if r.StorageThreshold <= 0 {
			return errors.NewValidationError("storage_threshold must be greater than 0 when strategy is by_storage")
		}
	}

	// 至少包含一种状态的种子
	if !r.IncludeCompleted && !r.IncludeDownloading && !r.IncludePaused {
		return errors.NewValidationError("at least one of include_completed, include_downloading, or include_paused must be true")
	}

	return nil
}
