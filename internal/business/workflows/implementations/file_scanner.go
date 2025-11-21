// Package implementations 提供动作系统的具体实现
package implementations

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"moviepilot-go/internal/business/workflows/base"
	"moviepilot-go/internal/business/workflows/interfaces"
	"moviepilot-go/internal/business/workflows/types"
	"moviepilot-go/pkg/logger"
)

// FileScanner 文件扫描动作
// 对应Python版本的ScanFileAction
type FileScanner struct {
	*base.Action
	config *FileScannerConfig
	mutex  sync.RWMutex
}

// FileScannerConfig 文件扫描器配置
type FileScannerConfig struct {
	ScanPath         []string          `json:"scan_path" description:"扫描路径列表"`
	ExcludePatterns  []string          `json:"exclude_patterns" description:"排除模式"`
	IncludePatterns  []string          `json:"include_patterns" description:"包含模式"`
	MaxFileSize      int64             `json:"max_file_size" description:"最大文件大小(字节)"`
	EnableHashCheck  bool              `json:"enable_hash_check" description:"启用哈希校验"`
	EnableVirusScan  bool              `json:"enable_virus_scan" description:"启用病毒扫描"`
	CacheTimeout     time.Duration     `json:"cache_timeout" description:"缓存超时时间"`
	ParallelScans    int               `json:"parallel_scans" description:"并行扫描数量"`
	CustomValidators []string          `json:"custom_validators" description:"自定义验证器"`
	FileTypeFilters  map[string]string `json:"file_type_filters" description:"文件类型过滤器"`
}

// ScanResult 扫描结果
type ScanResult struct {
	FilePath      string            `json:"file_path"`
	FileName      string            `json:"file_name"`
	FileSize      int64             `json:"file_size"`
	FileType      string            `json:"file_type"`
	Hash          string            `json:"hash,omitempty"`
	IsSafe        bool              `json:"is_safe"`
	ScanTime      time.Time         `json:"scan_time"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	Threats       []string          `json:"threats,omitempty"`
	ErrorMessage  string            `json:"error_message,omitempty"`
}

// NewFileScanner 创建文件扫描器实例
func NewFileScanner() interfaces.Action {
	return &FileScanner{
		Action: base.NewAction("FileScanner", "文件扫描器，支持安全检查、病毒扫描和内容分析"),
		config: &FileScannerConfig{
			ScanPath:         []string{"/downloads", "/media"},
			ExcludePatterns:  []string{"*.tmp", "*.part", ".*"},
			IncludePatterns:  []string{"*.mp4", "*.mkv", "*.avi", "*.mov"},
			MaxFileSize:      10 * 1024 * 1024 * 1024, // 10GB
			EnableHashCheck:  true,
			EnableVirusScan:  false,
			CacheTimeout:     30 * time.Minute,
			ParallelScans:    4,
			FileTypeFilters: map[string]string{
				"video": ".mp4,.mkv,.avi,.mov,.wmv,.flv,.webm",
				"audio": ".mp3,.flac,.wav,.aac,.ogg",
				"subtitle": ".srt,.ass,.ssa,.vtt",
			},
		},
	}
}

// Execute 执行文件扫描
func (fs *FileScanner) Execute(ctx context.Context, workflowID int64, params map[string]interface{}, actionContext *types.ActionContext) (*types.ActionContext, error) {
	logger.Debug("FileScanner execution started", 
		"workflow_id", workflowID,
		"action", "FileScanner")

	// 检查缓存
	if fs.CheckCache(ctx, workflowID, "scan_result") {
		logger.Info("Using cached scan result", "workflow_id", workflowID)
		cachedData, err := fs.GetCache(ctx, workflowID, "scan_result")
		if err == nil {
			if result, ok := cachedData.(*ScanResult); ok {
				fs.SetData("scan_result", result)
				fs.SetDone(fmt.Sprintf("使用缓存结果: %s", result.FilePath))
				return actionContext, nil
			}
		}
	}

	// 解析参数
	config, err := fs.parseConfig(params)
	if err != nil {
		fs.SetError(fmt.Sprintf("参数解析失败: %v", err))
		return actionContext, err
	}

	// 执行扫描
	result, err := fs.performScan(ctx, config, actionContext)
	if err != nil {
		fs.SetError(fmt.Sprintf("文件扫描失败: %v", err))
		return actionContext, err
	}

	// 保存缓存
	if err := fs.SaveCache(ctx, workflowID, "scan_result", result, fs.config.CacheTimeout); err != nil {
		logger.Warn("Failed to save scan result cache", "error", err)
	}

	// 设置结果
	fs.SetData("scan_result", result)
	fs.SetData("file_count", len(result))
	fs.SetDone(fmt.Sprintf("成功扫描 %d 个文件", len(result)))

	logger.Info("FileScanner execution completed", 
		"workflow_id", workflowID,
		"file_count", len(result))

	return actionContext, nil
}

// parseConfig 解析配置参数
func (fs *FileScanner) parseConfig(params map[string]interface{}) (*FileScannerConfig, error) {
	config := *fs.config // 复制默认配置

	if scanPath, ok := params["scan_path"].([]string); ok {
		config.ScanPath = scanPath
	}

	if excludePatterns, ok := params["exclude_patterns"].([]string); ok {
		config.ExcludePatterns = excludePatterns
	}

	if includePatterns, ok := params["include_patterns"].([]string); ok {
		config.IncludePatterns = includePatterns
	}

	if maxFileSize, ok := params["max_file_size"].(float64); ok {
		config.MaxFileSize = int64(maxFileSize)
	}

	if enableHashCheck, ok := params["enable_hash_check"].(bool); ok {
		config.EnableHashCheck = enableHashCheck
	}

	if enableVirusScan, ok := params["enable_virus_scan"].(bool); ok {
		config.EnableVirusScan = enableVirusScan
	}

	return &config, nil
}

// performScan 执行实际的文件扫描
func (fs *FileScanner) performScan(ctx context.Context, config *FileScannerConfig, actionContext *types.ActionContext) ([]*ScanResult, error) {
	var results []*ScanResult
	var mu sync.Mutex

	// 创建工作池
	jobs := make(chan string, config.ParallelScans*2)
	var wg sync.WaitGroup

	// 启动工作协程
	for i := 0; i < config.ParallelScans; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filePath := range jobs {
				result := fs.scanFile(ctx, filePath, config)
				mu.Lock()
				results = append(results, result)
				mu.Unlock()
			}
		}()
	}

	// 遍历扫描路径
	for _, scanPath := range config.ScanPath {
		err := filepath.Walk(scanPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				logger.Warn("Failed to access path", "path", path, "error", err)
				return nil
			}

			// 检查上下文是否已取消
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			// 跳过目录
			if info.IsDir() {
				return nil
			}

			// 应用过滤器
			if fs.shouldSkipFile(path, info, config) {
				return nil
			}

			// 添加到工作队列
			select {
			case jobs <- path:
			case <-ctx.Done():
				return ctx.Err()
			}

			return nil
		})

		if err != nil {
			logger.Warn("Failed to scan path", "path", scanPath, "error", err)
		}
	}

	// 关闭工作队列
	close(jobs)
	wg.Wait()

	return results, nil
}

// shouldSkipFile 检查是否应该跳过文件
func (fs *FileScanner) shouldSkipFile(filePath string, info os.FileInfo, config *FileScannerConfig) bool {
	fileName := filepath.Base(filePath)

	// 检查文件大小
	if info.Size() > config.MaxFileSize {
		return true
	}

	// 检查排除模式
	for _, pattern := range config.ExcludePatterns {
		if matched, _ := filepath.Match(pattern, fileName); matched {
			return true
		}
	}

	// 检查包含模式
	if len(config.IncludePatterns) > 0 {
		included := false
		for _, pattern := range config.IncludePatterns {
			if matched, _ := filepath.Match(pattern, fileName); matched {
				included = true
				break
			}
		}
		if !included {
			return true
		}
	}

	return false
}

// scanFile 扫描单个文件
func (fs *FileScanner) scanFile(ctx context.Context, filePath string, config *FileScannerConfig) *ScanResult {
	result := &ScanResult{
		FilePath: filePath,
		FileName: filepath.Base(filePath),
		ScanTime: time.Now(),
		IsSafe:   true,
		Metadata: make(map[string]string),
	}

	// 获取文件信息
	info, err := os.Stat(filePath)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("无法获取文件信息: %v", err)
		result.IsSafe = false
		return result
	}

	result.FileSize = info.Size()
	result.FileType = fs.getFileType(filePath)

	// 计算文件哈希
	if config.EnableHashCheck {
		hash, err := fs.calculateFileHash(filePath)
		if err != nil {
			result.ErrorMessage = fmt.Sprintf("哈希计算失败: %v", err)
			result.IsSafe = false
		} else {
			result.Hash = hash
		}
	}

	// 病毒扫描
	if config.EnableVirusScan {
		threats, err := fs.performVirusScan(ctx, filePath)
		if err != nil {
			result.ErrorMessage = fmt.Sprintf("病毒扫描失败: %v", err)
			result.IsSafe = false
		} else if len(threats) > 0 {
			result.Threats = threats
			result.IsSafe = false
		}
	}

	// 提取文件元数据
	fs.extractMetadata(filePath, result)

	return result
}

// getFileType 获取文件类型
func (fs *FileScanner) getFileType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	
	switch {
	case strings.Contains(".mp4,.mkv,.avi,.mov,.wmv,.flv,.webm", ext):
		return "video"
	case strings.Contains(".mp3,.flac,.wav,.aac,.ogg", ext):
		return "audio"
	case strings.Contains(".srt,.ass,.ssa,.vtt", ext):
		return "subtitle"
	default:
		return "other"
	}
}

// calculateFileHash 计算文件哈希
func (fs *FileScanner) calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// performVirusScan 执行病毒扫描
func (fs *FileScanner) performVirusScan(ctx context.Context, filePath string) ([]string, error) {
	// 这里可以集成实际的杀毒引擎
	// 目前返回空威胁列表表示安全
	return []string{}, nil
}

// extractMetadata 提取文件元数据
func (fs *FileScanner) extractMetadata(filePath string, result *ScanResult) {
	info, err := os.Stat(filePath)
	if err != nil {
		return
	}

	result.Metadata["mod_time"] = info.ModTime().Format(time.RFC3339)
	result.Metadata["size_mb"] = fmt.Sprintf("%.2f", float64(result.FileSize)/(1024*1024))
	result.Metadata["extension"] = filepath.Ext(filePath)
	result.Metadata["directory"] = filepath.Dir(filePath)
}

// Initialize 初始化文件扫描器
func (fs *FileScanner) Initialize() error {
	logger.Info("Initializing FileScanner", 
		"scan_paths", fs.config.ScanPath,
		"parallel_scans", fs.config.ParallelScans)
	return nil
}

// Cleanup 清理资源
func (fs *FileScanner) Cleanup() error {
	logger.Info("Cleaning up FileScanner")
	return nil
}