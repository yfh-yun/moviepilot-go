# 故障排查指南

## 🔍 故障排查概览

本文档提供了 MoviePilot Go 系统常见问题的诊断和解决方案，涵盖应用、数据库、网络、性能等各个方面。

## 🚨 紧急故障处理

### 1. 应用完全无响应

#### 症状
- 所有 HTTP 请求超时
- 无法连接到应用端口
- 容器状态异常

#### 排查步骤

```bash
# 1. 检查应用状态
docker ps | grep moviepilot
docker logs moviepilot-app --tail=100

# 2. 检查系统资源
top -p $(pgrep moviepilot)
free -h
df -h

# 3. 检查端口占用
netstat -tulpn | grep 3001
ss -tulpn | grep 3001

# 4. 检查进程状态
ps aux | grep moviepilot
kill -0 $(pgrep moviepilot) && echo "Process is running" || echo "Process is dead"
```

#### 解决方案

```bash
# 重启应用
docker restart moviepilot-app

# 如果容器无法启动，检查配置
docker run --rm -it moviepilot-go:latest /bin/sh

# 检查配置文件
cat /app/configs/config.yaml

# 重新部署
docker-compose down
docker-compose up -d
```

### 2. 数据库连接失败

#### 症状
- 应用日志显示数据库连接错误
- API 请求返回 500 错误
- 数据库相关功能不可用

#### 排查步骤

```bash
# 1. 检查数据库状态
docker ps | grep postgres
docker logs moviepilot-postgres --tail=50

# 2. 测试数据库连接
docker exec -it moviepilot-postgres psql -U moviepilot -d moviepilot -c "SELECT 1;"

# 3. 检查网络连接
docker exec moviepilot-app ping postgres
docker exec moviepilot-app telnet postgres 5432

# 4. 检查连接池状态
curl http://localhost:3001/metrics | grep db_connections
```

#### 解决方案

```sql
-- 检查数据库连接数
SELECT count(*) FROM pg_stat_activity;

-- 检查慢查询
SELECT query, mean_time, calls 
FROM pg_stat_statements 
ORDER BY mean_time DESC 
LIMIT 10;

-- 重置连接池
SELECT pg_terminate_backend(pid) 
FROM pg_stat_activity 
WHERE state = 'idle' 
  AND query_start < now() - interval '5 minutes';
```

```bash
# 重启数据库
docker restart moviepilot-postgres

# 调整连接池配置
# 编辑 configs/config.yaml
database:
  max_idle_conns: 20
  max_open_conns: 100
  conn_max_lifetime: "1h"
```

## 📊 性能问题排查

### 1. 响应时间过长

#### 症状
- API 响应时间超过 5 秒
- 用户体验明显下降
- 监控显示高延迟

#### 排查步骤

```bash
# 1. 检查系统资源
docker stats moviepilot-app
top -p $(pgrep moviepilot)

# 2. 分析慢查询
curl http://localhost:3001/metrics | grep http_request_duration

# 3. 检查数据库性能
docker exec moviepilot-postgres psql -U moviepilot -d moviepilot -c "
SELECT query, mean_time, calls 
FROM pg_stat_statements 
WHERE mean_time > 1000 
ORDER BY mean_time DESC 
LIMIT 5;"

# 4. 生成性能分析
curl http://localhost:3001/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof cpu.prof
```

#### 优化方案

```sql
-- 添加缺失索引
CREATE INDEX CONCURRENTLY idx_media_title_gin ON media USING gin(title gin_trgm_ops);
CREATE INDEX CONCURRENTLY idx_transfers_status_created ON transfers(status, created_at DESC);

-- 分析查询计划
EXPLAIN ANALYZE SELECT * FROM media WHERE title ILIKE '%keyword%' ORDER BY created_at DESC LIMIT 20;

-- 更新表统计信息
ANALYZE media;
ANALYZE transfers;
```

```go
// 优化查询示例
func (r *mediaRepository) SearchMediaOptimized(ctx context.Context, query string, limit int) ([]*models.Media, error) {
    var media []*models.Media
    
    // 使用预编译语句
    stmt := `
        SELECT id, title, type, year, rating, poster_url, created_at
        FROM media 
        WHERE search_vector @@ plainto_tsquery('english', $1)
        ORDER BY created_at DESC 
        LIMIT $2
    `
    
    err := r.db.WithContext(ctx).
        Raw(stmt, query, limit).
        Scan(&media).Error
    
    return media, err
}
```

### 2. 内存使用过高

#### 症状
- 容器内存使用率超过 90%
- 频繁的 OOM (Out of Memory) 错误
- 系统响应缓慢

#### 排查步骤

```bash
# 1. 检查内存使用
docker stats moviepilot-app --no-stream
free -h

# 2. 分析内存分布
curl http://localhost:3001/debug/pprof/heap > heap.prof
go tool pprof heap.prof

# 3. 检查 Goroutine 数量
curl http://localhost:3001/debug/pprof/goroutine?debug=1

# 4. 监控内存增长
watch -n 5 'docker stats moviepilot-app --no-stream'
```

#### 解决方案

```go
// 内存优化配置
func configureRuntime() {
    // 设置 GOMAXPROCS
    runtime.GOMAXPROCS(runtime.NumCPU())
    
    // 设置 GC 目标
    debug.SetGCPercent(100) // 默认值，可根据需要调整
    
    // 设置内存限制
    debug.SetMemoryLimit(1 << 30) // 1GB
}

// 对象池优化
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 0, 1024)
    },
}

func getBuffer() []byte {
    return bufferPool.Get().([]byte)[:0]
}

func putBuffer(buf []byte) {
    if cap(buf) <= 64*1024 { // 避免保留过大的缓冲区
        bufferPool.Put(buf)
    }
}
```

```bash
# 调整容器内存限制
docker-compose.yml:
services:
  app:
    deploy:
      resources:
        limits:
          memory: 1G
        reservations:
          memory: 512M
```

## 🌐 网络问题排查

### 1. 服务间通信失败

#### 症状
- gRPC 连接超时
- 插件服务无法访问
- 服务发现失败

#### 排查步骤

```bash
# 1. 检查网络连通性
docker exec moviepilot-app ping moviepilot-plugins
docker exec moviepilot-app telnet moviepilot-plugins 5000

# 2. 检查端口监听
netstat -tulpn | grep :5000
ss -tulpn | grep :3001

# 3. 检查 DNS 解析
docker exec moviepilot-app nslookup moviepilot-plugins
docker exec moviepilot-app cat /etc/resolv.conf

# 4. 检查防火墙规则
iptables -L -n
ufw status
```

#### 解决方案

```yaml
# Docker Compose 网络配置
version: '3.8'
services:
  app:
    networks:
      - moviepilot-network
    depends_on:
      - plugins
      
  plugins:
    networks:
      - moviepilot-network

networks:
  moviepilot-network:
    driver: bridge
    ipam:
      config:
        - subnet: 172.20.0.0/16
```

```go
// gRPC 连接配置优化
func NewGRPCClient(address string) (*grpc.ClientConn, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    conn, err := grpc.DialContext(ctx, address,
        grpc.WithTransportCredentials(insecure.NewCredentials()),
        grpc.WithBlock(),
        grpc.WithKeepaliveParams(keepalive.ClientParameters{
            Time:                10 * time.Second,
            Timeout:             3 * time.Second,
            PermitWithoutStream: true,
        }),
    )
    
    return conn, err
}
```

### 2. 外部 API 调用失败

#### 症状
- 第三方服务集成失败
- HTTP 客户端超时
- 代理服务器错误

#### 排查步骤

```bash
# 1. 测试网络连接
curl -v https://api.themoviedb.org/3/movie/550?api_key=YOUR_KEY
curl -I https://example.com

# 2. 检查 DNS 解析
nslookup api.themoviedb.org
dig api.themoviedb.org

# 3. 检查代理设置
echo $HTTP_PROXY
echo $HTTPS_PROXY

# 4. 测试 SSL 连接
openssl s_client -connect api.themoviedb.org:443 -servername api.themoviedb.org
```

#### 解决方案

```go
// HTTP 客户端配置优化
func NewHTTPClient() *http.Client {
    return &http.Client{
        Timeout: 30 * time.Second,
        Transport: &http.Transport{
            DialContext: (&net.Dialer{
                Timeout:   30 * time.Second,
                KeepAlive: 30 * time.Second,
            }).DialContext,
            MaxIdleConns:        100,
            MaxIdleConnsPerHost: 10,
            IdleConnTimeout:     90 * time.Second,
            TLSHandshakeTimeout: 10 * time.Second,
            ExpectContinueTimeout: 1 * time.Second,
        },
    }
}

// 重试机制
func WithRetry(client *http.Client, req *http.Request, maxRetries int) (*http.Response, error) {
    var resp *http.Response
    var err error
    
    for i := 0; i <= maxRetries; i++ {
        if i > 0 {
            time.Sleep(time.Duration(i) * time.Second)
        }
        
        resp, err = client.Do(req)
        if err == nil && resp.StatusCode < 500 {
            return resp, nil
        }
        
        if resp != nil {
            resp.Body.Close()
        }
    }
    
    return nil, err
}
```

## 🔧 插件问题排查

### 1. 插件加载失败

#### 症状
- 插件列表为空
- 插件状态显示错误
- 插件功能不可用

#### 排查步骤

```bash
# 1. 检查插件服务状态
docker ps | grep moviepilot-plugins
docker logs moviepilot-plugins --tail=50

# 2. 检查插件配置
curl http://localhost:5000/plugins
docker exec moviepilot-plugins ls -la /app/plugins/

# 3. 检查 gRPC 连接
docker exec moviepilot-app telnet moviepilot-plugins 5000

# 4. 验证插件文件
docker exec moviepilot-plugins python -c "
import json
with open('/app/plugins/example/plugin.json') as f:
    config = json.load(f)
    print(json.dumps(config, indent=2))
"
```

#### 解决方案

```python
# 插件配置验证
def validate_plugin_config(plugin_path):
    config_file = os.path.join(plugin_path, 'plugin.json')
    
    if not os.path.exists(config_file):
        raise FileNotFoundError(f"plugin.json not found in {plugin_path}")
    
    with open(config_file, 'r') as f:
        config = json.load(f)
    
    required_fields = ['id', 'name', 'version', 'type']
    for field in required_fields:
        if field not in config:
            raise ValueError(f"Missing required field: {field}")
    
    return config

# 插件依赖检查
def check_plugin_dependencies(plugin_path):
    requirements_file = os.path.join(plugin_path, 'requirements.txt')
    
    if os.path.exists(requirements_file):
        with open(requirements_file, 'r') as f:
            requirements = f.read().splitlines()
        
        for req in requirements:
            try:
                pkg_resources.require(req)
            except pkg_resources.DistributionNotFound:
                print(f"Installing missing dependency: {req}")
                subprocess.check_call(['pip', 'install', req])
```

```bash
# 重新安装插件依赖
docker exec moviepilot-plugins pip install -r /app/plugins/example/requirements.txt

# 重启插件服务
docker restart moviepilot-plugins

# 验证插件加载
curl http://localhost:3001/api/v1/plugins
```

### 2. 插件执行错误

#### 症状
- 插件方法调用失败
- 返回错误结果
- 插件进程崩溃

#### 排查步骤

```bash
# 1. 查看插件日志
docker logs moviepilot-plugins --tail=100 | grep ERROR

# 2. 测试插件接口
grpcurl -plaintext -d '{"plugin_id":"example","method":"search","params":"{\"keyword\":\"test\"}"}' \
  localhost:5000 plugin.PluginService.ExecutePlugin

# 3. 检查插件权限
docker exec moviepilot-plugins ls -la /app/plugins/example/
docker exec moviepilot-plugins python -c "
import sys
sys.path.append('/app/plugins/example')
try:
    import main
    print('Plugin module loaded successfully')
except Exception as e:
    print(f'Plugin load error: {e}')
"
```

#### 解决方案

```python
# 错误处理装饰器
def handle_plugin_errors(func):
    def wrapper(*args, **kwargs):
        try:
            return func(*args, **kwargs)
        except Exception as e:
            logger.error(f"Plugin error in {func.__name__}: {e}", exc_info=True)
            return {
                "success": False,
                "error": str(e),
                "error_type": type(e).__name__
            }
    return wrapper

# 插件健康检查
def health_check(plugin_id):
    try:
        # 测试插件基本功能
        result = execute_plugin_method(plugin_id, "health_check", {})
        return result.get("success", False)
    except Exception as e:
        logger.error(f"Health check failed for {plugin_id}: {e}")
        return False
```

## 📝 日志分析

### 1. 日志收集和分析

#### 常用日志查询

```bash
# 查看应用错误日志
docker logs moviepilot-app 2>&1 | grep ERROR | tail -20

# 查看 HTTP 请求日志
docker logs moviepilot-app | grep "HTTP" | tail -20

# 查看数据库查询日志
docker logs moviepilot-app | grep "SELECT\|INSERT\|UPDATE\|DELETE" | tail -20

# 查看插件相关日志
docker logs moviepilot-plugins 2>&1 | grep -E "(ERROR|WARN)" | tail -20

# 实时监控日志
docker logs -f moviepilot-app | grep -E "(ERROR|WARN|500)"
```

#### 日志分析工具

```bash
# 使用 jq 分析 JSON 日志
docker logs moviepilot-app | jq 'select(.level=="error")' | tail -10

# 统计错误类型
docker logs moviepilot-app | jq -r '.level' | sort | uniq -c

# 分析响应时间
docker logs moviepilot-app | jq -r 'select(.duration) | .duration' | awk '{sum+=$1; count++} END {print "Average:", sum/count}'
```

### 2. 结构化日志解析

```go
// 日志结构定义
type LogEntry struct {
    Timestamp time.Time              `json:"timestamp"`
    Level     string                 `json:"level"`
    Message   string                 `json:"message"`
    Module    string                 `json:"module"`
    RequestID string                 `json:"request_id,omitempty"`
    UserID    string                 `json:"user_id,omitempty"`
    Error     string                 `json:"error,omitempty"`
    Duration  string                 `json:"duration,omitempty"`
    Fields    map[string]interface{} `json:"fields,omitempty"`
}

// 日志分析器
type LogAnalyzer struct {
    entries []LogEntry
}

func (la *LogAnalyzer) ParseLogs(logData []byte) error {
    lines := strings.Split(string(logData), "\n")
    
    for _, line := range lines {
        if line == "" {
            continue
        }
        
        var entry LogEntry
        if err := json.Unmarshal([]byte(line), &entry); err != nil {
            continue
        }
        
        la.entries = append(la.entries, entry)
    }
    
    return nil
}

func (la *LogAnalyzer) GetErrorLogs() []LogEntry {
    var errors []LogEntry
    for _, entry := range la.entries {
        if entry.Level == "error" {
            errors = append(errors, entry)
        }
    }
    return errors
}

func (la *LogAnalyzer) GetSlowRequests(threshold time.Duration) []LogEntry {
    var slowRequests []LogEntry
    for _, entry := range la.entries {
        if entry.Duration != "" {
            if duration, err := time.ParseDuration(entry.Duration); err == nil {
                if duration > threshold {
                    slowRequests = append(slowRequests, entry)
                }
            }
        }
    }
    return slowRequests
}
```

## 🔧 调试工具

### 1. Go 调试

#### pprof 使用

```bash
# CPU 分析
curl http://localhost:3001/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof cpu.prof

# 内存分析
curl http://localhost:3001/debug/pprof/heap > heap.prof
go tool pprof heap.prof

# Goroutine 分析
curl http://localhost:3001/debug/pprof/goroutine > goroutine.prof
go tool pprof goroutine.prof

# 阻塞分析
curl http://localhost:3001/debug/pprof/block > block.prof
go tool pprof block.prof
```

#### Delve 调试器

```bash
# 安装 Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# 调试运行中的应用
dlv attach $(pgrep moviepilot)
dlv connect localhost:4000

# 调试测试
dlv test ./internal/services/

# 常用调试命令
(dlv) break main.go:50
(dlv) continue
(dlv) locals
(dlv) print variable
(dlv) stack
```

### 2. 数据库调试

#### 查询分析

```sql
-- 启用查询日志
ALTER SYSTEM SET log_statement = 'all';
ALTER SYSTEM SET log_min_duration_statement = 1000;  -- 记录超过1秒的查询
SELECT pg_reload_conf();

-- 查看活跃查询
SELECT pid, now() - pg_stat_activity.query_start AS duration, query 
FROM pg_stat_activity 
WHERE state = 'active' 
  AND now() - pg_stat_activity.query_start > interval '5 seconds';

-- 查看锁等待
SELECT blocked_locks.pid AS blocked_pid,
       blocked_activity.usename AS blocked_user,
       blocking_locks.pid AS blocking_pid,
       blocking_activity.usename AS blocking_user,
       blocked_activity.query AS blocked_statement,
       blocking_activity.query AS current_statement_in_blocking_process
FROM pg_catalog.pg_locks blocked_locks
JOIN pg_catalog.pg_stat_activity blocked_activity ON blocked_activity.pid = blocked_locks.pid
JOIN pg_catalog.pg_locks blocking_locks ON blocking_locks.locktype = blocked_locks.locktype
JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid
WHERE NOT blocked_locks.granted;
```

#### 性能调优

```sql
-- 查看表统计信息
SELECT schemaname, tablename, n_tup_ins, n_tup_upd, n_tup_del, n_live_tup, n_dead_tup
FROM pg_stat_user_tables
ORDER BY n_live_tup DESC;

-- 查看索引使用情况
SELECT schemaname, tablename, indexname, idx_scan, idx_tup_read, idx_tup_fetch
FROM pg_stat_user_indexes
ORDER BY idx_scan DESC;

-- 自动清理配置
SHOW autovacuum;
SHOW autovacuum_analyze_scale_factor;
SHOW autovacuum_vacuum_scale_factor;
```

## 📋 故障排查清单

### 应用层检查
- [ ] 应用进程是否运行
- [ ] 端口是否正常监听
- [ ] 健康检查是否通过
- [ ] 日志是否有错误信息
- [ ] 内存和 CPU 使用率是否正常
- [ ] 响应时间是否在可接受范围

### 数据库层检查
- [ ] 数据库服务是否运行
- [ ] 连接池是否正常
- [ ] 慢查询是否存在
- [ ] 索引是否合理
- [ ] 锁等待是否存在
- [ ] 磁盘空间是否充足

### 网络层检查
- [ ] 服务间网络是否通畅
- [ ] DNS 解析是否正常
- [ ] 防火墙规则是否正确
- [ ] 负载均衡是否正常
- [ ] SSL 证书是否有效
- [ ] 代理配置是否正确

### 插件系统检查
- [ ] 插件服务是否运行
- [ ] gRPC 连接是否正常
- [ ] 插件配置是否正确
- [ ] 插件依赖是否满足
- [ ] 插件权限是否足够
- [ ] 插件日志是否有错误

---

**注意**: 在处理生产环境问题时，请务必在维护窗口期间进行操作，并确保有完整的备份和回滚方案。对于复杂问题，建议收集足够的日志和监控数据后再进行分析。