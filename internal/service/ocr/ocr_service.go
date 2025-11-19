package ocr

import (
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/yfh-yun/moviepilot-go/internal/model"
	"go.uber.org/zap"
)

// OCRService OCR识别服务
type OCRService struct {
	logger    *zap.Logger
	engines   map[string]OCREngine
	processor *ImageProcessor
	cache     *OCRResultCache
}

// OCREngine OCR引擎接口
type OCREngine interface {
	// Name 引擎名称
	Name() string
	
	// IsAvailable 检查引擎是否可用
	IsAvailable() bool
	
	// RecognizeText 识别文本
	RecognizeText(ctx context.Context, img image.Image, options *OCROptions) (*OCRResult, error)
	
	// RecognizeFromPath 从文件路径识别
	RecognizeFromPath(ctx context.Context, imagePath string, options *OCROptions) (*OCRResult, error)
	
	// SupportedFormats 支持的格式
	SupportedFormats() []string
	
	// Priority 引擎优先级
	Priority() int
}

// OCROptions OCR选项
type OCROptions struct {
	Language     string            `json:"language"`     // 语言代码，如 "zh-CN", "en"
	Scale        float64           `json:"scale"`        // 缩放比例
	Confidence   float64           `json:"confidence"`   // 置信度阈值
	Preprocess   *PreprocessOptions `json:"preprocess"`   // 预处理选项
	Engine       string            `json:"engine"`       // 指定引擎
	Timeout      time.Duration     `json:"timeout"`      // 超时时间
}

// PreprocessOptions 预处理选项
type PreprocessOptions struct {
	Grayscale     bool    `json:"grayscale"`      // 灰度化
	Binarize      bool    `json:"binarize"`       // 二值化
	Resize        float64 `json:"resize"`         // 调整大小
	Contrast      float64 `json:"contrast"`       // 对比度调整
	Brightness    float64 `json:"brightness"`     // 亮度调整
	NoiseReduction bool   `json:"noise_reduction"` // 降噪
}

// OCRResult OCR识别结果
type OCRResult struct {
	Text        string             `json:"text"`        // 识别文本
	Confidence  float64            `json:"confidence"`  // 平均置信度
	Boxes       []*TextBox        `json:"boxes"`       // 文本框信息
	Engine      string             `json:"engine"`      // 使用的引擎
	Duration    time.Duration      `json:"duration"`    // 识别耗时
	Metadata    map[string]interface{} `json:"metadata"`    // 元数据
	Success     bool               `json:"success"`     // 是否成功
	Error       string             `json:"error"`       // 错误信息
}

// TextBox 文本框
type TextBox struct {
	Text       string    `json:"text"`
	Box        image.Rectangle `json:"box"`
	Confidence float64   `json:"confidence"`
	Language   string    `json:"language"`
}

// ImageProcessor 图像处理器
type ImageProcessor struct {
	logger *zap.Logger
}

// OCRResultCache OCR结果缓存
type OCRResultCache struct {
	cache   map[string]*CacheEntry
	mutex   sync.RWMutex
	ttl     time.Duration
	maxSize int
}

// CacheEntry 缓存条目
type CacheEntry struct {
	Result    *OCRResult
	ExpiresAt time.Time
}

// NewOCRService 创建OCR服务
func NewOCRService(logger *zap.Logger) *OCRService {
	service := &OCRService{
		logger:    logger,
		engines:   make(map[string]OCREngine),
		processor: NewImageProcessor(logger),
		cache:     NewOCRResultCache(time.Hour*24, 1000),
	}
	
	return service
}

// RegisterEngine 注册OCR引擎
func (s *OCRService) RegisterEngine(engine OCREngine) {
	s.engines[engine.Name()] = engine
	s.logger.Info("注册OCR引擎", 
		zap.String("name", engine.Name()),
		zap.Int("priority", engine.Priority()))
}

// RecognizeText 识别图片文本
func (s *OCRService) RecognizeText(ctx context.Context, img image.Image, options *OCROptions) (*OCRResult, error) {
	s.logger.Info("开始OCR文本识别", 
		zap.String("language", options.Language),
		zap.String("engine", options.Engine))

	// 1. 生成缓存键
	cacheKey := s.generateCacheKey(img, options)
	
	// 2. 尝试从缓存获取
	if cached := s.cache.Get(cacheKey); cached != nil {
		s.logger.Debug("从缓存获取OCR结果", zap.String("cache_key", cacheKey))
		return cached, nil
	}

	// 3. 图像预处理
	processedImg, err := s.processor.PreprocessImage(img, options.Preprocess)
	if err != nil {
		return nil, fmt.Errorf("图像预处理失败: %w", err)
	}

	// 4. 选择OCR引擎并执行识别
	result, err := s.executeOCREngine(ctx, processedImg, options)
	if err != nil {
		return nil, fmt.Errorf("OCR识别失败: %w", err)
	}

	// 5. 缓存结果
	s.cache.Set(cacheKey, result)

	s.logger.Info("OCR文本识别完成",
		zap.String("engine", result.Engine),
		zap.Float64("confidence", result.Confidence),
		zap.Duration("duration", result.Duration),
		zap.Int("text_boxes", len(result.Boxes)))

	return result, nil
}

// RecognizeFromPath 从文件路径识别
func (s *OCRService) RecognizeFromPath(ctx context.Context, imagePath string, options *OCROptions) (*OCRResult, error) {
	s.logger.Info("从文件路径进行OCR识别", zap.String("path", imagePath))

	// 1. 验证文件
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("文件不存在: %s", imagePath)
	}

	// 2. 检查文件格式
	ext := strings.ToLower(filepath.Ext(imagePath))
	if !s.isSupportedFormat(ext) {
		return nil, fmt.Errorf("不支持的文件格式: %s", ext)
	}

	// 3. 加载图片
	img, err := s.loadImage(imagePath)
	if err != nil {
		return nil, fmt.Errorf("加载图片失败: %w", err)
	}

	// 4. 执行OCR识别
	return s.RecognizeText(ctx, img, options)
}

// RecognizeCaptcha 识别验证码
func (s *OCRService) RecognizeCaptcha(ctx context.Context, img image.Image) (*CaptchaResult, error) {
	s.logger.Info("开始验证码识别")

	options := &OCROptions{
		Language: "en", // 验证码通常是英文
		Scale:    2.0,  // 放大识别
		Confidence: 0.6,
		Preprocess: &PreprocessOptions{
			Grayscale:     true,
			Binarize:      true,
			Contrast:      1.5,
			NoiseReduction: true,
		},
		Timeout: time.Second * 10,
	}

	result, err := s.RecognizeText(ctx, img, options)
	if err != nil {
		return nil, fmt.Errorf("验证码识别失败: %w", err)
	}

	// 后处理验证码文本
	captchaText := s.postProcessCaptchaText(result.Text)

	captchaResult := &CaptchaResult{
		Text:       captchaText,
		Confidence: result.Confidence,
		Engine:     result.Engine,
		Success:    result.Success,
	}

	s.logger.Info("验证码识别完成",
		zap.String("text", captchaText),
		zap.Float64("confidence", result.Confidence))

	return captchaResult, nil
}

// RecognizeQrCode 识别二维码
func (s *OCRService) RecognizeQrCode(ctx context.Context, img image.Image) (*QRCodeResult, error) {
	s.logger.Info("开始二维码识别")

	// 这里应该集成专门的二维码识别库
	// 为了示例，返回一个基本的结果
	result := &QRCodeResult{
		Text:    "",
		Format:  "QR_CODE",
		Success: false,
		Error:   "二维码识别功能未实现",
	}

	return result, nil
}

// GetEngines 获取可用的OCR引擎
func (s *OCRService) GetEngines() []string {
	var engines []string
	for name, engine := range s.engines {
		if engine.IsAvailable() {
			engines = append(engines, name)
		}
	}
	return engines
}

// GetEngineInfo 获取引擎信息
func (s *OCRService) GetEngineInfo(name string) *EngineInfo {
	engine, exists := s.engines[name]
	if !exists {
		return nil
	}

	return &EngineInfo{
		Name:            engine.Name(),
		Available:       engine.IsAvailable(),
		Priority:        engine.Priority(),
		SupportedFormats: engine.SupportedFormats(),
	}
}

// executeOCREngine 执行OCR引擎
func (s *OCRService) executeOCREngine(ctx context.Context, img image.Image, options *OCROptions) (*OCRResult, error) {
	var selectedEngines []OCREngine

	// 选择引擎
	if options.Engine != "" {
		// 指定了特定引擎
		if engine, exists := s.engines[options.Engine]; exists && engine.IsAvailable() {
			selectedEngines = append(selectedEngines, engine)
		} else {
			return nil, fmt.Errorf("指定的OCR引擎不可用: %s", options.Engine)
		}
	} else {
		// 按优先级选择可用引擎
		for _, engine := range s.engines {
			if engine.IsAvailable() {
				selectedEngines = append(selectedEngines, engine)
			}
		}
	}

	if len(selectedEngines) == 0 {
		return nil, fmt.Errorf("没有可用的OCR引擎")
	}

	// 尝试每个引擎
	var lastErr error
	for _, engine := range selectedEngines {
		start := time.Now()
		result, err := engine.RecognizeText(ctx, img, options)
		duration := time.Since(start)

		if err != nil {
			s.logger.Warn("OCR引擎识别失败",
				zap.String("engine", engine.Name()),
				zap.Error(err))
			lastErr = err
			continue
		}

		// 记录引擎信息
		result.Engine = engine.Name()
		result.Duration = duration

		// 检查置信度
		if options.Confidence > 0 && result.Confidence < options.Confidence {
			s.logger.Warn("OCR结果置信度过低",
				zap.String("engine", engine.Name()),
				zap.Float64("confidence", result.Confidence),
				zap.Float64("threshold", options.Confidence))
			continue
		}

		s.logger.Info("OCR引擎识别成功",
			zap.String("engine", engine.Name()),
			zap.Float64("confidence", result.Confidence),
			zap.Duration("duration", duration))

		return result, nil
	}

	return nil, fmt.Errorf("所有OCR引擎都识别失败: %w", lastErr)
}

// generateCacheKey 生成缓存键
func (s *OCRService) generateCacheKey(img image.Image, options *OCROptions) string {
	// 简化的缓存键生成，实际应该基于图像内容哈希
	bounds := img.Bounds()
	key := fmt.Sprintf("%dx%d:%s:%s:%.2f", 
		bounds.Dx(), bounds.Dy(), 
		options.Language, options.Engine, options.Scale)
	return key
}

// loadImage 加载图片
func (s *OCRService) loadImage(imagePath string) (image.Image, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	return img, err
}

// isSupportedFormat 检查是否支持格式
func (s *OCRService) isSupportedFormat(ext string) bool {
	supportedFormats := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".bmp":  true,
		".tiff": true,
		".tif":  true,
	}
	return supportedFormats[ext]
}

// postProcessCaptchaText 后处理验证码文本
func (s *OCRService) postProcessCaptchaText(text string) string {
	// 移除空格和特殊字符
	text = strings.TrimSpace(text)
	
	// 只保留字母数字
	var result []rune
	for _, r := range text {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			result = append(result, r)
		}
	}
	
	return strings.ToLower(string(result))
}

// CaptchaResult 验证码识别结果
type CaptchaResult struct {
	Text       string        `json:"text"`
	Confidence float64       `json:"confidence"`
	Engine     string        `json:"engine"`
	Success    bool          `json:"success"`
	Error      string        `json:"error,omitempty"`
	Duration   time.Duration `json:"duration,omitempty"`
}

// QRCodeResult 二维码识别结果
type QRCodeResult struct {
	Text    string `json:"text"`
	Format  string `json:"format"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// EngineInfo 引擎信息
type EngineInfo struct {
	Name            string   `json:"name"`
	Available       bool     `json:"available"`
	Priority        int      `json:"priority"`
	SupportedFormats []string `json:"supported_formats"`
}

// NewImageProcessor 创建图像处理器
func NewImageProcessor(logger *zap.Logger) *ImageProcessor {
	return &ImageProcessor{
		logger: logger,
	}
}

// PreprocessImage 预处理图像
func (p *ImageProcessor) PreprocessImage(img image.Image, options *PreprocessOptions) (image.Image, error) {
	if options == nil {
		return img, nil
	}

	processed := img

	// 这里应该实现具体的图像预处理逻辑
	// 为了示例，返回原图
	p.logger.Debug("图像预处理", 
		zap.Bool("grayscale", options.Grayscale),
		zap.Bool("binarize", options.Binarize),
		zap.Float64("resize", options.Resize))

	return processed, nil
}

// NewOCRResultCache 创建OCR结果缓存
func NewOCRResultCache(ttl time.Duration, maxSize int) *OCRResultCache {
	return &OCRResultCache{
		cache:   make(map[string]*CacheEntry),
		ttl:     ttl,
		maxSize: maxSize,
	}
}

// Get 获取缓存
func (c *OCRResultCache) Get(key string) *OCRResult {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, exists := c.cache[key]
	if !exists || time.Now().After(entry.ExpiresAt) {
		return nil
	}
	return entry.Result
}

// Set 设置缓存
func (c *OCRResultCache) Set(key string, result *OCRResult) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 检查缓存大小限制
	if len(c.cache) >= c.maxSize {
		c.evictOldest()
	}

	c.cache[key] = &CacheEntry{
		Result:    result,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

// evictOldest 淘汰最旧的缓存
func (c *OCRResultCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.cache {
		if oldestKey == "" || entry.ExpiresAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.ExpiresAt
		}
	}

	if oldestKey != "" {
		delete(c.cache, oldestKey)
	}
}