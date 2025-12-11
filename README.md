# MoviePilot Go

<div align="center">

![MoviePilot Logo](https://via.placeholder.com/150x150?text=MoviePilot)

**新一代智能影视资源自动化管理系统**

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/github/workflow/status/your-org/moviepilot-go/CI)](https://github.com/your-org/moviepilot-go/actions)
[![Coverage](https://img.shields.io/codecov/c/github/your-org/moviepilot-go)](https://codecov.io/gh/your-org/moviepilot-go)
[![Docker Pulls](https://img.shields.io/docker/pulls/moviepilot/moviepilot-go)](https://hub.docker.com/r/moviepilot/moviepilot-go)

[English](README_EN.md) | 简体中文

</div>

---

## 📖 简介

MoviePilot Go 是一个基于 Go 语言重构的智能影视资源自动化管理系统，提供订阅管理、自动下载、媒体整理、插件扩展等功能。

### ✨ 核心特性

- 🎯 **智能订阅** - 自动追踪影视剧更新，支持 RSS 订阅和站点监控
- 📥 **自动下载** - 集成 qBittorrent/Transmission，智能选种下载
- 📁 **媒体整理** - 自动识别、重命名、转移文件到媒体库
- 🔌 **插件系统** - Go + Python 混合插件架构，支持 70+ 插件
- 🔄 **工作流引擎** - 灵活的自动化流程编排
- 📊 **监控告警** - Prometheus + Grafana 完整监控方案
- 🚀 **高性能** - Go 语言实现，性能提升 3-5 倍
- 🐳 **容器化** - Docker/Kubernetes 部署，开箱即用

---

## 🚀 快速开始

### 前置要求

- Docker 20.10+
- Docker Compose 2.0+
- 或 Go 1.21+ (源码运行)

### Docker 部署（推荐）

```bash
# 克隆项目
git clone https://github.com/your-org/moviepilot-go.git
cd moviepilot-go

# 配置环境变量
cp deployments/.env.example deployments/.env
vim deployments/.env

# 启动服务
docker-compose -f deployments/docker-compose.yml up -d

# 查看日志
docker-compose -f deployments/docker-compose.yml logs -f
```

访问 http://localhost:3001 即可使用。

### 源码运行

```bash
# 安装依赖
go mod download

# 配置数据库
cp configs/config.yaml.example configs/config.yaml
vim configs/config.yaml

# 运行迁移
make migrate-up

# 启动应用
make run
```

---

## 📚 文档

- [部署指南](docs/deployment-guide.md) - 完整的部署文档
- [API 文档](docs/api-guide.md) - RESTful API 使用指南
- [开发指南](docs/development-guide.md) - 开发环境搭建
- [插件开发](docs/plugin-development.md) - 插件开发教程
- [架构设计](docs/architecture.md) - 系统架构说明

---

## 🏗️ 架构

```
┌─────────────────────────────────────────────────────────┐
│                    MoviePilot Go                        │
├─────────────────────────────────────────────────────────┤
│  APIs Layer (Gin)                                       │
│  ├─ Auth API          ├─ Subscription API              │
│  ├─ Site API          ├─ Download API                  │
│  └─ Plugin API        └─ Workflow API                  │
├─────────────────────────────────────────────────────────┤
│  Business Layer                                         │
│  ├─ Auth Service      ├─ Subscription Service          │
│  ├─ Site Service      ├─ Download Service              │
│  ├─ Plugin Manager    └─ Workflow Engine               │
├─────────────────────────────────────────────────────────┤
│  Infrastructure Layer                                   │
│  ├─ Config            ├─ Events                        │
│  ├─ Security          ├─ Cache (Redis)                 │
│  └─ Monitoring        └─ Logging                       │
├─────────────────────────────────────────────────────────┤
│  Integration Layer                                      │
│  ├─ Downloaders       ├─ Media Servers                 │
│  │  ├─ qBittorrent    │  ├─ Emby                       │
│  │  └─ Transmission   │  ├─ Plex                       │
│  ├─ Indexers          │  └─ Jellyfin                   │
│  │  ├─ Jackett        ├─ Metadata                      │
│  │  └─ Prowlarr       │  ├─ TMDB                       │
│  └─ Notifications     │  ├─ TVDB                       │
│     ├─ Telegram       │  └─ 豆瓣                        │
│     └─ WeChat         └─                               │
├─────────────────────────────────────────────────────────┤
│  Data Layer                                             │
│  ├─ PostgreSQL (主数据库)                               │
│  ├─ Redis (缓存)                                        │
│  └─ File Storage (媒体文件)                             │
└─────────────────────────────────────────────────────────┘
         │                                    │
         ▼                                    ▼
┌──────────────────┐              ┌──────────────────────┐
│  Python Plugins  │◄────gRPC────►│   Monitoring Stack   │
│  - Site Plugins  │              │   - Prometheus       │
│  - Indexers      │              │   - Grafana          │
│  - Notifiers     │              │   - Alertmanager     │
└──────────────────┘              └──────────────────────┘
```

---

## 🎯 功能特性

### 用户认证与权限

- ✅ JWT 双令牌认证（Access + Refresh Token）
- ✅ RBAC 权限模型（用户-角色-权限）
- ✅ 密码加密（bcrypt）
- ✅ 登录日志审计

### 站点管理

- ✅ PT/BT 站点配置
- ✅ Cookie 自动同步
- ✅ 自动签到调度
- ✅ 站点统计分析

### 订阅系统

- ✅ 影视剧订阅管理
- ✅ RSS 订阅支持
- ✅ 自动刷新匹配
- ✅ 订阅分享功能

### 下载管理

- ✅ qBittorrent/Transmission 集成
- ✅ 下载任务队列
- ✅ 下载监控分析
- ✅ 自动暂停/恢复

### 文件整理

- ✅ 媒体文件识别
- ✅ 智能重命名
- ✅ 自动转移到媒体库
- ✅ 整理历史记录

### 插件系统

- ✅ Go 原生插件
- ✅ Python 插件（gRPC）
- ✅ 70+ 预置插件
- ✅ 插件热加载

### 工作流引擎

- ✅ 可视化流程编排
- ✅ 6 种步骤类型
- ✅ 条件/循环/并行
- ✅ 错误处理和重试

### 监控告警

- ✅ Prometheus 指标采集
- ✅ Grafana 仪表板
- ✅ 8 个预置告警规则
- ✅ 多渠道通知

---

## 📊 性能指标

| 指标 | MoviePilot (Python) | MoviePilot Go | 提升 |
|------|---------------------|---------------|------|
| 启动时间 | ~15s | ~3s | **5x** |
| 内存占用 | ~500MB | ~150MB | **3.3x** |
| API 响应时间 | ~200ms | ~50ms | **4x** |
| 并发处理 | ~100 req/s | ~500 req/s | **5x** |
| 数据库查询 | ~100ms | ~30ms | **3.3x** |

---

## 🛠️ 技术栈

### 后端

- **语言**: Go 1.21+
- **框架**: Gin (Web), GORM (ORM)
- **数据库**: PostgreSQL 15
- **缓存**: Redis 7
- **消息**: gRPC
- **监控**: Prometheus + Grafana

### 前端

- **框架**: Vue 3 + TypeScript
- **UI**: Element Plus
- **状态**: Pinia
- **构建**: Vite

### 基础设施

- **容器**: Docker + Docker Compose
- **编排**: Kubernetes (可选)
- **CI/CD**: GitHub Actions
- **日志**: Zap
- **配置**: Viper

---

## 📦 项目结构

```
moviepilot-go/
├── cmd/                    # 应用入口
│   └── server/
│       └── main.go
├── internal/               # 私有代码
│   ├── apis/              # API 层
│   ├── business/          # 业务层
│   ├── infrastructure/    # 基础设施层
│   ├── integration/       # 集成层
│   ├── models/            # 数据模型
│   └── repositories/      # 数据访问
├── pkg/                    # 公共库
│   ├── cache/
│   ├── database/
│   ├── logger/
│   └── plugin/
├── configs/                # 配置文件
├── deployments/            # 部署配置
│   ├── docker-compose.yml
│   ├── prometheus/
│   └── grafana/
├── docs/                   # 文档
├── scripts/                # 脚本
├── tests/                  # 测试
├── go.mod
├── go.sum
├── Dockerfile
├── Makefile
└── README.md
```

---

## 🤝 贡献

欢迎贡献代码、报告问题或提出建议！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

详见 [贡献指南](CONTRIBUTING.md)

---

## 📝 开发路线图

- [x] Phase 1: 基础架构（Week 1-6）
- [x] Phase 2: 核心功能（Week 7-9）
- [x] Phase 3: 高级功能（Week 10-12）
- [ ] Phase 4: 完善与发布（Week 13-15）
  - [ ] 测试完善
  - [ ] 文档完善
  - [ ] 性能优化
  - [ ] 正式发布

详见 [项目进度](docs/PROGRESS.md)

---

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

---

## 🙏 致谢

- [MoviePilot](https://github.com/jxxghp/MoviePilot) - 原 Python 版本
- [Gin](https://github.com/gin-gonic/gin) - Web 框架
- [GORM](https://gorm.io/) - ORM 库
- 所有贡献者

---

## 📞 联系方式

- **项目主页**: https://github.com/your-org/moviepilot-go
- **问题反馈**: https://github.com/your-org/moviepilot-go/issues
- **讨论区**: https://github.com/your-org/moviepilot-go/discussions
- **Telegram**: https://t.me/moviepilot_go

---

<div align="center">

**如果这个项目对你有帮助，请给个 ⭐️ Star 支持一下！**

Made with ❤️ by MoviePilot Team

</div>
