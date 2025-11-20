package migrations
import (
	"github.com/yfh-yun/moviepilot-go/internal/models"
	"github.com/yfh-yun/moviepilot-go/pkg/database"
)

// MigrateDBModels 迁移数据库模型
func MigrateDBModels() error {
	db := database.GetDB()
	
	// 自动迁移新增的模型表
	err := db.AutoMigrate(
		&model.UserConfig{},
		&model.SiteIcon{},
		&model.SiteUserData{},
		&model.SiteStatistic{},
		&model.SubscribeHistory{},
		&model.Workflow{},
		&model.TransferHistory{},
	)
	
	if err != nil {
		return err
	}
	
	return nil
}
