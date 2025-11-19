package meta

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

// MediaMetadata 媒体元数据
type MediaMetadata struct {
	ID            string                 `json:"id"`             // 媒体唯一标识
	Type          string                 `json:"type"`           // 媒体类型：movie, series, episode, season, anime, music
	Title         string                 `json:"title"`          // 标题
	OriginalTitle string                 `json:"original_title"` // 原始标题
	Year          int                    `json:"year"`           // 年份
	Released      time.Time              `json:"released"`       // 发布日期
	Poster        string                 `json:"poster"`         // 海报URL
	Backdrop      string                 `json:"backdrop"`       // 背景图URL
	Overview      string                 `json:"overview"`       // 简介
	Runtime       int                    `json:"runtime"`        // 运行时间（分钟）
	Rating        float64                `json:"rating"`         // 评分
	Genres        []string               `json:"genres"`         // 类型
	Countries     []string               `json:"countries"`      // 国家
	Languages     []string               `json:"languages"`      // 语言
	Directors     []string               `json:"directors"`      // 导演
	Writers       []string               `json:"writers"`        // 编剧
	Actors        []string               `json:"actors"`         // 演员
	Studios       []string               `json:"studios"`        // 工作室
	Status        string                 `json:"status"`         // 状态
	Tags          []string               `json:"tags"`           // 标签
	Collections   []string               `json:"collections"`    // 系列/合集
	ExternalIDs   map[string]string      `json:"external_ids"`   // 外部ID映射
	EpisodeCount  int                    `json:"episode_count"`  // 剧集数（针对系列）
	SeasonCount   int                    `json:"season_count"`   // 季数（针对系列）
	SeasonNumber  int                    `json:"season_number"`  // 季号（针对季/集）
	EpisodeNumber int                    `json:"episode_number"` // 集号（针对集）
	ParentID      string                 `json:"parent_id"`      // 父级ID（针对季/集）
	AlternativeTitles []string           `json:"alternative_titles"` // 替代标题
	Keywords      []string               `json:"keywords"`       // 关键词
	Trailers      []string               `json:"trailers"`       // 预告片URL
	StreamingLinks []*StreamingLink      `json:"streaming_links"` // 流媒体链接
	CustomFields  map[string]interface{} `json:"custom_fields"`  // 自定义字段
	CreateTime    time.Time              `json:"create_time"`    // 创建时间
	UpdateTime    time.Time              `json:"update_time"`    // 更新时间
	Providers     map[string]ProviderInfo `json:"providers"`     // 提供方信息
	Source        string                 `json:"source"`         // 数据来源
	IsAdult       bool                   `json:"is_adult"`       // 是否成人内容
}

// ProviderInfo 内容提供方信息
type ProviderInfo struct {
	ProviderID   string  `json:"provider_id"`    // 提供方ID
	ProviderName string  `json:"provider_name"`  // 提供方名称
	Link         string  `json:"link"`           // 链接
	Quality      string  `json:"quality"`        // 质量
	IsFree       bool    `json:"is_free"`        // 是否免费
	IsRentable   bool    `json:"is_rentable"`    // 是否可租赁
	IsBuyable    bool    `json:"is_buyable"`     // 是否可购买
	IsExclusive  bool    `json:"is_exclusive"`   // 是否独家
	Price        float64 `json:"price,omitempty"` // 价格
}

// MetadataProvider 元数据提供接口
type MetadataProvider interface {
	ID() string
	Name() string
	Description() string
	Priority() int
	SupportsMediaType(mediaType string) bool
	SearchMetadata(ctx context.Context, title string, year int, mediaType string) ([]*MediaMetadata, error)
	GetMetadataByID(ctx context.Context, id string, mediaType string) (*MediaMetadata, error)
	GetMetadataByExternalID(ctx context.Context, externalIDType string, externalID string, mediaType string) (*MediaMetadata, error)
	GetSeasonMetadata(ctx context.Context, seriesID string, seasonNumber int) (*MediaMetadata, error)
	GetEpisodeMetadata(ctx context.Context, seriesID string, seasonNumber int, episodeNumber int) (*MediaMetadata, error)
	GetSeriesEpisodes(ctx context.Context, seriesID string, seasonNumber int) ([]*MediaMetadata, error)
	UpdateMetadata(ctx context.Context, metadata *MediaMetadata) error
}

// MetadataCache 元数据缓存接口
type MetadataCache interface {
	Set(ctx context.Context, key string, value *MediaMetadata, expiration time.Duration) error
	Get(ctx context.Context, key string) (*MediaMetadata, error)
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error
	GetStats(ctx context.Context) map[string]interface{}
}

// MetadataManager 元数据管理器
type MetadataManager struct {
	sync.RWMutex
	providers        []MetadataProvider
	providersByID    map[string]MetadataProvider
	cache            MetadataCache
	defaultCacheTTL  time.Duration
	streamingManager *StreamingPlatformManager
	onlineMode       bool
	initialized      bool
	httpTimeout      time.Duration
	requestTimeout   time.Duration
	retryAttempts    int
	retryDelay       time.Duration
}

// MetadataManagerOption 元数据管理器配置选项
type MetadataManagerOption func(*MetadataManager)

// WithCache 设置元数据缓存
func WithCache(cache MetadataCache) MetadataManagerOption {
	return func(m *MetadataManager) {
		m.cache = cache
	}
}

// WithDefaultCacheTTL 设置默认缓存TTL
func WithDefaultCacheTTL(ttl time.Duration) MetadataManagerOption {
	return func(m *MetadataManager) {
		m.defaultCacheTTL = ttl
	}
}

// WithOnlineMode 设置在线模式
func WithOnlineMode(online bool) MetadataManagerOption {
	return func(m *MetadataManager) {
		m.onlineMode = online
	}
}

// WithHTTPTimeout 设置HTTP超时
func WithHTTPTimeout(timeout time.Duration) MetadataManagerOption {
	return func(m *MetadataManager) {
		m.httpTimeout = timeout
	}
}

// WithRequestTimeout 设置请求超时
func WithRequestTimeout(timeout time.Duration) MetadataManagerOption {
	return func(m *MetadataManager) {
		m.requestTimeout = timeout
	}
}

// WithRetryAttempts 设置重试次数
func WithRetryAttempts(attempts int) MetadataManagerOption {
	return func(m *MetadataManager) {
		if attempts > 0 {
			m.retryAttempts = attempts
		}
	}
}

// WithRetryDelay 设置重试延迟
func WithRetryDelay(delay time.Duration) MetadataManagerOption {
	return func(m *MetadataManager) {
		m.retryDelay = delay
	}
}

// NewMetadataManager 创建元数据管理器实例
func NewMetadataManager(options ...MetadataManagerOption) *MetadataManager {
	manager := &MetadataManager{
		providers:        make([]MetadataProvider, 0),
		providersByID:    make(map[string]MetadataProvider),
		defaultCacheTTL:  24 * time.Hour,
		streamingManager: NewStreamingPlatformManager(),
		onlineMode:       true,
		initialized:      false,
		httpTimeout:      30 * time.Second,
		requestTimeout:   60 * time.Second,
		retryAttempts:    3,
		retryDelay:       1 * time.Second,
	}

	// 应用配置选项
	for _, option := range options {
		option(manager)
	}

	return manager
}

// Initialize 初始化元数据管理器
func (m *MetadataManager) Initialize(ctx context.Context) error {
	m.Lock()
	defer m.Unlock()

	if m.initialized {
		return nil
	}

	// 初始化流媒体平台管理器
	if err := m.streamingManager.Initialize(ctx); err != nil {
		return fmt.Errorf("初始化流媒体平台管理器失败: %w", err)
	}

	m.initialized = true
	return nil
}

// RegisterProvider 注册元数据提供方
func (m *MetadataManager) RegisterProvider(ctx context.Context, provider MetadataProvider) error {
	if !m.initialized {
		if err := m.Initialize(ctx); err != nil {
			return fmt.Errorf("初始化元数据管理器失败: %w", err)
		}
	}

	if provider == nil {
		return errors.New("提供方不能为空")
	}

	providerID := provider.ID()
	if providerID == "" {
		return errors.New("提供方ID不能为空")
	}

	m.Lock()
	defer m.Unlock()

	// 检查是否已存在
	if _, exists := m.providersByID[providerID]; exists {
		return fmt.Errorf("提供方ID %s 已存在", providerID)
	}

	// 添加到列表和映射
	m.providers = append(m.providers, provider)
	m.providersByID[providerID] = provider

	// 按优先级排序（优先级高的在前）
	m.sortProviders()

	return nil
}

// UnregisterProvider 注销元数据提供方
func (m *MetadataManager) UnregisterProvider(ctx context.Context, providerID string) error {
	if !m.initialized {
		if err := m.Initialize(ctx); err != nil {
			return fmt.Errorf("初始化元数据管理器失败: %w", err)
		}
	}

	m.Lock()
	defer m.Unlock()

	// 检查是否存在
	if _, exists := m.providersByID[providerID]; !exists {
		return fmt.Errorf("提供方ID %s 不存在", providerID)
	}

	// 从列表中移除
	var newProviders []MetadataProvider
	for _, provider := range m.providers {
		if provider.ID() != providerID {
			newProviders = append(newProviders, provider)
		}
	}
	m.providers = newProviders

	// 从映射中移除
	delete(m.providersByID, providerID)

	return nil
}

// GetProvider 获取指定ID的元数据提供方
func (m *MetadataManager) GetProvider(ctx context.Context, providerID string) (MetadataProvider, error) {
	if !m.initialized {
		if err := m.Initialize(ctx); err != nil {
			return nil, fmt.Errorf("初始化元数据管理器失败: %w", err)
		}
	}

	m.RLock()
	defer m.RUnlock()

	provider, exists := m.providersByID[providerID]
	if !exists {
		return nil, fmt.Errorf("提供方ID %s 不存在", providerID)
	}

	return provider, nil
}

// GetAllProviders 获取所有元数据提供方
func (m *MetadataManager) GetAllProviders(ctx context.Context) ([]MetadataProvider, error) {
	if !m.initialized {
		if err := m.Initialize(ctx); err != nil {
			return nil, fmt.Errorf("初始化元数据管理器失败: %w", err)
		}
	}

	m.RLock()
	defer m.RUnlock()

	// 返回副本以避免并发问题
	providers := make([]MetadataProvider, len(m.providers))
	copy(providers, m.providers)

	return providers, nil
}

// GetProvidersByMediaType 获取支持指定媒体类型的元数据提供方
func (m *MetadataManager) GetProvidersByMediaType(ctx context.Context, mediaType string) ([]MetadataProvider, error) {
	if !m.initialized {
		if err := m.Initialize(ctx); err != nil {
			return nil, fmt.Errorf("初始化元数据管理器失败: %w", err)
		}
	}

	m.RLock()
	defer m.RUnlock()

	var result []MetadataProvider
	for _, provider := range m.providers {
		if provider.SupportsMediaType(mediaType) {
			result = append(result, provider)
		}
	}

	return result, nil
}

// SearchMetadata 搜索元数据
func (m *MetadataManager) SearchMetadata(ctx context.Context, title string, year int, mediaType string) ([]*MediaMetadata, error) {
	if !m.initialized {
		if err := m.Initialize(ctx); err != nil {
			return nil, fmt.Errorf("初始化元数据管理器失败: %w", err)
		}
	}

	if title == "" {
		return nil, errors.New("标题不能为空")
	}

	// 生成缓存键
	cacheKey := fmt.Sprintf("search:%s:%d:%s", title, year, mediaType)

	// 尝试从缓存获取
	if m.cache != nil {
		cachedResult, err := m.cache.Get(ctx, cacheKey)
		if err == nil && cachedResult != nil {
			// 缓存命中，返回结果的副本
			var result []*MediaMetadata
			if err := json.Unmarshal([]byte(fmt.Sprintf("[%v]", cachedResult)), &result); err == nil {
				return result, nil
			}
		}
	}

	// 在线模式下，从提供方获取数据
	if m.onlineMode {
		providers := m.getProvidersForMediaType(mediaType)
		if len(providers) == 0 {
			return nil, fmt.Errorf("没有找到支持媒体类型 %s 的提供方", mediaType)
		}

		// 并发请求所有提供方
		var wg sync.WaitGroup
		var mu sync.Mutex
		var results []*MediaMetadata
		var lastErr error

		for _, provider := range providers {
			wg.Add(1)
			go func(p MetadataProvider) {
				defer wg.Done()

				// 添加请求超时
				ctxWithTimeout, cancel := context.WithTimeout(ctx, m.requestTimeout)
				defer cancel()

				providerResults, err := m.retryWithBackoff(func() ([]*MediaMetadata, error) {
					return p.SearchMetadata(ctxWithTimeout, title, year, mediaType)
				})

				if err != nil {
					lastErr = err
					return
				}

				if len(providerResults) > 0 {
					mu.Lock()
					results = append(results, providerResults...)
					mu.Unlock()
				}
			}(provider)
		}

		wg.Wait()

		// 去重并排序
		uniqueResults := m.deduplicateMetadata(results)

		// 缓存结果
		if m.cache != nil && len(uniqueResults) > 0 {
			// 只缓存第一个结果
			m.cache.Set(ctx, cacheKey, uniqueResults[0], m.defaultCacheTTL)
		}

		return uniqueResults, lastErr
	}

	return nil, errors.New("离线模式下不支持搜索")
}

// GetMetadataByID 根据ID获取元数据
func (m *MetadataManager) GetMetadataByID(ctx context.Context, providerID string, id string, mediaType string) (*MediaMetadata, error) {
	if !m.initialized {
		return nil, fmt.Errorf("初始化元数据管理器失败: %w", m.Initialize(ctx))
	}

	if id == "" {
		return nil, errors.New("ID不能为空")
	}

	// 生成缓存键
	cacheKey := fmt.Sprintf("id:%s:%s:%s", providerID, id, mediaType)

	// 尝试从缓存获取
	if m.cache != nil {
		cachedMetadata, err := m.cache.Get(ctx, cacheKey)
		if err == nil && cachedMetadata != nil {
			return cachedMetadata, nil
		}
	}

	// 在线模式下，从指定提供方获取数据
	if m.onlineMode {
		provider, err := m.GetProvider(ctx, providerID)
		if err != nil {
			return nil, err
		}

		if !provider.SupportsMediaType(mediaType) {
			return nil, fmt.Errorf("提供方 %s 不支持媒体类型 %s", providerID, mediaType)
		}

		// 添加请求超时
		ctxWithTimeout, cancel := context.WithTimeout(ctx, m.requestTimeout)
		defer cancel()

		metadata, err := m.retryWithBackoff(func() (*MediaMetadata, error) {
			return provider.GetMetadataByID(ctxWithTimeout, id, mediaType)
		})

		if err != nil {
			return nil, err
		}

		// 更新元数据的来源
		if metadata != nil {
			metadata.Source = providerID
			metadata.UpdateTime = time.Now()

			// 缓存结果
			if m.cache != nil {
				m.cache.Set(ctx, cacheKey, metadata, m.defaultCacheTTL)
			}
		}

		return metadata, nil
	}

	return nil, errors.New("离线模式下不支持根据ID获取元数据")
}

// GetMetadataByExternalID 根据外部ID获取元数据
func (m *MetadataManager) GetMetadataByExternalID(ctx context.Context, externalIDType string, externalID string, mediaType string) (*MediaMetadata, error) {
	if !m.initialized {
		return nil, fmt.Errorf("初始化元数据管理器失败: %w", m.Initialize(ctx))
	}

	if externalIDType == "" || externalID == "" {
		return nil, errors.New("外部ID类型和ID不能为空")
	}

	// 生成缓存键
	cacheKey := fmt.Sprintf("external:%s:%s:%s", externalIDType, externalID, mediaType)

	// 尝试从缓存获取
	if m.cache != nil {
		cachedMetadata, err := m.cache.Get(ctx, cacheKey)
		if err == nil && cachedMetadata != nil {
			return cachedMetadata, nil
		}
	}

	// 在线模式下，从提供方获取数据
	if m.onlineMode {
		providers := m.getProvidersForMediaType(mediaType)
		if len(providers) == 0 {
			return nil, fmt.Errorf("没有找到支持媒体类型 %s 的提供方", mediaType)
		}

		// 依次尝试每个提供方
		var lastErr error
		for _, provider := range providers {
			// 添加请求超时
			ctxWithTimeout, cancel := context.WithTimeout(ctx, m.requestTimeout)
			defer cancel()

			metadata, err := m.retryWithBackoff(func() (*MediaMetadata, error) {
				return provider.GetMetadataByExternalID(ctxWithTimeout, externalIDType, externalID, mediaType)
			})

			if err != nil {
				lastErr = err
				continue
			}

			if metadata != nil {
				// 更新元数据的来源
				metadata.Source = provider.ID()
				metadata.UpdateTime = time.Now()

				// 缓存结果
				if m.cache != nil {
					m.cache.Set(ctx, cacheKey, metadata, m.defaultCacheTTL)
				}

				return metadata, nil
			}
		}

		return nil, fmt.Errorf("所有提供方都无法获取外部ID %s:%s 的元数据: %w", externalIDType, externalID, lastErr)
	}

	return nil, errors.New("离线模式下不支持根据外部ID获取元数据")
}

// GetSeasonMetadata 获取季元数据
func (m *MetadataManager) GetSeasonMetadata(ctx context.Context, providerID string, seriesID string, seasonNumber int) (*MediaMetadata, error) {
	if !m.initialized {
		return nil, fmt.Errorf("初始化元数据管理器失败: %w", m.Initialize(ctx))
	}

	if seriesID == "" || seasonNumber <= 0 {
		return nil, errors.New("系列ID和季号不能为空且季号必须大于0")
	}

	// 生成缓存键
	cacheKey := fmt.Sprintf("season:%s:%s:%d", providerID, seriesID, seasonNumber)

	// 尝试从缓存获取
	if m.cache != nil {
		cachedMetadata, err := m.cache.Get(ctx, cacheKey)
		if err == nil && cachedMetadata != nil {
			return cachedMetadata, nil
		}
	}

	// 在线模式下，从指定提供方获取数据
	if m.onlineMode {
		provider, err := m.GetProvider(ctx, providerID)
		if err != nil {
			return nil, err
		}

		// 添加请求超时
		ctxWithTimeout, cancel := context.WithTimeout(ctx, m.requestTimeout)
		defer cancel()

		metadata, err := m.retryWithBackoff(func() (*MediaMetadata, error) {
			return provider.GetSeasonMetadata(ctxWithTimeout, seriesID, seasonNumber)
		})

		if err != nil {
			return nil, err
		}

		// 更新元数据
		if metadata != nil {
			metadata.Source = providerID
			metadata.UpdateTime = time.Now()
			metadata.Type = "season"
			metadata.ParentID = seriesID

			// 缓存结果
			if m.cache != nil {
				m.cache.Set(ctx, cacheKey, metadata, m.defaultCacheTTL)
			}
		}

		return metadata, nil
	}

	return nil, errors.New("离线模式下不支持获取季元数据")
}

// GetEpisodeMetadata 获取集元数据
func (m *MetadataManager) GetEpisodeMetadata(ctx context.Context, providerID string, seriesID string, seasonNumber int, episodeNumber int) (*MediaMetadata, error) {
	if !m.initialized {
		return nil, fmt.Errorf("初始化元数据管理器失败: %w", m.Initialize(ctx))
	}

	if seriesID == "" || seasonNumber <= 0 || episodeNumber <= 0 {
		return nil, errors.New("系列ID、季号和集号不能为空且必须大于0")
	}

	// 生成缓存键
	cacheKey := fmt.Sprintf("episode:%s:%s:%d:%d", providerID, seriesID, seasonNumber, episodeNumber)

	// 尝试从缓存获取
	if m.cache != nil {
		cachedMetadata, err := m.cache.Get(ctx, cacheKey)
		if err == nil && cachedMetadata != nil {
			return cachedMetadata, nil
		}
	}

	// 在线模式下，从指定提供方获取数据
	if m.onlineMode {
		provider, err := m.GetProvider(ctx, providerID)
		if err != nil {
			return nil, err
		}

		// 添加请求超时
		ctxWithTimeout, cancel := context.WithTimeout(ctx, m.requestTimeout)
		defer cancel()

		metadata, err := m.retryWithBackoff(func() (*MediaMetadata, error) {
			return provider.GetEpisodeMetadata(ctxWithTimeout, seriesID, seasonNumber, episodeNumber)
		})

		if err != nil {
			return nil, err
		}

		// 更新元数据
		if metadata != nil {
			metadata.Source = providerID
			metadata.UpdateTime = time.Now()
			metadata.Type = "episode"
			metadata.ParentID = seriesID

			// 缓存结果
			if m.cache != nil {
				m.cache.Set(ctx, cacheKey, metadata, m.defaultCacheTTL)
			}
		}

		return metadata, nil
	}

	return nil, errors.New("离线模式下不支持获取集元数据")
}

// GetSeriesEpisodes 获取系列的所有剧集
func (m *MetadataManager) GetSeriesEpisodes(ctx context.Context, providerID string, seriesID string, seasonNumber int) ([]*MediaMetadata, error) {
	if !m.initialized {
		return nil, fmt.Errorf("初始化元数据管理器失败: %w", m.Initialize(ctx))
	}

	if seriesID == "" || seasonNumber <= 0 {
		return nil, errors.New("系列ID和季号不能为空且季号必须大于0")
	}

	// 生成缓存键
	cacheKey := fmt.Sprintf("episodes:%s:%s:%d", providerID, seriesID, seasonNumber)

	// 尝试从缓存获取
	if m.cache != nil {
		cachedMetadata, err := m.cache.Get(ctx, cacheKey)
		if err == nil && cachedMetadata != nil {
			// 缓存命中，解析结果
			var results []*MediaMetadata
			if err := json.Unmarshal([]byte(fmt.Sprintf("[%v]", cachedMetadata)), &results); err == nil {
				return results, nil
			}
		}
	}

	// 在线模式下，从指定提供方获取数据
	if m.onlineMode {
		provider, err := m.GetProvider(ctx, providerID)
		if err != nil {
			return nil, err
		}

		// 添加请求超时
		ctxWithTimeout, cancel := context.WithTimeout(ctx, m.requestTimeout)
		defer cancel()

		episodes, err := m.retryWithBackoff(func() ([]*MediaMetadata, error) {
			return provider.GetSeriesEpisodes(ctxWithTimeout, seriesID, seasonNumber)
		})

		if err != nil {
			return nil, err
		}

		// 更新所有剧集的元数据
		for _, episode := range episodes {
			episode.Source = providerID
			episode.UpdateTime = time.Now()
			episode.Type = "episode"
			episode.ParentID = seriesID
		}

		// 缓存结果
		if m.cache != nil && len(episodes) > 0 {
			// 只缓存第一个剧集作为代表
			m.cache.Set(ctx, cacheKey, episodes[0], m.defaultCacheTTL)
		}

		return episodes, nil
	}

	return nil, errors.New("离线模式下不支持获取系列剧集")
}

// UpdateMetadata 更新元数据
func (m *MetadataManager) UpdateMetadata(ctx context.Context, metadata *MediaMetadata) error {
	if !m.initialized {
		return fmt.Errorf("初始化元数据管理器失败: %w", m.Initialize(ctx))
	}

	if metadata == nil {
		return errors.New("元数据不能为空")
	}

	// 更新时间戳
	metadata.UpdateTime = time.Now()

	// 如果有来源，尝试使用对应的提供方更新
	if metadata.Source != "" {
		provider, err := m.GetProvider(ctx, metadata.Source)
		if err == nil {
			if err := provider.UpdateMetadata(ctx, metadata); err != nil {
				return err
			}
		}
	}

	// 清除相关缓存
	m.clearRelatedCache(ctx, metadata)

	return nil
}

// MergeMetadata 合并元数据
func (m *MetadataManager) MergeMetadata(source, target *MediaMetadata) *MediaMetadata {
	if target == nil {
		return source
	}

	if source == nil {
		return target
	}

	// 深拷贝目标元数据
	merged := &MediaMetadata{}
	*merged = *target

	// 合并基础字段（只在目标为空时合并）
	if merged.Title == "" {
		merged.Title = source.Title
	}

	if merged.OriginalTitle == "" {
		merged.OriginalTitle = source.OriginalTitle
	}

	if merged.Year == 0 {
		merged.Year = source.Year
	}

	if merged.Released.IsZero() {
		merged.Released = source.Released
	}

	if merged.Poster == "" {
		merged.Poster = source.Poster
	}

	if merged.Backdrop == "" {
		merged.Backdrop = source.Backdrop
	}

	if merged.Overview == "" {
		merged.Overview = source.Overview
	}

	if merged.Runtime == 0 {
		merged.Runtime = source.Runtime
	}

	if merged.Rating == 0 {
		merged.Rating = source.Rating
	}

	if merged.Status == "" {
		merged.Status = source.Status
	}

	// 合并数组字段（去重）
	merged.Genres = m.mergeStringSlice(merged.Genres, source.Genres)
	merged.Countries = m.mergeStringSlice(merged.Countries, source.Countries)
	merged.Languages = m.mergeStringSlice(merged.Languages, source.Languages)
	merged.Directors = m.mergeStringSlice(merged.Directors, source.Directors)
	merged.Writers = m.mergeStringSlice(merged.Writers, source.Writers)
	merged.Actors = m.mergeStringSlice(merged.Actors, source.Actors)
	merged.Studios = m.mergeStringSlice(merged.Studios, source.Studios)
	merged.Tags = m.mergeStringSlice(merged.Tags, source.Tags)
	merged.Collections = m.mergeStringSlice(merged.Collections, source.Collections)
	merged.AlternativeTitles = m.mergeStringSlice(merged.AlternativeTitles, source.AlternativeTitles)
	merged.Keywords = m.mergeStringSlice(merged.Keywords, source.Keywords)
	merged.Trailers = m.mergeStringSlice(merged.Trailers, source.Trailers)

	// 合并映射字段
	if merged.ExternalIDs == nil {
		merged.ExternalIDs = make(map[string]string)
	}
	for k, v := range source.ExternalIDs {
		if _, exists := merged.ExternalIDs[k]; !exists {
			merged.ExternalIDs[k] = v
		}
	}

	if merged.CustomFields == nil {
		merged.CustomFields = make(map[string]interface{})
	}
	for k, v := range source.CustomFields {
		merged.CustomFields[k] = v
	}

	if merged.Providers == nil {
		merged.Providers = make(map[string]ProviderInfo)
	}
	for k, v := range source.Providers {
		merged.Providers[k] = v
	}

	// 合并流媒体链接
	if len(source.StreamingLinks) > 0 {
		if merged.StreamingLinks == nil {
			merged.StreamingLinks = make([]*StreamingLink, 0)
		}
		merged.StreamingLinks = append(merged.StreamingLinks, source.StreamingLinks...)
	}

	// 更新更新时间
	merged.UpdateTime = time.Now()

	return merged
}

// ValidateMetadata 验证元数据
func (m *MetadataManager) ValidateMetadata(metadata *MediaMetadata) error {
	if metadata == nil {
		return errors.New("元数据不能为空")
	}

	if metadata.ID == "" {
		return errors.New("元数据ID不能为空")
	}

	if metadata.Type == "" {
		return errors.New("媒体类型不能为空")
	}

	if metadata.Title == "" {
		return errors.New("标题不能为空")
	}

	validTypes := map[string]bool{
		"movie":   true,
		"series":  true,
		"season":  true,
		"episode": true,
		"anime":   true,
		"music":   true,
	}

	if !validTypes[metadata.Type] {
		return fmt.Errorf("无效的媒体类型: %s", metadata.Type)
	}

	// 针对不同类型的特定验证
	switch metadata.Type {
	case "season":
		if metadata.SeasonNumber <= 0 {
			return errors.New("季号必须大于0")
		}
		if metadata.ParentID == "" {
			return errors.New("季必须有父级系列ID")
		}
	case "episode":
		if metadata.SeasonNumber <= 0 {
			return errors.New("季号必须大于0")
		}
		if metadata.EpisodeNumber <= 0 {
			return errors.New("集号必须大于0")
		}
		if metadata.ParentID == "" {
			return errors.New("集必须有父级系列ID")
		}
	}

	return nil
}

// SetCache 设置缓存
func (m *MetadataManager) SetCache(ctx context.Context, cache MetadataCache) error {
	if !m.initialized {
		if err := m.Initialize(ctx); err != nil {
			return fmt.Errorf("初始化元数据管理器失败: %w", err)
		}
	}

	m.Lock()
	defer m.Unlock()

	m.cache = cache
	return nil
}

// SetOnlineMode 设置在线模式
func (m *MetadataManager) SetOnlineMode(ctx context.Context, online bool) error {
	if !m.initialized {
		if err := m.Initialize(ctx); err != nil {
			return fmt.Errorf("初始化元数据管理器失败: %w", err)
		}
	}

	m.Lock()
	defer m.Unlock()

	m.onlineMode = online
	return nil
}

// GetStreamingPlatformManager 获取流媒体平台管理器
func (m *MetadataManager) GetStreamingPlatformManager() *StreamingPlatformManager {
	return m.streamingManager
}

// GetStats 获取统计信息
func (m *MetadataManager) GetStats(ctx context.Context) map[string]interface{} {
	stats := map[string]interface{}{
		"providers_count":   len(m.providers),
		"online_mode":       m.onlineMode,
		"initialized":       m.initialized,
		"default_cache_ttl": m.defaultCacheTTL,
		"retry_attempts":    m.retryAttempts,
		"retry_delay":       m.retryDelay,
		"http_timeout":      m.httpTimeout,
		"request_timeout":   m.requestTimeout,
	}

	// 添加缓存统计
	if m.cache != nil {
		cacheStats := m.cache.GetStats(ctx)
		stats["cache_stats"] = cacheStats
	}

	// 添加流媒体平台统计
	streamingStats := m.streamingManager.GetStatistics()
	stats["streaming_stats"] = streamingStats

	return stats
}

// 私有辅助方法

// sortProviders 按优先级排序提供方
func (m *MetadataManager) sortProviders() {
	// 优先级高的在前
	sort.Slice(m.providers, func(i, j int) bool {
		return m.providers[i].Priority() > m.providers[j].Priority()
	})
}

// getProvidersForMediaType 获取支持指定媒体类型的提供方（内部方法，不加锁）
func (m *MetadataManager) getProvidersForMediaType(mediaType string) []MetadataProvider {
	var providers []MetadataProvider
	for _, provider := range m.providers {
		if provider.SupportsMediaType(mediaType) {
			providers = append(providers, provider)
		}
	}
	return providers
}

// retryWithBackoff 带退避的重试逻辑
func (m *MetadataManager) retryWithBackoff(fn func() (interface{}, error)) (interface{}, error) {
	var lastErr error
	for attempt := 0; attempt <= m.retryAttempts; attempt++ {
		result, err := fn()
		if err == nil {
			return result, nil
		}

		lastErr = err

		// 最后一次尝试失败后不再重试
		if attempt == m.retryAttempts {
			break
		}

		// 退避延迟
		time.Sleep(m.retryDelay * time.Duration(attempt+1))
	}

	return nil, lastErr
}

// deduplicateMetadata 去重元数据
func (m *MetadataManager) deduplicateMetadata(metadataList []*MediaMetadata) []*MediaMetadata {
	if len(metadataList) == 0 {
		return metadataList
	}

	// 使用map去重，基于标题和年份
	seen := make(map[string]bool)
	var result []*MediaMetadata

	for _, metadata := range metadataList {
		// 生成唯一键
		key := fmt.Sprintf("%s:%d", metadata.Title, metadata.Year)
		if !seen[key] {
			seen[key] = true
			result = append(result, metadata)
		}
	}

	// 按评分排序（评分高的在前）
	sort.Slice(result, func(i, j int) bool {
		return result[i].Rating > result[j].Rating
	})

	return result
}

// mergeStringSlice 合并字符串切片并去重
func (m *MetadataManager) mergeStringSlice(target, source []string) []string {
	if len(source) == 0 {
		return target
	}

	if len(target) == 0 {
		// 创建副本以避免修改原始数据
		result := make([]string, len(source))
		copy(result, source)
		return result
	}

	// 使用map去重
	seen := make(map[string]bool)
	for _, item := range target {
		seen[item] = true
	}

	// 添加新元素
	for _, item := range source {
		if !seen[item] {
			seen[item] = true
			target = append(target, item)
		}
	}

	return target
}

// clearRelatedCache 清除相关缓存
func (m *MetadataManager) clearRelatedCache(ctx context.Context, metadata *MediaMetadata) {
	if m.cache == nil || metadata == nil {
		return
	}

	// 清除搜索缓存
	searchCacheKey := fmt.Sprintf("search:%s:%d:%s", metadata.Title, metadata.Year, metadata.Type)
	m.cache.Delete(ctx, searchCacheKey)

	// 清除ID缓存
	if metadata.Source != "" && metadata.ID != "" {
		idCacheKey := fmt.Sprintf("id:%s:%s:%s", metadata.Source, metadata.ID, metadata.Type)
		m.cache.Delete(ctx, idCacheKey)
	}

	// 清除外部ID缓存
	for extType, extID := range metadata.ExternalIDs {
		extCacheKey := fmt.Sprintf("external:%s:%s:%s", extType, extID, metadata.Type)
		m.cache.Delete(ctx, extCacheKey)
	}

	// 根据类型清除特定缓存
	switch metadata.Type {
	case "season":
		if metadata.Source != "" && metadata.ParentID != "" && metadata.SeasonNumber > 0 {
			seasonCacheKey := fmt.Sprintf("season:%s:%s:%d", metadata.Source, metadata.ParentID, metadata.SeasonNumber)
			episodeListCacheKey := fmt.Sprintf("episodes:%s:%s:%d", metadata.Source, metadata.ParentID, metadata.SeasonNumber)
			m.cache.Delete(ctx, seasonCacheKey)
			m.cache.Delete(ctx, episodeListCacheKey)
		}
	case "episode":
		if metadata.Source != "" && metadata.ParentID != "" && metadata.SeasonNumber > 0 && metadata.EpisodeNumber > 0 {
			episodeCacheKey := fmt.Sprintf("episode:%s:%s:%d:%d", metadata.Source, metadata.ParentID, metadata.SeasonNumber, metadata.EpisodeNumber)
			m.cache.Delete(ctx, episodeCacheKey)
		}
	}
}

// sort 包级别的辅助函数，用于排序
func sortProviders(providers []MetadataProvider) {
	// 优先级高的在前
	sort.Slice(providers, func(i, j int) bool {
		return providers[i].Priority() > providers[j].Priority()
	})
}