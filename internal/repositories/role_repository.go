package repositories

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"moviepilot-go/internal/models"
)

// RoleRepository 角色数据访问接口
type RoleRepository interface {
	Create(ctx context.Context, role *models.Role) error
	GetByID(ctx context.Context, id uint) (*models.Role, error)
	GetByName(ctx context.Context, name string) (*models.Role, error)
	Update(ctx context.Context, role *models.Role) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context) ([]*models.Role, error)
	GetWithPermissions(ctx context.Context, id uint) (*models.Role, error)
	AssignPermissions(ctx context.Context, roleID uint, permissionIDs []uint) error
}

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) Create(ctx context.Context, role *models.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *roleRepository) GetByID(ctx context.Context, id uint) (*models.Role, error) {
	var role models.Role
	err := r.db.WithContext(ctx).First(&role, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("角色不存在: id=%d", id)
		}
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) GetByName(ctx context.Context, name string) (*models.Role, error) {
	var role models.Role
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&role).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("角色不存在: name=%s", name)
		}
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) Update(ctx context.Context, role *models.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *roleRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Role{}, id).Error
}

func (r *roleRepository) List(ctx context.Context) ([]*models.Role, error) {
	var roles []*models.Role
	err := r.db.WithContext(ctx).Find(&roles).Error
	return roles, err
}

func (r *roleRepository) GetWithPermissions(ctx context.Context, id uint) (*models.Role, error) {
	var role models.Role
	err := r.db.WithContext(ctx).Preload("Permissions").First(&role, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("角色不存在: id=%d", id)
		}
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) AssignPermissions(ctx context.Context, roleID uint, permissionIDs []uint) error {
	role, err := r.GetByID(ctx, roleID)
	if err != nil {
		return err
	}

	var permissions []models.Permission
	if err := r.db.WithContext(ctx).Find(&permissions, permissionIDs).Error; err != nil {
		return err
	}

	return r.db.WithContext(ctx).Model(role).Association("Permissions").Replace(permissions)
}
