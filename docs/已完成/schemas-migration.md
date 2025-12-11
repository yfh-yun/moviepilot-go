# app/schemas → Go DTO 映射文档

> Python: `app/schemas/*.py`  
> Go: `internal/models/` + `internal/models/dto/` + `internal/business/domains/`

本文件按功能模块梳理 `app/schemas` 中的 Pydantic 模型，给出：

- 模型用途
- 推荐的 Go 结构体落点（DTO / Model / Domain）
- 关键枚举 / 字段注意事项

> 说明：本表是设计指导，不是代码生成器；实现时请结合具体业务细节与已有 Go 代码。

---

## 1. 基础枚举与通用类型（types.py）

**Python 文件**：`app/schemas/types.py`  
**用途**：集中定义全局枚举与常用类型别名，例如：
- `MediaType`：媒体类型（电影/电视剧/番剧等）
- `EventType`：系统事件类型（下载完成、订阅刷新、工作流执行等）
- `ModuleType`, `DownloaderType`, `MediaServerType`, `MessageChannel`, `StorageSchema`, `OtherModulesType`
- 其他用于站点、插件、任务状态的枚举

**Go 建议落点**：
- 目录：`internal/models/enums/` 或 `internal/business/domains/*/enums.go`，视业务域拆分
- 示例：
  - `internal/models/enums/media_type.go` → `MediaType`
  - `internal/models/enums/event_type.go` → `EventType`
  - `internal/models/enums/module_type.go` → `ModuleType`, `DownloaderType` 等

**注意事项**：
- Go 枚举建议用 `type Xxx string` + `const`，保持与 Python `.value` 一致，方便序列化与 DB 存储。
- `EventType` 与事件总线 `internal/infrastructure/events` 要一一对应。
- `MediaType` 要与 `internal/business/domains/media` / `internal/models/dto/media` 中字段保持统一。

---

## 2. 认证与用户（token.py, user.py, system.py）

### 2.1 token.py

**用途**：
- `Token`：登录返回的 Access Token（通常包含 `access_token`, `token_type`）
- `TokenPayload`：JWT 中的载荷（`sub`, `username`, `super_user`, `level`, `purpose`, `exp`, `iat`）

**Go 落点建议**：
- `internal/models/dto/auth/token.go`
  - `LoginResponse` / `TokenResponse`（对应 `Token`）
- `internal/infrastructure/security/jwt.go`
  - `TokenClaims` 结构体（对应 `TokenPayload`，已在 `security-migration-plan.md` 中设计）

**注意事项**：
- `purpose` 字段要与 Go 侧 `TokenPurpose` 枚举一致（`authentication` / `resource`）。
- `exp`, `iat` 在 Go 中可使用 `jwt.RegisteredClaims` 自带字段。

### 2.2 user.py

**用途**：
- 用户 DTO：
  - `User`, `UserCreate`, `UserUpdate`, `UserInDB` 等 Pydantic 模型
- API 请求/响应：
  - 登录时的 `UserLogin` 请求体
  - 用户列表项 DTO

**Go 落点建议**：
- `internal/models/user.go`
  - DB 模型：`UserModel`（对应表结构）
- `internal/models/dto/user/user_dto.go`
  - `UserDTO`, `UserCreateRequest`, `UserUpdateRequest`, `UserLoginRequest`, `UserResponse`
- `internal/business/domains/user/`（可选）
  - 若有复杂领域逻辑，可定义 `User` 领域对象

**注意事项**：
- 密码字段应只出现在：
  - 请求 DTO（如 `password`）
  - 数据库模型中的 hash 字段（如 `PasswordHash`），禁止出现在普通响应 DTO。
- 权限/角色字段与 `security` 模块的 `Level`、`super_user` 语义要对应好。

### 2.3 system.py

**用途**：
- 系统状态与配置视图：
  - 系统 info、版本信息、运行模式（DEV/DEBUG）、资源限制、健康检查结果等

**Go 落点建议**：
- `internal/models/dto/system/system_dto.go`
  - `SystemInfoResponse`, `HealthStatusResponse`, `ConfigSummaryResponse` 等
- 实现由：`internal/apis/handlers/system` + `internal/business/services/system` 提供

**注意事项**：
- 注意与 `config-migration-plan.md` 的配置结构相互映射， DTO 中通常只暴露只读/安全字段。

---

## 3. 站点 / 订阅 / 下载 / 整理链路

### 3.1 site.py

**用途**：
- 站点配置模型：域名、cookie、UA、代理、分类映射等
- 站点统计信息：可用性、抓取数量、错误次数

**Go 落点建议**：
- `internal/models/site.go`
  - DB 模型：`SiteModel`
- `internal/models/dto/site/site_dto.go`
  - `SiteDTO`, `SiteCreateRequest`, `SiteUpdateRequest`, `SiteStatDTO`
- `internal/business/services/site/`
  - 站点业务逻辑：Cookie 同步、状态刷新、统计等

**注意事项**：
- 枚举影响：某些字段可能使用 `SiteType`, `SiteCategory` 等枚举（见 `types.py`）。

### 3.2 subscribe.py

**用途**：
- 订阅定义：
  - 名称、关键字、媒体类型、状态、模式（RSS/爬虫）、刷新间隔等
- 请求/响应 DTO：创建、更新、列表

**Go 落点建议**：
- `internal/models/subscribe.go`：`SubscribeModel`
- `internal/models/dto/subscribe/subscribe_dto.go`
  - `SubscribeDTO`, `SubscribeCreateRequest`, `SubscribeUpdateRequest`, `SubscribeListItem`
- `internal/business/services/subscribe/`
  - 订阅刷新与搜索逻辑

**注意事项**：
- 媒体类型字段使用 `MediaType` 枚举。
- 与 `chain.subscribe` 和 `workflows` 中订阅相关工作流保持字段兼容。

### 3.3 download.py

**用途**：
- 下载任务 DTO：
  - 任务 ID、状态、进度、种子信息、下载器信息等
- 用户发起交互下载请求/批量下载请求的请求体

**Go 落点建议**：
- `internal/models/download.go`：DB 模型（也可与 `downloadhistory` 区分）
- `internal/models/dto/download/download_dto.go`
  - `DownloadTaskDTO`, `DownloadRequest`, `BatchDownloadRequest`
- `internal/business/services/download/`
  - 调用 downloader 模块的业务服务

**注意事项**：
- 部分字段与 `TorrentInfo`（见 `core/context.py` & `context.py`）重合，可统一抽象。

### 3.4 transfer.py

**用途**：
- 整理任务/结果 DTO（文件路径、目标路径、重命名格式、冲突信息等）

**Go 落点建议**：
- `internal/models/transfer_history.go`
- `internal/models/dto/transfer/transfer_dto.go`
- `internal/business/services/transfer/`

**注意事项**：
- 与工作流动作（如“整理文件” Action）以及链路中的 `Context` 统一字段语义。

### 3.5 history.py

**用途**：
- 各种历史记录视图：下载、整理、订阅执行记录等

**Go 落点建议**：
- DB 模型已经在 `db-migration.md` 提到（`DownloadHistory`, `TransferHistory` 等）
- DTO：`internal/models/dto/history/history_dto.go`
  - `DownloadHistoryDTO`, `TransferHistoryDTO` 等

---

## 4. 媒体 & 媒体服务器（mediaserver.py, tmdb.py, context.py）

### 4.1 mediaserver.py

**用途**：
- Emby/Plex/Jellyfin 等媒体服务器配置 DTO
- 同步状态/媒体库信息 DTO

**Go 落点建议**：
- `internal/models/mediaserver.go`
- `internal/models/dto/mediaserver/mediaserver_dto.go`
- `internal/business/services/mediaserver/`
  - 对应 `app/chain/mediaserver.py` 和 `modules/emby|plex|jellyfin` 的逻辑

### 4.2 tmdb.py

**用途**：
- TMDB 相关输出模型：搜索结果精简版、详情页 DTO

**Go 落点建议**：
- `internal/models/dto/media/tmdb_dto.go`
  - `TMDBSearchResult`, `TMDBDetailResponse` 等
- 与 `internal/platform/tmdb` client + `MediaInfo` Domain 结合使用

### 4.3 context.py

**用途**：
- 与 `core/context.py` 的 Domain 版本对应，是用于 API 返回的上下文 DTO：
  - `TorrentInfoDTO`, `MediaInfoDTO`, `ContextDTO` 等

**Go 落点建议**：
- `internal/models/dto/media/context_dto.go`
  - 与 `docs/context-migration-plan.md` 中的设计保持一致
- Domain 在：`internal/business/domains/media/`

**注意事项**：
- DTO 中使用 `string`/简单类型，不直接暴露 Go 内部使用的复杂字段（例如 `time.Time` → RFC3339 字符串）。

---

## 5. 消息 / 监控 / 仪表盘（message.py, monitoring.py, dashboard.py）

### 5.1 message.py

**用途**：
- 消息发送请求和返回结构：
  - 目标渠道（Telegram/Slack/WeChat/Email 等）
  - 消息内容、标题、优先级

**Go 落点建议**：
- `internal/models/dto/message/message_dto.go`
- `internal/business/services/notification/`
  - 将 DTO 映射到内部 Domain 或直接传递给模块

### 5.2 monitoring.py

**用途**：
- 系统监控指标 DTO（CPU、内存、磁盘、网络、进程等），用于 `/metrics` 或 Dashboard

**Go 落点建议**：
- 已在 `internal/monitor/metrics` & `system handler` 里开始实现
- DTO 可放在：`internal/models/dto/system/metrics_dto.go`

### 5.3 dashboard.py

**用途**：
- 仪表盘布局/组件/属性的结构化描述
- 插件仪表盘也依赖这些模型

**Go 落点建议**：
- `internal/models/dto/dashboard/dashboard_dto.go`
- `internal/business/services/dashboard/` 或 `plugin` 相关 service 使用

---

## 6. 插件 / 工作流 / 规则（plugin.py, workflow.py, rule.py）

### 6.1 plugin.py

**用途**：
- 插件元数据 DTO：
  - `id`, `plugin_name`, `plugin_desc`, `plugin_version`, `plugin_icon`, `repo_url`, `installed`, `state`, `has_update`, `labels`, `auth_level` 等
- 插件仪表盘：
  - `PluginDashboard`（布局、元素、渲染模式等）

**Go 落点建议**：
- `internal/models/dto/plugin/plugin_dto.go`
- 与 `pkg/plugin` 模块整合，参见 `plugin-migration-plan.md`

### 6.2 workflow.py

**用途**：
- 工作流定义 DTO：
  - `Workflow`, `Action`, `ActionContext` 等
- 供 API 创建/更新/查看工作流，以及工作流执行结果展示

**Go 落点建议**：
- `internal/models/workflow.go`：DB 模型（在 `db-migration.md` 中已有规划）
- `internal/models/dto/workflow/workflow_dto.go`
  - `WorkflowDTO`, `WorkflowCreateRequest`, `WorkflowUpdateRequest`, `ActionDTO`, `ActionContextDTO`
- 与 `internal/business/workflows` 和 `internal/business/domains/actions` 结合，参见 `workflow-migration-plan.md`

### 6.3 rule.py

**用途**：
- 各种规则定义（过滤规则、下载规则、整理规则等）的 Pydantic 模型

**Go 落点建议**：
- 若规则持久化：`internal/models/rule.go` + `internal/repositories/rule_repository.go`
- DTO：`internal/models/dto/rule/rule_dto.go`
- Domain：`internal/business/domains/rule/` 提供规则匹配逻辑

**注意事项**：
- 与工作流中的条件匹配（`event_conditions`）使用一致的操作符语义（equals/not_equals/contains/...）。

---

## 7. 外部系统 / 集成（servarr.py, servcookie.py, file.py）

### 7.1 servarr.py

**用途**：
- 与 Sonarr/Radarr 等 *arr 系统的交互 DTO（配置、任务、状态）

**Go 落点建议**：
- `internal/models/dto/servarr/servarr_dto.go`
- integration client：`internal/integration/servarr/`

### 7.2 servcookie.py

**用途**：
- CookieCloud / 站点 Cookie 同步相关 DTO

**Go 落点建议**：
- `internal/models/dto/cookie/cookie_dto.go`
- `internal/business/services/site/cookie_service.go`

### 7.3 file.py

**用途**：
- 文件/目录 DTO（路径、大小、类型、mtime 等），用于浏览/选择本地或远程文件

**Go 落点建议**：
- `internal/models/dto/file/file_dto.go`
- 部分可与已有 `pkg/utils/system.go` 文件操作结果对接

---

## 8. 异常与统一响应（exception.py, response.py）

### 8.1 exception.py

**用途**：
- 定义统一错误结构，用于 FastAPI 返回 JSON 错误

**Go 落点建议**：
- `internal/models/dto/common/error_dto.go`
  - `ErrorResponse`（code/message/detail）
- 对应：`internal/apis/middlewares/recover` / 统一错误处理器

### 8.2 response.py

**用途**：
- 通用响应包装，如：
  - 分页结构：`ListResponse`, `PageResponse`
  - 统一 `success/data/error` 类型响应

**Go 落点建议**：
- `internal/models/dto/common/response_dto.go`
  - `PagedResponse[T]`, `SimpleResponse[T]`（使用 Go 泛型）

---

## 9. 实施建议

1. **Step 1：建立 DTO 目录结构**
   - 在 `internal/models/dto/` 下按业务域拆分子目录：`user/`, `site/`, `subscribe/`, `download/`, `media/`, `plugin/`, `workflow/`, `common/` 等。

2. **Step 2：优先迁移核心 DTO**
   - 与当前已设计/实现的模块对齐：
     - `config`, `security`, `context`, `workflow`, `plugin`, `site`, `subscribe`, `download`, `user`。

3. **Step 3：统一枚举与错误响应**
   - 先完成 `types.py` → Go 枚举的迁移，确保各层使用统一类型。
   - 统一错误/响应 DTO，方便中间件和 Handler 复用。

4. **Step 4：逐步补全长尾 schema**
   - 监控、消息、external integrations（*arr、CookieCloud）等可以按业务优先级逐步映射。

---

## 10. 总结

- `app/schemas` 集中了承上启下的 **Pydantic DTO**：既为 API/前端服务，又与领域模型/数据库模型紧密相关。  
- 在 Go 中建议：
  - 枚举拆到 `internal/models/enums` 或各域 `enums.go`；
  - DTO 统一放在 `internal/models/dto/<domain>/`；
  - DB 模型放在 `internal/models/`，领域对象在 `internal/business/domains/`；
  - 通过清晰的 DTO <-> Domain <-> Model 映射，保证类型一致性和演进空间。

此文档可作为你实现 Go DTO 的蓝图，后续若某一子域（如 `subscribe` 或 `plugin`）进入具体开发阶段，可以再根据实际字段细化对应 DTO 结构。
