package command

import (
	"context"

	"go.uber.org/zap"

	domains_events "moviepilot-go/internal/business/domains/events"
	infrastructure_events "moviepilot-go/internal/infrastructure/events"
	"moviepilot-go/pkg/logger"
)

// Factory 命令工厂，用于创建和配置命令服务
type Factory struct {
	logger   *zap.Logger
	eventBus infrastructure_events.Bus
}

// NewFactory 创建命令工厂
func NewFactory(eventBus infrastructure_events.Bus) *Factory {
	return &Factory{
		logger:   logger.GetLogger(),
		eventBus: eventBus,
	}
}

// Create 创建并配置命令服务
func (f *Factory) Create() (Service, error) {
	// 创建命令服务
	commands := NewService()

	// 注册命令处理器
	if err := f.registerHandlers(commands); err != nil {
		return nil, err
	}

	// 订阅消息事件
	if err := f.subscribeToEvents(commands); err != nil {
		return nil, err
	}

	return commands, nil
}

// registerHandlers 注册命令处理器
func (f *Factory) registerHandlers(commands Service) error {
	// 注册帮助命令
	helpHandler := NewHelpHandler(commands)
	if err := commands.RegisterHandler(helpHandler); err != nil {
		return err
	}

	// 注册状态命令
	statusHandler := NewStatusHandler()
	if err := commands.RegisterHandler(statusHandler); err != nil {
		return err
	}

	// 注册订阅命令
	subscribeHandler := NewSubscribeHandler()
	if err := commands.RegisterHandler(subscribeHandler); err != nil {
		return err
	}

	// 注册取消订阅命令
	unsubscribeHandler := NewUnsubscribeHandler()
	if err := commands.RegisterHandler(unsubscribeHandler); err != nil {
		return err
	}

	// 注册清理缓存命令
	clearCacheHandler := NewClearCacheHandler()
	if err := commands.RegisterHandler(clearCacheHandler); err != nil {
		return err
	}

	// 注册版本命令
	versionHandler := NewVersionHandler()
	if err := commands.RegisterHandler(versionHandler); err != nil {
		return err
	}

	// 注册重启命令
	restartHandler := NewRestartHandler()
	if err := commands.RegisterHandler(restartHandler); err != nil {
		return err
	}

	// 注册正在下载命令
	downloadingHandler := NewDownloadingHandler()
	if err := commands.RegisterHandler(downloadingHandler); err != nil {
		return err
	}

	// 注册手动整理命令
	redoHandler := NewRedoHandler()
	if err := commands.RegisterHandler(redoHandler); err != nil {
		return err
	}

	// 注册站点相关命令
	sitesHandler := NewSitesHandler()
	if err := commands.RegisterHandler(sitesHandler); err != nil {
		return err
	}

	siteCookieHandler := NewSiteCookieHandler()
	if err := commands.RegisterHandler(siteCookieHandler); err != nil {
		return err
	}

	siteStatisticHandler := NewSiteStatisticHandler()
	if err := commands.RegisterHandler(siteStatisticHandler); err != nil {
		return err
	}

	siteEnableHandler := NewSiteEnableHandler()
	if err := commands.RegisterHandler(siteEnableHandler); err != nil {
		return err
	}

	siteDisableHandler := NewSiteDisableHandler()
	if err := commands.RegisterHandler(siteDisableHandler); err != nil {
		return err
	}

	siteRefreshHandler := NewSiteRefreshHandler()
	if err := commands.RegisterHandler(siteRefreshHandler); err != nil {
		return err
	}

	return nil
}

// subscribeToEvents 订阅消息事件
func (f *Factory) subscribeToEvents(commands Service) error {
	// 订阅用户消息事件
	userMessageType := domains_events.EventType("user.message")
	f.eventBus.SubscribeBroadcast(userMessageType, func(ctx context.Context, event *domains_events.Event) error {
		// 从事件数据中获取消息内容
		if message, ok := event.Data.(string); ok {
			// 执行命令
			if err := commands.Execute(ctx, message); err != nil {
				f.logger.Error("Failed to execute command",
					zap.String("message", message),
					zap.Error(err))
			}
		}
		return nil
	})

	return nil
}
