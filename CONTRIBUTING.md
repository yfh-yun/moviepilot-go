# 贡献指南

感谢你考虑为 MoviePilot Go 做出贡献！

## 📋 目录

- [行为准则](#行为准则)
- [如何贡献](#如何贡献)
- [开发流程](#开发流程)
- [代码规范](#代码规范)
- [提交规范](#提交规范)
- [测试要求](#测试要求)
- [文档要求](#文档要求)

---

## 行为准则

本项目采用 [Contributor Covenant](https://www.contributor-covenant.org/) 行为准则。参与本项目即表示你同意遵守其条款。

### 我们的承诺

- 使用友好和包容的语言
- 尊重不同的观点和经验
- 优雅地接受建设性批评
- 关注对社区最有利的事情
- 对其他社区成员表示同理心

---

## 如何贡献

### 报告 Bug

在提交 Bug 报告前，请：

1. 检查 [Issues](https://github.com/your-org/moviepilot-go/issues) 确保问题未被报告
2. 使用最新版本重现问题
3. 收集相关信息（日志、截图、环境信息）

**Bug 报告应包含**:
- 清晰的标题
- 详细的问题描述
- 重现步骤
- 预期行为
- 实际行为
- 环境信息（OS、Go 版本、Docker 版本等）
- 相关日志和截图

### 建议功能

功能建议应包含：
- 功能的详细描述
- 使用场景
- 可能的实现方案
- 是否愿意实现该功能

### 提交 Pull Request

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

---

## 开发流程

### 1. 环境准备

```bash
# 克隆仓库
git clone https://github.com/your-org/moviepilot-go.git
cd moviepilot-go

# 安装依赖
go mod download

# 安装开发工具
make install-tools

# 启动开发环境
make dev
```

### 2. 创建分支

分支命名规范：
- `feature/xxx` - 新功能
- `fix/xxx` - Bug 修复
- `docs/xxx` - 文档更新
- `refactor/xxx` - 代码重构
- `test/xxx` - 测试相关

### 3. 开发

- 遵循 [代码规范](#代码规范)
- 编写单元测试
- 更新相关文档
- 确保所有测试通过

### 4. 提交代码

```bash
# 格式化代码
make fmt

# 运行 lint
make lint

# 运行测试
make test

# 提交
git add .
git commit -m "feat: add amazing feature"
```

### 5. 推送并创建 PR

```bash
git push origin feature/AmazingFeature
```

然后在 GitHub 上创建 Pull Request。

---

## 代码规范

### Go 代码规范

遵循 [Effective Go](https://golang.org/doc/effective_go.html) 和 [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)。

**关键原则**:

1. **命名规范**
   ```go
   // 包名：小写、简短、有意义
   package transfer
   
   // 导出函数：驼峰命名，首字母大写
   func CreateTransferHistory() {}
   
   // 私有函数：驼峰命名，首字母小写
   func validatePath() {}
   
   // 常量：大写、下划线分隔
   const MAX_RETRY_COUNT = 3
   
   // 接口：以 -er 结尾
   type TransferRepository interface {}
   ```

2. **错误处理**
   ```go
   // ✅ 正确
   if err != nil {
       return fmt.Errorf("failed to create transfer: %w", err)
   }
   
   // ❌ 错误
   if err != nil {
       panic(err)
   }
   ```

3. **日志记录**
   ```go
   // ✅ 使用 pkg/logger
   logger.Info("Transfer created", "id", transferID)
   logger.Error("Failed to create transfer", "error", err)
   
   // ❌ 不要使用 fmt.Println
   fmt.Println("Transfer created")
   ```

4. **注释规范**
   ```go
   // CreateTransferHistory 创建转移历史记录
   // 参数:
   //   - req: 转移请求
   // 返回:
   //   - *TransferHistory: 创建的历史记录
   //   - error: 错误信息
   func CreateTransferHistory(req TransferRequest) (*TransferHistory, error) {
       // ...
   }
   ```

### 项目结构规范

遵循 [项目规则](projectrules.md) 中定义的分层架构：

```
internal/
├── apis/          # API 层（HTTP 处理）
├── business/      # 业务层（核心逻辑）
├── infrastructure/# 基础设施层（技术细节）
├── integration/   # 集成层（外部服务）
├── models/        # 数据模型
└── repositories/  # 数据访问
```

**分层依赖规则**:
- APIs → Business → Infrastructure → Repositories
- 禁止反向依赖
- 使用接口和依赖注入

---

## 提交规范

使用 [Conventional Commits](https://www.conventionalcommits.org/) 规范。

### 提交格式

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Type 类型

- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档更新
- `style`: 代码格式（不影响代码运行）
- `refactor`: 重构
- `perf`: 性能优化
- `test`: 测试相关
- `chore`: 构建过程或辅助工具变动
- `ci`: CI 配置文件和脚本变动

### 示例

```bash
# 新功能
git commit -m "feat(subscription): add auto refresh feature"

# Bug 修复
git commit -m "fix(download): resolve connection timeout issue"

# 文档
git commit -m "docs(api): update subscription API documentation"

# 重构
git commit -m "refactor(transfer): improve file naming strategy"
```

---

## 测试要求

### 单元测试

- 所有新功能必须包含单元测试
- 测试覆盖率应 > 80%
- 使用 table-driven tests

```go
func TestCreateTransferHistory(t *testing.T) {
    tests := []struct {
        name    string
        req     TransferRequest
        want    *TransferHistory
        wantErr bool
    }{
        {
            name: "valid request",
            req:  TransferRequest{/* ... */},
            want: &TransferHistory{/* ... */},
            wantErr: false,
        },
        // 更多测试用例...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := CreateTransferHistory(tt.req)
            if (err != nil) != tt.wantErr {
                t.Errorf("CreateTransferHistory() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            // 断言...
        })
    }
}
```

### 集成测试

- 测试多个组件的交互
- 使用测试数据库
- 清理测试数据

### 性能测试

```go
func BenchmarkCreateTransferHistory(b *testing.B) {
    for i := 0; i < b.N; i++ {
        CreateTransferHistory(testReq)
    }
}
```

### 运行测试

```bash
# 运行所有测试
make test

# 运行测试并生成覆盖率报告
make test-cover

# 运行性能测试
make bench
```

---

## 文档要求

### 代码文档

- 所有导出的函数、类型、常量必须有注释
- 注释应清晰描述功能、参数、返回值
- 复杂逻辑需要添加行内注释

### API 文档

使用 Swagger 注解：

```go
// @Summary Create transfer history
// @Description Create a new transfer history record
// @Tags transfers
// @Accept json
// @Produce json
// @Param transfer body TransferRequest true "Transfer data"
// @Success 201 {object} TransferHistory
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/transfers [post]
func CreateTransferHistory(c *gin.Context) {
    // ...
}
```

### README 更新

- 新功能需要更新 README.md
- 添加使用示例
- 更新功能列表

### Changelog

- 重要更改需要更新 CHANGELOG.md
- 遵循 [Keep a Changelog](https://keepachangelog.com/) 格式

---

## Pull Request 检查清单

提交 PR 前，请确保：

- [ ] 代码遵循项目规范
- [ ] 所有测试通过
- [ ] 测试覆盖率 > 80%
- [ ] 添加了必要的文档
- [ ] 更新了 CHANGELOG.md
- [ ] Commit 信息符合规范
- [ ] 没有引入新的 lint 警告
- [ ] PR 描述清晰，包含相关 Issue 链接

---

## 代码审查

### 审查重点

1. **功能正确性** - 代码是否实现了预期功能
2. **代码质量** - 是否遵循最佳实践
3. **测试充分性** - 测试是否覆盖主要场景
4. **性能影响** - 是否有性能问题
5. **安全性** - 是否有安全隐患
6. **文档完整性** - 文档是否清晰完整

### 审查流程

1. 自动化检查（CI）
2. 代码审查（至少 1 位维护者）
3. 测试验证
4. 合并到主分支

---

## 获取帮助

- **文档**: [docs/](docs/)
- **Issues**: [GitHub Issues](https://github.com/your-org/moviepilot-go/issues)
- **讨论**: [GitHub Discussions](https://github.com/your-org/moviepilot-go/discussions)
- **Telegram**: https://t.me/moviepilot_go

---

## 许可证

贡献的代码将采用与项目相同的 [MIT License](LICENSE)。

---

**感谢你的贡献！** ❤️
