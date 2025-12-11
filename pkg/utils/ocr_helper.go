package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// OcrHelper OCR帮助类
type OcrHelper struct {
	logger *zap.Logger
	ocrURL string
	mutex  sync.RWMutex
}

// OCRResponse OCR响应
type OCRResponse struct {
	Result string `json:"result"`
}

// NewOcrHelper 创建OCR帮助类实例
func NewOcrHelper(ocrHost string) *OcrHelper {
	return &OcrHelper{
		logger: logger.GetLogger(),
		ocrURL: fmt.Sprintf("%s/captcha/base64", ocrHost),
	}
}

// GetCaptchaText 获取验证码文本
func (h *OcrHelper) GetCaptchaText(imageURL, imageB64, cookie, ua string) (string, error) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	// 如果没有base64图片，从URL下载
	if imageB64 == "" && imageURL != "" {
		b64, err := h.downloadImageToBase64(imageURL, cookie, ua)
		if err != nil {
			h.logger.Error("下载图片失败", zap.Error(err))
			return "", err
		}
		imageB64 = b64
	}

	// 如果仍然没有base64图片，返回空
	if imageB64 == "" {
		h.logger.Error("没有图片数据")
		return "", fmt.Errorf("没有图片数据")
	}

	// 调用OCR服务
	return h.callOCRService(imageB64)
}

// downloadImageToBase64 下载图片并转换为base64
func (h *OcrHelper) downloadImageToBase64(imageURL, cookie, ua string) (string, error) {
	// 创建HTTP客户端
	client := &http.Client{}

	// 创建请求
	req, err := http.NewRequest("GET", imageURL, nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	if ua != "" {
		req.Header.Set("User-Agent", ua)
	}

	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	// 读取图片数据
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取图片数据失败: %w", err)
	}

	// 转换为base64
	return base64.StdEncoding.EncodeToString(imageData), nil
}

// callOCRService 调用OCR服务
func (h *OcrHelper) callOCRService(imageB64 string) (string, error) {
	// 创建请求体
	reqBody := map[string]string{
		"base64_img": imageB64,
	}

	// 序列化请求体
	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求体失败: %w", err)
	}

	// 创建HTTP客户端
	client := &http.Client{}

	// 创建请求
	req, err := http.NewRequest("POST", h.ocrURL, strings.NewReader(string(reqJSON)))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OCR服务请求失败，状态码: %d", resp.StatusCode)
	}

	// 读取响应数据
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应数据失败: %w", err)
	}

	// 解析响应
	var ocrResp OCRResponse
	if err := json.Unmarshal(respData, &ocrResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	return ocrResp.Result, nil
}

// SetOCRHost 设置OCR服务地址
func (h *OcrHelper) SetOCRHost(ocrHost string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.ocrURL = fmt.Sprintf("%s/captcha/base64", ocrHost)
}

// GetOCRHost 获取OCR服务地址
func (h *OcrHelper) GetOCRHost() string {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return strings.Split(h.ocrURL, "/captcha/base64")[0]
}
