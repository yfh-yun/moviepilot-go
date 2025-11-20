# 动作系统重构迁移指南

本文档描述了如何从旧的 `internal/actions` 和 `internal/service/actions` 架构迁移到新的统一动作层架构。

## 架构变更概览

### 旧架构
```
internal/actions                    # 基础结构定义
├── base_structs.go               # 基础接口和结构
└── ...

internal/service/actions            # 业务逻辑实现
├── base_action.go                 # 动作基类
├── action_manager.go              # 动作管理器
├── types/                         # 数据类型
└── [26个具体动作文件]              # 具体实现
```

### 新架构
```
internal/actions                    # 统一动作层
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
│   ├── file_scanner.go            # 文件扫描器
│   ├── media_fetcher.go           # 媒体获取器
│   ├── message_sender.go          # 消息发送器
│   ├── plugin_invoker.go          # 插件调用器
│   └── data_sources.go            # 数据源实现
├── registry/                      # 注册表
│   └── registry.go                # 动作注册表
├── examples/                      # 使用示例
│   └── example_usage.go           # 示例代码
└── tests/                         # 测试文件
    ├── actions_test.go            # 单元测试
    └── benchmark_test.go          # 基准测试
```

## 主要改进

1. **消除重复定义**: 统一了接口和实现，避免了重复代码
2. **清晰的分层**: 明确分离了接口、类型、基础实现和具体实现
3. **注册表模式**: 使用注册表管理动作，支持动态加载和卸载
4. **更好的测试支持**: 提供完整的单元测试和基准测试
5. **示例和文档**: 包含详细的使用示例和迁移指南

## 迁移步骤

### 步骤1: 更新导入路径

**旧导入**:
```go
import "github.com/yfh-yun/moviepilot-go/internal/actions"
import "github.com/yfh-yun/moviepilot-go/internal/service/actions"
```

**新导入**:
```go
import "github.com/yfh-yun/moviepilot-go/internal/actions"
import "github.com/yfh-yun/moviepilot-go/internal/actions/interfaces"
import "github.com/yfh-yun/moviepilot-go/internal/actions/types"
```

### 步骤2: 更新动作创建方式

**旧方式**:
```go
action := &actions.FileScanner{}
action.Initialize()
```

**新方式**:
```go
reg := registry.GetDefaultRegistry()
action, err := reg.CreateAction("file_scanner")
if err != nil {
    return err
}
err = action.Initialize()
```

### 步骤3: 更新接口调用

**旧接口**:
```go
type BaseAction interface {
    Name() string
    Description() string
    Data() interface{}
    Done() bool
    Success() bool
    Message() string
    SetDone(message string)
    Execute(workflowID int, params interface{}, context *ActionContext) (*ActionContext, error)
}
```

**新接口**:
```go
type Action interface {
    Name() string
    Description() string
    Version() string
    Author() string
    Category() string
    Tags() []string
    
    IsDone() bool
    IsSuccess() bool
    GetMessage() string
    SetDone(message string)
    SetError(message string)
    
    Execute(ctx context.Context, workflowID int64, params map[string]interface{}, context *types.ActionContext) (*types.ActionContext, error)
    
    GetData() map[string]interface{}
    SetData(key string, value interface{})
    
    CheckCache(ctx context.Context, workflowID int64, key string) bool
    GetCache(ctx context.Context, workflowID int64, key string) (interface{}, error)
    SaveCache(ctx context.Context, workflowID int64, key string, data interface{}, ttl time.Duration) error
    ClearCache(ctx context.Context, workflowID int64) error
    
    Initialize() error
    Cleanup() error
}
```

### 步骤4: 更新动作实现

**旧实现示例**:
```go
type FileScanner struct {
    actions.BaseAction
    config *FileScannerConfig
}

func (fs *FileScanner) Execute(workflowID int, params interface{}, context *actions.ActionContext) (*actions.ActionContext, error) {
    // 实现逻辑
}
```

**新实现示例**:
```go
type FileScanner struct {
    *base.Action
    config *FileScannerConfig
}

func NewFileScanner() interfaces.Action {
    return &FileScanner{
        Action: base.NewAction("FileScanner", "文件扫描器"),
        config: &FileScannerConfig{...},
    }
}

func (fs *FileScanner) Execute(ctx context.Context, workflowID int64, params map[string]interface{}, actionContext *types.ActionContext) (*types.ActionContext, error) {
    // 实现逻辑
}
```

### 步骤5: 更新上下文类型

**旧上下文**:
```go
type ActionContext struct {
    WorkflowID int
    Variables  map[string]interface{}
    Metadata   map[string]string
    CreatedAt  time.Time
}
```

**新上下文**:
```go
type ActionContext struct {
    WorkflowID int64
    Variables  map[string]interface{}
    Metadata   map[string]string
    CreatedAt  time.Time
    UpdatedAt  time.Time
    Status     string
}
```

## 具体动作迁移映射

| 旧动作文件 | 新动作实现 | 动作名称 |
|-----------|-----------|----------|
| `file_scanner.go` | `implementations/file_scanner.go` | `file_scanner` |
| `media_fetcher.go` | `implementations/media_fetcher.go` | `media_fetcher` |
| `message_sender.go` | `implementations/message_sender.go` | `message_sender` |
| `plugin_invoker.go` | `implementations/plugin_invoker.go` | `plugin_invoker` |
| `download_manager.go` | `implementations/download.go` | `download` |
| `scan_file.go` | `implementations/scan.go` | `scan` |

## 数据模型迁移

### 媒体信息模型
```go
// 新的媒体信息模型
type MediaInfo struct {
    TMDBID         int      `json:"tmdb_id"`
    IMDBID         string   `json:"imdb_id"`
    Title          string   `json:"title"`
    OriginalTitle  string   `json:"original_title"`
    Overview       string   `json:"overview"`
    Year           int      `json:"year"`
    Type           string   `json:"type"`
    Rating         float64  `json:"rating"`
    VoteCount      int      `json:"vote_count"`
    Popularity     float64  `json:"popularity"`
    PosterPath     string   `json:"poster_path"`
    PosterURL      string   `json:"poster_url"`
    BackdropPath   string   `json:"backdrop_path"`
    BackdropURL    string   `json:"backdrop_url"`
    Genres         []string `json:"genres"`
    GenreIDs       []int    `json:"genre_ids"`
    Source         string   `json:"source"`
}
```

## 测试迁移

### 运行测试
```bash
# 运行所有测试
go test ./internal/actions/...

# 运行单元测试
go test ./internal/actions/actions_test.go

# 运行基准测试
go test -bench=. ./internal/actions/benchmark_test.go

# 生成测试覆盖率报告
go test -cover ./internal/actions/...
```

### 编写新测试
```go
func TestMyAction(t *testing.T) {
    reg := registry.GetDefaultRegistry()
    
    action, err := reg.CreateAction("my_action")
    require.NoError(t, err)
    
    err = action.Initialize()
    require.NoError(t, err)
    defer action.Cleanup()
    
    params := map[string]interface{}{
        "param1": "value1",
        "param2": 123,
    }
    
    actionContext := &types.ActionContext{
        WorkflowID: 12345,
        Variables:  make(map[string]interface{}),
        Metadata:   make(map[string]string),
        CreatedAt:  time.Now(),
    }
    
    ctx := context.Background()
    updatedContext, err := action.Execute(ctx, 12345, params, actionContext)
    require.NoError(t, err)
    
    assert.True(t, action.IsDone())
    assert.True(t, action.IsSuccess())
}
```

## 向后兼容性

为了保持向后兼容性，新架构提供了迁移辅助函数：

```go
// migration.go 提供迁移辅助函数
func MigrateOldAction(oldAction interface{}) (interfaces.Action, error) {
    // 迁移逻辑
}

func ConvertOldContext(oldCtx *OldActionContext) *types.ActionContext {
    // 转换逻辑
}
```

## 性能改进

新架构提供了以下性能改进：

1. **更快的动作创建**: 使用工厂模式和对象池
2. **更好的缓存**: 分层缓存和智能缓存失效
3. **并发支持**: 原生支持并发执行
4. **内存优化**: 减少内存分配和垃圾回收压力

## 故障排除

### 常见问题

1. **导入错误**: 确保使用正确的导入路径
2. **接口不匹配**: 检查是否实现了所有必需的方法
3. **类型错误**: 确保使用正确的数据类型
4. **注册失败**: 检查动作名称是否唯一

### 调试技巧

1. 启用详细日志:
```go
logger.SetLevel(zap.DebugLevel)
```

2. 使用注册表调试:
```go
actions := reg.ListActions()
fmt.Printf("Registered actions: %v\n", actions)
```

3. 检查动作信息:
```go
info, err := reg.GetActionInfo("action_name")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Action info: %+v\n", info)
```

## 最佳实践

1. **动作设计**: 保持动作单一职责，避免复杂逻辑
2. **错误处理**: 使用结构化错误处理，提供详细错误信息
3. **日志记录**: 使用结构化日志，包含足够的上下文信息
4. **测试覆盖**: 确保每个动作都有完整的单元测试
5. **文档**: 为每个动作提供清晰的文档和示例

## 迁移检查清单

- [ ] 更新所有导入路径
- [ ] 重构动作实现以符合新接口
- [ ] 更新动作创建和注册代码
- [ ] 迁移数据模型和上下文类型
- [ ] 更新测试代码
- [ ] 验证所有功能正常工作
- [ ] 运行性能基准测试
- [ ] 更新文档和示例

## 获取帮助

如果在迁移过程中遇到问题，可以：

1. 查看示例代码: `internal/actions/examples/example_usage.go`
2. 运行测试: `go test ./internal/actions/...`
3. 查看文档: `internal/actions/README.md`
4. 检查迁移工具: `internal/actions/migration.go`

## 总结

新的统一动作层架构提供了更好的可维护性、可扩展性和性能。通过遵循本迁移指南，您可以平滑地从旧架构迁移到新架构，同时保持代码的功能性和稳定性。