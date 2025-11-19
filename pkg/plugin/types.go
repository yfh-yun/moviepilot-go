package plugin

import (
	"time"
)

// Event 事件结构
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// EventHandler 事件处理器接口
type EventHandler interface {
	HandleEvent(event Event) error
	GetEventSubscriptions() []string
}

// EventBus 事件总线接口
type EventBus interface {
	Subscribe(eventType string, handler EventHandler)
	Unsubscribe(eventType string, handler EventHandler)
	Publish(event Event)
}

// ConfigForm 配置表单
type ConfigForm struct {
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Fields      []FormField `json:"fields"`
}

// FormField 表单字段
type FormField struct {
	Name        string      `json:"name"`
	Label       string      `json:"label"`
	Type        string      `json:"type"` // text, number, boolean, select, etc.
	Required    bool        `json:"required"`
	Default     interface{} `json:"default"`
	Options     []Option    `json:"options,omitempty"`
	Description string      `json:"description,omitempty"`
}

// Option 选项
type Option struct {
	Label string      `json:"label"`
	Value interface{} `json:"value"`
}

// APIRoute API路由
type APIRoute struct {
	Method  string `json:"method"`
	Path    string `json:"path"`
	Handler string `json:"handler"`
}

// Command 命令
type Command struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Usage       string `json:"usage"`
	Handler     string `json:"handler"`
}

// Service 服务
type Service struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Handler     string `json:"handler"`
}

// PluginMetadata 插件元数据
type PluginMetadata struct {
	Author      string    `json:"author"`
	Homepage    string    `json:"homepage"`
	Repository  string    `json:"repository"`
	License     string    `json:"license"`
	Tags        []string  `json:"tags"`
	Keywords    []string  `json:"keywords"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	MinVersion  string    `json:"min_version"`
	MaxVersion  string    `json:"max_version"`
	Dependencies []string  `json:"dependencies"`
}

// PluginInfo 插件信息
type PluginInfo struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	State       string                 `json:"state"`
	Config      map[string]interface{} `json:"config"`
	Metadata    PluginMetadata         `json:"metadata"`
	LastError   error                  `json:"last_error,omitempty"`
	LoadTime    time.Time              `json:"load_time"`
}

// Logger 日志接口
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Fatal(msg string, args ...interface{})
}

// Config 配置结构
type Config struct {
	Plugins PluginConfig `json:"plugins"`
}

// PluginConfig 插件配置
type PluginConfig struct {
	Native  NativePluginConfig  `json:"native"`
	Python  PythonPluginConfig  `json:"python"`
	Web     WebPluginConfig     `json:"web"`
	Plugins []string            `json:"plugins"`
}

// NativePluginConfig 原生插件配置
type NativePluginConfig struct {
	Path    string `json:"path"`
	Enabled bool   `json:"enabled"`
}

// PythonPluginConfig Python插件配置
type PythonPluginConfig struct {
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Timeout int    `json:"timeout"`
	Enabled bool   `json:"enabled"`
}

// WebPluginConfig Web插件配置
type WebPluginConfig struct {
	Enabled bool   `json:"enabled"`
	Path    string `json:"path"`
}