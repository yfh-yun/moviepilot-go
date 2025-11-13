package models

import (
	"context"
	
	"gorm.io/gorm"
	
	"moviepilot-go/internal/db"
)

// DbOper 数据库操作基类
type DbOper struct {
	DB *gorm.DB
}

// NewDbOper 创建数据库操作实例
func NewDbOper(database *gorm.DB) *DbOper {
	return &DbOper{
		DB: database,
	}
}

// NewDbOperWithDefault 使用默认数据库创建数据库操作实例
func NewDbOperWithDefault() *DbOper {
	return &DbOper{
		DB: db.GetDB(),
	}
}

// Transaction 执行事务操作
func (d *DbOper) Transaction(fc func(tx *gorm.DB) error) error {
	return d.DB.Transaction(fc)
}

// WithContext 给数据库操作添加上下文
func (d *DbOper) WithContext(ctx context.Context) *gorm.DB {
	return d.DB.WithContext(ctx)
}

// GetDB 获取数据库实例
func (d *DbOper) GetDB() *gorm.DB {
	return d.DB
}

// SetDB 设置数据库实例
func (d *DbOper) SetDB(database *gorm.DB) {
	d.DB = database
}