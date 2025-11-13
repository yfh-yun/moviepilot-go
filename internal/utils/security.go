package utils

import (
	"crypto/sha256"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"moviepilot-go/internal/logger"
)

// SecurityUtils 安全工具�?type SecurityUtils struct{}

// NewSecurityUtils 创建新的安全工具类实�?func NewSecurityUtils() *SecurityUtils {
	return &SecurityUtils{}
}

// IsSafePath 验证用户提供的路径是否在基准目录内，并检查文件类型是否合法，防止目录遍历攻击
// base_path: 基准目录，允许访问的根目�?// user_path: 用户提供的路径，需检查其是否位于基准目录�?// allowed_suffixes: 允许的文件后缀名集合，用于验证文件类型
// 返回�? 如果用户路径安全且位于基准目录内，且文件类型合法，返�?true；否则返�?false
func (s *SecurityUtils) IsSafePath(basePath string, userPath string, allowedSuffixes []string) bool {
	log := logger.GetLoggerManager()
	
	// 将相对路径转换为绝对路径
	basePathResolved, err := filepath.Abs(basePath)
	if err != nil {
		log.Debug(fmt.Sprintf("Error occurred while validating paths: %v", err))
		return false
	}
	
	userPathResolved, err := filepath.Abs(userPath)
	if err != nil {
		log.Debug(fmt.Sprintf("Error occurred while validating paths: %v", err))
		return false
	}
	
	// 检查用户路径是否在基准目录或基准目录的子目录内
	relPath, err := filepath.Rel(basePathResolved, userPathResolved)
	if err != nil {
		log.Debug(fmt.Sprintf("Error occurred while validating paths: %v", err))
		return false
	}
	
	// 检查路径是否在基准目录之外（通过检查相对路径是否以 .. 开头）
	if strings.HasPrefix(relPath, "..") {
		return false
	}
	
	// 检查文件后缀是否合法
	if allowedSuffixes != nil {
		// 将后缀列表转换为小写集�?		allowedSet := make(map[string]bool)
		for _, suffix := range allowedSuffixes {
			allowedSet[strings.ToLower(suffix)] = true
		}
		
		// 获取文件后缀并转换为小写
		ext := strings.ToLower(filepath.Ext(userPathResolved))
		if _, exists := allowedSet[ext]; !exists {
			return false
		}
	}
	
	return true
}

// IsSafeURL 验证URL是否在允许的域名列表中，包括带有端口的域�?// url: 需要验证的 URL
// allowedDomains: 允许的域名集合，域名可以包含端口
// strict: 是否严格匹配一级域名（默认�?false，允许多级域名）
// 返回�? 如果URL合法且在允许的域名列表中，返�?true；否则返�?false
func (s *SecurityUtils) IsSafeURL(rawURL string, allowedDomains []string, strict bool) bool {
	log := logger.GetLoggerManager()
	
	// 解析URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		log.Debug(fmt.Sprintf("Error occurred while validating URL: %v", err))
		return false
	}
	
	// 如果 URL 没有包含有效�?scheme，或者无法从中提取到有效�?host，则认为�?URL 是无效的
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return false
	}
	
	// 仅允�?http �?https 协议
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return false
	}
	
	// 获取完整�?host（包�?IP 和端口）并转换为小写
	host := strings.ToLower(parsedURL.Host)
	if host == "" {
		return false
	}
	
	// 检查每个允许的域名
	allowedSet := make(map[string]bool)
	for _, domain := range allowedDomains {
		allowedSet[strings.ToLower(domain)] = true
	}
	
	for domain := range allowedSet {
		// 解析允许的域�?		parsedAllowedURL, err := url.Parse(domain)
		if err != nil {
			continue
		}
		
		allowedHost := strings.ToLower(parsedAllowedURL.Host)
		if allowedHost == "" {
			allowedHost = strings.ToLower(parsedAllowedURL.Path)
		}
		
		if strict {
			// 严格模式下，要求完全匹配域名和端�?			if host == allowedHost {
				return true
			}
		} else {
			// 非严格模式下，允许子域名匹配
			if host == allowedHost || strings.HasSuffix(host, "."+allowedHost) {
				return true
			}
		}
	}
	
	return false
}

// SanitizeURLPath �?URL 的路径部分进行编码，确保合法字符，并对路径长度进行压缩处理（如果超出最大长度）
// url: 需要处理的 URL
// maxLength: 路径允许的最大长度，超出时进行压�?// 返回�? 处理后的路径字符�?func (s *SecurityUtils) SanitizeURLPath(rawURL string, maxLength int) string {
	// 解析 URL，获取路径部�?	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	
	path := strings.TrimPrefix(parsedURL.Path, "/")
	
	// 对路径中的特殊字符进行编�?	safePath := url.QueryEscape(path)
	
	// 如果路径过长，进行压缩处�?	if len(safePath) > maxLength {
		// 使用 SHA-256 对路径进行哈希，取前 16 位作为压缩后的路�?		hash := sha256.Sum256([]byte(safePath))
		hashValue := fmt.Sprintf("%x", hash)[:16]
		
		// 使用哈希值代替过长的路径，同时保留文件扩展名
		var fileExtension string
		if ext := filepath.Ext(safePath); ext != "" {
			fileExtension = strings.ToLower(ext)
		}
		
		safePath = fmt.Sprintf("compressed_%s%s", hashValue, fileExtension)
	}
	
	return safePath
}
