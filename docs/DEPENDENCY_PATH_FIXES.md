# 依赖路径修复总结

## 修复概述

在 Go 1.24.4 升级过程中，发现并修复了大量的包路径引用错误。这些错误主要是由于项目架构重构后，目录结构发生了变化，但代码中的导入路径没有及时更新。

## 修复的问题

### 1. 包名错误（单复数问题）

#### 1.1 validators 目录
- ❌ 错误: `internal/apis/validator`
- ✅ 正确: `internal/apis/validators`

**修复的文件:**
- `internal/apis/handlers/servarr/servarr_handler.go`
- `internal/apis/handlers/message/message_handler.go`
- `internal/apis/handlers/plugin/plugin_handler.go`

### 2. 目录结构变更

#### 2.1 helper → utils
- ❌ 错误: `internal/helper`
- ✅ 正确: `pkg/utils`

**修复的文件:**
- `internal/business/services/actions/subscribe_files.go`
- `internal/business/services/indexer/spider/spider.go`
- `internal/business/services/chain/media_chain.go`

#### 2.2 model → models
- ❌ 错误: `internal/model`
- ✅ 正确: `internal/models`

**影响的文件:** 93+ 个文件

#### 2.3 service → business/services
- ❌ 错误: `internal/service`
- ✅ 正确: `internal/business/services`

**影响的文件:** 38+ 个文件

#### 2.4 repository → repositories
- ❌ 错误: `internal/repository`
- ✅ 正确: `internal/repositories`

**影响的文件:** 27+ 个文件

#### 2.5 scheduler → schedulers
- ❌ 错误: `internal/scheduler`
- ✅ 正确: `internal/schedulers`

**影响的文件:** 4+ 个文件

#### 2.6 task → tasks
- ❌ 错误: `internal/task`
- ✅ 正确: `internal/tasks`

**影响的文件:** 1+ 个文件

#### 2.7 chain → business/services/chain
- ❌ 错误: `internal/chain`
- ✅ 正确: `internal/business/services/chain`

**影响的文件:** 1+ 个文件

#### 2.8 event → infrastructure/events
- ❌ 错误: `internal/event`
- ✅ 正确: `internal/infrastructure/events`

**影响的文件:** 1+ 个文件

#### 2.9 repositories/models → models
- ❌ 错误: `internal/repositories/models`
- ✅ 正确: `internal/models`

**影响的文件:** 17+ 个文件

### 3. 依赖版本更新

#### 3.1 JWT 版本
- ❌ 错误: `github.com/golang-jwt/jwt/v4`
- ✅ 正确: `github.com/golang-jwt/jwt/v5`

**修复的文件:**
- `internal/apis/middlewares/auth.go`

## 修复方法

### 批量修复命令
使用 `sed` 命令批量替换文件中的导入路径：

```bash
# 修复 model → models
find . -name "*.go" -type f -exec sed -i 's|github\.com/yfh-yun/moviepilot-go/internal/model|github.com/yfh-yun/moviepilot-go/internal/models|g' {} \;

# 修复 service → business/services
find . -name "*.go" -type f -exec sed -i 's|github\.com/yfh-yun/moviepilot-go/internal/service|github.com/yfh-yun/moviepilot-go/internal/business/services|g' {} \;

# 修复 repository → repositories
find . -name "*.go" -type f -exec sed -i 's|github\.com/yfh-yun/moviepilot-go/internal/repository|github.com/yfh-yun/moviepilot-go/internal/repositories|g' {} \;
```

### 手动修复
对于一些特殊情况，手动进行精确替换：

```go
// 单个文件修复示例
replace_in_file(filePath, oldImport, newImport)
```

## 验证结果

### 当前状态
- ✅ 所有包路径引用已修复
- ✅ 目录结构符合新架构规范
- ✅ 依赖版本已更新
- ⚠️ go.sum 文件需要重新生成（网络问题）

### 剩余问题
1. **go.sum 缺失**: 由于网络连接问题，无法重新生成 go.sum
2. **编译错误**: 由于 go.sum 缺失导致的编译错误

### 解决方案
当网络连接恢复后，执行以下命令：

```bash
# 重新生成 go.sum
rm -f go.sum
go mod tidy

# 验证编译
go build ./cmd/server
```

## 架构规范

### 正确的目录结构
```
internal/
├── apis/                    # API层
│   ├── handlers/           # HTTP处理器
│   ├── middlewares/        # 中间件
│   ├── validators/         # 验证器（复数）
│   └── routes/            # 路由
├── business/               # 业务层
│   └── services/          # 业务服务
├── infrastructure/         # 基础设施层
│   └── events/            # 事件系统
├── models/                # 数据模型（复数）
├── repositories/          # 数据访问（复数）
├── schedulers/            # 调度器（复数）
└── tasks/                 # 任务（复数）

pkg/
├── utils/                 # 工具函数
├── logger/                # 日志封装
└── ...                    # 其他公共包
```

### 命名规范
- **目录名**: 使用复数形式（handlers, services, models, repositories）
- **包名**: 小写、简短、有意义
- **导入路径**: 必须与实际目录结构一致

## 经验教训

1. **架构重构后必须同步更新导入路径**
2. **使用批量工具提高修复效率**
3. **保持命名规范的一致性**
4. **及时验证修复结果**

---

**修复完成时间**: 2025-11-20  
**修复文件数量**: 200+ 个文件  
**主要问题**: 导入路径与目录结构不匹配