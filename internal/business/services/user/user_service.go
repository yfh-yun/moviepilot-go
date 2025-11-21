// Package user 用户管理服务层
package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"moviepilot-go/internal/repositories/interfaces"
	"moviepilot-go/internal/models"
)

// UserService 用户服务接口
type UserService interface {
	// GetUserProfile 获取用户详细信息
	GetUserProfile(ctx context.Context, userID uint) (*models.User, error)
	// UpdateUserProfile 更新用户基本信息
	UpdateUserProfile(ctx context.Context, userID uint, username, email, avatar string) error
	// UpdateUserSettings 更新用户个性化设置
	UpdateUserSettings(ctx context.Context, userID uint, settings map[string]interface{}) error
	// GetUserSettings 获取用户设置
	GetUserSettings(ctx context.Context, userID uint) (map[string]interface{}, error)
	// UpdateUserPermissions 更新用户权限
	UpdateUserPermissions(ctx context.Context, userID uint, permissions map[string]bool) error
	// GetUserPermissions 获取用户权限
	GetUserPermissions(ctx context.Context, userID uint) (map[string]bool, error)
	// GetUserStats 获取用户统计信息
	GetUserStats(ctx context.Context, userID uint) (*UserStats, error)
	// ListUsers 获取用户列表（管理员）
	ListUsers(ctx context.Context, offset, limit int) ([]*models.User, int64, error)
	// EnableUser 启用用户（管理员）
	EnableUser(ctx context.Context, adminID, userID uint) error
	// DisableUser 禁用用户（管理员）
	DisableUser(ctx context.Context, adminID, userID uint) error
	// PromoteToAdmin 提升为管理员（超管）
	PromoteToAdmin(ctx context.Context, superAdminID, userID uint) error
	// DemoteFromAdmin 降级为普通用户（超管）
	DemoteFromAdmin(ctx context.Context, superAdminID, userID uint) error
	// GetUserActivity 获取用户活动记录
	GetUserActivity(ctx context.Context, userID uint, days int) ([]*UserActivity, error)
}

// UserStats 用户统计信息
type UserStats struct {
	UserID              uint    `json:"user_id"`
	TotalSubscribes     int64   `json:"total_subscribes"`
	ActiveSubscribes    int64   `json:"active_subscribes"`
	TotalDownloads      int64   `json:"total_downloads"`
	SuccessfulDownloads int64   `json:"successful_downloads"`
	FailedDownloads     int64   `json:"failed_downloads"`
	StorageUsed         float64 `json:"storage_used"` // GB
	LastActivityTime    string  `json:"last_activity_time"`
}

// UserActivity 用户活动记录
type UserActivity struct {
	ID          uint   `json:"id"`
	Type        string `json:"type"`   // subscribe, download, media, etc.
	Action      string `json:"action"` // created, updated, deleted
	Target      string `json:"target"` // target name
	Description string `json:"description"`
	Timestamp   string `json:"timestamp"`
}

// userService 用户服务实现
type userService struct {
	userRepo interfaces.UserRepository
}

// NewUserService 创建用户服务实例
func NewUserService(userRepo interfaces.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

// GetUserProfile 获取用户详细信息
func (s *userService) GetUserProfile(ctx context.Context, userID uint) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user profile failed: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// 清除敏感信息
	user.PasswordHash = ""
	user.OTPSecret = ""

	return user, nil
}

// UpdateUserProfile 更新用户基本信息
func (s *userService) UpdateUserProfile(ctx context.Context, userID uint, username, email, avatar string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user failed: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}

	// 检查用户名是否被其他用户使用
	if username != "" && username != user.Name {
		exists, err := s.userRepo.Exists(username)
		if err != nil {
			return fmt.Errorf("check username exists failed: %w", err)
		}
		if exists {
			return errors.New("username already exists")
		}
		user.Name = username
	}

	// 检查邮箱是否被其他用户使用
	if email != "" && email != user.Email {
		exists, err := s.userRepo.ExistsEmail(email)
		if err != nil {
			return fmt.Errorf("check email exists failed: %w", err)
		}
		if exists {
			return errors.New("email already exists")
		}
		user.Email = email
	}

	if avatar != "" {
		user.Avatar = avatar
	}

	return s.userRepo.Update(ctx, user)
}

// UpdateUserSettings 更新用户个性化设置
func (s *userService) UpdateUserSettings(ctx context.Context, userID uint, settings map[string]interface{}) error {
	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("marshal settings failed: %w", err)
	}

	return s.userRepo.UpdateSettings(ctx, userID, string(settingsJSON))
}

// GetUserSettings 获取用户设置
func (s *userService) GetUserSettings(ctx context.Context, userID uint) (map[string]interface{}, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user failed: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	var settings map[string]interface{}
	if user.Settings != "" {
		if err := json.Unmarshal([]byte(user.Settings), &settings); err != nil {
			return nil, fmt.Errorf("unmarshal settings failed: %w", err)
		}
	}

	return settings, nil
}

// UpdateUserPermissions 更新用户权限
func (s *userService) UpdateUserPermissions(ctx context.Context, userID uint, permissions map[string]bool) error {
	permissionsJSON, err := json.Marshal(permissions)
	if err != nil {
		return fmt.Errorf("marshal permissions failed: %w", err)
	}

	return s.userRepo.UpdatePermissions(ctx, userID, string(permissionsJSON))
}

// GetUserPermissions 获取用户权限
func (s *userService) GetUserPermissions(ctx context.Context, userID uint) (map[string]bool, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user failed: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	var permissions map[string]bool
	if user.Permissions != "" {
		if err := json.Unmarshal([]byte(user.Permissions), &permissions); err != nil {
			return nil, fmt.Errorf("unmarshal permissions failed: %w", err)
		}
	}

	return permissions, nil
}

// GetUserStats 获取用户统计信息
func (s *userService) GetUserStats(ctx context.Context, userID uint) (*UserStats, error) {
	// 这里需要集成其他服务来获取统计信息
	// 目前返回模拟数据
	stats := &UserStats{
		UserID:              userID,
		TotalSubscribes:     10,
		ActiveSubscribes:    5,
		TotalDownloads:      100,
		SuccessfulDownloads: 95,
		FailedDownloads:     5,
		StorageUsed:         15.2,
		LastActivityTime:    "2025-11-16 15:30:00",
	}

	return stats, nil
}

// ListUsers 获取用户列表（管理员）
func (s *userService) ListUsers(ctx context.Context, offset, limit int) ([]*models.User, int64, error) {
	users, total, err := s.userRepo.List(ctx, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("list users failed: %w", err)
	}

	// 清除敏感信息
	for _, user := range users {
		user.PasswordHash = ""
		user.OTPSecret = ""
	}

	return users, total, nil
}

// EnableUser 启用用户（管理员）
func (s *userService) EnableUser(ctx context.Context, adminID, userID uint) error {
	// 检查管理员权限
	admin, err := s.userRepo.GetByID(ctx, adminID)
	if err != nil {
		return fmt.Errorf("get admin failed: %w", err)
	}
	if admin == nil || !admin.IsSuperuser {
		return errors.New("unauthorized")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user failed: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}

	user.IsActive = true
	return s.userRepo.Update(ctx, user)
}

// DisableUser 禁用用户（管理员）
func (s *userService) DisableUser(ctx context.Context, adminID, userID uint) error {
	// 检查管理员权限
	admin, err := s.userRepo.GetByID(ctx, adminID)
	if err != nil {
		return fmt.Errorf("get admin failed: %w", err)
	}
	if admin == nil || !admin.IsSuperuser {
		return errors.New("unauthorized")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user failed: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}

	user.IsActive = false
	return s.userRepo.Update(ctx, user)
}

// PromoteToAdmin 提升为管理员（超管）
func (s *userService) PromoteToAdmin(ctx context.Context, superAdminID, userID uint) error {
	// 检查超级管理员权限
	superAdmin, err := s.userRepo.GetByID(ctx, superAdminID)
	if err != nil {
		return fmt.Errorf("get super admin failed: %w", err)
	}
	if superAdmin == nil || !superAdmin.IsSuperuser {
		return errors.New("unauthorized")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user failed: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}

	user.IsSuperuser = true
	return s.userRepo.Update(ctx, user)
}

// DemoteFromAdmin 降级为普通用户（超管）
func (s *userService) DemoteFromAdmin(ctx context.Context, superAdminID, userID uint) error {
	// 检查超级管理员权限
	superAdmin, err := s.userRepo.GetByID(ctx, superAdminID)
	if err != nil {
		return fmt.Errorf("get super admin failed: %w", err)
	}
	if superAdmin == nil || !superAdmin.IsSuperuser {
		return errors.New("unauthorized")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user failed: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}

	user.IsSuperuser = false
	return s.userRepo.Update(ctx, user)
}

// GetUserActivity 获取用户活动记录
func (s *userService) GetUserActivity(ctx context.Context, userID uint, days int) ([]*UserActivity, error) {
	// 这里需要集成活动记录服务
	// 目前返回模拟数据
	activities := []*UserActivity{
		{
			ID:          1,
			Type:        "subscribe",
			Action:      "created",
			Target:      "权力的游戏 第八季",
			Description: "创建了新的订阅",
			Timestamp:   "2025-11-16 15:30:00",
		},
		{
			ID:          2,
			Type:        "download",
			Action:      "completed",
			Target:      "阿凡达2：水之道.mp4",
			Description: "下载完成",
			Timestamp:   "2025-11-16 14:20:00",
		},
	}

	return activities, nil
}
