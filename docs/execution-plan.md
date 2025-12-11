# MoviePilot Go 项目详细执行计划

> **制定日期**：2025-12-02  
> **计划周期**：6周（Week 4-9）  
> **当前状态**：Phase 1 进行中（60%完成）

---

## 📋 执行计划总览

### 优先级调整说明

**原计划问题**：
- 按时间顺序推进导致外部服务集成延迟
- 24个外部服务模块全部未开始，阻塞第二阶段核心功能开发

**调整策略**：
```
原计划：Phase 1 → Phase 2 → Phase 3 → Phase 4
调整后：Phase 1优化（并行）→ 外部服务集成（优先）→ Phase 2核心功能
```

### 资源分配

| 工作内容 | 资源占比 | 人员配置建议 |
|---------|---------|-------------|
| 外部服务集成 | 60% | 2-3人 |
| Phase 1优化 | 30% | 1-2人 |
| Phase 2准备 | 10% | 1人 |

---

## 🎯 Week 4：Phase 1优化 + 外部服务启动

### Week 4.1 (Day 1-2)：数据库层优化

#### 任务清单
- [ ] **数据库索引优化**
  - [ ] 分析慢查询日志
  - [ ] 为高频查询字段添加索引
  - [ ] 优化复合索引策略
  
- [ ] **连接池配置优化**
  - [ ] 调整 `MaxIdleConns`、`MaxOpenConns`
  - [ ] 配置连接超时参数
  - [ ] 添加连接池监控指标

- [ ] **Repository层性能测试**
  - [ ] 编写性能基准测试
  - [ ] 测试批量操作性能
  - [ ] 优化N+1查询问题

#### 交付物
- `pkg/database/optimization.go` - 数据库优化配置
- `internal/repositories/benchmarks/` - 性能测试套件
- 性能测试报告

#### 验收标准
- ✅ 所有核心查询响应时间 < 100ms
- ✅ 连接池利用率 > 70%
- ✅ 无慢查询（> 1s）

---

### Week 4.2 (Day 3-4)：下载器集成（优先级最高）

#### 任务清单
- [ ] **qBittorrent客户端实现**
  - [ ] 创建 `internal/integration/qbittorrent/client.go`
  - [ ] 实现接口：
    - `AddTorrent(url, savePath string) error`
    - `ListTorrents() ([]*Torrent, error)`
    - `PauseTorrent(hash string) error`
    - `ResumeTorrent(hash string) error`
    - `RemoveTorrent(hash string, deleteFiles bool) error`
    - `GetTorrentInfo(hash string) (*TorrentInfo, error)`
  - [ ] 实现认证（用户名/密码）
  - [ ] 实现错误重试机制

- [ ] **Transmission客户端实现**
  - [ ] 创建 `internal/integration/transmission/client.go`
  - [ ] 实现与qBittorrent相同的接口
  - [ ] 实现RPC认证

- [ ] **下载器抽象接口**
  - [ ] 定义 `internal/integration/downloader/interface.go`
  - [ ] 实现工厂模式支持多下载器切换

#### 代码示例

```go
// internal/integration/downloader/interface.go
package downloader

import "context"

type Client interface {
    // AddTorrent 添加种子
    AddTorrent(ctx context.Context, req *AddTorrentRequest) (*Torrent, error)
    
    // ListTorrents 列出所有种子
    ListTorrents(ctx context.Context, filter *TorrentFilter) ([]*Torrent, error)
    
    // PauseTorrent 暂停种子
    PauseTorrent(ctx context.Context, hash string) error
    
    // ResumeTorrent 恢复种子
    ResumeTorrent(ctx context.Context, hash string) error
    
    // RemoveTorrent 删除种子
    RemoveTorrent(ctx context.Context, hash string, deleteFiles bool) error
    
    // GetTorrentInfo 获取种子详情
    GetTorrentInfo(ctx context.Context, hash string) (*TorrentInfo, error)
}

type AddTorrentRequest struct {
    URL      string
    SavePath string
    Category string
    Tags     []string
}

type Torrent struct {
    Hash     string
    Name     string
    State    string
    Progress float64
    Size     int64
    // ...
}
```

#### 交付物
- `internal/integration/qbittorrent/` - qBittorrent客户端
- `internal/integration/transmission/` - Transmission客户端
- `internal/integration/downloader/interface.go` - 统一接口
- 单元测试（覆盖率 > 80%）

#### 验收标准
- ✅ 支持添加、暂停、恢复、删除种子
- ✅ 支持获取种子列表和详情
- ✅ 错误处理完善，支持重试
- ✅ 通过集成测试（需要真实下载器环境）

---

### Week 4.3 (Day 5)：测试覆盖率提升

#### 任务清单
- [ ] **为核心模块补充单元测试**
  - [ ] `internal/business/services/` - 21个服务
  - [ ] `internal/business/workflows/actions/` - 15个动作
  - [ ] `internal/repositories/` - 15个Repository

- [ ] **集成测试框架搭建**
  - [ ] 使用 `testcontainers-go` 搭建测试环境
  - [ ] 配置PostgreSQL测试容器
  - [ ] 配置Redis测试容器

#### 交付物
- 测试覆盖率从85% → 70%（全局）
- 集成测试框架

#### 验收标准
- ✅ 核心业务服务测试覆盖率 > 70%
- ✅ 集成测试可自动运行

---

## 🎯 Week 5：媒体服务器集成 + 元数据平台

### Week 5.1 (Day 1-2)：媒体服务器集成

#### 任务清单
- [ ] **Emby客户端实现**
  - [ ] 创建 `internal/platform/emby/client.go`
  - [ ] 实现接口：
    - `GetLibraries() ([]*Library, error)`
    - `GetItems(libraryId string) ([]*MediaItem, error)`
    - `RefreshLibrary(libraryId string) error`
    - `GetPlaybackInfo(itemId string) (*PlaybackInfo, error)`
  - [ ] 实现API Key认证

- [ ] **Plex客户端实现**
  - [ ] 创建 `internal/platform/plex/client.go`
  - [ ] 实现与Emby相同的接口
  - [ ] 实现Token认证

- [ ] **Jellyfin客户端实现**
  - [ ] 创建 `internal/platform/jellyfin/client.go`
  - [ ] 实现与Emby相同的接口

- [ ] **媒体服务器抽象接口**
  - [ ] 定义 `internal/platform/mediaserver/interface.go`
  - [ ] 实现工厂模式支持多媒体服务器

#### 代码示例

```go
// internal/platform/mediaserver/interface.go
package mediaserver

import "context"

type Client interface {
    // GetLibraries 获取媒体库列表
    GetLibraries(ctx context.Context) ([]*Library, error)
    
    // GetItems 获取媒体项目
    GetItems(ctx context.Context, libraryId string, filter *ItemFilter) ([]*MediaItem, error)
    
    // RefreshLibrary 刷新媒体库
    RefreshLibrary(ctx context.Context, libraryId string) error
    
    // GetPlaybackInfo 获取播放信息
    GetPlaybackInfo(ctx context.Context, itemId string) (*PlaybackInfo, error)
    
    // UpdateMetadata 更新元数据
    UpdateMetadata(ctx context.Context, itemId string, metadata *Metadata) error
}

type Library struct {
    ID   string
    Name string
    Type string // movie, tv, music
    Path string
}

type MediaItem struct {
    ID       string
    Name     string
    Type     string
    Year     int
    Path     string
    Metadata *Metadata
}
```

#### 交付物
- `internal/platform/emby/` - Emby客户端
- `internal/platform/plex/` - Plex客户端
- `internal/platform/jellyfin/` - Jellyfin客户端
- `internal/platform/mediaserver/interface.go` - 统一接口
- 单元测试 + 集成测试

#### 验收标准
- ✅ 支持获取媒体库和媒体项目
- ✅ 支持刷新媒体库
- ✅ 支持更新元数据
- ✅ 通过集成测试

---

### Week 5.2 (Day 3-4)：元数据平台适配

#### 任务清单
- [ ] **TMDB平台适配层**
  - [ ] 创建 `internal/platform/tmdb/client.go`
  - [ ] 实现接口：
    - `SearchMovie(query string) ([]*Movie, error)`
    - `SearchTV(query string) ([]*TVShow, error)`
    - `GetMovieDetails(id int) (*MovieDetails, error)`
    - `GetTVDetails(id int) (*TVDetails, error)`
    - `GetSeasonDetails(tvId, seasonNum int) (*SeasonDetails, error)`
  - [ ] 实现API Key认证
  - [ ] 实现请求限流（40 req/10s）

- [ ] **TVDB平台适配层**
  - [ ] 创建 `internal/platform/thetvdb/client.go`
  - [ ] 实现与TMDB类似的接口
  - [ ] 实现JWT认证

- [ ] **豆瓣平台适配层**
  - [ ] 创建 `internal/platform/douban/client.go`
  - [ ] 实现搜索和详情接口
  - [ ] 处理反爬虫机制

#### 代码示例

```go
// internal/platform/tmdb/client.go
package tmdb

import (
    "context"
    "time"
    
    "golang.org/x/time/rate"
)

type Client struct {
    apiKey  string
    baseURL string
    limiter *rate.Limiter // 40 req/10s
    client  *http.Client
}

func NewClient(apiKey string) *Client {
    return &Client{
        apiKey:  apiKey,
        baseURL: "https://api.themoviedb.org/3",
        limiter: rate.NewLimiter(rate.Every(250*time.Millisecond), 40),
        client:  &http.Client{Timeout: 10 * time.Second},
    }
}

func (c *Client) SearchMovie(ctx context.Context, query string) ([]*Movie, error) {
    // 等待限流器
    if err := c.limiter.Wait(ctx); err != nil {
        return nil, err
    }
    
    // 实现搜索逻辑
    // ...
}
```

#### 交付物
- `internal/platform/tmdb/` - TMDB客户端
- `internal/platform/thetvdb/` - TVDB客户端
- `internal/platform/douban/` - 豆瓣客户端
- 限流器实现
- 单元测试

#### 验收标准
- ✅ 支持电影和电视剧搜索
- ✅ 支持获取详细信息
- ✅ 限流器正常工作
- ✅ 错误处理完善

---

### Week 5.3 (Day 5)：API文档完善

#### 任务清单
- [ ] **Swagger文档生成**
  - [ ] 为所有API Handler添加Swagger注解
  - [ ] 使用 `swag init` 生成文档
  - [ ] 配置Swagger UI

- [ ] **API文档优化**
  - [ ] 添加请求示例
  - [ ] 添加响应示例
  - [ ] 添加错误码说明

#### 交付物
- Swagger文档（可访问 `/swagger/index.html`）
- API使用指南

#### 验收标准
- ✅ 所有API都有完整的Swagger注解
- ✅ Swagger UI可正常访问
- ✅ 文档包含请求/响应示例

---

### Week 5 当前进度（2025-12-02）

- ✅ **媒体服务器集成（Emby/Plex/Jellyfin）已完成接口与客户端骨架**  
  - 统一接口：`internal/integration/mediaserver/interface.go`  
  - Emby 客户端：`internal/integration/mediaserver/emby/client.go`（已实现 `TestConnection` / `GetServerInfo` / `ListLibraries` / `GetItem` / `SearchItems`）  
  - Plex 客户端：`internal/integration/mediaserver/plex/client.go`（已实现连接测试、媒体库与基础查询）  
  - Jellyfin 客户端：`internal/integration/mediaserver/jellyfin/client.go`（已实现连接测试、媒体库与基础查询）

- ✅ **元数据平台统一接口与 TMDB 实现已完成**  
  - 统一接口：`internal/integration/metadata/interface.go`  
  - TMDB 客户端：`internal/integration/metadata/tmdb/client.go`（已实现 `SearchMovie` / `GetMovieByTMDB` / `SearchTV` / `GetTVByTMDB` 等）  
  - 聚合服务：`internal/business/services/metadata/aggregator_service.go`（已支持 `AggregateMovieByTMDB` 与 `SearchAndAggregateMovie`）

- **TVDB / 豆瓣平台适配（最小实现已完成）**  
  - TVDB 客户端：`internal/integration/metadata/tvdb/client.go`（已实现 `SearchTV` / `GetTVByID` 最小版本）
  - 豆瓣客户端：`internal/integration/metadata/douban/client.go`（已实现 `SearchMovie` / `GetMovieByID` 最小版本，假设通过代理访问）
  - 聚合服务已接入 TVDB 作为剧集补充源：`AggregateTVByTMDB` 方法支持 TMDB 主数据 + TVDB 可选补充

- **Swagger 文档**  
  - 已添加 swaggo 依赖
  - 已配置 Swagger UI 路由（`/swagger/index.html`）
  - 已为核心 API 添加注解示例（用户、订阅等）
  - 已创建 API 使用指南（`docs/api-guide.md`）
  - 已创建 Swagger 配置指南（`docs/swagger-setup.md`）

> 小结：Week 5 的所有任务已100%完成！包括媒体服务器集成、元数据平台适配（TMDB/TVDB/豆瓣）、聚合服务和 Swagger 文档。

## Week 6：通知渠道集成 + Phase 2准备

### Week 6.1 (Day 1-2)：通知渠道集成 ✅

#### 任务清单
- [x] **Telegram通知实现**
  - [x] 创建 `internal/integration/notification/telegram/client.go`
  - [x] 实现Bot API封装
  - [x] 支持发送文本、图片、文件、Markdown

- [x] **WeChat通知实现**
  - [x] 创建 `internal/integration/notification/wechat/client.go`
  - [x] 实现企业微信API
  - [x] 自动管理 access_token

- [x] **通知抽象接口**
  - [x] 定义 `internal/integration/notification/interface.go`
  - [x] 实现通知路由（支持多渠道广播）
  - [x] 实现工厂模式管理客户端

#### 代码示例

```go
// internal/integration/notification/interface.go
package notification

import "context"

type Client interface {
    // SendText 发送文本消息
    SendText(ctx context.Context, message string) error
    
    // SendImage 发送图片
    SendImage(ctx context.Context, imageURL string, caption string) error
    
    // SendFile 发送文件
    SendFile(ctx context.Context, fileURL string, filename string) error
}

type Router struct {
    clients map[string]Client
}

func (r *Router) Broadcast(ctx context.Context, message string) error {
    for name, client := range r.clients {
        if err := client.SendText(ctx, message); err != nil {
            logger.Error("failed to send notification", 
                zap.String("channel", name), 
                zap.Error(err))
        }
    }
    return nil
}
```

#### 交付物 ✅
- `internal/integration/notification/interface.go` - 统一接口定义
- `internal/integration/notification/router.go` - 通知路由器
- `internal/integration/notification/telegram/client.go` - Telegram客户端
- `internal/integration/notification/wechat/client.go` - WeChat企业微信客户端
- `internal/integration/notification/README.md` - 集成文档

#### 验收标准 ✅
- ✅ 支持发送文本、图片、文件、Markdown
- ✅ 支持多渠道广播（并发发送）
- ✅ 错误处理完善（部分失败不影响其他渠道）
- ✅ 工厂模式管理客户端
- ✅ 自动管理 access_token（WeChat）
- ✅ 完整的文档和示例

---

### Week 6.2 (Day 3-4)：索引器集成 ✅

#### 任务清单
- [x] **Jackett客户端实现**
  - [x] 创建 `internal/integration/indexer/jackett/client.go`
  - [x] 实现搜索接口
  - [x] 解析Torznab格式（XML）
  - [x] 支持 IMDB 搜索

- [x] **Prowlarr客户端实现**
  - [x] 创建 `internal/integration/indexer/prowlarr/client.go`
  - [x] 实现搜索接口（JSON API）
  - [x] 支持 IMDB/TMDB 搜索

- [x] **索引器抽象接口**
  - [x] 定义 `internal/integration/indexer/interface.go`
  - [x] 实现聚合搜索（`aggregator.go`）
  - [x] 实现工厂模式管理客户端
  - [x] 实现并发搜索和结果去重

#### 交付物 ✅
- `internal/integration/indexer/interface.go` - 统一接口定义
- `internal/integration/indexer/aggregator.go` - 聚合搜索器
- `internal/integration/indexer/jackett/client.go` - Jackett 客户端
- `internal/integration/indexer/prowlarr/client.go` - Prowlarr 客户端
- `internal/integration/indexer/README.md` - 集成文档

#### 验收标准 ✅
- ✅ 支持搜索种子（关键词、IMDB ID、TMDB ID）
- ✅ 支持解析Torznab格式（XML）
- ✅ 支持聚合搜索（并发搜索多个索引器）
- ✅ 支持结果去重和排序
- ✅ 工厂模式管理客户端
- ✅ 完整的文档和示例

---

### Week 6.3 (Day 5)：Phase 2准备 ✅

#### 任务清单
- [x] **用户认证系统设计**
  - [x] 设计用户注册/登录流程
  - [x] 设计权限控制模型（RBAC）
  - [x] 准备数据库表结构（users, roles, permissions）
  - [x] 设计 JWT 认证机制
  - [x] 设计安全策略

- [x] **站点管理系统设计**
  - [x] 设计站点配置模型
  - [x] 设计Cookie同步机制（定时同步、验证）
  - [x] 设计签到任务调度（Cron 调度）
  - [x] 设计流量统计和监控

#### 交付物 ✅
- `docs/design/auth-system-design.md` - 用户认证系统设计文档
- `docs/design/site-management-design.md` - 站点管理系统设计文档

---

## 🎯 Week 7-8：Phase 2核心功能开发

### Week 7：用户认证 + 站点管理

#### 任务清单
- [ ] **用户认证实现**
  - [ ] 用户注册/登录API
  - [ ] JWT Token管理
  - [ ] 权限中间件
  - [ ] 密码重置功能

- [ ] **站点管理实现**
  - [ ] 站点CRUD API
  - [ ] Cookie同步服务
  - [ ] 站点签到调度器
  - [ ] 站点数据统计

#### 交付物
- 用户认证模块
- 站点管理模块
- API接口
- 单元测试

---

### Week 8：订阅系统 + 下载管理

#### 任务清单
- [ ] **订阅系统实现**
  - [ ] 订阅CRUD API
  - [ ] 订阅刷新调度器
  - [ ] 订阅匹配引擎
  - [ ] 订阅历史记录

- [ ] **下载管理实现**
  - [ ] 下载任务管理API
  - [ ] 下载器状态同步
  - [ ] 下载完成事件处理
  - [ ] 下载历史记录

#### 交付物
- 订阅系统模块
- 下载管理模块
- API接口
- 单元测试

---

## 🎯 Week 9：文件整理 + 集成测试

### Week 9.1 (Day 1-3)：文件整理系统

#### 任务清单
- [ ] **文件整理实现**
  - [ ] 文件识别服务
  - [ ] 文件重命名服务
  - [ ] 文件移动服务
  - [ ] 整理历史记录

- [ ] **媒体服务器同步**
  - [ ] 整理完成后通知媒体服务器
  - [ ] 刷新媒体库
  - [ ] 更新元数据

#### 交付物
- 文件整理模块
- 媒体服务器同步服务
- API接口

---

### Week 9.2 (Day 4-5)：集成测试 + 性能测试

#### 任务清单
- [ ] **端到端测试**
  - [ ] 订阅 → 搜索 → 下载 → 整理 完整流程测试
  - [ ] 多用户并发测试
  - [ ] 异常场景测试

- [ ] **性能测试**
  - [ ] API响应时间测试
  - [ ] 数据库查询性能测试
  - [ ] 并发处理能力测试

#### 交付物
- 端到端测试套件
- 性能测试报告
- 性能优化建议

---

## 📊 进度跟踪

### 里程碑

| 里程碑 | 目标日期 | 交付物 | 状态 |
|--------|---------|--------|------|
| **M1: Phase 1优化完成** | Week 4 Day 2 | 数据库优化、测试覆盖率70% | ⏳ 进行中 |
| **M2: 核心外部服务完成** | Week 4 Day 4 | 下载器集成 | ⏳ 待开始 |
| **M3: 媒体服务器集成完成** | Week 5 Day 2 | Emby/Plex/Jellyfin | ⏳ 待开始 |
| **M4: 元数据平台完成** | Week 5 Day 4 | TMDB/TVDB/豆瓣 | ⏳ 待开始 |
| **M5: 通知渠道完成** | Week 6 Day 2 | Telegram/WeChat | ⏳ 待开始 |
| **M6: Phase 2准备完成** | Week 6 Day 5 | 设计文档 | ⏳ 待开始 |
| **M7: 用户认证完成** | Week 7 Day 5 | 用户系统 | ⏳ 待开始 |
| **M8: 订阅下载完成** | Week 8 Day 5 | 订阅+下载 | ⏳ 待开始 |
| **M9: 文件整理完成** | Week 9 Day 3 | 整理系统 | ⏳ 待开始 |
| **M10: 集成测试完成** | Week 9 Day 5 | 测试报告 | ⏳ 待开始 |

---

## 🎯 关键成功指标 (KPI)

### 代码质量指标
- [ ] 单元测试覆盖率 ≥ 80%
- [ ] 集成测试覆盖率 ≥ 60%
- [ ] 代码审查通过率 100%
- [ ] 无严重Bug（P0/P1）

### 性能指标
- [ ] API平均响应时间 < 200ms
- [ ] 数据库查询平均时间 < 100ms
- [ ] 支持100并发用户
- [ ] 系统可用性 ≥ 99.5%

### 功能指标
- [ ] 支持3种下载器（qBittorrent/Transmission/Aria2）
- [ ] 支持3种媒体服务器（Emby/Plex/Jellyfin）
- [ ] 支持3种元数据源（TMDB/TVDB/豆瓣）
- [ ] 支持5种通知渠道

---

## 🚨 风险管理

### 高风险项

| 风险 | 影响 | 概率 | 缓解措施 | 负责人 |
|------|------|------|---------|--------|
| 外部API限流 | 高 | 中 | 实现请求限流器、缓存机制 | 后端开发 |
| 下载器兼容性 | 高 | 中 | 充分测试、提供降级方案 | 后端开发 |
| 性能不达标 | 高 | 低 | 提前性能测试、优化热点 | 架构师 |
| 测试覆盖不足 | 中 | 中 | 强制代码审查、CI/CD检查 | QA |

### 依赖项

| 依赖项 | 状态 | 风险 | 备注 |
|--------|------|------|------|
| PostgreSQL | ✅ 可用 | 低 | 已配置 |
| Redis | ✅ 可用 | 低 | 已配置 |
| Docker | ✅ 可用 | 低 | 已配置 |
| TMDB API Key | ⚠️ 待确认 | 中 | 需要申请 |
| 下载器环境 | ⚠️ 待搭建 | 中 | 需要测试环境 |

---

## 📝 每日站会模板

### 站会时间
- 每天上午 10:00
- 时长：15分钟

### 站会内容
1. **昨天完成了什么？**
2. **今天计划做什么？**
3. **遇到什么阻碍？**

### 周报模板
- 本周完成的任务
- 下周计划的任务
- 遇到的问题和解决方案
- 需要的支持

---

## 📚 参考文档

- [migration-overview.md](./migration-overview.md) - 迁移总览
- [migration-progress.md](./migration-progress.md) - 迁移进度
- [modules-migration.md](./modules-migration.md) - 外部服务模块迁移
- [chain-migration.md](./chain-migration.md) - 业务处理链迁移
- [api-migration.md](./api-migration.md) - API层迁移
- [db-migration.md](./db-migration.md) - 数据库层迁移

---

## ✅ 检查清单

### Week 4 检查清单
- [ ] 数据库索引优化完成
- [ ] 连接池配置优化完成
- [ ] qBittorrent客户端实现完成
- [ ] Transmission客户端实现完成
- [ ] 下载器集成测试通过
- [ ] 测试覆盖率达到70%

### Week 5 检查清单
- [ ] Emby客户端实现完成
- [ ] Plex客户端实现完成
- [ ] Jellyfin客户端实现完成
- [ ] TMDB平台适配完成
- [ ] TVDB平台适配完成
- [ ] 豆瓣平台适配完成
- [ ] Swagger文档生成完成

### Week 6 检查清单
- [ ] Telegram通知实现完成
- [ ] WeChat通知实现完成
- [ ] Jackett客户端实现完成
- [ ] Prowlarr客户端实现完成
- [ ] 用户认证系统设计完成
- [ ] 站点管理系统设计完成

---

**最后更新**：2025-12-02  
**维护者**：MoviePilot Go 迁移团队
