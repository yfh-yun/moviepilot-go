# scheduler.py 定时任务迁移设计

> Python: `app/scheduler.py`  
> Go: `internal/schedulers/`

---

## 1. Python `scheduler.py` 分析

### 1.1 核心能力

- 基于 **APScheduler** 的后台调度器
- 支持 **Cron 触发器** 和 **间隔触发器**
- 任务类型：
  - 内建任务（CookieCloud 同步、媒体服务器同步、订阅刷新等）
  - 插件任务（由插件注册）
  - 工作流任务
- 动态任务管理：
  - 添加/删除/暂停/恢复任务
  - 运行时更新任务配置
- 配置变更监听：通过事件系统响应配置更新

### 1.2 典型任务示例

```python
self._jobs = {
    "cookiecloud": {
        "name": "同步CookieCloud站点",
        "func": SiteChain().sync_cookies,
        "trigger": "interval",
        "minutes": settings.COOKIECLOUD_INTERVAL
    },
    "subscribe_refresh": {
        "name": "刷新订阅",
        "func": SubscribeChain().refresh,
        "trigger": "interval",
        "minutes": 30 if settings.SUBSCRIBE_MODE == "rss" else 5
    },
    "mediaserver_sync": {
        "name": "同步媒体服务器",
        "func": MediaServerChain().sync,
        "trigger": "interval",
        "hours": settings.MEDIASERVER_SYNC_INTERVAL
    }
}
```

---

## 2. Go 设计方案

### 2.1 目录结构

```
internal/schedulers/
├── scheduler.go        # 调度器核心
├── job.go              # 任务定义与接口
├── cron.go             # Cron 表达式解析
├── registry.go         # 任务注册表
└── builtin/            # 内建任务
    ├── cookiecloud.go
    ├── subscribe.go
    ├── mediaserver.go
    └ ...
```

### 2.2 核心接口设计

**scheduler.go**：

```go
package schedulers

import (
    "context"
    "time"
)

// Job 任务接口
type Job interface {
    ID() string
    Name() string
    Run(ctx context.Context) error
}

// Trigger 触发器接口
type Trigger interface {
    Next(after time.Time) time.Time
}

// Scheduler 调度器接口
type Scheduler interface {
    Start() error
    Stop() error
    AddJob(job Job, trigger Trigger) error
    RemoveJob(id string) error
    PauseJob(id string) error
    ResumeJob(id string) error
    RunJob(id string) error  // 立即执行
    ListJobs() []JobInfo
}

// JobInfo 任务信息
type JobInfo struct {
    ID          string
    Name        string
    Status      string  // running/paused/stopped
    NextRunTime time.Time
    LastRunTime time.Time
    LastError   string
}
```

### 2.3 实现方案

**方案 A：使用第三方库（推荐）**

使用 `github.com/robfig/cron/v3`：

```go
package schedulers

import (
    "context"
    "sync"

    "github.com/robfig/cron/v3"
    "go.uber.org/zap"

    "moviepilot-go/pkg/logger"
)

type scheduler struct {
    cron   *cron.Cron
    jobs   map[string]*jobEntry
    mu     sync.RWMutex
    logger *zap.Logger
}

type jobEntry struct {
    job     Job
    entryID cron.EntryID
    paused  bool
}

func NewScheduler() Scheduler {
    return &scheduler{
        cron:   cron.New(cron.WithSeconds()),
        jobs:   make(map[string]*jobEntry),
        logger: logger.GetLogger(),
    }
}

func (s *scheduler) Start() error {
    s.cron.Start()
    s.logger.Info("scheduler started")
    return nil
}

func (s *scheduler) Stop() error {
    ctx := s.cron.Stop()
    <-ctx.Done()
    s.logger.Info("scheduler stopped")
    return nil
}

func (s *scheduler) AddJob(job Job, trigger Trigger) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    // 将 Trigger 转换为 cron spec
    spec := triggerToCronSpec(trigger)

    entryID, err := s.cron.AddFunc(spec, func() {
        s.runJob(job)
    })
    if err != nil {
        return err
    }

    s.jobs[job.ID()] = &jobEntry{
        job:     job,
        entryID: entryID,
        paused:  false,
    }

    s.logger.Info("job added",
        zap.String("id", job.ID()),
        zap.String("name", job.Name()),
        zap.String("spec", spec))
    return nil
}

func (s *scheduler) runJob(job Job) {
    ctx := context.Background()
    s.logger.Info("job started", zap.String("id", job.ID()))

    if err := job.Run(ctx); err != nil {
        s.logger.Error("job failed",
            zap.String("id", job.ID()),
            zap.Error(err))
    } else {
        s.logger.Info("job completed", zap.String("id", job.ID()))
    }
}
```

**方案 B：自实现（轻量级）**

如果不想引入第三方库，可以基于 `time.Ticker` + goroutine 自实现：

```go
func (s *scheduler) scheduleJob(job Job, interval time.Duration) {
    ticker := time.NewTicker(interval)
    go func() {
        for {
            select {
            case <-ticker.C:
                s.runJob(job)
            case <-s.stopChan:
                ticker.Stop()
                return
            }
        }
    }()
}
```

---

## 3. 内建任务迁移

### 3.1 任务注册

在 `internal/schedulers/builtin/` 下实现各内建任务：

**cookiecloud.go**：

```go
package builtin

import (
    "context"

    "moviepilot-go/internal/business/services/site"
)

type CookieCloudJob struct {
    siteService site.Service
}

func NewCookieCloudJob(siteService site.Service) *CookieCloudJob {
    return &CookieCloudJob{siteService: siteService}
}

func (j *CookieCloudJob) ID() string {
    return "cookiecloud"
}

func (j *CookieCloudJob) Name() string {
    return "同步CookieCloud站点"
}

func (j *CookieCloudJob) Run(ctx context.Context) error {
    return j.siteService.SyncCookies(ctx)
}
```

### 3.2 任务注册到调度器

在 `bootstrap.InitScheduler()` 中：

```go
func InitScheduler(cfg *config.Config, modules *ModuleRegistry) (*scheduler.Scheduler, error) {
    sched := scheduler.NewScheduler()

    // 注册 CookieCloud 任务
    if cfg.CookieCloud.Enabled {
        job := builtin.NewCookieCloudJob(modules.SiteService)
        trigger := scheduler.IntervalTrigger(time.Duration(cfg.CookieCloud.Interval) * time.Minute)
        sched.AddJob(job, trigger)
    }

    // 注册订阅刷新任务
    if cfg.Subscribe.Enabled {
        job := builtin.NewSubscribeRefreshJob(modules.SubscribeService)
        interval := 30 * time.Minute
        if cfg.Subscribe.Mode == "spider" {
            interval = 5 * time.Minute
        }
        trigger := scheduler.IntervalTrigger(interval)
        sched.AddJob(job, trigger)
    }

    // ... 其他任务

    return sched, nil
}
```

---

## 4. 插件任务注册

插件可以通过接口注册自己的定时任务：

**pkg/plugin/interface.go**：

```go
type Plugin interface {
    // ... 其他方法

    // ScheduledJobs 返回插件需要注册的定时任务
    ScheduledJobs() []ScheduledJob
}

type ScheduledJob struct {
    ID       string
    Name     string
    Spec     string  // Cron 表达式
    Handler  func(ctx context.Context) error
}
```

在插件初始化时，将插件任务注册到调度器：

```go
for _, plugin := range pluginMgr.ListPlugins() {
    for _, job := range plugin.ScheduledJobs() {
        sched.AddJob(
            &pluginJob{id: job.ID, name: job.Name, handler: job.Handler},
            scheduler.CronTrigger(job.Spec),
        )
    }
}
```

---

## 5. 配置变更响应

通过事件系统监听配置变更，动态更新任务：

```go
func (s *scheduler) handleConfigChanged(event *events.Event) {
    // 重新初始化调度器
    s.Stop()
    s.init()
    s.Start()
}
```

---

## 6. API 接口

暴露调度器管理 API：

**internal/apis/handlers/scheduler/handler.go**：

```go
// ListJobs 列出所有任务
func (h *Handler) ListJobs(c *gin.Context) {
    jobs := h.scheduler.ListJobs()
    c.JSON(200, gin.H{"jobs": jobs})
}

// RunJob 立即执行任务
func (h *Handler) RunJob(c *gin.Context) {
    jobID := c.Param("id")
    if err := h.scheduler.RunJob(jobID); err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    c.JSON(200, gin.H{"message": "job started"})
}

// PauseJob 暂停任务
func (h *Handler) PauseJob(c *gin.Context) {
    jobID := c.Param("id")
    if err := h.scheduler.PauseJob(jobID); err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    c.JSON(200, gin.H{"message": "job paused"})
}
```

---

## 7. 与 Python 的差异

| 特性 | Python (APScheduler) | Go (自实现 / cron) |
|------|----------------------|--------------------|
| 触发器 | interval / cron / date | interval / cron |
| 任务执行 | 线程池 | goroutine |
| 持久化 | 支持（可选） | 需自实现 |
| 动态管理 | 内建支持 | 需自实现 |
| 错误处理 | 内建重试 | 需自实现 |

---

## 8. 实现优先级

1. **Phase 1**：基础调度器 + 内建任务（CookieCloud、订阅、媒体服务器）
2. **Phase 2**：插件任务注册
3. **Phase 3**：动态管理 API
4. **Phase 4**：持久化 + 高级特性（重试、超时控制）

---

**相关文档**：
- [startup-migration.md](./startup-migration.md)
- [plugins-migration.md](./plugins-migration.md)
- [chain-migration.md](./chain-migration.md)
