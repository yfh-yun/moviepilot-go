# chain/ 业务处理链迁移设计

> Python: `app/chain/`  
> Go: `internal/business/services/` + `internal/business/workflows/`

---

## 1. Python `chain/` 分析

### 1.1 Chain 模式

Python 中的 `chain` 是一种**业务处理链模式**：

- 每个 Chain 负责一个业务领域（下载、订阅、站点、媒体服务器等）
- Chain 之间可以相互调用
- Chain 继承自 `ChainBase`，提供通用能力（事件发布、日志等）

### 1.2 主要 Chain 模块

| Chain 文件 | 职责 | Go 对应 | 状态 |
|-----------|------|---------|------|
| `download.py` | 下载管理（种子解析、下载器调用） | `services/download/` | ✅ 已实现 |
| `subscribe.py` | 订阅管理（订阅刷新、搜索、匹配） | `services/subscribe/` | ✅ 已实现 |
| `site.py` | 站点管理（Cookie 同步、数据统计） | `services/site/` | ✅ 已实现 |
| `transfer.py` | 文件整理（识别、重命名、移动） | `services/transfer/` | ✅ 已实现 |
| `media.py` | 媒体识别（TMDB/豆瓣查询、元数据） | `services/media/` | ✅ 已实现 |
| `mediaserver.py` | 媒体服务器（Emby/Plex/Jellyfin 同步） | `services/mediaserver/` | ✅ 已实现 |
| `search.py` | 搜索（站点搜索、资源匹配） | `services/search/` | ✅ 已实现 |
| `message.py` | 消息通知（Telegram/WeChat/Slack） | `services/notification/` | ✅ 已实现 |
| `tmdb.py` | TMDB API 封装 | `services/media/tmdb/` | ✅ 已实现 |
| `douban.py` | 豆瓣 API 封装 | `services/media/douban/` | ✅ 已实现 |
| `workflow.py` | 工作流执行 | `workflows/` | ✅ 已实现基础框架 |
| `user.py` | 用户管理 | `services/user/` | ✅ 已实现 |
| `plugin.py` | 插件管理 | `services/plugin/` | ✅ 已实现 |
| `storage.py` | 存储管理 | `services/storage/` | ✅ 已实现 |
| `dashboard.py` | 仪表板数据 | `services/dashboard/` | ✅ 已实现 |

---

## 2. Go 设计方案

### 2.1 目录结构

```
internal/business/
├── services/           # 业务服务（对应 Python Chain）
│   ├── auth/           # 认证服务
│   ├── bangumi/        # 番剧服务
│   ├── base/           # 基础服务
│   ├── command/        # 命令服务
│   ├── dashboard/      # 仪表板服务
│   ├── download/       # 下载管理
│   ├── douban/         # 豆瓣服务
│   ├── history/        # 历史记录服务
│   ├── media/          # 媒体识别
│   ├── media/tmdb/     # TMDB API 封装
│   ├── mediaserver/    # 媒体服务器同步
│   ├── message/        # 消息服务
│   ├── notification/   # 通知服务
│   ├── performance/    # 性能监控服务
│   ├── plugin/         # 插件管理
│   ├── pluginmedia/    # 插件媒体服务
│   ├── progress/       # 进度服务
│   ├── recommend/      # 推荐服务
│   ├── rule/           # 规则服务
│   ├── scraper/        # 刮削服务
│   ├── search/         # 搜索服务
│   ├── site/           # 站点管理
│   ├── statistics/     # 统计服务
│   ├── storage/        # 存储管理
│   ├── subscribe/      # 订阅管理
│   ├── system/         # 系统服务
│   ├── systemconfig/   # 系统配置服务
│   ├── tmdb/           # TMDB服务
│   ├── torrents/       # 种子服务
│   ├── transfer/       # 文件整理
│   ├── tvdb/           # TVDB服务
│   ├── user/           # 用户管理
│   ├── webhook/        # Webhook服务
│   └── workflow/       # 工作流服务
├── workflows/          # 工作流编排
│   ├── engine.go       # 工作流引擎
│   ├── workflow.go     # 工作流基础定义
│   └── actions/        # 工作流动作
│       ├── adapters/   # 动作适配器
│       └── params/     # 动作参数
└── domains/            # 领域模型
    ├── events/         # 事件定义
    ├── media/          # 媒体领域模型
    └── module/         # 模块领域模型
```

### 2.2 Service 接口设计

每个 Service 定义清晰的接口，通过依赖注入组合：

**services/download/service.go**：

```go
package download

import (
    "context"

    "moviepilot-go/internal/business/domains"
    "moviepilot-go/internal/repositories"
)

type Service interface {
    // AddDownload 添加下载任务
    AddDownload(ctx context.Context, torrent *domains.Torrent) error

    // ListDownloads 列出下载任务
    ListDownloads(ctx context.Context) ([]*domains.Download, error)

    // RemoveDownload 删除下载任务
    RemoveDownload(ctx context.Context, id string) error

    // PauseDownload 暂停下载
    PauseDownload(ctx context.Context, id string) error

    // ResumeDownload 恢复下载
    ResumeDownload(ctx context.Context, id string) error
}

type service struct {
    downloadRepo repositories.DownloadRepository
    siteService  site.Service  // 依赖其他 service
    logger       *zap.Logger
}

func NewService(
    downloadRepo repositories.DownloadRepository,
    siteService site.Service,
) Service {
    return &service{
        downloadRepo: downloadRepo,
        siteService:  siteService,
        logger:       logger.GetLogger(),
    }
}

func (s *service) AddDownload(ctx context.Context, torrent *domains.Torrent) error {
    s.logger.Info("adding download", zap.String("name", torrent.Name))

    // 业务逻辑
    // ...

    return s.downloadRepo.Create(ctx, torrent)
}
```

---

## 3. Chain 之间的依赖关系

### 3.1 Python 中的依赖

Python Chain 通过直接导入相互调用：

```python
class DownloadChain(ChainBase):
    def download(self, torrent):
        # 调用 SiteChain
        site_info = SiteChain().get_site(torrent.site_id)
        # 调用 MediaChain
        media_info = MediaChain().recognize(torrent.title)
        # ...
```

### 3.2 Go 中的依赖注入

Go 通过接口 + 构造函数注入：

```go
type downloadService struct {
    siteService  site.Service
    mediaService media.Service
    // ...
}

func NewService(
    siteService site.Service,
    mediaService media.Service,
) Service {
    return &downloadService{
        siteService:  siteService,
        mediaService: mediaService,
    }
}

func (s *downloadService) Download(ctx context.Context, torrent *domains.Torrent) error {
    // 调用 siteService
    siteInfo, err := s.siteService.GetSite(ctx, torrent.SiteID)
    if err != nil {
        return err
    }

    // 调用 mediaService
    mediaInfo, err := s.mediaService.Recognize(ctx, torrent.Title)
    if err != nil {
        return err
    }

    // ...
}
```

---

## 4. 工作流编排

复杂业务流程（如订阅刷新、文件整理）使用 Workflow 编排：

**workflows/workflow.go**：

```go
package workflows

import (
    "context"
    "errors"
    "time"

    "go.uber.org/zap"

    "moviepilot-go/internal/infrastructure/events"
    "moviepilot-go/pkg/logger"
)

// Workflow 工作流接口
type Workflow interface {
    // ID 获取工作流ID
    ID() string
    // Name 获取工作流名称
    Name() string
    // Type 获取工作流类型
    Type() string
    // Description 获取工作流描述
    Description() string
    // IsEnabled 工作流是否启用
    IsEnabled() bool
    // Execute 执行工作流
    Execute(ctx context.Context, data map[string]interface{}) error
    // Validate 验证工作流
    Validate() error
}

// BaseWorkflow 工作流基础结构
type BaseWorkflow struct {
    id          string
    name        string
    workflowType  string
    description string
    isEnabled   bool
    triggers    []Trigger
    actions     []Action
    conditions  []Condition
    eventBus    *events.EventBus
    logger      *zap.Logger
}

// Execute 执行工作流
func (w *BaseWorkflow) Execute(ctx context.Context, data map[string]interface{}) error {
    // 检查工作流是否启用
    if !w.isEnabled {
        return errors.New("workflow is disabled")
    }

    // 执行条件检查
    if !w.checkConditions(ctx, data) {
        return nil
    }

    // 执行动作
    for _, action := range w.actions {
        if err := action.Execute(ctx, data); err != nil {
            w.logger.Error("action execution failed", zap.String("action", action.Name()), zap.Error(err))
            // 根据配置决定是否继续执行
            if !action.ContinueOnError() {
                return err
            }
        }
    }

    return nil
}
```

---

## 5. 事件驱动

Chain 之间通过事件解耦：

**发布事件**：

```go
func (s *downloadService) AddDownload(ctx context.Context, torrent *domains.Torrent) error {
    // 添加下载
    err := s.downloadRepo.Create(ctx, torrent)
    if err != nil {
        return err
    }

    // 发布事件
    s.eventBus.Publish(ctx, events.Event{
        Type: events.DownloadAdded,
        Data: torrent,
    })

    return nil
}
```

**订阅事件**：

```go
func (s *transferService) init() {
    // 订阅下载完成事件
    ch, _ := s.eventBus.Subscribe(events.DownloadCompleted)
    go func() {
        for event := range ch {
            s.handleDownloadCompleted(event)
        }
    }()
}

func (s *transferService) handleDownloadCompleted(event events.Event) {
    download := event.Data.(*domains.Download)
    // 自动整理文件
    s.Transfer(context.Background(), download.Path)
}
```

---

## 6. 典型 Service 实现示例

### 6.1 SubscribeService

**services/subscribe/service.go**：

```go
package subscribe

import (
    "context"

    "moviepilot-go/internal/business/domains"
    "moviepilot-go/internal/repositories"
)

type Service interface {
    Create(ctx context.Context, sub *domains.Subscription) error
    Update(ctx context.Context, sub *domains.Subscription) error
    Delete(ctx context.Context, id int64) error
    Get(ctx context.Context, id int64) (*domains.Subscription, error)
    List(ctx context.Context) ([]*domains.Subscription, error)
    ListActive(ctx context.Context) ([]*domains.Subscription, error)
    Refresh(ctx context.Context) error
    Search(ctx context.Context) error
}

type service struct {
    repo          repositories.SubscribeRepository
    searchService search.Service
    downloadService download.Service
    logger        *zap.Logger
}

func NewService(
    repo repositories.SubscribeRepository,
    searchService search.Service,
    downloadService download.Service,
) Service {
    return &service{
        repo:            repo,
        searchService:   searchService,
        downloadService: downloadService,
        logger:          logger.GetLogger(),
    }
}

func (s *service) Refresh(ctx context.Context) error {
    s.logger.Info("refreshing subscriptions")

    subs, err := s.ListActive(ctx)
    if err != nil {
        return err
    }

    for _, sub := range subs {
        if err := s.refreshOne(ctx, sub); err != nil {
            s.logger.Error("failed to refresh subscription",
                zap.Int64("id", sub.ID),
                zap.Error(err))
        }
    }

    return nil
}

func (s *service) refreshOne(ctx context.Context, sub *domains.Subscription) error {
    // 搜索
    results, err := s.searchService.SearchBySub(ctx, sub)
    if err != nil {
        return err
    }

    // 下载
    for _, result := range results {
        if s.shouldDownload(sub, result) {
            s.downloadService.AddDownload(ctx, result.Torrent)
        }
    }

    return nil
}
```

---

## 7. 与 Python 的差异

| 特性 | Python Chain | Go Service |
|------|--------------|------------|
| 组织方式 | 单文件大类 | 多文件小模块 |
| 依赖管理 | 直接导入 | 接口注入 |
| 并发 | 线程池 | goroutine |
| 错误处理 | try/except | error 返回 |
| 事件 | eventmanager | event bus |
| 工作流 | 简单流程 | 完整的工作流引擎 |
| 扩展性 | 有限 | 高度可扩展，支持插件化 |

---

## 8. 迁移进度

| 阶段 | 任务 | 状态 | 完成时间 |
|------|------|------|----------|
| **Phase 1** | 核心 Service（site、media、download、subscribe） | ✅ 已完成 | 已实现 |
| **Phase 2** | Transfer + MediaServer + 其他服务 | ✅ 已完成 | 已实现 |
| **Phase 3** | Workflow 编排基础框架 | ✅ 已完成 | 已实现 |
| **Phase 4** | 事件驱动优化 | ✅ 已实现 | 已实现 |
| **Phase 5** | 工作流动作扩展 | ⏳ 进行中 | 待扩展 |

---

## 9. 迁移计划归纳

### 9.1 迁移目标

将 Python 项目的 `app/chain/` 目录迁移到 Go 项目的 `internal/business/services/` + `internal/business/workflows/` 目录，实现从 Python Chain 模式到 Go Service 模式的迁移，包括：

- 业务逻辑迁移
- 依赖关系重构
- 工作流编排升级
- 事件驱动优化

### 9.2 迁移策略

1. **Service 迁移**：将每个 Python Chain 转换为一个 Go Service，拆分为多个小模块
2. **依赖注入**：使用接口 + 构造函数注入替代直接导入，提高可测试性和扩展性
3. **工作流升级**：从简单的流程调用升级为完整的工作流引擎
4. **事件驱动**：使用事件总线解耦服务之间的依赖关系
5. **模块化设计**：将大型服务拆分为多个小模块，提高代码复用性和可维护性

### 9.3 迁移成果

1. **完整的 Service 层**：已实现所有核心业务服务，包括下载、订阅、站点、媒体服务器等
2. **工作流引擎**：已实现基础的工作流框架，支持条件判断、动作执行等
3. **事件驱动架构**：已实现事件总线，支持服务之间的解耦通信
4. **模块化设计**：服务拆分为多个小模块，提高了代码的可维护性和扩展性
5. **依赖注入**：所有服务通过接口注入依赖，提高了可测试性

### 9.4 后续优化方向

1. **工作流动作扩展**：增加更多工作流动作，支持复杂业务流程
2. **事件驱动优化**：优化事件总线性能，支持事件持久化
3. **服务监控**：增加服务监控，支持性能分析和故障排查
4. **自动化测试**：完善单元测试和集成测试
5. **文档完善**：增加服务接口文档和使用示例

---

## 10. 结论

业务处理链迁移已基本完成，核心功能已实现，系统可以正常运行。迁移后的代码结构更清晰，扩展性更强，可维护性更高。后续将继续优化工作流动作和事件驱动机制，提高系统的灵活性和性能。

---

**相关文档**：
- [db-migration.md](./db-migration.md)
- [api-migration.md](./api-migration.md)
- [scheduler-migration.md](./scheduler-migration.md)
