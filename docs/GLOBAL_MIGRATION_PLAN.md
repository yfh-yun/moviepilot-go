# MoviePilot Python → Go 全局分阶段迁移方案

> **项目目标**: 将 MoviePilot Python 2.8.1-1 完整迁移至 Go 微服务架构

---

## 📊 当前完成度: 30%

### ✅ 已完成模块

#### 1. 基础架构 (100%)
- 项目骨架、配置系统、日志系统
- 数据库层 (PostgreSQL + GORM)
- 缓存系统 (Redis)
- HTTP 客户端、JWT 认证、参数验证

#### 2. 数据模型 (80%)
- 核心模型: User, Site, Subscribe, DownloadHistory, TransferHistory, Media
- 文件模型: FileItem, TransferHistory
- Cookie 模型

#### 3. 数据访问层 (70%)
- 所有 Repository 接口定义
- PostgreSQL 实现完成

#### 4. 工具函数 (85%)
- 文件操作、媒体服务器、下载器、种子处理
- 订阅管理、通知系统、RSS、规则引擎
- 加密、Cookie、NFO、OCR、壁纸等

#### 5. Actions 框架 (100%)
- ScanFileAction, ScrapeFileAction
- FetchTorrentsAction, TransferFileAction

#### 6. 业务服务 (40%)
- Storage Service, Media Service, Transfer Service

#### 7. 监控系统 (100%)
- Metrics, Health Check, Prometheus

#### 8. 插件系统 (60%)
- 插件接口、管理器、Python 桥接

#### 9. 工作流系统 (70%)
- Workflow 框架、本地文件工作流、CLI 入口

### 🚧 待迁移模块 (70%)

**Python 源码统计**:
- Actions: 16 个 (4 个已完成，12 个待迁移)
- Chains: 21 个 (全部待迁移)
- API Endpoints: 28 个 (全部待迁移)
- Modules: 108 个 (部分已迁移)

---

## 📅 分阶段迁移计划

## 第一阶段: 核心链路完善 (Week 1-4)

### Week 1: API 入口与测试
- [ ] 完成 Workflow API Handler
- [ ] 路由注册与中间件
- [ ] 单元测试与集成测试

### Week 2: 真实刮削能力
- [ ] TMDB 集成 (迁移 themoviedb 模块)
- [ ] 元数据识别增强
- [ ] NFO 支持

### Week 3: 转移能力精细化
- [ ] 命名规则引擎
- [ ] 目录结构策略
- [ ] 冲突处理

### Week 4: 监控与优化
- [ ] 性能优化 (并发、批量)
- [ ] 监控指标与 Grafana
- [ ] 文档与部署

---

## 第二阶段: 订阅与下载链路 (Week 5-10)

### Week 5-6: 订阅系统
- [ ] Subscribe Service 实现
- [ ] 订阅匹配算法
- [ ] 缺集检测
- [ ] API 端点

### Week 7-8: 搜索与种子
- [ ] Site Service (站点管理)
- [ ] Search Service (多站点搜索)
- [ ] Torrents Service (种子解析)

### Week 9-10: 下载系统
- [ ] Download Service
- [ ] qBittorrent/Transmission 实现
- [ ] 下载完成钩子

---

## 第三阶段: 媒体服务器与通知 (Week 11-14)

### Week 11-12: 媒体服务器
- [ ] MediaServer Service
- [ ] Emby/Jellyfin/Plex 实现
- [ ] 库同步

### Week 13-14: 消息通知
- [ ] Message Service
- [ ] Telegram/WeChat/Slack 等渠道

---

## 第四阶段: 调度与监控 (Week 15-18)

### Week 15-16: 定时任务
- [ ] Scheduler Service (基于 cron)
- [ ] 订阅刷新、站点刷新等任务

### Week 17-18: 文件监控
- [ ] Monitor Service (基于 fsnotify)
- [ ] 自动化工作流

---

## 第五阶段: 插件系统与扩展 (Week 19-24)

### Week 19-21: 插件系统
- [ ] 插件生命周期管理
- [ ] Python 插件桥接优化
- [ ] Go 原生插件开发

### Week 22-24: 高级功能
- [ ] Douban/Bangumi 集成
- [ ] 推荐系统
- [ ] Dashboard
- [ ] Webhook 系统

---

## 第六阶段: 性能优化与部署 (Week 25-28)

### Week 25-26: 性能优化
- [ ] 并发优化、数据库优化
- [ ] 缓存优化、压力测试

### Week 27-28: 部署自动化
- [ ] CI/CD 完善
- [ ] Kubernetes 支持
- [ ] 监控告警

---

## 📈 迁移统计

| 模块 | 总数 | 已完成 | 完成率 |
|------|------|--------|--------|
| Actions | 16 | 4 | 25% |
| Chains | 21 | 0 | 0% |
| API Endpoints | 28 | 0 | 0% |
| Modules | 108 | 30 | 28% |
| 总计 | 211 | 59 | **28%** |

---

## 🎯 关键里程碑

- **M1 (Week 4)**: MVP 发布 - 本地文件链路可用
- **M2 (Week 10)**: 核心功能 - 订阅下载链路可用
- **M3 (Week 18)**: 功能对齐 - 媒体服务器、通知、调度
- **M4 (Week 28)**: 生产就绪 - 性能优化、部署自动化

---

## 📝 迁移原则

1. **竖切优先**: 优先打通完整业务链路
2. **渐进式**: 保持 Python 版本可运行
3. **测试驱动**: 每个模块必须有测试
4. **性能优先**: 利用 Go 并发特性
5. **兼容性**: 保持 API 接口兼容
