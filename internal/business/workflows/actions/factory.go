package actions

import (
	"fmt"
	"sync"

	"go.uber.org/zap"

	"moviepilot-go/internal/business/workflows/actions/common"
	coreActions "moviepilot-go/internal/business/workflows/actions/core"
	fileActions "moviepilot-go/internal/business/workflows/actions/file"
	systemActions "moviepilot-go/internal/business/workflows/actions/system"
)

// ActionFactory Action工厂
type ActionFactory struct {
	logger   *zap.Logger
	registry map[string]ActionCreator
	mu       sync.RWMutex
	services *Services
}

// ActionCreator Action创建函数
type ActionCreator func(logger *zap.Logger, services *Services) Action

// NewActionFactory 创建Action工厂
func NewActionFactory(logger *zap.Logger, services *Services) *ActionFactory {
	factory := &ActionFactory{
		logger:   logger,
		registry: make(map[string]ActionCreator),
		services: services,
	}

	// 注册所有内置Actions
	factory.registerBuiltinActions()

	return factory
}

// registerBuiltinActions 注册所有内置Actions
func (f *ActionFactory) registerBuiltinActions() {
	// 文件处理类
	f.Register("scan_file", func(logger *zap.Logger, services *Services) Action {
		return fileActions.NewScanAction()
	})

	f.Register("scrape_file", func(logger *zap.Logger, services *Services) Action {
		return fileActions.NewScrapeAction()
	})

	f.Register("transfer_file", func(logger *zap.Logger, services *Services) Action {
		return fileActions.NewTransferAction()
	})

	f.Register("delete_file", func(logger *zap.Logger, services *Services) Action {
		return common.NewBaseAction("delete_file", ActionTypeFile)
	})

	f.Register("copy_file", func(logger *zap.Logger, services *Services) Action {
		return common.NewBaseAction("copy_file", ActionTypeFile)
	})

	f.Register("move_file", func(logger *zap.Logger, services *Services) Action {
		return common.NewBaseAction("move_file", ActionTypeFile)
	})

	// 资源获取类
	f.Register("fetch_torrents", func(logger *zap.Logger, services *Services) Action {
		return common.NewBaseAction("fetch_torrents", ActionTypeResource)
	})

	f.Register("fetch_rss", func(logger *zap.Logger, services *Services) Action {
		return common.NewBaseAction("fetch_rss", ActionTypeResource)
	})

	f.Register("fetch_medias", func(logger *zap.Logger, services *Services) Action {
		return common.NewBaseAction("fetch_medias", ActionTypeResource)
	})

	f.Register("fetch_downloads", func(logger *zap.Logger, services *Services) Action {
		return common.NewBaseAction("fetch_downloads", ActionTypeResource)
	})

	// 过滤类
	f.Register("filter_torrents", func(logger *zap.Logger, services *Services) Action {
		return coreActions.NewFilterTorrentsAction()
	})

	f.Register("filter_medias", func(logger *zap.Logger, services *Services) Action {
		return coreActions.NewFilterMediasAction()
	})

	// 核心业务类
	f.Register("add_download", func(logger *zap.Logger, services *Services) Action {
		return coreActions.NewDownloadAction()
	})

	f.Register("add_subscribe", func(logger *zap.Logger, services *Services) Action {
		return coreActions.NewSubscribeAction()
	})

	f.Register("send_message", func(logger *zap.Logger, services *Services) Action {
		return coreActions.NewMessageAction()
	})

	// 系统功能类
	f.Register("invoke_plugin", func(logger *zap.Logger, services *Services) Action {
		return coreActions.NewInvokePluginAction()
	})

	f.Register("send_event", func(logger *zap.Logger, services *Services) Action {
		return systemActions.NewEventAction()
	})

	f.Register("note", func(logger *zap.Logger, services *Services) Action {
		return systemActions.NewNoteAction()
	})
}

// Register 注册Action创建函数
func (f *ActionFactory) Register(name string, creator ActionCreator) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.registry[name] = creator
	f.logger.Info("Action registered", zap.String("name", name))
}

// Create 创建Action实例
func (f *ActionFactory) Create(name string) (Action, error) {
	f.mu.RLock()
	creator, exists := f.registry[name]
	f.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("action not found: %s", name)
	}

	return creator(f.logger, f.services), nil
}

// List 列出所有已注册的Actions
func (f *ActionFactory) List() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	names := make([]string, 0, len(f.registry))
	for name := range f.registry {
		names = append(names, name)
	}

	return names
}

// Exists 检查Action是否存在
func (f *ActionFactory) Exists(name string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	_, exists := f.registry[name]
	return exists
}

// GetServices 获取服务容器
func (f *ActionFactory) GetServices() *Services {
	return f.services
}
