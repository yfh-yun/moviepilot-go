package actions

import (
	"time"

	"moviepilot-go/pkg/plugin"
)

// PluginType 插件类型枚举
type PluginType string

const (
	PluginTypeSite        PluginType = "site"        // 站点插件
	PluginTypeIndexer     PluginType = "indexer"     // 索引器插件
	PluginTypeMediaServer PluginType = "mediaserver" // 媒体服务器插件
	PluginTypeNotification PluginType = "notification" // 通知插件
	PluginTypeDownloader  PluginType = "downloader"  // 下载器插件
	PluginTypeScraper     PluginType = "scraper"     // 刮削器插件
	PluginTypeOther       PluginType = "other"       // 其他插件
)

// PluginStatus 插件状态枚举
type PluginStatus string

const (
	PluginStatusNotInstalled PluginStatus = "not_installed" // 未安装
	PluginStatusInstalled    PluginStatus = "installed"    // 已安装
	PluginStatusRunning      PluginStatus = "running"      // 运行中
	PluginStatusStopped      PluginStatus = "stopped"      // 已停止
	PluginStatusError        PluginStatus = "error"        // 错误
	PluginStatusUpdating     PluginStatus = "updating"     // 更新中
	PluginStatusStarting     PluginStatus = "starting"     // 启动中
	PluginStatusStopping     PluginStatus = "stopping"     // 停止中
)

// PluginInvokeType 插件调用类型
type PluginInvokeType string

const (
	PluginInvokeTypeFunction PluginInvokeType = "function" // 函数调用
	PluginInvokeTypeMethod   PluginInvokeType = "method"   // 方法调用
	PluginInvokeTypeCommand  PluginInvokeType = "command"  // 命令调用
)

// PluginInvokeRequest 插件调用请求
type PluginInvokeRequest struct {
	// 插件ID
	PluginID string `json:"plugin_id" binding:"required" validate:"min=1,max=100"`

	// 插件类型
	PluginType PluginType `json:"plugin_type" binding:"required" validate:"plugin_type"`

	// 调用类型
	InvokeType PluginInvokeType `json:"invoke_type" binding:"required" validate:"oneof=function method command"`

	// 调用目标（函数名、方法名、命令名）
	Target string `json:"target" binding:"required" validate:"min=1,max=200"`

	// 调用参数
	Params map[string]interface{} `json:"params"`

	// 超时时间（秒）
	Timeout int `json:"timeout,omitempty" validate:"min=1,max=3600"`

	// 是否异步调用
	Async bool `json:"async,omitempty"`

	// 调用上下文
	Context *PluginContext `json:"context,omitempty"`
}

// PluginContext 插件调用上下文
type PluginContext struct {
	// 用户ID
	UserID string `json:"user_id,omitempty"`

	// 请求ID
	RequestID string `json:"request_id,omitempty"`

	// 会话ID
	SessionID string `json:"session_id,omitempty"`

	// 调用时间戳
	Timestamp time.Time `json:"timestamp"`

	// 元数据
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// PluginInvokeResponse 插件调用响应
type PluginInvokeResponse struct {
	// 调用是否成功
	Success bool `json:"success"`

	// 调用结果
	Result interface{} `json:"result,omitempty"`

	// 错误信息
	Error *PluginError `json:"error,omitempty"`

	// 执行时间（毫秒）
	ExecutionTime int64 `json:"execution_time"`

	// 调用ID
	CallID string `json:"call_id,omitempty"`

	// 插件信息
	PluginInfo *PluginInfo `json:"plugin_info,omitempty"`

	// 异步任务ID（如果是异步调用）
	TaskID string `json:"task_id,omitempty"`
}

// PluginError 插件错误信息
type PluginError struct {
	// 错误代码
	Code string `json:"code"`

	// 错误消息
	Message string `json:"message"`

	// 错误详情
	Details interface{} `json:"details,omitempty"`

	// 错误类型
	ErrorType string `json:"error_type,omitempty"`

	// 原始错误
	OriginalError string `json:"original_error,omitempty"`
}

// PluginInfo 插件信息
type PluginInfo struct {
	// 插件ID
	ID string `json:"id"`

	// 插件名称
	Name string `json:"name"`

	// 插件版本
	Version string `json:"version"`

	// 插件类型
	Type PluginType `json:"type"`

	// 插件描述
	Description string `json:"description,omitempty"`

	// 插件作者
	Author string `json:"author,omitempty"`

	// 插件状态
	Status PluginStatus `json:"status"`

	// 插件路径
	Path string `json:"path,omitempty"`

	// 插件语言
	Language string `json:"language,omitempty"` // go, python

	// 插件依赖
	Dependencies []string `json:"dependencies,omitempty"`

	// 插件配置信息
	ConfigSchema map[string]interface{} `json:"config_schema,omitempty"`

	// 可用方法/函数列表
	Methods []string `json:"methods,omitempty"`

	// 健康状态
	Healthy bool `json:"healthy"`

	// 注册时间
	RegisterTime time.Time `json:"register_time,omitempty"`

	// 上次更新时间
	LastUpdateTime time.Time `json:"last_update_time,omitempty"`

	// 连接信息（gRPC地址等）
	ConnectionInfo string `json:"connection_info,omitempty"`
}

// PluginListResponse 插件列表响应
type PluginListResponse struct {
	// 插件总数
	Total int `json:"total"`

	// 插件列表
	Plugins []*PluginInfo `json:"plugins"`

	// 各类型插件数量统计
	TypeStats map[string]int `json:"type_stats,omitempty"`

	// 各状态插件数量统计
	StatusStats map[string]int `json:"status_stats,omitempty"`
}

// PluginManagerStats 插件管理器统计信息
type PluginManagerStats struct {
	// 已安装插件总数
	InstalledCount int `json:"installed_count"`

	// 运行中插件数量
	RunningCount int `json:"running_count"`

	// 错误插件数量
	ErrorCount int `json:"error_count"`

	// Python插件数量
	PythonPluginCount int `json:"python_plugin_count"`

	// Go插件数量
	GoPluginCount int `json:"go_plugin_count"`

	// 插件调用总次数
	TotalCalls int64 `json:"total_calls"`

	// 成功调用次数
	SuccessCalls int64 `json:"success_calls"`

	// 失败调用次数
	FailedCalls int64 `json:"failed_calls"`

	// 平均调用响应时间（毫秒）
	AvgResponseTime float64 `json:"avg_response_time"`

	// 最后调用时间
	LastCallTime time.Time `json:"last_call_time,omitempty"`
}

// PluginHealthCheck 插件健康检查结果
type PluginHealthCheck struct {
	// 插件ID
	PluginID string `json:"plugin_id"`

	// 是否健康
	Healthy bool `json:"healthy"`

	// 检查时间
	CheckTime time.Time `json:"check_time"`

	// 检查详情
	Details map[string]interface{} `json:"details,omitempty"`

	// 错误信息
	Error string `json:"error,omitempty"`

	// 响应时间（毫秒）
	ResponseTime int64 `json:"response_time"`
}

// PluginMethodInfo 插件方法信息
type PluginMethodInfo struct {
	// 方法名
	Name string `json:"name"`

	// 方法描述
	Description string `json:"description,omitempty"`

	// 参数规范
	Parameters []*MethodParameter `json:"parameters,omitempty"`

	// 返回值规范
	Returns *MethodReturn `json:"returns,omitempty"`

	// 是否需要认证
	RequiresAuth bool `json:"requires_auth"`

	// 最大执行时间（秒）
	MaxExecutionTime int `json:"max_execution_time,omitempty"`

	// 是否支持异步调用
	AsyncSupported bool `json:"async_supported"`
}

// MethodParameter 方法参数信息
type MethodParameter struct {
	// 参数名
	Name string `json:"name"`

	// 参数类型
	Type string `json:"type"` // string, number, boolean, object, array, etc.

	// 是否必需
	Required bool `json:"required"`

	// 参数描述
	Description string `json:"description,omitempty"`

	// 默认值
	Default interface{} `json:"default,omitempty"`

	// 验证规则
	Validation interface{} `json:"validation,omitempty"`
}

// MethodReturn 方法返回值信息
type MethodReturn struct {
	// 返回值类型
	Type string `json:"type"` // string, number, boolean, object, array, etc.

	// 返回值描述
	Description string `json:"description,omitempty"`

	// 示例返回值
	Example interface{} `json:"example,omitempty"`
}

// PluginConfigUpdate 插件配置更新请求
type PluginConfigUpdate struct {
	// 插件ID
	PluginID string `json:"plugin_id" binding:"required"`

	// 配置项
	Config map[string]interface{} `json:"config" binding:"required"`

	// 是否立即应用
	ApplyNow bool `json:"apply_now,omitempty"`
}

// PluginInstallRequest 插件安装请求
type PluginInstallRequest struct {
	// 插件ID
	PluginID string `json:"plugin_id" binding:"required"`

	// 插件类型
	PluginType PluginType `json:"plugin_type" binding:"required"`

	// 插件源
	Source string `json:"source,omitempty"` // git_url, local_path, etc.

	// 版本
	Version string `json:"version,omitempty"`

	// 安装选项
	Options map[string]interface{} `json:"options,omitempty"`
}

// PluginUninstallRequest 插件卸载请求
type PluginUninstallRequest struct {
	// 插件ID
	PluginID string `json:"plugin_id" binding:"required"`

	// 是否保留配置
	KeepConfig bool `json:"keep_config,omitempty"`

	// 是否保留数据
	KeepData bool `json:"keep_data,omitempty"`
}

// PluginUpdateRequest 插件更新请求
type PluginUpdateRequest struct {
	// 插件ID
	PluginID string `json:"plugin_id" binding:"required"`

	// 更新源
	Source string `json:"source,omitempty"`

	// 版本
	Version string `json:"version,omitempty"`

	// 是否保留配置
	KeepConfig bool `json:"keep_config,omitempty"`
}

// PluginStartStopRequest 插件启动/停止请求
type PluginStartStopRequest struct {
	// 插件ID
	PluginID string `json:"plugin_id" binding:"required"`

	// 启动/停止选项
	Options map[string]interface{} `json:"options,omitempty"`
}

// PluginMetadata 插件元数据
type PluginMetadata struct {
	// 插件ID
	ID string `json:"id"`

	// 插件名称
	Name string `json:"name"`

	// 插件版本
	Version string `json:"version"`

	// 插件类型
	Type PluginType `json:"type"`

	// 插件描述
	Description string `json:"description,omitempty"`

	// 插件作者
	Author string `json:"author,omitempty"`

	// 插件主页
	Homepage string `json:"homepage,omitempty"`

	// 许可证
	License string `json:"license,omitempty"`

	// 依赖项
	Dependencies map[string]string `json:"dependencies,omitempty"`

	// 兼容性信息
	Compatibility map[string]string `json:"compatibility,omitempty"`

	// 最低版本要求
	MinVersion string `json:"min_version,omitempty"`

	// 图标URL
	IconURL string `json:"icon_url,omitempty"`

	// 标签
	Tags []string `json:"tags,omitempty"`
}

// PluginSearchRequest 插件搜索请求
type PluginSearchRequest struct {
	// 搜索关键词
	Keyword string `json:"keyword,omitempty"`

	// 插件类型筛选
	Type PluginType `json:"type,omitempty"`

	// 分页参数
	Page int `json:"page,omitempty" validate:"min=1"`
	Limit int `json:"limit,omitempty" validate:"min=1,max=100"`

	// 排序字段
	SortBy string `json:"sort_by,omitempty" validate:"oneof=name version popularity rating"`

	// 排序方向
	SortOrder string `json:"sort_order,omitempty" validate:"oneof=asc desc"`
}

// PluginSearchResponse 插件搜索响应
type PluginSearchResponse struct {
	// 总结果数
	Total int `json:"total"`

	// 插件列表
	Plugins []*PluginMetadata `json:"plugins"`

	// 当前页码
	Page int `json:"page"`

	// 每页数量
	Limit int `json:"limit"`
}

// AsyncTaskStatus 异步任务状态
type AsyncTaskStatus string

const (
	AsyncTaskStatusPending   AsyncTaskStatus = "pending"   // 等待中
	AsyncTaskStatusRunning   AsyncTaskStatus = "running"   // 运行中
	AsyncTaskStatusCompleted AsyncTaskStatus = "completed" // 已完成
	AsyncTaskStatusFailed    AsyncTaskStatus = "failed"    // 失败
	AsyncTaskStatusCancelled AsyncTaskStatus = "cancelled" // 已取消
)

// AsyncTaskInfo 异步任务信息
type AsyncTaskInfo struct {
	// 任务ID
	TaskID string `json:"task_id"`

	// 任务状态
	Status AsyncTaskStatus `json:"status"`

	// 任务进度（0-100）
	Progress float64 `json:"progress"`

	// 任务描述
	Description string `json:"description,omitempty"`

	// 开始时间
	StartTime time.Time `json:"start_time,omitempty"`

	// 更新时间
	UpdateTime time.Time `json:"update_time,omitempty"`

	// 完成时间
	CompleteTime *time.Time `json:"complete_time,omitempty"`

	// 任务结果
	Result interface{} `json:"result,omitempty"`

	// 错误信息
	Error *PluginError `json:"error,omitempty"`

	// 插件信息
	PluginInfo *PluginInfo `json:"plugin_info,omitempty"`

	// 调用详情
	InvokeDetail *PluginInvokeRequest `json:"invoke_detail,omitempty"`
}

// TaskStatusRequest 任务状态查询请求
type TaskStatusRequest struct {
	// 任务ID
	TaskID string `json:"task_id" binding:"required"`

	// 是否返回详细信息
	Detailed bool `json:"detailed,omitempty"`
}

// TaskCancelRequest 任务取消请求
type TaskCancelRequest struct {
	// 任务ID
	TaskID string `json:"task_id" binding:"required"`

	// 是否强制取消
	Force bool `json:"force,omitempty"`
}

// TaskListRequest 任务列表查询请求
type TaskListRequest struct {
	// 插件ID筛选
	PluginID string `json:"plugin_id,omitempty"`

	// 状态筛选
	Status AsyncTaskStatus `json:"status,omitempty"`

	// 时间范围筛选
	StartTimeFrom *time.Time `json:"start_time_from,omitempty"`
	StartTimeTo *time.Time `json:"start_time_to,omitempty"`

	// 分页参数
	Page int `json:"page,omitempty" validate:"min=1"`
	Limit int `json:"limit,omitempty" validate:"min=1,max=100"`

	// 排序字段
	SortBy string `json:"sort_by,omitempty" validate:"oneof=start_time update_time status"`

	// 排序方向
	SortOrder string `json:"sort_order,omitempty" validate:"oneof=asc desc"`
}

// TaskListResponse 任务列表响应
type TaskListResponse struct {
	// 总任务数
	Total int `json:"total"`

	// 任务列表
	Tasks []*AsyncTaskInfo `json:"tasks"`

	// 当前页码
	Page int `json:"page"`

	// 每页数量
	Limit int `json:"limit"`
}

// PluginEvent 插件事件
type PluginEvent struct {
	// 事件ID
	EventID string `json:"event_id"`

	// 事件类型
	EventType string `json:"event_type"`

	// 插件ID
	PluginID string `json:"plugin_id"`

	// 事件数据
	Data map[string]interface{} `json:"data,omitempty"`

	// 事件时间
	EventTime time.Time `json:"event_time"`

	// 事件来源
	Source string `json:"source,omitempty"`

	// 事件严重级别
	Severity string `json:"severity,omitempty"` // info, warning, error, critical
}

// PluginMetrics 插件性能指标
type PluginMetrics struct {
	// 插件ID
	PluginID string `json:"plugin_id"`

	// 调用次数
	CallCount int64 `json:"call_count"`

	// 平均响应时间（毫秒）
	AvgResponseTime float64 `json:"avg_response_time"`

	// 最长响应时间（毫秒）
	MaxResponseTime int64 `json:"max_response_time"`

	// 最短响应时间（毫秒）
	MinResponseTime int64 `json:"min_response_time"`

	// 失败率（百分比）
	FailureRate float64 `json:"failure_rate"`

	// 内存使用（字节）
	MemoryUsage int64 `json:"memory_usage,omitempty"`

	// CPU使用率（百分比）
	CpuUsage float64 `json:"cpu_usage,omitempty"`

	// 最后更新时间
	UpdateTime time.Time `json:"update_time"`
}

// ConvertPluginStatus 转换插件状态从pkg/plugin到actions
func ConvertPluginStatus(status plugin.PluginStatus) PluginStatus {
	switch status {
	case plugin.PluginStatusNotInstalled:
		return PluginStatusNotInstalled
	case plugin.PluginStatusInstalled:
		return PluginStatusInstalled
	case plugin.PluginStatusRunning:
		return PluginStatusRunning
	case plugin.PluginStatusStopped:
		return PluginStatusStopped
	case plugin.PluginStatusError:
		return PluginStatusError
	case plugin.PluginStatusUpdating:
		return PluginStatusUpdating
	case plugin.PluginStatusStarting:
		return PluginStatusStarting
	case plugin.PluginStatusStopping:
		return PluginStatusStopping
	default:
		return PluginStatusUnknown
	}
}

// PluginStatusUnknown 未知插件状态
const PluginStatusUnknown PluginStatus = "unknown"

// SortOrderAsc 升序排序
const SortOrderAsc string = "asc"

// SortOrderDesc 降序排序
const SortOrderDesc string = "desc"
