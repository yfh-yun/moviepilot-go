# MoviePilot Go

MoviePilot Go语言版本 - 自动化媒体库管理工具

## 项目概述

MoviePilot Go 是一个基于Go语言开发的自动化媒体库管理工具，支持电影、电视剧的自动下载、整理、刮削等功能。采用微服务架构，Go主应用 + Python插件服务，通过gRPC通信。

## 技术栈

- **Go版本**: 1.24.4+
- **Web框架**: Gin
- **ORM**: GORM
- **数据库**: PostgreSQL
- **缓存**: Redis
- **日志**: Zap
- **配置**: Viper
- **容器化**: Docker + Docker Compose
- **API文档**: Swagger
- **插件系统**: gRPC + Python

## 项目结构

```
moviepilot-go/                       # Go主应用
├── cmd/
│   └── server/
│       └── main.go                 # 应用入口
├── internal/                       # 私有代码
│   ├── api/                        # API层
│   │   ├── handlers/               # HTTP处理器
│   │   ├── middleware/             # 中间件
│   │   └── routes/                 # 路由定义
│   ├── service/                    # 业务逻辑层
│   │   ├── user/                   # 用户服务
│   │   ├── subscribe/              # 订阅服务
│   │   ├── download/               # 下载服务
│   │   ├── transfer/               # 转移服务
│   │   └── plugin/                 # 插件管理服务
│   ├── repository/                 # 数据访问层
│   │   ├── postgres/               # PostgreSQL实现
│   │   └── redis/                  # Redis实现
│   ├── model/                      # 数据模型
│   │   ├── user.go
│   │   ├── subscribe.go
│   │   ├── download.go
│   │   └── plugin.go
│   └── config/                     # 配置管理
├── pkg/                            # 公共库
│   ├── database/                   # 数据库连接
│   ├── cache/                      # 缓存封装
│   ├── logger/                     # 日志封装
│   ├── plugin/                     # 插件系统核心
│   ├── utils/                      # 工具函数
│   └── errors/                     # 错误处理
├── configs/                        # 配置文件
├── scripts/                        # 脚本文件
├── deployments/                    # 部署配置
│   ├── docker-compose.yml          # 容器编排
│   ├── docker-compose.dev.yml      # 开发环境
│   └── docker-compose.prod.yml     # 生产环境
├── tests/                          # 测试文件
├── docs/                           # 文档
├── shared/                         # 共享资源
│   ├── proto/                      # gRPC协议定义
│   └── schemas/                    # 数据模式定义
├── go.mod
├── go.sum
├── Dockerfile                      # Go应用容器
└── README.md
```

## 快速开始

### 环境要求

- Go 1.24.4+
- Docker & Docker Compose
- PostgreSQL 14+
- Redis 6+

### 安装运行

1. 克隆项目
```bash
git clone https://github.com/moviepilot/moviepilot-go.git
cd moviepilot-go
```

2. 安装依赖
```bash
go mod download
```

3. 配置环境
```bash
cp configs/config.yaml.sample configs/config.yaml
# 编辑配置文件
```

4. 启动服务
```bash
# 开发环境
docker-compose -f deployments/docker-compose.dev.yml up -d

# 生产环境
docker-compose -f deployments/docker-compose.prod.yml up -d
```

5. 运行应用
```bash
go run cmd/server/main.go
```

### API文档

启动服务后访问: http://localhost:3001/swagger/index.html

## 开发指南

### 代码规范

- 遵循Go官方代码规范
- 使用gofmt格式化代码
- 使用golint检查代码质量
- 所有公共函数必须有注释
- 单元测试覆盖率不低于80%

### 提交规范

- feat: 新功能
- fix: 修复bug
- docs: 文档更新
- style: 代码格式调整
- refactor: 代码重构
- test: 测试相关
- chore: 构建过程或辅助工具的变动

### 分支管理

- main: 主分支，用于生产环境
- develop: 开发分支
- feature/*: 功能分支
- hotfix/*: 热修复分支

## 插件开发

MoviePilot Go 支持插件扩展，插件可以是Go原生插件或Python插件。

### Go插件

在 `plugins/` 目录下创建插件代码

### Python插件

参考 `python-plugins/` 目录下的插件示例

## 部署

### Docker部署

```bash
# 构建镜像
docker build -t moviepilot-go .

# 运行容器
docker run -d -p 3001:3001 moviepilot-go
```

### Kubernetes部署

```bash
kubectl apply -f deployments/kubernetes/
```

## 监控

- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000

## 贡献

欢迎提交Issue和Pull Request！

## 许可证

MIT License

## 联系方式

- 邮箱: support@moviepilot.com
- 网站: http://www.moviepilot.com