package monitor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
	"go.uber.org/zap"
)

// BrowserMonitor 浏览器监控器
type BrowserMonitor struct {
	config   BrowserMonitorConfig
	logger   *zap.Logger
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	running  bool
	mu       sync.RWMutex
	
	// 浏览器实例
	browsers map[string]*BrowserInstance
	
	// 指标缓存
	lastMetrics map[string]BrowserMetrics
	lastUpdate  time.Time
}

// BrowserMonitorConfig 浏览器监控配置
type BrowserMonitorConfig struct {
	Enabled          bool          `json:"enabled"`
	Interval         time.Duration `json:"interval"`
	Timeout          time.Duration `json:"timeout"`
	Headless         bool          `json:"headless"`
	UserAgent        string        `json:"user_agent"`
	WindowSize       WindowSize    `json:"window_size"`
	Proxy            string        `json:"proxy,omitempty"`
	DisableImages    bool          `json:"disable_images"`
	DisableJavaScript bool         `json:"disable_javascript"`
}

// WindowSize 窗口大小
type WindowSize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// BrowserInstance 浏览器实例
type BrowserInstance struct {
	ID          string
	AllocCtx    context.Context
	CancelCtx   context.CancelFunc
	BrowserCtx  context.Context
	TabCtx      context.Context
	IsRunning   bool
	StartTime   time.Time
	CurrentURL  string
	WindowTitle string
	Metrics     BrowserMetrics
	LastUpdate  time.Time
}

// NewBrowserMonitor 创建浏览器监控器
func NewBrowserMonitor(config BrowserMonitorConfig, logger *zap.Logger) *BrowserMonitor {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &BrowserMonitor{
		config:      config,
		logger:      logger,
		ctx:         ctx,
		cancel:      cancel,
		browsers:    make(map[string]*BrowserInstance),
		lastMetrics: make(map[string]BrowserMetrics),
	}
}

// Start 启动浏览器监控
func (bm *BrowserMonitor) Start() error {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	
	if bm.running {
		return fmt.Errorf("浏览器监控已经在运行")
	}
	
	if !bm.config.Enabled {
		bm.logger.Info("浏览器监控已禁用")
		return nil
	}
	
	bm.logger.Info("启动浏览器监控器",
		zap.Duration("interval", bm.config.Interval),
		zap.Bool("headless", bm.config.Headless),
		zap.Duration("timeout", bm.config.Timeout))
	
	// 启动指标收集循环
	bm.wg.Add(1)
	go bm.collectLoop()
	
	bm.running = true
	return nil
}

// Stop 停止浏览器监控
func (bm *BrowserMonitor) Stop() error {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	
	if !bm.running {
		return nil
	}
	
	bm.logger.Info("停止浏览器监控器")
	
	// 停止所有浏览器实例
	for _, instance := range bm.browsers {
		bm.closeBrowser(instance.ID)
	}
	
	bm.cancel()
	bm.wg.Wait()
	
	bm.running = false
	return nil
}

// collectLoop 指标收集循环
func (bm *BrowserMonitor) collectLoop() {
	defer bm.wg.Done()
	
	ticker := time.NewTicker(bm.config.Interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-bm.ctx.Done():
			return
		case <-ticker.C:
			bm.collectBrowserMetrics()
		}
	}
}

// CreateBrowser 创建浏览器实例
func (bm *BrowserMonitor) CreateBrowser(browserID string) (*BrowserInstance, error) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	
	if _, exists := bm.browsers[browserID]; exists {
		return nil, fmt.Errorf("浏览器实例已存在: %s", browserID)
	}
	
	// 创建浏览器上下文
	allocCtx, cancelCtx := chromedp.NewExecAllocator(bm.ctx,
		chromedp.Flag("headless", bm.config.Headless),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("disable-features", "VizDisplayCompositor"),
		chromedp.Flag("window-size", fmt.Sprintf("%d,%d", bm.config.WindowSize.Width, bm.config.WindowSize.Height)),
	)
	
	if bm.config.UserAgent != "" {
		allocCtx, cancelCtx = chromedp.NewExecAllocator(bm.ctx,
			chromedp.Flag("headless", bm.config.Headless),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-sandbox", true),
			chromedp.Flag("disable-dev-shm-usage", true),
			chromedp.Flag("disable-setuid-sandbox", true),
			chromedp.Flag("disable-web-security", true),
			chromedp.Flag("disable-features", "VizDisplayCompositor"),
			chromedp.Flag("window-size", fmt.Sprintf("%d,%d", bm.config.WindowSize.Width, bm.config.WindowSize.Height)),
			chromedp.UserAgent(bm.config.UserAgent),
		)
	}
	
	browserCtx, _ := chromedp.NewContext(allocCtx)
	tabCtx, _ := chromedp.NewContext(browserCtx)
	
	instance := &BrowserInstance{
		ID:         browserID,
		AllocCtx:   allocCtx,
		CancelCtx:  cancelCtx,
		BrowserCtx: browserCtx,
		TabCtx:     tabCtx,
		IsRunning:  true,
		StartTime:  time.Now(),
	}
	
	bm.browsers[browserID] = instance
	
	bm.logger.Info("创建浏览器实例", zap.String("browser_id", browserID))
	return instance, nil
}

// CloseBrowser 关闭浏览器实例
func (bm *BrowserMonitor) CloseBrowser(browserID string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	
	return bm.closeBrowser(browserID)
}

// closeBrowser 内部关闭浏览器方法
func (bm *BrowserMonitor) closeBrowser(browserID string) error {
	instance, exists := bm.browsers[browserID]
	if !exists {
		return fmt.Errorf("浏览器实例不存在: %s", browserID)
	}
	
	if instance.CancelCtx != nil {
		instance.CancelCtx()
	}
	
	instance.IsRunning = false
	delete(bm.browsers, browserID)
	delete(bm.lastMetrics, browserID)
	
	bm.logger.Info("关闭浏览器实例", zap.String("browser_id", browserID))
	return nil
}

// Navigate 导航到URL
func (bm *BrowserMonitor) Navigate(browserID string, url string) error {
	bm.mu.RLock()
	instance, exists := bm.browsers[browserID]
	bm.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("浏览器实例不存在: %s", browserID)
	}
	
	if !instance.IsRunning {
		return fmt.Errorf("浏览器实例未运行: %s", browserID)
	}
	
	ctx, cancel := context.WithTimeout(instance.TabCtx, bm.config.Timeout)
	defer cancel()
	
	var title string
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Title(&title),
	)
	
	if err != nil {
		return fmt.Errorf("导航失败: %w", err)
	}
	
	// 更新实例状态
	instance.CurrentURL = url
	instance.WindowTitle = title
	
	bm.logger.Info("浏览器导航完成",
		zap.String("browser_id", browserID),
		zap.String("url", url),
		zap.String("title", title))
	
	return nil
}

// ExecuteAction 执行浏览器动作
func (bm *BrowserMonitor) ExecuteAction(browserID string, action BrowserAction) (*BrowserActionResult, error) {
	start := time.Now()
	
	result := &BrowserActionResult{
		Success:  false,
		Message:  "动作执行失败",
		Duration: time.Since(start),
	}
	
	bm.mu.RLock()
	instance, exists := bm.browsers[browserID]
	bm.mu.RUnlock()
	
	if !exists {
		result.Error = fmt.Sprintf("浏览器实例不存在: %s", browserID)
		return result, result.Error
	}
	
	if !instance.IsRunning {
		result.Error = fmt.Sprintf("浏览器实例未运行: %s", browserID)
		return result, result.Error
	}
	
	ctx, cancel := context.WithTimeout(instance.TabCtx, action.Timeout)
	defer cancel()
	
	switch action.Type {
	case "click":
		err := chromedp.Run(ctx, chromedp.Click(action.Selector))
		if err == nil {
			result.Success = true
			result.Message = "点击成功"
		} else {
			result.Error = err.Error()
		}
		
	case "input":
		err := chromedp.Run(ctx,
			chromedp.SendKeys(action.Selector, action.Value),
		)
		if err == nil {
			result.Success = true
			result.Message = "输入成功"
		} else {
			result.Error = err.Error()
		}
		
	case "wait":
		time.Sleep(action.WaitTime)
		result.Success = true
		result.Message = "等待完成"
		
	case "screenshot":
		var screenshot []byte
		err := chromedp.Run(ctx, chromedp.CaptureScreenshot(&screenshot))
		if err == nil {
			result.Success = true
			result.Message = "截图成功"
			result.Screenshot = fmt.Sprintf("data:image/png;base64,%s", screenshot)
		} else {
			result.Error = err.Error()
		}
		
	case "get_text":
		var text string
		err := chromedp.Run(ctx, chromedp.Text(action.Selector, &text))
		if err == nil {
			result.Success = true
			result.Message = "获取文本成功"
			if result.Data == nil {
				result.Data = make(map[string]interface{})
			}
			result.Data["text"] = text
		} else {
			result.Error = err.Error()
		}
		
	case "get_title":
		var title string
		err := chromedp.Run(ctx, chromedp.Title(&title))
		if err == nil {
			result.Success = true
			result.Message = "获取标题成功"
			result.Title = title
		} else {
			result.Error = err.Error()
		}
		
	case "get_url":
		var url string
		err := chromedp.Run(ctx, chromedp.Location(&url))
		if err == nil {
			result.Success = true
			result.Message = "获取URL成功"
			result.URL = url
		} else {
			result.Error = err.Error()
		}
		
	default:
		result.Error = fmt.Sprintf("不支持的动作类型: %s", action.Type)
	}
	
	result.Duration = time.Since(start)
	
	bm.logger.Info("浏览器动作执行完成",
		zap.String("browser_id", browserID),
		zap.String("action_type", action.Type),
		zap.Bool("success", result.Success),
		zap.Duration("duration", result.Duration))
	
	return result, result.Error
}

// collectBrowserMetrics 收集浏览器指标
func (bm *BrowserMonitor) collectBrowserMetrics() {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	
	for browserID, instance := range bm.browsers {
		if !instance.IsRunning {
			continue
		}
		
		metrics, err := bm.collectInstanceMetrics(instance)
		if err != nil {
			bm.logger.Error("收集浏览器指标失败",
				zap.String("browser_id", browserID),
				zap.Error(err))
			continue
		}
		
		instance.Metrics = *metrics
		instance.LastUpdate = time.Now()
		bm.lastMetrics[browserID] = *metrics
	}
	
	bm.lastUpdate = time.Now()
}

// collectInstanceMetrics 收集实例指标
func (bm *BrowserMonitor) collectInstanceMetrics(instance *BrowserInstance) (*BrowserMetrics, error) {
	metrics := &BrowserMetrics{
		BrowserID:   instance.ID,
		IsRunning:   instance.IsRunning,
		WindowTitle: instance.WindowTitle,
		URL:         instance.CurrentURL,
		StartTime:   instance.StartTime,
	}
	
	ctx, cancel := context.WithTimeout(instance.TabCtx, 5*time.Second)
	defer cancel()
	
	// 获取性能指标
	var performanceInfo struct {
		FCP  float64 `json:"firstContentfulPaint"`
		LCP  float64 `json:"largestContentfulPaint"`
		FID  float64 `json:"firstInputDelay"`
		CLS  float64 `json:"cumulativeLayoutShift"`
		TTFB float64 `json:"timeToFirstByte"`
	}
	
	err := chromedp.Evaluate(ctx, `
		performance.getEntriesByType('navigation').forEach(entry => {
			window.perfData = {
				firstContentfulPaint: entry.loadEventEnd - entry.fetchStart,
				largestContentfulPaint: 0,
				firstInputDelay: 0,
				cumulativeLayoutShift: 0,
				timeToFirstByte: entry.responseStart - entry.fetchStart
			};
		});
		
		// 获取 LCP
		if ('PerformanceObserver' in window) {
			new PerformanceObserver((list) => {
				const entries = list.getEntries();
				if (entries.length > 0) {
					window.perfData.largestContentfulPaint = entries[entries.length - 1].renderTime || entries[entries.length - 1].loadTime;
				}
			}).observe({ entryTypes: ['largest-contentful-paint'] });
		}
		
		window.perfData || {};
	`, &performanceInfo)
	
	if err == nil {
		metrics.Performance.FCP = performanceInfo.FCP
		metrics.Performance.LCP = performanceInfo.LCP
		metrics.Performance.FID = performanceInfo.FID
		metrics.Performance.CLS = performanceInfo.CLS
		metrics.Performance.TTFB = performanceInfo.TTFB
	}
	
	// 获取内存指标
	var memoryInfo struct {
		UsedJSHeapSize   float64 `json:"usedJSHeapSize"`
		TotalJSHeapSize  float64 `json:"totalJSHeapSize"`
		JSHeapSizeLimit  float64 `json:"jsHeapSizeLimit"`
	}
	
	err = chromedp.Evaluate(ctx, `performance.memory || {}`, &memoryInfo)
	if err == nil {
		metrics.Javascript.JSHeapUsedSize = uint64(memoryInfo.UsedJSHeapSize)
		metrics.Javascript.JSHeapTotalSize = uint64(memoryInfo.TotalJSHeapSize)
		metrics.Javascript.JSHeapSizeLimit = uint64(memoryInfo.JSHeapSizeLimit)
	}
	
	return metrics, nil
}

// GetMetrics 获取浏览器指标
func (bm *BrowserMonitor) GetMetrics() map[string]BrowserMetrics {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	
	result := make(map[string]BrowserMetrics)
	for k, v := range bm.lastMetrics {
		result[k] = v
	}
	return result
}

// GetBrowserInstance 获取浏览器实例
func (bm *BrowserMonitor) GetBrowserInstance(browserID string) (*BrowserInstance, error) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	
	instance, exists := bm.browsers[browserID]
	if !exists {
		return nil, fmt.Errorf("浏览器实例不存在: %s", browserID)
	}
	
	return instance, nil
}

// ListBrowsers 列出所有浏览器实例
func (bm *BrowserMonitor) ListBrowsers() []string {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	
	var browserIDs []string
	for id := range bm.browsers {
		browserIDs = append(browserIDs, id)
	}
	
	return browserIDs
}

// IsRunning 检查监控器是否运行
func (bm *BrowserMonitor) IsRunning() bool {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	return bm.running
}

// UpdateConfig 更新配置
func (bm *BrowserMonitor) UpdateConfig(config BrowserMonitorConfig) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	
	oldConfig := bm.config
	bm.config = config
	
	bm.logger.Info("更新浏览器监控配置",
		zap.Bool("old_enabled", oldConfig.Enabled),
		zap.Bool("new_enabled", config.Enabled),
		zap.Duration("old_interval", oldConfig.Interval),
		zap.Duration("new_interval", config.Interval))
	
	return nil
}