package mediaserver

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// PlaybackService 播放状态服务
// 负责管理播放进度的收集、同步、分析和统计功能
type PlaybackService struct {
	logger         *zap.Logger
	mediaServerSvc *MediaServerService
}

// NewPlaybackService 创建播放状态服务实例
func NewPlaybackService(logger *zap.Logger, mediaServerSvc *MediaServerService) *PlaybackService {
	return &PlaybackService{
		logger:         logger,
		mediaServerSvc: mediaServerSvc,
	}
}

// PlaybackSession 播放会话信息
type PlaybackSession struct {
	ID          string    `json:"id"`           // 会话ID
	UserID      string    `json:"user_id"`      // 用户ID
	Username    string    `json:"username"`     // 用户名
	ItemID      string    `json:"item_id"`      // 媒体项ID
	ItemName    string    `json:"item_name"`    // 媒体项名称
	ItemType    string    `json:"item_type"`    // 媒体项类型
	Progress    float64   `json:"progress"`     // 播放进度 (0-1)
	Duration    float64   `json:"duration"`     // 总时长(秒)
	IsPlaying   bool      `json:"is_playing"`   // 是否正在播放
	LastPlayed  time.Time `json:"last_played"`  // 最后播放时间
	PlayedCount int       `json:"played_count"` // 播放次数
	ServerType  string    `json:"server_type"`  // 服务器类型
}

// PlaybackStatus 播放状态信息
type PlaybackStatus struct {
	Progress    float64   `json:"progress"`     // 播放进度
	IsPlaying   bool      `json:"is_playing"`   // 是否正在播放
	LastPlayed  time.Time `json:"last_played"`  // 最后播放时间
	PlayedCount int       `json:"played_count"` // 播放次数
}

// GetPlaybackSessions 获取播放会话列表
// 支持按服务器类型、用户ID、媒体项ID过滤
func (s *PlaybackService) GetPlaybackSessions(ctx context.Context, serverType, userID, itemID string) ([]*PlaybackSession, error) {
	s.logger.Debug("获取播放会话列表",
		zap.String("server_type", serverType),
		zap.String("user_id", userID),
		zap.String("item_id", itemID),
	)

	// 如果指定了服务器类型，只查询该服务器
	if serverType != "" {
		server, err := s.mediaServerSvc.GetServer(serverType)
		if err != nil {
			return nil, errors.Wrap(err, "获取媒体服务器实例失败")
		}

		sessions, err := server.GetPlaybackSessions(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "获取播放会话失败")
		}

		return s.filterSessions(sessions, userID, itemID), nil
	}

	// 查询所有启用的服务器
	servers := s.mediaServerSvc.GetEnabledServers()
	allSessions := make([]*PlaybackSession, 0)

	for _, server := range servers {
		sessions, err := server.GetPlaybackSessions(ctx)
		if err != nil {
			s.logger.Warn("获取服务器播放会话失败",
				zap.String("server_type", server.GetType()),
				zap.Error(err),
			)
			continue
		}

		allSessions = append(allSessions, sessions...)
	}

	return s.filterSessions(allSessions, userID, itemID), nil
}

// UpdatePlaybackStatus 更新播放状态
// 支持跨服务器同步播放状态
func (s *PlaybackService) UpdatePlaybackStatus(ctx context.Context, serverType, itemID, userID string, status *PlaybackStatus) error {
	s.logger.Info("更新播放状态",
		zap.String("server_type", serverType),
		zap.String("item_id", itemID),
		zap.String("user_id", userID),
		zap.Float64("progress", status.Progress),
		zap.Bool("is_playing", status.IsPlaying),
	)

	// 获取目标服务器
	server, err := s.mediaServerSvc.GetServer(serverType)
	if err != nil {
		return errors.Wrap(err, "获取媒体服务器实例失败")
	}

	// 检查连接状态
	if !server.IsConnected() {
		s.logger.Warn("媒体服务器未连接，尝试重新连接",
			zap.String("server_type", serverType),
		)

		if err := server.Connect(ctx); err != nil {
			return errors.Wrap(err, "媒体服务器连接失败")
		}
	}

	// 更新播放状态
	if err := server.UpdatePlaybackStatus(ctx, itemID, *status); err != nil {
		return errors.Wrap(err, "更新播放状态失败")
	}

	s.logger.Info("播放状态更新成功",
		zap.String("server_type", serverType),
		zap.String("item_id", itemID),
	)

	return nil
}

// SyncPlaybackStatus 同步播放状态
// 将播放状态从一个服务器同步到另一个服务器
func (s *PlaybackService) SyncPlaybackStatus(ctx context.Context, sourceServer, targetServer, userID string) error {
	s.logger.Info("同步播放状态",
		zap.String("source_server", sourceServer),
		zap.String("target_server", targetServer),
		zap.String("user_id", userID),
	)

	// 获取源服务器的播放会话
	sourceSessions, err := s.GetPlaybackSessions(ctx, sourceServer, userID, "")
	if err != nil {
		return errors.Wrap(err, "获取源服务器播放会话失败")
	}

	// 同步每个播放会话
	for _, session := range sourceSessions {
		status := &PlaybackStatus{
			Progress:    session.Progress,
			IsPlaying:   session.IsPlaying,
			LastPlayed:  session.LastPlayed,
			PlayedCount: session.PlayedCount,
		}

		if err := s.UpdatePlaybackStatus(ctx, targetServer, session.ItemID, session.UserID, status); err != nil {
			s.logger.Warn("同步单个播放状态失败",
				zap.String("session_id", session.ID),
				zap.String("item_id", session.ItemID),
				zap.Error(err),
			)
		} else {
			s.logger.Debug("播放状态同步成功",
				zap.String("session_id", session.ID),
				zap.String("item_id", session.ItemID),
			)
		}
	}

	s.logger.Info("播放状态同步完成",
		zap.Int("synced_count", len(sourceSessions)),
	)

	return nil
}

// GetPlaybackHistory 获取播放历史记录
// 支持按时间范围、用户、媒体类型过滤
func (s *PlaybackService) GetPlaybackHistory(ctx context.Context, startTime, endTime time.Time, userID, itemType string) ([]*PlaybackSession, error) {
	s.logger.Debug("获取播放历史记录",
		zap.Time("start_time", startTime),
		zap.Time("end_time", endTime),
		zap.String("user_id", userID),
		zap.String("item_type", itemType),
	)

	// 这里实现播放历史查询逻辑
	// 可以从数据库或服务器获取历史记录
	history := make([]*PlaybackSession, 0)

	return history, nil
}

// GetPlaybackStatistics 获取播放统计信息
// 提供播放次数、时长、进度等统计数据分析
func (s *PlaybackService) GetPlaybackStatistics(ctx context.Context, startTime, endTime time.Time, userID string) (*PlaybackStatistics, error) {
	s.logger.Debug("获取播放统计信息",
		zap.Time("start_time", startTime),
		zap.Time("end_time", endTime),
		zap.String("user_id", userID),
	)

	// 获取播放历史记录
	history, err := s.GetPlaybackHistory(ctx, startTime, endTime, userID, "")
	if err != nil {
		return nil, errors.Wrap(err, "获取播放历史失败")
	}

	// 计算统计信息
	stats := s.calculateStatistics(history)

	s.logger.Info("播放统计信息计算完成",
		zap.Int("total_sessions", len(history)),
		zap.Int("total_play_count", stats.TotalPlayCount),
		zap.Float64("total_duration_hours", stats.TotalPlayDurationHours),
	)

	return stats, nil
}

// PlaybackStatistics 播放统计信息
type PlaybackStatistics struct {
	TotalPlayCount         int            `json:"total_play_count"`          // 总播放次数
	TotalPlayDurationHours float64        `json:"total_play_duration_hours"` // 总播放时长(小时)
	AverageProgress        float64        `json:"average_progress"`          // 平均播放进度
	MostPlayedItem         string         `json:"most_played_item"`          // 最常播放项
	MostPlayedCount        int            `json:"most_played_count"`         // 最常播放次数
	PlayByType             map[string]int `json:"play_by_type"`              // 按类型统计
	PlayByServer           map[string]int `json:"play_by_server"`            // 按服务器统计
}

// filterSessions 过滤播放会话
func (s *PlaybackService) filterSessions(sessions []*PlaybackSession, userID, itemID string) []*PlaybackSession {
	if userID == "" && itemID == "" {
		return sessions
	}

	filtered := make([]*PlaybackSession, 0)

	for _, session := range sessions {
		// 过滤用户
		if userID != "" && session.UserID != userID {
			continue
		}

		// 过滤媒体项
		if itemID != "" && session.ItemID != itemID {
			continue
		}

		filtered = append(filtered, session)
	}

	return filtered
}

// calculateStatistics 计算播放统计信息
func (s *PlaybackService) calculateStatistics(sessions []*PlaybackSession) *PlaybackStatistics {
	stats := &PlaybackStatistics{
		PlayByType:   make(map[string]int),
		PlayByServer: make(map[string]int),
	}

	// 计算基本统计
	stats.TotalPlayCount = len(sessions)

	// 计算时长和进度
	totalDuration := 0.0
	totalProgress := 0.0

	itemPlayCount := make(map[string]int)

	for _, session := range sessions {
		// 累计时长
		totalDuration += session.Duration
		totalProgress += session.Progress

		// 统计播放次数
		stats.PlayByType[session.ItemType]++
		stats.PlayByServer[session.ServerType]++

		// 统计单个媒体项播放次数
		itemPlayCount[session.ItemID]++
	}

	// 计算平均值
	if len(sessions) > 0 {
		stats.TotalPlayDurationHours = totalDuration / 3600.0
		stats.AverageProgress = totalProgress / float64(len(sessions))
	}

	// 找出最常播放的媒体项
	if len(itemPlayCount) > 0 {
		maxCount := 0
		var mostPlayedItem string

		for itemID, count := range itemPlayCount {
			if count > maxCount {
				maxCount = count
				mostPlayedItem = itemID
			}
		}

		stats.MostPlayedItem = mostPlayedItem
		stats.MostPlayedCount = maxCount
	}

	return stats
}
