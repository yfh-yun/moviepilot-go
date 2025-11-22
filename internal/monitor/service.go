package monitor

import (
	"context"

	"net/http"
	"sync"
	"time"

	"moviepilot-go/pkg/errors"
	"moviepilot-go/pkg/logger"

	"go.uber.org/zap"
)

// MonitorService 监控服务
type MonitorService struct {
	systemMonitor  *SystemMonitor
	browserMonitor *BrowserMonitor
	cfHandler      *CloudflareHandler
	config         MonitorServiceConfig
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	running        bool
	mu             sync.RWMutex
}

// MonitorServiceConfig 监控服务配置
type MonitorServiceConfig struct {
	SystemMonitor  MonitorConfig        `json:"system_monitor"`
	BrowserMonitor BrowserMonitorConfig `json:"browser_monitor"`
	Cloudflare     CloudflareConfig     `json:"cloudflare"`
	Enabled        bool                 `json:"enabled"`
	LogLevel       string               `json:"log_level"`
}

// NewMonitorService 创建监控服务
func NewMonitorService(config MonitorServiceConfig) *MonitorService {
	logger.Debug("Creating new MonitorService instance",
		zap.String("func", "NewMonitorService"),
		zap.Bool("enabled", config.Enabled),
		zap.Bool("system_monitor_enabled", config.SystemMonitor.Enabled),
		zap.Bool("browser_monitor_enabled", config.BrowserMonitor.Enabled),
		zap.Bool("cloudflare_enabled", config.Cloudflare.Enabled))

	ctx, cancel := context.WithCancel(context.Background())

	service := &MonitorService{
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}

	// 初始化系统监控器
	if config.SystemMonitor.Enabled {
		service.systemMonitor = NewSystemMonitor(config.SystemMonitor, logger.GetLogger())
	}

	// 初始化浏览器监控器
	if config.BrowserMonitor.Enabled {
		service.browserMonitor = NewBrowserMonitor(config.BrowserMonitor, logger.GetLogger())
	}

	// 初始化Cloudflare处理器
	if config.Cloudflare.Enabled {
		service.cfHandler = NewCloudflareHandler(config.Cloudflare, logger.GetLogger())
		// 关联浏览器监控器
		if service.browserMonitor != nil {
			service.cfHandler.SetBrowserMonitor(service.browserMonitor)
		}
	}

	return service
}

// Start 启动监控服务
func (ms *MonitorService) Start() error {
	logger.Debug("Starting monitor service",
		zap.String("func", "MonitorService.Start"))

	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.running {
		logger.Warn("Monitor service is already running",
			zap.String("func", "MonitorService.Start"))
		return errors.NewAppError(http.StatusConflict, "监控服务已经在运行", "SERVICE_ALREADY_RUNNING")
	}

	if !ms.config.Enabled {
		logger.Info("监控服务已禁用",
			zap.String("func", "MonitorService.Start"))
		return nil
	}

	logger.Info("启动监控服务",
		zap.String("func", "MonitorService.Start"))

	// 启动系统监控
	if ms.systemMonitor != nil {
		if err := ms.systemMonitor.Start(); err != nil {
			logger.Error("Failed to start system monitor",
				zap.String("func", "MonitorService.Start"),
				zap.Error(err))
			return errors.WrapError(err, "启动系统监控失败")
		}
		logger.Info("System monitor started successfully",
			zap.String("func", "MonitorService.Start"))
	}

	// 启动浏览器监控
	if ms.browserMonitor != nil {
		if err := ms.browserMonitor.Start(); err != nil {
			logger.Error("Failed to start browser monitor",
				zap.String("func", "MonitorService.Start"),
				zap.Error(err))
			return errors.WrapError(err, "启动浏览器监控失败")
		}
		logger.Info("Browser monitor started successfully",
			zap.String("func", "MonitorService.Start"))
	}

	ms.running = true
	return nil
}

// Stop 停止监控服务
func (ms *MonitorService) Stop() error {
	logger.Debug("Stopping monitor service",
		zap.String("func", "MonitorService.Stop"))

	ms.mu.Lock()
	defer ms.mu.Unlock()

	if !ms.running {
		logger.Debug("Monitor service is not running",
			zap.String("func", "MonitorService.Stop"))
		return nil
	}

	logger.Info("停止监控服务",
		zap.String("func", "MonitorService.Stop"))

	// 停止系统监控
	if ms.systemMonitor != nil {
		if err := ms.systemMonitor.Stop(); err != nil {
			logger.Error("停止系统监控失败",
				zap.String("func", "MonitorService.Stop"),
				zap.Error(err))
		}
	}

	// 停止浏览器监控
	if ms.browserMonitor != nil {
		if err := ms.browserMonitor.Stop(); err != nil {
			logger.Error("停止浏览器监控失败",
				zap.String("func", "MonitorService.Stop"),
				zap.Error(err))
		}
	}

	ms.cancel()
	ms.wg.Wait()

	ms.running = false
	return nil
}

// GetSystemMetrics 获取系统指标
func (ms *MonitorService) GetSystemMetrics() (map[string]float64, error) {
	logger.Debug("Getting system metrics",
		zap.String("func", "MonitorService.GetSystemMetrics"))

	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if ms.systemMonitor == nil {
		logger.Warn("System monitor is not enabled",
			zap.String("func", "MonitorService.GetSystemMetrics"))
		return nil, errors.NewAppError(http.StatusServiceUnavailable, "系统监控未启用", "SYSTEM_MONITOR_NOT_ENABLED")
	}

	metrics := ms.systemMonitor.GetMetrics()
	logger.Debug("System metrics retrieved successfully",
		zap.String("func", "MonitorService.GetSystemMetrics"),
		zap.Int("metrics_count", len(metrics)))

	return metrics, nil
}

// GetBrowserMetrics 获取浏览器指标
func (ms *MonitorService) GetBrowserMetrics() (map[string]BrowserMetrics, error) {
	logger.Debug("Getting browser metrics",
		zap.String("func", "MonitorService.GetBrowserMetrics"))

	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if ms.browserMonitor == nil {
		logger.Warn("Browser monitor is not enabled",
			zap.String("func", "MonitorService.GetBrowserMetrics"))
		return nil, errors.NewAppError(http.StatusServiceUnavailable, "浏览器监控未启用", "BROWSER_MONITOR_NOT_ENABLED")
	}

	metrics := ms.browserMonitor.GetMetrics()
	logger.Debug("Browser metrics retrieved successfully",
		zap.String("func", "MonitorService.GetBrowserMetrics"),
		zap.Int("metrics_count", len(metrics)))

	return metrics, nil
}

// CreateBrowser 创建浏览器实例
func (ms *MonitorService) CreateBrowser(browserID string) (*BrowserInstance, error) {
	logger.Debug("Creating browser instance",
		zap.String("func", "MonitorService.CreateBrowser"),
		zap.String("browser_id", browserID))

	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if ms.browserMonitor == nil {
		logger.Warn("Browser monitor is not enabled",
			zap.String("func", "MonitorService.CreateBrowser"),
			zap.String("browser_id", browserID))
		return nil, errors.NewAppError(http.StatusServiceUnavailable, "浏览器监控未启用", "BROWSER_MONITOR_NOT_ENABLED")
	}

	instance, err := ms.browserMonitor.CreateBrowser(browserID)
	if err != nil {
		logger.Error("Failed to create browser instance",
			zap.String("func", "MonitorService.CreateBrowser"),
			zap.String("browser_id", browserID),
			zap.Error(err))
		return nil, errors.WrapError(err, "创建浏览器实例失败")
	}

	logger.Info("Browser instance created successfully",
		zap.String("func", "MonitorService.CreateBrowser"),
		zap.String("browser_id", browserID))

	return instance, nil
}

// CloseBrowser 关闭浏览器实例
func (ms *MonitorService) CloseBrowser(browserID string) error {
	logger.Debug("Closing browser instance",
		zap.String("func", "MonitorService.CloseBrowser"),
		zap.String("browser_id", browserID))

	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if ms.browserMonitor == nil {
		logger.Warn("Browser monitor is not enabled",
			zap.String("func", "MonitorService.CloseBrowser"),
			zap.String("browser_id", browserID))
		return errors.NewAppError(http.StatusServiceUnavailable, "浏览器监控未启用", "BROWSER_MONITOR_NOT_ENABLED")
	}

	err := ms.browserMonitor.CloseBrowser(browserID)
	if err != nil {
		logger.Error("Failed to close browser instance",
			zap.String("func", "MonitorService.CloseBrowser"),
			zap.String("browser_id", browserID),
			zap.Error(err))
		return errors.WrapError(err, "关闭浏览器实例失败")
	}

	logger.Info("Browser instance closed successfully",
		zap.String("func", "MonitorService.CloseBrowser"),
		zap.String("browser_id", browserID))

	return nil
}

// NavigateBrowser 浏览器导航
func (ms *MonitorService) NavigateBrowser(browserID string, url string) error {
	logger.Debug("Navigating browser",
		zap.String("func", "MonitorService.NavigateBrowser"),
		zap.String("browser_id", browserID),
		zap.String("url", url))

	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if ms.browserMonitor == nil {
		logger.Warn("Browser monitor is not enabled",
			zap.String("func", "MonitorService.NavigateBrowser"),
			zap.String("browser_id", browserID),
			zap.String("url", url))
		return errors.NewAppError(http.StatusServiceUnavailable, "浏览器监控未启用", "BROWSER_MONITOR_NOT_ENABLED")
	}

	err := ms.browserMonitor.Navigate(browserID, url)
	if err != nil {
		logger.Error("Failed to navigate browser",
			zap.String("func", "MonitorService.NavigateBrowser"),
			zap.String("browser_id", browserID),
			zap.String("url", url),
			zap.Error(err))
		return errors.WrapError(err, "浏览器导航失败")
	}

	logger.Info("Browser navigated successfully",
		zap.String("func", "MonitorService.NavigateBrowser"),
		zap.String("browser_id", browserID),
		zap.String("url", url))

	return nil
}

// ExecuteBrowserAction 执行浏览器动作
func (ms *MonitorService) ExecuteBrowserAction(browserID string, action BrowserAction) (*BrowserActionResult, error) {
	logger.Debug("Executing browser action",
		zap.String("func", "MonitorService.ExecuteBrowserAction"),
		zap.String("browser_id", browserID),
		zap.String("action_type", action.Type))

	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if ms.browserMonitor == nil {
		logger.Warn("Browser monitor is not enabled",
			zap.String("func", "MonitorService.ExecuteBrowserAction"),
			zap.String("browser_id", browserID),
			zap.String("action_type", action.Type))
		return nil, errors.NewAppError(http.StatusServiceUnavailable, "浏览器监控未启用", "BROWSER_MONITOR_NOT_ENABLED")
	}

	result, err := ms.browserMonitor.ExecuteAction(browserID, action)
	if err != nil {
		logger.Error("Failed to execute browser action",
			zap.String("func", "MonitorService.ExecuteBrowserAction"),
			zap.String("browser_id", browserID),
			zap.String("action_type", action.Type),
			zap.Error(err))
		return nil, errors.WrapError(err, "执行浏览器动作失败")
	}

	logger.Debug("Browser action executed successfully",
		zap.String("func", "MonitorService.ExecuteBrowserAction"),
		zap.String("browser_id", browserID),
		zap.String("action_type", action.Type),
		zap.Bool("success", result.Success))

	return result, nil
}

// HandleCloudflareChallenge 处理Cloudflare挑战
func (ms *MonitorService) HandleCloudflareChallenge(targetURL string, httpClient *http.Client) (*CloudflareInfo, error) {
	logger.Debug("Handling Cloudflare challenge",
		zap.String("func", "MonitorService.HandleCloudflareChallenge"),
		zap.String("url", targetURL))

	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if ms.cfHandler == nil {
		logger.Warn("Cloudflare handler is not enabled",
			zap.String("func", "MonitorService.HandleCloudflareChallenge"),
			zap.String("url", targetURL))
		return nil, errors.NewAppError(http.StatusServiceUnavailable, "Cloudflare处理器未启用", "CLOUDFLARE_HANDLER_NOT_ENABLED")
	}

	info, err := ms.cfHandler.HandleChallenge(targetURL, httpClient)
	if err != nil {
		logger.Error("Failed to handle Cloudflare challenge",
			zap.String("func", "MonitorService.HandleCloudflareChallenge"),
			zap.String("url", targetURL),
			zap.Error(err))
		return nil, errors.WrapError(err, "处理Cloudflare挑战失败")
	}

	logger.Info("Cloudflare challenge handled successfully",
		zap.String("func", "MonitorService.HandleCloudflareChallenge"),
		zap.String("url", targetURL),
		zap.Bool("challenge_detected", info.ChallengeDetected),
		zap.Bool("solved", info.Solved))

	return info, nil
}

// GetSystemInfo 获取系统信息
func (ms *MonitorService) GetSystemInfo() (*SystemInfo, error) {
	logger.Debug("Getting system info",
		zap.String("func", "MonitorService.GetSystemInfo"))

	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if ms.systemMonitor == nil {
		logger.Warn("System monitor is not enabled",
			zap.String("func", "MonitorService.GetSystemInfo"))
		return nil, errors.NewAppError(http.StatusServiceUnavailable, "系统监控未启用", "SYSTEM_MONITOR_NOT_ENABLED")
	}

	info, err := ms.systemMonitor.GetSystemInfo()
	if err != nil {
		logger.Error("Failed to get system info",
			zap.String("func", "MonitorService.GetSystemInfo"),
			zap.Error(err))
		return nil, errors.WrapError(err, "获取系统信息失败")
	}

	logger.Debug("System info retrieved successfully",
		zap.String("func", "MonitorService.GetSystemInfo"),
		zap.String("hostname", info.Hostname),
		zap.String("os", info.OS))

	return info, nil
}

// GetActiveAlerts 获取活跃告警
func (ms *MonitorService) GetActiveAlerts() ([]Alert, error) {
	logger.Debug("Getting active alerts",
		zap.String("func", "MonitorService.GetActiveAlerts"))

	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if ms.systemMonitor == nil {
		logger.Warn("System monitor is not enabled",
			zap.String("func", "MonitorService.GetActiveAlerts"))
		return nil, errors.NewAppError(http.StatusServiceUnavailable, "系统监控未启用", "SYSTEM_MONITOR_NOT_ENABLED")
	}

	alerts := ms.systemMonitor.alertManager.GetActiveAlerts()
	logger.Debug("Active alerts retrieved successfully",
		zap.String("func", "MonitorService.GetActiveAlerts"),
		zap.Int("alerts_count", len(alerts)))

	return alerts, nil
}

// GetAlertHistory 获取告警历史
func (ms *MonitorService) GetAlertHistory(limit int) ([]Alert, error) {
	logger.Debug("Getting alert history",
		zap.String("func", "MonitorService.GetAlertHistory"),
		zap.Int("limit", limit))

	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if ms.systemMonitor == nil {
		logger.Warn("System monitor is not enabled",
			zap.String("func", "MonitorService.GetAlertHistory"),
			zap.Int("limit", limit))
		return nil, errors.NewAppError(http.StatusServiceUnavailable, "系统监控未启用", "SYSTEM_MONITOR_NOT_ENABLED")
	}

	alerts := ms.systemMonitor.alertManager.GetAlertHistory(limit)
	logger.Debug("Alert history retrieved successfully",
		zap.String("func", "MonitorService.GetAlertHistory"),
		zap.Int("limit", limit),
		zap.Int("alerts_count", len(alerts)))

	return alerts, nil
}

// ListBrowsers 列出浏览器实例
// ListBrowsers 列出浏览器实例
func (ms *MonitorService) ListBrowsers() ([]string, error) {
	logger.Debug("Listing browser instances",
		zap.String("func", "MonitorService.ListBrowsers"))

	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if ms.browserMonitor == nil {
		logger.Warn("Browser monitor is not enabled",
			zap.String("func", "MonitorService.ListBrowsers"))
		return nil, errors.NewAppError(http.StatusServiceUnavailable, "浏览器监控未启用", "BROWSER_MONITOR_NOT_ENABLED")
	}

	browserInstances := ms.browserMonitor.ListBrowsers()
	browserIDs := make([]string, 0, len(browserInstances))
	for id := range browserInstances {
		browserIDs = append(browserIDs, id)
	}

	logger.Debug("Browser instances listed successfully",
		zap.String("func", "MonitorService.ListBrowsers"),
		zap.Int("browsers_count", len(browserIDs)))

	return browserIDs, nil
}

// GetBrowserInstance 获取浏览器实例
func (ms *MonitorService) GetBrowserInstance(browserID string) (*BrowserInstance, error) {
	logger.Debug("Getting browser instance",
		zap.String("func", "MonitorService.GetBrowserInstance"),
		zap.String("browser_id", browserID))

	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if ms.browserMonitor == nil {
		logger.Warn("Browser monitor is not enabled",
			zap.String("func", "MonitorService.GetBrowserInstance"),
			zap.String("browser_id", browserID))
		return nil, errors.NewAppError(http.StatusServiceUnavailable, "浏览器监控未启用", "BROWSER_MONITOR_NOT_ENABLED")
	}

	instance, err := ms.browserMonitor.GetBrowserInstance(browserID)
	if err != nil {
		logger.Error("Failed to get browser instance",
			zap.String("func", "MonitorService.GetBrowserInstance"),
			zap.String("browser_id", browserID),
			zap.Error(err))
		return nil, errors.WrapError(err, "获取浏览器实例失败")
	}

	logger.Debug("Browser instance retrieved successfully",
		zap.String("func", "MonitorService.GetBrowserInstance"),
		zap.String("browser_id", browserID))

	return instance, nil
}

// IsRunning 检查服务是否运行
func (ms *MonitorService) IsRunning() bool {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	logger.Debug("Checking if service is running",
		zap.String("func", "MonitorService.IsRunning"),
		zap.Bool("running", ms.running))

	return ms.running
}

// UpdateConfig 更新配置
func (ms *MonitorService) UpdateConfig(config MonitorServiceConfig) error {
	logger.Debug("Updating monitor service configuration",
		zap.String("func", "MonitorService.UpdateConfig"),
		zap.Bool("enabled", config.Enabled),
		zap.Bool("system_monitor_enabled", config.SystemMonitor.Enabled),
		zap.Bool("browser_monitor_enabled", config.BrowserMonitor.Enabled),
		zap.Bool("cloudflare_enabled", config.Cloudflare.Enabled))

	ms.mu.Lock()
	defer ms.mu.Unlock()

	oldConfig := ms.config
	ms.config = config

	// 更新系统监控配置
	if ms.systemMonitor != nil {
		if err := ms.systemMonitor.UpdateConfig(config.SystemMonitor); err != nil {
			logger.Error("Failed to update system monitor config",
				zap.String("func", "MonitorService.UpdateConfig"),
				zap.Error(err))
			return errors.WrapError(err, "更新系统监控配置失败")
		}
	}

	// 更新浏览器监控配置
	if ms.browserMonitor != nil {
		if err := ms.browserMonitor.UpdateConfig(config.BrowserMonitor); err != nil {
			logger.Error("Failed to update browser monitor config",
				zap.String("func", "MonitorService.UpdateConfig"),
				zap.Error(err))
			return errors.WrapError(err, "更新浏览器监控配置失败")
		}
	}

	// 更新Cloudflare配置
	if ms.cfHandler != nil {
		ms.cfHandler.UpdateConfig(config.Cloudflare)
	}

	logger.Info("监控服务配置已更新",
		zap.String("func", "MonitorService.UpdateConfig"),
		zap.Bool("old_enabled", oldConfig.Enabled),
		zap.Bool("new_enabled", config.Enabled))

	return nil
}

// GetStatus 获取服务状态
func (ms *MonitorService) GetStatus() map[string]interface{} {
	logger.Debug("Getting monitor service status",
		zap.String("func", "MonitorService.GetStatus"))

	ms.mu.RLock()
	defer ms.mu.RUnlock()

	status := map[string]interface{}{
		"running":           ms.running,
		"enabled":           ms.config.Enabled,
		"system_monitor":    false,
		"browser_monitor":   false,
		"cloudflare":        false,
		"browser_instances": 0,
		"active_alerts":     0,
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
	logger.Debug("Getting monitor service health",
		zap.String("func", "MonitorService.GetHealth"))

	health := map[string]interface{}{
		"status":     "healthy",
		"timestamp":  time.Now(),
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
				"status": "unhealthy",
				"reason": "not_running",
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
