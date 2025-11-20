// Package types 定义动作系统使用的数据模型
package types

import (
	"time"
)

// WorkflowExecution 工作流执行记录
type WorkflowExecution struct {
	ID           int64                  `json:"id"`
	WorkflowID   int                    `json:"workflow_id"`
	WorkflowName string                 `json:"workflow_name"`
	Trigger      string                 `json:"trigger"`      // manual, schedule, event
	Status       string                 `json:"status"`       // running, completed, failed, cancelled, paused
	StartTime    time.Time              `json:"start_time"`
	EndTime      *time.Time             `json:"end_time,omitempty"`
	Duration     time.Duration          `json:"duration"`
	Progress     int                    `json:"progress"`     // 0-100
	Message      string                 `json:"message"`
	Error        string                 `json:"error,omitempty"`
	RetryCount   int                    `json:"retry_count"`
	CurrentStep  int                    `json:"current_step"`
	TotalSteps   int                    `json:"total_steps"`
	
	// 执行参数
	Params       map[string]interface{} `json:"params"`
	Context      *ActionContext         `json:"context"`
	
	// 执行结果
	Result       map[string]interface{} `json:"result"`
	Output       string                 `json:"output"`
	
	// 用户信息
	UserID       uint                   `json:"user_id"`
	Username     string                 `json:"username"`
	
	// 时间戳
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
}

// WorkflowSchedule 工作流调度
type WorkflowSchedule struct {
	ID           int64     `json:"id"`
	WorkflowID   int       `json:"workflow_id"`
	WorkflowName string    `json:"workflow_name"`
	
	// 调度配置
	Cron         string    `json:"cron"`          // cron表达式
	Timezone     string    `json:"timezone"`      // 时区
	Enabled      bool      `json:"enabled"`       // 是否启用
	
	// 执行配置
	MaxRuns      int       `json:"max_runs"`      // 最大执行次数
	CurrentRuns  int       `json:"current_runs"`  // 当前执行次数
	Timeout      int       `json:"timeout"`       // 超时时间（秒）
	Retries      int       `json:"retries"`       // 重试次数
	
	// 状态信息
	Status       string    `json:"status"`        // active, inactive, paused, error
	LastRun      *time.Time `json:"last_run,omitempty"`
	NextRun      *time.Time `json:"next_run,omitempty"`
	LastSuccess  *time.Time `json:"last_success,omitempty"`
	LastError    string    `json:"last_error,omitempty"`
	
	// 时间戳
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// WorkflowTemplate 工作流模板
type WorkflowTemplate struct {
	ID           int                    `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Category     string                 `json:"category"`     // media, download, file, system
	Tags         []string               `json:"tags"`
	Icon         string                 `json:"icon"`
	Version      string                 `json:"version"`
	Author       string                 `json:"author"`
	
	// 模板配置
	Config       string                 `json:"config"`       // JSON配置
	Actions      []WorkflowAction       `json:"actions"`      // 动作列表
	Variables    []WorkflowVariable     `json:"variables"`    // 变量定义
	
	// 使用统计
	DownloadCount int64                  `json:"download_count"`
	UseCount      int64                  `json:"use_count"`
	Rating        float64                `json:"rating"`
	
	// 状态信息
	Status        string                 `json:"status"`       // draft, published, deprecated
	Featured      bool                   `json:"featured"`     // 是否推荐
	Official      bool                   `json:"official"`     // 是否官方
	
	// 时间戳
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	PublishedAt   *time.Time             `json:"published_at,omitempty"`
}

// WorkflowVariable 工作流变量
type WorkflowVariable struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"`         // string, int, float, bool, array, object
	Description  string      `json:"description"`
	Required     bool        `json:"required"`
	DefaultValue interface{} `json:"default_value"`
	Options      []string    `json:"options,omitempty"`      // 选项列表
	Validation   string      `json:"validation,omitempty"`    // 验证规则
	Example      string      `json:"example,omitempty"`       // 示例值
}

// WorkflowAction 工作流动作
type WorkflowAction struct {
	ID       int                    `json:"id"`
	Type     string                 `json:"type"`        // action类型
	Name     string                 `json:"name"`        // 动作名称
	Description string              `json:"description"` // 动作描述
	Config   map[string]interface{} `json:"config"`      // 动作配置
	Order    int                    `json:"order"`       // 执行顺序
	Enabled  bool                   `json:"enabled"`     // 是否启用
	Timeout  int                    `json:"timeout"`     // 超时时间（秒）
	Retry    int                    `json:"retry"`       // 重试次数
	Condition string                 `json:"condition,omitempty"` // 执行条件
	
	// 输入输出映射
	InputMapping  map[string]string `json:"input_mapping,omitempty"`   // 输入映射
	OutputMapping map[string]string `json:"output_mapping,omitempty"`  // 输出映射
	
	// 错误处理
	ErrorHandling string            `json:"error_handling,omitempty"` // stop, continue, retry
	FallbackAction *WorkflowAction   `json:"fallback_action,omitempty"` // 备用动作
}

// ActionConfig 动作配置
type ActionConfig struct {
	ActionID     string                 `json:"action_id"`
	ActionType   string                 `json:"action_type"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Version      string                 `json:"version"`
	
	// 配置模式
	Schema       map[string]interface{} `json:"schema"`       // JSON Schema
	Defaults     map[string]interface{} `json:"defaults"`     // 默认值
	Required     []string               `json:"required"`     // 必填字段
	
	// 权限和安全
	Permissions  []string               `json:"permissions"`  // 所需权限
	Sandbox      bool                   `json:"sandbox"`      // 是否沙箱运行
	ResourceLimits *ResourceLimits     `json:"resource_limits,omitempty"` // 资源限制
	
	// 元数据
	Tags         []string               `json:"tags"`
	Category     string                 `json:"category"`
	Author       string                 `json:"author"`
	Homepage     string                 `json:"homepage"`
	
	// 依赖
	Dependencies []string              `json:"dependencies"` // 依赖的动作类型
	Conflicts    []string               `json:"conflicts"`    // 冲突的动作类型
	
	// 时间戳
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// ResourceLimits 资源限制
type ResourceLimits struct {
	MaxMemory    int64 `json:"max_memory"`    // 最大内存（字节）
	MaxCPU       int   `json:"max_cpu"`        // 最大CPU使用率（百分比）
	MaxDuration  int64 `json:"max_duration"`   // 最大执行时间（秒）
	MaxFileSize  int64 `json:"max_file_size"`  // 最大文件大小（字节）
	MaxNetwork   int64 `json:"max_network"`   // 最大网络流量（字节）
}

// ActionExecution 动作执行记录
type ActionExecution struct {
	ID           int64                  `json:"id"`
	WorkflowID   int64                  `json:"workflow_id"`
	ActionID     string                 `json:"action_id"`
	ActionName   string                 `json:"action_name"`
	ActionType   string                 `json:"action_type"`
	
	// 执行信息
	Status       string                 `json:"status"`        // running, completed, failed, cancelled, skipped
	StartTime    time.Time              `json:"start_time"`
	EndTime      *time.Time             `json:"end_time,omitempty"`
	Duration     time.Duration          `json:"duration"`
	
	// 输入输出
	Input        map[string]interface{} `json:"input"`
	Output       map[string]interface{} `json:"output"`
	Result       interface{}            `json:"result"`
	
	// 状态信息
	Success      bool                   `json:"success"`
	Message      string                 `json:"message"`
	Error        string                 `json:"error,omitempty"`
	RetryCount   int                    `json:"retry_count"`
	
	// 性能指标
	MemoryUsage  int64                  `json:"memory_usage"`  // 内存使用量（字节）
	CPUUsage     float64                `json:"cpu_usage"`     // CPU使用率
	NetworkIO    int64                  `json:"network_io"`    // 网络IO（字节）
	
	// 时间戳
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// ActionMetrics 动作指标
type ActionMetrics struct {
	ActionID     string    `json:"action_id"`
	ActionName   string    `json:"action_name"`
	ActionType   string    `json:"action_type"`
	
	// 执行统计
	TotalExecutions    int64   `json:"total_executions"`
	SuccessfulExecutions int64 `json:"successful_executions"`
	FailedExecutions    int64   `json:"failed_executions"`
	SuccessRate         float64 `json:"success_rate"`
	
	// 性能统计
	AverageExecutionTime float64 `json:"average_execution_time"`
	MinExecutionTime     float64 `json:"min_execution_time"`
	MaxExecutionTime     float64 `json:"max_execution_time"`
	P95ExecutionTime     float64 `json:"p95_execution_time"`
	P99ExecutionTime     float64 `json:"p99_execution_time"`
	
	// 资源统计
	AverageMemoryUsage  int64   `json:"average_memory_usage"`
	PeakMemoryUsage     int64   `json:"peak_memory_usage"`
	AverageCPUUsage     float64 `json:"average_cpu_usage"`
	PeakCPUUsage        float64 `json:"peak_cpu_usage"`
	
	// 错误统计
	ErrorTypes          map[string]int64 `json:"error_types"`
	RetryRate           float64          `json:"retry_rate"`
	TimeoutRate         float64          `json:"timeout_rate"`
	
	// 时间范围
	StartTime           time.Time        `json:"start_time"`
	EndTime             time.Time        `json:"end_time"`
	LastUpdateTime      time.Time        `json:"last_update_time"`
}

// EventData 事件数据
type EventData map[string]interface{}

// EventFilter 事件过滤器
type EventFilter struct {
	UserID    string    `json:"user_id"`
	Type      string    `json:"type"`
	Source    string    `json:"source"`
	Level     string    `json:"level"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Tags      []string  `json:"tags"`
	Keyword   string    `json:"keyword"`
	Limit     int       `json:"limit"`
	Offset    int       `json:"offset"`
}

// MessageFilter 消息过滤器
type MessageFilter struct {
	UserID    uint      `json:"user_id"`
	Type      string    `json:"type"`
	Channel   string    `json:"channel"`
	Status    string    `json:"status"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Keyword   string    `json:"keyword"`
	Limit     int       `json:"limit"`
	Offset    int       `json:"offset"`
}

// DownloadFilter 下载过滤器
type DownloadFilter struct {
	UserID     uint      `json:"user_id"`
	Status     string    `json:"status"`
	Type       string    `json:"type"`
	Downloader string    `json:"downloader"`
	SiteID     string    `json:"site_id"`
	MediaID    int       `json:"media_id"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	Keyword    string    `json:"keyword"`
	Limit      int       `json:"limit"`
	Offset     int       `json:"offset"`
}

// SubscribeFilter 订阅过滤器
type SubscribeFilter struct {
	UserID     uint      `json:"user_id"`
	Type       string    `json:"type"`
	Status     string    `json:"status"`
	Username   string    `json:"username"`
	SiteID     string    `json:"site_id"`
	MediaID    int       `json:"media_id"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	Keyword    string    `json:"keyword"`
	Limit      int       `json:"limit"`
	Offset     int       `json:"offset"`
}

// FileFilter 文件过滤器
type FileFilter struct {
	UserID      uint      `json:"user_id"`
	Type        string    `json:"type"`
	MediaType   string    `json:"media_type"`
	Status      string    `json:"status"`
	Path        string    `json:"path"`
	Extension   string    `json:"extension"`
	MinSize     int64     `json:"min_size"`
	MaxSize     int64     `json:"max_size"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Keyword     string    `json:"keyword"`
	Limit       int       `json:"limit"`
	Offset      int       `json:"offset"`
}

// PluginFilter 插件过滤器
type PluginFilter struct {
	UserID     uint      `json:"user_id"`
	Status     string    `json:"status"`
	Type       string    `json:"type"`
	Author     string    `json:"author"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	Keyword    string    `json:"keyword"`
	Limit      int       `json:"limit"`
	Offset     int       `json:"offset"`
}

// SiteFilter 站点过滤器
type SiteFilter struct {
	UserID     uint      `json:"user_id"`
	Type       string    `json:"type"`
	Status     string    `json:"status"`
	Keyword    string    `json:"keyword"`
	Limit      int       `json:"limit"`
	Offset     int       `json:"offset"`
}

// WorkflowFilter 工作流过滤器
type WorkflowFilter struct {
	UserID     uint      `json:"user_id"`
	Status     string    `json:"status"`
	Trigger    string    `json:"trigger"`
	Category   string    `json:"category"`
	Tags       []string  `json:"tags"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	Keyword    string    `json:"keyword"`
	Limit      int       `json:"limit"`
	Offset     int       `json:"offset"`
}