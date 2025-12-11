package user

import (
	"context"

	"moviepilot-go/internal/models/database"
	"moviepilot-go/internal/repositories/interfaces"
	"moviepilot-go/pkg/errors"
	"moviepilot-go/pkg/logger"

	"go.uber.org/zap"
)

// Service 定义用户业务操作接口
type Service interface {
	CreateUser(ctx context.Context, user *database.User) (*database.User, error)
	GetUserByID(ctx context.Context, id string) (*database.User, error)
	GetUserByUsername(ctx context.Context, username string) (*database.User, error)
	ListUsers(ctx context.Context, params interfaces.ListUserParams) ([]*database.User, int64, error)
	UpdateUser(ctx context.Context, user *database.User) (*database.User, error)
	DeleteUser(ctx context.Context, id string) error
	UpdatePassword(ctx context.Context, userID, password string) error
	UpdateLastLogin(ctx context.Context, userID string) error
}

type service struct {
	repo interfaces.UserRepository
}

// NewService 创建用户服务实例
func NewService(repo interfaces.UserRepository) Service {
	if repo == nil {
		panic("user repository cannot be nil")
	}
	return &service{repo: repo}
}

func (s *service) CreateUser(ctx context.Context, user *database.User) (*database.User, error) {
	log := logger.WithContext(ctx)
	log.Debug("Creating user", zap.String("name", user.Name))

	if err := s.repo.Create(ctx, user); err != nil {
		log.Error("Create user failed", zap.Error(err))
		return nil, errors.WrapError(err, "create user failed")
	}

	log.Info("User created", zap.String("name", user.Name), zap.Uint("user_id", user.ID))
	return user, nil
}

func (s *service) GetUserByID(ctx context.Context, id string) (*database.User, error) {
	log := logger.WithContext(ctx)
	log.Debug("Fetching user by id", zap.String("user_id", id))

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		log.Error("Get user by id failed", zap.String("user_id", id), zap.Error(err))
		return nil, errors.WrapError(err, "get user by id failed")
	}

	return user, nil
}

func (s *service) GetUserByUsername(ctx context.Context, username string) (*database.User, error) {
	log := logger.WithContext(ctx)
	log.Debug("Fetching user by username", zap.String("username", username))

	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		log.Error("Get user by username failed", zap.String("username", username), zap.Error(err))
		return nil, errors.WrapError(err, "get user by username failed")
	}

	return user, nil
}

func (s *service) ListUsers(ctx context.Context, params interfaces.ListUserParams) ([]*database.User, int64, error) {
	log := logger.WithContext(ctx)
	log.Debug("Listing users", zap.Any("params", params))

	users, total, err := s.repo.List(ctx, params)
	if err != nil {
		log.Error("List users failed", zap.Error(err))
		return nil, 0, errors.WrapError(err, "list users failed")
	}

	return users, total, nil
}

func (s *service) UpdateUser(ctx context.Context, user *database.User) (*database.User, error) {
	log := logger.WithContext(ctx)
	log.Debug("Updating user", zap.Uint("user_id", user.ID))

	if err := s.repo.Update(ctx, user); err != nil {
		log.Error("Update user failed", zap.Uint("user_id", user.ID), zap.Error(err))
		return nil, errors.WrapError(err, "update user failed")
	}

	log.Info("User updated", zap.Uint("user_id", user.ID))
	return user, nil
}

func (s *service) DeleteUser(ctx context.Context, id string) error {
	log := logger.WithContext(ctx)
	log.Debug("Deleting user", zap.String("user_id", id))

	if err := s.repo.Delete(ctx, id); err != nil {
		log.Error("Delete user failed", zap.String("user_id", id), zap.Error(err))
		return errors.WrapError(err, "delete user failed")
	}

	log.Info("User deleted", zap.String("user_id", id))
	return nil
}

func (s *service) UpdatePassword(ctx context.Context, userID, password string) error {
	log := logger.WithContext(ctx)
	log.Debug("Updating user password", zap.String("user_id", userID))

	if err := s.repo.UpdatePassword(ctx, userID, password); err != nil {
		log.Error("Update password failed", zap.String("user_id", userID), zap.Error(err))
		return errors.WrapError(err, "update password failed")
	}

	log.Info("User password updated", zap.String("user_id", userID))
	return nil
}

func (s *service) UpdateLastLogin(ctx context.Context, userID string) error {
	log := logger.WithContext(ctx)
	log.Debug("Updating last login", zap.String("user_id", userID))

	if err := s.repo.UpdateLastLogin(ctx, userID); err != nil {
		log.Error("Update last login failed", zap.String("user_id", userID), zap.Error(err))
		return errors.WrapError(err, "update last login failed")
	}

	log.Info("User last login updated", zap.String("user_id", userID))
	return nil
}
