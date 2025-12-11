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
	if err := commands.Register(helpHandler); err != nil {
		return err
	}

	// 注册状态命令
	statusHandler := NewStatusHandler()
	if err := commands.Register(statusHandler); err != nil {
		return err
	}

	// 注册订阅命令
	subscribeHandler := NewSubscribeHandler()
	if err := commands.Register(subscribeHandler); err != nil {
		return err
	}

	// 注册取消订阅命令
	unsubscribeHandler := NewUnsubscribeHandler()
	if err := commands.Register(unsubscribeHandler); err != nil {
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
