package validator

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// Validator 验证器接口
type Validator interface {
	Validate(interface{}) error
}

// StringValidator 字符串验证器
type StringValidator struct {
	Required bool
	Min      int
	Max      int
	Pattern  string
}

// Validate 验证字符串
func (v *StringValidator) Validate(obj interface{}) error {
	str, ok := obj.(string)
	if !ok {
		return &ValidationError{"Value must be a string"}
	}
	
	if v.Required && str == "" {
		return &ValidationError{"Field is required"}
	}
	
	if v.Min > 0 && len(str) < v.Min {
		return &ValidationError{fmt.Sprintf("Length must be at least %d", v.Min)}
	}
	
	if v.Max > 0 && len(str) > v.Max {
		return &ValidationError{fmt.Sprintf("Length must be at most %d", v.Max)}
	}
	
	if v.Pattern != "" {
		matched, err := regexp.MatchString(v.Pattern, str)
		if err != nil {
			return &ValidationError{"Invalid pattern"}
		}
		if !matched {
			return &ValidationError{"Value does not match pattern"}
		}
	}
	
	return nil
}

// ValidationError 验证错误
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// ValidateStruct 验证结构体
func ValidateStruct(obj interface{}) error {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	
	if val.Kind() != reflect.Struct {
		return &ValidationError{"Object must be a struct or pointer to struct"}
	}
	
	typ := val.Type()
	var errors []string
	
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)
		
		// 检查required标签
		required := false
		if tag := fieldType.Tag.Get("validate"); tag != "" {
			tags := strings.Split(tag, ",")
			for _, t := range tags {
				if t == "required" {
					required = true
				}
			}
		}
		
		// 验证必填字段
		if required && isZero(field) {
			errors = append(errors, fmt.Sprintf("%s is required", fieldType.Name))
		}
	}
	
	if len(errors) > 0 {
		return &ValidationError{strings.Join(errors, "; ")}
	}
	
	return nil
}

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Interface:
		return v.IsNil()
	default:
		return false
	}
}