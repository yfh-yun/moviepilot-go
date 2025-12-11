package adapters

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// SubscribeType 定义订阅类型
type SubscribeType string

// SubscribeType 常量定义
const (
	SubscribeTypeMovie = SubscribeType("movie") // 电影订阅
	SubscribeTypeTV    = SubscribeType("tv")    // 剧集订阅
	SubscribeTypeRSS   = SubscribeType("rss")   // RSS订阅
	SubscribeTypeOther = SubscribeType("other") // 其他类型订阅
)

// SubscribeStatus 定义订阅状态
type SubscribeStatus string

// SubscribeStatus 常量定义
const (
	SubscribeStatusActive    = SubscribeStatus("active")    // 活跃
	SubscribeStatusPaused    = SubscribeStatus("paused")    // 已暂停
	SubscribeStatusCompleted = SubscribeStatus("completed") // 已完成
	SubscribeStatusCancelled = SubscribeStatus("cancelled") // 已取消
	SubscribeStatusError     = SubscribeStatus("error")     // 错误
)

// SubscribeItem 定义订阅项
type SubscribeItem struct {
	ID                  string          `json:"id"`                   // 订阅ID
	Title               string          `json:"title"`                // 订阅标题
	Keyword             string          `json:"keyword"`              // 订阅关键词
	Type                SubscribeType   `json:"type"`                 // 订阅类型
	Status              SubscribeStatus `json:"status"`               // 订阅状态
	LastCheck           time.Time       `json:"last_check"`           // 最后检查时间
	NextCheck           time.Time       `json:"next_check"`           // 下次检查时间
	CheckInterval       time.Duration   `json:"check_interval"`       // 检查间隔
	FilterRules         []FilterRule    `json:"filter_rules"`         // 过滤规则
	AutoDownload        bool            `json:"auto_download"`        // 是否自动下载
	NotificationEnabled bool            `json:"notification_enabled"` // 是否启用通知
	Metadata            map[string]any  `json:"metadata"`             // 元数据
	CreatedAt           time.Time       `json:"created_at"`           // 创建时间
	UpdatedAt           time.Time       `json:"updated_at"`           // 更新时间
}

// FilterRule 定义过滤规则（复用自params包）
type FilterRule struct {
	Field    string `json:"field"`    // 字段名称
	Operator string `json:"operator"` // 操作符
	Value    any    `json:"value"`    // 比较值
}

// SubscribeService 定义订阅服务接口
type SubscribeService interface {
	// CreateSubscribe 创建订阅
	CreateSubscribe(ctx context.Context, subscribe SubscribeItem) (string, error)

	// UpdateSubscribe 更新订阅
	UpdateSubscribe(ctx context.Context, subscribeID string, updates map[string]any) error

	// DeleteSubscribe 删除订阅
	DeleteSubscribe(ctx context.Context, subscribeID string) error

	// GetSubscribe 获取单个订阅
	GetSubscribe(ctx context.Context, subscribeID string) (*SubscribeItem, error)

	// GetSubscribes 获取订阅列表
	GetSubscribes(ctx context.Context, params GetSubscribesParams) ([]SubscribeItem, error)

	// ActivateSubscribe 激活订阅
	ActivateSubscribe(ctx context.Context, subscribeID string) error

	// PauseSubscribe 暂停订阅
	PauseSubscribe(ctx context.Context, subscribeID string) error

	// RefreshSubscribe 刷新订阅
	RefreshSubscribe(ctx context.Context, subscribeID string) (*SubscribeItem, error)

	// TriggerSubscribe 手动触发订阅检查
	TriggerSubscribe(ctx context.Context, subscribeID string) error

	// GetSubscribeTypes 获取支持的订阅类型
	GetSubscribeTypes(ctx context.Context) ([]SubscribeType, error)
}

// GetSubscribesParams 获取订阅列表参数
type GetSubscribesParams struct {
	Type      SubscribeType   `json:"type"`       // 订阅类型过滤
	Status    SubscribeStatus `json:"status"`     // 订阅状态过滤
	Keyword   string          `json:"keyword"`    // 关键词过滤
	Limit     int             `json:"limit"`      // 返回结果数量限制
	Offset    int             `json:"offset"`     // 偏移量
	SortBy    string          `json:"sort_by"`    // 排序字段
	SortOrder string          `json:"sort_order"` // 排序顺序
}

// SubscribeServiceAdapter 订阅服务适配器实现
type SubscribeServiceAdapter struct {
	logger *zap.Logger
	// 实际的订阅服务客户端可以在这里注入
}

// NewSubscribeServiceAdapter 创建新的订阅服务适配器实例
func NewSubscribeServiceAdapter(logger *zap.Logger) *SubscribeServiceAdapter {
	return &SubscribeServiceAdapter{
		logger: logger,
	}
}

// CreateSubscribe 创建订阅
func (a *SubscribeServiceAdapter) CreateSubscribe(ctx context.Context, subscribe SubscribeItem) (string, error) {
	// 实际实现中，这里应该调用核心业务服务的订阅API
	// 这里使用模拟实现，返回一个随机生成的订阅ID
	a.logger.Info("Creating subscribe", zap.String("title", subscribe.Title), zap.String("type", string(subscribe.Type)))
	return "subscribe-" + time.Now().Format("20060102150405"), nil
}

// UpdateSubscribe 更新订阅
func (a *SubscribeServiceAdapter) UpdateSubscribe(ctx context.Context, subscribeID string, updates map[string]any) error {
	// 实际实现中，这里应该调用核心业务服务的订阅API
	// 这里使用模拟实现，返回nil
	a.logger.Info("Updating subscribe", zap.String("subscribe_id", subscribeID))
	return nil
}

// DeleteSubscribe 删除订阅
func (a *SubscribeServiceAdapter) DeleteSubscribe(ctx context.Context, subscribeID string) error {
	// 实际实现中，这里应该调用核心业务服务的订阅API
	// 这里使用模拟实现，返回nil
	a.logger.Info("Deleting subscribe", zap.String("subscribe_id", subscribeID))
	return nil
}

// GetSubscribe 获取单个订阅
func (a *SubscribeServiceAdapter) GetSubscribe(ctx context.Context, subscribeID string) (*SubscribeItem, error) {
	// 实际实现中，这里应该调用核心业务服务的订阅API
	// 这里使用模拟实现，返回nil
	a.logger.Info("Getting subscribe", zap.String("subscribe_id", subscribeID))
	return nil, nil
}

// GetSubscribes 获取订阅列表
func (a *SubscribeServiceAdapter) GetSubscribes(ctx context.Context, params GetSubscribesParams) ([]SubscribeItem, error) {
	// 实际实现中，这里应该调用核心业务服务的订阅API
	// 这里使用模拟实现，返回空列表
	a.logger.Info("Getting subscribes", zap.String("type", string(params.Type)), zap.String("status", string(params.Status)))
	return []SubscribeItem{}, nil
}

// ActivateSubscribe 激活订阅
func (a *SubscribeServiceAdapter) ActivateSubscribe(ctx context.Context, subscribeID string) error {
	// 实际实现中，这里应该调用核心业务服务的订阅API
	// 这里使用模拟实现，返回nil
	a.logger.Info("Activating subscribe", zap.String("subscribe_id", subscribeID))
	return nil
}

// PauseSubscribe 暂停订阅
func (a *SubscribeServiceAdapter) PauseSubscribe(ctx context.Context, subscribeID string) error {
	// 实际实现中，这里应该调用核心业务服务的订阅API
	// 这里使用模拟实现，返回nil
	a.logger.Info("Pausing subscribe", zap.String("subscribe_id", subscribeID))
	return nil
}

// RefreshSubscribe 刷新订阅
func (a *SubscribeServiceAdapter) RefreshSubscribe(ctx context.Context, subscribeID string) (*SubscribeItem, error) {
	// 实际实现中，这里应该调用核心业务服务的订阅API
	// 这里使用模拟实现，返回nil
	a.logger.Info("Refreshing subscribe", zap.String("subscribe_id", subscribeID))
	return nil, nil
}

// TriggerSubscribe 手动触发订阅检查
func (a *SubscribeServiceAdapter) TriggerSubscribe(ctx context.Context, subscribeID string) error {
	// 实际实现中，这里应该调用核心业务服务的订阅API
	// 这里使用模拟实现，返回nil
	a.logger.Info("Triggering subscribe check", zap.String("subscribe_id", subscribeID))
	return nil
}

// GetSubscribeTypes 获取支持的订阅类型
func (a *SubscribeServiceAdapter) GetSubscribeTypes(ctx context.Context) ([]SubscribeType, error) {
	// 实际实现中，这里应该调用核心业务服务的订阅API
	// 这里使用模拟实现，返回所有支持的订阅类型
	a.logger.Info("Getting subscribe types")
	return []SubscribeType{
		SubscribeTypeMovie,
		SubscribeTypeTV,
		SubscribeTypeRSS,
		SubscribeTypeOther,
	}, nil
}

// MockSubscribeService 模拟订阅服务实现，用于测试
type MockSubscribeService struct {
	logger     *zap.Logger
	subscribes map[string]SubscribeItem
}

// NewMockSubscribeService 创建新的模拟订阅服务实例
func NewMockSubscribeService(logger *zap.Logger) *MockSubscribeService {
	return &MockSubscribeService{
		logger:     logger,
		subscribes: make(map[string]SubscribeItem),
	}
}

// CreateSubscribe 创建订阅（模拟实现）
func (m *MockSubscribeService) CreateSubscribe(ctx context.Context, subscribe SubscribeItem) (string, error) {
	m.logger.Info("Mock creating subscribe", zap.String("title", subscribe.Title), zap.String("type", string(subscribe.Type)))

	// 创建模拟订阅
	subscribeID := "mock-subscribe-" + time.Now().Format("20060102150405")
	newSubscribe := SubscribeItem{
		ID:                  subscribeID,
		Title:               subscribe.Title,
		Keyword:             subscribe.Keyword,
		Type:                subscribe.Type,
		Status:              subscribe.Status,
		LastCheck:           time.Now(),
		NextCheck:           time.Now().Add(subscribe.CheckInterval),
		CheckInterval:       subscribe.CheckInterval,
		FilterRules:         subscribe.FilterRules,
		AutoDownload:        subscribe.AutoDownload,
		NotificationEnabled: subscribe.NotificationEnabled,
		Metadata:            subscribe.Metadata,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	// 如果未设置订阅类型，使用默认值
	if newSubscribe.Type == "" {
		newSubscribe.Type = SubscribeTypeTV
	}

	// 如果未设置订阅状态，使用默认值
	if newSubscribe.Status == "" {
		newSubscribe.Status = SubscribeStatusActive
	}

	// 如果未设置检查间隔，使用默认值
	if newSubscribe.CheckInterval == 0 {
		newSubscribe.CheckInterval = time.Hour * 1
	}

	// 如果未设置下次检查时间，使用当前时间加上检查间隔
	if newSubscribe.NextCheck.IsZero() {
		newSubscribe.NextCheck = time.Now().Add(newSubscribe.CheckInterval)
	}

	m.subscribes[subscribeID] = newSubscribe
	return subscribeID, nil
}

// UpdateSubscribe 更新订阅（模拟实现）
func (m *MockSubscribeService) UpdateSubscribe(ctx context.Context, subscribeID string, updates map[string]any) error {
	m.logger.Info("Mock updating subscribe", zap.String("subscribe_id", subscribeID))

	// 更新模拟订阅
	subscribe, exists := m.subscribes[subscribeID]
	if !exists {
		return nil
	}

	// 这里简单处理，实际应该根据updates更新对应的字段
	// 为了简化，这里只更新状态
	subscribe.UpdatedAt = time.Now()
	m.subscribes[subscribeID] = subscribe

	return nil
}

// DeleteSubscribe 删除订阅（模拟实现）
func (m *MockSubscribeService) DeleteSubscribe(ctx context.Context, subscribeID string) error {
	m.logger.Info("Mock deleting subscribe", zap.String("subscribe_id", subscribeID))

	// 从模拟订阅列表中删除
	delete(m.subscribes, subscribeID)
	return nil
}

// GetSubscribe 获取单个订阅（模拟实现）
func (m *MockSubscribeService) GetSubscribe(ctx context.Context, subscribeID string) (*SubscribeItem, error) {
	m.logger.Info("Mock getting subscribe", zap.String("subscribe_id", subscribeID))

	subscribe, exists := m.subscribes[subscribeID]
	if !exists {
		return nil, nil
	}

	return &subscribe, nil
}

// GetSubscribes 获取订阅列表（模拟实现）
func (m *MockSubscribeService) GetSubscribes(ctx context.Context, params GetSubscribesParams) ([]SubscribeItem, error) {
	m.logger.Info("Mock getting subscribes", zap.String("type", string(params.Type)), zap.String("status", string(params.Status)))

	var subscribes []SubscribeItem
	for _, subscribe := range m.subscribes {
		if (params.Type == "" || subscribe.Type == params.Type) &&
			(params.Status == "" || subscribe.Status == params.Status) &&
			(params.Keyword == "" || subscribe.Keyword == params.Keyword) {
			subscribes = append(subscribes, subscribe)
		}
	}

	return subscribes, nil
}

// ActivateSubscribe 激活订阅（模拟实现）
func (m *MockSubscribeService) ActivateSubscribe(ctx context.Context, subscribeID string) error {
	m.logger.Info("Mock activating subscribe", zap.String("subscribe_id", subscribeID))

	// 更新模拟订阅状态
	if subscribe, exists := m.subscribes[subscribeID]; exists {
		subscribe.Status = SubscribeStatusActive
		subscribe.UpdatedAt = time.Now()
		m.subscribes[subscribeID] = subscribe
	}

	return nil
}

// PauseSubscribe 暂停订阅（模拟实现）
func (m *MockSubscribeService) PauseSubscribe(ctx context.Context, subscribeID string) error {
	m.logger.Info("Mock pausing subscribe", zap.String("subscribe_id", subscribeID))

	// 更新模拟订阅状态
	if subscribe, exists := m.subscribes[subscribeID]; exists {
		subscribe.Status = SubscribeStatusPaused
		subscribe.UpdatedAt = time.Now()
		m.subscribes[subscribeID] = subscribe
	}

	return nil
}

// RefreshSubscribe 刷新订阅（模拟实现）
func (m *MockSubscribeService) RefreshSubscribe(ctx context.Context, subscribeID string) (*SubscribeItem, error) {
	m.logger.Info("Mock refreshing subscribe", zap.String("subscribe_id", subscribeID))

	subscribe, exists := m.subscribes[subscribeID]
	if !exists {
		return nil, nil
	}

	// 更新订阅的最后检查时间和下次检查时间
	subscribe.LastCheck = time.Now()
	subscribe.NextCheck = time.Now().Add(subscribe.CheckInterval)
	subscribe.UpdatedAt = time.Now()
	m.subscribes[subscribeID] = subscribe

	return &subscribe, nil
}

// TriggerSubscribe 手动触发订阅检查（模拟实现）
func (m *MockSubscribeService) TriggerSubscribe(ctx context.Context, subscribeID string) error {
	m.logger.Info("Mock triggering subscribe check", zap.String("subscribe_id", subscribeID))

	// 模拟触发订阅检查
	_, err := m.RefreshSubscribe(ctx, subscribeID)
	return err
}

// GetSubscribeTypes 获取支持的订阅类型（模拟实现）
func (m *MockSubscribeService) GetSubscribeTypes(ctx context.Context) ([]SubscribeType, error) {
	m.logger.Info("Mock getting subscribe types")
	return []SubscribeType{
		SubscribeTypeMovie,
		SubscribeTypeTV,
		SubscribeTypeRSS,
		SubscribeTypeOther,
	}, nil
}
