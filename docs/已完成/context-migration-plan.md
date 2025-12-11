# Context 上下文系统迁移计划

> Python: `app/core/context.py`  
> Go: `internal/business/domains/media/` + `internal/models/dto/`

---

## 1. Python 上下文系统概览

### 1.1 核心功能

- **上下文聚合**：整合种子信息、媒体信息、元数据等
- **上下文传递**：在业务流程中传递上下文数据
- **上下文扩展**：支持动态添加上下文字段

### 1.2 典型使用场景

- 种子下载 → 媒体识别 → 文件整理 → 媒体服务器同步
- 订阅检查 → 资源搜索 → 下载添加 → 通知发送

---

## 2. Go 目标设计

### 2.1 分层设计

| 层级 | 位置 | 职责 |
|------|------|------|
| 领域层 | `internal/business/domains/media/` | 核心上下文模型定义 |
| 数据传输层 | `internal/models/dto/` | 上下文数据传输对象 |

### 2.2 核心模型

**种子上下文**：

```go
// internal/business/domains/media/torrent.go
type TorrentInfo struct {
    // 站点相关
    SiteID         int    `json:"site_id"`
    SiteName       string `json:"site_name"`
    SiteCookie     string `json:"site_cookie"`
    SiteUserAgent  string `json:"site_ua"`
    SiteProxy      bool   `json:"site_proxy"`
    
    // 资源内容
    Title       string  `json:"title"`
    Description string  `json:"description"`
    IMDBID      string  `json:"imdbid"`
    Enclosure   string  `json:"enclosure"`
    PageURL     string  `json:"page_url"`
    Size        float64 `json:"size"`
    
    // 实时状态
    Seeders int       `json:"seeders"`
    Peers   int       `json:"peers"`
    Grabs   int       `json:"grabs"`
    PubDate time.Time `json:"pubdate"`
}
```

**媒体上下文**：

```go
// internal/business/domains/media/media.go
type MediaInfo struct {
    Source   MediaSource `json:"source"`
    Type     MediaType   `json:"type"`
    
    Title    string `json:"title"`
    EnTitle  string `json:"en_title"`
    Year     string `json:"year"`
    Season   int    `json:"season"`
    
    TMDBID   int64  `json:"tmdb_id"`
    IMDBID   string `json:"imdb_id"`
    TVDBID   int64  `json:"tvdb_id"`
    DoubanID string `json:"douban_id"`
    
    Overview    string  `json:"overview"`
    VoteAverage float64 `json:"vote_average"`
}
```

**上下文聚合**：

```go
// internal/business/domains/media/context.go
type Context struct {
    MetaInfo                MetaBase   // 元数据信息
    MediaInfo               *MediaInfo // 媒体信息
    TorrentInfo             *TorrentInfo // 种子信息
    MediaRecognizeFailCount int // 媒体识别失败次数
}
```

### 2.3 迁移步骤

1. **定义核心上下文模型**：在 `internal/business/domains/media/` 中定义种子、媒体和上下文聚合模型
2. **实现上下文数据传输对象**：在 `internal/models/dto/` 中定义上下文DTO
3. **实现上下文转换逻辑**：实现不同层级上下文之间的转换
4. **集成到业务流程**：在业务服务中使用上下文模型

---

## 3. 集成与使用

### 3.1 上下文创建示例

```go
// 创建上下文
ctx := &media.Context{
    TorrentInfo: &media.TorrentInfo{
        SiteName: "站点名称",
        Title:    "资源标题",
        Size:     1024,
    },
    MediaInfo: &media.MediaInfo{
        Title:  "媒体标题",
        Year:   "2023",
        TMDBID: 12345,
    },
}
```

### 3.2 上下文传递示例

```go
// 在业务流程中传递上下文
func (s *DownloadService) AddDownload(ctx context.Context, mediaCtx *media.Context) error {
    // 1. 添加下载任务
    // 2. 更新上下文状态
    // 3. 发布事件
    return s.eventBus.PublishBroadcast(ctx, events.EventTypeDownloadAdded, mediaCtx, 10)
}
```

---

## 4. 检查清单

- [x] 定义种子上下文模型
- [x] 定义媒体上下文模型
- [x] 定义上下文聚合模型
- [x] 实现上下文数据传输对象
- [x] 集成到业务流程中

---

## 5. 后续优化

1. **上下文扩展机制**：支持动态添加上下文字段
2. **上下文序列化**：支持上下文的JSON序列化和反序列化
3. **上下文验证**：添加上下文数据验证逻辑
4. **上下文版本控制**：支持不同版本上下文的兼容处理
