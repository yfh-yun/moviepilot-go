package plugin

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEventManager(t *testing.T) {
	ctx := context.Background()
	em := NewEventManager(ctx)
	defer em.Close()

	// 测试事件发布和订阅
	eventReceived := false
	var receivedEvent *Event

	// 订阅事件
	subscriptionID, err := em.SubscribeEvent(EventTypePluginStarted, func(ctx context.Context, event *Event) error {
		eventReceived = true
		receivedEvent = event
		return nil
	}, nil)

	assert.NoError(t, err)
	assert.NotEmpty(t, subscriptionID)

	// 发布事件
	event := &Event{
		Type:   EventTypePluginStarted,
		Source: "test-plugin",
		Data: map[string]interface{}{
			"plugin_id":   "test-plugin",
			"plugin_name": "Test Plugin",
			"plugin_version": "1.0.0",
		},
	}

	err = em.PublishEvent(ctx, event)
	assert.NoError(t, err)

	// 等待事件处理完成
	time.Sleep(100 * time.Millisecond)

	// 验证事件是否被接收
	assert.True(t, eventReceived)
	assert.Equal(t, event.Type, receivedEvent.Type)
	assert.Equal(t, event.Source, receivedEvent.Source)
	assert.Equal(t, event.Data["plugin_id"], receivedEvent.Data["plugin_id"])

	// 测试取消订阅
	eventReceived = false
	err = em.UnsubscribeEvent(subscriptionID)
	assert.NoError(t, err)

	// 再次发布事件，应该不会被接收
	err = em.PublishEvent(ctx, event)
	assert.NoError(t, err)

	// 等待事件处理完成
	time.Sleep(100 * time.Millisecond)

	// 验证事件是否没有被接收
	assert.False(t, eventReceived)

	// 测试批量订阅
	eventTypes := []EventType{EventTypePluginStarted, EventTypePluginStopped, EventTypePluginError}
	eventCount := 0

	subscriptionIDs, err := em.SubscribeMultipleEvents(eventTypes, func(ctx context.Context, event *Event) error {
		eventCount++
		return nil
	}, nil)

	assert.NoError(t, err)
	assert.Len(t, subscriptionIDs, len(eventTypes))

	// 发布多个事件
	for _, eventType := range eventTypes {
		err = em.PublishEvent(ctx, &Event{
			Type:   eventType,
			Source: "test-plugin",
		})
		assert.NoError(t, err)
	}

	// 等待事件处理完成
	time.Sleep(100 * time.Millisecond)

	// 验证所有事件都被接收
	assert.Equal(t, len(eventTypes), eventCount)

	// 取消之前的订阅，避免影响后续测试
	for _, id := range subscriptionIDs {
		_ = em.UnsubscribeEvent(id)
	}

	// 测试异步事件发布
	asyncEventReceived := false
	_, _ = em.SubscribeEvent(EventTypePluginConfigChanged, func(ctx context.Context, event *Event) error {
		asyncEventReceived = true
		return nil
	}, nil)

	em.PublishEventAsync(&Event{
		Type:   EventTypePluginConfigChanged,
		Source: "test-plugin",
	})

	// 等待事件处理完成
	time.Sleep(100 * time.Millisecond)

	// 验证事件是否被接收
	assert.True(t, asyncEventReceived)
}
