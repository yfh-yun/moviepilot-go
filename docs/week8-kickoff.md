# Week 8 启动文档

> **Phase 2 核心功能开发 - Week 8**  
> **任务**: 订阅系统 + 下载管理  
> **开始时间**: 2025-12-03  
> **预计完成**: 2025-12-07

---

## 🎯 本周目标

实现 MoviePilot Go 的订阅系统和下载管理功能，包括：
- 订阅的创建、管理和自动刷新
- 订阅匹配引擎（RSS 解析、种子筛选）
- 下载任务的创建、监控和管理
- 下载状态同步和历史记录

---

## 📋 前置准备

### 已完成的基础（Week 1-7）

**基础设施**:
- ✅ 数据库连接池优化
- ✅ 日志系统
- ✅ 配置管理

**集成模块**:
- ✅ 下载器集成（qBittorrent + Transmission）
- ✅ 媒体服务器集成（Emby + Plex + Jellyfin）
- ✅ 元数据平台集成（TMDB + TVDB + 豆瓣）
- ✅ 通知渠道集成（Telegram + WeChat）
- ✅ 索引器集成（Jackett + Prowlarr）

**核心功能**:
- ✅ 用户认证系统
- ✅ 站点管理系统

### 需要的依赖

**已有**:
- gorm.io/gorm - ORM
- github.com/gin-gonic/gin - Web 框架
- github.com/robfig/cron/v3 - 任务调度
- github.com/golang-jwt/jwt/v5 - JWT 认证

**可能需要新增**:
- github.com/mmcdole/gofeed - RSS 解析
- github.com/PuerkitoBio/goquery - HTML 解析

---

## 🗓️ 5 天实施计划

### Day 1: 订阅数据模型和 Repository（2025-12-03）

**任务**:
1. 创建订阅相关数据库迁移
   - subscriptions 表
   - subscription_items 表
   - subscription_history 表

2. 定义 GORM 模型
   - Subscription 模型
   - SubscriptionItem 模型
   - SubscriptionHistory 模型

3. 实现 Repository 层
   - SubscriptionRepository
   - SubscriptionItemRepository

**交付物**:
- 3 个迁移脚本
- 3 个 GORM 模型
- 2 个 Repository

### Day 2: 订阅服务和匹配引擎（2025-12-04）

**任务**:
1. 实现订阅服务
   - SubscriptionService（CRUD 操作）
   - SubscriptionRefreshService（刷新逻辑）

2. 实现匹配引擎
   - RSSParser（RSS 解析）
   - TorrentMatcher（种子匹配）
   - FilterEngine（过滤规则）

3. 实现订阅调度器
   - SubscriptionRefreshScheduler

**交付物**:
- 2 个服务文件
- 3 个匹配引擎组件
- 1 个调度器

### Day 3: 下载数据模型和 Repository（2025-12-05）

**任务**:
1. 创建下载相关数据库迁移
   - download_tasks 表
   - download_history 表

2. 定义 GORM 模型
   - DownloadTask 模型
   - DownloadHistory 模型

3. 实现 Repository 层
   - DownloadTaskRepository
   - DownloadHistoryRepository

**交付物**:
- 2 个迁移脚本
- 2 个 GORM 模型
- 2 个 Repository

### Day 4: 下载服务和状态同步（2025-12-06）

**任务**:
1. 实现下载服务
   - DownloadService（任务管理）
   - DownloadSyncService（状态同步）

2. 实现下载调度器
   - DownloadSyncScheduler（定期同步）

3. 集成下载器客户端
   - 使用已有的 qBittorrent/Transmission 客户端

**交付物**:
- 2 个服务文件
- 1 个调度器

### Day 5: API 接口和集成测试（2025-12-07）

**任务**:
1. 实现订阅 API
   - SubscriptionHandler（8 个接口）

2. 实现下载 API
   - DownloadHandler（6 个接口）

3. 编写集成测试
   - 订阅流程测试
   - 下载流程测试

4. 更新文档
   - API 文档
   - 使用指南

**交付物**:
- 2 个 API Handler
- 集成测试
- 文档更新

---

## 📊 数据库设计

### 订阅表（subscriptions）

```sql
CREATE TABLE subscriptions (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL,  -- movie, tv
    tmdb_id INT,
    imdb_id VARCHAR(20),
    season INT,
    quality VARCHAR(50),
    filter_rules JSONB,
    enabled BOOLEAN DEFAULT true,
    auto_download BOOLEAN DEFAULT true,
    notify_on_match BOOLEAN DEFAULT false,
    last_refresh_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

### 订阅项表（subscription_items）

```sql
CREATE TABLE subscription_items (
    id SERIAL PRIMARY KEY,
    subscription_id INT NOT NULL,
    title VARCHAR(200) NOT NULL,
    torrent_url VARCHAR(500),
    size BIGINT,
    seeders INT,
    leechers INT,
    publish_date TIMESTAMP,
    matched_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    downloaded BOOLEAN DEFAULT false,
    FOREIGN KEY (subscription_id) REFERENCES subscriptions(id)
);
```

### 下载任务表（download_tasks）

```sql
CREATE TABLE download_tasks (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    subscription_id INT,
    downloader_type VARCHAR(20) NOT NULL,  -- qbittorrent, transmission
    hash VARCHAR(64) UNIQUE NOT NULL,
    name VARCHAR(200) NOT NULL,
    save_path VARCHAR(500),
    size BIGINT,
    status VARCHAR(20) DEFAULT 'downloading',  -- downloading, completed, error, paused
    progress DECIMAL(5,2) DEFAULT 0,
    download_speed BIGINT DEFAULT 0,
    upload_speed BIGINT DEFAULT 0,
    error_message TEXT,
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (subscription_id) REFERENCES subscriptions(id)
);
```

---

## 🎯 核心功能设计

### 订阅系统

**订阅流程**:
1. 用户创建订阅（指定 TMDB ID、质量、过滤规则）
2. 调度器定期刷新订阅（每 30 分钟）
3. RSS 解析器获取最新种子
4. 匹配引擎筛选符合条件的种子
5. 自动下载或通知用户

**匹配规则**:
- 标题匹配（正则表达式）
- 质量匹配（1080p、4K 等）
- 大小范围
- 做种数量
- 发布时间

### 下载管理

**下载流程**:
1. 创建下载任务
2. 调用下载器 API 添加种子
3. 定期同步下载状态（每 5 分钟）
4. 下载完成后通知
5. 记录下载历史

**状态同步**:
- 进度更新
- 速度更新
- 状态变更（下载中、完成、错误）
- 错误处理

---

## 🔌 API 接口设计

### 订阅 API（8 个）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/v1/subscriptions | 创建订阅 |
| GET | /api/v1/subscriptions | 获取订阅列表 |
| GET | /api/v1/subscriptions/:id | 获取订阅详情 |
| PUT | /api/v1/subscriptions/:id | 更新订阅 |
| DELETE | /api/v1/subscriptions/:id | 删除订阅 |
| POST | /api/v1/subscriptions/:id/refresh | 手动刷新订阅 |
| GET | /api/v1/subscriptions/:id/items | 获取订阅项 |
| POST | /api/v1/subscriptions/:id/items/:item_id/download | 下载订阅项 |

### 下载 API（6 个）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/v1/downloads | 创建下载任务 |
| GET | /api/v1/downloads | 获取下载列表 |
| GET | /api/v1/downloads/:id | 获取下载详情 |
| PUT | /api/v1/downloads/:id/pause | 暂停下载 |
| PUT | /api/v1/downloads/:id/resume | 恢复下载 |
| DELETE | /api/v1/downloads/:id | 删除下载 |

---

## 📊 预期成果

### 代码统计

| 类别 | 预计文件数 | 预计代码行数 |
|------|-----------|-------------|
| 数据库迁移 | 10 | - |
| GORM 模型 | 5 | 300 |
| Repository | 4 | 400 |
| 业务服务 | 6 | 800 |
| 匹配引擎 | 3 | 400 |
| 调度器 | 2 | 150 |
| API Handler | 2 | 500 |
| **总计** | **32** | **2,550** |

---

## 🎯 成功标准

### 功能完整性
- ✅ 订阅 CRUD 操作正常
- ✅ 订阅自动刷新工作
- ✅ 匹配引擎准确筛选
- ✅ 下载任务正常创建
- ✅ 下载状态正确同步

### 性能指标
- ✅ 订阅刷新时间 < 30 秒
- ✅ 匹配引擎处理 < 5 秒
- ✅ 状态同步延迟 < 10 秒
- ✅ API 响应时间 < 200ms

### 代码质量
- ✅ 所有代码编译通过
- ✅ 单元测试覆盖率 > 60%
- ✅ 集成测试通过
- ✅ 文档完整

---

## 🚀 开发环境设置

### 数据库准备
```bash
# 运行迁移
make migrate-up

# 插入种子数据
make migrate-seed
```

### 启动服务
```bash
# 开发模式
make dev

# 生产模式
make run
```

### 运行测试
```bash
# 单元测试
make test

# 集成测试
make test-integration
```

---

## 📝 进度跟踪

### Day 1 进度
- [ ] 订阅数据库迁移
- [ ] 订阅 GORM 模型
- [ ] 订阅 Repository

### Day 2 进度
- [ ] 订阅服务
- [ ] 匹配引擎
- [ ] 订阅调度器

### Day 3 进度
- [ ] 下载数据库迁移
- [ ] 下载 GORM 模型
- [ ] 下载 Repository

### Day 4 进度
- [ ] 下载服务
- [ ] 状态同步
- [ ] 下载调度器

### Day 5 进度
- [ ] 订阅 API
- [ ] 下载 API
- [ ] 集成测试
- [ ] 文档更新

---

## 💡 技术要点

### RSS 解析
- 使用 gofeed 库解析 RSS
- 支持多种 RSS 格式
- 提取种子信息

### 种子匹配
- 正则表达式匹配
- 质量识别（分辨率、编码）
- 大小过滤
- 做种数过滤

### 下载管理
- 使用已有的下载器客户端
- 定期同步状态
- 错误重试机制
- 完成通知

---

## 🔍 风险和应对

### 潜在风险

1. **RSS 解析复杂度**
   - 风险：不同站点 RSS 格式差异大
   - 应对：使用成熟的 gofeed 库，支持多格式

2. **匹配准确性**
   - 风险：匹配规则可能不够精确
   - 应对：提供灵活的过滤规则配置

3. **下载器集成**
   - 风险：下载器 API 可能不稳定
   - 应对：添加重试机制和错误处理

4. **性能问题**
   - 风险：大量订阅可能影响性能
   - 应对：使用并发处理和缓存

---

## 📚 参考资料

### 相关文档
- Week 7 完成总结
- 下载器集成文档
- 索引器集成文档

### 技术文档
- gofeed 文档：https://github.com/mmcdole/gofeed
- RSS 2.0 规范
- Torrent 文件格式

---

**准备状态**: ✅ 就绪  
**开始时间**: 2025-12-03  
**预计完成**: 2025-12-07

---

**Week 8，我们来了！** 🚀
