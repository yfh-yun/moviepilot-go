// Package registry 提供动作注册和管理功能
package registry

import (
	"fmt"
	"sync"

	"moviepilot-go/internal/business/workflows/interfaces"
	"moviepilot-go/internal/business/workflows/implementations"
	"moviepilot-go/pkg/logger"
)

// ActionRegistry 动作注册表
type ActionRegistry struct {
	actions map[string]interfaces.ActionFactory
	mutex   sync.RWMutex
}

var (
	defaultRegistry *ActionRegistry
	once           sync.Once
)

// GetDefaultRegistry 获取默认注册表
func GetDefaultRegistry() *ActionRegistry {
	once.Do(func() {
		defaultRegistry = NewActionRegistry()
		// 注册内置动作
		defaultRegistry.registerBuiltinActions()
	})
	return defaultRegistry
}

// NewActionRegistry 创建新的动作注册表
func NewActionRegistry() *ActionRegistry {
	return &ActionRegistry{
		actions: make(map[string]interfaces.ActionFactory),
	}
}

// Register 注册动作工厂
func (r *ActionRegistry) Register(name string, factory interfaces.ActionFactory) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.actions[name]; exists {
		return fmt.Errorf("action '%s' already registered", name)
	}

	r.actions[name] = factory
	logger.Info("Action registered", "name", name)
	return nil
}

// Unregister 注销动作
func (r *ActionRegistry) Unregister(name string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.actions[name]; !exists {
		return fmt.Errorf("action '%s' not found", name)
	}

	delete(r.actions, name)
	logger.Info("Action unregistered", "name", name)
	return nil
}

// GetFactory 获取动作工厂
func (r *ActionRegistry) GetFactory(name string) (interfaces.ActionFactory, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	factory, exists := r.actions[name]
	if !exists {
		return nil, fmt.Errorf("action '%s' not found", name)
	}

	return factory, nil
}

// CreateAction 创建动作实例
func (r *ActionRegistry) CreateAction(name string) (interfaces.Action, error) {
	factory, err := r.GetFactory(name)
	if err != nil {
		return nil, err
	}

	return factory.Create(), nil
}

// ListActions 列出所有已注册的动作
func (r *ActionRegistry) ListActions() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var names []string
	for name := range r.actions {
		names = append(names, name)
	}

	return names
}

// GetActionInfo 获取动作信息
func (r *ActionRegistry) GetActionInfo(name string) (*ActionInfo, error) {
	factory, err := r.GetFactory(name)
	if err != nil {
		return nil, err
	}

	action := factory.Create()
	return &ActionInfo{
		Name:        action.Name(),
		Description: action.Description(),
		Version:     action.Version(),
		Author:      action.Author(),
		Category:    action.Category(),
		Tags:        action.Tags(),
	}, nil
}

// ActionInfo 动作信息
type ActionInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Author      string   `json:"author"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
}

// registerBuiltinActions 注册内置动作
func (r *ActionRegistry) registerBuiltinActions() {
	// 注册下载动作
	r.Register("download", implementations.NewDownloadAction)
	
	// 注册扫描动作
	r.Register("scan", implementations.NewScanAction)
	
	// 注册文件扫描器
	r.Register("file_scanner", implementations.NewFileScanner)
	
	// 注册媒体获取器
	r.Register("media_fetcher", implementations.NewMediaFetcher)
	
	// 注册消息发送器
	r.Register("message_sender", implementations.NewMessageSender)
	
	// 注册插件调用器
	r.Register("plugin_invoker", implementations.NewPluginInvoker)
	
	// 注册RSS解析器
	r.Register("rss_parser", implementations.NewRSSParser)
	
	// 注册订阅管理器
	r.Register("subscribe_manager", implementations.NewSubscribeManager)
	
	// 注册文件转移管理器
	r.Register("transfer_manager", implementations.NewTransferManager)
	
	// 注册工作流缓存
	r.Register("workflow_cache", implementations.NewWorkflowCache)
	
	logger.Info("Builtin actions registered", "count", len(r.actions))
}