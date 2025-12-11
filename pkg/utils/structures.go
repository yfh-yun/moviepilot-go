package utils

// FilterKeysToSubset 过滤 source map，使其键成为 reference map 键的子集
// 对应 Python DictUtils.filter_keys_to_subset
func FilterKeysToSubset[K comparable, V any](source, reference map[K]V) map[K]V {
	result := make(map[K]V)
	if source == nil || reference == nil {
		return result
	}
	for k, v := range source {
		if _, ok := reference[k]; ok {
			result[k] = v
		}
	}
	return result
}

// IsKeysSubset 判断 source 的键是否为 reference 键的子集
// 对应 Python DictUtils.is_keys_subset
func IsKeysSubset[K comparable, V any](source, reference map[K]V) bool {
	if source == nil || reference == nil {
		return false
	}
	for k := range source {
		if _, ok := reference[k]; !ok {
			return false
		}
	}
	return true
}

// FlattenList 将嵌套切片展平成一层，对应 Python ListUtils.flatten
// 支持任意深度嵌套，与原实现保持一致
func FlattenList(nested any) []any {
	result := make([]any, 0)
	
	switch v := nested.(type) {
	case []any:
		for _, item := range v {
			// 递归展平每个元素
			result = append(result, FlattenList(item)...)
		}
	default:
		// 非切片类型直接添加
		result = append(result, v)
	}
	
	return result
}

// FlattenSet 将嵌套集合展开为一个集合，对应 Python SetUtils.flatten
// 使用 map[T]struct{} 模拟集合，与原实现保持一致（仅展开一层）
func FlattenSet[T comparable](nested []map[T]struct{}) map[T]struct{} {
	result := make(map[T]struct{})
	if nested == nil {
		return result
	}

	// 检查是否嵌套，若不嵌套直接返回
	isNested := false
	for _, subset := range nested {
		if len(subset) > 0 {
			isNested = true
			break
		}
	}

	if !isNested {
		return result
	}

	for _, subset := range nested {
		for item := range subset {
			result[item] = struct{}{}
		}
	}

	return result
}
