package policies

import (
	"moviepilot-go/internal/business/domains"
)

// PolicyManager 策略管理器
type PolicyManager struct {
	SubscriptionPolicy *SubscriptionPolicy
	QualityPolicy      *QualityPolicy
	DownloadPolicy     *DownloadPolicy
}

// NewPolicyManager 创建策略管理器
func NewPolicyManager() *PolicyManager {
	return &PolicyManager{
		SubscriptionPolicy: NewSubscriptionPolicy(),
		QualityPolicy:      NewQualityPolicy(),
		DownloadPolicy:     NewDownloadPolicy(),
	}
}

// GetSubscriptionPolicy 获取订阅策略
func (pm *PolicyManager) GetSubscriptionPolicy() *SubscriptionPolicy {
	return pm.SubscriptionPolicy
}

// GetQualityPolicy 获取质量策略
func (pm *PolicyManager) GetQualityPolicy() *QualityPolicy {
	return pm.QualityPolicy
}

// GetDownloadPolicy 获取下载策略
func (pm *PolicyManager) GetDownloadPolicy() *DownloadPolicy {
	return pm.DownloadPolicy
}

// ValidateAll 验证所有策略
func (pm *PolicyManager) ValidateAll(subscription *domains.Subscribe, downloadConfig *domains.DownloadConfig) error {
	// 验证订阅策略
	if subscription != nil {
		if err := pm.SubscriptionPolicy.ValidateSubscription(subscription, []domains.Subscribe{}); err != nil {
			return err
		}
	}
	
	// 验证下载策略
	if downloadConfig != nil {
		if err := pm.DownloadPolicy.ValidateDownload(downloadConfig, 0, 0); err != nil {
			return err
		}
	}
	
	return nil
}

// GetRecommendedSettings 获取推荐设置
func (pm *PolicyManager) GetRecommendedSettings(userPreferences map[string]interface{}) map[string]interface{} {
	settings := make(map[string]interface{})
	
	// 推荐质量
	settings["recommended_quality"] = pm.QualityPolicy.SuggestQuality(userPreferences)
	
	// 推荐下载配置
	settings["recommended_download_config"] = domains.DefaultDownloadConfig()
	
	// 推荐订阅设置
	settings["max_subscriptions"] = pm.SubscriptionPolicy.MaxSubscriptionsPerUser
	
	return settings
}