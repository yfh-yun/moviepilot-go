package repository

import (
	"github.com/yfh-yun/moviepilot-go/internal/repository/repositories"
	"github.com/yfh-yun/moviepilot-go/internal/repository/interfaces"
)

// Repositories 所有仓储集合
type Repositories struct {
	UserConfig     interfaces.UserConfigRepository
	SiteIcon       interfaces.SiteIconRepository
	SiteStatistic  interfaces.SiteStatisticRepository
	SiteUserData    interfaces.SiteUserDataRepository
	SubscribeHistory interfaces.SubscribeHistoryRepository
	Workflow        interfaces.WorkflowRepository
	
	// 其他已有的Repository
	// 这些在factory.go中实现
}

// NewRepositories 创建所有仓储集合
func NewRepositories() *Repositories {
	return &Repositories{
		UserConfig:     repositories.NewUserConfigRepository(),
		SiteIcon:       repositories.NewSiteIconRepository(),
		SiteStatistic:  repositories.NewSiteStatisticRepository(),
		SiteUserData:    repositories.NewSiteUserDataRepository(),
		SubscribeHistory: repositories.NewSubscribeHistoryRepository(),
		Workflow:        repositories.NewWorkflowRepository(),
	}
}
