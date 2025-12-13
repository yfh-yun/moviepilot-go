package command

// InitializeCommands 初始化所有命令
func InitializeCommands(service Service) error {
	// 注册处理器命令
	handlers := []Handler{
		NewHelpHandler(service),
		NewStatusHandler(),
		NewSubscribeHandler(),
		NewUnsubscribeHandler(),
		NewClearCacheHandler(),
		NewVersionHandler(),
		NewRestartHandler(),
		NewDownloadingHandler(),
		NewRedoHandler(),
	}

	for _, handler := range handlers {
		if err := service.RegisterHandler(handler); err != nil {
			return err
		}
	}

	// 注册调度器命令
	schedulerHandlers := []SchedulerHandler{
		NewCookieCloudSchedulerHandler(),
		NewMediaServerSyncSchedulerHandler(),
		NewSubscribeRefreshSchedulerHandler(),
		NewSubscribeSearchSchedulerHandler(),
		NewSubscribeTmdbSchedulerHandler(),
		NewTransferSchedulerHandler(),
	}

	for _, handler := range schedulerHandlers {
		if err := service.RegisterScheduler(handler); err != nil {
			return err
		}
	}

	// 插件命令将在插件加载时动态注册

	return nil
}
