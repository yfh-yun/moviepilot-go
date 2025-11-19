package module

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"moviepilot-go/pkg/logger"
)

// 模块状态常量
const (
	StatusStopped   = "stopped"
	StatusStarting  = "starting"
	StatusRunning   = "running"
	StatusStopping  = "stopping"
	StatusError     = "error"
	StatusDisabled  = "disabled"
)

// 预定义的错误
var (
	ErrModuleNotFound       = errors.New("module not found")
	ErrModuleAlreadyExists  = errors.New("module already exists")
	ErrModuleNotInitialized = errors.New("module not initialized")
	ErrModuleNotRunning     = errors.New("module not running")
	ErrModuleDependency     = errors.New("module dependency error")
	ErrModuleDisabled       = errors.New("module is disabled")
	ErrModuleFailed         = errors.New("module failed to start")
)

// ModuleInfo 模块元数据信息
type ModuleInfo struct {
	ID          string            // 模块唯一标识符
	Name        string            // 模块名称
	Description string            // 模块描述
	Version     string            // 模块版本
	Author      string            // 模块作者
	Status      string            // 模块状态
	Priority    int               // 启动优先级，数字越小优先级越高
	Dependencies []string         // 依赖的其他模块ID列表
	Config      map[string]interface{} // 模块配置
	StartTime   time.Time         // 启动时间
	StopTime    time.Time         // 停止时间
	Error       string            // 错误信息
}

// Module 模块接口定义
type Module interface {
	// GetInfo 获取模块信息
	GetInfo() *ModuleInfo

	// Initialize 初始化模块
	Initialize(config map[string]interface{}, logger logger.Logger) error

	// Start 启动模块
	Start() error

	// Stop 停止模块
	Stop() error

	// IsRunning 检查模块是否运行中
	IsRunning() bool

	// GetStatus 获取模块当前状态
	GetStatus() string

	// GetConfig 获取模块配置
	GetConfig() map[string]interface{}

	// UpdateConfig 更新模块配置
	UpdateConfig(config map[string]interface{}) error

	// GetDependencyIDs 获取模块依赖
	GetDependencyIDs() []string

	// GetLogger 获取模块日志器
	GetLogger() logger.Logger
}

// ModuleBase 模块基础实现，提供通用功能
type ModuleBase struct {
	info        *ModuleInfo
	logger      logger.Logger
	mutex       sync.RWMutex
	initialized bool
}

// NewModuleBase 创建模块基础实例
func NewModuleBase(id, name, description, version, author string, priority int, dependencies []string) *ModuleBase {
	return &ModuleBase{
		info: &ModuleInfo{
			ID:          id,
			Name:        name,
			Description: description,
			Version:     version,
			Author:      author,
			Status:      StatusStopped,
			Priority:    priority,
			Dependencies: dependencies,
			Config:      make(map[string]interface{}),
		},
		initialized: false,
	}
}

// GetInfo 获取模块信息
func (m *ModuleBase) GetInfo() *ModuleInfo {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	// 返回副本避免并发问题
	infoCopy := *m.info
	return &infoCopy
}

// Initialize 初始化模块
func (m *ModuleBase) Initialize(config map[string]interface{}, logger logger.Logger) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.initialized {
		return nil
	}

	m.logger = logger
	m.info.Config = make(map[string]interface{})

	// 合并配置
	if config != nil {
		for k, v := range config {
			m.info.Config[k] = v
		}
	}

	m.initialized = true
	m.logger.Debug("Module initialized", "module", m.info.ID)
	return nil
}

// GetStatus 获取模块状态
func (m *ModuleBase) GetStatus() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.info.Status
}

// SetStatus 设置模块状态
func (m *ModuleBase) SetStatus(status string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	oldStatus := m.info.Status
	m.info.Status = status
	
	// 记录状态变化
	if m.logger != nil {
		if status == StatusRunning {
			m.info.StartTime = time.Now()
		} else if status == StatusStopped {
			m.info.StopTime = time.Now()
		}
		m.logger.Debug("Module status changed", "module", m.info.ID, "from", oldStatus, "to", status)
	}
}

// IsRunning 检查模块是否运行中
func (m *ModuleBase) IsRunning() bool {
	return m.GetStatus() == StatusRunning
}

// GetConfig 获取模块配置
func (m *ModuleBase) GetConfig() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	// 返回副本避免并发问题
	configCopy := make(map[string]interface{})
	for k, v := range m.info.Config {
		configCopy[k] = v
	}
	return configCopy
}

// UpdateConfig 更新模块配置
func (m *ModuleBase) UpdateConfig(config map[string]interface{}) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.initialized {
		return ErrModuleNotInitialized
	}

	// 合并新配置
	if config != nil {
		for k, v := range config {
			m.info.Config[k] = v
		}
	}

	if m.logger != nil {
		m.logger.Debug("Module config updated", "module", m.info.ID)
	}
	return nil
}

// GetDependencyIDs 获取模块依赖
func (m *ModuleBase) GetDependencyIDs() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	// 返回副本避免并发问题
	deps := make([]string, len(m.info.Dependencies))
	copy(deps, m.info.Dependencies)
	return deps
}

// GetLogger 获取模块日志器
func (m *ModuleBase) GetLogger() logger.Logger {
	return m.logger
}

// SetError 设置模块错误信息
func (m *ModuleBase) SetError(err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if err != nil {
		m.info.Error = err.Error()
		m.info.Status = StatusError
		if m.logger != nil {
			m.logger.Error("Module error", "module", m.info.ID, "error", err.Error())
		}
	} else {
		m.info.Error = ""
	}
}

// ClearError 清除模块错误信息
func (m *ModuleBase) ClearError() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.info.Error = ""
}

// IsInitialized 检查模块是否已初始化
func (m *ModuleBase) IsInitialized() bool {
	return m.initialized
}

// ValidateDependencies 验证模块依赖（基础实现）
func (m *ModuleBase) ValidateDependencies(moduleManager ModuleManager) error {
	if moduleManager == nil {
		return fmt.Errorf("module manager is nil")
	}

	dependencies := m.GetDependencyIDs()
	for _, depID := range dependencies {
		depModule, err := moduleManager.GetModule(depID)
		if err != nil {
			return fmt.Errorf("dependency module %s not found: %w", depID, err)
		}

		if depModule.GetStatus() != StatusRunning {
			return fmt.Errorf("dependency module %s is not running", depID)
		}
	}

	return nil
}

// Start 启动模块的基本实现（需要被子类重写）
func (m *ModuleBase) Start() error {
	if !m.initialized {
		return ErrModuleNotInitialized
	}
	
	m.SetStatus(StatusStarting)
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("panic in module %s: %v", m.info.ID, r)
			m.SetError(err)
		}
	}()

	return nil
}

// Stop 停止模块的基本实现（需要被子类重写）
func (m *ModuleBase) Stop() error {
	if !m.initialized {
		return ErrModuleNotInitialized
	}
	
	m.SetStatus(StatusStopping)
	defer func() {
		if r := recover(); r != nil {
			m.logger.Error("Panic during module stop", "module", m.info.ID, "error", r)
		}
		m.SetStatus(StatusStopped)
	}()

	return nil
}