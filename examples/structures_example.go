package main

import (
	"fmt"

	"moviepilot-go/internal/utils"
)

func main() {
	fmt.Println("=== 数据结构工具示例 ===")

	// 测试字典工具
	fmt.Println("\n--- 字典工具测试 ---")
	testDictUtils()

	// 测试列表工具
	fmt.Println("\n--- 列表工具测试 ---")
	testListUtils()

	// 测试集合工具
	fmt.Println("\n--- 集合工具测试 ---")
	testSetUtils()
}

func testDictUtils() {
	dictUtils := utils.NewDictUtils()

	// 创建测试数据
	source := map[interface{}]interface{}{
		"name":    "张三",
		"age":     25,
		"city":    "北京",
		"country": "中国",
	}

	reference := map[interface{}]interface{}{
		"name": "李四",
		"age":  30,
		"email": "test@example.com",
	}

	// 测试 FilterKeysToSubset
	fmt.Println("原始source字典:")
	printDict(source)
	fmt.Println("参考reference字典:")
	printDict(reference)

	filtered := dictUtils.FilterKeysToSubset(source, reference)
	fmt.Println("过滤后的字典:")
	printDict(filtered)

	// 测试 IsKeysSubset
	isSubset := dictUtils.IsKeysSubset(source, reference)
	fmt.Printf("source的键是否为reference的键子集: %v\n", isSubset)

	// 创建一个键子集字典进行测试
	subsetDict := map[interface{}]interface{}{
		"name": "王五",
		"age":  35,
	}

	isSubset = dictUtils.IsKeysSubset(subsetDict, reference)
	fmt.Printf("subsetDict的键是否为reference的键子集: %v\n", isSubset)

	// 测试边界情况
	emptyResult := dictUtils.FilterKeysToSubset(nil, reference)
	fmt.Printf("nil字典过滤结果长度: %d\n", len(emptyResult))

	emptyResult = dictUtils.FilterKeysToSubset(source, nil)
	fmt.Printf("reference为nil的过滤结果长�? %d\n", len(emptyResult))

	boolResult := dictUtils.IsKeysSubset(nil, reference)
	fmt.Printf("nil字典是否为子�? %v\n", boolResult)
}

func testListUtils() {
	listUtils := utils.NewListUtils()

	// 创建嵌套列表
	nestedList := []interface{}{
		[]interface{}{"a", "b", "c"},
		[]interface{}{1, 2, 3},
		[]interface{}{"x", "y"},
	}

	fmt.Println("原始嵌套列表:")
	printList(nestedList)

	// 测试 Flatten
	flattened := listUtils.Flatten(nestedList)
	fmt.Println("展平后的列表:")
	printList(flattened)

	// 测试非嵌套列�?	nonNestedList := []interface{}{"a", "b", "c", 1, 2, 3}
	fmt.Println("\n非嵌套列�?")
	printList(nonNestedList)

	flattened = listUtils.Flatten(nonNestedList)
	fmt.Println("非嵌套列表展平后:")
	printList(flattened)

	// 测试边界情况
	emptyResult := listUtils.Flatten(nil)
	fmt.Printf("nil列表展平结果长度: %d\n", len(emptyResult))
}

func testSetUtils() {
	setUtils := utils.NewSetUtils()

	// 创建嵌套集合 (使用map[interface{}]bool来模拟set)
	nestedSets := map[interface{}]bool{
		// 子集�?
		map[interface{}]bool{
			"a": true,
			"b": true,
			"c": true,
		}: true,
		// 子集�?
		map[interface{}]bool{
			1: true,
			2: true,
			3: true,
		}: true,
		// 子集�?
		map[interface{}]bool{
			"x": true,
			"y": true,
		}: true,
	}

	fmt.Println("原始嵌套集合:")
	printSet(nestedSets)

	// 测试 Flatten
	flattened := setUtils.Flatten(nestedSets)
	fmt.Println("展开后的集合:")
	printSet(flattened)

	// 测试非嵌套集�?	nonNestedSet := map[interface{}]bool{
		"a": true,
		"b": true,
		"c": true,
		1:   true,
		2:   true,
		3:   true,
	}

	fmt.Println("\n非嵌套集�?")
	printSet(nonNestedSet)

	flattened = setUtils.Flatten(nonNestedSet)
	fmt.Println("非嵌套集合展开�?")
	printSet(flattened)

	// 测试边界情况
	emptyResult := setUtils.Flatten(nil)
	fmt.Printf("nil集合展开结果长度: %d\n", len(emptyResult))
}

// 辅助函数：打印字�?func printDict(dict map[interface{}]interface{}) {
	for key, value := range dict {
		fmt.Printf("  %v: %v\n", key, value)
	}
}

// 辅助函数：打印列�?func printList(list []interface{}) {
	fmt.Print("  [")
	for i, item := range list {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Printf("%v", item)
	}
	fmt.Println("]")
}

// 辅助函数：打印集�?func printSet(set map[interface{}]bool) {
	fmt.Print("  {")
	first := true
	for key := range set {
		if !first {
			fmt.Print(", ")
		}
		fmt.Printf("%v", key)
		first = false
	}
	fmt.Println("}")
}
