package utils

import (
	"testing"
)

func TestFilterKeysToSubset(t *testing.T) {
	tests := []struct {
		name     string
		source   map[string]int
		reference map[string]int
		expected map[string]int
	}{
		{
			name:     "正常情况，过滤部分键",
			source:   map[string]int{"a": 1, "b": 2, "c": 3},
			reference: map[string]int{"a": 10, "c": 30},
			expected: map[string]int{"a": 1, "c": 3},
		},
		{
			name:     "源字典为空",
			source:   map[string]int{},
			reference: map[string]int{"a": 10},
			expected: map[string]int{},
		},
		{
			name:     "参考字典为空",
			source:   map[string]int{"a": 1, "b": 2},
			reference: map[string]int{},
			expected: map[string]int{},
		},
		{
			name:     "源字典和参考字典都为空",
			source:   map[string]int{},
			reference: map[string]int{},
			expected: map[string]int{},
		},
		{
			name:     "源字典为nil",
			source:   nil,
			reference: map[string]int{"a": 10},
			expected: map[string]int{},
		},
		{
			name:     "参考字典为nil",
			source:   map[string]int{"a": 1},
			reference: nil,
			expected: map[string]int{},
		},
		{
			name:     "所有键都匹配",
			source:   map[string]int{"a": 1, "b": 2},
			reference: map[string]int{"a": 10, "b": 20, "c": 30},
			expected: map[string]int{"a": 1, "b": 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterKeysToSubset(tt.source, tt.reference)
			if len(result) != len(tt.expected) {
				t.Errorf("长度不匹配：期望 %d, 得到 %d", len(tt.expected), len(result))
				return
			}

			for key, expectedVal := range tt.expected {
				if actualVal, exists := result[key]; !exists || actualVal != expectedVal {
					t.Errorf("键 %s 不匹配：期望 %d, 得到 %d", key, expectedVal, actualVal)
				}
			}
		})
	}
}

func TestIsKeysSubset(t *testing.T) {
	tests := []struct {
		name     string
		source   map[string]int
		reference map[string]int
		expected bool
	}{
		{
			name:     "源字典键是参考字典键的子集",
			source:   map[string]int{"a": 1, "c": 3},
			reference: map[string]int{"a": 10, "b": 20, "c": 30},
			expected: true,
		},
		{
			name:     "源字典键不是参考字典键的子集",
			source:   map[string]int{"a": 1, "d": 4},
			reference: map[string]int{"a": 10, "b": 20, "c": 30},
			expected: false,
		},
		{
			name:     "源字典为空",
			source:   map[string]int{},
			reference: map[string]int{"a": 10},
			expected: true,
		},
		{
			name:     "参考字典为空",
			source:   map[string]int{"a": 1},
			reference: map[string]int{},
			expected: false,
		},
		{
			name:     "源字典和参考字典都为空",
			source:   map[string]int{},
			reference: map[string]int{},
			expected: true,
		},
		{
			name:     "源字典为nil",
			source:   nil,
			reference: map[string]int{"a": 10},
			expected: false,
		},
		{
			name:     "参考字典为nil",
			source:   map[string]int{"a": 1},
			reference: nil,
			expected: false,
		},
		{
			name:     "键完全匹配",
			source:   map[string]int{"a": 1, "b": 2},
			reference: map[string]int{"a": 10, "b": 20},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsKeysSubset(tt.source, tt.reference)
			if result != tt.expected {
				t.Errorf("测试 %s 失败：期望 %t, 得到 %t", tt.name, tt.expected, result)
			}
		})
	}
}

func TestFlattenList(t *testing.T) {
	tests := []struct {
		name     string
		nested   any
		expected []any
	}{
		{
			name:     "单层切片",
			nested:   []any{1, 2, 3, 4, 5},
			expected: []any{1, 2, 3, 4, 5},
		},
		{
			name:     "嵌套切片",
			nested:   []any{[]any{1, 2}, []any{3, 4}, 5},
			expected: []any{1, 2, 3, 4, 5},
		},
		{
			name:     "深层嵌套切片",
			nested:   []any{[]any{[]any{1, 2}, []any{3}}, []any{4, 5}},
			expected: []any{1, 2, 3, 4, 5},
		},
		{
			name:     "空切片",
			nested:   []any{},
			expected: []any{},
		},
		{
			name:     "只有嵌套空切片",
			nested:   []any{[]any{}, []any{[]any{}}},
			expected: []any{},
		},
		{
			name:     "单个元素",
			nested:   1,
			expected: []any{1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FlattenList(tt.nested)
			if len(result) != len(tt.expected) {
				t.Errorf("长度不匹配：期望 %d, 得到 %d", len(tt.expected), len(result))
				return
			}

			for i, expectedVal := range tt.expected {
				if result[i] != expectedVal {
					t.Errorf("索引 %d 不匹配：期望 %v, 得到 %v", i, expectedVal, result[i])
				}
			}
		})
	}
}

func TestFlattenSet(t *testing.T) {
	// 创建集合辅助函数
	createIntSet := func(items ...int) map[int]struct{} {
		set := make(map[int]struct{})
		for _, item := range items {
			set[item] = struct{}{}
		}
		return set
	}

	// 检查集合是否相等
	areIntSetsEqual := func(set1, set2 map[int]struct{}) bool {
		if len(set1) != len(set2) {
			return false
		}

		for key := range set1 {
			if _, exists := set2[key]; !exists {
				return false
			}
		}

		return true
	}

	tests := []struct {
		name     string
		nested   []map[int]struct{}
		expected map[int]struct{}
	}{
		{
			name:     "嵌套集合",
			nested:   []map[int]struct{}{createIntSet(1, 2), createIntSet(3, 4), createIntSet(5)},
			expected: createIntSet(1, 2, 3, 4, 5),
		},
		{
			name:     "空切片",
			nested:   []map[int]struct{}{},
			expected: make(map[int]struct{}),
		},
		{
			name:     "nil切片",
			nested:   nil,
			expected: make(map[int]struct{}),
		},
		{
			name:     "只有嵌套空集合",
			nested:   []map[int]struct{}{make(map[int]struct{}), make(map[int]struct{})},
			expected: make(map[int]struct{}),
		},
		{
			name:     "包含重复元素的集合",
			nested:   []map[int]struct{}{createIntSet(1, 2), createIntSet(2, 3), createIntSet(3)}, 
			expected: createIntSet(1, 2, 3),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FlattenSet(tt.nested)
			if !areIntSetsEqual(result, tt.expected) {
				t.Errorf("集合不匹配：期望 %v, 得到 %v", tt.expected, result)
			}
		})
	}
}
