package monitor

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"go.uber.org/zap"
)

// CloudflareHandler Cloudflare挑战处理器
type CloudflareHandler struct {
	config   CloudflareConfig
	logger   *zap.Logger
	browser  *BrowserMonitor
}

// CloudflareConfig Cloudflare配置
type CloudflareConfig struct {
	Enabled          bool          `json:"enabled"`
	Timeout          time.Duration `json:"timeout"`
	MaxRetries       int           `json:"max_retries"`
	UserAgent        string        `json:"user_agent"`
	WaitForJS        time.Duration `json:"wait_for_js"`
	CheckInterval    time.Duration `json:"check_interval"`
	SolveMethod      string        `json:"solve_method"` // "browser", "api"
	JavaScriptEnabled bool         `json:"javascript_enabled"`
}

// NewCloudflareHandler 创建Cloudflare处理器
func NewCloudflareHandler(config CloudflareConfig, logger *zap.Logger) *CloudflareHandler {
	return &CloudflareHandler{
		config: config,
		logger: logger,
	}
}

// SetBrowserMonitor 设置浏览器监控器
func (cf *CloudflareHandler) SetBrowserMonitor(browser *BrowserMonitor) {
	cf.browser = browser
}

// HandleChallenge 处理Cloudflare挑战
func (cf *CloudflareHandler) HandleChallenge(targetURL string, httpClient *http.Client) (*CloudflareInfo, error) {
	if !cf.config.Enabled {
		return nil, fmt.Errorf("Cloudflare挑战处理已禁用")
	}
	
	cf.logger.Info("开始处理Cloudflare挑战", zap.String("url", targetURL))
	
	info := &CloudflareInfo{
		ChallengeDetected: false,
		Solved:            false,
		RequiredAction:    "",
		ChallengeType:     "",
		UserAgent:         cf.config.UserAgent,
		Cookies:           make(map[string]string),
		Headers:           make(map[string]string),
	}
	
	// 检查是否存在挑战
	challengeDetected, challengeType, err := cf.detectChallenge(targetURL, httpClient)
	if err != nil {
		return info, fmt.Errorf("检测挑战失败: %w", err)
	}
	
	info.ChallengeDetected = challengeDetected
	info.ChallengeType = challengeType
	
	if !challengeDetected {
		cf.logger.Info("未检测到Cloudflare挑战")
		return info, nil
	}
	
	cf.logger.Info("检测到Cloudflare挑战", zap.String("type", challengeType))
	
	// 尝试解决挑战
	switch cf.config.SolveMethod {
	case "browser":
		err = cf.solveWithBrowser(targetURL, info)
	case "api":
		err = cf.solveWithAPI(targetURL, info)
	default:
		return info, fmt.Errorf("不支持的解决方法: %s", cf.config.SolveMethod)
	}
	
	if err != nil {
		return info, fmt.Errorf("解决挑战失败: %w", err)
	}
	
	return info, nil
}

// detectChallenge 检测Cloudflare挑战
func (cf *CloudflareHandler) detectChallenge(targetURL string, httpClient *http.Client) (bool, string, error) {
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return false, "", err
	}
	
	if cf.config.UserAgent != "" {
		req.Header.Set("User-Agent", cf.config.UserAgent)
	}
	
	resp, err := httpClient.Do(req)
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()
	
	// 检查响应头
	cfHeaders := []string{
		"cf-ray",
		"cf-cache-status",
		"server", // 通常为 "cloudflare"
	}
	
	hasCFHeaders := false
	for _, header := range cfHeaders {
		if resp.Header.Get(header) != "" {
			hasCFHeaders = true
			break
		}
	}
	
	// 读取响应体
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return hasCFHeaders, "header_only", nil // 至少检测到CF头
	}
	
	// 检查Cloudflare相关元素
	challengeIndicators := []string{
		"cf-browser-verification",
		"cf-im-under-attack",
		"challenge-form",
		"cf-challenge-running",
		"turnstile-wrapper",
		"cf-turnstile",
	}
	
	var detectedTypes []string
	doc.Find("*").Each(func(i int, s *goquery.Selection) {
		class, _ := s.Attr("class")
		id, _ := s.Attr("id")
		
		for _, indicator := range challengeIndicators {
			if strings.Contains(class, indicator) || strings.Contains(id, indicator) {
				detectedTypes = append(detectedTypes, indicator)
				return
			}
		}
	})
	
	// 检查标题
	title := doc.Find("title").Text()
	if strings.Contains(strings.ToLower(title), "cloudflare") {
		detectedTypes = append(detectedTypes, "title")
	}
	
	// 检查脚本内容
	html, err := doc.Html()
	if err == nil {
		cfPatterns := []string{
			`cf-spinner`,
			`challenge-platform`,
			`turnstile`,
			`cf_chl_rc`,
		}
		
		for _, pattern := range cfPatterns {
			if matched, _ := regexp.MatchString(pattern, html); matched {
				detectedTypes = append(detectedTypes, "script_"+pattern)
			}
		}
	}
	
	if len(detectedTypes) == 0 {
		return hasCFHeaders, "header_only", nil
	}
	
	challengeType := "unknown"
	for _, t := range detectedTypes {
		if strings.Contains(t, "turnstile") {
			challengeType = "turnstile"
			break
		} else if strings.Contains(t, "challenge") {
			challengeType = "js_challenge"
			break
		} else if strings.Contains(t, "attack") {
			challengeType = "under_attack"
			break
		}
	}
	
	return true, challengeType, nil
}

// solveWithBrowser 使用浏览器解决挑战
func (cf *CloudflareHandler) solveWithBrowser(targetURL string, info *CloudflareInfo) error {
	if cf.browser == nil {
		return fmt.Errorf("浏览器监控器未设置")
	}
	
	browserID := fmt.Sprintf("cf_solve_%d", time.Now().Unix())
	instance, err := cf.browser.CreateBrowser(browserID)
	if err != nil {
		return fmt.Errorf("创建浏览器实例失败: %w", err)
	}
	defer cf.browser.CloseBrowser(browserID)
	
	startTime := time.Now()
	
	// 导航到目标URL
	err = cf.browser.Navigate(browserID, targetURL)
	if err != nil {
		return fmt.Errorf("导航失败: %w", err)
	}
	
	cf.logger.Info("等待Cloudflare挑战解决", zap.String("browser_id", browserID))
	
	// 等待挑战解决
	maxWait := cf.config.Timeout
	checkInterval := cf.config.CheckInterval
	
	for {
		if time.Since(startTime) > maxWait {
			return fmt.Errorf("解决挑战超时")
		}
		
		// 检查是否还在挑战页面
		challengeResult, err := cf.browser.ExecuteAction(browserID, BrowserAction{
			Type:    "get_title",
			Timeout: 5 * time.Second,
		})
		
		if err != nil {
			time.Sleep(checkInterval)
			continue
		}
		
		title := challengeResult.Title
		
		// 检查是否成功通过挑战
		if !cf.isChallengePage(title) {
			// 获取最终的Cookie
			cf.logger.Info("Cloudflare挑战已解决", 
				zap.String("browser_id", browserID),
				zap.Duration("solve_time", time.Since(startTime)))
			
			info.Solved = true
			info.SolveTime = time.Since(startTime)
			info.RequiredAction = "browser_solve"
			
			// 这里可以提取Cookie和Headers
			// 实际实现中需要通过CDP获取Cookie
			
			break
		}
		
		time.Sleep(checkInterval)
	}
	
	return nil
}

// solveWithAPI 使用API解决挑战
func (cf *CloudflareHandler) solveWithAPI(targetURL string, info *CloudflareInfo) error {
	// API解决通常需要第三方服务
	// 这里提供一个框架，实际实现需要集成具体的解决服务
	
	cf.logger.Warn("API解决Cloudflare挑战暂未实现")
	
	info.RequiredAction = "api_solve_not_implemented"
	info.Solved = false
	
	return fmt.Errorf("API解决方法尚未实现")
}

// isChallengePage 检查是否是挑战页面
func (cf *CloudflareHandler) isChallengePage(title string) bool {
	title = strings.ToLower(title)
	
	challengeKeywords := []string{
		"cloudflare",
		"checking your browser",
		"just a moment",
		"attention required",
		"verify you are human",
		"browser check",
		"ddos protection",
	}
	
	for _, keyword := range challengeKeywords {
		if strings.Contains(title, keyword) {
			return true
		}
	}
	
	return false
}

// GetOptimalUserAgent 获取最优的User-Agent
func (cf *CloudflareHandler) GetOptimalUserAgent() string {
	if cf.config.UserAgent != "" {
		return cf.config.UserAgent
	}
	
	// 返回一个常用的现代浏览器User-Agent
	return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
}

// BuildHTTPClient 构建HTTP客户端
func (cf *CloudflareHandler) BuildHTTPClient(cookies map[string]string, headers map[string]string) *http.Client {
	client := &http.Client{
		Timeout: cf.config.Timeout,
	}
	
	// 创建Transport来处理Cookie和Headers
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}
	
	client.Transport = transport
	
	return client
}

// ValidateChallengeSolution 验证挑战解决方案
func (cf *CloudflareHandler) ValidateChallengeSolution(targetURL string, cookies map[string]string) (bool, error) {
	client := cf.BuildHTTPClient(cookies, cf.config.Headers)
	
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return false, err
	}
	
	// 设置Cookie
	for name, value := range cookies {
		req.AddCookie(&http.Cookie{Name: name, Value: value})
	}
	
	if cf.config.UserAgent != "" {
		req.Header.Set("User-Agent", cf.config.UserAgent)
	}
	
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	
	// 检查是否还返回挑战页面
	challengeDetected, _, err := cf.detectChallenge(targetURL, client)
	return !challengeDetected, err
}

// GetChallengeType 获取挑战类型描述
func (cf *CloudflareHandler) GetChallengeType(challengeType string) string {
	switch challengeType {
	case "js_challenge":
		return "JavaScript挑战"
	case "turnstile":
		return "Turnstile验证"
	case "under_attack":
		return "DDoS保护模式"
	case "captcha":
		return "CAPTCHA验证"
	case "header_only":
		return "Cloudflare代理"
	default:
		return "未知挑战类型"
	}
}

// UpdateConfig 更新配置
func (cf *CloudflareHandler) UpdateConfig(config CloudflareConfig) {
	cf.config = config
	
	cf.logger.Info("更新Cloudflare处理器配置",
		zap.Bool("enabled", config.Enabled),
		zap.Duration("timeout", config.Timeout),
		zap.Int("max_retries", config.MaxRetries),
		zap.String("solve_method", config.SolveMethod))
}