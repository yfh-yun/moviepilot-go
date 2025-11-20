# Actions 动作系统

统一动作层架构，提供完整的动作定义、管理和执行功能。

## 架构概览

```
internal/actions/                    # 统一动作层
├── interfaces/                    # 接口定义
│   ├── action.go                  # 动作接口
│   └── manager.go                 # 管理器接口
├── types/                         # 数据类型
│   ├── context.go                 # 上下文定义
│   └── models.go                  # 数据模型
├── base/                          # 基础实现
│   └── action.go                  # 基础动作实现
├── manager/                       # 管理器实现
│   └── action_manager.go          # 动作管理器
├── implementations/                # 具体实现
│   ├── download.go                # 下载动作
│   ├── scan.go                    # 扫描动作
│   └── ...                        # 其他动作
└── README.md                      # 文档
```

## 核心组件

### 1. 接口层 (interfaces/)

定义了动作系统的核心契约：

- **Action**: 动作基础接口，定义执行、状态管理、缓存等核心功能
- **ActionChain**: 动作链接口，支持动作编排和流水线执行
- **Manager**: 动作管理器接口，负责动作注册、执行和监控
- **ActionFactory**: 动作工厂接口，支持动态创建动作
- **ActionValidator**: 动作验证器接口，提供参数和配置验证
- **ActionObserver**: 动作观察者接口，支持事件通知

### 2. 类型层 (types/)

定义了动作系统使用的数据结构：

- **ActionContext**: 动作执行上下文，包含执行状态、数据、错误等
- **WorkflowExecution**: 工作流执行记录
- **ActionMetrics**: 动作执行指标
- **各种过滤器**: EventFilter, MessageFilter, DownloadFilter 等

### 3. 基础实现层 (base/)

提供动作的基础实现：

- **BaseAction**: 实现了 Action 接口的大部分功能
- 支持状态管理、缓存操作、统计信息、观察者模式
- 提供生命周期管理（Initialize, Cleanup）

### 4. 管理器层 (manager/)

实现动作系统的核心管理功能：

- **ActionManager**: 动作管理器实现
- 支持动作注册、执行监控、配置管理
- 提供健康检查和指标收集

### 5. 具体实现层 (implementations/)

具体的动作实现：

- **DownloadAction**: 下载动作，处理下载任务创建和管理
- **ScanAction**: 文件扫描动作，扫描目录下的媒体文件
- 每个动作都包含完整的配置、验证和执行逻辑

## 设计模式

### 1. 工厂模式
```go
type ActionFactory interface {
    CreateAction(actionType string, config map[string]interface{}) (Action, error)
    GetSupportedActionTypes() []string
    RegisterActionType(actionType string, creator ActionCreator) error
}
```

### 2. 观察者模式
```go
type ActionObserver interface {
    OnActionStart(ctx context.Context, action Action, workflowID int64)
    OnActionComplete(ctx context.Context, action Action, workflowID int64, result *ActionContext, err error)
    OnActionError(ctx context.Context, action Action, workflowID int64, err error)
}
```

### 3. 策略模式
通过 ActionValidator 接口支持不同的验证策略。

### 4. 模板方法模式
BaseAction 提供了执行模板，子类实现具体的 doExecute 方法。

## 使用示例

### 1. 创建动作
```go
// 创建下载动作
downloadAction := implementations.NewDownloadAction(
    "download_001",
    downloadService,
    cache,
)

// 配置动作
config := map[string]interface{}{
    "downloader": "qbittorrent",
    "save_path":  "/downloads",
    "auto_start": true,
}
downloadAction.SetConfig(config)
```

### 2. 注册动作
```go
manager := manager.NewActionManager(nil)
err := manager.RegisterAction(downloadAction)
```

### 3. 执行动作
```go
ctx := context.Background()
workflowID := int64(12345)
params := types.ActionParams{
    "url":   "magnet:?xt=urn:btih:...",
    "title": "Movie Title",
}
context := types.NewActionContext(workflowID)

result, err := manager.ExecuteAction(ctx, "download_001", workflowID, params, context)
```

### 4. 创建动作链
```go
chain := manager.CreateChain("download_chain")
chain.AddAction(downloadAction)
chain.AddAction(scanAction)

result, err := manager.ExecuteChain(ctx, "download_chain", workflowID, params, context)
```

## 配置管理

### 动作配置
每个动作都有详细的配置选项，支持：

- 基础配置：ID、名称、描述
- 执行配置：超时、重试、并发
- 资源限制：内存、CPU、文件大小
- 通知配置：完成、错误、进度通知

### 管理器配置
```go
config := &interfaces.ManagerConfig{
    DefaultTimeout:    300,
    MaxConcurrency:    10,
    EnableRetry:       true,
    DefaultRetryCount: 3,
    CacheEnabled:      true,
    CacheTTL:          3600,
    MetricsEnabled:    true,
    WorkerPoolSize:    5,
    QueueSize:         100,
}
```

## 监控和指标

### 执行指标
- 总执行次数、成功次数、失败次数
- 平均执行时间、P95、P99 延迟
- 资源使用情况：内存、CPU、网络

### 健康检查
```go
health := manager.HealthCheck()
fmt.Printf("状态: %s, 检查项: %v\n", health.Status, health.Checks)
```

### 指标获取
```go
metrics := manager.GetMetrics()
fmt.Printf("成功率: %.2f%%\n", metrics.SuccessRate)
```

## 扩展性

### 添加新动作
1. 在 `implementations/` 目录下创建新的动作文件
2. 实现 `Action` 接口
3. 可选：实现 `ActionValidator` 接口
4. 注册到管理器

### 自定义验证器
```go
type CustomValidator struct{}

func (cv *CustomValidator) ValidateConfig(config map[string]interface{}) error {
    // 自定义验证逻辑
    return nil
}
```

### 添加观察者
```go
type CustomObserver struct{}

func (co *CustomObserver) OnActionStart(ctx context.Context, action Action, workflowID int64) {
    // 自定义通知逻辑
}

manager.AddObserver(&CustomObserver{})
```

## 最佳实践

1. **动作设计**: 保持动作单一职责，避免过于复杂
2. **错误处理**: 使用结构化错误信息，支持重试机制
3. **配置验证**: 实现完整的参数和配置验证
4. **日志记录**: 使用结构化日志，包含足够的上下文信息
5. **测试**: 为每个动作编写单元测试和集成测试
6. **性能**: 注意资源使用，避免内存泄漏
7. **监控**: 利用内置的指标和健康检查功能

## 与原架构的对比

### 原架构问题
- `internal/actions` 和 `internal/service/actions` 重复定义
- 接口分散，依赖关系混乱
- 缺乏统一的管理和监控机制

### 新架构优势
- 统一的接口定义和实现
- 清晰的分层架构
- 完整的管理和监控功能
- 更好的扩展性和可维护性
- 支持工厂模式、观察者模式等设计模式

## 迁移指南

1. 将现有的动作迁移到 `implementations/` 目录
2. 使用 `BaseAction` 作为基础类
3. 实现相应的验证器
4. 更新调用代码使用新的管理器接口
5. 逐步移除旧的重复定义

这个重构后的架构提供了一个完整、可扩展、易维护的动作系统，为 MoviePilot 项目提供了强大的工作流执行能力。