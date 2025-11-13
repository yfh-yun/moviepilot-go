package main

import (
	"fmt"
	"moviepilot-go/internal/helper"
)

func main() {
	// 创建OCR帮助类实�?	ocrHelper := helper.NewOcrHelper()
	
	// 示例：使用base64图片数据进行OCR识别
	imageB64 := "示例base64图片数据"
	
	// 调用OCR识别方法
	result := ocrHelper.GetCaptchaText(nil, &imageB64, nil, nil)
	
	fmt.Printf("OCR识别结果: %s\n", result)
	
	// 示例：使用图片URL进行OCR识别
	imageUrl := "https://example.com/captcha.jpg"
	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
	cookie := "session_id=abc123"
	
	result = ocrHelper.GetCaptchaText(&imageUrl, nil, &cookie, &userAgent)
	
	fmt.Printf("通过URL识别的OCR结果: %s\n", result)
}
