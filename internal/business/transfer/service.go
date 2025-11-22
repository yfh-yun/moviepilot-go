package transfer

import (
	"time"

	"go.uber.org/zap"

	"moviepilot-go/internal/business/storage"
	"moviepilot-go/internal/models"
)

// Task 描述一次转移任务所需的核心字段。
type Task struct {
	Media      models.Media
	SourcePath string
	TargetPath string
	Mode       storage.TransferMode
	Overwrite  bool
	Category   string
}

// Service 定义转移相关的业务操作。
type Service interface {
	Execute(tasks []Task) ([]models.TransferHistory, error)
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

// Execute 将业务任务转换为 storage.TransferTask，并返回转移历史记录（当前为占位实现）。
func (s *DefaultService) Execute(tasks []Task) ([]models.TransferHistory, error) {
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

	histories := make([]models.TransferHistory, 0, len(tasks))
	now := time.Now()
	for _, task := range tasks {
		histories = append(histories, models.TransferHistory{
			BaseModel: models.BaseModel{
				CreatedAt: now,
				UpdatedAt: now,
			},
			Type:          task.Media.Type,
			Title:         task.Media.Title,
			Year:          task.Media.Year,
			Season:        task.Media.Season,
			Episode:       task.Media.Episode,
			Source:        task.SourcePath,
			SourcePath:    task.SourcePath,
			Target:        task.TargetPath,
			TargetPath:    task.TargetPath,
			Status:        "planned",
			Note:          `{"mode":"` + string(task.Mode) + `"}`,
			MediaCategory: task.Category,
		})
	}

	if s.logger != nil {
		s.logger.Info("transfer tasks planned", zap.Int("count", len(histories)))
	}

	return histories, nil
}
