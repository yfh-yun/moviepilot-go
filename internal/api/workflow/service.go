package workflowapi

import (
	"context"
	"fmt"
	"time"

	"moviepilot-go/internal/business/media"
	"moviepilot-go/internal/business/storage"
	"moviepilot-go/internal/business/transfer"
	wf "moviepilot-go/internal/platform/workflow"
	"moviepilot-go/pkg/logger"

	"go.uber.org/zap"
)

type workflowManager interface {
	RegisterWorkflow(workflow *wf.Workflow) error
	RunWorkflow(id string, ctx context.Context, variables map[string]interface{}) error
	WaitForWorkflow(id string) error
	GetWorkflowResult(id string) (*wf.TaskResult, error)
}

// Service 聚合工作流和业务依赖，供 Handler 调用。
type Service struct {
	manager workflowManager
	cfg     wf.LocalFileWorkflowConfig
	logger  *zap.Logger
}

// NewService 创建 Service，可注入自定义依赖；若未提供则使用占位实现。
func NewService(manager workflowManager, cfg wf.LocalFileWorkflowConfig, log *zap.Logger) *Service {
	if log == nil {
		log = logger.GetLogger()
	}
	if cfg.Logger == nil {
		cfg.Logger = log
	}
	if cfg.StorageService == nil {
		cfg.StorageService = storage.NewLocalService(cfg.Logger)
	}
	if cfg.MediaService == nil {
		cfg.MediaService = media.NewDefaultCompositeService(cfg.Logger)
	}
	if cfg.TransferService == nil {
		cfg.TransferService = transfer.NewDefaultService(cfg.StorageService, cfg.Logger)
	}

	return &Service{
		manager: manager,
		cfg:     cfg,
		logger:  log,
	}
}

// StartLocalFileWorkflow 根据请求参数构建并执行“本地文件 → 刮削 → 转移”工作流。
func (s *Service) StartLocalFileWorkflow(ctx context.Context, req StartLocalFileWorkflowRequest) (*StartLocalFileWorkflowResponse, error) {
	ctxLogger := logger.WithContext(ctx)
	ctxLogger.Info("building local file workflow",
		zap.String("root_path", req.RootPath),
		zap.String("target_root", req.TargetRoot),
		zap.Bool("include_fetch", req.IncludeFetch),
		zap.Bool("wait_for_completion", req.WaitForCompletion),
	)

	opts := wf.LocalFileWorkflowOptions{
		ID:           fmt.Sprintf("local_file_scrape_transfer_%d", time.Now().UnixNano()),
		IncludeFetch: req.IncludeFetch,
		ActionParams: map[string]any{
			"scan_file": map[string]any{
				"root_path":        req.RootPath,
				"include_patterns": req.Include,
				"exclude_patterns": req.Exclude,
				"max_depth":        req.MaxDepth,
				"follow_symlinks":  req.FollowSymlink,
			},
			"scrape_file": map[string]any{
				"force_refresh": req.ForceRefresh,
				"source":        req.Source,
			},
			"transfer_file": map[string]any{
				"target_root":  req.TargetRoot,
				"mode":         req.Mode,
				"category":     req.Category,
				"overwrite":    req.Overwrite,
				"preserve_dir": req.PreserveDir,
				"dry_run":      req.DryRun,
			},
		},
	}

	if req.IncludeFetch {
		opts.ActionParams["fetch_torrents"] = map[string]any{
			"keywords": req.FetchKeywords,
		}
	}

	workflow, err := wf.BuildLocalFileScrapeTransferWorkflow(s.cfg, opts)
	if err != nil {
		ctxLogger.Error("build workflow failed", zap.Error(err))
		return nil, fmt.Errorf("build workflow failed: %w", err)
	}

	if err := s.manager.RegisterWorkflow(workflow); err != nil {
		ctxLogger.Error("register workflow failed", zap.Error(err))
		return nil, err
	}

	if err := s.manager.RunWorkflow(workflow.ID, ctx, nil); err != nil {
		ctxLogger.Error("run workflow failed", zap.Error(err), zap.String("workflow_id", workflow.ID))
		return nil, err
	}

	resp := &StartLocalFileWorkflowResponse{
		WorkflowID: workflow.ID,
		Status:     "running",
		Message:    "workflow started",
	}
	ctxLogger.Info("workflow started",
		zap.String("workflow_id", workflow.ID),
		zap.String("status", resp.Status),
	)

	if req.WaitForCompletion {
		if err := s.manager.WaitForWorkflow(workflow.ID); err != nil {
			ctxLogger.Error("wait workflow failed", zap.Error(err), zap.String("workflow_id", workflow.ID))
			return nil, err
		}
		result, err := s.manager.GetWorkflowResult(workflow.ID)
		if err != nil {
			ctxLogger.Error("get workflow result failed", zap.Error(err), zap.String("workflow_id", workflow.ID))
			return nil, err
		}
		if result != nil {
			resp.Status = string(result.Status)
			resp.Result = result.Output
			resp.Message = "workflow completed"
		}
		ctxLogger.Info("workflow completed",
			zap.String("workflow_id", workflow.ID),
			zap.String("status", resp.Status),
		)
	}

	return resp, nil
}
