package utils

import "regexp"

// Tokens 对应 Python app/utils/tokens.py 中的 Tokens 类
// 负责按特定分隔符将文本拆分为 token，并提供顺序访问接口。
type Tokens struct {
	text   string
	index  int
	tokens []string
}

// NewTokens 创建 Tokens 实例并加载文本
func NewTokens(text string) *Tokens {
	t := &Tokens{}
	t.LoadText(text)
	return t
}

// LoadText 按与 Python 版相同的分隔符拆分文本
func (t *Tokens) LoadText(text string) {
	// 使用与 Python re.split(r"\. |\s+|\(|\)|\[|]|-|【|】|/|～|;|&|\||#|_|「|」|~", text) 等价的正则
	re := regexp.MustCompile(`\.|\s+|\(|\)|\[|]|-|【|】|/|～|;|&|\||#|_|「|」|~`)
	parts := re.Split(text, -1)

	t.tokens = t.tokens[:0]
	for _, p := range parts {
		if p != "" {
			t.tokens = append(t.tokens, p)
		}
	}
	t.text = text
	t.index = 0
}

// Cur 返回当前 token，不前进索引
func (t *Tokens) Cur() *string {
	if t.index >= len(t.tokens) {
		return nil
	}
	tok := t.tokens[t.index]
	return &tok
}

// Next 返回当前 token 并前进索引
func (t *Tokens) Next() *string {
	tok := t.Cur()
	if tok != nil {
		t.index++
	}
	return tok
}

// Peek 返回下一个 token，但不前进索引
func (t *Tokens) Peek() *string {
	idx := t.index + 1
	if idx >= len(t.tokens) {
		return nil
	}
	tok := t.tokens[idx]
	return &tok
}

// All 返回全部 token 切片的拷贝
func (t *Tokens) All() []string {
	out := make([]string, len(t.tokens))
	copy(out, t.tokens)
	return out
}
