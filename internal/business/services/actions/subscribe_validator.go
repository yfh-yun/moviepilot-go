// Package actions 提供订阅管理的参数验证器实现
package actions

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"moviepilot-go/pkg/logger"
)

// SubscribeValidator 订阅参数验证器
type SubscribeValidator struct {
	logger logger.Logger
}

// ValidationError 验证错误结构体
type ValidationError struct {
	Field   string
	Message string
}

// Error 实现error接口
func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// NewSubscribeValidator 创建订阅验证器实例
func NewSubscribeValidator() *SubscribeValidator {
	return &SubscribeValidator{
		logger: logger.NewLogger("subscribe_validator"),
	}
}

// ValidateAddSubscribe 验证添加订阅参数
func (v *SubscribeValidator) ValidateAddSubscribe(params *AddSubscribeParams) error {
	if params == nil {
		return errors.New("订阅参数不能为空")
	}

	// 验证必填字段
	if err := v.validateRequiredFields(params); err != nil {
		return err
	}

	// 验证订阅类型
	if err := v.validateSubscribeType(params.Type); err != nil {
		return err
	}

	// 验证订阅配置
	if err := v.validateSubscribeConfig(params.Type, params.Config); err != nil {
		return err
	}

	// 验证下载配置
	if err := v.validateDownloadConfig(params.DownloadConfig); err != nil {
		return err
	}

	// 验证过滤配置
	if err := v.validateFilterConfig(params.FilterConfig); err != nil {
		return err
	}

	// 验证更新周期
	if err := v.validateUpdateInterval(params.UpdateInterval); err != nil {
		return err
	}

	// 验证规则配置
	if err := v.validateRules(params.Rules); err != nil {
		return err
	}

	// 验证通知配置
	if err := v.validateNotificationConfig(params.NotifyConfig); err != nil {
		return err
	}

	return nil
}

// ValidateUpdateSubscribe 验证更新订阅参数
func (v *SubscribeValidator) ValidateUpdateSubscribe(params *UpdateSubscribeParams) error {
	if params == nil {
		return errors.New("订阅更新参数不能为空")
	}

	// 如果指定了订阅类型，验证类型
	if params.Type != "" {
		if err := v.validateSubscribeType(params.Type); err != nil {
			return err
		}

		// 如果同时提供了配置，验证配置
		if params.Config != nil {
			if err := v.validateSubscribeConfig(params.Type, params.Config); err != nil {
				return err
			}
		}
	}

	// 验证下载配置
	if params.DownloadConfig != nil {
		if err := v.validateDownloadConfig(params.DownloadConfig); err != nil {
			return err
		}
	}

	// 验证过滤配置
	if params.FilterConfig != nil {
		if err := v.validateFilterConfig(params.FilterConfig); err != nil {
			return err
		}
	}

	// 验证更新周期
	if params.UpdateInterval != nil {
		if err := v.validateUpdateInterval(*params.UpdateInterval); err != nil {
			return err
		}
	}

	// 验证规则配置
	if params.Rules != nil {
		if err := v.validateRules(params.Rules); err != nil {
			return err
		}
	}

	// 验证通知配置
	if params.NotifyConfig != nil {
		if err := v.validateNotificationConfig(params.NotifyConfig); err != nil {
			return err
		}
	}

	return nil
}

// validateRequiredFields 验证必填字段
func (v *SubscribeValidator) validateRequiredFields(params *AddSubscribeParams) error {
	if strings.TrimSpace(params.Name) == "" {
		return &ValidationError{Field: "name", Message: "订阅名称不能为空"}
	}

	if strings.TrimSpace(string(params.Type)) == "" {
		return &ValidationError{Field: "type", Message: "订阅类型不能为空"}
	}

	if params.Config == nil {
		return &ValidationError{Field: "config", Message: "订阅配置不能为空"}
	}

	return nil
}

// validateSubscribeType 验证订阅类型
func (v *SubscribeValidator) validateSubscribeType(subscribeType SubscribeType) error {
	validTypes := []SubscribeType{
		SubscribeTypeRSS,
		SubscribeTypeTorrent,
		SubscribeTypeMedia,
		SubscribeTypeKeyword,
		SubscribeTypeCustom,
	}

	for _, validType := range validTypes {
		if subscribeType == validType {
			return nil
		}
	}

	return &ValidationError{Field: "type", Message: "无效的订阅类型"}
}

// validateSubscribeConfig 验证订阅配置
func (v *SubscribeValidator) validateSubscribeConfig(subscribeType SubscribeType, config map[string]interface{}) error {
	if config == nil {
		return &ValidationError{Field: "config", Message: "订阅配置不能为空"}
	}

	switch subscribeType {
	case SubscribeTypeRSS:
		if _, ok := config["url"]; !ok {
			return &ValidationError{Field: "config.url", Message: "RSS URL 不能为空"}
		}

		url, ok := config["url"].(string)
		if !ok || !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			return &ValidationError{Field: "config.url", Message: "RSS URL 格式无效"}
		}

	case SubscribeTypeTorrent:
		if _, ok := config["torrent_url"]; !ok {
			return &ValidationError{Field: "config.torrent_url", Message: "种子URL不能为空"}
		}

	case SubscribeTypeMedia:
		if _, ok := config["media_id"]; !ok {
			return &ValidationError{Field: "config.media_id", Message: "媒体ID不能为空"}
		}

	case SubscribeTypeKeyword:
		if _, ok := config["keyword"]; !ok {
			return &ValidationError{Field: "config.keyword", Message: "关键词不能为空"}
		}

		keyword, ok := config["keyword"].(string)
		if !ok || strings.TrimSpace(keyword) == "" {
			return &ValidationError{Field: "config.keyword", Message: "关键词不能为空或空白"}
		}

	case SubscribeTypeCustom:
		if _, ok := config["custom_handler"]; !ok {
			return &ValidationError{Field: "config.custom_handler", Message: "自定义处理器名称不能为空"}
		}
	}

	return nil
}

// validateDownloadConfig 验证下载配置
func (v *SubscribeValidator) validateDownloadConfig(config *DownloadConfig) error {
	if config == nil {
		return nil // 下载配置可选
	}

	// 验证下载器类型
	if config.Downloader != "" {
		validDownloaders := []string{"qBittorrent", "Transmission", "Aria2", "NZBGet"}
		found := false
		for _, validDownloader := range validDownloaders {
			if config.Downloader == validDownloader {
				found = true
				break
			}
		}

		if !found {
			return &ValidationError{Field: "download_config.downloader", Message: "无效的下载器类型"}
		}
	}

	// 验证保存路径
	if config.SavePath != "" && !strings.HasPrefix(config.SavePath, "/") {
		return &ValidationError{Field: "download_config.save_path", Message: "保存路径必须是绝对路径"}
	}

	// 验证下载速度限制
	if config.MaxSpeed > 0 && config.MaxSpeed < 1024 { // 至少1KB/s
		return &ValidationError{Field: "download_config.max_speed", Message: "下载速度限制过小"}
	}

	return nil
}

// validateFilterConfig 验证过滤配置
func (v *SubscribeValidator) validateFilterConfig(config *FilterConfig) error {
	if config == nil {
		return nil // 过滤配置可选
	}

	// 验证质量过滤
	if len(config.QualityFilters) > 0 {
		validQualities := []string{"1080p", "720p", "2160p", "4K", "SD", "HD"}
		for _, quality := range config.QualityFilters {
			found := false
			for _, validQuality := range validQualities {
				if quality == validQuality {
					found = true
					break
				}
			}

			if !found {
				return &ValidationError{Field: "filter_config.quality_filters", Message: "无效的质量过滤值: " + quality}
			}
		}
	}

	// 验证大小过滤
	if config.MinSize > 0 && config.MaxSize > 0 && config.MinSize > config.MaxSize {
		return &ValidationError{Field: "filter_config", Message: "最小大小不能大于最大大小"}
	}

	// 验证评分过滤
	if config.MinRating > 0 && (config.MinRating < 0 || config.MinRating > 10) {
		return &ValidationError{Field: "filter_config.min_rating", Message: "评分必须在0-10之间"}
	}

	return nil
}

// validateUpdateInterval 验证更新周期
func (v *SubscribeValidator) validateUpdateInterval(interval int) error {
	if interval <= 0 {
		return &ValidationError{Field: "update_interval", Message: "更新周期必须大于0"}
	}

	if interval < 5 {
		return &ValidationError{Field: "update_interval", Message: "更新周期不能小于5分钟"}
	}

	if interval > 10080 { // 7天
		return &ValidationError{Field: "update_interval", Message: "更新周期不能超过7天"}
	}

	return nil
}

// validateRules 验证规则配置
func (v *SubscribeValidator) validateRules(rules []Rule) error {
	if rules == nil {
		return nil // 规则配置可选
	}

	for i, rule := range rules {
		field := fmt.Sprintf("rules[%d]", i)

		// 验证规则类型
		if rule.Type != "include" && rule.Type != "exclude" {
			return &ValidationError{Field: field + ".type", Message: "规则类型必须是include或exclude"}
		}

		// 验证规则字段
		if rule.Field == "" {
			return &ValidationError{Field: field + ".field", Message: "规则字段不能为空"}
		}

		// 验证规则值
		if rule.Value == "" {
			return &ValidationError{Field: field + ".value", Message: "规则值不能为空"}
		}

		// 验证规则匹配方式
		validModes := []string{"equals", "contains", "startswith", "endswith", "regex"}
		found := false
		for _, mode := range validModes {
			if rule.Mode == mode {
				found = true
				break
			}
		}

		if !found {
			return &ValidationError{Field: field + ".mode", Message: "无效的规则匹配方式"}
		}
	}

	return nil
}

// validateNotificationConfig 验证通知配置
func (v *SubscribeValidator) validateNotificationConfig(config *NotificationConfig) error {
	if config == nil {
		return nil // 通知配置可选
	}

	// 验证通知类型
	if len(config.EnabledTypes) > 0 {
		validTypes := []string{"new", "download", "error", "all"}
		for _, notifyType := range config.EnabledTypes {
			found := false
			for _, validType := range validTypes {
				if notifyType == validType {
					found = true
					break
				}
			}

			if !found {
				return &ValidationError{Field: "notify_config.enabled_types", Message: "无效的通知类型: " + notifyType}
			}
		}
	}

	return nil
}

// ValidateSubscribeFilter 验证订阅列表过滤参数
func (v *SubscribeValidator) ValidateSubscribeFilter(filter *SubscribeFilter) error {
	if filter == nil {
		return errors.New("过滤参数不能为空")
	}

	// 设置默认值
	if filter.Limit <= 0 {
		filter.Limit = 20
	}

	if filter.Limit > 100 {
		filter.Limit = 100 // 最大100
	}

	if filter.Offset < 0 {
		filter.Offset = 0
	}

	// 验证订阅类型
	for _, subscribeType := range filter.Types {
		if err := v.validateSubscribeType(subscribeType); err != nil {
			return err
		}
	}

	// 验证订阅状态
	validStatuses := []SubscribeStatus{
		SubscribeStatusActive,
		SubscribeStatusPaused,
		SubscribeStatusError,
		SubscribeStatusCompleted,
	}

	for _, status := range filter.Statuses {
		found := false
		for _, validStatus := range validStatuses {
			if status == validStatus {
				found = true
				break
			}
		}

		if !found {
			return &ValidationError{Field: "status", Message: "无效的订阅状态: " + string(status)}
		}
	}

	return nil
}

// ValidateSubscribeItemFilter 验证订阅项列表过滤参数
func (v *SubscribeValidator) ValidateSubscribeItemFilter(filter *SubscribeItemFilter) error {
	if filter == nil {
		return errors.New("过滤参数不能为空")
	}

	// 设置默认值
	if filter.Limit <= 0 {
		filter.Limit = 50
	}

	if filter.Limit > 200 {
		filter.Limit = 200 // 最大200
	}

	if filter.Offset < 0 {
		filter.Offset = 0
	}

	// 验证排序字段
	if filter.OrderBy == "" {
		filter.OrderBy = "created_at"
	}

	validOrderFields := []string{"created_at", "updated_at", "title", "size", "downloaded"}
	found := false
	for _, field := range validOrderFields {
		if filter.OrderBy == field {
			found = true
			break
		}
	}

	if !found {
		return &ValidationError{Field: "order_by", Message: "无效的排序字段"}
	}

	// 验证排序方向
	if filter.OrderDir == "" {
		filter.OrderDir = "desc"
	}

	if filter.OrderDir != "asc" && filter.OrderDir != "desc" {
		return &ValidationError{Field: "order_dir", Message: "排序方向必须是asc或desc"}
	}

	return nil
}
