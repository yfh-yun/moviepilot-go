package media

import (
	"strings"
)

// IdentifierConfigProvider 识别词配置提供者接口
type IdentifierConfigProvider interface {
	GetCustomIdentifiers() ([]string, error)
}

// WordsMatcher 识别词匹配器
type WordsMatcher struct {
	provider IdentifierConfigProvider
}

// NewWordsMatcher 创建新的WordsMatcher实例
func NewWordsMatcher(provider IdentifierConfigProvider) *WordsMatcher {
	return &WordsMatcher{
		provider: provider,
	}
}

// Prepare 预处理标题，应用自定义识别词
func (wm *WordsMatcher) Prepare(title string, customWords []string) (string, []string) {
	if title == "" {
		return title, nil
	}

	// 合并自定义识别词
	allWords := customWords
	if wm.provider != nil {
		configWords, err := wm.provider.GetCustomIdentifiers()
		if err == nil {
			allWords = append(allWords, configWords...)
		}
	}

	// 应用识别词
	var appliedWords []string
	processedTitle := title

	for _, word := range allWords {
		word = strings.TrimSpace(word)
		if word == "" {
			continue
		}

		// 简单替换规则：A => B
		if strings.Contains(word, "=>") {
			parts := strings.SplitN(word, "=>", 2)
			if len(parts) == 2 {
				oldStr := strings.TrimSpace(parts[0])
				newStr := strings.TrimSpace(parts[1])
				if oldStr != "" && strings.Contains(processedTitle, oldStr) {
					processedTitle = strings.ReplaceAll(processedTitle, oldStr, newStr)
					appliedWords = append(appliedWords, word)
				}
			}
		} else if strings.HasPrefix(word, "-") {
			// 屏蔽词规则：-A
			blockWord := strings.TrimPrefix(word, "-")
			if blockWord != "" && strings.Contains(processedTitle, blockWord) {
				processedTitle = strings.ReplaceAll(processedTitle, blockWord, "")
				appliedWords = append(appliedWords, word)
			}
		}
	}

	// 清理多余空格
	processedTitle = strings.TrimSpace(processedTitle)
	processedTitle = strings.Join(strings.Fields(processedTitle), " ")

	return processedTitle, appliedWords
}

// SimpleWordsMatcher 简单的WordsMatcher实现，不依赖外部配置
type SimpleWordsMatcher struct{}

// NewSimpleWordsMatcher 创建简单的WordsMatcher实例
func NewSimpleWordsMatcher() *WordsMatcher {
	return &WordsMatcher{
		provider: &SimpleWordsMatcher{},
	}
}

// GetCustomIdentifiers 获取自定义识别词
func (swm *SimpleWordsMatcher) GetCustomIdentifiers() ([]string, error) {
	// 简单实现，返回空列表
	return []string{}, nil
}
