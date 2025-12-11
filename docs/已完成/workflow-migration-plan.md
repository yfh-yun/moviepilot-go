# 工作流系统迁移计划

> Python: `app/core/workflow.py`  
> Go: `internal/business/workflows/` + `internal/business/services/workflow/`

---

## 1. Python 工作流系统概览

### 1.1 核心功能

- **工作流定义**：节点、条件、动作的配置
- **工作流执行**：顺序执行、条件分支、错误处理
- **事件驱动**：基于事件触发工作流
- **工作流管理**：创建、更新、删除、启用/禁用

### 1.2 典型使用场景

- 下载完成 → 执行文件整理工作流
- 订阅更新 → 执行订阅检查工作流
- 站点更新 → 执行站点检查工作流

---

## 2. Go 目标设计

### 2.1 分层设计

| 层级 | 位置 | 职责 |
|------|------|------|
| 业务层 | `internal/business/workflows/` | 工作流引擎、动作实现 |
| 服务层 | `internal/business/services/workflow/` | 工作流管理服务 |

### 2.2 核心模块

**工作流引擎**：

```go
// internal/business/workflows/engine.go
type EventEngine struct {
    logger         *zap.Logger
    workflowRepo   interfaces.WorkflowRepository
    eventBus       infrastructure_events.Bus
    actionManager  *actions.Manager
    eventWorkflows map[string][]uint // eventType -> workflowIDs
}

func (e *EventEngine) Init(ctx context.Context) error
func (e *EventEngine) UpdateWorkflowEvent(ctx context.Context, workflow *database.Workflow) error
func (e *EventEngine) LoadWorkflowEvents(ctx context.Context, workflowID string) error
func (e *EventEngine) handleEvent(ctx context.Context, ev *domains_events.Event) error
```

**工作流动作**：

```go
// internal/business/workflows/actions/factory.go
type Manager struct {
    logger   *zap.Logger
    handlers map[string]Handler
}

func (m *Manager) Register(handler Handler) error
func (m *Manager) Get(handlerType string) (Handler, error)
func (m *Manager) IsRunning(workflowID uint) bool

// internal/business/workflows/actions/chain.go
type Handler interface {
    ID() string
    Name() string
    Description() string
    Execute(ctx context.Context, data map[string]interface{}) error
}
```

### 2.3 迁移步骤

1. **实现工作流引擎**：在 `internal/business/workflows/` 中实现事件驱动的工作流引擎
2. **实现动作系统**：在 `internal/business/workflows/actions/` 中实现工作流动作
3. **实现工作流服务**：在 `internal/business/services/workflow/` 中实现工作流管理服务
4. **替换 Python 工作流调用**：将 Python 中的工作流相关调用替换为 Go 实现
5. **集成到事件系统**：将工作流引擎与事件系统集成

---

## 3. 集成与使用

### 3.1 工作流执行示例

```go
// 事件触发工作流
event := domains_events.NewEvent("download.completed", map[string]interface{}{
    "file_path": "/path/to/file",
})

eventBus.PublishBroadcast(ctx, event.Type, event.Data, event.Priority)

// 工作流引擎处理事件
eventEngine.handleEvent(ctx, event)
```

### 3.2 动作执行示例

```go
// 执行添加下载动作
action, err := actionManager.Get("add_download")
if err != nil {
    return err
}

err = action.Execute(ctx, map[string]interface{}{
    "url": "/path/to/torrent",
    "name": "资源名称",
})
```

---

## 4. 检查清单

- [x] 实现事件驱动工作流引擎
- [x] 实现工作流动作系统
- [x] 实现工作流管理服务
- [x] 集成到事件系统
- [x] 实现基本工作流动作（添加下载、添加订阅等）

---

## 5. 后续优化

1. **支持更多动作类型**：实现更多工作流动作
2. **工作流可视化**：添加工作流可视化配置界面
3. **工作流版本控制**：支持工作流的版本管理
4. **工作流执行监控**：添加工作流执行监控和日志
5. **工作流模板**：支持工作流模板的创建和使用
