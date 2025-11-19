package notification

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// NotificationConfig 通知系统配置
type NotificationConfig struct {
	Enabled      bool                      `json:"enabled"`       // 是否启用通知系统
	DefaultLevel NotificationLevel         `json:"default_level"` // 默认通知级别
	Channels     map[string]*ChannelConfig `json:"channels"`      // 渠道配置
	RoutingRules []RoutingRule             `json:"routing_rules"` // 路由规则
}

// ChannelConfig 渠道配置
type ChannelConfig struct {
	Name       string `json:"name"`        // 渠道名称
	Enabled    bool   `json:"enabled"`     // 是否启用
	Priority   int    `json:"priority"`    // 优先级（1-10，数字越大优先级越高）
	RetryCount int    `json:"retry_count"` // 重试次数
	RetryDelay int    `json:"retry_delay"` // 重试延迟（秒）
}

// RoutingRule 路由规则
type RoutingRule struct {
	Name     string            `json:"name"`     // 规则名称
	Enabled  bool              `json:"enabled"`  // 是否启用
	Patterns []string          `json:"patterns"` // 匹配模式（支持通配符）
	Level    NotificationLevel `json:"level"`    // 目标级别
	Channels []string          `json:"channels"` // 目标渠道
	Weight   int               `json:"weight"`   // 权重（用于规则优先级）
}

// DefaultNotificationConfig 默认通知配置
func DefaultNotificationConfig() *NotificationConfig {
	return &NotificationConfig{
		Enabled:      true,
		DefaultLevel: LevelInfo,
		Channels: map[string]*ChannelConfig{
			"wechat": {
				Name:       "wechat",
				Enabled:    false,
				Priority:   5,
				RetryCount: 3,
				RetryDelay: 5,
			},
			"wechat_work": {
				Name:       "wechat_work",
				Enabled:    false,
				Priority:   5,
				RetryCount: 3,
				RetryDelay: 5,
			},
			"telegram": {
				Name:       "telegram",
				Enabled:    false,
				Priority:   4,
				RetryCount: 3,
				RetryDelay: 5,
			},
			"dingtalk": {
				Name:       "dingtalk",
				Enabled:    false,
				Priority:   4,
				RetryCount: 3,
				RetryDelay: 5,
			},
			"feishu": {
				Name:       "feishu",
				Enabled:    false,
				Priority:   4,
				RetryCount: 3,
				RetryDelay: 5,
			},
			"email": {
				Name:       "email",
				Enabled:    false,
				Priority:   3,
				RetryCount: 2,
				RetryDelay: 10,
			},
			"webpush": {
				Name:       "webpush",
				Enabled:    false,
				Priority:   2,
				RetryCount: 3,
				RetryDelay: 5,
			},
		},
		RoutingRules: []RoutingRule{
			{
				Name:     "error-high-priority",
				Enabled:  true,
				Patterns: []string{"*error*", "*failed*", "*critical*"},
				Level:    LevelError,
				Channels: []string{"wechat", "telegram", "email"},
				Weight:   100,
			},
			{
				Name:     "warning-medium-priority",
				Enabled:  true,
				Patterns: []string{"*warning*", "*alert*", "*issue*"},
				Level:    LevelWarning,
				Channels: []string{"wechat", "email"},
				Weight:   50,
			},
			{
				Name:     "success-low-priority",
				Enabled:  true,
				Patterns: []string{"*success*", "*completed*", "*finished*"},
				Level:    LevelSuccess,
				Channels: []string{"email"},
				Weight:   25,
			},
		},
	}
}

// LoadNotificationConfig 从文件加载通知配置
func LoadNotificationConfig(configPath string) (*NotificationConfig, error) {
	// 如果文件不存在，使用默认配置
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultNotificationConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config NotificationConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// 确保所有必需字段都有默认值
	if config.DefaultLevel == 0 {
		config.DefaultLevel = LevelInfo
	}

	if config.Channels == nil {
		config.Channels = make(map[string]*ChannelConfig)
	}

	return &config, nil
}

// SaveNotificationConfig 保存通知配置到文件
func SaveNotificationConfig(config *NotificationConfig, configPath string) error {
	// 确保目录存在
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ApplyRoutingRules 应用路由规则到消息
func (c *NotificationConfig) ApplyRoutingRules(message *Message) {
	// 按照权重从高到低排序规则
	rules := make([]RoutingRule, len(c.RoutingRules))
	copy(rules, c.RoutingRules)

	// 简单的冒泡排序（权重从小到大，然后反转）
	for i := 0; i < len(rules)-1; i++ {
		for j := 0; j < len(rules)-i-1; j++ {
			if rules[j].Weight > rules[j+1].Weight {
				rules[j], rules[j+1] = rules[j+1], rules[j]
			}
		}
	}

	// 反转以获得从高到低的顺序
	for i, j := 0, len(rules)-1; i < j; i, j = i+1, j-1 {
		rules[i], rules[j] = rules[j], rules[i]
	}

	// 应用规则
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		// 检查是否匹配任何模式
		matched := false
		text := strings.ToLower(message.Title + " " + message.Content)

		for _, pattern := range rule.Patterns {
			if wildcardMatch(pattern, text) {
				matched = true
				break
			}
		}

		if matched {
			// 更新消息级别和目标渠道
			message.Level = rule.Level
			// 这里只是设置规则匹配，实际发送渠道由管理器决定
			break
		}
	}
}

// wildcardMatch 简单的通配符匹配（支持 * 和 ?）
func wildcardMatch(pattern, text string) bool {
	if pattern == "*" {
		return true
	}

	// 简单的通配符匹配实现
	pattern = strings.ToLower(pattern)
	text = strings.ToLower(text)

	if pattern == "*" {
		return true
	}

	if strings.Contains(pattern, "*") {
		parts := strings.Split(pattern, "*")
		for i, part := range parts {
			if part == "" {
				continue
			}

			if i == 0 {
				// 第一部分必须在开头
				if !strings.HasPrefix(text, part) {
					return false
				}
				text = text[len(part):]
			} else if i == len(parts)-1 {
				// 最后一部分必须在结尾
				if !strings.HasSuffix(text, part) {
					return false
				}
			} else {
				// 中间部分必须在文本中的某个位置
				index := strings.Index(text, part)
				if index == -1 {
					return false
				}
				text = text[index+len(part):]
			}
		}
		return true
	}

	// 精确匹配
	return pattern == text
}

// GetEnabledChannels 获取启用的渠道列表
func (c *NotificationConfig) GetEnabledChannels() []string {
	var enabledChannels []string

	for name, channel := range c.Channels {
		if channel.Enabled {
			enabledChannels = append(enabledChannels, name)
		}
	}

	return enabledChannels
}

// GetChannelConfig 获取渠道配置
func (c *NotificationConfig) GetChannelConfig(name string) (*ChannelConfig, bool) {
	config, exists := c.Channels[name]
	return config, exists
}

// ValidateConfig 验证配置
func (c *NotificationConfig) ValidateConfig() error {
	if !c.Enabled {
		return nil // 如果禁用，配置无效也无所谓
	}

	// 检查默认级别
	if c.DefaultLevel < LevelInfo || c.DefaultLevel > LevelSuccess {
		return fmt.Errorf("invalid default level: %d", c.DefaultLevel)
	}

	// 检查路由规则
	for _, rule := range c.RoutingRules {
		if rule.Name == "" {
			return fmt.Errorf("routing rule name cannot be empty")
		}

		if len(rule.Patterns) == 0 {
			return fmt.Errorf("routing rule %s must have at least one pattern", rule.Name)
		}

		if rule.Level < LevelInfo || rule.Level > LevelSuccess {
			return fmt.Errorf("invalid level in routing rule %s: %d", rule.Name, rule.Level)
		}
	}

	return nil
}
