package utils

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OCRHelper OCR识别助手
type OCRHelper struct {
	ocrBase64URL string
	httpClient   *http.Client
}

// OCRResult OCR识别结果
type OCRResult struct {
	Text     string  `json:"text"`
	Confidence float64 `json:"confidence,omitempty"`
	Boxes    []Box   `json:"boxes,omitempty"`
}

// Box 文本框信息
type Box struct {
	X      int     `json:"x"`
	Y      int     `json:"y"`
	Width  int     `json:"width"`
	Height int     `json:"height"`
	Text   string  `json:"text"`
	Score  float64 `json:"score"`
}

// NewOCRHelper 创建OCR助手实例
func NewOCRHelper(ocrHost string) *OCRHelper {
	if ocrHost == "" {
		ocrHost = "http://localhost:8080"
	}

	return &OCRHelper{
		ocrBase64URL: fmt.Sprintf("%s/captcha/base64", ocrHost),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetCaptchaText 根据图片地址获取验证码文本
func (ocr *OCRHelper) GetCaptchaText(imageURL, imageB64, cookie, userAgent string) (string, error) {
	var base64Image string
	var err error

	if imageURL != "" {
		// 从URL下载图片
		base64Image, err = ocr.downloadImageToBase64(imageURL, cookie, userAgent)
		if err != nil {
			return "", fmt.Errorf("failed to download image: %v", err)
		}
	} else if imageB64 != "" {
		// 使用提供的base64图片
		base64Image = imageB64
	} else {
		return "", fmt.Errorf("either image URL or base64 image must be provided")
	}

	if base64Image == "" {
		return "", fmt.Errorf("image data is empty")
	}

	// 调用OCR服务识别文本
	return ocr.recognizeText(base64Image)
}

// downloadImageToBase64 下载图片并转换为base64
func (ocr *OCRHelper) downloadImageToBase64(imageURL, cookie, userAgent string) (string, error) {
	req, err := http.NewRequest("GET", imageURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	// 设置请求头
	if userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}
	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}

	resp, err := ocr.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download image: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// 读取图片数据
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read image data: %v", err)
	}

	if len(imageData) == 0 {
		return "", fmt.Errorf("downloaded image is empty")
	}

	// 转换为base64
	mimeType := ocr.detectMimeType(imageData)
	if mimeType == "" {
		mimeType = "image/jpeg" // 默认
	}

	base64Str := base64.StdEncoding.EncodeToString(imageData)
	return fmt.Sprintf("data:%s;base64,%s", mimeType, base64Str), nil
}

// recognizeText 识别文本
func (ocr *OCRHelper) recognizeText(base64Image string) (string, error) {
	// 构建请求体
	requestBody := map[string]string{
		"base64_img": base64Image,
	}

	// 发送POST请求到OCR服务
	result, err := ocr.sendOCRRequest(requestBody)
	if err != nil {
		return "", err
	}

	return result.Text, nil
}

// sendOCRRequest 发送OCR请求
func (ocr *OCRHelper) sendOCRRequest(requestBody map[string]string) (*OCRResult, error) {
	// 这里应该实现实际的HTTP请求
	// 由于是示例，我们返回一个模拟的结果
	return &OCRResult{
		Text: "sample_ocr_result",
	}, nil
}

// detectMimeType 检测图片MIME类型
func (ocr *OCRHelper) detectMimeType(imageData []byte) string {
	if len(imageData) < 12 {
		return ""
	}

	// 检测常见图片格式的文件头
	if bytes.HasPrefix(imageData, []byte{0xFF, 0xD8, 0xFF}) {
		return "image/jpeg"
	}
	if bytes.HasPrefix(imageData, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}) {
		return "image/png"
	}
	if bytes.HasPrefix(imageData, []byte{0x47, 0x49, 0x46, 0x38}) {
		return "image/gif"
	}
	if bytes.HasPrefix(imageData, []byte{0x42, 0x4D}) {
		return "image/bmp"
	}
	if bytes.HasPrefix(imageData, []byte{0x52, 0x49, 0x46, 0x46}) && len(imageData) > 8 && 
	   bytes.HasPrefix(imageData[8:12], []byte{0x57, 0x45, 0x42, 0x50}) {
		return "image/webp"
	}

	return ""
}

// RecognizeFromFile 从文件识别文本
func (ocr *OCRHelper) RecognizeFromFile(filePath string) (string, error) {
	// 这里应该读取文件并转换为base64
	// 简化实现
	return "", fmt.Errorf("file recognition not implemented")
}

// RecognizeFromBytes 从字节数据识别文本
func (ocr *OCRHelper) RecognizeFromBytes(imageData []byte) (string, error) {
	if len(imageData) == 0 {
		return "", fmt.Errorf("image data is empty")
	}

	mimeType := ocr.detectMimeType(imageData)
	if mimeType == "" {
		mimeType = "image/jpeg"
	}

	base64Str := base64.StdEncoding.EncodeToString(imageData)
	base64Image := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Str)

	return ocr.recognizeText(base64Image)
}

// ValidateImage 验证图片格式
func (ocr *OCRHelper) ValidateImage(imageData []byte) error {
	if len(imageData) == 0 {
		return fmt.Errorf("image data is empty")
	}

	if len(imageData) > 10*1024*1024 { // 10MB限制
		return fmt.Errorf("image size too large (max 10MB)")
	}

	mimeType := ocr.detectMimeType(imageData)
	if mimeType == "" {
		return fmt.Errorf("unsupported image format")
	}

	supportedTypes := []string{
		"image/jpeg",
		"image/png",
		"image/gif",
		"image/bmp",
		"image/webp",
	}

	for _, supportedType := range supportedTypes {
		if mimeType == supportedType {
			return nil
		}
	}

	return fmt.Errorf("unsupported image type: %s", mimeType)
}

// CleanText 清理识别的文本
func (ocr *OCRHelper) CleanText(text string) string {
	// 移除多余的空白字符
	text = strings.TrimSpace(text)
	
	// 替换常见的OCR错误
	replacements := map[string]string{
		"0": "O", // 数字0替换为字母O（在某些验证码中常见）
		"I": "1", // 字母I替换为数字1
		"l": "1", // 小写l替换为数字1
		"O": "0", // 字母O替换为数字0
	}

	for old, new := range replacements {
		text = strings.ReplaceAll(text, old, new)
	}

	// 移除非字母数字字符（保留常见的验证码字符）
	var result strings.Builder
	for _, char := range text {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || 
		   (char >= '0' && char <= '9') {
			result.WriteRune(char)
		}
	}

	return result.String()
}

// SetTimeout 设置HTTP客户端超时
func (ocr *OCRHelper) SetTimeout(timeout time.Duration) {
	ocr.httpClient.Timeout = timeout
}

// GetTimeout 获取HTTP客户端超时
func (ocr *OCRHelper) GetTimeout() time.Duration {
	return ocr.httpClient.Timeout
}

// SetOCRURL 设置OCR服务URL
func (ocr *OCRHelper) SetOCRURL(ocrHost string) {
	if ocrHost == "" {
		ocrHost = "http://localhost:8080"
	}
	ocr.ocrBase64URL = fmt.Sprintf("%s/captcha/base64", ocrHost)
}

// GetOCRURL 获取OCR服务URL
func (ocr *OCRHelper) GetOCRURL() string {
	return ocr.ocrBase64URL
}

// IsHealthy 检查OCR服务健康状态
func (ocr *OCRHelper) IsHealthy() bool {
	// 发送健康检查请求
	resp, err := ocr.httpClient.Get(ocr.ocrBase64URL + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// BatchRecognize 批量识别
func (ocr *OCRHelper) BatchRecognize(base64Images []string) ([]*OCRResult, error) {
	if len(base64Images) == 0 {
		return nil, fmt.Errorf("no images provided")
	}

	results := make([]*OCRResult, 0, len(base64Images))
	for _, base64Image := range base64Images {
		text, err := ocr.recognizeText(base64Image)
		if err != nil {
			// 对于批量处理，记录错误但继续处理其他图片
			results = append(results, &OCRResult{
				Text: "",
				Confidence: 0,
			})
			continue
		}

		results = append(results, &OCRResult{
			Text: text,
			Confidence: 1.0, // 默认置信度
		})
	}

	return results, nil
}

// ExtractCaptchaCode 提取验证码代码
func (ocr *OCRHelper) ExtractCaptchaCode(text string) string {
	// 清理文本
	cleaned := ocr.CleanText(text)
	
	// 验证码通常是4-6位字符
	if len(cleaned) < 4 || len(cleaned) > 6 {
		// 尝试提取数字和字母的组合
		var result strings.Builder
		count := 0
		for _, char := range cleaned {
			if ((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || 
			    (char >= '0' && char <= '9')) && count < 6 {
				result.WriteRune(char)
				count++
			}
		}
		cleaned = result.String()
	}

	return cleaned
}