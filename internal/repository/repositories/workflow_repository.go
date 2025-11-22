package repositories

import (
	"context"
	"errors"
	
	"gorm.io/gorm"
	
	"moviepilot-go/internal/repository/interfaces"
	"moviepilot-go/internal/models"
)

// WorkflowRepositoryImpl 工作流仓储实现
type WorkflowRepositoryImpl struct {
	db *gorm.DB
}

// NewWorkflowRepository 创建工作流仓储实例
func NewWorkflowRepository(db *gorm.DB) interfaces.WorkflowRepository {
	return &WorkflowRepositoryImpl{db: db}
}

// Create 创建工作流
func (r *WorkflowRepositoryImpl) Create(ctx context.Context, workflow *models.Workflow) error {
	return r.db.WithContext(ctx).Create(workflow).Error
}

// GetByID 根据ID获取工作流
func (r *WorkflowRepositoryImpl) GetByID(ctx context.Context, id string) (*models.Workflow, error) {
	var workflow models.Workflow
	err := r.db.WithContext(ctx).First(&workflow, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &workflow, nil
}

// Update 更新工作流
func (r *WorkflowRepositoryImpl) Update(ctx context.Context, workflow *models.Workflow) error {
	return r.db.WithContext(ctx).Save(workflow).Error
}

// Delete 删除工作流
func (r *WorkflowRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Workflow{}, "id = ?", id).Error
}

// List 获取工作流列表
func (r *WorkflowRepositoryImpl) List(ctx context.Context, params interfaces.ListWorkflowParams) ([]*models.Workflow, int64, error) {
	var workflows []*models.Workflow
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.Workflow{})
	
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
	
	return workflows, total, err
}

// GetActiveWorkflows 获取活跃工作流
func (r *WorkflowRepositoryImpl) GetActiveWorkflows(ctx context.Context) ([]*models.Workflow, error) {
	var workflows []*models.Workflow
	err := r.db.WithContext(ctx).Where("status = ?", "active").Find(&workflows).Error
	return workflows, err
}

// GetByType 根据类型获取工作流
func (r *WorkflowRepositoryImpl) GetByType(ctx context.Context, workflowType string) ([]*models.Workflow, error) {
	var workflows []*models.Workflow
	err := r.db.WithContext(ctx).Where("type = ?", workflowType).Find(&workflows).Error
	return workflows, err
}

// UpdateStatus 更新工作流状态
func (r *WorkflowRepositoryImpl) UpdateStatus(ctx context.Context, id, status string) error {
	return r.db.WithContext(ctx).Model(&models.Workflow{}).Where("id = ?", id).Update("status", status).Error
}