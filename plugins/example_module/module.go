package main

// ExamplePlugin 示例插件结构�?type ExamplePlugin struct {
	Name string
}

// GetName 获取插件名称
func (p *ExamplePlugin) GetName() string {
	return p.Name
}

// Process 处理方法
func (p *ExamplePlugin) Process(data string) string {
	return "Processed: " + data
}

// Exported variables for plugin symbol lookup
var (
	// PluginInstance 插件实例
	PluginInstance = &ExamplePlugin{Name: "example_module"}
	
	// Version 插件版本
	Version = "1.0.0"
	
	// Description 插件描述
	Description = "示例模块插件"
)
