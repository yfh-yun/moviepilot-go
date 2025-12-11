# MoviePilot `app/core` [app-core] 到 Go 架构的实现映射

> 版本：草案
> 目标：指导从 Python MoviePilot `app/core` 迁移到 Go 项目 `moviepilot-go` 的分层与目录设计。

---

## 1. Python `app/core` 模块概览

Python 端 `app/core` 目录结构：

- `cache.py`：统一缓存抽象（内存 / Redis / 文件），函数级缓存装饰器。
- `config.py`：配置加载与校验（Pydantic + `.env`），业务级配置模型 `ConfigModel` + `Settings`。
- `context.py`：请求/任务上下文管理（trace_id、user_id、workflow_id 等）。
- `event.py`：应用内事件总线（发布/订阅）、事件模型。
- `meta/`：元数据反射、模型元信息工具。
- `metainfo.py`：系统/媒体元信息（版本、依赖、运行环境等）。
- `module.py`：模块加载与管理（内部“模块系统”，不是插件）。
- `plugin.py`：插件核心（插件元数据、加载、启停、hook 调用）。
- `security.py`：安全相关核心（加解密策略、权限/安全配置桥接）。
- `workflow.py`：工作流引擎（节点、执行上下文、状态流转）。

Python 中的 `core` 是一个"粗粒度核心层"，在 Go 中需要拆分进 `pkg/` 与 `internal/` 的多个子系统。

---

## 2. 总体映射关系

| Python `app/core`      | Go 对应位置                                             | 说明 |
|------------------------|---------------------------------------------------------|------|
| `cache.py`             | `pkg/cache/`                                            | 通用缓存库（内存 / Redis / 文件），供业务层与插件使用 |
| `config.py`            | `internal/infrastructure/config/` + `configs/`          | 配置模型、加载、默认值、`.env` / 文件加载与动态更新 |
| `context.py`           | `internal/infrastructure/context/`                      | 请求/任务上下文封装，与 logger、trace 集成 |
| `event.py`             | `internal/infrastructure/events/`                       | 事件总线与事件类型定义 |
| `meta/` + `metainfo.py`| `internal/infrastructure/meta/` + `shared/docs/`        | 元数据反射、版本/构建信息、schema/文档 |
| `module.py`            | `internal/platform/` 或 `internal/integration/`         | 内部“模块”注册与发现，可与插件系统/调度器协作 |
| `plugin.py`            | 已有 `pkg/plugin/` + `python-plugins/`                 | Go 插件核心与 Python gRPC 插件桥接 |
| `security.py`          | `internal/infrastructure/security/` + `pkg/utils/crypto.go` | 安全策略、加解密封装、白名单配置、token 策略 |
| `workflow.py`          | `internal/business/workflows/` + `internal/business/domains/` | 工作流编排与领域工作流模型 |

---

## 3. `cache.py` [缓存核心] → `pkg/cache/`

### 3.1 Python 能力

- 抽象基类：`CacheBackend` / `AsyncCacheBackend`。
- 内存缓存：`MemoryBackend` / `AsyncMemoryBackend`（TTL / LRU，按 region 分区）。
- Redis 缓存：`RedisBackend` / `AsyncRedisBackend`，基于 `RedisHelper` 封装。
- 文件缓存：`FileBackend` / `AsyncFileBackend`，按 `base / region / key` 存储文件。
- 工厂：`Cache()` / `AsyncCache()` / `FileCache()` / `AsyncFileCache()`。
- 函数结果缓存装饰器：`@cached(region, maxsize, ttl, skip_none, skip_empty)`。
- 代理：`CacheProxy` + `TTLCache` / `LRUCache` 兼容 `cachetools` 接口。

### 3.2 Go 设计

**位置**：`pkg/cache/`

**接口（同步）示意：**

```go
package cache

type Backend interface {
    Set(region, key string, value any, ttlSeconds int64) error
    Get(region, key string, dest any) (bool, error)
    Exists(region, key string) (bool, error)
    Delete(region, key string) error
    Clear(region string) error // 为空代表清空所有 region
    Close() error
}
```

**实现建议：**

- `memory.go`：`MemoryBackend`
  - 结构：`map[region]map[key]*entry` + `sync.RWMutex`。
  - TTL：在 entry 中记录 `expireAt time.Time`，Get 时检查是否过期。
  - LRU：可选使用第三方库（如 `golang-lru`），先期也可只实现 TTL。

- `redis.go`：`RedisBackend`
  - 使用 `github.com/redis/go-redis/v9` 或项目已有 Redis 封装。
  - key 规则：`region:{region}:{key}`。

- `file.go`：`FileBackend`
  - 结构：`baseDir / region / key`。
  - 文件名建议使用 key 的 hash，避免非法字符。
  - 写入使用 `os.CreateTemp` + `os.Rename` 保证原子替换。

**工厂（对应 Cache/FileCache/AsyncCache/...）：**

```go
type Type string

const (
    BackendMemory Type = "memory"
    BackendRedis  Type = "redis"
    BackendFile   Type = "file"
)

type Config struct {
    Type        Type
    DefaultTTL  time.Duration
    MaxSize     int
    FileBaseDir string
    // Redis 连接信息：可以从 internal/infrastructure/config 注入
}

func NewBackend(cfg Config) (Backend, error) {
    switch cfg.Type {
    case BackendRedis:
        return NewRedisBackend(...), nil
    case BackendFile:
        return NewFileBackend(cfg.FileBaseDir), nil
    default:
        return NewMemoryBackend(cfg.DefaultTTL, cfg.MaxSize), nil
    }
}
```

**函数级缓存（对应 `@cached`）：**

Go 无装饰器，可通过包装函数实现：

```go
// 无参 -> (T, error) 的示例
func CachedFunc[T any](backend Backend, region string, ttl time.Duration,
    keyBuilder func() string,
    fn func() (T, error),
) (T, error) {
    var zero T
    key := keyBuilder()

    var cached T
    hit, err := backend.Get(region, key, &cached)
    if err == nil && hit {
        return cached, nil
    }

    res, err := fn()
    if err != nil {
        return zero, err
    }

    _ = backend.Set(region, key, res, int64(ttl.Seconds()))
    return res, nil
}
```

业务代码通过传入 `keyBuilder(args...)` 模拟 Python 中根据函数参数自动生成 key 的行为。

---

## 4. `config.py` [配置系统] → `internal/infrastructure/config/`

### 4.1 Python 能力

- `ConfigModel(BaseModel)`：完整的配置 schema + 默认值（应用、安全、数据库、缓存、代理、媒体、站点、订阅、下载、插件、性能、安全、工作流、Docker 等）。
- `Settings(BaseSettings, ConfigModel, LogConfigModel)`：
  - 集成 Pydantic 的 `.env` / 环境变量加载能力。
  - `generic_type_validator` + `generic_type_converter`：根据字段类型自动转换/纠错，并写回 `app.env`。
  - `update_setting(s)`：运行时更新配置项 + 落盘。
  - 派生属性：`CONFIG_PATH/TEMP_PATH/CACHE_PATH/ROOT_PATH/LOG_PATH/COOKIE_PATH/...`。
  - 组合配置：`CONF`（根据 `BIG_MEMORY_MODE` 返回不同缓存/线程池配额）。
  - HTTP 相关：`USER_AGENT/NORMAL_USER_AGENT/PROXY/PROXY_SERVER/GITHUB_HEADERS/...`。

### 4.2 Go 设计

**位置**：`internal/infrastructure/config/`

**文件建议：**

- `model.go`：
  - `AppConfig`（应用级：端口、HOST、TZ、调试开关等）。
  - `DatabaseConfig`（PostgreSQL/SQLite）。
  - `CacheConfig`（`CACHE_BACKEND_TYPE/CACHE_BACKEND_URL/...`）。
  - `MonitoringConfig`（已有：Prometheus、health check）。
  - `SecurityConfig`（允许域名、图片后缀等）。
  - `PerformanceConfig`（`BIG_MEMORY_MODE`、`MEMORY_GC_INTERVAL` 等）。

- `loader.go`：
  - 使用 `viper` 从 `configs/` + 环境变量 + `.env` 读取配置，填充 `AppConfig`。
  - 保持与 Python 默认值尽可能一致。

- `paths.go`：

  利用 `pkg/utils/system.go` 中的 `GetConfigPath/GetEnvPath`：

  ```go
  func ConfigPath() string
  func TempPath() string
  func CachePath() string
  func LogPath() string
  func CookiePath() string
  ```

- `dynamic_update.go`：
  - 暴露 `UpdateSetting(key string, value any)`，内部：
    - 做类型转换/校验（可参考 `generic_type_converter`）。
    - 写回 env 或配置存储（可选）。
    - 更新运行时的全局 config 实例（类似 `settings`）。

**对接：**

- `internal/apis/handlers/system/handler.go` 里的 `/api/system/env` & `/api/system/setting/{key}` 调用 service；
- service 再调用 config 模块的 `UpdateSetting(s)`。

---

## 5. `context.py` [上下文] → `internal/infrastructure/context/`

**目标**：统一在 Go 的 `context.Context` 之上，携带 request_id/user_id/trace_id/workflow_id 等，并与 `pkg/logger` 集成。

**设计要点：**

- `context.go`：
  - 定义内部 key 类型（自定义 struct 避免冲突）。
  - `WithRequestID(ctx, id)` / `RequestIDFrom(ctx)` 等 helper。
  - `WithUserID/WithTraceID/WithWorkflowID` 等。

- 中间件集成：
  - 在 `internal/apis/handlers/middlewares/` 中，HTTP 请求进入时：
    - 生成 request_id（如 UUID）。
    - 注入到 ctx，并在 logger 中设为字段。

---

## 6. `event.py` [事件总线] → `internal/infrastructure/events/`

**目标**：在 Go 中提供应用内事件发布/订阅机制，供业务层、monitor、workflow 使用。

**设计要点：**

- `event.go`：

  ```go
  type Event struct {
      Type      string
      Payload   any
      Timestamp time.Time
      Source    string
  }
  ```

- `bus.go`：

  ```go
  type Bus interface {
      Publish(ctx context.Context, event Event) error
      Subscribe(types ...string) (<-chan Event, func())
  }
  ```

- 简单实现：基于 channel + goroutine 的内存总线；后续可替换为 Kafka/Redis Stream。

- 业务层在关键动作（如订阅更新、下载完成、workflow 结束）发事件，监控与审计模块可订阅。

---

## 7. `meta/` + `metainfo.py` [元数据] → `internal/infrastructure/meta/` + `shared/docs/`

**目标**：

- 通过反射获取结构体字段、tag 等元信息（供文档/调试/动态 UI）。
- 暴露系统版本、构建时间、运行环境等信息。

**设计要点：**

- `internal/infrastructure/meta/info.go`：
  - 从 `pkg/version` 获取版本号/构建信息。
  - 暴露 api：`GetBuildInfo()` / `GetRuntimeInfo()`。

- `internal/infrastructure/meta/reflect.go`：
  - 提供基于 `reflect` 的字段/标签枚举工具。

- 已存在的 `shared/`（proto/schema/docs）继续承载跨服务共享的 schema。

---

## 8. `module.py` [模块系统] → `internal/platform/` 或 `internal/integration/`

**目标**：为内部“模块”（文件管理、订阅、下载等）提供统一生命周期管理，而不是作为插件。

**设计要点：**

- 定义接口：

  ```go
  type Module interface {
      ID() string
      Init(ctx context.Context) error
      Start(ctx context.Context) error
      Stop(ctx context.Context) error
  }
  ```

- 在 `cmd/server/main.go` 中维护模块注册表，按启动顺序初始化。

- 模块实现可以分布在：
  - `internal/platform/filemanager`、`internal/platform/mediaserver` 等；
  - 或 `internal/integration/*`（与第三方系统通信）。

---

## 9. `plugin.py` [插件核心] → 已有 `pkg/plugin/` + `python-plugins/`

Python `plugin.py` 是单进程插件框架；Go 版已经重构为：

- Go 插件核心：`pkg/plugin/`
  - `interface.go`：插件接口定义。
  - `manager.go`：插件管理器。
  - `bridge.go`：Python 插件桥接（gRPC）。
  - `registry.go`：插件注册表。

- Go 原生插件目录：`/plugins` 下分类（downloader/scraper/notifier 等）。

- Python 插件服务：`python-plugins/` 独立进程，通过 `shared/proto/plugin.proto` 与 Go 通信。

迁移重点是保持：

- 插件元数据字段在 proto/json schema 中与 Python 一致；
- 业务层仅依赖 `PluginManager` 接口，而不感知具体实现（Go 原生 vs Python 插件）。

---

## 10. `security.py` [安全核心] → `internal/infrastructure/security/` + `pkg/utils/crypto.go`

**目标**：统一安全策略（允许域名/后缀、token 策略）、基础加解密封装。

**设计要点：**

- `internal/infrastructure/security/config.go`：
  - 读取 `SecurityConfig`（如 `SECURITY_IMAGE_DOMAINS/SECURITY_IMAGE_SUFFIXES`）。
  - 暴露校验函数：`IsAllowedImageDomain(host)`、`IsAllowedImageSuffix(ext)`。

- `internal/infrastructure/security/token.go`：
  - API token/JWT 生成与校验；超时时间配置来自 `Config`。

- `pkg/utils/crypto.go`：
  - AES 加解密、hash 等基础函数（注意 staticcheck 关于 `cipher.NewCFB*` 的警告，后续可改为 AEAD 模式）。

API 层通过 middleware 和 security 模块统一完成认证/授权，避免在 handler 里散落安全逻辑。

---

## 11. `workflow.py` [工作流引擎] → `internal/business/workflows/`

**目标**：

- 在业务层实现工作流编排（订阅、下载、整理等），替代 Python Workflow 引擎。

**设计要点：**

- `internal/business/workflows/types.go`：

  ```go
  type Context struct {
      ID        string
      Input     any
      State     map[string]any
      StartedAt time.Time
      // 其他：trace_id/user_id 等，可复用 context 模块
  }

  type Step interface {
      Name() string
      Run(ctx context.Context, wf *Context) error
  }

  type Workflow struct {
      ID    string
      Steps []Step
  }
  ```

- `engine.go`：
  - 顺序执行/条件分支/错误重试策略。
  - 与 `pkg/logger` 集成，记录关键步骤与错误。

- 与 `internal/schedulers` 集成：调度器触发特定 workflow 执行。

---

## 12. `context` / `event` / `workflow` 与监控子系统的关系

当前 Go 项目监控相关：

- `internal/business/services/performance/monitor.go`：应用 runtime 性能指标（GC/内存/goroutine）。
- `internal/monitor/metrics/collector.go`：系统级监控采集器（CPU/内存/磁盘/主机/Go runtime），基于 gopsutil。

建议：

- `context`：在 ctx 中注入 trace_id/workflow_id。
- `event`：在 workflow 生命周期中发布事件（start/success/fail）。
- `monitor`：订阅事件 + 调用 `collector`，对外提供性能/系统指标 API（已部分在 `internal/apis/handlers/performance` 与 `system` 中实现）。

---

## 13. 迁移优先级建议

结合你的迁移路线规则：

1. **配置与日志优先**（第二阶段核心）：
   - 补齐 `internal/infrastructure/config`，尽量覆盖 `ConfigModel` 的关键项。
   - 确保所有新功能通过 `pkg/logger` 记录日志。

2. **缓存系统**：
   - 在 `pkg/cache` 中实现内存后端（TTL）+ 接口。
   - 后续接入 Redis 与文件后端。

3. **上下文与事件**：
   - 标准化 `context` 与 `events`，方便工作流/监控/审计集成。

4. **工作流**：
   - 在 `internal/business/workflows` 中实现基础工作流引擎，逐步迁移 Python Workflow。

这份文档仅作为总体设计蓝图，具体实现可以按模块分支逐步落地与重构。