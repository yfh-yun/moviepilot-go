# 其余模块迁移快速索引

> 本文档快速列出剩余模块的迁移方向，详细设计后续补充。

---

## 1. command.py - 命令管理

### Python 职责
- 全局命令注册与分发
- 消息命令解析（如 `/subscribe`）
- 插件命令注册
- 内建命令（帮助、状态、重启等）

### Go 对应
- **位置**：`internal/business/services/command/`
- **设计**：
  ```go
  type CommandService interface {
      RegisterCommand(cmd Command) error
      ExecuteCommand(ctx context.Context, input string) error
      ListCommands() []CommandInfo
  }
  
  type Command interface {
      Name() string
      Description() string
      Execute(ctx context.Context, args []string) error
  }
  ```
- **集成**：通过事件总线监听消息事件，解析命令并执行

---

## 2. factory.py - 应用工厂

### Python 职责
- 创建 FastAPI 实例
- 配置 CORS 中间件
- 设置 lifespan（启动/关闭钩子）

### Go 对应
- **位置**：`cmd/server/main.go` + `internal/apis/routes/router.go`
- **设计**：
  ```go
  func NewRouter(cfg *config.Config) *gin.Engine {
      r := gin.New()
      r.Use(gin.Recovery())
      r.Use(middleware.CORS())
      r.Use(middleware.Logger())
      
      // 注册路由
      v1 := r.Group("/api/v1")
      {
          v1.GET("/health", handlers.Health)
          // ...
      }
      
      return r
  }
  ```

---

## 3. log.py - 日志系统

### Python 职责
- 日志配置（级别、格式、颜色）
- 异步文件写入
- 日志轮转

### Go 对应
- **位置**：`pkg/logger/` ✅ **已实现**
- **技术栈**：zap + lumberjack
- **特性**：
  - 结构化日志
  - 日志轮转
  - 多输出（控制台 + 文件）

---

## 4. monitor.py - 文件监控

### Python 职责
- 监控目录变化（watchdog）
- 文件变更事件处理
- 缓存快照对比

### Go 对应
- **位置**：`internal/monitor/filewatch/`
- **技术栈**：`github.com/fsnotify/fsnotify`
- **设计**：
  ```go
  type FileWatcher interface {
      Watch(path string, handler EventHandler) error
      Stop() error
  }
  
  type EventHandler func(event Event)
  
  type Event struct {
      Path string
      Op   Operation  // Create/Write/Remove/Rename
  }
  ```

---

## 5. main.py - 应用入口

### Python 职责
- 设置进程名
- 信号处理（SIGTERM、SIGINT）
- 启动 uvicorn
- 托盘图标（可选）
- 数据库初始化

### Go 对应
- **位置**：`cmd/server/main.go`
- **设计**：
  ```go
  func main() {
      // 初始化各组件
      app, err := bootstrap.Bootstrap()
      if err != nil {
          log.Fatal(err)
      }
      
      // 启动 HTTP 服务器
      srv := &http.Server{Addr: ":3001", Handler: app.Router}
      go srv.ListenAndServe()
      
      // 优雅关闭
      quit := make(chan os.Signal, 1)
      signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
      <-quit
      
      ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
      defer cancel()
      app.Shutdown(ctx)
  }
  ```

---

## 6. schemas/ - 数据模型

### Python 职责
- Pydantic 数据模型（DTO）
- 请求/响应验证
- 类型定义

### Go 对应
- **位置**：`internal/models/dto/` + `shared/schemas/`
- **技术栈**：struct + `validator` 标签
- **示例**：
  ```go
  type SubscribeCreateRequest struct {
      Name    string `json:"name" binding:"required"`
      Type    string `json:"type" binding:"required,oneof=movie tv"`
      Keyword string `json:"keyword" binding:"required"`
      TMDBID  int64  `json:"tmdb_id"`
  }
  
  type SubscribeResponse struct {
      ID        int64     `json:"id"`
      Name      string    `json:"name"`
      Enabled   bool      `json:"enabled"`
      CreatedAt time.Time `json:"created_at"`
  }
  ```

---

## 7. api/ - API 层

### Python 职责
- FastAPI 路由定义
- Endpoint 实现
- 请求验证
- 响应序列化

### Go 对应
- **位置**：`internal/apis/handlers/` + `internal/apis/routes/`
- **设计**：
  ```
  internal/apis/
  ├── handlers/
  │   ├── user/
  │   │   └── handler.go
  │   ├── subscribe/
  │   │   └── handler.go
  │   └── site/
  │       └── handler.go
  └── routes/
      └── router.go
  ```

---

## 8. helper/ - 辅助工具

### Python 职责
- 浏览器自动化（Playwright）
- Cookie 管理
- CookieCloud 同步
- RSS 解析
- 消息发送
- 站点解析
- 种子解析

### Go 对应

| Python Helper | Go 位置 | 技术栈 |
|---------------|---------|--------|
| `browser.py` | `pkg/browser/` | chromedp |
| `cookie.py` | `internal/infrastructure/cookie/` | net/http/cookiejar |
| `cookiecloud.py` | `internal/business/services/site/cookiecloud.go` | HTTP client |
| `rss.py` | `pkg/rss/` | gofeed |
| `message.py` | `internal/business/services/notification/` | 各平台 SDK |
| `torrent.py` | `pkg/torrent/` | anacrolix/torrent |

---

## 9. modules/ - 外部服务模块

### Python 职责
- TMDB API 封装
- Emby/Plex/Jellyfin API
- qBittorrent/Transmission API
- Telegram/WeChat/Slack API
- 索引器（Jackett、Prowlarr）

### Go 对应
- **位置**：`internal/platform/` + `internal/integration/`
- **设计**：
  ```
  internal/platform/
  ├── tmdb/
  │   ├── client.go
  │   └── types.go
  ├── emby/
  ├── plex/
  └── jellyfin/
  
  internal/integration/
  ├── qbittorrent/
  ├── transmission/
  ├── telegram/
  └── indexer/
  ```

---

## 10. plugins/ - 插件系统

### Python 职责
- 插件加载与管理
- 插件生命周期
- 插件事件处理
- 插件配置管理

### Go 对应
- **位置**：`pkg/plugin/` (核心) + `plugins/` (Go 插件) + `python-plugins/` (Python 插件服务)
- **架构**：
  ```
  Go 主应用 (moviepilot-go)
      ↓ gRPC
  Python 插件服务 (python-plugins)
      ↓
  Python 插件实现
  ```
- **详见**：[plugins-migration.md](./plugins-migration.md)（待创建）

---

## 11. actions/ - 动作处理

### Python 职责
- 工作流动作定义
- 动作执行器

### Go 对应
- **位置**：`internal/business/workflows/actions/`
- **设计**：
  ```go
  type Action interface {
      Type() string
      Execute(ctx context.Context, input interface{}) (interface{}, error)
  }
  
  type DownloadAction struct{}
  func (a *DownloadAction) Execute(ctx context.Context, input interface{}) (interface{}, error) {
      // 下载逻辑
  }
  ```

---

## 12. 迁移优先级总结

### 高优先级（Week 1-4）
1. ✅ log.py → `pkg/logger/`
2. factory.py → `cmd/server/main.go` + `internal/apis/routes/`
3. main.py → `cmd/server/main.go`
4. db/ → `internal/repositories/` + `internal/models/`
5. schemas/ → `internal/models/dto/`

### 中优先级（Week 5-8）
6. chain/ → `internal/business/services/`
7. api/ → `internal/apis/handlers/`
8. scheduler.py → `internal/schedulers/`
9. command.py → `internal/business/services/command/`

### 低优先级（Week 9-12）
10. monitor.py → `internal/monitor/filewatch/`
11. helper/ → `pkg/` + `internal/infrastructure/`
12. modules/ → `internal/platform/` + `internal/integration/`
13. plugins/ → `pkg/plugin/` + `python-plugins/`
14. actions/ → `internal/business/workflows/actions/`

---

## 13. 后续任务

为以下模块创建详细设计文档：

- [ ] `api-migration.md`
- [ ] `schemas-migration.md`
- [ ] `helper-migration.md`
- [ ] `modules-migration.md`
- [ ] `plugins-migration.md`
- [ ] `command-migration.md`
- [ ] `factory-migration.md`
- [ ] `log-migration.md`（已有实现，补充文档）
- [ ] `monitor-migration.md`
- [ ] `main-migration.md`

---

**相关文档**：
- [migration-overview.md](./migration-overview.md)
- [startup-migration.md](./startup-migration.md)
- [scheduler-migration.md](./scheduler-migration.md)
- [chain-migration.md](./chain-migration.md)
- [db-migration.md](./db-migration.md)
