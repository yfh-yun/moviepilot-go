package transfer

import (
	"context"
	"strconv"

	"go.uber.org/zap"

	"moviepilot-go/internal/models/database"
	"moviepilot-go/internal/models/dto"
	"moviepilot-go/internal/repositories/interfaces"
)

// HistoryService 定义转移历史相关的业务操作
// 负责从仓储中查询/删除历史记录并转换为 DTO。
type HistoryService interface {
	ListHistory(ctx context.Context, page, size int, status string) (*dto.TransferHistoryResponse, error)
	DeleteHistory(ctx context.Context, id uint) error
	ExecuteAndSave(ctx context.Context, tasks []Task) ([]database.TransferHistory, error)
}

// historyService 默认实现
type historyService struct {
	repo     interfaces.TransferHistoryRepository
	executor Service
	logger   *zap.Logger
}

// NewHistoryService 创建 HistoryService 实例
func NewHistoryService(repo interfaces.TransferHistoryRepository, executor Service, logger *zap.Logger) HistoryService {
	return &historyService{repo: repo, executor: executor, logger: logger}
}

// ListHistory 分页查询转移历史
func (s *historyService) ListHistory(ctx context.Context, page, size int, status string) (*dto.TransferHistoryResponse, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}

	var statusPtr *bool
	if status != "" {
		switch status {
		case "completed", "success":
			v := true
			statusPtr = &v
		case "failed", "error", "cancelled":
			v := false
			statusPtr = &v
		}
	}

	params := interfaces.ListTransferHistoryParams{
		Page:     page,
		PageSize: size,
		Status:   statusPtr,
	}

	histories, total, err := s.repo.ListByPage(ctx, params)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("List transfer history failed", zap.Error(err))
		}
		return nil, err
	}

	records := make([]dto.TransferRecord, 0, len(histories))
	for _, h := range histories {
		if h == nil {
			continue
		}

		statusStr := "failed"
		if h.Status {
			statusStr = "completed"
		}

		record := dto.TransferRecord{
			ID:          strconv.FormatUint(uint64(h.ID), 10),
			TaskID:      "history_" + strconv.FormatUint(uint64(h.ID), 10),
			SourcePath:  h.SourcePath,
			TargetPath:  h.TargetPath,
			MediaID:     "", // 当前模型中没有严格的一一对应，先留空
			Status:      statusStr,
			FileSize:    0,
			Speed:       0,
			Duration:    0,
			ErrorMsg:    h.ErrMsg,
			CreatedAt:   h.CreatedAt.Format("2006-01-02 15:04:05"),
			CompletedAt: "", // TransferHistory 目前没有单独的完成时间字段
		}
		records = append(records, record)
	}

	resp := &dto.TransferHistoryResponse{
		Page:    page,
		Size:    size,
		Total:   int(total),
		Status:  status,
		Records: records,
		Message: "History retrieved successfully",
	}

	return resp, nil
}

// DeleteHistory 删除指定 ID 的转移历史
func (s *historyService) DeleteHistory(ctx context.Context, id uint) error {
	// 先检查是否存在
	h, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("Get transfer history failed before delete", zap.Error(err), zap.Uint("id", id))
		}
		return err
	}
	if h == nil {
		// 直接返回 nil，让上层决定返回 404 还是 200
		return nil
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		if s.logger != nil {
			s.logger.Error("Delete transfer history failed", zap.Error(err), zap.Uint("id", id))
		}
		return err
	}
	return nil
}

// ExecuteAndSave 调用底层转移服务执行任务，并将返回的历史记录写入仓储
func (s *historyService) ExecuteAndSave(ctx context.Context, tasks []Task) ([]database.TransferHistory, error) {
	if s.executor == nil {
		if s.logger != nil {
			s.logger.Error("transfer executor is not configured")
		}
		return nil, nil
	}

	histories, err := s.executor.Execute(tasks)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("execute transfer tasks failed", zap.Error(err))
		}
		return nil, err
	}

	for i := range histories {
		if err := s.repo.Create(ctx, &histories[i]); err != nil {
			if s.logger != nil {
				s.logger.Error("persist transfer history failed", zap.Error(err))
			}
			return nil, err
		}
	}

	return histories, nil
}
