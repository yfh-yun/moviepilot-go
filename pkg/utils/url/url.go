package url

import (
	"fmt"
	"mime"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	security "moviepilot-go/pkg/utils/security"
)

// ReplaceProxyPlaceholders 替换代理占位符
func ReplaceProxyPlaceholders(input string) string {
	if input == "" {
		return input
	}

	// 替换Github代理占位符
	githubProxy := getGithubProxy()
	if githubProxy != "" {
		input = strings.ReplaceAll(input, "{GITHUB_PROXY}", githubProxy)
	}

	// 替换PIP代理占位符
	pipProxy := getPIPProxy()
	if pipProxy != "" {
		input = strings.ReplaceAll(input, "{PIP_PROXY}", pipProxy)
	}

	// 替换TMDB API Key占位符
	tmdbAPIKey := getTMDBAPIKey()
	if tmdbAPIKey != "" {
		input = strings.ReplaceAll(input, "{TMDBAPIKEY}", tmdbAPIKey)
	}

	return input
}

// StandardizeBaseURL 标准化主机地址，确保有协议前缀且以 / 结尾
// 行为参考 Python UrlUtils.standardize_base_url
func StandardizeBaseURL(host string) string {
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

// AdaptRequestURL 基于 host/endpoint 适配完整 URL
// 对齐 Python UrlUtils.adapt_request_url
func AdaptRequestURL(host, endpoint string) string {
	if host == "" && endpoint == "" {
		return ""
	}
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		return endpoint
	}
	host = StandardizeBaseURL(host)
	if host == "" {
		return endpoint
	}
	return joinURL(host, endpoint)
}

// CombineURL 使用 host/path/query 组合生成完整 URL
// 参考 Python UrlUtils.combine_url
func CombineURL(host string, p string, query map[string][]string) (string, error) {
	if p == "" {
		p = "/"
	}
	host = StandardizeBaseURL(host)

	base, err := url.Parse(host)
	if err != nil {
		return "", err
	}

	// 使用 ResolveReference 合并路径
	ref, err := url.Parse(p)
	if err != nil {
		return "", err
	}

	merged := base.ResolveReference(ref)

	// 合并查询参数
	existing := merged.Query()
	for k, vs := range query {
		// 覆盖同名参数，保持与 Python 行为接近
		existing[k] = vs
	}
	merged.RawQuery = existing.Encode()

	return merged.String(), nil
}

// ReplaceImageURLPlaceholders 替换图片URL中的占位符
func ReplaceImageURLPlaceholders(imageURL string) string {
	if imageURL == "" {
		return imageURL
	}

	// 替换通用代理占位符
	imageURL = ReplaceProxyPlaceholders(imageURL)

	// 图片域名特殊处理
	if strings.Contains(imageURL, "tmdb.org") || strings.Contains(imageURL, "themoviedb.org") {
		// 如果配置了TMDB镜像域名，则替换
		if tmdbMirror := getTMDBMirror(); tmdbMirror != "" {
			imageURL = strings.ReplaceAll(imageURL, "image.tmdb.org", tmdbMirror)
			imageURL = strings.ReplaceAll(imageURL, "themoviedb.org", tmdbMirror)
		}
	}

	return imageURL
}

// IsImageURL 检查是否为图片URL
func IsImageURL(urlStr string) bool {
	if urlStr == "" {
		return false
	}

	lowerURL := strings.ToLower(urlStr)
	imageExtensions := []string{
		".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".svg", ".ico",
	}

	for _, ext := range imageExtensions {
		if strings.HasSuffix(lowerURL, ext) {
			return true
		}
	}

	// 检查URL参数中是否有format参数
	if parsedURL, err := url.Parse(urlStr); err == nil {
		if format := parsedURL.Query().Get("format"); format != "" {
			lowerFormat := strings.ToLower(format)
			for _, ext := range imageExtensions {
				if strings.HasPrefix(lowerFormat, strings.TrimPrefix(ext, ".")) {
					return true
				}
			}
		}
	}

	return false
}

// GetImageMimeType 根据URL推断MIME类型
func GetImageMimeType(urlStr string) string {
	if urlStr == "" {
		return "application/octet-stream"
	}

	lowerURL := strings.ToLower(urlStr)

	mimeMap := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".bmp":  "image/bmp",
		".svg":  "image/svg+xml",
		".ico":  "image/x-icon",
	}

	for ext, mimeType := range mimeMap {
		if strings.HasSuffix(lowerURL, ext) {
			return mimeType
		}
	}

	// 从URL参数中获取格式
	if parsedURL, err := url.Parse(urlStr); err == nil {
		if format := parsedURL.Query().Get("format"); format != "" {
			lowerFormat := strings.ToLower(format)
			switch {
			case strings.HasPrefix(lowerFormat, "jpg") || strings.HasPrefix(lowerFormat, "jpeg"):
				return "image/jpeg"
			case strings.HasPrefix(lowerFormat, "png"):
				return "image/png"
			case strings.HasPrefix(lowerFormat, "gif"):
				return "image/gif"
			case strings.HasPrefix(lowerFormat, "webp"):
				return "image/webp"
			}
		}
	}

	return "application/octet-stream"
}

// GetMimeType 根据文件路径或 URL 获取 MIME 类型，对齐 Python UrlUtils.get_mime_type
// 若无法推断类型则返回 defaultType（默认 application/octet-stream）
func GetMimeType(pathOrURL string, defaultType string) string {
	if defaultType == "" {
		defaultType = "application/octet-stream"
	}
	if pathOrURL == "" {
		return defaultType
	}

	// 尝试从 URL 中提取路径部分
	if u, err := url.Parse(pathOrURL); err == nil && u.Path != "" {
		pathOrURL = u.Path
	}

	ext := strings.ToLower(filepath.Ext(pathOrURL))
	if ext == "" {
		return defaultType
	}

	if mimeType := mime.TypeByExtension(ext); mimeType != "" {
		return mimeType
	}

	return defaultType
}

// Quote 对字符串做 URL 安全编码，对齐 Python UrlUtils.quote
func Quote(s string) string {
	return url.QueryEscape(s)
}

// URLParams 解析后的 URL 关键参数
type URLParams struct {
	Scheme   string
	Hostname string
	Port     int
	Path     string
}

// ParseURLParams 解析 URL，提取协议/主机/端口/路径，对齐 Python UrlUtils.parse_url_params
func ParseURLParams(raw string) (*URLParams, error) {
	if raw == "" {
		return nil, fmt.Errorf("empty url")
	}

	std := StandardizeBaseURL(raw)
	parsed, err := url.Parse(std)
	if err != nil {
		return nil, err
	}
	if parsed.Hostname() == "" {
		return nil, fmt.Errorf("invalid hostname")
	}

	scheme := parsed.Scheme
	hostname := parsed.Hostname()
	port := parsed.Port()
	if port == "" {
		if scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return nil, err
	}

	pth := parsed.Path
	if pth == "" {
		pth = "/"
	}

	return &URLParams{
		Scheme:   scheme,
		Hostname: hostname,
		Port:     portNum,
		Path:     pth,
	}, nil
}

// getGithubProxy 获取Github代理URL
func getGithubProxy() string {
	if proxy := os.Getenv("GITHUB_PROXY"); proxy != "" {
		return strings.TrimSuffix(proxy, "/")
	}
	return ""
}

// getPIPProxy 获取PIP代理URL
func getPIPProxy() string {
	if proxy := os.Getenv("PIP_PROXY"); proxy != "" {
		return strings.TrimSuffix(proxy, "/")
	}
	return ""
}

// getTMDBAPIKey 获取TMDB API Key
func getTMDBAPIKey() string {
	return os.Getenv("TMDB_API_KEY")
}

// getTMDBMirror 获取TMDB镜像域名
func getTMDBMirror() string {
	if mirror := os.Getenv("TMDB_IMAGE_DOMAIN"); mirror != "" {
		return strings.TrimSuffix(mirror, "/")
	}
	return ""
}

// IsValidURL 检查URL格式是否有效
func IsValidURL(urlStr string) bool {
	if urlStr == "" {
		return false
	}

	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		return false
	}

	_, err := url.Parse(urlStr)
	return err == nil
}

// ExtractDomain 提取URL的域名
func ExtractDomain(urlStr string) string {
	if !IsValidURL(urlStr) {
		return ""
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}

	return parsedURL.Host
}

// IsSafeImageURL 判断图片 URL 是否安全，辅助 pkg/utils/security
// 使用 security.IsSafeImagePath 以及 URL 合法性检查
func IsSafeImageURL(urlStr string) bool {
	if !IsValidURL(urlStr) {
		return false
	}
	return security.IsSafeImagePath(urlStr)
}

// BuildCacheKey 构建缓存键
func BuildCacheKey(urlStr, prefix string) string {
	if urlStr == "" {
		return ""
	}

	// 简单的哈希算法，实际可以使用更复杂的哈希
	if prefix == "" {
		prefix = "img_cache"
	}

	return fmt.Sprintf("%s_%x", prefix, urlStr)
}

func joinURL(base, rel string) string {
	if base == "" {
		return rel
	}
	ub, err := url.Parse(base)
	if err != nil {
		return base
	}
	ur, err := url.Parse(rel)
	if err != nil {
		return base
	}
	return ub.ResolveReference(ur).String()
}
