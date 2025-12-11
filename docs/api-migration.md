# api/ 模块迁移计划

> Python: `app/api/`  \
> Go: `internal/apis/handlers/` + `internal/apis/routes/`

---

## 1. Python `app/api` 概览

- **职责**：
  - FastAPI 路由注册
  - Endpoint 实现（业务入口）
  - 请求体验证、响应序列化
  - 认证、权限控制（依赖全局中间件）
- **典型结构**（示意）：
  - `app/api/deps.py`（依赖注入）
  - `app/api/routes/user.py`
  - `app/api/routes/subscribe.py`
  - `app/api/routes/site.py`
  - `app/api/routes/system.py`

---

## 2. Go 目标结构

### 2.1 目录结构

```text
internal/apis/
├── handlers/              # 每个领域一个子目录
│   ├── base/             # 基础handler
│   ├── user/             # 用户相关接口
│   ├── subscribe/        # 订阅相关接口
│   ├── site/             # 站点相关接口
│   ├── system/           # 系统相关接口
│   ├── search/           # 搜索相关接口
│   ├── download/         # 下载相关接口
│   ├── media/            # 媒体相关接口
│   ├── plugin/           # 插件相关接口
│   ├── transfer/         # 转移相关接口
│   ├── dashboard/        # 仪表板相关接口
│   ├── login/            # 登录相关接口
│   ├── bangumi/          # 番剧相关接口
│   ├── douban/           # 豆瓣相关接口
│   ├── tmdb/             # TMDB相关接口
│   ├── workflow/         # 工作流相关接口
│   ├── scraper/          # 刮削相关接口
│   ├── performance/      # 性能相关接口
│   ├── notification/     # 通知相关接口
│   ├── message/          # 消息相关接口
│   ├── mediaserver/      # 媒体服务器相关接口
│   ├── history/          # 历史记录相关接口
│   └── pluginmedia/      # 插件媒体相关接口
├── routes/              # 统一路由注册
│   └── routes.go         # 路由注册文件
├── middleware/          # 认证、日志等横切逻辑
│   ├── auth.go          # 认证中间件
│   └── logging.go       # 日志中间件
└── common/              # 通用组件
    ├── error.go         # 错误处理
    ├── response.go      # 统一响应
    └── validator.go     # 验证器
```

### 2.2 Handler 实现模式

每个 Handler 都遵循以下模式：

```go
// handlers/user/handler.go
package user

import (
    "net/http"

    "github.com/gin-gonic/gin"

    "moviepilot-go/internal/business/services/user"
    "moviepilot-go/internal/models/dto"
    "moviepilot-go/internal/apis/common"
    "moviepilot-go/pkg/logger"
)

type Handler struct {
    userService user.Service
}

func NewHandler(userService user.Service) *Handler {
    return &Handler{
        userService: userService,
    }
}

// GetUserByID 获取用户信息
// @Summary 获取用户信息
// @Description 根据用户ID获取用户信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} common.Response{data=dto.UserResponse}
// @Failure 400 {object} common.Response
// @Failure 404 {object} common.Response
// @Failure 500 {object} common.Response
// @Router /api/v1/users/{id} [get]
func (h *Handler) GetUserByID(c *gin.Context) {
    // 参数绑定与验证
    var req dto.GetUserRequest
    if err := c.ShouldBindUri(&req); err != nil {
        logger.Error("Invalid request parameters", zap.Error(err))
        common.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters")
        return
    }

    // 调用Service
    user, err := h.userService.GetUserByID(c.Request.Context(), req.ID)
    if err != nil {
        logger.Error("Failed to get user", zap.Error(err))
        common.ErrorResponse(c, http.StatusInternalServerError, "Failed to get user")
        return
    }

    // 构造响应
    resp := dto.UserResponse{
        ID:   user.ID,
        Name: user.Name,
        // 其他字段...
    }

    common.SuccessResponse(c, resp)
}
```

---

## 3. 路由与 Handler 映射

| Python 路由 | Go Handler 目标 | 状态 | 备注 |
|-------------|-----------------|------|------|
| `routes/user.py` | `internal/apis/handlers/user/handler.go` | ✅ 已实现 | 用户相关接口 |
| `routes/subscribe.py` | `internal/apis/handlers/subscribe/handler.go` | ✅ 已实现 | 订阅管理 |
| `routes/site.py` | `internal/apis/handlers/site/handler.go` | ✅ 已实现 | 站点管理/签到/RSS |
| `routes/system.py` | `internal/apis/handlers/system/handler.go` | ✅ 已实现 | 系统配置、状态 |
| `routes/search.py` | `internal/apis/handlers/search/handler.go` | ✅ 已实现 | 搜索接口 |
| `routes/download.py` | `internal/apis/handlers/download/` | ✅ 已实现 | 下载管理 |
| `routes/media.py` | `internal/apis/handlers/media/handler.go` | ✅ 已实现 | 媒体库相关 |
| `routes/plugin.py` | `internal/apis/handlers/plugin/handler.go` | ✅ 已实现 | 插件管理 |
| `routes/transfer.py` | `internal/apis/handlers/transfer/handler.go` | ✅ 已实现 | 文件转移 |
| `routes/dashboard.py` | `internal/apis/handlers/dashboard/handler.go` | ✅ 已实现 | 仪表板 |
| `routes/login.py` | `internal/apis/handlers/login/handler.go` | ✅ 已实现 | 登录认证 |
| `routes/bangumi.py` | `internal/apis/handlers/bangumi/handler.go` | ✅ 已实现 | 番剧相关 |
| `routes/douban.py` | `internal/apis/handlers/douban/handler.go` | ✅ 已实现 | 豆瓣相关 |
| `routes/tmdb.py` | `internal/apis/handlers/tmdb/handler.go` | ✅ 已实现 | TMDB相关 |
| `routes/workflow.py` | `internal/apis/handlers/workflow/handler.go` | ✅ 已实现 | 工作流相关 |
| `routes/scraper.py` | `internal/apis/handlers/scraper/handler.go` | ✅ 已实现 | 刮削相关 |
| `routes/performance.py` | `internal/apis/handlers/performance/handler.go` | ✅ 已实现 | 性能相关 |
| `routes/notification.py` | `internal/apis/handlers/notification/handler.go` | ✅ 已实现 | 通知相关 |
| `routes/message.py` | `internal/apis/handlers/message/handler.go` | ✅ 已实现 | 消息相关 |
| `routes/mediaserver.py` | `internal/apis/handlers/mediaserver/handler.go` | ✅ 已实现 | 媒体服务器相关 |
| `routes/history.py` | `internal/apis/handlers/history/handler.go` | ✅ 已实现 | 历史记录相关 |
| `routes/pluginmedia.py` | `internal/apis/handlers/pluginmedia/handler.go` | ✅ 已实现 | 插件媒体相关 |

---

## 4. 输入/输出模型迁移

- **Python**：Pydantic 模型位于 `app/schemas/`，通过 FastAPI 自动绑定。
- **Go**：
  - 请求体：`internal/models/dto/` 中的 DTO 结构体，配合 Gin `binding` 标签验证。
  - 响应体：复用 DTO 或定义专用 Response 结构体，通过 `c.JSON` 返回。

**约定：**
- API 层只做：
  - 参数绑定 + 基础校验
  - 调用 Service
  - 统一响应（成功/错误包装）
- 不在 Handler 中写业务逻辑，不直接访问数据库。

---

## 5. 中间件与依赖注入

- **Python**：依赖 `deps.py` / FastAPI Depends 实现认证、DB Session 等注入。
- **Go**：
  - 认证、日志、限流等放在 `internal/apis/middleware/`。
  - 业务依赖（Service、Repository）通过：
    - 在 `routes.Register` 中组装依赖并注入到 Handler 结构体。

**依赖注入示例：**

```go
// routes/routes.go
func RegisterRoutes(r *gin.Engine, di *dependency.Dependency) {
    // 初始化各领域 Service
    userService := user.NewService(di.UserRepository)
    subscribeService := subscribe.NewService(di.SubscribeRepository)
    
    // 初始化各领域 Handler
    userHandler := user.NewHandler(userService)
    subscribeHandler := subscribe.NewHandler(subscribeService)
    
    // 注册路由
    api := r.Group("/api/v1")
    {
        // 用户路由
        users := api.Group("/users")
        {
            users.GET("/:id", userHandler.GetUserByID)
            users.POST("/", userHandler.CreateUser)
            users.PUT("/:id", userHandler.UpdateUser)
            users.DELETE("/:id", userHandler.DeleteUser)
        }
        
        // 订阅路由
        subscribes := api.Group("/subscribes")
        {
            subscribes.GET("/", subscribeHandler.ListSubscribes)
            subscribes.POST("/", subscribeHandler.CreateSubscribe)
            subscribes.PUT("/:id", subscribeHandler.UpdateSubscribe)
            subscribes.DELETE("/:id", subscribeHandler.DeleteSubscribe)
        }
    }
}
```

---

## 6. 迁移步骤

### 6.1 已完成步骤

1. **目录结构搭建**：创建了 `internal/apis/handlers/`、`internal/apis/routes/`、`internal/apis/middleware/` 和 `internal/apis/common/` 目录
2. **通用组件实现**：实现了统一响应、错误处理、验证器等通用组件
3. **中间件实现**：实现了认证、日志等中间件
4. **Handler 实现**：实现了所有核心领域的 Handler
5. **路由注册**：实现了统一路由注册
6. **DTO 定义**：为所有 API 定义了对应的 DTO 结构体
7. **依赖注入**：实现了基于构造函数的依赖注入

### 6.2 后续优化步骤

1. **完善 Swagger 文档**：为所有 API 添加详细的 Swagger 注解
2. **完善单元测试**：为所有 Handler 编写单元测试
3. **优化中间件**：增强中间件的功能和性能
4. **API 版本管理**：实现 API 版本管理
5. **限流与熔断**：添加限流和熔断机制
6. **监控与追踪**：添加 API 监控和分布式追踪

---

## 7. 检查清单

- [x] 所有 Python 路由在表格中都有对应的 Go Handler 条目。
- [x] Handler 不包含业务逻辑，只负责输入输出。
- [x] 所有 Handler 使用 `pkg/logger` 记录关键日志。
- [x] 所有需要认证的接口都经过统一中间件处理。
- [x] 所有 API 都使用统一的响应格式。
- [x] 所有 API 都使用 DTO 进行参数绑定和验证。
- [x] 所有 Handler 都通过依赖注入获取所需服务。
- [x] 关键 API 已补充单元测试与 Swagger 文档。

---

## 8. 迁移计划归纳

### 8.1 迁移目标

将 Python 项目的 `app/api/` 目录迁移到 Go 项目的 `internal/apis/handlers/` + `internal/apis/routes/` 目录，实现从 FastAPI 到 Gin 框架的迁移，包括：

- API 路由迁移
- 请求处理迁移
- 响应处理迁移
- 中间件迁移
- 依赖注入迁移

### 8.2 迁移策略

1. **目录结构先行**：先搭建好目标目录结构
2. **通用组件优先**：先实现统一响应、错误处理、验证器等通用组件
3. **中间件实现**：实现认证、日志等中间件
4. **Handler 迁移**：逐个迁移 Python 路由到 Go Handler
5. **路由注册**：统一注册所有路由
6. **依赖注入**：实现基于构造函数的依赖注入
7. **测试与优化**：编写单元测试并优化性能

### 8.3 迁移成果

1. **完整的 API 层**：已实现所有核心领域的 API，包括用户、订阅、站点、系统、搜索、下载、媒体、插件等
2. **清晰的目录结构**：按领域划分 Handler，便于维护和扩展
3. **统一的响应格式**：所有 API 使用统一的响应格式
4. **完善的中间件**：实现了认证、日志等中间件
5. **依赖注入**：通过依赖注入提高了可测试性和可扩展性
6. **符合 RESTful 规范**：API 设计符合 RESTful 规范

### 8.4 后续优化方向

1. **API 文档**：完善 Swagger 文档
2. **测试覆盖**：提高单元测试覆盖率
3. **性能优化**：优化 API 响应性能
4. **安全性增强**：增强 API 的安全性
5. **监控与追踪**：添加 API 监控和分布式追踪
6. **版本管理**：实现 API 版本管理

---

## 9. 结论

API 模块迁移已基本完成，所有核心领域的 API 都已实现，包括用户、订阅、站点、系统、搜索、下载、媒体、插件等。迁移后的 API 层结构清晰，符合 Go 语言的最佳实践，使用了依赖注入、统一响应、中间件等设计模式，提高了代码的可维护性和可扩展性。

后续将继续完善 API 文档、单元测试和性能优化，确保 API 层的高质量和高性能。

---

**相关文档**：
- [chain-migration.md](./chain-migration.md)
- [db-migration.md](./db-migration.md)
- [workflow-migration-plan.md](./workflow-migration-plan.md)
