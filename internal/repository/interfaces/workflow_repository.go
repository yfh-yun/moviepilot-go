package interfaces

import (
	"github.com/yfh-yun/moviepilot-go/internal/model"
	"time"
)

// WorkflowRepository 工作流仓储接口
type WorkflowRepository interface {
	// Create 创建工作流
	Create(workflow *model.Workflow) error

	// GetByID 根据ID获取工作流
	GetByID(id uint) (*model.Workflow, error)

	// GetByName 根据名称获取工作流
	GetByName(name string) (*model.Workflow, error)

	// GetByType 根据类型获取工作流列表
	GetByType(workflowType string) ([]*model.Workflow, error)

	// GetActive 获取活跃的工作流列表
	GetActive() ([]*model.Workflow, error)

	// GetByTrigger 根据触发器获取工作流列表
	GetByTrigger(trigger string) ([]*model.Workflow, error)

	// Update 更新工作流
	Update(workflow *model.Workflow) error

	// Delete 删除工作流
	Delete(id uint) error

	// List 分页获取工作流列表
	List(offset, limit int) ([]*model.Workflow, int64, error)

	// Search 搜索工作流
	Search(keyword string, offset, limit int) ([]*model.Workflow, int64, error)

	// Count 统计工作流数量
	Count() (int64, error)

	// UpdateStatus 更新工作流状态
	UpdateStatus(id uint, status string) error

	// UpdateLastExecution 更新最后执行时间
	UpdateLastExecution(id uint, executionTime time.Time) error
}
