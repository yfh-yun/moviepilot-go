package rule

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

// NewRuleCompiler 创建规则编译器
func NewRuleCompiler(logger *zap.Logger) *RuleCompiler {
	return &RuleCompiler{
		logger: logger,
	}
}

// Compile 编译规则
func (c *RuleCompiler) Compile(rule *Rule) error {
	c.logger.Debug("编译规则",
		zap.String("rule_id", rule.ID),
		zap.String("rule_name", rule.Name))

	// 编译条件
	for _, condition := range rule.Conditions {
		if err := c.compileCondition(condition); err != nil {
			return fmt.Errorf("编译条件失败: %w", err)
		}
	}

	// 编译动作
	for _, action := range rule.Actions {
		if err := c.compileAction(action); err != nil {
			return fmt.Errorf("编译动作失败: %w", err)
		}
	}

	c.logger.Debug("规则编译成功", zap.String("rule_id", rule.ID))
	return nil
}

// compileCondition 编译条件
func (c *RuleCompiler) compileCondition(condition *Condition) error {
	// 编译字段路径
	if err := c.compileFieldPath(condition); err != nil {
		return err
	}

	// 编译值
	if err := c.compileValue(condition); err != nil {
		return err
	}

	// 编译参数
	if err := c.compileParameters(condition); err != nil {
		return err
	}

	return nil
}

// compileFieldPath 编译字段路径
func (c *RuleCompiler) compileFieldPath(condition *Condition) error {
	// 支持嵌套字段访问，如 "user.profile.name"
	// 编译时验证字段路径的有效性
	fields := strings.Split(condition.Field, ".")
	if len(fields) == 0 {
		return fmt.Errorf("字段路径不能为空")
	}

	// 验证字段名
	for _, field := range fields {
		if !isValidFieldName(field) {
			return fmt.Errorf("无效的字段名: %s", field)
		}
	}

	return nil
}

// compileValue 编译值
func (c *RuleCompiler) compileValue(condition *Condition) error {
	switch condition.Type {
	case ConditionTypeEquals, ConditionTypeNotEquals:
		return c.compileEqualsValue(condition)
	case ConditionTypeGreaterThan, ConditionTypeLessThan,
		ConditionTypeGreaterEqual, ConditionTypeLessEqual:
		return c.compileNumberValue(condition)
	case ConditionTypeContains, ConditionTypeStartsWith, ConditionTypeEndsWith:
		return c.compileStringValue(condition)
	case ConditionTypeMatches:
		return c.compileRegexValue(condition)
	case ConditionTypeIn, ConditionTypeNotIn:
		return c.compileArrayValue(condition)
	case ConditionTypeIsEmpty, ConditionTypeIsNotEmpty:
		return nil // 无需编译
	case ConditionTypeIsNull, ConditionTypeIsNotNull:
		return nil // 无需编译
	default:
		return nil // 自定义类型在运行时处理
	}
}

// compileEqualsValue 编译等于比较的值
func (c *RuleCompiler) compileEqualsValue(condition *Condition) error {
	// 确保值是可比较的类型
	if !isComparable(condition.Value) {
		return fmt.Errorf("值不可比较: %T", condition.Value)
	}
	return nil
}

// compileNumberValue 编译数字比较的值
func (c *RuleCompiler) compileNumberValue(condition *Condition) error {
	// 尝试将值转换为数字
	switch v := condition.Value.(type) {
	case int, int32, int64, float32, float64:
		return nil
	case string:
		if _, err := strconv.ParseFloat(v, 64); err != nil {
			return fmt.Errorf("无法转换为数字: %s", v)
		}
		return nil
	default:
		return fmt.Errorf("不支持的数字类型: %T", v)
	}
}

// compileStringValue 编码字符串操作的值
func (c *RuleCompiler) compileStringValue(condition *Condition) error {
	// 确保值是字符串或可转换为字符串
	if condition.Value == nil {
		return fmt.Errorf("字符串操作值不能为空")
	}

	strValue := fmt.Sprintf("%v", condition.Value)
	if strValue == "" {
		return fmt.Errorf("字符串操作值不能为空字符串")
	}

	// 缓存转换后的字符串
	condition.Parameters["compiled_string"] = strValue
	return nil
}

// compileRegexValue 编译正则表达式的值
func (c *RuleCompiler) compileRegexValue(condition *Condition) error {
	if strValue, ok := condition.Value.(string); ok {
		// 验证正则表达式语法
		_, err := regexp.Compile(strValue)
		if err != nil {
			return fmt.Errorf("无效的正则表达式: %w", err)
		}
		// 缓存编译后的正则表达式
		condition.Parameters["compiled_regex"] = strValue
		return nil
	}
	return fmt.Errorf("正则表达式值必须是字符串")
}

// compileArrayValue 编译数组操作的值
func (c *RuleCompiler) compileArrayValue(condition *Condition) error {
	// 确保值是数组或可以转换为数组
	switch v := condition.Value.(type) {
	case []interface{}:
		return nil // 已经是数组
	case []string:
		// 转换为 []interface{}
		array := make([]interface{}, len(v))
		for i, item := range v {
			array[i] = item
		}
		condition.Value = array
		return nil
	case string:
		// 解析JSON数组
		return nil // 简化实现
	default:
		return fmt.Errorf("不支持的数组类型: %T", v)
	}
}

// compileParameters 编译参数
func (c *RuleCompiler) compileParameters(condition *Condition) error {
	if condition.Parameters == nil {
		condition.Parameters = make(map[string]interface{})
		return nil
	}

	// 验证参数类型
	for key, value := range condition.Parameters {
		if !isValidParameterType(value) {
			return fmt.Errorf("无效的参数类型，键: %s, 类型: %T", key, value)
		}
	}

	return nil
}

// compileAction 编译动作
func (c *RuleCompiler) compileAction(action *Action) error {
	// 验证动作参数
	if action.Parameters == nil {
		action.Parameters = make(map[string]interface{})
	}

	// 根据动作类型进行特定编译
	switch action.Type {
	case ActionTypeSetVariable:
		return c.compileSetVariableAction(action)
	case ActionTypeExecuteFunction:
		return c.compileExecuteFunctionAction(action)
	case ActionTypeSendWebhook:
		return c.compileSendWebhookAction(action)
	case ActionTypeLogMessage:
		return c.compileLogMessageAction(action)
	default:
		return nil // 自定义动作在运行时处理
	}
}

// compileSetVariableAction 编译设置变量动作
func (c *RuleCompiler) compileSetVariableAction(action *Action) error {
	// 检查必需参数
	if _, exists := action.Parameters["name"]; !exists {
		return fmt.Errorf("设置变量动作缺少name参数")
	}
	if _, exists := action.Parameters["value"]; !exists {
		return fmt.Errorf("设置变量动作缺少value参数")
	}

	// 编译变量名
	if name, ok := action.Parameters["name"].(string); ok {
		if !isValidVariableName(name) {
			return fmt.Errorf("无效的变量名: %s", name)
		}
	} else {
		return fmt.Errorf("变量名必须是字符串")
	}

	return nil
}

// compileExecuteFunctionAction 编译执行函数动作
func (c *RuleCompiler) compileExecuteFunctionAction(action *Action) error {
	// 检查必需参数
	if _, exists := action.Parameters["function"]; !exists {
		return fmt.Errorf("执行函数动作缺少function参数")
	}

	// 编译函数名
	if functionName, ok := action.Parameters["function"].(string); ok {
		if !isValidFunctionName(functionName) {
			return fmt.Errorf("无效的函数名: %s", functionName)
		}
	} else {
		return fmt.Errorf("函数名必须是字符串")
	}

	// 编译函数参数
	if params, exists := action.Parameters["params"]; exists {
		if err := c.compileFunctionParams(params); err != nil {
			return fmt.Errorf("编译函数参数失败: %w", err)
		}
	}

	return nil
}

// compileSendWebhookAction 编译发送Webhook动作
func (c *RuleCompiler) compileSendWebhookAction(action *Action) error {
	// 检查必需参数
	if _, exists := action.Parameters["url"]; !exists {
		return fmt.Errorf("发送Webhook动作缺少url参数")
	}

	// 编译URL
	if url, ok := action.Parameters["url"].(string); ok {
		if !isValidURL(url) {
			return fmt.Errorf("无效的URL: %s", url)
		}
	} else {
		return fmt.Errorf("URL必须是字符串")
	}

	return nil
}

// compileLogMessageAction 编译日志消息动作
func (c *RuleCompiler) compileLogMessageAction(action *Action) error {
	// 检查必需参数
	if _, exists := action.Parameters["message"]; !exists {
		return fmt.Errorf("日志消息动作缺少message参数")
	}

	// 编译消息
	if message, ok := action.Parameters["message"].(string); ok {
		if message == "" {
			return fmt.Errorf("日志消息不能为空")
		}
	} else {
		return fmt.Errorf("消息必须是字符串")
	}

	// 编译日志级别
	if level, exists := action.Parameters["level"]; exists {
		if levelStr, ok := level.(string); ok {
			if !isValidLogLevel(levelStr) {
				return fmt.Errorf("无效的日志级别: %s", levelStr)
			}
		} else {
			return fmt.Errorf("日志级别必须是字符串")
		}
	}

	return nil
}

// compileFunctionParams 编译函数参数
func (c *RuleCompiler) compileFunctionParams(params interface{}) error {
	switch v := params.(type) {
	case map[string]interface{}:
		// 验证每个参数的类型
		for key, value := range v {
			if !isValidParameterType(value) {
				return fmt.Errorf("无效的函数参数类型，键: %s, 类型: %T", key, value)
			}
		}
		return nil
	case []interface{}:
		// 验证数组参数
		for i, value := range v {
			if !isValidParameterType(value) {
				return fmt.Errorf("无效的函数参数类型，索引: %d, 类型: %T", i, value)
			}
		}
		return nil
	default:
		return fmt.Errorf("函数参数必须是对象或数组")
	}
}

// 辅助验证函数
func isValidFieldName(field string) bool {
	if field == "" {
		return false
	}
	// 字段名可以包含字母、数字、下划线、点
	for _, r := range field {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || 
			 (r >= '0' && r <= '9') || r == '_' || r == '.') {
			return false
		}
	}
	return true
}

func isComparable(value interface{}) bool {
	// 检查值是否可以进行比较
	switch value.(type) {
	case nil, string, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64, bool:
		return true
	default:
		// 检查指针类型
		if reflect.ValueOf(value).Kind() == reflect.Ptr {
			return isComparable(reflect.ValueOf(value).Elem().Interface())
		}
		return false
	}
}

func isValidParameterType(value interface{}) bool {
	// 检查参数类型是否有效
	switch value.(type) {
	case nil, string, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64, bool:
		return true
	case []interface{}, map[string]interface{}:
		return true
	default:
		return false
	}
}

func isValidVariableName(name string) bool {
	if name == "" {
		return false
	}
	// 变量名以字母开头，只能包含字母、数字、下划线
	for i, r := range name {
		if i == 0 && !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')) {
			return false
		}
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || 
			 (r >= '0' && r <= '9') || r == '_') {
			return false
		}
	}
	return true
}

func isValidFunctionName(name string) bool {
	if name == "" {
		return false
	}
	// 函数名以字母开头，只能包含字母、数字、下划线
	for i, r := range name {
		if i == 0 && !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')) {
			return false
		}
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || 
			 (r >= '0' && r <= '9') || r == '_') {
			return false
		}
	}
	return true
}

func isValidURL(url string) bool {
	// 简化的URL验证
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}

func isValidLogLevel(level string) bool {
	validLevels := []string{"debug", "info", "warn", "error", "fatal"}
	for _, validLevel := range validLevels {
		if level == validLevel {
			return true
		}
	}
	return false
}