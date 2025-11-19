# 开发指南

## 概述

本文档介绍如何参与MoviePilot Go项目的开发。

## 开发环境设置

### 前置要求

- Go 1.21+
- Docker & Docker Compose
- Git
- IDE (推荐VSCode或GoLand)

### 环境配置

1. 克隆项目
```bash
git clone https://github.com/moviepilot/moviepilot-go.git
cd moviepilot-go
```

2. 安装依赖
```bash
go mod download
```

3. 设置开发环境
```bash
make setup-dev
```

4. 启动开发服务
```bash
make dev
```

### 开发工具

推荐安装以下VSCode扩展：
- Go (官方)
- Docker
- GitLens
- Thunder Client (API测试)

## 代码规范

### 命名规范

#### 包名
- 小写，简短，有意义
- 单个单词优先
- 避免下划线和驼峰

```go
// 好的包名
package user
package cache
package utils

// 不好的包名
package user_service
package cacheUtil
package MyPackage
```

#### 文件名
- 小写字母和下划线
- 描述文件内容
- 避免缩写

```go
// 好的文件名
user_service.go
cache_repository.go
http_handler.go

// 不好的文件名
us.go
cr.go
handler.go
```

#### 变量名
- 驼峰命名法
- 有意义的名称
- 避免单字母变量（循环变量除外）

```go
// 好的变量名
var userService UserService
var cacheManager CacheManager
var isActive bool

// 不好的变量名
var us UserService
var cm CacheManager
var flag bool
```

#### 函数名
- 导出函数：大驼峰
- 私有函数：小驼峰
- 动词开头

```go
// 好的函数名
func GetUserByID(id int) (*User, error)
func validateEmail(email string) bool
func (s *UserService) CreateUser(user *User) error

// 不好的函数名
func user(id int) (*User, error)
func check(email string) bool
func (s *UserService) add(user *User) error
```

#### 接口名
- 大驼峰
- 通常以-er结尾
- 描述行为

```go
// 好的接口名
type UserRepository interface
type CacheManager interface
type HTTPHandler interface

// 不好的接口名
type UserRepo interface
type Cache interface
type Handler interface
```

### 注释规范

#### 包注释
```go
// Package user provides user management functionality.
// It includes user authentication, authorization, and profile management.
package user
```

#### 公共函数注释
```go
// GetUserByID retrieves a user by their ID.
// It returns the user if found, otherwise returns an error.
//
// Parameters:
//   - id: The user's unique identifier
//
// Returns:
//   - *User: The user object if found
//   - error: An error if the user is not found or database error occurs
func GetUserByID(id int) (*User, error) {
    // implementation
}
```

#### 结构体注释
```go
// User represents a user in the system.
// It contains authentication information and profile data.
type User struct {
    ID       int    `json:"id" gorm:"primaryKey"`
    Username string `json:"username" gorm:"unique;not null"`
    Email    string `json:"email" gorm:"unique;not null"`
    Password string `json:"-" gorm:"not null"` // 密码不序列化
}
```

### 错误处理

#### 错误定义
```go
// 使用pkg/errors包
import "moviepilot-go/pkg/errors"

// 定义业务错误
var (
    ErrUserNotFound = errors.New("user not found")
    ErrInvalidEmail = errors.New("invalid email format")
    ErrDuplicateUser = errors.New("user already exists")
)
```

#### 错误处理
```go
// 好的错误处理
func GetUserByID(id int) (*User, error) {
    user, err := repository.FindByID(id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, errors.Wrap(ErrUserNotFound, "failed to get user")
        }
        return nil, errors.Wrap(err, "database error")
    }
    return user, nil
}

// 不好的错误处理
func GetUserByID(id int) (*User, error) {
    user, err := repository.FindByID(id)
    if err != nil {
        return nil, err // 直接返回原始错误
    }
    return user, nil
}
```

### 日志规范

#### 日志级别
- Debug: 调试信息（开发环境）
- Info: 一般信息（生产环境可见）
- Warn: 警告信息
- Error: 错误信息
- Fatal: 致命错误（程序退出）

#### 日志使用
```go
import "moviepilot-go/pkg/logger"

func GetUserByID(id int) (*User, error) {
    logger.Debug("GetUserByID called", "user_id", id)
    
    user, err := repository.FindByID(id)
    if err != nil {
        logger.Error("Failed to get user", 
            "user_id", id, 
            "error", err.Error())
        return nil, err
    }
    
    logger.Info("User retrieved successfully", 
        "user_id", id, 
        "username", user.Username)
    return user, nil
}
```

## 项目结构

### 目录说明

```
moviepilot-go/
├── cmd/                    # 应用入口
│   └── server/
│       └── main.go        # 主程序入口
├── internal/              # 私有代码
│   ├── api/              # API层
│   ├── service/          # 业务逻辑层
│   ├── repository/       # 数据访问层
│   └── model/           # 数据模型
├── pkg/                 # 公共库
│   ├── logger/          # 日志封装
│   ├── database/        # 数据库连接
│   ├── cache/           # 缓存封装
│   └── utils/           # 工具函数
├── configs/             # 配置文件
├── tests/               # 测试文件
├── docs/                # 文档
└── deployments/         # 部署配置
```

### 分层架构

```
┌─────────────────┐
│     API层       │ HTTP请求处理
├─────────────────┤
│   Service层     │ 业务逻辑
├─────────────────┤
│ Repository层    │ 数据访问
├─────────────────┤
│   Database层    │ 数据库
└─────────────────┘
```

## 开发流程

### 分支管理

1. **main分支**: 生产环境代码
2. **develop分支**: 开发环境代码
3. **feature分支**: 新功能开发
4. **hotfix分支**: 紧急修复

### 提交规范

使用Conventional Commits规范：

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

#### 类型说明
- feat: 新功能
- fix: 修复bug
- docs: 文档更新
- style: 代码格式调整
- refactor: 代码重构
- test: 测试相关
- chore: 构建过程或辅助工具的变动

#### 示例
```bash
git commit -m "feat(auth): add JWT authentication"
git commit -m "fix(user): resolve user creation validation error"
git commit -m "docs(api): update authentication documentation"
```

### 开发步骤

1. 创建功能分支
```bash
git checkout -b feature/new-feature
```

2. 开发功能
```bash
# 编写代码
# 运行测试
make test
# 代码检查
make lint
```

3. 提交代码
```bash
git add .
git commit -m "feat: add new feature"
```

4. 推送分支
```bash
git push origin feature/new-feature
```

5. 创建Pull Request

6. 代码审查

7. 合并到develop分支

8. 发布到生产环境

## 测试

### 测试类型

1. **单元测试**: 测试单个函数或方法
2. **集成测试**: 测试多个组件的协作
3. **端到端测试**: 测试完整的用户场景

### 测试编写

#### 单元测试
```go
func TestGetUserByID(t *testing.T) {
    // 准备
    repo := &MockUserRepository{}
    service := NewUserService(repo)
    
    user := &User{ID: 1, Username: "test"}
    repo.On("FindByID", 1).Return(user, nil)
    
    // 执行
    result, err := service.GetUserByID(1)
    
    // 断言
    assert.NoError(t, err)
    assert.Equal(t, user, result)
    repo.AssertExpectations(t)
}
```

#### 集成测试
```go
func TestUserAPI(t *testing.T) {
    // 设置测试数据库
    db := setupTestDB()
    defer cleanupTestDB(db)
    
    // 设置测试路由
    router := setupTestRouter(db)
    
    // 测试请求
    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/api/v1/users/1", nil)
    router.ServeHTTP(w, req)
    
    // 断言
    assert.Equal(t, 200, w.Code)
    
    var user User
    err := json.Unmarshal(w.Body.Bytes(), &user)
    assert.NoError(t, err)
    assert.Equal(t, 1, user.ID)
}
```

### 测试运行

```bash
# 运行所有测试
make test

# 运行特定测试
go test ./internal/service/user

# 运行测试并生成覆盖率
make test-cover

# 运行性能测试
make bench
```

## 性能优化

### 数据库优化

1. 使用连接池
2. 合理使用索引
3. 批量操作
4. 预加载关联数据

```go
// 好的查询
users, err := db.Preload("Profile").Where("active = ?", true).Find(&users).Error

// 批量插入
err := db.CreateInBatches(users, 100).Error
```

### 缓存优化

1. 合理设置缓存过期时间
2. 使用多层缓存
3. 缓存预热
4. 缓存穿透保护

```go
// 多层缓存
func GetUserByID(id int) (*User, error) {
    // 1. 检查内存缓存
    if user := memoryCache.Get(id); user != nil {
        return user, nil
    }
    
    // 2. 检查Redis缓存
    if user, err := redisCache.Get(id); err == nil {
        memoryCache.Set(id, user, 5*time.Minute)
        return user, nil
    }
    
    // 3. 查询数据库
    user, err := repository.FindByID(id)
    if err != nil {
        return nil, err
    }
    
    // 4. 设置缓存
    redisCache.Set(id, user, 1*time.Hour)
    memoryCache.Set(id, user, 5*time.Minute)
    
    return user, nil
}
```

### 并发优化

1. 使用Goroutines处理并发
2. 使用Channels进行通信
3. 避免共享内存
4. 使用sync包进行同步

```go
// 并发处理
func ProcessUsers(users []User) error {
    var wg sync.WaitGroup
    errChan := make(chan error, len(users))
    
    for _, user := range users {
        wg.Add(1)
        go func(u User) {
            defer wg.Done()
            if err := processUser(u); err != nil {
                errChan <- err
            }
        }(user)
    }
    
    wg.Wait()
    close(errChan)
    
    for err := range errChan {
        if err != nil {
            return err
        }
    }
    
    return nil
}
```

## 调试

### 日志调试

```go
logger.Debug("Entering function", "func", "GetUserByID", "user_id", id)
logger.Debug("Database query", "query", "SELECT * FROM users WHERE id = ?", id)
logger.Debug("Query result", "user", user)
```

### 性能分析

```go
// 使用pprof
import _ "net/http/pprof"

func main() {
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
    
    // 应用代码
}
```

### 断点调试

使用Delve调试器：

```bash
# 安装delve
go install github.com/go-delve/delve/cmd/dlv@latest

# 调试应用
dlv debug cmd/server/main.go
```

## 常见问题

### 依赖问题

```bash
# 清理模块缓存
go clean -modcache

# 重新下载依赖
go mod download

# 更新依赖
go get -u ./...
```

### 测试问题

```bash
# 清理测试缓存
go clean -testcache

# 运行特定测试
go test -run TestGetUserByID ./internal/service/user

# 详细测试输出
go test -v ./internal/service/user
```

### 构建问题

```bash
# 交叉编译
GOOS=linux GOARCH=amd64 go build -o moviepilot-linux cmd/server/main.go

# 静态编译
CGO_ENABLED=0 go build -a -installsuffix cgo -o moviepilot cmd/server/main.go
```

## 贡献指南

### 提交代码前检查

1. 代码格式化
```bash
make fmt
```

2. 代码检查
```bash
make lint
```

3. 运行测试
```bash
make test
```

4. 静态分析
```bash
make vet
```

### Pull Request要求

1. 清晰的标题和描述
2. 关联相关Issue
3. 包含测试用例
4. 更新相关文档
5. 通过所有检查

### 代码审查

1. 检查代码逻辑
2. 验证测试覆盖
3. 确认性能影响
4. 检查安全性
5. 验证向后兼容性

## 资源链接

- [Go官方文档](https://golang.org/doc/)
- [Gin框架文档](https://gin-gonic.com/docs/)
- [GORM文档](https://gorm.io/docs/)
- [Docker文档](https://docs.docker.com/)
- [Kubernetes文档](https://kubernetes.io/docs/)