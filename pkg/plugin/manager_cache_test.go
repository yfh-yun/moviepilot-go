package plugin

import (
	"context"
	"testing"
	"time"
)

// mockConfigStore 模拟配置存储

type mockConfigStore struct{}

func (m *mockConfigStore) Get(ctx context.Context, pluginID string) (map[string]any, error) {
	return nil, nil
}

func (m *mockConfigStore) Set(ctx context.Context, pluginID string, config map[string]any) error {
	return nil
}

func (m *mockConfigStore) Delete(ctx context.Context, pluginID string) error {
	return nil
}

// mockDataStore 模拟数据存储

type mockDataStore struct{}

func (m *mockDataStore) Get(ctx context.Context, pluginID, key string) (any, error) {
	return nil, nil
}

func (m *mockDataStore) Set(ctx context.Context, pluginID, key string, value any) error {
	return nil
}

func (m *mockDataStore) Delete(ctx context.Context, pluginID, key string) error {
	return nil
}

func (m *mockDataStore) DeleteAll(ctx context.Context, pluginID string) error {
	return nil
}

// mockEventManager 模拟事件管理器

type mockEventManager struct{}

func (m *mockEventManager) PublishEvent(ctx context.Context, event *Event) error {
	return nil
}

func (m *mockEventManager) PublishEventAsync(event *Event) {
	// do nothing
}

func (m *mockEventManager) SubscribeEvent(eventType EventType, handler EventHandler, filter EventFilter) (string, error) {
	return "", nil
}

func (m *mockEventManager) UnsubscribeEvent(subscriptionID string) error {
	return nil
}

func (m *mockEventManager) SubscribeMultipleEvents(eventTypes []EventType, handler EventHandler, filter EventFilter) ([]string, error) {
	return []string{}, nil
}

func (m *mockEventManager) UnsubscribeAllEvents() error {
	return nil
}

func (m *mockEventManager) GetSubscriptions() []*EventSubscription {
	return nil
}

func (m *mockEventManager) GetSubscriptionsByEventType(eventType EventType) []*EventSubscription {
	return nil
}

func (m *mockEventManager) Close() error {
	return nil
}

// TestManagerCache 测试插件管理器缓存功能
func TestManagerCache(t *testing.T) {
	// 创建模拟依赖
	configStore := &mockConfigStore{}
	dataStore := &mockDataStore{}
	eventManager := &mockEventManager{}

	// 创建插件管理器
	manager := NewPluginManager(configStore, dataStore, eventManager)

	// 测试GetOnlinePlugins缓存
	t.Run("GetOnlinePluginsCache", func(t *testing.T) {
		ctx := context.Background()

		// 第一次调用，应该从插件市场获取（模拟返回空列表）
		plugins1, err := manager.GetOnlinePlugins(ctx, false)
		if err != nil {
			t.Fatalf("第一次调用GetOnlinePlugins失败: %v", err)
		}
		if len(plugins1) != 0 {
			t.Fatalf("第一次调用GetOnlinePlugins应该返回空列表，实际返回: %v", len(plugins1))
		}

		// 第二次调用，应该从缓存获取
		plugins2, err := manager.GetOnlinePlugins(ctx, false)
		if err != nil {
			t.Fatalf("第二次调用GetOnlinePlugins失败: %v", err)
		}
		if len(plugins2) != 0 {
			t.Fatalf("第二次调用GetOnlinePlugins应该返回空列表，实际返回: %v", len(plugins2))
		}
	})

	// 测试AsyncGetOnlinePlugins缓存
	t.Run("AsyncGetOnlinePluginsCache", func(t *testing.T) {
		ctx := context.Background()

		// 第一次调用，应该从插件市场获取（模拟返回空列表）
		plugins1, err := manager.AsyncGetOnlinePlugins(ctx, false)
		if err != nil {
			t.Fatalf("第一次调用AsyncGetOnlinePlugins失败: %v", err)
		}
		if len(plugins1) != 0 {
			t.Fatalf("第一次调用AsyncGetOnlinePlugins应该返回空列表，实际返回: %v", len(plugins1))
		}

		// 第二次调用，应该从缓存获取
		plugins2, err := manager.AsyncGetOnlinePlugins(ctx, false)
		if err != nil {
			t.Fatalf("第二次调用AsyncGetOnlinePlugins失败: %v", err)
		}
		if len(plugins2) != 0 {
			t.Fatalf("第二次调用AsyncGetOnlinePlugins应该返回空列表，实际返回: %v", len(plugins2))
		}
	})

	// 测试ClearOnlinePluginsCache
	t.Run("ClearOnlinePluginsCache", func(t *testing.T) {
		ctx := context.Background()

		// 调用清除缓存方法
		err := manager.ClearOnlinePluginsCache()
		if err != nil {
			t.Fatalf("调用ClearOnlinePluginsCache失败: %v", err)
		}

		// 再次调用GetOnlinePlugins，应该重新从插件市场获取
		plugins, err := manager.GetOnlinePlugins(ctx, false)
		if err != nil {
			t.Fatalf("清除缓存后调用GetOnlinePlugins失败: %v", err)
		}
		if len(plugins) != 0 {
			t.Fatalf("清除缓存后调用GetOnlinePlugins应该返回空列表，实际返回: %v", len(plugins))
		}
	})

	// 测试强制刷新
	t.Run("ForceRefresh", func(t *testing.T) {
		ctx := context.Background()

		// 第一次调用，应该从插件市场获取
		plugins1, err := manager.GetOnlinePlugins(ctx, false)
		if err != nil {
			t.Fatalf("第一次调用GetOnlinePlugins失败: %v", err)
		}

		// 等待1秒
		time.Sleep(1 * time.Second)

		// 第二次调用，使用force=true，应该重新从插件市场获取
		plugins2, err := manager.GetOnlinePlugins(ctx, true)
		if err != nil {
			t.Fatalf("强制刷新调用GetOnlinePlugins失败: %v", err)
		}

		// 两次调用的结果应该相同（因为都是模拟返回空列表）
		if len(plugins1) != len(plugins2) {
			t.Fatalf("强制刷新前后调用GetOnlinePlugins返回结果长度不一致: %v vs %v", len(plugins1), len(plugins2))
		}
	})
}
