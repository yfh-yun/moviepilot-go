package filter

import (
	"strings"
	"sync"
)

// RuleParser 规则解析�?type RuleParser struct {
	// 使用字符串处理简单的逻辑表达式解�?}

var (
	parserInstance *RuleParser
	parserOnce     sync.Once
)

// GetRuleParser 获取规则解析器单�?func GetRuleParser() *RuleParser {
	parserOnce.Do(func() {
		parserInstance = &RuleParser{}
	})
	return parserInstance
}

// Parse 解析给定的表达式
func (r *RuleParser) Parse(expression string) []interface{} {
	// 移除多余空格
	expression = strings.TrimSpace(expression)
	
	// 首先处理括号表达�?	result := r.parseExpression(expression)
	
	return []interface{}{result}
}

// parseExpression 解析表达式，处理括号、逻辑非、逻辑与、逻辑或操作符
func (r *RuleParser) parseExpression(expr string) []interface{} {
	// 递归处理表达�?	
	// 先处理括�?	for strings.Contains(expr, "(") {
		// 找到最后一个左括号
		lastOpen := strings.LastIndex(expr, "(")
		// 找到对应的右括号
		nextClose := strings.Index(expr[lastOpen:], ")")
		if nextClose == -1 {
			break
		}
		nextClose += lastOpen
		
		// 提取括号内的内容
		innerExpr := expr[lastOpen+1 : nextClose]
		// 递归解析括号内容
		innerResult := r.parseExpression(innerExpr)
		
		// 替换原表达式中的括号部分
		expr = expr[:lastOpen] + "EXPR_PLACEHOLDER" + expr[nextClose+1:]
		
		// 如果是逻辑非操作符后面跟括�?		if lastOpen > 0 && expr[lastOpen-1] == '!' {
			return []interface{}{"not", innerResult}
		}
	}
	
	// 按空格分�?	parts := strings.Fields(expr)
	
	// 处理逻辑非操作符
	for i := 0; i < len(parts); i++ {
		if parts[i] == "!" && i+1 < len(parts) {
			parts[i] = "not"
			parts[i+1] = "(" + parts[i+1] + ")"
		}
	}
	
	// 处理逻辑与操作符
	for i := 0; i < len(parts); i++ {
		if parts[i] == "&" && i > 0 && i+1 < len(parts) {
			left := parts[i-1]
			right := parts[i+1]
			parts = append(parts[:i-1], append([]string{"(" + left + " and " + right + ")"}, parts[i+2:]...)...)
			i-- // 调整索引
		}
	}
	
	// 处理逻辑或操作符
	for i := 0; i < len(parts); i++ {
		if parts[i] == "|" && i > 0 && i+1 < len(parts) {
			left := parts[i-1]
			right := parts[i+1]
			parts = append(parts[:i-1], append([]string{"(" + left + " or " + right + ")"}, parts[i+2:]...)...)
			i-- // 调整索引
		}
	}
	
	// 如果只剩下一个元素，直接返回
	if len(parts) == 1 {
		// 检查是否是特殊标记的表达式
		if strings.HasPrefix(parts[0], "(") && strings.HasSuffix(parts[0], ")") {
			inner := parts[0][1 : len(parts[0])-1]
			return r.parseExpression(inner)
		}
		return []interface{}{parts[0]}
	}
	
	// 否则构造嵌套结�?	var result []interface{}
	for _, part := range parts {
		if part == "and" || part == "or" {
			result = append(result, part)
		} else if strings.HasPrefix(part, "(") && strings.HasSuffix(part, ")") {
			inner := part[1 : len(part)-1]
			result = append(result, r.parseExpression(inner))
		} else {
			result = append(result, part)
		}
	}
	
	return result
}
