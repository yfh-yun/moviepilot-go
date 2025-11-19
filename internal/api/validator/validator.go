// Package validator 提供数据验证功能
package validator

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// Validator 验证器
type Validator struct {
	validate *validator.Validate
	logger   *zap.Logger
}

// NewValidator 创建新的验证器
func NewValidator(logger *zap.Logger) *Validator {
	v := validator.New()
	
	// 注册自定义验证规则
	v.RegisterValidation("required_if", validateRequiredIf)
	v.RegisterValidation("unique", validateUnique)
	
	return &Validator{
		validate: v,
		logger:   logger,
	}
}

// Validate 验证结构体
func (v *Validator) Validate(s interface{}) error {
	if err := v.validate.Struct(s); err != nil {
		return err
	}
	return nil
}

// BindJSONAndValidate 绑定JSON并验证
func (v *Validator) BindJSONAndValidate(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		return fmt.Errorf("请求格式错误: %w", err)
	}
	
	if err := v.Validate(obj); err != nil {
		return fmt.Errorf("参数验证失败: %w", err)
	}
	
	return nil
}

// HandleValidationError 处理验证错误
func (v *Validator) HandleValidationError(c *gin.Context, err error) {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		messages := make([]string, len(validationErrors))
		for i, fieldError := range validationErrors {
			messages[i] = v.getErrorMessage(fieldError)
		}
		c.JSON(400, gin.H{
			"success": false,
			"error":   strings.Join(messages, "; "),
		})
		return
	}
	
	c.JSON(400, gin.H{
		"success": false,
		"error":   err.Error(),
	})
}

// TranslateError 翻译验证错误
func (v *Validator) TranslateError(err error) string {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		var messages []string
		for _, fieldError := range validationErrors {
			messages = append(messages, v.getErrorMessage(fieldError))
		}
		return strings.Join(messages, "; ")
	}
	return err.Error()
}

// getErrorMessage 获取错误消息
func (v *Validator) getErrorMessage(fe validator.FieldError) string {
	field := fe.Field()
	tag := fe.Tag()
	param := fe.Param()
	
	switch tag {
	case "required":
		return fmt.Sprintf("%s 是必填字段", field)
	case "min":
		return fmt.Sprintf("%s 长度不能少于 %s", field, param)
	case "max":
		return fmt.Sprintf("%s 长度不能超过 %s", field, param)
	case "email":
		return fmt.Sprintf("%s 必须是有效的邮箱地址", field)
	case "alphanum":
		return fmt.Sprintf("%s 只能包含字母和数字", field)
	case "containsany":
		return fmt.Sprintf("%s 必须包含特殊字符", field)
	case "required_if":
		return fmt.Sprintf("%s 在特定条件下为必填字段", field)
	case "unique":
		return fmt.Sprintf("%s 必须唯一", field)
	default:
		return fmt.Sprintf("%s 验证失败", field)
	}
}

// GetUintParam 获取uint参数
func GetUintParam(c *gin.Context, key string) (uint, error) {
	value := c.Param(key)
	if value == "" {
		return 0, fmt.Errorf("参数 %s 不能为空", key)
	}
	
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("参数 %s 必须是数字", key)
	}
	
	if intValue < 0 {
		return 0, fmt.Errorf("参数 %s 不能为负数", key)
	}
	
	return uint(intValue), nil
}

// GetIntParam 获取int参数
func GetIntParam(c *gin.Context, key string) (int, error) {
	value := c.Param(key)
	if value == "" {
		return 0, fmt.Errorf("参数 %s 不能为空", key)
	}
	
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("参数 %s 必须是数字", key)
	}
	
	return intValue, nil
}

// GetStringParam 获取string参数
func GetStringParam(c *gin.Context, key string) (string, error) {
	value := c.Param(key)
	if value == "" {
		return "", fmt.Errorf("参数 %s 不能为空", key)
	}
	
	return value, nil
}

// GetQueryParam 获取查询参数
func GetQueryParam(c *gin.Context, key string, defaultValue string) string {
	value := c.Query(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetIntQueryParam 获取整数查询参数
func GetIntQueryParam(c *gin.Context, key string, defaultValue int) (int, error) {
	value := c.Query(key)
	if value == "" {
		return defaultValue, nil
	}
	
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("查询参数 %s 必须是数字", key)
	}
	
	return intValue, nil
}

// validateRequiredIf 自定义验证：在特定条件下必填
func validateRequiredIf(fl validator.FieldLevel) bool {
	field := fl.Field()
	param := fl.Param()
	
	// 解析参数: field1=value1,field2=value2
	conditions := strings.Split(param, ",")
	for _, condition := range conditions {
		parts := strings.Split(condition, "=")
		if len(parts) != 2 {
			continue
		}
		
		otherField := parts[0]
		expectedValue := parts[1]
		
		// 获取其他字段的值
		otherFieldValue := fl.Top().FieldByName(otherField)
		if !otherFieldValue.IsValid() {
			continue
		}
		
		// 检查条件是否满足
		if fmt.Sprintf("%v", otherFieldValue.Interface()) == expectedValue {
			// 如果条件满足，检查当前字段是否有值
			return !isEmpty(field)
		}
	}
	
	return true
}

// validateUnique 自定义验证：唯一性验证（需要数据库查询）
func validateUnique(fl validator.FieldLevel) bool {
	// 这里只是一个示例，实际使用时需要结合数据库查询
	// 可以通过上下文或者依赖注入的方式传入数据库查询函数
	return true
}

// isEmpty 检查字段是否为空
func isEmpty(field reflect.Value) bool {
	switch field.Kind() {
	case reflect.String:
		return field.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return field.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return field.Float() == 0
	case reflect.Bool:
		return !field.Bool()
	case reflect.Slice, reflect.Array, reflect.Map:
		return field.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return field.IsNil()
	default:
		return false
	}
}

// ValidatePage 验证分页参数
func ValidatePage(page, pageSize int) error {
	if page < 1 {
		return errors.New("页码必须大于0")
	}
	if pageSize < 1 || pageSize > 100 {
		return errors.New("每页数量必须在1-100之间")
	}
	return nil
}

// NormalizePage 规范化分页参数
func NormalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}