// Package actions 提供动作系统的业务逻辑实现
package actions

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

	"github.com/yfh-yun/moviepilot-go/pkg/logger"

	"go.uber.org/zap"
)

// FileScanner 文件扫描器
// 提供文件安全检查、病毒扫描和内容分析功能
type FileScanner struct {
	antivirusEngine AntivirusEngine
	hashDatabase    HashDatabase
	logger          *zap.Logger
	scanCache       *ScanCache
	mutex           sync.RWMutex
}

// AntivirusEngine 杀毒引擎接口
type AntivirusEngine interface {
	ScanFile(ctx context.Context, filePath string) (*ScanResult, error)
	GetEngineInfo() *EngineInfo
	UpdateVirusDefinitions() error
}

// HashDatabase 哈希数据库接口
type HashDatabase interface {
	CheckHash(hash string) (*HashResult, error)
	AddHash(hash string, result *HashResult) error
	GetMaliciousHashes() ([]string, error)
}

// NewFileScanner 创建文件扫描器实例
func NewFileScanner(
	antivirusEngine AntivirusEngine,
	hashDatabase HashDatabase,
) *FileScanner {
	return &FileScanner{
		antivirusEngine: antivirusEngine,
		hashDatabase:    hashDatabase,
		logger:          logger.NewLogger("file_scanner"),
		scanCache:       NewScanCache(),
	}
}

// ScanFile 扫描文件
func (s *FileScanner) ScanFile(ctx context.Context, request *ScanRequest) (*ScanResponse, error) {
	s.logger.Info("开始文件扫描",
		zap.String("file_path", request.FilePath),
		zap.String("scan_type", request.ScanType))

	// 1. 基础文件检查
	fileInfo, err := s.validateFile(request.FilePath)
	if err != nil {
		return nil, fmt.Errorf("文件验证失败: %w", err)
	}

	// 2. 计算文件哈希
	fileHash, err := s.calculateFileHash(request.FilePath)
	if err != nil {
		return nil, fmt.Errorf("文件哈希计算失败: %w", err)
	}

	// 3. 检查缓存
	if cachedResult, exists := s.scanCache.Get(fileHash); exists {
		s.logger.Info("使用缓存扫描结果", zap.String("file_hash", fileHash))
		return &ScanResponse{
			FilePath:   request.FilePath,
			FileHash:   fileHash,
			ScanResult: cachedResult,
			FromCache:  true,
			ScannedAt:  time.Now(),
		}, nil
	}

	// 4. 执行扫描
	scanResult, err := s.performScan(ctx, request, fileHash)
	if err != nil {
		return nil, fmt.Errorf("文件扫描失败: %w", err)
	}

	// 5. 缓存结果
	s.scanCache.Set(fileHash, scanResult, 24*time.Hour)

	response := &ScanResponse{
		FilePath:   request.FilePath,
		FileHash:   fileHash,
		FileSize:   fileInfo.Size(),
		ScanResult: scanResult,
		FromCache:  false,
		ScannedAt:  time.Now(),
	}

	s.logger.Info("文件扫描完成",
		zap.String("file_path", request.FilePath),
		zap.String("scan_result", scanResult.ThreatLevel),
		zap.Bool("is_threat", scanResult.IsThreat))

	return response, nil
}

// validateFile 验证文件
func (s *FileScanner) validateFile(filePath string) (os.FileInfo, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("文件不存在或无法访问: %w", err)
	}

	if fileInfo.IsDir() {
		return nil, fmt.Errorf("路径指向目录而非文件")
	}

	if fileInfo.Size() == 0 {
		return nil, fmt.Errorf("文件为空")
	}

	if fileInfo.Size() > 100*1024*1024 { // 100MB限制
		return nil, fmt.Errorf("文件过大: %d bytes", fileInfo.Size())
	}

	return fileInfo, nil
}

// calculateFileHash 计算文件哈希
func (s *FileScanner) calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("读取文件失败: %w", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// performScan 执行扫描
func (s *FileScanner) performScan(ctx context.Context, request *ScanRequest, fileHash string) (*ScanResult, error) {
	scanResult := &ScanResult{
		FilePath:    request.FilePath,
		FileHash:    fileHash,
		ScanType:    request.ScanType,
		ThreatLevel: "clean",
		IsThreat:    false,
		ScanTime:    time.Now(),
	}

	// 1. 哈希数据库检查
	if hashResult, err := s.hashDatabase.CheckHash(fileHash); err == nil && hashResult != nil {
		if hashResult.IsMalicious {
			scanResult.ThreatLevel = "malicious"
			scanResult.IsThreat = true
			scanResult.Detections = append(scanResult.Detections, "Known malicious file hash")
			return scanResult, nil
		}
	}

	// 2. 文件类型分析
	if err := s.analyzeFileType(request.FilePath, scanResult); err != nil {
		s.logger.Warn("文件类型分析失败", zap.Error(err))
	}

	// 3. 病毒扫描（如果启用）
	if request.EnableAntivirus {
		if antivirusResult, err := s.antivirusEngine.ScanFile(ctx, request.FilePath); err == nil {
			if antivirusResult.IsThreat {
				scanResult.ThreatLevel = "infected"
				scanResult.IsThreat = true
				scanResult.Detections = append(scanResult.Detections, antivirusResult.ThreatName)
			}
		} else {
			s.logger.Warn("病毒扫描失败", zap.Error(err))
		}
	}

	// 4. 启发式分析
	if err := s.heuristicAnalysis(request.FilePath, scanResult); err != nil {
		s.logger.Warn("启发式分析失败", zap.Error(err))
	}

	// 5. 风险评估
	s.assessRiskLevel(scanResult)

	return scanResult, nil
}

// analyzeFileType 分析文件类型
func (s *FileScanner) analyzeFileType(filePath string, result *ScanResult) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 读取文件头（前512字节）
	header := make([]byte, 512)
	n, err := file.Read(header)
	if err != nil && err != io.EOF {
		return err
	}

	// 文件类型检测
	fileType := s.detectFileType(header[:n], filepath.Ext(filePath))
	result.FileType = fileType

	// 检查可疑文件类型
	if s.isSuspiciousFileType(fileType) {
		result.RiskScore += 30
		result.Warnings = append(result.Warnings, fmt.Sprintf("可疑文件类型: %s", fileType))
	}

	return nil
}

// detectFileType 检测文件类型
func (s *FileScanner) detectFileType(header []byte, extension string) string {
	// 基于文件头的类型检测
	magicNumbers := map[string]string{
		"\x89PNG\r\n\x1a\n": "PNG Image",
		"\xff\xd8\xff":      "JPEG Image",
		"GIF87a":            "GIF Image",
		"GIF89a":            "GIF Image",
		"%PDF":              "PDF Document",
		"PK\x03\x04":        "ZIP Archive",
		"Rar!":              "RAR Archive",
		"MZ":                "Windows Executable",
		"\x7fELF":           "ELF Executable",
		"#!/":               "Script File",
	}

	for magic, fileType := range magicNumbers {
		if len(header) >= len(magic) && string(header[:len(magic)]) == magic {
			return fileType
		}
	}

	// 基于扩展名的类型检测
	extensionMap := map[string]string{
		".exe":  "Windows Executable",
		".dll":  "Windows Library",
		".bat":  "Batch Script",
		".cmd":  "Command Script",
		".ps1":  "PowerShell Script",
		".sh":   "Shell Script",
		".py":   "Python Script",
		".js":   "JavaScript",
		".html": "HTML Document",
		".htm":  "HTML Document",
	}

	if fileType, exists := extensionMap[strings.ToLower(extension)]; exists {
		return fileType
	}

	return "Unknown"
}

// isSuspiciousFileType 检查可疑文件类型
func (s *FileScanner) isSuspiciousFileType(fileType string) bool {
	suspiciousTypes := []string{
		"Windows Executable",
		"Windows Library",
		"Batch Script",
		"Command Script",
		"PowerShell Script",
		"Shell Script",
		"Python Script",
		"JavaScript",
	}

	for _, suspiciousType := range suspiciousTypes {
		if fileType == suspiciousType {
			return true
		}
	}

	return false
}

// heuristicAnalysis 启发式分析
func (s *FileScanner) heuristicAnalysis(filePath string, result *ScanResult) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	contentStr := string(content)

	// 检查可疑字符串模式
	suspiciousPatterns := []string{
		"eval(",
		"exec(",
		"system(",
		"shell_exec(",
		"passthru(",
		"base64_decode",
		"gzinflate",
		"document.cookie",
		"<script>",
		"javascript:",
	}

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(contentStr, pattern) {
			result.RiskScore += 10
			result.Warnings = append(result.Warnings, fmt.Sprintf("发现可疑模式: %s", pattern))
		}
	}

	// 检查文件编码异常
	if s.detectEncodingAnomalies(content) {
		result.RiskScore += 15
		result.Warnings = append(result.Warnings, "检测到编码异常")
	}

	return nil
}

// detectEncodingAnomalies 检测编码异常
func (s *FileScanner) detectEncodingAnomalies(content []byte) bool {
	// 检查是否存在过多的控制字符
	controlCharCount := 0
	for _, b := range content {
		if b < 32 && b != 9 && b != 10 && b != 13 { // 排除制表符、换行符、回车符
			controlCharCount++
		}
	}

	// 如果控制字符超过内容的5%，则认为是异常
	if len(content) > 0 && float64(controlCharCount)/float64(len(content)) > 0.05 {
		return true
	}

	return false
}

// assessRiskLevel 风险评估
func (s *FileScanner) assessRiskLevel(result *ScanResult) {
	// 根据风险分数确定威胁等级
	if result.RiskScore >= 80 {
		result.ThreatLevel = "high"
		result.IsThreat = true
	} else if result.RiskScore >= 60 {
		result.ThreatLevel = "medium"
		result.IsThreat = true
	} else if result.RiskScore >= 40 {
		result.ThreatLevel = "low"
		result.IsThreat = false
	} else {
		result.ThreatLevel = "clean"
		result.IsThreat = false
	}
}

// BatchScanFiles 批量扫描文件
func (s *FileScanner) BatchScanFiles(ctx context.Context, requests []*ScanRequest) ([]*ScanResponse, error) {
	var wg sync.WaitGroup
	responses := make([]*ScanResponse, len(requests))
	errors := make([]error, len(requests))

	for i, request := range requests {
		wg.Add(1)

		go func(index int, req *ScanRequest) {
			defer wg.Done()

			response, err := s.ScanFile(ctx, req)
			responses[index] = response
			errors[index] = err
		}(i, request)
	}

	wg.Wait()

	// 检查错误
	for _, err := range errors {
		if err != nil {
			return nil, fmt.Errorf("批量扫描失败: %w", err)
		}
	}

	return responses, nil
}

// GetScanStatistics 获取扫描统计信息
func (s *FileScanner) GetScanStatistics() *ScanStatistics {
	return &ScanStatistics{
		TotalScans:   s.scanCache.GetTotalScans(),
		CacheHits:    s.scanCache.GetCacheHits(),
		ThreatsFound: s.scanCache.GetThreatsFound(),
		LastScanTime: time.Now(),
	}
}

// ScanRequest 扫描请求
type ScanRequest struct {
	FilePath        string `json:"file_path"`
	ScanType        string `json:"scan_type"` // "quick", "full", "custom"
	EnableAntivirus bool   `json:"enable_antivirus"`
	ScanDepth       int    `json:"scan_depth"` // 扫描深度
}

// ScanResponse 扫描响应
type ScanResponse struct {
	FilePath   string      `json:"file_path"`
	FileHash   string      `json:"file_hash"`
	FileSize   int64       `json:"file_size"`
	ScanResult *ScanResult `json:"scan_result"`
	FromCache  bool        `json:"from_cache"`
	ScannedAt  time.Time   `json:"scanned_at"`
}

// ScanResult 扫描结果
type ScanResult struct {
	FilePath    string    `json:"file_path"`
	FileHash    string    `json:"file_hash"`
	FileType    string    `json:"file_type"`
	ScanType    string    `json:"scan_type"`
	ThreatLevel string    `json:"threat_level"` // "clean", "low", "medium", "high", "malicious", "infected"
	IsThreat    bool      `json:"is_threat"`
	RiskScore   int       `json:"risk_score"`
	Detections  []string  `json:"detections"`
	Warnings    []string  `json:"warnings"`
	ScanTime    time.Time `json:"scan_time"`
}

// ScanStatistics 扫描统计
type ScanStatistics struct {
	TotalScans   int       `json:"total_scans"`
	CacheHits    int       `json:"cache_hits"`
	ThreatsFound int       `json:"threats_found"`
	LastScanTime time.Time `json:"last_scan_time"`
}

// EngineInfo 引擎信息
type EngineInfo struct {
	Name                    string    `json:"name"`
	Version                 string    `json:"version"`
	VirusDefinitionsVersion string    `json:"virus_definitions_version"`
	LastUpdate              time.Time `json:"last_update"`
}

// HashResult 哈希结果
type HashResult struct {
	Hash        string    `json:"hash"`
	IsMalicious bool      `json:"is_malicious"`
	ThreatName  string    `json:"threat_name"`
	FirstSeen   time.Time `json:"first_seen"`
	LastSeen    time.Time `json:"last_seen"`
}

// ScanCache 扫描缓存
type ScanCache struct {
	cache        map[string]*cacheEntry
	mutex        sync.RWMutex
	totalScans   int
	cacheHits    int
	threatsFound int
}

// cacheEntry 缓存条目
type cacheEntry struct {
	result    *ScanResult
	expiresAt time.Time
}

// NewScanCache 创建扫描缓存
func NewScanCache() *ScanCache {
	return &ScanCache{
		cache: make(map[string]*cacheEntry),
	}
}

// Get 获取缓存结果
func (c *ScanCache) Get(hash string) (*ScanResult, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, exists := c.cache[hash]
	if !exists || time.Now().After(entry.expiresAt) {
		return nil, false
	}

	c.cacheHits++
	if entry.result.IsThreat {
		c.threatsFound++
	}

	return entry.result, true
}

// Set 设置缓存结果
func (c *ScanCache) Set(hash string, result *ScanResult, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache[hash] = &cacheEntry{
		result:    result,
		expiresAt: time.Now().Add(ttl),
	}
	c.totalScans++
}

// GetTotalScans 获取总扫描次数
func (c *ScanCache) GetTotalScans() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.totalScans
}

// GetCacheHits 获取缓存命中次数
func (c *ScanCache) GetCacheHits() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.cacheHits
}

// GetThreatsFound 获取发现的威胁数量
func (c *ScanCache) GetThreatsFound() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.threatsFound
}
