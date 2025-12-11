# app/core/meta 子包拆分与 Go 归档建议

> Python 源：`app/core/meta/`
>
> - `metabase.py`
> - `metavideo.py`
> - `metaanime.py`
> - `words.py`
> - `releasegroup.py`
> - `customization.py`
> - `streamingplatform.py`
>
> Go 目标：`internal/business/domains/media/` + 少量辅助放在 `internal/models/enums/`

本文件是在 `docs/metainfo-migration-plan.md` 的基础上，进一步细化 **meta 子包内部的职责拆分** 以及在 Go 侧的落盘结构，便于后续实现 `domains/media` 领域层代码。

---

## 1. 设计原则回顾

- **领域逻辑归 domains 层**：
  - 标题解析、季集推导、分辨率/编码/字幕组匹配等都属于核心领域规则，统一放在 `internal/business/domains/media`。
- **DTO / DB 模型与解析逻辑分离**：
  - DTO：`internal/models/dto/media/*`
  - GORM 模型：`internal/models/*`
  - 解析器 & 规则：`internal/business/domains/media/*`
- **可配置化识别规则**：
  - `WordsMatcher`、`ReleaseGroupsMatcher`、`CustomizationMatcher` 都依赖系统配置（`SystemConfigKey.*`），在 Go 中通过 **业务 Service + Repository** 提供配置数据，解析器只关心接口。
- **可测试性**：
  - 解析函数尽量纯函数化：输入字符串/配置快照 → 输出 `Meta` 结构体，方便单测覆盖各种命名样例。

---

## 2. Python meta 子包职责拆分概览

### 2.1 核心基类

- `metabase.py`
  - `MetaBase`：统一的元信息基础结构与公共逻辑：
    - 名称：`cn_name`, `en_name`, `name` property
    - 年份：`year`
    - 季：`begin_season`, `end_season`, `total_season` + 派生：`season`, `sea`, `season_seq`, `season_list`
    - 集：`begin_episode`, `end_episode`, `total_episode` + 派生：`episode`, `episode_list`, `episodes`, `episode_seq`, `episode_seqs`
    - 媒体类型：`type: MediaType`
    - 资源属性：`resource_type`, `resource_effect`, `resource_pix`, `resource_team`, `customization`, `web_source`, `video_encode`, `audio_encode`
    - 其他：`part`, `tmdbid`, `doubanid`, `apply_words`
  - 通用逻辑：
    - 副标题解析 `init_subtitle`：解析“第X季/全X季/第X集-第Y集/全X集”等中文、英文混合表达。
    - 季/集判断 & 修改：`is_in_season`, `is_in_episode`, `set_season`, `set_episode`, `set_episodes`。
    - 元信息合并：`merge(meta)`。
    - 序列化：`to_dict()`，补充 `season_episode`, `edition`, `episode_list` 等派生字段。

### 2.2 具体解析实现

- `metavideo.py` — 电影/电视剧标题解析：
  - `MetaVideo(MetaBase)`：
    - 使用 `Tokens` 分词，对每个 token 依次尝试匹配：名称、年份、季、集、资源版本、分辨率、流媒体平台、编码等。
    - 利用大量正则处理 Season/EP/分辨率/版本/编码格式。
    - 调用：
      - `ReleaseGroupsMatcher` → `resource_team`
      - `CustomizationMatcher` → `customization`
      - `StreamingPlatforms` → `web_source`

- `metaanime.py` — 动漫标题解析：
  - `MetaAnime(MetaBase)`：
    - 依赖 `anitopy` 解析动漫命名：`anime_title`, `anime_year`, `anime_season`, `episode_number`, `video_resolution`, `video_term`, `audio_term`, `release_group` 等。
    - 按动漫圈习惯拆分中/英文名，清理噪音字符，推导季/集/分辨率。
    - 同样调用 `ReleaseGroupsMatcher` 与 `CustomizationMatcher`。

### 2.3 辅助匹配器

- `words.py` — 识别词预处理：
  - `WordsMatcher`：
    - 从系统配置 `SystemConfigKey.CustomIdentifiers` 读取“自定义识别词”，支持：
      - 屏蔽词
      - 简单替换：`A => B`
      - 集数偏移：`... && 前定位 <> 后定位 >> EP+N` 之类表达式。
    - 在 `MetaInfo` 入口前对标题做归一化处理。

- `releasegroup.py` — 制作组/字幕组识别：
  - `ReleaseGroupsMatcher`：
    - 内置大量 PT/字幕组正则 + 用户自定义 `CustomReleaseGroups`。
    - 从标题中提取组名，返回 `"A@B@C"` 形式字符串。

- `customization.py` — 自定义占位符识别：
  - `CustomizationMatcher`：
    - 从配置 `SystemConfigKey.Customization` 读取占位符列表，通过正则在标题中匹配，保留出现顺序，用 `@` 连接。

- `streamingplatform.py` — 流媒体平台映射：
  - `StreamingPlatforms`：
    - 维护 `(short, full)` 列表，构建大小写无关的查找表。
    - 提供 `is_streaming_platform` 和 `get_streaming_platform_name`。

---

## 3. Go 侧 domains/media 包结构草图

建议在 Go 中将上述职责归档至 `internal/business/domains/media`，并保持“结构清晰 + 易测”的拆分：

```text
internal/business/domains/media/
├── meta_base.go          # MetaBase 领域结构与通用方法
├── meta_video.go         # MetaVideo 标题解析实现（电影/电视剧）
├── meta_anime.go         # MetaAnime 标题解析实现（动漫）
├── matcher_words.go      # WordsMatcher：自定义识别词预处理
├── matcher_release.go    # ReleaseGroupsMatcher：字幕组/制作组匹配
├── matcher_custom.go     # CustomizationMatcher：自定义占位符匹配
├── streaming_platforms.go# StreamingPlatforms：流媒体平台映射
├── meta_service.go       # 对外暴露的解析服务接口（MetaInfo/MetaInfoPath 等）
└── meta_types.go         # 本领域内部使用的辅助类型、常量
```

### 3.1 meta_base.go

- 主要内容：
  - Go 版 `MetaBase` 领域结构（可参考 `metainfo-migration-plan.md` 中的结构提案）：
    - `Title` / `OrgString` / `Subtitle`
    - `Type`（依赖 `enums.MediaType`）
    - `CnName`, `EnName`, `Year`
    - 季/集相关字段与派生方法：
      - `BeginSeason`, `EndSeason`, `TotalSeason`
      - `BeginEpisode`, `EndEpisode`, `TotalEpisode`
      - `Season()`, `SeasonSeq()`, `SeasonList()`
      - `Episode()`, `EpisodeList()`, `Episodes()`, `EpisodeSeq()`, `EpisodeSeqs()`
    - 资源属性：`ResourceType`, `ResourceEffect`, `ResourcePix`, `ResourceTeam`, `Customization`, `WebSource`, `VideoEncode`, `AudioEncode`。
    - 其他：`Part`, `TMDBID`, `DoubanID`, `AppliedWords`。
  - 通用方法：
    - `InitSubtitle(title, subtitle string)`
    - `IsInSeason(...)`, `IsInEpisode(...)`
    - `SetSeason(...)`, `SetEpisode(...)`, `SetEpisodes(...)`
    - `Merge(other *MetaBase)`

> 注意：Go 中可以让 `MetaVideo` / `MetaAnime` **组合** `MetaBase`（匿名嵌入），而不是继承。

### 3.2 meta_video.go

- 职责：实现与 `metavideo.py` 等价的影视标题解析逻辑。
- 建议：
  - 将大块私有方法按职能拆分为 Go 函数：
    - `parseNameTokens(...)`
    - `parseYearToken(...)`
    - `parseSeasonToken(...)`
    - `parseEpisodeToken(...)`
    - `parseResourceTypeToken(...)`
    - `parseResolutionToken(...)`
    - `parseVideoEncodeToken(...)`
    - `parseAudioEncodeToken(...)`
    - `parseWebSourceToken(...)`
  - 领域对象：

```go
type MetaVideo struct {
    *MetaBase
    // 解析过程内部状态可以用未导出字段保存
}

func NewMetaVideo(title, subtitle string, isFile bool, deps MetaParserDeps) *MetaVideo
```

- 其中 `MetaParserDeps` 用于注入依赖：
  - `WordsMatcher`、`ReleaseGroupsMatcher`、`CustomizationMatcher`、`StreamingPlatforms` 等（或其接口）。

### 3.3 meta_anime.go

- 职责：实现 `MetaAnime` 动漫标题解析：
  - 是否直接调用 `anitopy` 对应的 Go 版本/重写逻辑，取决于依赖可用性；
  - 也可以先实现一个简化版本，覆盖常见命名格式。
- 结构建议：

```go
type MetaAnime struct {
    *MetaBase
}

func NewMetaAnime(title, subtitle string, isFile bool, deps MetaParserDeps) *MetaAnime
```

- 逻辑步骤与 Python 大体同步：
  - 解析 `AnimeTitle` → 拆分中英文名
  - 解析 `AnimeYear` / `AnimeSeason` / `EpisodeNumber`
  - 填充分辨率、编码、release_group、自定义占位符
  - 调用 `InitSubtitle` 进一步补齐季/集

### 3.4 matcher_words.go

- 职责：Go 实现 `WordsMatcher`：
  - 接口设计：

```go
// 配置读取通过业务层接口注入，避免直接依赖 repositories

type IdentifierConfigProvider interface {
    GetCustomIdentifiers(ctx context.Context) ([]string, error)
}

type WordsMatcher struct {
    provider IdentifierConfigProvider
}

func (m *WordsMatcher) Prepare(title string, customWords []string) (string, []string, error)
```

- 内部实现：
  - 解析配置语法（三类规则），利用 `regexp` 与 `cn2an` 等库实现替换和集数偏移。

### 3.5 matcher_release.go

- 职责：Go 实现 `ReleaseGroupsMatcher`：

```go
type ReleaseGroupsMatcher struct {
    builtInPatterns []string
    customProvider  ReleaseGroupConfigProvider // 从配置读取自定义组
}

func (m *ReleaseGroupsMatcher) Match(title string, groupsPattern string) string
```

- 内置规则可以直接用常量表或从 JSON/配置文件生成，便于后续维护。

### 3.6 matcher_custom.go

- 职责：Go 实现 `CustomizationMatcher`：
  - 从配置读取字符串/数组，构建正则；
  - 返回 `"A@B@C"` 形式字符串；
  - 支持自定义分隔符（如后续需要）。

### 3.7 streaming_platforms.go

- 职责：Go 实现 `StreamingPlatforms`：

```go
type StreamingPlatforms struct {
    lookup map[string]string // UPPER(alias) -> canonical name
}

func NewStreamingPlatforms() *StreamingPlatforms
func (s *StreamingPlatforms) IsStreamingPlatform(name string) bool
func (s *StreamingPlatforms) GetName(code string) string
```

- 常量表可以放在同文件顶部，或拆到 `streaming_platforms_data.go` 以保持源码清爽。

### 3.8 meta_service.go

- 职责：将 `metainfo.py` 入口函数概念化为 **领域服务接口**：

```go
type MetaService interface {
    MetaInfo(ctx context.Context, title, subtitle string, isFile bool, customWords []string) (*MetaBase, error)
    MetaInfoPath(ctx context.Context, path string) (*MetaBase, error)
    IsAnime(title string) bool
}

type metaService struct {
    wordsMatcher       *WordsMatcher
    releaseMatcher     *ReleaseGroupsMatcher
    customizationMatch *CustomizationMatcher
    streamingPlatforms *StreamingPlatforms
    // 其他依赖...
}
```

- `MetaInfo` 实现：
  - 读取/注入配置 → 调 `WordsMatcher.Prepare`
  - 调用 `IsAnime` 判定动漫与否 → 构造 `MetaAnime` 或 `MetaVideo`
  - 返回 `MetaBase` 指针（嵌入在具体类型中），供上层 `domains/context` / `services/subscribe` / `services/transfer` 使用。

---

## 4. 与其它层的边界

- **与 repositories：**
  - `WordsMatcher` / `ReleaseGroupsMatcher` / `CustomizationMatcher` 不直接依赖 DB；通过接口 `IdentifierConfigProvider` 等由业务服务注入，从 `internal/repositories/systemconfig` 获取配置。
- **与 DTO：**
  - DTO 中的字段尽量是 `string`/简单枚举，不暴露这些解析器的内部状态字段；仅用 `MetaBase` 里的核心语义字段。
- **与 config：**
  - 与 `settings.RMT_MEDIAEXT` 等配置的对应关系，在 `metainfo-migration-plan.md` 中已有说明；在 Go 中从 `config` 包读取，传给解析器。

---

## 5. 实施优先级建议

1. **先实现 MetaBase + StreamingPlatforms**：
   - 提供基础数据结构和平台映射工具；
   - 可以立即被后续解析逻辑和 Context 域模型使用。
2. **实现 WordsMatcher / ReleaseGroupsMatcher / CustomizationMatcher**：
   - 与配置子系统打通，保证可通过配置调整识别行为。
3. **实现 MetaVideo**：
   - 覆盖电影/电视剧主流程命名；
   - 编写大量单元测试，对照 Python 行为。
4. **实现 MetaAnime**：
   - 视是否需要完全等价于 `anitopy` 决定实现复杂度；
   - 优先支持常见番剧命名。
5. **实现 MetaService**：
   - 统一对外暴露 `MetaInfo`/`MetaInfoPath` 等接口；
   - 上层只依赖接口，不直接依赖具体解析类。

---

## 6. 总结

- `app/core/meta` 是元信息解析系统的“核心零件”，在 Go 中应集中归档于 `internal/business/domains/media`，并通过 **领域对象 + 匹配器 + 解析服务** 的形式呈现。
- 本文档给出了推荐的文件划分与职责边界，可作为后续实现 `domains/media` 包时的骨架参考。