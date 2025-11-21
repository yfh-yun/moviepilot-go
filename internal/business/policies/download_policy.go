package policies

import (
	"errors"
	"fmt"
	"moviepilot-go/internal/business/domains"
	"time"
)

var (
	ErrDownloadLimitExceeded = errors.New("download limit exceeded")
	ErrInvalidDownloadSize   = errors.New("invalid download size")
	ErrDownloadNotAllowed    = errors.New("download not allowed")
)

// DownloadPolicy 下载策略
type DownloadPolicy struct {
	MaxConcurrentDownloads    int
	MaxDailyDownloadCount     int
	MaxDownloadSize           int64 // bytes
	MaxDownloadSpeed          int64 // bytes per second
	AllowedFileTypes          []string
	BlockedDomains            []string
	DownloadTimeout           time.Duration
	RetryPolicy               *RetryPolicy
	QualityPolicy             *QualityPolicy
}

// RetryPolicy 重试策略
type RetryPolicy struct {
	MaxRetries      int
	BaseDelay       time.Duration
	MaxDelay        time.Duration
	BackoffFactor   float64
	RetryableErrors []string
}

// NewDownloadPolicy 创建新的下载策略
func NewDownloadPolicy() *DownloadPolicy {
	return &DownloadPolicy{
		MaxConcurrentDownloads: 3,
		MaxDailyDownloadCount:  50,
		MaxDownloadSize:        50 * 1024 * 1024 * 1024, // 50GB
		MaxDownloadSpeed:       10 * 1024 * 1024,        // 10MB/s
		AllowedFileTypes: []string{
			".mp4", ".mkv", ".avi", ".mov", ".wmv",
			".mp3", ".flac", ".wav", ".aac",
			".srt", ".ass", ".vtt",
			".nfo", ".txt", ".jpg", ".png",
		},
		BlockedDomains: []string{
			"malware-site.com",
			"blocked-domain.com",
		},
		DownloadTimeout: 30 * time.Minute,
		RetryPolicy: &RetryPolicy{
			MaxRetries:    3,
			BaseDelay:     5 * time.Second,
			MaxDelay:      5 * time.Minute,
			BackoffFactor: 2.0,
			RetryableErrors: []string{
				"timeout",
				"connection refused",
				"network unreachable",
				"temporary failure",
			},
		},
		QualityPolicy: NewQualityPolicy(),
	}
}

// ValidateDownload 验证下载请求
func (p *DownloadPolicy) ValidateDownload(request *domains.DownloadConfig, currentDownloads int, todayDownloads int) error {
	// 检查并发下载限制
	if currentDownloads >= p.MaxConcurrentDownloads {
		return ErrDownloadLimitExceeded
	}
	
	// 检查每日下载限制
	if todayDownloads >= p.MaxDailyDownloadCount {
		return ErrDownloadLimitExceeded
	}
	
	// 检查下载大小限制
	if request.Headers != nil {
		if sizeStr, ok := request.Headers["Content-Length"]; ok {
			// 这里应该解析大小并比较
		}
	}
	
	// 检查超时设置
	if request.Timeout == 0 {
		request.Timeout = p.DownloadTimeout
	} else if request.Timeout > p.DownloadTimeout*2 {
		request.Timeout = p.DownloadTimeout * 2
	}
	
	return nil
}

// ShouldRetry 判断是否应该重试
func (p *DownloadPolicy) ShouldRetry(errorMsg string, attemptCount int) (bool, time.Duration) {
	if attemptCount >= p.RetryPolicy.MaxRetries {
		return false, 0
	}
	
	// 检查是否为可重试错误
	for _, retryableError := range p.RetryPolicy.RetryableErrors {
		if contains(errorMsg, retryableError) {
			delay := p.calculateRetryDelay(attemptCount)
			return true, delay
		}
	}
	
	return false, 0
}

// calculateRetryDelay 计算重试延迟
func (p *DownloadPolicy) calculateRetryDelay(attemptCount int) time.Duration {
	delay := time.Duration(float64(p.RetryPolicy.BaseDelay) * 
		p.RetryPolicy.BackoffFactor * float64(attemptCount))
	
	if delay > p.RetryPolicy.MaxDelay {
		delay = p.RetryPolicy.MaxDelay
	}
	
	return delay
}

// IsAllowedDomain 检查域名是否被允许
func (p *DownloadPolicy) IsAllowedDomain(url string) bool {
	// 简化实现，实际应该解析URL并检查域名
	for _, blocked := range p.BlockedDomains {
		if contains(url, blocked) {
			return false
		}
	}
	return true
}

// IsAllowedFileType 检查文件类型是否被允许
func (p *DownloadPolicy) IsAllowedFileType(filename string) bool {
	for _, allowedType := range p.AllowedFileTypes {
		if endsWith(filename, allowedType) {
			return true
		}
	}
	return false
}

// GetDownloadPriority 获取下载优先级
func (p *DownloadPolicy) GetDownloadPriority(request *domains.DownloadConfig) int {
	priority := 0
	
	// 根据质量调整优先级
	if request.Headers != nil {
		if quality, ok := request.Headers["X-Quality"].(string); ok {
			qualityObj := p.QualityPolicy.ParseQuality(quality)
			score := p.QualityPolicy.GetQualityScore(qualityObj)
			priority += int(score)
		}
	}
	
	// 根据用户类型调整优先级
	if request.Headers != nil {
		if userType, ok := request.Headers["X-User-Type"].(string); ok {
			switch userType {
			case "vip":
				priority += 100
			case "premium":
				priority += 50
			}
		}
	}
	
	return priority
}

// EstimateDownloadTime 估算下载时间
func (p *DownloadPolicy) EstimateDownloadTime(fileSize int64) time.Duration {
	speed := p.MaxDownloadSpeed
	if speed == 0 {
		speed = 1024 * 1024 // 默认1MB/s
	}
	
	return time.Duration(fileSize/speed) * time.Second
}

// ShouldThrottle 判断是否应该限速
func (p *DownloadPolicy) ShouldThrottle(currentTime time.Time) bool {
	// 在高峰时段限速
	hour := currentTime.Hour()
	return (hour >= 19 && hour <= 23) || (hour >= 0 && hour <= 6)
}

// GetThrottledSpeed 获取限速后的速度
func (p *DownloadPolicy) GetThrottledSpeed(originalSpeed int64) int64 {
	if !p.ShouldThrottle(time.Now()) {
		return originalSpeed
	}
	
	// 高峰时段限速到50%
	return originalSpeed / 2
}

// ValidateDownloadURL 验证下载URL
func (p *DownloadPolicy) ValidateDownloadURL(url string) error {
	if url == "" {
		return errors.New("download URL cannot be empty")
	}
	
	if !p.IsAllowedDomain(url) {
		return ErrDownloadNotAllowed
	}
	
	// 可以添加更多URL验证逻辑
	
	return nil
}

// GetRetryStrategy 获取重试策略
func (p *DownloadPolicy) GetRetryStrategy() *RetryPolicy {
	return p.RetryPolicy
}

// UpdateDownloadStats 更新下载统计
func (p *DownloadPolicy) UpdateDownloadStats(stats map[string]interface{}, downloadSize int64, duration time.Duration) {
	stats["total_downloads"] = stats["total_downloads"].(int) + 1
	stats["total_bytes"] = stats["total_bytes"].(int64) + downloadSize
	stats["total_duration"] = stats["total_duration"].(time.Duration) + duration
	
	// 计算平均速度
	if duration > 0 {
		avgSpeed := downloadSize / int64(duration.Seconds())
		if stats["avg_speed"] == nil {
			stats["avg_speed"] = avgSpeed
		} else {
			// 简单的移动平均
			currentAvg := stats["avg_speed"].(int64)
			stats["avg_speed"] = (currentAvg + avgSpeed) / 2
		}
	}
}

// IsDownloadHealthy 检查下载是否健康
func (p *DownloadPolicy) IsDownloadHealthystats(stats map[string]interface{}) bool {
	// 检查失败率
	if total, ok := stats["total_downloads"].(int); ok && total > 0 {
		if failed, ok := stats["failed_downloads"].(int); ok {
			failureRate := float64(failed) / float64(total)
			if failureRate > 0.2 { // 失败率超过20%
				return false
			}
		}
	}
	
	// 检查平均速度
	if avgSpeed, ok := stats["avg_speed"].(int64); ok {
		if avgSpeed < 1024*1024 { // 平均速度低于1MB/s
			return false
		}
	}
	
	return true
}

// GetOptimizationSuggestions 获取优化建议
func (p *DownloadPolicy) GetOptimizationSuggestions(stats map[string]interface{}) []string {
	var suggestions []string
	
	// 检查并发下载
	if concurrent, ok := stats["current_concurrent"].(int); ok {
		if concurrent >= p.MaxConcurrentDownloads {
			suggestions = append(suggestions, "考虑增加最大并发下载数")
		}
	}
	
	// 检查下载速度
	if avgSpeed, ok := stats["avg_speed"].(int64); ok {
		if avgSpeed < p.MaxDownloadSpeed/2 {
			suggestions = append(suggestions, "检查网络连接或考虑更换下载源")
		}
	}
	
	// 检查失败率
	if total, ok := stats["total_downloads"].(int); ok && total > 0 {
		if failed, ok := stats["failed_downloads"].(int); ok {
			failureRate := float64(failed) / float64(total)
			if failureRate > 0.1 {
				suggestions = append(suggestions, "失败率较高，建议检查下载源质量")
			}
		}
	}
	
	return suggestions
}

// 辅助函数
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && 
			(s[:len(substr)] == substr || 
			 s[len(s)-len(substr):] == substr || 
			 indexOf(s, substr) >= 0)))
}

func endsWith(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}