# 第一阶段详细执行计划 (Week 1-4)

> **目标**: 完善"本地文件 → 刮削 → 转移"链路，实现 MVP

---

## Week 1: API 入口与测试完善

### Day 1-2: API Handler 实现 ✅

#### 任务清单
- [x] 完成 `internal/apis/workflow/handler.go`
  ```go
  type Handler struct {
      service workflowapi.WorkflowService
      logger  *zap.Logger
  }
  
  func (h *Handler) StartLocalFileWorkflow(c *gin.Context) {
      // 1. validator.BindAndValidate
      // 2. 使用 logger.WithContext 记录请求上下文
      // 3. 调用 service.StartLocalFileWorkflow
      // 4. 根据 wait_for_completion 选择 202/200
  }
  ```

- [x] 实现请求参数校验
  ```go
  type StartLocalFileWorkflowRequest struct {
      RootPath      string   `json:"root_path" binding:"required"`
      Include       []string `json:"include" binding:"omitempty,dive,min=1"`
      Exclude       []string `json:"exclude" binding:"omitempty,dive,min=1"`
      MaxDepth      int      `json:"max_depth" binding:"gte=0"`
      FollowSymlink bool     `json:"follow_symlink"`

      TargetRoot  string `json:"target_root" binding:"required"`
      Mode        string `json:"mode" binding:"omitempty,oneof=move copy link hardlink softlink"`
      Category    string `json:"category" binding:"omitempty,max=64"`
      Overwrite   bool   `json:"overwrite"`
      PreserveDir bool   `json:"preserve_dir"`
      DryRun      bool   `json:"dry_run"`

      IncludeFetch      bool     `json:"include_fetch"`
      FetchKeywords     []string `json:"fetch_keywords" binding:"omitempty,dive,min=1"`
      WaitForCompletion bool     `json:"wait_for_completion"`

      ForceRefresh bool   `json:"force_refresh"`
      Source       string `json:"source" binding:"omitempty,max=32"`
  }
  ```

- [x] 实现响应格式
  ```go
  type StartLocalFileWorkflowResponse struct {
      WorkflowID string      `json:"workflow_id"`
      Status     string      `json:"status"`
      Message    string      `json:"message,omitempty"`
      Result     interface{} `json:"result,omitempty"`
  }
  ```

- [x] 支持同步/异步两种模式
  - 异步: 返回 202 + workflow_id（默认路径）
  - 同步: `wait_for_completion=true` 时等待 WorkflowManager，200 + 结果
  - 统一通过 `pkg/response` 包装成功/错误响应

#### 验收标准
- [x] API 可以接收请求并返回正确格式（`tests/api/local_workflow_test.go` 验证）
- [x] 参数校验正确 (必填字段、枚举值)
- [x] 错误处理统一 (400/500 错误码，`response.ErrorWithLog`)

#### 测试覆盖现状（截至 2024-11-22）
- [x] `tests/api/local_workflow_test.go`：验证 Handler + Service + WorkflowManager happy path（同步等待 + 日志输出）
- [ ] `tests/api/workflow_handler_test.go`：待补全参数校验失败、异步模式等单元测试
- [ ] `tests/actions/*_test.go`：Action/Business 层仍为规划阶段，需在 Week1 Day4-5 开始实现
- [ ] `tests/integration/local_workflow_test.go`：端到端测试脚手架写在计划中，待 Week1 Day6-7 执行

---

### Day 3: 路由注册与中间件 （部分完成）

#### 任务清单
- [x] 在 `cmd/server/main.go` 中初始化 WorkflowManager / Service / Handler
  ```go
  zapLogger := logger.GetLogger()

  workflowManager := wf.NewWorkflowManager(zapLogger)
  workflowService := workflowapi.NewService(
      workflowManager,
      wf.LocalFileWorkflowConfig{Logger: zapLogger},
      zapLogger,
  )
  workflowHandler := workflowapi.NewHandler(workflowService, zapLogger)
  ```

- [x] 注册路由
  ```go
  engine.POST("/api/workflows/local-file-scrape-transfer", workflowHandler.StartLocalFileWorkflow)
  ```

- [ ] 添加中间件
  - [ ] 请求日志中间件
  - [ ] CORS 中间件
  - [ ] 认证中间件 (JWT)
  - [ ] 限流中间件
  - [ ] 错误恢复中间件

#### 验收标准
- [ ] 路由可访问
- [ ] 中间件生效 (日志、CORS、认证)
- [ ] 健康检查端点正常: `GET /health`

#### 中间件实施计划
- [ ] **请求 ID + 结构化日志**：计划使用 `pkg/middlewares/requestid`（待实现）为每次请求注入 `request_id`，配合 `pkg/logger` 输出
- [ ] **统一恢复**：封装 `RecoveryMiddleware` 捕获 panic，输出 JSON 错误并记录 `stacktrace`
- [ ] **CORS**：根据 `config/server` 配置允许前端域名访问，默认允许本地开发域
- [ ] **认证**：在 `/api/**` 挂载 JWT 校验中间件，Workflow API 可保留白名单（用于 CLI）
- [ ] **限流**：结合 `pkg/utils/redis` 实现令牌桶限流，优先保护 Workflow API，参数来自 `config.rate_limit`

---

### Day 4-5: 单元测试

#### 测试范围

##### 1. Actions 层测试
```go
// tests/actions/scan_file_action_test.go
func TestScanFileAction_Execute(t *testing.T) {
    // 测试正常扫描
    // 测试空目录
    // 测试无权限目录
    // 测试包含/排除规则
}

// tests/actions/scrape_file_action_test.go
func TestScrapeFileAction_Execute(t *testing.T) {
    // 测试有效文件名
    // 测试模糊文件名
    // 测试异常文件名
}

// tests/actions/transfer_file_action_test.go
func TestTransferFileAction_Execute(t *testing.T) {
    // 测试正常转移
    // 测试目标冲突
    // 测试无权限场景
    // 测试 dry-run 模式
}
```

##### 2. Business Services 测试
```go
// tests/business/storage_service_test.go
func TestStorageService_Scan(t *testing.T) {}
func TestStorageService_Transfer(t *testing.T) {}

// tests/business/media_service_test.go
func TestMediaService_Identify(t *testing.T) {}

// tests/business/transfer_service_test.go
func TestTransferService_Execute(t *testing.T) {}
```

##### 3. API Handler 测试
```go
// tests/api/workflow_handler_test.go
func TestWorkflowHandler_StartLocalFileWorkflow(t *testing.T) {
    // Mock WorkflowManager
    // 测试参数校验失败 (400)
    // 测试正常请求 (202/200)
    // 测试异步模式
    // 测试同步模式
}
```

#### 验收标准
- [ ] 测试覆盖率 > 70%
- [ ] 所有测试通过
- [ ] 测试数据自动清理

---

### Day 6-7: 集成测试

#### 测试场景
```go
// tests/integration/local_workflow_test.go
func TestLocalFileWorkflow_EndToEnd(t *testing.T) {
    // 1. 准备测试数据
    testDir := createTestFiles(t, []string{
        "Movie.2023.1080p.mkv",
        "TV.Show.S01E01.mkv",
    })
    defer cleanupTestFiles(t, testDir)
    
    // 2. 发送 API 请求
    req := StartLocalFileWorkflowRequest{
        RootPath:   testDir,
        TargetRoot: "/tmp/target",
        Mode:       "copy",
    }
    
    resp := callAPI(t, req)
    assert.Equal(t, 202, resp.StatusCode)
    
    // 3. 等待完成
    waitForWorkflow(t, resp.WorkflowID, 30*time.Second)
    
    // 4. 验证结果
    result := getWorkflowResult(t, resp.WorkflowID)
    assert.Equal(t, "completed", result.Status)
    assert.Equal(t, 2, len(result.Transfers))
    
    // 5. 验证文件已转移
    assertFileExists(t, "/tmp/target/Movie (2023)/Movie.2023.1080p.mkv")
    assertFileExists(t, "/tmp/target/TV Show/Season 01/TV.Show.S01E01.mkv")
}
```

#### 验收标准
- [ ] 端到端测试通过
- [ ] 文件正确转移
- [ ] 数据库记录正确
- [ ] 日志完整

---

## Week 2: 真实刮削能力接入 ✅

### Day 1-3: TMDB 集成 ✅

#### 任务清单

##### 1. 迁移 TMDB 模块结构 ✅
```
internal/business/media/tmdb/
├── client.go          # TMDB API 客户端
├── movie.go           # 电影相关 API
├── tv.go              # 电视剧相关 API
├── search.go          # 搜索 API
├── person.go          # 人物 API
├── collection.go      # 合集 API
├── discover.go        # 发现 API
├── trending.go        # 趋势 API
├── cache.go           # 缓存策略
└── types.go           # 数据类型
```

##### 2. 实现核心功能 ✅
- [x] **TMDB Client**
  ```go
  type Client struct {
      apiKey     string
      baseURL    string
      httpClient *httpclient.Client
      cache      cache.Cache
      limiter    *rate.Limiter
  }
  
  func NewClient(apiKey string, cache cache.Cache) *Client
  ```

- [x] **电影搜索**
  ```go
  func (c *Client) SearchMovie(query string, year int) ([]Movie, error)
  func (c *Client) GetMovieDetails(id int) (*MovieDetails, error)
  func (c *Client) GetMovieCredits(id int) (*Credits, error)
  func (c *Client) GetMovieImages(id int) (*Images, error)
  ```

- [x] **电视剧搜索**
  ```go
  func (c *Client) SearchTV(query string, year int) ([]TVShow, error)
  func (c *Client) GetTVDetails(id int) (*TVDetails, error)
  func (c *Client) GetSeasonDetails(tvID, seasonNum int) (*Season, error)
  func (c *Client) GetEpisodeDetails(tvID, seasonNum, episodeNum int) (*Episode, error)
  ```

##### 3. 缓存策略 ✅
- [x] 搜索结果缓存 (1小时)
- [x] 详情缓存 (24小时)
- [x] 图片 URL 缓存 (7天)

##### 4. 限流与重试 ✅
- [x] 限流: 40 req/10s (TMDB API 限制)
- [x] 重试: 3次，指数退避

#### 验收标准 ✅
- [x] 可以搜索电影/电视剧
- [x] 可以获取详情信息
- [x] 缓存生效
- [x] 限流生效
- [x] 错误处理正确

---

### Day 4-5: 元数据识别增强 ✅

#### 任务清单

##### 1. 实现 TmdbService ✅
```go
// internal/business/media/tmdb_service.go
type TmdbService struct {
    client *tmdb.Client
    cache  cache.Cache
    logger *logger.Logger
}

func (s *TmdbService) Identify(files []FileItem, opts IdentifyOptions) ([]models.Media, error) {
    var medias []models.Media
    
    for _, file := range files {
        // 1. 解析文件名
        meta := parseFileName(file.Path)
        
        // 2. 搜索 TMDB
        results, err := s.searchTMDB(meta)
        if err != nil {
            continue
        }
        
        // 3. 选择最佳匹配
        best := selectBestMatch(results, meta)
        
        // 4. 获取详情
        details, err := s.getDetails(best.ID, meta.Type)
        if err != nil {
            continue
        }
        
        // 5. 转换为 Media 模型
        media := convertToMedia(details, file)
        medias = append(medias, media)
    }
    
    return medias, nil
}
```

##### 2. 文件名解析优化 ✅
- [x] 支持常见命名格式
  - `Movie.Title.2023.1080p.BluRay.mkv`
  - `TV.Show.S01E01.1080p.WEB-DL.mkv`
  - `[Group] Anime Title - 01 [1080p].mkv`
- [x] 提取关键信息
  - 标题、年份、季、集
  - 分辨率、来源、编码
  - 字幕、音频、发布组

##### 3. 匹配策略 ✅
- [x] 精确匹配 (标题 + 年份)
- [x] 模糊匹配 (相似度 > 80%)
- [x] 多结果时选择最佳 (评分、流行度)

#### 验收标准 ✅
- [x] 识别准确率 > 90%
- [x] 支持多种命名格式
- [x] 匹配逻辑合理

---

### Day 6: NFO 支持 ✅

#### 任务清单
- [x] 读取现有 NFO 文件
  ```go
  func ReadNFO(path string) (*NFOData, error)
  ```

- [x] 生成 NFO 文件
  ```go
  func WriteNFO(media *models.Media, path string) error
  ```

- [x] 支持多种格式
  - Kodi/XBMC
  - Emby
  - Jellyfin

#### 验收标准 ✅
- [x] 可以读取 NFO
- [x] 可以生成 NFO
- [x] 格式正确

---

### Day 7: 测试与验证 ✅

#### 测试数据
准备真实电影/剧集文件名进行测试:
```
测试集1: 电影
- The.Matrix.1999.1080p.BluRay.x264.mkv ✅
- Inception.2010.2160p.UHD.BluRay.x265.mkv ✅
- 肖申克的救赎.The.Shawshank.Redemption.1994.mkv ✅

测试集2: 电视剧
- Breaking.Bad.S01E01.1080p.WEB-DL.mkv ✅
- Game.of.Thrones.S08E06.FINAL.1080p.mkv ✅
- 权力的游戏.Game.of.Thrones.S01E01.mkv ✅

测试集3: 动漫
- [SubsPlease] Demon Slayer - 01 [1080p].mkv ✅
- 进击的巨人.Attack.on.Titan.S04E01.mkv ✅
```

#### 验收标准 ✅
- [x] 所有测试文件正确识别
- [x] 多语言支持 (中英文)
- [x] 错误处理正确

#### 完成报告
详见: [WEEK2_COMPLETION_REPORT.md](./WEEK2_COMPLETION_REPORT.md)

---

## Week 3: 转移能力精细化 ✅

### Day 1-3: 命名规则引擎 ✅

#### 任务清单

##### 1. 模板引擎实现
```go
// pkg/utils/naming.go
type NamingTemplate struct {
    template string
    vars     map[string]string
}

func ParseTemplate(template string) (*NamingTemplate, error)
func (t *NamingTemplate) Render(media *models.Media) (string, error)
```

##### 2. 支持的变量
```
电影:
- ${title}          # 标题
- ${year}           # 年份
- ${resolution}     # 分辨率
- ${source}         # 来源
- ${codec}          # 编码
- ${audio}          # 音频
- ${subtitle}       # 字幕
- ${group}          # 发布组

电视剧:
- ${title}          # 剧名
- ${season}         # 季 (S01)
- ${season_num}     # 季号 (1)
- ${episode}        # 集 (E01)
- ${episode_num}    # 集号 (1)
- ${episode_title}  # 集标题
- ${year}           # 年份
```

##### 3. 默认模板
```
电影: ${title} (${year})/${title}.${year}.${resolution}.${source}.${codec}${ext}
电视剧: ${title}/Season ${season_num}/${title}.${season}${episode}.${episode_title}${ext}
动漫: ${title}/Season ${season_num}/[${group}] ${title} - ${episode_num}${ext}
```

##### 4. 特殊字符处理 ✅
- [x] 移除非法字符: `<>:"/\|?*`
- [x] 替换空格: `_` 或 `.`
- [x] 文件名长度限制: 255 字符

#### 验收标准 ✅
- [x] 模板解析正确
- [x] 变量替换正确
- [x] 特殊字符处理正确
- [x] 支持自定义模板

---

### Day 4-5: 目录结构策略 ✅

#### 任务清单

##### 1. 目录生成器
```go
// internal/business/transfer/directory.go
type DirectoryStrategy interface {
    GeneratePath(media *models.Media, file *FileItem) (string, error)
}

type MovieStrategy struct {
    template string
}

type TVStrategy struct {
    template string
}

type AnimeStrategy struct {
    template string
}
```

##### 2. 实现策略
```go
// 电影策略
func (s *MovieStrategy) GeneratePath(media *models.Media, file *FileItem) (string, error) {
    // Movies/Title (Year)/Title.Year.Resolution.mkv
    return fmt.Sprintf("Movies/%s (%d)/%s", media.Title, media.Year, fileName), nil
}

// 电视剧策略
func (s *TVStrategy) GeneratePath(media *models.Media, file *FileItem) (string, error) {
    // TV/Show/Season XX/Show.SXXEXX.mkv
    return fmt.Sprintf("TV/%s/Season %02d/%s", media.Title, season, fileName), nil
}
```

##### 3. 配置支持 ✅
```yaml
transfer:
  movie_template: "Movies/${title} (${year})/${title}.${year}.${resolution}${ext}"
  tv_template: "TV/${title}/Season ${season_num}/${title}.${season}${episode}${ext}"
  anime_template: "Anime/${title}/Season ${season_num}/[${group}] ${title} - ${episode_num}${ext}"
```

#### 验收标准 ✅
- [x] 目录结构正确
- [x] 支持多种策略
- [x] 配置生效

---

### Day 6: 冲突处理 ✅

#### 任务清单

##### 1. 冲突检测
```go
func (s *TransferService) checkConflict(targetPath string) (ConflictType, error) {
    if !fileExists(targetPath) {
        return NoConflict, nil
    }
    
    // 比较文件大小、修改时间、MD5
    if isSameFile(sourcePath, targetPath) {
        return SameFile, nil
    }
    
    return DifferentFile, nil
}
```

##### 2. 冲突策略
```go
type ConflictStrategy string

const (
    StrategyOverwrite ConflictStrategy = "overwrite"  // 覆盖
    StrategySkip      ConflictStrategy = "skip"       // 跳过
    StrategyRename    ConflictStrategy = "rename"     // 重命名
)

func (s *TransferService) handleConflict(strategy ConflictStrategy, targetPath string) (string, error) {
    switch strategy {
    case StrategyOverwrite:
        return targetPath, nil
    case StrategySkip:
        return "", ErrSkipped
    case StrategyRename:
        return generateNewName(targetPath), nil
    }
}
```

##### 3. 文件完整性校验 ✅
```go
func verifyTransfer(source, target string) error {
    // 比较文件大小
    if getFileSize(source) != getFileSize(target) {
        return ErrSizeMismatch
    }
    
    // 可选: 计算 MD5/SHA256
    if opts.VerifyChecksum {
        if getMD5(source) != getMD5(target) {
            return ErrChecksumMismatch
        }
    }
    
    return nil
}
```

#### 验收标准 ✅
- [x] 冲突检测正确
- [x] 各种策略生效
- [x] 文件完整性校验正确

---

### Day 7: 测试 ✅

#### 测试场景 ✅
```go
func TestTransferWithNaming(t *testing.T) {
    // 测试各种命名规则 ✅
}

func TestTransferWithConflict(t *testing.T) {
    // 测试冲突处理 ✅
}

func TestTransferWithVerification(t *testing.T) {
    // 测试文件校验 ✅
}
```

#### 完成报告
详见: [WEEK3_COMPLETION_REPORT.md](./WEEK3_COMPLETION_REPORT.md)

---

## Week 4: 监控与优化 ✅

### Day 1-2: 性能优化 ✅

#### 任务清单

##### 1. 并发扫描
```go
func (s *StorageService) ScanConcurrent(opts ScanOptions) ([]FileItem, error) {
    var (
        wg    sync.WaitGroup
        mu    sync.Mutex
        files []FileItem
    )
    
    // 使用 Goroutine 池
    pool := workerpool.New(opts.Concurrency)
    
    filepath.WalkDir(opts.RootPath, func(path string, d fs.DirEntry, err error) error {
        if d.IsDir() {
            return nil
        }
        
        wg.Add(1)
        pool.Submit(func() {
            defer wg.Done()
            
            file := processFile(path)
            
            mu.Lock()
            files = append(files, file)
            mu.Unlock()
        })
        
        return nil
    })
    
    wg.Wait()
    return files, nil
}
```

##### 2. 批量刮削
```go
func (s *TmdbService) IdentifyBatch(files []FileItem) ([]models.Media, error) {
    // 批量请求，减少 API 调用
    var medias []models.Media
    
    for i := 0; i < len(files); i += batchSize {
        end := min(i+batchSize, len(files))
        batch := files[i:end]
        
        results, err := s.identifyBatch(batch)
        if err != nil {
            continue
        }
        
        medias = append(medias, results...)
    }
    
    return medias, nil
}
```

##### 3. 数据库批量写入
```go
func (r *TransferHistoryRepository) BatchCreate(histories []*models.TransferHistory) error {
    return r.db.CreateInBatches(histories, 100).Error
}
```

#### 验收标准 ✅
- [x] 扫描速度提升 > 3x
- [x] 刮削速度提升 > 2x
- [x] 数据库写入速度提升 > 5x

---

### Day 3-4: 监控指标 ✅

#### 任务清单

##### 1. Workflow 指标
```go
// internal/monitor/metrics/workflow.go
var (
    workflowExecutionTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "workflow_execution_total",
            Help: "Total number of workflow executions",
        },
        []string{"workflow_type", "status"},
    )
    
    workflowExecutionDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "workflow_execution_duration_seconds",
            Help: "Workflow execution duration in seconds",
        },
        []string{"workflow_type"},
    )
)
```

##### 2. Actions 指标
```go
var (
    actionExecutionTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "action_execution_total",
            Help: "Total number of action executions",
        },
        []string{"action_name", "status"},
    )
)
```

##### 3. Grafana Dashboard
创建 Dashboard 配置文件:
```json
{
  "dashboard": {
    "title": "MoviePilot Workflows",
    "panels": [
      {
        "title": "Workflow Execution Rate",
        "targets": [
          {
            "expr": "rate(workflow_execution_total[5m])"
          }
        ]
      },
      {
        "title": "Workflow Success Rate",
        "targets": [
          {
            "expr": "sum(rate(workflow_execution_total{status=\"success\"}[5m])) / sum(rate(workflow_execution_total[5m]))"
          }
        ]
      }
    ]
  }
}
```

#### 验收标准 ✅
- [x] 指标正确采集
- [x] Prometheus 可查询
- [x] Grafana Dashboard 可视化

#### 完成报告
详见: [WEEK4_COMPLETION_REPORT.md](./WEEK4_COMPLETION_REPORT.md)

---

### Day 5: 日志优化

#### 任务清单
- [ ] 结构化日志完善
- [ ] 日志级别调整
- [ ] 敏感信息脱敏 (密码、Token)
- [ ] 日志轮转配置

---

### Day 6-7: 文档与部署

#### 文档任务
- [ ] API 文档 (Swagger)
- [ ] 部署文档更新
- [ ] 用户手册
- [ ] 开发者文档

#### 部署任务
- [ ] Docker Compose 优化
- [ ] 环境变量配置
- [ ] 健康检查配置
- [ ] 日志收集配置

---

## 验收标准总结

### 功能验收
- [ ] 本地文件扫描正常
- [ ] TMDB 刮削准确率 > 90%
- [ ] 文件转移成功率 > 95%
- [ ] API 响应时间 < 500ms
- [ ] 并发处理能力 > 100 req/s

### 质量验收
- [ ] 单元测试覆盖率 > 70%
- [ ] 集成测试通过
- [ ] 无内存泄漏
- [ ] 无 Goroutine 泄漏

### 文档验收
- [ ] API 文档完整
- [ ] 部署文档清晰
- [ ] 代码注释充分

---

## 风险与应对

### 风险1: TMDB API 限流
**应对**: 
- 实现本地缓存
- 使用代理池
- 降级到本地识别

### 风险2: 文件权限问题
**应对**:
- 提前检查权限
- 提供详细错误信息
- 支持 sudo 模式

### 风险3: 性能不达标
**应对**:
- 增加并发度
- 优化数据库查询
- 使用缓存

---

## 下一步

第一阶段完成后，进入第二阶段: 订阅与下载链路 (Week 5-10)
