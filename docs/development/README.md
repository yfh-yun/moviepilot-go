# 开发指南

## 🛠️ 开发环境搭建

### 环境要求
- **Go**: 1.21+
- **Git**: 2.30+
- **Docker**: 20.10+
- **Docker Compose**: 2.0+
- **IDE**: VS Code / GoLand / Vim

### 快速开始

```bash
# 1. 克隆项目
git clone https://github.com/yfh-yun/moviepilot-go.git
cd moviepilot-go

# 2. 安装 Go 依赖
go mod download

# 3. 安装开发工具
make install-tools

# 4. 启动开发环境
make dev

# 5. 运行测试
make test
```

## 📁 项目结构详解

### 目录说明

```
moviepilot-go/
├── cmd/                           # 应用入口点
│   └── server/
│       ├── main.go                # 主应用入口
│       └── main_simple.go         # 简化版入口（测试用）
├── internal/                      # 私有应用代码
│   ├── apis/                      # API 层
│   │   ├── handlers/              # HTTP 处理器
│   │   │   ├── v1/                # API v1 版本
│   │   │   └── middlewares/       # 中间件
│   │   └── routes/                # 路由定义
│   ├── business/                  # 业务逻辑层
│   │   ├── domains/               # 领域模型
│   │   ├── services/              # 业务服务
│   │   └── workflows/             # 工作流
│   ├── infrastructure/            # 基础设施层
│   │   ├── config/                # 配置管理
│   │   ├── security/              # 安全组件
│   │   └── events/                # 事件系统
│   ├── models/                    # 数据模型
│   ├── repositories/              # 数据访问层
│   └── monitor/                   # 监控系统
├── pkg/                           # 公共库
│   ├── database/                  # 数据库封装
│   ├── cache/                     # 缓存封装
│   ├── logger/                    # 日志封装
│   ├── plugin/                    # 插件系统
│   └── utils/                     # 工具函数
├── configs/                       # 配置文件
├── deployments/                   # 部署配置
├── scripts/                       # 脚本文件
├── tests/                         # 测试文件
└── docs/                          # 文档
```

### 分层架构原则

```
┌─────────────────┐
│   APIs Layer    │ ← HTTP 请求处理
├─────────────────┤
│ Business Layer  │ ← 业务逻辑
├─────────────────┤
│Infrastructure  │ ← 技术实现
├─────────────────┤
│Repositories     │ ← 数据访问
└─────────────────┘
```

**依赖方向**: APIs → Business → Infrastructure → Repositories

## 🧩 核心组件开发

### 1. API Handler 开发

#### 创建新 Handler
```go
// internal/apis/handlers/v1/user.go
package v1

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "moviepilot-go/internal/business/services"
    "moviepilot-go/pkg/logger"
)

type UserHandler struct {
    userService services.UserService
}

func NewUserHandler(userService services.UserService) *UserHandler {
    return &UserHandler{
        userService: userService,
    }
}

// GetUserProfile 获取用户信息
// @Summary 获取用户信息
// @Description 获取当前用户的详细信息
// @Tags users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} model.User
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/users/profile [get]
func (h *UserHandler) GetUserProfile(c *gin.Context) {
    userID := c.GetString("user_id")
    
    logger.Debug("GetUserProfile called", "user_id", userID)
    
    user, err := h.userService.GetUserByID(userID)
    if err != nil {
        logger.Error("Failed to get user profile", "error", err.Error(), "user_id", userID)
        c.JSON(http.StatusInternalServerError, gin.H{
            "code":    500,
            "message": "Internal server error",
            "error":   err.Error(),
        })
        return
    }
    
    logger.Info("User profile retrieved successfully", "user_id", userID)
    c.JSON(http.StatusOK, gin.H{
        "code":    200,
        "message": "success",
        "data":    user,
    })
}
```

#### 注册路由
```go
// internal/apis/routes/v1.go
package routes

import (
    "github.com/gin-gonic/gin"
    "moviepilot-go/internal/apis/handlers/v1"
)

func RegisterV1Routes(r *gin.Engine, handlers *v1.Handlers) {
    v1 := r.Group("/api/v1")
    {
        // 用户路由
        users := v1.Group("/users")
        {
            users.GET("/profile", handlers.User.GetUserProfile)
            users.PUT("/profile", handlers.User.UpdateUserProfile)
        }
        
        // 媒体路由
        media := v1.Group("/media")
        {
            media.GET("", handlers.Media.GetMediaList)
            media.POST("", handlers.Media.CreateMedia)
            media.GET("/:id", handlers.Media.GetMedia)
            media.PUT("/:id", handlers.Media.UpdateMedia)
            media.DELETE("/:id", handlers.Media.DeleteMedia)
        }
    }
}
```

### 2. Service 开发

#### 接口定义
```go
// internal/business/services/user.go
package services

import (
    "context"
    "moviepilot-go/internal/models"
)

type UserService interface {
    GetUserByID(ctx context.Context, id string) (*models.User, error)
    UpdateUser(ctx context.Context, user *models.User) error
    CreateUser(ctx context.Context, user *models.User) error
    DeleteUser(ctx context.Context, id string) error
}

type userService struct {
    userRepo repositories.UserRepository
    logger   logger.Logger
}

func NewUserService(userRepo repositories.UserRepository, logger logger.Logger) UserService {
    return &userService{
        userRepo: userRepo,
        logger:   logger,
    }
}

func (s *userService) GetUserByID(ctx context.Context, id string) (*models.User, error) {
    s.logger.Debug("Getting user by ID", "user_id", id)
    
    user, err := s.userRepo.GetByID(ctx, id)
    if err != nil {
        s.logger.Error("Failed to get user", "error", err.Error(), "user_id", id)
        return nil, err
    }
    
    s.logger.Info("User retrieved successfully", "user_id", id)
    return user, nil
}
```

### 3. Repository 开发

#### 接口定义
```go
// internal/repositories/user.go
package repositories

import (
    "context"
    "moviepilot-go/internal/models"
)

type UserRepository interface {
    GetByID(ctx context.Context, id string) (*models.User, error)
    GetByEmail(ctx context.Context, email string) (*models.User, error)
    Create(ctx context.Context, user *models.User) error
    Update(ctx context.Context, user *models.User) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, offset, limit int) ([]*models.User, error)
}
```

#### 实现
```go
// internal/repositories/user_gorm.go
package repositories

import (
    "context"
    "moviepilot-go/internal/models"
    "moviepilot-go/pkg/database"
    "moviepilot-go/pkg/logger"
)

type userRepository struct {
    db     *gorm.DB
    logger logger.Logger
}

func NewUserRepository(db *database.DB, logger logger.Logger) UserRepository {
    return &userRepository{
        db:     db.GetDB(),
        logger: logger,
    }
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
    var user models.User
    
    err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrUserNotFound
        }
        r.logger.Error("Failed to get user by ID", "error", err.Error(), "user_id", id)
        return nil, err
    }
    
    r.logger.Debug("User retrieved from database", "user_id", id)
    return &user, nil
}
```

## 🧪 测试开发

### 1. 单元测试

#### Service 测试
```go
// tests/services/user_test.go
package services_test

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "moviepilot-go/internal/business/services"
    "moviepilot-go/internal/models"
)

// Mock Repository
type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
    args := m.Called(ctx, id)
    return args.Get(0).(*models.User), args.Error(1)
}

func TestUserService_GetUserByID(t *testing.T) {
    // 准备
    mockRepo := new(MockUserRepository)
    mockLogger := logger.NewNopLogger()
    
    service := services.NewUserService(mockRepo, mockLogger)
    
    expectedUser := &models.User{
        ID:       "user-123",
        Username: "testuser",
        Email:    "test@example.com",
    }
    
    mockRepo.On("GetByID", mock.Anything, "user-123").Return(expectedUser, nil)
    
    // 执行
    user, err := service.GetUserByID(context.Background(), "user-123")
    
    // 验证
    assert.NoError(t, err)
    assert.Equal(t, expectedUser.Username, user.Username)
    assert.Equal(t, expectedUser.Email, user.Email)
    mockRepo.AssertExpectations(t)
}
```

### 2. 集成测试

#### API 测试
```go
// tests/api/user_test.go
package api_test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
)

func TestUserHandler_GetUserProfile(t *testing.T) {
    // 设置 Gin 测试模式
    gin.SetMode(gin.TestMode)
    
    // 创建测试路由
    router := gin.New()
    
    // 设置测试 Handler
    mockService := new(MockUserService)
    handler := v1.NewUserHandler(mockService)
    
    router.GET("/api/v1/users/profile", handler.GetUserProfile)
    
    // 准备测试数据
    expectedUser := &models.User{
        ID:       "user-123",
        Username: "testuser",
        Email:    "test@example.com",
    }
    
    mockService.On("GetUserByID", mock.Anything, "user-123").Return(expectedUser, nil)
    
    // 创建测试请求
    req, _ := http.NewRequest("GET", "/api/v1/users/profile", nil)
    req.Header.Set("Authorization", "Bearer test-token")
    
    // 模拟中间件设置 user_id
    router.Use(func(c *gin.Context) {
        c.Set("user_id", "user-123")
        c.Next()
    })
    
    // 执行请求
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    // 验证响应
    assert.Equal(t, http.StatusOK, w.Code)
    
    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.Equal(t, float64(200), response["code"])
    assert.Equal(t, "success", response["message"])
    
    data := response["data"].(map[string]interface{})
    assert.Equal(t, "testuser", data["username"])
    assert.Equal(t, "test@example.com", data["email"])
}
```

### 3. 测试数据库

#### 测试配置
```go
// tests/database/test_db.go
package database

import (
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "moviepilot-go/internal/models"
)

func SetupTestDB(t *testing.T) *gorm.DB {
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    if err != nil {
        t.Fatalf("Failed to connect to test database: %v", err)
    }
    
    // 自动迁移测试表
    err = db.AutoMigrate(&models.User{}, &models.Media{}, &models.Transfer{})
    if err != nil {
        t.Fatalf("Failed to migrate test database: %v", err)
    }
    
    return db
}

func CleanupTestDB(t *testing.T, db *gorm.DB) {
    sqlDB, err := db.DB()
    if err != nil {
        t.Fatalf("Failed to get underlying sql.DB: %v", err)
    }
    sqlDB.Close()
}
```

## 🔧 开发工具

### 1. Makefile 命令

```makefile
# Makefile
.PHONY: help build test lint clean dev

help: ## 显示帮助信息
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## 构建应用
	go build -o bin/moviepilot cmd/server/main.go

test: ## 运行测试
	go test -v ./...

test-coverage: ## 运行测试并生成覆盖率报告
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint: ## 代码检查
	golangci-lint run

fmt: ## 格式化代码
	go fmt ./...
	goimports -w .

dev: ## 启动开发环境
	docker-compose -f deployments/docker-compose.dev.yml up -d
	go run cmd/server/main.go

clean: ## 清理构建文件
	rm -rf bin/
	rm -f coverage.out coverage.html

install-tools: ## 安装开发工具
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
```

### 2. VS Code 配置

#### .vscode/settings.json
```json
{
    "go.toolsManagement.checkForUpdates": "local",
    "go.useLanguageServer": true,
    "go.gopath": "",
    "go.goroot": "",
    "go.formatTool": "goimports",
    "go.lintTool": "golangci-lint",
    "go.lintOnSave": "package",
    "go.testOnSave": true,
    "go.coverOnSave": true,
    "go.coverageDecorator": {
        "type": "gutter",
        "coveredHighlightColor": "rgba(64,128,64,0.5)",
        "uncoveredHighlightColor": "rgba(128,64,64,0.25)"
    },
    "files.exclude": {
        "**/bin": true,
        "**/coverage.out": true,
        "**/coverage.html": true
    }
}
```

#### .vscode/tasks.json
```json
{
    "version": "2.0.0",
    "tasks": [
        {
            "label": "go: build",
            "type": "shell",
            "command": "make",
            "args": ["build"],
            "group": {
                "kind": "build",
                "isDefault": true
            }
        },
        {
            "label": "go: test",
            "type": "shell",
            "command": "make",
            "args": ["test"],
            "group": {
                "kind": "test",
                "isDefault": true
            }
        },
        {
            "label": "go: lint",
            "type": "shell",
            "command": "make",
            "args": ["lint"]
        }
    ]
}
```

### 3. Git Hooks

#### .pre-commit-config.yaml
```yaml
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.4.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-yaml
      - id: check-added-large-files

  - repo: https://github.com/golangci/golangci-lint
    rev: v1.54.2
    hooks:
      - id: golangci-lint

  - repo: local
    hooks:
      - id: go-test
        name: go test
        entry: go test ./...
        language: system
        pass_filenames: false
```

## 📝 代码规范

### 1. 命名规范

#### 文件命名
```
✅ user_handler.go
✅ transfer_service.go
✅ media_repository.go

❌ userHandler.go
❌ transfer-service.go
❌ MediaRepository.go
```

#### 变量命名
```go
✅ userID, userName, isActive
✅ userService, transferHistory
✅ MAX_RETRY_COUNT, DEFAULT_TIMEOUT

❌ userId, UserName, is_active
❌ UserService, transfer_history
❌ max_retry_count, defaultTimeout
```

#### 函数命名
```go
✅ GetUserByID(), CreateMedia(), UpdateTransferStatus()
✅ isValidEmail(), parseConfig(), handleError()

❌ getUserById(), createmedia(), update_transfer_status()
❌ IsValidEmail(), ParseConfig(), HandleError()
```

### 2. 注释规范

#### 包注释
```go
// Package services provides business logic implementations for the MoviePilot application.
// It contains service interfaces and implementations that coordinate between
// different domain objects and external dependencies.
package services
```

#### 函数注释
```go
// GetUserByID retrieves a user by their unique identifier.
// It returns the user object if found, otherwise returns an error.
// The ctx parameter is used for request cancellation and timeouts.
func (s *userService) GetUserByID(ctx context.Context, id string) (*models.User, error) {
    // Implementation...
}
```

#### 结构体注释
```go
// User represents a user in the system with authentication and profile information.
type User struct {
    ID       string    `json:"id" gorm:"primaryKey;type:varchar(36)"` // Unique identifier
    Username string    `json:"username" gorm:"unique;not null"`       // Unique username
    Email    string    `json:"email" gorm:"unique;not null"`           // Email address
    Password string    `json:"-" gorm:"not null"`                      // Hashed password
    CreatedAt time.Time `json:"created_at"`                           // Creation timestamp
    UpdatedAt time.Time `json:"updated_at"`                           // Last update timestamp
}
```

### 3. 错误处理规范

#### 错误定义
```go
// pkg/errors/errors.go
package errors

import (
    "errors"
    "fmt"
)

var (
    ErrUserNotFound     = errors.New("user not found")
    ErrInvalidCredentials = errors.New("invalid credentials")
    ErrDuplicateUser    = errors.New("user already exists")
)

// NewUserError creates a new user-related error with context
func NewUserError(message string, userID string) error {
    return fmt.Errorf("user error: %s (user_id: %s)", message, userID)
}
```

#### 错误处理
```go
func (s *userService) GetUserByID(ctx context.Context, id string) (*models.User, error) {
    user, err := s.userRepo.GetByID(ctx, id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            s.logger.Warn("User not found", "user_id", id)
            return nil, ErrUserNotFound
        }
        
        s.logger.Error("Failed to get user", "error", err.Error(), "user_id", id)
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    
    return user, nil
}
```

## 🚀 性能优化

### 1. 数据库优化

#### 查询优化
```go
// 使用预加载避免 N+1 查询
func (r *mediaRepository) GetMediaWithDetails(ctx context.Context, id string) (*models.Media, error) {
    var media models.Media
    err := r.db.WithContext(ctx).
        Preload("Transfers").
        Preload("Subtitles").
        Where("id = ?", id).
        First(&media).Error
    return &media, err
}

// 使用索引优化查询
func (r *mediaRepository) SearchMedia(ctx context.Context, query string, limit int) ([]*models.Media, error) {
    var media []*models.Media
    err := r.db.WithContext(ctx).
        Where("title ILIKE ? OR overview ILIKE ?", "%"+query+"%", "%"+query+"%").
        Order("created_at DESC").
        Limit(limit).
        Find(&media).Error
    return media, err
}
```

#### 批量操作
```go
func (r *userRepository) CreateUsers(ctx context.Context, users []*models.User) error {
    if len(users) == 0 {
        return nil
    }
    
    // 使用批量插入
    err := r.db.WithContext(ctx).CreateInBatches(users, 100).Error
    if err != nil {
        return fmt.Errorf("failed to create users in batch: %w", err)
    }
    
    return nil
}
```

### 2. 缓存优化

#### Redis 缓存
```go
func (s *userService) GetUserByID(ctx context.Context, id string) (*models.User, error) {
    // 尝试从缓存获取
    cacheKey := fmt.Sprintf("user:%s", id)
    
    if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
        var user models.User
        if json.Unmarshal([]byte(cached), &user) == nil {
            s.logger.Debug("User retrieved from cache", "user_id", id)
            return &user, nil
        }
    }
    
    // 从数据库获取
    user, err := s.userRepo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // 写入缓存
    if data, err := json.Marshal(user); err == nil {
        s.cache.Set(ctx, cacheKey, string(data), time.Hour)
    }
    
    return user, nil
}
```

---

**注意**: 遵循这些开发规范可以确保代码质量和团队协作效率。在提交代码前，请确保所有测试通过并且代码符合规范要求。