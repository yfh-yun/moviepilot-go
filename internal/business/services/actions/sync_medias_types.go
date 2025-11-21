package actions

import (
	"time"
)

// SyncStrategy 同步策略枚举
type SyncStrategy string

const (
	// SyncStrategyFull 完全同步
	SyncStrategyFull SyncStrategy = "full"
	// SyncStrategyIncremental 增量同步
	SyncStrategyIncremental SyncStrategy = "incremental"
	// SyncStrategyOnlyNew 仅同步新增
	SyncStrategyOnlyNew SyncStrategy = "only_new"
	// SyncStrategyOnlyUpdate 仅同步更新
	SyncStrategyOnlyUpdate SyncStrategy = "only_update"
)

// SyncSource 同步源枚举
type SyncSource string

const (
	// SyncSourceLocal 本地文件系统
	SyncSourceLocal SyncSource = "local"
	// SyncSourceRemote 远程媒体服务器
	SyncSourceRemote SyncSource = "remote"
	// SyncSourceDatabase 数据库
	SyncSourceDatabase SyncSource = "database"
	// SyncSourceAPI 第三方API
	SyncSourceAPI SyncSource = "api"
)

// SyncStatus 同步状态枚举
type SyncStatus string

const (
	// SyncStatusPending 等待同步
	SyncStatusPending SyncStatus = "pending"
	// SyncStatusInProgress 同步中
	SyncStatusInProgress SyncStatus = "in_progress"
	// SyncStatusCompleted 同步完成
	SyncStatusCompleted SyncStatus = "completed"
	// SyncStatusFailed 同步失败
	SyncStatusFailed SyncStatus = "failed"
	// SyncStatusCancelled 同步取消
	SyncStatusCancelled SyncStatus = "cancelled"
	// SyncStatusPartiallyCompleted 部分完成
	SyncStatusPartiallyCompleted SyncStatus = "partially_completed"
)

// SyncField 同步字段枚举
type SyncField string

const (
	// SyncFieldAll 所有字段
	SyncFieldAll SyncField = "all"
	// SyncFieldBasic 基本信息
	SyncFieldBasic SyncField = "basic"
	// SyncFieldMetadata 元数据
	SyncFieldMetadata SyncField = "metadata"
	// SyncFieldFiles 文件信息
	SyncFieldFiles SyncField = "files"
	// SyncFieldStatus 状态信息
	SyncFieldStatus SyncField = "status"
	// SyncFieldTags 标签信息
	SyncFieldTags SyncField = "tags"
)

// MediaSyncRequest 媒体同步请求
type MediaSyncRequest struct {
	// 同步源
	Source SyncSource `json:"source" binding:"required"`
	// 同步策略
	Strategy SyncStrategy `json:"strategy" binding:"required"`
	// 同步字段
	SyncFields []SyncField `json:"sync_fields"`
	// 媒体类型过滤
	MediaType string `json:"media_type"`
	// 媒体ID列表（可选）
	MediaIDs []string `json:"media_ids"`
	// 源路径（本地同步时使用）
	SourcePath string `json:"source_path"`
	// 目标路径（本地同步时使用）
	TargetPath string `json:"target_path"`
	// 远程服务器ID（远程同步时使用）
	RemoteServerID string `json:"remote_server_id"`
	// 是否删除不存在的媒体
	DeleteMissing bool `json:"delete_missing"`
	// 是否覆盖已存在的媒体
	OverwriteExisting bool `json:"overwrite_existing"`
	// 并发数
	Concurrency int `json:"concurrency"`
	// 超时时间（秒）
	Timeout int `json:"timeout"`
	// 最大重试次数
	MaxRetries int `json:"max_retries"`
	// 是否跳过验证
	SkipValidation bool `json:"skip_validation"`
	// 是否启用日志
	EnableLogging bool `json:"enable_logging"`
	// 自定义过滤条件
	Filter map[string]interface{} `json:"filter"`
}

// MediaSyncResponse 媒体同步响应
type MediaSyncResponse struct {
	// 任务ID
	TaskID string `json:"task_id"`
	// 同步状态
	Status SyncStatus `json:"status"`
	// 总媒体数量
	Total int `json:"total"`
	// 已同步数量
	Synced int `json:"synced"`
	// 失败数量
	Failed int `json:"failed"`
	// 跳过数量
	Skipped int `json:"skipped"`
	// 删除数量
	Deleted int `json:"deleted"`
	// 创建时间
	StartTime time.Time `json:"start_time"`
	// 消息
	Message string `json:"message"`
}

// MediaSyncResult 单个媒体同步结果
type MediaSyncResult struct {
	// 媒体ID
	MediaID string `json:"media_id"`
	// 媒体标题
	Title string `json:"title"`
	// 是否成功
	Success bool `json:"success"`
	// 同步操作类型（新增/更新/删除/跳过）
	Operation string `json:"operation"`
	// 错误信息
	Error string `json:"error,omitempty"`
	// 同步字段
	SyncedFields []SyncField `json:"synced_fields"`
	// 执行时间（毫秒）
	ExecutionTime int64 `json:"execution_time"`
	// 同步时间
	SyncedAt time.Time `json:"synced_at"`
}

// MediaSyncTask 媒体同步任务
type MediaSyncTask struct {
	// 任务ID
	ID string `json:"id"`
	// 请求参数
	Request *MediaSyncRequest `json:"request"`
	// 状态
	Status SyncStatus `json:"status"`
	// 进度
	Progress *SyncProgress `json:"progress"`
	// 统计信息
	Stats *SyncStats `json:"stats"`
	// 创建时间
	CreatedAt time.Time `json:"created_at"`
	// 更新时间
	UpdatedAt time.Time `json:"updated_at"`
	// 完成时间
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	// 错误信息
	Error string `json:"error,omitempty"`
	// 日志信息
	Logs []SyncLog `json:"logs,omitempty"`
}

// SyncProgress 同步进度
type SyncProgress struct {
	// 总媒体数量
	Total int `json:"total"`
	// 已处理数量
	Processed int `json:"processed"`
	// 进度百分比
	Percentage int `json:"percentage"`
	// 当前操作
	CurrentOperation string `json:"current_operation"`
	// 预计剩余时间（秒）
	EstimatedRemaining int `json:"estimated_remaining,omitempty"`
}

// SyncStats 同步统计信息
type SyncStats struct {
	// 成功数量
	Success int `json:"success"`
	// 失败数量
	Failed int `json:"failed"`
	// 跳过数量
	Skipped int `json:"skipped"`
	// 新增数量
	Added int `json:"added"`
	// 更新数量
	Updated int `json:"updated"`
	// 删除数量
	Deleted int `json:"deleted"`
	// 总耗时（毫秒）
	TotalTime int64 `json:"total_time"`
	// 平均耗时（毫秒）
	AverageTime int64 `json:"average_time"`
	// 错误详情
	ErrorDetails map[string]int `json:"error_details"`
}

// SyncLog 同步日志
type SyncLog struct {
	// 日志级别
	Level string `json:"level"`
	// 时间
	Timestamp time.Time `json:"timestamp"`
	// 消息
	Message string `json:"message"`
	// 媒体ID
	MediaID string `json:"media_id,omitempty"`
	// 操作类型
	Operation string `json:"operation,omitempty"`
}

// MediaSyncTaskQuery 媒体同步任务查询参数
type MediaSyncTaskQuery struct {
	// 页码
	Page int `json:"page" form:"page"`
	// 每页数量
	PageSize int `json:"page_size" form:"page_size"`
	// 任务状态
	Status string `json:"status" form:"status"`
	// 同步源
	Source string `json:"source" form:"source"`
	// 媒体类型
	MediaType string `json:"media_type" form:"media_type"`
	// 开始时间
	StartTime string `json:"start_time" form:"start_time"`
	// 结束时间
	EndTime string `json:"end_time" form:"end_time"`
	// 搜索关键词
	Keyword string `json:"keyword" form:"keyword"`
}

// MediaSyncTaskListResponse 媒体同步任务列表响应
type MediaSyncTaskListResponse struct {
	// 任务列表
	Tasks []*MediaSyncTask `json:"tasks"`
	// 总数
	Total int64 `json:"total"`
	// 页码
	Page int `json:"page"`
	// 每页数量
	PageSize int `json:"page_size"`
	// 总页数
	TotalPages int `json:"total_pages"`
}

// MediaSyncStats 媒体同步统计总览
type MediaSyncStats struct {
	// 今日同步统计
	Today *DailySyncStats `json:"today"`
	// 本周同步统计
	Week *WeeklySyncStats `json:"week"`
	// 本月同步统计
	Month *MonthlySyncStats `json:"month"`
	// 任务状态统计
	TaskStatuses map[SyncStatus]int `json:"task_statuses"`
	// 同步源统计
	Sources map[SyncSource]int `json:"sources"`
	// 成功率
	SuccessRate float64 `json:"success_rate"`
	// 平均同步时间
	AverageSyncTime float64 `json:"average_sync_time"`
}

// DailySyncStats 每日同步统计
type DailySyncStats struct {
	// 任务数量
	TaskCount int `json:"task_count"`
	// 媒体数量
	MediaCount int `json:"media_count"`
	// 成功数量
	SuccessCount int `json:"success_count"`
	// 失败数量
	FailureCount int `json:"failure_count"`
	// 总耗时（分钟）
	TotalDuration int `json:"total_duration"`
}

// WeeklySyncStats 每周同步统计
type WeeklySyncStats struct {
	// 任务数量
	TaskCount int `json:"task_count"`
	// 媒体数量
	MediaCount int `json:"media_count"`
	// 平均成功率
	AverageSuccessRate float64 `json:"average_success_rate"`
	// 每日统计
	DailyStats []*DailySyncStats `json:"daily_stats"`
}

// MonthlySyncStats 每月同步统计
type MonthlySyncStats struct {
	// 任务数量
	TaskCount int `json:"task_count"`
	// 媒体数量
	MediaCount int `json:"media_count"`
	// 成功率趋势
	SuccessRateTrend []float64 `json:"success_rate_trend"`
	// 每周统计
	WeeklyStats []*WeeklySyncStats `json:"weekly_stats"`
}

// SyncConflict 同步冲突信息
type SyncConflict struct {
	// 冲突ID
	ID string `json:"id"`
	// 媒体ID
	MediaID string `json:"media_id"`
	// 媒体标题
	Title string `json:"title"`
	// 冲突类型
	ConflictType string `json:"conflict_type"`
	// 本地版本
	LocalVersion *MediaVersion `json:"local_version,omitempty"`
	// 远程版本
	RemoteVersion *MediaVersion `json:"remote_version,omitempty"`
	// 创建时间
	CreatedAt time.Time `json:"created_at"`
	// 状态
	Status string `json:"status"`
}

// MediaVersion 媒体版本信息
type MediaVersion struct {
	// 版本号
	Version string `json:"version"`
	// 更新时间
	UpdatedAt time.Time `json:"updated_at"`
	// 修改者
	ModifiedBy string `json:"modified_by"`
	// 文件大小
	FileSize int64 `json:"file_size"`
	// 文件哈希
	FileHash string `json:"file_hash"`
}

// SyncConflictResolution 同步冲突解决方案
type SyncConflictResolution struct {
	// 冲突ID
	ConflictID string `json:"conflict_id" binding:"required"`
	// 解决方案（keep_local/keep_remote/merge）
	Resolution string `json:"resolution" binding:"required"`
	// 自定义字段映射
	FieldMapping map[string]string `json:"field_mapping"`
}
