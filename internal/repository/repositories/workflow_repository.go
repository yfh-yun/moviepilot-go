package repositories

import (
	"errors"
	"fmt"
	"github.com/yfh-yun/moviepilot-go/internal/repository/interfaces"
	"github.com/yfh-yun/moviepilot-go/internal/model"
	"time"

	"gorm.io/gorm"
)

// WorkflowRepository 工作流仓储实现
type WorkflowRepository struct {
	db *gorm.DB
}

// NewWorkflowRepository 创建工作流仓储实例
func NewWorkflowRepository(db *gorm.DB) interfaces.WorkflowRepository {
	return &model.WorkflowRepository{db: db}
}

// Create 创建工作流
func (r *WorkflowRepository) Create(workflow *model.Workflow) error {
	if workflow == nil {
		return errors.New("workflow cannot be nil")
	}

	// 检查名称是否已存在
	var existingWorkflow model.Workflow
	if err := r.db.Where("name = ?", workflow.Name).First(&existingWorkflow).Error; err == nil {
		return fmt.Errorf("workflow with name '%s' already exists", workflow.Name)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to check workflow name: %w", err)
	}

	return r.db.Create(workflow).Error
}

// GetByID 根据ID获取工作流
func (r *WorkflowRepository) GetByID(id uint) (*model.Workflow, error) {
	var workflow model.Workflow
	err := r.db.First(&workflow, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow by id %d: %w", id, err)
	}
	return &workflow, nil
}

// GetByName 根据名称获取工作流
func (r *WorkflowRepository) GetByName(name string) (*model.Workflow, error) {
	var workflow model.Workflow
	err := r.db.Where("name = ?", name).First(&workflow).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow by name '%s': %w", name, err)
	}
	return &workflow, nil
}

// GetByType 根据类型获取工作流列表
func (r *WorkflowRepository) GetByType(workflowType string) ([]*model.Workflow, error) {
	var workflows []*model.Workflow
	err := r.db.Where("type = ?", workflowType).Find(&workflows).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get workflows by type '%s': %w", workflowType, err)
	}
	return workflows, nil
}

// GetActive 获取活跃的工作流列表
func (r *WorkflowRepository) GetActive() ([]*model.Workflow, error) {
	var workflows []*model.Workflow
	err := r.db.Where("enabled = ?", true).Find(&workflows).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get active workflows: %w", err)
	}
	return workflows, nil
}

// GetByTrigger 根据触发器获取工作流列表
func (r *WorkflowRepository) GetByTrigger(trigger string) ([]*model.Workflow, error) {
	var workflows []*model.Workflow
	err := r.db.Where("trigger = ? AND enabled = ?", trigger, true).Find(&workflows).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get workflows by trigger '%s': %w", trigger, err)
	}
	return workflows, nil
}

// Update 更新工作流
func (r *WorkflowRepository) Update(workflow *model.Workflow) error {
	if workflow == nil {
		return errors.New("workflow cannot be nil")
	}

	// 检查名称是否与其他工作流冲突
	var existingWorkflow model.Workflow
	err := r.db.Where("name = ? AND id != ?", workflow.Name, workflow.ID).First(&existingWorkflow).Error
	if err == nil {
		return fmt.Errorf("workflow with name '%s' already exists", workflow.Name)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to check workflow name: %w", err)
	}

	return r.db.Save(workflow).Error
}

// Delete 删除工作流
func (r *WorkflowRepository) Delete(id uint) error {
	result := r.db.Delete(&model.Workflow{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete workflow %d: %w", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("workflow with id %d not found", id)
	}
	return nil
}

// List 分页获取工作流列表
func (r *WorkflowRepository) List(offset, limit int) ([]*model.Workflow, int64, error) {
	var workflows []*model.Workflow
	var total int64

	// 获取总数
	if err := r.db.Model(&model.Workflow{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count workflows: %w", err)
	}

	// 获取分页数据
	if err := r.db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&workflows).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list workflows: %w", err)
	}

	return workflows, total, nil
}

// Search 搜索工作流
func (r *WorkflowRepository) Search(keyword string, offset, limit int) ([]*model.Workflow, int64, error) {
	var workflows []*model.Workflow
	var total int64

	query := r.db.Model(&model.Workflow{})
	if keyword != "" {
		likeKeyword := "%" + keyword + "%"
		query = query.Where("name LIKE ? OR description LIKE ?", likeKeyword, likeKeyword)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count workflows: %w", err)
	}

	// 获取分页数据
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&workflows).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to search workflows: %w", err)
	}

	return workflows, total, nil
}

// Count 统计工作流数量
func (r *WorkflowRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.Workflow{}).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count workflows: %w", err)
	}
	return count, nil
}

// UpdateStatus 更新工作流状态
func (r *WorkflowRepository) UpdateStatus(id uint, status string) error {
	result := r.db.Model(&model.Workflow{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("failed to update workflow status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("workflow with id %d not found", id)
	}
	return nil
}

// UpdateLastExecution 更新最后执行时间
func (r *WorkflowRepository) UpdateLastExecution(id uint, executionTime time.Time) error {
	result := r.db.Model(&model.Workflow{}).Where("id = ?", id).Update("last_execution", executionTime)
	if result.Error != nil {
		return fmt.Errorf("failed to update workflow last execution: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("workflow with id %d not found", id)
	}
	return nil
}
