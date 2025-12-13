package plugin

import (
	"context"
)

// State 插件状态枚举
type State string

const (
	StateDisabled State = "disabled"
	StateEnabled  State = "enabled"
)

// EventPublisher 事件发布者接口
type EventPublisher interface {
	// PublishEvent 发布事件
	PublishEvent(ctx context.Context, event *Event) error
	
	// PublishEventAsync 异步发布事件
	PublishEventAsync(event *Event)
}

// EventSubscriber 事件订阅者接口
type EventSubscriber interface {
	// SubscribeEvent 订阅事件
	SubscribeEvent(eventType EventType, handler EventHandler, filter EventFilter) (string, error)
	
	// UnsubscribeEvent 取消订阅
	UnsubscribeEvent(subscriptionID string) error
	
	// SubscribeMultipleEvents 订阅多个事件
	SubscribeMultipleEvents(eventTypes []EventType, handler EventHandler, filter EventFilter) ([]string, error)
	
	// UnsubscribeAllEvents 取消所有订阅
	UnsubscribeAllEvents() error
}

// EventManager 事件管理器接口，同时具备发布和订阅功能
type EventManager interface {
	EventPublisher
	EventSubscriber
	
	// GetSubscriptions 获取所有订阅
	GetSubscriptions() []*EventSubscription
	
	// GetSubscriptionsByEventType 获取指定事件类型的订阅
	GetSubscriptionsByEventType(eventType EventType) []*EventSubscription
	
	// Close 关闭事件管理器
	Close() error
}

// Plugin 插件接口定义（Go 侧抽象）
type Plugin interface {
	// ID 获取插件ID
	ID() string
	// Name 获取插件名称
	Name() string
	// Version 获取插件版本
	Version() string
	// Description 获取插件描述
	Description() string
	// Icon 获取插件图标
	Icon() string
	// Author 获取插件作者
	Author() string
	// AuthorURL 获取作者主页
	AuthorURL() string
	// Label 获取插件标签
	Label() string

	// Init 初始化插件
	Init(ctx context.Context, cfg map[string]any) error
	// Stop 停止插件
	Stop(ctx context.Context) error

	// State 获取插件状态
	State(ctx context.Context) State
	// SetState 设置插件状态
	SetState(ctx context.Context, state State) error

	// Config 获取插件配置
	Config(ctx context.Context) (map[string]any, error)
	// SetConfig 设置插件配置
	SetConfig(ctx context.Context, cfg map[string]any) error

	// Commands 获取插件命令
	Commands(ctx context.Context) ([]Command, error)
	// APIs 获取插件API
	APIs(ctx context.Context) ([]API, error)
	// Services 获取插件服务
	Services(ctx context.Context) ([]Service, error)
	// Actions 获取插件动作
	Actions(ctx context.Context) ([]ActionGroup, error)
	// DashboardMeta 获取仪表盘元信息
	DashboardMeta(ctx context.Context) ([]DashboardMeta, error)
	// Dashboard 获取仪表盘
	Dashboard(ctx context.Context, key, userAgent string) (*Dashboard, error)
	// RenderMode 获取渲染模式
	RenderMode(ctx context.Context) (string, string) // mode, distPath
	
	// OnEvent 事件处理方法，插件可以实现此方法来处理事件
	OnEvent(ctx context.Context, event *Event) error
}

// Command 插件命令结构
type Command struct {
	Cmd   string         `json:"cmd"`
	Event string         `json:"event"`
	Desc  string         `json:"desc"`
	Data  map[string]any `json:"data"`
}

// API 插件API结构
type API struct {
	Path           string   `json:"path"`
	Methods        []string `json:"methods"`
	Summary        string   `json:"summary"`
	Description    string   `json:"description"`
	AllowAnonymous bool     `json:"allow_anonymous"`
	Auth           string   `json:"auth"`
}

// Service 插件服务结构
type Service struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Trigger  string         `json:"trigger"`
	Func     string         `json:"func"`
	Args     []any          `json:"args"`
	FuncArgs map[string]any `json:"func_kwargs"`
}

// ActionGroup 插件动作组结构
type ActionGroup struct {
	Name    string   `json:"name"`
	Actions []Action `json:"actions"`
}

// Action 插件动作结构
type Action struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Args        []map[string]any `json:"args"`
}

// DashboardMeta 仪表盘元信息结构
type DashboardMeta struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

// Dashboard 仪表盘结构
type Dashboard struct {
	Name       string           `json:"name"`
	Key        string           `json:"key"`
	RenderMode string           `json:"render_mode"`
	Attrs      map[string]any   `json:"attrs"`
	Cols       map[string]any   `json:"cols"`
	Elements   []map[string]any `json:"elements"`
}

// PluginInfo 插件信息结构
type PluginInfo struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	Version       string          `json:"version"`
	Description   string          `json:"description"`
	Icon          string          `json:"icon"`
	Author        string          `json:"author"`
	AuthorURL     string          `json:"author_url"`
	Label         string          `json:"label"`
	State         string          `json:"state"`
	HasPage       bool            `json:"has_page"`
	HasUpdate     bool            `json:"has_update"`
	IsLocal       bool            `json:"is_local"`
	RepoURL       string          `json:"repo_url"`
	InstallCount  int             `json:"install_count"`
	Config        map[string]any  `json:"config"`
	Commands      []Command       `json:"commands"`
	APIs          []API           `json:"apis"`
	Services      []Service       `json:"services"`
	Actions       []ActionGroup   `json:"actions"`
	DashboardMeta []DashboardMeta `json:"dashboard_meta"`
}

// ConfigStore 插件配置存储接口
type ConfigStore interface {
	// Get 获取插件配置
	Get(ctx context.Context, pluginID string) (map[string]any, error)
	// Set 设置插件配置
	Set(ctx context.Context, pluginID string, config map[string]any) error
	// Delete 删除插件配置
	Delete(ctx context.Context, pluginID string) error
}

// DataStore 插件数据存储接口
type DataStore interface {
	// Get 获取插件数据
	Get(ctx context.Context, pluginID, key string) (any, error)
	// Set 设置插件数据
	Set(ctx context.Context, pluginID, key string, value any) error
	// Delete 删除插件数据
	Delete(ctx context.Context, pluginID, key string) error
	// DeleteAll 删除插件所有数据
	DeleteAll(ctx context.Context, pluginID string) error
}
