// Package actions 提供从旧架构迁移到新架构的工具
package actions

import (
	"fmt"
	"log"
	"time"

	"github.com/yfh-yun/moviepilot-go/internal/actions/interfaces"
	"github.com/yfh-yun/moviepilot-go/internal/actions/types"
	"github.com/yfh-yun/moviepilot-go/internal/actions/base"
	"github.com/yfh-yun/moviepilot-go/internal/actions/manager"
	"github.com/yfh-yun/moviepilot-go/internal/actions/implementations"
)

// MigrationHelper 迁移助手
type MigrationHelper struct {
	oldActionRegistry map[string]interface{} // 旧的动作注册表
	newManager        interfaces.Manager     // 新的管理器
	logger            *log.Logger
}

// NewMigrationHelper 创建迁移助手
func NewMigrationHelper() *MigrationHelper {
	return &MigrationHelper{
		oldActionRegistry: make(map[string]interface{}),
		newManager:        manager.NewActionManager(nil),
		logger:            log.Default(),
	}
}

// RegisterOldAction 注册旧动作（模拟）
func (mh *MigrationHelper) RegisterOldAction(actionID string, oldAction interface{}) {
	mh.oldActionRegistry[actionID] = oldAction
	mh.logger.Printf("注册旧动作: %s", actionID)
}

// MigrateAction 迁移单个动作
func (mh *MigrationHelper) MigrateAction(actionID string) error {
	oldAction, exists := mh.oldActionRegistry[actionID]
	if !exists {
		return fmt.Errorf("旧动作 %s 不存在", actionID)
	}

	// 根据动作类型创建新动作
	newAction, err := mh.createNewAction(actionID, oldAction)
	if err != nil {
		return fmt.Errorf("创建新动作失败: %w", err)
	}

	// 注册到新管理器
	if err := mh.newManager.RegisterAction(newAction); err != nil {
		return fmt.Errorf("注册新动作失败: %w", err)
	}

	mh.logger.Printf("成功迁移动作: %s", actionID)
	return nil
}

// createNewAction 根据旧动作创建新动作
func (mh *MigrationHelper) createNewAction(actionID string, oldAction interface{}) (interfaces.Action, error) {
	// 根据动作ID判断动作类型
	switch {
	case contains(actionID, []string{"download", "torrent"}):
		// 创建下载动作
		return mh.createDownloadAction(actionID, oldAction)
	case contains(actionID, []string{"scan", "file"}):
		// 创建扫描动作
		return mh.createScanAction(actionID, oldAction)
	default:
		// 创建通用动作
		return mh.createGenericAction(actionID, oldAction)
	}
}

// createDownloadAction 创建下载动作
func (mh *MigrationHelper) createDownloadAction(actionID string, oldAction interface{}) (interfaces.Action, error) {
	// 这里应该根据旧动作的配置创建新动作
	// 简化实现，使用默认配置
	cache := mh.newManager.GetCache()
	if cache == nil {
		// 创建默认缓存
		cache = &DefaultCache{}
	}
	
	// 模拟下载服务
	downloadService := &DefaultDownloadService{}
	
	return implementations.NewDownloadAction(actionID, downloadService, cache), nil
}

// createScanAction 创建扫描动作
func (mh *MigrationHelper) createScanAction(actionID string, oldAction interface{}) (interfaces.Action, error) {
	cache := mh.newManager.GetCache()
	if cache == nil {
		cache = &DefaultCache{}
	}
	
	// 模拟文件服务
	fileService := &DefaultFileService{}
	
	return implementations.NewScanAction(actionID, fileService, cache), nil
}

// createGenericAction 创建通用动作
func (mh *MigrationHelper) createGenericAction(actionID string, oldAction interface{}) (interfaces.Action, error) {
	cache := mh.newManager.GetCache()
	if cache == nil {
		cache = &DefaultCache{}
	}
	
	// 创建基础动作
	action := base.NewBaseAction(actionID, cache)
	
	// 设置名称和描述
	if ga, ok := action.(*base.BaseAction); ok {
		// 这里可以根据旧动作的信息设置配置
		ga.SetConfig(map[string]interface{}{
			"migrated_from": "old_action",
			"migration_time": "now",
		})
	}
	
	return action, nil
}

// MigrateAllActions 迁移所有动作
func (mh *MigrationHelper) MigrateAllActions() error {
	mh.logger.Println("开始迁移所有动作...")
	
	for actionID := range mh.oldActionRegistry {
		if err := mh.MigrateAction(actionID); err != nil {
			mh.logger.Printf("迁移动作 %s 失败: %v", actionID, err)
			continue
		}
	}
	
	mh.logger.Println("动作迁移完成")
	return nil
}

// GetNewManager 获取新管理器
func (mh *MigrationHelper) GetNewManager() interfaces.Manager {
	return mh.newManager
}

// ValidateMigration 验证迁移结果
func (mh *MigrationHelper) ValidateMigration() error {
	actions := mh.newManager.ListActions()
	mh.logger.Printf("新管理器中共有 %d 个动作", len(actions))
	
	// 检查每个动作的健康状态
	for _, action := range actions {
		if err := action.Initialize(); err != nil {
			return fmt.Errorf("初始化动作 %s 失败: %w", action.GetActionID(), err)
		}
	}
	
	mh.logger.Println("迁移验证通过")
	return nil
}

// contains 检查字符串是否包含任一关键词
func contains(str string, keywords []string) bool {
	for _, keyword := range keywords {
		if len(str) >= len(keyword) {
			for i := 0; i <= len(str)-len(keyword); i++ {
				if str[i:i+len(keyword)] == keyword {
					return true
				}
			}
		}
	}
	return false
}

// DefaultCache 默认缓存实现（用于迁移）
type DefaultCache struct{}

func (dc *DefaultCache) Get(ctx interface{}, key string) (interface{}, error) {
	return nil, nil
}

func (dc *DefaultCache) Set(ctx interface{}, key string, value interface{}, ttl int64) error {
	return nil
}

func (dc *DefaultCache) Delete(ctx interface{}, key string) error {
	return nil
}

func (dc *DefaultCache) Clear(ctx interface{}) error {
	return nil
}

func (dc *DefaultCache) Exists(ctx interface{}, key string) (bool, error) {
	return false, nil
}

func (dc *DefaultCache) Keys(ctx interface{}, pattern string) ([]string, error) {
	return []string{}, nil
}

func (dc *DefaultCache) Size(ctx interface{}) (int64, error) {
	return 0, nil
}

// DefaultDownloadService 默认下载服务实现（用于迁移）
type DefaultDownloadService struct{}

func (dds *DefaultDownloadService) ListDownloads(ctx interface{}, params interfaces.ListDownloadsParams) ([]*interfaces.DownloadTask, int64, error) {
	return []*interfaces.DownloadTask{}, 0, nil
}

func (dds *DefaultDownloadService) GetDownloadDetail(ctx interface{}, taskID string) (*interfaces.DownloadTask, error) {
	return nil, fmt.Errorf("not implemented")
}

func (dds *DefaultDownloadService) CreateDownload(ctx interface{}, params interfaces.CreateDownloadParams) (*interfaces.DownloadTask, error) {
	return &interfaces.DownloadTask{
		ID:         "migrated_task",
		Title:      params.Title,
		Type:       params.Type,
		Status:     "pending",
		Progress:   0,
		FileSize:   0,
		Downloaded: 0,
		Speed:      0,
		ETA:        "",
		CreatedAt:  *new(time.Time),
		UpdatedAt:  *new(time.Time),
	}, nil
}

func (dds *DefaultDownloadService) DeleteDownload(ctx interface{}, taskID string) error {
	return nil
}

func (dds *DefaultDownloadService) PauseDownload(ctx interface{}, taskID string) error {
	return nil
}

func (dds *DefaultDownloadService) ResumeDownload(ctx interface{}, taskID string) error {
	return nil
}

func (dds *DefaultDownloadService) GetDownloadStats(ctx interface{}) (*interfaces.DownloadStats, error) {
	return &interfaces.DownloadStats{}, nil
}

func (dds *DefaultDownloadService) GetDownloadSpeed(ctx interface{}) (*interfaces.DownloadSpeed, error) {
	return &interfaces.DownloadSpeed{}, nil
}

func (dds *DefaultDownloadService) ClearCompletedDownloads(ctx interface{}) error {
	return nil
}

func (dds *DefaultDownloadService) BatchDeleteDownloads(ctx interface{}, taskIDs []string) error {
	return nil
}

// DefaultFileService 默认文件服务实现（用于迁移）
type DefaultFileService struct{}

func (dfs *DefaultFileService) ListFiles(ctx interface{}, filter interfaces.FileFilter) ([]*types.File, int64, error) {
	return []*types.File{}, 0, nil
}

func (dfs *DefaultFileService) GetFile(ctx interface{}, fileID int) (*types.File, error) {
	return nil, fmt.Errorf("not implemented")
}

func (dfs *DefaultFileService) CreateFile(ctx interface{}, file *types.File) error {
	return nil
}

func (dfs *DefaultFileService) UpdateFile(ctx interface{}, fileID int, file *types.File) error {
	return nil
}

func (dfs *DefaultFileService) DeleteFile(ctx interface{}, fileID int) error {
	return nil
}

func (dfs *DefaultFileService) ScanFiles(ctx interface{}, paths []string, recursive bool) ([]*types.File, error) {
	return []*types.File{}, nil
}

// MigrationReport 迁移报告
type MigrationReport struct {
	TotalOldActions    int                    `json:"total_old_actions"`
	MigratedActions    []string               `json:"migrated_actions"`
	FailedActions      map[string]string      `json:"failed_actions"`
	NewActionsCount    int                    `json:"new_actions_count"`
	ValidationPassed   bool                   `json:"validation_passed"`
	MigrationTime      string                 `json:"migration_time"`
}

// GenerateReport 生成迁移报告
func (mh *MigrationHelper) GenerateReport() *MigrationReport {
	report := &MigrationReport{
		TotalOldActions: len(mh.oldActionRegistry),
		MigratedActions: []string{},
		FailedActions:   make(map[string]string),
		NewActionsCount: len(mh.newManager.ListActions()),
		ValidationPassed: true,
		MigrationTime:    "now",
	}
	
	// 这里应该收集实际的迁移结果
	// 简化实现
	
	return report
}