# startup/ 启动初始化迁移设计

> Python: `app/startup/`  
> Go: `cmd/server/main.go` + `internal/infrastructure/bootstrap/`

---

## 1. Python `startup/` 模块分析

### 1.1 目录结构

```
app/startup/
├── command_initializer.py      # 命令系统初始化
├── lifecycle.py                # FastAPI lifespan 管理（启动/关闭）
├── modules_initializer.py      # 内部模块初始化
├── monitor_initializer.py      # 文件监控初始化
├── plugins_initializer.py      # 插件系统初始化
├── routers_initializer.py      # API 路由注册
├── scheduler_initializer.py    # 定时任务初始化
└── workflow_initializer.py     # 工作流初始化
```

### 1.2 核心职责

- **lifecycle.py**：FastAPI 的 `@asynccontextmanager` lifespan
  - 启动时：初始化各子系统（模块、插件、调度器、监控、命令、路由、工作流）
  - 关闭时：优雅停止各子系统

- **各 initializer**：
  - 按顺序初始化对应子系统
  - 注册到全局单例或事件管理器
  - 处理初始化失败

---

## 2. Go 设计方案

### 2.1 目录结构

```
moviepilot-go/
├── cmd/server/
│   └── main.go                 # 应用入口 + 启动流程
└── internal/infrastructure/bootstrap/
    ├── bootstrap.go            # 启动协调器
    ├── config.go               # 配置初始化
    ├── database.go             # 数据库初始化
    ├── cache.go                # 缓存初始化
    ├── logger.go               # 日志初始化
    ├── scheduler.go            # 调度器初始化
    ├── monitor.go              # 监控初始化
    ├── plugin.go               # 插件初始化
    ├── module.go               # 模块初始化
    └── router.go               # 路由初始化
```

### 2.2 启动流程设计

**cmd/server/main.go**：

```go
package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"
    "time"

    "moviepilot-go/internal/infrastructure/bootstrap"
    "moviepilot-go/pkg/logger"
)

func main() {
    // 1. 初始化日志（最先）
    if err := bootstrap.InitLogger(); err != nil {
        panic("failed to initialize logger: " + err.Error())
    }
    log := logger.GetLogger()

    // 2. 加载配置
    cfg, err := bootstrap.InitConfig()
    if err != nil {
        log.Fatal("failed to load config", zap.Error(err))
    }

    // 3. 初始化数据库
    db, err := bootstrap.InitDatabase(cfg.Database)
    if err != nil {
        log.Fatal("failed to initialize database", zap.Error(err))
    }
    defer db.Close()

    // 4. 初始化缓存
    cache, err := bootstrap.InitCache(cfg.Cache)
    if err != nil {
        log.Fatal("failed to initialize cache", zap.Error(err))
    }
    defer cache.Close()

    // 5. 初始化模块（内部服务）
    modules, err := bootstrap.InitModules(cfg, db, cache)
    if err != nil {
        log.Fatal("failed to initialize modules", zap.Error(err))
    }

    // 6. 初始化插件
    pluginMgr, err := bootstrap.InitPlugins(cfg, db, cache)
    if err != nil {
        log.Fatal("failed to initialize plugins", zap.Error(err))
    }

    // 7. 初始化调度器
    scheduler, err := bootstrap.InitScheduler(cfg, modules, pluginMgr)
    if err != nil {
        log.Fatal("failed to initialize scheduler", zap.Error(err))
    }
    scheduler.Start()
    defer scheduler.Stop()

    // 8. 初始化文件监控
    monitor, err := bootstrap.InitMonitor(cfg, modules)
    if err != nil {
        log.Fatal("failed to initialize monitor", zap.Error(err))
    }
    monitor.Start()
    defer monitor.Stop()

    // 9. 初始化 HTTP 服务器（Gin）
    router := bootstrap.InitRouter(cfg, modules, pluginMgr)
    srv := &http.Server{
        Addr:    fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
        Handler: router,
    }

    // 10. 启动 HTTP 服务器
    go func() {
        log.Info("starting HTTP server", zap.String("addr", srv.Addr))
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatal("HTTP server error", zap.Error(err))
        }
    }()

    // 11. 优雅关闭
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Info("shutting down server...")
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        log.Error("server forced to shutdown", zap.Error(err))
    }

    log.Info("server exited")
}
```

### 2.3 bootstrap 包设计

**internal/infrastructure/bootstrap/bootstrap.go**：

```go
package bootstrap

import (
    "moviepilot-go/internal/infrastructure/config"
    "moviepilot-go/pkg/cache"
    "moviepilot-go/pkg/database"
    "moviepilot-go/pkg/logger"
)

// App 应用上下文，持有所有已初始化的组件
type App struct {
    Config      *config.Config
    DB          database.Database
    Cache       cache.Backend
    Modules     *ModuleRegistry
    PluginMgr   *plugin.Manager
    Scheduler   *scheduler.Scheduler
    Monitor     *monitor.Monitor
    Router      *gin.Engine
}

// Bootstrap 统一初始化入口
func Bootstrap() (*App, error) {
    app := &App{}

    // 按顺序初始化各组件
    if err := initLogger(); err != nil {
        return nil, err
    }

    if err := initConfig(app); err != nil {
        return nil, err
    }

    if err := initDatabase(app); err != nil {
        return nil, err
    }

    if err := initCache(app); err != nil {
        return nil, err
    }

    if err := initModules(app); err != nil {
        return nil, err
    }

    if err := initPlugins(app); err != nil {
        return nil, err
    }

    if err := initScheduler(app); err != nil {
        return nil, err
    }

    if err := initMonitor(app); err != nil {
        return nil, err
    }

    if err := initRouter(app); err != nil {
        return nil, err
    }

    return app, nil
}

// Shutdown 优雅关闭
func (app *App) Shutdown(ctx context.Context) error {
    // 按相反顺序关闭
    app.Monitor.Stop()
    app.Scheduler.Stop()
    app.PluginMgr.Stop()
    app.Modules.Stop()
    app.Cache.Close()
    app.DB.Close()
    return nil
}
```

---

## 3. 各子系统初始化对应关系

| Python Initializer | Go Bootstrap 函数 | 说明 |
|--------------------|-------------------|------|
| `lifecycle.py` | `bootstrap.Bootstrap()` | 统一启动流程 |
| `modules_initializer.py` | `bootstrap.InitModules()` | 初始化内部模块（文件管理、媒体服务器等） |
| `plugins_initializer.py` | `bootstrap.InitPlugins()` | 加载插件（Go + Python gRPC） |
| `scheduler_initializer.py` | `bootstrap.InitScheduler()` | 启动定时任务 |
| `monitor_initializer.py` | `bootstrap.InitMonitor()` | 启动文件监控 |
| `routers_initializer.py` | `bootstrap.InitRouter()` | 注册 API 路由 |
| `command_initializer.py` | 集成到 `InitPlugins` + 事件系统 | 命令通过事件总线注册 |
| `workflow_initializer.py` | 集成到 `InitModules` | 工作流作为模块初始化 |

---

## 4. 关键设计点

### 4.1 初始化顺序

严格按依赖关系初始化：

```
Logger → Config → Database → Cache → Modules → Plugins → Scheduler → Monitor → Router
```

### 4.2 错误处理

- 任何初始化失败都应 **立即退出**（fail-fast）。
- 记录详细错误日志，包括失败的组件和原因。

### 4.3 优雅关闭

- 监听 `SIGINT` / `SIGTERM` 信号。
- 按**相反顺序**关闭各组件。
- 给每个组件设置关闭超时（如 30 秒）。

### 4.4 健康检查

在 `bootstrap` 完成后，暴露 `/health` 端点：

```go
router.GET("/health", func(c *gin.Context) {
    c.JSON(200, gin.H{
        "status": "healthy",
        "uptime": time.Since(startTime).Seconds(),
    })
})
```

---

## 5. 与 Python 的差异

| 特性 | Python | Go |
|------|--------|-----|
| 启动方式 | FastAPI lifespan (async) | main.go 同步启动 |
| 组件管理 | 全局单例 + 事件管理器 | App 结构体 + 依赖注入 |
| 错误处理 | try/except + 日志 | error 返回 + Fatal |
| 关闭流程 | lifespan yield 后执行 | defer + signal 监听 |

---

## 6. 实现优先级

1. **Phase 1**：基础 bootstrap（logger + config + database）
2. **Phase 2**：缓存 + 模块初始化
3. **Phase 3**：插件 + 调度器
4. **Phase 4**：监控 + 路由

---

## 7. 测试建议

- **单元测试**：每个 `Init*` 函数独立测试。
- **集成测试**：完整 `Bootstrap()` 流程测试。
- **失败场景**：模拟各组件初始化失败，验证错误处理。

---

**相关文档**：
- [scheduler-migration.md](./scheduler-migration.md)
- [monitor-migration.md](./monitor-migration.md)
- [plugins-migration.md](./plugins-migration.md)
