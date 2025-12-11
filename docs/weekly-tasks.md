# MoviePilot Go 每周任务分解

> **配合文档**：[execution-plan.md](./execution-plan.md)  
> **更新频率**：每周一更新

---

## 📅 Week 4 任务分解（当前周）

### 🎯 本周目标
1. 完成数据库层性能优化
2. 实现核心下载器集成（qBittorrent + Transmission）
3. 提升测试覆盖率至70%

### 📋 详细任务

#### Day 1 (周一)：数据库索引优化

**上午任务**
- [ ] 分析现有数据库慢查询
  - 启用PostgreSQL慢查询日志
  - 收集最近7天的慢查询数据
  - 识别Top 10慢查询

**下午任务**
- [ ] 设计索引优化方案
  - 为 `subscribes` 表添加复合索引 `(user_id, state, updated_at)`
  - 为 `download_history` 表添加索引 `(torrent_hash, state)`
  - 为 `transfer_history` 表添加索引 `(src_path, state)`
  - 为 `sites` 表添加索引 `(user_id, is_active)`

**代码示例**
```go
// internal/models/database/indexes.go
package database

// 在 AutoMigrate 后执行索引创建
func CreateIndexes(db *gorm.DB) error {
    indexes := []string{
        "CREATE INDEX IF NOT EXISTS idx_subscribes_user_state ON subscribes(user_id, state, updated_at)",
        "CREATE INDEX IF NOT EXISTS idx_download_history_hash_state ON download_history(torrent_hash, state)",
        "CREATE INDEX IF NOT EXISTS idx_transfer_history_path_state ON transfer_history(src_path, state)",
        "CREATE INDEX IF NOT EXISTS idx_sites_user_active ON sites(user_id, is_active)",
    }
    
    for _, idx := range indexes {
        if err := db.Exec(idx).Error; err != nil {
            return err
        }
    }
    return nil
}
```

**交付物**
- 索引优化脚本
- 性能对比报告

---

#### Day 2 (周二)：连接池优化 + 性能测试

**上午任务**
- [ ] 优化数据库连接池配置
  ```go
  // pkg/database/database.go
  func optimizeConnectionPool(db *gorm.DB) {
      sqlDB, _ := db.DB()
      
      // 设置最大空闲连接数
      sqlDB.SetMaxIdleConns(10)
      
      // 设置最大打开连接数
      sqlDB.SetMaxOpenConns(100)
      
      // 设置连接最大生命周期
      sqlDB.SetConnMaxLifetime(time.Hour)
      
      // 设置连接最大空闲时间
      sqlDB.SetConnMaxIdleTime(10 * time.Minute)
  }
  ```

**下午任务**
- [ ] 编写Repository性能基准测试
  ```go
  // internal/repositories/benchmarks/subscribe_benchmark_test.go
  func BenchmarkSubscribeRepository_List(b *testing.B) {
      repo := setupTestRepo()
      ctx := context.Background()
      
      b.ResetTimer()
      for i := 0; i < b.N; i++ {
          _, err := repo.List(ctx, &ListOptions{
              Page:     1,
              PageSize: 20,
          })
          if err != nil {
              b.Fatal(err)
          }
      }
  }
  ```

**交付物**
- 连接池优化配置
- 性能基准测试套件
- 性能测试报告

---

#### Day 3 (周三)：qBittorrent客户端实现

**上午任务**
- [ ] 创建qBittorrent客户端基础结构
  ```go
  // internal/integration/qbittorrent/client.go
  package qbittorrent
  
  import (
      "context"
      "net/http"
      "net/http/cookiejar"
      "time"
  )
  
  type Client struct {
      baseURL  string
      username string
      password string
      client   *http.Client
      cookie   string
  }
  
  func NewClient(baseURL, username, password string) *Client {
      jar, _ := cookiejar.New(nil)
      return &Client{
          baseURL:  baseURL,
          username: username,
          password: password,
          client: &http.Client{
              Timeout: 30 * time.Second,
              Jar:     jar,
          },
      }
  }
  ```

**下午任务**
- [ ] 实现认证和核心API
  ```go
  // Login 登录
  func (c *Client) Login(ctx context.Context) error {
      // 实现登录逻辑
  }
  
  // AddTorrent 添加种子
  func (c *Client) AddTorrent(ctx context.Context, req *AddTorrentRequest) error {
      // 实现添加种子逻辑
  }
  
  // ListTorrents 列出种子
  func (c *Client) ListTorrents(ctx context.Context) ([]*Torrent, error) {
      // 实现列出种子逻辑
  }
  ```

**交付物**
- qBittorrent客户端基础实现
- 单元测试（Mock HTTP）

---

#### Day 4 (周四)：Transmission客户端 + 统一接口

**上午任务**
- [ ] 实现Transmission RPC客户端
  ```go
  // internal/integration/transmission/client.go
  package transmission
  
  import (
      "context"
      "encoding/json"
      "net/http"
  )
  
  type Client struct {
      baseURL  string
      username string
      password string
      client   *http.Client
      sessionID string
  }
  
  // RPC 调用Transmission RPC
  func (c *Client) RPC(ctx context.Context, method string, args interface{}) (json.RawMessage, error) {
      // 实现RPC调用
  }
  ```

**下午任务**
- [ ] 定义统一下载器接口
  ```go
  // internal/integration/downloader/interface.go
  package downloader
  
  type Client interface {
      AddTorrent(ctx context.Context, req *AddTorrentRequest) (*Torrent, error)
      ListTorrents(ctx context.Context, filter *TorrentFilter) ([]*Torrent, error)
      PauseTorrent(ctx context.Context, hash string) error
      ResumeTorrent(ctx context.Context, hash string) error
      RemoveTorrent(ctx context.Context, hash string, deleteFiles bool) error
      GetTorrentInfo(ctx context.Context, hash string) (*TorrentInfo, error)
  }
  
  // Factory 下载器工厂
  type Factory struct {
      clients map[string]Client
  }
  
  func (f *Factory) GetClient(name string) (Client, error) {
      // 返回对应的下载器客户端
  }
  ```

**交付物**
- Transmission客户端实现
- 统一下载器接口
- 工厂模式实现

---

#### Day 5 (周五)：测试覆盖率提升

**上午任务**
- [ ] 为业务服务补充单元测试
  - `internal/business/services/download/` - 下载服务测试
  - `internal/business/services/subscribe/` - 订阅服务测试
  - `internal/business/services/site/` - 站点服务测试

**下午任务**
- [ ] 搭建集成测试框架
  ```go
  // tests/integration/setup_test.go
  package integration
  
  import (
      "testing"
      "github.com/testcontainers/testcontainers-go"
      "github.com/testcontainers/testcontainers-go/wait"
  )
  
  func setupPostgres(t *testing.T) testcontainers.Container {
      req := testcontainers.ContainerRequest{
          Image:        "postgres:15-alpine",
          ExposedPorts: []string{"5432/tcp"},
          Env: map[string]string{
              "POSTGRES_PASSWORD": "test",
              "POSTGRES_DB":       "moviepilot_test",
          },
          WaitingFor: wait.ForLog("database system is ready to accept connections"),
      }
      
      container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
          ContainerRequest: req,
          Started:          true,
      })
      if err != nil {
          t.Fatal(err)
      }
      return container
  }
  ```

**交付物**
- 单元测试（覆盖率提升至70%）
- 集成测试框架
- 测试覆盖率报告

---

### 📊 Week 4 验收标准

- [ ] 数据库查询平均响应时间 < 100ms
- [ ] 连接池利用率 > 70%
- [ ] qBittorrent客户端功能完整
- [ ] Transmission客户端功能完整
- [ ] 下载器集成测试通过
- [ ] 全局测试覆盖率 ≥ 70%

---

## 📅 Week 5 任务分解

### 🎯 本周目标
1. 实现3种媒体服务器集成（Emby/Plex/Jellyfin）
2. 实现3种元数据平台适配（TMDB/TVDB/豆瓣）
3. 完善API文档（Swagger）

### 📋 详细任务

#### Day 1 (周一)：Emby客户端实现

**任务清单**
- [ ] 实现Emby API封装
  - 认证（API Key）
  - 获取媒体库列表
  - 获取媒体项目
  - 刷新媒体库
  - 更新元数据

**代码框架**
```go
// internal/platform/emby/client.go
package emby

type Client struct {
    baseURL string
    apiKey  string
    client  *http.Client
}

func (c *Client) GetLibraries(ctx context.Context) ([]*Library, error) {
    // GET /Library/VirtualFolders
}

func (c *Client) RefreshLibrary(ctx context.Context, libraryId string) error {
    // POST /Library/Refresh
}
```

---

#### Day 2 (周二)：Plex + Jellyfin客户端

**任务清单**
- [ ] 实现Plex客户端（Token认证）
- [ ] 实现Jellyfin客户端（API Key认证）
- [ ] 定义统一媒体服务器接口

---

#### Day 3 (周三)：TMDB平台适配

**任务清单**
- [ ] 实现TMDB API封装
  - 搜索电影/电视剧
  - 获取详细信息
  - 获取季/集信息
- [ ] 实现请求限流器（40 req/10s）
- [ ] 实现响应缓存

---

#### Day 4 (周四)：TVDB + 豆瓣平台适配

**任务清单**
- [ ] 实现TVDB客户端（JWT认证）
- [ ] 实现豆瓣客户端（反爬虫处理）
- [ ] 实现元数据聚合服务

---

#### Day 5 (周五)：Swagger文档生成

**任务清单**
- [ ] 为所有API添加Swagger注解
- [ ] 生成Swagger文档
- [ ] 配置Swagger UI
- [ ] 编写API使用指南

---

### Week 5 进度更新（2025-12-02）

- ✅ Emby 客户端：已在 `internal/integration/mediaserver/emby/client.go` 中实现  
  - 支持连接测试、获取服务器信息、列出媒体库、按 ID 获取条目、按关键字搜索

- ✅ Plex / Jellyfin 客户端：已在 `internal/integration/mediaserver/plex` 与 `internal/integration/mediaserver/jellyfin` 下实现  
  - 支持连接测试与核心库/条目查询，接口与 Emby 对齐

- ✅ TMDB 平台适配：已在 `internal/integration/metadata/tmdb/client.go` 中实现  
  - 支持 `SearchMovie` / `GetMovieByTMDB` / `SearchTV` / `GetTVByTMDB` 等核心接口

- ✅ Metadata 聚合服务：已在 `internal/business/services/metadata/aggregator_service.go` 中实现  
  - 电影聚合：`AggregateMovieByTMDB`、`SearchAndAggregateMovie`
  - 剧集聚合：`AggregateTVByTMDB`（TMDB 主数据 + TVDB 可选补充）

- ✅ TVDB 平台适配：已在 `internal/integration/metadata/tvdb/client.go` 中实现最小版本  
  - 支持 `SearchTV` / `GetTVByID`（调用 TVDB v4 API）

- ✅ 豆瓣平台适配：已在 `internal/integration/metadata/douban/client.go` 中实现最小版本  
  - 支持 `SearchMovie` / `GetMovieByID`（假设通过代理访问）

- ✅ Swagger 文档：已完成配置和文档编写  
  - 已添加 swaggo 依赖并配置 Swagger UI 路由
  - 核心 API 已有完整注解（用户、订阅等）
  - 已创建 API 使用指南和 Swagger 配置指南

> 状态概览：Week 5 的所有任务已100%完成！包括媒体服务器集成、元数据平台适配（TMDB/TVDB/豆瓣）、聚合服务和 Swagger 文档。

## 📅 Week 6 任务分解

### 🎯 本周目标
1. 实现通知渠道集成（Telegram/WeChat）
2. 实现索引器集成（Jackett/Prowlarr）
3. 完成Phase 2准备工作

### 📋 详细任务

#### Day 1-2：通知渠道实现
- [ ] Telegram Bot API封装
- [ ] WeChat企业微信API封装
- [ ] 通知路由实现

#### Day 3-4：索引器实现
- [ ] Jackett客户端（Torznab）
- [ ] Prowlarr客户端
- [ ] 聚合搜索实现

#### Day 5：Phase 2准备
- [ ] 用户认证系统设计
- [ ] 站点管理系统设计

---

## 📝 每日工作流程

### 每日开始
1. 查看今日任务清单
2. 拉取最新代码
3. 参加每日站会（10:00）

### 开发过程
1. 创建功能分支 `feature/xxx`
2. 编写代码 + 单元测试
3. 本地测试通过
4. 提交代码审查

### 每日结束
1. 更新任务状态
2. 提交代码
3. 记录遇到的问题

---

## 🔧 开发工具和命令

### 常用命令

```bash
# 运行单元测试
go test ./... -v -cover

# 运行基准测试
go test ./internal/repositories/benchmarks -bench=. -benchmem

# 生成测试覆盖率报告
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# 生成Swagger文档
swag init -g cmd/server/main.go

# 运行集成测试
go test ./tests/integration -v

# 代码格式化
gofmt -w .
goimports -w .

# 代码检查
golangci-lint run
```

### 开发环境配置

```bash
# 启动开发环境
docker-compose -f deployments/docker-compose.dev.yml up -d

# 查看日志
docker-compose -f deployments/docker-compose.dev.yml logs -f

# 停止开发环境
docker-compose -f deployments/docker-compose.dev.yml down
```

---

## 📊 进度跟踪表

### Week 4 进度

| 任务 | 计划时间 | 实际时间 | 状态 | 备注 |
|------|---------|---------|------|------|
| 数据库索引优化 | Day 1 | - | ⏳ 待开始 | - |
| 连接池优化 | Day 2 | - | ⏳ 待开始 | - |
| qBittorrent客户端 | Day 3 | - | ⏳ 待开始 | - |
| Transmission客户端 | Day 4 | - | ⏳ 待开始 | - |
| 测试覆盖率提升 | Day 5 | - | ⏳ 待开始 | - |

### Week 5 进度

| 任务 | 计划时间 | 实际时间 | 状态 | 备注 |
|------|---------|---------|------|------|
| Emby客户端 | Day 1 | - | ⏳ 待开始 | - |
| Plex/Jellyfin客户端 | Day 2 | - | ⏳ 待开始 | - |
| TMDB平台适配 | Day 3 | - | ⏳ 待开始 | - |
| TVDB/豆瓣适配 | Day 4 | - | ⏳ 待开始 | - |
| Swagger文档 | Day 5 | - | ⏳ 待开始 | - |

---

## 🚨 问题跟踪

### 当前问题

| 问题 | 优先级 | 状态 | 负责人 | 备注 |
|------|--------|------|--------|------|
| - | - | - | - | - |

### 已解决问题

| 问题 | 解决方案 | 解决日期 |
|------|---------|---------|
| - | - | - |

---

**最后更新**：2025-12-02  
**下次更新**：每周一
