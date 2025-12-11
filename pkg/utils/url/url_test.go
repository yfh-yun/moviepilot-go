package url

import (
	"strings"
	"testing"
)

// TestStandardizeBaseURL 测试StandardizeBaseURL函数
func TestStandardizeBaseURL(t *testing.T) {
	// 测试正常情况：HTTP
	result := StandardizeBaseURL("example.com")
	expected := "http://example.com/"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	// 测试正常情况：HTTPS
	result = StandardizeBaseURL("https://example.com")
	expected = "https://example.com/"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	// 测试正常情况：已包含尾部斜杠
	result = StandardizeBaseURL("example.com/")
	expected = "http://example.com/"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	// 测试边界情况：空字符串
	result = StandardizeBaseURL("")
	expected = ""
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

// TestAdaptRequestURL 测试AdaptRequestURL函数
func TestAdaptRequestURL(t *testing.T) {
	// 测试正常情况：完整URL端点
	result := AdaptRequestURL("example.com", "https://full.url/path")
	expected := "https://full.url/path"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	// 测试正常情况：相对路径
	result = AdaptRequestURL("example.com", "api/v1/resource")
	expected = "http://example.com/api/v1/resource"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	// 测试边界情况：空主机
	result = AdaptRequestURL("", "api/v1/resource")
	expected = "api/v1/resource"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	// 测试边界情况：空端点
	result = AdaptRequestURL("example.com", "")
	expected = "http://example.com/"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	// 测试边界情况：都为空
	result = AdaptRequestURL("", "")
	expected = ""
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

// TestCombineURL 测试CombineURL函数
func TestCombineURL(t *testing.T) {
	// 测试正常情况：基本组合
	query := map[string][]string{
		"param1": {"value1"},
		"param2": {"value2"},
	}
	result, err := CombineURL("example.com", "api/v1/resource", query)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// 检查结果是否包含预期的部分
	if !contains(result, "http://example.com/api/v1/resource") {
		t.Errorf("Expected result to contain %s, got %s", "http://example.com/api/v1/resource", result)
	}
	if !contains(result, "param1=value1") {
		t.Errorf("Expected result to contain %s, got %s", "param1=value1", result)
	}
	if !contains(result, "param2=value2") {
		t.Errorf("Expected result to contain %s, got %s", "param2=value2", result)
	}

	// 测试边界情况：空路径
	result, err = CombineURL("example.com", "", query)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !contains(result, "http://example.com/?") {
		t.Errorf("Expected result to contain %s, got %s", "http://example.com/?", result)
	}

	// 测试边界情况：空查询
	result, err = CombineURL("example.com", "api/v1/resource", nil)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	expected := "http://example.com/api/v1/resource"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

// TestGetMimeType 测试GetMimeType函数
func TestGetMimeType(t *testing.T) {
	// 测试正常情况：常见文件类型
	result := GetMimeType("example.jpg", "")
	expected := "image/jpeg"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	// 测试正常情况：URL路径
	result = GetMimeType("https://example.com/file.png", "")
	expected = "image/png"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	// 测试正常情况：自定义默认类型
	result = GetMimeType("unknown.unknown", "application/custom")
	expected = "application/custom"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	// 测试边界情况：空路径
	result = GetMimeType("", "")
	expected = "application/octet-stream"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

// TestQuote 测试Quote函数
func TestQuote(t *testing.T) {
	// 测试正常情况：包含特殊字符
	result := Quote("hello world!")
	expected := "hello+world%21"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	// 测试正常情况：包含中文字符
	result = Quote("你好世界")
	// 中文应被正确编码
	if result == "你好世界" {
		t.Errorf("Expected Chinese characters to be encoded, got %s", result)
	}

	// 测试边界情况：空字符串
	result = Quote("")
	expected = ""
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

// TestParseURLParams 测试ParseURLParams函数
func TestParseURLParams(t *testing.T) {
	// 测试正常情况：HTTP URL
	result, err := ParseURLParams("http://example.com:8080/path")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.Scheme != "http" {
		t.Errorf("Expected scheme http, got %s", result.Scheme)
	}
	if result.Hostname != "example.com" {
		t.Errorf("Expected hostname example.com, got %s", result.Hostname)
	}
	if result.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", result.Port)
	}
	// 注意：由于StandardizeBaseURL会添加尾部斜杠，所以路径会是/path/
	if result.Path != "/path/" {
		t.Errorf("Expected path /path/, got %s", result.Path)
	}

	// 测试正常情况：HTTPS URL（默认端口）
	result, err = ParseURLParams("https://example.com/path")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.Scheme != "https" {
		t.Errorf("Expected scheme https, got %s", result.Scheme)
	}
	if result.Port != 443 {
		t.Errorf("Expected port 443, got %d", result.Port)
	}

	// 测试边界情况：空URL
	result, err = ParseURLParams("")
	if err == nil {
		t.Errorf("Expected error for empty URL, got nil")
	}
}

// TestIsValidURL 测试IsValidURL函数
func TestIsValidURL(t *testing.T) {
	// 测试正常情况：有效URL
	result := IsValidURL("https://example.com")
	if !result {
		t.Errorf("Expected true for valid URL, got false")
	}

	// 测试正常情况：无效URL
	result = IsValidURL("invalid-url")
	if result {
		t.Errorf("Expected false for invalid URL, got true")
	}

	// 测试边界情况：空字符串
	result = IsValidURL("")
	if result {
		t.Errorf("Expected false for empty string, got true")
	}
}

// 辅助函数：检查字符串是否包含指定子串
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
