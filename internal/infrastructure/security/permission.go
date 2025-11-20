package security

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"moviepilot-go/pkg/logger"
)

// MemoryPermissionManager 内存权限管理器实现
type MemoryPermissionManager struct {
	permissions      map[string]*Permission
	permissionCodeMap map[string]*Permission
	categoryMap      map[string][]*Permission
	mutex            sync.RWMutex
	roleManager      *MemoryRoleManager
	logger           logger.Logger
}

// NewMemoryPermissionManager 创建内存权限管理器
func NewMemoryPermissionManager(roleManager *MemoryRoleManager, logger logger.Logger) *MemoryPermissionManager {
	return &MemoryPermissionManager{
		permissions:      make(map[string]*Permission),
		permissionCodeMap: make(map[string]*Permission),
		categoryMap:      make(map[string][]*Permission),
		roleManager:      roleManager,
		logger:           logger,
	}
}

// CreatePermission 创建权限
func (pm *MemoryPermissionManager) CreatePermission(perm *Permission) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// 检查权限代码是否已存在
	permCode := strings.ToLower(perm.Code)
	if _, exists := pm.permissionCodeMap[permCode]; exists {
		return errors.New("permission code already exists")
	}

	// 设置默认值
	if perm.CreatedAt.IsZero() {
		perm.CreatedAt = time.Now()
	}
	perm.UpdatedAt = time.Now()
	if perm.Category == "" {
		perm.Category = "default"
	}

	// 存储权限
	pm.permissions[perm.ID] = perm
	pm.permissionCodeMap[permCode] = perm

	// 按分类存储
	category := strings.ToLower(perm.Category)
	pm.categoryMap[category] = append(pm.categoryMap[category], perm)

	pm.logger.Info("Permission created successfully", "permission_id", perm.ID, "permission_code", perm.Code)
	return nil
}

// GetPermissionByID 根据ID获取权限
func (pm *MemoryPermissionManager) GetPermissionByID(id string) (*Permission, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	perm, exists := pm.permissions[id]
	if !exists {
		return nil, errors.New("permission not found")
	}

	// 返回权限副本
	return perm.clone(), nil
}

// GetPermissionByCode 根据代码获取权限
func (pm *MemoryPermissionManager) GetPermissionByCode(code string) (*Permission, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	perm, exists := pm.permissionCodeMap[strings.ToLower(code)]
	if !exists {
		return nil, errors.New("permission not found")
	}

	// 返回权限副本
	return perm.clone(), nil
}

// UpdatePermission 更新权限
func (pm *MemoryPermissionManager) UpdatePermission(perm *Permission) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	originalPerm, exists := pm.permissions[perm.ID]
	if !exists {
		return errors.New("permission not found")
	}

	// 检查权限代码是否被其他权限使用
	permCode := strings.ToLower(perm.Code)
	if perm.Code != originalPerm.Code {
		if existingPerm, exists := pm.permissionCodeMap[permCode]; exists && existingPerm.ID != perm.ID {
			return errors.New("permission code already exists")
		}
		delete(pm.permissionCodeMap, strings.ToLower(originalPerm.Code))
		pm.permissionCodeMap[permCode] = perm
	}

	// 检查分类是否变更
	originalCategory := strings.ToLower(originalPerm.Category)
	newCategory := strings.ToLower(perm.Category)
	if newCategory == "" {
		newCategory = "default"
	}

	if originalCategory != newCategory {
		// 从原分类移除
		newCategoryList := []*Permission{}
		for _, p := range pm.categoryMap[originalCategory] {
			if p.ID != perm.ID {
				newCategoryList = append(newCategoryList, p)
			}
		}
		if len(newCategoryList) == 0 {
			delete(pm.categoryMap, originalCategory)
		} else {
			pm.categoryMap[originalCategory] = newCategoryList
		}

		// 添加到新分类
		pm.categoryMap[newCategory] = append(pm.categoryMap[newCategory], perm)
	}

	// 更新时间
	perm.UpdatedAt = time.Now()
	if perm.CreatedAt.IsZero() {
		perm.CreatedAt = originalPerm.CreatedAt
	}

	// 更新权限
	pm.permissions[perm.ID] = perm

	pm.logger.Info("Permission updated successfully", "permission_id", perm.ID, "permission_code", perm.Code)
	return nil
}

// DeletePermission 删除权限
func (pm *MemoryPermissionManager) DeletePermission(id string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	perm, exists := pm.permissions[id]
	if !exists {
		return errors.New("permission not found")
	}

	// 删除权限引用
	delete(pm.permissions, id)
	delete(pm.permissionCodeMap, strings.ToLower(perm.Code))

	// 从分类中移除
	category := strings.ToLower(perm.Category)
	if categoryList, exists := pm.categoryMap[category]; exists {
		newCategoryList := []*Permission{}
		for _, p := range categoryList {
			if p.ID != id {
				newCategoryList = append(newCategoryList, p)
			}
		}
		if len(newCategoryList) == 0 {
			delete(pm.categoryMap, category)
		} else {
			pm.categoryMap[category] = newCategoryList
		}
	}

	pm.logger.Info("Permission deleted successfully", "permission_id", id, "permission_code", perm.Code)
	return nil
}

// ListPermissions 列出所有权限
func (pm *MemoryPermissionManager) ListPermissions() ([]*Permission, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	permList := make([]*Permission, 0, len(pm.permissions))
	for _, perm := range pm.permissions {
		permList = append(permList, perm.clone())
	}

	return permList, nil
}

// ListPermissionsByCategory 按分类列出权限
func (pm *MemoryPermissionManager) ListPermissionsByCategory(category string) ([]*Permission, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	category = strings.ToLower(category)
	perms, exists := pm.categoryMap[category]
	if !exists {
		return []*Permission{}, nil
	}

	// 返回权限副本
	permList := make([]*Permission, len(perms))
	for i, perm := range perms {
		permList[i] = perm.clone()
	}

	return permList, nil
}

// CheckPermission 检查用户是否拥有指定权限
func (pm *MemoryPermissionManager) CheckPermission(userID, permissionCode string) (bool, error) {
	// 获取用户角色
	roles, err := pm.roleManager.GetUserRoles(userID)
	if err != nil {
		return false, fmt.Errorf("failed to get user roles: %w", err)
	}

	// 检查每个角色是否包含该权限
	permissionCode = strings.ToLower(permissionCode)
	for _, role := range roles {
		for _, perm := range role.Permissions {
			if strings.ToLower(perm) == permissionCode {
				return true, nil
			}
		}
	}

	return false, nil
}

// CheckUserHasPermission 直接检查用户对象是否拥有指定权限
func (pm *MemoryPermissionManager) CheckUserHasPermission(user *User, permissionCode string) bool {
	if user == nil {
		return false
	}

	// 获取用户角色
	roles, err := pm.roleManager.GetUserRoles(user.ID)
	if err != nil {
		pm.logger.Error("Failed to get user roles", "user_id", user.ID, "error", err.Error())
		return false
	}

	// 检查每个角色是否包含该权限
	permissionCode = strings.ToLower(permissionCode)
	for _, role := range roles {
		for _, perm := range role.Permissions {
			if strings.ToLower(perm) == permissionCode {
				return true
			}
		}
	}

	return false
}

// GetUserPermissions 获取用户的所有权限
func (pm *MemoryPermissionManager) GetUserPermissions(userID string) ([]*Permission, error) {
	// 获取用户角色
	roles, err := pm.roleManager.GetUserRoles(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	// 收集所有权限
	permissionSet := make(map[string]*Permission)
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	for _, role := range roles {
		for _, permCode := range role.Permissions {
			if perm, exists := pm.permissionCodeMap[strings.ToLower(permCode)]; exists {
				permissionSet[perm.ID] = perm.clone()
			}
		}
	}

	// 转换为切片
	permissions := make([]*Permission, 0, len(permissionSet))
	for _, perm := range permissionSet {
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// GetRolePermissions 获取角色的所有权限详情
func (pm *MemoryPermissionManager) GetRolePermissions(roleID string) ([]*Permission, error) {
	// 获取角色
	role, err := pm.roleManager.GetRoleByID(roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	// 获取权限详情
	permissions := []*Permission{}
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	for _, permCode := range role.Permissions {
		if perm, exists := pm.permissionCodeMap[strings.ToLower(permCode)]; exists {
			permissions = append(permissions, perm.clone())
		}
	}

	return permissions, nil
}

// GetPermissionCategories 获取所有权限分类
func (pm *MemoryPermissionManager) GetPermissionCategories() []string {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	categories := make([]string, 0, len(pm.categoryMap))
	for category := range pm.categoryMap {
		categories = append(categories, category)
	}

	return categories
}

// SearchPermissions 搜索权限
func (pm *MemoryPermissionManager) SearchPermissions(keyword string) ([]*Permission, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	keyword = strings.ToLower(keyword)
	var results []*Permission

	for _, perm := range pm.permissions {
		if strings.Contains(strings.ToLower(perm.Code), keyword) ||
			strings.Contains(strings.ToLower(perm.Name), keyword) ||
			strings.Contains(strings.ToLower(perm.Description), keyword) {
			results = append(results, perm.clone())
		}
	}

	return results, nil
}

// ValidatePermissionCode 验证权限代码格式
func ValidatePermissionCode(code string) error {
	if code == "" {
		return errors.New("permission code cannot be empty")
	}

	// 权限代码格式：模块名:操作名 (例如：user:create, movie:read)
	if !strings.Contains(code, ":") {
		return errors.New("permission code must be in format 'module:action'")
	}

	parts := strings.Split(code, ":")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return errors.New("invalid permission code format")
	}

	// 检查模块名和操作名的字符
	for _, char := range parts[0] {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_') {
			return errors.New("module name can only contain letters, numbers and underscores")
		}
	}

	for _, char := range parts[1] {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_') {
			return errors.New("action name can only contain letters, numbers and underscores")
		}
	}

	return nil
}

// clone 克隆权限
func (p *Permission) clone() *Permission {
	clone := *p
	return &clone
}