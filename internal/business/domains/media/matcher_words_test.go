package media

import (
	"testing"
)

func TestWordsMatcher_Prepare(t *testing.T) {
	// 创建简单的WordsMatcher实例
	matcher := NewSimpleWordsMatcher()

	tests := []struct {
		name            string
		title           string
		customWords     []string
		expectedTitle   string
		expectedApplied []string
	}{
		{
			name:            "基本替换词",
			title:           "Test Movie 2023",
			customWords:     []string{"Movie => Film"},
			expectedTitle:   "Test Film 2023",
			expectedApplied: []string{"Movie => Film"},
		},
		{
			name:            "屏蔽词",
			title:           "Test Movie 2023",
			customWords:     []string{"Movie"},
			expectedTitle:   "Test  2023",
			expectedApplied: []string{"Movie"},
		},
		{
			name:            "空标题",
			title:           "",
			customWords:     []string{"Movie => Film"},
			expectedTitle:   "",
			expectedApplied: nil,
		},
		{
			name:            "空识别词列表",
			title:           "Test Movie 2023",
			customWords:     []string{},
			expectedTitle:   "Test Movie 2023",
			expectedApplied: nil,
		},
		{
			name:            "注释识别词",
			title:           "Test Movie 2023",
			customWords:     []string{"# This is a comment", "Movie => Film"},
			expectedTitle:   "Test Film 2023",
			expectedApplied: []string{"Movie => Film"},
		},
		{
			name:            "多个识别词",
			title:           "Test Movie 2023",
			customWords:     []string{"Test => Demo", "Movie => Film"},
			expectedTitle:   "Demo Film 2023",
			expectedApplied: []string{"Test => Demo", "Movie => Film"},
		},
		{
			name:            "复杂替换规则",
			title:           "Test Series S01E01",
			customWords:     []string{"Series => Show && S01E <> E >> EP+1"},
			expectedTitle:   "Test Show S01E02",
			expectedApplied: []string{"Series => Show && S01E <> E >> EP+1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultTitle, resultApplied := matcher.Prepare(tt.title, tt.customWords)
			if resultTitle != tt.expectedTitle {
				t.Errorf("Prepare() title = %q, want %q", resultTitle, tt.expectedTitle)
			}
			if len(resultApplied) != len(tt.expectedApplied) {
				t.Errorf("Prepare() applied words length = %d, want %d", len(resultApplied), len(tt.expectedApplied))
			} else {
				for i, applied := range resultApplied {
					if applied != tt.expectedApplied[i] {
						t.Errorf("Prepare() applied word %d = %q, want %q", i, applied, tt.expectedApplied[i])
					}
				}
			}
		})
	}
}

func TestWordsMatcher_ReplaceRegex(t *testing.T) {
	// 创建简单的WordsMatcher实例
	matcher := NewSimpleWordsMatcher()

	tests := []struct {
		name          string
		title         string
		replaced      string
		replace       string
		expectedTitle string
		expectedState bool
	}{
		{
			name:          "基本替换",
			title:         "Test Movie 2023",
			replaced:      "Movie",
			replace:       "Film",
			expectedTitle: "Test Film 2023",
			expectedState: true,
		},
		{
			name:          "正则替换",
			title:         "Test Movie 2023",
			replaced:      `M\w+`,
			replace:       "Film",
			expectedTitle: "Test Film 2023",
			expectedState: true,
		},
		{
			name:          "不存在的替换词",
			title:         "Test Movie 2023",
			replaced:      "Nonexistent",
			replace:       "Film",
			expectedTitle: "Test Movie 2023",
			expectedState: false,
		},
		{
			title:         "Test Movie 2023",
			replaced:      "",
			replace:       "Film",
			expectedTitle: "Test Movie 2023",
			expectedState: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultTitle, _, resultState := matcher.replaceRegex(tt.title, tt.replaced, tt.replace)
			if resultTitle != tt.expectedTitle {
				t.Errorf("replaceRegex() title = %q, want %q", resultTitle, tt.expectedTitle)
			}
			if resultState != tt.expectedState {
				t.Errorf("replaceRegex() state = %v, want %v", resultState, tt.expectedState)
			}
		})
	}
}

func TestWordsMatcher_EpisodeOffset(t *testing.T) {
	// 创建简单的WordsMatcher实例
	matcher := NewSimpleWordsMatcher()

	tests := []struct {
		name          string
		title         string
		front         string
		back          string
		offset        string
		expectedTitle string
		expectedState bool
	}{
		{
			name:          "基本集偏移",
			title:         "Test S01E01",
			front:         "S01E",
			back:          "",
			offset:        "EP+1",
			expectedTitle: "Test S01E02",
			expectedState: true,
		},
		{
			name:          "中文数字集偏移",
			title:         "Test 第一季第一集",
			front:         "第一季第",
			back:          "集",
			offset:        "EP+1",
			expectedTitle: "Test 第一季第二集",
			expectedState: true,
		},
		{
			name:          "前导零集偏移",
			title:         "Test S01E001",
			front:         "S01E",
			back:          "",
			offset:        "EP+1",
			expectedTitle: "Test S01E002",
			expectedState: true,
		},
		{
			name:          "不存在的集",
			title:         "Test Movie",
			front:         "S01E",
			back:          "",
			offset:        "EP+1",
			expectedTitle: "Test Movie",
			expectedState: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultTitle, _, resultState := matcher.episodeOffset(tt.title, tt.front, tt.back, tt.offset)
			if resultTitle != tt.expectedTitle {
				t.Errorf("episodeOffset() title = %q, want %q", resultTitle, tt.expectedTitle)
			}
			if resultState != tt.expectedState {
				t.Errorf("episodeOffset() state = %v, want %v", resultState, tt.expectedState)
			}
		})
	}
}

func TestWordsMatcher_CalculateOffset(t *testing.T) {
	// 创建简单的WordsMatcher实例
	matcher := NewSimpleWordsMatcher()

	tests := []struct {
		name          string
		expr          string
		expected      int
		expectedError bool
	}{
		{
			name:          "加法运算",
			expr:          "1+1",
			expected:      2,
			expectedError: false,
		},
		{
			name:          "减法运算",
			expr:          "5-2",
			expected:      3,
			expectedError: false,
		},
		{
			name:          "乘法运算",
			expr:          "3*4",
			expected:      12,
			expectedError: false,
		},
		{
			name:          "除法运算",
			expr:          "10/2",
			expected:      5,
			expectedError: false,
		},
		{
			name:          "整数直接返回",
			expr:          "42",
			expected:      42,
			expectedError: false,
		},
		{
			name:          "带空格的表达式",
			expr:          "10 + 20",
			expected:      30,
			expectedError: false,
		},
		{
			name:          "除以零",
			expr:          "10/0",
			expected:      0,
			expectedError: true,
		},
		{
			name:          "无效表达式",
			expr:          "invalid",
			expected:      0,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := matcher.calculateOffset(tt.expr)
			if tt.expectedError {
				if err == nil {
					t.Errorf("calculateOffset() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("calculateOffset() unexpected error: %v", err)
				} else if result != tt.expected {
					t.Errorf("calculateOffset() = %d, want %d", result, tt.expected)
				}
			}
		})
	}
}

func TestWordsMatcher_IsChineseNumber(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		expected bool
	}{
		{
			name:     "中文数字",
			s:        "一二三",
			expected: true,
		},
		{
			name:     "阿拉伯数字",
			s:        "123",
			expected: false,
		},
		{
			name:     "混合数字",
			s:        "一2三",
			expected: false,
		},
		{
			name:     "空字符串",
			s:        "",
			expected: false,
		},
		{
			name:     "单个中文数字",
			s:        "一",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isChineseNumber(tt.s)
			if result != tt.expected {
				t.Errorf("isChineseNumber(%q) = %v, want %v", tt.s, result, tt.expected)
			}
		})
	}
}
