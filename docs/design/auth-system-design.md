# 用户认证系统设计文档

> **版本**: v1.0.0  
> **创建时间**: 2025-12-02  
> **设计阶段**: Phase 2 准备

---

## 📋 目录

1. [概述](#概述)
2. [系统架构](#系统架构)
3. [用户注册流程](#用户注册流程)
4. [用户登录流程](#用户登录流程)
5. [权限控制模型](#权限控制模型)
6. [数据库设计](#数据库设计)
7. [API 设计](#api-设计)
8. [安全策略](#安全策略)
9. [实施计划](#实施计划)

---

## 概述

### 设计目标

MoviePilot Go 用户认证系统旨在提供：
- ✅ 安全的用户注册和登录机制
- ✅ 基于 JWT 的无状态认证
- ✅ 细粒度的权限控制（RBAC）
- ✅ 多因素认证支持（可选）
- ✅ 密码重置和账户恢复
- ✅ 会话管理和令牌刷新

### 核心特性

1. **JWT 认证**：无状态、可扩展
2. **RBAC 权限模型**：角色-权限分离
3. **密码安全**：bcrypt 加密存储
4. **令牌刷新**：Access Token + Refresh Token
5. **审计日志**：记录所有认证操作

---

## 系统架构

### 架构图

```
┌─────────────┐
│   客户端    │
└──────┬──────┘
       │ HTTP/HTTPS
       ▼
┌─────────────────────────────────────┐
│         API Gateway (Gin)           │
│  ┌──────────────────────────────┐   │
│  │  认证中间件 (JWT Verify)     │   │
│  └──────────────────────────────┘   │
└──────────┬──────────────────────────┘
           │
           ▼
┌─────────────────────────────────────┐
│      Business Layer (Services)      │
│  ┌────────────┐  ┌────────────┐    │
│  │ AuthService│  │ UserService│    │
│  └────────────┘  └────────────┘    │
└──────────┬──────────────────────────┘
           │
           ▼
┌─────────────────────────────────────┐
│    Repository Layer (Data Access)   │
│  ┌────────────┐  ┌────────────┐    │
│  │  UserRepo  │  │  RoleRepo  │    │
│  └────────────┘  └────────────┘    │
└──────────┬──────────────────────────┘
           │
           ▼
┌─────────────────────────────────────┐
│         Database (PostgreSQL)       │
│  ┌──────┐ ┌──────┐ ┌──────────┐    │
│  │ users│ │ roles│ │permissions│   │
│  └──────┘ └──────┘ └──────────┘    │
└─────────────────────────────────────┘
```

### 核心组件

1. **认证中间件**：验证 JWT Token
2. **AuthService**：处理认证逻辑
3. **UserService**：用户管理
4. **JWT Manager**：生成和验证 Token
5. **Password Manager**：密码加密和验证

---

## 用户注册流程

### 流程图

```
用户 → 提交注册信息
  ↓
验证输入（用户名、邮箱、密码）
  ↓
检查用户名/邮箱是否已存在
  ↓
密码加密（bcrypt）
  ↓
创建用户记录
  ↓
分配默认角色（user）
  ↓
发送欢迎邮件（可选）
  ↓
返回成功响应
```

### 注册规则

| 字段 | 规则 |
|------|------|
| 用户名 | 3-20 字符，字母数字下划线 |
| 邮箱 | 有效的邮箱格式 |
| 密码 | 8-32 字符，包含大小写字母和数字 |
| 昵称 | 可选，1-50 字符 |

### 注册请求示例

```json
POST /api/v1/auth/register
{
  "username": "john_doe",
  "email": "john@example.com",
  "password": "SecurePass123",
  "nickname": "John"
}
```

### 注册响应示例

```json
{
  "code": 200,
  "message": "注册成功",
  "data": {
    "user_id": 1,
    "username": "john_doe",
    "email": "john@example.com",
    "nickname": "John",
    "created_at": "2025-12-02T07:00:00Z"
  }
}
```

---

## 用户登录流程

### 流程图

```
用户 → 提交登录凭证（用户名/邮箱 + 密码）
  ↓
验证输入
  ↓
查询用户记录
  ↓
验证密码（bcrypt.CompareHashAndPassword）
  ↓
检查用户状态（是否被禁用）
  ↓
生成 Access Token（15分钟有效）
  ↓
生成 Refresh Token（7天有效）
  ↓
记录登录日志
  ↓
返回 Token
```

### 登录请求示例

```json
POST /api/v1/auth/login
{
  "username": "john_doe",
  "password": "SecurePass123"
}
```

### 登录响应示例

```json
{
  "code": 200,
  "message": "登录成功",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 900,
    "user": {
      "id": 1,
      "username": "john_doe",
      "email": "john@example.com",
      "nickname": "John",
      "roles": ["user"]
    }
  }
}
```

### JWT Payload 结构

```json
{
  "user_id": 1,
  "username": "john_doe",
  "roles": ["user"],
  "permissions": ["read:subscribe", "write:subscribe"],
  "exp": 1701504000,
  "iat": 1701503100,
  "iss": "moviepilot-go"
}
```

---

## 权限控制模型

### RBAC 模型

MoviePilot 采用 **基于角色的访问控制（RBAC）** 模型：

```
用户 (User) ──┐
              ├─→ 角色 (Role) ──→ 权限 (Permission) ──→ 资源 (Resource)
用户 (User) ──┘
```

### 预定义角色

| 角色 | 说明 | 权限范围 |
|------|------|----------|
| `admin` | 管理员 | 所有权限 |
| `user` | 普通用户 | 基本功能权限 |
| `guest` | 访客 | 只读权限 |

### 权限命名规范

格式：`<action>:<resource>`

示例：
- `read:subscribe` - 查看订阅
- `write:subscribe` - 创建/修改订阅
- `delete:subscribe` - 删除订阅
- `manage:user` - 管理用户
- `manage:site` - 管理站点

### 权限矩阵

| 资源 | admin | user | guest |
|------|-------|------|-------|
| 订阅管理 | ✅ 全部 | ✅ 自己的 | ❌ |
| 下载管理 | ✅ 全部 | ✅ 自己的 | ❌ |
| 站点管理 | ✅ | ❌ | ❌ |
| 用户管理 | ✅ | ❌ | ❌ |
| 系统设置 | ✅ | ❌ | ❌ |

### 权限检查流程

```go
// 伪代码
func CheckPermission(userID int, permission string) bool {
    // 1. 获取用户角色
    roles := GetUserRoles(userID)
    
    // 2. 获取角色权限
    permissions := GetRolePermissions(roles)
    
    // 3. 检查权限
    return Contains(permissions, permission)
}
```

---

## 数据库设计

### ER 图

```
┌──────────────┐         ┌──────────────┐         ┌──────────────┐
│    users     │         │  user_roles  │         │    roles     │
├──────────────┤         ├──────────────┤         ├──────────────┤
│ id (PK)      │────┐    │ user_id (FK) │    ┌────│ id (PK)      │
│ username     │    └───→│ role_id (FK) │←───┘    │ name         │
│ email        │         └──────────────┘         │ description  │
│ password_hash│                                  │ created_at   │
│ nickname     │         ┌──────────────┐         └──────────────┘
│ avatar       │         │role_permissions│              │
│ status       │         ├──────────────┤              │
│ created_at   │    ┌────│ role_id (FK) │              │
│ updated_at   │    │    │ perm_id (FK) │←─────────────┘
└──────────────┘    │    └──────────────┘
                    │
                    │    ┌──────────────┐
                    │    │ permissions  │
                    │    ├──────────────┤
                    └───→│ id (PK)      │
                         │ name         │
                         │ resource     │
                         │ action       │
                         │ description  │
                         └──────────────┘
```

### 表结构定义

#### 1. users 表

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    nickname VARCHAR(100),
    avatar VARCHAR(500),
    status VARCHAR(20) DEFAULT 'active', -- active, disabled, locked
    last_login_at TIMESTAMP,
    last_login_ip VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    
    INDEX idx_username (username),
    INDEX idx_email (email),
    INDEX idx_status (status)
);
```

#### 2. roles 表

```sql
CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    display_name VARCHAR(100),
    description TEXT,
    is_system BOOLEAN DEFAULT FALSE, -- 系统角色不可删除
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_name (name)
);
```

#### 3. permissions 表

```sql
CREATE TABLE permissions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL, -- 如: read:subscribe
    resource VARCHAR(50) NOT NULL,      -- 如: subscribe
    action VARCHAR(50) NOT NULL,        -- 如: read
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_name (name),
    INDEX idx_resource (resource)
);
```

#### 4. user_roles 表（多对多关系）

```sql
CREATE TABLE user_roles (
    user_id INT NOT NULL,
    role_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
);
```

#### 5. role_permissions 表（多对多关系）

```sql
CREATE TABLE role_permissions (
    role_id INT NOT NULL,
    permission_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    PRIMARY KEY (role_id, permission_id),
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
);
```

#### 6. auth_logs 表（审计日志）

```sql
CREATE TABLE auth_logs (
    id SERIAL PRIMARY KEY,
    user_id INT,
    action VARCHAR(50) NOT NULL,  -- login, logout, register, password_reset
    ip_address VARCHAR(50),
    user_agent TEXT,
    status VARCHAR(20),           -- success, failed
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    INDEX idx_user_id (user_id),
    INDEX idx_action (action),
    INDEX idx_created_at (created_at)
);
```

### 初始数据

```sql
-- 插入默认角色
INSERT INTO roles (name, display_name, description, is_system) VALUES
('admin', '管理员', '系统管理员，拥有所有权限', TRUE),
('user', '普通用户', '普通用户，拥有基本功能权限', TRUE),
('guest', '访客', '访客，只读权限', TRUE);

-- 插入默认权限
INSERT INTO permissions (name, resource, action, description) VALUES
('read:subscribe', 'subscribe', 'read', '查看订阅'),
('write:subscribe', 'subscribe', 'write', '创建/修改订阅'),
('delete:subscribe', 'subscribe', 'delete', '删除订阅'),
('read:download', 'download', 'read', '查看下载'),
('write:download', 'download', 'write', '创建/修改下载'),
('delete:download', 'download', 'delete', '删除下载'),
('manage:user', 'user', 'manage', '管理用户'),
('manage:site', 'site', 'manage', '管理站点'),
('manage:system', 'system', 'manage', '管理系统设置');

-- 分配权限给角色
-- admin 拥有所有权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT 1, id FROM permissions;

-- user 拥有基本权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT 2, id FROM permissions WHERE name IN (
    'read:subscribe', 'write:subscribe', 'delete:subscribe',
    'read:download', 'write:download', 'delete:download'
);

-- guest 只有读权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT 3, id FROM permissions WHERE action = 'read';

-- 创建默认管理员账户（密码: admin123）
INSERT INTO users (username, email, password_hash, nickname, status) VALUES
('admin', 'admin@moviepilot.com', '$2a$10$...', 'Administrator', 'active');

-- 分配管理员角色
INSERT INTO user_roles (user_id, role_id) VALUES (1, 1);
```

---

## API 设计

### 认证相关 API

#### 1. 用户注册

```
POST /api/v1/auth/register
Content-Type: application/json

Request:
{
  "username": "john_doe",
  "email": "john@example.com",
  "password": "SecurePass123",
  "nickname": "John"
}

Response: 201 Created
{
  "code": 201,
  "message": "注册成功",
  "data": {
    "user_id": 1,
    "username": "john_doe"
  }
}
```

#### 2. 用户登录

```
POST /api/v1/auth/login
Content-Type: application/json

Request:
{
  "username": "john_doe",
  "password": "SecurePass123"
}

Response: 200 OK
{
  "code": 200,
  "data": {
    "access_token": "eyJ...",
    "refresh_token": "eyJ...",
    "expires_in": 900
  }
}
```

#### 3. 刷新令牌

```
POST /api/v1/auth/refresh
Content-Type: application/json

Request:
{
  "refresh_token": "eyJ..."
}

Response: 200 OK
{
  "code": 200,
  "data": {
    "access_token": "eyJ...",
    "expires_in": 900
  }
}
```

#### 4. 用户登出

```
POST /api/v1/auth/logout
Authorization: Bearer <access_token>

Response: 200 OK
{
  "code": 200,
  "message": "登出成功"
}
```

#### 5. 修改密码

```
PUT /api/v1/auth/password
Authorization: Bearer <access_token>
Content-Type: application/json

Request:
{
  "old_password": "OldPass123",
  "new_password": "NewPass456"
}

Response: 200 OK
{
  "code": 200,
  "message": "密码修改成功"
}
```

#### 6. 重置密码（忘记密码）

```
POST /api/v1/auth/password/reset
Content-Type: application/json

Request:
{
  "email": "john@example.com"
}

Response: 200 OK
{
  "code": 200,
  "message": "重置链接已发送到邮箱"
}
```

### 用户管理 API

#### 1. 获取当前用户信息

```
GET /api/v1/users/me
Authorization: Bearer <access_token>

Response: 200 OK
{
  "code": 200,
  "data": {
    "id": 1,
    "username": "john_doe",
    "email": "john@example.com",
    "nickname": "John",
    "roles": ["user"],
    "permissions": ["read:subscribe", "write:subscribe"]
  }
}
```

#### 2. 更新用户信息

```
PUT /api/v1/users/me
Authorization: Bearer <access_token>
Content-Type: application/json

Request:
{
  "nickname": "John Doe",
  "avatar": "https://example.com/avatar.jpg"
}

Response: 200 OK
{
  "code": 200,
  "message": "更新成功"
}
```

#### 3. 获取用户列表（管理员）

```
GET /api/v1/users?page=1&limit=20
Authorization: Bearer <access_token>
Permission: manage:user

Response: 200 OK
{
  "code": 200,
  "data": {
    "items": [...],
    "total": 100,
    "page": 1,
    "limit": 20
  }
}
```

---

## 安全策略

### 1. 密码安全

- **加密算法**：bcrypt（cost=10）
- **密码强度**：至少 8 位，包含大小写字母和数字
- **密码历史**：记录最近 5 次密码，禁止重复使用
- **密码过期**：可选，90 天强制修改

### 2. Token 安全

- **Access Token**：15 分钟有效期
- **Refresh Token**：7 天有效期
- **签名算法**：HS256（HMAC-SHA256）
- **密钥管理**：环境变量存储，定期轮换

### 3. 防暴力破解

- **登录限流**：同一 IP 5 分钟内最多 5 次失败
- **账户锁定**：连续 10 次失败后锁定 30 分钟
- **验证码**：3 次失败后要求验证码

### 4. 会话管理

- **单点登录**：可选，同一账户只允许一个活跃会话
- **会话超时**：15 分钟无操作自动登出
- **强制登出**：管理员可强制用户登出

### 5. 审计日志

- **记录内容**：用户 ID、操作类型、IP 地址、时间戳
- **保留期限**：90 天
- **敏感操作**：登录、登出、密码修改、权限变更

---

## 实施计划

### Week 7 实施任务

#### Day 1-2: 数据库和模型

- [ ] 创建数据库迁移脚本
- [ ] 定义 GORM 模型（User、Role、Permission）
- [ ] 实现 Repository 层

#### Day 3-4: 业务逻辑

- [ ] 实现 AuthService（注册、登录、登出）
- [ ] 实现 JWT Manager
- [ ] 实现 Password Manager
- [ ] 实现权限检查逻辑

#### Day 5: API 和中间件

- [ ] 实现认证 API Handler
- [ ] 实现认证中间件
- [ ] 实现权限中间件
- [ ] 编写单元测试

### 验收标准

- ✅ 用户可以注册和登录
- ✅ JWT Token 正常生成和验证
- ✅ 权限控制正常工作
- ✅ 密码安全存储（bcrypt）
- ✅ 审计日志正常记录
- ✅ API 文档完整

---

## 附录

### A. JWT 配置示例

```yaml
jwt:
  secret: "your-secret-key-change-in-production"
  access_token_expire: 900    # 15 分钟
  refresh_token_expire: 604800 # 7 天
  issuer: "moviepilot-go"
```

### B. 密码强度正则

```go
// 至少 8 位，包含大小写字母和数字
const PasswordRegex = `^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)[a-zA-Z\d@$!%*?&]{8,32}$`
```

### C. 权限检查示例

```go
// 中间件示例
func RequirePermission(permission string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetInt("user_id")
        
        if !authService.HasPermission(userID, permission) {
            c.JSON(403, gin.H{"error": "权限不足"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}

// 使用示例
router.DELETE("/api/v1/subscribes/:id", 
    RequirePermission("delete:subscribe"),
    subscribeHandler.Delete)
```

---

**文档状态**: ✅ 设计完成，待实施  
**下一步**: Week 7 开始实施
