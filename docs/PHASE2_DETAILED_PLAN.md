# 第二阶段详细执行计划 (Week 5-10)

> **目标**: 实现订阅与下载链路，完成自动化工作流

---

## 概述

第二阶段将实现完整的订阅与下载功能，包括：
- 订阅系统 (RSS、站点监控)
- 下载器集成 (qBittorrent、Transmission)
- 自动化任务调度
- 通知系统 (Telegram、WeChat、Email)

---

## Week 5: 订阅系统基础

### Day 1-2: 订阅模型与 API

#### 任务清单

##### 1. 订阅数据模型
```go
// internal/models/subscribe.go
type Subscribe struct {
    BaseModel
    Name            string     `gorm:"size:500;not null;index" json:"name"`
    Year            *string    `gorm:"size:10" json:"year"`
    Type            string     `gorm:"size:20" json:"type"` // movie, tv
    Season          *int       `json:"season"`
    TMDBID          *int       `gorm:"index" json:"tmdb_id"`
    IMDBID          *string    `gorm:"size:20" json:"imdb_id"`
    
    // 订阅配置
    Quality         string     `gorm:"size:50" json:"quality"` // 1080p, 2160p
    Resolution      string     `gorm:"size:50" json:"resolution"`
    Source          string     `gorm:"size:50" json:"source"` // BluRay, WEB-DL
    
    // 过滤规则
    Include         string     `gorm:"type:text" json:"include"` // 包含关键词
    Exclude         string     `gorm:"type:text" json:"exclude"` // 排除关键词
    
    // 状态
    State           string     `gorm:"size:20;default:active" json:"state"` // active, paused, completed
    TotalEpisodes   *int       `json:"total_episodes"`
    CurrentEpisode  *int       `json:"current_episode"`
    
    // 时间
    LastUpdate      *time.Time `json:"last_update"`
    BestVersion     string     `gorm:"type:text" json:"best_version"`
}
```

##### 2. 订阅 Repository
```go
// internal/repository/subscribe_repository.go
type SubscribeRepository interface {
    Create(subscribe *models.Subscribe) error
    Update(subscribe *models.Subscribe) error
    Delete(id uint) error
    FindByID(id uint) (*models.Subscribe, error)
    FindAll(opts FindOptions) ([]models.Subscribe, error)
    FindActive() ([]models.Subscribe, error)
    FindByTMDBID(tmdbID int, mediaType string) (*models.Subscribe, error)
}
```

##### 3. 订阅 Service
```go
// internal/business/subscribe/service.go
type Service interface {
    CreateSubscribe(req CreateSubscribeRequest) (*models.Subscribe, error)
    UpdateSubscribe(id uint, req UpdateSubscribeRequest) error
    DeleteSubscribe(id uint) error
    GetSubscribe(id uint) (*models.Subscribe, error)
    ListSubscribes(opts ListOptions) ([]models.Subscribe, int64, error)
    PauseSubscribe(id uint) error
    ResumeSubscribe(id uint) error
}
```

##### 4. 订阅 API
```go
// internal/apis/subscribe/handler.go
type Handler struct {
    service subscribe.Service
    logger  *zap.Logger
}

// POST   /api/subscribes          - 创建订阅
// GET    /api/subscribes          - 列表
// GET    /api/subscribes/:id      - 详情
// PUT    /api/subscribes/:id      - 更新
// DELETE /api/subscribes/:id      - 删除
// POST   /api/subscribes/:id/pause   - 暂停
// POST   /api/subscribes/:id/resume  - 恢复
```

#### 验收标准
- [ ] 订阅模型创建完成
- [ ] Repository 实现完成
- [ ] Service 实现完成
- [ ] API 端点可用
- [ ] 单元测试通过

---

### Day 3-4: RSS 解析器

#### 任务清单

##### 1. RSS 解析器
```go
// pkg/rss/parser.go
type Parser struct {
    client  *http.Client
    cache   cache.Cache
    logger  *zap.Logger
}

type RSSFeed struct {
    Title       string
    Link        string
    Description string
    Items       []RSSItem
}

type RSSItem struct {
    Title       string
    Link        string
    Description string
    PubDate     time.Time
    Enclosure   Enclosure
    GUID        string
}

type Enclosure struct {
    URL    string
    Length int64
    Type   string
}

func (p *Parser) ParseURL(url string) (*RSSFeed, error)
func (p *Parser) ParseXML(data []byte) (*RSSFeed, error)
```

##### 2. Torrent 信息提取
```go
// pkg/rss/torrent.go
type TorrentInfo struct {
    Title      string
    Size       int64
    Seeders    int
    Leechers   int
    DownloadURL string
    MagnetLink string
    InfoHash   string
    
    // 解析出的媒体信息
    MediaTitle string
    Season     int
    Episode    int
    Quality    string
    Source     string
    Codec      string
    Group      string
}

func ExtractTorrentInfo(item RSSItem) (*TorrentInfo, error)
```

##### 3. RSS 订阅源管理
```go
// internal/business/subscribe/rss_source.go
type RSSSource struct {
    Name     string
    URL      string
    Interval time.Duration
    Enabled  bool
}

type RSSSourceManager struct {
    sources []RSSSource
    parser  *rss.Parser
    logger  *zap.Logger
}

func (m *RSSSourceManager) AddSource(source RSSSource) error
func (m *RSSSourceManager) RemoveSource(name string) error
func (m *RSSSourceManager) FetchAll() ([]rss.RSSItem, error)
func (m *RSSSourceManager) FetchSource(name string) ([]rss.RSSItem, error)
```

#### 验收标准
- [ ] RSS 解析器实现完成
- [ ] Torrent 信息提取正确
- [ ] 支持多个 RSS 源
- [ ] 缓存机制生效
- [ ] 测试通过

---

### Day 5-6: 订阅匹配引擎

#### 任务清单

##### 1. 匹配规则
```go
// internal/business/subscribe/matcher.go
type Matcher struct {
    logger *zap.Logger
}

type MatchRule struct {
    Subscribe   *models.Subscribe
    Quality     []string  // 质量优先级
    Source      []string  // 来源优先级
    Include     []string  // 包含关键词
    Exclude     []string  // 排除关键词
    MinSize     int64     // 最小文件大小
    MaxSize     int64     // 最大文件大小
}

func (m *Matcher) Match(torrent *rss.TorrentInfo, rule MatchRule) (bool, int)
func (m *Matcher) SelectBest(torrents []*rss.TorrentInfo, rule MatchRule) *rss.TorrentInfo
```

##### 2. 匹配评分系统
```go
type MatchScore struct {
    Total       int
    QualityScore int
    SourceScore  int
    SizeScore    int
    SeedScore    int
}

func CalculateScore(torrent *rss.TorrentInfo, rule MatchRule) MatchScore
```

##### 3. 订阅扫描器
```go
// internal/business/subscribe/scanner.go
type Scanner struct {
    rssManager *RSSSourceManager
    matcher    *Matcher
    repo       repository.SubscribeRepository
    logger     *zap.Logger
}

func (s *Scanner) ScanAll() ([]MatchResult, error)
func (s *Scanner) ScanSubscribe(subscribe *models.Subscribe) ([]MatchResult, error)
```

#### 验收标准
- [ ] 匹配规则实现完成
- [ ] 评分系统合理
- [ ] 能正确选择最佳资源
- [ ] 测试覆盖完整

---

### Day 7: 测试与优化

#### 测试场景
```go
func TestSubscribe_Create(t *testing.T) {
    // 测试创建订阅
}

func TestRSS_Parse(t *testing.T) {
    // 测试 RSS 解析
}

func TestMatcher_Match(t *testing.T) {
    // 测试匹配规则
}

func TestMatcher_SelectBest(t *testing.T) {
    // 测试最佳选择
}
```

#### 验收标准
- [ ] 所有测试通过
- [ ] 测试覆盖率 > 70%
- [ ] 性能测试通过

---

## Week 6: 下载器集成

### Day 1-3: qBittorrent 集成

#### 任务清单

##### 1. qBittorrent 客户端
```go
// pkg/downloader/qbittorrent/client.go
type Client struct {
    baseURL  string
    username string
    password string
    cookie   string
    client   *http.Client
    logger   *zap.Logger
}

func NewClient(config Config) (*Client, error)
func (c *Client) Login() error
func (c *Client) AddTorrent(opts AddTorrentOptions) (string, error)
func (c *Client) GetTorrents(filter string) ([]Torrent, error)
func (c *Client) GetTorrentInfo(hash string) (*TorrentInfo, error)
func (c *Client) DeleteTorrent(hash string, deleteFiles bool) error
func (c *Client) PauseTorrent(hash string) error
func (c *Client) ResumeTorrent(hash string) error
```

##### 2. Torrent 状态管理
```go
type TorrentStatus string

const (
    StatusDownloading TorrentStatus = "downloading"
    StatusPaused      TorrentStatus = "paused"
    StatusCompleted   TorrentStatus = "completed"
    StatusError       TorrentStatus = "error"
)

type Torrent struct {
    Hash        string
    Name        string
    Size        int64
    Progress    float64
    Status      TorrentStatus
    DownloadSpeed int64
    UploadSpeed   int64
    ETA         int64
    SavePath    string
}
```

##### 3. 下载任务管理
```go
// internal/business/download/service.go
type Service interface {
    AddDownload(req AddDownloadRequest) (*models.Download, error)
    GetDownload(id uint) (*models.Download, error)
    ListDownloads(opts ListOptions) ([]models.Download, int64, error)
    PauseDownload(id uint) error
    ResumeDownload(id uint) error
    DeleteDownload(id uint, deleteFiles bool) error
    SyncStatus() error
}
```

#### 验收标准
- [ ] qBittorrent 客户端实现完成
- [ ] 支持所有基本操作
- [ ] 状态同步正常
- [ ] 错误处理完善

---

### Day 4-5: Transmission 集成

#### 任务清单

##### 1. Transmission 客户端
```go
// pkg/downloader/transmission/client.go
type Client struct {
    baseURL    string
    username   string
    password   string
    sessionID  string
    client     *http.Client
    logger     *zap.Logger
}

// 实现与 qBittorrent 相同的接口
```

##### 2. 下载器抽象接口
```go
// pkg/downloader/interface.go
type Downloader interface {
    AddTorrent(opts AddTorrentOptions) (string, error)
    GetTorrents(filter string) ([]Torrent, error)
    GetTorrentInfo(hash string) (*TorrentInfo, error)
    DeleteTorrent(hash string, deleteFiles bool) error
    PauseTorrent(hash string) error
    ResumeTorrent(hash string) error
}
```

##### 3. 下载器工厂
```go
// pkg/downloader/factory.go
type Factory struct {
    config Config
    logger *zap.Logger
}

func (f *Factory) CreateDownloader(downloaderType string) (Downloader, error)
```

#### 验收标准
- [ ] Transmission 客户端实现完成
- [ ] 接口统一
- [ ] 支持多下载器切换

---

### Day 6-7: 下载监控

#### 任务清单

##### 1. 下载监控器
```go
// internal/business/download/monitor.go
type Monitor struct {
    downloader downloader.Downloader
    repo       repository.DownloadRepository
    interval   time.Duration
    logger     *zap.Logger
}

func (m *Monitor) Start() error
func (m *Monitor) Stop() error
func (m *Monitor) SyncStatus() error
func (m *Monitor) HandleCompleted(torrent *downloader.Torrent) error
```

##### 2. 下载完成处理
```go
type CompletionHandler struct {
    transferService transfer.Service
    mediaService    media.Service
    logger          *zap.Logger
}

func (h *CompletionHandler) Handle(download *models.Download) error {
    // 1. 识别媒体信息
    // 2. 转移文件
    // 3. 更新订阅状态
    // 4. 发送通知
}
```

#### 验收标准
- [ ] 监控器正常运行
- [ ] 状态同步准确
- [ ] 完成后自动处理

---

## Week 7: 自动化任务调度

### Day 1-3: 定时任务系统

#### 任务清单

##### 1. 任务调度器
```go
// internal/scheduler/scheduler.go
type Scheduler struct {
    cron   *cron.Cron
    tasks  map[string]Task
    logger *zap.Logger
}

type Task interface {
    Name() string
    Schedule() string  // Cron 表达式
    Execute(ctx context.Context) error
}

func (s *Scheduler) AddTask(task Task) error
func (s *Scheduler) RemoveTask(name string) error
func (s *Scheduler) Start() error
func (s *Scheduler) Stop() error
```

##### 2. 订阅扫描任务
```go
// internal/scheduler/tasks/subscribe_scan.go
type SubscribeScanTask struct {
    scanner *subscribe.Scanner
    logger  *zap.Logger
}

func (t *SubscribeScanTask) Execute(ctx context.Context) error {
    // 1. 扫描所有活跃订阅
    // 2. 匹配新资源
    // 3. 添加到下载队列
}
```

##### 3. 下载监控任务
```go
// internal/scheduler/tasks/download_monitor.go
type DownloadMonitorTask struct {
    monitor *download.Monitor
    logger  *zap.Logger
}

func (t *DownloadMonitorTask) Execute(ctx context.Context) error {
    // 同步下载状态
}
```

#### 验收标准
- [ ] 调度器实现完成
- [ ] 支持 Cron 表达式
- [ ] 任务执行正常
- [ ] 错误处理完善

---

### Day 4-5: 自动化工作流

#### 任务清单

##### 1. 订阅工作流
```
订阅扫描 → 资源匹配 → 添加下载 → 下载监控 → 完成处理 → 发送通知
```

##### 2. 工作流编排
```go
// internal/workflow/subscribe_workflow.go
type SubscribeWorkflow struct {
    scanner   *subscribe.Scanner
    matcher   *subscribe.Matcher
    downloader download.Service
    handler   *download.CompletionHandler
    logger    *zap.Logger
}

func (w *SubscribeWorkflow) Execute(ctx context.Context) error
```

#### 验收标准
- [ ] 工作流完整
- [ ] 各环节衔接正常
- [ ] 异常处理完善

---

### Day 6-7: 测试与优化

#### 集成测试
```go
func TestSubscribeWorkflow_EndToEnd(t *testing.T) {
    // 1. 创建订阅
    // 2. 模拟 RSS 更新
    // 3. 验证匹配结果
    // 4. 验证下载添加
    // 5. 模拟下载完成
    // 6. 验证文件转移
}
```

---

## Week 8-10: 通知系统与完善

### Week 8: 通知系统

#### 任务清单

##### 1. 通知接口
```go
// pkg/notification/interface.go
type Notifier interface {
    Send(message Message) error
}

type Message struct {
    Title   string
    Content string
    Level   Level
    Data    map[string]interface{}
}
```

##### 2. Telegram 通知
```go
// pkg/notification/telegram/client.go
type Client struct {
    botToken string
    chatID   string
    client   *http.Client
}
```

##### 3. 通知管理器
```go
// internal/business/notification/manager.go
type Manager struct {
    notifiers []notification.Notifier
    logger    *zap.Logger
}

func (m *Manager) Notify(event Event) error
```

---

### Week 9: 完善与优化

#### 任务清单
- [ ] 性能优化
- [ ] 错误处理完善
- [ ] 日志优化
- [ ] 监控指标添加

---

### Week 10: 测试与文档

#### 任务清单
- [ ] 集成测试
- [ ] 压力测试
- [ ] API 文档
- [ ] 用户手册
- [ ] 部署文档

---

## 验收标准总结

### 功能验收
- [ ] 订阅系统完整可用
- [ ] 下载器集成正常
- [ ] 自动化工作流运行正常
- [ ] 通知系统可用

### 质量验收
- [ ] 测试覆盖率 > 70%
- [ ] 所有测试通过
- [ ] 性能达标
- [ ] 文档完整

---

## 下一步

第二阶段完成后，进入第三阶段: 插件系统与扩展 (Week 11-15)
