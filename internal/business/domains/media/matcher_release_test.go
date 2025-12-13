package media

import (
	"testing"

	"moviepilot-go/pkg/cache"
)

// MockReleaseGroupConfigProvider 模拟ReleaseGroupConfigProvider接口
type MockReleaseGroupConfigProvider struct {
	customGroups []string
	err          error
}

// GetCustomReleaseGroups 实现ReleaseGroupConfigProvider接口
func (m *MockReleaseGroupConfigProvider) GetCustomReleaseGroups() ([]string, error) {
	return m.customGroups, m.err
}

func TestReleaseGroupsMatcher_Match(t *testing.T) {
	// 创建一个内存缓存实例
	cacheInstance := cache.Cache("memory", 1000, 24*60*60)
	// 创建一个模拟的配置提供者
	mockProvider := &MockReleaseGroupConfigProvider{
		customGroups: []string{`CustomGroup`},
		err:          nil,
	}
	matcher := NewReleaseGroupsMatcher(mockProvider, cacheInstance)

	tests := []struct {
		name     string
		title    string
		groups   string
		expected string
	}{
		{
			name:     "匹配CHDBits组",
			title:    "Test.Movie.2023.1080p-CHDBits.WEB-DL.H264",
			groups:   "",
			expected: "CHDBits",
		},
		{
			name:     "匹配MTeam组",
			title:    "Test.Movie.2023.1080p@MTeam.WEB-DL.H264",
			groups:   "",
			expected: "MTeam",
		},
		{
			name:     "匹配TTG组",
			title:    "Test.Movie.2023.1080p[TTG].WEB-DL.H264",
			groups:   "",
			expected: "TTG",
		},
		{
			name:     "匹配LoliHouse动漫组",
			title:    "Test.Anime.2023.1080p-LoliHouse.WEB-DL.H264",
			groups:   "",
			expected: "LoliHouse",
		},
		{
			name:     "匹配中文组",
			title:    "Test.Anime.2023.1080p[喵萌奶茶屋].WEB-DL.H264",
			groups:   "",
			expected: "喵萌奶茶屋",
		},
		{
			name:     "空标题",
			title:    "",
			groups:   "",
			expected: "",
		},
		{
			name:     "无匹配组",
			title:    "Test.Movie.2023.1080p.UnknownGroup.WEB-DL.H264",
			groups:   "",
			expected: "",
		},
		{
			name:     "使用自定义组",
			title:    "Test.Movie.2023.1080p-CustomGroup.WEB-DL.H264",
			groups:   "",
			expected: "CustomGroup",
		},
		{
			name:     "直接指定groups参数",
			title:    "Test.Movie.2023.1080p-MyGroup.WEB-DL.H264",
			groups:   "MyGroup",
			expected: "MyGroup",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matcher.Match(tt.title, tt.groups)
			if result != tt.expected {
				t.Errorf("Match() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestReleaseGroupsMatcher_Cache(t *testing.T) {
	// 创建一个内存缓存实例
	cacheInstance := cache.Cache("memory", 1000, 24*60*60)
	// 创建一个模拟的配置提供者
	mockProvider := &MockReleaseGroupConfigProvider{
		customGroups: []string{},
		err:          nil,
	}
	matcher := NewReleaseGroupsMatcher(mockProvider, cacheInstance)

	title := "Test.Movie.2023.1080p-CHDBits.WEB-DL.H264"
	groups := ""

	// 第一次调用，应该计算结果并缓存
	result1 := matcher.Match(title, groups)
	if result1 != "CHDBits" {
		t.Errorf("First Match() = %v, want CHDBits", result1)
	}

	// 第二次调用，应该从缓存获取结果
	result2 := matcher.Match(title, groups)
	if result2 != "CHDBits" {
		t.Errorf("Second Match() = %v, want CHDBits", result2)
	}

	// 验证结果一致
	if result1 != result2 {
		t.Errorf("First result %v != Second result %v", result1, result2)
	}
}

func TestReleaseGroupsMatcher_RegexPatterns(t *testing.T) {
	// 创建一个内存缓存实例
	cacheInstance := cache.Cache("memory", 1000, 24*60*60)
	// 创建一个模拟的配置提供者
	mockProvider := &MockReleaseGroupConfigProvider{
		customGroups: []string{},
		err:          nil,
	}
	matcher := NewReleaseGroupsMatcher(mockProvider, cacheInstance)

	// 测试一些特定的正则表达式模式
	tests := []struct {
		title    string
		expected string
	}{
		{"Test.2023-FFWEB.WEB-DL.H264", "FFWEB"},
		{"Test.2023@BeiTai.WEB-DL.H264", "BeiTai"},
		{"Test.2023[BtsCHOOL].WEB-DL.H264", "BtsCHOOL"},
		{"Test.2023-CHDBits.WEB-DL.H264", "CHDBits"},
		{"Test.2023-TLF.WEB-DL.H264", "TLF"},
		{"Test.2023-iNT-TLF.WEB-DL.H264", "iNT-TLF"},
		{"Test.2023@CMCT.WEB-DL.H264", "CMCT"},
		{"Test.2023-TJUPT.WEB-DL.H264", "TJUPT"},
		{"Test.2023-ANi.WEB-DL.H264", "ANi"},
		{"Test.2023-LoliHouse.WEB-DL.H264", "LoliHouse"},
		{"Test.2023-织梦字幕组.WEB-DL.H264", "织梦字幕组"},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			result := matcher.Match(tt.title, "")
			if result != tt.expected {
				t.Errorf("Match(%v) = %v, want %v", tt.title, result, tt.expected)
			}
		})
	}
}

func TestReleaseGroupsMatcher_CustomProviderError(t *testing.T) {
	// 创建一个内存缓存实例
	cacheInstance := cache.Cache("memory", 1000, 24*60*60)
	// 创建一个模拟的配置提供者，返回错误
	mockProvider := &MockReleaseGroupConfigProvider{
		customGroups: []string{},
		err:          &testError{"test error"},
	}
	matcher := NewReleaseGroupsMatcher(mockProvider, cacheInstance)

	// 测试当配置提供者返回错误时，应该只使用内置组
	title := "Test.Movie.2023.1080p-CHDBits.WEB-DL.H264"
	result := matcher.Match(title, "")
	if result != "CHDBits" {
		t.Errorf("Match() = %v, want CHDBits", result)
	}
}

// testError 简单的错误实现
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}
