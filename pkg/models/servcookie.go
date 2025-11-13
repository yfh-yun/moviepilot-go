package models

import (
	"errors"
	"regexp"
)

// CookieData Cookie数据模型
type CookieData struct {
	// 加密数据
	Encrypted string `json:"encrypted"`
	// UUID
	UUID string `json:"uuid"`
}

// CookiePassword Cookie密码模型
type CookiePassword struct {
	// 密码
	Password string `json:"password"`
}

// Validate 验证CookieData数据
func (c *CookieData) Validate() error {
	// 验证encrypted字段
	if len(c.Encrypted) < 1 || len(c.Encrypted) > 1024*1024*50 {
		return errors.New("encrypted length must be between 1 and 52428800")
	}

	// 验证uuid字段
	if len(c.UUID) < 5 {
		return errors.New("uuid length must be at least 5 characters")
	}

	// 验证uuid格式（只包含字母和数字）
	uuidPattern := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	if !uuidPattern.MatchString(c.UUID) {
		return errors.New("uuid must contain only alphanumeric characters")
	}

	return nil
}

// Validate 验证CookiePassword数据
func (c *CookiePassword) Validate() error {
	// 简单验证，确保密码不为�?	if len(c.Password) == 0 {
		return errors.New("password cannot be empty")
	}
	return nil
}

// NewCookieData 创建一个新�?CookieData 实例
func NewCookieData(encrypted, uuid string) *CookieData {
	return &CookieData{
		Encrypted: encrypted,
		UUID:      uuid,
	}
}

// NewCookiePassword 创建一个新�?CookiePassword 实例
func NewCookiePassword(password string) *CookiePassword {
	return &CookiePassword{
		Password: password,
	}
}
