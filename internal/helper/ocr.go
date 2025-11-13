package helper

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"moviepilot-go/internal/config"
	"moviepilot-go/internal/types"
	"moviepilot-go/internal/utils"
)

// OcrHelper OCR帮助�?type OcrHelper struct {
	ocrB64Url string
}

// NewOcrHelper 创建OcrHelper实例
func NewOcrHelper() *OcrHelper {
	// 注意：在Go版本中，我们使用默认的OCR_HOST，因为配置系统可能不�?	ocrHost := "https://movie-pilot.org"
	return &OcrHelper{
		ocrB64Url: fmt.Sprintf("%s/captcha/base64", ocrHost),
	}
}

// GetCaptchaText 根据图片地址，获取验证码图片，并识别内容
// :param imageUrl: 图片地址
// :param imageB64: 图片base64，跳过图片地址下载
// :param cookie: 下载图片使用的cookie
// :param ua: 下载图片使用的ua
func (o *OcrHelper) GetCaptchaText(imageUrl *string, imageB64 *string, cookie *string, ua *string) string {
	/*
	 * 根据图片地址，获取验证码图片，并识别内容
	 * :param imageUrl: 图片地址
	 * :param imageB64: 图片base64，跳过图片地址下载
	 * :param cookie: 下载图片使用的cookie
	 * :param ua: 下载图片使用的ua
	 */
	
	// 如果提供了图片URL，则下载图片并转换为base64
	if imageUrl != nil && *imageUrl != "" {
		// 设置请求�?		headers := make(map[string]string)
		if ua != nil && *ua != "" {
			headers["User-Agent"] = *ua
		}
		
		// 设置Cookie
		var cookies []*types.Cookie
		if cookie != nil && *cookie != "" {
			// 简化处理Cookie，实际应用中可能需要解析Cookie字符�?			cookies = []*types.Cookie{
				{Name: "cookie", Value: *cookie},
			}
		}
		
		// 发送GET请求下载图片
		resp, err := utils.RequestUtils.GetRes(*imageUrl, headers, nil, 0)
		if err == nil && resp != nil && resp.StatusCode == 200 && len(resp.Body) > 0 {
			// 将图片内容转换为base64
			encoded := base64.StdEncoding.EncodeToString(resp.Body)
			imageB64 = &encoded
		}
	}
	
	// 如果没有图片base64数据，则返回空字符串
	if imageB64 == nil || *imageB64 == "" {
		return ""
	}
	
	// 发送POST请求到OCR服务进行识别
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	
	// 构造请求数�?	requestData := map[string]string{
		"base64_img": *imageB64,
	}
	
	// 将数据转换为JSON
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return ""
	}
	
	// 发送POST请求
	resp, err := utils.RequestUtils.PostRes(o.ocrB64Url, jsonData, headers, nil, 0)
	if err != nil || resp == nil || resp.StatusCode != 200 {
		return ""
	}
	
	// 解析响应JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(resp.Content), &result); err != nil {
		return ""
	}
	
	// 返回识别结果
	if resultText, ok := result["result"].(string); ok {
		return resultText
	}
	
	return ""
}
