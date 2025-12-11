package transfer

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"moviepilot-go/internal/business/services/storage"
	"moviepilot-go/internal/models/database"
)

// Task 描述一次转移任务所需的核心字段。
type Task struct {
	Media      database.Media
	SourcePath string
	TargetPath string
	Mode       storage.TransferMode
	Overwrite  bool
	Category   string
}

// Service 定义转移相关的业务操作。
type Service interface {
	Execute(tasks []Task) ([]database.TransferHistory, error)
	// QueryName 查询整理后的名称
	QueryName(path, filetype string) (string, error)
	// GetQueue 查询整理队列
	GetQueue() ([]any, error)
	// RemoveFromQueue 从整理队列中删除任务
	RemoveFromQueue(path string) error
	// Process 立即执行下载器文件整理
	Process() error
}

// DefaultService 通过 storage.Service 完成底层转移。
type DefaultService struct {
	storageSvc storage.Service
	logger     *zap.Logger
}

// NewDefaultService 创建 DefaultService。
func NewDefaultService(storageSvc storage.Service, logger *zap.Logger) *DefaultService {
	return &DefaultService{storageSvc: storageSvc, logger: logger}
}

// NewTransferService 创建转移服务实例（路由使用）
func NewTransferService(db *gorm.DB, logger *zap.Logger) *DefaultService {
	storageSvc := storage.NewStorageService(db, logger)
	return &DefaultService{storageSvc: storageSvc, logger: logger}
}

// Execute 将业务任务转换为 storage.TransferTask，并返回转移历史记录（当前为占位实现）。
func (s *DefaultService) Execute(tasks []Task) ([]database.TransferHistory, error) {
	transferTasks := make([]storage.TransferTask, 0, len(tasks))
	for _, task := range tasks {
		transferTasks = append(transferTasks, storage.TransferTask{
			SourcePath: task.SourcePath,
			TargetPath: task.TargetPath,
			Mode:       task.Mode,
			Overwrite:  task.Overwrite,
		})
	}

	if _, err := s.storageSvc.Transfer(transferTasks); err != nil {
		return nil, err
	}

	histories := make([]database.TransferHistory, 0, len(tasks))
	now := time.Now()
	for _, task := range tasks {
		// 构建 Seasons 和 Episodes 字符串
		var seasons, episodes string
		if task.Media.Season != nil {
			seasons = fmt.Sprintf("S%02d", *task.Media.Season)
		}
		if task.Media.Episode != nil {
			episodes = fmt.Sprintf("E%02d", *task.Media.Episode)
		}

		histories = append(histories, database.TransferHistory{
			BaseModel: database.BaseModel{
				CreatedAt: now,
				UpdatedAt: now,
			},
			Type:          task.Media.Type,
			Title:         task.Media.Title,
			Year:          task.Media.Year,
			Seasons:       seasons,
			Episodes:      episodes,
			Src:           task.SourcePath,
			Source:        task.SourcePath,
			SourcePath:    task.SourcePath,
			Dest:          task.TargetPath,
			Target:        task.TargetPath,
			TargetPath:    task.TargetPath,
			Status:        true, // true 表示计划中/成功
			Mode:          string(task.Mode),
			Note:          `{"mode":"` + string(task.Mode) + `"}`,
			MediaCategory: task.Category,
			Date:          now.Format("2006-01-02 15:04:05"),
		})
	}

	if s.logger != nil {
		s.logger.Info("transfer tasks planned", zap.Int("count", len(histories)))
	}

	return histories, nil
}

// QueryName 查询整理后的名称
func (s *DefaultService) QueryName(path, filetype string) (string, error) {
	s.logger.Info("Querying renamed media name", zap.String("path", path), zap.String("filetype", filetype))

	// TODO: 实现查询整理后的名称逻辑
	// 1. 解析文件路径，获取元信息
	// 2. 识别媒体信息
	// 3. 生成推荐的新名称
	// 4. 根据filetype返回不同格式的名称

	// 目前返回文件名作为占位实现
	return path, nil
}

// GetQueue 查询整理队列
func (s *DefaultService) GetQueue() ([]any, error) {
	s.logger.Info("Getting transfer queue")

	// TODO: 实现查询整理队列逻辑
	// 1. 从全局变量或缓存中获取当前整理队列
	// 2. 转换为前端需要的格式

	// 目前返回空列表作为占位实现
	return []any{}, nil
}

// RemoveFromQueue 从整理队列中删除任务
func (s *DefaultService) RemoveFromQueue(path string) error {
	s.logger.Info("Removing task from transfer queue", zap.String("path", path))

	// TODO: 实现从整理队列中删除任务逻辑
	// 1. 从全局变量或缓存中找到对应的任务
	// 2. 删除任务
	// 3. 取消正在进行的转移

	// 目前返回nil作为占位实现
	return nil
}

// Process 立即执行下载器文件整理
func (s *DefaultService) Process() error {
	s.logger.Info("Processing downloader files")

	// TODO: 实现立即执行下载器文件整理逻辑
	// 1. 调用下载器文件整理逻辑
	// 2. 处理下载器中的文件

	// 目前返回nil作为占位实现
	return nil
}
