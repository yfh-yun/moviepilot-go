package utils

import (
	"regexp"
	"strings"
)

// Tokens 是一个用于处理文本标记的结构�?type Tokens struct {
	text   string   // 原始文本
	index  int      // 当前索引位置
	tokens []string // 分割后的标记数组
}

// NewTokens 创建一个新的Tokens实例
func NewTokens(text string) *Tokens {
	t := &Tokens{
		text:   text,
		index:  0,
		tokens: make([]string, 0),
	}
	t.loadText(text)
	return t
}

// loadText 将文本按指定分隔符分割成标记数组
func (t *Tokens) loadText(text string) {
	// 使用正则表达式分割文本，与Python版本保持一�?	re := regexp.MustCompile(`\.|\s+|\(|\)|\[|\]|-|【|】|/|～|;|&|\||#|_|「|」|~`)
	splittedText := re.Split(text, -1)
	
	for _, subText := range splittedText {
		if strings.TrimSpace(subText) != "" {
			t.tokens = append(t.tokens, subText)
		}
	}
}

// Cur 返回当前索引位置的标记，如果索引超出范围则返回空字符�?func (t *Tokens) Cur() string {
	if t.index >= len(t.tokens) {
		return ""
	}
	return t.tokens[t.index]
}

// GetNext 返回当前标记并将索引向前移动一�?func (t *Tokens) GetNext() string {
	token := t.Cur()
	if token != "" {
		t.index++
	}
	return token
}

// Peek 返回下一个标记但不移动索引位�?func (t *Tokens) Peek() string {
	index := t.index + 1
	if index >= len(t.tokens) {
		return ""
	}
	return t.tokens[index]
}

// Tokens 返回所有的标记数组
func (t *Tokens) Tokens() []string {
	return t.tokens
}
