package utils

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"moviepilot-go/internal/logger"
	"go.uber.org/zap"
)

// CookieHelper Cookie助手结构�?type CookieHelper struct{}

// SiteLoginXPath 站点登录界面元素XPath
var SiteLoginXPath = map[string][]string{
	"username": {
		`input[name="username"]`,
		`input#form_item_username`,
		`input#username`,
	},
	"password": {
		`input[name="password"]`,
		`input#form_item_password`,
		`input#password`,
		`input[type="password"]`,
	},
	"captcha": {
		`input[name="imagestring"]`,
		`input[name="captcha"]`,
		`input#form_item_captcha`,
		`input[placeholder="驗證�?]`,
	},
	"captchaImg": {
		`img[alt="captcha"]`,
		`img[alt="CAPTCHA"]`,
		`img[alt="SECURITY CODE"]`,
		`img#LAY-user-get-vercode`,
		`img[src*="/api/getCaptcha"]`,
	},
	"submit": {
		`input[type="submit"]`,
		`button[type="submit"]`,
		`button[lay-filter="login"]`,
		`button[lay-filter="formLogin"]`,
		`input[type="button"][value="登录"]`,
	},
	"error": {
		`table.main td.text`,
	},
	"twostep": {
		`input[name="two_step_code"]`,
		`input[name="2fa_secret"]`,
		`input[name="otp"]`,
	},
}

// NewCookieHelper 创建新的Cookie助手实例
func NewCookieHelper() *CookieHelper {
	return &CookieHelper{}
}

// ParseCookies 将浏览器返回的cookies转化为字符串
func (ch *CookieHelper) ParseCookies(cookies []map[string]interface{}) string {
	/*
		将浏览器返回的cookies转化为字符串
	*/
	if len(cookies) == 0 {
		return ""
	}
	cookieStr := ""
	for _, cookie := range cookies {
		if name, ok := cookie["name"].(string); ok {
			if value, ok := cookie["value"].(string); ok {
				cookieStr += fmt.Sprintf("%s=%s; ", name, value)
			}
		}
	}
	return cookieStr
}

// GetSiteCookieUA 获取站点cookie和ua
func (ch *CookieHelper) GetSiteCookieUA(
	url string,
	username string,
	password string,
	twoStepCode string,
	proxies map[string]interface{},
	timeout int) (cookie string, ua string, message string) {
	/*
		获取站点cookie和ua
		:param url: 站点地址
		:param username: 用户�?		:param password: 密码
		:param twoStepCode: 二步验证码或密钥
		:param proxies: 代理
		:param timeout: 超时时间
		:return: cookie、ua、message
	*/

	// 参数检�?	if url == "" || username == "" || password == "" {
		return "", "", "参数错误"
	}

	// 使用BrowserHelper访问页面
	browserHelper := NewDefaultBrowserHelper()
	pageSource, err := browserHelper.GetPageSource(url, "", "", proxies, false, timeout)
	if err != nil {
		logger.GetLoggerManager().Error("获取页面源码失败", zap.Error(err))
		return "", "", fmt.Sprintf("获取页面源码失败: %v", err)
	}

	if pageSource == "" {
		return "", "", "获取源码失败"
	}

	// 解析HTML文档
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(pageSource))
	if err != nil {
		logger.GetLoggerManager().Error("解析HTML文档失败", zap.Error(err))
		return "", "", fmt.Sprintf("解析HTML文档失败: %v", err)
	}

	// 查找用户名输入框
	usernameSelector := ""
	for _, selector := range SiteLoginXPath["username"] {
		if doc.Find(selector).Length() > 0 {
			usernameSelector = selector
			break
		}
	}
	if usernameSelector == "" {
		return "", "", "未找到用户名输入�?
	}

	// 查找密码输入�?	passwordSelector := ""
	for _, selector := range SiteLoginXPath["password"] {
		if doc.Find(selector).Length() > 0 {
			passwordSelector = selector
			break
		}
	}
	if passwordSelector == "" {
		return "", "", "未找到密码输入框"
	}

	// 处理二步验证�?	// TODO: 实现TwoFactorAuth功能
	otpCode := ""
	if twoStepCode != "" {
		// 这里需要实现TwoFactorAuth(two_step_code).get_code()
		// 暂时留空，后续实�?		otpCode = twoStepCode
	}

	// 查找二步验证码输入框
	twostepSelector := ""
	if otpCode != "" {
		for _, selector := range SiteLoginXPath["twostep"] {
			if doc.Find(selector).Length() > 0 {
				twostepSelector = selector
				break
			}
		}
	}

	// 查找验证码输入框
	captchaSelector := ""
	for _, selector := range SiteLoginXPath["captcha"] {
		if doc.Find(selector).Length() > 0 {
			captchaSelector = selector
			break
		}
	}

	// 查找验证码图�?	captchaImgURL := ""
	if captchaSelector != "" {
		for _, selector := range SiteLoginXPath["captchaImg"] {
			selection := doc.Find(selector)
			if selection.Length() > 0 {
				// 获取图片URL
				if src, exists := selection.Attr("src"); exists {
					captchaImgURL = src
					break
				}
			}
		}
		if captchaImgURL == "" {
			return "", "", "未找到验证码图片"
		}
	}

	// 查找登录按钮
	submitSelector := ""
	for _, selector := range SiteLoginXPath["submit"] {
		if doc.Find(selector).Length() > 0 {
			submitSelector = selector
			break
		}
	}
	if submitSelector == "" {
		return "", "", "未找到登录按�?
	}

	// TODO: 实现完整的登录流�?	// 由于Go语言没有直接对应Playwright的库，我们需要使用HTTP客户端模拟登录过�?	// 这里暂时返回一个模拟结�?
	// 模拟登录过程
	logger.GetLoggerManager().Info("模拟登录过程",
		zap.String("url", url),
		zap.String("username", username))

	// 模拟返回结果
	// 注意：这只是一个模拟实现，实际需要使用HTTP客户端实现完整的登录流程
	return "cookie=example", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36", ""
}

// getCaptchaText 识别验证码图片的内容
func (ch *CookieHelper) getCaptchaText(cookie string, ua string, codeURL string) string {
	/*
		识别验证码图片的内容
	*/
	if codeURL == "" {
		return ""
	}

	// TODO: 实现验证码识别功�?	// 这需要实现类似RequestUtils和OcrHelper的功�?	logger.GetLoggerManager().Info("模拟验证码识�?,
		zap.String("code_url", codeURL))

	// 暂时返回空字符串，实际需要实现OCR识别功能
	return ""
}

// getCaptchaURL 获取验证码图片的URL
func (ch *CookieHelper) getCaptchaURL(siteURL string, imageURL string) string {
	/*
		获取验证码图片的URL
	*/
	if siteURL == "" || imageURL == "" {
		return ""
	}
	
	// 确保siteURL以http://或https://开�?	if !strings.HasPrefix(siteURL, "http://") && !strings.HasPrefix(siteURL, "https://") {
		return ""
	}
	
	// 提取基础URL (协议 + 域名 + 端口)
	var baseURL string
	if strings.HasPrefix(siteURL, "https://") {
		baseURL = "https://" + strings.Split(strings.TrimPrefix(siteURL, "https://"), "/")[0]
	} else {
		baseURL = "http://" + strings.Split(strings.TrimPrefix(siteURL, "http://"), "/")[0]
	}
	
	if strings.HasPrefix(imageURL, "/") {
		imageURL = imageURL[1:]
	}
	
	return fmt.Sprintf("%s/%s", baseURL, imageURL)
}
