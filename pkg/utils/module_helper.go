package utils

import (
	"errors"
	"fmt"
	"plugin"
	"sync"

	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// Module 模块接口
type Module interface {
	// Name 获取模块名称
	Name() string

	// Version 获取模块版本
	Version() string

	// Description 获取模块描述
	Description() string

	// Initialize 初始化模块
	Initialize() error

	// Close 关闭模块
	Close() error
}

// ModuleHelper 模块帮助类
type ModuleHelper struct {
	modules map[string]Module
	mutex   sync.RWMutex
	logger  *zap.Logger
}

// NewModuleHelper 创建模块帮助类
func NewModuleHelper() *ModuleHelper {
	return &ModuleHelper{
		modules: make(map[string]Module),
		logger:  logger.GetLogger(),
	}
}

// LoadModule 加载单个模块
func (h *ModuleHelper) LoadModule(path string) (Module, error) {
	h.logger.Info("加载模块", zap.String("path", path))

	// 打开插件
	p, err := plugin.Open(path)
	if err != nil {
		h.logger.Error("打开插件失败", zap.Error(err))
		return nil, fmt.Errorf("打开插件失败: %w", err)
	}

	// 查找模块符号
	symbol, err := p.Lookup("Module")
	if err != nil {
		h.logger.Error("查找模块符号失败", zap.Error(err))
		return nil, fmt.Errorf("查找模块符号失败: %w", err)
	}

	// 断言为Module类型
	module, ok := symbol.(Module)
	if !ok {
		h.logger.Error("模块符号类型错误")
		return nil, errors.New("模块符号类型错误")
	}

	// 初始化模块
	if err := module.Initialize(); err != nil {
		h.logger.Error("初始化模块失败", zap.Error(err))
		return nil, fmt.Errorf("初始化模块失败: %w", err)
	}

	// 添加到模块列表
	h.mutex.Lock()
	h.modules[module.Name()] = module
	h.mutex.Unlock()

	h.logger.Info("模块加载成功", zap.String("name", module.Name()))
	return module, nil
}

// LoadModules 加载多个模块
func (h *ModuleHelper) LoadModules(paths []string) ([]Module, error) {
	var modules []Module

	for _, path := range paths {
		module, err := h.LoadModule(path)
		if err != nil {
			h.logger.Error("加载模块失败", zap.String("path", path), zap.Error(err))
			continue
		}
		modules = append(modules, module)
	}

	return modules, nil
}

// GetModule 获取模块
func (h *ModuleHelper) GetModule(name string) (Module, bool) {
	h.mutex.RLock()
	module, ok := h.modules[name]
	h.mutex.RUnlock()
	return module, ok
}

// GetModules 获取所有模块
func (h *ModuleHelper) GetModules() []Module {
	h.mutex.RLock()
	modules := make([]Module, 0, len(h.modules))
	for _, module := range h.modules {
		modules = append(modules, module)
	}
	h.mutex.RUnlock()
	return modules
}

// RemoveModule 移除模块
func (h *ModuleHelper) RemoveModule(name string) error {
	h.mutex.Lock()
	module, ok := h.modules[name]
	if !ok {
		h.mutex.Unlock()
		return fmt.Errorf("模块不存在: %s", name)
	}

	// 关闭模块
	if err := module.Close(); err != nil {
		h.logger.Error("关闭模块失败", zap.Error(err))
	}

	// 从模块列表中移除
	delete(h.modules, name)
	h.mutex.Unlock()

	h.logger.Info("模块移除成功", zap.String("name", name))
	return nil
}

// CloseAll 关闭所有模块
func (h *ModuleHelper) CloseAll() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	for name, module := range h.modules {
		h.logger.Info("关闭模块", zap.String("name", name))
		if err := module.Close(); err != nil {
			h.logger.Error("关闭模块失败", zap.Error(err))
		}
	}

	// 清空模块列表
	h.modules = make(map[string]Module)
	h.logger.Info("所有模块已关闭")
}
