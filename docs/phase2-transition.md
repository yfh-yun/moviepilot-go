# Phase 2 过渡文档

> **从 Week 6 到 Week 7**  
> **从设计到实施**  
> **时间**: 2025-12-02

---

## 📊 Phase 1 回顾（Week 1-6）

### 已完成的核心功能

| 周次 | 任务 | 状态 |
|------|------|------|
| Week 4 | 数据库优化 + 下载器集成 | ✅ 100% |
| Week 5 | 媒体服务器 + 元数据平台集成 | ✅ 100% |
| Week 6 | 通知渠道 + 索引器集成 + Phase 2 设计 | ✅ 100% |

### 技术积累

**已实现的集成模块**:
1. ✅ 下载器集成（qBittorrent + Transmission）
2. ✅ 媒体服务器集成（Emby + Plex + Jellyfin）
3. ✅ 元数据平台集成（TMDB + TVDB + 豆瓣）
4. ✅ 通知渠道集成（Telegram + WeChat）
5. ✅ 索引器集成（Jackett + Prowlarr）

**技术亮点**:
- 统一接口设计
- 工厂模式管理
- 并发处理
- 容错机制
- 协议支持（Torznab、JWT）

**代码统计**:
- 总代码行数：12,000+
- 总文档行数：8,500+
- 总文件数：57

---

## 🎯 Phase 2 目标（Week 7-9）

### 核心功能开发

**Week 7: 用户认证 + 站点管理**
- 用户注册/登录系统
- JWT 认证机制
- RBAC 权限控制
- 站点配置管理
- Cookie 自动同步
- 定时签到调度

**Week 8: 订阅系统 + 下载管理**
- 订阅 CRUD 操作
- 订阅刷新调度
- 订阅匹配引擎
- 下载任务管理
- 下载状态同步

**Week 9: 文件整理 + 集成测试**
- 文件识别和重命名
- 文件移动和整理
- 媒体服务器同步
- 端到端测试
- 性能测试

---

## 📋 Week 6 → Week 7 过渡

### Week 6 Day 5 设计成果

**用户认证系统设计**（`docs/design/auth-system-design.md`）:
- ✅ 系统架构设计
- ✅ 用户注册/登录流程
- ✅ RBAC 权限模型
- ✅ 数据库设计（6 张表）
- ✅ API 设计（10+ 接口）
- ✅ 安全策略

**站点管理系统设计**（`docs/design/site-management-design.md`）:
- ✅ 系统架构设计
- ✅ 站点配置模型
- ✅ Cookie 同步机制
- ✅ 签到任务调度
- ✅ 数据库设计（5 张表）
- ✅ API 设计（15+ 接口）

### Week 7 实施计划

**Day 1: 数据库迁移**
- 创建 11 张表
- 插入初始数据
- 验证迁移

**Day 2: 模型和 Repository**
- 定义 9 个 GORM 模型
- 实现 5 个 Repository
- 编写单元测试

**Day 3: 认证服务**
- 实现 JWT Manager
- 实现 Password Manager
- 实现 AuthService
- 实现 PermissionService

**Day 4: 站点服务和调度**
- 实现 SiteService
- 实现 CookieService
- 实现 CheckinService
- 实现任务调度器

**Day 5: API 和中间件**
- 实现认证 API（6 个接口）
- 实现站点管理 API（8 个接口）
- 实现认证中间件
- 实现权限中间件
- 编写集成测试

---

## 🔄 从设计到实施的映射

### 数据库表映射

| 设计文档中的表 | 迁移脚本 | GORM 模型 |
|---------------|---------|----------|
| users | 000001_create_users_table.sql | models.User |
| roles | 000002_create_roles_table.sql | models.Role |
| permissions | 000003_create_permissions_table.sql | models.Permission |
| user_roles | 000004_create_user_roles_table.sql | - |
| role_permissions | 000005_create_role_permissions_table.sql | - |
| auth_logs | 000006_create_auth_logs_table.sql | models.AuthLog |
| sites | 000007_create_sites_table.sql | models.Site |
| site_cookies | 000008_create_site_cookies_table.sql | models.SiteCookie |
| checkin_logs | 000009_create_checkin_logs_table.sql | models.CheckinLog |
| site_stats | 000010_create_site_stats_table.sql | models.SiteStats |
| sync_logs | 000011_create_sync_logs_table.sql | models.SyncLog |

### API 接口映射

| 设计文档中的 API | Handler 方法 | 路由 |
|-----------------|-------------|------|
| POST /auth/register | auth.Register | POST /api/v1/auth/register |
| POST /auth/login | auth.Login | POST /api/v1/auth/login |
| POST /auth/logout | auth.Logout | POST /api/v1/auth/logout |
| POST /auth/refresh | auth.RefreshToken | POST /api/v1/auth/refresh |
| PUT /auth/password | auth.ChangePassword | PUT /api/v1/auth/password |
| GET /users/me | auth.GetCurrentUser | GET /api/v1/users/me |
| POST /sites | site.Create | POST /api/v1/sites |
| GET /sites | site.List | GET /api/v1/sites |
| GET /sites/:id | site.Get | GET /api/v1/sites/:id |
| PUT /sites/:id | site.Update | PUT /api/v1/sites/:id |
| DELETE /sites/:id | site.Delete | DELETE /api/v1/sites/:id |
| POST /sites/:id/validate | site.ValidateCookie | POST /api/v1/sites/:id/validate |
| POST /sites/:id/sync | site.Sync | POST /api/v1/sites/:id/sync |
| POST /sites/:id/checkin | site.Checkin | POST /api/v1/sites/:id/checkin |

### 服务层映射

| 设计文档中的服务 | 实现文件 |
|-----------------|---------|
| JWT Manager | internal/business/services/auth/jwt_manager.go |
| Password Manager | internal/business/services/auth/password_manager.go |
| AuthService | internal/business/services/auth/auth_service.go |
| PermissionService | internal/business/services/auth/permission_service.go |
| SiteService | internal/business/services/site/site_service.go |
| CookieService | internal/business/services/site/cookie_service.go |
| CheckinService | internal/business/services/site/checkin_service.go |
| MonitorService | internal/business/services/site/monitor_service.go |

---

## 🛠️ 技术栈变化

### 新增依赖

**Week 7 需要添加的依赖**:

```bash
# JWT 认证
go get github.com/golang-jwt/jwt/v5

# 密码加密
go get golang.org/x/crypto/bcrypt

# 任务调度
go get github.com/robfig/cron/v3

# 数据库迁移
go get -tags 'postgres' github.com/golang-migrate/migrate/v4
```

### 配置文件更新

**新增配置项**:

```yaml
# JWT 配置
jwt:
  secret: "your-secret-key"
  access_token_expire: 900    # 15 分钟
  refresh_token_expire: 604800 # 7 天

# 密码策略
password:
  min_length: 8
  require_uppercase: true
  require_lowercase: true
  require_number: true

# 调度器配置
scheduler:
  cookie_sync_cron: "0 * * * *"    # 每小时
  checkin_cron: "0 8 * * *"        # 每天 8:00
```

---

## 📈 进度对比

### Week 6 vs Week 7

| 指标 | Week 6 | Week 7（预计） |
|------|--------|---------------|
| 任务类型 | 集成 + 设计 | 实施 |
| 新增文件 | 15 | 30+ |
| 新增代码行数 | 1,800+ | 3,000+ |
| 新增文档行数 | 3,000+ | 1,000+ |
| 数据库表 | 0 | 11 |
| API 接口 | 0 | 14 |
| 单元测试 | 0 | 20+ |

---

## 🎯 关键里程碑

### Phase 2 的第一个里程碑

**Week 7 完成后将实现**:
1. ✅ 用户可以注册和登录
2. ✅ 系统有完整的权限控制
3. ✅ 用户可以管理站点
4. ✅ 站点可以自动签到
5. ✅ Cookie 可以自动同步

**这意味着**:
- MoviePilot Go 有了用户系统
- 可以开始多用户支持
- 站点管理功能完整
- 为订阅系统打下基础

---

## 🚀 准备就绪检查清单

### 环境准备

- [x] PostgreSQL 数据库已安装
- [x] Go 1.21+ 已安装
- [x] 开发工具已配置
- [x] 环境变量已设置

### 文档准备

- [x] 设计文档已完成
- [x] 实施计划已制定
- [x] 启动文档已创建
- [x] 过渡文档已编写

### 技术准备

- [x] 依赖包已了解
- [x] 数据库设计已确认
- [x] API 设计已确认
- [x] 安全策略已确认

---

## 💡 实施建议

### 开发顺序

1. **先数据库，后代码**
   - 确保数据库结构正确
   - 数据库是基础

2. **先模型，后服务**
   - 模型定义清晰
   - 服务依赖模型

3. **先服务，后 API**
   - 业务逻辑独立
   - API 只是接口

4. **先功能，后测试**
   - 功能实现完整
   - 测试覆盖全面

### 质量保证

1. **每天提交代码**
   - 保持小步快跑
   - 便于回滚

2. **每天运行测试**
   - 及时发现问题
   - 保证质量

3. **每天更新文档**
   - 记录变更
   - 便于追溯

---

## 📝 下一步行动

### 立即开始

1. **创建数据库迁移脚本**
   - 参考设计文档中的表结构
   - 使用 golang-migrate 工具
   - 先创建 users 表

2. **设置开发环境**
   - 创建 .env 文件
   - 配置数据库连接
   - 安装必要依赖

3. **创建项目结构**
   - 创建必要的目录
   - 准备文件模板
   - 设置 Git 分支

### 本周目标

**Week 7 结束时**:
- ✅ 11 张数据库表创建完成
- ✅ 9 个 GORM 模型定义完成
- ✅ 7 个服务实现完成
- ✅ 14 个 API 接口实现完成
- ✅ 单元测试覆盖率 > 80%
- ✅ 集成测试通过

---

## 🎉 Let's Build!

从设计到实施，从 Week 6 到 Week 7，让我们开始构建 MoviePilot Go 的核心功能！

**下一步**: 开始 Week 7 Day 1 - 数据库迁移

---

**文档创建时间**: 2025-12-02  
**状态**: ✅ 准备就绪  
**信心指数**: ⭐⭐⭐⭐⭐
