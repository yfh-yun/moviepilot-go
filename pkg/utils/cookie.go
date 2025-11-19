package utils

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

// CookieHelper Cookie辅助工具
type CookieHelper struct {
	jar *http.CookieJar
}

// NewCookieHelper 创建Cookie辅助工具实例
func NewCookieHelper() *CookieHelper {
	jar, _ := cookiejar.New(nil)
	return &CookieHelper{
		jar: jar,
	}
}

// SetCookie 设置Cookie
func (c *CookieHelper) SetCookie(client *http.Client, rawURL, name, value string, domain string, path string, expires time.Time, secure bool, httpOnly bool) {
	u, _ := url.Parse(rawURL)
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Domain:   domain,
		Path:     path,
		Expires:  expires,
		Secure:   secure,
		HttpOnly: httpOnly,
	}
	client.Jar.SetCookies(u, []*http.Cookie{cookie})
}

// GetCookie 获取Cookie
func (c *CookieHelper) GetCookie(client *http.Client, rawURL, name string) (*http.Cookie, error) {
	u, _ := url.Parse(rawURL)
	cookies := client.Jar.Cookies(u)
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie, nil
		}
	}
	return nil, nil
}

// GetAllCookies 获取所有Cookie
func (c *CookieHelper) GetAllCookies(client *http.Client, rawURL string) []*http.Cookie {
	u, _ := url.Parse(rawURL)
	return client.Jar.Cookies(u)
}

// DeleteCookie 删除Cookie
func (c *CookieHelper) DeleteCookie(client *http.Client, rawURL, name string) {
	u, _ := url.Parse(rawURL)
	cookies := client.Jar.Cookies(u)
	var newCookies []*http.Cookie
	for _, cookie := range cookies {
		if cookie.Name != name {
			newCookies = append(newCookies, cookie)
		}
	}
	client.Jar.SetCookies(u, newCookies)
}

// ClearCookies 清除所有Cookie
func (c *CookieHelper) ClearCookies(client *http.Client, rawURL string) {
	u, _ := url.Parse(rawURL)
	client.Jar.SetCookies(u, []*http.Cookie{})
}

// ParseCookieString 解析Cookie字符串
func (c *CookieHelper) ParseCookieString(cookieString string) []*http.Cookie {
	var cookies []*http.Cookie
	parts := strings.Split(cookieString, ";")
	for _, part := range parts {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) == 2 {
			cookies = append(cookies, &http.Cookie{
				Name:  kv[0],
				Value: kv[1],
			})
		}
	}
	return cookies
}

// CookieToString 将Cookie转换为字符串
func (c *CookieHelper) CookieToString(cookies []*http.Cookie) string {
	var parts []string
	for _, cookie := range cookies {
		parts = append(parts, cookie.Name+"="+cookie.Value)
	}
	return strings.Join(parts, "; ")
}

// IsCookieExpired 检查Cookie是否过期
func (c *CookieHelper) IsCookieExpired(cookie *http.Cookie) bool {
	if cookie.Expires.IsZero() {
		return false // 会话Cookie不会过期
	}
	return time.Now().After(cookie.Expires)
}

// FilterExpiredCookies 过滤过期的Cookie
func (c *CookieHelper) FilterExpiredCookies(cookies []*http.Cookie) []*http.Cookie {
	var validCookies []*http.Cookie
	for _, cookie := range cookies {
		if !c.IsCookieExpired(cookie) {
			validCookies = append(validCookies, cookie)
		}
	}
	return validCookies
}