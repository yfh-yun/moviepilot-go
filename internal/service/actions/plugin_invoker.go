// Package actions 提供动作系统的业务逻辑实现
package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/model"

	"go.uber.org/zap"
)

// PluginInvoker 智能插件调用器
// 提供安全、高效的插件调用机制，支持多种插件类型
type PluginInvoker struct {
	pluginManager   PluginManager
	securityManager SecurityManager
	logger          *zap.Logger
	circuitBreaker  *CircuitBreaker
	cache           *PluginCache
	mutex           sync.RWMutex
}

// PluginManager 插件管理器接口
type PluginManager interface {
	GetPlugin(pluginID string) (*model.Plugin, error)
	CallPluginMethod(ctx context.Context, pluginID, method string, args map[string]interface{}) (interface{}, error)
	ValidatePluginCall(ctx context.Context, pluginID, method string) error
}

// SecurityManager 安全管理器接口
type SecurityManager interface {
	ValidatePluginAccess(ctx context.Context, pluginID, method string, userID string) error
	AuditPluginCall(ctx context.Context, pluginID, method string, args map[string]interface{}) error
}

// NewPluginInvoker 创建智能插件调用器实例
func NewPluginInvoker(
	pluginManager PluginManager,
	securityManager SecurityManager,
) *PluginInvoker {
	return &PluginInvoker{
		pluginManager:   pluginManager,
		securityManager: securityManager,
		logger:          logger.NewLogger("plugin_invoker"),
		circuitBreaker:  NewCircuitBreaker(),
		cache:           NewPluginCache(),
	}
}

// InvokePlugin 调用插件方法
func (p *PluginInvoker) InvokePlugin(ctx context.Context, request *PluginInvokeRequest) (*PluginInvokeResponse, error) {
	p.logger.Info("开始调用插件",
		zap.String("plugin_id", request.PluginID),
		zap.String("method", request.Method),
		zap.String("caller", request.Caller))

	// 1. 安全检查
	if err := p.securityManager.ValidatePluginAccess(ctx, request.PluginID, request.Method, request.Caller); err != nil {
		return nil, fmt.Errorf("插件访问安全检查失败: %w", err)
	}

	// 2. 熔断器检查
	if p.circuitBreaker.IsOpen(request.PluginID) {
		return nil, fmt.Errorf("插件调用被熔断: %s", request.PluginID)
	}

	// 3. 参数验证和转换
	validatedArgs, err := p.validateAndConvertArgs(ctx, request.PluginID, request.Method, request.Arguments)
	if err != nil {
		return nil, fmt.Errorf("参数验证失败: %w", err)
	}

	// 4. 缓存检查
	cacheKey := p.generateCacheKey(request.PluginID, request.Method, validatedArgs)
	if cachedResult, found := p.cache.Get(cacheKey); found {
		p.logger.Info("使用缓存结果", zap.String("plugin_id", request.PluginID))
		return &PluginInvokeResponse{
			PluginID:  request.PluginID,
			Method:    request.Method,
			Result:    cachedResult,
			FromCache: true,
			InvokedAt: time.Now(),
			Duration:  0, // 缓存无耗时
		}, nil
	}

	// 5. 执行插件调用
	startTime := time.Now()
	result, err := p.executePluginCall(ctx, request.PluginID, request.Method, validatedArgs)
	duration := time.Since(startTime)

	// 6. 处理调用结果
	if err != nil {
		// 更新熔断器状态
		p.circuitBreaker.RecordFailure(request.PluginID)

		p.logger.Error("插件调用失败",
			zap.String("plugin_id", request.PluginID),
			zap.String("method", request.Method),
			zap.Duration("duration", duration),
			zap.Error(err))

		return nil, fmt.Errorf("插件调用失败: %w", err)
	}

	// 7. 缓存成功结果
	p.cache.Set(cacheKey, result, 5*time.Minute)

	// 8. 审计日志
	if err := p.securityManager.AuditPluginCall(ctx, request.PluginID, request.Method, validatedArgs); err != nil {
		p.logger.Warn("审计日志记录失败", zap.Error(err))
	}

	response := &PluginInvokeResponse{
		PluginID:  request.PluginID,
		Method:    request.Method,
		Result:    result,
		FromCache: false,
		InvokedAt: startTime,
		Duration:  duration,
	}

	p.logger.Info("插件调用成功",
		zap.String("plugin_id", request.PluginID),
		zap.String("method", request.Method),
		zap.Duration("duration", duration),
		zap.Bool("from_cache", false))

	return response, nil
}

// validateAndConvertArgs 参数验证和转换
func (p *PluginInvoker) validateAndConvertArgs(ctx context.Context, pluginID, method string, args map[string]interface{}) (map[string]interface{}, error) {
	// 获取插件元数据
	plugin, err := p.pluginManager.GetPlugin(pluginID)
	if err != nil {
		return nil, fmt.Errorf("获取插件信息失败: %w", err)
	}

	// 验证方法参数
	if err := p.validateMethodArgs(plugin, method, args); err != nil {
		return nil, fmt.Errorf("方法参数验证失败: %w", err)
	}

	// 参数类型转换和标准化
	convertedArgs := p.convertArgs(args)

	// 安全检查：防止注入攻击
	if err := p.sanitizeArgs(convertedArgs); err != nil {
		return nil, fmt.Errorf("参数安全检查失败: %w", err)
	}

	return convertedArgs, nil
}

// validateMethodArgs 验证方法参数
func (p *PluginInvoker) validateMethodArgs(plugin *model.Plugin, method string, args map[string]interface{}) error {
	// 检查方法是否存在
	var methodExists bool
	for _, m := range plugin.Methods {
		if m.Name == method {
			methodExists = true

			// 验证必需参数
			for _, param := range m.Parameters {
				if param.Required {
					if _, exists := args[param.Name]; !exists {
						return fmt.Errorf("缺少必需参数: %s", param.Name)
					}
				}
			}
			break
		}
	}

	if !methodExists {
		return fmt.Errorf("插件方法不存在: %s", method)
	}

	return nil
}

// convertArgs 参数类型转换
func (p *PluginInvoker) convertArgs(args map[string]interface{}) map[string]interface{} {
	converted := make(map[string]interface{})

	for key, value := range args {
		switch v := value.(type) {
		case string:
			// 字符串类型不做转换
			converted[key] = v
		case float64:
			// JSON数字转换为int64
			converted[key] = int64(v)
		case bool:
			// 布尔类型
			converted[key] = v
		case []interface{}:
			// 数组类型
			converted[key] = p.convertArray(v)
		case map[string]interface{}:
			// 对象类型
			converted[key] = p.convertArgs(v)
		default:
			// 其他类型转换为字符串
			converted[key] = fmt.Sprintf("%v", v)
		}
	}

	return converted
}

// convertArray 转换数组类型
func (p *PluginInvoker) convertArray(arr []interface{}) []interface{} {
	var result []interface{}

	for _, item := range arr {
		switch v := item.(type) {
		case string:
			result = append(result, v)
		case float64:
			result = append(result, int64(v))
		case bool:
			result = append(result, v)
		case []interface{}:
			result = append(result, p.convertArray(v))
		case map[string]interface{}:
			result = append(result, p.convertArgs(v))
		default:
			result = append(result, fmt.Sprintf("%v", v))
		}
	}

	return result
}

// sanitizeArgs 参数安全检查
func (p *PluginInvoker) sanitizeArgs(args map[string]interface{}) error {
	for key, value := range args {
		// 检查键名安全性
		if !p.isSafeKey(key) {
			return fmt.Errorf("参数键名不安全: %s", key)
		}

		// 检查值安全性
		if err := p.sanitizeValue(value); err != nil {
			return fmt.Errorf("参数值不安全: %s", key)
		}
	}

	return nil
}

// isSafeKey 检查键名安全性
func (p *PluginInvoker) isSafeKey(key string) bool {
	// 允许字母、数字、下划线
	for _, char := range key {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '_') {
			return false
		}
	}

	// 不允许的关键词
	dangerousKeywords := []string{"eval", "exec", "system", "shell", "cmd"}
	for _, keyword := range dangerousKeywords {
		if key == keyword {
			return false
		}
	}

	return true
}

// sanitizeValue 检查值安全性
func (p *PluginInvoker) sanitizeValue(value interface{}) error {
	switch v := value.(type) {
	case string:
		// 检查字符串是否包含危险内容
		dangerousPatterns := []string{"<script>", "javascript:", "eval(", "exec("}
		for _, pattern := range dangerousPatterns {
			if contains(v, pattern) {
				return fmt.Errorf("字符串包含危险内容: %s", pattern)
			}
		}
	case []interface{}:
		// 递归检查数组元素
		for _, item := range v {
			if err := p.sanitizeValue(item); err != nil {
				return err
			}
		}
	case map[string]interface{}:
		// 递归检查对象属性
		for _, item := range v {
			if err := p.sanitizeValue(item); err != nil {
				return err
			}
		}
	}

	return nil
}

// executePluginCall 执行插件调用
func (p *PluginInvoker) executePluginCall(ctx context.Context, pluginID, method string, args map[string]interface{}) (interface{}, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// 调用插件管理器执行具体调用
	result, err := p.pluginManager.CallPluginMethod(ctx, pluginID, method, args)
	if err != nil {
		return nil, err
	}

	// 验证返回结果
	if err := p.validateResult(result); err != nil {
		return nil, fmt.Errorf("返回结果验证失败: %w", err)
	}

	return result, nil
}

// validateResult 验证返回结果
func (p *PluginInvoker) validateResult(result interface{}) error {
	// 检查结果大小（防止过大数据返回）
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("结果序列化失败: %w", err)
	}

	if len(resultJSON) > 10*1024*1024 { // 10MB限制
		return fmt.Errorf("返回结果过大: %d bytes", len(resultJSON))
	}

	// 检查结果类型安全性
	if err := p.sanitizeValue(result); err != nil {
		return fmt.Errorf("返回结果内容不安全: %w", err)
	}

	return nil
}

// generateCacheKey 生成缓存键
func (p *PluginInvoker) generateCacheKey(pluginID, method string, args map[string]interface{}) string {
	argsJSON, _ := json.Marshal(args)
	return fmt.Sprintf("plugin:%s:%s:%s", pluginID, method, string(argsJSON))
}

// BatchInvokePlugins 批量调用插件
func (p *PluginInvoker) BatchInvokePlugins(ctx context.Context, requests []*PluginInvokeRequest) ([]*PluginInvokeResponse, error) {
	var wg sync.WaitGroup
	responses := make([]*PluginInvokeResponse, len(requests))
	errors := make([]error, len(requests))

	for i, request := range requests {
		wg.Add(1)

		go func(index int, req *PluginInvokeRequest) {
			defer wg.Done()

			response, err := p.InvokePlugin(ctx, req)
			responses[index] = response
			errors[index] = err
		}(i, request)
	}

	wg.Wait()

	// 检查错误
	for _, err := range errors {
		if err != nil {
			return nil, fmt.Errorf("批量调用失败: %w", err)
		}
	}

	return responses, nil
}

// PluginInvokeRequest 插件调用请求
type PluginInvokeRequest struct {
	PluginID  string                 `json:"plugin_id"`
	Method    string                 `json:"method"`
	Arguments map[string]interface{} `json:"arguments"`
	Caller    string                 `json:"caller"`
	Timeout   time.Duration          `json:"timeout"`
}

// PluginInvokeResponse 插件调用响应
type PluginInvokeResponse struct {
	PluginID  string        `json:"plugin_id"`
	Method    string        `json:"method"`
	Result    interface{}   `json:"result"`
	FromCache bool          `json:"from_cache"`
	InvokedAt time.Time     `json:"invoked_at"`
	Duration  time.Duration `json:"duration"`
}

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	states map[string]*circuitState
	mutex  sync.RWMutex
}

// circuitState 熔断器状态
type circuitState struct {
	failures    int
	lastFailure time.Time
	state       string // "closed", "open", "half-open"
}

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker() *CircuitBreaker {
	return &CircuitBreaker{
		states: make(map[string]*circuitState),
	}
}

// IsOpen 检查熔断器是否打开
func (cb *CircuitBreaker) IsOpen(pluginID string) bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	state, exists := cb.states[pluginID]
	if !exists {
		return false
	}

	return state.state == "open"
}

// RecordFailure 记录失败
func (cb *CircuitBreaker) RecordFailure(pluginID string) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	state, exists := cb.states[pluginID]
	if !exists {
		state = &circuitState{}
		cb.states[pluginID] = state
	}

	state.failures++
	state.lastFailure = time.Now()

	// 失败次数超过阈值，打开熔断器
	if state.failures >= 5 {
		state.state = "open"
	}
}

// RecordSuccess 记录成功
func (cb *CircuitBreaker) RecordSuccess(pluginID string) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	state, exists := cb.states[pluginID]
	if exists {
		state.failures = 0
		state.state = "closed"
	}
}

// PluginCache 插件缓存
type PluginCache struct {
	cache map[string]*cacheEntry
	mutex sync.RWMutex
}

// cacheEntry 缓存条目
type cacheEntry struct {
	value     interface{}
	expiresAt time.Time
}

// NewPluginCache 创建插件缓存
func NewPluginCache() *PluginCache {
	return &PluginCache{
		cache: make(map[string]*cacheEntry),
	}
}

// Get 获取缓存值
func (c *PluginCache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, exists := c.cache[key]
	if !exists || time.Now().After(entry.expiresAt) {
		return nil, false
	}

	return entry.value, true
}

// Set 设置缓存值
func (c *PluginCache) Set(key string, value interface{}, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache[key] = &cacheEntry{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
}

// 辅助函数：检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && contains(s[1:], substr))
}
