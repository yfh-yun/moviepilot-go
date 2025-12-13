package repositories

import (
	"gorm.io/gorm"

	"moviepilot-go/internal/repositories/interfaces"
	"moviepilot-go/internal/repositories/repositories"
)

// Repositories 所有仓储集合
type Repositories struct {
	UserConfig       interfaces.UserConfigRepository
	SiteIcon         interfaces.SiteIconRepository
	SiteStatistic    interfaces.SiteStatisticRepository
	SiteUserData     interfaces.SiteUserDataRepository
	SubscribeHistory interfaces.SubscribeHistoryRepository
	Workflow         interfaces.WorkflowRepository
	Subscribe        interfaces.SubscribeRepository
	Media            interfaces.MediaRepository
	User             interfaces.UserRepository
	Message          interfaces.MessageRepository
}

// NewRepositories 创建所有仓储集合
func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		UserConfig:       repositories.NewUserConfigRepository(db),
		SiteIcon:         repositories.NewSiteIconRepository(db),
		SiteStatistic:    repositories.NewSiteStatisticRepository(db),
		SiteUserData:     repositories.NewSiteUserDataRepository(db),
		SubscribeHistory: repositories.NewSubscribeHistoryRepository(db),
		Workflow:         repositories.NewWorkflowRepository(db),
		Subscribe:        repositories.NewSubscribeRepository(db),
		Media:            repositories.NewMediaRepository(db),
		User:             repositories.NewUserRepository(db),
		Message:          repositories.NewMessageRepository(db),
	}
}
