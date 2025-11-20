package monitoring

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// HealthStatus 健康状态
type HealthStatus string

const (
	HealthStatusUp       HealthStatus = "UP"
	HealthStatusDown     HealthStatus = "DOWN"
	HealthStatusDegraded HealthStatus = "DEGRADED"
)

// HealthCheckResult 健康检查结果
type HealthCheckResult struct {
	Name      string        `json:"name"`
	Status    HealthStatus  `json:"status"`
	Details   interface{}   `json:"details,omitempty"`
	Error     string        `json:"error,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
	Duration  time.Duration `json:"duration"`
}

// HealthCheck 健康检查接口
type HealthCheck interface {
	Name() string
	Check(ctx context.Context) (*HealthCheckResult, error)
}

// HealthChecker 健康检查器
type HealthChecker struct {
	checks map[string]HealthCheck
	mu     sync.RWMutex
}

// NewHealthChecker 创建新的健康检查器
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		checks: make(map[string]HealthCheck),
	}
}

// RegisterHealthCheck 注册健康检查
func (h *HealthChecker) RegisterHealthCheck(check HealthCheck) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	name := check.Name()
	if _, exists := h.checks[name]; exists {
		return fmt.Errorf("health check %s already registered", name)
	}

	h.checks[name] = check
	return nil
}

// UnregisterHealthCheck 注销健康检查
func (h *HealthChecker) UnregisterHealthCheck(name string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.checks[name]; !exists {
		return fmt.Errorf("health check %s not found", name)
	}

	delete(h.checks, name)
	return nil
}

// CheckAll 执行所有健康检查
func (h *HealthChecker) CheckAll(ctx context.Context) []*HealthCheckResult {
	h.mu.RLock()
	checks := make([]HealthCheck, 0, len(h.checks))
	for _, check := range h.checks {
		checks = append(checks, check)
	}
	h.mu.RUnlock()

	results := make([]*HealthCheckResult, len(checks))

	var wg sync.WaitGroup
	for i, check := range checks {
		wg.Add(1)
		go func(idx int, check HealthCheck) {
			defer wg.Done()
			result, err := check.Check(ctx)
			if err != nil {
				results[idx] = &HealthCheckResult{
					Name:      check.Name(),
					Status:    HealthStatusDown,
					Error:     err.Error(),
					Timestamp: time.Now(),
				}
			} else {
				results[idx] = result
			}
		}(i, check)
	}
	wg.Wait()

	return results
}

// Check 执行指定健康检查
func (h *HealthChecker) Check(ctx context.Context, name string) (*HealthCheckResult, error) {
	h.mu.RLock()
	check, exists := h.checks[name]
	h.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("health check %s not found", name)
	}

	return check.Check(ctx)
}

// OverallStatus 获取整体健康状态
func (h *HealthChecker) OverallStatus(ctx context.Context) HealthStatus {
	results := h.CheckAll(ctx)

	hasDown := false
	hasDegraded := false

	for _, result := range results {
		if result.Status == HealthStatusDown {
			hasDown = true
		}
		if result.Status == HealthStatusDegraded {
			hasDegraded = true
		}
	}

	if hasDown {
		return HealthStatusDown
	}
	if hasDegraded {
		return HealthStatusDegraded
	}
	return HealthStatusUp
}

// Handler 获取HTTP处理程序
func (h *HealthChecker) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		results := h.CheckAll(ctx)

		statusCode := http.StatusOK
		overallStatus := h.OverallStatus(ctx)

		if overallStatus == HealthStatusDown {
			statusCode = http.StatusServiceUnavailable
		} else if overallStatus == HealthStatusDegraded {
			statusCode = http.StatusOK // 降级状态仍返回200
		}

		response := map[string]interface{}{
			"status":    overallStatus,
			"checks":    results,
			"timestamp": time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)

		json.NewEncoder(w).Encode(response)
	}
}

// DatabaseHealthCheck 数据库健康检查
type DatabaseHealthCheck struct {
	name string
	db   *sql.DB
}

// NewDatabaseHealthCheck 创建数据库健康检查
func NewDatabaseHealthCheck(name string, db *sql.DB) *DatabaseHealthCheck {
	return &DatabaseHealthCheck{
		name: name,
		db:   db,
	}
}

// Name 返回检查名称
func (d *DatabaseHealthCheck) Name() string {
	return d.name
}

// Check 执行健康检查
func (d *DatabaseHealthCheck) Check(ctx context.Context) (*HealthCheckResult, error) {
	start := time.Now()

	var err error
	if d.db == nil {
		err = fmt.Errorf("database connection is nil")
	} else {
		err = d.db.PingContext(ctx)
	}

	duration := time.Since(start)

	if err != nil {
		return &HealthCheckResult{
			Name:      d.name,
			Status:    HealthStatusDown,
			Error:     err.Error(),
			Timestamp: time.Now(),
			Duration:  duration,
		}, nil
	}

	// 获取数据库统计信息
	stats := d.db.Stats()

	return &HealthCheckResult{
		Name:   d.name,
		Status: HealthStatusUp,
		Details: map[string]interface{}{
			"open_connections": stats.OpenConnections,
			"in_use":           stats.InUse,
			"idle":             stats.Idle,
			"wait_count":       stats.WaitCount,
			"wait_duration":    stats.WaitDuration.String(),
		},
		Timestamp: time.Now(),
		Duration:  duration,
	}, nil
}

// RedisHealthCheck Redis健康检查
type RedisHealthCheck struct {
	name   string
	client *redis.Client
}

// NewRedisHealthCheck 创建Redis健康检查
func NewRedisHealthCheck(name string, client *redis.Client) *RedisHealthCheck {
	return &RedisHealthCheck{
		name:   name,
		client: client,
	}
}

// Name 返回检查名称
func (r *RedisHealthCheck) Name() string {
	return r.name
}

// Check 执行健康检查
func (r *RedisHealthCheck) Check(ctx context.Context) (*HealthCheckResult, error) {
	start := time.Now()

	var err error
	if r.client == nil {
		err = fmt.Errorf("redis client is nil")
	} else {
		_, err = r.client.Ping(ctx).Result()
	}

	duration := time.Since(start)

	if err != nil {
		return &HealthCheckResult{
			Name:      r.name,
			Status:    HealthStatusDown,
			Error:     err.Error(),
			Timestamp: time.Now(),
			Duration:  duration,
		}, nil
	}

	// 获取Redis统计信息
	info := r.client.Info(ctx, "stats", "memory", "clients").Val()

	return &HealthCheckResult{
		Name:   r.name,
		Status: HealthStatusUp,
		Details: map[string]interface{}{
			"info": info,
		},
		Timestamp: time.Now(),
		Duration:  duration,
	}, nil
}

// DiskSpaceHealthCheck 磁盘空间健康检查
type DiskSpaceHealthCheck struct {
	name      string
	path      string
	threshold uint64 // 阈值（字节）
}

// NewDiskSpaceHealthCheck 创建磁盘空间健康检查
func NewDiskSpaceHealthCheck(name, path string, threshold uint64) *DiskSpaceHealthCheck {
	return &DiskSpaceHealthCheck{
		name:      name,
		path:      path,
		threshold: threshold,
	}
}

// Name 返回检查名称
func (d *DiskSpaceHealthCheck) Name() string {
	return d.name
}

// Check 执行健康检查
func (d *DiskSpaceHealthCheck) Check(ctx context.Context) (*HealthCheckResult, error) {
	start := time.Now()

	// 这里简化实现，实际应该使用syscall.Statfs等系统调用
	// 暂时返回模拟数据
	var status HealthStatus = HealthStatusUp
	var errorMsg string

	// 模拟磁盘检查
	availableSpace := uint64(1024 * 1024 * 1024) // 1GB
	if availableSpace < d.threshold {
		status = HealthStatusDegraded
		errorMsg = fmt.Sprintf("disk space low: %d bytes available", availableSpace)
	}

	duration := time.Since(start)

	return &HealthCheckResult{
		Name:   d.name,
		Status: status,
		Details: map[string]interface{}{
			"available_space": availableSpace,
			"threshold":       d.threshold,
			"path":            d.path,
		},
		Error:     errorMsg,
		Timestamp: time.Now(),
		Duration:  duration,
	}, nil
}

// MemoryHealthCheck 内存健康检查
type MemoryHealthCheck struct {
	name      string
	threshold float64 // 内存使用率阈值（0-1）
}

// NewMemoryHealthCheck 创建内存健康检查
func NewMemoryHealthCheck(name string, threshold float64) *MemoryHealthCheck {
	return &MemoryHealthCheck{
		name:      name,
		threshold: threshold,
	}
}

// Name 返回检查名称
func (m *MemoryHealthCheck) Name() string {
	return m.name
}

// Check 执行健康检查
func (m *MemoryHealthCheck) Check(ctx context.Context) (*HealthCheckResult, error) {
	start := time.Now()

	// 这里简化实现，实际应该使用runtime包获取内存信息
	// 暂时返回模拟数据
	var status HealthStatus = HealthStatusUp
	var errorMsg string

	// 模拟内存检查
	memoryUsage := 0.7 // 70%内存使用率
	if memoryUsage > m.threshold {
		status = HealthStatusDegraded
		errorMsg = fmt.Sprintf("memory usage high: %.2f%%", memoryUsage*100)
	}

	duration := time.Since(start)

	return &HealthCheckResult{
		Name:   m.name,
		Status: status,
		Details: map[string]interface{}{
			"memory_usage": memoryUsage,
			"threshold":    m.threshold,
		},
		Error:     errorMsg,
		Timestamp: time.Now(),
		Duration:  duration,
	}, nil
}

// ServiceHealthCheck 服务健康检查
type ServiceHealthCheck struct {
	name    string
	url     string
	timeout time.Duration
}

// NewServiceHealthCheck 创建服务健康检查
func NewServiceHealthCheck(name, url string, timeout time.Duration) *ServiceHealthCheck {
	return &ServiceHealthCheck{
		name:    name,
		url:     url,
		timeout: timeout,
	}
}

// Name 返回检查名称
func (s *ServiceHealthCheck) Name() string {
	return s.name
}

// Check 执行健康检查
func (s *ServiceHealthCheck) Check(ctx context.Context) (*HealthCheckResult, error) {
	start := time.Now()

	httpClient := &http.Client{
		Timeout: s.timeout,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", s.url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := httpClient.Do(req)
	duration := time.Since(start)

	if err != nil {
		return &HealthCheckResult{
			Name:      s.name,
			Status:    HealthStatusDown,
			Error:     err.Error(),
			Timestamp: time.Now(),
			Duration:  duration,
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return &HealthCheckResult{
			Name:   s.name,
			Status: HealthStatusUp,
			Details: map[string]interface{}{
				"status_code": resp.StatusCode,
				"url":         s.url,
			},
			Timestamp: time.Now(),
			Duration:  duration,
		}, nil
	}

	return &HealthCheckResult{
		Name:   s.name,
		Status: HealthStatusDown,
		Details: map[string]interface{}{
			"status_code": resp.StatusCode,
			"url":         s.url,
		},
		Error:     fmt.Sprintf("HTTP status: %d", resp.StatusCode),
		Timestamp: time.Now(),
		Duration:  duration,
	}, nil
}
