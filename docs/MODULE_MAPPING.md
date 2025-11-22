# Python → Go 模块映射表

## Actions 层映射

| Python 模块 | Go 模块 | 状态 | 优先级 |
|------------|---------|------|--------|
| `actions/scan_file.py` | `internal/actions/scan_file_action.go` | ✅ 已完成 | - |
| `actions/scrape_file.py` | `internal/actions/scrape_file_action.go` | ✅ 已完成 | - |
| `actions/transfer_file.py` | `internal/actions/transfer_file_action.go` | ✅ 已完成 | - |
| `actions/fetch_torrents.py` | `internal/actions/fetch_torrents_action.go` | ✅ 已完成 | - |
| `actions/add_download.py` | `internal/actions/add_download_action.go` | ⏳ 待迁移 | P0 |
| `actions/add_subscribe.py` | `internal/actions/add_subscribe_action.go` | ⏳ 待迁移 | P0 |
| `actions/fetch_downloads.py` | `internal/actions/fetch_downloads_action.go` | ⏳ 待迁移 | P1 |
| `actions/fetch_medias.py` | `internal/actions/fetch_medias_action.go` | ⏳ 待迁移 | P1 |
| `actions/fetch_rss.py` | `internal/actions/fetch_rss_action.go` | ⏳ 待迁移 | P1 |
| `actions/filter_medias.py` | `internal/actions/filter_medias_action.go` | ⏳ 待迁移 | P2 |
| `actions/filter_torrents.py` | `internal/actions/filter_torrents_action.go` | ⏳ 待迁移 | P1 |
| `actions/invoke_plugin.py` | `internal/actions/invoke_plugin_action.go` | ⏳ 待迁移 | P2 |
| `actions/note.py` | `internal/actions/note_action.go` | ⏳ 待迁移 | P3 |
| `actions/send_event.py` | `internal/actions/send_event_action.go` | ⏳ 待迁移 | P2 |
| `actions/send_message.py` | `internal/actions/send_message_action.go` | ⏳ 待迁移 | P1 |

## Chain 层映射 (业务逻辑链)

| Python Chain | Go Service/Chain | 状态 | 优先级 | 说明 |
|-------------|------------------|------|--------|------|
| `chain/media.py` | `internal/business/media/` | 🔄 部分完成 | P0 | 媒体识别核心 |
| `chain/transfer.py` | `internal/business/transfer/` | 🔄 部分完成 | P0 | 转移核心 |
| `chain/storage.py` | `internal/business/storage/` | ✅ 已完成 | P0 | 存储操作 |
| `chain/subscribe.py` | `internal/business/subscribe/` | ⏳ 待迁移 | P0 | 订阅系统 |
| `chain/download.py` | `internal/business/download/` | ⏳ 待迁移 | P0 | 下载管理 |
| `chain/search.py` | `internal/business/search/` | ⏳ 待迁移 | P0 | 搜索服务 |
| `chain/site.py` | `internal/business/site/` | ⏳ 待迁移 | P0 | 站点管理 |
| `chain/torrents.py` | `internal/business/torrents/` | ⏳ 待迁移 | P0 | 种子管理 |
| `chain/message.py` | `internal/business/message/` | ⏳ 待迁移 | P1 | 消息处理 |
| `chain/mediaserver.py` | `internal/business/mediaserver/` | ⏳ 待迁移 | P1 | 媒体服务器 |
| `chain/tmdb.py` | `internal/business/media/tmdb/` | ⏳ 待迁移 | P0 | TMDB 集成 |
| `chain/douban.py` | `internal/business/media/douban/` | ⏳ 待迁移 | P2 | 豆瓣集成 |
| `chain/bangumi.py` | `internal/business/media/bangumi/` | ⏳ 待迁移 | P2 | Bangumi 集成 |
| `chain/recommend.py` | `internal/business/recommend/` | ⏳ 待迁移 | P2 | 推荐系统 |
| `chain/dashboard.py` | `internal/business/dashboard/` | ⏳ 待迁移 | P2 | 仪表板 |
| `chain/user.py` | `internal/business/user/` | ⏳ 待迁移 | P1 | 用户管理 |
| `chain/system.py` | `internal/business/system/` | ⏳ 待迁移 | P2 | 系统管理 |
| `chain/webhook.py` | `internal/business/webhook/` | ⏳ 待迁移 | P2 | Webhook |
| `chain/workflow.py` | `internal/platform/workflow/` | 🔄 部分完成 | P1 | 工作流引擎 |
| `chain/tvdb.py` | `internal/business/media/tvdb/` | ⏳ 待迁移 | P3 | TVDB 集成 |

## API 端点映射

| Python API | Go API | 状态 | 优先级 |
|-----------|--------|------|--------|
| `api/endpoints/login.py` | `internal/apis/auth/` | ⏳ 待迁移 | P0 |
| `api/endpoints/user.py` | `internal/apis/user/` | ⏳ 待迁移 | P0 |
| `api/endpoints/subscribe.py` | `internal/apis/subscribe/` | ⏳ 待迁移 | P0 |
| `api/endpoints/download.py` | `internal/apis/download/` | ⏳ 待迁移 | P0 |
| `api/endpoints/search.py` | `internal/apis/search/` | ⏳ 待迁移 | P0 |
| `api/endpoints/site.py` | `internal/apis/site/` | ⏳ 待迁移 | P0 |
| `api/endpoints/transfer.py` | `internal/apis/transfer/` | ⏳ 待迁移 | P0 |
| `api/endpoints/media.py` | `internal/apis/media/` | ⏳ 待迁移 | P0 |
| `api/endpoints/workflow.py` | `internal/apis/workflow/` | 🔄 部分完成 | P0 |
| `api/endpoints/mediaserver.py` | `internal/apis/mediaserver/` | ⏳ 待迁移 | P1 |
| `api/endpoints/message.py` | `internal/apis/message/` | ⏳ 待迁移 | P1 |
| `api/endpoints/plugin.py` | `internal/apis/plugin/` | ⏳ 待迁移 | P1 |
| `api/endpoints/storage.py` | `internal/apis/storage/` | ⏳ 待迁移 | P1 |
| `api/endpoints/history.py` | `internal/apis/history/` | ⏳ 待迁移 | P1 |
| `api/endpoints/dashboard.py` | `internal/apis/dashboard/` | ⏳ 待迁移 | P2 |
| `api/endpoints/system.py` | `internal/apis/system/` | ⏳ 待迁移 | P2 |
| `api/endpoints/tmdb.py` | `internal/apis/tmdb/` | ⏳ 待迁移 | P2 |
| `api/endpoints/douban.py` | `internal/apis/douban/` | ⏳ 待迁移 | P2 |
| `api/endpoints/bangumi.py` | `internal/apis/bangumi/` | ⏳ 待迁移 | P2 |
| `api/endpoints/recommend.py` | `internal/apis/recommend/` | ⏳ 待迁移 | P2 |
| `api/endpoints/discover.py` | `internal/apis/discover/` | ⏳ 待迁移 | P3 |
| `api/endpoints/torrent.py` | `internal/apis/torrent/` | ⏳ 待迁移 | P1 |
| `api/endpoints/webhook.py` | `internal/apis/webhook/` | ⏳ 待迁移 | P2 |

## 核心组件映射

| Python 组件 | Go 组件 | 状态 | 优先级 |
|-----------|---------|------|--------|
| `core/config.py` | `config/` | ✅ 已完成 | - |
| `core/cache.py` | `pkg/utils/redis.go` | ✅ 已完成 | - |
| `core/event.py` | `internal/platform/event/` | ⏳ 待迁移 | P0 |
| `core/metainfo.py` | `pkg/utils/metainfo.go` | ⏳ 待迁移 | P0 |
| `core/module.py` | `pkg/plugin/` | 🔄 部分完成 | P1 |
| `core/plugin.py` | `pkg/plugin/` | 🔄 部分完成 | P1 |
| `core/context.py` | `internal/models/context.go` | ⏳ 待迁移 | P0 |
| `log.py` | `pkg/logger/` | ✅ 已完成 | - |
| `monitor.py` | `internal/platform/monitor/` | ⏳ 待迁移 | P1 |
| `scheduler.py` | `internal/platform/scheduler/` | ⏳ 待迁移 | P1 |
| `command.py` | `internal/platform/command/` | ⏳ 待迁移 | P2 |

## Modules 映射 (第三方服务集成)

### 下载器模块
| Python | Go | 状态 |
|--------|-----|------|
| `modules/qbittorrent/` | `pkg/utils/downloader.go` (qBittorrent) | ✅ 已完成 |
| `modules/transmission/` | `pkg/utils/downloader.go` (Transmission) | ✅ 已完成 |

### 媒体服务器模块
| Python | Go | 状态 |
|--------|-----|------|
| `modules/emby/` | `pkg/utils/mediaserver.go` (Emby) | 🔄 部分完成 |
| `modules/jellyfin/` | `pkg/utils/mediaserver.go` (Jellyfin) | 🔄 部分完成 |
| `modules/plex/` | `pkg/utils/mediaserver.go` (Plex) | 🔄 部分完成 |

### 元数据源模块
| Python | Go | 状态 |
|--------|-----|------|
| `modules/themoviedb/` (34 文件) | `internal/business/media/tmdb/` | ⏳ 待迁移 |
| `modules/douban/` | `internal/business/media/douban/` | ⏳ 待迁移 |
| `modules/bangumi/` | `internal/business/media/bangumi/` | ⏳ 待迁移 |
| `modules/thetvdb/` | `internal/business/media/tvdb/` | ⏳ 待迁移 |
| `modules/fanart/` | `internal/business/media/fanart/` | ⏳ 待迁移 |

### 通知模块
| Python | Go | 状态 |
|--------|-----|------|
| `modules/telegram/` | `pkg/utils/notification.go` (Telegram) | 🔄 部分完成 |
| `modules/wechat/` | `pkg/utils/notification.go` (WeChat) | 🔄 部分完成 |
| `modules/slack/` | `pkg/utils/notification.go` (Slack) | 🔄 部分完成 |
| `modules/webpush/` | `pkg/utils/notification.go` (WebPush) | ⏳ 待迁移 |
| `modules/vocechat/` | `pkg/utils/notification.go` (VoceChat) | ⏳ 待迁移 |
| `modules/synologychat/` | `pkg/utils/notification.go` (SynologyChat) | ⏳ 待迁移 |

### 索引器模块
| Python | Go | 状态 |
|--------|-----|------|
| `modules/indexer/` (25 个站点) | `internal/business/indexer/` | ⏳ 待迁移 |

### 其他模块
| Python | Go | 状态 |
|--------|-----|------|
| `modules/filemanager/` | `pkg/utils/file.go`, `storage.go` | ✅ 已完成 |
| `modules/filter/` | `pkg/utils/rule.go` | ✅ 已完成 |
| `modules/subtitle/` | `pkg/utils/subtitle.go` | ⏳ 待迁移 |
| `modules/postgresql/` | `pkg/database/` | ✅ 已完成 |
| `modules/redis/` | `pkg/utils/redis.go` | ✅ 已完成 |
| `modules/trimemedia/` | `internal/business/media/` | ⏳ 待迁移 |

## 工具函数映射

| Python Utils | Go Utils | 状态 |
|-------------|----------|------|
| `utils/` (20 个文件) | `pkg/utils/` (37 个文件) | 🔄 大部分完成 |

---

## 优先级说明

- **P0**: 核心功能，必须优先完成
- **P1**: 重要功能，第二批完成
- **P2**: 增强功能，第三批完成
- **P3**: 可选功能，最后完成

## 状态说明

- ✅ **已完成**: 功能完整实现并测试通过
- 🔄 **部分完成**: 基础框架完成，部分功能待实现
- ⏳ **待迁移**: 尚未开始迁移

---

## 迁移建议

### 第一批 (P0 - Week 1-10)
1. 完善 Media Chain (TMDB 集成)
2. Subscribe Chain 完整实现
3. Download Chain 完整实现
4. Search Chain 完整实现
5. Site Chain 完整实现
6. Torrents Chain 完整实现
7. 对应的 API 端点

### 第二批 (P1 - Week 11-18)
1. MediaServer Chain
2. Message Chain
3. Event 系统
4. Monitor 系统
5. Scheduler 系统
6. Plugin 系统完善

### 第三批 (P2 - Week 19-24)
1. Douban/Bangumi 集成
2. Recommend Chain
3. Dashboard Chain
4. Webhook Chain
5. System Chain

### 第四批 (P3 - Week 25-28)
1. TVDB 集成
2. 其他可选功能
3. 性能优化
4. 部署自动化
