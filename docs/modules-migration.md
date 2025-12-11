# modules/ 外部服务模块迁移计划

> Python: `app/modules/`  \
> Go: `internal/platform/` + `internal/integration/`

---

## 1. Python `modules/` 概览

- **职责**：对外部服务做封装：
  - 媒体服务器：Emby / Plex / Jellyfin
  - 下载器：qBittorrent / Transmission
  - 通知渠道：Telegram / WeChat / Slack 等
  - 索引器：Jackett / Prowlarr
  - 第三方 API：TMDB 等（部分已拆到 core/meta）

---

## 2. Go 目标结构

- **平台适配层**：`internal/platform/`
  - 聚焦“媒体服务器”等长连接或状态丰富的外部系统。
- **集成层**：`internal/integration/`
  - 聚焦“HTTP API 封装”，如下载器、通知平台等。

示例结构：

```text
internal/platform/
  tmdb/
  emby/
  plex/
  jellyfin/

internal/integration/
  qbittorrent/
  transmission/
  telegram/
  wechat/
  indexer/
```

---

## 3. 映射表（按 app/modules 真实结构）

> 目录来自 `MoviePilot-2.8.1-1/app/modules/`，后续可按子模块文件继续细化。

| Python 子目录/模块 | Go 目标位置 | 状态 | 说明 |
|---------------------|-------------|------|------|
| `bangumi/` | `internal/business/services/bangumi/` + `internal/platform/bangumi/` | ⏳ 规划中 | Bangumi 数据源封装（部分已在 domains/media 中） |
| `douban/` | `internal/business/services/douban/` + `internal/platform/douban/` | ⏳ 规划中 | 豆瓣抓取/评分接口 |
| `emby/` | `internal/platform/emby/` | ⏳ 规划中 | Emby 媒体服务器客户端 |
| `fanart/` | `internal/platform/fanart/` | ⏳ 规划中 | Fanart.tv / 海报图像服务 |
| `filemanager/` | `internal/integration/filemanager/` 或 `internal/infrastructure/storage/` | ⏳ 规划中 | 文件移动/删除/清理等文件管理操作 |
| `filter/` | `internal/business/policies/filter/` | ⏳ 规划中 | 资源筛选规则，可与 rule 系统整合 |
| `indexer/` | `internal/integration/indexer/` | ⏳ 规划中 | Jackett / Prowlarr 等索引器客户端 |
| `jellyfin/` | `internal/platform/jellyfin/` | ⏳ 规划中 | Jellyfin 媒体服务器客户端 |
| `plex/` | `internal/platform/plex/` | ⏳ 规划中 | Plex 媒体服务器客户端 |
| `postgresql/` | `internal/infrastructure/db/postgresql/` | ⏳ 规划中 | PostgreSQL 特定功能封装（如扩展/维护脚本） |
| `qbittorrent/` | `internal/integration/qbittorrent/` | ⏳ 规划中 | qBittorrent HTTP API 封装 |
| `redis/` | `pkg/cache` + `internal/infrastructure/cache/` | ⏳ 规划中 | Redis 专用工具，部分已在 `pkg/cache` 中实现 |
| `slack/` | `internal/integration/slack/` | ⏳ 规划中 | Slack 通知集成 |
| `subtitle/` | `internal/integration/subtitle/` + `internal/business/services/media/` | ⏳ 规划中 | 字幕下载/管理 |
| `synologychat/` | `internal/integration/synologychat/` | ⏳ 规划中 | 群晖 Chat 通知 |
| `telegram/` | `internal/integration/telegram/` | ⏳ 规划中 | Telegram Bot / 通知 |
| `themoviedb/` | `internal/business/services/media/tmdb/` + `internal/platform/tmdb/` | ✅ 核心已迁，平台层待补 | TMDB API 已在 `services/media/tmdb` 中实现，后续可抽离部分为 platform 层 |
| `thetvdb/` | `internal/business/services/tvdb/` + `internal/platform/thetvdb/` | ⏳ 部分已迁 | 已有 TVDB Service，平台适配层待设计 |
| `transmission/` | `internal/integration/transmission/` | ⏳ 规划中 | Transmission RPC 客户端 |
| `trimemedia/` | `internal/integration/trimemedia/` 或并入 `mediaserver/` | ⏳ 规划中 | 第三方媒体管理服务 |
| `vocechat/` | `internal/integration/vocechat/` | ⏳ 规划中 | VoceChat 通知集成 |
| `webpush/` | `internal/integration/webpush/` | ⏳ 规划中 | WebPush 通知/浏览器推送 |
| `wechat/` | `internal/integration/wechat/` | ⏳ 规划中 | WeChat 推送/机器人 |

---

## 4. 设计原则

1. **接口优先**：
   - 在 `platform`/`integration` 定义 interface，由业务 Service 依赖接口，而不是具体实现。
2. **配置驱动**：
   - 所需 URL、密钥等配置来源于 `internal/infrastructure/config`。
3. **错误转换**：
   - 技术错误（网络超时、解析失败）转换为领域可理解的错误。
4. **日志与重试**：
   - 统一使用 `pkg/logger`，为关键外部调用增加重试/熔断策略（可放到后续优化阶段）。

---

## 5. 迁移步骤

1. 为每类外部系统定义最小接口（例如 `MediaServerClient`、`DownloaderClient`）。
2. 将 Python `modules/*` 中的逻辑按系统拆分到对应包中：
   - 只保留“调用外部服务”的代码，不混入业务决策。
3. 在业务 Service 中依赖这些接口：
   - 例如 `mediaserver.Service` 注入 `MediaServerClient`。
4. 为关键外部交互编写集成测试（可使用 mock server）。

---

## 6. 检查清单

- [ ] 所有 `app/modules/` 文件在映射表中有对应条目。
- [ ] 业务 Service 不直接依赖具体第三方 SDK，而是依赖接口。
- [ ] 外部请求统一经过配置与日志封装。
