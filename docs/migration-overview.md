# MoviePilot Python → Go 迁移总览

> 版本：草案 v1.0  
> 目标：指导从 Python MoviePilot 到 Go moviepilot-go 的整体架构迁移。

---

## 1. Python 项目核心模块结构

```
MoviePilot-2.8.1-1/app/
├── startup/          # 应用启动初始化器（模块、插件、调度器、路由等）
├── command.py        # 全局命令管理（消息命令、插件命令、内建命令）
├── factory.py        # FastAPI 应用工厂（CORS、中间件、lifespan）
├── log.py            # 日志系统（异步文件写入、格式化、级别控制）
├── monitor.py        # 文件监控（watchdog、目录监控、快照对比）
├── main.py           # 应用入口（uvicorn、信号处理、数据库初始化）
├── scheduler.py      # 定时任务管理（APScheduler、任务注册、执行）
├── schemas/          # Pydantic 数据模型（DTO、事件、类型定义）
├── db/               # 数据库操作层（SQLAlchemy ORM、各表 CRUD）
├── chain/            # 业务处理链（下载、订阅、站点、媒体服务器等）
├── api/              # FastAPI 路由与 endpoints
├── helper/           # 辅助工具（浏览器、Cookie、消息、RSS、站点等）
├── modules/          # 外部服务模块（TMDB、Emby、qBittorrent 等）
├── plugins/          # 插件目录（Python 插件实现）
└── actions/          # 动作处理（未在列表中，可能是工作流动作）
```

---

## 2. Go 项目对应目录映射

| Python 模块/文件 | Go 对应位置 | 文档 |
|------------------|-------------|------|
| `startup/` | `cmd/server/main.go` + `internal/infrastructure/bootstrap/` | [startup-migration.md](./startup-migration.md) |
| `command.py` | `internal/business/services/command/` + `internal/infrastructure/events/` | [command-migration.md](./command-migration.md) |
| `factory.py` | `cmd/server/main.go` + `internal/apis/routes/` | [factory-migration.md](./factory-migration.md) |
| `log.py` | `pkg/logger/` (已实现) | [log-migration.md](./log-migration.md) |
| `monitor.py` | `internal/monitor/filewatch/` + `internal/business/services/transfer/` | [monitor-migration.md](./monitor-migration.md) |
| `main.py` | `cmd/server/main.go` | [main-migration.md](./main-migration.md) |
| `scheduler.py` | `internal/schedulers/` | [scheduler-migration.md](./scheduler-migration.md) |
| `schemas/` | `internal/models/dto/` + `shared/schemas/` | [schemas-migration.md](./schemas-migration.md) |
| `db/` | `internal/repositories/` + `internal/models/` | [db-migration.md](./db-migration.md) |
| `chain/` | `internal/business/services/` + `internal/business/workflows/` | [chain-migration.md](./chain-migration.md) |
| `api/` | `internal/apis/handlers/` + `internal/apis/routes/` | [api-migration.md](./api-migration.md) |
| `helper/` | `pkg/` + `internal/infrastructure/` | [helper-migration.md](./helper-migration.md) |
| `modules/` | `internal/platform/` + `internal/integration/` | [modules-migration.md](./modules-migration.md) |
| `plugins/` | `plugins/` (Go) + `python-plugins/` (gRPC) | [plugins-migration.md](./plugins-migration.md) |

---

## 3. 分层架构对比

### Python 架构（隐式分层）

```
FastAPI App (factory.py)
    ↓
API Endpoints (api/)
    ↓
Chain (业务处理链)
    ↓
Helper + Modules (辅助工具 + 外部服务)
    ↓
DB (SQLAlchemy ORM)
```

### Go 架构（显式分层 - DDD + Clean Architecture）

```
cmd/server/main.go (应用入口)
    ↓
internal/apis/ (API 层 - Gin handlers + routes)
    ↓
internal/business/ (业务层 - services + workflows + domains)
    ↓
internal/infrastructure/ (基础设施层 - config + events + security)
    ↓
internal/repositories/ (数据访问层 - GORM)
    ↓
internal/models/ (数据模型)
    ↓
pkg/ (通用库 - logger + cache + utils + plugin)
```

---

## 4. 关键设计原则

### 4.1 依赖方向（严格单向）

```
apis → business → infrastructure → repositories/models
```

- ❌ 禁止反向依赖（business 不能依赖 apis）
- ✅ 所有跨层调用通过接口 + 依赖注入

### 4.2 日志规范

- 所有模块必须通过 `pkg/logger/` 记录日志
- 禁止使用 `fmt.Println` / `log.Print`
- 日志必须包含上下文（request_id、user_id、trace_id）

### 4.3 错误处理

- 业务层返回领域特定错误
- Infrastructure 层处理技术异常并转换
- API 层统一错误响应格式

### 4.4 配置管理

- 配置从 `internal/infrastructure/config/` 加载
- 支持环境变量、`.env` 文件、YAML 配置
- 运行时配置更新通过 service 层

---

## 5. 迁移优先级

### 第一阶段（Week 1-3）：基础架构

- [x] 项目骨架、Docker 配置
- [x] 日志系统 (`pkg/logger/`)
- [x] 系统工具 (`pkg/utils/system.go`)
- [x] 监控采集器 (`internal/monitor/metrics/`)
- [ ] 配置系统完善 (`internal/infrastructure/config/`)
- [ ] 缓存系统 (`pkg/cache/`)
- [ ] 数据库迁移与 ORM 封装

### 第二阶段（Week 4-8）：核心功能

- [ ] 用户认证与授权
- [ ] 站点管理
- [ ] 订阅系统
- [ ] 下载管理
- [ ] 文件整理（transfer）

### 第三阶段（Week 9-11）：插件与扩展

- [ ] 插件系统重构（Go + Python gRPC）
- [ ] 工作流引擎
- [ ] 消息通知

### 第四阶段（Week 12-15）：优化与部署

- [ ] 性能优化
- [ ] 监控与告警（Prometheus + Grafana）
- [ ] CI/CD 自动化
- [ ] 文档完善

---

## 6. 各模块详细设计文档

请参考以下独立文档了解每个模块的详细迁移方案：

1. [startup/ 启动初始化](./startup-migration.md)
2. [command.py 命令管理](./command-migration.md)
3. [factory.py 应用工厂](./factory-migration.md)
4. [log.py 日志系统](./log-migration.md)
5. [monitor.py 文件监控](./monitor-migration.md)
6. [main.py 应用入口](./main-migration.md)
7. [scheduler.py 定时任务](./scheduler-migration.md)
8. [schemas/ 数据模型](./schemas-migration.md)
9. [db/ 数据库层](./db-migration.md)
10. [chain/ 业务处理链](./chain-migration.md)
11. [api/ API 层](./api-migration.md)
12. [helper/ 辅助工具](./helper-migration.md)
13. [modules/ 外部服务模块](./modules-migration.md)
14. [plugins/ 插件系统](./plugins-migration.md)

---

## 7. 技术栈对比

| 功能 | Python | Go |
|------|--------|-----|
| Web 框架 | FastAPI | Gin |
| ORM | SQLAlchemy | GORM |
| 定时任务 | APScheduler | cron + 自定义调度器 |
| 日志 | logging + 自定义 | zap |
| 配置 | Pydantic + dotenv | viper + 自定义 |
| 缓存 | cachetools + Redis | 自实现 + go-redis |
| 文件监控 | watchdog | fsnotify |
| 消息队列 | 内存 channel | channel + 可选 Redis/Kafka |
| 插件 | Python 动态加载 | Go plugin + Python gRPC |
| 数据校验 | Pydantic | validator + 自定义 |

---

## 8. 注意事项

### 8.1 Python 特性在 Go 中的替代

- **装饰器**：用函数包装或中间件替代
- **async/await**：用 goroutine + channel 替代
- **动态类型**：用接口 + 泛型（Go 1.18+）
- **列表推导式**：用 for 循环 + append
- **多重继承**：用组合（composition）替代

### 8.2 性能优化点

- 使用 goroutine 池避免过度并发
- 合理使用缓存减少数据库查询
- 避免在热路径上使用反射
- 使用 sync.Pool 复用对象

### 8.3 兼容性考虑

- API 接口保持与 Python 版本兼容
- 数据库 schema 保持一致
- 配置文件格式兼容
- 插件接口向后兼容

---

## 9. 测试策略

- **单元测试**：每个 service/repository 必须有测试
- **集成测试**：API 端到端测试
- **性能测试**：关键路径压测
- **兼容性测试**：与 Python 版本对比测试

---

## 10. 文档维护

本文档及各子文档需要随迁移进度持续更新：

- 标记已完成的模块
- 记录遇到的问题与解决方案
- 更新设计决策
- 补充代码示例

---

**最后更新**：2025-11-26  
**维护者**：MoviePilot Go 迁移团队
