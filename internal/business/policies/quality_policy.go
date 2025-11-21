package policies

import (
	"fmt"
	"moviepilot-go/internal/business/domains"
	"strings"
)

// QualityPolicy 质量策略
type QualityPolicy struct {
	QualityHierarchy map[string]int
	SourceHierarchy  map[string]int
	CodecHierarchy   map[string]int
	MinQuality       domains.Quality
	MaxQuality       domains.Quality
	PreferredQualities []domains.Quality
}

// NewQualityPolicy 创建新的质量策略
func NewQualityPolicy() *QualityPolicy {
	return &QualityPolicy{
		QualityHierarchy: map[string]int{
			"480p": 1,
			"576p": 2,
			"720p": 3,
			"1080p": 4,
			"1440p": 5,
			"2K": 5,
			"4K": 6,
			"8K": 7,
		},
		SourceHierarchy: map[string]int{
			"CAM":      1,
			"TS":       2,
			"TC":       3,
			"DVDSCR":   4,
			"Screener": 4,
			"DVD":      5,
			"HDTV":     6,
			"HDCAM":    6,
			"HDTS":     6,
			"WEB-DL":   7,
			"WEBRip":   7,
			"WEB":      7,
			"BDRip":    8,
			"BluRay":   9,
			"Remux":    10,
			"UHD":      11,
		},
		CodecHierarchy: map[string]int{
			"MPEG":    1,
			"XVID":    2,
			"DIVX":    2,
			"H.264":   3,
			"AVC":     3,
			"H.265":   4,
			"HEVC":    4,
			"AV1":     5,
			"VP9":     4,
		},
		MinQuality: domains.Quality{
			Resolution: "720p",
			Source:     "HDTV",
			Codec:      "H.264",
		},
		MaxQuality: domains.Quality{
			Resolution: "4K",
			Source:     "BluRay",
			Codec:      "H.265",
		},
		PreferredQualities: []domains.Quality{
			{Resolution: "1080p", Source: "BluRay", Codec: "H.264"},
			{Resolution: "1080p", Source: "WEB-DL", Codec: "H.264"},
			{Resolution: "720p", Source: "BluRay", Codec: "H.264"},
			{Resolution: "720p", Source: "WEB-DL", Codec: "H.264"},
		},
	}
}

// ParseQuality 解析质量字符串
func (p *QualityPolicy) ParseQuality(qualityStr string) domains.Quality {
	parts := strings.Split(qualityStr, ".")
	quality := domains.Quality{}
	
	for _, part := range parts {
		part = strings.ToUpper(strings.TrimSpace(part))
		
		// 检查分辨率
		if p.QualityHierarchy[part] > 0 {
			quality.Resolution = part
			continue
		}
		
		// 检查来源
		if p.SourceHierarchy[part] > 0 {
			quality.Source = part
			continue
		}
		
		// 检查编码
		if p.CodecHierarchy[part] > 0 {
			quality.Codec = part
			continue
		}
		
		// 特殊处理
		switch part {
		case "HDCAM":
			quality.Resolution = "720p"
			quality.Source = "HDCAM"
		case "HDTS":
			quality.Resolution = "720p"
			quality.Source = "HDTS"
		case "BRRIP":
			quality.Resolution = "720p"
			quality.Source = "BDRip"
		case "BDRIP":
			quality.Resolution = "1080p"
			quality.Source = "BDRip"
		case "WEBRIP":
			quality.Resolution = "1080p"
			quality.Source = "WEBRip"
		case "WEB":
			quality.Resolution = "1080p"
			quality.Source = "WEB-DL"
		case "UHD":
			quality.Resolution = "4K"
			quality.Source = "UHD"
		}
	}
	
	return quality
}

// CompareQualities 比较两个质量
func (p *QualityPolicy) CompareQualities(q1, q2 domains.Quality) int {
	// 比较分辨率
	res1Score := p.QualityHierarchy[q1.Resolution]
	res2Score := p.QualityHierarchy[q2.Resolution]
	
	if res1Score > res2Score {
		return 1
	} else if res1Score < res2Score {
		return -1
	}
	
	// 分辨率相同，比较来源
	src1Score := p.SourceHierarchy[q1.Source]
	src2Score := p.SourceHierarchy[q2.Source]
	
	if src1Score > src2Score {
		return 1
	} else if src1Score < src2Score {
		return -1
	}
	
	// 来源相同，比较编码
	codec1Score := p.CodecHierarchy[q1.Codec]
	codec2Score := p.CodecHierarchy[q2.Codec]
	
	if codec1Score > codec2Score {
		return 1
	} else if codec1Score < codec2Score {
		return -1
	}
	
	return 0 // 相等
}

// IsBetterThan 检查q1是否优于q2
func (p *QualityPolicy) IsBetterThan(q1, q2 domains.Quality) bool {
	return p.CompareQualities(q1, q2) > 0
}

// IsAcceptable 检查质量是否可接受
func (p *QualityPolicy) IsAcceptable(quality domains.Quality) bool {
	// 检查是否高于最低质量
	if p.CompareQualities(quality, p.MinQuality) < 0 {
		return false
	}
	
	// 检查是否低于最高质量（可选）
	// return p.CompareQualities(quality, p.MaxQuality) <= 0
	
	return true
}

// GetBestQuality 从多个质量中选择最佳质量
func (p *QualityPolicy) GetBestQuality(qualities []domains.Quality) domains.Quality {
	if len(qualities) == 0 {
		return domains.Quality{}
	}
	
	best := qualities[0]
	for _, quality := range qualities[1:] {
		if p.IsBetterThan(quality, best) {
			best = quality
		}
	}
	
	return best
}

// FilterQualities 过滤质量列表
func (p *QualityPolicy) FilterQualities(qualities []domains.Quality, minQuality, maxQuality *domains.Quality) []domains.Quality {
	var filtered []domains.Quality
	
	for _, quality := range qualities {
		// 检查最低质量要求
		if minQuality != nil && p.CompareQualities(quality, *minQuality) < 0 {
			continue
		}
		
		// 检查最高质量限制
		if maxQuality != nil && p.CompareQualities(quality, *maxQuality) > 0 {
			continue
		}
		
		filtered = append(filtered, quality)
	}
	
	return filtered
}

// GetQualityScore 获取质量分数
func (p *QualityPolicy) GetQualityScore(quality domains.Quality) float64 {
	resScore := float64(p.QualityHierarchy[quality.Resolution])
	srcScore := float64(p.SourceHierarchy[quality.Source]) * 0.8
	codecScore := float64(p.CodecHierarchy[quality.Codec]) * 0.6
	
	return resScore + srcScore + codecScore
}

// GetQualityDescription 获取质量描述
func (p *QualityPolicy) GetQualityDescription(quality domains.Quality) string {
	if quality.Resolution == "" {
		return "未知质量"
	}
	
	description := quality.Resolution
	
	if quality.Source != "" {
		description += " " + quality.Source
	}
	
	if quality.Codec != "" {
		description += " " + quality.Codec
	}
	
	// 添加质量评级
	score := p.GetQualityScore(quality)
	var rating string
	switch {
	case score >= 15:
		rating = " (极佳)"
	case score >= 12:
		rating = " (优秀)"
	case score >= 9:
		rating = " (良好)"
	case score >= 6:
		rating = " (一般)"
	default:
		rating = " (较差)"
	}
	
	return description + rating
}

// SuggestQuality 根据用户偏好建议质量
func (p *QualityPolicy) SuggestQuality(userPreferences map[string]interface{}) domains.Quality {
	// 从用户偏好中获取质量要求
	if qualityStr, ok := userPreferences["preferred_quality"].(string); ok {
		return p.ParseQuality(qualityStr)
	}
	
	// 根据网络条件建议质量
	if networkSpeed, ok := userPreferences["network_speed"].(float64); ok {
		if networkSpeed < 10 { // Mbps
			return domains.Quality{Resolution: "720p", Source: "WEB-DL", Codec: "H.264"}
		} else if networkSpeed < 50 {
			return domains.Quality{Resolution: "1080p", Source: "WEB-DL", Codec: "H.264"}
		} else {
			return domains.Quality{Resolution: "4K", Source: "WEB-DL", Codec: "H.265"}
		}
	}
	
	// 根据存储空间建议质量
	if storageSpace, ok := userPreferences["storage_space"].(float64); ok { // GB
		if storageSpace < 500 {
			return domains.Quality{Resolution: "720p", Source: "WEB-DL", Codec: "H.264"}
		} else if storageSpace < 2000 {
			return domains.Quality{Resolution: "1080p", Source: "WEB-DL", Codec: "H.264"}
		} else {
			return domains.Quality{Resolution: "4K", Source: "BluRay", Codec: "H.265"}
		}
	}
	
	// 返回默认推荐质量
	return p.PreferredQualities[0]
}

// ValidateQualityRequirement 验证质量要求
func (p *QualityPolicy) ValidateQualityRequirement(requirement string) error {
	quality := p.ParseQuality(requirement)
	
	if quality.Resolution == "" {
		return fmt.Errorf("invalid quality requirement: %s", requirement)
	}
	
	if !p.IsAcceptable(quality) {
		return fmt.Errorf("quality %s is below minimum acceptable quality", requirement)
	}
	
	return nil
}