package main

import (
	"fmt"
	"reflect"

	"moviepilot-go/internal/utils"
)

// 示例函数用于测试
func exampleFunc1(a int, b string) bool {
	return true
}

func exampleFunc2() {
	// 空函�?}

func exampleFunc3(x float64, y []int, z map[string]interface{}) string {
	return "test"
}

func main() {
	fmt.Println("=== 对象工具类示�?===")

	// 创建对象工具类实�?	objUtils := utils.NewObjectUtils()

	// 测试 IsObj 方法
	fmt.Println("\n--- IsObj 方法测试 ---")
	testIsObj(objUtils)

	// 测试 IsObjStr 方法
	fmt.Println("\n--- IsObjStr 方法测试 ---")
	testIsObjStr(objUtils)

	// 测试 Arguments 方法
	fmt.Println("\n--- Arguments 方法测试 ---")
	testArguments(objUtils)

	// 测试 CheckMethod 方法
	fmt.Println("\n--- CheckMethod 方法测试 ---")
	testCheckMethod(objUtils)

	// 测试 CheckSignature 方法
	fmt.Println("\n--- CheckSignature 方法测试 ---")
	testCheckSignature(objUtils)
}

func testIsObj(objUtils *utils.ObjectUtils) {
	// 测试基本类型
	fmt.Printf("IsObj(123): %v\n", objUtils.IsObj(123))                 // false
	fmt.Printf("IsObj(123.45): %v\n", objUtils.IsObj(123.45))           // false
	fmt.Printf("IsObj(true): %v\n", objUtils.IsObj(true))               // false
	fmt.Printf("IsObj(\"hello\"): %v\n", objUtils.IsObj("hello"))       // false

	// 测试复合类型
	fmt.Printf("IsObj([]int{1, 2, 3}): %v\n", objUtils.IsObj([]int{1, 2, 3}))     // true
	fmt.Printf("IsObj(map[string]int{\"a\": 1}): %v\n", objUtils.IsObj(map[string]int{"a": 1})) // true

	// 测试结构�?	type Person struct {
		Name string
		Age  int
	}
	person := Person{Name: "张三", Age: 30}
	fmt.Printf("IsObj(Person{Name: \"张三\", Age: 30}): %v\n", objUtils.IsObj(person)) // true

	// 测试nil
	fmt.Printf("IsObj(nil): %v\n", objUtils.IsObj(nil)) // true
}

func testIsObjStr(objUtils *utils.ObjectUtils) {
	// 测试字符串是否表示对�?	fmt.Printf("IsObjStr(\"{\\\"name\\\": \\\"张三\\\"}\"): %v\n", objUtils.IsObjStr("{\"name\": \"张三\"}")) // true
	fmt.Printf("IsObjStr(\"[1, 2, 3]\"): %v\n", objUtils.IsObjStr("[1, 2, 3]"))                         // true
	fmt.Printf("IsObjStr(\"(1, 2, 3)\"): %v\n", objUtils.IsObjStr("(1, 2, 3)"))                         // true
	fmt.Printf("IsObjStr(\"hello\"): %v\n", objUtils.IsObjStr("hello"))                                 // false
	fmt.Printf("IsObjStr(123): %v\n", objUtils.IsObjStr(123))                                           // false
	fmt.Printf("IsObjStr(\" { \\\"name\\\": \\\"张三\\\" } \"): %v\n", objUtils.IsObjStr(" { \"name\": \"张三\" } ")) // true
}

func testArguments(objUtils *utils.ObjectUtils) {
	// 测试函数参数个数
	fmt.Printf("Arguments(exampleFunc1): %d\n", objUtils.Arguments(exampleFunc1)) // 2
	fmt.Printf("Arguments(exampleFunc2): %d\n", objUtils.Arguments(exampleFunc2)) // 0
	fmt.Printf("Arguments(exampleFunc3): %d\n", objUtils.Arguments(exampleFunc3)) // 3

	// 测试匿名函数
	anonymousFunc := func(a, b, c int) {}
	fmt.Printf("Arguments(anonymousFunc): %d\n", objUtils.Arguments(anonymousFunc)) // 3

	// 测试非函数类�?	fmt.Printf("Arguments(\"not a function\"): %d\n", objUtils.Arguments("not a function")) // 0
}

func testCheckMethod(objUtils *utils.ObjectUtils) {
	// 测试函数是否已实�?	fmt.Printf("CheckMethod(exampleFunc1): %v\n", objUtils.CheckMethod(exampleFunc1))     // true
	fmt.Printf("CheckMethod(exampleFunc2): %v\n", objUtils.CheckMethod(exampleFunc2))     // true
	fmt.Printf("CheckMethod(exampleFunc3): %v\n", objUtils.CheckMethod(exampleFunc3))     // true

	// 测试匿名函数
	anonymousFunc := func() { fmt.Println("匿名函数") }
	fmt.Printf("CheckMethod(anonymousFunc): %v\n", objUtils.CheckMethod(anonymousFunc))   // true

	// 测试nil
	fmt.Printf("CheckMethod(nil): %v\n", objUtils.CheckMethod(nil))                       // false

	// 测试非函数类�?	fmt.Printf("CheckMethod(\"not a function\"): %v\n", objUtils.CheckMethod("not a function")) // false
}

func testCheckSignature(objUtils *utils.ObjectUtils) {
	// 测试函数签名检�?	fmt.Printf("CheckSignature(exampleFunc1, 123, \"hello\"): %v\n",
		objUtils.CheckSignature(exampleFunc1, 123, "hello")) // true

	fmt.Printf("CheckSignature(exampleFunc1, 123, 456): %v\n",
		objUtils.CheckSignature(exampleFunc1, 123, 456)) // false (第二个参数应该是string)

	fmt.Printf("CheckSignature(exampleFunc1, 123): %v\n",
		objUtils.CheckSignature(exampleFunc1, 123)) // false (参数个数不匹�?

	// 测试exampleFunc3
	slice := []int{1, 2, 3}
	mapping := map[string]interface{}{"key": "value"}
	fmt.Printf("CheckSignature(exampleFunc3, 123.45, slice, mapping): %v\n",
		objUtils.CheckSignature(exampleFunc3, 123.45, slice, mapping)) // true

	// 测试参数类型不匹�?	fmt.Printf("CheckSignature(exampleFunc3, \"string\", slice, mapping): %v\n",
		objUtils.CheckSignature(exampleFunc3, "string", slice, mapping)) // false (第一个参数应该是float64)

	// 测试nil函数
	fmt.Printf("CheckSignature(nil, 123): %v\n", objUtils.CheckSignature(nil, 123)) // false
}
