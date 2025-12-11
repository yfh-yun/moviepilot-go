package common

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// ValidateRequest 验证请求参数
func ValidateRequest(c *gin.Context, req any) error {
	// 绑定JSON
	if err := c.ShouldBindJSON(req); err != nil {
		return NewAPIError(ErrCodeBadRequest, "请求参数格式错误", err)
	}

	// 验证结构体
	if err := validate.Struct(req); err != nil {
		return NewAPIError(ErrCodeValidationFailed, "参数验证失败", err)
	}

	return nil
}

// ValidateQuery 验证查询参数
func ValidateQuery(c *gin.Context, req any) error {
	// 绑定查询参数
	if err := c.ShouldBindQuery(req); err != nil {
		return NewAPIError(ErrCodeBadRequest, "查询参数格式错误", err)
	}

	// 验证结构体
	if err := validate.Struct(req); err != nil {
		return NewAPIError(ErrCodeValidationFailed, "参数验证失败", err)
	}

	return nil
}

// ValidateURI 验证URI参数
func ValidateURI(c *gin.Context, req any) error {
	// 绑定URI参数
	if err := c.ShouldBindUri(req); err != nil {
		return NewAPIError(ErrCodeBadRequest, "URI参数格式错误", err)
	}

	// 验证结构体
	if err := validate.Struct(req); err != nil {
		return NewAPIError(ErrCodeValidationFailed, "参数验证失败", err)
	}

	return nil
}

// GetValidator 获取验证器实例
func GetValidator() *validator.Validate {
	return validate
}
