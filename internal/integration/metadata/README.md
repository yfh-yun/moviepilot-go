# Metadata 统一接口设计（TMDB / TVDB / 豆瓣）

> 统一封装电影与剧集元数据查询，屏蔽不同平台（TMDB/TVDB/豆瓣）的差异，为业务层提供一致的数据结构和调用方式。

---

## 目录结构

```text
internal/integration/metadata/
  interface.go          # 统一接口与核心领域结构
  tmdb/
    client.go           # TMDB 实现（电影 & 剧集主来源）
  tvdb/
    client.go           # TVDB 实现（以剧集为主）
  douban/
    client.go           # 豆瓣实现（通常经由代理）
```

---

## 核心概念

### 1. ProviderName

```go
type ProviderName string

const (
    ProviderTMDB   ProviderName = "tmdb"
    ProviderTVDB   ProviderName = "tvdb"
    ProviderDouban ProviderName = "douban"
)
```

### 2. 统一数据结构

#### MovieInfo

```go
type MovieInfo struct {
    ID          string       // 提供方内部ID
    Provider    ProviderName // tmdb/tvdb/douban
    Title       string
    Original    string
    Year        int
    Overview    string
    PosterURL   string
    BackdropURL string
    TMDBID      *int
    IMDBID      *string
    DoubanID    *string
}
```

#### TVShowInfo

```go
type TVShowInfo struct {
    ID        string
    Provider  ProviderName
    Title     string
    Original  string
    Year      int
    Overview  string
    PosterURL string
    TMDBID    *int
    TVDBID    *int
    DoubanID  *string
    Seasons   []TVSeasonInfo
}
```

#### SearchOptions

```go
type SearchOptions struct {
    Language Language // ISO-639-1，如 "zh-CN"、"en-US"
    Year     int
    Page     int
    Limit    int
}
```

---

## 统一接口：MetadataProvider

```go
type MetadataProvider interface {
    Name() ProviderName

    // 搜索电影/剧集
    SearchMovie(ctx context.Context, keyword string, opts SearchOptions) ([]*MovieInfo, error)
    SearchTV(ctx context.Context, keyword string, opts SearchOptions) ([]*TVShowInfo, error)

    // 通过 TMDB ID 获取详情
    GetMovieByTMDB(ctx context.Context, tmdbID int, lang Language) (*MovieInfo, error)
    GetTVByTMDB(ctx context.Context, tmdbID int, lang Language) (*TVShowInfo, error)

    // 通过提供方自身 ID 获取详情
    GetMovieByID(ctx context.Context, id string, lang Language) (*MovieInfo, error)
    GetTVByID(ctx context.Context, id string, lang Language) (*TVShowInfo, error)
}
```

> 业务层**只依赖该接口**，不感知底层具体是 TMDB、TVDB 还是豆瓣。

---

## 工厂用法

```go
factory := metadata.NewFactory()

// 注册具体实现（在初始化阶段完成）
factory.Register(tmdbClient)
factory.Register(tvdbClient)
factory.Register(doubanClient)

// 在业务代码中按名称获取
provider, ok := factory.Get(metadata.ProviderTMDB)
if !ok {
    // 处理未注册情况
}

movies, err := provider.SearchMovie(ctx, "Inception", metadata.SearchOptions{
    Language: "zh-CN",
    Year:     2010,
    Limit:    5,
})
```

---

## TMDB 客户端

**文件**：`metadata/tmdb/client.go`

### 支持的能力

- ✅ `SearchMovie` → `GET /search/movie`
- ✅ `GetMovieByTMDB` → `GET /movie/{id}`
- ✅ `GetMovieByID` → 复用 `GetMovieByTMDB`
- ✅ `SearchTV` → `GET /search/tv`
- ✅ `GetTVByTMDB` → `GET /tv/{id}`
- ✅ `GetTVByID` → 复用 `GetTVByTMDB`
- ⏳ 计划：补充剧集季/集详情（`/tv/{id}/season/{s}` 等）

### 配置与创建

```go
cfg := tmdb.Config{
    APIKey:  "<TMDB_API_KEY>",
    BaseURL: "https://api.themoviedb.org/3", // 可选
}
client, err := tmdb.NewClient(cfg)
```

---

## TVDB 客户端

**文件**：`metadata/tvdb/client.go`

当前定位：
- 以**剧集**为主的数据源
- 代码中预留了 `APIKey` / `JWTToken` 字段，便于后续对接 TVDB v4+ 登录流程

### 支持的能力

- ✅ `SearchTV` → `GET /search?q={keyword}&year={year?}`（最小实现）
- ✅ `GetTVByID` → `GET /series/{id}`（最小实现）
- ⏳ `GetTVByTMDB` → 占位实现（返回"暂未实现映射逻辑"）
- ❌ 电影相关方法默认返回"不支持"错误

### 配置与创建

```go
cfg := tvdb.Config{
    APIKey:  "<TVDB_API_KEY>",
    BaseURL: "https://api4.thetvdb.com/v4", // 可选
}
client, err := tvdb.NewClient(cfg)
```

> 后续可以在需要更精细剧集数据（季/集）时，再逐步接入 TVDB 更多 API。

---

## 豆瓣客户端

**文件**：`metadata/douban/client.go`

由于豆瓣官方 API 限制较多，建议通过**自建代理服务**间接访问。

### 支持的能力

- ✅ `SearchMovie` → `GET /search/movie?q={keyword}`（最小实现，假设代理提供）
- ✅ `GetMovieByID` → `GET /movie/{id}`（最小实现，假设代理提供）
- ⏳ `SearchTV` / `GetTVByID` → 占位实现（返回空或错误）
- ❌ `GetMovieByTMDB` / `GetTVByTMDB` → 返回"不支持通过 TMDB 直接查询"

### 配置与创建

```go
cfg := douban.Config{
    BaseURL: "https://api.your-proxy.com/douban", // 自建代理地址
}
client, err := douban.NewClient(cfg)
```

> 注意：豆瓣评分等字段暂未映射到 `MovieInfo` 中（当前接口无 `Rating` 字段），后续可扩展。

> 后续可以在代理层统一输出与 TMDB/TVDB 接近的数据结构，再在此处做简单映射。

---

## 与业务层的集成思路

1. **在启动时初始化所有 Provider 并注册到 Factory**：
   - TMDB：主电影/剧集来源
   - TVDB：补充剧集元数据（可选）
   - 豆瓣：做本地化评分/标签补充（可选）

2. **业务层使用统一接口**：
   - `SearchMovie` / `SearchTV` 查找候选
   - `GetMovieByTMDB` / `GetTVByTMDB` 做精确匹配
   - 根据需要组合多个 Provider 的结果（如 TMDB 基本信息 + 豆瓣评分）

3. **与媒体服务器联动**：
   - 媒体服务器返回的 TMDB/TVDB/IMDB ID
   - 使用 metadata Provider 拉取详情
   - 最终形成统一的领域对象，供订阅/下载/重命名等模块使用

---

## 下一步计划

- [ ] 为 TMDB 客户端补充错误重试和速率限制处理
- [ ] 为 metadata 模块增加单元测试（使用 http.RoundTripper mock）
- [ ] 在 docs/execution-plan.md 中补充 Week 5 metadata 集成说明
- [ ] 在业务 service 层增加元数据聚合服务（优先 TMDB，次选 TVDB/豆瓣）
