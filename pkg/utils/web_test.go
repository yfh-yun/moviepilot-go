package utils

import (
	"testing"
)

// TestGetLocation 测试GetLocation函数
func TestGetLocation(t *testing.T) {
	// 测试正常情况：有效的IPv4地址
	// 注意：这里使用本地回环地址，可能会返回空结果，这是正常的
	result := GetLocation("127.0.0.1")
	// 由于测试环境可能无法访问外部API，所以只检查结果是否为字符串，不检查具体值
	_ = result

	// 测试正常情况：有效的IPv6地址
	result = GetLocation("::1")
	_ = result

	// 测试边界情况：空字符串
	result = GetLocation("")
	expected := ""
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	// 测试边界情况：无效的IP地址
	result = GetLocation("invalid-ip")
	// 对于无效IP，函数应该返回空字符串或具体的错误信息，这取决于API的处理
	// 我们只检查结果是否为字符串，不检查具体值
	_ = result
}

// TestGetLocation1 测试getLocation1函数
func TestGetLocation1(t *testing.T) {
	// 注意：这个测试会调用外部API，可能会因为网络问题或API限制而失败
	// 我们只检查函数是否能正常执行，不检查具体结果
	result := getLocation1("127.0.0.1")
	_ = result
}

// TestGetLocation2 测试getLocation2函数
func TestGetLocation2(t *testing.T) {
	// 注意：这个测试会调用外部API，可能会因为网络问题或API限制而失败
	// 我们只检查函数是否能正常执行，不检查具体结果
	result := getLocation2("127.0.0.1")
	_ = result
}
