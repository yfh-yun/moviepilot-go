package utils

import (
	"fmt"
	"mime"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
)

// UrlUtils 提供URL处理相关的工具函�?type UrlUtils struct{}

// StandardizeBaseURL 标准化提供的主机地址，确保它以http://或https://开头，并且以斜�?/)结尾
// 参数: host - 提供的主机地址字符�?// 返回: 标准化后的主机地址字符�?func (u *UrlUtils) StandardizeBaseURL(host string) string {
	if host == "" {
		return host
	}
	
	if !strings.HasSuffix(host, "/") {
		host += "/"
	}
	
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = "http://" + host
	}
	
	return host
}

// AdaptRequestUrl 基于传入的host，适配请求的URL，确保每个请求的URL是完整的
// 参数: 
//   host - 主机�?//   endpoint - 端点
// 返回: 完整的请求URL字符串，如果无法组合则返回空字符�?func (u *UrlUtils) AdaptRequestUrl(host string, endpoint string) string {
	if host == "" && endpoint == "" {
		return ""
	}
	
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		return endpoint
	}
	
	host = u.StandardizeBaseURL(host)
	return u.joinUrl(host, endpoint)
}

// CombineUrl 使用给定的主机头、路径和查询参数组合生成完整的URL
// 参数:
//   host - 主机头，例如 https://example.com
//   path - 包含路径和可能已经包含的查询参数的端点，例如 /path/to/resource?current=1
//   query - 可选，额外的查询参数，例如 map[string]interface{}{"key": "value"}
// 返回: 完整的请求URL字符�?func (u *UrlUtils) CombineUrl(host string, path string, query map[string]interface{}) string {
	// 如果路径为空，则默认�?'/'
	if path == "" {
		path = "/"
	}
	
	host = u.StandardizeBaseURL(host)
	
	// 使用 urljoin 合并 host �?path
	urlStr := u.joinUrl(host, path)
	
	// 解析当前 URL
	parsedUrl, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}
	
	// 解析已存在的查询参数
	queryParams := parsedUrl.Query()
	
	// 与额外的查询参数合并
	if query != nil {
		for key, value := range query {
			// 将值转换为字符�?			var strValue string
			switch v := value.(type) {
			case string:
				strValue = v
			case int:
				strValue = strconv.Itoa(v)
			case bool:
				strValue = strconv.FormatBool(v)
			default:
				strValue = fmt.Sprintf("%v", v)
			}
			queryParams.Set(key, strValue)
		}
	}
	
	// 设置新的查询参数
	parsedUrl.RawQuery = queryParams.Encode()
	
	return parsedUrl.String()
}

// GetMimeType 根据文件路径�?URL 获取 MIME 类型，如果无法获取则返回默认类型
// 参数:
//   pathOrUrl - 文件路径�?URL
//   defaultType - 无法获取类型时返回的默认 MIME 类型
// 返回: 获取到的 MIME 类型或默认类�?func (u *UrlUtils) GetMimeType(pathOrUrl string, defaultType string) string {
	if defaultType == "" {
		defaultType = "application/octet-stream"
	}
	
	// 尝试根据路径�?URL 获取 MIME 类型
	mimeType := mime.TypeByExtension(filepath.Ext(pathOrUrl))
	
	// 如果无法推测到类型，返回默认类型
	if mimeType == "" {
		return defaultType
	}
	
	return mimeType
}

// Quote 将字符串编码�?URL 安全的格�?// 参数: s - 要编码的字符�?// 返回: 编码后的字符�?func (u *UrlUtils) Quote(s string) string {
	return url.QueryEscape(s)
}

// ParseUrlParams 解析给定�?URL，并提取协议、主机名、端口和路径信息
// 参数: urlStr - 需要解析的 URL 字符�?// 返回:
//   protocol - 协议（例如："http", "https"�?//   hostname - 主机名或 IP 地址
//   port - 端口�?//   path - URL 的路径部�?//   ok - 是否解析成功
func (u *UrlUtils) ParseUrlParams(urlStr string) (protocol, hostname string, port int64, path string, ok bool) {
	if urlStr == "" {
		return "", "", 0, "", false
	}
	
	urlStr = u.StandardizeBaseURL(urlStr)
	
	parsed, err := url.Parse(urlStr)
	if err != nil || parsed.Hostname() == "" {
		return "", "", 0, "", false
	}
	
	protocol = parsed.Scheme
	hostname = parsed.Hostname()
	
	// 获取端口
	portStr := parsed.Port()
	if portStr != "" {
		var portErr error
		port, portErr = strconv.ParseInt(portStr, 10, 64)
		if portErr != nil {
			port = 0
		}
	} else {
		// 默认端口
		if protocol == "https" {
			port = 443
		} else {
			port = 80
		}
	}
	
	path = parsed.Path
	if path == "" {
		path = "/"
	}
	
	return protocol, hostname, port, path, true
}

// joinUrl 模拟Python的urljoin功能
func (u *UrlUtils) joinUrl(base, ref string) string {
	if ref == "" {
		return base
	}
	
	// 如果ref是绝对路径，则直接返回ref
	if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
		return ref
	}
	
	// 确保base�?结尾
	base = strings.TrimSuffix(base, "/") + "/"
	
	// 处理ref开头的/
	ref = strings.TrimPrefix(ref, "/")
	
	return base + ref
}

// NewUrlUtils 创建一个新的UrlUtils实例
func NewUrlUtils() *UrlUtils {
	return &UrlUtils{}
}
