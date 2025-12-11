package security

import (
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

// IsSafeURL 检查URL是否在允许的域名列表中
// 对应 Python SecurityUtils.is_safe_url：
// - 仅允许 http/https 协议
// - 支持带端口的域名
// - 支持严格模式和非严格模式
func IsSafeURL(rawURL string, allowedDomains []string, strict bool) bool {
	if rawURL == "" {
		return false
	}

	// 优先使用 net/url 严格解析
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		// 回退到简单域名提取
		domain := extractDomain(rawURL)
		if domain == "" {
			return false
		}
		for _, allowed := range allowedDomains {
			allowedLower := strings.ToLower(strings.TrimSpace(allowed))
			if allowedLower == "" {
				continue
			}
			domainLower := strings.ToLower(domain)
			if strict {
				if domainLower == allowedLower {
					return true
				}
			} else {
				if domainLower == allowedLower || strings.HasSuffix(domainLower, "."+allowedLower) {
					return true
				}
			}
		}
		return false
	}

	// 仅允许 http/https
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}

	netloc := strings.ToLower(parsed.Host)
	if netloc == "" {
		return false
	}

	// 允许 allowedDomains 中既可以是裸域名，也可以是完整URL
	for _, allowed := range allowedDomains {
		allowed = strings.TrimSpace(allowed)
		if allowed == "" {
			continue
		}

		ap, err := url.Parse(allowed)
		if err != nil {
			continue
		}
		allowedHost := strings.ToLower(ap.Host)
		if allowedHost == "" {
			// 传入的可能是裸域名
			allowedHost = strings.ToLower(ap.Path)
		}
		if allowedHost == "" {
			continue
		}

		if strict {
			// 严格模式：完全匹配域名和端口
			if netloc == allowedHost {
				return true
			}
		} else {
			// 非严格模式：支持子域名
			if netloc == allowedHost || strings.HasSuffix(netloc, "."+allowedHost) {
				return true
			}
		}
	}

	return false
}

// IsSafeURLWithDefaults 使用默认参数调用 IsSafeURL（非严格模式）
// 兼容原有函数签名
func IsSafeURLWithDefaults(rawURL string, allowedDomains []string) bool {
	return IsSafeURL(rawURL, allowedDomains, false)
}

// IsSafePath 检查路径是否安全，防止目录遍历攻击
// 对应 Python SecurityUtils.is_safe_path
func IsSafePath(basePath, targetPath string, allowedSuffixes []string) bool {
	if basePath == "" || targetPath == "" {
		return false
	}

	// 规范化路径
	absBase, err := filepath.Abs(basePath)
	if err != nil {
		return false
	}

	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return false
	}

	// 检查目标路径是否在基础路径内
	if !strings.HasPrefix(absTarget, absBase) {
		return false
	}

	// 检查文件后缀
	if len(allowedSuffixes) > 0 {
		// 获取文件后缀
		suffix := strings.ToLower(filepath.Ext(targetPath))
		if suffix == "" {
			return false
		}

		// 检查后缀是否在允许列表中
		allowedSet := make(map[string]bool)
		for _, allowed := range allowedSuffixes {
			allowedSet[strings.ToLower(allowed)] = true
		}

		if !allowedSet[suffix] {
			return false
		}
	}

	return true
}

// IsSafePathNoSuffix 检查路径是否安全，不检查文件后缀
// 兼容原有函数签名
func IsSafePathNoSuffix(basePath, targetPath string) bool {
	return IsSafePath(basePath, targetPath, nil)
}

// IsSafeLogPath 检查日志文件路径是否安全
func IsSafeLogPath(logPath, baseLogPath string) bool {
	if logPath == "" {
		return false
	}

	// 设置默认基础日志路径
	if baseLogPath == "" {
		baseLogPath = "/app/logs"
	}

	// 检查是否在日志目录内
	if !strings.HasPrefix(logPath, baseLogPath) {
		return false
	}

	// 检查是否为 .log 后缀
	if len(logPath) < 4 || logPath[len(logPath)-4:] != ".log" {
		return false
	}

	// 检查是否包含路径遍历字符
	if strings.Contains(logPath, "..") {
		return false
	}

	return true
}

// SanitizePath 清理路径，移除危险字符
func SanitizePath(path string) string {
	if path == "" {
		return ""
	}

	// 移除路径遍历字符
	path = strings.ReplaceAll(path, "..", "")
	path = strings.ReplaceAll(path, "~", "")

	// 规范化路径分隔符
	path = filepath.ToSlash(path)

	// 移除开头的斜杠
	path = strings.TrimLeft(path, "/")

	return path
}

// SanitizeURLPath 将 URL 的路径部分进行清理并在必要时压缩，适合用作缓存路径
// 对齐 Python SecurityUtils.sanitize_url_path：
// - 仅使用 URL 的 path 部分
// - 对特殊字符进行转义
// - 当路径过长时，使用 sha256 哈希压缩并保留扩展名
func SanitizeURLPath(rawURL string, maxLength int) string {
	if rawURL == "" {
		return ""
	}
	if maxLength <= 0 {
		maxLength = 120
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		// 解析失败时退化为普通路径清理
		return SanitizePath(rawURL)
	}

	// 只取路径部分并移除开头的斜杠
	pathPart := strings.TrimLeft(parsed.Path, "/")
	if pathPart == "" {
		pathPart = "index"
	}

	// 对路径进行 URL 安全转义
	safePath := url.PathEscape(pathPart)

	// 如果路径过长则进行压缩
	if len(safePath) > maxLength {
		// 计算 sha256 哈希并取前16位
		h := sha256.Sum256([]byte(safePath))
		hashStr := hex.EncodeToString(h[:])[:16]

		// 尝试保留扩展名
		ext := filepath.Ext(safePath)
		if len(ext) > 16 {
			ext = ""
		}

		safePath = "compressed_" + hashStr + ext
	}

	return safePath
}

// IsValidTokenFormat 检查token格式是否有效
func IsValidTokenFormat(token string) bool {
	if token == "" {
		return false
	}

	// 基本长度检查
	if len(token) < 8 || len(token) > 256 {
		return false
	}

	// 检查是否只包含允许的字符
	validChars := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	return validChars.MatchString(token)
}

// extractDomain 从URL中提取域名
func extractDomain(url string) string {
	// 移除协议
	if strings.HasPrefix(url, "http://") {
		url = strings.TrimPrefix(url, "http://")
	} else if strings.HasPrefix(url, "https://") {
		url = strings.TrimPrefix(url, "https://")
	}

	// 移除路径
	parts := strings.Split(url, "/")
	if len(parts) == 0 {
		return ""
	}

	// 移除端口
	domain := strings.Split(parts[0], ":")[0]

	return domain
}

// GetAllowedImageDomains 获取允许的图片域名列表
func GetAllowedImageDomains() []string {
	return []string{
		"themoviedb.org",
		"tmdb.org",
		"image.tmdb.org",
		"doubanio.com",
		"img1.doubanio.com",
		"img2.doubanio.com",
		"img3.doubanio.com",
		"img4.doubanio.com",
		"img5.doubanio.com",
		"img9.doubanio.com",
		"douban.com",
		"fanart.tv",
		"webservice.fanart.tv",
	}
}

// IsSafeImagePath 检查图片URL是否安全
func IsSafeImagePath(url string) bool {
	allowedDomains := GetAllowedImageDomains()
	return IsSafeURL(url, allowedDomains, false)
}
