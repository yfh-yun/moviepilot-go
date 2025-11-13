package utils

import (
	"reflect"
	"strings"
)

// ObjectUtils 对象工具�?type ObjectUtils struct{}

// NewObjectUtils 创建新的对象工具类实�?func NewObjectUtils() *ObjectUtils {
	return &ObjectUtils{}
}

// IsObj 判断对象是否为复杂对象（list、dict、tuple等）
// 在Go中对应的是切片、映射、数组、结构体等复合类�?func (o *ObjectUtils) IsObj(obj interface{}) bool {
	if obj == nil {
		return true
	}

	// 获取对象的反射类�?	t := reflect.TypeOf(obj)

	// 检查是否为基本类型
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.Bool, reflect.String:
		return false
	case reflect.Slice, reflect.Map, reflect.Struct, reflect.Array:
		return true
	default:
		return true
	}
}

// IsObjStr 判断字符串是否表示一个对象（�?{ [ ( 开头）
func (o *ObjectUtils) IsObjStr(obj interface{}) bool {
	// 检查是否为字符串类�?	str, ok := obj.(string)
	if !ok {
		return false
	}

	// 去除首尾空格
	str = strings.TrimSpace(str)
	if len(str) == 0 {
		return false
	}

	// 检查是否以 { [ ( 开�?	return strings.HasPrefix(str, "{") ||
		strings.HasPrefix(str, "[") ||
		strings.HasPrefix(str, "(")
}

// Arguments 返回函数的参数个�?// 注意：在Go中无法直接获取函数参数个数，因为Go没有泛型反射机制
// 这里提供一个基于反射的方法，但功能有限
func (o *ObjectUtils) Arguments(fn interface{}) int {
	if fn == nil {
		return 0
	}

	// 获取函数的反射�?	fnValue := reflect.ValueOf(fn)
	if fnValue.Kind() != reflect.Func {
		return 0
	}

	// 获取函数类型
	fnType := fnValue.Type()
	
	// 返回输入参数个数
	return fnType.NumIn()
}

// CheckMethod 检查函数是否已实现
// 在Go中，这个功能比较难实现，因为Go编译时会检查函数实�?// 这里提供一个简单的检查方法，通过解析源代码来判断
func (o *ObjectUtils) CheckMethod(fn interface{}) bool {
	// 在Go中，如果函数能被编译和调用，说明已经实现
	// 这个检查主要用于Python，在Go中意义不�?	// 但为了保持接口一致性，我们仍然提供实现
	
	if fn == nil {
		return false
	}

	// 获取函数的反射�?	fnValue := reflect.ValueOf(fn)
	if fnValue.Kind() != reflect.Func {
		return false
	}

	// 对于Go函数，如果能获取到则说明已实�?	// 我们尝试解析源代码来检查实�?	// 但Go的反射机制不提供源代码信息，所以这里只能做基本检�?	return !fnValue.IsNil()
}

// CheckSignature 检查输出与函数的参数类型是否一�?// 在Go中，类型检查在编译时完成，运行时检查主要用于反射场�?func (o *ObjectUtils) CheckSignature(fn interface{}, args ...interface{}) bool {
	if fn == nil {
		return false
	}

	// 获取函数的反射�?	fnValue := reflect.ValueOf(fn)
	if fnValue.Kind() != reflect.Func {
		return false
	}

	// 获取函数类型
	fnType := fnValue.Type()

	// 检查参数个数是否匹�?	if len(args) != fnType.NumIn() {
		return false
	}

	// 检查每个参数的类型是否匹配
	for i, arg := range args {
		argType := reflect.TypeOf(arg)
		paramType := fnType.In(i)
		
		// 如果参数类型不匹配，返回false
		if !argType.AssignableTo(paramType) {
			return false
		}
	}

	return true
}
