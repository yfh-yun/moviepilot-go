# Logger - MoviePilot 日志系统

## 概述

Logger是MoviePilot项目的日志管理模块，提供统一的日志记录功能，支持结构化日志、上下文信息、多级别日志和灵活的配置选项。

## 特性

- ✅ 支持多级别日志：Debug、Info、Warn、Error、Fatal、Panic
- ✅ 结构化日志输出（JSON格式）
- ✅ 上下文信息支持（request_id、user_id、trace_id）
- ✅ 环境变量配置支持
- ✅ 文件轮转和压缩
- ✅ 多种输出方式：标准输出、文件输出、两者兼顾
- ✅ 调用者信息和堆栈跟踪
- ✅ 格式化日志方法支持

## 快速开始

### 初始化

在应用启动时初始化日志系统：

```go
import "moviepilot-go/pkg/logger"

func main() {
    // 初始化日志系统
    if err := logger.Init(); err != nil {
        panic("日志初始化失败: " + err.Error())
    }
    defer logger.Sync() // 确保日志缓冲区被刷新
    
    // 应用逻辑...
}
```

### 基本用法

```go
// 使用不同级别的日志
logger.Debug("调试信息", "key", "value")
logger.Info("用户登录", "user_id", "123", "ip", "192.168.1.1")
logger.Warn("配置文件缺失，使用默认配置", "config_file", "/etc/moviepilot/config.yaml")
logger.Error("数据库连接失败", "error", err.Error(), "retry_count", 3)

// 使用格式化日志
logger.Debugf("处理请求: %s, 耗时: %dms", path, duration)
logger.Infof("用户 %s 成功注册", username)
logger.Warnf("磁盘空间不足: %d%%", usage)
logger.Errorf("API调用失败: %v", err)
```

### 带上下文的日志

```go
import (
    "context"
    "moviepilot-go/pkg/logger"
)

func handler(ctx context.Context) {
    // 添加上下文信息
    ctx = context.WithValue(ctx, logger.ContextKeyRequestID, "req-123456")
    ctx = context.WithValue(ctx, logger.ContextKeyUserID, "user-7890")
    
    // 获取带上下文的日志实例
    log := logger.WithContext(ctx)
    
    // 使用带上下文的日志
    log.Info("处理请求开始")
    // 业务逻辑...
    log.Info("处理请求完成", "status", "success")
}
```

## 配置选项

### 环境变量配置

日志系统支持通过以下环境变量进行配置：

| 环境变量 | 描述 | 默认值 |
|---------|------|-------|
| LOGGER_LEVEL | 日志级别（debug/info/warn/error/fatal/panic） | info |
| LOGGER_FORMAT | 日志格式（json/console） | json |
| LOGGER_OUTPUT | 输出方式（stdout/file/both） | file |
| LOGGER_FILE | 日志文件路径 | /var/log/moviepilot/app.log |
| LOGGER_MAX_SIZE | 单个日志文件最大大小（MB） | 100 |
| LOGGER_MAX_BACKUPS | 保留的最大日志文件数量 | 3 |
| LOGGER_MAX_AGE | 日志保留天数 | 28 |
| LOGGER_COMPRESS | 是否压缩旧日志文件 | true |

### 配置文件配置

日志系统也支持通过配置文件进行配置（优先级低于环境变量）：

```yaml
log:
  level: "info"  # 日志级别
  format: "json"  # 日志格式
  output: "file"  # 输出方式
  file_path: "/var/log/moviepilot/app.log"  # 日志文件路径
  max_size: 100  # 单个日志文件最大大小（MB）
  max_backups: 3  # 保留的最大日志文件数量
  max_age: 28  # 日志保留天数
```

## 日志格式

JSON格式的日志包含以下字段：

- `timestamp`: ISO8601格式的时间戳
- `level`: 日志级别（debug/info/warn/error/fatal/panic）
- `caller`: 调用者信息（文件:行号）
- `message`: 日志消息
- `stacktrace`: 错误堆栈信息（仅错误级别）
- 自定义字段: 通过fields参数添加的自定义键值对
- 上下文字段: request_id, user_id, trace_id（如果提供）

## 最佳实践

### 1. 在不同层的使用

#### API层

```go
func (h *Handler) CreateUser(c *gin.Context) {
    // 获取或生成请求ID
    requestID := generateRequestID()
    
    // 创建带上下文的日志
    ctx := context.WithValue(c.Request.Context(), logger.ContextKeyRequestID, requestID)
    log := logger.WithContext(ctx)
    
    // 记录请求开始
    log.Info("API请求开始", 
        "method", c.Request.Method,
        "path", c.Request.URL.Path,
        "ip", c.ClientIP())
    
    // 业务逻辑...
    
    // 记录请求结束
    log.Info("API请求结束", 
        "status_code", http.StatusCreated,
        "duration", time.Since(startTime).Milliseconds())
}
```

#### 服务层

```go
func (s *UserService) CreateUser(req *UserCreateRequest) (*User, error) {
    // 记录服务方法开始
    logger.Debug("CreateUser服务方法开始", 
        "func", "UserService.CreateUser",
        "username", req.Username)
    
    // 业务逻辑...
    
    // 记录关键业务节点
    logger.Info("用户创建成功", 
        "user_id", user.ID,
        "username", user.Username)
    
    return user, nil
}
```

#### 数据访问层

```go
func (r *UserRepository) GetByID(id string) (*User, error) {
    // 记录数据库操作
    logger.Debug("数据库查询开始", 
        "func", "UserRepository.GetByID",
        "user_id", id)
    
    // 执行查询...
    
    // 记录查询结果
    logger.Info("数据库查询完成", 
        "user_id", id,
        "found", user != nil,
        "duration", time.Since(startTime).Milliseconds())
    
    return user, nil
}
```

### 2. 错误处理

```go
if err != nil {
    logger.Error("操作失败", 
        "error", err.Error(),
        "operation", "process_data",
        "entity_id", entityID,
        "retry_count", retryCount)
    
    // 根据错误类型决定下一步操作
    return nil, err
}
```

### 3. 避免记录敏感信息

**禁止**在日志中记录以下敏感信息：
- 密码
- 访问令牌/密钥
- 个人身份信息（PII）
- 财务信息

```go
// ❌ 错误示例
logger.Info("用户登录", "username", username, "password", password)

// ✅ 正确示例
logger.Info("用户登录", "username", username)
```

## 高级特性

### 1. 上下文传递

在请求处理链路中传递上下文，确保整个调用链的日志包含相同的跟踪信息：

```go
func middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 创建新的上下文
        ctx := r.Context()
        ctx = context.WithValue(ctx, logger.ContextKeyRequestID, generateRequestID())
        
        // 注入到请求中
        r = r.WithContext(ctx)
        
        next.ServeHTTP(w, r)
    })
}
```

### 2. 性能监控

使用日志记录关键操作的执行时间，用于性能分析：

```go
startTime := time.Now()

// 执行操作...

duration := time.Since(startTime)
logger.Info("操作完成", 
    "operation", "process_batch",
    "items", len(items),
    "duration_ms", duration.Milliseconds())
```

## 故障排除

### 日志文件权限问题

如果日志文件创建失败，请检查日志目录的权限设置：

```bash
mkdir -p /var/log/moviepilot
chmod 755 /var/log/moviepilot
```

### 日志级别调整

在开发环境中，可以将日志级别设置为debug以获取更详细的信息：

```bash
export LOGGER_LEVEL=debug
```

在生产环境中，建议使用info级别：

```bash
export LOGGER_LEVEL=info
```

## 示例

详细的使用示例可以查看 [example.go](example.go) 文件。

## 测试

运行日志系统的单元测试：

```bash
go test ./pkg/logger -v
```
