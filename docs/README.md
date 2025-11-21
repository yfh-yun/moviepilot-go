# MoviePilot Go 项目文档

## 📋 项目概述

MoviePilot Go 是一个现代化的媒体资源管理系统，采用 Go 语言重构的微服务架构。本项目从原始 Python 版本迁移而来，在性能、可维护性和部署效率方面有显著提升。

## 🏗️ 架构概览

```
moviepilot-go/                       # Go主应用
├── cmd/server/main.go               # 应用入口
├── internal/                        # 私有代码
│   ├── apis/                        # API层
│   │   ├── handlers/                # HTTP处理器
│   │   ├── middlewares/             # 中间件
│   │   └── routes/                  # 路由定义
│   ├── business/                    # 业务层
│   │   ├── domains/                 # 领域模型
│   │   ├── services/                # 业务服务
│   │   └── workflows/               # 工作流
│   ├── infrastructure/              # 基础设施层
│   ├── models/                      # 数据模型
│   ├── repositories/                # 数据访问层
│   └── monitor/                     # 监控系统
├── pkg/                             # 公共库
│   ├── database/                    # 数据库连接
│   ├── cache/                       # 缓存封装
│   ├── logger/                      # 日志封装
│   └── plugin/                      # 插件系统
├── configs/                         # 配置文件
└── deployments/                     # 部署配置
```

## 🚀 快速开始

### 环境要求

- Go 1.21+
- Docker & Docker Compose
- PostgreSQL 14+
- Redis 6+

### 本地开发

```bash
# 克隆项目
git clone https://github.com/yfh-yun/moviepilot-go.git
cd moviepilot-go

# 安装依赖
go mod download

# 启动开发环境
make dev

# 或使用 Docker Compose
docker-compose -f deployments/docker-compose.dev.yml up -d
```

### 配置说明

1. 复制配置文件：`cp configs/config.yaml.sample configs/config.yaml`
2. 修改数据库连接、Redis 等配置
3. 启动应用：`go run cmd/server/main.go`

## 📚 文档导航

### 核心文档
- [架构设计](./architecture/README.md) - 系统架构和设计原则
- [API文档](./api/README.md) - REST API 接口说明
- [部署指南](./deployment/README.md) - 部署和运维指南

### 开发文档
- [开发规范](./development/README.md) - 代码规范和最佳实践
- [插件开发](./plugins/README.md) - 插件系统开发指南
- [数据库设计](./database/README.md) - 数据模型和迁移

### 运维文档
- [监控告警](./monitoring/README.md) - 系统监控和告警配置
- [性能优化](./performance/README.md) - 性能调优指南
- [故障排查](./troubleshooting/README.md) - 常见问题解决

## 🔧 技术栈

### 后端技术
- **框架**: Gin (HTTP), gRPC (服务间通信)
- **ORM**: GORM
- **数据库**: PostgreSQL (主), Redis (缓存)
- **日志**: Zap + 自定义封装
- **配置**: Viper
- **容器化**: Docker + Docker Compose

### 监控运维
- **监控**: Prometheus + Grafana
- **日志**: 结构化日志 (JSON格式)
- **链路追踪**: OpenTelemetry
- **健康检查**: 内置健康检查端点

## 📊 性能指标

| 指标 | 原Python版本 | Go版本 | 提升比例 |
|------|-------------|--------|----------|
| 启动时间 | ~8-12秒 | ~1-2秒 | **5-6x** |
| 内存占用 | ~200-300MB | ~50-100MB | **2-3x** |
| 并发处理 | ~100 req/s | ~1000+ req/s | **10x+** |
| CPU使用率 | 高负载下60-80% | 高负载下20-30% | **2-3x** |

## 🔄 迁移状态

### ✅ 已完成
- [x] 基础架构搭建
- [x] 用户系统迁移
- [x] 监控系统集成
- [x] Docker 容器化
- [x] CI/CD 流水线

### 🚧 进行中
- [ ] 插件系统重构
- [ ] 媒体管理功能
- [ ] 订阅系统迁移

### 📋 待开始
- [ ] 搜索功能优化
- [ ] 通知系统迁移
- [ ] 性能基准测试

## 🤝 贡献指南

1. Fork 项目
2. 创建特性分支：`git checkout -b feature/amazing-feature`
3. 提交更改：`git commit -m 'Add amazing feature'`
4. 推送分支：`git push origin feature/amazing-feature`
5. 提交 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 📞 联系方式

- 项目主页: https://github.com/yfh-yun/moviepilot-go
- 问题反馈: https://github.com/yfh-yun/moviepilot-go/issues
- 文档站点: https://docs.moviepilot-go.com

---

**注意**: 本项目正在积极开发中，部分功能可能不稳定。建议在生产环境使用前进行充分测试。