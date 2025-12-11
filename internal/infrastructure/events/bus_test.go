package events

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	domains "moviepilot-go/internal/business/domains/events"
)

func TestBus_SubscribeBroadcast(t *testing.T) {
	bus := NewBus()
	ctx := context.Background()

	// 定义测试事件类型
	testEventType := domains.EventType("test.broadcast")

	// 测试计数器和同步锁
	var wg sync.WaitGroup
	var count int
	var mu sync.Mutex

	// 订阅广播事件
	handler := func(ctx context.Context, event *domains.Event) error {
		mu.Lock()
		count++
		mu.Unlock()
		wg.Done()
		return nil
	}

	bus.SubscribeBroadcast(testEventType, handler)

	// 发布广播事件
	for i := 0; i < 5; i++ {
		wg.Add(1)
		err := bus.PublishBroadcast(ctx, testEventType, "test-data", 10)
		assert.NoError(t, err)
	}

	// 等待所有处理器执行完成
	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		mu.Lock()
		assert.Equal(t, 5, count)
		mu.Unlock()
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for broadcast events to be handled")
	}
}

func TestBus_SubscribeChain(t *testing.T) {
	bus := NewBus()
	ctx := context.Background()

	// 定义测试事件类型
	testEventType := domains.ChainEventType("test.chain")

	// 测试结果和同步锁
	var results []string
	var mu sync.Mutex

	// 定义三个处理器，按优先级执行
	handler1 := func(ctx context.Context, event *domains.Event) error {
		mu.Lock()
		results = append(results, "handler1")
		mu.Unlock()
		return nil
	}

	handler2 := func(ctx context.Context, event *domains.Event) error {
		mu.Lock()
		results = append(results, "handler2")
		mu.Unlock()
		return nil
	}

	handler3 := func(ctx context.Context, event *domains.Event) error {
		mu.Lock()
		results = append(results, "handler3")
		mu.Unlock()
		return nil
	}

	// 按不同优先级订阅链式事件
	bus.SubscribeChain(testEventType, 20, handler1) // 低优先级
	bus.SubscribeChain(testEventType, 10, handler2) // 中优先级
	bus.SubscribeChain(testEventType, 5, handler3)  // 高优先级

	// 分发链式事件
	_, err := bus.DispatchChain(ctx, testEventType, "test-data", 10)
	assert.NoError(t, err)

	// 验证处理器按优先级顺序执行
	expectedResults := []string{"handler3", "handler2", "handler1"}
	assert.Equal(t, expectedResults, results)
}

func TestBus_DisableHandler(t *testing.T) {
	bus := NewBus()
	ctx := context.Background()

	// 定义测试事件类型
	testEventType := domains.EventType("test.disable")

	// 测试计数器和同步锁
	var count int
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 订阅广播事件，获取handlerID
	handler := func(ctx context.Context, event *domains.Event) error {
		mu.Lock()
		count++
		mu.Unlock()
		wg.Done()
		return nil
	}

	handlerID := bus.SubscribeBroadcast(testEventType, handler)

	// 测试1：检查处理器列表，默认应该是启用的
	handlers := bus.ListHandlers()
	found := false
	for _, h := range handlers {
		if h.HandlerID == handlerID {
			found = true
			assert.Equal(t, "enabled", h.Status)
			break
		}
	}
	assert.True(t, found, "handler not found in list")

	// 测试2：禁用处理器
	bus.DisableHandler(handlerID)

	// 检查处理器列表，应该是禁用的
	handlers = bus.ListHandlers()
	found = false
	for _, h := range handlers {
		if h.HandlerID == handlerID {
			found = true
			assert.Equal(t, "disabled", h.Status)
			break
		}
	}
	assert.True(t, found, "handler not found in list")

	// 测试3：启用处理器
	bus.EnableHandler(handlerID)

	// 检查处理器列表，应该是启用的
	handlers = bus.ListHandlers()
	found = false
	for _, h := range handlers {
		if h.HandlerID == handlerID {
			found = true
			assert.Equal(t, "enabled", h.Status)
			break
		}
	}
	assert.True(t, found, "handler not found in list")

	// 测试4：处理器应该能正常执行
	wg.Add(1)
	err := bus.PublishBroadcast(ctx, testEventType, "test-data", 10)
	assert.NoError(t, err)

	// 等待处理器执行完成
	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		mu.Lock()
		assert.Equal(t, 1, count)
		mu.Unlock()
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for handler to execute")
	}
}

func TestBus_ListHandlers(t *testing.T) {
	bus := NewBus()

	// 定义测试事件类型
	broadcastType := domains.EventType("test.list.broadcast")
	chainType := domains.ChainEventType("test.list.chain")

	// 订阅事件
	handler1 := func(ctx context.Context, event *domains.Event) error {
		return nil
	}

	handler2 := func(ctx context.Context, event *domains.Event) error {
		return nil
	}

	bus.SubscribeBroadcast(broadcastType, handler1)
	bus.SubscribeChain(chainType, 10, handler2)

	// 获取处理器列表
	handlers := bus.ListHandlers()

	// 验证处理器数量
	assert.Equal(t, 2, len(handlers))

	// 验证处理器信息
	handlerMap := make(map[string]HandlerInfo)
	for _, h := range handlers {
		handlerMap[h.EventType] = h
	}

	// 验证广播事件处理器
	broadcastHandler, exists := handlerMap[string(broadcastType)]
	assert.True(t, exists)
	assert.Equal(t, string(broadcastType), broadcastHandler.EventType)
	assert.Equal(t, "enabled", broadcastHandler.Status)
	assert.Nil(t, broadcastHandler.Priority)

	// 验证链式事件处理器
	chainHandler, exists := handlerMap[string(chainType)]
	assert.True(t, exists)
	assert.Equal(t, string(chainType), chainHandler.EventType)
	assert.Equal(t, "enabled", chainHandler.Status)
	assert.NotNil(t, chainHandler.Priority)
	assert.Equal(t, 10, *chainHandler.Priority)
}
