# API架构整理总结

## 整理前的问题

### 1. 重复的路由定义
- `routes/route.go` 包含了所有路由的集中定义
- `routes/` 目录下还有单独的路由文件：
  - `plugin_manager_route.go`
  - `search_route.go` 
  - `servarr_route.go`
  - `site_route.go`
  - `user_route.go`
- `handlers/history/routes.go` 重复定义了历史记录路由

### 2. 不一致的路由注册方式
- 有些使用集中式注册（`route.go`）
- 有些使用分散式注册（各个独立的route文件）

### 3. schemas目录内容不当
- `schemas/servcookie.go` 包含的是业务模型，应该放在 `internal/model/` 中

### 4. handlers目录结构混乱
- 有一个 `handlers/base_handler.go` 基础处理器
- 又有各个子目录的处理器（auth/, dashboard/, discover/等）

## 整理后的架构

### 目录结构
```
internal/api/
├── handlers/                 # HTTP处理器
│   ├── base_handler.go      # 基础处理器（提供通用响应方法）
│   ├── auth/                # 认证相关处理器
│   ├── dashboard/           # 仪表板处理器
│   ├── discover/            # 发现内容处理器
│   ├── download/            # 下载管理处理器
│   ├── file/                # 文件管理处理器
│   ├── history/             # 历史记录处理器
│   ├── media/               # 媒体管理处理器
│   ├── message/             # 消息管理处理器
│   ├── plugin/              # 插件管理处理器
│   ├── search/              # 搜索处理器
│   └── servarr/             # ServArr集成处理器
├── middleware/              # 中间件
│   ├── api_key.go          # API Key中间件
│   ├── auth.go             # 认证中间件
│   ├── cors.go             # CORS中间件
│   ├── global.go           # 全局中间件
│   └── logger.go           # 日志中间件
├── routes/                  # 路由定义
│   └── route.go            # 统一路由配置
└── validator/               # 请求验证器
    └── validator.go        # 验证器实现
```

### 关键改进

#### 1. 移除重复的路由文件
删除了以下重复的路由文件：
- `routes/plugin_manager_route.go`
- `routes/search_route.go`
- `routes/servarr_route.go`
- `routes/site_route.go`
- `routes/user_route.go`
- `handlers/history/routes.go`

#### 2. 统一路由管理
- 所有路由现在集中在 `routes/route.go` 中管理
- 使用 `RouterConfig` 结构体统一管理所有处理器依赖
- 按功能模块组织路由设置函数

#### 3. 正确的模型位置
- 将 `schemas/servcookie.go` 移动到 `internal/model/cookie/cookie.go`
- 删除了空的 `schemas/` 目录

#### 4. 清理的处理器架构
- `base_handler.go` 只提供通用的响应方法和辅助函数
- 各个专门的处理器负责具体的业务逻辑
- 路由设置函数使用正确的处理器类型

### RouterConfig结构
```go
type RouterConfig struct {
    BaseHandler      *handlers.BaseHandler
    AuthHandler      *auth.AuthHandler
    DashboardHandler *dashboard.DashboardHandler
    DiscoverHandler  *discover.DiscoverHandler
    DownloadHandler  *download.DownloadHandler
    FileHandler      *file.FileHandler
    HistoryHandler   *history.HistoryHandler
    MediaHandler     *media.MediaHandler
    MessageHandler   *message.MessageHandler
    PluginHandler    *plugin.PluginHandler
    SearchHandler    *search.SearchHandler
    ServarrHandler   *servarr.ServArrHandler
}
```

### 路由组织
- **公开路由**：认证、系统信息、发现内容
- **受保护路由**：需要认证中间件的所有业务功能
- **模块化路由设置**：每个功能模块有独立的路由设置函数

## 优势

1. **消除重复**：移除了重复的路由定义和处理器
2. **统一管理**：所有路由集中管理，便于维护
3. **清晰分层**：handlers、middleware、routes职责明确
4. **易于扩展**：新功能只需添加对应的处理器和路由设置函数
5. **类型安全**：使用正确的处理器类型，避免编译错误

## 使用方式

```go
// 创建路由配置
config := &routes.RouterConfig{
    BaseHandler:      handlers.NewBaseHandler(),
    AuthHandler:      auth.NewAuthHandler(authService),
    MediaHandler:     media.NewMediaHandler(mediaService),
    // ... 其他处理器
}

// 设置路由
router := routes.SetupRouter(config)
```

这个整理后的架构符合Go项目的最佳实践，提供了清晰的分层结构和统一的依赖管理。