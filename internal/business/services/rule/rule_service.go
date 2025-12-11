package rule

import (
	"context"

	"moviepilot-go/internal/business/services/base"
	"moviepilot-go/pkg/logger"
	ruleutil "moviepilot-go/pkg/utils"

	"go.uber.org/zap"
)

// RuleService 规则服务
type RuleService struct {
	*base.ServiceBase
	ruleHelper *ruleutil.RuleHelper
}

// NewRuleService 创建RuleService实例
func NewRuleService() *RuleService {
	return &RuleService{
		ServiceBase: base.NewServiceBase(),
		ruleHelper:  ruleutil.NewRuleHelper(),
	}
}

// Initialize 初始化服务
func (s *RuleService) Initialize() error {
	logger.Info("Initializing RuleService")
	return nil
}

// Name 获取服务名称
func (s *RuleService) Name() string {
	return "RuleService"
}

// Close 关闭服务
func (s *RuleService) Close() error {
	logger.Info("Closing RuleService")
	return nil
}

// CreateRuleGroup 创建规则组
func (s *RuleService) CreateRuleGroup(ctx context.Context, group *ruleutil.FilterRuleGroup) error {
	logger.Info("Creating rule group",
		zap.String("name", group.Name),
		zap.String("media_type", string(group.MediaType)))

	// 添加规则组
	if err := s.ruleHelper.AddRuleGroup(group); err != nil {
		logger.Error("Failed to create rule group", zap.Error(err))
		return err
	}

	logger.Info("Rule group created successfully", zap.String("name", group.Name))
	return nil
}

// UpdateRuleGroup 更新规则组
func (s *RuleService) UpdateRuleGroup(ctx context.Context, groupName string, updates map[string]any) error {
	logger.Info("Updating rule group",
		zap.String("group_name", groupName),
		zap.Any("updates", updates))

	if err := s.ruleHelper.UpdateRuleGroup(groupName, updates); err != nil {
		logger.Error("Failed to update rule group", zap.Error(err))
		return err
	}

	logger.Info("Rule group updated successfully", zap.String("group_name", groupName))
	return nil
}

// DeleteRuleGroup 删除规则组
func (s *RuleService) DeleteRuleGroup(ctx context.Context, groupName string) error {
	logger.Info("Deleting rule group", zap.String("group_name", groupName))

	if err := s.ruleHelper.RemoveRuleGroup(groupName); err != nil {
		logger.Error("Failed to delete rule group", zap.Error(err))
		return err
	}

	logger.Info("Rule group deleted successfully", zap.String("group_name", groupName))
	return nil
}

// GetRuleGroup 获取规则组
func (s *RuleService) GetRuleGroup(ctx context.Context, groupName string) (*ruleutil.FilterRuleGroup, error) {
	logger.Debug("Getting rule group", zap.String("group_name", groupName))

	group, err := s.ruleHelper.GetRuleGroup(groupName)
	if err != nil {
		logger.Error("Failed to get rule group", zap.Error(err))
		return nil, err
	}

	return group, nil
}

// ListRuleGroups 获取规则组列表
func (s *RuleService) ListRuleGroups(ctx context.Context) ([]*ruleutil.FilterRuleGroup, error) {
	logger.Debug("Listing rule groups")

	groups := s.ruleHelper.GetRuleGroups()

	return groups, nil
}

// AddRule 添加规则
func (s *RuleService) AddRule(ctx context.Context, groupName string, rule *ruleutil.FilterRule) error {
	logger.Info("Adding rule",
		zap.String("group_name", groupName),
		zap.String("rule_name", rule.Name))

	// 添加规则
	if err := s.ruleHelper.AddRule(groupName, rule); err != nil {
		logger.Error("Failed to add rule", zap.Error(err))
		return err
	}

	logger.Info("Rule added successfully",
		zap.String("group_name", groupName),
		zap.String("rule_name", rule.Name))
	return nil
}

// UpdateRule 更新规则
func (s *RuleService) UpdateRule(ctx context.Context, groupName string, ruleName string, updates map[string]any) error {
	logger.Info("Updating rule",
		zap.String("group_name", groupName),
		zap.String("rule_name", ruleName))

	if err := s.ruleHelper.UpdateRule(groupName, ruleName, updates); err != nil {
		logger.Error("Failed to update rule", zap.Error(err))
		return err
	}

	logger.Info("Rule updated successfully",
		zap.String("group_name", groupName),
		zap.String("rule_name", ruleName))
	return nil
}

// DeleteRule 删除规则
func (s *RuleService) DeleteRule(ctx context.Context, groupName string, ruleName string) error {
	logger.Info("Deleting rule",
		zap.String("group_name", groupName),
		zap.String("rule_name", ruleName))

	if err := s.ruleHelper.RemoveRule(groupName, ruleName); err != nil {
		logger.Error("Failed to delete rule", zap.Error(err))
		return err
	}

	logger.Info("Rule deleted successfully",
		zap.String("group_name", groupName),
		zap.String("rule_name", ruleName))
	return nil
}

// GetRule 获取规则
func (s *RuleService) GetRule(ctx context.Context, groupName string, ruleName string) (*ruleutil.FilterRule, error) {
	logger.Debug("Getting rule",
		zap.String("group_name", groupName),
		zap.String("rule_name", ruleName))

	rule, err := s.ruleHelper.GetRule(groupName, ruleName)
	if err != nil {
		logger.Error("Failed to get rule", zap.Error(err))
		return nil, err
	}

	return rule, nil
}

// ApplyRules 应用规则
func (s *RuleService) ApplyRules(ctx context.Context, media *ruleutil.MediaInfo, groupNames []string) (bool, []string, error) {
	logger.Info("Applying rules to media",
		zap.String("media_title", media.Title),
		zap.Strings("group_names", groupNames))

	// 应用规则
	allowed, matchedRules := s.ruleHelper.ApplyRules(media, groupNames)

	logger.Info("Rules applied",
		zap.String("media_title", media.Title),
		zap.Bool("allowed", allowed),
		zap.Int("matched_rules_count", len(matchedRules)))

	return allowed, matchedRules, nil
}

// GetRuleStatistics 获取规则统计信息
func (s *RuleService) GetRuleStatistics(ctx context.Context) (map[string]int, error) {
	logger.Debug("Getting rule statistics")

	stats := map[string]int{
		"group_count": s.ruleHelper.GetGroupCount(),
		"rule_count":  s.ruleHelper.GetRuleCount(),
	}

	return stats, nil
}

// ImportRules 导入规则
func (s *RuleService) ImportRules(ctx context.Context, rules map[string]*ruleutil.FilterRuleGroup) error {
	logger.Info("Importing rules", zap.Int("group_count", len(rules)))

	if err := s.ruleHelper.ImportRules(rules); err != nil {
		logger.Error("Failed to import rules", zap.Error(err))
		return err
	}

	logger.Info("Rules imported successfully", zap.Int("group_count", len(rules)))
	return nil
}

// ExportRules 导出规则
func (s *RuleService) ExportRules(ctx context.Context) (map[string]*ruleutil.FilterRuleGroup, error) {
	logger.Debug("Exporting rules")

	rules := s.ruleHelper.ExportRules()

	logger.Info("Rules exported successfully", zap.Int("group_count", len(rules)))
	return rules, nil
}

// ClearRules 清空所有规则
func (s *RuleService) ClearRules(ctx context.Context) error {
	logger.Info("Clearing all rules")

	s.ruleHelper.ClearRules()

	logger.Info("All rules cleared")
	return nil
}
