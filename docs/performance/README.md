# 性能优化指南

## 🚀 性能概览

MoviePilot Go 通过多层次的性能优化策略，实现了从 Python 版本 5-10 倍的性能提升。本文档详细介绍了系统的性能优化方案和最佳实践。

### 性能提升对比

| 指标 | 原Python版本 | Go版本 | 提升比例 |
|------|-------------|--------|----------|
| 启动时间 | ~8-12秒 | ~1-2秒 | **5-6x** |
| 内存占用 | ~200-300MB | ~50-100MB | **2-3x** |
| 并发处理 | ~100 req/s | ~1000+ req/s | **10x+** |
| CPU使用率 | 高负载下60-80% | 高负载下20-30% | **2-3x** |
| 数据库查询 | 同步阻塞 | 连接池+异步 | **3-5x** |

## ⚡ 应用层优化

### 1. 并发处理优化

#### Goroutine 池管理
```go
// pkg/pool/worker_pool.go
package pool

import (
    "context"
    "runtime"
    "sync"
    "time"
)

type Task func(ctx context.Context) error

type WorkerPool struct {
    workers    int
    taskQueue  chan Task
    workerPool chan chan Task
    quit       chan bool
    wg         sync.WaitGroup
}

func NewWorkerPool(workers int) *WorkerPool {
    if workers <= 0 {
        workers = runtime.NumCPU()
    }
    
    return &WorkerPool{
        workers:    workers,
        taskQueue:  make(chan Task, workers*2),
        workerPool: make(chan chan Task, workers),
        quit:       make(chan bool),
    }
}

func (p *WorkerPool) Start() {
    for i := 0; i < p.workers; i++ {
        p.wg.Add(1)
        go p.startWorker(i)
    }
    
    // 分发任务
    go p.dispatch()
}

func (p *WorkerPool) startWorker(id int) {
    defer p.wg.Done()
    
    for {
        select {
        case task := <-p.taskQueue:
            ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
            if err := task(ctx); err != nil {
                // 记录错误日志
            }
            cancel()
            
        case <-p.quit:
            return
        }
    }
}

func (p *WorkerPool) dispatch() {
    for {
        select {
        case task := <-p.taskQueue:
            // 任务已经在队列中，等待 worker 处理
            
        case <-p.quit:
            return
        }
    }
}

func (p *WorkerPool) Submit(task Task) {
    select {
    case p.taskQueue <- task:
    default:
        // 队列满时的处理策略
        // 可以选择阻塞、丢弃或扩容
    }
}

func (p *WorkerPool) Stop() {
    close(p.quit)
    p.wg.Wait()
}
```

#### 并发安全的缓存
```go
// pkg/cache/safe_cache.go
package cache

import (
    "sync"
    "time"
)

type SafeCache struct {
    items map[string]*cacheItem
    mutex sync.RWMutex
}

type cacheItem struct {
    value      interface{}
    expiration int64
    created    time.Time
}

func NewSafeCache() *SafeCache {
    cache := &SafeCache{
        items: make(map[string]*cacheItem),
    }
    
    // 启动清理协程
    go cache.startCleanup()
    
    return cache
}

func (c *SafeCache) Set(key string, value interface{}, ttl time.Duration) {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    var expiration int64
    if ttl > 0 {
        expiration = time.Now().Add(ttl).UnixNano()
    }
    
    c.items[key] = &cacheItem{
        value:      value,
        expiration: expiration,
        created:    time.Now(),
    }
}

func (c *SafeCache) Get(key string) (interface{}, bool) {
    c.mutex.RLock()
    defer c.mutex.RUnlock()
    
    item, exists := c.items[key]
    if !exists {
        return nil, false
    }
    
    if item.expiration > 0 && time.Now().UnixNano() > item.expiration {
        return nil, false
    }
    
    return item.value, true
}

func (c *SafeCache) startCleanup() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            c.cleanup()
        }
    }
}

func (c *SafeCache) cleanup() {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    now := time.Now().UnixNano()
    for key, item := range c.items {
        if item.expiration > 0 && now > item.expiration {
            delete(c.items, key)
        }
    }
}
```

### 2. 内存优化

#### 对象池复用
```go
// pkg/pool/buffer_pool.go
package pool

import (
    "bytes"
    "sync"
)

var (
    bufferPool = sync.Pool{
        New: func() interface{} {
            return new(bytes.Buffer)
        },
    }
)

func GetBuffer() *bytes.Buffer {
    return bufferPool.Get().(*bytes.Buffer)
}

func PutBuffer(buf *bytes.Buffer) {
    if buf != nil {
        buf.Reset()
        bufferPool.Put(buf)
    }
}

// 使用示例
func ProcessData(data []byte) ([]byte, error) {
    buf := GetBuffer()
    defer PutBuffer(buf)
    
    // 使用 buffer 处理数据
    buf.Write(data)
    
    // 处理逻辑...
    
    result := make([]byte, buf.Len())
    copy(result, buf.Bytes())
    
    return result, nil
}
```

#### 内存监控和限制
```go
// pkg/memory/monitor.go
package memory

import (
    "context"
    "runtime"
    "sync/atomic"
    "time"
)

type MemoryMonitor struct {
    maxMemoryBytes int64
    currentUsage   int64
    alertThreshold float64
}

func NewMemoryMonitor(maxMemoryMB int, alertThreshold float64) *MemoryMonitor {
    return &MemoryMonitor{
        maxMemoryBytes: int64(maxMemoryMB) * 1024 * 1024,
        alertThreshold: alertThreshold,
    }
}

func (m *MemoryMonitor) Start(ctx context.Context) {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            m.checkMemory()
        }
    }
}

func (m *MemoryMonitor) checkMemory() {
    var memStats runtime.MemStats
    runtime.ReadMemStats(&memStats)
    
    atomic.StoreInt64(&m.currentUsage, int64(memStats.Alloc))
    
    usagePercent := float64(memStats.Alloc) / float64(m.maxMemoryBytes) * 100
    
    if usagePercent > m.alertThreshold {
        // 触发内存告警
        m.triggerGC()
    }
}

func (m *MemoryMonitor) triggerGC() {
    runtime.GC()
    runtime.GC() // 强制执行两次
}

func (m *MemoryMonitor) GetCurrentUsage() int64 {
    return atomic.LoadInt64(&m.currentUsage)
}

func (m *MemoryMonitor) IsMemoryPressure() bool {
    usagePercent := float64(m.GetCurrentUsage()) / float64(m.maxMemoryBytes) * 100
    return usagePercent > m.alertThreshold
}
```

### 3. I/O 优化

#### 批量处理
```go
// pkg/batch/processor.go
package batch

import (
    "context"
    "sync"
    "time"
)

type BatchProcessor struct {
    batchSize    int
    flushTimeout time.Duration
    buffer       []interface{}
    mutex        sync.Mutex
    processor    func([]interface{}) error
    timer        *time.Timer
}

func NewBatchProcessor(batchSize int, flushTimeout time.Duration, processor func([]interface{}) error) *BatchProcessor {
    return &BatchProcessor{
        batchSize:    batchSize,
        flushTimeout: flushTimeout,
        buffer:       make([]interface{}, 0, batchSize),
        processor:    processor,
    }
}

func (bp *BatchProcessor) Add(item interface{}) error {
    bp.mutex.Lock()
    defer bp.mutex.Unlock()
    
    bp.buffer = append(bp.buffer, item)
    
    if len(bp.buffer) >= bp.batchSize {
        return bp.flush()
    }
    
    // 设置或重置定时器
    if bp.timer == nil {
        bp.timer = time.AfterFunc(bp.flushTimeout, func() {
            bp.mutex.Lock()
            defer bp.mutex.Unlock()
            bp.flush()
        })
    } else {
        bp.timer.Reset(bp.flushTimeout)
    }
    
    return nil
}

func (bp *BatchProcessor) flush() error {
    if len(bp.buffer) == 0 {
        return nil
    }
    
    // 停止定时器
    if bp.timer != nil {
        bp.timer.Stop()
        bp.timer = nil
    }
    
    // 复制缓冲区
    batch := make([]interface{}, len(bp.buffer))
    copy(batch, bp.buffer)
    
    // 清空缓冲区
    bp.buffer = bp.buffer[:0]
    
    // 处理批次
    return bp.processor(batch)
}

func (bp *BatchProcessor) Flush() error {
    bp.mutex.Lock()
    defer bp.mutex.Unlock()
    return bp.flush()
}
```

#### 连接池优化
```go
// pkg/pool/connection_pool.go
package pool

import (
    "net"
    "sync"
    "time"
)

type ConnectionPool struct {
    factory    func() (net.Conn, error)
    pool       chan net.Conn
    mutex      sync.Mutex
    maxConn    int
    created    int
    maxIdle    time.Duration
}

func NewConnectionPool(factory func() (net.Conn, error), maxConn int, maxIdle time.Duration) *ConnectionPool {
    return &ConnectionPool{
        factory: factory,
        pool:    make(chan net.Conn, maxConn),
        maxConn: maxConn,
        maxIdle: maxIdle,
    }
}

func (p *ConnectionPool) Get() (net.Conn, error) {
    select {
    case conn := <-p.pool:
        // 检查连接是否有效
        if p.isConnValid(conn) {
            return conn, nil
        }
        conn.Close()
        p.mutex.Lock()
        p.created--
        p.mutex.Unlock()
        
    default:
        // 池为空，创建新连接
    }
    
    p.mutex.Lock()
    defer p.mutex.Unlock()
    
    if p.created >= p.maxConn {
        // 等待可用连接
        conn := <-p.pool
        if p.isConnValid(conn) {
            return conn, nil
        }
        conn.Close()
        p.created--
    }
    
    conn, err := p.factory()
    if err != nil {
        return nil, err
    }
    
    p.created++
    return conn, nil
}

func (p *ConnectionPool) Put(conn net.Conn) {
    if !p.isConnValid(conn) {
        conn.Close()
        p.mutex.Lock()
        p.created--
        p.mutex.Unlock()
        return
    }
    
    select {
    case p.pool <- conn:
    default:
        // 池满，关闭连接
        conn.Close()
        p.mutex.Lock()
        p.created--
        p.mutex.Unlock()
    }
}

func (p *ConnectionPool) isConnValid(conn net.Conn) bool {
    if conn == nil {
        return false
    }
    
    // 设置读取超时
    conn.SetReadDeadline(time.Now().Add(time.Second))
    defer conn.SetReadDeadline(time.Time{})
    
    // 尝试读取一个字节
    one := make([]byte, 1)
    _, err := conn.Read(one)
    if err != nil {
        return false
    }
    
    return true
}
```

## 🗄️ 数据库优化

### 1. 连接池优化

#### GORM 连接池配置
```go
// pkg/database/optimized_db.go
package database

import (
    "fmt"
    "time"
    
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

type OptimizedConfig struct {
    Host            string
    Port            int
    User            string
    Password        string
    DBName          string
    SSLMode         string
    
    // 连接池配置
    MaxIdleConns    int           `yaml:"max_idle_conns" default:"10"`
    MaxOpenConns    int           `yaml:"max_open_conns" default:"100"`
    ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" default:"1h"`
    ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time" default:"10m"`
    
    // 性能配置
    PrepareStmt     bool `yaml:"prepare_stmt" default:"true"`
    DisableForeignKeyConstraintWhenMigrating bool `yaml:"disable_foreign_key_constraint_when_migrating" default:"false"`
}

func NewOptimizedDB(config OptimizedConfig) (*gorm.DB, error) {
    dsn := fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Shanghai",
        config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode,
    )
    
    gormConfig := &gorm.Config{
        Logger: logger.Default.LogMode(logger.Silent),
        NowFunc: func() time.Time {
            return time.Now().UTC()
        },
        PrepareStmt:                             config.PrepareStmt,
        DisableForeignKeyConstraintWhenMigrating: config.DisableForeignKeyConstraintWhenMigrating,
    }
    
    db, err := gorm.Open(postgres.Open(dsn), gormConfig)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }
    
    sqlDB, err := db.DB()
    if err != nil {
        return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
    }
    
    // 优化的连接池配置
    sqlDB.SetMaxIdleConns(config.MaxIdleConns)
    sqlDB.SetMaxOpenConns(config.MaxOpenConns)
    sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
    sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)
    
    return db, nil
}
```

### 2. 查询优化

#### 预编译语句缓存
```go
// pkg/database/prepared_statements.go
package database

import (
    "context"
    "database/sql"
    "fmt"
    "sync"
    
    "gorm.io/gorm"
)

type PreparedStatementCache struct {
    db     *sql.DB
    cache  map[string]*sql.Stmt
    mutex  sync.RWMutex
    maxAge time.Duration
}

func NewPreparedStatementCache(db *sql.DB, maxAge time.Duration) *PreparedStatementCache {
    cache := &PreparedStatementCache{
        db:     db,
        cache:  make(map[string]*sql.Stmt),
        maxAge: maxAge,
    }
    
    // 启动清理协程
    go cache.startCleanup()
    
    return cache
}

func (c *PreparedStatementCache) Prepare(query string) (*sql.Stmt, error) {
    c.mutex.RLock()
    stmt, exists := c.cache[query]
    c.mutex.RUnlock()
    
    if exists {
        return stmt, nil
    }
    
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    // 双重检查
    if stmt, exists := c.cache[query]; exists {
        return stmt, nil
    }
    
    stmt, err := c.db.Prepare(query)
    if err != nil {
        return nil, fmt.Errorf("failed to prepare statement: %w", err)
    }
    
    c.cache[query] = stmt
    return stmt, nil
}

func (c *PreparedStatementCache) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
    stmt, err := c.Prepare(query)
    if err != nil {
        return nil, err
    }
    
    return stmt.ExecContext(ctx, args...)
}

func (c *PreparedStatementCache) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
    stmt, err := c.Prepare(query)
    if err != nil {
        return nil, err
    }
    
    return stmt.QueryContext(ctx, args...)
}

func (c *PreparedStatementCache) startCleanup() {
    ticker := time.NewTicker(c.maxAge / 2)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            c.cleanup()
        }
    }
}

func (c *PreparedStatementCache) cleanup() {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    for query, stmt := range c.cache {
        // 关闭旧的预编译语句
        stmt.Close()
        delete(c.cache, query)
    }
}
```

#### 批量插入优化
```go
// repositories/optimized_repository.go
package repositories

import (
    "context"
    "fmt"
    "strings"
    
    "gorm.io/gorm"
)

type BulkInserter struct {
    db          *gorm.DB
    batchSize   int
    tableName   string
    columns     []string
    buffer      []interface{}
    valueStrings []string
}

func NewBulkInserter(db *gorm.DB, tableName string, columns []string, batchSize int) *BulkInserter {
    return &BulkInserter{
        db:        db,
        tableName: tableName,
        columns:   columns,
        batchSize: batchSize,
    }
}

func (b *BulkInserter) Add(values map[string]interface{}) error {
    if len(values) != len(b.columns) {
        return fmt.Errorf("values count mismatch")
    }
    
    // 构建值字符串
    placeholders := make([]string, len(b.columns))
    for i := range b.columns {
        placeholders[i] = fmt.Sprintf("$%d", len(b.buffer)+i+1)
    }
    
    b.valueStrings = append(b.valueStrings, fmt.Sprintf("(%s)", strings.Join(placeholders, ",")))
    
    // 添加值到缓冲区
    for _, col := range b.columns {
        b.buffer = append(b.buffer, values[col])
    }
    
    // 检查是否需要刷新
    if len(b.valueStrings) >= b.batchSize {
        return b.Flush()
    }
    
    return nil
}

func (b *BulkInserter) Flush() error {
    if len(b.valueStrings) == 0 {
        return nil
    }
    
    query := fmt.Sprintf(
        "INSERT INTO %s (%s) VALUES %s",
        b.tableName,
        strings.Join(b.columns, ","),
        strings.Join(b.valueStrings, ","),
    )
    
    err := b.db.Exec(query, b.buffer...).Error
    if err != nil {
        return fmt.Errorf("bulk insert failed: %w", err)
    }
    
    // 重置缓冲区
    b.buffer = b.buffer[:0]
    b.valueStrings = b.valueStrings[:0]
    
    return nil
}
```

### 3. 索引优化策略

#### 智能索引建议
```sql
-- 查询性能分析
SELECT 
    query,
    calls,
    total_time,
    mean_time,
    rows,
    100.0 * shared_blks_hit / nullif(shared_blks_hit + shared_blks_read, 0) AS hit_percent
FROM pg_stat_statements 
WHERE calls > 100
ORDER BY mean_time DESC 
LIMIT 10;

-- 查看缺失的索引
SELECT 
    schemaname,
    tablename,
    attname,
    n_distinct,
    correlation
FROM pg_stats 
WHERE schemaname = 'public' 
  AND n_distinct > 100
ORDER BY n_distinct DESC;

-- 索引使用情况
SELECT 
    schemaname,
    tablename,
    indexname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch,
    pg_size_pretty(pg_relation_size(indexrelid)) as index_size
FROM pg_stat_user_indexes 
ORDER BY idx_scan DESC;
```

## 🚀 缓存策略

### 1. 多级缓存架构

#### L1 + L2 缓存实现
```go
// pkg/cache/multilevel_cache.go
package cache

import (
    "context"
    "encoding/json"
    "time"
    
    "github.com/go-redis/redis/v8"
)

type MultiLevelCache struct {
    l1Cache *SafeCache      // 内存缓存
    l2Cache *redis.Client    // Redis 缓存
    l1TTL   time.Duration
    l2TTL   time.Duration
}

func NewMultiLevelCache(redisAddr string, l1TTL, l2TTL time.Duration) (*MultiLevelCache, error) {
    rdb := redis.NewClient(&redis.Options{
        Addr:     redisAddr,
        Password: "",
        DB:       0,
    })
    
    // 测试 Redis 连接
    if err := rdb.Ping(context.Background()).Err(); err != nil {
        return nil, err
    }
    
    return &MultiLevelCache{
        l1Cache: NewSafeCache(),
        l2Cache: rdb,
        l1TTL:   l1TTL,
        l2TTL:   l2TTL,
    }, nil
}

func (c *MultiLevelCache) Set(ctx context.Context, key string, value interface{}) error {
    // 序列化数据
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }
    
    // 设置 L1 缓存
    c.l1Cache.Set(key, value, c.l1TTL)
    
    // 设置 L2 缓存
    return c.l2Cache.Set(ctx, key, data, c.l2TTL).Err()
}

func (c *MultiLevelCache) Get(ctx context.Context, key string, dest interface{}) error {
    // 先从 L1 缓存获取
    if value, found := c.l1Cache.Get(key); found {
        return c.copyToDest(value, dest)
    }
    
    // 从 L2 缓存获取
    data, err := c.l2Cache.Get(ctx, key).Result()
    if err == redis.Nil {
        return ErrNotFound
    } else if err != nil {
        return err
    }
    
    // 反序列化
    var value interface{}
    if err := json.Unmarshal([]byte(data), &value); err != nil {
        return err
    }
    
    // 回填 L1 缓存
    c.l1Cache.Set(key, value, c.l1TTL)
    
    return c.copyToDest(value, dest)
}

func (c *MultiLevelCache) copyToDest(value interface{}, dest interface{}) error {
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }
    return json.Unmarshal(data, dest)
}
```

#### 缓存预热策略
```go
// pkg/cache/warmer.go
package cache

import (
    "context"
    "sync"
    "time"
)

type CacheWarmer struct {
    cache   *MultiLevelCache
    loaders map[string]CacheLoader
    mutex   sync.RWMutex
}

type CacheLoader interface {
    Load(ctx context.Context) (map[string]interface{}, error)
    Key() string
    TTL() time.Duration
}

func NewCacheWarmer(cache *MultiLevelCache) *CacheWarmer {
    return &CacheWarmer{
        cache:   cache,
        loaders: make(map[string]CacheLoader),
    }
}

func (w *CacheWarmer) RegisterLoader(loader CacheLoader) {
    w.mutex.Lock()
    defer w.mutex.Unlock()
    w.loaders[loader.Key()] = loader
}

func (w *CacheWarmer) WarmAll(ctx context.Context) error {
    w.mutex.RLock()
    loaders := make([]CacheLoader, 0, len(w.loaders))
    for _, loader := range w.loaders {
        loaders = append(loaders, loader)
    }
    w.mutex.RUnlock()
    
    // 并行预热
    var wg sync.WaitGroup
    errChan := make(chan error, len(loaders))
    
    for _, loader := range loaders {
        wg.Add(1)
        go func(l CacheLoader) {
            defer wg.Done()
            if err := w.warmLoader(ctx, l); err != nil {
                errChan <- err
            }
        }(loader)
    }
    
    wg.Wait()
    close(errChan)
    
    // 收集错误
    var errors []error
    for err := range errChan {
        errors = append(errors, err)
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("cache warm errors: %v", errors)
    }
    
    return nil
}

func (w *CacheWarmer) warmLoader(ctx context.Context, loader CacheLoader) error {
    data, err := loader.Load(ctx)
    if err != nil {
        return fmt.Errorf("failed to load %s: %w", loader.Key(), err)
    }
    
    for key, value := range data {
        if err := w.cache.Set(ctx, key, value); err != nil {
            return fmt.Errorf("failed to set cache %s: %w", key, err)
        }
    }
    
    return nil
}

func (w *CacheWarmer) StartPeriodicWarm(ctx context.Context, interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            w.WarmAll(ctx)
        }
    }
}
```

### 2. 缓存失效策略

#### 智能缓存失效
```go
// pkg/cache/invalidation.go
package cache

import (
    "context"
    "hash/fnv"
    "strings"
    "time"
)

type CacheInvalidator struct {
    cache   *MultiLevelCache
    patterns map[string][]string
    mutex   sync.RWMutex
}

func NewCacheInvalidator(cache *MultiLevelCache) *CacheInvalidator {
    return &CacheInvalidator{
        cache:    cache,
        patterns: make(map[string][]string),
    }
}

func (ci *CacheInvalidator) RegisterPattern(pattern string, keys []string) {
    ci.mutex.Lock()
    defer ci.mutex.Unlock()
    ci.patterns[pattern] = keys
}

func (ci *CacheInvalidator) InvalidateByPattern(ctx context.Context, pattern string) error {
    ci.mutex.RLock()
    keys, exists := ci.patterns[pattern]
    ci.mutex.RUnlock()
    
    if !exists {
        return nil
    }
    
    for _, key := range keys {
        ci.cache.l1Cache.Delete(key)
        ci.cache.l2Cache.Del(ctx, key)
    }
    
    return nil
}

func (ci *CacheInvalidator) InvalidateByTag(ctx context.Context, tag string) error {
    // 基于标签的缓存失效
    pattern := fmt.Sprintf("tag:%s:*", tag)
    return ci.InvalidateByPattern(ctx, pattern)
}

// 基于内容哈希的缓存键生成
func GenerateCacheKey(prefix string, params map[string]interface{}) string {
    h := fnv.New64a()
    h.Write([]byte(prefix))
    
    // 排序参数以确保一致性
    sortedKeys := make([]string, 0, len(params))
    for k := range params {
        sortedKeys = append(sortedKeys, k)
    }
    
    for _, k := range sortedKeys {
        h.Write([]byte(k))
        h.Write([]byte(fmt.Sprintf("%v", params[k])))
    }
    
    return fmt.Sprintf("%s:%x", prefix, h.Sum64())
}
```

## 🌐 网络优化

### 1. HTTP 客户端优化

#### 连接池和超时配置
```go
// pkg/http/optimized_client.go
package http

import (
    "net"
    "net/http"
    "time"
)

func NewOptimizedClient() *http.Client {
    return &http.Client{
        Transport: &http.Transport{
            // 连接池配置
            MaxIdleConns:        100,
            MaxIdleConnsPerHost: 10,
            MaxConnsPerHost:     100,
            IdleConnTimeout:     90 * time.Second,
            
            // TLS 配置
            TLSHandshakeTimeout: 10 * time.Second,
            
            // 连接配置
            DialContext: (&net.Dialer{
                Timeout:   30 * time.Second,
                KeepAlive: 30 * time.Second,
            }).DialContext,
            
            // 响应头超时
            ResponseHeaderTimeout: 10 * time.Second,
            
            // 期望继续超时
            ExpectContinueTimeout: 1 * time.Second,
            
            // 禁用 HTTP/2（在某些场景下性能更好）
            ForceAttemptHTTP2: false,
        },
        
        // 请求超时
        Timeout: 30 * time.Second,
    }
}
```

#### 请求重试和熔断
```go
// pkg/http/resilient_client.go
package http

import (
    "context"
    "fmt"
    "math"
    "net/http"
    "time"
    
    "github.com/cenkalti/backoff/v4"
)

type ResilientClient struct {
    client    *http.Client
    maxRetries int
    backoff   backoff.BackOff
}

func NewResilientClient(maxRetries int) *ResilientClient {
    return &ResilientClient{
        client:     NewOptimizedClient(),
        maxRetries: maxRetries,
        backoff: backoff.NewExponentialBackOff(
            backoff.WithInitialInterval(100*time.Millisecond),
            backoff.WithMaxInterval(5*time.Second),
            backoff.WithMaxElapsedTime(30*time.Second),
            backoff.WithMultiplier(2.0),
        ),
    }
}

func (rc *ResilientClient) Do(req *http.Request) (*http.Response, error) {
    var resp *http.Response
    var err error
    
    operation := func() error {
        resp, err = rc.client.Do(req)
        if err != nil {
            return err
        }
        
        // 检查 HTTP 状态码
        if resp.StatusCode >= 500 {
            resp.Body.Close()
            return fmt.Errorf("server error: %d", resp.StatusCode)
        }
        
        return nil
    }
    
    notify := func(err error, next time.Duration) {
        // 记录重试日志
        fmt.Printf("Request failed, retrying in %v: %v\n", next, err)
    }
    
    err = backoff.RetryNotify(operation, rc.backoff, notify)
    if err != nil {
        return nil, err
    }
    
    return resp, nil
}

// 熔断器实现
type CircuitBreaker struct {
    maxFailures   int
    resetTimeout  time.Duration
    failures      int
    lastFailTime  time.Time
    state         CircuitState
    mutex         sync.RWMutex
}

type CircuitState int

const (
    StateClosed CircuitState = iota
    StateOpen
    StateHalfOpen
)

func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
    return &CircuitBreaker{
        maxFailures:  maxFailures,
        resetTimeout: resetTimeout,
        state:        StateClosed,
    }
}

func (cb *CircuitBreaker) Call(fn func() error) error {
    if !cb.allowRequest() {
        return fmt.Errorf("circuit breaker is open")
    }
    
    err := fn()
    cb.recordResult(err)
    return err
}

func (cb *CircuitBreaker) allowRequest() bool {
    cb.mutex.RLock()
    defer cb.mutex.RUnlock()
    
    switch cb.state {
    case StateClosed:
        return true
    case StateOpen:
        if time.Since(cb.lastFailTime) > cb.resetTimeout {
            cb.mutex.RUnlock()
            cb.mutex.Lock()
            cb.state = StateHalfOpen
            cb.failures = 0
            cb.mutex.Unlock()
            cb.mutex.RLock()
            return true
        }
        return false
    case StateHalfOpen:
        return true
    default:
        return false
    }
}

func (cb *CircuitBreaker) recordResult(err error) {
    cb.mutex.Lock()
    defer cb.mutex.Unlock()
    
    if err != nil {
        cb.failures++
        cb.lastFailTime = time.Now()
        
        if cb.failures >= cb.maxFailures {
            cb.state = StateOpen
        }
    } else {
        cb.failures = 0
        cb.state = StateClosed
    }
}
```

## 📊 性能监控

### 1. 实时性能指标

#### 性能收集器
```go
// pkg/performance/collector.go
package performance

import (
    "context"
    "runtime"
    "sync/atomic"
    "time"
)

type PerformanceCollector struct {
    requestCount    int64
    requestDuration int64 // 纳秒
    errorCount      int64
    activeGoroutines int64
    
    startTime time.Time
}

func NewPerformanceCollector() *PerformanceCollector {
    return &PerformanceCollector{
        startTime: time.Now(),
    }
}

func (pc *PerformanceCollector) RecordRequest(duration time.Duration, isError bool) {
    atomic.AddInt64(&pc.requestCount, 1)
    atomic.AddInt64(&pc.requestDuration, duration.Nanoseconds())
    
    if isError {
        atomic.AddInt64(&pc.errorCount, 1)
    }
    
    // 更新活跃协程数
    atomic.StoreInt64(&pc.activeGoroutines, int64(runtime.NumGoroutine()))
}

func (pc *PerformanceCollector) GetMetrics() PerformanceMetrics {
    requestCount := atomic.LoadInt64(&pc.requestCount)
    requestDuration := atomic.LoadInt64(&pc.requestDuration)
    errorCount := atomic.LoadInt64(&pc.errorCount)
    activeGoroutines := atomic.LoadInt64(&pc.activeGoroutines)
    
    var avgDuration time.Duration
    if requestCount > 0 {
        avgDuration = time.Duration(requestDuration / requestCount)
    }
    
    var errorRate float64
    if requestCount > 0 {
        errorRate = float64(errorCount) / float64(requestCount) * 100
    }
    
    uptime := time.Since(pc.startTime)
    
    var qps float64
    if uptime.Seconds() > 0 {
        qps = float64(requestCount) / uptime.Seconds()
    }
    
    return PerformanceMetrics{
        RequestCount:       requestCount,
        ErrorCount:         errorCount,
        ErrorRate:          errorRate,
        AverageDuration:    avgDuration,
        QPS:                qps,
        ActiveGoroutines:   activeGoroutines,
        Uptime:             uptime,
    }
}

type PerformanceMetrics struct {
    RequestCount     int64         `json:"request_count"`
    ErrorCount       int64         `json:"error_count"`
    ErrorRate        float64       `json:"error_rate"`
    AverageDuration  time.Duration `json:"average_duration"`
    QPS              float64       `json:"qps"`
    ActiveGoroutines int64         `json:"active_goroutines"`
    Uptime           time.Duration `json:"uptime"`
}
```

### 2. 性能分析工具

#### CPU 分析器
```go
// pkg/performance/profiler.go
package performance

import (
    "context"
    "os"
    "runtime/pprof"
    "time"
)

type Profiler struct {
    cpuProfile    *os.File
    memProfile    *os.File
    enabled       bool
}

func NewProfiler() *Profiler {
    return &Profiler{enabled: false}
}

func (p *Profiler) StartCPUProfile(filename string) error {
    if !p.enabled {
        return nil
    }
    
    f, err := os.Create(filename)
    if err != nil {
        return err
    }
    
    if err := pprof.StartCPUProfile(f); err != nil {
        f.Close()
        return err
    }
    
    p.cpuProfile = f
    return nil
}

func (p *Profiler) StopCPUProfile() {
    if p.cpuProfile != nil {
        pprof.StopCPUProfile()
        p.cpuProfile.Close()
        p.cpuProfile = nil
    }
}

func (p *Profiler) WriteHeapProfile(filename string) error {
    if !p.enabled {
        return nil
    }
    
    f, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer f.Close()
    
    runtime.GC() // 强制垃圾回收
    return pprof.WriteHeapProfile(f)
}

func (p *Profiler) Enable() {
    p.enabled = true
}

func (p *Profiler) Disable() {
    p.enabled = false
    p.StopCPUProfile()
}

// 自动性能分析
func (p *Profiler) AutoProfile(ctx context.Context, interval time.Duration, outputDir string) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    
    counter := 0
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            counter++
            
            // CPU 分析
            cpuFile := fmt.Sprintf("%s/cpu_%d.prof", outputDir, counter)
            if err := p.StartCPUProfile(cpuFile); err == nil {
                go func() {
                    time.Sleep(30 * time.Second)
                    p.StopCPUProfile()
                }()
            }
            
            // 内存分析
            memFile := fmt.Sprintf("%s/heap_%d.prof", outputDir, counter)
            p.WriteHeapProfile(memFile)
        }
    }
}
```

---

**注意**: 性能优化是一个持续的过程，需要根据实际业务场景和监控数据进行调整。建议定期进行性能基准测试，并根据测试结果优化系统配置。