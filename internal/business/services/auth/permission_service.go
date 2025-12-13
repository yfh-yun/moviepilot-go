package auth

import (
	"context"
	"fmt"

	"moviepilot-go/internal/models"
	"moviepilot-go/internal/repositories"
)

// PermissionService 权限服务接口
type PermissionService interface {
	// CheckPermission 检查用户是否有指定权限
	CheckPermission(ctx context.Context, userID uint, permissionName string) (bool, error)
	// CheckPermissions 检查用户是否有多个权限（需要全部拥有）
	CheckPermissions(ctx context.Context, userID uint, permissionNames []string) (bool, error)
	// CheckAnyPermission 检查用户是否有任意一个权限
	CheckAnyPermission(ctx context.Context, userID uint, permissionNames []string) (bool, error)
	// GetUserPermissions 获取用户的所有权限
	GetUserPermissions(ctx context.Context, userID uint) ([]*models.Permission, error)
	// HasRole 检查用户是否有指定角色
	HasRole(ctx context.Context, userID uint, roleName string) (bool, error)
	// IsAdmin 检查用户是否是管理员
	IsAdmin(ctx context.Context, userID uint) (bool, error)
}

// permissionService 权限服务实现
type permissionService struct {
	userRepo       repositories.UserRepository
	permissionRepo repositories.PermissionRepository
}

// NewPermissionService 创建权限服务
func NewPermissionService(
	userRepo repositories.UserRepository,
	permissionRepo repositories.PermissionRepository,
) PermissionService {
	return &permissionService{
		userRepo:       userRepo,
		permissionRepo: permissionRepo,
	}
}

// CheckPermission 检查用户是否有指定权限
func (s *permissionService) CheckPermission(ctx context.Context, userID uint, permissionName string) (bool, error) {
	// 获取用户及其角色
	user, err := s.userRepo.GetWithRoles(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("获取用户失败: %w", err)
	}

	// 检查用户状态
	if !user.IsActiveUser() {
		return false, fmt.Errorf("用户已被禁用")
	}

	// 遍历用户的所有角色
	for _, role := range user.Roles {
		// 获取角色的权限
		permissions, err := s.permissionRepo.GetByRoleID(ctx, role.ID)
		if err != nil {
			continue
		}

		// 检查是否有目标权限
		for _, permission := range permissions {
			if permission.Name == permissionName {
				return true, nil
			}
		}
	}

	return false, nil
}

// CheckPermissions 检查用户是否有多个权限（需要全部拥有）
func (s *permissionService) CheckPermissions(ctx context.Context, userID uint, permissionNames []string) (bool, error) {
	for _, permissionName := range permissionNames {
		hasPermission, err := s.CheckPermission(ctx, userID, permissionName)
		if err != nil {
			return false, err
		}
		if !hasPermission {
			return false, nil
		}
	}
	return true, nil
}

// CheckAnyPermission 检查用户是否有任意一个权限
func (s *permissionService) CheckAnyPermission(ctx context.Context, userID uint, permissionNames []string) (bool, error) {
	for _, permissionName := range permissionNames {
		hasPermission, err := s.CheckPermission(ctx, userID, permissionName)
		if err != nil {
			return false, err
		}
		if hasPermission {
			return true, nil
		}
	}
	return false, nil
}

// GetUserPermissions 获取用户的所有权限
func (s *permissionService) GetUserPermissions(ctx context.Context, userID uint) ([]*models.Permission, error) {
	// 获取用户及其角色
	user, err := s.userRepo.GetWithRoles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户失败: %w", err)
	}

	// 使用 map 去重
	permissionMap := make(map[uint]*models.Permission)

	// 遍历用户的所有角色
	for _, role := range user.Roles {
		// 获取角色的权限
		permissions, err := s.permissionRepo.GetByRoleID(ctx, role.ID)
		if err != nil {
			continue
		}

		// 添加到 map 中
		for _, permission := range permissions {
			permissionMap[permission.ID] = permission
		}
	}

	// 转换为切片
	result := make([]*models.Permission, 0, len(permissionMap))
	for _, permission := range permissionMap {
		result = append(result, permission)
	}

	return result, nil
}

// HasRole 检查用户是否有指定角色
func (s *permissionService) HasRole(ctx context.Context, userID uint, roleName string) (bool, error) {
	// 获取用户及其角色
	user, err := s.userRepo.GetWithRoles(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("获取用户失败: %w", err)
	}

	// 检查用户状态
	if !user.IsActiveUser() {
		return false, fmt.Errorf("用户已被禁用")
	}

	// 遍历用户的所有角色
	for _, role := range user.Roles {
		if role.Name == roleName {
			return true, nil
		}
	}

	return false, nil
}

// IsAdmin 检查用户是否是管理员
func (s *permissionService) IsAdmin(ctx context.Context, userID uint) (bool, error) {
	return s.HasRole(ctx, userID, "admin")
}
