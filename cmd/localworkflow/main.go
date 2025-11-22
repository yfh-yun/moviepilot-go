package main

import (
	"context"
	"fmt"
	"time"

	"moviepilot-go/internal/business/media"
	"moviepilot-go/internal/business/storage"
	"moviepilot-go/internal/business/transfer"
	"moviepilot-go/internal/platform/workflow"
	"moviepilot-go/pkg/logger"

	"go.uber.org/zap"
)

func main() {
	if err := logger.Init(); err != nil {
		panic(fmt.Sprintf("failed to init logger: %v", err))
	}
	defer logger.Sync()

	log := logger.GetLogger()
	storageSvc := storage.NewLocalService(log)
	mediaSvc := media.NewDefaultCompositeService(log)
	transferSvc := transfer.NewDefaultService(storageSvc, log)

	cfg := workflow.LocalFileWorkflowConfig{
		Logger:          log,
		StorageService:  storageSvc,
		MediaService:    mediaSvc,
		TransferService: transferSvc,
	}

	opts := workflow.LocalFileWorkflowOptions{
		IncludeFetch: false,
		ActionParams: map[string]any{
			"scan_file": map[string]any{
				"root_path":       "./",
				"include":         []string{"*.go"},
				"exclude":         []string{},
				"max_depth":       1,
				"follow_symlinks": false,
			},
			"scrape_file": map[string]any{},
			"transfer_file": map[string]any{
				"target_root": "./output",
				"mode":        "copy",
				"category":    "test",
			},
		},
	}

	wf, err := workflow.BuildLocalFileScrapeTransferWorkflow(cfg, opts)
	if err != nil {
		log.Fatal("build workflow failed", zap.Error(err))
	}

	manager := workflow.NewWorkflowManager(log)
	if err := manager.RegisterWorkflow(wf); err != nil {
		log.Fatal("register workflow failed", zap.Error(err))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := manager.RunWorkflow(wf.ID, ctx, nil); err != nil {
		log.Fatal("run workflow failed", zap.Error(err))
	}

	if err := manager.WaitForWorkflow(wf.ID); err != nil {
		log.Fatal("wait workflow failed", zap.Error(err))
	}

	result, err := manager.GetWorkflowResult(wf.ID)
	if err != nil {
		log.Fatal("get workflow result failed", zap.Error(err))
	}

	log.Info("workflow execution finished", zap.String("status", string(result.Status)))
}
