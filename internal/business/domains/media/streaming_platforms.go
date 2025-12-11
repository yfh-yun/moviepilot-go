package media

import (
	"strings"
)

// StreamingPlatforms 流媒体平台映射
type StreamingPlatforms struct {
	lookup map[string]string // UPPER(alias) -> canonical name
}

// streamingPlatformData 流媒体平台数据
var streamingPlatformData = map[string][]string{
	"Netflix":   {"NF", "Netflix", "netflix"},
	"Disney+":   {"DSNP", "Disney+", "DisneyPlus", "disney+"},
	"HBO Max":   {"HBOMAX", "HBO", "HBO Max", "hbomax"},
	"Amazon":    {"AMZN", "Amazon", "Prime", "Prime Video", "amazon"},
	"Apple TV+": {"APPLE", "Apple", "Apple TV+", "AppleTV+"},
	"Hulu":      {"HULU", "Hulu", "hulu"},
	"Tencent":   {"TX", "Tencent", "腾讯", "腾讯视频"},
	"iQiyi":     {"IQIYI", "iQiyi", "爱奇艺"},
	"Youku":     {"YK", "Youku", "优酷"},
	"Bilibili":  {"BILIBILI", "B站", "哔哩哔哩"},
}

// NewStreamingPlatforms 创建新的StreamingPlatforms实例
func NewStreamingPlatforms() *StreamingPlatforms {
	sp := &StreamingPlatforms{
		lookup: make(map[string]string),
	}

	// 构建查找表
	for name, aliases := range streamingPlatformData {
		for _, alias := range aliases {
			sp.lookup[strings.ToUpper(alias)] = name
		}
	}

	return sp
}

// IsStreamingPlatform 判断是否为流媒体平台
func (s *StreamingPlatforms) IsStreamingPlatform(name string) bool {
	if name == "" {
		return false
	}
	_, exists := s.lookup[strings.ToUpper(name)]
	return exists
}

// GetName 获取流媒体平台名称
func (s *StreamingPlatforms) GetName(code string) string {
	if code == "" {
		return ""
	}
	if name, exists := s.lookup[strings.ToUpper(code)]; exists {
		return name
	}
	return ""
}
