package repositories

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"moviepilot-go/internal/models"
)

// PermissionRepository 权限数据访问接口
type PermissionRepository interface {
	Create(ctx context.Context, permission *models.Permission) error
	GetByID(ctx context.Context, id uint) (*models.Permission, error)
	GetByName(ctx context.Context, name string) (*models.Permission, error)
	Update(ctx context.Context, permission *models.Permission) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context) ([]*models.Permission, error)
	GetByResource(ctx context.Context, resource string) ([]*models.Permission, error)
	GetByRoleID(ctx context.Context, roleID uint) ([]*models.Permission, error)
}

type permissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) PermissionRepository {
	return &permissionRepository{db: db}
}

func (r *permissionRepository) Create(ctx context.Context, permission *models.Permission) error {
	return r.db.WithContext(ctx).Create(permission).Error
}

func (r *permissionRepository) GetByID(ctx context.Context, id uint) (*models.Permission, error) {
	var permission models.Permission
	err := r.db.WithContext(ctx).First(&permission, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("权限不存在: id=%d", id)
		}
		return nil, err
	}
	return &permission, nil
}

func (r *permissionRepository) GetByName(ctx context.Context, name string) (*models.Permission, error) {
	var permission models.Permission
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&permission).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("权限不存在: name=%s", name)
		}
		return nil, err
	}
	return &permission, nil
}

func (r *permissionRepository) Update(ctx context.Context, permission *models.Permission) error {
	return r.db.WithContext(ctx).Save(permission).Error
}

func (r *permissionRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Permission{}, id).Error
}

func (r *permissionRepository) List(ctx context.Context) ([]*models.Permission, error) {
	var permissions []*models.Permission
	err := r.db.WithContext(ctx).Find(&permissions).Error
	return permissions, err
}

func (r *permissionRepository) GetByResource(ctx context.Context, resource string) ([]*models.Permission, error) {
	var permissions []*models.Permission
	err := r.db.WithContext(ctx).Where("resource = ?", resource).Find(&permissions).Error
	return permissions, err
}

func (r *permissionRepository) GetByRoleID(ctx context.Context, roleID uint) ([]*models.Permission, error) {
	var permissions []*models.Permission
	err := r.db.WithContext(ctx).
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ?", roleID).
		Find(&permissions).Error
	return permissions, err
}
