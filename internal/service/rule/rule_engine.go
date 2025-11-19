package rule

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/yfh-yun/moviepilot-go/internal/model"
	"go.uber.org/zap"
)

// RuleEngine 规则引擎
type RuleEngine struct {
	logger       *zap.Logger
	rules        map[string]*Rule
	ruleSets     map[string]*RuleSet
	compiler     *RuleCompiler
	evaluator    *RuleEvaluator
	executor     *RuleExecutor
	configStore  *DynamicConfigStore
	mu           sync.RWMutex
}

// Rule 规则定义
type Rule struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Enabled     bool                  `json:"enabled"`
	Priority    int                   `json:"priority"`
	Conditions  []*Condition          `json:"conditions"`
	Actions     []*Action             `json:"actions"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
}

// Condition 条件
type Condition struct {
	ID         string                 `json:"id"`
	Type       ConditionType          `json:"type"`
	Field      string                 `json:"field"`
	Operator   Operator              `json:"operator"`
	Value      interface{}            `json:"value"`
	Parameters map[string]interface{} `json:"parameters"`
	Negate     bool                   `json:"negate"`
}

// Action 动作
type Action struct {
	ID         string                 `json:"id"`
	Type       ActionType             `json:"type"`
	Parameters map[string]interface{} `json:"parameters"`
	Async      bool                   `json:"async"`
}

// RuleSet 规则集
type RuleSet struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Rules       []string              `json:"rules"` // 规则ID列表
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
}

// ConditionType 条件类型
type ConditionType string

const (
	ConditionTypeEquals       ConditionType = "equals"
	ConditionTypeNotEquals    ConditionType = "not_equals"
	ConditionTypeContains     ConditionType = "contains"
	ConditionTypeStartsWith   ConditionType = "starts_with"
	ConditionTypeEndsWith     ConditionType = "ends_with"
	ConditionTypeMatches      ConditionType = "matches"
	ConditionTypeIn          ConditionType = "in"
	ConditionTypeGreaterThan ConditionType = "greater_than"
	ConditionTypeLessThan    ConditionType = "less_than"
	ConditionTypeIsEmpty     ConditionType = "is_empty"
	ConditionTypeIsNotEmpty  ConditionType = "is_not_empty"
	ConditionTypeIsNull      ConditionType = "is_null"
	ConditionTypeIsNotNull   ConditionType = "is_not_null"
	ConditionTypeCustom      ConditionType = "custom"
)

// ActionType 动作类型
type ActionType string

const (
	ActionTypeSetVariable   ActionType = "set_variable"
	ActionTypeExecuteFunction ActionType = "execute_function"
	ActionTypeSendWebhook   ActionType = "send_webhook"
	ActionTypeLogMessage    ActionType = "log_message"
	ActionTypeCustom        ActionType = "custom"
)

// Operator 操作符
type Operator string

const (
	OperatorEquals        Operator = "=="
	OperatorNotEquals     Operator = "!="
	OperatorGreaterThan   Operator = ">"
	OperatorLessThan      Operator = "<"
	OperatorGreaterEqual Operator = ">="
	OperatorLessEqual    Operator = "<="
	OperatorContains     Operator = "contains"
	OperatorNotContains  Operator = "not_contains"
	OperatorStartsWith   Operator = "starts_with"
	OperatorEndsWith     Operator = "ends_with"
	OperatorMatches      Operator = "matches"
	OperatorIn          Operator = "in"
	OperatorNotIn       Operator = "not_in"
	OperatorIsEmpty      Operator = "is_empty"
	OperatorIsNotEmpty   Operator = "is_not_empty"
)

// RuleCompiler 规则编译器
type RuleCompiler struct {
	logger *zap.Logger
}

// RuleEvaluator 规则评估器
type RuleEvaluator struct {
	logger      *zap.Logger
	functions   map[string]Function
	validators  map[ConditionType]Validator
}

// RuleExecutor 规则执行器
type RuleExecutor struct {
	logger    *zap.Logger
	actions   map[ActionType]ActionHandler
	variables map[string]interface{}
	mu        sync.RWMutex
}

// DynamicConfigStore 动态配置存储
type DynamicConfigStore struct {
	configs map[string]*Config
	mu      sync.RWMutex
}

// Config 配置项
type Config struct {
	Key         string                 `json:"key"`
	Value       interface{}            `json:"value"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// Function 函数接口
type Function interface {
	Name() string
	Validate(params map[string]interface{}) error
	Execute(ctx context.Context, params map[string]interface{}) (interface{}, error)
}

// Validator 验证器接口
type Validator interface {
	Validate(condition *Condition, context map[string]interface{}) (bool, error)
}

// ActionHandler 动作处理器接口
type ActionHandler interface {
	Execute(ctx context.Context, action *Action, context map[string]interface{}) error
}

// ExecutionContext 执行上下文
type ExecutionContext struct {
	Context   map[string]interface{}
	Variables map[string]interface{}
	Logger    *zap.Logger
}

// NewRuleEngine 创建规则引擎
func NewRuleEngine(logger *zap.Logger) *RuleEngine {
	engine := &RuleEngine{
		logger:      logger,
		rules:       make(map[string]*Rule),
		ruleSets:    make(map[string]*RuleSet),
		compiler:    NewRuleCompiler(logger),
		evaluator:   NewRuleEvaluator(logger),
		executor:    NewRuleExecutor(logger),
		configStore: NewDynamicConfigStore(),
	}

	// 注册内置函数和动作
	engine.registerBuiltins()

	return engine
}

// LoadRule 加载规则
func (e *RuleEngine) LoadRule(rule *Rule) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 验证规则
	if err := e.validateRule(rule); err != nil {
		return fmt.Errorf("规则验证失败: %w", err)
	}

	// 编译规则
	if err := e.compiler.Compile(rule); err != nil {
		return fmt.Errorf("规则编译失败: %w", err)
	}

	e.rules[rule.ID] = rule
	e.logger.Info("加载规则成功",
		zap.String("rule_id", rule.ID),
		zap.String("rule_name", rule.Name))

	return nil
}

// LoadRuleSet 加载规则集
func (e *RuleEngine) LoadRuleSet(ruleSet *RuleSet) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 验证规则集
	for _, ruleID := range ruleSet.Rules {
		if _, exists := e.rules[ruleID]; !exists {
			return fmt.Errorf("规则不存在: %s", ruleID)
		}
	}

	e.ruleSets[ruleSet.ID] = ruleSet
	e.logger.Info("加载规则集成功",
		zap.String("ruleset_id", ruleSet.ID),
		zap.String("ruleset_name", ruleSet.Name),
		zap.Strings("rule_ids", ruleSet.Rules))

	return nil
}

// ExecuteRules 执行规则
func (e *RuleEngine) ExecuteRules(ctx context.Context, ruleSetID string, data map[string]interface{}) ([]*RuleResult, error) {
	e.logger.Info("开始执行规则",
		zap.String("ruleset_id", ruleSetID))

	// 获取规则集
	ruleSet, exists := e.ruleSets[ruleSetID]
	if !exists {
		return nil, fmt.Errorf("规则集不存在: %s", ruleSetID)
	}

	var results []*RuleResult
	executionContext := &ExecutionContext{
		Context:   data,
		Variables: make(map[string]interface{}),
		Logger:    e.logger,
	}

	// 按优先级排序执行规则
	sortedRules := e.getSortedRules(ruleSet.Rules)

	for _, ruleID := range sortedRules {
		rule, exists := e.rules[ruleID]
		if !exists {
			e.logger.Warn("规则不存在，跳过", zap.String("rule_id", ruleID))
			continue
		}

		if !rule.Enabled {
			e.logger.Debug("规则已禁用，跳过", zap.String("rule_id", ruleID))
			continue
		}

		// 执行规则
		result := e.executeRule(ctx, rule, executionContext)
		results = append(results, result)

		// 如果规则匹配且动作执行失败，可以选择停止
		if result.Matched && !result.ActionSuccess {
			e.logger.Error("规则动作执行失败",
				zap.String("rule_id", ruleID),
				zap.Error(result.Error))
		}
	}

	e.logger.Info("规则执行完成",
		zap.String("ruleset_id", ruleSetID),
		zap.Int("total_rules", len(sortedRules)),
		zap.Int("matched_rules", e.countMatchedRules(results)))

	return results, nil
}

// executeRule 执行单个规则
func (e *RuleEngine) executeRule(ctx context.Context, rule *Rule, execContext *ExecutionContext) *RuleResult {
	result := &RuleResult{
		RuleID:    rule.ID,
		RuleName:  rule.Name,
		Timestamp: time.Now(),
	}

	e.logger.Debug("执行规则",
		zap.String("rule_id", rule.ID),
		zap.String("rule_name", rule.Name))

	// 评估条件
	matched, err := e.evaluator.Evaluate(ctx, rule, execContext)
	if err != nil {
		result.Error = fmt.Errorf("条件评估失败: %w", err)
		return result
	}

	result.Matched = matched

	if !matched {
		e.logger.Debug("规则条件不匹配", zap.String("rule_id", rule.ID))
		return result
	}

	e.logger.Info("规则条件匹配，执行动作", zap.String("rule_id", rule.ID))

	// 执行动作
	actionSuccess := true
	for _, action := range rule.Actions {
		if err := e.executor.Execute(ctx, action, execContext); err != nil {
			e.logger.Error("动作执行失败",
				zap.String("rule_id", rule.ID),
				zap.String("action_id", action.ID),
				zap.Error(err))
			actionSuccess = false
			result.Error = err
		}
	}

	result.ActionSuccess = actionSuccess
	return result
}

// UpdateConfig 更新动态配置
func (e *RuleEngine) UpdateConfig(key string, value interface{}, configType string) error {
	config := &Config{
		Key:       key,
		Value:     value,
		Type:      configType,
		UpdatedAt: time.Now(),
	}

	return e.configStore.Set(key, config)
}

// GetConfig 获取动态配置
func (e *RuleEngine) GetConfig(key string) (interface{}, bool) {
	config, exists := e.configStore.Get(key)
	if !exists {
		return nil, false
	}
	return config.Value, true
}

// ListRules 列出所有规则
func (e *RuleEngine) ListRules() []*Rule {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var rules []*Rule
	for _, rule := range e.rules {
		rules = append(rules, rule)
	}
	return rules
}

// ListRuleSets 列出所有规则集
func (e *RuleEngine) ListRuleSets() []*RuleSet {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var ruleSets []*RuleSet
	for _, ruleSet := range e.ruleSets {
		ruleSets = append(ruleSets, ruleSet)
	}
	return ruleSets
}

// validateRule 验证规则
func (e *RuleEngine) validateRule(rule *Rule) error {
	if rule.ID == "" {
		return fmt.Errorf("规则ID不能为空")
	}
	if rule.Name == "" {
		return fmt.Errorf("规则名称不能为空")
	}
	if len(rule.Conditions) == 0 {
		return fmt.Errorf("规则必须包含至少一个条件")
	}
	if len(rule.Actions) == 0 {
		return fmt.Errorf("规则必须包含至少一个动作")
	}

	// 验证条件
	for _, condition := range rule.Conditions {
		if err := e.validateCondition(condition); err != nil {
			return fmt.Errorf("条件验证失败: %w", err)
		}
	}

	// 验证动作
	for _, action := range rule.Actions {
		if err := e.validateAction(action); err != nil {
			return fmt.Errorf("动作验证失败: %w", err)
		}
	}

	return nil
}

// validateCondition 验证条件
func (e *RuleEngine) validateCondition(condition *Condition) error {
	if condition.ID == "" {
		return fmt.Errorf("条件ID不能为空")
	}
	if condition.Field == "" {
		return fmt.Errorf("条件字段不能为空")
	}

	// 验证操作符
	validOperators := map[Operator]bool{
		OperatorEquals:        true,
		OperatorNotEquals:     true,
		OperatorGreaterThan:   true,
		OperatorLessThan:      true,
		OperatorGreaterEqual: true,
		OperatorLessEqual:    true,
		OperatorContains:     true,
		OperatorNotContains:  true,
		OperatorStartsWith:   true,
		OperatorEndsWith:     true,
		OperatorMatches:      true,
		OperatorIn:          true,
		OperatorNotIn:       true,
		OperatorIsEmpty:      true,
		OperatorIsNotEmpty:   true,
	}

	if !validOperators[condition.Operator] {
		return fmt.Errorf("不支持的操作符: %s", condition.Operator)
	}

	return nil
}

// validateAction 验证动作
func (e *RuleEngine) validateAction(action *Action) error {
	if action.ID == "" {
		return fmt.Errorf("动作ID不能为空")
	}

	// 验证动作类型
	validTypes := map[ActionType]bool{
		ActionTypeSetVariable:   true,
		ActionTypeExecuteFunction: true,
		ActionTypeSendWebhook:   true,
		ActionTypeLogMessage:    true,
		ActionTypeCustom:        true,
	}

	if !validTypes[action.Type] {
		return fmt.Errorf("不支持的动作类型: %s", action.Type)
	}

	return nil
}

// getSortedRules 按优先级排序规则
func (e *RuleEngine) getSortedRules(ruleIDs []string) []string {
	var rules []struct {
		ID       string
		Priority int
	}

	for _, ruleID := range ruleIDs {
		if rule, exists := e.rules[ruleID]; exists {
			rules = append(rules, struct {
				ID       string
				Priority int
			}{ID: ruleID, Priority: rule.Priority})
		}
	}

	// 按优先级排序（数值越小优先级越高）
	for i := 0; i < len(rules)-1; i++ {
		for j := i + 1; j < len(rules); j++ {
			if rules[i].Priority > rules[j].Priority {
				rules[i], rules[j] = rules[j], rules[i]
			}
		}
	}

	var sortedIDs []string
	for _, rule := range rules {
		sortedIDs = append(sortedIDs, rule.ID)
	}

	return sortedIDs
}

// countMatchedRules 统计匹配的规则数
func (e *RuleEngine) countMatchedRules(results []*RuleResult) int {
	count := 0
	for _, result := range results {
		if result.Matched {
			count++
		}
	}
	return count
}

// registerBuiltins 注册内置函数和动作
func (e *RuleEngine) registerBuiltins() {
	// 注册内置函数
	e.evaluator.RegisterFunction(&StringFunction{})
	e.evaluator.RegisterFunction(&NumberFunction{})
	e.evaluator.RegisterFunction(&DateFunction{})

	// 注册内置动作处理器
	e.executor.RegisterHandler(ActionTypeSetVariable, &SetVariableHandler{})
	e.executor.RegisterHandler(ActionTypeExecuteFunction, &ExecuteFunctionHandler{})
	e.executor.RegisterHandler(ActionTypeSendWebhook, &SendWebhookHandler{})
	e.executor.RegisterHandler(ActionTypeLogMessage, &LogMessageHandler{})
}

// RuleResult 规则执行结果
type RuleResult struct {
	RuleID        string      `json:"rule_id"`
	RuleName      string      `json:"rule_name"`
	Matched       bool        `json:"matched"`
	ActionSuccess bool        `json:"action_success"`
	Timestamp     time.Time   `json:"timestamp"`
	Error         error       `json:"error,omitempty"`
	Metadata      interface{} `json:"metadata,omitempty"`
}