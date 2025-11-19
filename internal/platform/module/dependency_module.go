package module

import (
	"fmt"
	"sync"
	"time"

	"moviepilot-go/pkg/logger"
)

// DependencyModule 依赖服务模块
type DependencyModule struct {
	*ModuleBase
	initialized bool
	running     bool
	services    map[string]interface{}
	serviceMutex sync.RWMutex
	config      map[string]interface{}
	shutdownCh  chan struct{}
	workWg      sync.WaitGroup
}

// NewDependencyModule 创建依赖服务模块
func NewDependencyModule() *DependencyModule {
	// 创建基础模块
	base := NewModuleBase(
		"dependency_module",
		"依赖服务模块",
		"提供依赖服务管理，作为其他模块的基础服务",
		"1.0.0",
		"MoviePilot Team",
		50,  // 优先级较高，因为其他模块可能依赖它
		[]string{}, // 不依赖其他模块
	)

	return &DependencyModule{
		ModuleBase:   base,
		initialized:  false,
		running:      false,
		services:     make(map[string]interface{}),
		shutdownCh:   make(chan struct{}),
	}
}

// Initialize 初始化依赖服务模块
func (dm *DependencyModule) Initialize(config map[string]interface{}, logger logger.Logger) error {
	// 调用基础类初始化
	if err := dm.ModuleBase.Initialize(config, logger); err != nil {
		return err
	}

	dm.logger.Info("Initializing dependency module")
	dm.config = dm.GetConfig()

	// 初始化默认配置
	if dm.config["cache_size"] == nil {
		dm.config["cache_size"] = 1000
	}

	if dm.config["cleanup_interval"] == nil {
		dm.config["cleanup_interval"] = 3600 // 默认1小时清理一次
	}

	if dm.config["max_service_wait"] == nil {
		dm.config["max_service_wait"] = 30 // 默认等待30秒
	}

	// 初始化服务缓存
	dm.services = make(map[string]interface{})
	dm.initialized = true

	dm.logger.Info("Dependency module initialized successfully")
	return nil
}

// Start 启动依赖服务模块
func (dm *DependencyModule) Start() error {
	// 调用基础类的Start方法
	if err := dm.ModuleBase.Start(); err != nil {
		return err
	}

	// 检查是否已初始化
	if !dm.initialized {
		dm.SetError(fmt.Errorf("module not initialized"))
		return fmt.Errorf("module not initialized")
	}

	dm.logger.Info("Starting dependency module")
	dm.running = true

	// 启动清理任务
	dm.workWg.Add(1)
	go dm.cleanupTask()

	// 启动健康检查
	dm.workWg.Add(1)
	go dm.healthCheckTask()

	dm.SetStatus(StatusRunning)
	dm.logger.Info("Dependency module started successfully")
	return nil
}

// Stop 停止依赖服务模块
func (dm *DependencyModule) Stop() error {
	// 调用基础类的Stop方法
	if err := dm.ModuleBase.Stop(); err != nil {
		return err
	}

	dm.logger.Info("Stopping dependency module")
	dm.running = false

	// 发送关闭信号
	close(dm.shutdownCh)

	// 等待所有工作协程结束
	dm.workWg.Wait()

	// 清理所有服务
	dm.clearAllServices()

	dm.SetStatus(StatusStopped)
	dm.logger.Info("Dependency module stopped successfully")
	return nil
}

// RegisterService 注册服务
func (dm *DependencyModule) RegisterService(name string, service interface{}) error {
	if !dm.IsRunning() {
		return ErrModuleNotRunning
	}

	dm.serviceMutex.Lock()
	defer dm.serviceMutex.Unlock()

	if _, exists := dm.services[name]; exists {
		return fmt.Errorf("service %s already exists", name)
	}

	dm.services[name] = service
	dm.logger.Debug("Service registered", "service", name, "type", fmt.Sprintf("%T", service))
	return nil
}

// UnregisterService 注销服务
func (dm *DependencyModule) UnregisterService(name string) error {
	if !dm.IsRunning() {
		return ErrModuleNotRunning
	}

	dm.serviceMutex.Lock()
	defer dm.serviceMutex.Unlock()

	if _, exists := dm.services[name]; !exists {
		return fmt.Errorf("service %s not found", name)
	}

	delete(dm.services, name)
	dm.logger.Debug("Service unregistered", "service", name)
	return nil
}

// GetService 获取服务
func (dm *DependencyModule) GetService(name string) (interface{}, error) {
	if !dm.IsRunning() {
		return nil, ErrModuleNotRunning
	}

	dm.serviceMutex.RLock()
	defer dm.serviceMutex.RUnlock()

	service, exists := dm.services[name]
	if !exists {
		return nil, fmt.Errorf("service %s not found", name)
	}

	return service, nil
}

// ListServices 列出所有服务
func (dm *DependencyModule) ListServices() []string {
	dm.serviceMutex.RLock()
	defer dm.serviceMutex.RUnlock()

	serviceNames := make([]string, 0, len(dm.services))
	for name := range dm.services {
		serviceNames = append(serviceNames, name)
	}

	return serviceNames
}

// WaitForService 等待服务可用
func (dm *DependencyModule) WaitForService(name string) (interface{}, error) {
	if !dm.IsRunning() {
		return nil, ErrModuleNotRunning
	}

	// 获取最大等待时间
	maxWait, ok := dm.config["max_service_wait"].(int)
	if !ok {
		maxWait = 30
	}

	timeout := time.After(time.Duration(maxWait) * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 尝试获取服务
			service, err := dm.GetService(name)
			if err == nil {
				return service, nil
			}
		case <-timeout:
			return nil, fmt.Errorf("timeout waiting for service %s", name)
		case <-dm.shutdownCh:
			return nil, fmt.Errorf("module shutting down")
		}
	}
}

// cleanupTask 清理过期服务的任务
func (dm *DependencyModule) cleanupTask() {
	defer dm.workWg.Done()

	interval, ok := dm.config["cleanup_interval"].(int)
	if !ok {
		interval = 3600 // 默认1小时
	}

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	dm.logger.Info("Service cleanup task started", "interval", interval)

	for {
		select {
		case <-ticker.C:
			// 执行清理（这里只是示例，实际根据需要实现）
			dm.logger.Debug("Performing service cleanup")
		case <-dm.shutdownCh:
			dm.logger.Info("Service cleanup task stopped")
			return
		}
	}
}

// healthCheckTask 健康检查任务
func (dm *DependencyModule) healthCheckTask() {
	defer dm.workWg.Done()

	ticker := time.NewTicker(30 * time.Second) // 每30秒检查一次
	defer ticker.Stop()

	dm.logger.Info("Health check task started")

	for {
		select {
		case <-ticker.C:
			// 执行健康检查
			dm.performHealthCheck()
		case <-dm.shutdownCh:
			dm.logger.Info("Health check task stopped")
			return
		}
	}
}

// performHealthCheck 执行健康检查
func (dm *DependencyModule) performHealthCheck() {
	dm.serviceMutex.RLock()
	serviceCount := len(dm.services)
	dm.serviceMutex.RUnlock()

	dm.logger.Debug("Health check performed", "service_count", serviceCount)
}

// clearAllServices 清理所有服务
func (dm *DependencyModule) clearAllServices() {
	dm.serviceMutex.Lock()
	defer dm.serviceMutex.Unlock()

	for name := range dm.services {
		dm.logger.Debug("Cleaning up service during shutdown", "service", name)
	}

	dm.services = make(map[string]interface{})
}

// GetServiceStats 获取服务统计信息
func (dm *DependencyModule) GetServiceStats() map[string]interface{} {
	dm.serviceMutex.RLock()
	defer dm.serviceMutex.RUnlock()

	serviceDetails := make(map[string]string)
	for name, service := range dm.services {
		serviceDetails[name] = fmt.Sprintf("%T", service)
	}

	return map[string]interface{}{
		"initialized":    dm.initialized,
		"running":        dm.running,
		"service_count":  len(dm.services),
		"services":       serviceDetails,
		"config":         dm.config,
	}
}

// 扩展Module接口的服务方法

// ServiceProvider 服务提供者接口
type ServiceProvider interface {
	// RegisterService 注册服务
	RegisterService(name string, service interface{}) error

	// UnregisterService 注销服务
	UnregisterService(name string) error

	// GetService 获取服务
	GetService(name string) (interface{}, error)

	// ListServices 列出所有服务
	ListServices() []string

	// WaitForService 等待服务可用
	WaitForService(name string) (interface{}, error)
}

// ServiceDependent 依赖服务的模块接口
type ServiceDependent interface {
	// GetRequiredServices 获取所需的服务列表
	GetRequiredServices() []string

	// InjectServices 注入服务
	InjectServices(provider ServiceProvider) error
}

// 实现服务依赖注入的模块示例

// DependentModule 依赖服务的模块示例
type DependentModule struct {
	*ModuleBase
	dependencyModule *DependencyModule
	services         map[string]interface{}
	initialized      bool
	running          bool
	config           map[string]interface{}
}

// NewDependentModule 创建依赖服务的模块
func NewDependentModule() *DependentModule {
	// 创建基础模块，依赖dependency_module
	base := NewModuleBase(
		"dependent_module",
		"依赖服务的模块",
		"演示如何依赖和使用其他模块提供的服务",
		"1.0.0",
		"MoviePilot Team",
		150, // 优先级低于依赖模块
		[]string{"dependency_module"}, // 依赖dependency_module
	)

	return &DependentModule{
		ModuleBase: base,
		services:   make(map[string]interface{}),
	}
}

// Initialize 初始化依赖服务的模块
func (dm *DependentModule) Initialize(config map[string]interface{}, logger logger.Logger) error {
	// 调用基础类初始化
	if err := dm.ModuleBase.Initialize(config, logger); err != nil {
		return err
	}

	dm.logger.Info("Initializing dependent module")
	dm.config = dm.GetConfig()
	dm.initialized = true

	// 初始化默认配置
	if dm.config["retry_count"] == nil {
		dm.config["retry_count"] = 3
	}

	dm.logger.Info("Dependent module initialized successfully")
	return nil
}

// Start 启动依赖服务的模块
func (dm *DependentModule) Start() error {
	// 调用基础类的Start方法
	if err := dm.ModuleBase.Start(); err != nil {
		return err
	}

	// 检查是否已初始化
	if !dm.initialized {
		dm.SetError(fmt.Errorf("module not initialized"))
		return fmt.Errorf("module not initialized")
	}

	dm.logger.Info("Starting dependent module")

	// 获取依赖的模块
	moduleManager := dm.getModuleManager()
	if moduleManager == nil {
		dm.SetError(fmt.Errorf("cannot get module manager"))
		return fmt.Errorf("cannot get module manager")
	}

	// 获取依赖服务模块
	depModule, err := moduleManager.GetModule("dependency_module")
	if err != nil {
		dm.SetError(fmt.Errorf("cannot get dependency module: %w", err))
		return fmt.Errorf("cannot get dependency module: %w", err)
	}

	// 转换为依赖服务模块类型
	dm.dependencyModule = depModule.(*DependencyModule)

	// 注入服务
	if err := dm.InjectServices(dm.dependencyModule); err != nil {
		dm.SetError(fmt.Errorf("failed to inject services: %w", err))
		return fmt.Errorf("failed to inject services: %w", err)
	}

	dm.running = true
	dm.SetStatus(StatusRunning)
	dm.logger.Info("Dependent module started successfully")
	return nil
}

// Stop 停止依赖服务的模块
func (dm *DependentModule) Stop() error {
	// 调用基础类的Stop方法
	if err := dm.ModuleBase.Stop(); err != nil {
		return err
	}

	dm.logger.Info("Stopping dependent module")

	// 清理服务引用
	dm.services = make(map[string]interface{})
	dm.dependencyModule = nil

	dm.running = false
	dm.SetStatus(StatusStopped)
	dm.logger.Info("Dependent module stopped successfully")
	return nil
}

// GetRequiredServices 获取所需的服务列表
func (dm *DependentModule) GetRequiredServices() []string {
	// 定义此模块需要的服务
	return []string{"example_service", "data_service"}
}

// InjectServices 注入服务
func (dm *DependentModule) InjectServices(provider ServiceProvider) error {
	requiredServices := dm.GetRequiredServices()
	retryCount, ok := dm.config["retry_count"].(int)
	if !ok {
		retryCount = 3
	}

	for _, serviceName := range requiredServices {
		dm.logger.Debug("Waiting for service", "service", serviceName)

		// 尝试获取服务，支持重试
		var service interface{}
		var err error

		for i := 0; i < retryCount; i++ {
			service, err = provider.WaitForService(serviceName)
			if err == nil {
				break
			}

			dm.logger.Warn("Failed to get service, retrying", 
				"service", serviceName, 
				"retry", i+1, 
				"max_retries", retryCount,
				"error", err.Error())

			time.Sleep(1 * time.Second)
		}

		if err != nil {
			return fmt.Errorf("failed to get required service %s: %w", serviceName, err)
		}

		dm.services[serviceName] = service
		dm.logger.Debug("Service injected", "service", serviceName, "type", fmt.Sprintf("%T", service))
	}

	return nil
}

// GetService 获取注入的服务
func (dm *DependentModule) GetService(name string) (interface{}, error) {
	service, exists := dm.services[name]
	if !exists {
		return nil, fmt.Errorf("service %s not injected", name)
	}

	return service, nil
}

// getModuleManager 获取模块管理器（示例方法，实际实现可能需要从外部传入）
func (dm *DependentModule) getModuleManager() ModuleManager {
	// 这里应该返回实际的模块管理器实例
	// 为了演示目的，这里简单返回nil
	// 在实际应用中，模块管理器应该通过某种方式注入或获取
	return nil
}

// LifecycleHook 生命周期钩子接口
type LifecycleHook interface {
	// OnBeforeInit 初始化前钩子
	OnBeforeInit(config map[string]interface{}) error

	// OnAfterInit 初始化后钩子
	OnAfterInit() error

	// OnBeforeStart 启动前钩子
	OnBeforeStart() error

	// OnAfterStart 启动后钩子
	OnAfterStart() error

	// OnBeforeStop 停止前钩子
	OnBeforeStop() error

	// OnAfterStop 停止后钩子
	OnAfterStop() error

	// OnConfigUpdate 配置更新钩子
	OnConfigUpdate(config map[string]interface{}) error
}

// HookModule 带生命周期钩子的模块基类
type HookModule struct {
	*ModuleBase
	hooks []LifecycleHook
}

// NewHookModule 创建带生命周期钩子的模块
func NewHookModule(id, name, description, version, author string, priority int, dependencies []string) *HookModule {
	base := NewModuleBase(id, name, description, version, author, priority, dependencies)
	return &HookModule{
		ModuleBase: base,
		hooks:      make([]LifecycleHook, 0),
	}
}

// RegisterHook 注册生命周期钩子
func (hm *HookModule) RegisterHook(hook LifecycleHook) {
	hm.hooks = append(hm.hooks, hook)
}

// Initialize 重写初始化方法，添加钩子调用
func (hm *HookModule) Initialize(config map[string]interface{}, logger logger.Logger) error {
	// 调用初始化前钩子
	for _, hook := range hm.hooks {
		if err := hook.OnBeforeInit(config); err != nil {
			return fmt.Errorf("error in OnBeforeInit hook: %w", err)
		}
	}

	// 调用基础类初始化
	if err := hm.ModuleBase.Initialize(config, logger); err != nil {
		return err
	}

	// 调用初始化后钩子
	for _, hook := range hm.hooks {
		if err := hook.OnAfterInit(); err != nil {
			return fmt.Errorf("error in OnAfterInit hook: %w", err)
		}
	}

	return nil
}

// Start 重写启动方法，添加钩子调用
func (hm *HookModule) Start() error {
	// 调用启动前钩子
	for _, hook := range hm.hooks {
		if err := hook.OnBeforeStart(); err != nil {
			return fmt.Errorf("error in OnBeforeStart hook: %w", err)
		}
	}

	// 调用基础类启动
	if err := hm.ModuleBase.Start(); err != nil {
		return err
	}

	// 调用启动后钩子
	for _, hook := range hm.hooks {
		if err := hook.OnAfterStart(); err != nil {
			return fmt.Errorf("error in OnAfterStart hook: %w", err)
		}
	}

	return nil
}

// Stop 重写停止方法，添加钩子调用
func (hm *HookModule) Stop() error {
	// 调用停止前钩子
	for _, hook := range hm.hooks {
		if err := hook.OnBeforeStop(); err != nil {
			return fmt.Errorf("error in OnBeforeStop hook: %w", err)
		}
	}

	// 调用基础类停止
	if err := hm.ModuleBase.Stop(); err != nil {
		return err
	}

	// 调用停止后钩子
	for _, hook := range hm.hooks {
		if err := hook.OnAfterStop(); err != nil {
			return fmt.Errorf("error in OnAfterStop hook: %w", err)
		}
	}

	return nil
}

// UpdateConfig 重写配置更新方法，添加钩子调用
func (hm *HookModule) UpdateConfig(config map[string]interface{}) error {
	// 调用基础类更新配置
	if err := hm.ModuleBase.UpdateConfig(config); err != nil {
		return err
	}

	// 调用配置更新钩子
	for _, hook := range hm.hooks {
		if err := hook.OnConfigUpdate(config); err != nil {
			return fmt.Errorf("error in OnConfigUpdate hook: %w", err)
		}
	}

	return nil
}