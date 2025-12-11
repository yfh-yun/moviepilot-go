package bootstrap

// ModuleRegistry 模块注册表
type ModuleRegistry struct {
	modules []Module
}

// Module 模块接口
type Module interface {
	ID() string
	Name() string
	Init() error
	Start() error
	Stop() error
}

// initModules 初始化内部模块
func initModules(app *App) error {
	// 创建模块注册表
	registry := &ModuleRegistry{
		modules: make([]Module, 0),
	}

	// 简化实现：直接返回空的模块注册表
	app.Modules = registry
	return nil
}

// Start 启动所有模块
func (r *ModuleRegistry) Start() error {
	for _, module := range r.modules {
		if err := module.Start(); err != nil {
			return err
		}
	}
	return nil
}

// Stop 停止所有模块
func (r *ModuleRegistry) Stop() error {
	for _, module := range r.modules {
		if err := module.Stop(); err != nil {
			return err
		}
	}
	return nil
}
