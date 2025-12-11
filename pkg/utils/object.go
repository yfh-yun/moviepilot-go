package utils

import (
	"reflect"
	"strings"
)

// IsObj 判断是否为“对象类型”（参考 Python ObjectUtils.is_obj）
// 这里约定：切片、数组、映射、结构体视为对象；基础数值/布尔/字符串/字节视为非对象。
func IsObj(v any) bool {
	if v == nil {
		return false
	}
	t := reflect.TypeOf(v)
	k := t.Kind()

	switch k {
	case reflect.Slice, reflect.Array, reflect.Map, reflect.Struct:
		return true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64,
		reflect.Bool,
		reflect.String:
		return false
	default:
		return true
	}
}

// IsObjStr 判断一个值是否为“看起来像对象的字符串”，即以 '{' '[' '(' 开头
func IsObjStr(v any) bool {
	s, ok := v.(string)
	if !ok {
		return false
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	return strings.HasPrefix(s, "{") || strings.HasPrefix(s, "[") || strings.HasPrefix(s, "(")
}

// Arguments 返回函数的参数个数，对应 Python ObjectUtils.arguments
func Arguments(fn any) int {
	t := reflect.TypeOf(fn)
	if t == nil || t.Kind() != reflect.Func {
		return 0
	}
	return t.NumIn()
}

// CheckMethod 检查函数是否已实现，对应 Python ObjectUtils.check_method
// 在 Go 中，函数声明即表示已实现，但可以检测接口类型的零值
func CheckMethod(fn any) bool {
	if fn == nil {
		return false
	}
	
	t := reflect.TypeOf(fn)
	k := t.Kind()
	
	// 如果是函数类型，直接返回 true（Go中函数声明即实现）
	if k == reflect.Func {
		return true
	}
	
	// 检查是否为接口类型的零值
	if k == reflect.Interface {
		v := reflect.ValueOf(fn)
		return !v.IsNil()
	}
	
	// 其他类型返回 true
	return true
}

// CheckSignature 检查实参是否与函数形参类型兼容，对应 Python ObjectUtils.check_signature
func CheckSignature(fn any, args ...any) bool {
	t := reflect.TypeOf(fn)
	if t == nil || t.Kind() != reflect.Func {
		return false
	}
	if len(args) != t.NumIn() {
		return false
	}

	for i, arg := range args {
		paramType := t.In(i)
		if arg == nil {
			// 对于接口类型，允许 nil
			if paramType.Kind() == reflect.Interface || paramType.Kind() == reflect.Slice || paramType.Kind() == reflect.Map || paramType.Kind() == reflect.Ptr || paramType.Kind() == reflect.Func || paramType.Kind() == reflect.Chan {
				continue
			}
			return false
		}

		argType := reflect.TypeOf(arg)
		if argType.AssignableTo(paramType) {
			continue
		}
		if argType.ConvertibleTo(paramType) {
			continue
		}
		return false
	}

	return true
}
