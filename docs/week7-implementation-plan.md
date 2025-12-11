# Week 7 实施计划

> **任务**: 用户认证 + 站点管理实施  
> **周期**: Week 7 (Day 1-5)  
> **基于设计**: `docs/design/auth-system-design.md` + `docs/design/site-management-design.md`

---

## 📋 总体目标

实现 Phase 2 的核心功能：
1. ✅ 用户认证系统（注册、登录、权限控制）
2. ✅ 站点管理系统（配置、Cookie 同步、签到调度）

---

## 📅 Day 1: 数据库迁移脚本

### 任务清单

- [ ] 创建数据库迁移目录结构
- [ ] 编写用户认证相关表迁移
  - [ ] `users` 表
  - [ ] `roles` 表
  - [ ] `permissions` 表
  - [ ] `user_roles` 表
  - [ ] `role_permissions` 表
  - [ ] `auth_logs` 表
- [ ] 编写站点管理相关表迁移
  - [ ] `sites` 表
  - [ ] `site_cookies` 表
  - [ ] `checkin_logs` 表
  - [ ] `site_stats` 表
  - [ ] `sync_logs` 表
- [ ] 编写初始数据种子
  - [ ] 默认角色（admin, user, guest）
  - [ ] 默认权限
  - [ ] 角色权限关联
  - [ ] 默认管理员账户

### 文件结构

```
database/
├── migrations/
│   ├── 000001_create_users_table.up.sql
│   ├── 000001_create_users_table.down.sql
│   ├── 000002_create_roles_table.up.sql
│   ├── 000002_create_roles_table.down.sql
│   ├── 000003_create_permissions_table.up.sql
│   ├── 000003_create_permissions_table.down.sql
│   ├── 000004_create_user_roles_table.up.sql
│   ├── 000004_create_user_roles_table.down.sql
│   ├── 000005_create_role_permissions_table.up.sql
│   ├── 000005_create_role_permissions_table.down.sql
│   ├── 000006_create_auth_logs_table.up.sql
│   ├── 000006_create_auth_logs_table.down.sql
│   ├── 000007_create_sites_table.up.sql
│   ├── 000007_create_sites_table.down.sql
│   ├── 000008_create_site_cookies_table.up.sql
│   ├── 000008_create_site_cookies_table.down.sql
│   ├── 000009_create_checkin_logs_table.up.sql
│   ├── 000009_create_checkin_logs_table.down.sql
│   ├── 000010_create_site_stats_table.up.sql
│   ├── 000010_create_site_stats_table.down.sql
│   ├── 000011_create_sync_logs_table.up.sql
│   └── 000011_create_sync_logs_table.down.sql
└── seeds/
    ├── 001_insert_default_roles.sql
    ├── 002_insert_default_permissions.sql
    ├── 003_insert_role_permissions.sql
    └── 004_insert_default_admin.sql
```

### 验收标准

- ✅ 所有迁移脚本可以成功执行
- ✅ 可以回滚迁移
- ✅ 初始数据正确插入
- ✅ 索引创建正确

---

## 📅 Day 2: GORM 模型定义 + Repository 层

### 任务清单

- [ ] 定义用户认证相关模型
  - [ ] `User` 模型
  - [ ] `Role` 模型
  - [ ] `Permission` 模型
  - [ ] `AuthLog` 模型
- [ ] 定义站点管理相关模型
  - [ ] `Site` 模型
  - [ ] `SiteCookie` 模型
  - [ ] `CheckinLog` 模型
  - [ ] `SiteStats` 模型
  - [ ] `SyncLog` 模型
- [ ] 实现 Repository 层
  - [ ] `UserRepository`
  - [ ] `RoleRepository`
  - [ ] `PermissionRepository`
  - [ ] `SiteRepository`
  - [ ] `CheckinLogRepository`

### 文件结构

```
internal/
├── models/
│   ├── user.go
│   ├── role.go
│   ├── permission.go
│   ├── auth_log.go
│   ├── site.go
│   ├── site_cookie.go
│   ├── checkin_log.go
│   ├── site_stats.go
│   └── sync_log.go
└── repositories/
    ├── user_repository.go
    ├── role_repository.go
    ├── permission_repository.go
    ├── site_repository.go
    └── checkin_log_repository.go
```

### 示例代码

#### User 模型

```go
package models

import (
    "time"
    "gorm.io/gorm"
)

type User struct {
    ID            uint           `gorm:"primaryKey" json:"id"`
    Username      string         `gorm:"uniqueIndex;size:50;not null" json:"username"`
    Email         string         `gorm:"uniqueIndex;size:100;not null" json:"email"`
    PasswordHash  string         `gorm:"size:255;not null" json:"-"`
    Nickname      string         `gorm:"size:100" json:"nickname"`
    Avatar        string         `gorm:"size:500" json:"avatar"`
    Status        string         `gorm:"size:20;default:active" json:"status"` // active, disabled, locked
    LastLoginAt   *time.Time     `json:"last_login_at"`
    LastLoginIP   string         `gorm:"size:50" json:"last_login_ip"`
    CreatedAt     time.Time      `json:"created_at"`
    UpdatedAt     time.Time      `json:"updated_at"`
    DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
    
    // 关联
    Roles         []Role         `gorm:"many2many:user_roles;" json:"roles,omitempty"`
}
```

#### UserRepository

```go
package repositories

import (
    "context"
    "gorm.io/gorm"
    "moviepilot-go/internal/models"
)

type UserRepository interface {
    Create(ctx context.Context, user *models.User) error
    GetByID(ctx context.Context, id uint) (*models.User, error)
    GetByUsername(ctx context.Context, username string) (*models.User, error)
    GetByEmail(ctx context.Context, email string) (*models.User, error)
    Update(ctx context.Context, user *models.User) error
    Delete(ctx context.Context, id uint) error
    List(ctx context.Context, page, limit int) ([]*models.User, int64, error)
}

type userRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
    return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
    return r.db.WithContext(ctx).Create(user).Error
}

// ... 其他方法实现
```

### 验收标准

- ✅ 所有模型定义完整
- ✅ GORM 标签正确
- ✅ Repository 接口定义清晰
- ✅ Repository 实现完整

---

## 📅 Day 3: AuthService + JWT Manager + Password Manager

### 任务清单

- [ ] 实现 JWT Manager
  - [ ] 生成 Access Token
  - [ ] 生成 Refresh Token
  - [ ] 验证 Token
  - [ ] 解析 Claims
- [ ] 实现 Password Manager
  - [ ] 密码加密（bcrypt）
  - [ ] 密码验证
  - [ ] 密码强度检查
- [ ] 实现 AuthService
  - [ ] 用户注册
  - [ ] 用户登录
  - [ ] 用户登出
  - [ ] 刷新 Token
  - [ ] 修改密码
  - [ ] 重置密码
- [ ] 实现权限检查逻辑
  - [ ] 检查用户权限
  - [ ] 检查角色权限
  - [ ] 获取用户所有权限

### 文件结构

```
internal/business/services/
├── auth/
│   ├── jwt_manager.go
│   ├── password_manager.go
│   ├── auth_service.go
│   └── permission_service.go
pkg/
└── security/
    ├── jwt.go
    └── password.go
```

### 示例代码

#### JWT Manager

```go
package auth

import (
    "time"
    "github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
    secretKey            string
    accessTokenDuration  time.Duration
    refreshTokenDuration time.Duration
}

type Claims struct {
    UserID      uint     `json:"user_id"`
    Username    string   `json:"username"`
    Roles       []string `json:"roles"`
    Permissions []string `json:"permissions"`
    jwt.RegisteredClaims
}

func (m *JWTManager) GenerateAccessToken(userID uint, username string, roles, permissions []string) (string, error) {
    claims := &Claims{
        UserID:      userID,
        Username:    username,
        Roles:       roles,
        Permissions: permissions,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.accessTokenDuration)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    "moviepilot-go",
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(m.secretKey))
}

// ... 其他方法
```

### 验收标准

- ✅ JWT Token 可以正常生成和验证
- ✅ 密码可以安全加密和验证
- ✅ 用户可以注册和登录
- ✅ 权限检查逻辑正确

---

## 📅 Day 4: SiteService + Cookie 同步 + 签到调度

### 任务清单

- [ ] 实现 SiteService
  - [ ] 站点 CRUD 操作
  - [ ] 站点状态管理
  - [ ] 站点配置验证
- [ ] 实现 CookieService
  - [ ] Cookie 验证
  - [ ] Cookie 同步
  - [ ] Cookie 刷新
  - [ ] 提取用户信息
- [ ] 实现 CheckinService
  - [ ] 手动签到
  - [ ] 自动签到
  - [ ] 签到结果解析
  - [ ] 签到日志记录
- [ ] 实现任务调度器
  - [ ] Cookie 同步任务（每小时）
  - [ ] 自动签到任务（每天）
  - [ ] 任务状态监控

### 文件结构

```
internal/business/services/
├── site/
│   ├── site_service.go
│   ├── cookie_service.go
│   ├── checkin_service.go
│   └── monitor_service.go
internal/schedulers/
├── cookie_sync_scheduler.go
└── checkin_scheduler.go
```

### 示例代码

#### CookieService

```go
package site

import (
    "context"
    "net/http"
    "io"
)

type CookieService struct {
    client *http.Client
}

func (s *CookieService) ValidateCookie(ctx context.Context, site *models.Site) error {
    req, _ := http.NewRequestWithContext(ctx, "GET", site.URL, nil)
    req.Header.Set("Cookie", site.Cookie)
    req.Header.Set("User-Agent", site.UserAgent)
    
    resp, err := s.client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != 200 {
        return fmt.Errorf("invalid status code: %d", resp.StatusCode)
    }
    
    body, _ := io.ReadAll(resp.Body)
    if strings.Contains(string(body), "login") {
        return fmt.Errorf("cookie expired")
    }
    
    return nil
}

// ... 其他方法
```

### 验收标准

- ✅ 站点可以正常添加、编辑、删除
- ✅ Cookie 可以自动同步和验证
- ✅ 签到任务可以定时执行
- ✅ 任务调度器正常工作

---

## 📅 Day 5: API Handler + 中间件 + 单元测试

### 任务清单

- [ ] 实现认证 API Handler
  - [ ] POST `/api/v1/auth/register` - 用户注册
  - [ ] POST `/api/v1/auth/login` - 用户登录
  - [ ] POST `/api/v1/auth/logout` - 用户登出
  - [ ] POST `/api/v1/auth/refresh` - 刷新 Token
  - [ ] PUT `/api/v1/auth/password` - 修改密码
  - [ ] GET `/api/v1/users/me` - 获取当前用户信息
- [ ] 实现站点管理 API Handler
  - [ ] POST `/api/v1/sites` - 创建站点
  - [ ] GET `/api/v1/sites` - 获取站点列表
  - [ ] GET `/api/v1/sites/:id` - 获取站点详情
  - [ ] PUT `/api/v1/sites/:id` - 更新站点
  - [ ] DELETE `/api/v1/sites/:id` - 删除站点
  - [ ] POST `/api/v1/sites/:id/validate` - 验证 Cookie
  - [ ] POST `/api/v1/sites/:id/sync` - 同步站点信息
  - [ ] POST `/api/v1/sites/:id/checkin` - 手动签到
- [ ] 实现中间件
  - [ ] 认证中间件（JWT 验证）
  - [ ] 权限中间件（RBAC 检查）
  - [ ] 限流中间件
- [ ] 编写单元测试
  - [ ] JWT Manager 测试
  - [ ] Password Manager 测试
  - [ ] AuthService 测试
  - [ ] SiteService 测试
  - [ ] API Handler 测试

### 文件结构

```
internal/apis/handlers/
├── auth/
│   └── handler.go
├── site/
│   └── handler.go
internal/apis/middlewares/
├── auth_middleware.go
└── permission_middleware.go
tests/
├── unit/
│   ├── jwt_manager_test.go
│   ├── password_manager_test.go
│   ├── auth_service_test.go
│   └── site_service_test.go
└── integration/
    ├── auth_api_test.go
    └── site_api_test.go
```

### 示例代码

#### 认证中间件

```go
package middlewares

import (
    "github.com/gin-gonic/gin"
    "moviepilot-go/internal/business/services/auth"
)

func AuthMiddleware(jwtManager *auth.JWTManager) gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.JSON(401, gin.H{"error": "未提供令牌"})
            c.Abort()
            return
        }
        
        // 移除 "Bearer " 前缀
        if len(token) > 7 && token[:7] == "Bearer " {
            token = token[7:]
        }
        
        claims, err := jwtManager.VerifyToken(token)
        if err != nil {
            c.JSON(401, gin.H{"error": "无效的令牌"})
            c.Abort()
            return
        }
        
        c.Set("user_id", claims.UserID)
        c.Set("username", claims.Username)
        c.Set("roles", claims.Roles)
        c.Set("permissions", claims.Permissions)
        
        c.Next()
    }
}
```

### 验收标准

- ✅ 所有 API 可以正常访问
- ✅ 认证中间件正常工作
- ✅ 权限中间件正常工作
- ✅ 单元测试覆盖率 > 80%
- ✅ 集成测试通过

---

## 📊 Week 7 交付物

### 代码文件

**数据库**:
1. 11 个迁移脚本（up + down）
2. 4 个种子数据脚本

**模型**:
3. 9 个 GORM 模型文件

**Repository**:
4. 5 个 Repository 接口和实现

**Service**:
5. 7 个 Service 文件（Auth + Site）

**API Handler**:
6. 2 个 Handler 文件

**中间件**:
7. 2 个中间件文件

**调度器**:
8. 2 个调度器文件

**测试**:
9. 10+ 个测试文件

### 文档

1. Week 7 实施计划（本文档）
2. Week 7 完成总结
3. API 文档更新

---

## 🎯 验收标准

### 功能验收

- ✅ 用户可以注册和登录
- ✅ JWT Token 正常生成和验证
- ✅ 权限控制正常工作
- ✅ 站点可以正常管理
- ✅ Cookie 可以自动同步
- ✅ 签到任务可以定时执行

### 质量验收

- ✅ 代码编译通过
- ✅ 单元测试覆盖率 > 80%
- ✅ 集成测试通过
- ✅ 代码符合规范
- ✅ 文档完整

### 性能验收

- ✅ 登录响应时间 < 500ms
- ✅ API 响应时间 < 1s
- ✅ 并发 100 用户无压力

---

## 📝 注意事项

1. **安全第一**
   - 密码必须使用 bcrypt 加密
   - JWT Secret 必须从环境变量读取
   - 敏感信息不能记录到日志

2. **错误处理**
   - 所有错误必须有详细的错误信息
   - 使用统一的错误响应格式
   - 记录错误日志

3. **性能优化**
   - 使用数据库索引
   - 使用连接池
   - 避免 N+1 查询

4. **测试覆盖**
   - 核心逻辑必须有单元测试
   - API 必须有集成测试
   - 边界条件必须测试

---

**文档状态**: ✅ 计划完成，准备实施  
**开始时间**: 2025-12-02  
**预计完成**: 2025-12-06
