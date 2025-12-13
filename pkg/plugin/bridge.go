package plugin

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"

	"moviepilot-go/internal/proto/plugin"
	"moviepilot-go/pkg/logger"
)

// PythonPlugin Python插件代理，实现Plugin接口
type PythonPlugin struct {
	id          string
	name        string
	version     string
	description string
	icon        string
	author      string
	authorURL   string
	label       string
	state       State
	config      map[string]any
	logger      *zap.Logger
	grpcClient  *GRPCClient
}

// NewPythonPlugin 创建Python插件代理
func NewPythonPlugin(id, name, version, description, icon, author, authorURL, label string, grpcClient *GRPCClient) *PythonPlugin {
	return &PythonPlugin{
		id:          id,
		name:        name,
		version:     version,
		description: description,
		icon:        icon,
		author:      author,
		authorURL:   authorURL,
		label:       label,
		state:       StateDisabled,
		config:      make(map[string]any),
		logger:      logger.GetLogger(),
		grpcClient:  grpcClient,
	}
}

// ID 获取插件ID
func (p *PythonPlugin) ID() string {
	return p.id
}

// Name 获取插件名称
func (p *PythonPlugin) Name() string {
	return p.name
}

// Version 获取插件版本
func (p *PythonPlugin) Version() string {
	return p.version
}

// Description 获取插件描述
func (p *PythonPlugin) Description() string {
	return p.description
}

// Icon 获取插件图标
func (p *PythonPlugin) Icon() string {
	return p.icon
}

// Author 获取插件作者
func (p *PythonPlugin) Author() string {
	return p.author
}

// AuthorURL 获取作者主页
func (p *PythonPlugin) AuthorURL() string {
	return p.authorURL
}

// Label 获取插件标签
func (p *PythonPlugin) Label() string {
	return p.label
}

// Init 初始化插件
func (p *PythonPlugin) Init(ctx context.Context, cfg map[string]any) error {
	p.logger.Info("初始化Python插件", zap.String("id", p.id))
	p.config = cfg

	// 通过gRPC调用Python插件服务的StartPlugin方法
	if err := p.grpcClient.StartPlugin(ctx, p.id, cfg); err != nil {
		p.logger.Error("初始化Python插件失败", zap.String("id", p.id), zap.Error(err))
		return err
	}

	p.state = StateEnabled
	return nil
}

// Stop 停止插件
func (p *PythonPlugin) Stop(ctx context.Context) error {
	p.logger.Info("停止Python插件", zap.String("id", p.id))

	// 通过gRPC调用Python插件服务的StopPlugin方法
	if err := p.grpcClient.StopPlugin(ctx, p.id, false); err != nil {
		p.logger.Error("停止Python插件失败", zap.String("id", p.id), zap.Error(err))
		return err
	}

	p.state = StateDisabled
	return nil
}

// State 获取插件状态
func (p *PythonPlugin) State(ctx context.Context) State {
	return p.state
}

// SetState 设置插件状态
func (p *PythonPlugin) SetState(ctx context.Context, state State) error {
	p.logger.Info("设置Python插件状态", zap.String("id", p.id), zap.String("state", string(state)))

	// 调用相应的gRPC方法
	var err error
	if state == StateEnabled {
		err = p.grpcClient.StartPlugin(ctx, p.id, p.config)
	} else {
		err = p.grpcClient.StopPlugin(ctx, p.id, false)
	}

	if err != nil {
		p.logger.Error("设置Python插件状态失败", zap.String("id", p.id), zap.String("state", string(state)), zap.Error(err))
		return err
	}

	p.state = state
	return nil
}

// Config 获取插件配置
func (p *PythonPlugin) Config(ctx context.Context) (map[string]any, error) {
	return p.config, nil
}

// SetConfig 设置插件配置
func (p *PythonPlugin) SetConfig(ctx context.Context, cfg map[string]any) error {
	p.logger.Info("设置Python插件配置", zap.String("id", p.id))

	// 将配置转换为gRPC ConfigItem格式
	configItems := make([]any, 0, len(cfg))
	for k, v := range cfg {
		configItems = append(configItems, map[string]any{
			"key":   k,
			"value": v,
		})
	}

	// 将配置转换为JSON格式
	configJSON, err := json.Marshal(configItems)
	if err != nil {
		p.logger.Error("序列化插件配置失败", zap.String("id", p.id), zap.Error(err))
		return err
	}

	// 通过gRPC调用Python插件服务的ExecutePlugin方法，执行SetConfig操作
	if _, err := p.grpcClient.ExecutePlugin(ctx, p.id, "SetConfig", configJSON, 30); err != nil {
		p.logger.Error("设置Python插件配置失败", zap.String("id", p.id), zap.Error(err))
		return err
	}

	p.config = cfg
	return nil
}

// Commands 获取插件命令
func (p *PythonPlugin) Commands(ctx context.Context) ([]Command, error) {
	p.logger.Info("获取Python插件命令", zap.String("id", p.id))

	// 通过gRPC调用Python插件服务的ExecutePlugin方法，执行Commands操作
	result, err := p.grpcClient.ExecutePlugin(ctx, p.id, "Commands", nil, 30)
	if err != nil {
		p.logger.Error("获取Python插件命令失败", zap.String("id", p.id), zap.Error(err))
		return make([]Command, 0), nil
	}

	// 解析结果
	var commands []Command
	if err := json.Unmarshal(result, &commands); err != nil {
		p.logger.Error("解析插件命令失败", zap.String("id", p.id), zap.Error(err))
		return make([]Command, 0), nil
	}

	return commands, nil
}

// APIs 获取插件API
func (p *PythonPlugin) APIs(ctx context.Context) ([]API, error) {
	p.logger.Info("获取Python插件API", zap.String("id", p.id))

	// 通过gRPC调用Python插件服务的ExecutePlugin方法，执行APIs操作
	result, err := p.grpcClient.ExecutePlugin(ctx, p.id, "APIs", nil, 30)
	if err != nil {
		p.logger.Error("获取Python插件API失败", zap.String("id", p.id), zap.Error(err))
		return make([]API, 0), nil
	}

	// 解析结果
	var apis []API
	if err := json.Unmarshal(result, &apis); err != nil {
		p.logger.Error("解析插件API失败", zap.String("id", p.id), zap.Error(err))
		return make([]API, 0), nil
	}

	return apis, nil
}

// Services 获取插件服务
func (p *PythonPlugin) Services(ctx context.Context) ([]Service, error) {
	p.logger.Info("获取Python插件服务", zap.String("id", p.id))

	// 通过gRPC调用Python插件服务的ExecutePlugin方法，执行Services操作
	result, err := p.grpcClient.ExecutePlugin(ctx, p.id, "Services", nil, 30)
	if err != nil {
		p.logger.Error("获取Python插件服务失败", zap.String("id", p.id), zap.Error(err))
		return make([]Service, 0), nil
	}

	// 解析结果
	var services []Service
	if err := json.Unmarshal(result, &services); err != nil {
		p.logger.Error("解析插件服务失败", zap.String("id", p.id), zap.Error(err))
		return make([]Service, 0), nil
	}

	return services, nil
}

// Actions 获取插件动作
func (p *PythonPlugin) Actions(ctx context.Context) ([]ActionGroup, error) {
	p.logger.Info("获取Python插件动作", zap.String("id", p.id))

	// 通过gRPC调用Python插件服务的ExecutePlugin方法，执行Actions操作
	result, err := p.grpcClient.ExecutePlugin(ctx, p.id, "Actions", nil, 30)
	if err != nil {
		p.logger.Error("获取Python插件动作失败", zap.String("id", p.id), zap.Error(err))
		return make([]ActionGroup, 0), nil
	}

	// 解析结果
	var actions []ActionGroup
	if err := json.Unmarshal(result, &actions); err != nil {
		p.logger.Error("解析插件动作失败", zap.String("id", p.id), zap.Error(err))
		return make([]ActionGroup, 0), nil
	}

	return actions, nil
}

// DashboardMeta 获取仪表盘元信息
func (p *PythonPlugin) DashboardMeta(ctx context.Context) ([]DashboardMeta, error) {
	p.logger.Info("获取Python插件仪表盘元信息", zap.String("id", p.id))

	// 通过gRPC调用Python插件服务的ExecutePlugin方法，执行DashboardMeta操作
	result, err := p.grpcClient.ExecutePlugin(ctx, p.id, "DashboardMeta", nil, 30)
	if err != nil {
		p.logger.Error("获取Python插件仪表盘元信息失败", zap.String("id", p.id), zap.Error(err))
		return make([]DashboardMeta, 0), nil
	}

	// 解析结果
	var dashboardMeta []DashboardMeta
	if err := json.Unmarshal(result, &dashboardMeta); err != nil {
		p.logger.Error("解析仪表盘元信息失败", zap.String("id", p.id), zap.Error(err))
		return make([]DashboardMeta, 0), nil
	}

	return dashboardMeta, nil
}

// Dashboard 获取仪表盘
func (p *PythonPlugin) Dashboard(ctx context.Context, key, userAgent string) (*Dashboard, error) {
	p.logger.Info("获取Python插件仪表盘", zap.String("id", p.id), zap.String("key", key))

	// 构建请求参数
	params := map[string]any{
		"key":       key,
		"userAgent": userAgent,
	}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		p.logger.Error("序列化仪表盘请求参数失败", zap.String("id", p.id), zap.Error(err))
		return nil, fmt.Errorf("序列化仪表盘请求参数失败: %w", err)
	}

	// 通过gRPC调用Python插件服务的ExecutePlugin方法，执行Dashboard操作
	result, err := p.grpcClient.ExecutePlugin(ctx, p.id, "Dashboard", paramsJSON, 30)
	if err != nil {
		p.logger.Error("获取Python插件仪表盘失败", zap.String("id", p.id), zap.Error(err))
		return nil, fmt.Errorf("获取Python插件仪表盘失败: %w", err)
	}

	// 解析结果
	var dashboard Dashboard
	if err := json.Unmarshal(result, &dashboard); err != nil {
		p.logger.Error("解析仪表盘失败", zap.String("id", p.id), zap.Error(err))
		return nil, fmt.Errorf("解析仪表盘失败: %w", err)
	}

	return &dashboard, nil
}

// RenderMode 获取渲染模式
func (p *PythonPlugin) RenderMode(ctx context.Context) (string, string) {
	p.logger.Info("获取Python插件渲染模式", zap.String("id", p.id))

	// 通过gRPC调用Python插件服务的ExecutePlugin方法，执行RenderMode操作
	result, err := p.grpcClient.ExecutePlugin(ctx, p.id, "RenderMode", nil, 30)
	if err != nil {
		p.logger.Error("获取Python插件渲染模式失败", zap.String("id", p.id), zap.Error(err))
		return "", ""
	}

	// 解析结果
	var renderMode []string
	if err := json.Unmarshal(result, &renderMode); err != nil {
		p.logger.Error("解析渲染模式失败", zap.String("id", p.id), zap.Error(err))
		return "", ""
	}

	if len(renderMode) >= 2 {
		return renderMode[0], renderMode[1]
	}

	return "", ""
}

// OnEvent 事件处理方法
func (p *PythonPlugin) OnEvent(ctx context.Context, event *Event) error {
	// 这里可以实现Python插件的事件处理逻辑
	// 目前简单实现，后续可以通过gRPC调用Python插件的事件处理方法
	p.logger.Debug("Python插件接收到事件", 
		zap.String("plugin_id", p.id),
		zap.String("event_type", string(event.Type)),
		zap.String("event_source", event.Source))
	return nil
}

// PythonPluginBridge Python插件桥接，用于管理Python插件
type PythonPluginBridge struct {
	logger     *zap.Logger
	plugins    map[string]*PythonPlugin
	grpcClient *GRPCClient
}

// NewPythonPluginBridge 创建Python插件桥接
func NewPythonPluginBridge(grpcClient *GRPCClient) *PythonPluginBridge {
	return &PythonPluginBridge{
		logger:     logger.GetLogger(),
		plugins:    make(map[string]*PythonPlugin),
		grpcClient: grpcClient,
	}
}

// RefreshPlugins 从Python插件服务刷新插件列表
func (b *PythonPluginBridge) RefreshPlugins() error {
	b.logger.Info("刷新Python插件列表")
	
	// 调用Python插件服务获取所有插件信息
	plugins, err := b.grpcClient.ListPlugins(context.Background(), plugin.PluginType_PLUGIN_TYPE_UNSPECIFIED, plugin.PluginStatus_PLUGIN_STATUS_UNSPECIFIED)
	if err != nil {
		b.logger.Error("获取Python插件列表失败", zap.Error(err))
		return err
	}
	
	// 注册插件
	for _, pluginInfo := range plugins {
		// 检查插件是否已注册
		if _, exists := b.plugins[pluginInfo.Id]; exists {
			continue
		}
		
		// 创建并注册Python插件代理
		pythonPlugin := NewPythonPlugin(
			pluginInfo.Id,
			pluginInfo.Name,
			pluginInfo.Version,
			pluginInfo.Description,
			pluginInfo.Icon,
			pluginInfo.Author,
			pluginInfo.Homepage,
			pluginInfo.Metadata["label"],
			b.grpcClient,
		)
		
		// 根据插件状态设置初始状态
		if pluginInfo.Status == plugin.PluginStatus_PLUGIN_STATUS_RUNNING {
			pythonPlugin.state = StateEnabled
		} else {
			pythonPlugin.state = StateDisabled
		}
		
		b.RegisterPythonPlugin(pythonPlugin)
	}
	
	return nil
}

// RegisterPythonPlugin 注册Python插件
func (b *PythonPluginBridge) RegisterPythonPlugin(plugin *PythonPlugin) {
	b.plugins[plugin.ID()] = plugin
	b.logger.Info("Python插件已注册", zap.String("id", plugin.ID()), zap.String("name", plugin.Name()))
}

// UnregisterPythonPlugin 注销Python插件
func (b *PythonPluginBridge) UnregisterPythonPlugin(id string) error {
	if _, exists := b.plugins[id]; !exists {
		return fmt.Errorf("python插件不存在: %s", id)
	}
	delete(b.plugins, id)
	b.logger.Info("Python插件已注销", zap.String("id", id))
	return nil
}

// GetPythonPlugin 获取Python插件
func (b *PythonPluginBridge) GetPythonPlugin(id string) (*PythonPlugin, error) {
	plugin, exists := b.plugins[id]
	if !exists {
		return nil, fmt.Errorf("python插件不存在: %s", id)
	}
	return plugin, nil
}

// ListPythonPlugins 列出所有Python插件
func (b *PythonPluginBridge) ListPythonPlugins() []*PythonPlugin {
	plugins := make([]*PythonPlugin, 0, len(b.plugins))
	for _, plugin := range b.plugins {
		plugins = append(plugins, plugin)
	}
	return plugins
}
