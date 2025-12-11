package utils

import (
	"fmt"
	"time"

	"github.com/playwright-community/playwright-go"
	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// PlaywrightHelper Playwright浏览器自动化助手
type PlaywrightHelper struct {
	browserType string
	logger      *zap.Logger
}

// NewPlaywrightHelper 创建Playwright浏览器自动化助手
func NewPlaywrightHelper(browserType string) *PlaywrightHelper {
	return &PlaywrightHelper{
		browserType: browserType,
		logger:      logger.GetLogger(),
	}
}

// PassCloudflare 尝试跳过Cloudflare验证
func (p *PlaywrightHelper) PassCloudflare(url string) (string, error) {
	p.logger.Info("尝试跳过Cloudflare验证", zap.String("url", url))

	// 初始化Playwright
	pw, err := playwright.Run()
	if err != nil {
		return "", fmt.Errorf("初始化Playwright失败: %w", err)
	}
	defer pw.Stop()

	// 启动浏览器
	browser, err := pw.Chromium.Launch(
		playwright.BrowserTypeLaunchOptions{
			Headless: playwright.Bool(true),
		},
	)
	if err != nil {
		return "", fmt.Errorf("启动浏览器失败: %w", err)
	}
	defer browser.Close()

	// 创建页面
	page, err := browser.NewPage()
	if err != nil {
		return "", fmt.Errorf("创建页面失败: %w", err)
	}

	// 导航到目标URL
	if _, err := page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(30000),
	}); err != nil {
		return "", fmt.Errorf("导航到URL失败: %w", err)
	}

	// 等待页面加载完成
	time.Sleep(5 * time.Second)

	// 获取Cookie
	cookies, err := page.Context().Cookies()
	if err != nil {
		return "", fmt.Errorf("获取Cookie失败: %w", err)
	}

	// 构建Cookie字符串
	cookieStr := ""
	for _, cookie := range cookies {
		if cookie.Name != "" && cookie.Value != "" {
			if cookieStr != "" {
				cookieStr += "; "
			}
			cookieStr += fmt.Sprintf("%s=%s", cookie.Name, cookie.Value)
		}
	}

	p.logger.Info("Cloudflare验证跳过成功", zap.String("url", url))
	return cookieStr, nil
}

// GetPageContent 获取页面内容
func (p *PlaywrightHelper) GetPageContent(url string) (string, error) {
	p.logger.Info("获取页面内容", zap.String("url", url))

	// 初始化Playwright
	pw, err := playwright.Run()
	if err != nil {
		return "", fmt.Errorf("初始化Playwright失败: %w", err)
	}
	defer pw.Stop()

	// 启动浏览器
	browser, err := pw.Chromium.Launch(
		playwright.BrowserTypeLaunchOptions{
			Headless: playwright.Bool(true),
		},
	)
	if err != nil {
		return "", fmt.Errorf("启动浏览器失败: %w", err)
	}
	defer browser.Close()

	// 创建页面
	page, err := browser.NewPage()
	if err != nil {
		return "", fmt.Errorf("创建页面失败: %w", err)
	}

	// 导航到目标URL
	if _, err := page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(30000),
	}); err != nil {
		return "", fmt.Errorf("导航到URL失败: %w", err)
	}

	// 等待页面加载完成
	time.Sleep(3 * time.Second)

	// 获取页面内容
	content, err := page.Content()
	if err != nil {
		return "", fmt.Errorf("获取页面内容失败: %w", err)
	}

	p.logger.Info("页面内容获取成功", zap.String("url", url))
	return content, nil
}
