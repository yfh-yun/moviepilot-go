package user

import (
	"context"

	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// PermissionService 权限服务接口
type PermissionService interface {
	// CheckPermission 检查权限
	CheckPermission(ctx context.Context, userID uint, resource, action string) (bool, error)

	// GrantPermission 授予权限
	GrantPermission(ctx context.Context, userID uint, resource, action string) error

	// RevokePermission 撤销权限
	RevokePermission(ctx context.Context, userID uint, resource, action string) error

	// GetUserPermissions 获取用户权限
	GetUserPermissions(ctx context.Context, userID uint) ([]Permission, error)

	// GetRolePermissions 获取角色权限
	GetRolePermissions(ctx context.Context, role string) ([]Permission, error)
}

// permissionService 权限服务实现
type permissionService struct {
	rolePermissions map[string][]Permission
	logger          *zap.Logger
}

// NewPermissionService 创建权限服务
func NewPermissionService() PermissionService {
	service := &permissionService{
		rolePermissions: make(map[string][]Permission),
		logger:          logger.GetLogger(),
	}

	// 初始化默认角色权限
	service.initDefaultPermissions()

	return service
}

// Permission 权限
type Permission struct {
	Resource string `json:"resource"` // 资源：subscribe, site, download, etc.
	Action   string `json:"action"`   // 操作：read, write, delete, execute
}

// Role 角色
type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
	RoleGuest Role = "guest"
)

// CheckPermission 检查权限
func (s *permissionService) CheckPermission(ctx context.Context, userID uint, resource, action string) (bool, error) {
	s.logger.Debug("检查权限",
		zap.Uint("user_id", userID),
		zap.String("resource", resource),
		zap.String("action", action),
	)

	// TODO: 从数据库获取用户角色
	role := "admin" // 简化版，默认管理员

	// 获取角色权限
	permissions, err := s.GetRolePermissions(ctx, role)
	if err != nil {
		return false, err
	}

	// 检查是否有权限
	for _, perm := range permissions {
		if perm.Resource == resource && perm.Action == action {
			return true, nil
		}
		// 检查通配符权限
		if perm.Resource == "*" || perm.Action == "*" {
			return true, nil
		}
	}

	return false, nil
}

// GrantPermission 授予权限
func (s *permissionService) GrantPermission(ctx context.Context, userID uint, resource, action string) error {
	s.logger.Info("授予权限",
		zap.Uint("user_id", userID),
		zap.String("resource", resource),
		zap.String("action", action),
	)

	// TODO: 保存到数据库
	return nil
}

// RevokePermission 撤销权限
func (s *permissionService) RevokePermission(ctx context.Context, userID uint, resource, action string) error {
	s.logger.Info("撤销权限",
		zap.Uint("user_id", userID),
		zap.String("resource", resource),
		zap.String("action", action),
	)

	// TODO: 从数据库删除
	return nil
}

// GetUserPermissions 获取用户权限
func (s *permissionService) GetUserPermissions(ctx context.Context, userID uint) ([]Permission, error) {
	// TODO: 从数据库获取用户角色和权限
	role := "admin"
	return s.GetRolePermissions(ctx, role)
}

// GetRolePermissions 获取角色权限
func (s *permissionService) GetRolePermissions(ctx context.Context, role string) ([]Permission, error) {
	if permissions, ok := s.rolePermissions[role]; ok {
		return permissions, nil
	}
	return []Permission{}, nil
}

// initDefaultPermissions 初始化默认权限
func (s *permissionService) initDefaultPermissions() {
	// 管理员权限（所有权限）
	s.rolePermissions["admin"] = []Permission{
		{Resource: "*", Action: "*"},
	}

	// 普通用户权限
	s.rolePermissions["user"] = []Permission{
		{Resource: "subscribe", Action: "read"},
		{Resource: "subscribe", Action: "write"},
		{Resource: "search", Action: "read"},
		{Resource: "download", Action: "read"},
		{Resource: "site", Action: "read"},
	}

	// 访客权限（只读）
	s.rolePermissions["guest"] = []Permission{
		{Resource: "subscribe", Action: "read"},
		{Resource: "search", Action: "read"},
		{Resource: "download", Action: "read"},
	}
}

// HasPermission 辅助函数：检查权限
func HasPermission(permissions []Permission, resource, action string) bool {
	for _, perm := range permissions {
		if (perm.Resource == resource || perm.Resource == "*") &&
			(perm.Action == action || perm.Action == "*") {
			return true
		}
	}
	return false
}
