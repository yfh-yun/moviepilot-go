package modules

import (
	"fmt"
	"sync"

	"go.uber.org/zap"

	"moviepilot-go/pkg/cache"
)

// Module 模块接口
// 原Python: Module in app/core/module.py
type Module interface {
	// GetID 获取模块ID
	GetID() string

	// GetName 获取模块名称
	GetName() string

	// GetPriority 获取优先级（数字越小优先级越高）
	GetPriority() int

	// Initialize 初始化模块
	Initialize() error

	// Stop 停止模块
	Stop() error
}

// BaseModule 基础模块
type BaseModule struct {
	ID       string
	Name     string
	Priority int
}

// GetID 获取模块ID
func (m *BaseModule) GetID() string {
	return m.ID
}

// GetName 获取模块名称
func (m *BaseModule) GetName() string {
	return m.Name
}

// GetPriority 获取优先级
func (m *BaseModule) GetPriority() int {
	if m.Priority == 0 {
		return 10 // 默认优先级
	}
	return m.Priority
}

// Initialize 初始化模块
func (m *BaseModule) Initialize() error {
	return nil
}

// Stop 停止模块
func (m *BaseModule) Stop() error {
	return nil
}

// Manager 模块管理器
// 原Python: ModuleManager in app/core/module.py
type Manager struct {
	modules  map[string]Module   // 模块ID -> 模块
	methods  map[string][]Module // 方法名 -> 模块列表
	mu       sync.RWMutex
	logger   *zap.Logger
	cache    cache.Backend // 缓存后端
	cacheTTL int64         // 缓存过期时间（秒）
}

// NewManager 创建模块管理器
func NewManager(logger *zap.Logger, cache cache.Backend) *Manager {
	return &Manager{
		modules:  make(map[string]Module),
		methods:  make(map[string][]Module),
		logger:   logger,
		cache:    cache,
		cacheTTL: 24 * 3600, // 默认缓存24小时
	}
}

// Register 注册模块
func (m *Manager) Register(module Module) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	moduleID := module.GetID()

	if _, exists := m.modules[moduleID]; exists {
		return fmt.Errorf("模块已存在: %s", moduleID)
	}

	// 初始化模块
	if err := module.Initialize(); err != nil {
		return fmt.Errorf("初始化模块失败: %w", err)
	}

	m.modules[moduleID] = module

	// 清理模块相关缓存
	if m.cache != nil {
		if err := m.cache.Clear("module_method"); err != nil {
			m.logger.Error("清理模块方法缓存失败", zap.Error(err))
		}
	}

	m.logger.Info("注册模块",
		zap.String("module_id", moduleID),
		zap.String("module_name", module.GetName()))

	return nil
}

// Unregister 注销模块
func (m *Manager) Unregister(moduleID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	module, exists := m.modules[moduleID]
	if !exists {
		return fmt.Errorf("模块不存在: %s", moduleID)
	}

	// 停止模块
	if err := module.Stop(); err != nil {
		m.logger.Error("停止模块失败",
			zap.String("module_id", moduleID),
			zap.Error(err))
	}

	delete(m.modules, moduleID)

	// 从方法映射中删除
	for method, modules := range m.methods {
		for i, mod := range modules {
			if mod.GetID() == moduleID {
				m.methods[method] = append(modules[:i], modules[i+1:]...)
				break
			}
		}
	}

	// 清理模块相关缓存
	if m.cache != nil {
		if err := m.cache.Clear("module_method"); err != nil {
			m.logger.Error("清理模块方法缓存失败", zap.Error(err))
		}
	}

	m.logger.Info("注销模块", zap.String("module_id", moduleID))

	return nil
}

// Get 获取模块
func (m *Manager) Get(moduleID string) (Module, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	module, exists := m.modules[moduleID]
	if !exists {
		return nil, fmt.Errorf("模块不存在: %s", moduleID)
	}

	return module, nil
}

// GetAll 获取所有模块
func (m *Manager) GetAll() []Module {
	m.mu.RLock()
	defer m.mu.RUnlock()

	modules := make([]Module, 0, len(m.modules))
	for _, module := range m.modules {
		modules = append(modules, module)
	}

	return modules
}

// RegisterMethod 注册模块方法
// 将模块注册到特定方法名下，用于模块调用
func (m *Manager) RegisterMethod(methodName string, module Module) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.methods[methodName] = append(m.methods[methodName], module)

	// 清理该方法的缓存
	if m.cache != nil {
		cacheKey := fmt.Sprintf("running_modules_by_method:%s", methodName)
		if err := m.cache.Delete("module_method", cacheKey); err != nil {
			m.logger.Error("清理方法缓存失败", zap.String("method", methodName), zap.Error(err))
		}
	}

	m.logger.Debug("注册模块方法",
		zap.String("method", methodName),
		zap.String("module_id", module.GetID()))
}

// GetRunningModules 获取运行中的模块（按优先级排序）
// 原Python: get_running_modules(method)
func (m *Manager) GetRunningModules(methodName string) []Module {
	// 生成缓存键
	cacheKey := fmt.Sprintf("running_modules_by_method:%s", methodName)

	// 检查缓存
	if m.cache != nil {
		var cachedModules []Module
		hit, err := m.cache.Get("module_method", cacheKey, &cachedModules)
		if err == nil && hit {
			return cachedModules
		}
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	modules, exists := m.methods[methodName]
	if !exists {
		return []Module{}
	}

	// 复制模块列表
	result := make([]Module, len(modules))
	copy(result, modules)

	// 按优先级排序（冒泡排序，简单实现）
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].GetPriority() > result[j].GetPriority() {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	// 更新缓存
	if m.cache != nil {
		m.cache.Set("module_method", cacheKey, result, m.cacheTTL)
	}

	return result
}

// StopAll 停止所有模块
func (m *Manager) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for moduleID, module := range m.modules {
		if err := module.Stop(); err != nil {
			m.logger.Error("停止模块失败",
				zap.String("module_id", moduleID),
				zap.Error(err))
		}
	}

	m.logger.Info("所有模块已停止")
}

// GetModuleCount 获取模块数量
func (m *Manager) GetModuleCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.modules)
}
