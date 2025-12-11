# 数据库迁移文档

> **创建时间**: 2025-12-02  
> **Week 7 Day 1**: 数据库迁移脚本

---

## 📋 迁移清单

### 用户认证相关表（6 张）

| 序号 | 表名 | 说明 | 文件 |
|------|------|------|------|
| 000001 | users | 用户表 | 000001_create_users_table.sql |
| 000002 | roles | 角色表 | 000002_create_roles_table.sql |
| 000003 | permissions | 权限表 | 000003_create_permissions_table.sql |
| 000004 | user_roles | 用户角色关联表 | 000004_create_user_roles_table.sql |
| 000005 | role_permissions | 角色权限关联表 | 000005_create_role_permissions_table.sql |
| 000006 | auth_logs | 认证日志表 | 000006_create_auth_logs_table.sql |

### 站点管理相关表（5 张）

| 序号 | 表名 | 说明 | 文件 |
|------|------|------|------|
| 000007 | sites | 站点表 | 000007_create_sites_table.sql |
| 000008 | site_cookies | Cookie历史表 | 000008_create_site_cookies_table.sql |
| 000009 | checkin_logs | 签到日志表 | 000009_create_checkin_logs_table.sql |
| 000010 | site_stats | 流量统计表 | 000010_create_site_stats_table.sql |
| 000011 | sync_logs | 同步日志表 | 000011_create_sync_logs_table.sql |

---

## 🔧 使用方法

### 安装迁移工具

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### 执行迁移

```bash
# 向上迁移（创建表）
migrate -path database/migrations -database "postgresql://moviepilot:moviepilot@localhost:5432/moviepilot?sslmode=disable" up

# 向下迁移（删除表）
migrate -path database/migrations -database "postgresql://moviepilot:moviepilot@localhost:5432/moviepilot?sslmode=disable" down

# 迁移到指定版本
migrate -path database/migrations -database "postgresql://moviepilot:moviepilot@localhost:5432/moviepilot?sslmode=disable" goto 5

# 查看当前版本
migrate -path database/migrations -database "postgresql://moviepilot:moviepilot@localhost:5432/moviepilot?sslmode=disable" version
```

### 使用 Makefile

```bash
# 向上迁移
make migrate-up

# 向下迁移
make migrate-down

# 重置数据库
make migrate-reset
```

---

## 📊 表结构说明

### 1. users 表

用户基本信息表。

**字段**:
- `id`: 主键
- `username`: 用户名（唯一）
- `email`: 邮箱（唯一）
- `password_hash`: 密码哈希（bcrypt）
- `nickname`: 昵称
- `avatar`: 头像URL
- `status`: 状态（active/disabled/locked）
- `last_login_at`: 最后登录时间
- `last_login_ip`: 最后登录IP
- `created_at`: 创建时间
- `updated_at`: 更新时间
- `deleted_at`: 删除时间（软删除）

**索引**:
- `idx_users_username`
- `idx_users_email`
- `idx_users_status`
- `idx_users_deleted_at`

### 2. roles 表

角色定义表。

**字段**:
- `id`: 主键
- `name`: 角色名称（唯一）
- `display_name`: 显示名称
- `description`: 描述
- `is_system`: 是否系统角色
- `created_at`: 创建时间
- `updated_at`: 更新时间

**预定义角色**:
- `admin`: 管理员
- `user`: 普通用户
- `guest`: 访客

### 3. permissions 表

权限定义表。

**字段**:
- `id`: 主键
- `name`: 权限名称（唯一，格式：action:resource）
- `resource`: 资源名称
- `action`: 操作名称
- `description`: 描述
- `created_at`: 创建时间

**权限命名规范**:
- `read:subscribe` - 查看订阅
- `write:subscribe` - 创建/修改订阅
- `delete:subscribe` - 删除订阅
- `manage:user` - 管理用户
- `manage:site` - 管理站点

### 4. user_roles 表

用户角色关联表（多对多）。

**字段**:
- `user_id`: 用户ID（外键）
- `role_id`: 角色ID（外键）
- `created_at`: 创建时间

**主键**: (user_id, role_id)

### 5. role_permissions 表

角色权限关联表（多对多）。

**字段**:
- `role_id`: 角色ID（外键）
- `permission_id`: 权限ID（外键）
- `created_at`: 创建时间

**主键**: (role_id, permission_id)

### 6. auth_logs 表

认证日志表（审计）。

**字段**:
- `id`: 主键
- `user_id`: 用户ID（外键，可为空）
- `action`: 操作类型（login/logout/register等）
- `ip_address`: IP地址
- `user_agent`: User-Agent
- `status`: 状态（success/failed）
- `error_message`: 错误信息
- `created_at`: 创建时间

### 7. sites 表

站点配置表。

**字段**:
- `id`: 主键
- `user_id`: 所属用户ID（外键）
- `name`: 站点名称
- `url`: 站点URL
- `type`: 站点类型（pt/public/rss）
- `priority`: 优先级（1-10）
- `enabled`: 是否启用
- `cookie`: Cookie
- `user_agent`: User-Agent
- `proxy`: 代理地址
- `checkin_enabled`: 是否启用签到
- `checkin_cron`: 签到Cron表达式
- `checkin_url`: 签到URL
- `upload`: 上传量（字节）
- `download`: 下载量（字节）
- `ratio`: 分享率
- `status`: 状态（active/error/disabled）
- `last_checkin`: 最后签到时间
- `last_sync`: 最后同步时间
- `error_message`: 错误信息
- `created_at`: 创建时间
- `updated_at`: 更新时间
- `deleted_at`: 删除时间

### 8. site_cookies 表

Cookie历史表。

**字段**:
- `id`: 主键
- `site_id`: 站点ID（外键）
- `cookie`: Cookie内容
- `is_valid`: 是否有效
- `expires_at`: 过期时间
- `created_at`: 创建时间

### 9. checkin_logs 表

签到日志表。

**字段**:
- `id`: 主键
- `site_id`: 站点ID（外键）
- `success`: 是否成功
- `message`: 消息
- `bonus`: 获得的魔力值/积分
- `continue_days`: 连续签到天数
- `error_message`: 错误信息
- `checkin_time`: 签到时间

### 10. site_stats 表

流量统计表。

**字段**:
- `id`: 主键
- `site_id`: 站点ID（外键）
- `date`: 日期
- `upload_delta`: 当天上传增量
- `download_delta`: 当天下载增量
- `upload_total`: 总上传
- `download_total`: 总下载
- `ratio`: 分享率
- `created_at`: 创建时间

**唯一约束**: (site_id, date)

### 11. sync_logs 表

同步日志表。

**字段**:
- `id`: 主键
- `site_id`: 站点ID（外键）
- `success`: 是否成功
- `duration_ms`: 同步耗时（毫秒）
- `error_message`: 错误信息
- `synced_at`: 同步时间

---

## 🌱 初始数据

### 种子数据文件

1. `001_insert_default_roles.sql` - 插入默认角色
2. `002_insert_default_permissions.sql` - 插入默认权限
3. `003_insert_role_permissions.sql` - 分配角色权限
4. `004_insert_default_admin.sql` - 创建默认管理员

### 执行种子数据

```bash
# 按顺序执行
psql -d moviepilot -f database/seeds/001_insert_default_roles.sql
psql -d moviepilot -f database/seeds/002_insert_default_permissions.sql
psql -d moviepilot -f database/seeds/003_insert_role_permissions.sql
psql -d moviepilot -f database/seeds/004_insert_default_admin.sql
```

---

## ✅ 验证

### 检查表是否创建成功

```sql
-- 查看所有表
\dt

-- 查看表结构
\d users
\d roles
\d permissions

-- 查看索引
\di

-- 查看外键约束
SELECT
    tc.table_name, 
    kcu.column_name, 
    ccu.table_name AS foreign_table_name,
    ccu.column_name AS foreign_column_name 
FROM 
    information_schema.table_constraints AS tc 
    JOIN information_schema.key_column_usage AS kcu
      ON tc.constraint_name = kcu.constraint_name
    JOIN information_schema.constraint_column_usage AS ccu
      ON ccu.constraint_name = tc.constraint_name
WHERE tc.constraint_type = 'FOREIGN KEY';
```

### 检查初始数据

```sql
-- 查看角色
SELECT * FROM roles;

-- 查看权限
SELECT * FROM permissions;

-- 查看角色权限关联
SELECT r.name as role, p.name as permission
FROM roles r
JOIN role_permissions rp ON r.id = rp.role_id
JOIN permissions p ON p.id = rp.permission_id
ORDER BY r.name, p.name;

-- 查看默认管理员
SELECT * FROM users WHERE username = 'admin';
```

---

## 🔄 回滚

如果需要回滚迁移：

```bash
# 回滚最后一次迁移
migrate -path database/migrations -database "postgresql://..." down 1

# 回滚所有迁移
migrate -path database/migrations -database "postgresql://..." down

# 回滚到指定版本
migrate -path database/migrations -database "postgresql://..." goto 5
```

---

## 📝 注意事项

1. **备份数据**
   - 在生产环境执行迁移前，务必备份数据库
   - 使用 `pg_dump` 创建备份

2. **测试迁移**
   - 先在测试环境验证迁移脚本
   - 确保可以正常执行和回滚

3. **外键约束**
   - 注意表的创建顺序
   - 先创建被引用的表

4. **索引优化**
   - 根据实际查询需求调整索引
   - 避免创建过多索引

5. **数据一致性**
   - 使用事务保证数据一致性
   - 迁移失败时自动回滚

---

**文档状态**: ✅ 完成  
**迁移脚本数量**: 22 个（11 up + 11 down）  
**种子数据脚本**: 4 个
