# MoviePilot Python 到 Go 功能映射表

## 📋 映射概述

本文档详细描述了 MoviePilot Python 版本功能到 Go 版本的映射关系，包括功能状态、实现复杂度和技术方案。

---

## 🎯 功能模块映射总览

| 模块 | Python功能 | Go实现 | 状态 | 复杂度 | 优先级 |
|------|-----------|--------|------|--------|--------|
| **用户系统** | Flask-Login | Gin + JWT | 🔄 60% | 中 | 高 |
| **订阅管理** | SQLAlchemy + APScheduler | GORM + Cron | 🔄 30% | 高 | 高 |
| **下载管理** | TransmissionAPI | Go下载客户端 | 🔄 20% | 高 | 高 |
| **媒体管理** | MediaInfo + FFmpeg | Go bindings | ⏳ 0% | 高 | 中 |
| **插件系统** | Python动态导入 | gRPC + Python | ⏳ 0% | 极高 | 高 |
| **Web界面** | Jinja2 + Vue.js | Go templates + API | ⏳ 0% | 高 | 中 |
| **搜索系统** | 多站点聚合 | Go并发搜索 | ⏳ 0% | 高 | 中 |
| **通知系统** | 多渠道通知 | Go通知客户端 | ⏳ 0% | 中 | 低 |

---

## 👤 用户系统映射

### 认证授权
| Python实现 | Go实现 | 状态 | 技术方案 |
|-----------|--------|------|----------|
| Flask-Login会话 | JWT Token | 🔄 80% | gin-jwt中间件 |
| RBAC权限装饰器 | Casbin RBAC | 🔄 60% | casbin权限框架 |
| 密码加密(bcrypt) | bcrypt | ✅ 100% | golang.org/x/crypto |
| OAuth2登录 | OAuth2库 | 🔄 40% | go-oauth2 |

### 用户管理
| 功能 | Python代码 | Go代码 | 状态 |
|------|-----------|--------|------|
| 用户注册 | `@app.route('/register')` | `POST /api/v1/register` | 🔄 70% |
| 用户登录 | `@app.route('/login')` | `POST /api/v1/login` | 🔄 80% |
| 用户信息 | `@app.route('/profile')` | `GET /api/v1/users/:id` | 🔄 50% |
| 密码修改 | `@app.route('/change-password')` | `PUT /api/v1/users/:id/password` | 🔄 30% |

### 数据模型映射
```python
# Python SQLAlchemy模型
class User(db.Model):
    id = db.Column(db.Integer, primary_key=True)
    username = db.Column(db.String(80), unique=True)
    email = db.Column(db.String(120), unique=True)
    password_hash = db.Column(db.String(128))
    is_active = db.Column(db.Boolean, default=True)
```

```go
// Go GORM模型
type User struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    Username  string    `gorm:"uniqueIndex;size:80" json:"username"`
    Email     string    `gorm:"uniqueIndex;size:120" json:"email"`
    Password  string    `gorm:"size:128" json:"-"`
    IsActive  bool      `gorm:"default:true" json:"is_active"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

---

## 📺 订阅管理映射

### 订阅类型
| Python实现 | Go实现 | 状态 | 说明 |
|-----------|--------|------|------|
| 电影订阅 | MovieSubscription | 🔄 40% | TMDB API集成 |
| 剧集订阅 | TVSubscription | 🔄 20% | 季/集跟踪 |
| 资源订阅 | ResourceSubscription | ⏳ 0% | 关键词匹配 |

### 自动化逻辑
| 功能 | Python方案 | Go方案 | 状态 |
|------|-----------|--------|------|
| 定时检查 | APScheduler | Cron库 | 🔄 60% |
| 资源匹配 | 正则表达式 | Go regexp | 🔄 50% |
| 下载触发 | 事件系统 | Channel事件 | 🔄 30% |
| 状态更新 | 数据库操作 | GORM | 🔄 70% |

### 数据模型对比
```python
# Python订阅模型
class Subscription(db.Model):
    id = db.Column(db.Integer, primary_key=True)
    name = db.Column(db.String(100))
    type = db.Column(db.String(20))  # movie, tv
    tmdb_id = db.Column(db.Integer)
    season = db.Column(db.Integer)  # 仅剧集
    quality = db.Column(db.String(20))
    auto_download = db.Column(db.Boolean, default=True)
```

```go
// Go订阅模型
type Subscription struct {
    ID           uint      `gorm:"primaryKey" json:"id"`
    Name         string    `gorm:"size:100" json:"name"`
    Type         string    `gorm:"size:20" json:"type"` // movie, tv
    TMDBID       int       `json:"tmdb_id"`
    Season       *int      `json:"season,omitempty"`
    Quality      string    `gorm:"size:20" json:"quality"`
    AutoDownload bool      `gorm:"default:true" json:"auto_download"`
    UserID       uint      `json:"user_id"`
    User         User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}
```

---

## ⬇️ 下载管理映射

### 下载器支持
| 下载器 | Python实现 | Go实现 | 状态 | 备注 |
|--------|-----------|--------|------|------|
| Transmission | transmissionrpc | go-transmission | 🔄 40% | 主要下载器 |
| qBittorrent | python-qbittorrent | go-qbittorrent | ⏳ 0% | 备选方案 |
| Aria2 | aria2p | go-aria2 | ⏳ 0% | HTTP/FTP下载 |

### 下载功能映射
| 功能 | Python代码 | Go代码 | 状态 |
|------|-----------|--------|------|
| 添加下载 | `client.add_torrent()` | `client.AddTorrent()` | 🔄 50% |
| 查询状态 | `client.get_torrents()` | `client.GetTorrents()` | 🔄 40% |
| 暂停/恢复 | `client.pause_torrent()` | `client.PauseTorrent()` | 🔄 30% |
| 删除任务 | `client.remove_torrent()` | `client.RemoveTorrent()` | 🔄 20% |

### 下载状态管理
```python
# Python状态枚举
class DownloadStatus:
    PENDING = "pending"
    DOWNLOADING = "downloading"
    COMPLETED = "completed"
    FAILED = "failed"
    PAUSED = "paused"
```

```go
// Go状态常量
const (
    DownloadStatusPending    = "pending"
    DownloadStatusDownloading = "downloading"
    DownloadStatusCompleted  = "completed"
    DownloadStatusFailed     = "failed"
    DownloadStatusPaused     = "paused"
)

type DownloadStatus string
```

---

## 🔌 插件系统映射

### 架构变化
| 方面 | Python版本 | Go版本 | 状态 |
|------|-----------|--------|------|
| 插件加载 | 动态导入 | gRPC服务 | ⏳ 0% |
| 通信方式 | 函数调用 | gRPC + HTTP | ⏳ 0% |
| 插件语言 | 仅Python | Python + Go | ⏳ 0% |
| 热加载 | 支持 | 支持 | ⏳ 0% |

### 插件类型映射
| 插件类型 | Python实现 | Go实现 | 状态 |
|----------|-----------|--------|------|
| 站点插件 | SitePlugin类 | gRPC接口 | ⏳ 0% |
| 刮削器 | ScraperPlugin | gRPC接口 | ⏳ 0% |
| 通知器 | NotifierPlugin | gRPC接口 | ⏳ 0% |
| 媒体服务器 | MediaServerPlugin | gRPC接口 | ⏳ 0% |

### gRPC接口定义
```protobuf
// 插件服务接口
service PluginService {
    rpc LoadPlugin(LoadPluginRequest) returns (LoadPluginResponse);
    rpc UnloadPlugin(UnloadPluginRequest) returns (UnloadPluginResponse);
    rpc ExecutePlugin(ExecutePluginRequest) returns (ExecutePluginResponse);
    rpc GetPluginStatus(GetPluginStatusRequest) returns (GetPluginStatusResponse);
}
```

---

## 🎬 媒体管理映射

### 媒体处理
| 功能 | Python库 | Go库 | 状态 |
|------|-----------|------|------|
| 视频信息 | pymediainfo | go-mediainfo | ⏳ 0% |
| 视频转码 | ffmpeg-python | exec + ffmpeg | ⏳ 0% |
| 字幕处理 | pysrt | go-srt | ⏳ 0% |
| 封面提取 | ffmpegthumbnailer | exec + ffmpeg | ⏳ 0% |

### 媒体库管理
| 功能 | Python实现 | Go实现 | 状态 |
|------|-----------|--------|------|
| 扫描媒体库 | os.walk | filepath.Walk | ⏳ 0% |
| 识别媒体 | guessit | go-guessit | ⏳ 0% |
| 重命名文件 | pathlib | os.Rename | ⏳ 0% |
| 组织目录 | shutil | io/ioutil | ⏳ 0% |

---

## 🌐 Web界面映射

### 后端API
| Python路由 | Go路由 | 状态 | 说明 |
|-----------|--------|------|------|
| `@app.route('/api/subscriptions')` | `GET /api/v1/subscriptions` | 🔄 70% | RESTful API |
| `@app.route('/api/downloads')` | `GET /api/v1/downloads` | 🔄 50% | 下载状态API |
| `@app.route('/api/media')` | `GET /api/v1/media` | ⏳ 0% | 媒体库API |

### 前端技术
| 技术 | Python版本 | Go版本 | 状态 |
|------|-----------|--------|------|
| 模板引擎 | Jinja2 | Go templates | 🔄 40% |
| 前端框架 | Vue.js | 保持Vue.js | ⏳ 0% |
| API通信 | Axios | 保持Axios | ⏳ 0% |
| 状态管理 | Vuex | 保持Vuex | ⏳ 0% |

---

## 🔍 搜索系统映射

### 搜索引擎
| 搜索源 | Python实现 | Go实现 | 状态 |
|--------|-----------|--------|------|
| 站点搜索 | 自定义爬虫 | Go HTTP客户端 | ⏳ 0% |
| 豆瓣搜索 | douban-api | Go HTTP客户端 | ⏳ 0% |
| TMDB搜索 | tmdbv3-api | go-tmdb | ⏳ 0% |
| 磁力链接 | DHT爬虫 | Go DHT客户端 | ⏳ 0% |

### 搜索功能
| 功能 | Python方案 | Go方案 | 状态 |
|------|-----------|--------|------|
| 并发搜索 | ThreadPoolExecutor | Goroutines | ⏳ 0% |
| 结果去重 | 集合操作 | Map去重 | ⏳ 0% |
| 结果排序 | 自定义排序 | Go sort包 | ⏳ 0% |
| 缓存结果 | Redis | Redis | 🔄 60% |

---

## 📢 通知系统映射

### 通知渠道
| 渠道 | Python库 | Go实现 | 状态 |
|------|-----------|--------|------|
| 邮件 | smtplib | net/smtp | ⏳ 0% |
| 微信 | Server酱 | HTTP API | ⏳ 0% |
| Telegram | python-telegram-bot | go-telegram-bot-api | ⏳ 0% |
| 钉钉 | dingtalk-sdk | HTTP API | ⏳ 0% |

### 通知模板
| 功能 | Python实现 | Go实现 | 状态 |
|------|-----------|--------|------|
| 模板渲染 | Jinja2 | Go templates | ⏳ 0% |
| 消息格式化 | f-string | fmt.Sprintf | ⏳ 0% |
| 多媒体支持 | Base64编码 | Base64编码 | ⏳ 0% |

---

## 📊 迁移复杂度评估

### 高复杂度模块
1. **插件系统** - 需要重新设计架构
2. **媒体处理** - 涉及FFmpeg集成
3. **搜索系统** - 需要重写爬虫逻辑
4. **Web界面** - 前后端分离改造

### 中等复杂度模块
1. **订阅管理** - 业务逻辑复杂
2. **下载管理** - 第三方客户端集成
3. **用户系统** - 认证授权逻辑
4. **通知系统** - 多渠道集成

### 低复杂度模块
1. **配置管理** - 直接替换
2. **日志系统** - 功能对等替换
3. **数据库操作** - ORM映射
4. **基础API** - 路由映射

---

## 🎯 迁移优先级

### P0 - 核心功能 (必须完成)
- [x] 基础架构
- [🔄] 用户系统
- [🔄] 订阅管理
- [🔄] 下载管理

### P1 - 重要功能 (应该完成)
- [⏳] 插件系统
- [⏳] 媒体管理
- [⏳] 搜索系统
- [⏳] Web界面

### P2 - 增强功能 (可以完成)
- [⏳] 通知系统
- [⏳] 监控系统
- [⏳] 性能优化
- [⏳] 高级配置

---

## 📋 实现检查清单

### 代码迁移检查
- [ ] API接口完全映射
- [ ] 数据模型字段对应
- [ ] 业务逻辑一致性
- [ ] 错误处理完整性
- [ ] 日志记录完整性

### 功能验证检查
- [ ] 单元测试覆盖
- [ ] 集成测试通过
- [ ] 性能基准达标
- [ ] 安全测试通过
- [ ] 用户体验测试

---

*文档最后更新: 2025-11-21*  
*版本: v1.0*  
*负责人: 技术架构团队*