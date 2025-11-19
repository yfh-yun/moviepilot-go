package monitoring

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Dashboard 监控仪表板
type Dashboard struct {
	metrics       *Metrics
	healthChecker *HealthChecker
	mu            sync.RWMutex
	data          map[string]interface{}
}

// DashboardData 仪表板数据
type DashboardData struct {
	Timestamp    time.Time              `json:"timestamp"`
	SystemInfo   SystemInfo             `json:"system_info"`
	HealthStatus HealthStatus           `json:"health_status"`
	Metrics      map[string]interface{} `json:"metrics"`
	Alerts       []Alert                `json:"alerts"`
	Services     []ServiceStatus        `json:"services"`
}

// SystemInfo 系统信息
type SystemInfo struct {
	Uptime       string  `json:"uptime"`
	MemoryUsage  string  `json:"memory_usage"`
	CPUUsage     float64 `json:"cpu_usage"`
	DiskUsage    string  `json:"disk_usage"`
	ActiveUsers  int     `json:"active_users"`
	Requests     int64   `json:"requests"`
	Errors       int64   `json:"errors"`
	ResponseTime float64 `json:"response_time"`
}

// Alert 告警信息
type Alert struct {
	ID           string    `json:"id"`
	Level        string    `json:"level"`
	Message      string    `json:"message"`
	Timestamp    time.Time `json:"timestamp"`
	Service      string    `json:"service"`
	Acknowledged bool      `json:"acknowledged"`
}

// ServiceStatus 服务状态
type ServiceStatus struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Uptime    string    `json:"uptime"`
	Requests  int64     `json:"requests"`
	Errors    int64     `json:"errors"`
	Latency   float64   `json:"latency"`
	LastCheck time.Time `json:"last_check"`
}

// NewDashboard 创建新的监控仪表板
func NewDashboard(metrics *Metrics, healthChecker *HealthChecker) *Dashboard {
	dashboard := &Dashboard{
		metrics:       metrics,
		healthChecker: healthChecker,
		data:          make(map[string]interface{}),
	}

	// 启动数据收集循环
	go dashboard.collectDataLoop()

	return dashboard
}

// collectDataLoop 数据收集循环
func (d *Dashboard) collectDataLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		d.updateData()
	}
}

// updateData 更新仪表板数据
func (d *Dashboard) updateData() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.data["timestamp"] = time.Now()
	d.data["system_info"] = d.collectSystemInfo()
	d.data["health_status"] = d.collectHealthStatus()
	d.data["metrics"] = d.collectMetrics()
	d.data["alerts"] = d.collectAlerts()
	d.data["services"] = d.collectServiceStatus()
}

// collectSystemInfo 收集系统信息
func (d *Dashboard) collectSystemInfo() SystemInfo {
	return SystemInfo{
		Uptime:       fmt.Sprintf("%dh%dm", int(d.metrics.Uptime().Hours()), int(d.metrics.Uptime().Minutes())%60),
		MemoryUsage:  fmt.Sprintf("%.1f%%", d.metrics.MemoryUsage()),
		CPUUsage:     d.metrics.CPUUsage(),
		DiskUsage:    fmt.Sprintf("%.1f%%", d.metrics.DiskUsage()),
		ActiveUsers:  int(d.metrics.ActiveUsers()),
		Requests:     d.metrics.RequestsTotal(),
		Errors:       d.metrics.ErrorsTotal(),
		ResponseTime: d.metrics.ResponseTimeAvg(),
	}
}

// collectHealthStatus 收集健康状态
func (d *Dashboard) collectHealthStatus() HealthStatus {
	return d.healthChecker.GetStatus()
}

// collectMetrics 收集指标数据
func (d *Dashboard) collectMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	// 收集所有指标的最新值
	metrics["requests_total"] = d.metrics.RequestsTotal()
	metrics["errors_total"] = d.metrics.ErrorsTotal()
	metrics["response_time_avg"] = d.metrics.ResponseTimeAvg()
	metrics["active_users"] = d.metrics.ActiveUsers()
	metrics["memory_usage"] = d.metrics.MemoryUsage()
	metrics["cpu_usage"] = d.metrics.CPUUsage()
	metrics["disk_usage"] = d.metrics.DiskUsage()

	return metrics
}

// collectAlerts 收集告警信息
func (d *Dashboard) collectAlerts() []Alert {
	// 从health checker获取告警信息
	status := d.healthChecker.GetStatus()
	var alerts []Alert

	for service, serviceStatus := range status.Services {
		if !serviceStatus.Healthy {
			alerts = append(alerts, Alert{
				ID:           fmt.Sprintf("service_%s", service),
				Level:        "error",
				Message:      fmt.Sprintf("服务 %s 异常: %s", service, serviceStatus.Error),
				Timestamp:    serviceStatus.LastChecked,
				Service:      service,
				Acknowledged: false,
			})
		}
	}

	// 按时间排序
	sort.Slice(alerts, func(i, j int) bool {
		return alerts[i].Timestamp.After(alerts[j].Timestamp)
	})

	return alerts
}

// collectServiceStatus 收集服务状态
func (d *Dashboard) collectServiceStatus() []ServiceStatus {
	status := d.healthChecker.GetStatus()
	var services []ServiceStatus

	for name, serviceStatus := range status.Services {
		services = append(services, ServiceStatus{
			Name:      name,
			Status:    serviceStatus.HealthyString(),
			Uptime:    serviceStatus.Uptime.String(),
			Requests:  serviceStatus.Metrics["requests"],
			Errors:    serviceStatus.Metrics["errors"],
			Latency:   serviceStatus.Metrics["latency"],
			LastCheck: serviceStatus.LastChecked,
		})
	}

	// 按名称排序
	sort.Slice(services, func(i, j int) bool {
		return services[i].Name < services[j].Name
	})

	return services
}

// GetData 获取仪表板数据
func (d *Dashboard) GetData() DashboardData {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return DashboardData{
		Timestamp:    d.data["timestamp"].(time.Time),
		SystemInfo:   d.data["system_info"].(SystemInfo),
		HealthStatus: d.data["health_status"].(HealthStatus),
		Metrics:      d.data["metrics"].(map[string]interface{}),
		Alerts:       d.data["alerts"].([]Alert),
		Services:     d.data["services"].([]ServiceStatus),
	}
}

// ServeHTTP 提供HTTP接口
func (d *Dashboard) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/metrics":
		promhttp.Handler().ServeHTTP(w, r)
	case "/api/health":
		d.serveHealthAPI(w, r)
	case "/api/metrics":
		d.serveMetricsAPI(w, r)
	case "/api/alerts":
		d.serveAlertsAPI(w, r)
	case "/api/services":
		d.serveServicesAPI(w, r)
	case "/api/dashboard":
		d.serveDashboardAPI(w, r)
	case "/", "/dashboard":
		d.serveDashboardUI(w, r)
	default:
		http.NotFound(w, r)
	}
}

// serveHealthAPI 提供健康检查API
func (d *Dashboard) serveHealthAPI(w http.ResponseWriter, r *http.Request) {
	status := d.healthChecker.GetStatus()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// serveMetricsAPI 提供指标API
func (d *Dashboard) serveMetricsAPI(w http.ResponseWriter, r *http.Request) {
	data := d.GetData()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data.Metrics)
}

// serveAlertsAPI 提供告警API
func (d *Dashboard) serveAlertsAPI(w http.ResponseWriter, r *http.Request) {
	data := d.GetData()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data.Alerts)
}

// serveServicesAPI 提供服务API
func (d *Dashboard) serveServicesAPI(w http.ResponseWriter, r *http.Request) {
	data := d.GetData()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data.Services)
}

// serveDashboardAPI 提供仪表板API
func (d *Dashboard) serveDashboardAPI(w http.ResponseWriter, r *http.Request) {
	data := d.GetData()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// serveDashboardUI 提供仪表板UI
func (d *Dashboard) serveDashboardUI(w http.ResponseWriter, r *http.Request) {
	// 简单的HTML仪表板
	tmpl := `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>MoviePilot 监控仪表板</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background-color: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; }
        .header { background: #2c3e50; color: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; }
        .status-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; margin-bottom: 20px; }
        .status-card { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .status-card.healthy { border-left: 4px solid #27ae60; }
        .status-card.warning { border-left: 4px solid #f39c12; }
        .status-card.error { border-left: 4px solid #e74c3c; }
        .alert-list { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .alert-item { padding: 10px; margin: 5px 0; border-radius: 4px; }
        .alert-error { background: #ffebee; color: #c62828; }
        .alert-warning { background: #fff3e0; color: #ef6c00; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎬 MoviePilot 监控仪表板</h1>
            <p>最后更新: <span id="timestamp">{{.Timestamp.Format "2006-01-02 15:04:05"}}</span></p>
        </div>
        
        <div class="status-grid">
            <div class="status-card {{if .HealthStatus.Healthy}}healthy{{else}}error{{end}}">
                <h3>系统状态</h3>
                <p>{{if .HealthStatus.Healthy}}✅ 健康{{else}}❌ 异常{{end}}</p>
                <p>运行时间: {{.SystemInfo.Uptime}}</p>
            </div>
            
            <div class="status-card">
                <h3>系统资源</h3>
                <p>CPU: {{printf "%.1f%%" .SystemInfo.CPUUsage}}</p>
                <p>内存: {{.SystemInfo.MemoryUsage}}</p>
                <p>磁盘: {{.SystemInfo.DiskUsage}}</p>
            </div>
            
            <div class="status-card">
                <h3>性能指标</h3>
                <p>请求数: {{.SystemInfo.Requests}}</p>
                <p>错误数: {{.SystemInfo.Errors}}</p>
                <p>响应时间: {{printf "%.2fms" .SystemInfo.ResponseTime}}</p>
            </div>
            
            <div class="status-card">
                <h3>活跃用户</h3>
                <p>当前: {{.SystemInfo.ActiveUsers}}</p>
            </div>
        </div>
        
        {{if .Alerts}}
        <div class="alert-list">
            <h3>告警信息</h3>
            {{range .Alerts}}
            <div class="alert-item alert-{{.Level}}">
                <strong>{{.Level | upper}}:</strong> {{.Message}}<br>
                <small>{{.Service}} - {{.Timestamp.Format "2006-01-02 15:04:05"}}</small>
            </div>
            {{end}}
        </div>
        {{end}}
        
        <div class="alert-list">
            <h3>服务状态</h3>
            {{range .Services}}
            <div class="status-card {{if eq .Status "healthy"}}healthy{{else}}error{{end}}">
                <h4>{{.Name}}</h4>
                <p>状态: {{.Status}}</p>
                <p>请求数: {{.Requests}}</p>
                <p>错误数: {{.Errors}}</p>
                <p>延迟: {{printf "%.2fms" .Latency}}</p>
            </div>
            {{end}}
        </div>
    </div>
    
    <script>
        // 定时刷新数据
        setInterval(() => {
            fetch('/api/dashboard')
                .then(response => response.json())
                .then(data => {
                    document.getElementById('timestamp').textContent = 
                        new Date(data.timestamp).toLocaleString();
                });
        }, 10000);
    </script>
</body>
</html>`

	t, err := template.New("dashboard").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := d.GetData()
	if err := t.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// RegisterMetrics 注册Prometheus指标
func (d *Dashboard) RegisterMetrics() {
	prometheus.MustRegister(
		prometheus.NewGaugeFunc(prometheus.GaugeOpts{
			Name: "system_uptime_hours",
			Help: "系统运行时间（小时）",
		}, func() float64 {
			return d.metrics.Uptime().Hours()
		}),

		prometheus.NewGaugeFunc(prometheus.GaugeOpts{
			Name: "system_memory_usage_percent",
			Help: "内存使用率",
		}, func() float64 {
			return d.metrics.MemoryUsage()
		}),

		prometheus.NewGaugeFunc(prometheus.GaugeOpts{
			Name: "system_cpu_usage_percent",
			Help: "CPU使用率",
		}, func() float64 {
			return d.metrics.CPUUsage()
		}),
	)
}
