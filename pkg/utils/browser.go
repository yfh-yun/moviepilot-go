package utils

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/chromedp/chromedp"
)

// BrowserHelper 浏览器辅助工具
type BrowserHelper struct {
	client *http.Client
}

// NewBrowserHelper 创建浏览器辅助工具实例
func NewBrowserHelper() *BrowserHelper {
	return &BrowserHelper{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Navigate 导航到指定URL
func (b *BrowserHelper) Navigate(ctx context.Context, url string) error {
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	return chromedp.Run(ctx, chromedp.Navigate(url))
}

// GetPageTitle 获取页面标题
func (b *BrowserHelper) GetPageTitle(ctx context.Context, url string) (string, error) {
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	var title string
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Title(&title),
	)
	if err != nil {
		return "", fmt.Errorf("获取页面标题失败: %w", err)
	}
	return title, nil
}

// GetPageHTML 获取页面HTML
func (b *BrowserHelper) GetPageHTML(ctx context.Context, url string) (string, error) {
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	var html string
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.OuterHTML("html", &html),
	)
	if err != nil {
		return "", fmt.Errorf("获取页面HTML失败: %w", err)
	}
	return html, nil
}

// ClickElement 点击元素
func (b *BrowserHelper) ClickElement(ctx context.Context, selector string) error {
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	err := chromedp.Run(ctx, chromedp.Click(selector))
	if err != nil {
		return fmt.Errorf("点击元素失败: %w", err)
	}
	return nil
}

// InputText 输入文本
func (b *BrowserHelper) InputText(ctx context.Context, selector, text string) error {
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	err := chromedp.Run(ctx, chromedp.SendKeys(selector, text))
	if err != nil {
		return fmt.Errorf("输入文本失败: %w", err)
	}
	return nil
}

// WaitForElement 等待元素出现
func (b *BrowserHelper) WaitForElement(ctx context.Context, selector string) error {
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	err := chromedp.Run(ctx, chromedp.WaitVisible(selector))
	if err != nil {
		return fmt.Errorf("等待元素失败: %w", err)
	}
	return nil
}

// GetElementText 获取元素文本
func (b *BrowserHelper) GetElementText(ctx context.Context, selector string) (string, error) {
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	var text string
	err := chromedp.Run(ctx, chromedp.Text(selector, &text))
	if err != nil {
		return "", fmt.Errorf("获取元素文本失败: %w", err)
	}
	return text, nil
}

// Screenshot 截图
func (b *BrowserHelper) Screenshot(ctx context.Context, url string) ([]byte, error) {
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	var screenshot []byte
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.FullScreenshot(&screenshot, 100),
	)
	if err != nil {
		return nil, fmt.Errorf("截图失败: %w", err)
	}
	return screenshot, nil
}

// ExecuteJavaScript 执行JavaScript
func (b *BrowserHelper) ExecuteJavaScript(ctx context.Context, script string, result interface{}) error {
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	err := chromedp.Run(ctx, chromedp.Evaluate(script, result))
	if err != nil {
		return fmt.Errorf("执行JavaScript失败: %w", err)
	}
	return nil
}