package policies

import (
	"errors"
	"fmt"
	"moviepilot-go/internal/business/domains"
	"time"
)

var (
	ErrDuplicateSubscription = errors.New("duplicate subscription")
	ErrInvalidQuality        = errors.New("invalid quality requirement")
	ErrExceedSubscriptionLimit = errors.New("exceed subscription limit")
)

// SubscriptionPolicy 订阅策略
type SubscriptionPolicy struct {
	MaxSubscriptionsPerUser int
	AllowedQualities       []domains.Quality
	MinQuality             domains.Quality
	MaxQuality             domains.Quality
}

// NewSubscriptionPolicy 创建新的订阅策略
func NewSubscriptionPolicy() *SubscriptionPolicy {
	return &SubscriptionPolicy{
		MaxSubscriptionsPerUser: 100,
		AllowedQualities: []domains.Quality{
			{Resolution: "720p", Source: "HDTV"},
			{Resolution: "720p", Source: "WEB-DL"},
			{Resolution: "1080p", Source: "HDTV"},
			{Resolution: "1080p", Source: "WEB-DL"},
			{Resolution: "1080p", Source: "BluRay"},
			{Resolution: "4K", Source: "WEB-DL"},
			{Resolution: "4K", Source: "BluRay"},
		},
		MinQuality: domains.Quality{Resolution: "720p", Source: "HDTV"},
		MaxQuality: domains.Quality{Resolution: "4K", Source: "BluRay"},
	}
}

// ValidateSubscription 验证订阅是否符合策略
func (p *SubscriptionPolicy) ValidateSubscription(subscription *domains.Subscribe, existingSubscriptions []domains.Subscribe) error {
	// 检查重复订阅
	for _, existing := range existingSubscriptions {
		if existing.MediaID == subscription.MediaID && 
		   existing.Season == subscription.Season && 
		   existing.IsActive {
			return ErrDuplicateSubscription
		}
	}
	
	// 检查订阅数量限制
	activeCount := 0
	for _, existing := range existingSubscriptions {
		if existing.IsActive {
			activeCount++
		}
	}
	
	if activeCount >= p.MaxSubscriptionsPerUser {
		return ErrExceedSubscriptionLimit
	}
	
	// 验证质量要求
	if err := p.validateQuality(subscription.Quality); err != nil {
		return err
	}
	
	// 验证电视剧订阅的特殊规则
	if subscription.IsTVSeries() {
		if subscription.Season < 0 || subscription.Episode < 0 {
			return errors.New("invalid season or episode number")
		}
	}
	
	return nil
}

// validateQuality 验证质量要求
func (p *SubscriptionPolicy) validateQuality(qualityStr string) error {
	if qualityStr == "" {
		return nil // 未指定质量，使用默认策略
	}
	
	// 这里需要解析质量字符串并与允许的质量进行比较
	// 简化实现，实际应该有完整的质量解析逻辑
	
	return nil
}

// GetRecommendedQuality 获取推荐质量
func (p *SubscriptionPolicy) GetRecommendedQuality(userPreferences map[string]interface{}) domains.Quality {
	// 根据用户偏好和策略推荐质量
	// 简化实现，返回默认推荐质量
	return domains.Quality{
		Resolution: "1080p",
		Source:     "WEB-DL",
		Codec:      "H.264",
	}
}

// ShouldAutoDownload 判断是否应该自动下载
func (p *SubscriptionPolicy) ShouldAutoDownload(subscription *domains.Subscribe, availableQuality domains.Quality) bool {
	// 检查质量是否符合要求
	if subscription.Quality != "" {
		// 如果订阅指定了质量要求，检查是否匹配
		// 这里应该实现质量匹配逻辑
	}
	
	// 检查是否在允许的质量范围内
	for _, allowed := range p.AllowedQualities {
		if availableQuality.Resolution == allowed.Resolution && 
		   availableQuality.Source == allowed.Source {
			return true
		}
	}
	
	return false
}

// GetNextEpisode 获取下一集信息
func (p *SubscriptionPolicy) GetNextEpisode(subscription *domains.Subscribe, currentEpisodes []int) (int, int, error) {
	if !subscription.IsTVSeries() {
		return 0, 0, errors.New("not a TV series subscription")
	}
	
	// 如果是整季订阅，返回下一季
	if subscription.Episode == 0 {
		nextSeason := subscription.Season + 1
		return nextSeason, 0, nil
	}
	
	// 如果是单集订阅，返回下一集
	nextEpisode := subscription.Episode + 1
	return subscription.Season, nextEpisode, nil
}

// CalculateRetryDelay 计算重试延迟
func (p *SubscriptionPolicy) CalculateRetryDelay(attemptCount int, lastFailure time.Time) time.Duration {
	// 指数退避策略
	baseDelay := time.Hour
	maxDelay := 24 * time.Hour
	
	delay := time.Duration(attemptCount) * baseDelay
	if delay > maxDelay {
		delay = maxDelay
	}
	
	// 考虑上次失败时间
	timeSinceFailure := time.Since(lastFailure)
	if timeSinceFailure < delay {
		delay = delay - timeSinceFailure
	} else {
		delay = time.Minute // 立即重试
	}
	
	return delay
}

// IsSubscriptionExpired 检查订阅是否过期
func (p *SubscriptionPolicy) IsSubscriptionExpired(subscription *domains.Subscribe) bool {
	// 简化实现：订阅超过30天未更新视为过期
	return time.Since(subscription.UpdatedAt) > 30*24*time.Hour
}

// RenewSubscription 续订订阅
func (p *SubscriptionPolicy) RenewSubscription(subscription *domains.Subscribe) error {
	if p.IsSubscriptionExpired(subscription) {
		return errors.New("subscription is too old to renew")
	}
	
	// 更新订阅时间
	subscription.UpdatedAt = time.Now()
	subscription.IsActive = true
	
	return nil
}

// GetSubscriptionStats 获取订阅统计信息
func (p *SubscriptionPolicy) GetSubscriptionStats(subscriptions []domains.Subscribe) map[string]interface{} {
	stats := make(map[string]interface{})
	
	activeCount := 0
	movieCount := 0
	tvCount := 0
	
	for _, sub := range subscriptions {
		if sub.IsActive {
			activeCount++
		}
		
		if sub.IsTVSeries() {
			tvCount++
		} else {
			movieCount++
		}
	}
	
	stats["total"] = len(subscriptions)
	stats["active"] = activeCount
	stats["movies"] = movieCount
	stats["tv_series"] = tvCount
	stats["inactive"] = len(subscriptions) - activeCount
	
	return stats
}