package main

import (
	"fmt"

	"github.com/yfh-yun/moviepilot-go/pkg/plugin"
)

// HelloPlugin 示例插件
type HelloPlugin struct {
	id      string
	name    string
	version string
	state   plugin.PluginState
	config  map[string]interface{}
}

// NewPlugin 创建插件实例
func NewPlugin() plugin.Plugin {
	return &HelloPlugin{
		id:      "hello-plugin",
		name:    "Hello Plugin",
		version: "1.0.0",
		state:   plugin.StateUnloaded,
		config:  make(map[string]interface{}),
	}
}

// ID 返回插件ID
func (p *HelloPlugin) ID() string {
	return p.id
}

// Name 返回插件名称
func (p *HelloPlugin) Name() string {
	return p.name
}

// Version 返回插件版本
func (p *HelloPlugin) Version() string {
	return p.version
}

// Type 返回插件类型
func (p *HelloPlugin) Type() plugin.PluginType {
	return plugin.PluginTypeNative
}

// Description 返回插件描述
func (p *HelloPlugin) Description() string {
	return "A simple hello world plugin example"
}

// Initialize 初始化插件
func (p *HelloPlugin) Initialize(config map[string]interface{}) error {
	p.config = config
	p.state = plugin.StateInitialized
	
	message := "Hello Plugin initialized"
	if greeting, ok := config["greeting"].(string); ok {
		message = fmt.Sprintf("Hello Plugin initialized with greeting: %s", greeting)
	}
	fmt.Println(message)
	
	return nil
}

// Start 启动插件
func (p *HelloPlugin) Start() error {
	p.state = plugin.StateRunning
	
	greeting := "Hello, World!"
	if g, ok := p.config["greeting"].(string); ok {
		greeting = g
	}
	
	fmt.Printf("Hello Plugin started: %s\n", greeting)
	return nil
}

// Stop 停止插件
func (p *HelloPlugin) Stop() error {
	p.state = plugin.StateStopped
	fmt.Println("Hello Plugin stopped")
	return nil
}

// Destroy 销毁插件
func (p *HelloPlugin) Destroy() error {
	p.state = plugin.StateUnloaded
	fmt.Println("Hello Plugin destroyed")
	return nil
}

// GetState 获取插件状态
func (p *HelloPlugin) GetState() plugin.PluginState {
	return p.state
}

// HandleEvent 处理事件
func (p *HelloPlugin) HandleEvent(event plugin.Event) error {
	fmt.Printf("Hello Plugin received event: %s\n", event.Type)
	return nil
}

// GetConfigForm 获取配置表单
func (p *HelloPlugin) GetConfigForm() *plugin.ConfigForm {
	return &plugin.ConfigForm{
		Title:       "Hello Plugin Configuration",
		Description: "Configure the hello world plugin",
		Fields: []plugin.FormField{
			{
				Name:        "greeting",
				Label:       "Greeting Message",
				Type:        "text",
				Required:    true,
				Default:     "Hello, World!",
				Description: "The greeting message to display",
			},
			{
				Name:        "enabled",
				Label:       "Enable Plugin",
				Type:        "boolean",
				Required:    false,
				Default:     true,
				Description: "Whether to enable the plugin",
			},
		},
	}
}

// GetAPIRoutes 获取API路由
func (p *HelloPlugin) GetAPIRoutes() []plugin.APIRoute {
	return []plugin.APIRoute{
		{
			Method:  "GET",
			Path:    "/hello",
			Handler: "handleHello",
		},
		{
			Method:  "POST",
			Path:    "/hello/config",
			Handler: "handleConfig",
		},
	}
}

// GetCommands 获取命令
func (p *HelloPlugin) GetCommands() []plugin.Command {
	return []plugin.Command{
		{
			Name:        "hello",
			Description: "Say hello",
			Usage:       "hello [name]",
			Handler:     "handleHelloCommand",
		},
	}
}

// GetServices 获取服务
func (p *HelloPlugin) GetServices() []plugin.Service {
	return []plugin.Service{
		{
			Name:        "hello-service",
			Description: "Hello world service",
			Handler:     "handleHelloService",
		},
	}
}

// 为了编译为.so文件，需要一个空的main函数
func main() {}