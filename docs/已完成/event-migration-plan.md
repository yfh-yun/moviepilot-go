# Event 事件系统迁移计划

> Python: `app/core/event.py`  
> Go: `internal/business/domains/events/` + `internal/infrastructure/events/`

---

## 1. Python 事件系统概览

### 1.1 核心功能

- **事件发布/订阅**：`send_event` / `subscribe`
- **事件类型**：20+ 种预定义事件类型（如 `transfer.complete`、`download.added`）
- **事件数据**：`event_type` + `data` 结构
- **异步处理**：事件发布后异步执行订阅者回调

### 1.2 典型使用场景

- 站点更新 → 触发订阅检查
- 下载完成 → 触发文件整理
- 订阅完成 → 发送通知

---

## 2. Go 目标设计

### 2.1 分层设计

| 层级 | 位置 | 职责 |
|------|------|------|
| 领域层 | `internal/business/domains/events/` | 事件模型、领域事件定义 |
| 基础设施层 | `internal/infrastructure/events/` | 事件总线、事件管理器实现 |

### 2.2 核心接口

**领域事件模型**：

```go
// internal/business/domains/events/event.go
type EventType string

type Event struct {
    ID        string      // 事件唯一标识
    Type      EventType   // 事件类型
    Data      interface{} // 事件数据
    Priority  int         // 事件优先级
    CreatedAt time.Time   // 事件创建时间
}
```

**事件总线接口**：

```go
// internal/infrastructure/events/bus.go
type Bus interface {
    // 订阅广播事件
    SubscribeBroadcast(eventType EventType, handler HandlerFunc) string
    // 发布广播事件
    PublishBroadcast(ctx context.Context, eventType EventType, data interface{}, priority int) error
    // 订阅链式事件
    SubscribeChain(eventType ChainEventType, priority int, handler HandlerFunc) string
    // 分发链式事件
    DispatchChain(ctx context.Context, eventType ChainEventType, data interface{}, priority int) (*Event, error)
}
```

### 2.3 事件类型映射

| Python 事件类型 | Go 事件类型 | 说明 |
|----------------|-------------|------|
| `transfer.complete` | `EventTypeTransferComplete` | 整理完成 |
| `download.added` | `EventTypeDownloadAdded` | 添加下载 |
| `subscribe.complete` | `EventTypeSubscribeComplete` | 订阅已完成 |
| `metadata.scrape` | `EventTypeMetadataScrape` | 刮削元数据 |
| `config.updated` | `EventTypeConfigChanged` | 配置项更新 |
| `plugin.reload` | `EventTypePluginReload` | 插件重载 |

---

## 3. 迁移步骤

1. **定义事件类型**：在 `internal/models/types/types.go` 中定义所有事件类型常量
2. **实现领域事件模型**：在 `internal/business/domains/events/` 中实现事件模型
3. **实现事件总线**：在 `internal/infrastructure/events/` 中实现事件总线
4. **替换 Python 事件调用**：将 Python 中的 `send_event` 替换为 Go 中的 `eventBus.PublishBroadcast`
5. **实现事件订阅**：在各模块中实现事件订阅逻辑

---

## 4. 集成与使用

### 4.1 事件发布示例

```go
// 发布事件
eventBus.PublishBroadcast(ctx, EventTypeTransferComplete, map[string]interface{}{
    "file_path": "/path/to/file",
    "status":    "success",
}, 10)
```

### 4.2 事件订阅示例

```go
// 订阅事件
eventBus.SubscribeBroadcast(EventTypeTransferComplete, func(ctx context.Context, event *Event) error {
    // 处理事件
    return nil
})
```

---

## 5. 检查清单

- [x] 定义所有事件类型常量
- [x] 实现领域事件模型
- [x] 实现事件总线
- [x] 实现事件订阅机制
- [x] 实现事件发布机制
- [x] 集成到业务流程中

---

## 6. 后续优化

1. **事件持久化**：支持将事件持久化到数据库
2. **事件重试机制**：支持事件处理失败重试
3. **事件溯源**：支持基于事件重建系统状态
4. **事件监控**：添加事件处理监控和统计
