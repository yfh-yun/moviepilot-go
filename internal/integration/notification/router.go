package notification

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// Router 通知路由器
// 支持向多个通知渠道广播消息
type Router struct {
	factory *Factory
	logger  *zap.Logger
	mu      sync.RWMutex
}

// NewRouter 创建通知路由器
func NewRouter(factory *Factory) *Router {
	return &Router{
		factory: factory,
		logger:  logger.GetLogger(),
	}
}

// BroadcastText 向所有渠道广播文本消息
func (r *Router) BroadcastText(ctx context.Context, message string) error {
	r.mu.RLock()
	clients := r.factory.List()
	r.mu.RUnlock()

	if len(clients) == 0 {
		r.logger.Warn("没有可用的通知渠道")
		return fmt.Errorf("no notification channels available")
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(clients))

	for _, name := range clients {
		wg.Add(1)
		go func(channelName string) {
			defer wg.Done()

			client, ok := r.factory.Get(channelName)
			if !ok {
				r.logger.Warn("通知渠道不存在", zap.String("channel", channelName))
				return
			}

			if err := client.SendText(ctx, message); err != nil {
				r.logger.Error("发送通知失败",
					zap.String("channel", channelName),
					zap.Error(err))
				errChan <- fmt.Errorf("channel %s: %w", channelName, err)
			} else {
				r.logger.Info("通知发送成功", zap.String("channel", channelName))
			}
		}(name)
	}

	wg.Wait()
	close(errChan)

	// 收集所有错误
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("部分通知发送失败: %v", errs)
	}

	return nil
}

// BroadcastImage 向所有渠道广播图片消息
func (r *Router) BroadcastImage(ctx context.Context, imageURL string, caption string) error {
	r.mu.RLock()
	clients := r.factory.List()
	r.mu.RUnlock()

	if len(clients) == 0 {
		return fmt.Errorf("no notification channels available")
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(clients))

	for _, name := range clients {
		wg.Add(1)
		go func(channelName string) {
			defer wg.Done()

			client, ok := r.factory.Get(channelName)
			if !ok {
				return
			}

			if err := client.SendImage(ctx, imageURL, caption); err != nil {
				r.logger.Error("发送图片通知失败",
					zap.String("channel", channelName),
					zap.Error(err))
				errChan <- fmt.Errorf("channel %s: %w", channelName, err)
			}
		}(name)
	}

	wg.Wait()
	close(errChan)

	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("部分图片通知发送失败: %v", errs)
	}

	return nil
}

// Broadcast 向所有渠道广播通用消息
func (r *Router) Broadcast(ctx context.Context, msg *Message) error {
	r.mu.RLock()
	clients := r.factory.List()
	r.mu.RUnlock()

	if len(clients) == 0 {
		return fmt.Errorf("no notification channels available")
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(clients))

	for _, name := range clients {
		wg.Add(1)
		go func(channelName string) {
			defer wg.Done()

			client, ok := r.factory.Get(channelName)
			if !ok {
				return
			}

			if err := client.Send(ctx, msg); err != nil {
				r.logger.Error("发送通知失败",
					zap.String("channel", channelName),
					zap.String("type", string(msg.Type)),
					zap.Error(err))
				errChan <- fmt.Errorf("channel %s: %w", channelName, err)
			}
		}(name)
	}

	wg.Wait()
	close(errChan)

	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("部分通知发送失败: %v", errs)
	}

	return nil
}

// SendToChannel 向指定渠道发送消息
func (r *Router) SendToChannel(ctx context.Context, channelName string, msg *Message) error {
	client, ok := r.factory.Get(channelName)
	if !ok {
		return fmt.Errorf("notification channel not found: %s", channelName)
	}

	if err := client.Send(ctx, msg); err != nil {
		r.logger.Error("发送通知失败",
			zap.String("channel", channelName),
			zap.String("type", string(msg.Type)),
			zap.Error(err))
		return err
	}

	r.logger.Info("通知发送成功",
		zap.String("channel", channelName),
		zap.String("type", string(msg.Type)))

	return nil
}

// TestAllChannels 测试所有通知渠道的连接
func (r *Router) TestAllChannels(ctx context.Context) map[string]error {
	r.mu.RLock()
	clients := r.factory.List()
	r.mu.RUnlock()

	results := make(map[string]error)
	var mu sync.Mutex

	var wg sync.WaitGroup
	for _, name := range clients {
		wg.Add(1)
		go func(channelName string) {
			defer wg.Done()

			client, ok := r.factory.Get(channelName)
			if !ok {
				mu.Lock()
				results[channelName] = fmt.Errorf("channel not found")
				mu.Unlock()
				return
			}

			err := client.TestConnection(ctx)
			mu.Lock()
			results[channelName] = err
			mu.Unlock()

			if err != nil {
				r.logger.Error("通知渠道连接测试失败",
					zap.String("channel", channelName),
					zap.Error(err))
			} else {
				r.logger.Info("通知渠道连接测试成功",
					zap.String("channel", channelName))
			}
		}(name)
	}

	wg.Wait()
	return results
}
