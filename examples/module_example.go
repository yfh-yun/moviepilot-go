package main

import (
	"fmt"
	"reflect"

	"moviepilot-go/internal/helper"
)

// ExampleStruct 示例结构�?type ExampleStruct struct {
	Name string
	Age  int
}

// ExampleFunction 示例函数
func ExampleFunction() string {
	return "Hello from example function"
}

// ExampleMethod 示例方法
func (e *ExampleStruct) ExampleMethod() string {
	return fmt.Sprintf("Hello, I'm %s, %d years old", e.Name, e.Age)
}

func main() {
	// 创建模块帮助类实�?	moduleHelper := helper.NewModuleHelper()
	
	// 示例1: 使用Load方法加载模块
	fmt.Println("=== 使用Load方法 ===")
	modules := moduleHelper.Load("internal.helper", nil)
	fmt.Printf("加载到的模块数量: %d\n", len(modules))
	
	// 示例2: 使用带过滤器的Load方法
	fmt.Println("\n=== 使用带过滤器的Load方法 ===")
	filterFunc := func(name string, obj interface{}) bool {
		// 只加载名称包�?Example"的模�?		return name != "" && obj != nil && reflect.TypeOf(obj) != nil
	}
	
	filteredModules := moduleHelper.Load("internal.helper", filterFunc)
	fmt.Printf("过滤后加载到的模块数�? %d\n", len(filteredModules))
	
	// 示例3: 使用LoadWithPreFilter方法
	fmt.Println("\n=== 使用LoadWithPreFilter方法 ===")
	preFilteredModules := moduleHelper.LoadWithPreFilter("internal.helper", nil)
	fmt.Printf("预过滤后加载到的模块数量: %d\n", len(preFilteredModules))
	
	// 示例4: 动态导入所有模�?	fmt.Println("\n=== 动态导入所有模�?===")
	allModules := moduleHelper.DynamicImportAllModules("internal/helper", "helper")
	fmt.Printf("目录中的模块: %v\n", allModules)
	
	// 示例5: 创建实例并调用方�?	fmt.Println("\n=== 创建实例并调用方�?===")
	example := &ExampleStruct{Name: "Alice", Age: 30}
	result := example.ExampleMethod()
	fmt.Println(result)
	
	// 示例6: 调用函数
	fmt.Println("\n=== 调用函数 ===")
	funcResult := ExampleFunction()
	fmt.Println(funcResult)
}
