# Week 7 启动文档

> **阶段**: Phase 2 核心功能开发  
> **任务**: 用户认证 + 站点管理实施  
> **时间**: 2025-12-02 开始

---

## 🎯 本周目标

实现 MoviePilot Go 的两个核心系统：

1. **用户认证系统**
   - 用户注册/登录
   - JWT Token 管理
   - RBAC 权限控制
   - 密码安全管理

2. **站点管理系统**
   - 站点配置管理
   - Cookie 自动同步
   - 定时签到调度
   - 流量统计监控

---

## 📚 前置准备

### 已完成的设计文档

1. ✅ `docs/design/auth-system-design.md` - 用户认证系统设计
2. ✅ `docs/design/site-management-design.md` - 站点管理系统设计

### 技术栈确认

- **数据库**: PostgreSQL 14+
- **ORM**: GORM v2
- **JWT**: golang-jwt/jwt v5
- **密码加密**: bcrypt
- **调度器**: robfig/cron v3
- **HTTP 框架**: Gin

---

## 📅 5 天实施计划

### Day 1: 数据库迁移 (2025-12-02)

**目标**: 创建所有数据库表和初始数据

**任务**:
- 创建 11 张表的迁移脚本
- 编写初始数据种子
- 验证迁移可以正常执行和回滚

**交付物**:
- `database/migrations/` - 22 个迁移文件（up + down）
- `database/seeds/` - 4 个种子数据文件

**验收**:
```bash
# 执行迁移
make migrate-up

# 验证表结构
psql -d moviepilot -c "\dt"

# 验证初始数据
psql -d moviepilot -c "SELECT * FROM roles;"
```

---

### Day 2: 模型和 Repository (2025-12-03)

**目标**: 定义所有 GORM 模型和数据访问层

**任务**:
- 定义 9 个 GORM 模型
- 实现 5 个 Repository 接口
- 编写 Repository 单元测试

**交付物**:
- `internal/models/` - 9 个模型文件
- `internal/repositories/` - 5 个 Repository 文件
- `tests/unit/repositories/` - 5 个测试文件

**验收**:
```bash
# 编译检查
go build ./internal/models/...
go build ./internal/repositories/...

# 运行测试
go test ./internal/repositories/... -v
```

---

### Day 3: 认证服务 (2025-12-04)

**目标**: 实现完整的用户认证逻辑

**任务**:
- 实现 JWT Manager
- 实现 Password Manager
- 实现 AuthService
- 实现 PermissionService
- 编写服务层单元测试

**交付物**:
- `internal/business/services/auth/` - 4 个服务文件
- `pkg/security/` - 2 个安全工具文件
- `tests/unit/services/` - 4 个测试文件

**验收**:
```bash
# 运行测试
go test ./internal/business/services/auth/... -v

# 测试覆盖率
go test ./internal/business/services/auth/... -cover
```

---

### Day 4: 站点服务和调度 (2025-12-05)

**目标**: 实现站点管理和自动化任务

**任务**:
- 实现 SiteService
- 实现 CookieService
- 实现 CheckinService
- 实现任务调度器
- 编写服务层单元测试

**交付物**:
- `internal/business/services/site/` - 4 个服务文件
- `internal/schedulers/` - 2 个调度器文件
- `tests/unit/services/` - 4 个测试文件

**验收**:
```bash
# 运行测试
go test ./internal/business/services/site/... -v

# 测试调度器
go test ./internal/schedulers/... -v
```

---

### Day 5: API 和中间件 (2025-12-06)

**目标**: 实现所有 API 接口和中间件

**任务**:
- 实现认证 API Handler（6 个接口）
- 实现站点管理 API Handler（8 个接口）
- 实现认证中间件
- 实现权限中间件
- 编写集成测试
- 更新 Swagger 文档

**交付物**:
- `internal/apis/handlers/auth/` - 1 个 Handler 文件
- `internal/apis/handlers/site/` - 1 个 Handler 文件
- `internal/apis/middlewares/` - 2 个中间件文件
- `tests/integration/` - 2 个集成测试文件

**验收**:
```bash
# 运行集成测试
go test ./tests/integration/... -v

# 启动服务测试
make run

# 测试 API
curl -X POST http://localhost:3001/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","email":"test@example.com","password":"Test123456"}'
```

---

## 🔧 开发环境设置

### 1. 数据库准备

```bash
# 创建数据库
createdb moviepilot

# 或使用 Docker
docker run -d \
  --name moviepilot-postgres \
  -e POSTGRES_DB=moviepilot \
  -e POSTGRES_USER=moviepilot \
  -e POSTGRES_PASSWORD=moviepilot \
  -p 5432:5432 \
  postgres:14
```

### 2. 环境变量配置

创建 `.env` 文件：

```env
# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=moviepilot
DB_PASSWORD=moviepilot
DB_NAME=moviepilot

# JWT 配置
JWT_SECRET=your-secret-key-change-in-production
JWT_ACCESS_TOKEN_EXPIRE=900
JWT_REFRESH_TOKEN_EXPIRE=604800

# 服务器配置
SERVER_PORT=3001
SERVER_ENV=development
```

### 3. 安装依赖

```bash
# 安装 Go 依赖
go mod download

# 安装迁移工具
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# 安装 Cron 调度器
go get github.com/robfig/cron/v3
```

---

## 📊 进度跟踪

### 每日检查清单

**Day 1**:
- [ ] 迁移脚本创建完成
- [ ] 迁移可以成功执行
- [ ] 初始数据正确插入
- [ ] 可以回滚迁移

**Day 2**:
- [ ] 所有模型定义完成
- [ ] Repository 接口定义完成
- [ ] Repository 实现完成
- [ ] 单元测试通过

**Day 3**:
- [ ] JWT Manager 实现完成
- [ ] Password Manager 实现完成
- [ ] AuthService 实现完成
- [ ] 单元测试覆盖率 > 80%

**Day 4**:
- [ ] SiteService 实现完成
- [ ] CookieService 实现完成
- [ ] CheckinService 实现完成
- [ ] 调度器正常工作

**Day 5**:
- [ ] 所有 API 实现完成
- [ ] 中间件实现完成
- [ ] 集成测试通过
- [ ] Swagger 文档更新

---

## 🎯 成功标准

### 功能标准

1. **用户认证**
   - ✅ 用户可以注册（用户名、邮箱、密码）
   - ✅ 用户可以登录（返回 JWT Token）
   - ✅ 用户可以登出
   - ✅ Token 可以刷新
   - ✅ 用户可以修改密码
   - ✅ 权限控制正常工作

2. **站点管理**
   - ✅ 用户可以添加站点
   - ✅ 用户可以查看站点列表
   - ✅ 用户可以编辑站点
   - ✅ 用户可以删除站点
   - ✅ Cookie 可以自动同步（每小时）
   - ✅ 签到可以定时执行（每天）

### 质量标准

1. **代码质量**
   - ✅ 代码编译无错误
   - ✅ 代码符合 Go 规范
   - ✅ 代码有适当的注释
   - ✅ 错误处理完善

2. **测试覆盖**
   - ✅ 单元测试覆盖率 > 80%
   - ✅ 集成测试通过
   - ✅ 边界条件测试

3. **性能指标**
   - ✅ 登录响应时间 < 500ms
   - ✅ API 响应时间 < 1s
   - ✅ 支持 100 并发用户

---

## 🚨 风险和应对

### 潜在风险

1. **数据库迁移失败**
   - 风险：迁移脚本有错误
   - 应对：先在测试环境验证，保留回滚脚本

2. **JWT 安全问题**
   - 风险：密钥泄露、Token 被盗用
   - 应对：使用环境变量存储密钥，设置合理的过期时间

3. **Cookie 同步失败**
   - 风险：站点反爬虫、Cookie 格式变化
   - 应对：添加重试机制，记录详细日志

4. **性能瓶颈**
   - 风险：并发访问导致数据库压力
   - 应对：使用连接池，添加缓存

---

## 📝 开发规范

### 代码规范

1. **命名规范**
   - 文件名：小写蛇形（`user_service.go`）
   - 包名：小写单数（`auth`, `site`）
   - 接口：名词或动词+er（`UserRepository`, `Validator`）
   - 函数：驼峰命名（`GetUserByID`）

2. **错误处理**
   - 使用 `fmt.Errorf` 包装错误
   - 记录错误日志
   - 返回有意义的错误信息

3. **日志规范**
   - 使用 `pkg/logger`
   - 记录关键操作
   - 不记录敏感信息

### Git 提交规范

```
feat: 添加用户注册功能
fix: 修复 JWT Token 验证错误
docs: 更新 API 文档
test: 添加 AuthService 单元测试
refactor: 重构 Repository 层
```

---

## 📚 参考资源

### 内部文档

- `docs/design/auth-system-design.md` - 认证系统设计
- `docs/design/site-management-design.md` - 站点管理设计
- `docs/week7-implementation-plan.md` - 实施计划

### 外部资源

- [GORM 文档](https://gorm.io/docs/)
- [Gin 文档](https://gin-gonic.com/docs/)
- [JWT 最佳实践](https://tools.ietf.org/html/rfc7519)
- [bcrypt 使用指南](https://pkg.go.dev/golang.org/x/crypto/bcrypt)

---

## 🎉 Let's Go!

Week 7 是 Phase 2 的关键一周，让我们开始实施吧！

**下一步**: 开始 Day 1 - 创建数据库迁移脚本

---

**文档创建时间**: 2025-12-02  
**状态**: ✅ 准备就绪
