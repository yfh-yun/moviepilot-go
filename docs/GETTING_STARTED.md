# MoviePilot Go - 快速开始指南

## 🚀 项目概述

MoviePilot Go 是一个基于 Go 语言开发的自动化媒体库管理工具，采用现代化的微服务架构设计。

### 🏗️ 技术栈

- **Web框架**: Gin
- **ORM**: GORM
- **数据库**: PostgreSQL
- **缓存**: Redis
- **日志**: Zap
- **配置**: Viper
- **容器化**: Docker + Docker Compose
- **API文档**: Swagger

## 📋 系统要求

- Go 1.21+
- PostgreSQL 12+
- Redis 6+
- Docker & Docker Compose (可选)

## 🛠️ 安装和运行

### 方式一：使用 Docker Compose (推荐)

1. **克隆项目**
```bash
git clone https://github.com/yfh-yun/moviepilot-go.git
cd moviepilot-go
```

2. **配置环境变量**
```bash
cp configs/config.yaml.sample configs/config.yaml
# 编辑 configs/config.yaml 文件，配置数据库和Redis连接信息
```

3. **启动服务**
```bash
docker-compose up -d
```

4. **访问服务**
- API服务: http://localhost:3001
- 健康检查: http://localhost:3001/health
- API文档: http://localhost:3001/swagger/index.html

### 方式二：本地开发环境

1. **安装依赖**
```bash
go mod download
```

2. **配置数据库和Redis**
```bash
# 创建PostgreSQL数据库
createdb moviepilot

# 启动Redis (如果本地没有)
redis-server
```

3. **配置文件**
```bash
cp configs/config.yaml.sample configs/config.yaml
# 编辑配置文件
```

4. **运行应用**
```bash
# 使用简化版本 (推荐用于快速测试)
go run cmd/server/main_simple.go

# 或使用完整版本 (需要所有依赖)
go run cmd/server/main.go
```

## ⚙️ 配置说明

### 基础配置 (config.yaml)

```yaml
server:
  host: "0.0.0.0"
  port: 3001
  env: "development"  # development, production
  read_timeout: 30
  write_timeout: 30
  idle_timeout: 60
  shutdown_timeout: 30

database:
  type: "postgresql"
  host: "localhost"
  port: 5432
  name: "moviepilot"
  username: "postgres"
  password: "password"
  ssl_mode: "prefer"

redis:
  host: "localhost"
  port: 6379
  password: ""
  database: 0

logging:
  level: "info"  # debug, info, warn, error
  format: "json"  # json, text
```

### 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `SERVER_PORT` | 服务器端口 | `3001` |
| `SERVER_ENV` | 运行环境 | `development` |
| `DB_HOST` | 数据库主机 | `localhost` |
| `DB_PORT` | 数据库端口 | `5432` |
| `DB_NAME` | 数据库名称 | `moviepilot` |
| `REDIS_HOST` | Redis主机 | `localhost` |
| `REDIS_PORT` | Redis端口 | `6379` |

## 📚 API 文档

### 健康检查

```http
GET /health
```

响应示例：
```json
{
  "status": "ok",
  "timestamp": 1703123456,
  "version": "2.8.1",
  "uptime": "5m30s"
}
```

### API v1 端点

```http
GET /api/v1/ping
```

响应示例：
```json
{
  "message": "pong",
  "time": 1703123456
}
```

## 🏗️ 项目结构

```
moviepilot-go/
├── cmd/                    # 应用入口点
│   └── server/
│       ├── main.go         # 完整版本
│       └── main_simple.go  # 简化版本
├── internal/               # 私有应用代码
│   ├── api/               # API层
│   ├── config/            # 配置管理
│   ├── core/              # 核心业务逻辑
│   ├── integration/       # 第三方服务集成
│   ├── model/             # 数据模型
│   ├── repository/        # 数据访问层
│   ├── scheduler/         # 任务调度
│   └── service/           # 业务服务层
├── pkg/                   # 可复用的公共库
│   ├── cache/            # 缓存封装
│   ├── database/         # 数据库连接
│   ├── errors/           # 错误处理
│   ├── logger/           # 日志封装
│   └── utils/            # 工具函数
├── configs/              # 配置文件
├── deployments/           # 部署配置
├── docs/                 # 文档
└── tests/                # 测试文件
```

## 🔧 开发工具

### 构建项目

```bash
# 使用 Makefile
make build

# 手动构建
go build -o bin/moviepilot cmd/server/main.go
```

### 运行测试

```bash
# 运行所有测试
make test

# 运行单元测试
go test ./tests/unit/...

# 运行测试并生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 代码格式化

```bash
# 格式化代码
go fmt ./...

# 运行 linter
golangci-lint run
```

## 🐳 Docker 部署

### 开发环境

```bash
docker-compose -f deployments/docker-compose.dev.yml up -d
```

### 生产环境

```bash
docker-compose -f deployments/docker-compose.prod.yml up -d
```

## 📊 监控和日志

### 日志级别

- `debug`: 调试信息
- `info`: 一般信息
- `warn`: 警告信息
- `error`: 错误信息
- `fatal`: 致命错误

### 监控端点

- `/health` - 健康检查
- `/metrics` - Prometheus 指标 (计划中)

## 🤝 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

## 📝 开发计划

- [ ] 完整的用户认证系统
- [ ] 媒体库管理功能
- [ ] 自动下载功能
- [ ] 插件系统
- [ ] 监控和指标收集
- [ ] API 文档完善

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🆘 故障排除

### 常见问题

1. **数据库连接失败**
   - 检查PostgreSQL是否运行
   - 验证配置文件中的数据库连接信息

2. **Redis连接失败**
   - 检查Redis是否运行
   - 验证Redis连接配置

3. **端口被占用**
   - 修改配置文件中的端口号
   - 或停止占用端口的进程

4. **依赖安装失败**
   ```bash
   go clean -modcache
   go mod download
   ```

### 获取帮助

- 提交 Issue: https://github.com/yfh-yun/moviepilot-go/issues
- 查看文档: [docs/](./docs/)
- 联系维护者: support@moviepilot.com