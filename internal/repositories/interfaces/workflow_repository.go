package interfaces

import (
	"context"
	"moviepilot-go/internal/models/database"
)

// WorkflowRepository 工作流仓储接口
type WorkflowRepository interface {
	// Create 创建工作流
	Create(ctx context.Context, workflow *database.Workflow) error

	// GetByID 根据ID获取工作流
	GetByID(ctx context.Context, id string) (*database.Workflow, error)

	// Update 更新工作流
	Update(ctx context.Context, workflow *database.Workflow) error

	// Delete 删除工作流
	Delete(ctx context.Context, id string) error

	// List 获取工作流列表
	List(ctx context.Context, params ListWorkflowParams) ([]*database.Workflow, int64, error)

	// GetActiveWorkflows 获取活跃工作流
	GetActiveWorkflows(ctx context.Context) ([]*database.Workflow, error)

	// GetByType 根据类型获取工作流
	GetByType(ctx context.Context, workflowType string) ([]*database.Workflow, error)

	// UpdateStatus 更新工作流状态
	UpdateStatus(ctx context.Context, id, status string) error
}

// ListWorkflowParams 工作流列表查询参数
type ListWorkflowParams struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Type     string `json:"type"`
	Status   string `json:"status"`
	UserID   string `json:"user_id"`
}
