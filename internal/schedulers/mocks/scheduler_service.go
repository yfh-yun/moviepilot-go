package mocks

import (
	"context"
	"moviepilot-go/internal/schedulers"

	"github.com/stretchr/testify/mock"
)

// SchedulerService 是 scheduler.SchedulerService 的模拟类型
type SchedulerService struct {
	mock.Mock
}

// Start 模拟 Start 方法
func (m *SchedulerService) Start() error {
	args := m.Called()
	return args.Error(0)
}

// Stop 模拟 Stop 方法
func (m *SchedulerService) Stop() {
	m.Called()
}

// ExecuteWorkflow 模拟 ExecuteWorkflow 方法
func (m *SchedulerService) ExecuteWorkflow(ctx context.Context, workflowID string, triggerData map[string]interface{}) (*scheduler.WorkflowInstance, error) {
	args := m.Called(ctx, workflowID, triggerData)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*scheduler.WorkflowInstance), args.Error(1)
}

// GetWorkflowInstance 模拟 GetWorkflowInstance 方法
func (m *SchedulerService) GetWorkflowInstance(instanceID string) (*scheduler.WorkflowInstance, error) {
	args := m.Called(instanceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*scheduler.WorkflowInstance), args.Error(1)
}

// ListJobStatus 模拟 ListJobStatus 方法
func (m *SchedulerService) ListJobStatus() []scheduler.JobConfig {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]scheduler.JobConfig)
}

// ListWorkflowInstances 模拟 ListWorkflowInstances 方法
func (m *SchedulerService) ListWorkflowInstances() []scheduler.WorkflowInstance {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]scheduler.WorkflowInstance)
}

// AddJob 模拟 AddJob 方法
func (m *SchedulerService) AddJob(config *scheduler.JobConfig) error {
	args := m.Called(config)
	return args.Error(0)
}

// EnableJob 模拟 EnableJob 方法
func (m *SchedulerService) EnableJob(jobID string) error {
	args := m.Called(jobID)
	return args.Error(0)
}

// DisableJob 模拟 DisableJob 方法
func (m *SchedulerService) DisableJob(jobID string) error {
	args := m.Called(jobID)
	return args.Error(0)
}
