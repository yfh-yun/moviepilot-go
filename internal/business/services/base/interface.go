package base

import "context"

// Service 服务接口
// 所有Service类都应实现此接口
type Service interface {
	// Initialize 初始化服务
	Initialize() error

	// Name 获取服务名称
	Name() string

	// Close 关闭服务，释放资源
	Close() error
}

// ModuleRunner 模块运行器接口
type ModuleRunner interface {
	// RunModule 运行模块
	RunModule(method string, args ...any) any

	// RunModuleAsync 异步运行模块
	RunModuleAsync(ctx context.Context, method string, args ...any) (any, error)
}

// CacheManager 缓存管理器接口
type CacheManager interface {
	// LoadCache 加载缓存
	LoadCache(filename string) (any, error)

	// SaveCache 保存缓存
	SaveCache(data any, filename string) error

	// RemoveCache 删除缓存
	RemoveCache(filename string) error

	// AsyncLoadCache 异步加载缓存
	AsyncLoadCache(ctx context.Context, filename string) (any, error)

	// AsyncSaveCache 异步保存缓存
	AsyncSaveCache(ctx context.Context, data any, filename string) error

	// AsyncRemoveCache 异步删除缓存
	AsyncRemoveCache(ctx context.Context, filename string) error
}

// MessageSender 消息发送器接口
type MessageSender interface {
	// PostMessage 发送消息
	PostMessage(msg any) error
}

// EventSender 事件发送器接口
type EventSender interface {
	// PostEvent 发送事件
	PostEvent(eventType string, data map[string]any) error
}
