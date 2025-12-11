package download

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"moviepilot-go/internal/business/services/base"
	"moviepilot-go/internal/infrastructure/events"
	"moviepilot-go/internal/models/database"
	"moviepilot-go/internal/models/dto"
	"moviepilot-go/internal/models/types"
	"moviepilot-go/pkg/logger"
)

// DownloadService 下载服务
// 原DownloadChain，负责下载管理、任务控制等功能
type DownloadService struct {
	*base.ServiceBase
	db               *gorm.DB
	repository       Repository
	queueService     QueueService
	monitor          *Monitor
	eventManager     *events.Manager
	torrentHandler   *TorrentHandler
	batchDownloader  *BatchDownloader
	existenceChecker *ExistenceChecker
	logger           *zap.Logger
}

// NewDownloadService 创建DownloadService实例
func NewDownloadService(
	db *gorm.DB,
	repository Repository,
	queueService QueueService,
	eventManager *events.Manager,
) *DownloadService {
	log := logger.GetLogger()

	service := &DownloadService{
		ServiceBase:  base.NewServiceBase(),
		db:           db,
		repository:   repository,
		queueService: queueService,
		eventManager: eventManager,
		logger:       log,
	}

	// 初始化其他组件
	service.torrentHandler = NewTorrentHandler(log)
	service.batchDownloader = NewBatchDownloader(service, log)
	service.existenceChecker = NewExistenceChecker(db, log)
	// 简化监控器，移除对downloaderMgr的依赖
	service.monitor = NewMonitor(MonitorConfig{
		Repository: repository,
		Logger:     log,
		Interval:   30 * time.Second,
	})

	return service
}

// Initialize 初始化服务
func (s *DownloadService) Initialize() error {
	s.logger.Info("初始化下载服务")

	// 启动监控器
	if err := s.monitor.Start(context.Background()); err != nil {
		s.logger.Error("启动下载监控器失败", zap.Error(err))
		return err
	}

	// 注册下载完成处理器
	s.monitor.RegisterCompletionHandler(s.handleDownloadCompletion)

	// 注册事件监听器
	s.registerEventListeners()

	s.logger.Info("下载服务初始化完成")
	return nil
}

// Name 获取服务名称
func (s *DownloadService) Name() string {
	return "DownloadService"
}

// Close 关闭服务
func (s *DownloadService) Close() error {
	s.logger.Info("关闭下载服务")

	if s.monitor != nil {
		if err := s.monitor.Stop(); err != nil {
			s.logger.Error("停止下载监控器失败", zap.Error(err))
		}
	}

	s.logger.Info("下载服务已关闭")
	return nil
}

// AddDownload 添加下载任务
func (s *DownloadService) AddDownload(ctx context.Context, torrentContext *dto.Context, downloader string) (*dto.DownloadTask, error) {
	// 基础实现：创建下载任务
	if torrentContext == nil {
		return nil, errors.New("torrent context cannot be nil")
	}

	// 生成下载任务ID
	taskID := fmt.Sprintf("download_%d", time.Now().Unix())

	task := &dto.DownloadTask{
		DownloadID: taskID,
		Downloader: downloader,
		Path:       "/downloads/" + taskID, // TODO: 根据配置生成路径
		Completed:  false,
	}

	return task, nil
}

// DeleteDownload 删除下载任务
func (s *DownloadService) DeleteDownload(ctx context.Context, downloadID string, deleteFiles bool) error {
	// TODO: 实现删除下载逻辑
	return nil
}

// GetDownloads 获取下载列表
func (s *DownloadService) GetDownloads(ctx context.Context, downloader string) ([]*dto.DownloadTask, error) {
	// 基础实现：返回模拟下载列表
	tasks := []*dto.DownloadTask{
		{
			DownloadID: "download_001",
			Downloader: downloader,
			Path:       "/downloads/download_001",
			Completed:  false,
		},
		{
			DownloadID: "download_002",
			Downloader: downloader,
			Path:       "/downloads/download_002",
			Completed:  true,
		},
	}

	return tasks, nil
}

// GetDownload 获取下载任务详情
func (s *DownloadService) GetDownload(ctx context.Context, downloadID string) (*dto.DownloadTask, error) {
	// TODO: 实现获取下载详情逻辑
	return nil, nil
}

// PauseDownload 暂停下载
func (s *DownloadService) PauseDownload(ctx context.Context, downloadID string) error {
	// TODO: 实现暂停下载逻辑
	return nil
}

// ResumeDownload 恢复下载
func (s *DownloadService) ResumeDownload(ctx context.Context, downloadID string) error {
	// TODO: 实现恢复下载逻辑
	return nil
}

// GetDownloadHistory 获取下载历史
func (s *DownloadService) GetDownloadHistory(ctx context.Context, page, pageSize int) ([]*dto.DownloadHistory, error) {
	// TODO: 实现获取下载历史逻辑
	return nil, nil
}

// DeleteDownloadHistory 删除下载历史
func (s *DownloadService) DeleteDownloadHistory(ctx context.Context, historyID int) error {
	s.logger.Info("删除下载历史", zap.Int("history_id", historyID))

	if err := s.db.WithContext(ctx).Delete(&database.DownloadHistory{}, historyID).Error; err != nil {
		s.logger.Error("删除下载历史失败", zap.Int("history_id", historyID), zap.Error(err))
		return err
	}

	return nil
}

// DownloadTorrent 下载种子文件
// 对应Python的download_torrent方法
// 返回：种子内容，种子目录名，种子文件清单
func (s *DownloadService) DownloadTorrent(ctx context.Context, url string, cookie string, ua string, proxy bool) ([]byte, string, []string, error) {
	s.logger.Info("下载种子文件", zap.String("url", url))
	return s.torrentHandler.DownloadTorrent(ctx, url, cookie, ua, proxy)
}

// MediaExists 检查媒体是否已存在
// 对应Python的media_exists方法
func (s *DownloadService) MediaExists(ctx context.Context, mediaInfo *dto.MediaInfo) (*dto.MediaExistsInfo, error) {
	s.logger.Info("检查媒体是否存在", zap.String("title", mediaInfo.Title))
	
	// 构建媒体存在信息
	// TODO: 实现完整的媒体存在检查逻辑
	mediaExistsInfo := &dto.MediaExistsInfo{
		Exists:   false,
		Seasons:  make(map[int][]int),
	}
	
	// 模拟实现：根据媒体类型返回不同结果
	if mediaInfo.Type == "movie" {
		// 电影：检查是否存在
		mediaExistsInfo.Exists = false
	} else {
		// 电视剧：检查季节和集数
		mediaExistsInfo.Exists = false
		mediaExistsInfo.Seasons = make(map[int][]int)
	}
	
	return mediaExistsInfo, nil
}

// DownloadSingle 添加下载任务（匹配Python的download_single方法签名）
// 对应Python的download_single方法
func (s *DownloadService) DownloadSingle(ctx context.Context, torrentContext *dto.Context, downloader, savePath, labels string) (*dto.DownloadTask, error) {
	s.logger.Info("添加下载任务", 
		zap.String("title", torrentContext.TorrentInfo.Title),
		zap.String("downloader", downloader),
		zap.String("save_path", savePath),
		zap.String("labels", labels),
	)
	
	// 生成下载任务ID
	taskID := fmt.Sprintf("download_%d", time.Now().Unix())
	
	task := &dto.DownloadTask{
		DownloadID: taskID,
		Downloader: downloader,
		Path:       savePath + "/" + taskID, // 使用指定的保存路径
		Completed:  false,
	}
	
	return task, nil
}

// DownloadSingleLegacy 下载单个资源及发送通知（旧方法签名，保持兼容）
// 对应Python的download_single方法
func (s *DownloadService) DownloadSingleLegacy(ctx context.Context, torrentContext *dto.Context, torrentPath string, torrentContent []byte, episodes []int, downloader string) error {
	s.logger.Info("下载单个资源",
		zap.String("title", torrentContext.TorrentInfo.Title),
		zap.String("downloader", downloader),
	)

	// 添加到下载器
	hash, err := s.torrentHandler.AddToDownloader(ctx, torrentContext, torrentPath, torrentContent, downloader)
	if err != nil {
		s.logger.Error("添加到下载器失败", zap.Error(err))
		return err
	}

	// 创建下载记录
	downloadRecord := &database.Download{
		Hash:       hash,
		Title:      torrentContext.TorrentInfo.Title,
		Size:       int64(torrentContext.TorrentInfo.Size),
		Status:     "downloading",
		Downloader: downloader,
		Category:   torrentContext.MediaInfo.Category,
	}

	if torrentContext.MediaInfo != nil {
		if torrentContext.MediaInfo.TmdbID != nil {
			mediaID := uint(*torrentContext.MediaInfo.TmdbID)
			downloadRecord.MediaID = &mediaID
		}
		if torrentContext.MediaInfo.Season != nil {
			downloadRecord.Season = torrentContext.MediaInfo.Season
		}
	}

	if err := s.repository.Create(ctx, downloadRecord); err != nil {
		s.logger.Error("创建下载记录失败", zap.Error(err))
		return err
	}

	// 发送下载添加事件
	err = s.eventManager.SendEvent(string(types.EventTypeDownloadAdded), map[string]any{
		"download_record": downloadRecord,
	})
	if err != nil {
		s.logger.Error("发送下载添加事件失败", zap.Error(err))
	}

	s.logger.Info("下载任务已添加", zap.String("hash", hash))
	return nil
}

// BatchDownload 根据缺失数据，从种子列表中组合择优下载
// 对应Python的batch_download方法
// 返回：已经下载的资源列表、剩余未下载到的剧集
func (s *DownloadService) BatchDownload(ctx context.Context, torrentList []*dto.Context, needEpisodes []int, downloader string) ([]*dto.Context, []int, error) {
	s.logger.Info("批量下载",
		zap.Int("torrent_count", len(torrentList)),
		zap.Ints("need_episodes", needEpisodes),
	)

	return s.batchDownloader.BatchDownload(ctx, torrentList, needEpisodes, downloader)
}

// GetNoExistsInfo 检查媒体库，查询是否存在
// 对应Python的get_no_exists_info方法
// 返回：当前媒体是否缺失，各标题总的季集和缺失的季集
func (s *DownloadService) GetNoExistsInfo(ctx context.Context, meta *dto.MetaInfo, mediaInfo *dto.MediaInfo) (bool, map[int][]int, map[int][]int, error) {
	s.logger.Info("检查媒体库存在性",
		zap.String("title", mediaInfo.Title),
		zap.String("type", mediaInfo.Type),
	)

	return s.existenceChecker.GetNoExistsInfo(ctx, meta, mediaInfo)
}

// RemoteDownloading 查询正在下载的任务，并发送消息
// 对应Python的remote_downloading方法
func (s *DownloadService) RemoteDownloading(ctx context.Context, channel string) error {
	s.logger.Info("查询正在下载的任务")

	// 获取正在下载的任务
	downloads, err := s.Downloading(ctx)
	if err != nil {
		return err
	}

	if len(downloads) == 0 {
		s.logger.Info("没有正在下载的任务")
		return nil
	}

	// 发送通知消息
	// TODO: 集成通知服务
	s.logger.Info("正在下载的任务数", zap.Int("count", len(downloads)))

	return nil
}

// Downloading 查询正在下载的任务
// 对应Python的downloading方法
// 返回：正在下载的任务列表
func (s *DownloadService) Downloading(ctx context.Context) ([]*database.Download, error) {
	s.logger.Debug("查询正在下载的任务")

	// 使用硬编码的状态值，移除对downloader包的依赖
	downloads, err := s.repository.ListByStatus(ctx, []string{
		"downloading",
		"queued",
		"checking",
	})

	if err != nil {
		s.logger.Error("查询下载任务失败", zap.Error(err))
		return nil, err
	}

	return downloads, nil
}

// SetDownloading 控制下载任务 start/stop
// 对应Python的set_downloading方法
func (s *DownloadService) SetDownloading(ctx context.Context, hash string, action string) error {
	s.logger.Info("控制下载任务",
		zap.String("hash", hash),
		zap.String("action", action),
	)

	// 获取下载记录
	download, err := s.repository.GetByHash(ctx, hash)
	if err != nil {
		s.logger.Error("获取下载记录失败", zap.Error(err))
		return err
	}

	// 执行操作 - 简化实现，仅更新状态
	var newStatus string
	switch action {
	case "start", "resume":
		newStatus = "downloading"
	case "stop", "pause":
		newStatus = "paused"
	default:
		return fmt.Errorf("不支持的操作: %s", action)
	}

	download.Status = newStatus

	// 更新记录
	if err := s.repository.Update(ctx, download); err != nil {
		s.logger.Error("更新下载记录失败", zap.Error(err))
		return err
	}

	s.logger.Info("下载任务控制成功", zap.String("hash", hash), zap.String("action", action))
	return nil
}

// RemoveDownloading 删除下载任务
// 对应Python的remove_downloading方法
func (s *DownloadService) RemoveDownloading(ctx context.Context, hash string, deleteFiles bool) error {
	s.logger.Info("删除下载任务",
		zap.String("hash", hash),
		zap.Bool("delete_files", deleteFiles),
	)

	// 获取下载记录
	download, err := s.repository.GetByHash(ctx, hash)
	if err != nil {
		s.logger.Error("获取下载记录失败", zap.Error(err))
		return err
	}

	// 删除数据库记录
	if err := s.repository.Delete(ctx, download.ID); err != nil {
		s.logger.Error("删除下载记录失败", zap.Error(err))
		return err
	}

	// 发送删除事件
	eventData := map[string]any{
		"download": download,
		"hash":     hash,
	}
	if err := s.eventManager.SendEvent(string(types.EventTypeDownloadDeleted), eventData); err != nil {
		s.logger.Warn("发送删除事件失败", zap.Error(err))
	}

	s.logger.Info("下载任务已删除", zap.String("hash", hash))
	return nil
}

// handleDownloadCompletion 处理下载完成
func (s *DownloadService) handleDownloadCompletion(ctx context.Context, download *database.Download) error {
	s.logger.Info("处理下载完成", zap.String("hash", download.Hash))

	// 发送下载完成事件
	eventData := map[string]any{
		"download": download,
		"hash":     download.Hash,
	}
	if err := s.eventManager.SendEvent(string(types.EventTypeTransferComplete), eventData); err != nil {
		s.logger.Warn("发送下载完成事件失败", zap.Error(err))
	}

	return nil
}

// registerEventListeners 注册事件监听器
func (s *DownloadService) registerEventListeners() {
	s.logger.Info("注册下载服务事件监听器")

	// 监听下载文件删除事件
	s.eventManager.Subscribe(string(types.EventTypeDownloadFileDeleted), func(evt *events.Event) error {
		s.logger.Info("收到下载文件删除事件")

		// 从事件数据中获取hash
		hash, ok := evt.Data["hash"].(string)
		if !ok || hash == "" {
			s.logger.Warn("事件数据中缺少hash")
			return nil
		}

		// 同步删除下载任务
		return s.RemoveDownloading(context.Background(), hash, false)
	})

	s.logger.Info("事件监听器注册完成")
}
