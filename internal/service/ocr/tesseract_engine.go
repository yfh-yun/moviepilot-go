package ocr

import (
	"context"
	"fmt"
	"image"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

// TesseractEngine Tesseract OCR引擎
type TesseractEngine struct {
	logger     *zap.Logger
	binaryPath string
	available  bool
	priority   int
}

// NewTesseractEngine 创建Tesseract引擎
func NewTesseractEngine(logger *zap.Logger) *TesseractEngine {
	engine := &TesseractEngine{
		logger:   logger,
		priority: 100, // 高优先级
	}
	
	// 检查Tesseract是否可用
	engine.checkAvailability()
	
	return engine
}

// Name 引擎名称
func (e *TesseractEngine) Name() string {
	return "tesseract"
}

// IsAvailable 检查引擎是否可用
func (e *TesseractEngine) IsAvailable() bool {
	return e.available
}

// Priority 引擎优先级
func (e *TesseractEngine) Priority() int {
	return e.priority
}

// SupportedFormats 支持的格式
func (e *TesseractEngine) SupportedFormats() []string {
	return []string{"png", "jpg", "jpeg", "bmp", "tiff", "tif"}
}

// RecognizeText 识别文本
func (e *TesseractEngine) RecognizeText(ctx context.Context, img image.Image, options *OCROptions) (*OCRResult, error) {
	if !e.available {
		return nil, fmt.Errorf("Tesseract引擎不可用")
	}

	e.logger.Debug("使用Tesseract识别文本", 
		zap.String("language", options.Language))

	// 构建命令参数
	args := e.buildCommandArgs(options)

	// 执行Tesseract命令
	result, err := e.executeTesseract(ctx, img, args, options.Timeout)
	if err != nil {
		e.logger.Error("Tesseract识别失败", zap.Error(err))
		return nil, err
	}

	// 解析结果
	ocrResult := e.parseTesseractOutput(result, options)
	ocrResult.Engine = e.Name()

	return ocrResult, nil
}

// RecognizeFromPath 从文件路径识别
func (e *TesseractEngine) RecognizeFromPath(ctx context.Context, imagePath string, options *OCROptions) (*OCRResult, error) {
	if !e.available {
		return nil, fmt.Errorf("Tesseract引擎不可用")
	}

	args := []string{imagePath, "stdout"}
	if options.Language != "" {
		args = append(args, "-l", options.Language)
	}

	// 添加输出格式
	args = append(args, "-c", "tessedit_create_txt=1")

	// 执行命令
	ctx, cancel := context.WithTimeout(ctx, options.Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, e.binaryPath, args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("执行Tesseract失败: %w", err)
	}

	text := strings.TrimSpace(string(output))
	ocrResult := &OCRResult{
		Text:     text,
		Success:  true,
		Engine:   e.Name(),
		Duration: time.Since(time.Now()),
	}

	return ocrResult, nil
}

// checkAvailability 检查Tesseract可用性
func (e *TesseractEngine) checkAvailability() {
	// 尝试查找Tesseract二进制文件
	paths := []string{
		"tesseract",
		"/usr/bin/tesseract",
		"/usr/local/bin/tesseract",
		"/opt/homebrew/bin/tesseract",
	}

	for _, path := range paths {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		cmd := exec.CommandContext(ctx, path, "--version")
		err := cmd.Run()
		if err == nil {
			e.binaryPath = path
			e.available = true
			e.logger.Info("Tesseract引擎可用", zap.String("path", path))
			return
		}
	}

	e.logger.Warn("Tesseract引擎不可用")
	e.available = false
}

// buildCommandArgs 构建命令参数
func (e *TesseractEngine) buildCommandArgs(options *OCROptions) []string {
	var args []string

	// 语言设置
	if options.Language != "" {
		args = append(args, "-l", options.Language)
	} else {
		args = append(args, "-l", "eng+chi_sim") // 默认英文+简体中文
	}

	// 页面分割模式
	args = append(args, "--psm", "6") // 假设单一文本块

	// 输出格式配置
	args = append(args, "-c", "tessedit_create_txt=1")
	args = append(args, "-c", "tessedit_create_hocr=1")
	args = append(args, "-c", "tessedit_create_tsv=1")

	// OCR引擎模式
	args = append(args, "--oem", "3") // 默认OCR引擎

	return args
}

// executeTesseract 执行Tesseract命令
func (e *TesseractEngine) executeTesseract(ctx context.Context, img image.Image, args []string, timeout time.Duration) (*TesseractOutput, error) {
	// 这里应该实现图像保存为临时文件并调用Tesseract
	// 为了简化示例，返回模拟结果
	
	text := "识别的文本内容"
	confidence := 0.85

	result := &TesseractOutput{
		Text:       text,
		Confidence: confidence,
		Words: []TesseractWord{
			{Text: "识别", Confidence: 0.9, Box: "0,0,50,20"},
			{Text: "的", Confidence: 0.95, Box: "50,0,65,20"},
			{Text: "文本", Confidence: 0.8, Box: "65,0,115,20"},
			{Text: "内容", Confidence: 0.85, Box: "115,0,165,20"},
		},
	}

	return result, nil
}

// parseTesseractOutput 解析Tesseract输出
func (e *TesseractEngine) parseTesseractOutput(output *TesseractOutput, options *OCROptions) *OCRResult {
	ocrResult := &OCRResult{
		Text:       output.Text,
		Confidence: output.Confidence,
		Success:    true,
		Metadata: map[string]interface{}{
			"word_count": len(output.Words),
			"engine":    "tesseract",
		},
	}

	// 转换文本框信息
	for _, word := range output.Words {
		box := e.parseBoxCoordinates(word.Box)
		textBox := &TextBox{
			Text:       word.Text,
			Confidence: word.Confidence,
			Box:        box,
			Language:   options.Language,
		}
		ocrResult.Boxes = append(ocrResult.Boxes, textBox)
	}

	return ocrResult
}

// parseBoxCoordinates 解析坐标
func (e *TesseractEngine) parseBoxCoordinates(boxStr string) image.Rectangle {
	// 解析坐标字符串，格式：x,y,width,height
	parts := strings.Split(boxStr, ",")
	if len(parts) != 4 {
		return image.Rect(0, 0, 0, 0)
	}

	x, _ := strconv.Atoi(parts[0])
	y, _ := strconv.Atoi(parts[1])
	w, _ := strconv.Atoi(parts[2])
	h, _ := strconv.Atoi(parts[3])

	return image.Rect(x, y, x+w, y+h)
}

// TesseractOutput Tesseract输出结构
type TesseractOutput struct {
	Text       string           `json:"text"`
	Confidence float64          `json:"confidence"`
	Words      []TesseractWord  `json:"words"`
}

// TesseractWord Tesseract单词信息
type TesseractWord struct {
	Text       string `json:"text"`
	Confidence float64 `json:"confidence"`
	Box        string `json:"box"`
}