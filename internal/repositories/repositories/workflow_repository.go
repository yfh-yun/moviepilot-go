package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"moviepilot-go/internal/models/database"
	"moviepilot-go/internal/repositories/interfaces"
	"moviepilot-go/pkg/cache"
)

// WorkflowRepositoryImpl 工作流仓储实现
type WorkflowRepositoryImpl struct {
	db    *gorm.DB
	cache cache.CacheBackend
}

// NewWorkflowRepository 创建工作流仓储实例
func NewWorkflowRepository(db *gorm.DB) interfaces.WorkflowRepository {
	return &WorkflowRepositoryImpl{
		db:    db,
		cache: cache.Cache("ttl", 1000, 3600),
	}
}

// Create 创建工作流
func (r *WorkflowRepositoryImpl) Create(ctx context.Context, workflow *database.Workflow) error {
	err := r.db.WithContext(ctx).Create(workflow).Error
	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("workflow")
	}
	return err
}

// GetByID 根据ID获取工作流
func (r *WorkflowRepositoryImpl) GetByID(ctx context.Context, id string) (*database.Workflow, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("workflow:id:%s", id)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "workflow")
		if err == nil && hit {
			if workflow, ok := cachedValue.(*database.Workflow); ok {
				return workflow, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var workflow database.Workflow
	err := r.db.WithContext(ctx).First(&workflow, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, &workflow, 3600*time.Second, "workflow")
	}

	return &workflow, nil
}

// Update 更新工作流
func (r *WorkflowRepositoryImpl) Update(ctx context.Context, workflow *database.Workflow) error {
	err := r.db.WithContext(ctx).Save(workflow).Error
	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("workflow")
	}
	return err
}

// Delete 删除工作流
func (r *WorkflowRepositoryImpl) Delete(ctx context.Context, id string) error {
	err := r.db.WithContext(ctx).Delete(&database.Workflow{}, "id = ?", id).Error
	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("workflow")
	}
	return err
}

// List 获取工作流列表
func (r *WorkflowRepositoryImpl) List(ctx context.Context, params interfaces.ListWorkflowParams) ([]*database.Workflow, int64, error) {
	// 生成缓存键，包含分页和过滤参数
	cacheKey := fmt.Sprintf("workflow:list:page:%d:page_size:%d:type:%s:status:%s:user_id:%s",
		params.Page, params.PageSize, params.Type, params.Status, params.UserID)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "workflow")
		if err == nil && hit {
			if cacheData, ok := cachedValue.(struct {
				Workflows []*database.Workflow
				Total     int64
			}); ok {
				return cacheData.Workflows, cacheData.Total, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var workflows []*database.Workflow
	var total int64

	query := r.db.WithContext(ctx).Model(&database.Workflow{})

	// 添加过滤条件
	if params.Type != "" {
		query = query.Where("type = ?", params.Type)
	}
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.UserID != "" {
		query = query.Where("user_id = ?", params.UserID)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (params.Page - 1) * params.PageSize
	err := query.Offset(offset).Limit(params.PageSize).Order("created_at DESC").Find(&workflows).Error
	if err != nil {
		return nil, 0, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, struct {
			Workflows []*database.Workflow
			Total     int64
		}{Workflows: workflows, Total: total}, 3600*time.Second, "workflow")
	}

	return workflows, total, err
}

// GetActiveWorkflows 获取活跃工作流
func (r *WorkflowRepositoryImpl) GetActiveWorkflows(ctx context.Context) ([]*database.Workflow, error) {
	// 生成缓存键
	cacheKey := "workflow:active"

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "workflow")
		if err == nil && hit {
			if workflows, ok := cachedValue.([]*database.Workflow); ok {
				return workflows, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var workflows []*database.Workflow
	err := r.db.WithContext(ctx).Where("status = ?", "active").Find(&workflows).Error
	if err != nil {
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, workflows, 3600*time.Second, "workflow")
	}

	return workflows, nil
}

// GetByType 根据类型获取工作流
func (r *WorkflowRepositoryImpl) GetByType(ctx context.Context, workflowType string) ([]*database.Workflow, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("workflow:type:%s", workflowType)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "workflow")
		if err == nil && hit {
			if workflows, ok := cachedValue.([]*database.Workflow); ok {
				return workflows, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var workflows []*database.Workflow
	err := r.db.WithContext(ctx).Where("type = ?", workflowType).Find(&workflows).Error
	if err != nil {
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, workflows, 3600*time.Second, "workflow")
	}

	return workflows, nil
}

// UpdateStatus 更新工作流状态
func (r *WorkflowRepositoryImpl) UpdateStatus(ctx context.Context, id, status string) error {
	err := r.db.WithContext(ctx).Model(&database.Workflow{}).Where("id = ?", id).Update("status", status).Error
	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("workflow")
	}
	return err
}
