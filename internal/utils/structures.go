package utils

// DictUtils 字典工具�?type DictUtils struct{}

// NewDictUtils 创建新的字典工具类实�?func NewDictUtils() *DictUtils {
	return &DictUtils{}
}

// FilterKeysToSubset 过滤 source 字典，使其键成为 reference 字典键的子集
// source: 要被过滤的字�?// reference: 参考字典，定义允许的键
// 返回�? 过滤后的字典，只包含�?reference 中存在的�?func (d *DictUtils) FilterKeysToSubset(source map[interface{}]interface{}, reference map[interface{}]interface{}) map[interface{}]interface{} {
	// 检查输入参数是否为有效的字�?	if source == nil || reference == nil {
		return make(map[interface{}]interface{})
	}

	// 创建结果字典
	result := make(map[interface{}]interface{})

	// 遍历source字典，只保留reference中存在的�?	for key, value := range source {
		if _, exists := reference[key]; exists {
			result[key] = value
		}
	}

	return result
}

// IsKeysSubset 判断 source 字典的键是否�?reference 字典键的子集
// source: 要检查的字典
// reference: 参考字�?// 返回�? 如果 source 的键�?reference 的键子集，则返回 true，否则返�?false
func (d *DictUtils) IsKeysSubset(source map[interface{}]interface{}, reference map[interface{}]interface{}) bool {
	// 检查输入参数是否为有效的字�?	if source == nil || reference == nil {
		return false
	}

	// 检查source中的每个键是否都在reference中存�?	for key := range source {
		if _, exists := reference[key]; !exists {
			return false
		}
	}

	return true
}

// ListUtils 列表工具�?type ListUtils struct{}

// NewListUtils 创建新的列表工具类实�?func NewListUtils() *ListUtils {
	return &ListUtils{}
}

// Flatten 将嵌套的列表展平成单个列�?// nestedList: 嵌套的列�?// 返回�? 展平后的列表
func (l *ListUtils) Flatten(nestedList []interface{}) []interface{} {
	// 检查输入参数是否为有效的列�?	if nestedList == nil {
		return []interface{}{}
	}

	// 检查是否嵌套，若不嵌套直接返回
	isNested := false
	for _, item := range nestedList {
		if _, ok := item.([]interface{}); ok {
			isNested = true
			break
		}
	}

	// 如果不嵌套，直接返回
	if !isNested {
		return nestedList
	}

	// 展平嵌套列表
	result := make([]interface{}, 0)
	for _, item := range nestedList {
		// 检查是否为子列�?		if subList, ok := item.([]interface{}); ok {
			// 将子列表中的元素添加到结果中
			result = append(result, subList...)
		}
	}

	return result
}

// SetUtils 集合工具�?type SetUtils struct{}

// NewSetUtils 创建新的集合工具类实�?func NewSetUtils() *SetUtils {
	return &SetUtils{}
}

// Flatten 将嵌套的集合展开为单个集�?// nestedSets: 嵌套的集�?// 返回�? 展开的集�?func (s *SetUtils) Flatten(nestedSets map[interface{}]bool) map[interface{}]bool {
	// 检查输入参数是否为有效的集�?	if nestedSets == nil {
		return make(map[interface{}]bool)
	}

	// 检查是否嵌套，若不嵌套直接返回
	isNested := false
	for item := range nestedSets {
		if _, ok := item.(map[interface{}]bool); ok {
			isNested = true
			break
		}
	}

	// 如果不嵌套，直接返回
	if !isNested {
		// 创建一个新的map作为返回值，避免直接返回原集�?		result := make(map[interface{}]bool)
		for key, value := range nestedSets {
			result[key] = value
		}
		return result
	}

	// 展开嵌套集合
	result := make(map[interface{}]bool)
	for item := range nestedSets {
		// 检查是否为子集�?		if subSet, ok := item.(map[interface{}]bool); ok {
			// 将子集合中的元素添加到结果中
			for key, value := range subSet {
				result[key] = value
			}
		}
	}

	return result
}
