# metainfo.py 元信息解析系统迁移计划

> Python: `app/core/metainfo.py`  
> Go: `internal/business/domains/meta/` + `internal/business/services/media/`

---

## 1. Python metainfo.py 分析

### 1.1 主要函数

```python
def MetaInfo(title: str, subtitle: Optional[str] = None, custom_words: List[str] = None) -> MetaBase

def MetaInfoPath(path: Path) -> MetaBase

def is_anime(name: str) -> bool

def find_metainfo(title: str) -> Tuple[str, dict]
```

### 1.2 MetaInfo(title, subtitle, custom_words)

核心职责：
- 对 **标题/种子名/文件名** 做一层预处理 + 元信息提取，返回 `MetaBase` 子类（`MetaAnime` 或 `MetaVideo`）

流程：

1. **记录原标题**：`org_title = title`
2. **预处理标题**：
   - 调用 `WordsMatcher().prepare(title, custom_words)`：
     - 去掉无关前后缀、站点标记等
     - 应用自定义识别词（如强制指定季/集/年份）
     - 返回：
       - 清理后的 `title`
       - 使用到的识别词 `apply_words`
3. **内嵌元信息提取**：
   - `title, metainfo = find_metainfo(title)`
   - 作用：从标题中的特殊标记解析出：
     - `tmdbid`, `doubanid`, `type`, `begin_season`, `end_season`, `total_season`, `begin_episode`, `end_episode`, `total_episode`
   - 并将 `{[tmdbid=xxx;type=movies;s=1-2;e=1-10]}` 等标记从标题字符串中剥离
4. **判断是否文件名**：
   - 如果 `Path(title).suffix.lower()` 在 `settings.RMT_MEDIAEXT` 中 → `isfile = True` 且去掉后缀
   - 否则 `isfile = False`
5. **选择解析器**：
   - `MetaAnime(title, subtitle, isfile)` if `is_anime(title)` else `MetaVideo(title, subtitle, isfile)`
6. **修正/补充元信息**：根据 `metainfo` 覆盖 `meta` 对象中的字段：
   - `tmdbid`, `doubanid`, `type`, `begin_season`, `end_season`, `total_season`, `begin_episode`, `end_episode`, `total_episode`
7. 返回 `meta`（MetaAnime / MetaVideo 实例）

依赖：
- `app.core.meta.MetaAnime`, `MetaVideo`, `MetaBase`
- `app.core.meta.words.WordsMatcher`
- `settings.RMT_MEDIAEXT`
- `MediaType` 枚举
- 日志 `logger`

### 1.3 MetaInfoPath(path)

职责：根据一个 **完整路径** 推导元信息，利用“多级目录名 + 文件名”的组合增强识别：

流程：

1. `file_meta = MetaInfo(title=path.name)`
2. `dir_meta = MetaInfo(title=path.parent.name)`
3. `file_meta.merge(dir_meta)`
4. `root_meta = MetaInfo(title=path.parent.parent.name)`
5. `file_meta.merge(root_meta)`
6. 返回 `file_meta`

其中 `merge` 逻辑在 `MetaBase` 中实现，用来合并季/集/年份等字段。

### 1.4 is_anime(name)

职责：判断一个名称是否更像“动漫（番剧）”而不是普通电影/剧集，用于决定选择 `MetaAnime` 还是 `MetaVideo`。

规则（正则）：
- 如果匹配 `【[+0-9XVPI-]+】\s*【` → 认为是动漫
- 或匹配 `\s+\-\s+[\dv]{1,4}\s+` → 动漫
- 如果包含明显的 Sxx/Epxx 格式：`S01- S02`, `EP01-EP10` 等 → 更偏向普通剧集 → 返回 False
- 其他某些 `[xx][xx]` 样式也认为是动漫

### 1.5 find_metainfo(title)

职责：从标题中剥离内嵌的“元信息标签”，并返回：
- 清理后的标题
- 一个包含季/集/ID/类型信息的 dict

默认返回结构：

```python
metainfo = {
    'tmdbid': None,
    'doubanid': None,
    'type': None,
    'begin_season': None,
    'end_season': None,
    'total_season': None,
    'begin_episode': None,
    'end_episode': None,
    'total_episode': None,
}
```

支持的标签格式：

1. **自定义 `{[ ... ]}` 格式**：

   ```
   {[tmdbid=12345;type=movies;s=1-2;e=1-10]}
   ```

   - 正则 `(?<={\[)[\W\w]+(?=]})`
   - 解析：
     - `tmdbid=\d+`
     - `doubanid=\d+`
     - `type=movies|tv` → 映射到 `MediaType.MOVIE/TV`
     - `s=1` / `s=1-3` → `begin_season`, `end_season`
     - `e=1` / `e=1-10` → `begin_episode`, `end_episode`
   - 只要出现过这些信息，就把整个 `{[... ]}` 从 title 中删掉

2. **Emby 风格标签（tmdbid）**：

   - `[tmdbid=xxxx]` / `[tmdbid-xxxx]`
   - `[tmdb=xxxx]` / `[tmdb-xxxx]`
   - `{tmdbid=xxxx}` / `{tmdbid-xxxx}`
   - `{tmdb=xxxx}` / `{tmdb-xxxx}`

   匹配到后同时 **更新 metainfo['tmdbid'] 并从 title 中删掉标签**。

3. **自动计算总季/总集**：

   - 如果有 `begin_season` 和 `end_season`：
     - 如果 `begin > end` 交换
     - `total_season = end - begin + 1`
   - 如果只有 `begin_season`：`total_season = 1`
   - 集数同理：`begin_episode`/`end_episode` → `total_episode`

---

## 2. Go 侧总体设计

### 2.1 放置位置

根据你的分层规范：

- **元信息解析属于核心领域逻辑** → 放在 `internal/business/domains/meta/`
- 与链路处理（下载、订阅、整理）紧耦合的部分 → 可由 `internal/business/services/media/` 调用

建议结构：

```text
internal/business/domains/meta/
├── meta.go           # MetaBase / MetaVideo / MetaAnime（对应 Python app/core/meta）
├── metainfo.go       # 本文件对应的逻辑：MetaInfo / MetaInfoPath / is_anime / find_metainfo
└── words.go          # WordsMatcher 的 Go 实现
```

> 注：`app/core/meta.py` 已在之前的 design 中分析过，这里只聚焦 `metainfo.py` 这一层包装逻辑。

### 2.2 Go 版本函数设计

#### 2.2.1 MetaInfo

```go
// internal/business/domains/meta/metainfo.go

package meta

type MetaBase interface {
    Merge(other MetaBase)
    // ... 其他字段访问方法
}

func MetaInfo(title string, subtitle *string, customWords []string, cfg *config.Config) MetaBase {
    // 1. 原标题
    orgTitle := title

    // 2. 预处理标题（WordsMatcher）
    wm := NewWordsMatcher(cfg) // 从配置读取自定义识别词等
    cleanedTitle, applyWords := wm.Prepare(title, customWords)

    // 3. 内嵌元信息解析
    cleanedTitle, metainfo := FindMetaInfo(cleanedTitle)

    // 4. 判断是否文件名
    isFile := false
    if cleanedTitle != "" {
        ext := strings.ToLower(filepath.Ext(cleanedTitle))
        if contains(cfg.Media.SupportedExt, ext) { // 对应 settings.RMT_MEDIAEXT
            isFile = true
            cleanedTitle = strings.TrimSuffix(cleanedTitle, ext)
        }
    }

    // 5. 选择解析器：动漫 or 普通视频
    var metaObj MetaBase
    if IsAnime(cleanedTitle) {
        metaObj = NewMetaAnime(cleanedTitle, subtitle, isFile)
    } else {
        metaObj = NewMetaVideo(cleanedTitle, subtitle, isFile)
    }

    // 6. 补充字段（原标题、使用的识别词、tmdbid 等）
    metaObj.SetOriginalTitle(orgTitle)
    metaObj.SetApplyWords(applyWords)

    if metainfo.TMDBID != 0 {
        metaObj.SetTMDBID(metainfo.TMDBID)
    }
    if metainfo.DoubanID != "" {
        metaObj.SetDoubanID(metainfo.DoubanID)
    }
    // ... 季/集相关字段

    return metaObj
}
```

为此需要：
- 在 `MetaBase`/`MetaVideo`/`MetaAnime` 中补充必要 setter 方法
- `cfg.Media.SupportedExt` 对应 Python 的 `settings.RMT_MEDIAEXT`

#### 2.2.2 MetaInfoPath

```go
func MetaInfoPath(path string, cfg *config.Config) MetaBase {
    p := filepath.Clean(path)

    fileName := filepath.Base(p)
    fileMeta := MetaInfo(fileName, nil, nil, cfg)

    parent := filepath.Base(filepath.Dir(p))
    if parent != "." && parent != "/" {
        dirMeta := MetaInfo(parent, nil, nil, cfg)
        fileMeta.Merge(dirMeta)
    }

    grandParent := filepath.Base(filepath.Dir(filepath.Dir(p)))
    if grandParent != "." && grandParent != "/" {
        rootMeta := MetaInfo(grandParent, nil, nil, cfg)
        fileMeta.Merge(rootMeta)
    }

    return fileMeta
}
```

#### 2.2.3 IsAnime

```go
var (
    reAnimeStyle1 = regexp.MustCompile(`【[+0-9XVPI-]+】\s*【`)
    reAnimeStyle2 = regexp.MustCompile(`\s+\-\s+[\dv]{1,4}\s+`)
    reTvPattern   = regexp.MustCompile(`S\d{2}\s*\-\s*S\d{2}|S\d{2}|\s+S\d{1,2}|EP?\d{2,4}\s*\-\s*EP?\d{2,4}|EP?\d{2,4}|\s+EP?\d{1,4}`)
    reAnimeStyle3 = regexp.MustCompile(`\[[+0-9XVPI-]+]\s*\[`)
)

func IsAnime(name string) bool {
    if strings.TrimSpace(name) == "" {
        return false
    }
    if reAnimeStyle1.MatchString(name) {
        return true
    }
    if reAnimeStyle2.MatchString(name) {
        return true
    }
    if reTvPattern.MatchString(name) {
        return false
    }
    if reAnimeStyle3.MatchString(name) {
        return true
    }
    return false
}
```

#### 2.2.4 FindMetaInfo

设计一个结构体承载结果：

```go
type MetaInfoResult struct {
    TMDBID       int64
    DoubanID     string
    Type         MediaType
    BeginSeason  *int
    EndSeason    *int
    TotalSeason  *int
    BeginEpisode *int
    EndEpisode   *int
    TotalEpisode *int
}

func FindMetaInfo(title string) (string, *MetaInfoResult) {
    result := &MetaInfoResult{}
    if title == "" {
        return title, result
    }

    // 1. 处理 {[ ... ]} 格式
    //  - 使用 regexp.FindAllStringSubmatch 来提取 tmdbid/doubanid/type/s/e
    //  - 与 Python 一致，匹配到后从 title 中删除该段

    // 2. 处理 Emby 风格 [tmdbid=xxxx] / [tmdb=xxxx] / {tmdbid=xxxx} / {tmdb=xxxx}
    //  - 顺序与 Python 保持一致，优先级相同

    // 3. 计算 total_season / total_episode

    return strings.TrimSpace(title), result
}
```

> 这里要确保正则和删减逻辑与 Python 行为兼容，尤其是多标签叠加场景需要测试。

---

## 3. 与其他模块的关系

### 3.1 与 `app/core/meta.py`（MetaBase/MetaVideo/MetaAnime）

- `MetaInfo` 是 **入口包装层**：负责预处理 + find_metainfo + Anime 判定 + 调用具体 Meta 解析器
- 具体解析逻辑（如季集解析、版本判断）在 `MetaVideo` / `MetaAnime` 内部实现

Go 侧：
- `internal/business/domains/meta/meta.go` 定义：
  - `MetaBase`, `MetaVideo`, `MetaAnime`
  - 提供 `Merge` 与字段访问/设置方法
- `metainfo.go` 只做：
  - 标题预处理（WordsMatcher）
  - 路径组合（MetaInfoPath）
  - Anime/Video 分流
  - 内嵌标签解析（FindMetaInfo）

### 3.2 与 `WordsMatcher`

- `WordsMatcher().prepare(title, custom_words)` 会：
  - 应用自定义词库（比如“第 x 季”、“完结”等）
  - 返回处理后的 title 和匹配到的识别词

Go 侧：
- `internal/business/domains/meta/words.go` 提供 `WordsMatcher`
- 由配置系统提供词库：
  - 配置路径：`internal/infrastructure/config` 中的某个数组字段

### 3.3 与配置系统

- 使用了 `settings.RMT_MEDIAEXT` 判断“是否文件”
- Go 侧对应：
  - `cfg.Media.SupportedExt []string`

---

## 4. 实施计划

### Phase 1：基础结构与函数骨架（Week 1）

- [ ] 创建 `internal/business/domains/meta/` 目录
- [ ] 在其中创建 `metainfo.go`，定义：
  - `MetaInfo`
  - `MetaInfoPath`
  - `IsAnime`
  - `MetaInfoResult`, `FindMetaInfo`
- [ ] 与现有/规划中的 `MetaBase`/`MetaVideo`/`MetaAnime` 接口对齐（只写调用，不实现内部解析）

### Phase 2：正则与解析逻辑迁移（Week 2）

- [ ] 完整移植 `IsAnime` 中的正则规则，并编写测试用例：
  - 典型动漫命名样例
  - 明显非动漫剧集样例
- [ ] 完整移植 `find_metainfo` 中的解析逻辑：
  - `{[... ]}` 自定义标签
  - 各种 `tmdbid`/`tmdb` 格式
  - 季/集范围 + total 计算
- [ ] 在测试中构造与 Python 相同的样例，比对结果

### Phase 3：与 Meta/WordsMatcher 集成（Week 3）

- [ ] 实现/接入 `WordsMatcher` 的 Go 版本
- [ ] 实现 `MetaVideo` / `MetaAnime` 的基本版本（至少支持季/集/年份）
- [ ] 在 `MetaInfo` 中打通完整链路

### Phase 4：业务接入与优化（Week 4）

- [ ] 修改媒体链路相关 service（search/download/transfer/subscribe）使用新的 `MetaInfo`/`MetaInfoPath`
- [ ] 加上必要的日志（通过 `pkg/logger`），记录：
  - 标题解析结果
  - Anime 判定
  - 内嵌标签解析结果
- [ ] 根据实际使用情况调整正则与规则，增加更多测试样例

---

## 5. 测试策略

### 5.1 单元测试

- `MetaInfo`：
  - 不同标题 + 副标题 + custom_words 的组合
  - 检查 `tmdbid/doubanid/type/season/episode` 是否与 Python 一致
- `MetaInfoPath`：
  - 典型目录结构：`/Show/Season 01/Show.S01E01.mkv` 等
  - 验证目录信息是否被正确合并到 file_meta
- `IsAnime`：
  - 多种命名风格样例（动漫 / 非动漫）
- `FindMetaInfo`：
  - 含/不含 `{[... ]}` 标签
  - 含多种 Emby ID 标签

### 5.2 对比测试

- 在 Python 侧导出一组 `(title, subtitle)` 示例及对应 `MetaInfo(...).to_dict()`
- 在 Go 侧用相同输入调用新实现，比较核心字段（tmdbid/doubanid/type/season/episode/title）

---

## 6. 注意事项

- 需要与 `MetaBase` 设计一起考虑，避免在 `metainfo.go` 中直接操作太多字段，保持职责干净：
  - `metainfo.py` 只负责“发现 + 修正”；
  - 深度解析放在 `MetaVideo`/`MetaAnime` 内部
- 正则表达式尽量与 Python 一致，注意 Go 语法差异（转义）
- 日志统一使用 `pkg/logger`，不要使用 `fmt.Println`

---

## 7. 结论

`metainfo.py` 是所有媒体命名解析的入口层封装：
- 负责把“带各种标记的原始标题/路径”变成干净的解析输入 + 补充 ID/季集信息
- 与 `MetaVideo`/`MetaAnime`、`WordsMatcher`、配置系统紧密协作

在 Go 中，通过 `internal/business/domains/meta/metainfo.go` 实现同等功能，并配合 `MetaBase` 体系和配置系统，可以无缝替代 Python 版本，为后续的搜索、下载、整理业务提供统一的“元信息解析”能力。