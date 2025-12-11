# helper/ 与 utils/ 模块迁移计划

> Python: `app/helper/` + `app/utils/`  \
> Go: `pkg/` + `internal/infrastructure/` + 各业务 `services/`

---

## 1. Python `helper/` 与 `utils/` 概览

- **helper/**：
  - 浏览器自动化（`browser.py`）
  - Cookie 与 CookieCloud（`cookie.py`, `cookiecloud.py`）
  - RSS 解析（`rss.py`）
  - 消息发送（`message.py`）
  - 站点与种子解析（`site.py`, `torrent.py` 等）
- **utils/**：
  - 通用工具函数（字符串处理、时间、路径、网络等）
  - 与业务弱相关或无关

---

## 2. 迁移原则

1. **通用工具** → `pkg/utils/` 下按领域拆分：
   - 时间/日期：`pkg/utils/timeutil`（示例）
   - 字符串处理：`pkg/utils/strutil`
2. **基础设施类 helper** → `internal/infrastructure/`：
   - Cookie、网络、浏览器自动化、监控等。
3. **强业务相关 helper** → 对应 `internal/business/services/`：
   - 站点业务逻辑、种子解析逻辑等。
4. **日志**：统一通过 `pkg/logger`，禁止在 helper 中使用 `fmt.Println`。

---

## 3. 映射总表（初版）

| Python Helper/Utils | Go 目标位置 | 状态 | 备注 |
|---------------------|-------------|------|------|
| `helper/browser.py` | `pkg/browser/` | ⏳ 规划中 | 浏览器自动化（chromedp/rod 等） |
| `helper/cookie.py` | `internal/infrastructure/cookie/` | ⏳ 规划中 | Cookie 读写、持久化 |
| `helper/cookiecloud.py` | `internal/business/services/site/` | ⏳ 规划中 | CookieCloud 同步逻辑 |
| `helper/rss.py` | `pkg/rss/` | ⏳ 规划中 | RSS 解析（gofeed） |
| `helper/message.py` | `internal/business/services/notification/` | ⏳ 规划中 | 通知发送统一封装 |
| `helper/site.py` | `internal/business/services/site/` | ⏳ 规划中 | 站点登录、检测、请求封装 |
| `helper/torrent.py` | `pkg/torrent/` + `internal/business/domains/media/` | ⏳ 规划中 | 种子元信息解析 |
| `utils/time.py` | `pkg/utils/timeutil/` | ⏳ 规划中 | 时间处理工具 |
| `utils/string.py` | `pkg/utils/strutil/` | ⏳ 规划中 | 字符串处理 |

> 上表只是起点，可在实际梳理时按文件粒度继续细化，并更新状态列。

---

## 4. Go 侧目录建议

- `pkg/utils/`：
  - 只放“无业务语义”的通用工具包。
  - 包名小写、语义明确，例如 `timeutil`, `pathutil`, `strutil` 等。
- `internal/infrastructure/`：
  - `cookie/`：Cookie 存储与管理。
  - `network/`（若后续拆分）：代理、HTTP 客户端配置。
  - `browser/`：浏览器控制（如有）。
- `internal/business/services/`：
  - `site/`：站点相关业务逻辑。
  - `notification/`：消息/通知逻辑。

---

## 5. 迁移步骤

1. 列出 `app/helper/` 与 `app/utils/` 中所有文件与函数。
2. 对每个文件评估其职责：
   - 与“技术细节/外部服务”相关 → `internal/infrastructure/`。
   - 与“领域业务”相关 → 对应 `internal/business/services/` 或 `domains/`。
   - 纯工具 → `pkg/utils/`。
3. 为每个目标目录设计最小可行 API，并将原有函数/类重写为 Go 函数/结构体。
4. 替换业务代码中的旧调用：
   - Python 调用 → Go 侧通过 Service/工具函数实现。
5. 删除不再使用的历史 helper/utils 入口（完成迁移后）。

---

## 6. 检查清单

- [ ] `app/helper/` 中所有文件在映射表中都有一条记录。
- [ ] `app/utils/` 中所有公共函数都有明确迁移目标。
- [ ] Go 侧不出现“万能 util 包”，而是按领域拆分。
- [ ] helper 中不再直接依赖日志实现，而统一通过 `pkg/logger`。
- [ ] 新增的 Go 包均遵守命名与分层规范（不跨层依赖）。
