package module

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"moviepilot-go/pkg/logger"
)

// ModuleManager 模块管理器接口
type ModuleManager interface {
	// RegisterModule 注册模块
	RegisterModule(module Module) error

	// UnregisterModule 注销模块
	UnregisterModule(moduleID string) error

	// GetModule 获取模块
	GetModule(moduleID string) (Module, error)

	// ListModules 列出所有模块
	ListModules() []Module

	// InitializeAll 初始化所有模块
	InitializeAll(config map[string]interface{}) error

	// StartAll 启动所有模块
	StartAll() error

	// StopAll 停止所有模块
	StopAll() error

	// StartModule 启动指定模块
	StartModule(moduleID string) error

	// StopModule 停止指定模块
	StopModule(moduleID string) error

	// RestartModule 重启指定模块
	RestartModule(moduleID string) error

	// GetModuleStatus 获取模块状态
	GetModuleStatus(moduleID string) (string, error)

	// GetModuleInfo 获取模块详细信息
	GetModuleInfo(moduleID string) (*ModuleInfo, error)

	// UpdateModuleConfig 更新模块配置
	UpdateModuleConfig(moduleID string, config map[string]interface{}) error

	// CheckModuleDependencies 检查模块依赖
	CheckModuleDependencies(moduleID string) error

	// GetRunningModules 获取运行中的模块
	GetRunningModules() []Module

	// GetStoppedModules 获取已停止的模块
	GetStoppedModules() []Module

	// GetErrorModules 获取出错的模块
	GetErrorModules() []Module

	// EnableModule 启用模块
	EnableModule(moduleID string) error

	// DisableModule 禁用模块
	DisableModule(moduleID string) error

	// IsModuleEnabled 检查模块是否启用
	IsModuleEnabled(moduleID string) (bool, error)

	// GetModuleStats 获取模块统计信息
	GetModuleStats() map[string]interface{}
}

// ModuleManagerImpl 模块管理器实现
type ModuleManagerImpl struct {
	modules     map[string]Module
	logger      logger.Logger
	config      map[string]interface{}
	initialized bool
	mutex       sync.RWMutex
}

// NewModuleManager 创建模块管理器
func NewModuleManager(logger logger.Logger) *ModuleManagerImpl {
	return &ModuleManagerImpl{
		modules: make(map[string]Module),
		logger:  logger,
		config:  make(map[string]interface{}),
	}
}

// RegisterModule 注册模块
func (mm *ModuleManagerImpl) RegisterModule(module Module) error {
	if module == nil {
		return errors.New("module cannot be nil")
	}

	moduleInfo := module.GetInfo()
	if moduleInfo == nil || moduleInfo.ID == "" {
		return errors.New("invalid module ID")
	}

	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	if _, exists := mm.modules[moduleInfo.ID]; exists {
		return fmt.Errorf("%w: %s", ErrModuleAlreadyExists, moduleInfo.ID)
	}

	mm.modules[moduleInfo.ID] = module
	mm.logger.Info("Module registered", "module", moduleInfo.ID, "name", moduleInfo.Name)
	return nil
}

// UnregisterModule 注销模块
func (mm *ModuleManagerImpl) UnregisterModule(moduleID string) error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	module, exists := mm.modules[moduleID]
	if !exists {
		return fmt.Errorf("%w: %s", ErrModuleNotFound, moduleID)
	}

	// 停止模块
	if module.IsRunning() {
		if err := module.Stop(); err != nil {
			mm.logger.Warn("Failed to stop module during unregistration", "module", moduleID, "error", err.Error())
		}
	}

	delete(mm.modules, moduleID)
	mm.logger.Info("Module unregistered", "module", moduleID)
	return nil
}

// GetModule 获取模块
func (mm *ModuleManagerImpl) GetModule(moduleID string) (Module, error) {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	module, exists := mm.modules[moduleID]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrModuleNotFound, moduleID)
	}

	return module, nil
}

// ListModules 列出所有模块
func (mm *ModuleManagerImpl) ListModules() []Module {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	modules := make([]Module, 0, len(mm.modules))
	for _, module := range mm.modules {
		modules = append(modules, module)
	}

	// 按优先级排序
	sort.Slice(modules, func(i, j int) bool {
		return modules[i].GetInfo().Priority < modules[j].GetInfo().Priority
	})

	return modules
}

// InitializeAll 初始化所有模块
func (mm *ModuleManagerImpl) InitializeAll(config map[string]interface{}) error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	// 保存配置
	if config != nil {
		mm.config = config
	}

	// 获取按优先级排序的模块
	modules := mm.ListModules()

	// 初始化每个模块
	for _, module := range modules {
		moduleID := module.GetInfo().ID
		moduleConfig := make(map[string]interface{})

		// 提取模块特定配置
		if moduleConfigs, ok := mm.config["modules"].(map[string]interface{}); ok {
			if mc, ok := moduleConfigs[moduleID].(map[string]interface{}); ok {
				moduleConfig = mc
			}
		}

		// 初始化模块
		if err := module.Initialize(moduleConfig, mm.logger); err != nil {
			mm.logger.Error("Failed to initialize module", "module", moduleID, "error", err.Error())
			return fmt.Errorf("failed to initialize module %s: %w", moduleID, err)
		}
	}

	mm.initialized = true
	mm.logger.Info("All modules initialized successfully")
	return nil
}

// StartAll 启动所有模块
func (mm *ModuleManagerImpl) StartAll() error {
	if !mm.initialized {
		return errors.New("module manager not initialized")
	}

	// 计算模块启动顺序
	orderedModules, err := mm.calculateStartupOrder()
	if err != nil {
		mm.logger.Error("Failed to calculate module startup order", "error", err.Error())
		return err
	}

	// 启动每个模块
	for _, module := range orderedModules {
		moduleID := module.GetInfo().ID

		// 跳过已禁用的模块
		if module.GetStatus() == StatusDisabled {
			mm.logger.Debug("Skipping disabled module", "module", moduleID)
			continue
		}

		// 启动模块
		if err := module.Start(); err != nil {
			mm.logger.Error("Failed to start module", "module", moduleID, "error", err.Error())
			return fmt.Errorf("failed to start module %s: %w", moduleID, err)
		}
	}

	mm.logger.Info("All modules started successfully")
	return nil
}

// StopAll 停止所有模块
func (mm *ModuleManagerImpl) StopAll() error {
	// 获取所有模块（按反向优先级排序，确保先启动的后停止）
	modules := mm.ListModules()
	sort.Slice(modules, func(i, j int) bool {
		return modules[i].GetInfo().Priority > modules[j].GetInfo().Priority
	})

	var lastError error

	// 停止每个模块
	for _, module := range modules {
		moduleID := module.GetInfo().ID
		if module.IsRunning() {
			if err := module.Stop(); err != nil {
				mm.logger.Error("Failed to stop module", "module", moduleID, "error", err.Error())
				lastError = err
			}
		}
	}

	if lastError != nil {
		return fmt.Errorf("some modules failed to stop: %w", lastError)
	}

	mm.logger.Info("All modules stopped successfully")
	return nil
}

// StartModule 启动指定模块
func (mm *ModuleManagerImpl) StartModule(moduleID string) error {
	module, err := mm.GetModule(moduleID)
	if err != nil {
		return err
	}

	// 检查模块是否已运行
	if module.IsRunning() {
		mm.logger.Debug("Module already running", "module", moduleID)
		return nil
	}

	// 检查模块是否禁用
	if module.GetStatus() == StatusDisabled {
		return ErrModuleDisabled
	}

	// 检查依赖
	if err := mm.CheckModuleDependencies(moduleID); err != nil {
		return fmt.Errorf("%w: %s", ErrModuleDependency, err.Error())
	}

	// 启动模块
	if err := module.Start(); err != nil {
		mm.logger.Error("Failed to start module", "module", moduleID, "error", err.Error())
		return err
	}

	return nil
}

// StopModule 停止指定模块
func (mm *ModuleManagerImpl) StopModule(moduleID string) error {
	module, err := mm.GetModule(moduleID)
	if err != nil {
		return err
	}

	// 检查是否有其他模块依赖此模块
	if err := mm.checkDependentModules(moduleID); err != nil {
		return err
	}

	// 停止模块
	if err := module.Stop(); err != nil {
		mm.logger.Error("Failed to stop module", "module", moduleID, "error", err.Error())
		return err
	}

	return nil
}

// RestartModule 重启指定模块
func (mm *ModuleManagerImpl) RestartModule(moduleID string) error {
	// 停止模块
	if err := mm.StopModule(moduleID); err != nil {
		return fmt.Errorf("failed to stop module for restart: %w", err)
	}

	// 等待一小段时间确保模块完全停止
	time.Sleep(100 * time.Millisecond)

	// 启动模块
	if err := mm.StartModule(moduleID); err != nil {
		return fmt.Errorf("failed to start module after restart: %w", err)
	}

	return nil
}

// GetModuleStatus 获取模块状态
func (mm *ModuleManagerImpl) GetModuleStatus(moduleID string) (string, error) {
	module, err := mm.GetModule(moduleID)
	if err != nil {
		return "", err
	}

	return module.GetStatus(), nil
}

// GetModuleInfo 获取模块详细信息
func (mm *ModuleManagerImpl) GetModuleInfo(moduleID string) (*ModuleInfo, error) {
	module, err := mm.GetModule(moduleID)
	if err != nil {
		return nil, err
	}

	return module.GetInfo(), nil
}

// UpdateModuleConfig 更新模块配置
func (mm *ModuleManagerImpl) UpdateModuleConfig(moduleID string, config map[string]interface{}) error {
	module, err := mm.GetModule(moduleID)
	if err != nil {
		return err
	}

	// 更新配置
	if err := module.UpdateConfig(config); err != nil {
		mm.logger.Error("Failed to update module config", "module", moduleID, "error", err.Error())
		return err
	}

	// 更新全局配置
	if mm.config["modules"] == nil {
		mm.config["modules"] = make(map[string]interface{})
	}

	if moduleConfigs, ok := mm.config["modules"].(map[string]interface{}); ok {
		if moduleConfigs[moduleID] == nil {
			moduleConfigs[moduleID] = make(map[string]interface{})
		}

		if mc, ok := moduleConfigs[moduleID].(map[string]interface{}); ok {
			for k, v := range config {
				mc[k] = v
			}
		}
	}

	return nil
}

// CheckModuleDependencies 检查模块依赖
func (mm *ModuleManagerImpl) CheckModuleDependencies(moduleID string) error {
	module, err := mm.GetModule(moduleID)
	if err != nil {
		return err
	}

	// 获取依赖列表
	dependencies := module.GetDependencyIDs()
	for _, depID := range dependencies {
		depModule, err := mm.GetModule(depID)
		if err != nil {
			return fmt.Errorf("dependency module %s not found: %w", depID, err)
		}

		// 检查依赖模块状态
		depStatus := depModule.GetStatus()
		if depStatus != StatusRunning {
			return fmt.Errorf("dependency module %s is not running (status: %s)", depID, depStatus)
		}
	}

	return nil
}

// GetRunningModules 获取运行中的模块
func (mm *ModuleManagerImpl) GetRunningModules() []Module {
	var runningModules []Module

	for _, module := range mm.ListModules() {
		if module.IsRunning() {
			runningModules = append(runningModules, module)
		}
	}

	return runningModules
}

// GetStoppedModules 获取已停止的模块
func (mm *ModuleManagerImpl) GetStoppedModules() []Module {
	var stoppedModules []Module

	for _, module := range mm.ListModules() {
		if module.GetStatus() == StatusStopped {
			stoppedModules = append(stoppedModules, module)
		}
	}

	return stoppedModules
}

// GetErrorModules 获取出错的模块
func (mm *ModuleManagerImpl) GetErrorModules() []Module {
	var errorModules []Module

	for _, module := range mm.ListModules() {
		if module.GetStatus() == StatusError {
			errorModules = append(errorModules, module)
		}
	}

	return errorModules
}

// EnableModule 启用模块
func (mm *ModuleManagerImpl) EnableModule(moduleID string) error {
	module, err := mm.GetModule(moduleID)
	if err != nil {
		return err
	}

	// 更新状态
	if module.GetStatus() == StatusDisabled {
		module.(*ModuleBase).SetStatus(StatusStopped)
		mm.logger.Info("Module enabled", "module", moduleID)
	}

	return nil
}

// DisableModule 禁用模块
func (mm *ModuleManagerImpl) DisableModule(moduleID string) error {
	module, err := mm.GetModule(moduleID)
	if err != nil {
		return err
	}

	// 检查是否有其他模块依赖此模块
	if err := mm.checkDependentModules(moduleID); err != nil {
		return err
	}

	// 停止模块
	if module.IsRunning() {
		if err := module.Stop(); err != nil {
			mm.logger.Warn("Failed to stop module during disable", "module", moduleID, "error", err.Error())
		}
	}

	// 设置为禁用状态
	module.(*ModuleBase).SetStatus(StatusDisabled)
	mm.logger.Info("Module disabled", "module", moduleID)

	return nil
}

// IsModuleEnabled 检查模块是否启用
func (mm *ModuleManagerImpl) IsModuleEnabled(moduleID string) (bool, error) {
	module, err := mm.GetModule(moduleID)
	if err != nil {
		return false, err
	}

	return module.GetStatus() != StatusDisabled, nil
}

// GetModuleStats 获取模块统计信息
func (mm *ModuleManagerImpl) GetModuleStats() map[string]interface{} {
	modules := mm.ListModules()
	stats := make(map[string]interface{})

	runningCount := 0
	stoppedCount := 0
	errorCount := 0
	disabledCount := 0

	moduleDetails := make([]map[string]interface{}, 0, len(modules))

	for _, module := range modules {
		info := module.GetInfo()
		moduleStats := map[string]interface{}{
			"id":           info.ID,
			"name":         info.Name,
			"version":      info.Version,
			"status":       info.Status,
			"priority":     info.Priority,
			"dependencies": info.Dependencies,
		}

		moduleDetails = append(moduleDetails, moduleStats)

		// 统计状态
		switch info.Status {
		case StatusRunning:
			runningCount++
		case StatusStopped:
			stoppedCount++
		case StatusError:
			errorCount++
		case StatusDisabled:
			disabledCount++
		}
	}

	stats["total"] = len(modules)
	stats["running"] = runningCount
	stats["stopped"] = stoppedCount
	stats["error"] = errorCount
	stats["disabled"] = disabledCount
	stats["modules"] = moduleDetails

	return stats
}

// calculateStartupOrder 计算模块启动顺序（基于依赖关系）
func (mm *ModuleManagerImpl) calculateStartupOrder() ([]Module, error) {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	// 构建依赖图
	dependencyGraph := make(map[string][]string)
	visited := make(map[string]bool)
	result := make([]Module, 0)

	// 构建依赖图
	for _, module := range mm.modules {
		info := module.GetInfo()
		if info.Status != StatusDisabled {
			dependencyGraph[info.ID] = info.Dependencies
		}
	}

	// 深度优先搜索计算启动顺序
	for moduleID, _ := range dependencyGraph {
		if !visited[moduleID] {
			if err := mm.dfs(moduleID, dependencyGraph, visited, &result); err != nil {
				return nil, err
			}
		}
	}

	// 按优先级排序（优先级相同时保持DFS顺序）
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].GetInfo().Priority < result[j].GetInfo().Priority
	})

	return result, nil
}

// dfs 深度优先搜索遍历依赖图
func (mm *ModuleManagerImpl) dfs(moduleID string, graph map[string][]string, visited map[string]bool, result *[]Module) error {
	visited[moduleID] = true

	// 访问所有依赖
	for _, depID := range graph[moduleID] {
		// 检查依赖是否存在
		if _, exists := mm.modules[depID]; !exists {
			return fmt.Errorf("dependency module %s not found for module %s", depID, moduleID)
		}

		// 检查依赖是否禁用
		if mm.modules[depID].GetStatus() == StatusDisabled {
			return fmt.Errorf("dependency module %s is disabled", depID)
		}

		// 如果依赖未访问，递归访问
		if !visited[depID] {
			if err := mm.dfs(depID, graph, visited, result); err != nil {
				return err
			}
		}
	}

	// 将当前模块添加到结果中
	*result = append(*result, mm.modules[moduleID])
	return nil
}

// checkDependentModules 检查是否有其他模块依赖指定模块
func (mm *ModuleManagerImpl) checkDependentModules(moduleID string) error {
	for _, module := range mm.ListModules() {
		dependencies := module.GetDependencyIDs()
		for _, depID := range dependencies {
			if depID == moduleID && module.IsRunning() {
				return fmt.Errorf("cannot stop module %s, it is required by module %s", moduleID, module.GetInfo().ID)
			}
		}
	}

	return nil
}
