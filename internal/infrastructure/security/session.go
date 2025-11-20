package security

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
	"time"

	"moviepilot-go/pkg/logger"
)

// MemoryUserSessionManager 内存用户会话管理器实现
type MemoryUserSessionManager struct {
	sessions      map[string]*UserSession
	userSessions  map[string]map[string]*UserSession // userID -> map[sessionID]*UserSession
	mutex         sync.RWMutex
	config        *SecurityConfig
	jwtManager    *JWTManager
	logger        logger.Logger
	sessionTimeout time.Duration
}

// NewMemoryUserSessionManager 创建内存用户会话管理器
func NewMemoryUserSessionManager(config *SecurityConfig, jwtManager *JWTManager, logger logger.Logger) *MemoryUserSessionManager {
	return &MemoryUserSessionManager{
		sessions:       make(map[string]*UserSession),
		userSessions:   make(map[string]map[string]*UserSession),
		config:         config,
		jwtManager:     jwtManager,
		logger:         logger,
		sessionTimeout: time.Duration(config.SessionTimeout) * time.Minute,
	}
}

// CreateSession 创建用户会话
func (sm *MemoryUserSessionManager) CreateSession(userID, ip, userAgent string) (*UserSession, error) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// 生成会话ID
	sessionID, err := sm.generateSessionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	// 创建JWT令牌
	token, err := sm.jwtManager.GenerateToken(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// 创建会话
	session := &UserSession{
		SessionID:    sessionID,
		UserID:       userID,
		Token:        token,
		IP:           ip,
		UserAgent:    userAgent,
		LastActivity: time.Now(),
		ExpiresAt:    time.Now().Add(sm.sessionTimeout),
		IsActive:     true,
	}

	// 存储会话
	sm.sessions[sessionID] = session

	// 更新用户会话映射
	if _, exists := sm.userSessions[userID]; !exists {
		sm.userSessions[userID] = make(map[string]*UserSession)
	}
	sm.userSessions[userID][sessionID] = session

	// 如果配置了最大会话数，检查并删除最旧的会话
	if sm.config.MaxSessionsPerUser > 0 {
		sm.enforceMaxSessions(userID)
	}

	sm.logger.Info("User session created", "user_id", userID, "session_id", sessionID)
	return session, nil
}

// GetSession 获取会话
func (sm *MemoryUserSessionManager) GetSession(sessionID string) (*UserSession, error) {
	sm.mutex.RLock()
	session, exists := sm.sessions[sessionID]
	sm.mutex.RUnlock()

	if !exists {
		return nil, errors.New("session not found")
	}

	// 检查会话是否过期
	if time.Now().After(session.ExpiresAt) {
		// 异步清理过期会话
		go sm.InvalidateSession(sessionID)
		return nil, errors.New("session expired")
	}

	// 检查会话是否活跃
	if !session.IsActive {
		return nil, errors.New("session inactive")
	}

	// 验证JWT令牌
	claims, err := sm.jwtManager.ValidateToken(session.Token)
	if err != nil {
		// 令牌无效，使会话失效
		go sm.InvalidateSession(sessionID)
		return nil, errors.New("invalid token")
	}

	// 确保令牌中的用户ID与会话匹配
	if claims.UserID != session.UserID {
		go sm.InvalidateSession(sessionID)
		return nil, errors.New("token user mismatch")
	}

	// 返回会话副本
	return session.clone(), nil
}

// UpdateSessionActivity 更新会话活动时间
func (sm *MemoryUserSessionManager) UpdateSessionActivity(sessionID string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return errors.New("session not found")
	}

	// 更新活动时间和过期时间
	session.LastActivity = time.Now()
	session.ExpiresAt = time.Now().Add(sm.sessionTimeout)

	return nil
}

// InvalidateSession 使会话失效
func (sm *MemoryUserSessionManager) InvalidateSession(sessionID string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil // 会话不存在，视为已失效
	}

	// 从用户会话映射中删除
	if userSessions, exists := sm.userSessions[session.UserID]; exists {
		delete(userSessions, sessionID)
		if len(userSessions) == 0 {
			delete(sm.userSessions, session.UserID)
		}
	}

	// 删除会话
	delete(sm.sessions, sessionID)

	sm.logger.Info("User session invalidated", "user_id", session.UserID, "session_id", sessionID)
	return nil
}

// InvalidateUserSessions 使用户的所有会话失效
func (sm *MemoryUserSessionManager) InvalidateUserSessions(userID string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// 获取用户的所有会话
	userSessions, exists := sm.userSessions[userID]
	if !exists {
		return nil // 用户没有会话
	}

	// 删除每个会话
	for sessionID := range userSessions {
		delete(sm.sessions, sessionID)
	}

	// 删除用户会话映射
	delete(sm.userSessions, userID)

	sm.logger.Info("All user sessions invalidated", "user_id", userID, "session_count", len(userSessions))
	return nil
}

// ListUserSessions 列出用户的所有会话
func (sm *MemoryUserSessionManager) ListUserSessions(userID string) ([]*UserSession, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	userSessions, exists := sm.userSessions[userID]
	if !exists {
		return []*UserSession{}, nil
	}

	// 转换为切片并返回副本
	sessions := make([]*UserSession, 0, len(userSessions))
	for _, session := range userSessions {
		sessions = append(sessions, session.clone())
	}

	return sessions, nil
}

// GetActiveSessionsCount 获取活跃会话数量
func (sm *MemoryUserSessionManager) GetActiveSessionsCount() int {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	return len(sm.sessions)
}

// RefreshSessionToken 刷新会话令牌
func (sm *MemoryUserSessionManager) RefreshSessionToken(sessionID string) (string, error) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return "", errors.New("session not found")
	}

	// 检查会话是否活跃
	if !session.IsActive || time.Now().After(session.ExpiresAt) {
		return "", errors.New("session inactive or expired")
	}

	// 生成新的JWT令牌
	newToken, err := sm.jwtManager.GenerateToken(session.UserID)
	if err != nil {
		return "", fmt.Errorf("failed to generate new token: %w", err)
	}

	// 更新会话
	session.Token = newToken
	session.LastActivity = time.Now()
	session.ExpiresAt = time.Now().Add(sm.sessionTimeout)

	sm.logger.Info("Session token refreshed", "user_id", session.UserID, "session_id", sessionID)
	return newToken, nil
}

// VerifySessionToken 验证会话令牌
func (sm *MemoryUserSessionManager) VerifySessionToken(token string) (*UserSession, error) {
	// 验证JWT令牌
	claims, err := sm.jwtManager.ValidateToken(token)
	if err != nil {
		return nil, errors.New("invalid token")
	}

	// 查找对应的会话
	userID := claims.UserID

	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	// 遍历用户的所有会话查找匹配的令牌
	if userSessions, exists := sm.userSessions[userID]; exists {
		for _, session := range userSessions {
			if session.Token == token && session.IsActive && time.Now().Before(session.ExpiresAt) {
				return session.clone(), nil
			}
		}
	}

	return nil, errors.New("session not found for token")
}

// CleanupExpiredSessions 清理过期会话
func (sm *MemoryUserSessionManager) CleanupExpiredSessions() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	now := time.Now()
	expiredCount := 0

	// 找出所有过期会话
	for sessionID, session := range sm.sessions {
		if now.After(session.ExpiresAt) {
			// 从用户会话映射中删除
			if userSessions, exists := sm.userSessions[session.UserID]; exists {
				delete(userSessions, sessionID)
				if len(userSessions) == 0 {
					delete(sm.userSessions, session.UserID)
				}
			}

			// 删除会话
			delete(sm.sessions, sessionID)
			expiredCount++
		}
	}

	if expiredCount > 0 {
		sm.logger.Info("Cleaned up expired sessions", "count", expiredCount)
	}
}

// StartCleanupTask 启动定期清理任务
func (sm *MemoryUserSessionManager) StartCleanupTask() {
	ticker := time.NewTicker(30 * time.Minute) // 每30分钟清理一次
	go func() {
		for range ticker.C {
			sm.CleanupExpiredSessions()
		}
	}()

	sm.logger.Info("Session cleanup task started")
}

// generateSessionID 生成会话ID
func (sm *MemoryUserSessionManager) generateSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// enforceMaxSessions 强制执行最大会话数限制
func (sm *MemoryUserSessionManager) enforceMaxSessions(userID string) {
	userSessions, exists := sm.userSessions[userID]
	if !exists || len(userSessions) <= sm.config.MaxSessionsPerUser {
		return
	}

	// 将会话按最后活动时间排序
	sessions := make([]*UserSession, 0, len(userSessions))
	for _, session := range userSessions {
		sessions = append(sessions, session)
	}

	// 简单选择排序（按最后活动时间升序）
	for i := 0; i < len(sessions)-1; i++ {
		minIndex := i
		for j := i + 1; j < len(sessions); j++ {
			if sessions[j].LastActivity.Before(sessions[minIndex].LastActivity) {
				minIndex = j
			}
		}
		sessions[i], sessions[minIndex] = sessions[minIndex], sessions[i]
	}

	// 删除超出限制的最旧会话
	excessCount := len(sessions) - sm.config.MaxSessionsPerUser
	for i := 0; i < excessCount; i++ {
		sessionID := sessions[i].SessionID
		delete(sm.sessions, sessionID)
		delete(userSessions, sessionID)
		sm.logger.Info("Oldest session removed due to max sessions limit", 
			"user_id", userID, "session_id", sessionID)
	}
}

// IsSessionValid 检查会话是否有效
func (sm *MemoryUserSessionManager) IsSessionValid(sessionID string) bool {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return false
	}

	return session.IsActive && time.Now().Before(session.ExpiresAt)
}

// GetSessionStats 获取会话统计信息
func (sm *MemoryUserSessionManager) GetSessionStats() map[string]interface{} {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	// 计算每个用户的会话数
	userSessionCounts := make(map[string]int)
	for userID, sessions := range sm.userSessions {
		userSessionCounts[userID] = len(sessions)
	}

	// 计算活跃会话数（非过期的）
	now := time.Now()
	activeCount := 0
	for _, session := range sm.sessions {
		if session.IsActive && now.Before(session.ExpiresAt) {
			activeCount++
		}
	}

	return map[string]interface{}{
		"total_sessions":     len(sm.sessions),
		"active_sessions":    activeCount,
		"unique_users":       len(sm.userSessions),
		"user_session_counts": userSessionCounts,
	}
}

// clone 克隆会话
func (s *UserSession) clone() *UserSession {
	clone := *s
	return &clone
}