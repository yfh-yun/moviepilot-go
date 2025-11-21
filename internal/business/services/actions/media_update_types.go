package actions

import (
	"time"
)

// MediaUpdateStatus 媒体更新状态枚举
type MediaUpdateStatus string

const (
	UpdateStatusPending   MediaUpdateStatus = "pending"
	UpdateStatusInProgress MediaUpdateStatus = "in_progress"
	UpdateStatusCompleted MediaUpdateStatus = "completed"
	UpdateStatusFailed    MediaUpdateStatus = "failed"
	UpdateStatusSkipped   MediaUpdateStatus = "skipped"
)

// MediaUpdateField 需要更新的媒体字段枚举
type MediaUpdateField string

const (
	UpdateFieldTitle       MediaUpdateField = "title"
	UpdateFieldOverview    MediaUpdateField = "overview"
	UpdateFieldPoster      MediaUpdateField = "poster"
	UpdateFieldBackdrop    MediaUpdateField = "backdrop"
	UpdateFieldActors      MediaUpdateField = "actors"
	UpdateFieldDirectors   MediaUpdateField = "directors"
	UpdateFieldGenres      MediaUpdateField = "genres"
	UpdateFieldReleaseDate MediaUpdateField = "release_date"
	UpdateFieldRating      MediaUpdateField = "rating"
	UpdateFieldRuntime     MediaUpdateField = "runtime"
	UpdateFieldStudio      MediaUpdateField = "studio"
	UpdateFieldTags        MediaUpdateField = "tags"
	UpdateFieldAll         MediaUpdateField = "all"
)

// MediaUpdateStrategy 更新策略枚举
type MediaUpdateStrategy string

const (
	StrategyForceUpdate   MediaUpdateStrategy = "force_update"
	StrategyIncremental   MediaUpdateStrategy = "incremental"
	StrategyOnlyMissing   MediaUpdateStrategy = "only_missing"
	StrategyOnlyOutdated  MediaUpdateStrategy = "only_outdated"
)

// MediaUpdatePriority 更新优先级枚举
type MediaUpdatePriority int

const (
	PriorityLow    MediaUpdatePriority = 0
	PriorityMedium MediaUpdatePriority = 1
	PriorityHigh   MediaUpdatePriority = 2
	PriorityCritical MediaUpdatePriority = 3
)

// MediaUpdateRequest 媒体更新请求
type MediaUpdateRequest struct {
	// MediaID 媒体ID，为空时表示批量更新
	MediaID string `json:"media_id" binding:"omitempty"`
	// MediaIDs 媒体ID列表，用于批量更新
	MediaIDs []string `json:"media_ids" binding:"omitempty,dive"`
	// MediaType 媒体类型 (movie, series)
	MediaType string `json:"media_type" binding:"omitempty,oneof=movie series"`
	// UpdateFields 需要更新的字段列表
	UpdateFields []MediaUpdateField `json:"update_fields" binding:"omitempty,dive,oneof=title overview poster backdrop actors directors genres release_date rating runtime studio tags all"`
	// UpdateStrategy 更新策略
	UpdateStrategy MediaUpdateStrategy `json:"update_strategy" binding:"omitempty,oneof=force_update incremental only_missing only_outdated"`
	// Priority 更新优先级
	Priority MediaUpdatePriority `json:"priority" binding:"omitempty,min=0,max=3"`
	// SkipIfNoChanges 是否在没有变更时跳过更新
	SkipIfNoChanges bool `json:"skip_if_no_changes" binding:"omitempty"`
	// NotifyOnCompletion 是否在完成时通知
	NotifyOnCompletion bool `json:"notify_on_completion" binding:"omitempty"`
	// TriggerBy 触发来源
	TriggerBy string `json:"trigger_by" binding:"omitempty"`
}

// MediaUpdateResponse 媒体更新响应
type MediaUpdateResponse struct {
	// TaskID 更新任务ID
	TaskID string `json:"task_id"`
	// Status 任务状态
	Status MediaUpdateStatus `json:"status"`
	// Total 总更新数量
	Total int `json:"total"`
	// Completed 已完成数量
	Completed int `json:"completed"`
	// Failed 失败数量
	Failed int `json:"failed"`
	// Skipped 跳过数量
	Skipped int `json:"skipped"`
	// StartTime 开始时间
	StartTime time.Time `json:"start_time"`
	// EndTime 结束时间，未完成时为空
	EndTime *time.Time `json:"end_time,omitempty"`
	// Message 任务状态消息
	Message string `json:"message"`
}

// MediaUpdateTask 媒体更新任务
type MediaUpdateTask struct {
	// ID 任务ID
	ID string `json:"id"`
	// Request 更新请求
	Request MediaUpdateRequest `json:"request"`
	// Status 任务状态
	Status MediaUpdateStatus `json:"status"`
	// Progress 进度信息
	Progress MediaUpdateProgress `json:"progress"`
	// Errors 错误信息
	Errors []MediaUpdateError `json:"errors"`
	// CreatedAt 创建时间
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt time.Time `json:"updated_at"`
	// CompletedAt 完成时间
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// MediaUpdateProgress 媒体更新进度
type MediaUpdateProgress struct {
	// Total 总更新数量
	Total int `json:"total"`
	// Completed 已完成数量
	Completed int `json:"completed"`
	// Failed 失败数量
	Failed int `json:"failed"`
	// Skipped 跳过数量
	Skipped int `json:"skipped"`
	// Current 当前正在更新的媒体ID
	Current string `json:"current,omitempty"`
	// Percentage 进度百分比
	Percentage float64 `json:"percentage"`
}

// MediaUpdateError 媒体更新错误
type MediaUpdateError struct {
	// MediaID 媒体ID
	MediaID string `json:"media_id"`
	// ErrorCode 错误代码
	ErrorCode string `json:"error_code"`
	// ErrorMessage 错误消息
	ErrorMessage string `json:"error_message"`
	// Timestamp 错误发生时间
	Timestamp time.Time `json:"timestamp"`
}

// MediaUpdateResult 单个媒体更新结果
type MediaUpdateResult struct {
	// MediaID 媒体ID
	MediaID string `json:"media_id"`
	// Success 是否成功
	Success bool `json:"success"`
	// UpdatedFields 实际更新的字段
	UpdatedFields []MediaUpdateField `json:"updated_fields"`
	// ChangedFields 发生变更的字段
	ChangedFields []MediaUpdateField `json:"changed_fields"`
	// Error 错误信息（如果失败）
	Error string `json:"error,omitempty"`
	// ExecutionTime 执行时间（毫秒）
	ExecutionTime int64 `json:"execution_time"`
	// UpdatedAt 更新时间
	UpdatedAt time.Time `json:"updated_at"`
}

// MediaUpdateStats 媒体更新统计
type MediaUpdateStats struct {
	// TotalTasks 总任务数
	TotalTasks int `json:"total_tasks"`
	// ActiveTasks 活跃任务数
	ActiveTasks int `json:"active_tasks"`
	// CompletedTasks 已完成任务数
	CompletedTasks int `json:"completed_tasks"`
	// FailedTasks 失败任务数
	FailedTasks int `json:"failed_tasks"`
	// AverageDuration 平均执行时间（秒）
	AverageDuration float64 `json:"average_duration"`
	// LastUpdateTime 最后更新时间
	LastUpdateTime *time.Time `json:"last_update_time,omitempty"`
}

// MediaUpdateTaskQuery 查询媒体更新任务的参数
type MediaUpdateTaskQuery struct {
	// TaskID 任务ID，精确匹配
	TaskID string `json:"task_id" binding:"omitempty"`
	// Status 任务状态过滤
	Status MediaUpdateStatus `json:"status" binding:"omitempty,oneof=pending in_progress completed failed skipped"`
	// MediaID 媒体ID，查询包含该媒体的任务
	MediaID string `json:"media_id" binding:"omitempty"`
	// TriggerBy 触发来源过滤
	TriggerBy string `json:"trigger_by" binding:"omitempty"`
	// StartTime 开始时间范围（开始）
	StartTime *time.Time `json:"start_time" binding:"omitempty"`
	// EndTime 开始时间范围（结束）
	EndTime *time.Time `json:"end_time" binding:"omitempty"`
	// Page 页码
	Page int `json:"page" binding:"omitempty,min=1"`
	// PageSize 每页大小
	PageSize int `json:"page_size" binding:"omitempty,min=1,max=100"`
	// OrderBy 排序字段
	OrderBy string `json:"order_by" binding:"omitempty,oneof=created_at updated_at completed_at status"`
	// OrderDir 排序方向
	OrderDir string `json:"order_dir" binding:"omitempty,oneof=asc desc"`
}

// MediaUpdateTaskListResponse 任务列表响应
type MediaUpdateTaskListResponse struct {
	// Tasks 任务列表
	Tasks []MediaUpdateTask `json:"tasks"`
	// Total 总任务数
	Total int64 `json:"total"`
	// Page 页码
	Page int `json:"page"`
	// PageSize 每页大小
	PageSize int `json:"page_size"`
	// TotalPages 总页数
	TotalPages int `json:"total_pages"`
}

// MediaUpdateConfig 媒体更新配置
type MediaUpdateConfig struct {
	// MaxConcurrentUpdates 最大并发更新数量
	MaxConcurrentUpdates int `json:"max_concurrent_updates" binding:"min=1,max=50"`
	// UpdateTimeout 更新超时时间（秒）
	UpdateTimeout int `json:"update_timeout" binding:"min=10,max=3600"`
	// RetryCount 失败重试次数
	RetryCount int `json:"retry_count" binding:"min=0,max=5"`
	// RetryInterval 重试间隔（秒）
	RetryInterval int `json:"retry_interval" binding:"min=1,max=600"`
	// CacheTTL 缓存过期时间（秒）
	CacheTTL int `json:"cache_ttl" binding:"min=60,max=86400"`
	// MaxBatchSize 最大批量大小
	MaxBatchSize int `json:"max_batch_size" binding:"min=1,max=1000"`
}

// MediaUpdateFieldMapping 字段映射配置
type MediaUpdateFieldMapping struct {
	// Source 来源字段
	Source string `json:"source"`
	// Target 目标字段
	Target string `json:"target"`
	// Transform 转换函数名称
	Transform string `json:"transform,omitempty"`
}

// MediaUpdateHook 媒体更新钩子配置
type MediaUpdateHook struct {
	// Name 钩子名称
	Name string `json:"name"`
	// Type 钩子类型 (pre_update, post_update, on_error)
	Type string `json:"type" binding:"oneof=pre_update post_update on_error"`
	// Priority 优先级
	Priority int `json:"priority" binding:"min=0,max=100"`
	// Enabled 是否启用
	Enabled bool `json:"enabled"`
	// Config 钩子配置
	Config map[string]interface{} `json:"config,omitempty"`
}
