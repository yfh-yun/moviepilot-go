package helper

import (
	"testing"
)

func TestTwoFactorAuth_NewTwoFactorAuth(t *testing.T) {
	// 测试使用短字符串初始�?	tfa1 := NewTwoFactorAuth("123456")
	if tfa1.Code != "123456" {
		t.Errorf("Expected Code to be '123456', got '%s'", tfa1.Code)
	}
	if tfa1.Secret != "" {
		t.Errorf("Expected Secret to be empty, got '%s'", tfa1.Secret)
	}

	// 测试使用长字符串初始�?	tfa2 := NewTwoFactorAuth("1234567890123456")
	if tfa2.Code != "" {
		t.Errorf("Expected Code to be empty, got '%s'", tfa2.Code)
	}
	if tfa2.Secret != "1234567890123456" {
		t.Errorf("Expected Secret to be '1234567890123456', got '%s'", tfa2.Secret)
	}
}

func TestTwoFactorAuth_GetCode(t *testing.T) {
	// 测试预设验证�?	tfa1 := NewTwoFactorAuth("123456")
	code := tfa1.GetCode()
	if code != "123456" {
		t.Errorf("Expected code to be '123456', got '%s'", code)
	}

	// 测试动态计算验证码（使用空密钥�?	tfa2 := NewTwoFactorAuth("")
	code = tfa2.GetCode()
	if code != "" {
		t.Errorf("Expected code to be empty, got '%s'", code)
	}
}
