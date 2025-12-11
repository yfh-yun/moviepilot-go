# db/ 数据库层迁移设计

> Python: `app/db/`  
> Go: `internal/repositories/` + `internal/models/`

---

## 1. Python `db/` 分析

### 1.1 目录结构

```
app/db/
├── __init__.py                  # 数据库初始化、Session 管理
├── models/                      # SQLAlchemy ORM 模型
│   ├── downloadhistory.py
│   ├── mediaserver.py
│   ├── message.py
│   ├── plugindata.py
│   ├── site.py
│   ├── subscribe.py
│   ├── systemconfig.py
│   ├── transferhistory.py
│   ├── user.py
│   ├── userconfig.py
│   └── workflow.py
└── *_oper.py                    # 各表的 CRUD 操作类
    ├── downloadhistory_oper.py
    ├── site_oper.py
    ├── subscribe_oper.py
    └── ...
```

### 1.2 核心能力

- **ORM**：SQLAlchemy
- **连接池**：QueuePool / NullPool
- **数据库支持**：SQLite / PostgreSQL
- **事务管理**：Session + context manager
- **操作封装**：每个表一个 `*_oper.py`，提供 CRUD 方法

---

## 2. Go 设计方案

### 2.1 目录结构

```
internal/
├── models/                      # 数据模型（GORM）
│   └── database/                # 数据库模型
│       ├── models.go            # 所有数据库表模型
│       ├── base.go              # 基础模型（包含ID、创建时间、更新时间）
│       └── ...
├── repositories/                # 数据访问层
│   ├── interfaces/              # Repository 接口
│   │   ├── user_repository.go
│   │   ├── subscribe_repository.go
│   │   ├── site_repository.go
│   │   └── ...
│   ├── repositories/            # Repository 实现
│   │   ├── user_repository.go
│   │   ├── subscribe_repository.go
│   │   ├── site_repository.go
│   │   └── ...
│   └── migrations/              # 数据库迁移工具
└── pkg/database/                # 数据库连接封装
    ├── database.go
    ├── sqlite.go
    └── postgres.go
```

### 2.2 数据模型定义

所有数据模型已统一在 `internal/models/database/models.go` 中实现，包含：
- User 用户表
- Media 媒体表
- Subscribe 订阅表
- Download 下载任务表
- DownloadHistory 下载历史表
- TransferHistory 转移历史表
- MediaServer 媒体服务器配置表
- MediaServerItem 媒体服务器项目表
- Message 消息表
- PluginData 插件数据表
- SystemConfig 系统配置表
- UserConfig 用户配置表
- Site 站点表
- SiteIcon 站点图标表
- SiteUserData 站点用户数据表
- SiteStatistic 站点统计表
- SubscribeHistory 订阅历史表
- DownloadFiles 下载文件表
- Workflow 工作流表
- Search 搜索历史表
- Plugin 插件表
- ScrapeEvent 刮削事件表
- ScrapeResult 刮削结果表
- MediaInfo 媒体信息表

### 2.3 Repository 接口与实现

已实现的 Repository 接口和实现：

| 模型 | 接口 | 实现 |
|------|------|------|
| User | UserRepository | user_repository.go |
| Subscribe | SubscribeRepository | subscribe_repository.go |
| Site | SiteRepository | site_repository.go |
| DownloadHistory | DownloadHistoryRepository | download_history_repository.go |
| TransferHistory | TransferHistoryRepository | transfer_history_repository.go |
| MediaServer | MediaServerRepository | media_server_repository.go |
| Message | MessageRepository | message_repository.go |
| PluginData | PluginDataRepository | plugin_data_repository.go |
| SystemConfig | SystemConfigRepository | system_config_repository.go |
| UserConfig | UserConfigRepository | user_config_repository.go |
| Workflow | WorkflowRepository | workflow_repository.go |
| Media | MediaRepository | media_repository.go |
| SubscribeHistory | SubscribeHistoryRepository | subscribe_history_repository.go |
| SiteUserData | SiteUserDataRepository | site_user_data_repository.go |
| SiteStatistic | SiteStatisticRepository | site_statistic_repository.go |
| SiteIcon | SiteIconRepository | site_icon_repository.go |

### 2.4 数据库初始化

数据库初始化已在 `pkg/database/database.go` 中实现，支持 SQLite 和 PostgreSQL，使用 GORM 的 AutoMigrate 进行数据库迁移。

### 2.5 事务管理

事务管理已在 `pkg/database/database.go` 中通过 GORM 的 Transaction 方法实现。

---

## 3. 迁移进度

| 阶段 | 任务 | 状态 | 完成时间 |
|------|------|------|----------|
| **Phase 1** | 核心表模型迁移 | ✅ 已完成 | 已实现 |
| **Phase 2** | Repository 接口与实现 | ✅ 已完成 | 已实现 |
| **Phase 3** | 事务管理 + 迁移工具 | ✅ 已完成 | 已实现 |
| **Phase 4** | 性能优化（索引、连接池） | ⏳ 进行中 | 待优化 |

---

## 4. 迁移计划归纳

### 4.1 迁移目标

将 Python 项目的 `app/db/` 目录迁移到 Go 项目的 `internal/repositories/` + `internal/models/` 目录，实现从 SQLAlchemy 到 GORM 的 ORM 迁移，从 SQLite 到 PostgreSQL 的数据库迁移。

### 4.2 迁移策略

1. **模型迁移**：将 SQLAlchemy 模型转换为 GORM 模型，统一放在 `internal/models/database/` 目录下
2. **Repository 迁移**：将 `*_oper.py` 转换为 Go 的 Repository 接口和实现，放在 `internal/repositories/` 目录下
3. **数据库连接**：使用 GORM 实现 SQLite 和 PostgreSQL 的数据库连接
4. **迁移工具**：使用 GORM 的 AutoMigrate 进行数据库迁移
5. **事务管理**：使用 GORM 的 Transaction 方法进行事务管理

### 4.3 技术栈对比

| 特性 | Python (SQLAlchemy) | Go (GORM) |
|------|---------------------|-----------|
| ORM | SQLAlchemy | GORM |
| Session | scoped_session | db.WithContext |
| 事务 | with session.begin() | db.Transaction |
| 迁移 | Alembic | GORM AutoMigrate |
| 查询 | Query API | Method Chain |
| 数据库支持 | SQLite | SQLite, PostgreSQL |

### 4.4 迁移成果

1. **统一的数据模型**：所有数据库表模型都已统一在 `internal/models/database/models.go` 中实现
2. **完整的 Repository 层**：所有核心模型都已实现对应的 Repository 接口和实现
3. **灵活的数据库支持**：支持 SQLite 和 PostgreSQL，可通过配置切换
4. **自动化的数据库迁移**：使用 GORM 的 AutoMigrate 实现自动化的数据库迁移
5. **完善的事务管理**：支持使用 GORM 的 Transaction 方法进行事务管理

### 4.5 后续优化方向

1. **性能优化**：添加合适的索引，优化查询性能
2. **连接池优化**：配置合适的连接池参数
3. **日志优化**：添加更详细的数据库操作日志
4. **监控优化**：添加数据库连接和查询监控
5. **测试优化**：完善数据库相关的单元测试和集成测试

---

## 5. 结论

数据库层迁移已基本完成，核心功能已实现，系统可以正常运行。后续将继续优化性能，完善监控和测试，确保系统的稳定性和可靠性。
