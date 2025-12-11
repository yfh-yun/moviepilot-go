# Swagger 文档配置指南

> **更新时间**: 2025-12-02  
> **适用版本**: MoviePilot Go v2.8.1

---

## 📋 目录

1. [快速开始](#快速开始)
2. [Swagger 注解规范](#swagger-注解规范)
3. [生成文档](#生成文档)
4. [常见问题](#常见问题)

---

## 快速开始

### 1. 安装 Swag 工具

```bash
make install-tools
```

或手动安装：

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

### 2. 生成 Swagger 文档

```bash
make swagger
```

### 3. 启动服务并访问文档

```bash
make run
```

访问：`http://localhost:3001/swagger/index.html`

---

## Swagger 注解规范

### 主应用注解（cmd/server/main.go）

```go
// @title MoviePilot API
// @version 2.8.1
// @description MoviePilot Go version - Automated media library management tool
// @termsOfService http://swagger.io/terms/

// @contact.name MoviePilot Team
// @contact.url http://www.moviepilot.com
// @contact.email support@moviepilot.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:3001
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
    // ...
}
```

### API 处理器注解

#### 基础模板

```go
// HandlerName 处理器描述
// @Summary 简短描述
// @Description 详细描述
// @Tags 标签名称
// @Accept json
// @Produce json
// @Param paramName paramType dataType required "参数描述"
// @Success 200 {object} ResponseType
// @Failure 400 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/path [method]
func (h *Handler) HandlerName(c *gin.Context) {
    // ...
}
```

#### 参数类型说明

| paramType | 说明 | 示例 |
|-----------|------|------|
| `path` | 路径参数 | `/api/users/{id}` |
| `query` | 查询参数 | `/api/users?name=xxx` |
| `body` | 请求体 | JSON 数据 |
| `header` | 请求头 | `Authorization` |
| `formData` | 表单数据 | 文件上传 |

#### 数据类型

| dataType | 说明 |
|----------|------|
| `string` | 字符串 |
| `int` | 整数 |
| `bool` | 布尔值 |
| `object` | 对象（需指定类型） |
| `array` | 数组（需指定元素类型） |
| `file` | 文件 |

### 示例：用户登录 API

```go
// Login 用户登录
// @Summary 用户登录
// @Description 用户登录获取访问令牌
// @Tags user
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "登录凭证"
// @Success 200 {object} userbiz.AuthToken
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "用户名或密码错误"
// @Router /api/user/login [post]
func (h *Handler) Login(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    // ...
}

// LoginRequest 登录请求
type LoginRequest struct {
    Username string `json:"username" binding:"required" example:"admin"`
    Password string `json:"password" binding:"required" example:"password123"`
}
```

### 示例：列表查询 API

```go
// ListSubscribes 列出订阅
// @Summary 列出订阅
// @Description 分页列出所有订阅
// @Tags 订阅
// @Security BearerAuth
// @Produce json
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Param type query string false "类型过滤" Enums(movie, tv)
// @Param state query string false "状态过滤" Enums(active, paused, completed)
// @Success 200 {object} ListSubscribesResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/subscribes [get]
func (h *Handler) ListSubscribes(c *gin.Context) {
    // ...
}

// ListSubscribesResponse 订阅列表响应
type ListSubscribesResponse struct {
    Items      []database.Subscribe `json:"items"`
    Total      int                  `json:"total"`
    Page       int                  `json:"page"`
    Limit      int                  `json:"limit"`
    TotalPages int                  `json:"total_pages"`
}
```

### 示例：文件上传 API

```go
// UploadPoster 上传海报
// @Summary 上传海报
// @Description 上传媒体海报图片
// @Tags media
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "海报文件"
// @Param media_id formData int true "媒体ID"
// @Success 200 {object} map[string]interface{} "上传成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Router /api/media/poster [post]
func (h *Handler) UploadPoster(c *gin.Context) {
    // ...
}
```

### 示例：需要认证的 API

```go
// GetUserProfile 获取用户资料
// @Summary 获取用户资料
// @Description 获取当前登录用户的资料信息
// @Tags user
// @Security BearerAuth
// @Produce json
// @Success 200 {object} database.User
// @Failure 401 {object} map[string]interface{} "未认证"
// @Router /api/user/profile [get]
func (h *Handler) GetUserProfile(c *gin.Context) {
    // ...
}
```

---

## 生成文档

### 方式一：使用 Makefile

```bash
make swagger
```

### 方式二：直接使用 swag 命令

```bash
swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal
```

### 参数说明

- `-g`: 指定包含主注解的 Go 文件
- `-o`: 输出目录
- `--parseDependency`: 解析依赖包中的注解
- `--parseInternal`: 解析 internal 包中的注解

### 生成的文件

```
docs/
├── docs.go          # Go 代码
├── swagger.json     # JSON 格式文档
└── swagger.yaml     # YAML 格式文档
```

---

## 常见问题

### 1. 找不到类型定义

**错误**:
```
cannot find type definition: dto.MediaInfo
```

**解决方案**:
- 确保类型定义在同一包或已导入的包中
- 使用完整的包路径：`moviepilot-go/internal/models.User`
- 或使用 `map[string]interface{}` 作为通用响应类型

### 2. 注解格式错误

**错误**:
```
ParseComment error: invalid format
```

**解决方案**:
- 检查注解格式是否正确
- 确保每个注解独占一行
- 参数之间用空格分隔，不要用逗号

### 3. 数组类型注解

**正确写法**:
```go
// @Success 200 {array} database.Subscribe
```

**错误写法**:
```go
// @Success 200 {object} []database.Subscribe  // ❌
```

### 4. 枚举值注解

**正确写法**:
```go
// @Param type query string false "类型" Enums(movie, tv)
```

### 5. 示例值注解

在结构体中添加 `example` 标签：

```go
type LoginRequest struct {
    Username string `json:"username" example:"admin"`
    Password string `json:"password" example:"password123"`
}
```

### 6. 响应示例

使用 `@Success` 和 `@Failure` 注解：

```go
// @Success 200 {object} database.User "成功返回用户信息"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未认证"
// @Failure 404 {object} map[string]interface{} "用户不存在"
```

---

## 最佳实践

### 1. 统一响应格式

定义统一的响应结构：

```go
// APIResponse 统一API响应
type APIResponse struct {
    Code    int         `json:"code" example:"200"`
    Message string      `json:"message" example:"success"`
    Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
    Error   string                 `json:"error" example:"invalid_request"`
    Message string                 `json:"message" example:"请求参数错误"`
    Details map[string]interface{} `json:"details,omitempty"`
}
```

### 2. 分组管理

使用 `@Tags` 对 API 进行分组：

```go
// @Tags user       // 用户相关
// @Tags subscribe  // 订阅相关
// @Tags download   // 下载相关
// @Tags system     // 系统相关
```

### 3. 安全定义

在需要认证的 API 上添加：

```go
// @Security BearerAuth
```

### 4. 版本管理

在路由中体现版本：

```go
// @Router /api/v1/users [get]
// @Router /api/v2/users [get]
```

### 5. 文档注释

为每个 API 添加详细的描述：

```go
// @Summary 简短描述（一句话）
// @Description 详细描述（可以多行）
// @Description 第二行描述
// @Description 第三行描述
```

---

## 集成到 CI/CD

### GitHub Actions 示例

```yaml
name: Generate Swagger Docs

on:
  push:
    branches: [ main, develop ]

jobs:
  swagger:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Install swag
        run: go install github.com/swaggo/swag/cmd/swag@latest
      
      - name: Generate Swagger docs
        run: make swagger
      
      - name: Commit docs
        run: |
          git config --local user.email "action@github.com"
          git config --local user.name "GitHub Action"
          git add docs/
          git commit -m "docs: update Swagger documentation" || exit 0
          git push
```

---

## 参考资源

- [Swag 官方文档](https://github.com/swaggo/swag)
- [Swagger 规范](https://swagger.io/specification/)
- [OpenAPI 3.0 规范](https://spec.openapis.org/oas/v3.0.0)
- [gin-swagger](https://github.com/swaggo/gin-swagger)

---

## 更新日志

### 2025-12-02

- ✅ 初始化 Swagger 配置
- ✅ 添加用户 API 注解示例
- ✅ 添加订阅 API 注解示例
- ✅ 配置 Swagger UI 路由
- ✅ 编写配置指南文档
