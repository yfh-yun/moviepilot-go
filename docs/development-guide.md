# 开发指南

> **MoviePilot Go 开发环境搭建与开发流程**  
> **更新时间**: 2025-12-02

---

## 📋 目录

- [环境要求](#环境要求)
- [快速开始](#快速开始)
- [项目结构](#项目结构)
- [开发流程](#开发流程)
- [调试技巧](#调试技巧)
- [常见问题](#常见问题)

---

## 环境要求

### 必需软件

| 软件 | 版本 | 说明 |
|------|------|------|
| Go | 1.21+ | 编程语言 |
| PostgreSQL | 15+ | 数据库 |
| Redis | 7+ | 缓存 |
| Docker | 20.10+ | 容器化（可选） |
| Docker Compose | 2.0+ | 容器编排（可选） |
| Git | 2.30+ | 版本控制 |

### 开发工具

```bash
# 安装 golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 安装 swag (API 文档生成)
go install github.com/swaggo/swag/cmd/swag@latest

# 安装 goimports (代码格式化)
go install golang.org/x/tools/cmd/goimports@latest

# 安装 migrate (数据库迁移)
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# 或使用 Makefile 一键安装
make install-tools
```

---

## 快速开始

### 1. 克隆项目

```bash
git clone https://github.com/your-org/moviepilot-go.git
cd moviepilot-go
```

### 2. 配置环境

**方式一：使用 Docker（推荐）**

```bash
# 启动开发环境（包含数据库和 Redis）
make dev

# 或手动启动
docker-compose -f deployments/docker-compose.dev.yml up -d
```

**方式二：本地安装**

```bash
# 安装 PostgreSQL
sudo apt-get install postgresql-15

# 安装 Redis
sudo apt-get install redis-server

# 创建数据库
createdb moviepilot

# 启动 Redis
redis-server
```

### 3. 配置文件

```bash
# 复制配置文件
cp configs/config.yaml.example configs/config.yaml

# 编辑配置
vim configs/config.yaml
```

**关键配置项**:

```yaml
server:
  port: 3001
  mode: debug  # debug/release

database:
  host: localhost
  port: 5432
  name: moviepilot
  user: moviepilot
  password: moviepilot123

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

log:
  level: debug  # debug/info/warn/error
  file: logs/app.log
```

### 4. 数据库迁移

```bash
# 执行迁移
make migrate-up

# 插入种子数据
make migrate-seed

# 查看迁移状态
make migrate-status
```

### 5. 启动应用

```bash
# 开发模式运行
make run

# 或直接运行
go run cmd/server/main.go
```

访问 http://localhost:3001

---

## 项目结构

```
moviepilot-go/
├── cmd/                        # 应用入口
│   └── server/
│       └── main.go            # 主程序
├── internal/                   # 私有代码
│   ├── apis/                  # API 层
│   │   ├── handlers/          # HTTP 处理器
│   │   ├── middlewares/       # 中间件
│   │   └── routes/            # 路由定义
│   ├── business/              # 业务层
│   │   ├── domains/           # 领域模型
│   │   ├── services/          # 业务服务
│   │   └── workflows/         # 工作流
│   ├── infrastructure/        # 基础设施层
│   │   ├── config/            # 配置管理
│   │   ├── security/          # 安全组件
│   │   └── events/            # 事件系统
│   ├── integration/           # 集成层
│   │   ├── downloader/        # 下载器
│   │   ├── mediaserver/       # 媒体服务器
│   │   ├── metadata/          # 元数据
│   │   ├── notification/      # 通知
│   │   └── indexer/           # 索引器
│   ├── models/                # 数据模型
│   │   └── database/          # 数据库模型
│   └── repositories/          # 数据访问
│       └── repositories/      # Repository 实现
├── pkg/                        # 公共库
│   ├── cache/                 # 缓存封装
│   ├── database/              # 数据库连接
│   ├── logger/                # 日志封装
│   ├── plugin/                # 插件系统
│   └── utils/                 # 工具函数
├── configs/                    # 配置文件
├── database/                   # 数据库相关
│   ├── migrations/            # 迁移脚本
│   └── seeds/                 # 种子数据
├── deployments/                # 部署配置
│   ├── docker-compose.yml
│   ├── prometheus/
│   └── grafana/
├── docs/                       # 文档
├── scripts/                    # 脚本
├── tests/                      # 测试
├── go.mod                      # Go 模块
├── go.sum                      # 依赖锁定
├── Dockerfile                  # Docker 镜像
├── Makefile                    # 构建脚本
└── README.md                   # 项目说明
```

---

## 开发流程

### 1. 创建新功能

#### 步骤 1: 创建分支

```bash
git checkout -b feature/amazing-feature
```

#### 步骤 2: 定义数据模型

```go
// internal/models/database/models.go
type AmazingFeature struct {
    ID        uint      `gorm:"primaryKey"`
    Name      string    `gorm:"size:100;not null"`
    Status    string    `gorm:"size:20;default:'active'"`
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

#### 步骤 3: 创建迁移脚本

```bash
# 创建迁移文件
migrate create -ext sql -dir database/migrations -seq create_amazing_features_table
```

```sql
-- database/migrations/xxx_create_amazing_features_table.up.sql
CREATE TABLE amazing_features (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### 步骤 4: 创建 Repository

```go
// internal/repositories/repositories/amazing_feature_repository.go
package repositories

type AmazingFeatureRepository interface {
    Create(ctx context.Context, feature *models.AmazingFeature) error
    FindByID(ctx context.Context, id uint) (*models.AmazingFeature, error)
    List(ctx context.Context) ([]*models.AmazingFeature, error)
}

type amazingFeatureRepository struct {
    db *gorm.DB
}

func NewAmazingFeatureRepository(db *gorm.DB) AmazingFeatureRepository {
    return &amazingFeatureRepository{db: db}
}
```

#### 步骤 5: 创建 Service

```go
// internal/business/services/amazing/service.go
package amazing

type Service interface {
    CreateFeature(ctx context.Context, req CreateRequest) (*Response, error)
    GetFeature(ctx context.Context, id uint) (*Response, error)
}

type service struct {
    repo repositories.AmazingFeatureRepository
    logger *zap.Logger
}

func NewService(repo repositories.AmazingFeatureRepository) Service {
    return &service{
        repo: repo,
        logger: logger.GetLogger(),
    }
}
```

#### 步骤 6: 创建 Handler

```go
// internal/apis/handlers/amazing/handler.go
package amazing

type Handler struct {
    service amazing.Service
}

func NewHandler(service amazing.Service) *Handler {
    return &Handler{service: service}
}

// @Summary Create amazing feature
// @Tags amazing
// @Accept json
// @Produce json
// @Param feature body CreateRequest true "Feature data"
// @Success 201 {object} Response
// @Router /api/v1/amazing [post]
func (h *Handler) Create(c *gin.Context) {
    // 实现...
}
```

#### 步骤 7: 注册路由

```go
// internal/apis/routes/routes.go
func RegisterRoutes(r *gin.Engine) {
    v1 := r.Group("/api/v1")
    {
        amazing := v1.Group("/amazing")
        {
            handler := handlers.NewAmazingHandler(/* ... */)
            amazing.POST("", handler.Create)
            amazing.GET("/:id", handler.Get)
        }
    }
}
```

#### 步骤 8: 编写测试

```go
// internal/business/services/amazing/service_test.go
func TestService_CreateFeature(t *testing.T) {
    // 测试实现...
}
```

### 2. 运行测试

```bash
# 运行所有测试
make test

# 运行特定包的测试
go test -v ./internal/business/services/amazing/

# 生成覆盖率报告
make test-cover
```

### 3. 代码检查

```bash
# 格式化代码
make fmt

# 运行 lint
make lint

# 静态分析
make vet
```

### 4. 生成文档

```bash
# 生成 Swagger 文档
make swagger

# 访问 http://localhost:3001/swagger/index.html
```

### 5. 提交代码

```bash
git add .
git commit -m "feat(amazing): add amazing feature"
git push origin feature/amazing-feature
```

---

## 调试技巧

### 1. 使用 Delve 调试器

```bash
# 安装 Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# 启动调试
dlv debug cmd/server/main.go

# 设置断点
(dlv) break main.main
(dlv) continue
```

### 2. 日志调试

```go
import "moviepilot-go/pkg/logger"

logger.Debug("Debug message", "key", "value")
logger.Info("Info message", "user_id", userID)
logger.Error("Error occurred", "error", err)
```

### 3. 性能分析

```bash
# CPU 分析
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof

# 内存分析
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof

# HTTP 性能分析
curl http://localhost:3001/debug/pprof/heap > heap.prof
go tool pprof heap.prof
```

### 4. 数据库调试

```go
// 启用 SQL 日志
db.Debug().Where("id = ?", id).First(&user)

// 查看生成的 SQL
sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
    return tx.Where("id = ?", id).First(&user)
})
fmt.Println(sql)
```

---

## 常见问题

### Q1: 数据库连接失败

**问题**: `dial tcp 127.0.0.1:5432: connect: connection refused`

**解决**:
```bash
# 检查 PostgreSQL 是否运行
sudo systemctl status postgresql

# 启动 PostgreSQL
sudo systemctl start postgresql

# 或使用 Docker
docker-compose -f deployments/docker-compose.dev.yml up -d postgres
```

### Q2: Redis 连接失败

**问题**: `dial tcp 127.0.0.1:6379: connect: connection refused`

**解决**:
```bash
# 检查 Redis 是否运行
redis-cli ping

# 启动 Redis
redis-server

# 或使用 Docker
docker-compose -f deployments/docker-compose.dev.yml up -d redis
```

### Q3: 端口被占用

**问题**: `bind: address already in use`

**解决**:
```bash
# 查找占用端口的进程
lsof -i :3001

# 杀死进程
kill -9 <PID>

# 或修改配置文件中的端口
vim configs/config.yaml
```

### Q4: 依赖下载失败

**问题**: `go: downloading ... timeout`

**解决**:
```bash
# 设置 Go 代理
go env -w GOPROXY=https://goproxy.cn,direct

# 重新下载
go mod download
```

### Q5: 测试失败

**问题**: 测试数据库连接失败

**解决**:
```bash
# 创建测试数据库
createdb moviepilot_test

# 设置测试环境变量
export DB_NAME=moviepilot_test
export DB_HOST=localhost

# 运行测试
make test
```

---

## 开发工具推荐

### IDE

- **GoLand** - JetBrains 官方 Go IDE
- **VS Code** - 配合 Go 插件
- **Vim/Neovim** - 配合 vim-go

### VS Code 插件

- Go (官方)
- Go Test Explorer
- GitLens
- Docker
- REST Client

### 浏览器插件

- Swagger UI
- JSON Formatter
- Postman

---

## 性能优化建议

1. **数据库查询优化**
   - 使用索引
   - 避免 N+1 查询
   - 使用批量操作

2. **缓存使用**
   - 热点数据缓存
   - 查询结果缓存
   - 设置合理的 TTL

3. **并发处理**
   - 使用 Goroutine 池
   - 控制并发数量
   - 避免 Goroutine 泄漏

4. **内存管理**
   - 及时释放资源
   - 使用对象池
   - 避免内存泄漏

---

## 参考资源

- [Go 官方文档](https://golang.org/doc/)
- [Gin 文档](https://gin-gonic.com/docs/)
- [GORM 文档](https://gorm.io/docs/)
- [项目 Wiki](https://github.com/your-org/moviepilot-go/wiki)

---

**祝开发愉快！** 🚀
