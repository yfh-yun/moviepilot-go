# API目录重构总结

## 重构目标
按照架构要求整理 `moviepilot-go/internal/api` 目录，移除多余文件，统一架构模式。

## 重构前问题

### 1. 重复和冗余的路由定义
- 存在多个独立路由文件：`plugin_manager_route.go`, `search_route.go`, `servarr_route.go`, `site_route.go`, `user_route.go`
- `handlers/history/routes.go` 重复定义历史记录路由
- 路由注册方式不统一

### 2. 不当的文件组织
- `schemas/servcookie.go` 包含业务模型，应放在 `internal/model/`
- handlers目录结构混乱，基础处理器与业务处理器混合

### 3. 缺失的路由配置
- 许多处理器目录没有对应的路由配置
- RouterConfig结构体不完整

## 重构执行

### 1. 清理多余文件
```bash
# 删除重复的路由文件
rm -f routes/plugin_manager_route.go
rm -f routes/search_route.go  
rm -f routes/servarr_route.go
rm -f routes/site_route.go
rm -f routes/user_route.go
rm -f handlers/history/routes.go

# 移动模型文件到正确位置
mv internal/api/schemas/servcookie.go internal/model/cookie/cookie.go
rmdir internal/api/schemas
```

### 2. 统一路由管理
- 所有路由集中在 `routes/route.go` 管理
- 创建完整的 `RouterConfig` 结构体，包含所有处理器
- 按功能模块组织路由设置函数

### 3. 完善处理器架构
- `base_handler.go` 只提供通用响应方法和辅助函数
- 各专门处理器负责具体业务逻辑
- 路由设置函数使用正确的处理器类型

## 重构后架构

### 目录结构
```
internal/api/
├── handlers/                 # HTTP处理器 (23个子模块)
│   ├── base_handler.go      # 基础处理器
│   ├── auth/                # 认证
│   ├── dashboard/           # 仪表板
│   ├── discover/            # 发现内容
│   ├── douban/              # 豆瓣集成
│   ├── download/            # 下载管理
│   ├── file/                # 文件管理
│   ├── history/             # 历史记录
│   ├── media/               # 媒体管理
│   ├── mediaserver/         # 媒体服务器
│   ├── message/             # 消息管理
│   ├── notification/        # 通知管理
│   ├── plugin/              # 插件管理
│   ├── recommend/           # 推荐系统
│   ├── scheduler/           # 任务调度
│   ├── search/              # 搜索功能
│   ├── servarr/             # ServArr集成
│   ├── settings/            # 系统设置
│   ├── site/                # 站点管理
│   ├── storage/             # 存储管理
│   ├── subscribe/           # 订阅功能
│   ├── subscription/        # 订阅管理
│   ├── system/              # 系统管理
│   ├── torrent/             # 种子管理
│   ├── transfer/            # 传输管理
│   ├── user/                # 用户管理
│   ├── webhook/             # Webhook
│   └── workflow/            # 工作流
├── middleware/              # 中间件 (5个)
│   ├── api_key.go
│   ├── auth.go
│   ├── cors.go
│   ├── global.go
│   └── logger.go
├── routes/                  # 路由定义
│   └── route.go            # 统一路由配置
└── validator/               # 请求验证器
    └── validator.go
```

### RouterConfig结构
包含24个处理器，完整覆盖所有功能模块：
- 基础系统：BaseHandler, SystemHandler, SettingsHandler
- 用户认证：AuthHandler, UserHandler
- 媒体相关：MediaHandler, MediaServerHandler, DoubanHandler
- 下载传输：DownloadHandler, TorrentHandler, TransferHandler
- 文件存储：FileHandler, StorageHandler
- 订阅站点：SubscribeHandler, SubscriptionHandler, SiteHandler
- 搜索推荐：SearchHandler, RecommendHandler, DiscoverHandler
- 系统功能：HistoryHandler, MessageHandler, NotificationHandler
- 集成扩展：PluginHandler, ServarrHandler, WebhookHandler
- 管理界面：DashboardHandler, SchedulerHandler, WorkflowHandler

### 路由组织
- **公开路由**：认证、系统信息、发现内容、豆瓣
- **受保护路由**：需要认证的所有业务功能
- **模块化设置**：每个功能模块有独立的路由设置函数

## 重构收益

### 1. 消除重复
- 移除了6个重复的路由文件
- 统一了路由注册方式
- 避免了代码重复和维护成本

### 2. 架构清晰
- 分层明确：handlers → middleware → routes
- 职责单一：每个处理器负责特定功能
- 依赖清晰：RouterConfig统一管理依赖注入

### 3. 易于维护
- 集中式路由管理，便于查找和修改
- 模块化设计，便于功能扩展
- 类型安全，编译时检查错误

### 4. 符合规范
- 遵循Go项目标准布局
- 符合领域驱动设计原则
- 满足项目架构要求

## 使用示例

```go
// 创建所有处理器
config := &routes.RouterConfig{
    BaseHandler:        handlers.NewBaseHandler(),
    AuthHandler:        auth.NewAuthHandler(authService),
    MediaHandler:       media.NewMediaHandler(mediaService),
    // ... 其他22个处理器
}

// 设置路由
router := routes.SetupRouter(config)
```

## 验证结果

- ✅ 删除了所有多余文件
- ✅ 统一了路由管理方式
- ✅ 完整了处理器架构
- ✅ 通过了lint检查
- ✅ 保持了功能完整性

重构后的API目录结构清晰、职责明确、易于维护，完全符合项目的架构要求。