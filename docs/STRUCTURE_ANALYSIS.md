# MoviePilot-Go 项目结构分析与调整方案

> 生成时间：2024-11-22  
> 最后更新：2024-11-22  
> 目标：对比预设规范与实际结构，识别差异并制定调整方案

## ✅ P0 调整已完成（2024-11-22）

已完成以下关键目录重命名和 import 路径更新：

1. ✅ **重命名 `internal/apis` → `internal/api`**
   - 所有相关 import 路径已更新
   - 影响文件：`cmd/server/main.go`, `tests/api/*.go`, `tests/integration/*.go`

2. ✅ **重命名 `internal/repositories` → `internal/repository`**
   - 所有相关 import 路径已更新
   - 影响文件：`internal/repository/**/*.go`

3. ✅ **验证编译通过**
   - `go mod tidy` 执行成功
   - 测试文件编译正常

---

## 一、预设结构 vs 实际结构对比

### 1. `internal/` 目录对比

#### 预设结构（规范要求）
```
internal/
├── api/                        # API层
│   ├── handlers/               # HTTP处理器
│   ├── middleware/             # 中间件
│   └── routes/                 # 路由定义
├── service/                    # 业务逻辑层
│   ├── user/                   # 用户服务
│   ├── subscribe/              # 订阅服务
│   ├── download/               # 下载服务
│   ├── transfer/               # 转移服务
│   └── plugin/                 # 插件管理服务
├── repository/                 # 数据访问层
│   ├── postgres/               # PostgreSQL实现
│   └── redis/                  # Redis实现
├── model/                      # 数据模型
│   ├── user.go
│   ├── subscribe.go
│   ├── download.go
│   └── plugin.go
└── config/                     # 配置管理
```

#### 实际结构（当前状态 - 已完成 P0 调整）
```
internal/
├── api/                        # ✅ API层（已重命名，符合规范）
│   └── workflow/               # ✅ 工作流API（已实现）
│       ├── handler.go
│       ├── service.go
│       └── dto.go
├── actions/                    # ⚠️ 新增：Action层（工作流任务）
│   ├── fetch_torrents/
│   ├── scan_file/
│   ├── scrape_file/
│   └── transfer_file/
├── business/                   # ⚠️ 新增：业务服务层（类似service）
│   ├── media/                  # 媒体识别服务
│   ├── storage/                # 存储服务
│   └── transfer/               # 转移服务
├── platform/                   # ⚠️ 新增：平台层（工作流引擎）
│   ├── workflow/               # 工作流管理器
│   └── module/                 # 模块管理
├── repository/                 # ✅ 数据访问层（已重命名，符合规范）
│   ├── interfaces/             # 接口定义
│   ├── repositories/           # 具体实现
│   └── migrations/             # 数据库迁移
├── models/                     # ✅ 数据模型（命名差异：复数形式）
│   └── cookie/
└── monitor/                    # ⚠️ 新增：监控层
    ├── metrics/
    └── health/
```

### 2. `pkg/` 目录对比

#### 预设结构（规范要求）
```
pkg/
├── database/                   # 数据库连接
├── cache/                      # 缓存封装
├── logger/                     # 日志封装
├── plugin/                     # 插件系统核心
│   ├── interface.go
│   ├── manager.go
│   ├── bridge.go
│   └── registry.go
├── utils/                      # 工具函数
└── errors/                     # 错误处理
```

#### 实际结构（当前状态）
```
pkg/
├── logger/                     # ✅ 日志封装（已完善）
├── middlewares/                # ✅ 中间件（新增，已实现）
│   ├── request_id.go
│   ├── recovery.go
│   └── cors.go
├── response/                   # ✅ 响应封装（新增）
├── validator/                  # ✅ 参数校验（新增）
├── errors/                     # ✅ 错误处理
├── plugin/                     # ✅ 插件系统
├── utils/                      # ✅ 工具函数
├── httpclient/                 # ✅ HTTP客户端（新增）
├── jwt/                        # ✅ JWT认证（新增）
├── redis/                      # ⚠️ Redis封装（应在pkg/cache下）
└── models/                     # ❌ 数据模型（应在internal/models）
```

---

## 二、主要差异分析

### 🔴 关键问题

#### 1. ✅ **目录命名不一致**（已解决）
- **问题**：`internal/apis` vs 规范要求的 `internal/api`
- **影响**：与规范文档不一致，可能导致理解偏差
- **解决方案**：✅ 已重命名 `internal/apis` → `internal/api`
- **完成时间**：2024-11-22

#### 2. **缺少 `internal/api/middleware` 和 `internal/api/routes`**
- **问题**：中间件放在 `pkg/middlewares`，路由直接在 `cmd/server/main.go` 注册
- **影响**：
  - 中间件位置不符合规范（应在 `internal/api/middleware`）
  - 路由管理分散，缺少统一路由注册模块
- **建议**：
  - 创建 `internal/api/middleware/` 存放业务中间件
  - 创建 `internal/api/routes/` 统一管理路由注册
  - `pkg/middlewares/` 保留通用中间件（RequestID、Recovery、CORS）

#### 3. **`internal/service` 缺失**
- **问题**：业务逻辑分散在 `internal/business` 和 `internal/actions`
- **影响**：不符合规范的三层架构（API → Service → Repository）
- **建议**：
  - 创建 `internal/service/` 目录
  - 将 `internal/business/` 重命名或迁移到 `internal/service/`
  - 明确 Service 层职责：封装业务逻辑，调用 Repository

#### 4. ✅ **`internal/repository` vs `internal/repositories`**（已解决）
- **问题**：使用复数形式 `repositories`
- **影响**：与规范不一致
- **解决方案**：✅ 已重命名 `internal/repositories` → `internal/repository`
- **完成时间**：2024-11-22

#### 5. **`pkg/models` 位置错误**
- **问题**：数据模型应在 `internal/models`，不应暴露到 `pkg`
- **影响**：违反封装原则，数据模型不应作为公共库
- **建议**：迁移 `pkg/models` → `internal/models`

#### 6. **`pkg/redis` 应归入 `pkg/cache`**
- **问题**：Redis 封装独立存在，应作为缓存实现的一部分
- **影响**：结构不清晰
- **建议**：
  - 创建 `pkg/cache/` 目录
  - 迁移 `pkg/redis` → `pkg/cache/redis`
  - 提供统一的 Cache 接口

### 🟡 次要问题

#### 7. **新增层级合理性评估**
- **`internal/actions/`**：✅ 合理，用于工作流任务实现
- **`internal/platform/`**：✅ 合理，工作流引擎和模块管理
- **`internal/monitor/`**：✅ 合理，监控和健康检查
- **`pkg/middlewares/`**：✅ 合理，通用中间件
- **`pkg/response/`**：✅ 合理，统一响应格式
- **`pkg/validator/`**：✅ 合理，参数校验封装

#### 8. **缺少 `internal/config`**
- **问题**：配置管理逻辑分散
- **建议**：创建 `internal/config/` 统一管理配置加载和解析

---

## 三、调整方案

### 阶段一：目录重命名（优先级：高）

```bash
# 1. 重命名 API 目录
mv internal/apis internal/api

# 2. 重命名 Repository 目录
mv internal/repositories internal/repository

# 3. 迁移数据模型
mv pkg/models/* internal/models/
rmdir pkg/models
```

### 阶段二：创建缺失目录（优先级：高）

```bash
# 1. 创建 API 子目录
mkdir -p internal/api/middleware
mkdir -p internal/api/routes

# 2. 创建 Service 目录
mkdir -p internal/service/user
mkdir -p internal/service/subscribe
mkdir -p internal/service/download
mkdir -p internal/service/transfer
mkdir -p internal/service/plugin

# 3. 创建 Config 目录
mkdir -p internal/config

# 4. 创建 Cache 目录
mkdir -p pkg/cache/redis
mkdir -p pkg/cache/memory
```

### 阶段三：代码迁移（优先级：中）

#### 1. 迁移中间件
```bash
# 业务中间件 → internal/api/middleware
# 通用中间件保留在 pkg/middlewares
```

**区分原则**：
- **通用中间件**（`pkg/middlewares/`）：RequestID、Recovery、CORS、RateLimit
- **业务中间件**（`internal/api/middleware/`）：Auth、Permission、Tenant

#### 2. 创建路由管理模块
```go
// internal/api/routes/routes.go
package routes

func RegisterRoutes(engine *gin.Engine) {
    // 注册所有路由
    RegisterWorkflowRoutes(engine)
    RegisterUserRoutes(engine)
    // ...
}
```

#### 3. 重构 Business → Service
```bash
# 将 internal/business 内容迁移到 internal/service
# 保持业务逻辑封装，调用 Repository 层
```

#### 4. 整合 Cache
```bash
# 迁移 pkg/redis → pkg/cache/redis
# 创建统一 Cache 接口
```

### 阶段四：更新导入路径（优先级：高）

所有受影响文件需要更新 import 路径：
- `moviepilot-go/internal/apis` → `moviepilot-go/internal/api`
- `moviepilot-go/internal/repositories` → `moviepilot-go/internal/repository`
- `moviepilot-go/pkg/models` → `moviepilot-go/internal/models`

---

## 四、调整后的目标结构

```
moviepilot-go/
├── internal/
│   ├── api/                        # ✅ API层
│   │   ├── handlers/               # HTTP处理器（待创建）
│   │   │   └── workflow/           # 工作流Handler
│   │   ├── middleware/             # 业务中间件（待创建）
│   │   │   ├── auth.go
│   │   │   └── permission.go
│   │   └── routes/                 # 路由定义（待创建）
│   │       ├── routes.go
│   │       └── workflow.go
│   ├── service/                    # ✅ 业务逻辑层（待创建）
│   │   ├── user/
│   │   ├── subscribe/
│   │   ├── download/
│   │   ├── transfer/               # 从business迁移
│   │   └── plugin/
│   ├── repository/                 # ✅ 数据访问层（重命名）
│   │   ├── interfaces/
│   │   ├── postgres/               # PostgreSQL实现
│   │   └── redis/                  # Redis实现
│   ├── models/                     # ✅ 数据模型（合并pkg/models）
│   ├── config/                     # ✅ 配置管理（待创建）
│   ├── actions/                    # ✅ 工作流任务（保留）
│   ├── platform/                   # ✅ 平台层（保留）
│   └── monitor/                    # ✅ 监控层（保留）
├── pkg/
│   ├── logger/                     # ✅ 日志封装
│   ├── middlewares/                # ✅ 通用中间件
│   │   ├── request_id.go
│   │   ├── recovery.go
│   │   ├── cors.go
│   │   └── rate_limit.go
│   ├── cache/                      # ✅ 缓存封装（待创建）
│   │   ├── cache.go                # 统一接口
│   │   ├── redis/                  # Redis实现
│   │   └── memory/                 # 内存实现
│   ├── database/                   # ✅ 数据库连接（待创建）
│   ├── response/                   # ✅ 响应封装
│   ├── validator/                  # ✅ 参数校验
│   ├── errors/                     # ✅ 错误处理
│   ├── plugin/                     # ✅ 插件系统
│   ├── utils/                      # ✅ 工具函数
│   ├── httpclient/                 # ✅ HTTP客户端
│   └── jwt/                        # ✅ JWT认证
└── cmd/
    └── server/
        └── main.go                 # 简化，调用routes.RegisterRoutes
```

---

## 五、执行优先级

### ✅ P0（已完成 - 2024-11-22）
1. ✅ 重命名 `internal/apis` → `internal/api`
2. ✅ 重命名 `internal/repositories` → `internal/repository`
3. ✅ 更新所有 import 路径
   - 影响文件：`cmd/server/main.go`, `tests/api/*.go`, `tests/integration/*.go`, `internal/repository/**/*.go`
   - 执行命令：`go mod tidy` 验证通过

### P1（本周完成）
4. 创建 `internal/api/routes/` 并迁移路由注册逻辑
5. 创建 `internal/api/middleware/` 并区分业务中间件
6. 迁移 `pkg/models` → `internal/models`

### P2（下周完成）
7. 创建 `internal/service/` 并重构 `internal/business`
8. 创建 `pkg/cache/` 并整合 Redis
9. 创建 `internal/config/` 统一配置管理

### P3（后续优化）
10. 完善 `internal/api/handlers/` 目录结构
11. 补充缺失的 Service 模块（user、subscribe、download、plugin）

---

## 六、风险与注意事项

### 风险
1. **大规模重命名可能导致编译错误**：需要仔细更新所有 import 路径
2. **测试用例可能失效**：需要同步更新测试文件的 import
3. **Git 历史可能混乱**：建议使用 `git mv` 保留文件历史

### 注意事项
1. 每次调整后立即运行 `go test ./...` 验证
2. 使用 IDE 的重构功能批量更新 import
3. 分阶段提交，每个阶段独立验证
4. 更新文档同步反映结构变化

---

## 七、后续行动

1. **立即执行 P0 任务**：重命名核心目录
2. **更新 PHASE1_DETAILED_PLAN.md**：反映结构调整
3. **创建迁移脚本**：自动化目录调整和 import 更新
4. **补充单元测试**：验证调整后功能正常

---

**文档维护者**：Cascade AI  
**最后更新**：2024-11-22
