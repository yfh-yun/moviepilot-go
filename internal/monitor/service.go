package monitor

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// MonitorService 监控服务
type MonitorService struct {
	systemMonitor  *SystemMonitor
	browserMonitor *BrowserMonitor
	cfHandler      *CloudflareHandler
	config         MonitorServiceConfig
	logger         *zap.Logger
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	running        bool
	mu             sync.RWMutex
}

// MonitorServiceConfig 监控服务配置
type MonitorServiceConfig struct {
	SystemMonitor  MonitorConfig       `json:"system_monitor"`
	BrowserMonitor BrowserMonitorConfig `json:"browser_monitor"`
	Cloudflare     CloudflareConfig     `json:"cloudflare"`
	Enabled        bool                 `json:"enabled"`
	LogLevel       string               `json:"log_level"`
}

// NewMonitorService 创建监控服务
func NewMonitorService(config MonitorServiceConfig, logger *zap.Logger) *MonitorService {
	ctx, cancel := context.WithCancel(context.Background())
	
	service := &MonitorService{
		config: config,
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
	}
	
	// 初始化系统监控器
	if config.SystemMonitor.Enabled {
		service.systemMonitor = NewSystemMonitor(config.SystemMonitor, logger)
	}
	
	// 初始化浏览器监控器
	if config.BrowserMonitor.Enabled {
		service.browserMonitor = NewBrowserMonitor(config.BrowserMonitor, logger)
	}
	
	// 初始化Cloudflare处理器
	if config.Cloudflare.Enabled {
		service.cfHandler = NewCloudflareHandler(config.Cloudflare, logger)
		// 关联浏览器监控器
		if service.browserMonitor != nil {
			service.cfHandler.SetBrowserMonitor(service.browserMonitor)
		}
	}
	
	return service
}

// Start 启动监控服务
func (ms *MonitorService) Start() error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	
	if ms.running {
		return fmt.Errorf("监控服务已经在运行")
	}
	
	if !ms.config.Enabled {
		ms.logger.Info("监控服务已禁用")
		return nil
	}
	
	ms.logger.Info("启动监控服务")
	
	// 启动系统监控
	if ms.systemMonitor != nil {
		if err := ms.systemMonitor.Start(); err != nil {
			return fmt.Errorf("启动系统监控失败: %w", err)
		}
	}
	
	// 启动浏览器监控
	if ms.browserMonitor != nil {
		if err := ms.browserMonitor.Start(); err != nil {
			return fmt.Errorf("启动浏览器监控失败: %w", err)
		}
	}
	
	ms.running = true
	return nil
}

// Stop 停止监控服务
func (ms *MonitorService) Stop() error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	
	if !ms.running {
		return nil
	}
	
	ms.logger.Info("停止监控服务")
	
	// 停止系统监控
	if ms.systemMonitor != nil {
		if err := ms.systemMonitor.Stop(); err != nil {
			ms.logger.Error("停止系统监控失败", zap.Error(err))
		}
	}
	
	// 停止浏览器监控
	if ms.browserMonitor != nil {
		if err := ms.browserMonitor.Stop(); err != nil {
			ms.logger.Error("停止浏览器监控失败", zap.Error(err))
		}
	}
	
	ms.cancel()
	ms.wg.Wait()
	
	ms.running = false
	return nil
}

// GetSystemMetrics 获取系统指标
func (ms *MonitorService) GetSystemMetrics() (map[string]float64, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	if ms.systemMonitor == nil {
		return nil, fmt.Errorf("系统监控未启用")
	}
	
	return ms.systemMonitor.GetMetrics(), nil
}

// GetBrowserMetrics 获取浏览器指标
func (ms *MonitorService) GetBrowserMetrics() (map[string]BrowserMetrics, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	if ms.browserMonitor == nil {
		return nil, fmt.Errorf("浏览器监控未启用")
	}
	
	return ms.browserMonitor.GetMetrics(), nil
}

// CreateBrowser 创建浏览器实例
func (ms *MonitorService) CreateBrowser(browserID string) (*BrowserInstance, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	if ms.browserMonitor == nil {
		return nil, fmt.Errorf("浏览器监控未启用")
	}
	
	return ms.browserMonitor.CreateBrowser(browserID)
}

// CloseBrowser 关闭浏览器实例
func (ms *MonitorService) CloseBrowser(browserID string) error {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	if ms.browserMonitor == nil {
		return fmt.Errorf("浏览器监控未启用")
	}
	
	return ms.browserMonitor.CloseBrowser(browserID)
}

// NavigateBrowser 浏览器导航
func (ms *MonitorService) NavigateBrowser(browserID string, url string) error {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	if ms.browserMonitor == nil {
		return fmt.Errorf("浏览器监控未启用")
	}
	
	return ms.browserMonitor.Navigate(browserID, url)
}

// ExecuteBrowserAction 执行浏览器动作
func (ms *MonitorService) ExecuteBrowserAction(browserID string, action BrowserAction) (*BrowserActionResult, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	if ms.browserMonitor == nil {
		return nil, fmt.Errorf("浏览器监控未启用")
	}
	
	return ms.browserMonitor.ExecuteAction(browserID, action)
}

// HandleCloudflareChallenge 处理Cloudflare挑战
func (ms *MonitorService) HandleCloudflareChallenge(targetURL string, httpClient *http.Client) (*CloudflareInfo, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	if ms.cfHandler == nil {
		return nil, fmt.Errorf("Cloudflare处理器未启用")
	}
	
	return ms.cfHandler.HandleChallenge(targetURL, httpClient)
}

// GetSystemInfo 获取系统信息
func (ms *MonitorService) GetSystemInfo() (*SystemInfo, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	if ms.systemMonitor == nil {
		return nil, fmt.Errorf("系统监控未启用")
	}
	
	return ms.systemMonitor.GetSystemInfo()
}

// GetActiveAlerts 获取活跃告警
func (ms *MonitorService) GetActiveAlerts() ([]Alert, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	if ms.systemMonitor == nil {
		return nil, fmt.Errorf("系统监控未启用")
	}
	
	return ms.systemMonitor.alertManager.GetActiveAlerts(), nil
}

// GetAlertHistory 获取告警历史
func (ms *MonitorService) GetAlertHistory(limit int) ([]Alert, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	if ms.systemMonitor == nil {
		return nil, fmt.Errorf("系统监控未启用")
	}
	
	return ms.systemMonitor.alertManager.GetAlertHistory(limit), nil
}

// ListBrowsers 列出浏览器实例
func (ms *MonitorService) ListBrowsers() ([]string, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	if ms.browserMonitor == nil {
		return nil, fmt.Errorf("浏览器监控未启用")
	}
	
	return ms.browserMonitor.ListBrowsers(), nil
}

// GetBrowserInstance 获取浏览器实例
func (ms *MonitorService) GetBrowserInstance(browserID string) (*BrowserInstance, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	if ms.browserMonitor == nil {
		return nil, fmt.Errorf("浏览器监控未启用")
	}
	
	return ms.browserMonitor.GetBrowserInstance(browserID)
}

// IsRunning 检查服务是否运行
func (ms *MonitorService) IsRunning() bool {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.running
}

// UpdateConfig 更新配置
func (ms *MonitorService) UpdateConfig(config MonitorServiceConfig) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	
	oldConfig := ms.config
	ms.config = config
	
	// 更新系统监控配置
	if ms.systemMonitor != nil {
		if err := ms.systemMonitor.UpdateConfig(config.SystemMonitor); err != nil {
			return fmt.Errorf("更新系统监控配置失败: %w", err)
		}
	}
	
	// 更新浏览器监控配置
	if ms.browserMonitor != nil {
		if err := ms.browserMonitor.UpdateConfig(config.BrowserMonitor); err != nil {
			return fmt.Errorf("更新浏览器监控配置失败: %w", err)
		}
	}
	
	// 更新Cloudflare配置
	if ms.cfHandler != nil {
		ms.cfHandler.UpdateConfig(config.Cloudflare)
	}
	
	ms.logger.Info("监控服务配置已更新")
	
	return nil
}

// GetStatus 获取服务状态
func (ms *MonitorService) GetStatus() map[string]interface{} {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	status := map[string]interface{}{
		"running":           ms.running,
		"enabled":           ms.config.Enabled,
		"system_monitor":    false,
		"browser_monitor":   false,
		"cloudflare":        false,
		"browser_instances":  0,
		"active_alerts":    0,
	}
	
	if ms.systemMonitor != nil && ms.systemMonitor.IsRunning() {
		status["system_monitor"] = true
		activeAlerts := ms.systemMonitor.alertManager.GetActiveAlerts()
		status["active_alerts"] = len(activeAlerts)
	}
	
	if ms.browserMonitor != nil && ms.browserMonitor.IsRunning() {
		status["browser_monitor"] = true
		browsers := ms.browserMonitor.ListBrowsers()
		status["browser_instances"] = len(browsers)
	}
	
	if ms.cfHandler != nil {
		status["cloudflare"] = true
	}
	
	return status
}

// GetHealth 获取健康状态
func (ms *MonitorService) GetHealth() map[string]interface{} {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"components": make(map[string]interface{}),
	}
	
	components := health["components"].(map[string]interface{})
	
	// 检查系统监控健康状态
	if ms.systemMonitor != nil {
		if ms.systemMonitor.IsRunning() {
			components["system_monitor"] = map[string]interface{}{
				"status":  "healthy",
				"uptime":  time.Since(ms.systemMonitor.lastUpdate),
				"metrics": len(ms.systemMonitor.lastMetrics),
			}
		} else {
			components["system_monitor"] = map[string]interface{}{
				"status":  "unhealthy",
				"reason":  "not_running",
			}
			health["status"] = "degraded"
		}
	}
	
	// 检查浏览器监控健康状态
	if ms.browserMonitor != nil {
		if ms.browserMonitor.IsRunning() {
			browsers := ms.browserMonitor.ListBrowsers()
			components["browser_monitor"] = map[string]interface{}{
				"status":    "healthy",
				"instances": len(browsers),
			}
		} else {
			components["browser_monitor"] = map[string]interface{}{
				"status": "unhealthy",
				"reason": "not_running",
			}
			health["status"] = "degraded"
		}
	}
	
	// 检查Cloudflare处理器健康状态
	if ms.cfHandler != nil {
		components["cloudflare"] = map[string]interface{}{
			"status":  "healthy",
			"enabled": ms.cfHandler.config.Enabled,
		}
	}
	
	return health
}