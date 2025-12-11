package utils

import (
	"testing"
)

// TestNewTokens 测试NewTokens函数
func TestNewTokens(t *testing.T) {
	text := "Hello World. This is a test-token【with】special/characters～"
	tokens := NewTokens(text)
	if tokens == nil {
		t.Errorf("Expected non-nil Tokens instance, got nil")
	}
	allTokens := tokens.All()
	if len(allTokens) == 0 {
		t.Errorf("Expected tokens, got empty slice")
	}
}

// TestLoadText 测试LoadText方法
func TestLoadText(t *testing.T) {
	text := "Hello.World (test) [123] - 【中文】/英文～"
	tokens := NewTokens("")
	tokens.LoadText(text)
	allTokens := tokens.All()

	// 验证分割结果
	expectedCount := 6 // Hello, World, test, 123, 中文, 英文
	if len(allTokens) != expectedCount {
		t.Errorf("Expected %d tokens, got %d: %v", expectedCount, len(allTokens), allTokens)
	}
}

// TestCur 测试Cur方法
func TestCur(t *testing.T) {
	text := "First Second Third"
	tokens := NewTokens(text)

	// 初始状态下，Cur应返回第一个token
	curToken := tokens.Cur()
	if curToken == nil || *curToken != "First" {
		t.Errorf("Expected 'First', got %v", curToken)
	}

	// 多次调用Cur应返回相同的token
	curToken2 := tokens.Cur()
	if curToken2 == nil || *curToken2 != "First" {
		t.Errorf("Expected 'First' again, got %v", curToken2)
	}
}

// TestNext 测试Next方法
func TestNext(t *testing.T) {
	text := "First Second Third"
	tokens := NewTokens(text)

	// 第一次调用Next应返回First，并前进索引
	token := tokens.Next()
	if token == nil || *token != "First" {
		t.Errorf("Expected 'First', got %v", token)
	}

	// 第二次调用Next应返回Second，并前进索引
	token = tokens.Next()
	if token == nil || *token != "Second" {
		t.Errorf("Expected 'Second', got %v", token)
	}

	// 第三次调用Next应返回Third，并前进索引
	token = tokens.Next()
	if token == nil || *token != "Third" {
		t.Errorf("Expected 'Third', got %v", token)
	}

	// 第四次调用Next应返回nil
	token = tokens.Next()
	if token != nil {
		t.Errorf("Expected nil for out-of-bounds, got %v", token)
	}
}

// TestPeek 测试Peek方法
func TestPeek(t *testing.T) {
	text := "First Second Third"
	tokens := NewTokens(text)

	// 初始状态下，Peek应返回Second
	peekToken := tokens.Peek()
	if peekToken == nil || *peekToken != "Second" {
		t.Errorf("Expected 'Second', got %v", peekToken)
	}

	// 调用Cur应返回First，因为Peek不前进索引
	curToken := tokens.Cur()
	if curToken == nil || *curToken != "First" {
		t.Errorf("Expected 'First', got %v", curToken)
	}

	// 调用Next前进索引，然后Peek应返回Third
	tokens.Next()
	peekToken = tokens.Peek()
	if peekToken == nil || *peekToken != "Third" {
		t.Errorf("Expected 'Third', got %v", peekToken)
	}

	// 当索引指向最后一个token时，Peek应返回nil
	tokens.Next()
	peekToken = tokens.Peek()
	if peekToken != nil {
		t.Errorf("Expected nil for out-of-bounds peek, got %v", peekToken)
	}
}

// TestAll 测试All方法
func TestAll(t *testing.T) {
	text := "A.B C_D (E) [F] - G【H】/I～J;K&L|M#N「O」~P"
	tokens := NewTokens(text)
	allTokens := tokens.All()

	// 验证返回的切片是拷贝，修改不影响原切片
	if len(allTokens) > 0 {
		allTokens[0] = "Modified"
		original := tokens.Cur()
		if original == nil || *original == "Modified" {
			t.Errorf("Expected original token to remain unchanged, got %v", original)
		}
	}

	// 验证分割结果
	expectedCount := 16 // A, B, C, D, E, F, G, H, I, J, K, L, M, N, O, P
	if len(allTokens) != expectedCount {
		t.Errorf("Expected %d tokens, got %d: %v", expectedCount, len(allTokens), allTokens)
	}
}

// TestEmptyText 测试空文本处理
func TestEmptyText(t *testing.T) {
	text := ""
	tokens := NewTokens(text)
	allTokens := tokens.All()
	if len(allTokens) != 0 {
		t.Errorf("Expected empty slice for empty text, got %v", allTokens)
	}

	curToken := tokens.Cur()
	if curToken != nil {
		t.Errorf("Expected nil for empty text Cur(), got %v", curToken)
	}

	nextToken := tokens.Next()
	if nextToken != nil {
		t.Errorf("Expected nil for empty text Next(), got %v", nextToken)
	}

	peekToken := tokens.Peek()
	if peekToken != nil {
		t.Errorf("Expected nil for empty text Peek(), got %v", peekToken)
	}
}

// TestMultipleDelimiters 测试连续分隔符处理
func TestMultipleDelimiters(t *testing.T) {
	text := "A..B   C--D\t\nE||F"
	tokens := NewTokens(text)
	allTokens := tokens.All()

	// 连续分隔符应被视为一个，预期结果：A, B, C, D, E, F
	expectedCount := 6
	if len(allTokens) != expectedCount {
		t.Errorf("Expected %d tokens for multiple delimiters, got %d: %v", expectedCount, len(allTokens), allTokens)
	}
}
