package plugin

import (
	"context"
	"time"
)

// EventType 事件类型枚举
type EventType string

// 定义插件相关的事件类型
const (
	// 插件生命周期事件
	EventTypePluginLoaded    EventType = "plugin.loaded"    // 插件加载完成
	EventTypePluginStarted   EventType = "plugin.started"   // 插件启动成功
	EventTypePluginStopped   EventType = "plugin.stopped"   // 插件停止成功
	EventTypePluginError     EventType = "plugin.error"     // 插件发生错误
	EventTypePluginConfigChanged EventType = "plugin.config_changed" // 插件配置变更

	// 插件命令事件
	EventTypePluginCommandExecuted EventType = "plugin.command_executed" // 插件命令执行完成
	EventTypePluginAPIExecuted     EventType = "plugin.api_executed"     // 插件API执行完成

	// 系统事件
	EventTypeSystemStarted     EventType = "system.started"     // 系统启动完成
	EventTypeSystemShutdown    EventType = "system.shutdown"    // 系统开始关闭
	EventTypeSystemConfigChanged EventType = "system.config_changed" // 系统配置变更

	// 媒体相关事件
	EventTypeMediaAdded      EventType = "media.added"      // 媒体添加
	EventTypeMediaUpdated    EventType = "media.updated"    // 媒体更新
	EventTypeMediaDeleted    EventType = "media.deleted"    // 媒体删除
	EventTypeMediaDownloaded EventType = "media.downloaded" // 媒体下载完成
	EventTypeMediaTransfered EventType = "media.transfered" // 媒体转移完成

	// 任务相关事件
	EventTypeTaskCreated     EventType = "task.created"     // 任务创建
	EventTypeTaskStarted     EventType = "task.started"     // 任务开始
	EventTypeTaskCompleted   EventType = "task.completed"   // 任务完成
	EventTypeTaskFailed      EventType = "task.failed"      // 任务失败
	EventTypeTaskCanceled    EventType = "task.canceled"    // 任务取消
)

// Event 事件结构体
type Event struct {
	ID        string                 `json:"id"`        // 事件ID
	Type      EventType              `json:"type"`      // 事件类型
	Source    string                 `json:"source"`    // 事件源（插件ID或系统）
	Timestamp time.Time              `json:"timestamp"` // 事件发生时间
	Data      map[string]interface{} `json:"data"`      // 事件数据
}

// EventHandler 事件处理器类型
type EventHandler func(ctx context.Context, event *Event) error

// EventFilter 事件过滤器类型
type EventFilter func(event *Event) bool

// EventSubscription 事件订阅信息
type EventSubscription struct {
	ID        string        // 订阅ID
	EventType EventType     // 订阅的事件类型
	Handler   EventHandler  // 事件处理器
	Filter    EventFilter   // 事件过滤器
}
