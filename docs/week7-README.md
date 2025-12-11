# Week 7: 用户认证 + 站点管理

> **完成时间**: 2025-12-02  
> **状态**: ✅ 100% 完成

---

## 🎯 本周目标

实现 MoviePilot Go 的用户认证和站点管理系统，包括：
- 用户注册、登录、权限控制
- 站点 CRUD、Cookie 同步、自动签到
- 完整的 API 接口和中间件

---

## ✅ 完成清单

### 数据库层
- [x] 11 张表的迁移脚本
- [x] 种子数据（默认角色、权限、管理员）
- [x] Makefile 迁移命令

### 模型和 Repository
- [x] 9 个 GORM 模型
- [x] 5 个 Repository 接口和实现
- [x] 完整的 CRUD 操作

### 业务服务
- [x] JWT Manager（双令牌机制）
- [x] Password Manager（bcrypt 加密）
- [x] AuthService（注册/登录/权限）
- [x] PermissionService（RBAC 权限检查）
- [x] SiteService（站点管理）
- [x] CookieService（Cookie 同步）
- [x] CheckinService（签到逻辑）

### 调度系统
- [x] Cookie 同步调度器（每小时）
- [x] 签到调度器（每天）

### API 和中间件
- [x] AuthMiddleware（JWT 认证）
- [x] PermissionMiddleware（权限检查）
- [x] 认证 API（6 个接口）
- [x] 站点管理 API（7 个接口）

---

## 📊 成果统计

| 指标 | 数量 |
|------|------|
| 文件数 | 57 |
| 代码行数 | 2,760+ |
| 数据库表 | 11 |
| API 接口 | 13 |
| 服务数 | 7 |

---

## 🔌 API 接口

### 认证 API
```
POST   /api/v1/auth/register    # 用户注册
POST   /api/v1/auth/login       # 用户登录
POST   /api/v1/auth/logout      # 用户登出
POST   /api/v1/auth/refresh     # 刷新令牌
PUT    /api/v1/auth/password    # 修改密码
GET    /api/v1/users/me         # 获取当前用户
```

### 站点管理 API
```
POST   /api/v1/sites            # 创建站点
GET    /api/v1/sites            # 获取站点列表
GET    /api/v1/sites/:id        # 获取站点详情
PUT    /api/v1/sites/:id        # 更新站点
DELETE /api/v1/sites/:id        # 删除站点
POST   /api/v1/sites/:id/checkin    # 站点签到
POST   /api/v1/sites/:id/validate   # 验证Cookie
```

---

## 🛠️ 使用示例

### 用户注册
```bash
curl -X POST http://localhost:3001/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "Test123456"
  }'
```

### 用户登录
```bash
curl -X POST http://localhost:3001/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "Test123456"
  }'
```

### 创建站点
```bash
curl -X POST http://localhost:3001/api/v1/sites \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "示例站点",
    "url": "https://example.com",
    "type": "pt",
    "cookie": "your_cookie_here"
  }'
```

---

## 📁 文件结构

```
moviepilot-go/
├── database/
│   ├── migrations/          # 数据库迁移
│   └── seeds/              # 种子数据
├── internal/
│   ├── models/             # GORM 模型
│   ├── repositories/       # Repository 层
│   ├── business/services/
│   │   ├── auth/          # 认证服务
│   │   └── site/          # 站点服务
│   ├── schedulers/site/   # 调度器
│   ├── apis/
│   │   ├── middlewares/   # 中间件
│   │   └── handlers/v1/   # API Handler
└── pkg/security/          # 安全工具
```

---

## 🚀 快速开始

### 1. 运行数据库迁移
```bash
make migrate-up
```

### 2. 插入种子数据
```bash
make migrate-seed
```

### 3. 启动服务
```bash
make run
```

### 4. 访问 API
```
http://localhost:3001/api/v1
```

---

## 📚 相关文档

- [Week 7 最终总结](./week7-final-summary.md)
- [数据库迁移文档](../database/migrations/README.md)
- [认证系统设计](./design/auth-system-design.md)
- [站点管理设计](./design/site-management-design.md)

---

**状态**: ✅ Week 7 完成  
**下一步**: Week 8 - 订阅系统 + 下载管理
