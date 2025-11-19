// Package types 定义动作上下文相关类型
package types

import (
	"time"
)

// ActionContext 动作上下文
// 在动作执行过程中传递的数据和状态信息
type ActionContext struct {
	// 基础信息
	WorkflowID int64                  `json:"workflow_id"`
	Progress   int                    `json:"progress"`
	Message    string                 `json:"message"`
	Data       map[string]interface{} `json:"data"`
	
	// 执行状态
	Status     ExecutionStatus         `json:"status"`
	StartTime  time.Time              `json:"start_time"`
	UpdateTime time.Time              `json:"update_time"`
	Duration   time.Duration          `json:"duration"`
	
	// 错误信息
	Error      *ExecutionError        `json:"error,omitempty"`
	RetryCount int                    `json:"retry_count"`
	
	// 媒体相关
	Medias     []*MediaInfo           `json:"medias"`
	Torrents   []*TorrentInfo         `json:"torrents"`
	
	// 下载相关
	Downloads  []*Download            `json:"downloads"`
	
	// 订阅相关
	Subscribes []*Subscribe           `json:"subscribes"`
	
	// 文件相关
	Files      []*File                `json:"files"`
	
	// 消息相关
	Messages   []*Message             `json:"messages"`
	
	// 事件相关
	Events     []*Event               `json:"events"`
	
	// 插件相关
	Plugins    []*Plugin              `json:"plugins"`
	
	// 备注相关
	Notes      []*Note                `json:"notes"`
	
	// 站点相关
	Sites      []*Site                `json:"sites"`
	
	// 工作流相关
	Workflow   *Workflow              `json:"workflow,omitempty"`
	Execution  *WorkflowExecution    `json:"execution,omitempty"`
	
	// 用户相关
	UserID     uint                   `json:"user_id"`
	Username   string                 `json:"username"`
	UserConfig map[string]interface{} `json:"user_config"`
	
	// 系统配置
	SystemConfig map[string]interface{} `json:"system_config"`
	
	// 元数据
	Metadata   map[string]interface{} `json:"metadata"`
	Tags       []string               `json:"tags"`
	
	// 控制标志
	ShouldStop bool                   `json:"should_stop"`
	ShouldPause bool                  `json:"should_pause"`
	ShouldRetry bool                  `json:"should_retry"`
	
	// 缓存键
	CacheKeys  map[string]bool        `json:"cache_keys"`
	
	// 日志信息
	Logs       []LogEntry             `json:"logs"`
}

// ExecutionStatus 执行状态
type ExecutionStatus string

const (
	StatusPending    ExecutionStatus = "pending"
	StatusRunning    ExecutionStatus = "running"
	StatusCompleted  ExecutionStatus = "completed"
	StatusFailed     ExecutionStatus = "failed"
	StatusCancelled  ExecutionStatus = "cancelled"
	StatusPaused     ExecutionStatus = "paused"
	StatusRetrying   ExecutionStatus = "retrying"
)

// ExecutionError 执行错误
type ExecutionError struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
	StackTrace string                `json:"stack_trace,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Retryable bool                   `json:"retryable"`
}

// LogEntry 日志条目
type LogEntry struct {
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	Source    string                 `json:"source"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// ActionParams 动作参数
type ActionParams map[string]interface{}

// NewActionContext 创建新的动作上下文
func NewActionContext(workflowID int64) *ActionContext {
	return &ActionContext{
		WorkflowID:   workflowID,
		Progress:     0,
		Status:       StatusPending,
		StartTime:    time.Now(),
		UpdateTime:   time.Now(),
		Data:         make(map[string]interface{}),
		Metadata:     make(map[string]interface{}),
		CacheKeys:    make(map[string]bool),
		Logs:         make([]LogEntry, 0),
		SystemConfig: make(map[string]interface{}),
		UserConfig:   make(map[string]interface{}),
		Tags:         make([]string, 0),
	}
}

// Clone 克隆动作上下文
func (ctx *ActionContext) Clone() *ActionContext {
	clone := *ctx
	
	// 深拷贝切片
	if ctx.Medias != nil {
		clone.Medias = make([]*MediaInfo, len(ctx.Medias))
		copy(clone.Medias, ctx.Medias)
	}
	if ctx.Torrents != nil {
		clone.Torrents = make([]*TorrentInfo, len(ctx.Torrents))
		copy(clone.Torrents, ctx.Torrents)
	}
	if ctx.Downloads != nil {
		clone.Downloads = make([]*Download, len(ctx.Downloads))
		copy(clone.Downloads, ctx.Downloads)
	}
	if ctx.Subscribes != nil {
		clone.Subscribes = make([]*Subscribe, len(ctx.Subscribes))
		copy(clone.Subscribes, ctx.Subscribes)
	}
	if ctx.Files != nil {
		clone.Files = make([]*File, len(ctx.Files))
		copy(clone.Files, ctx.Files)
	}
	if ctx.Messages != nil {
		clone.Messages = make([]*Message, len(ctx.Messages))
		copy(clone.Messages, ctx.Messages)
	}
	if ctx.Events != nil {
		clone.Events = make([]*Event, len(ctx.Events))
		copy(clone.Events, ctx.Events)
	}
	if ctx.Plugins != nil {
		clone.Plugins = make([]*Plugin, len(ctx.Plugins))
		copy(clone.Plugins, ctx.Plugins)
	}
	if ctx.Notes != nil {
		clone.Notes = make([]*Note, len(ctx.Notes))
		copy(clone.Notes, ctx.Notes)
	}
	if ctx.Sites != nil {
		clone.Sites = make([]*Site, len(ctx.Sites))
		copy(clone.Sites, ctx.Sites)
	}
	if ctx.Logs != nil {
		clone.Logs = make([]LogEntry, len(ctx.Logs))
		copy(clone.Logs, ctx.Logs)
	}
	if ctx.Tags != nil {
		clone.Tags = make([]string, len(ctx.Tags))
		copy(clone.Tags, ctx.Tags)
	}
	
	// 深拷贝map
	clone.Data = make(map[string]interface{})
	for k, v := range ctx.Data {
		clone.Data[k] = v
	}
	clone.Metadata = make(map[string]interface{})
	for k, v := range ctx.Metadata {
		clone.Metadata[k] = v
	}
	clone.CacheKeys = make(map[string]bool)
	for k, v := range ctx.CacheKeys {
		clone.CacheKeys[k] = v
	}
	clone.SystemConfig = make(map[string]interface{})
	for k, v := range ctx.SystemConfig {
		clone.SystemConfig[k] = v
	}
	clone.UserConfig = make(map[string]interface{})
	for k, v := range ctx.UserConfig {
		clone.UserConfig[k] = v
	}
	
	return &clone
}

// AddMedia 添加媒体信息
func (ctx *ActionContext) AddMedia(media *MediaInfo) {
	if ctx.Medias == nil {
		ctx.Medias = make([]*MediaInfo, 0)
	}
	ctx.Medias = append(ctx.Medias, media)
}

// AddTorrent 添加种子信息
func (ctx *ActionContext) AddTorrent(torrent *TorrentInfo) {
	if ctx.Torrents == nil {
		ctx.Torrents = make([]*TorrentInfo, 0)
	}
	ctx.Torrents = append(ctx.Torrents, torrent)
}

// AddDownload 添加下载信息
func (ctx *ActionContext) AddDownload(download *Download) {
	if ctx.Downloads == nil {
		ctx.Downloads = make([]*Download, 0)
	}
	ctx.Downloads = append(ctx.Downloads, download)
}

// AddFile 添加文件信息
func (ctx *ActionContext) AddFile(file *File) {
	if ctx.Files == nil {
		ctx.Files = make([]*File, 0)
	}
	ctx.Files = append(ctx.Files, file)
}

// AddMessage 添加消息
func (ctx *ActionContext) AddMessage(message *Message) {
	if ctx.Messages == nil {
		ctx.Messages = make([]*Message, 0)
	}
	ctx.Messages = append(ctx.Messages, message)
}

// AddEvent 添加事件
func (ctx *ActionContext) AddEvent(event *Event) {
	if ctx.Events == nil {
		ctx.Events = make([]*Event, 0)
	}
	ctx.Events = append(ctx.Events, event)
}

// AddLog 添加日志
func (ctx *ActionContext) AddLog(level, message, source string, data map[string]interface{}) {
	if ctx.Logs == nil {
		ctx.Logs = make([]LogEntry, 0)
	}
	
	log := LogEntry{
		Level:     level,
		Message:   message,
		Timestamp: time.Now(),
		Source:    source,
		Data:      data,
	}
	
	ctx.Logs = append(ctx.Logs, log)
}

// SetError 设置错误
func (ctx *ActionContext) SetError(code, message string, details map[string]interface{}, retryable bool) {
	ctx.Error = &ExecutionError{
		Code:       code,
		Message:    message,
		Details:    details,
		Timestamp:  time.Now(),
		Retryable:  retryable,
	}
	ctx.Status = StatusFailed
}

// ClearError 清除错误
func (ctx *ActionContext) ClearError() {
	ctx.Error = nil
}

// UpdateProgress 更新进度
func (ctx *ActionContext) UpdateProgress(progress int, message string) {
	ctx.Progress = progress
	ctx.Message = message
	ctx.UpdateTime = time.Now()
	ctx.Duration = ctx.UpdateTime.Sub(ctx.StartTime)
}

// SetCompleted 设置完成状态
func (ctx *ActionContext) SetCompleted(message string) {
	ctx.Status = StatusCompleted
	ctx.Progress = 100
	ctx.Message = message
	ctx.UpdateTime = time.Now()
	ctx.Duration = ctx.UpdateTime.Sub(ctx.StartTime)
}

// SetRunning 设置运行状态
func (ctx *ActionContext) SetRunning(message string) {
	ctx.Status = StatusRunning
	ctx.Message = message
	ctx.UpdateTime = time.Now()
	ctx.Duration = ctx.UpdateTime.Sub(ctx.StartTime)
}

// IsCompleted 是否已完成
func (ctx *ActionContext) IsCompleted() bool {
	return ctx.Status == StatusCompleted
}

// IsRunning 是否正在运行
func (ctx *ActionContext) IsRunning() bool {
	return ctx.Status == StatusRunning
}

// IsFailed 是否失败
func (ctx *ActionContext) IsFailed() bool {
	return ctx.Status == StatusFailed
}

// HasError 是否有错误
func (ctx *ActionContext) HasError() bool {
	return ctx.Error != nil
}

// IsRetryable 是否可重试
func (ctx *ActionContext) IsRetryable() bool {
	return ctx.HasError() && ctx.Error.Retryable
}

// GetDuration 获取执行时长
func (ctx *ActionContext) GetDuration() time.Duration {
	if ctx.UpdateTime.IsZero() {
		return 0
	}
	return ctx.UpdateTime.Sub(ctx.StartTime)
}