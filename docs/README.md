# MoviePilot Go 迁移文档索引

> 本目录包含从 Python MoviePilot 到 Go moviepilot-go 的完整迁移设计文档。

---

## 📚 文档列表

### 核心文档

1. **[migration-overview.md](./migration-overview.md)** - 迁移总览
   - 整体架构对比
   - 分层设计原则
   - 迁移路线图
   - 技术栈对比

2. **[core-migration-app-core.md](./core-migration-app-core.md)** - app/core 模块迁移
   - 配置管理（config.py）
   - 上下文管理（context.py）
   - 事件总线（event.py）
   - 缓存系统（cache.py）
   - 安全模块（security.py）

### 启动与调度

3. **[startup-migration.md](./startup-migration.md)** - 启动初始化
   - Bootstrap 流程设计
   - 组件初始化顺序
   - 优雅关闭机制

4. **[scheduler-migration.md](./scheduler-migration.md)** - 定时任务
   - 调度器设计
   - 内建任务迁移
   - 插件任务注册
   - Cron 表达式支持

### 业务层

5. **[chain-migration.md](./chain-migration.md)** - 业务处理链
   - Service 层设计
   - 依赖注入
   - 工作流编排
   - 事件驱动

### 数据层

6. **[db-migration.md](./db-migration.md)** - 数据库层
   - GORM 模型定义
   - Repository 模式
   - 事务管理
   - 数据库迁移

### 其他模块

7. **[remaining-modules-migration.md](./remaining-modules-migration.md)** - 其余模块快速索引
   - command.py - 命令管理
   - factory.py - 应用工厂
   - log.py - 日志系统
   - monitor.py - 文件监控
   - main.py - 应用入口
   - schemas/ - 数据模型
   - api/ - API 层
   - helper/ - 辅助工具
   - modules/ - 外部服务模块
   - plugins/ - 插件系统
   - actions/ - 动作处理

---

## 🗂️ 文档状态

| 文档 | 状态 | 完成度 |
|------|------|--------|
| migration-overview.md | ✅ 完成 | 100% |
| core-migration-app-core.md | ✅ 完成 | 100% |
| startup-migration.md | ✅ 完成 | 100% |
| scheduler-migration.md | ✅ 完成 | 100% |
| chain-migration.md | ✅ 完成 | 100% |
| db-migration.md | ✅ 完成 | 100% |
| remaining-modules-migration.md | ✅ 完成 | 80% |
| api-migration.md | 📝 待创建 | 0% |
| schemas-migration.md | 📝 待创建 | 0% |
| helper-migration.md | 📝 待创建 | 0% |
| modules-migration.md | 📝 待创建 | 0% |
| plugins-migration.md | 📝 待创建 | 0% |
| command-migration.md | 📝 待创建 | 0% |
| factory-migration.md | 📝 待创建 | 0% |
| log-migration.md | 📝 待创建 | 0% |
| monitor-migration.md | 📝 待创建 | 0% |
| main-migration.md | 📝 待创建 | 0% |

---

## 📖 阅读顺序建议

### 新手入门
1. [migration-overview.md](./migration-overview.md) - 了解整体架构
2. [core-migration-app-core.md](./core-migration-app-core.md) - 理解核心模块
3. [startup-migration.md](./startup-migration.md) - 学习启动流程

### 开发者
1. [chain-migration.md](./chain-migration.md) - 业务层设计
2. [db-migration.md](./db-migration.md) - 数据层设计
3. [scheduler-migration.md](./scheduler-migration.md) - 定时任务设计

### 架构师
1. [migration-overview.md](./migration-overview.md) - 整体架构
2. 所有核心文档 - 深入理解各模块

---

## 🎯 迁移路线图

### Phase 1: 基础架构（Week 1-3）
- [x] 项目骨架
- [x] 日志系统 (`pkg/logger/`)
- [x] 系统工具 (`pkg/utils/system.go`)
- [x] 监控采集器 (`internal/monitor/metrics/`)
- [ ] 配置系统完善
- [ ] 缓存系统
- [ ] 数据库迁移

### Phase 2: 核心功能（Week 4-8）
- [ ] 用户认证与授权
- [ ] 站点管理
- [ ] 订阅系统
- [ ] 下载管理
- [ ] 文件整理（transfer）

### Phase 3: 插件与扩展（Week 9-11）
- [ ] 插件系统重构（Go + Python gRPC）
- [ ] 工作流引擎
- [ ] 消息通知

### Phase 4: 优化与部署（Week 12-15）
- [ ] 性能优化
- [ ] 监控与告警（Prometheus + Grafana）
- [ ] CI/CD 自动化
- [ ] 文档完善

---

## 🔗 相关资源

### 代码仓库
- **Go 主应用**：`/workspaces/moviepilot/moviepilot-go-project/moviepilot-go/`
- **Python 原版**：`/workspaces/moviepilot/MoviePilot-2.8.1-1/`

### 技术栈文档
- [Gin Web Framework](https://gin-gonic.com/)
- [GORM](https://gorm.io/)
- [Zap Logger](https://github.com/uber-go/zap)
- [Viper Config](https://github.com/spf13/viper)
- [Cron](https://github.com/robfig/cron)

### 设计模式
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Domain-Driven Design](https://martinfowler.com/bliki/DomainDrivenDesign.html)
- [Repository Pattern](https://martinfowler.com/eaaCatalog/repository.html)

---

## 📝 文档维护

### 更新规则
- 每完成一个模块迁移，更新对应文档状态
- 记录遇到的问题与解决方案
- 补充代码示例和最佳实践

### 贡献指南
1. 创建新文档时，遵循现有文档格式
2. 使用清晰的标题和章节结构
3. 提供代码示例和对比
4. 更新本 README 的文档列表

---

## ❓ FAQ

### Q: 为什么选择 Go 而不是继续用 Python？
A: Go 提供更好的性能、并发支持和部署便利性，适合长期运行的服务。

### Q: 是否完全抛弃 Python？
A: 不是。插件系统仍支持 Python，通过 gRPC 与 Go 主应用通信。

### Q: 迁移过程中如何保证兼容性？
A: API 接口、数据库 schema、配置文件格式保持与 Python 版本兼容。

### Q: 预计迁移周期多长？
A: 完整迁移预计 12-15 周，分 4 个阶段进行。

---

**最后更新**：2025-11-26  
**维护者**：MoviePilot Go 迁移团队
