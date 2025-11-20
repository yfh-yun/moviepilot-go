package security

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"moviepilot-go/pkg/logger"
)

// MemoryRoleManager 内存角色管理器实现
type MemoryRoleManager struct {
	roles      map[string]*Role
	roleNameMap map[string]*Role
	userRoles  map[string][]string // userID -> []roleID
	mutex      sync.RWMutex
	logger     logger.Logger
}

// NewMemoryRoleManager 创建内存角色管理器
func NewMemoryRoleManager(logger logger.Logger) *MemoryRoleManager {
	return &MemoryRoleManager{
		roles:      make(map[string]*Role),
		roleNameMap: make(map[string]*Role),
		userRoles:  make(map[string][]string),
		logger:     logger,
	}
}

// CreateRole 创建角色
func (rm *MemoryRoleManager) CreateRole(role *Role) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	// 检查角色名是否已存在
	roleName := strings.ToLower(role.Name)
	if _, exists := rm.roleNameMap[roleName]; exists {
		return errors.New("role name already exists")
	}

	// 设置默认值
	if role.CreatedAt.IsZero() {
		role.CreatedAt = time.Now()
	}
	role.UpdatedAt = time.Now()
	if role.Permissions == nil {
		role.Permissions = []string{}
	}

	// 存储角色
	rm.roles[role.ID] = role
	rm.roleNameMap[roleName] = role

	rm.logger.Info("Role created successfully", "role_id", role.ID, "role_name", role.Name)
	return nil
}

// GetRoleByID 根据ID获取角色
func (rm *MemoryRoleManager) GetRoleByID(id string) (*Role, error) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	role, exists := rm.roles[id]
	if !exists {
		return nil, errors.New("role not found")
	}

	// 返回角色副本
	return role.clone(), nil
}

// GetRoleByName 根据名称获取角色
func (rm *MemoryRoleManager) GetRoleByName(name string) (*Role, error) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	role, exists := rm.roleNameMap[strings.ToLower(name)]
	if !exists {
		return nil, errors.New("role not found")
	}

	// 返回角色副本
	return role.clone(), nil
}

// UpdateRole 更新角色
func (rm *MemoryRoleManager) UpdateRole(role *Role) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	originalRole, exists := rm.roles[role.ID]
	if !exists {
		return errors.New("role not found")
	}

	// 检查角色名是否被其他角色使用
	roleName := strings.ToLower(role.Name)
	if role.Name != originalRole.Name {
		if existingRole, exists := rm.roleNameMap[roleName]; exists && existingRole.ID != role.ID {
			return errors.New("role name already exists")
		}
		delete(rm.roleNameMap, strings.ToLower(originalRole.Name))
		rm.roleNameMap[roleName] = role
	}

	// 更新时间
	role.UpdatedAt = time.Now()
	if role.CreatedAt.IsZero() {
		role.CreatedAt = originalRole.CreatedAt
	}
	if role.Permissions == nil {
		role.Permissions = []string{}
	}

	// 更新角色
	rm.roles[role.ID] = role

	rm.logger.Info("Role updated successfully", "role_id", role.ID, "role_name", role.Name)
	return nil
}

// DeleteRole 删除角色
func (rm *MemoryRoleManager) DeleteRole(id string) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	role, exists := rm.roles[id]
	if !exists {
		return errors.New("role not found")
	}

	// 删除角色引用
	delete(rm.roles, id)
	delete(rm.roleNameMap, strings.ToLower(role.Name))

	// 从所有用户中移除该角色
	for userID, roleIDs := range rm.userRoles {
		newRoleIDs := []string{}
		for _, roleID := range roleIDs {
			if roleID != id {
				newRoleIDs = append(newRoleIDs, roleID)
			}
		}
		if len(newRoleIDs) == 0 {
			delete(rm.userRoles, userID)
		} else {
			rm.userRoles[userID] = newRoleIDs
		}
	}

	rm.logger.Info("Role deleted successfully", "role_id", id, "role_name", role.Name)
	return nil
}

// ListRoles 列出所有角色
func (rm *MemoryRoleManager) ListRoles() ([]*Role, error) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	roleList := make([]*Role, 0, len(rm.roles))
	for _, role := range rm.roles {
		roleList = append(roleList, role.clone())
	}

	return roleList, nil
}

// AddPermissionToRole 为角色添加权限
func (rm *MemoryRoleManager) AddPermissionToRole(roleID, permissionCode string) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	role, exists := rm.roles[roleID]
	if !exists {
		return errors.New("role not found")
	}

	// 检查权限是否已存在
	for _, perm := range role.Permissions {
		if strings.EqualFold(perm, permissionCode) {
			return nil // 权限已存在，无需添加
		}
	}

	// 添加权限
	role.Permissions = append(role.Permissions, permissionCode)
	role.UpdatedAt = time.Now()

	rm.logger.Info("Permission added to role", "role_id", roleID, "permission_code", permissionCode)
	return nil
}

// RemovePermissionFromRole 从角色中移除权限
func (rm *MemoryRoleManager) RemovePermissionFromRole(roleID, permissionCode string) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	role, exists := rm.roles[roleID]
	if !exists {
		return errors.New("role not found")
	}

	// 移除权限
	newPermissions := []string{}
	removed := false
	for _, perm := range role.Permissions {
		if !strings.EqualFold(perm, permissionCode) {
			newPermissions = append(newPermissions, perm)
		} else {
			removed = true
		}
	}

	if removed {
		role.Permissions = newPermissions
		role.UpdatedAt = time.Now()
		rm.logger.Info("Permission removed from role", "role_id", roleID, "permission_code", permissionCode)
	}

	return nil
}

// GetUserRoles 获取用户的所有角色
func (rm *MemoryRoleManager) GetUserRoles(userID string) ([]*Role, error) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	roleIDs, exists := rm.userRoles[userID]
	if !exists {
		return []*Role{}, nil
	}

	roles := []*Role{}
	for _, roleID := range roleIDs {
		if role, exists := rm.roles[roleID]; exists {
			roles = append(roles, role.clone())
		}
	}

	return roles, nil
}

// AssignRoleToUser 为用户分配角色
func (rm *MemoryRoleManager) AssignRoleToUser(userID, roleID string) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	// 检查角色是否存在
	if _, exists := rm.roles[roleID]; !exists {
		return errors.New("role not found")
	}

	// 检查用户是否已拥有该角色
	if roleIDs, exists := rm.userRoles[userID]; exists {
		for _, rid := range roleIDs {
			if rid == roleID {
				return nil // 用户已拥有该角色，无需分配
			}
		}
		// 添加角色
		rm.userRoles[userID] = append(roleIDs, roleID)
	} else {
		// 用户首次分配角色
		rm.userRoles[userID] = []string{roleID}
	}

	rm.logger.Info("Role assigned to user", "user_id", userID, "role_id", roleID)
	return nil
}

// RemoveRoleFromUser 从用户移除角色
func (rm *MemoryRoleManager) RemoveRoleFromUser(userID, roleID string) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	// 检查用户是否拥有角色
	if roleIDs, exists := rm.userRoles[userID]; exists {
		newRoleIDs := []string{}
		removed := false
		for _, rid := range roleIDs {
			if rid != roleID {
				newRoleIDs = append(newRoleIDs, rid)
			} else {
				removed = true
			}
		}

		if removed {
			if len(newRoleIDs) == 0 {
				delete(rm.userRoles, userID)
			} else {
				rm.userRoles[userID] = newRoleIDs
			}
			rm.logger.Info("Role removed from user", "user_id", userID, "role_id", roleID)
		}
	}

	return nil
}

// IsUserInRole 检查用户是否拥有指定角色
func (rm *MemoryRoleManager) IsUserInRole(userID, roleID string) bool {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	if roleIDs, exists := rm.userRoles[userID]; exists {
		for _, rid := range roleIDs {
			if rid == roleID {
				return true
			}
		}
	}

	return false
}

// IsUserInRoleByName 检查用户是否拥有指定名称的角色
func (rm *MemoryRoleManager) IsUserInRoleByName(userID, roleName string) bool {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	// 获取角色
	role, exists := rm.roleNameMap[strings.ToLower(roleName)]
	if !exists {
		return false
	}

	// 检查用户是否拥有该角色
	if roleIDs, exists := rm.userRoles[userID]; exists {
		for _, rid := range roleIDs {
			if rid == role.ID {
				return true
			}
		}
	}

	return false
}

// GetUsersInRole 获取拥有指定角色的所有用户ID
func (rm *MemoryRoleManager) GetUsersInRole(roleID string) []string {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	var userIDs []string
	for userID, roleIDs := range rm.userRoles {
		for _, rid := range roleIDs {
			if rid == roleID {
				userIDs = append(userIDs, userID)
				break
			}
		}
	}

	return userIDs
}

// GetRolePermissionMap 获取角色权限映射
func (rm *MemoryRoleManager) GetRolePermissionMap() map[string][]string {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	result := make(map[string][]string)
	for _, role := range rm.roles {
		permissions := make([]string, len(role.Permissions))
		copy(permissions, role.Permissions)
		result[role.ID] = permissions
	}

	return result
}

// clone 克隆角色
func (r *Role) clone() *Role {
	clone := *r
	// 深拷贝权限列表
	clone.Permissions = make([]string, len(r.Permissions))
	copy(clone.Permissions, r.Permissions)
	return &clone
}