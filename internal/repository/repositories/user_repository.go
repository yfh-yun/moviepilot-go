package repositories

import (
	"errors"
	"github.com/yfh-yun/moviepilot-go/internal/repository/interfaces"
	"github.com/yfh-yun/moviepilot-go/internal/model"

	"gorm.io/gorm"
)

// userRepository 用户仓储实现
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓储
func NewUserRepository(db *gorm.DB) interfaces.UserRepository {
	return &userRepository{db: db}
}

// Create 创建用户
func (r *userRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// Update 更新用户
func (r *userRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}

// Delete 删除用户
func (r *userRepository) Delete(id uint) error {
	return r.db.Delete(&model.User{}, id).Error
}

// GetByID 根据ID获取用户
func (r *userRepository) GetByID(id uint) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByUsername 根据用户名获取用户
func (r *userRepository) FindByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Where("name = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetByEmail 根据邮箱获取用户
func (r *userRepository) GetByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// List 获取用户列表
func (r *userRepository) List(offset, limit int) ([]*model.User, int64, error) {
	var users []*model.User
	var total int64

	err := r.db.Model(&model.User{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Offset(offset).Limit(limit).Find(&users).Error
	return users, total, err
}

// UpdateOTP 更新OTP设置
func (r *userRepository) UpdateOTP(id uint, isOTP bool, secret string) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_otp":     isOTP,
		"otp_secret": secret,
	}).Error
}

// UpdatePassword 更新密码
func (r *userRepository) UpdatePassword(id uint, passwordHash string) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).Update("hashed_password", passwordHash).Error
}

// UpdateAvatar 更新头像
func (r *userRepository) UpdateAvatar(id uint, avatar string) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).Update("avatar", avatar).Error
}

// UpdateSettings 更新用户设置
func (r *userRepository) UpdateSettings(id uint, settings string) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).Update("settings", settings).Error
}

// UpdatePermissions 更新用户权限
func (r *userRepository) UpdatePermissions(id uint, permissions string) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).Update("permissions", permissions).Error
}

// Count 统计用户数量
func (r *userRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.User{}).Count(&count).Error
	return count, err
}

// Exists 检查用户名是否存在
func (r *userRepository) Exists(username string) (bool, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("name = ?", username).Count(&count).Error
	return count > 0, err
}

// ExistsEmail 检查邮箱是否存在
func (r *userRepository) ExistsEmail(email string) (bool, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}
