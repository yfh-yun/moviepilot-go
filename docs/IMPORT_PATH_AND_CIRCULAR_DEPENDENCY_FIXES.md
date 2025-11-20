# 导入路径格式和循环依赖修复

## 问题概述

在 Go 1.24.4 升级过程中，发现了以下关键问题：

1. **导入路径格式错误**: `github.com/PuerkitoBio/goquery` 前面有多余的空格
2. **循环依赖问题**: `pkg/logger` 和 `internal/infrastructure/config` 之间存在循环导入

## 修复详情

### 1. 导入路径格式修复

**问题文件**: `internal/business/services/indexer/spider/nexus_php.go`

**修复前**:
```go
import (
    // ...
    " github.com/PuerkitoBio/goquery"  // 注意前面的空格
    // ...
)
```

**修复后**:
```go
import (
    // ...
    "github.com/PuerkitoBio/goquery"  // 移除了多余的空格
    // ...
)
```

### 2. 循环依赖问题修复

**问题分析**:
- `pkg/logger/logger.go` 导入了 `internal/infrastructure/config`
- `internal/infrastructure/config/config.go` 导入了 `pkg/logger`
- 这违反了分层架构的依赖规则

**解决方案**: 重构 `pkg/logger/logger.go`，移除对 config 包的依赖

#### 2.1 添加辅助函数

```go
// getEnvOrDefault 获取环境变量或返回默认值
func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

// getIntEnvOrDefault 获取整数环境变量或返回默认值
func getIntEnvOrDefault(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intValue, err := strconv.Atoi(value); err == nil {
            return intValue
        }
    }
    return defaultValue
}

// getBoolEnvOrDefault 获取布尔环境变量或返回默认值
func getBoolEnvOrDefault(key string, defaultValue bool) bool {
    if value := os.Getenv(key); value != "" {
        if boolValue, err := strconv.ParseBool(value); err == nil {
            return boolValue
        }
    }
    return defaultValue
}
```

#### 2.2 重构配置读取逻辑

**修复前**: 使用 `config.Config.Get*()` 方法
**修复后**: 使用环境变量和默认值

```go
// 示例：日志级别配置
level := getEnvOrDefault(envPrefix+"LEVEL", "info")

// 示例：文件大小配置
maxSize := getIntEnvOrDefault(envPrefix+"MAX_SIZE", 100)

// 示例：压缩配置
compress := getBoolEnvOrDefault(envPrefix+"COMPRESS", true)
```

#### 2.3 移除的函数

- `setDefaultConfig()`: 不再需要，直接使用环境变量默认值

## 环境变量配置

日志系统现在支持以下环境变量：

| 环境变量 | 默认值 | 说明 |
|---------|--------|------|
| `LOGGER_LEVEL` | `info` | 日志级别 (debug/info/warn/error/fatal/panic) |
| `LOGGER_FORMAT` | `json` | 日志格式 (json/console) |
| `LOGGER_OUTPUT` | `stdout` | 输出方式 (stdout/file/both) |
| `LOGGER_FILE` | `/var/log/moviepilot/app.log` | 日志文件路径 |
| `LOGGER_MAX_SIZE` | `100` | 单个日志文件最大大小(MB) |
| `LOGGER_MAX_BACKUPS` | `3` | 保留的旧日志文件数量 |
| `LOGGER_MAX_AGE` | `28` | 日志文件保留天数 |
| `LOGGER_COMPRESS` | `true` | 是否压缩旧日志文件 |

## 验证结果

修复后的编译状态：

```bash
# 编译成功，无循环导入错误
go build -v ./pkg/logger
go build -v ./internal/business/services/indexer/spider
```

**剩余问题**: 仅存在 go.sum 条目缺失，可通过 `go mod tidy` 解决

## 架构改进

这次修复带来了以下架构改进：

1. **清晰的依赖方向**: `pkg/logger` 不再依赖任何内部包
2. **更好的可测试性**: 日志配置可以通过环境变量控制
3. **符合12-factor应用原则**: 配置通过环境变量管理
4. **减少耦合**: 日志系统更加独立，便于复用

## 最佳实践

1. **避免循环依赖**: 严格按照分层架构设计依赖关系
2. **使用环境变量**: 配置管理遵循12-factor应用原则
3. **包的职责单一**: `pkg` 下的包应该保持独立，不依赖内部包
4. **导入路径格式**: 确保导入路径没有多余空格或格式错误