package models

// WorkflowConfig 工作流配置
type WorkflowConfig struct {
	// 工作流数据共享
	StatisticShare bool `mapstructure:"WORKFLOW_STATISTIC_SHARE" default:"true"`
}