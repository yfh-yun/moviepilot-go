package helper

import (
	"testing"
)

func TestOcrHelper(t *testing.T) {
	// 测试创建OcrHelper实例
	t.Run("创建OcrHelper实例", func(t *testing.T) {
		ocrHelper := NewOcrHelper()
		if ocrHelper == nil {
			t.Error("无法创建OcrHelper实例")
		}
	})

	// 测试GetCaptchaText方法
	t.Run("测试GetCaptchaText方法", func(t *testing.T) {
		ocrHelper := NewOcrHelper()
		
		// 测试imageB64为nil的情�?		result := ocrHelper.GetCaptchaText(nil, nil, nil, nil)
		if result != "" {
			t.Error("当imageB64为nil时，应该返回空字符串")
		}
		
		// 测试空的imageB64
		emptyImageB64 := ""
		result = ocrHelper.GetCaptchaText(nil, &emptyImageB64, nil, nil)
		if result != "" {
			t.Error("当imageB64为空时，应该返回空字符串")
		}
	})
}
