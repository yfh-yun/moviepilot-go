# Indexer 索引器模块

> **版本**: v1.0.0  
> **更新时间**: 2025-12-02

---

## 📋 目录

1. [概述](#概述)
2. [架构设计](#架构设计)
3. [支持的索引器](#支持的索引器)
4. [快速开始](#快速开始)
5. [API 参考](#api-参考)
6. [最佳实践](#最佳实践)

---

## 概述

索引器模块提供了统一的种子搜索接口，支持多种索引器（Jackett、Prowlarr），并提供聚合搜索功能，可以同时从多个索引器搜索并合并结果。

### 核心特性

- ✅ **统一接口**：所有索引器实现相同的接口
- ✅ **多索引器支持**：Jackett、Prowlarr
- ✅ **聚合搜索**：并发搜索多个索引器并合并结果
- ✅ **Torznab 支持**：完整支持 Torznab 协议
- ✅ **IMDB/TMDB 搜索**：支持通过 IMDB ID 和 TMDB ID 搜索
- ✅ **结果去重**：自动去除重复的种子
- ✅ **智能排序**：按做种数自动排序

---

## 架构设计

### 核心组件

```
indexer/
├── interface.go           # 统一接口定义
├── aggregator.go          # 聚合搜索器
├── jackett/               # Jackett 客户端
│   └── client.go
└── prowlarr/              # Prowlarr 客户端
    └── client.go
```

### 接口定义

```go
type Client interface {
    Name() string
    Search(ctx context.Context, opts SearchOptions) ([]*Torrent, error)
    TestConnection(ctx context.Context) error
    GetCapabilities(ctx context.Context) (*Capabilities, error)
}
```

### 种子结构

```go
type Torrent struct {
    Title       string          // 标题
    Link        string          // 种子链接
    MagnetURL   string          // 磁力链接
    Size        int64           // 文件大小（字节）
    Seeders     int             // 做种数
    Leechers    int             // 下载数
    PublishDate time.Time       // 发布时间
    Category    TorrentCategory // 分类
    IndexerName string          // 索引器名称
    IMDBID      string          // IMDB ID
    TMDBID      int             // TMDB ID
}
```

---

## 支持的索引器

### 1. Jackett

**特性**：
- ✅ 支持 Torznab 协议
- ✅ 支持 IMDB 搜索
- ✅ 支持多分类搜索
- ✅ 聚合多个 Tracker

**配置示例**：

```go
cfg := jackett.Config{
    BaseURL: "http://localhost:9117",
    APIKey:  "your_api_key",
    Timeout: 30 * time.Second,
}

client, err := jackett.NewClient(cfg)
```

**获取 API Key**：
1. 访问 Jackett Web UI
2. 点击右上角的 "API Key"
3. 复制显示的 API Key

### 2. Prowlarr

**特性**：
- ✅ 支持 IMDB 搜索
- ✅ 支持 TMDB 搜索
- ✅ 现代化 REST API
- ✅ 更好的索引器管理

**配置示例**：

```go
cfg := prowlarr.Config{
    BaseURL: "http://localhost:9696",
    APIKey:  "your_api_key",
    Timeout: 30 * time.Second,
}

client, err := prowlarr.NewClient(cfg)
```

**获取 API Key**：
1. 访问 Prowlarr Web UI
2. 进入 Settings → General
3. 复制 API Key

---

## 快速开始

### 1. 创建索引器客户端

```go
package main

import (
    "context"
    "fmt"
    "time"

    "moviepilot-go/internal/integration/indexer"
    "moviepilot-go/internal/integration/indexer/jackett"
    "moviepilot-go/internal/integration/indexer/prowlarr"
)

func main() {
    // 创建 Jackett 客户端
    jackettClient, err := jackett.NewClient(jackett.Config{
        BaseURL: "http://localhost:9117",
        APIKey:  "your_jackett_api_key",
    })
    if err != nil {
        panic(err)
    }

    // 创建 Prowlarr 客户端
    prowlarrClient, err := prowlarr.NewClient(prowlarr.Config{
        BaseURL: "http://localhost:9696",
        APIKey:  "your_prowlarr_api_key",
    })
    if err != nil {
        panic(err)
    }

    // 创建工厂并注册客户端
    factory := indexer.NewFactory()
    factory.Register(jackettClient)
    factory.Register(prowlarrClient)

    // 创建聚合搜索器
    aggregator := indexer.NewAggregator(factory)

    // 搜索种子
    ctx := context.Background()
    results, err := aggregator.Search(ctx, indexer.SearchOptions{
        Query:      "The Last of Us S01E05",
        Category:   indexer.CategoryTV,
        MinSeeders: 5,
        Limit:      20,
    })
    if err != nil {
        panic(err)
    }

    // 打印结果
    for _, torrent := range results {
        fmt.Printf("标题: %s\n", torrent.Title)
        fmt.Printf("做种: %d | 下载: %d | 大小: %d MB\n",
            torrent.Seeders, torrent.Leechers, torrent.Size/1024/1024)
        fmt.Printf("链接: %s\n\n", torrent.MagnetURL)
    }
}
```

### 2. 搜索不同类型的内容

```go
// 搜索电影
results, err := aggregator.Search(ctx, indexer.SearchOptions{
    Query:    "Inception 2010",
    Category: indexer.CategoryMovie,
    Limit:    10,
})

// 通过 IMDB ID 搜索
results, err := aggregator.Search(ctx, indexer.SearchOptions{
    IMDBID:   "tt1375666",
    Category: indexer.CategoryMovie,
})

// 通过 TMDB ID 搜索（仅 Prowlarr）
results, err := aggregator.Search(ctx, indexer.SearchOptions{
    TMDBID:   27205,
    Category: indexer.CategoryMovie,
})

// 搜索动漫
results, err := aggregator.Search(ctx, indexer.SearchOptions{
    Query:    "進撃の巨人",
    Category: indexer.CategoryAnime,
})
```

### 3. 从指定索引器搜索

```go
// 只从 Jackett 搜索
results, err := aggregator.SearchByIndexer(ctx, "jackett", indexer.SearchOptions{
    Query: "The Last of Us",
})

// 只从 Prowlarr 搜索
results, err := aggregator.SearchByIndexer(ctx, "prowlarr", indexer.SearchOptions{
    Query: "The Last of Us",
})
```

### 4. 测试连接

```go
// 测试所有索引器
results := aggregator.TestAllIndexers(ctx)
for name, err := range results {
    if err != nil {
        fmt.Printf("❌ %s: %v\n", name, err)
    } else {
        fmt.Printf("✅ %s: 连接成功\n", name)
    }
}
```

---

## API 参考

### Aggregator 方法

#### Search

聚合搜索所有索引器。

```go
func (a *Aggregator) Search(ctx context.Context, opts SearchOptions) ([]*Torrent, error)
```

**参数**：
- `ctx`: 上下文
- `opts`: 搜索选项
  - `Query`: 搜索关键词
  - `Category`: 分类过滤
  - `IMDBID`: IMDB ID 过滤
  - `TMDBID`: TMDB ID 过滤
  - `Limit`: 结果数量限制
  - `MinSeeders`: 最小做种数

**返回**：
- `[]*Torrent`: 种子列表（已去重和排序）
- `error`: 错误信息

#### SearchByIndexer

从指定索引器搜索。

```go
func (a *Aggregator) SearchByIndexer(ctx context.Context, indexerName string, opts SearchOptions) ([]*Torrent, error)
```

#### TestAllIndexers

测试所有索引器的连接。

```go
func (a *Aggregator) TestAllIndexers(ctx context.Context) map[string]error
```

#### GetAllCapabilities

获取所有索引器的能力。

```go
func (a *Aggregator) GetAllCapabilities(ctx context.Context) map[string]*Capabilities
```

---

## 最佳实践

### 1. 超时控制

使用 context 控制超时：

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

results, err := aggregator.Search(ctx, opts)
```

### 2. 结果过滤

使用 `MinSeeders` 过滤低质量种子：

```go
results, err := aggregator.Search(ctx, indexer.SearchOptions{
    Query:      "The Last of Us",
    MinSeeders: 10, // 至少 10 个做种
})
```

### 3. 自定义过滤

使用 `FilterTorrents` 方法自定义过滤：

```go
// 过滤大于 5GB 的种子
filtered := aggregator.FilterTorrents(results, func(t *indexer.Torrent) bool {
    return t.Size > 5*1024*1024*1024
})

// 过滤包含特定关键词的种子
filtered := aggregator.FilterTorrents(results, func(t *indexer.Torrent) bool {
    return strings.Contains(strings.ToLower(t.Title), "1080p")
})
```

### 4. 错误处理

聚合搜索时，部分索引器失败不影响其他索引器：

```go
results, err := aggregator.Search(ctx, opts)
// err 为 nil 表示至少有一个索引器成功
// 即使部分索引器失败，仍会返回可用结果
```

### 5. 性能优化

对于大量搜索，建议使用 goroutine 池：

```go
type SearchTask struct {
    Query string
    Opts  indexer.SearchOptions
}

func searchBatch(tasks []SearchTask) {
    var wg sync.WaitGroup
    sem := make(chan struct{}, 5) // 限制并发数

    for _, task := range tasks {
        wg.Add(1)
        go func(t SearchTask) {
            defer wg.Done()
            sem <- struct{}{}        // 获取信号量
            defer func() { <-sem }() // 释放信号量

            results, err := aggregator.Search(ctx, t.Opts)
            // 处理结果...
        }(task)
    }

    wg.Wait()
}
```

---

## Torznab 协议

### 什么是 Torznab？

Torznab 是一个标准化的种子搜索 API 协议，基于 Newznab 协议，专门用于 BitTorrent 索引器。

### 支持的参数

| 参数 | 说明 | 示例 |
|------|------|------|
| `q` | 搜索关键词 | `The Last of Us` |
| `cat` | 分类 ID | `5000` (TV) |
| `imdbid` | IMDB ID | `tt1234567` |
| `limit` | 结果数量 | `50` |
| `offset` | 结果偏移 | `0` |

### 分类 ID

| 分类 | ID |
|------|-----|
| Movies | 2000 |
| TV | 5000 |
| Anime | 5070 |
| Audio | 3000 |

---

## 常见问题

### 1. Jackett 搜索失败

**问题**：`Jackett API 返回错误: status=401`

**解决方案**：
- 检查 API Key 是否正确
- 确保 Jackett 服务正在运行
- 检查 BaseURL 是否正确（包括端口）

### 2. Prowlarr 搜索失败

**问题**：`Prowlarr API 返回错误: status=404`

**解决方案**：
- 检查 Prowlarr 版本（需要 v1.0+）
- 确保 API 路径正确（`/api/v1/search`）
- 检查是否已添加索引器

### 3. 搜索结果为空

**问题**：搜索返回 0 个结果

**解决方案**：
- 检查搜索关键词是否正确
- 降低 `MinSeeders` 要求
- 检查索引器是否已配置 Tracker
- 尝试使用 IMDB ID 或 TMDB ID 搜索

### 4. 搜索超时

**问题**：`context deadline exceeded`

**解决方案**：
- 增加超时时间：`Timeout: 60 * time.Second`
- 减少同时搜索的索引器数量
- 检查网络连接

---

## 扩展新的索引器

### 1. 实现 Client 接口

```go
package myindexer

import (
    "context"
    "moviepilot-go/internal/integration/indexer"
)

type Client struct {
    // ...
}

func (c *Client) Name() string {
    return "myindexer"
}

func (c *Client) Search(ctx context.Context, opts indexer.SearchOptions) ([]*indexer.Torrent, error) {
    // 实现搜索逻辑
    return nil, nil
}

// 实现其他接口方法...

var _ indexer.Client = (*Client)(nil)
```

### 2. 注册到工厂

```go
myClient := myindexer.NewClient(config)
factory.Register(myClient)
```

---

## 更新日志

### v1.0.0 (2025-12-02)

- ✅ 初始版本
- ✅ 支持 Jackett
- ✅ 支持 Prowlarr
- ✅ 实现聚合搜索
- ✅ 支持 Torznab 协议
- ✅ 支持 IMDB/TMDB 搜索
- ✅ 自动去重和排序

---

## 参考资源

- [Jackett GitHub](https://github.com/Jackett/Jackett)
- [Prowlarr GitHub](https://github.com/Prowlarr/Prowlarr)
- [Torznab 规范](https://torznab.github.io/spec-1.3-draft/)
- [MoviePilot Go 项目文档](../../docs/README.md)
