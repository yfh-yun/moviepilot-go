# MoviePilot Go 架构文档

## 项目概述

MoviePilot Go 是一个基于Go语言开发的自动化媒体库管理工具，采用分层架构和微服务设计。

## 架构原则

### 1. 分层架构

```
┌─────────────────────────────────────────────────────────┐
│                    API层 (HTTP)                        │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐       │
│  │  Handlers   │ │ Middleware  │ │   Routes    │       │
│  └─────────────┘ └─────────────┘ └─────────────┘       │
├─────────────────────────────────────────────────────────┤
│                  业务逻辑层 (Service)                   │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐       │
│  │   Core      │ │ Integration │ │   Chains    │       │
│  └─────────────┘ └─────────────┘ └─────────────┘       │
├─────────────────────────────────────────────────────────┤
│                  数据访问层 (Repository)                │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐       │
│  │ PostgreSQL  │ │    Redis    │ │   Models    │       │
│  └─────────────┘ └─────────────┘ └─────────────┘       │
├─────────────────────────────────────────────────────────┤
│                    基础设施层 (Infrastructure)           │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐       │
│  │   Config    │ │   Logger    │ │    Cache    │       │
│  └─────────────┘ └─────────────┘ └─────────────┘       │
└─────────────────────────────────────────────────────────┘
```

### 2. 依赖方向

- 高层模块不依赖低层模块
- 抽象不依赖具体实现
- 依赖关系单向流动

### 3. 领域驱动设计

- 按业务领域组织代码
- 明确的边界上下文
- 统一的语言和模型

## 目录结构

```
moviepilot-go/
├── cmd/                    # 应用入口
│   └── server/
│       └── main.go        # 主程序入口
├── internal/              # 私有代码
│   ├── api/              # API层
│   │   ├── handlers/     # HTTP处理器
│   │   ├── middleware/   # 中间件
│   │   ├── routes/       # 路由定义
│   │   ├── schemas/      # 请求/响应模式
│   │   ├── validator/    # 数据验证
│   │   └── response/     # 响应封装
│   ├── service/          # 业务逻辑层
│   │   ├── auth/         # 认证服务
│   │   ├── user/         # 用户服务
│   │   ├── media/        # 媒体服务
│   │   ├── download/     # 下载服务
│   │   ├── subscribe/    # 订阅服务
│   │   ├── site/         # 站点服务
│   │   ├── notification/ # 通知服务
│   │   ├── plugin/       # 插件服务
│   │   ├── chain/        # 业务链
│   │   └── template/     # 模板服务
│   ├── repository/       # 数据访问层
│   │   ├── postgres/     # PostgreSQL实现
│   │   └── redis/        # Redis实现
│   ├── model/            # 数据模型
│   │   └── models.go     # 模型定义
│   ├── core/             # 核心功能
│   │   ├── context/      # 上下文管理
│   │   ├── event/        # 事件系统
│   │   ├── meta/         # 元数据处理
│   │   ├── metainfo/     # 元信息解析
│   │   └── security/     # 安全功能
│   ├── integration/      # 第三方集成
│   │   ├── bangumi/      # Bangumi集成
│   │   ├── douban/       # 豆瓣集成
│   │   ├── emby/         # Emby集成
│   │   ├── jellyfin/     # Jellyfin集成
│   │   ├── plex/         # Plex集成
│   │   ├── tmdb/         # TMDB集成
│   │   └── tvdb/         # TVDB集成
│   ├── cache/            # 缓存管理
│   ├── config/           # 配置管理
│   ├── monitor/          # 监控功能
│   └── scheduler/        # 任务调度
├── pkg/                  # 公共库
│   ├── database/         # 数据库连接
│   ├── logger/           # 日志封装
│   ├── utils/            # 工具函数
│   ├── plugin/           # 插件系统
│   ├── response/         # 响应封装
│   ├── jwt/              # JWT处理
│   ├── httpclient/       # HTTP客户端
│   └── models/           # 公共模型
├── configs/              # 配置文件
├── deployments/          # 部署配置
├── docs/                 # 文档
├── scripts/              # 脚本文件
├── shared/               # 共享资源
├── tests/                # 测试文件
├── go.mod                # 模块定义
├── go.sum                # 依赖校验
├── Dockerfile            # 容器化
├── Makefile              # 构建工具
└── README.md             # 项目说明
```

## 核心组件

### 1. API层

负责HTTP请求处理和响应：
- **Handlers**: 处理具体的HTTP请求
- **Middleware**: 请求预处理和后处理
- **Routes**: 路由定义和注册
- **Schemas**: 请求和响应的数据结构
- **Validator**: 数据验证逻辑

### 2. Service层

实现核心业务逻辑：
- **业务服务**: 按领域划分的服务
- **业务链**: 复杂业务流程的编排
- **集成服务**: 第三方服务的集成

### 3. Repository层

数据访问抽象：
- **PostgreSQL**: 关系型数据存储
- **Redis**: 缓存和会话存储
- **Models**: 数据模型定义

### 4. 基础设施层

提供基础服务：
- **配置管理**: 动态配置加载
- **日志系统**: 结构化日志
- **缓存系统**: 多级缓存
- **插件系统**: 可扩展的插件架构

## 数据流

### 1. 请求处理流程

```
HTTP请求 → Middleware → Routes → Handlers → Services → Repository → Database
    ↑                                                    ↓
Response ← JSON封装 ← Error处理 ← Business Logic ← Data Access
```

### 2. 业务流程

```
事件触发 → 业务链 → 服务编排 → 数据操作 → 事件发布
    ↑                                      ↓
通知系统 ← 状态更新 ← 结果处理 ← 业务执行
```

## 设计模式

### 1. Repository模式

抽象数据访问，提供统一的数据操作接口：

```go
type UserRepository interface {
    Create(user *User) error
    GetByID(id int) (*User, error)
    Update(user *User) error
    Delete(id int) error
    List(offset, limit int) ([]*User, int64, error)
}
```

### 2. Service模式

封装业务逻辑，提供高级业务操作：

```go
type UserService interface {
    Register(req *RegisterRequest) (*User, error)
    Login(req *LoginRequest) (*LoginResponse, error)
    UpdateProfile(id int, req *UpdateProfileRequest) error
}
```

### 3. Chain模式

编排复杂的业务流程：

```go
type Chain interface {
    Execute(ctx context.Context, input interface{}) (interface{}, error)
    AddHandler(handler Handler) Chain
    SetErrorHandler(handler ErrorHandler) Chain
}
```

### 4. Plugin模式

支持功能的动态扩展：

```go
type Plugin interface {
    ID() string
    Name() string
    Version() string
    Start() error
    Stop() error
}
```

## 安全架构

### 1. 认证机制

- JWT Token认证
- API Key认证
- 会话管理

### 2. 授权机制

- 基于角色的访问控制(RBAC)
- 资源级权限控制
- API级权限验证

### 3. 数据安全

- 敏感数据加密
- SQL注入防护
- XSS攻击防护

## 性能优化

### 1. 数据库优化

- 连接池管理
- 查询优化
- 索引设计

### 2. 缓存策略

- 多级缓存
- 缓存预热
- 缓存穿透保护

### 3. 并发处理

- Goroutine池
- Channel通信
- 锁优化

## 监控体系

### 1. 应用监控

- 性能指标收集
- 错误率监控
- 业务指标统计

### 2. 基础设施监控

- 服务器资源监控
- 数据库性能监控
- 网络监控

### 3. 日志管理

- 结构化日志
- 日志聚合
- 日志分析

## 部署架构

### 1. 容器化部署

- Docker容器化
- Docker Compose编排
- Kubernetes部署

### 2. 微服务架构

- Go主应用
- Python插件服务
- 数据库服务
- 缓存服务

### 3. 负载均衡

- Nginx反向代理
- 应用层负载均衡
- 数据库读写分离

## 扩展性设计

### 1. 水平扩展

- 无状态设计
- 分布式缓存
- 微服务拆分

### 2. 垂直扩展

- 资源监控
- 性能调优
- 容量规划

### 3. 功能扩展

- 插件系统
- 业务链扩展
- API版本控制

## 技术选型

### 1. 核心框架

- **Web框架**: Gin
- **ORM**: GORM
- **配置管理**: Viper
- **日志**: Zap

### 2. 数据存储

- **关系数据库**: PostgreSQL
- **缓存**: Redis
- **搜索引擎**: Elasticsearch(可选)

### 3. 中间件

- **消息队列**: Redis Streams
- **监控**: Prometheus + Grafana
- **链路追踪**: Jaeger(可选)

### 4. 开发工具

- **API文档**: Swagger
- **测试**: Testify
- **代码质量**: golangci-lint
- **容器化**: Docker

## 最佳实践

### 1. 代码规范

- 遵循Go官方代码规范
- 使用统一的命名约定
- 编写清晰的注释

### 2. 错误处理

- 使用pkg/errors包装错误
- 提供详细的错误信息
- 记录错误上下文

### 3. 测试策略

- 单元测试覆盖率>80%
- 集成测试覆盖核心流程
- 端到端测试验证业务场景

### 4. 文档维护

- API文档自动生成
- 架构文档及时更新
- 部署文档详细准确

## 未来规划

### 1. 短期目标

- 完善核心功能
- 优化性能表现
- 增强监控能力

### 2. 中期目标

- 微服务拆分
- 云原生改造
- 智能化功能

### 3. 长期目标

- 多租户支持
- 国际化支持
- 生态系统建设