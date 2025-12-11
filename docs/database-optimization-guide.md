# 数据库优化指南

> **执行计划**: Week 4 Day 1-2  
> **目标**: 完成数据库层性能优化

---

## 📋 任务清单

### Day 1 任务

- [x] 创建索引优化模块 (`pkg/database/indexes.go`)
- [x] 创建连接池优化模块 (`pkg/database/optimization.go`)
- [x] 创建数据库优化脚本 (`scripts/db_optimize.go`)
- [ ] 分析慢查询日志
- [ ] 创建优化索引
- [ ] 测试索引效果

### Day 2 任务

- [ ] 优化连接池配置
- [ ] 编写性能基准测试
- [ ] 执行性能测试
- [ ] 生成性能报告

---

## 🚀 快速开始

### 1. 分析当前数据库状态

```bash
cd /workspaces/moviepilot/moviepilot-go-project/moviepilot-go

# 分析数据库（连接池、表大小、索引使用情况）
go run scripts/db_optimize.go -action=analyze \
  -host=localhost \
  -port=5432 \
  -user=postgres \
  -password=your_password \
  -dbname=moviepilot
```

**输出示例**：
```
=== 开始数据库分析 ===

--- 连接池统计 ---
max_open_connections: 25
open_connections: 5
in_use: 2
idle: 3
utilization_percentage: 8.00

--- 表大小统计 ---
表: subscribes, 总大小: 2.5 MB, 表大小: 1.8 MB, 索引大小: 700 KB
表: downloadhistories, 总大小: 15 MB, 表大小: 12 MB, 索引大小: 3 MB
...

--- 索引使用情况 ---
未使用的索引: subscribes.idx_old_index (大小: 500 KB)
...
```

---

### 2. 启用慢查询日志

```bash
# 启用 pg_stat_statements 扩展
psql -U postgres -d moviepilot -c "CREATE EXTENSION IF NOT EXISTS pg_stat_statements;"

# 分析慢查询（平均执行时间 > 100ms）
go run scripts/db_optimize.go -action=slow-queries \
  -host=localhost \
  -dbname=moviepilot
```

---

### 3. 创建优化索引

```bash
# 创建所有优化索引
go run scripts/db_optimize.go -action=create-indexes \
  -host=localhost \
  -dbname=moviepilot
```

**创建的索引**：

| 表名 | 索引名 | 字段 | 类型 |
|------|--------|------|------|
| subscribes | idx_subscribes_user_state_updated | (username, state, updated_at DESC) | 复合索引 |
| subscribes | idx_subscribes_tmdb_type | (tmdb_id, type) | 部分索引 |
| subscribes | idx_subscribes_state_last_update | (state, last_update DESC) | 部分索引 |
| downloadhistories | idx_download_history_hash_state | (download_hash, state) | 复合索引 |
| downloadhistories | idx_download_history_user_date | (username, date DESC) | 复合索引 |
| transfer_histories | idx_transfer_history_src_status | (src, status) | 复合索引 |
| sites | idx_sites_active_pri | (is_active, pri DESC) | 部分索引 |
| downloads | idx_downloads_status_created | (status, created_at DESC) | 复合索引 |

---

### 4. 优化连接池配置

```bash
# 应用生产环境连接池配置
go run scripts/db_optimize.go -action=optimize \
  -host=localhost \
  -dbname=moviepilot
```

**优化配置**：
- `MaxOpenConns`: 25 → **100**
- `MaxIdleConns`: 5 → **10**
- `ConnMaxLifetime`: 1小时
- `ConnMaxIdleTime`: 10分钟

---

### 5. 执行 VACUUM ANALYZE

```bash
# 清理表并更新统计信息
go run scripts/db_optimize.go -action=vacuum \
  -host=localhost \
  -dbname=moviepilot
```

---

## 📊 性能测试

### 运行基准测试

```bash
# 运行所有基准测试
go test ./internal/repositories/benchmarks -bench=. -benchmem

# 运行特定测试
go test ./internal/repositories/benchmarks -bench=BenchmarkSubscribeRepository_List -benchmem

# 生成性能报告
go test ./internal/repositories/benchmarks -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof
```

**预期结果**：

| 操作 | 优化前 | 优化后 | 提升 |
|------|--------|--------|------|
| List (1000条) | 150 ms | < 50 ms | 3x |
| GetByID | 15 ms | < 5 ms | 3x |
| ListByState | 200 ms | < 80 ms | 2.5x |
| BatchCreate (100条) | 500 ms | < 200 ms | 2.5x |

---

## 🔍 索引设计说明

### 1. Subscribe 表索引

#### idx_subscribes_user_state_updated
```sql
CREATE INDEX idx_subscribes_user_state_updated 
ON subscribes(username, state, updated_at DESC)
```
**用途**: 用户订阅列表查询（最常用）
**查询示例**:
```sql
SELECT * FROM subscribes 
WHERE username = 'user1' AND state = 'R' 
ORDER BY updated_at DESC;
```

#### idx_subscribes_state_last_update
```sql
CREATE INDEX idx_subscribes_state_last_update 
ON subscribes(state, last_update DESC) 
WHERE state IN ('R', 'N')
```
**用途**: 订阅刷新任务（只索引活跃订阅）
**查询示例**:
```sql
SELECT * FROM subscribes 
WHERE state = 'R' 
ORDER BY last_update ASC 
LIMIT 100;
```

---

### 2. DownloadHistory 表索引

#### idx_download_history_hash_state
```sql
CREATE INDEX idx_download_history_hash_state 
ON downloadhistories(download_hash, state)
```
**用途**: 下载状态查询
**查询示例**:
```sql
SELECT * FROM downloadhistories 
WHERE download_hash = 'abc123' AND state = 'completed';
```

---

### 3. TransferHistory 表索引

#### idx_transfer_history_src_status
```sql
CREATE INDEX idx_transfer_history_src_status 
ON transfer_histories(src, status)
```
**用途**: 文件转移历史查询
**查询示例**:
```sql
SELECT * FROM transfer_histories 
WHERE src = '/path/to/file' AND status = true;
```

---

## 📈 性能监控

### 1. 实时监控连接池

```go
// 在应用启动时启动监控
stop := make(chan struct{})
go database.MonitorConnectionPool(db, 30*time.Second, stop)

// 应用关闭时停止监控
close(stop)
```

### 2. 查看索引使用情况

```sql
-- 查看索引扫描次数
SELECT 
    schemaname,
    tablename,
    indexname,
    idx_scan as scans,
    idx_tup_read as tuples_read,
    pg_size_pretty(pg_relation_size(indexrelid)) as size
FROM pg_stat_user_indexes
WHERE schemaname = 'public'
ORDER BY idx_scan ASC;
```

### 3. 查看慢查询

```sql
-- 需要先启用 pg_stat_statements
SELECT 
    query,
    calls,
    mean_exec_time,
    max_exec_time
FROM pg_stat_statements
WHERE mean_exec_time > 100
ORDER BY mean_exec_time DESC
LIMIT 20;
```

---

## ✅ 验收标准

### Day 1 完成标准
- [x] 索引优化模块创建完成
- [ ] 慢查询分析完成（识别Top 10慢查询）
- [ ] 优化索引创建完成（20+个索引）
- [ ] 索引使用情况分析完成

### Day 2 完成标准
- [ ] 连接池配置优化完成
- [ ] 性能基准测试编写完成
- [ ] 性能测试报告生成
- [ ] 所有核心查询响应时间 < 100ms
- [ ] 连接池利用率 > 70%

---

## 🚨 注意事项

### 1. 索引创建时机
- ⚠️ 在**低峰期**创建索引（避免锁表）
- ⚠️ 大表创建索引可能需要较长时间
- ⚠️ 使用 `CONCURRENTLY` 选项避免阻塞（PostgreSQL）

### 2. 连接池配置
- ⚠️ `MaxOpenConns` 不要设置过大（避免数据库过载）
- ⚠️ 根据实际负载调整参数
- ⚠️ 监控连接池利用率

### 3. VACUUM 操作
- ⚠️ 定期执行 VACUUM ANALYZE
- ⚠️ 大表 VACUUM 可能耗时较长
- ⚠️ 考虑使用 autovacuum

---

## 📚 参考资料

- [PostgreSQL 索引优化](https://www.postgresql.org/docs/current/indexes.html)
- [GORM 性能优化](https://gorm.io/docs/performance.html)
- [Go 数据库连接池](https://go.dev/doc/database/manage-connections)

---

**最后更新**: 2025-12-02  
**负责人**: 数据库优化小组
