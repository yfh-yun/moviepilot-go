package actions

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"moviepilot-go/pkg/logger"
)

// PluginValidator 插件验证器接口
type PluginValidator interface {
	// ValidateInvokeRequest 验证插件调用请求
	ValidateInvokeRequest(req *PluginInvokeRequest) error
	// ValidatePluginID 验证插件ID
	ValidatePluginID(pluginID string) error
	// ValidatePluginMethod 验证插件方法名
	ValidatePluginMethod(method string) error
	// ValidatePluginType 验证插件类型
	ValidatePluginType(pluginType string) error
	// ValidateHistoryParams 验证历史记录查询参数
	ValidateHistoryParams(params *PluginHistoryParams) error
	// RegisterValidators 注册验证器到Gin框架
	RegisterValidators() error
}

// pluginValidator 插件验证器实现
type pluginValidator struct {
	logger logger.Logger
}

// NewPluginValidator 创建插件验证器实例
func NewPluginValidator(logger logger.Logger) PluginValidator {
	return &pluginValidator{
		logger: logger,
	}
}

// ValidateInvokeRequest 验证插件调用请求
func (v *pluginValidator) ValidateInvokeRequest(req *PluginInvokeRequest) error {
	// 验证插件ID
	if err := v.ValidatePluginID(req.PluginID); err != nil {
		return err
	}

	// 验证方法名
	if err := v.ValidatePluginMethod(req.Method); err != nil {
		return err
	}

	// 验证调用者
	if req.Caller != "" {
		if len(req.Caller) > 50 {
			return errors.New("调用者名称长度不能超过50个字符")
		}
		if !isValidCaller(req.Caller) {
			return errors.New("调用者名称只能包含字母、数字、下划线和中文字符")
		}
	}

	// 验证超时时间
	if req.Timeout < 0 || req.Timeout > 300*time.Second {
		return errors.New("超时时间必须在0-300秒之间")
	}

	// 验证参数
	if err := v.validateArguments(req.Arguments); err != nil {
		return err
	}

	return nil
}

// ValidatePluginID 验证插件ID
func (v *pluginValidator) ValidatePluginID(pluginID string) error {
	if pluginID == "" {
		return errors.New("插件ID不能为空")
	}

	if len(pluginID) < 3 || len(pluginID) > 50 {
		return errors.New("插件ID长度必须在3-50个字符之间")
	}

	// 插件ID格式验证：只能包含字母、数字、下划线、连字符和点
	matched, err := regexp.MatchString(`^[a-zA-Z0-9_\-\.]+$`, pluginID)
	if err != nil || !matched {
		return errors.New("插件ID只能包含字母、数字、下划线、连字符和点")
	}

	return nil
}

// ValidatePluginMethod 验证插件方法名
func (v *pluginValidator) ValidatePluginMethod(method string) error {
	if method == "" {
		return errors.New("方法名不能为空")
	}

	if len(method) < 2 || len(method) > 50 {
		return errors.New("方法名长度必须在2-50个字符之间")
	}

	// 方法名格式验证：只能包含字母、数字、下划线，且必须以字母开头
	matched, err := regexp.MatchString(`^[a-zA-Z][a-zA-Z0-9_]*$`, method)
	if err != nil || !matched {
		return errors.New("方法名只能包含字母、数字、下划线，且必须以字母开头")
	}

	// 禁止的方法名
	forbiddenMethods := []string{"__init__", "__del__", "__call__", "__enter__", "__exit__"}
	for _, forbidden := range forbiddenMethods {
		if method == forbidden {
			return errors.New("方法名不能使用保留关键字")
		}
	}

	return nil
}

// ValidatePluginType 验证插件类型
func (v *pluginValidator) ValidatePluginType(pluginType string) error {
	if pluginType == "" {
		return nil // 空类型表示不限制
	}

	// 支持的插件类型
	supportedTypes := map[string]bool{
		"site":        true,
		"indexer":     true,
		"mediaserver": true,
		"notification": true,
		"downloader":  true,
		"scraper":     true,
		"filter":      true,
	}

	if !supportedTypes[strings.ToLower(pluginType)] {
		return errors.New("不支持的插件类型")
	}

	return nil
}

// ValidateHistoryParams 验证历史记录查询参数
func (v *pluginValidator) ValidateHistoryParams(params *PluginHistoryParams) error {
	// 验证插件ID（可选）
	if params.PluginID != "" {
		if err := v.ValidatePluginID(params.PluginID); err != nil {
			return err
		}
	}

	// 验证分页参数
	if params.Page < 0 {
		return errors.New("页码不能为负数")
	}
	if params.PageSize < 0 || params.PageSize > 100 {
		return errors.New("每页大小必须在0-100之间")
	}

	// 验证时间范围
	if params.StartTime > params.EndTime && params.EndTime != 0 {
		return errors.New("开始时间不能晚于结束时间")
	}

	// 验证排序字段
	if params.OrderBy != "" {
		supportedOrders := map[string]bool{
			"invoke_time": true,
			"duration":    true,
			"success":     true,
		}
		if !supportedOrders[strings.ToLower(params.OrderBy)] {
			return errors.New("不支持的排序字段")
		}
	}

	// 验证排序方向
	if params.OrderDir != "" && params.OrderDir != "asc" && params.OrderDir != "desc" {
		return errors.New("排序方向只能是asc或desc")
	}

	return nil
}

// RegisterValidators 注册验证器到Gin框架
func (v *pluginValidator) RegisterValidators() error {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// 注册插件ID验证
		if err := v.RegisterValidation("plugin_id", v.validatePluginIDTag); err != nil {
			v.logger.Error("注册插件ID验证器失败", "error", err.Error())
			return err
		}

		// 注册插件方法验证
		if err := v.RegisterValidation("plugin_method", v.validatePluginMethodTag); err != nil {
			v.logger.Error("注册插件方法验证器失败", "error", err.Error())
			return err
		}

		// 注册插件类型验证
		if err := v.RegisterValidation("plugin_type", v.validatePluginTypeTag); err != nil {
			v.logger.Error("注册插件类型验证器失败", "error", err.Error())
			return err
		}

		v.logger.Info("插件验证器注册成功")
		return nil
	}

	return errors.New("无法获取验证器引擎")
}

// validateArguments 验证插件参数
func (v *pluginValidator) validateArguments(args map[string]interface{}) error {
	if args == nil {
		return nil
	}

	// 参数数量限制
	if len(args) > 20 {
		return errors.New("参数数量不能超过20个")
	}

	// 验证每个参数
	for key, value := range args {
		// 验证键名
		if err := v.validateArgumentKey(key); err != nil {
			return err
		}

		// 验证值
		if err := v.validateArgumentValue(value); err != nil {
			return errors.New("参数值验证失败: " + err.Error())
		}
	}

	return nil
}

// validateArgumentKey 验证参数键名
func (v *pluginValidator) validateArgumentKey(key string) error {
	if key == "" {
		return errors.New("参数键名不能为空")
	}

	if len(key) > 30 {
		return errors.New("参数键名长度不能超过30个字符")
	}

	// 键名格式验证：只能包含字母、数字、下划线，且必须以字母开头
	matched, err := regexp.MatchString(`^[a-zA-Z][a-zA-Z0-9_]*$`, key)
	if err != nil || !matched {
		return errors.New("参数键名只能包含字母、数字、下划线，且必须以字母开头")
	}

	return nil
}

// validateArgumentValue 验证参数值
func (v *pluginValidator) validateArgumentValue(value interface{}) error {
	if value == nil {
		return nil
	}

	// 检查值的大小和类型
	switch v := value.(type) {
	case string:
		if len(v) > 1000 {
			return errors.New("字符串参数长度不能超过1000个字符")
		}
	case []interface{}:
		if len(v) > 100 {
			return errors.New("数组参数长度不能超过100个元素")
		}
		// 递归验证数组元素
		for _, item := range v {
			if err := v.validateArgumentValue(item); err != nil {
				return err
			}
		}
	case map[string]interface{}:
		if len(v) > 20 {
			return errors.New("对象参数不能超过20个字段")
		}
		// 递归验证对象字段
		for k, item := range v {
			if err := v.validateArgumentKey(k); err != nil {
				return err
			}
			if err := v.validateArgumentValue(item); err != nil {
				return err
			}
		}
	}

	return nil
}

// validatePluginIDTag 验证标签函数
func (v *pluginValidator) validatePluginIDTag(fl validator.FieldLevel) bool {
	pluginID := fl.Field().String()
	err := v.ValidatePluginID(pluginID)
	return err == nil
}

// validatePluginMethodTag 验证标签函数
func (v *pluginValidator) validatePluginMethodTag(fl validator.FieldLevel) bool {
	method := fl.Field().String()
	err := v.ValidatePluginMethod(method)
	return err == nil
}

// validatePluginTypeTag 验证标签函数
func (v *pluginValidator) validatePluginTypeTag(fl validator.FieldLevel) bool {
	pluginType := fl.Field().String()
	err := v.ValidatePluginType(pluginType)
	return err == nil
}

// isValidCaller 验证调用者名称
func isValidCaller(caller string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_\u4e00-\u9fa5]+$`, caller)
	return matched
}

// ValidatePluginVersion 验证插件版本号
func (v *pluginValidator) ValidatePluginVersion(version string) error {
	if version == "" {
		return errors.New("版本号不能为空")
	}

	// 简单的语义化版本验证
	matched, err := regexp.MatchString(`^\d+\.\d+\.\d+(-[a-zA-Z0-9_\.]+)?$`, version)
	if err != nil || !matched {
		return errors.New("版本号格式不正确，应为X.Y.Z或X.Y.Z-xxx格式")
	}

	return nil
}

// ParseDuration 解析持续时间字符串
func (v *pluginValidator) ParseDuration(durationStr string) (time.Duration, error) {
	if durationStr == "" {
		return 0, errors.New("持续时间不能为空")
	}

	// 支持秒数或时间格式
	if seconds, err := strconv.Atoi(durationStr); err == nil {
		return time.Duration(seconds) * time.Second, nil
	}

	return time.ParseDuration(durationStr)
}
