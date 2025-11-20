// Package service 业务逻辑层
package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/yfh-yun/moviepilot-go/internal/repositories/interfaces"
	"github.com/yfh-yun/moviepilot-go/internal/models"
	"github.com/yfh-yun/moviepilot-go/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

// AuthService 认证服务接口
type AuthService interface {
	// Login 用户登录
	Login(ctx context.Context, username, password string) (*jwt.TokenPair, error)
	// Register 用户注册
	Register(ctx context.Context, username, password, email string) (*models.User, error)
	// RefreshToken 刷新Token
	RefreshToken(ctx context.Context, refreshToken string) (*jwt.TokenPair, error)
	// Logout 用户登出
	Logout(ctx context.Context, userID uint) error
	// ChangePassword 修改密码
	ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error
	// GetUserByToken 通过Token获取用户信息
	GetUserByToken(ctx context.Context, token string) (*models.User, error)
}

// authService 认证服务实现
type authService struct {
	userRepo interfaces.UserRepository
}

// NewAuthService 创建认证服务实例
func NewAuthService(userRepo interfaces.UserRepository) AuthService {
	return &authService{
		userRepo: userRepo,
	}
}

// Login 用户登录
func (s *authService) Login(ctx context.Context, username, password string) (*jwt.TokenPair, error) {
	// 1. 查找用户
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 2. 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid password")
	}

	// 3. 检查用户状态
	if !user.IsActive {
		return nil, errors.New("user is disabled")
	}

	// 4. 生成Token
	tokenPair, err := jwt.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, fmt.Errorf("generate token failed: %w", err)
	}

	return tokenPair, nil
}

// Register 用户注册
func (s *authService) Register(ctx context.Context, username, password, email string) (*models.User, error) {
	// 1. 检查用户是否已存在
	existingUser, _ := s.userRepo.FindByUsername(ctx, username)
	if existingUser != nil {
		return nil, errors.New("username already exists")
	}

	// 2. 密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password failed: %w", err)
	}

	// 3. 创建用户
	user := &models.User{
		Username: username,
		Password: string(hashedPassword),
		Email:    email,
		IsActive: true,
		Role:     "user", // 默认角色
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user failed: %w", err)
	}

	// 4. 清除密码字段
	user.Password = ""

	return user, nil
}

// RefreshToken 刷新Token
func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*jwt.TokenPair, error) {
	// 1. 刷新Token
	tokenPair, err := jwt.RefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("refresh token failed: %w", err)
	}

	return tokenPair, nil
}

// Logout 用户登出
func (s *authService) Logout(ctx context.Context, userID uint) error {
	// 实际实现中可能需要将Token加入黑名单
	// 这里简化处理
	return nil
}

// ChangePassword 修改密码
func (s *authService) ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error {
	// 1. 查找用户
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// 2. 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("invalid old password")
	}

	// 3. 哈希新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password failed: %w", err)
	}

	// 4. 更新密码
	user.Password = string(hashedPassword)
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("update password failed: %w", err)
	}

	return nil
}

// GetUserByToken 通过Token获取用户信息
func (s *authService) GetUserByToken(ctx context.Context, token string) (*models.User, error) {
	// 1. 解析Token获取用户ID
	userID, err := jwt.GetUserIDFromToken(token)
	if err != nil {
		return nil, fmt.Errorf("parse token failed: %w", err)
	}

	// 2. 查找用户
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 3. 清除密码字段
	user.Password = ""

	return user, nil
}
