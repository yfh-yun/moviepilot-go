package media

import (
	"testing"
)

func TestStreamingPlatforms_GetStreamingPlatformName(t *testing.T) {
	sp := NewStreamingPlatforms()

	tests := []struct {
		name           string
		platformCode   string
		expectedName   string
	}{
		{
			name:         "获取Amazon标准名称",
			platformCode: "AMZN",
			expectedName: "Amazon",
		},
		{
			name:         "获取Netflix标准名称",
			platformCode: "NF",
			expectedName: "Netflix",
		},
		{
			name:         "获取Disney+标准名称",
			platformCode: "DSNP",
			expectedName: "Disney+",
		},
		{
			name:         "获取Apple TV+标准名称",
			platformCode: "ATVP",
			expectedName: "Apple TV+",
		},
		{
			name:         "空字符串输入",
			platformCode: "",
			expectedName: "",
		},
		{
			name:         "不存在的平台代码",
			platformCode: "INVALID",
			expectedName: "",
		},
		{
			name:         "小写平台代码",
			platformCode: "netflix",
			expectedName: "Netflix",
		},
		{
			name:         "获取Max标准名称",
			platformCode: "HMAX",
			expectedName: "Max",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sp.GetStreamingPlatformName(tt.platformCode)
			if result != tt.expectedName {
				t.Errorf("GetStreamingPlatformName(%q) = %q, want %q", tt.platformCode, result, tt.expectedName)
			}
		})
	}
}

func TestStreamingPlatforms_IsStreamingPlatform(t *testing.T) {
	sp := NewStreamingPlatforms()

	tests := []struct {
		name     string
		platform string
		expected bool
	}{
		{
			name:     "验证AMZN是流媒体平台",
			platform: "AMZN",
			expected: true,
		},
		{
			name:     "验证Netflix是流媒体平台",
			platform: "Netflix",
			expected: true,
		},
		{
			name:     "验证小写netflix是流媒体平台",
			platform: "netflix",
			expected: true,
		},
		{
			name:     "验证Disney+是流媒体平台",
			platform: "Disney+",
			expected: true,
		},
		{
			name:     "空字符串不是流媒体平台",
			platform: "",
			expected: false,
		},
		{
			name:     "不存在的平台不是流媒体平台",
			platform: "INVALID",
			expected: false,
		},
		{
			name:     "验证HULU是流媒体平台",
			platform: "HULU",
			expected: true,
		},
		{
			name:     "验证Prime不是流媒体平台（根据Python版本定义）",
			platform: "Prime",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sp.IsStreamingPlatform(tt.platform)
			if result != tt.expected {
				t.Errorf("IsStreamingPlatform(%q) = %v, want %v", tt.platform, result, tt.expected)
			}
		})
	}
}

func TestStreamingPlatforms_Singleton(t *testing.T) {
	// 获取第一个实例
	instance1 := NewStreamingPlatforms()
	// 获取第二个实例
	instance2 := NewStreamingPlatforms()

	// 验证两个实例是同一个对象
	if instance1 != instance2 {
		t.Error("NewStreamingPlatforms() 没有返回单例实例")
	}

	// 验证实例的查找表不为空
	if len(instance1.lookup) == 0 {
		t.Error("StreamingPlatforms实例的查找表为空")
	}
}

func TestStreamingPlatforms_BuildCache(t *testing.T) {
	sp := NewStreamingPlatforms()
	// 记录初始缓存大小
	initialSize := len(sp.lookup)

	// 重新构建缓存
	sp.buildCache()
	// 验证缓存大小不变
	if len(sp.lookup) != initialSize {
		t.Errorf("buildCache() 改变了缓存大小: 初始 %d, 重建后 %d", initialSize, len(sp.lookup))
	}

	// 验证缓存内容不为空
	if len(sp.lookup) == 0 {
		t.Error("buildCache() 生成的缓存为空")
	}
}
