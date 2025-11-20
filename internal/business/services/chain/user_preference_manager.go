package chain

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/models"
	"github.com/yfh-yun/moviepilot-go/internal/repositories"
	"github.com/yfh-yun/moviepilot-go/pkg/cache"
	"github.com/yfh-yun/moviepilot-go/pkg/utils"
)

// UserActivity 用户活动记录
type UserActivity struct {
	ID         int64                  `json:"id"`
	UserID     int64                  `json:"user_id"`
	Action     string                 `json:"action"`
	Resource   string                 `json:"resource"`
	Details    map[string]interface{} `json:"details"`
	IPAddress  string                 `json:"ip_address"`
	UserAgent  string                 `json:"user_agent"`
	CreateTime time.Time              `json:"create_time"`
}

// UserPreference 用户偏好设置
type UserPreference struct {
	UserID      int64                  `json:"user_id"`
	Preferences map[string]interface{} `json:"preferences"`
	Settings    map[string]interface{} `json:"settings"`
	UpdateTime  time.Time              `json:"update_time"`
}

// UserStats 用户统计信息
type UserStats struct {
	UserID             int64     `json:"user_id"`
	LoginCount         int64     `json:"login_count"`
	LastLoginTime      time.Time `json:"last_login_time"`
	LastActiveTime     time.Time `json:"last_active_time"`
	TotalDownloads     int64     `json:"total_downloads"`
	TotalSubscribes    int64     `json:"total_subscribes"`
	TotalSearches      int64     `json:"total_searches"`
	FavoriteGenres     []string  `json:"favorite_genres"`
	FavoriteCategories []string  `json:"favorite_categories"`
	PreferredQuality   []string  `json:"preferred_quality"`
	PreferredLanguage  []string  `json:"preferred_language"`
}

// UserPreferenceManager 用户偏好管理器
type UserPreferenceManager struct {
	cache      *cache.Cache
	logger     *logger.Logger
	userRepo   *repository.UserRepository
	expiration time.Duration
}

// NewUserPreferenceManager 创建用户偏好管理器
func NewUserPreferenceManager(cache *cache.Cache, logger *logger.Logger, userRepo *repository.UserRepository) *UserPreferenceManager {
	return &UserPreferenceManager{
		cache:      cache,
		logger:     logger,
		userRepo:   userRepo,
		expiration: 24 * time.Hour,
	}
}

// GetUserPreferences 获取用户偏好设置
func (m *UserPreferenceManager) GetUserPreferences(ctx context.Context, userID int64) (*UserPreference, error) {
	m.logger.Info("获取用户偏好设置", "userID", userID)

	// 先从缓存获取
	cacheKey := fmt.Sprintf("user:preference:%d", userID)
	if cached, err := m.cache.Get(ctx, cacheKey); err == nil {
		var preference UserPreference
		if err := json.Unmarshal([]byte(cached), &preference); err == nil {
			return &preference, nil
		}
	}

	// 从数据库获取
	user, err := m.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("用户不存在")
	}

	preference := &UserPreference{
		UserID:      userID,
		Preferences: user.Settings,
		Settings:    make(map[string]interface{}),
		UpdateTime:  time.Now(),
	}

	// 缓存结果
	if data, err := json.Marshal(preference); err == nil {
		m.cache.Set(ctx, cacheKey, string(data), m.expiration)
	}

	return preference, nil
}

// UpdateUserPreferences 更新用户偏好设置
func (m *UserPreferenceManager) UpdateUserPreferences(ctx context.Context, userID int64, preferences map[string]interface{}) error {
	m.logger.Info("更新用户偏好设置", "userID", userID)

	user, err := m.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	if user == nil {
		return fmt.Errorf("用户不存在")
	}

	// 合并偏好设置
	if user.Settings == nil {
		user.Settings = make(map[string]interface{})
	}

	for key, value := range preferences {
		user.Settings[key] = value
	}

	// 更新数据库
	updateData := model.UserUpdateData{
		Settings: user.Settings,
	}

	_, err = m.userRepo.UpdateUser(ctx, userID, updateData)
	if err != nil {
		return err
	}

	// 清除缓存
	cacheKey := fmt.Sprintf("user:preference:%d", userID)
	m.cache.Delete(ctx, cacheKey)

	m.logger.Info("更新用户偏好设置成功", "userID", userID)
	return nil
}

// GetFavoriteGenres 获取用户喜欢的类型
func (m *UserPreferenceManager) GetFavoriteGenres(ctx context.Context, userID int64) ([]string, error) {
	preference, err := m.GetUserPreferences(ctx, userID)
	if err != nil {
		return nil, err
	}

	genres, ok := preference.Preferences["favorite_genres"].([]interface{})
	if !ok {
		return []string{}, nil
	}

	var result []string
	for _, genre := range genres {
		if genreStr, ok := genre.(string); ok {
			result = append(result, genreStr)
		}
	}

	return result, nil
}

// GetPreferredQuality 获取用户偏好的质量
func (m *UserPreferenceManager) GetPreferredQuality(ctx context.Context, userID int64) ([]string, error) {
	preference, err := m.GetUserPreferences(ctx, userID)
	if err != nil {
		return nil, err
	}

	qualities, ok := preference.Preferences["preferred_quality"].([]interface{})
	if !ok {
		return []string{"1080p", "720p"}, nil // 默认偏好
	}

	var result []string
	for _, quality := range qualities {
		if qualityStr, ok := quality.(string); ok {
			result = append(result, qualityStr)
		}
	}

	return result, nil
}

// GetPreferredLanguage 获取用户偏好的语言
func (m *UserPreferenceManager) GetPreferredLanguage(ctx context.Context, userID int64) ([]string, error) {
	preference, err := m.GetUserPreferences(ctx, userID)
	if err != nil {
		return nil, err
	}

	languages, ok := preference.Preferences["preferred_language"].([]interface{})
	if !ok {
		return []string{"zh", "en"}, nil // 默认偏好
	}

	var result []string
	for _, language := range languages {
		if langStr, ok := language.(string); ok {
			result = append(result, langStr)
		}
	}

	return result, nil
}

// UserActivityTracker 用户活动跟踪器
type UserActivityTracker struct {
	cache      *cache.Cache
	logger     *logger.Logger
	userRepo   *repository.UserRepository
	expiration time.Duration
}

// NewUserActivityTracker 创建用户活动跟踪器
func NewUserActivityTracker(cache *cache.Cache, logger *logger.Logger, userRepo *repository.UserRepository) *UserActivityTracker {
	return &UserActivityTracker{
		cache:      cache,
		logger:     logger,
		userRepo:   userRepo,
		expiration: 7 * 24 * time.Hour, // 7天
	}
}

// TrackActivity 跟踪用户活动
func (t *UserActivityTracker) TrackActivity(ctx context.Context, activity *UserActivity) error {
	t.logger.Info("跟踪用户活动", "userID", activity.UserID, "action", activity.Action)

	// 设置创建时间
	if activity.CreateTime.IsZero() {
		activity.CreateTime = time.Now()
	}

	// 生成活动ID
	activity.ID = utils.GenerateSnowflakeID()

	// 存储到缓存（实际项目中应该存储到数据库）
	cacheKey := fmt.Sprintf("user:activity:%d:%d", activity.UserID, activity.ID)
	activityData, err := json.Marshal(activity)
	if err != nil {
		return err
	}

	t.cache.Set(ctx, cacheKey, string(activityData), t.expiration)

	// 更新用户最后活动时间
	userUpdateData := model.UserUpdateData{
		LastActiveAt: &activity.CreateTime,
	}

	_, err = t.userRepo.UpdateUser(ctx, activity.UserID, userUpdateData)
	if err != nil {
		t.logger.Error("更新用户最后活动时间失败", "userID", activity.UserID, "error", err)
	}

	// 更新用户统计信息
	go t.updateUserStats(activity.UserID, activity.Action)

	return nil
}

// GetUserActivities 获取用户活动列表
func (t *UserActivityTracker) GetUserActivities(ctx context.Context, userID int64, limit int) ([]*UserActivity, error) {
	t.logger.Info("获取用户活动列表", "userID", userID, "limit", limit)

	// 实际项目中应该从数据库查询
	// 这里简化实现，返回模拟数据
	activities := make([]*UserActivity, 0)

	// 从缓存获取活动数据
	pattern := fmt.Sprintf("user:activity:%d:*", userID)
	keys, err := t.cache.Keys(ctx, pattern)
	if err != nil {
		return nil, err
	}

	for i, key := range keys {
		if i >= limit {
			break
		}

		cached, err := t.cache.Get(ctx, key)
		if err != nil {
			continue
		}

		var activity UserActivity
		if err := json.Unmarshal([]byte(cached), &activity); err == nil {
			activities = append(activities, &activity)
		}
	}

	return activities, nil
}

// GetUserStats 获取用户统计信息
func (t *UserActivityTracker) GetUserStats(ctx context.Context, userID int64) (*UserStats, error) {
	t.logger.Info("获取用户统计信息", "userID", userID)

	// 先从缓存获取
	cacheKey := fmt.Sprintf("user:stats:%d", userID)
	if cached, err := t.cache.Get(ctx, cacheKey); err == nil {
		var stats UserStats
		if err := json.Unmarshal([]byte(cached), &stats); err == nil {
			return &stats, nil
		}
	}

	// 从数据库获取用户基本信息
	user, err := t.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("用户不存在")
	}

	// 获取用户活动并统计
	activities, err := t.GetUserActivities(ctx, userID, 1000)
	if err != nil {
		return nil, err
	}

	stats := t.calculateUserStats(userID, user, activities)

	// 缓存结果
	if data, err := json.Marshal(stats); err == nil {
		t.cache.Set(ctx, cacheKey, string(data), time.Hour)
	}

	return stats, nil
}

// calculateUserStats 计算用户统计信息
func (t *UserActivityTracker) calculateUserStats(userID int64, user *model.User, activities []*UserActivity) *UserStats {
	stats := &UserStats{
		UserID:         userID,
		LastActiveTime: time.Now(),
		FavoriteGenres: []string{},
	}

	// 设置最后登录时间
	if user.LastLoginAt != nil {
		stats.LastLoginTime = *user.LastLoginAt
	}

	// 统计各种活动
	loginCount := int64(0)
	downloadCount := int64(0)
	subscribeCount := int64(0)
	searchCount := int64(0)

	for _, activity := range activities {
		switch activity.Action {
		case "login":
			loginCount++
		case "download":
			downloadCount++
		case "subscribe":
			subscribeCount++
		case "search":
			searchCount++
		}
	}

	stats.LoginCount = loginCount
	stats.TotalDownloads = downloadCount
	stats.TotalSubscribes = subscribeCount
	stats.TotalSearches = searchCount

	// 从用户设置中获取偏好信息
	if user.Settings != nil {
		if genres, ok := user.Settings["favorite_genres"].([]interface{}); ok {
			for _, genre := range genres {
				if genreStr, ok := genre.(string); ok {
					stats.FavoriteGenres = append(stats.FavoriteGenres, genreStr)
				}
			}
		}

		if qualities, ok := user.Settings["preferred_quality"].([]interface{}); ok {
			for _, quality := range qualities {
				if qualityStr, ok := quality.(string); ok {
					stats.PreferredQuality = append(stats.PreferredQuality, qualityStr)
				}
			}
		}

		if languages, ok := user.Settings["preferred_language"].([]interface{}); ok {
			for _, language := range languages {
				if langStr, ok := language.(string); ok {
					stats.PreferredLanguage = append(stats.PreferredLanguage, langStr)
				}
			}
		}
	}

	return stats
}

// updateUserStats 更新用户统计信息
func (t *UserActivityTracker) updateUserStats(userID int64, action string) {
	ctx := context.Background()

	// 获取当前统计信息
	stats, err := t.GetUserStats(ctx, userID)
	if err != nil {
		t.logger.Error("获取用户统计信息失败", "userID", userID, "error", err)
		return
	}

	// 更新统计
	switch action {
	case "login":
		stats.LoginCount++
		stats.LastLoginTime = time.Now()
	case "download":
		stats.TotalDownloads++
	case "subscribe":
		stats.TotalSubscribes++
	case "search":
		stats.TotalSearches++
	}

	// 更新缓存
	cacheKey := fmt.Sprintf("user:stats:%d", userID)
	if data, err := json.Marshal(stats); err == nil {
		t.cache.Set(ctx, cacheKey, string(data), time.Hour)
	}
}

// PersonalizedRecommendationEngine 个性化推荐引擎
type PersonalizedRecommendationEngine struct {
	preferenceManager *UserPreferenceManager
	activityTracker   *UserActivityTracker
	logger            *logger.Logger
}

// NewPersonalizedRecommendationEngine 创建个性化推荐引擎
func NewPersonalizedRecommendationEngine(
	preferenceManager *UserPreferenceManager,
	activityTracker *UserActivityTracker,
	logger *logger.Logger,
) *PersonalizedRecommendationEngine {
	return &PersonalizedRecommendationEngine{
		preferenceManager: preferenceManager,
		activityTracker:   activityTracker,
		logger:            logger,
	}
}

// GetPersonalizedRecommendations 获取个性化推荐
func (e *PersonalizedRecommendationEngine) GetPersonalizedRecommendations(ctx context.Context, userID int64, limit int) (*model.PersonalizedRecommendations, error) {
	e.logger.Info("获取个性化推荐", "userID", userID, "limit", limit)

	// 获取用户偏好
	preferences, err := e.preferenceManager.GetUserPreferences(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 获取用户统计
	stats, err := e.activityTracker.GetUserStats(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 获取用户喜欢的类型
	favoriteGenres, err := e.preferenceManager.GetFavoriteGenres(ctx, userID)
	if err != nil {
		favoriteGenres = []string{}
	}

	// 获取用户偏好的质量
	preferredQuality, err := e.preferenceManager.GetPreferredQuality(ctx, userID)
	if err != nil {
		preferredQuality = []string{"1080p", "720p"}
	}

	// 获取用户偏好的语言
	preferredLanguage, err := e.preferenceManager.GetPreferredLanguage(ctx, userID)
	if err != nil {
		preferredLanguage = []string{"zh", "en"}
	}

	// 生成个性化推荐
	recommendations := &model.PersonalizedRecommendations{
		UserID:            userID,
		FavoriteGenres:    favoriteGenres,
		PreferredQuality:  preferredQuality,
		PreferredLanguage: preferredLanguage,
		Recommendations:   []model.MediaItem{},
	}

	// 基于用户偏好生成推荐（这里简化实现）
	// 实际项目中应该使用更复杂的推荐算法

	e.logger.Info("生成个性化推荐成功", "userID", userID, "genreCount", len(favoriteGenres))
	return recommendations, nil
}

// UpdateRecommendationWeights 更新推荐权重
func (e *PersonalizedRecommendationEngine) UpdateRecommendationWeights(ctx context.Context, userID int64, action string, mediaInfo *model.MediaInfo) error {
	e.logger.Info("更新推荐权重", "userID", userID, "action", action)

	// 根据用户行为更新推荐权重
	// 例如：用户下载了某个类型的电影，增加该类型的权重
	// 用户收藏了某个演员，增加该演员的作品权重

	// 这里实现简化的权重更新逻辑
	preferences := make(map[string]interface{})

	switch action {
	case "download":
		// 下载行为增加相关类型的权重
		if mediaInfo != nil && len(mediaInfo.Genres) > 0 {
			preferences["favorite_genres"] = mediaInfo.Genres
		}
	case "favorite":
		// 收藏行为增加相关演员的权重
		if mediaInfo != nil && len(mediaInfo.Actors) > 0 {
			preferences["favorite_actors"] = mediaInfo.Actors
		}
	case "like":
		// 点赞行为增加相关导演的权重
		if mediaInfo != nil && mediaInfo.Director != "" {
			preferences["favorite_directors"] = []string{mediaInfo.Director}
		}
	}

	// 更新用户偏好
	if len(preferences) > 0 {
		err := e.preferenceManager.UpdateUserPreferences(ctx, userID, preferences)
		if err != nil {
			return err
		}
	}

	e.logger.Info("更新推荐权重成功", "userID", userID)
	return nil
}
