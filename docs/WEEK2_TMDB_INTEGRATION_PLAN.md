# Week 2 TMDB集成开发计划

> **开发周期**: Week 2 (2024-11-23 - 2024-11-29)  
> **主要目标**: 完成TMDB API集成，实现智能媒体识别和刮削功能

---

## 📋 当前状态评估

### ✅ 已完成功能
- TMDB客户端基础框架 (`internal/business/media/tmdb/client.go`)
- 基础数据类型定义 (`internal/business/media/tmdb/types.go`)
- 图片下载功能 (`internal/business/media/tmdb/image.go`)
- TMDB服务层基础结构 (`internal/business/media/tmdb_service.go`)
- 基础搜索功能：SearchMovie, SearchTV, SearchMulti
- 基础详情功能：GetMovieDetails, GetTVDetails, GetSeasonDetails, GetEpisodeDetails
- 演职员功能：GetMovieCredits
- 趋势功能：GetTrending

### ❌ 待实现功能
- GetMovieImages/GetTVImages (图片获取)
- GetTVCredits (电视剧演职员)
- 更完善的缓存策略
- 限流机制优化
- 错误重试机制
- 文件名解析与TMDB匹配
- 智能媒体识别逻辑

---

## 🎯 Week 2 开发目标

### Day 1-2: TMDB API功能完善

#### 1. 图片API实现
```go
// 需要实现的方法
func (c *Client) GetMovieImages(ctx context.Context, id int) (*Images, error)
func (c *Client) GetTVImages(ctx context.Context, id int) (*Images, error)
func (c *Client) GetSeasonImages(ctx context.Context, tvID, seasonNum int) (*Images, error)
```

#### 2. 演职员API完善
```go
// 需要实现的方法
func (c *Client) GetTVCredits(ctx context.Context, id int) (*Credits, error)
func (c *Client) GetSeasonCredits(ctx context.Context, tvID, seasonNum int) (*Credits, error)
func (c *Client) GetEpisodeCredits(ctx context.Context, tvID, seasonNum, episodeNum int) (*Credits, error)
```

#### 3. 发现API实现
```go
// 需要实现的方法
func (c *Client) DiscoverMovies(ctx context.Context, params DiscoverParams) (*MovieSearchResponse, error)
func (c *Client) DiscoverTV(ctx context.Context, params DiscoverParams) (*TVSearchResponse, error)
```

### Day 3-4: 智能媒体识别

#### 1. 文件名解析器
```go
// 需要实现的结构体和方法
type FileNameParser struct {
    patterns []string
    logger   *zap.Logger
}

func (p *FileNameParser) Parse(fileName string) (*ParseResult, error)
func (p *FileNameParser) ExtractYear(fileName string) (int, error)
func (p *FileNameParser) ExtractSeasonEpisode(fileName string) (season, episode int, err error)
```

#### 2. TMDB匹配器
```go
// 需要实现的结构体和方法
type TMDBMatcher struct {
    client   *tmdb.Client
    cache    cache.Cache
    logger   *zap.Logger
}

func (m *TMDBMatcher) MatchMovie(parsed *ParseResult) (*MovieMatch, error)
func (m *TMDBMatcher) MatchTV(parsed *ParseResult) (*TVMatch, error)
func (m *TMDBMatcher) CalculateSimilarity(str1, str2 string) float64
```

### Day 5: 集成测试与优化

#### 1. 性能优化
- 实现连接池
- 优化缓存策略
- 实现批量请求

#### 2. 错误处理完善
- 实现指数退避重试
- 添加详细的错误分类
- 完善日志记录

#### 3. 集成测试
- 端到端测试文件识别流程
- 性能基准测试
- 边界条件测试

---

## 📁 需要创建的文件

### 1. TMDB API扩展
```
internal/business/media/tmdb/
├── images.go          # 图片API实现
├── credits.go         # 演职员API实现  
├── discover.go        # 发现API实现
├── collection.go      # 合集API实现
└── network.go         # 电视网API实现
```

### 2. 媒体识别模块
```
internal/business/media/
├── parser.go          # 文件名解析器
├── matcher.go         # TMDB匹配器
├── identifier.go      # 媒体识别器
└── scorer.go          # 相似度评分
```

### 3. 测试文件
```
tests/business/media/
├── tmdb_client_test.go      # TMDB客户端测试
├── tmdb_images_test.go      # 图片功能测试
├── parser_test.go           # 文件名解析测试
├── matcher_test.go          # 匹配器测试
└── integration_test.go      # 集成测试
```

---

## 🔧 技术实现细节

### 1. 缓存策略优化
```go
// 缓存层级
1. L1: 内存缓存 (sync.Map) - 热点数据，5分钟TTL
2. L2: Redis缓存 - 共享数据，按类型设置TTL
   - 搜索结果: 1小时
   - 详情信息: 24小时  
   - 图片URL: 7天
   - 演职员: 24小时
```

### 2. 限流机制
```go
// TMDB API限制: 40 req/10s
type RateLimiter struct {
    limiter *rate.Limiter
    logger  *zap.Logger
}

// 实现令牌桶算法
limiter := rate.NewLimiter(rate.Every(250*time.Millisecond), 10)
```

### 3. 重试策略
```go
// 指数退避重试
type RetryConfig struct {
    MaxRetries int
    BaseDelay  time.Duration
    MaxDelay   time.Duration
}

// 重试逻辑
for attempt := 0; attempt < config.MaxRetries; attempt++ {
    err := doRequest()
    if err == nil || !isRetryable(err) {
        break
    }
    delay := calculateBackoff(attempt)
    time.Sleep(delay)
}
```

---

## 📊 验收标准

### 1. 功能验收
- [ ] 可以搜索电影/电视剧并获取完整信息
- [ ] 可以下载海报、背景图等图片资源
- [ ] 可以获取演职员信息
- [ ] 可以智能解析文件名并匹配TMDB内容
- [ ] 缓存和限流机制正常工作

### 2. 性能验收
- [ ] 搜索响应时间 < 2秒
- [ ] 详情获取响应时间 < 1秒
- [ ] 图片下载成功率 > 95%
- [ ] 缓存命中率 > 80%

### 3. 稳定性验收
- [ ] API调用失败自动重试
- [ ] 网络异常优雅降级
- [ ] 内存使用稳定，无泄漏
- [ ] 并发安全

---

## 🚀 开发优先级

### P0 (必须完成)
1. GetMovieImages/GetTVImages 实现
2. 文件名解析器
3. TMDB匹配器
4. 基础集成测试

### P1 (重要)
1. 演职员API完善
2. 发现API实现
3. 缓存策略优化
4. 性能测试

### P2 (可选)
1. 合集API
2. 电视网API
3. 高级匹配算法
4. 批量操作优化

---

## 📈 进度跟踪

### Day 1 (11-23)
- [ ] 实现GetMovieImages/GetTVImages
- [ ] 创建图片相关测试
- [ ] 验证图片下载功能

### Day 2 (11-24)  
- [ ] 实现GetTVCredits等演职员API
- [ ] 创建演职员相关测试
- [ ] 验证演职员信息获取

### Day 3 (11-25)
- [ ] 实现文件名解析器
- [ ] 创建解析器测试
- [ ] 验证各种文件名格式

### Day 4 (11-26)
- [ ] 实现TMDB匹配器
- [ ] 创建匹配器测试
- [ ] 验证匹配准确性

### Day 5 (11-27)
- [ ] 集成测试
- [ ] 性能优化
- [ ] 文档更新

### Day 6-7 (11-28/29)
- [ ] 代码审查
- [ ] Bug修复
- [ ] 准备Week 3计划

---

## 🎯 成功指标

### 技术指标
- API功能完整度: 100%
- 测试覆盖率: >90%
- 性能基准达标率: 100%

### 业务指标
- 媒体识别准确率: >90%
- 用户满意度: >95%
- 系统稳定性: 99.9%

---

**创建时间**: 2024-11-22  
**负责人**: AI Assistant  
**状态**: 📋 计划就绪