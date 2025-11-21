# 监控系统文档

## 📊 监控概览

MoviePilot Go 采用多层次监控架构，提供全方位的系统可观测性，包括指标监控、日志聚合、链路追踪和告警管理。

### 监控架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Application    │    │   Prometheus    │    │     Grafana     │
│   (Metrics)     │───►│   (Storage)     │───►│  (Visualization)│
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Application    │    │   Loki/ELK      │    │   AlertManager  │
│   (Logs)         │───►│   (Log Storage) │───►│  (Alerting)     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Application    │    │    Jaeger       │    │   Webhook/Slack │
│   (Traces)       │───►│  (Trace Storage)│───►│  (Notifications)│
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 📈 指标监控

### 1. Prometheus 配置

#### prometheus.yml
```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "rules/*.yml"

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093

scrape_configs:
  # MoviePilot 应用监控
  - job_name: 'moviepilot'
    static_configs:
      - targets: ['app:3001']
    metrics_path: '/metrics'
    scrape_interval: 5s
    scrape_timeout: 5s

  # 数据库监控
  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres-exporter:9187']
    scrape_interval: 10s

  # Redis 监控
  - job_name: 'redis'
    static_configs:
      - targets: ['redis-exporter:9121']
    scrape_interval: 10s

  # 容器监控
  - job_name: 'cadvisor'
    static_configs:
      - targets: ['cadvisor:8080']
    scrape_interval: 10s

  # Node 监控
  - job_name: 'node'
    static_configs:
      - targets: ['node-exporter:9100']
    scrape_interval: 10s
```

### 2. 应用指标实现

#### metrics/metrics.go
```go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // HTTP 请求计数器
    httpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )

    // HTTP 请求持续时间
    httpRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint"},
    )

    // 数据库连接池
    dbConnectionsActive = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "db_connections_active",
            Help: "Number of active database connections",
        },
    )

    dbConnectionsIdle = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "db_connections_idle",
            Help: "Number of idle database connections",
        },
    )

    // 传输任务计数器
    transferTasksTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "transfer_tasks_total",
            Help: "Total number of transfer tasks",
        },
        []string{"status", "source_type"},
    )

    // 传输任务持续时间
    transferTaskDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "transfer_task_duration_seconds",
            Help:    "Transfer task duration in seconds",
            Buckets: []float64{1, 5, 10, 30, 60, 300, 600, 1800, 3600},
        },
        []string{"source_type"},
    )

    // 媒体文件计数
    mediaFilesTotal = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "media_files_total",
            Help: "Total number of media files",
        },
        []string{"type", "status"},
    )

    // 插件状态
    pluginStatus = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "plugin_status",
            Help: "Plugin status (1=enabled, 0=disabled)",
        },
        []string{"plugin_id", "plugin_type"},
    )

    // 系统资源使用
    systemCPUUsage = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "system_cpu_usage_percent",
            Help: "System CPU usage percentage",
        },
    )

    systemMemoryUsage = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "system_memory_usage_bytes",
            Help: "System memory usage in bytes",
        },
    )

    systemDiskUsage = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "system_disk_usage_bytes",
            Help: "System disk usage in bytes",
        },
        []string{"mount_point"},
    )
)

// HTTP 指标记录函数
func RecordHTTPRequest(method, endpoint, status string) {
    httpRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
}

func RecordHTTPRequestDuration(method, endpoint string, duration float64) {
    httpRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
}

// 数据库指标记录函数
func RecordDBConnections(active, idle int) {
    dbConnectionsActive.Set(float64(active))
    dbConnectionsIdle.Set(float64(idle))
}

// 传输任务指标记录函数
func RecordTransferTask(status, sourceType string) {
    transferTasksTotal.WithLabelValues(status, sourceType).Inc()
}

func RecordTransferTaskDuration(sourceType string, duration float64) {
    transferTaskDuration.WithLabelValues(sourceType).Observe(duration)
}

// 媒体文件指标记录函数
func RecordMediaFilesTotal(mediaType, status string, count float64) {
    mediaFilesTotal.WithLabelValues(mediaType, status).Set(count)
}

// 插件状态记录函数
func RecordPluginStatus(pluginID, pluginType string, enabled bool) {
    value := 0.0
    if enabled {
        value = 1.0
    }
    pluginStatus.WithLabelValues(pluginID, pluginType).Set(value)
}

// 系统资源指标记录函数
func RecordSystemCPUUsage(usage float64) {
    systemCPUUsage.Set(usage)
}

func RecordSystemMemoryUsage(usage int64) {
    systemMemoryUsage.Set(float64(usage))
}

func RecordSystemDiskUsage(mountPoint string, usage int64) {
    systemDiskUsage.WithLabelValues(mountPoint).Set(float64(usage))
}
```

#### 中间件集成
```go
// internal/apis/middlewares/metrics.go
package middlewares

import (
    "strconv"
    "time"
    
    "github.com/gin-gonic/gin"
    "moviepilot-go/pkg/metrics"
)

func MetricsMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        // 处理请求
        c.Next()
        
        // 计算持续时间
        duration := time.Since(start).Seconds()
        
        // 记录指标
        method := c.Request.Method
        endpoint := c.FullPath()
        status := strconv.Itoa(c.Writer.Status())
        
        metrics.RecordHTTPRequest(method, endpoint, status)
        metrics.RecordHTTPRequestDuration(method, endpoint, duration)
    }
}
```

### 3. 自定义监控

#### 系统监控
```go
// internal/monitor/system_monitor.go
package monitor

import (
    "context"
    "runtime"
    "time"
    
    "moviepilot-go/pkg/logger"
    "moviepilot-go/pkg/metrics"
)

type SystemMonitor struct {
    logger logger.Logger
}

func NewSystemMonitor(logger logger.Logger) *SystemMonitor {
    return &SystemMonitor{logger: logger}
}

func (m *SystemMonitor) Start(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            m.collectMetrics()
        }
    }
}

func (m *SystemMonitor) collectMetrics() {
    // CPU 使用率
    cpuUsage := m.getCPUUsage()
    metrics.RecordSystemCPUUsage(cpuUsage)
    
    // 内存使用
    var memStats runtime.MemStats
    runtime.ReadMemStats(&memStats)
    metrics.RecordSystemMemoryUsage(int64(memStats.Alloc))
    
    // 磁盘使用
    diskUsage := m.getDiskUsage()
    for mountPoint, usage := range diskUsage {
        metrics.RecordSystemDiskUsage(mountPoint, usage)
    }
    
    m.logger.Debug("System metrics collected",
        "cpu_usage", cpuUsage,
        "memory_usage", memStats.Alloc,
        "disk_usage", diskUsage,
    )
}

func (m *SystemMonitor) getCPUUsage() float64 {
    // 实现 CPU 使用率计算
    // 这里简化为模拟值
    return 25.5
}

func (m *SystemMonitor) getDiskUsage() map[string]int64 {
    // 实现磁盘使用率计算
    return map[string]int64{
        "/": 1024 * 1024 * 1024 * 100, // 100GB
        "/data": 1024 * 1024 * 1024 * 500, // 500GB
    }
}
```

#### 业务监控
```go
// internal/monitor/business_monitor.go
package monitor

import (
    "context"
    "time"
    
    "gorm.io/gorm"
    
    "moviepilot-go/internal/models"
    "moviepilot-go/pkg/logger"
    "moviepilot-go/pkg/metrics"
)

type BusinessMonitor struct {
    db     *gorm.DB
    logger logger.Logger
}

func NewBusinessMonitor(db *gorm.DB, logger logger.Logger) *BusinessMonitor {
    return &BusinessMonitor{
        db:     db,
        logger: logger,
    }
}

func (m *BusinessMonitor) Start(ctx context.Context) {
    ticker := time.NewTicker(60 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            m.collectBusinessMetrics()
        }
    }
}

func (m *BusinessMonitor) collectBusinessMetrics() {
    // 媒体文件统计
    m.collectMediaMetrics()
    
    // 传输任务统计
    m.collectTransferMetrics()
    
    // 用户活动统计
    m.collectUserMetrics()
}

func (m *BusinessMonitor) collectMediaMetrics() {
    var movieCount, tvCount, docCount int64
    
    // 统计各类型媒体数量
    m.db.Model(&models.Media{}).Where("type = ? AND status = ?", "movie", "active").Count(&movieCount)
    m.db.Model(&models.Media{}).Where("type = ? AND status = ?", "tv", "active").Count(&tvCount)
    m.db.Model(&models.Media{}).Where("type = ? AND status = ?", "documentary", "active").Count(&docCount)
    
    metrics.RecordMediaFilesTotal("movie", "active", float64(movieCount))
    metrics.RecordMediaFilesTotal("tv", "active", float64(tvCount))
    metrics.RecordMediaFilesTotal("documentary", "active", float64(docCount))
}

func (m *BusinessMonitor) collectTransferMetrics() {
    var completedCount, runningCount, failedCount int64
    
    m.db.Model(&models.Transfer{}).Where("status = ?", "completed").Count(&completedCount)
    m.db.Model(&models.Transfer{}).Where("status = ?", "running").Count(&runningCount)
    m.db.Model(&models.Transfer{}).Where("status = ?", "failed").Count(&failedCount)
    
    metrics.RecordTransferTask("completed", "all")
    metrics.RecordTransferTask("running", "all")
    metrics.RecordTransferTask("failed", "all")
}

func (m *BusinessMonitor) collectUserMetrics() {
    var activeUsers int64
    m.db.Model(&models.User{}).Where("is_active = ?", true).Count(&activeUsers)
    
    // 记录活跃用户数
    // metrics.RecordActiveUsers(float64(activeUsers))
}
```

## 📊 日志监控

### 1. 结构化日志配置

#### logger/logger.go
```go
package logger

import (
    "os"
    
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

type Logger interface {
    Debug(msg string, fields ...interface{})
    Info(msg string, fields ...interface{})
    Warn(msg string, fields ...interface{})
    Error(msg string, fields ...interface{})
    Fatal(msg string, fields ...interface{})
}

type ZapLogger struct {
    *zap.Logger
}

func NewLogger(level string) (*ZapLogger, error) {
    config := zap.NewProductionConfig()
    
    // 设置日志级别
    switch level {
    case "debug":
        config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
    case "info":
        config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
    case "warn":
        config.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
    case "error":
        config.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
    }
    
    // 设置编码格式为 JSON
    config.Encoding = "json"
    
    // 设置时间格式
    config.EncoderConfig.TimeKey = "timestamp"
    config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
    
    // 设置调用者信息
    config.EncoderConfig.CallerKey = "caller"
    config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
    
    logger, err := config.Build()
    if err != nil {
        return nil, err
    }
    
    return &ZapLogger{Logger: logger}, nil
}

func (l *ZapLogger) Debug(msg string, fields ...interface{}) {
    l.Logger.Debug(msg, l.toZapFields(fields...)...)
}

func (l *ZapLogger) Info(msg string, fields ...interface{}) {
    l.Logger.Info(msg, l.toZapFields(fields...)...)
}

func (l *ZapLogger) Warn(msg string, fields ...interface{}) {
    l.Logger.Warn(msg, l.toZapFields(fields...)...)
}

func (l *ZapLogger) Error(msg string, fields ...interface{}) {
    l.Logger.Error(msg, l.toZapFields(fields...)...)
}

func (l *ZapLogger) Fatal(msg string, fields ...interface{}) {
    l.Logger.Fatal(msg, l.toZapFields(fields...)...)
}

func (l *ZapLogger) toZapFields(fields ...interface{}) []zap.Field {
    if len(fields)%2 != 0 {
        fields = append(fields, "")
    }
    
    zapFields := make([]zap.Field, 0, len(fields)/2)
    for i := 0; i < len(fields); i += 2 {
        key, ok := fields[i].(string)
        if !ok {
            key = "unknown"
        }
        value := fields[i+1]
        zapFields = append(zapFields, zap.Any(key, value))
    }
    
    return zapFields
}
```

### 2. 日志聚合配置

#### Loki 配置 (docker-compose.yml)
```yaml
services:
  loki:
    image: grafana/loki:2.9.0
    container_name: moviepilot-loki
    restart: unless-stopped
    ports:
      - "3100:3100"
    volumes:
      - ./monitoring/loki/config.yml:/etc/loki/local-config.yaml
      - loki_data:/loki
    command: -config.file=/etc/loki/local-config.yaml
    networks:
      - moviepilot-network

  promtail:
    image: grafana/promtail:2.9.0
    container_name: moviepilot-promtail
    restart: unless-stopped
    volumes:
      - ./monitoring/promtail/config.yml:/etc/promtail/config.yml
      - /var/log:/var/log:ro
      - /var/lib/docker/containers:/var/lib/docker/containers:ro
    command: -config.file=/etc/promtail/config.yml
    networks:
      - moviepilot-network

volumes:
  loki_data:
```

#### promtail/config.yml
```yaml
server:
  http_listen_port: 9080
  grpc_listen_port: 0

positions:
  filename: /tmp/positions.yaml

clients:
  - url: http://loki:3100/loki/api/v1/push

scrape_configs:
  - job_name: containers
    static_configs:
      - targets:
          - localhost
        labels:
          job: containerlogs
          __path__: /var/lib/docker/containers/*/*log

    pipeline_stages:
      - json:
          expressions:
            output: log
            stream: stream
            attrs:
      - json:
          expressions:
            tag:
          source: attrs
      - regex:
          expression: (?P<container_name>(?:[^|]*))\|
          source: tag
      - timestamp:
          format: RFC3339Nano
          source: time
      - labels:
          stream:
          container_name:
      - output:
          source: output

  - job_name: moviepilot-app
    static_configs:
      - targets:
          - localhost
        labels:
          job: moviepilot
          __path__: /var/log/moviepilot/*.log

    pipeline_stages:
      - json:
          expressions:
            level:
            timestamp:
            message:
            caller:
            error:
      - timestamp:
          format: RFC3339Nano
          source: timestamp
      - labels:
          level:
          caller:
      - output:
          source: message
```

## 🚨 告警配置

### 1. AlertManager 配置

#### alertmanager.yml
```yaml
global:
  smtp_smarthost: 'smtp.gmail.com:587'
  smtp_from: 'alerts@moviepilot.com'
  smtp_auth_username: 'alerts@moviepilot.com'
  smtp_auth_password: 'your-app-password'

route:
  group_by: ['alertname', 'cluster', 'service']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'web.hook'
  routes:
    - match:
        severity: critical
      receiver: 'critical-alerts'
    - match:
        severity: warning
      receiver: 'warning-alerts'

receivers:
  - name: 'web.hook'
    webhook_configs:
      - url: 'http://localhost:5001/webhook'
        send_resolved: true

  - name: 'critical-alerts'
    email_configs:
      - to: 'admin@moviepilot.com'
        subject: '[CRITICAL] MoviePilot Alert'
        body: |
          {{ range .Alerts }}
          Alert: {{ .Annotations.summary }}
          Description: {{ .Annotations.description }}
          {{ end }}
    slack_configs:
      - api_url: 'YOUR_SLACK_WEBHOOK_URL'
        channel: '#alerts'
        title: 'Critical Alert'
        text: '{{ range .Alerts }}{{ .Annotations.summary }}{{ end }}'

  - name: 'warning-alerts'
    email_configs:
      - to: 'team@moviepilot.com'
        subject: '[WARNING] MoviePilot Alert'

inhibit_rules:
  - source_match:
      severity: 'critical'
    target_match:
      severity: 'warning'
    equal: ['alertname', 'cluster', 'service']
```

### 2. 告警规则

#### rules/moviepilot.yml
```yaml
groups:
  - name: moviepilot.rules
    rules:
      # 应用可用性告警
      - alert: MoviePilotDown
        expr: up{job="moviepilot"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "MoviePilot is down"
          description: "MoviePilot has been down for more than 1 minute."

      # 高错误率告警
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value }} errors per second."

      # 高响应时间告警
      - alert: HighResponseTime
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High response time detected"
          description: "95th percentile response time is {{ $value }} seconds."

      # 数据库连接告警
      - alert: DatabaseConnectionHigh
        expr: db_connections_active > 80
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High database connections"
          description: "Database has {{ $value }} active connections."

      # 磁盘空间告警
      - alert: DiskSpaceLow
        expr: (system_disk_usage_bytes / system_disk_capacity_bytes) > 0.9
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Disk space low"
          description: "Disk usage is above 90% on {{ $labels.mount_point }}."

      # 内存使用告警
      - alert: MemoryUsageHigh
        expr: (system_memory_usage_bytes / system_memory_capacity_bytes) > 0.9
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High memory usage"
          description: "Memory usage is above 90%."

      # 传输任务失败告警
      - alert: TransferTaskFailures
        expr: rate(transfer_tasks_total{status="failed"}[5m]) > 0.1
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "Transfer task failures"
          description: "Transfer task failure rate is {{ $value }} per second."

      # 插件离线告警
      - alert: PluginOffline
        expr: plugin_status == 0
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Plugin is offline"
          description: "Plugin {{ $labels.plugin_id }} is offline."
```

## 📊 Grafana 仪表板

### 1. 系统概览仪表板

#### dashboards/system-overview.json
```json
{
  "dashboard": {
    "id": null,
    "title": "MoviePilot System Overview",
    "tags": ["moviepilot", "system"],
    "timezone": "browser",
    "panels": [
      {
        "id": 1,
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total[5m])",
            "legendFormat": "{{method}} {{endpoint}}"
          }
        ],
        "yAxes": [
          {
            "label": "Requests/sec"
          }
        ]
      },
      {
        "id": 2,
        "title": "Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          },
          {
            "expr": "histogram_quantile(0.50, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "50th percentile"
          }
        ],
        "yAxes": [
          {
            "label": "Seconds"
          }
        ]
      },
      {
        "id": 3,
        "title": "Error Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total{status=~\"4..\"}[5m])",
            "legendFormat": "4xx errors"
          },
          {
            "expr": "rate(http_requests_total{status=~\"5..\"}[5m])",
            "legendFormat": "5xx errors"
          }
        ],
        "yAxes": [
          {
            "label": "Errors/sec"
          }
        ]
      },
      {
        "id": 4,
        "title": "Database Connections",
        "type": "graph",
        "targets": [
          {
            "expr": "db_connections_active",
            "legendFormat": "Active"
          },
          {
            "expr": "db_connections_idle",
            "legendFormat": "Idle"
          }
        ],
        "yAxes": [
          {
            "label": "Connections"
          }
        ]
      },
      {
        "id": 5,
        "title": "System Resources",
        "type": "graph",
        "targets": [
          {
            "expr": "system_cpu_usage_percent",
            "legendFormat": "CPU %"
          },
          {
            "expr": "system_memory_usage_bytes / 1024 / 1024 / 1024",
            "legendFormat": "Memory GB"
          }
        ],
        "yAxes": [
          {
            "label": "Percentage / GB"
          }
        ]
      },
      {
        "id": 6,
        "title": "Media Files Count",
        "type": "stat",
        "targets": [
          {
            "expr": "sum(media_files_total)",
            "legendFormat": "Total Media Files"
          }
        ]
      }
    ],
    "time": {
      "from": "now-1h",
      "to": "now"
    },
    "refresh": "30s"
  }
}
```

### 2. 业务指标仪表板

#### dashboards/business-metrics.json
```json
{
  "dashboard": {
    "id": null,
    "title": "MoviePilot Business Metrics",
    "tags": ["moviepilot", "business"],
    "panels": [
      {
        "id": 1,
        "title": "Transfer Tasks Status",
        "type": "piechart",
        "targets": [
          {
            "expr": "sum by (status) (transfer_tasks_total)",
            "legendFormat": "{{status}}"
          }
        ]
      },
      {
        "id": 2,
        "title": "Media Files by Type",
        "type": "piechart",
        "targets": [
          {
            "expr": "sum by (type) (media_files_total)",
            "legendFormat": "{{type}}"
          }
        ]
      },
      {
        "id": 3,
        "title": "Transfer Task Duration",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(transfer_task_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          }
        ],
        "yAxes": [
          {
            "label": "Seconds"
          }
        ]
      },
      {
        "id": 4,
        "title": "Plugin Status",
        "type": "table",
        "targets": [
          {
            "expr": "plugin_status",
            "legendFormat": "{{plugin_id}} - {{plugin_type}}"
          },
          {
            "expr": "plugin_status == 1",
            "legendFormat": "Enabled"
          }
        ]
      }
    ],
    "time": {
      "from": "now-24h",
      "to": "now"
    },
    "refresh": "1m"
  }
}
```

## 🔍 链路追踪

### 1. OpenTelemetry 配置

#### tracing/tracer.go
```go
package tracing

import (
    "context"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/jaeger"
    "go.opentelemetry.io/otel/sdk/resource"
    "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func InitTracer(serviceName, jaegerURL string) error {
    // 创建 Jaeger exporter
    exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerURL)))
    if err != nil {
        return err
    }
    
    // 创建资源
    res, err := resource.New(context.Background(),
        resource.WithAttributes(
            semconv.ServiceNameKey.String(serviceName),
        ),
    )
    if err != nil {
        return err
    }
    
    // 创建 TracerProvider
    tp := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
        trace.WithResource(res),
    )
    
    // 注册为全局 TracerProvider
    otel.SetTracerProvider(tp)
    
    return nil
}
```

#### 中间件集成
```go
// internal/apis/middlewares/tracing.go
package middlewares

import (
    "net/http"
    
    "github.com/gin-gonic/gin"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/trace"
)

func TracingMiddleware(serviceName string) gin.HandlerFunc {
    tracer := otel.Tracer(serviceName)
    
    return func(c *gin.Context) {
        ctx, span := tracer.Start(c.Request.Context(), c.Request.URL.Path)
        defer span.End()
        
        // 设置属性
        span.SetAttributes(
            attribute.String("http.method", c.Request.Method),
            attribute.String("http.url", c.Request.URL.String()),
            attribute.String("http.host", c.Request.Host),
            attribute.String("http.user_agent", c.Request.UserAgent()),
        )
        
        // 更新上下文
        c.Request = c.Request.WithContext(ctx)
        
        // 处理请求
        c.Next()
        
        // 设置响应属性
        span.SetAttributes(
            attribute.Int("http.status_code", c.Writer.Status()),
        )
    }
}
```

## 🔧 监控部署

### 1. Docker Compose 完整配置

#### monitoring/docker-compose.monitoring.yml
```yaml
version: '3.8'

services:
  prometheus:
    image: prom/prometheus:v2.45.0
    container_name: moviepilot-prometheus
    restart: unless-stopped
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus/config.yml:/etc/prometheus/prometheus.yml
      - ./prometheus/rules:/etc/prometheus/rules
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
      - '--storage.tsdb.retention.time=30d'
      - '--web.enable-lifecycle'
    networks:
      - moviepilot-network

  alertmanager:
    image: prom/alertmanager:v0.25.0
    container_name: moviepilot-alertmanager
    restart: unless-stopped
    ports:
      - "9093:9093"
    volumes:
      - ./alertmanager/config.yml:/etc/alertmanager/alertmanager.yml
      - alertmanager_data:/alertmanager
    networks:
      - moviepilot-network

  grafana:
    image: grafana/grafana:10.0.0
    container_name: moviepilot-grafana
    restart: unless-stopped
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
      - GF_USERS_ALLOW_SIGN_UP=false
    volumes:
      - grafana_data:/var/lib/grafana
      - ./grafana/provisioning:/etc/grafana/provisioning
      - ./grafana/dashboards:/var/lib/grafana/dashboards
    networks:
      - moviepilot-network

  loki:
    image: grafana/loki:2.9.0
    container_name: moviepilot-loki
    restart: unless-stopped
    ports:
      - "3100:3100"
    volumes:
      - ./loki/config.yml:/etc/loki/local-config.yaml
      - loki_data:/loki
    command: -config.file=/etc/loki/local-config.yaml
    networks:
      - moviepilot-network

  promtail:
    image: grafana/promtail:2.9.0
    container_name: moviepilot-promtail
    restart: unless-stopped
    volumes:
      - ./promtail/config.yml:/etc/promtail/config.yml
      - /var/log:/var/log:ro
      - /var/lib/docker/containers:/var/lib/docker/containers:ro
    command: -config.file=/etc/promtail/config.yml
    networks:
      - moviepilot-network

  jaeger:
    image: jaegertracing/all-in-one:1.47
    container_name: moviepilot-jaeger
    restart: unless-stopped
    ports:
      - "16686:16686"
      - "14268:14268"
    environment:
      - COLLECTOR_OTLP_ENABLED=true
    networks:
      - moviepilot-network

  node-exporter:
    image: prom/node-exporter:v1.6.0
    container_name: moviepilot-node-exporter
    restart: unless-stopped
    ports:
      - "9100:9100"
    volumes:
      - /proc:/host/proc:ro
      - /sys:/host/sys:ro
      - /:/rootfs:ro
    command:
      - '--path.procfs=/host/proc'
      - '--path.rootfs=/rootfs'
      - '--path.sysfs=/host/sys'
      - '--collector.filesystem.mount-points-exclude=^/(sys|proc|dev|host|etc)($$|/)'
    networks:
      - moviepilot-network

  postgres-exporter:
    image: prometheuscommunity/postgres-exporter:v0.11.1
    container_name: moviepilot-postgres-exporter
    restart: unless-stopped
    ports:
      - "9187:9187"
    environment:
      - DATA_SOURCE_NAME=postgresql://moviepilot:password@postgres:5432/moviepilot?sslmode=disable
    networks:
      - moviepilot-network

  redis-exporter:
    image: oliver006/redis_exporter:v1.43.0
    container_name: moviepilot-redis-exporter
    restart: unless-stopped
    ports:
      - "9121:9121"
    environment:
      - REDIS_ADDR=redis://redis:6379
    networks:
      - moviepilot-network

  cadvisor:
    image: gcr.io/cadvisor/cadvisor:v0.47.0
    container_name: moviepilot-cadvisor
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - /:/rootfs:ro
      - /var/run:/var/run:rw
      - /sys:/sys:ro
      - /var/lib/docker/:/var/lib/docker:ro
    privileged: true
    networks:
      - moviepilot-network

volumes:
  prometheus_data:
  alertmanager_data:
  grafana_data:
  loki_data:

networks:
  moviepilot-network:
    external: true
```

---

**注意**: 监控系统需要根据实际业务需求进行调整和优化。建议定期检查告警规则的有效性，并根据系统变化更新监控指标。