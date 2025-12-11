package transfer

// TransferManualRequest 手动转移请求
type TransferManualRequest struct {
	SourcePath string `json:"source_path" binding:"required"` // 源文件路径
	TargetPath string `json:"target_path" binding:"required"` // 目标路径
	MediaID    string `json:"media_id,omitempty"`             // 媒体ID
	Overwrite  bool   `json:"overwrite,omitempty"`            // 是否覆盖现有文件
	Preserve   bool   `json:"preserve,omitempty"`             // 是否保留源文件
}

// TransferManualResponse 手动转移响应
type TransferManualResponse struct {
	TaskID     string `json:"task_id"`     // 任务ID
	SourcePath string `json:"source_path"` // 源文件路径
	TargetPath string `json:"target_path"` // 目标路径
	MediaID    string `json:"media_id"`    // 媒体ID
	Status     string `json:"status"`      // 状态: created, processing, completed, failed
	Message    string `json:"message"`     // 状态消息
	Progress   int    `json:"progress"`    // 进度百分比 (0-100)
}

// TransferRecord 转移记录
type TransferRecord struct {
	ID          string `json:"id"`           // 记录ID
	TaskID      string `json:"task_id"`      // 任务ID
	SourcePath  string `json:"source_path"`  // 源文件路径
	TargetPath  string `json:"target_path"`  // 目标路径
	MediaID     string `json:"media_id"`     // 媒体ID
	Status      string `json:"status"`       // 状态: completed, failed, cancelled
	FileSize    int64  `json:"file_size"`    // 文件大小
	Speed       int64  `json:"speed"`        // 转移速度 (bytes/sec)
	Duration    int    `json:"duration"`     // 转移耗时 (seconds)
	ErrorMsg    string `json:"error_msg"`    // 错误信息
	CreatedAt   string `json:"created_at"`   // 创建时间
	CompletedAt string `json:"completed_at"` // 完成时间
}

// TransferHistoryRequest 转移历史请求
type TransferHistoryRequest struct {
	Page      int    `json:"page,omitempty"`       // 页码
	Size      int    `json:"size,omitempty"`       // 每页数量
	Status    string `json:"status,omitempty"`     // 状态过滤
	MediaID   string `json:"media_id,omitempty"`   // 媒体ID过滤
	StartDate string `json:"start_date,omitempty"` // 开始日期
	EndDate   string `json:"end_date,omitempty"`   // 结束日期
}

// TransferHistoryResponse 转移历史响应
type TransferHistoryResponse struct {
	Page    int              `json:"page"`    // 当前页
	Size    int              `json:"size"`    // 每页数量
	Total   int              `json:"total"`   // 总数量
	Status  string           `json:"status"`  // 当前过滤状态
	Records []TransferRecord `json:"records"` // 转移记录
	Message string           `json:"message"` // 状态消息
}

// TransferStatusRequest 转移状态请求
type TransferStatusRequest struct {
	TaskID string `json:"task_id,omitempty"` // 任务ID（可选，不提供则返回所有活动任务）
}

// TransferStatusResponse 转移状态响应
type TransferStatusResponse struct {
	TaskID    string `json:"task_id"`    // 任务ID
	Status    string `json:"status"`     // 状态: pending, processing, completed, failed, cancelled
	Progress  int    `json:"progress"`   // 进度百分比 (0-100)
	Speed     int64  `json:"speed"`      // 转移速度 (bytes/sec)
	TotalSize int64  `json:"total_size"` // 总文件大小
	Completed int64  `json:"completed"`  // 已完成字节数
	Remaining int    `json:"remaining"`  // 预计剩余时间 (seconds)
	Message   string `json:"message"`    // 状态消息
	StartTime string `json:"start_time"` // 开始时间
	ETA       string `json:"eta"`        // 预计完成时间
}

// TransferConfig 转移配置
type TransferConfig struct {
	Enable             bool     `json:"enable"`                 // 是否启用自动转移
	SourceDir          string   `json:"source_dir"`             // 源目录
	TargetDir          string   `json:"target_dir"`             // 目标目录
	FilePattern        string   `json:"file_pattern"`           // 文件匹配模式
	DirectoryStructure string   `json:"directory_structure"`    // 目录结构模板
	Overwrite          bool     `json:"overwrite"`              // 是否覆盖
	PreserveSource     bool     `json:"preserve_source"`        // 是否保留源文件
	MinSize            int64    `json:"min_size,omitempty"`     // 最小文件大小过滤
	MaxSize            int64    `json:"max_size,omitempty"`     // 最大文件大小过滤
	AllowedExts        []string `json:"allowed_exts,omitempty"` // 允许的文件扩展名
	IgnoreExts         []string `json:"ignore_exts,omitempty"`  // 忽略的文件扩展名
}

// TransferConfigRequest 转移配置请求
type TransferConfigRequest struct {
	Config TransferConfig `json:"config" binding:"required"` // 转移配置
}

// TransferConfigResponse 转移配置响应
type TransferConfigResponse struct {
	Config  TransferConfig `json:"config"`  // 转移配置
	Message string         `json:"message"` // 状态消息
}

// TransferStatistics 转移统计
type TransferStatistics struct {
	TotalFiles      int     `json:"total_files"`      // 总文件数
	TotalSize       int64   `json:"total_size"`       // 总文件大小
	TransferredSize int64   `json:"transferred_size"` // 已转移大小
	SuccessCount    int     `json:"success_count"`    // 成功数量
	FailedCount     int     `json:"failed_count"`     // 失败数量
	SuccessRate     float64 `json:"success_rate"`     // 成功率
	AverageSpeed    int64   `json:"average_speed"`    // 平均速度
	LastTransfer    string  `json:"last_transfer"`    // 最后转移时间
}

// TransferStatisticsRequest 转移统计请求
type TransferStatisticsRequest struct {
	Period string `json:"period,omitempty"` // 统计周期: day, week, month, year
}

// TransferStatisticsResponse 转移统计响应
type TransferStatisticsResponse struct {
	Statistics TransferStatistics `json:"statistics"` // 转移统计
	Message    string             `json:"message"`    // 状态消息
}

// TransferAutoRequest 自动转移请求
type TransferAutoRequest struct {
	Enable bool   `json:"enable"`           // 是否启用自动转移
	Source string `json:"source,omitempty"` // 源目录（覆盖配置）
	Target string `json:"target,omitempty"` // 目标目录（覆盖配置）
}

// TransferAutoResponse 自动转移响应
type TransferAutoResponse struct {
	TaskID  string `json:"task_id"` // 任务ID
	Status  string `json:"status"`  // 状态
	Message string `json:"message"` // 状态消息
}
