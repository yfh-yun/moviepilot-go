package chain

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"moviepilot-go/internal/models"
	"moviepilot-go/pkg/utils"
	"go.uber.org/zap"
)

// ConditionEvaluator 条件评估器
type ConditionEvaluator struct {
	logger *zap.Logger
}

// NewConditionEvaluator 创建条件评估器
func NewConditionEvaluator(logger *zap.Logger) *ConditionEvaluator {
	return &ConditionEvaluator{
		logger: logger,
	}
}

// EvaluateConditions 评估条件组
func (e *ConditionEvaluator) EvaluateConditions(ctx context.Context, conditions []model.Condition, logicalOperator model.LogicalOperator) (bool, error) {
	if len(conditions) == 0 {
		return true, nil
	}

	// 处理条件组的逻辑关系
	if len(conditions) == 1 {
		return e.evaluateSingleCondition(ctx, conditions[0])
	}

	results := make([]bool, len(conditions))
	for i, condition := range conditions {
		result, err := e.evaluateSingleCondition(ctx, condition)
		if err != nil {
			e.logger.Error("评估条件失败", 
				zap.Int("condition_index", i),
				zap.String("field", condition.Field),
				zap.Error(err))
			return false, fmt.Errorf("评估条件失败: %w", err)
		}
		results[i] = result
	}

	// 根据逻辑操作符组合结果
	switch logicalOperator {
	case model.LogicalOperatorAnd:
		for _, result := range results {
			if !result {
				return false, nil
			}
		}
		return true, nil
	case model.LogicalOperatorOr:
		for _, result := range results {
			if result {
				return true, nil
			}
		}
		return false, nil
	default:
		return false, fmt.Errorf("不支持的逻辑操作符: %s", logicalOperator)
	}
}

// evaluateSingleCondition 评估单个条件
func (e *ConditionEvaluator) evaluateSingleCondition(ctx context.Context, condition model.Condition) (bool, error) {
	// 解析左值（变量或常量）
	left, err := e.resolveValue(condition.Left, ctx)
	if err != nil {
		return false, fmt.Errorf("解析左值失败: %w", err)
	}

	// 解析右值（变量或常量）
	right, err := e.resolveValue(condition.Right, ctx)
	if err != nil {
		return false, fmt.Errorf("解析右值失败: %w", err)
	}

	// 执行比较操作
	result, err := e.compareValues(left, right, condition.Operator, condition.CaseSensitive)
	if err != nil {
		return false, fmt.Errorf("比较值失败: %w", err)
	}

	e.logger.Debug("条件评估结果",
		zap.String("field", condition.Field),
		zap.String("operator", string(condition.Operator)),
		zap.Any("left", left),
		zap.Any("right", right),
		zap.Bool("result", result))

	return result, nil
}

// resolveValue 解析值（支持变量引用和常量）
func (e *ConditionEvaluator) resolveValue(value interface{}, context *model.WorkflowContext) (interface{}, error) {
	switch v := value.(type) {
	case string:
		// 检查是否是变量引用（格式：${variable}）
		if strings.HasPrefix(v, "${") && strings.HasSuffix(v, "}") {
			varName := strings.TrimPrefix(strings.TrimSuffix(v, "}"), "${")
			return e.resolveStringValue(varName, context)
		}
		return v, nil
	case map[string]interface{}:
		return e.resolveObjectValue(v, context)
	default:
		return value, nil
	}
}

// resolveStringValue 解析字符串值（支持嵌套变量）
func (e *ConditionEvaluator) resolveStringValue(value string, context *model.WorkflowContext) (interface{}, error) {
	// 检查是否是嵌套变量路径（如：${user.profile.name}）
	parts := strings.Split(value, ".")
	if len(parts) == 1 {
		// 简单变量
		if val, exists := context.GetVariable(value); exists {
			return val, nil
		}
		return nil, fmt.Errorf("变量不存在: %s", value)
	}

	// 嵌套变量
	rootVar := parts[0]
	if val, exists := context.GetVariable(rootVar); exists {
		return e.getNestedValue(val, parts[1:])
	}
	return nil, fmt.Errorf("根变量不存在: %s", rootVar)
}

// getNestedValue 获取嵌套对象的值
func (e *ConditionEvaluator) getNestedValue(obj interface{}, path []string) (interface{}, error) {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if len(path) == 0 {
		return obj, nil
	}

	currentField := path[0]
	remainingPath := path[1:]

	switch val.Kind() {
	case reflect.Struct:
		field := val.FieldByName(currentField)
		if !field.IsValid() {
			return nil, fmt.Errorf("字段不存在: %s", currentField)
		}
		if len(remainingPath) > 0 {
			return e.getNestedValue(field.Interface(), remainingPath)
		}
		return field.Interface(), nil

	case reflect.Map:
		mapValue := val.MapIndex(reflect.ValueOf(currentField))
		if !mapValue.IsValid() {
			return nil, fmt.Errorf("键不存在: %s", currentField)
		}
		if len(remainingPath) > 0 {
			return e.getNestedValue(mapValue.Interface(), remainingPath)
		}
		return mapValue.Interface(), nil

	default:
		return nil, fmt.Errorf("不支持的类型: %v", val.Kind())
	}
}

// resolveObjectValue 解析对象值
func (e *ConditionEvaluator) resolveObjectValue(obj map[string]interface{}, context *model.WorkflowContext) (interface{}, error) {
	result := make(map[string]interface{})
	for key, value := range obj {
		resolved, err := e.resolveValue(value, context)
		if err != nil {
			return nil, err
		}
		result[key] = resolved
	}
	return result, nil
}

// compareValues 比较两个值
func (e *ConditionEvaluator) compareValues(left, right interface{}, operator model.Operator, caseSensitive bool) (bool, error) {
	switch operator {
	case model.OperatorEquals:
		return e.equals(left, right, caseSensitive), nil
	case model.OperatorNotEquals:
		return !e.equals(left, right, caseSensitive), nil
	case model.OperatorGreaterThan:
		return e.greaterThan(left, right)
	case model.OperatorGreaterThanOrEqual:
		return e.greaterThanOrEqual(left, right)
	case model.OperatorLessThan:
		return e.lessThan(left, right)
	case model.OperatorLessThanOrEqual:
		return e.lessThanOrEqual(left, right)
	case model.OperatorContains:
		return e.contains(left, right, caseSensitive), nil
	case model.OperatorNotContains:
		return !e.contains(left, right, caseSensitive), nil
	case model.OperatorStartsWith:
		return e.startsWith(left, right, caseSensitive), nil
	case model.OperatorEndsWith:
		return e.endsWith(left, right, caseSensitive), nil
	case model.OperatorMatches:
		return e.matches(left, right, caseSensitive), nil
	case model.OperatorIn:
		return e.in(left, right), nil
	case model.OperatorNotIn:
		return !e.in(left, right), nil
	case model.OperatorIsEmpty:
		return e.isEmpty(left), nil
	case model.OperatorIsNotEmpty:
		return !e.isEmpty(left), nil
	case model.OperatorIsNull:
		return e.isNull(left), nil
	case model.OperatorIsNotNull:
		return !e.isNull(left), nil
	default:
		return false, fmt.Errorf("不支持的操作符: %s", operator)
	}
}

// equals 相等比较
func (e *ConditionEvaluator) equals(left, right interface{}, caseSensitive bool) bool {
	leftStr := utils.StringValue(left)
	rightStr := utils.StringValue(right)

	if !caseSensitive {
		leftStr = strings.ToLower(leftStr)
		rightStr = strings.ToLower(rightStr)
	}

	return leftStr == rightStr
}

// greaterThan 大于比较
func (e *ConditionEvaluator) greaterThan(left, right interface{}) (bool, error) {
	leftNum, err := e.toNumber(left)
	if err != nil {
		return false, fmt.Errorf("解析左值为数字失败: %w", err)
	}

	rightNum, err := e.toNumber(right)
	if err != nil {
		return false, fmt.Errorf("解析右值为数字失败: %w", err)
	}

	return leftNum > rightNum, nil
}

// greaterThanOrEqual 大于等于比较
func (e *ConditionEvaluator) greaterThanOrEqual(left, right interface{}) (bool, error) {
	leftNum, err := e.toNumber(left)
	if err != nil {
		return false, fmt.Errorf("解析左值为数字失败: %w", err)
	}

	rightNum, err := e.toNumber(right)
	if err != nil {
		return false, fmt.Errorf("解析右值为数字失败: %w", err)
	}

	return leftNum >= rightNum, nil
}

// lessThan 小于比较
func (e *ConditionEvaluator) lessThan(left, right interface{}) (bool, error) {
	leftNum, err := e.toNumber(left)
	if err != nil {
		return false, fmt.Errorf("解析左值为数字失败: %w", err)
	}

	rightNum, err := e.toNumber(right)
	if err != nil {
		return false, fmt.Errorf("解析右值为数字失败: %w", err)
	}

	return leftNum < rightNum, nil
}

// lessThanOrEqual 小于等于比较
func (e *ConditionEvaluator) lessThanOrEqual(left, right interface{}) (bool, error) {
	leftNum, err := e.toNumber(left)
	if err != nil {
		return false, fmt.Errorf("解析左值为数字失败: %w", err)
	}

	rightNum, err := e.toNumber(right)
	if err != nil {
		return false, fmt.Errorf("解析右值为数字失败: %w", err)
	}

	return leftNum <= rightNum, nil
}

// contains 包含比较
func (e *ConditionEvaluator) contains(left, right interface{}, caseSensitive bool) bool {
	leftStr := utils.StringValue(left)
	rightStr := utils.StringValue(right)

	if !caseSensitive {
		leftStr = strings.ToLower(leftStr)
		rightStr = strings.ToLower(rightStr)
	}

	return strings.Contains(leftStr, rightStr)
}

// startsWith 以...开头
func (e *ConditionEvaluator) startsWith(left, right interface{}, caseSensitive bool) bool {
	leftStr := utils.StringValue(left)
	rightStr := utils.StringValue(right)

	if !caseSensitive {
		leftStr = strings.ToLower(leftStr)
		rightStr = strings.ToLower(rightStr)
	}

	return strings.HasPrefix(leftStr, rightStr)
}

// endsWith 以...结尾
func (e *ConditionEvaluator) endsWith(left, right interface{}, caseSensitive bool) bool {
	leftStr := utils.StringValue(left)
	rightStr := utils.StringValue(right)

	if !caseSensitive {
		leftStr = strings.ToLower(leftStr)
		rightStr = strings.ToLower(rightStr)
	}

	return strings.HasSuffix(leftStr, rightStr)
}

// matches 正则匹配
func (e *ConditionEvaluator) matches(left, right interface{}, caseSensitive bool) bool {
	leftStr := utils.StringValue(left)
	rightStr := utils.StringValue(right)

	if !caseSensitive {
		rightStr = "(?i)" + rightStr
	}

	matched, err := regexp.MatchString(rightStr, leftStr)
	if err != nil {
		e.logger.Error("正则匹配失败", zap.Error(err))
		return false
	}

	return matched
}

// in 在列表中
func (e *ConditionEvaluator) in(left, right interface{}) bool {
	rightSlice, ok := right.([]interface{})
	if !ok {
		// 尝试将右值转换为切片
		rightStr := utils.StringValue(right)
		if rightStr == "" {
			return false
		}
		rightSlice = []interface{}{rightStr}
	}

	for _, item := range rightSlice {
		if e.equals(left, item, true) {
			return true
		}
	}

	return false
}

// isEmpty 是否为空
func (e *ConditionEvaluator) isEmpty(value interface{}) bool {
	if value == nil {
		return true
	}

	switch v := value.(type) {
	case string:
		return v == ""
	case []interface{}:
		return len(v) == 0
	case map[string]interface{}:
		return len(v) == 0
	default:
		return utils.StringValue(value) == ""
	}
}

// isNull 是否为null
func (e *ConditionEvaluator) isNull(value interface{}) bool {
	return value == nil
}

// toNumber 转换为数字
func (e *ConditionEvaluator) toNumber(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("无法转换为数字: %T", value)
	}
}

// resolveObjectValue 解析对象值
func (e *ConditionEvaluator) resolveObjectValue(obj map[string]interface{}, context *model.WorkflowContext) (interface{}, error) {
	result := make(map[string]interface{})
	for key, value := range obj {
		resolved, err := e.resolveValue(value, context)
		if err != nil {
			return nil, err
		}
		result[key] = resolved
	}
	return result, nil
}