# actions/ 动作系统迁移计划

> Python: `app/actions/`  
> Go: `internal/business/workflows/actions/` + 领域服务

---

## 1. Python `actions/` 概览

- **职责**：
  - 描述工作流中的“单个动作”（如下载、整理、通知等）。
  - 将上游事件（订阅命中、下载完成等）转换为可执行步骤。
  - 与 modules/helper 等协作完成实际操作。

---

## 2. Go 目标设计

### 2.1 目录结构

```
internal/business/workflows/actions/
├── actions.go              # 动作接口定义
├── base_structs.go         # 基础结构体
├── manager.go              # 动作管理器
├── factory.go              # 动作工厂
├── services.go             # 服务依赖注入
├── chain.go                # 动作链
├── add_download_action.go  # 具体动作实现
├── add_subscribe_action.go # 具体动作实现
├── fetch_downloads_action.go # 具体动作实现
├── fetch_medias_action.go  # 具体动作实现
├── fetch_rss_action.go     # 具体动作实现
├── fetch_torrents_action.go # 具体动作实现
├── filter_medias_action.go # 具体动作实现
├── filter_torrents_action.go # 具体动作实现
├── invoke_plugin_action.go # 具体动作实现
├── note_action.go          # 具体动作实现
├── scan_file_action.go     # 具体动作实现
├── scrape_file_action.go   # 具体动作实现
├── send_event_action.go    # 具体动作实现
├── send_message_action.go  # 具体动作实现
├── transfer_file_action.go # 具体动作实现
├── params/                 # 动作参数定义
│   ├── base_params.go      # 基础参数
│   ├── add_download_params.go # 具体动作参数
│   └── ...
└── adapters/               # 适配器
    ├── download_adapter.go # 下载服务适配器
    ├── media_adapter.go    # 媒体服务适配器
    └── ...
```

### 2.2 核心接口设计

```go
// Action 动作接口
type Action interface {
    // ID 获取动作ID
    ID() string
    // Name 获取动作名称
    Name() string
    // Description 获取动作描述
    Description() string
    // Type 获取动作类型
    Type() string
    // Execute 执行动作
    Execute(ctx context.Context, data map[string]interface{}) error
    // Validate 验证动作参数
    Validate() error
    // ContinueOnError 执行失败时是否继续
    ContinueOnError() bool
}

// ActionWithParams 带参数的动作接口
type ActionWithParams interface {
    Action
    // Params 获取动作参数
    Params() interface{}
}
```

### 2.3 协作关系

- 上游：workflow 引擎（`internal/business/workflows/engine.go`）。
- 下游：业务 Service（下载、媒体、通知等）和 integration/modules 封装。
- 依赖注入：通过 services.go 注入所需服务。

---

## 3. 映射表（按实际实现补全）

| Python 动作模块 | Go 目标位置 | 实现文件 | 状态 | 说明 |
|-----------------|-------------|----------|------|------|
| `add_download.py` | `internal/business/workflows/actions/` + `services/download/` | `add_download_action.go` | ✅ 已实现 | 将命中结果加入下载队列 |
| `add_subscribe.py` | `internal/business/workflows/actions/` + `services/subscribe/` | `add_subscribe_action.go` | ✅ 已实现 | 创建/更新订阅 |
| `fetch_downloads.py` | `internal/business/workflows/actions/` + `services/download/` | `fetch_downloads_action.go` | ✅ 已实现 | 从下载器拉取任务状态 |
| `fetch_medias.py` | `internal/business/workflows/actions/` + `services/media/` | `fetch_medias_action.go` | ✅ 已实现 | 拉取/刷新媒体信息 |
| `fetch_rss.py` | `internal/business/workflows/actions/` + `services/subscribe/` | `fetch_rss_action.go` | ✅ 已实现 | 拉取 RSS / 站点订阅源 |
| `fetch_torrents.py` | `internal/business/workflows/actions/` + `services/search/` | `fetch_torrents_action.go` | ✅ 已实现 | 抓取种子列表 |
| `filter_medias.py` | `internal/business/workflows/actions/` + `business/policies/` | `filter_medias_action.go` | ✅ 已实现 | 按规则过滤媒体候选 |
| `filter_torrents.py` | `internal/business/workflows/actions/` + `business/policies/` | `filter_torrents_action.go` | ✅ 已实现 | 按规则过滤种子 |
| `invoke_plugin.py` | `internal/business/workflows/actions/` + `services/plugin/` | `invoke_plugin_action.go` | ✅ 已实现 | 调用插件完成自定义动作 |
| `note.py` | `internal/business/workflows/actions/` | `note_action.go` | ✅ 已实现 | 记录备注/日志型动作 |
| `scan_file.py` | `internal/business/workflows/actions/` + `services/storage/` | `scan_file_action.go` | ✅ 已实现 | 扫描本地文件，构建待处理列表 |
| `scrape_file.py` | `internal/business/workflows/actions/` + `services/media/` | `scrape_file_action.go` | ✅ 已实现 | 对文件进行刮削识别 |
| `send_event.py` | `internal/business/workflows/actions/` + `events/` | `send_event_action.go` | ✅ 已实现 | 抛出领域事件，驱动后续流程 |
| `send_message.py` | `internal/business/workflows/actions/` + `services/notification/` | `send_message_action.go` | ✅ 已实现 | 发送通知/消息 |
| `transfer_file.py` | `internal/business/workflows/actions/` + `services/transfer/` | `transfer_file_action.go` | ✅ 已实现 | 文件转存/整理入库 |

---

## 4. 迁移步骤

### 4.1 已完成步骤

1. **动作接口定义**：定义了 Action 和 ActionWithParams 接口
2. **基础结构实现**：实现了基础结构体和辅助功能
3. **动作实现**：实现了所有 Python 动作对应的 Go 动作
4. **参数定义**：为每个动作定义了对应的参数结构体
5. **适配器实现**：实现了各种服务的适配器
6. **依赖注入**：通过 services.go 实现了服务的依赖注入
7. **动作工厂**：实现了动作工厂，用于创建动作实例
8. **动作管理器**：实现了动作管理器，用于管理动作
9. **单元测试**：为部分动作编写了单元测试

### 4.2 后续优化步骤

1. **完善单元测试**：为所有动作编写单元测试
2. **优化动作链**：增强动作链的功能和灵活性
3. **事件驱动优化**：优化事件驱动机制
4. **性能优化**：优化动作执行性能
5. **文档完善**：完善动作接口和使用文档

---

## 5. 实现细节

### 5.1 动作执行流程

1. 工作流引擎调用动作工厂创建动作实例
2. 动作工厂根据动作类型创建对应的动作实例
3. 动作实例通过依赖注入获取所需服务
4. 工作流引擎调用动作的 Execute 方法执行动作
5. 动作执行过程中调用对应的服务完成实际操作
6. 动作执行完成后返回结果

### 5.2 依赖注入

所有动作都通过依赖注入获取所需服务，避免了直接依赖具体实现，提高了可测试性和可扩展性。

```go
// services.go
type ServiceProvider struct {
    downloadService  download.Service
    subscribeService subscribe.Service
    mediaService     media.Service
    // 其他服务...
}

// Action 结构体包含 ServiceProvider
type AddDownloadAction struct {
    ServiceProvider
    // 其他字段...
}
```

### 5.3 参数验证

所有动作都实现了 Validate 方法，用于验证动作参数的合法性。

```go
// add_download_action.go
func (a *AddDownloadAction) Validate() error {
    if a.Params().TorrentURL == "" {
        return errors.New("torrent url is required")
    }
    return nil
}
```

---

## 6. 检查清单

- [x] `app/actions/` 中的每个动作在映射表中都有对应条目。
- [x] Go 动作实现不包含直接的 HTTP/DB 调用，而是通过 Service/接口完成。
- [x] 动作执行过程有足够的结构化日志，便于排查工作流问题。
- [x] 所有动作都实现了 Action 接口。
- [x] 所有动作都通过依赖注入获取所需服务。
- [x] 所有动作都实现了参数验证。

---

## 7. 迁移计划归纳

### 7.1 迁移目标

将 Python 项目的 `app/actions/` 目录迁移到 Go 项目的 `internal/business/workflows/actions/` 目录，实现从 Python 动作系统到 Go 动作系统的迁移，包括：

- 动作接口设计
- 动作实现
- 参数定义
- 依赖注入
- 动作工厂和管理器
- 单元测试

### 7.2 迁移策略

1. **接口先行**：先定义动作接口，明确动作的职责和协作关系
2. **基础实现**：实现基础结构体和辅助功能
3. **动作迁移**：逐个迁移 Python 动作到 Go 动作
4. **依赖注入**：使用依赖注入管理服务依赖
5. **测试驱动**：为每个动作编写单元测试
6. **优化迭代**：不断优化动作系统的性能和可扩展性

### 7.3 迁移成果

1. **完整的动作系统**：已实现所有 Python 动作对应的 Go 动作
2. **清晰的接口设计**：定义了明确的动作接口
3. **依赖注入**：通过依赖注入管理服务依赖，提高了可测试性和可扩展性
4. **完善的参数验证**：所有动作都实现了参数验证
5. **结构化日志**：动作执行过程有足够的结构化日志
6. **动作工厂和管理器**：实现了动作工厂和管理器，便于管理和使用动作

### 7.4 后续优化方向

1. **完善单元测试**：为所有动作编写单元测试
2. **优化动作链**：增强动作链的功能和灵活性
3. **事件驱动优化**：优化事件驱动机制
4. **性能优化**：优化动作执行性能
5. **文档完善**：完善动作接口和使用文档

---

## 8. 结论

动作系统迁移已基本完成，所有 Python 动作都已在 Go 中实现，并且实现了完善的接口设计、依赖注入、参数验证和结构化日志。动作系统可以正常运行，支持工作流引擎的调用。后续将继续优化动作系统的性能和可扩展性，完善单元测试和文档。

---

**相关文档**：
- [chain-migration.md](./chain-migration.md)
- [workflow-migration-plan.md](./workflow-migration-plan.md)
