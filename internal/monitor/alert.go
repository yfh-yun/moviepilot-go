package monitor

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// AlertManager 告警管理器
type AlertManager struct {
	rules    []AlertRule
	logger   *zap.Logger
	mu       sync.RWMutex
	active   map[string]*Alert
	history  []Alert
}

// NewAlertManager 创建告警管理器
func NewAlertManager(rules []AlertRule, logger *zap.Logger) *AlertManager {
	return &AlertManager{
		rules:   rules,
		logger:  logger,
		active:  make(map[string]*Alert),
		history: make([]Alert, 0),
	}
}

// CheckAlerts 检查告警
func (am *AlertManager) CheckAlerts(metrics map[string]float64) []Alert {
	am.mu.Lock()
	defer am.mu.Unlock()
	
	var triggeredAlerts []Alert
	now := time.Now()
	
	for _, rule := range am.rules {
		if !rule.Enabled {
			continue
		}
		
		value, exists := metrics[rule.Metric]
		if !exists {
			continue
		}
		
		alertKey := fmt.Sprintf("%s:%s", rule.Metric, rule.ID)
		
		// 检查告警条件
		triggered := am.evaluateCondition(value, rule.Operator, rule.Threshold)
		
		if triggered {
			// 告警触发
			if alert, exists := am.active[alertKey]; exists {
				// 更新现有告警
				alert.Value = value
				alert.StartTime = now
			} else {
				// 创建新告警
				alert := &Alert{
					ID:          am.generateAlertID(),
					RuleID:      rule.ID,
					RuleName:    rule.Name,
					Level:       rule.Level,
					Message:     am.buildAlertMessage(rule, value),
					Value:       value,
					Threshold:   rule.Threshold,
					Metric:      rule.Metric,
					Tags:        map[string]string{},
					StartTime:   now,
					Status:      "firing",
					Annotations: map[string]string{"description": rule.Description},
				}
				
				am.active[alertKey] = alert
				triggeredAlerts = append(triggeredAlerts, *alert)
				
				am.logger.Warn("告警触发",
					zap.String("rule", rule.Name),
					zap.String("metric", rule.Metric),
					zap.Float64("value", value),
					zap.Float64("threshold", rule.Threshold),
					zap.String("level", string(rule.Level)))
			}
		} else {
			// 告警恢复
			if alert, exists := am.active[alertKey]; exists {
				endTime := now
				alert.EndTime = &endTime
				alert.Status = "resolved"
				
				// 移除活跃告警并添加到历史记录
				delete(am.active, alertKey)
				am.history = append(am.history, *alert)
				
				// 限制历史记录数量
				if len(am.history) > 1000 {
					am.history = am.history[1:]
				}
				
				am.logger.Info("告警恢复",
					zap.String("rule", alert.RuleName),
					zap.String("metric", alert.Metric),
					zap.Duration("duration", endTime.Sub(alert.StartTime)))
			}
		}
	}
	
	return triggeredAlerts
}

// evaluateCondition 评估条件
func (am *AlertManager) evaluateCondition(value float64, operator string, threshold float64) bool {
	switch operator {
	case ">":
		return value > threshold
	case ">=":
		return value >= threshold
	case "<":
		return value < threshold
	case "<=":
		return value <= threshold
	case "==":
		return math.Abs(value-threshold) < 0.0001
	case "!=":
		return math.Abs(value-threshold) >= 0.0001
	default:
		return false
	}
}

// buildAlertMessage 构建告警消息
func (am *AlertManager) buildAlertMessage(rule AlertRule, value float64) string {
	var operatorText string
	switch rule.Operator {
	case ">":
		operatorText = "大于"
	case ">=":
		operatorText = "大于等于"
	case "<":
		operatorText = "小于"
	case "<=":
		operatorText = "小于等于"
	case "==":
		operatorText = "等于"
	case "!=":
		operatorText = "不等于"
	default:
		operatorText = rule.Operator
	}
	
	return fmt.Sprintf("指标 %s 当前值 %.2f %s 阈值 %.2f", 
		rule.Metric, value, operatorText, rule.Threshold)
}

// generateAlertID 生成告警ID
func (am *AlertManager) generateAlertID() string {
	return fmt.Sprintf("alert_%d", time.Now().UnixNano())
}

// GetActiveAlerts 获取活跃告警
func (am *AlertManager) GetActiveAlerts() []Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()
	
	var alerts []Alert
	for _, alert := range am.active {
		alerts = append(alerts, *alert)
	}
	return alerts
}

// GetAlertHistory 获取告警历史
func (am *AlertManager) GetAlertHistory(limit int) []Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()
	
	if limit <= 0 || limit > len(am.history) {
		limit = len(am.history)
	}
	
	start := len(am.history) - limit
	if start < 0 {
		start = 0
	}
	
	// 返回副本
	history := make([]Alert, limit)
	copy(history, am.history[start:])
	
	return history
}

// UpdateRules 更新告警规则
func (am *AlertManager) UpdateRules(rules []AlertRule) {
	am.mu.Lock()
	defer am.mu.Unlock()
	
	am.rules = rules
	
	// 清除已删除规则的活跃告警
	activeRules := make(map[string]bool)
	for _, rule := range rules {
		activeRules[rule.ID] = true
	}
	
	for key, alert := range am.active {
		if !activeRules[alert.RuleID] {
			endTime := time.Now()
			alert.EndTime = &endTime
			alert.Status = "resolved"
			am.history = append(am.history, *alert)
			delete(am.active, key)
			
			am.logger.Info("告警规则已删除，标记告警为恢复状态",
				zap.String("rule", alert.RuleName))
		}
	}
}

// GetRules 获取告警规则
func (am *AlertManager) GetRules() []AlertRule {
	am.mu.RLock()
	defer am.mu.RUnlock()
	
	rules := make([]AlertRule, len(am.rules))
	copy(rules, am.rules)
	return rules
}

// AcknowledgeAlert 确认告警
func (am *AlertManager) AcknowledgeAlert(alertID string) error {
	am.mu.Lock()
	defer am.mu.Unlock()
	
	for key, alert := range am.active {
		if alert.ID == alertID {
			alert.Annotations["acknowledged"] = "true"
			alert.Annotations["acknowledged_at"] = time.Now().Format(time.RFC3339)
			am.active[key] = alert
			
			am.logger.Info("告警已确认", zap.String("alert_id", alertID))
			return nil
		}
	}
	
	return fmt.Errorf("未找到告警: %s", alertID)
}

// GetAlertStatistics 获取告警统计
func (am *AlertManager) GetAlertStatistics() map[string]interface{} {
	am.mu.RLock()
	defer am.mu.RUnlock()
	
	stats := map[string]interface{}{
		"active_count":   len(am.active),
		"history_count":  len(am.history),
		"rules_count":    len(am.rules),
		"by_level":       make(map[string]int),
		"by_status":      make(map[string]int),
		"by_metric":      make(map[string]int),
	}
	
	// 统计活跃告警
	for _, alert := range am.active {
		stats["by_level"].(map[string]int)[string(alert.Level)]++
		stats["by_status"].(map[string]int)[alert.Status]++
		stats["by_metric"].(map[string]int)[alert.Metric]++
	}
	
	// 统计历史告警（最近100条）
	recentCount := 100
	if len(am.history) < recentCount {
		recentCount = len(am.history)
	}
	
	for i := len(am.history) - recentCount; i < len(am.history); i++ {
		alert := am.history[i]
		stats["by_level"].(map[string]int)[string(alert.Level)]++
		stats["by_status"].(map[string]int)[alert.Status]++
		stats["by_metric"].(map[string]int)[alert.Metric]++
	}
	
	return stats
}

// ValidateRule 验证告警规则
func (am *AlertManager) ValidateRule(rule AlertRule) error {
	if strings.TrimSpace(rule.Name) == "" {
		return fmt.Errorf("规则名称不能为空")
	}
	
	if strings.TrimSpace(rule.Metric) == "" {
		return fmt.Errorf("指标名称不能为空")
	}
	
	validOperators := []string{">", ">=", "<", "<=", "==", "!="}
	invalidOperator := true
	for _, op := range validOperators {
		if rule.Operator == op {
			invalidOperator = false
			break
		}
	}
	if invalidOperator {
		return fmt.Errorf("无效的操作符: %s", rule.Operator)
	}
	
	if rule.Duration < 0 {
		return fmt.Errorf("持续时间不能为负数")
	}
	
	validLevels := []string{string(AlertLevelInfo), string(AlertLevelWarning), string(AlertLevelError), string(AlertLevelCritical)}
	invalidLevel := true
	for _, level := range validLevels {
		if string(rule.Level) == level {
			invalidLevel = false
			break
		}
	}
	if invalidLevel {
		return fmt.Errorf("无效的告警级别: %s", rule.Level)
	}
	
	return nil
}