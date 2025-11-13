package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"moviepilot-go/internal/utils"
)

func main() {
	fmt.Println("=== 安全工具示例 ===")

	// 创建安全工具类实�?	securityUtils := utils.NewSecurityUtils()

	// 测试路径安全检�?	fmt.Println("\n--- 路径安全检�?---")
	testPathSafety(securityUtils)

	// 测试URL安全检�?	fmt.Println("\n--- URL安全检�?---")
	testURLSafety(securityUtils)

	// 测试URL路径清理
	fmt.Println("\n--- URL路径清理 ---")
	testSanitizeURLPath(securityUtils)
}

func testPathSafety(securityUtils *utils.SecurityUtils) {
	// 创建测试目录结构
	baseDir := "./test_base"
	testDir := filepath.Join(baseDir, "subdir")
	
	// 确保测试目录存在
	_ = os.MkdirAll(testDir, 0755)
	defer os.RemoveAll(baseDir)
	
	// 测试合法路径
	allowedSuffixes := []string{".txt", ".log"}
	
	// 合法路径 - 在基准目录内且后缀合法
	validPath := filepath.Join(baseDir, "test.txt")
	isSafe := securityUtils.IsSafePath(baseDir, validPath, allowedSuffixes)
	fmt.Printf("合法路径检�?(%s): %v\n", validPath, isSafe)
	
	// 合法路径 - 在子目录内且后缀合法
	validSubPath := filepath.Join(testDir, "sub_test.log")
	isSafe = securityUtils.IsSafePath(baseDir, validSubPath, allowedSuffixes)
	fmt.Printf("子目录合法路径检�?(%s): %v\n", validSubPath, isSafe)
	
	// 非法路径 - 目录遍历攻击
	invalidPath := filepath.Join(baseDir, "../outside.txt")
	isSafe = securityUtils.IsSafePath(baseDir, invalidPath, allowedSuffixes)
	fmt.Printf("非法路径检�?(%s): %v\n", invalidPath, isSafe)
	
	// 非法路径 - 后缀不合�?	invalidExtPath := filepath.Join(baseDir, "test.exe")
	isSafe = securityUtils.IsSafePath(baseDir, invalidExtPath, allowedSuffixes)
	fmt.Printf("非法后缀检�?(%s): %v\n", invalidExtPath, isSafe)
	
	// 不带后缀限制的检�?	isSafe = securityUtils.IsSafePath(baseDir, validPath, nil)
	fmt.Printf("无后缀限制检�?(%s): %v\n", validPath, isSafe)
}

func testURLSafety(securityUtils *utils.SecurityUtils) {
	// 允许的域名列�?	allowedDomains := []string{"example.com", "test.org:8080", "sub.domain.com"}
	
	// 测试合法URL
	validURL1 := "https://example.com/path"
	isSafe := securityUtils.IsSafeURL(validURL1, allowedDomains, false)
	fmt.Printf("合法URL检�?(%s): %v\n", validURL1, isSafe)
	
	// 测试子域名URL（非严格模式�?	validURL2 := "https://api.example.com/path"
	isSafe = securityUtils.IsSafeURL(validURL2, allowedDomains, false)
	fmt.Printf("子域名URL检�?(%s): %v\n", validURL2, isSafe)
	
	// 测试子域名URL（严格模式）
	isSafe = securityUtils.IsSafeURL(validURL2, allowedDomains, true)
	fmt.Printf("子域名URL严格模式检�?(%s): %v\n", validURL2, isSafe)
	
	// 测试带端口的合法URL
	validURL3 := "http://test.org:8080/api"
	isSafe = securityUtils.IsSafeURL(validURL3, allowedDomains, false)
	fmt.Printf("带端口URL检�?(%s): %v\n", validURL3, isSafe)
	
	// 测试非法URL - 不在允许列表�?	invalidURL1 := "https://malicious.com/path"
	isSafe = securityUtils.IsSafeURL(invalidURL1, allowedDomains, false)
	fmt.Printf("非法域名URL检�?(%s): %v\n", invalidURL1, isSafe)
	
	// 测试非法URL - 协议不合�?	invalidURL2 := "ftp://example.com/path"
	isSafe = securityUtils.IsSafeURL(invalidURL2, allowedDomains, false)
	fmt.Printf("非法协议URL检�?(%s): %v\n", invalidURL2, isSafe)
	
	// 测试无效URL格式
	invalidURL3 := "not_a_url"
	isSafe = securityUtils.IsSafeURL(invalidURL3, allowedDomains, false)
	fmt.Printf("无效URL格式检�?(%s): %v\n", invalidURL3, isSafe)
}

func testSanitizeURLPath(securityUtils *utils.SecurityUtils) {
	// 测试正常的URL路径
	url1 := "https://example.com/path/to/resource"
	safePath := securityUtils.SanitizeURLPath(url1, 120)
	fmt.Printf("正常路径清理 (%s): %s\n", url1, safePath)
	
	// 测试包含特殊字符的URL路径
	url2 := "https://example.com/path with spaces/文件�?txt"
	safePath = securityUtils.SanitizeURLPath(url2, 120)
	fmt.Printf("特殊字符路径清理 (%s): %s\n", url2, safePath)
	
	// 测试超长路径
	longPath := "https://example.com/" + strings.Repeat("very_long_path_segment/", 20) + "file.txt"
	safePath = securityUtils.SanitizeURLPath(longPath, 50)
	fmt.Printf("超长路径清理 (长度: %d): %s\n", len(longPath), safePath)
	
	// 测试边界情况 - 空URL
	emptyURL := ""
	safePath = securityUtils.SanitizeURLPath(emptyURL, 120)
	fmt.Printf("空URL路径清理: %s\n", safePath)
}
