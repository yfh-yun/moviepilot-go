// Package filter 规则解析器
package filter

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unicode"
)

// ExpressionType 表达式类型
type ExpressionType string

const (
	ExprTypeAtom ExpressionType = "atom"
	ExprTypeNot  ExpressionType = "not"
	ExprTypeAnd  ExpressionType = "and"
	ExprTypeOr   ExpressionType = "or"
)

// ParsedExpression 解析后的表达式
type ParsedExpression struct {
	Type     ExpressionType    `json:"type"`
	Value    string            `json:"value,omitempty"`
	Children []interface{}     `json:"children,omitempty"`
	Position int               `json:"position"`
}

// RuleParser 规则解析器
type RuleParser struct {
	mutex    sync.RWMutex
	initialized bool
}

// NewRuleParser 创建规则解析器
func NewRuleParser() *RuleParser {
	return &RuleParser{
		mutex:    sync.RWMutex{},
		initialized: false,
	}
}

// Parse 解析表达式
func (rp *RuleParser) Parse(expression string) (*ParsedExpression, error) {
	rp.mutex.RLock()
	defer rp.mutex.RUnlock()

	// 预处理表达式
	expression = rp.preprocessExpression(expression)
	if expression == "" {
		return nil, fmt.Errorf("empty expression")
	}

	// 解析表达式
	result, err := rp.parseExpression(expression, 0)
	if err != nil {
		return nil, err
	}

	// 检查是否完全解析
	if rp.skipWhitespace(expression[result.Position:]) != "" {
		return nil, fmt.Errorf("unexpected token at position %d", result.Position)
	}

	return result, nil
}

// preprocessExpression 预处理表达式
func (rp *RuleParser) preprocessExpression(expression string) string {
	// 移除多余的空白字符
	expression = regexp.MustCompile(`\s+`).ReplaceAllString(expression, " ")
	expression = strings.TrimSpace(expression)
	
	// 移除注释
	expression = regexp.MustCompile(`#.*`).ReplaceAllString(expression, "")
	
	return expression
}

// parseExpression 解析表达式
func (rp *RuleParser) parseExpression(expression string, pos int) (*ParsedExpression, error) {
	// 解析OR表达式（最低优先级）
	return rp.parseOrExpression(expression, pos)
}

// parseOrExpression 解析OR表达式
func (rp *RuleParser) parseOrExpression(expression string, pos int) (*ParsedExpression, error) {
	left, err := rp.parseAndExpression(expression, pos)
	if err != nil {
		return nil, err
	}

	remaining := rp.skipWhitespace(expression[left.Position:])
	for strings.HasPrefix(remaining, "|") {
		// 消耗 | 操作符
		left.Position += len(remaining) - len(remaining[1:])
		remaining = rp.skipWhitespace(expression[left.Position:])

		// 解析右侧表达式
		right, err := rp.parseAndExpression(expression, left.Position)
		if err != nil {
			return nil, err
		}

		// 创建OR表达式
		left = &ParsedExpression{
			Type:     ExprTypeOr,
			Children: []interface{}{left, right},
			Position: right.Position,
		}

		remaining = rp.skipWhitespace(expression[left.Position:])
	}

	return left, nil
}

// parseAndExpression 解析AND表达式
func (rp *RuleParser) parseAndExpression(expression string, pos int) (*ParsedExpression, error) {
	left, err := rp.parseNotExpression(expression, pos)
	if err != nil {
		return nil, err
	}

	remaining := rp.skipWhitespace(expression[left.Position:])
	for strings.HasPrefix(remaining, "&") {
		// 消耗 & 操作符
		left.Position += len(remaining) - len(remaining[1:])
		remaining = rp.skipWhitespace(expression[left.Position:])

		// 解析右侧表达式
		right, err := rp.parseNotExpression(expression, left.Position)
		if err != nil {
			return nil, err
		}

		// 创建AND表达式
		left = &ParsedExpression{
			Type:     ExprTypeAnd,
			Children: []interface{}{left, right},
			Position: right.Position,
		}

		remaining = rp.skipWhitespace(expression[left.Position:])
	}

	return left, nil
}

// parseNotExpression 解析NOT表达式
func (rp *RuleParser) parseNotExpression(expression string, pos int) (*ParsedExpression, error) {
	remaining := rp.skipWhitespace(expression[pos:])
	
	if strings.HasPrefix(remaining, "!") {
		// 消耗 ! 操作符
		pos += len(remaining) - len(remaining[1:])
		
		// 解析操作数
		operand, err := rp.parsePrimaryExpression(expression, pos)
		if err != nil {
			return nil, err
		}

		// 创建NOT表达式
		return &ParsedExpression{
			Type:     ExprTypeNot,
			Children: []interface{}{operand},
			Position: operand.Position,
		}, nil
	}

	return rp.parsePrimaryExpression(expression, pos)
}

// parsePrimaryExpression 解析基本表达式
func (rp *RuleParser) parsePrimaryExpression(expression string, pos int) (*ParsedExpression, error) {
	remaining := rp.skipWhitespace(expression[pos:])
	
	if strings.HasPrefix(remaining, "(") {
		// 解析括号表达式
		pos += 1 // 消耗 (
		
		// 解析括号内的表达式
		inner, err := rp.parseExpression(expression, pos)
		if err != nil {
			return nil, err
		}

		// 检查 )
		remaining = rp.skipWhitespace(expression[inner.Position:])
		if !strings.HasPrefix(remaining, ")") {
			return nil, fmt.Errorf("expected ')' at position %d", inner.Position)
		}

		inner.Position += 1 // 消耗 )
		return inner, nil
	}

	// 解析原子表达式
	return rp.parseAtomExpression(expression, pos)
}

// parseAtomExpression 解析原子表达式
func (rp *RuleParser) parseAtomExpression(expression string, pos int) (*ParsedExpression, error) {
	remaining := rp.skipWhitespace(expression[pos:])
	if remaining == "" {
		return nil, fmt.Errorf("unexpected end of expression at position %d", pos)
	}

	// 解析标识符或数字
	var atom strings.Builder
	i := 0
	
	for i < len(remaining) {
		char := rune(remaining[i])
		
		// 允许字母、数字、下划线、连字符
		if unicode.IsLetter(char) || unicode.IsDigit(char) || char == '_' || char == '-' {
			atom.WriteRune(char)
			i++
		} else {
			break
		}
	}

	if atom.Len() == 0 {
		return nil, fmt.Errorf("invalid atom at position %d", pos)
	}

	atomStr := atom.String()
	
	// 检查是否为数字
	if _, err := strconv.Atoi(atomStr); err == nil {
		// 纯数字，转换为标识符
		atomStr = "NUM_" + atomStr
	}

	return &ParsedExpression{
		Type:     ExprTypeAtom,
		Value:    atomStr,
		Position: pos + i,
	}, nil
}

// skipWhitespace 跳过空白字符
func (rp *RuleParser) skipWhitespace(s string) string {
	i := 0
	for i < len(s) && unicode.IsSpace(rune(s[i])) {
		i++
	}
	return s[i:]
}

// ValidateExpression 验证表达式
func (rp *RuleParser) ValidateExpression(expression string) error {
	_, err := rp.Parse(expression)
	return err
}

// OptimizeExpression 优化表达式
func (rp *RuleParser) OptimizeExpression(expr *ParsedExpression) *ParsedExpression {
	if expr == nil {
		return nil
	}

	switch expr.Type {
	case ExprTypeAtom:
		return expr
	case ExprTypeNot:
		if len(expr.Children) == 0 {
			return expr
		}
		child := rp.OptimizeExpression(expr.Children[0].(*ParsedExpression))
		
		// 双重否定优化
		if child.Type == ExprTypeNot && len(child.Children) > 0 {
			return rp.OptimizeExpression(child.Children[0].(*ParsedExpression))
		}
		
		return &ParsedExpression{
			Type:     ExprTypeNot,
			Children: []interface{}{child},
			Position: expr.Position,
		}
		
	case ExprTypeAnd:
		return rp.optimizeBinaryExpression(expr, ExprTypeAnd)
		
	case ExprTypeOr:
		return rp.optimizeBinaryExpression(expr, ExprTypeOr)
		
	default:
		return expr
	}
}

// optimizeBinaryExpression 优化二元表达式
func (rp *RuleParser) optimizeBinaryExpression(expr *ParsedExpression, exprType ExpressionType) *ParsedExpression {
	if len(expr.Children) < 2 {
		return expr
	}

	var optimizedChildren []interface{}
	for _, child := range expr.Children {
		optimized := rp.OptimizeExpression(child.(*ParsedExpression))
		
		// 如果子表达式类型相同且可以合并，则合并
		if optimized.Type == exprType && len(optimized.Children) > 0 {
			for _, grandChild := range optimized.Children {
				optimizedChildren = append(optimizedChildren, grandChild)
			}
		} else {
			optimizedChildren = append(optimizedChildren, optimized)
		}
	}

	// 移除重复的原子表达式
	seen := make(map[string]bool)
	var uniqueChildren []interface{}
	for _, child := range optimizedChildren {
		childExpr := child.(*ParsedExpression)
		if childExpr.Type == ExprTypeAtom {
			if !seen[childExpr.Value] {
				seen[childExpr.Value] = true
				uniqueChildren = append(uniqueChildren, child)
			}
		} else {
			uniqueChildren = append(uniqueChildren, child)
		}
	}

	return &ParsedExpression{
		Type:     exprType,
		Children: uniqueChildren,
		Position: expr.Position,
	}
}

// ExpressionToString 将表达式转换为字符串
func (rp *RuleParser) ExpressionToString(expr *ParsedExpression) string {
	if expr == nil {
		return ""
	}

	switch expr.Type {
	case ExprTypeAtom:
		return expr.Value
	case ExprTypeNot:
		if len(expr.Children) > 0 {
			return "!" + rp.ExpressionToString(expr.Children[0].(*ParsedExpression))
		}
		return "!"
	case ExprTypeAnd:
		return rp.binaryExpressionToString(expr.Children, "&")
	case ExprTypeOr:
		return rp.binaryExpressionToString(expr.Children, "|")
	default:
		return ""
	}
}

// binaryExpressionToString 二元表达式转字符串
func (rp *RuleParser) binaryExpressionToString(children []interface{}, operator string) string {
	if len(children) == 0 {
		return ""
	}
	
	var parts []string
	for _, child := range children {
		childStr := rp.ExpressionToString(child.(*ParsedExpression))
		if childStr != "" {
			parts = append(parts, childStr)
		}
	}
	
	return strings.Join(parts, " "+operator+" ")
}

// GetExpressionVariables 获取表达式中的变量
func (rp *RuleParser) GetExpressionVariables(expr *ParsedExpression) []string {
	var variables []string
	variableSet := make(map[string]bool)
	
	rp.collectVariables(expr, variableSet)
	
	for variable := range variableSet {
		variables = append(variables, variable)
	}
	
	return variables
}

// collectVariables 收集变量
func (rp *RuleParser) collectVariables(expr *ParsedExpression, variableSet map[string]bool) {
	if expr == nil {
		return
	}
	
	switch expr.Type {
	case ExprTypeAtom:
		variableSet[expr.Value] = true
	case ExprTypeNot, ExprTypeAnd, ExprTypeOr:
		for _, child := range expr.Children {
			rp.collectVariables(child.(*ParsedExpression), variableSet)
		}
	}
}

// SubstituteVariables 替换变量
func (rp *RuleParser) SubstituteVariables(expr *ParsedExpression, substitutions map[string]string) *ParsedExpression {
	if expr == nil {
		return nil
	}
	
	switch expr.Type {
	case ExprTypeAtom:
		if newValue, exists := substitutions[expr.Value]; exists {
			return &ParsedExpression{
				Type:     ExprTypeAtom,
				Value:    newValue,
				Position: expr.Position,
			}
		}
		return expr
		
	case ExprTypeNot:
		if len(expr.Children) > 0 {
			child := rp.SubstituteVariables(expr.Children[0].(*ParsedExpression), substitutions)
			return &ParsedExpression{
				Type:     ExprTypeNot,
				Children: []interface{}{child},
				Position: expr.Position,
			}
		}
		return expr
		
	case ExprTypeAnd, ExprTypeOr:
		var newChildren []interface{}
		for _, child := range expr.Children {
			newChild := rp.SubstituteVariables(child.(*ParsedExpression), substitutions)
			newChildren = append(newChildren, newChild)
		}
		return &ParsedExpression{
			Type:     expr.Type,
			Children: newChildren,
			Position: expr.Position,
		}
		
	default:
		return expr
	}
}