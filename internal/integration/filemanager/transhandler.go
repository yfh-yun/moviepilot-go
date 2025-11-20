// Package filemanager 文件传输处理器
package filemanager

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/models"
	"github.com/yfh-yun/moviepilot-go/internal/models"
	"github.com/yfh-yun/moviepilot-go/pkg/utils"

	"go.uber.org/zap"
)

// TransHandler 文件转移整理类
type TransHandler struct {
	logger      *zap.Logger
	mutex       sync.RWMutex
	result      *TransferInfo
	eventBus    EventBus
	metaHelper  *MetaHelper
	tmdbHelper  *TmdbHelper
}

// EventBus 事件总线接口
type EventBus interface {
	EmitEvent(ctx context.Context, eventType string, data interface{}) error
}

// MetaHelper 元数据助手接口
type MetaHelper interface {
	RecognizeFile(ctx context.Context, filePath string) (*models.MetaBase, error)
	GenerateFileName(ctx context.Context, mediaInfo *models.MediaInfo, metaBase *models.MetaBase) (string, error)
}

// TmdbHelper TMDB助手接口
type TmdbHelper interface {
	GetEpisodeInfo(ctx context.Context, tmdbID int, season int) ([]*models.TmdbEpisode, error)
	ScrapeMetadata(ctx context.Context, filePath string, mediaInfo *models.MediaInfo) error
}

// NewTransHandler 创建传输处理器
func NewTransHandler() *TransHandler {
	return &TransHandler{
		logger:   logger.Logger,
		eventBus: nil, // 需要注入
		result:   &TransferInfo{},
	}
}

// SetEventBus 设置事件总线
func (th *TransHandler) SetEventBus(eventBus EventBus) {
	th.eventBus = eventBus
}

// SetMetaHelper 设置元数据助手
func (th *TransHandler) SetMetaHelper(metaHelper MetaHelper) {
	th.metaHelper = metaHelper
}

// SetTmdbHelper 设置TMDB助手
func (th *TransHandler) SetTmdbHelper(tmdbHelper TmdbHelper) {
	th.tmdbHelper = tmdbHelper
}

// TransHandlerRequest 传输处理器请求
type TransHandlerRequest struct {
	FileItem        *FileItem                `json:"file_item"`
	InMeta          *models.MetaBase          `json:"in_meta"`
	MediaInfo       *models.MediaInfo         `json:"media_info"`
	TargetStorage   string                   `json:"target_storage"`
	TargetPath      string                   `json:"target_path"`
	TransferType    string                   `json:"transfer_type"`
	SourceOper      storage.Storage           `json:"source_oper"`
	TargetOper      storage.Storage           `json:"target_oper"`
	NeedScrape      bool                     `json:"need_scrape"`
	NeedRename      bool                     `json:"need_rename"`
	NeedNotify      bool                     `json:"need_notify"`
	OverwriteMode   string                   `json:"overwrite_mode"`
	EpisodesInfo    []*models.TmdbEpisode     `json:"episodes_info"`
}

// TransferMedia 转移媒体文件
func (th *TransHandler) TransferMedia(ctx context.Context, req *TransHandlerRequest) (*TransferInfo, error) {
	th.mutex.Lock()
	defer th.mutex.Unlock()

	// 重置结果
	th.resetResult()

	// 验证请求
	if err := th.validateRequest(req); err != nil {
		th.setResult(Success, false)
		th.setResult(Message, err.Error())
		return th.result, err
	}

	// 发送转移拦截事件
	if th.eventBus != nil {
		interceptData := &TransferInterceptEventData{
			FileItem:  req.FileItem,
			MediaInfo: req.MediaInfo,
			MetaBase:  req.InMeta,
		}
		if err := th.eventBus.EmitEvent(ctx, "transfer.intercept", interceptData); err != nil {
			th.logger.Warn("Failed to emit transfer intercept event", zap.Error(err))
		}
	}

	// 处理文件转移
	if req.FileItem.IsDir {
		// 处理目录
		if err := th.transferDirectory(ctx, req); err != nil {
			th.setResult(Success, false)
			th.setResult(Message, err.Error())
			return th.result, err
		}
	} else {
		// 处理单个文件
		if err := th.transferFile(ctx, req, req.FileItem); err != nil {
			th.setResult(Success, false)
			th.setResult(Message, err.Error())
			return th.result, err
		}
	}

	// 刮削元数据
	if req.NeedScrape && th.result.Success {
		if err := th.scrapeMetadata(ctx, req); err != nil {
			th.logger.Warn("Failed to scrape metadata", zap.Error(err))
			th.setResult(ScrapeStatus, false)
		} else {
			th.setResult(ScrapeStatus, true)
		}
	}

	// 发送通知
	if req.NeedNotify && th.result.Success {
		if err := th.sendNotification(ctx, req); err != nil {
			th.logger.Warn("Failed to send notification", zap.Error(err))
			th.setResult(NotifyStatus, false)
		} else {
			th.setResult(NotifyStatus, true)
		}
	}

	return th.result, nil
}

// MediaExists 检查媒体是否存在
func (th *TransHandler) MediaExists(ctx context.Context, req *MediaExistsHandlerRequest) (*ExistMediaInfo, error) {
	// 构建搜索路径
	searchPath := th.buildSearchPath(req.MediaInfo, req.TargetPath, req.Season)

	// 检查文件是否存在
	files, err := req.TargetOper.ListFiles(ctx, searchPath, true)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	if len(files) == 0 {
		return &ExistMediaInfo{
			Exists: false,
		}, nil
	}

	// 分析存在的文件
	existInfo := &ExistMediaInfo{
		Exists: true,
		Path:   searchPath,
		Files:  make([]string, 0),
	}

	var episodes []int
	for _, file := range files {
		if !file.IsDir {
			existInfo.Files = append(existInfo.Files, file.Path)
			
			// 尝试解析集数
			if episode := th.extractEpisodeNumber(file.Name); episode > 0 {
				episodes = append(episodes, episode)
			}
		}
	}

	existInfo.Episodes = episodes
	return existInfo, nil
}

// MoveFile 移动文件
func (th *TransHandler) MoveFile(ctx context.Context, req *MoveFileHandlerRequest) error {
	// 检查源文件是否存在
	sourceInfo, err := req.SourceOper.GetFileInfo(ctx, req.SourcePath)
	if err != nil {
		return fmt.Errorf("failed to get source file info: %w", err)
	}

	// 检查目标目录是否存在，不存在则创建
	targetDir := filepath.Dir(req.TargetPath)
	if err := req.TargetOper.CreateDirectory(ctx, targetDir); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// 执行移动操作
	if req.SourceOper == req.TargetOper {
		// 同一存储，使用重命名
		return req.SourceOper.RenameFile(ctx, req.SourcePath, req.TargetPath)
	} else {
		// 不同存储，使用复制+删除
		if err := th.copyFile(ctx, req.SourceOper, req.TargetOper, req.SourcePath, req.TargetPath, sourceInfo.Size); err != nil {
			return err
		}
		return req.SourceOper.DeleteFile(ctx, req.SourcePath)
	}
}

// validateRequest 验证请求
func (th *TransHandler) validateRequest(req *TransHandlerRequest) error {
	if req.FileItem == nil {
		return fmt.Errorf("file item is required")
	}
	if req.MediaInfo == nil {
		return fmt.Errorf("media info is required")
	}
	if req.SourceOper == nil {
		return fmt.Errorf("source storage operator is required")
	}
	if req.TargetOper == nil {
		return fmt.Errorf("target storage operator is required")
	}
	if req.TargetPath == "" {
		return fmt.Errorf("target path is required")
	}
	if req.TransferType == "" {
		req.TransferType = string(TransferTypeCopy)
	}
	return nil
}

// resetResult 重置结果
func (th *TransHandler) resetResult() {
	th.result = &TransferInfo{
		Success:      true,
		Message:      "",
		FailedFiles:  make([]string, 0),
		TransferCount: 0,
		TotalCount:   0,
		NeedScrape:   false,
		ScrapeStatus: false,
		NeedNotify:   false,
		NotifyStatus: false,
	}
}

// setResult 设置结果
func (th *TransHandler) setResult(key string, value interface{}) {
	switch key {
	case "Success":
		if success, ok := value.(bool); ok {
			th.result.Success = success
		}
	case "Message":
		if message, ok := value.(string); ok {
			th.result.Message = message
		}
	case "SourceFile":
		if file, ok := value.(string); ok {
			th.result.SourceFile = file
		}
	case "TargetFile":
		if file, ok := value.(string); ok {
			th.result.TargetFile = file
		}
	case "FailedFiles":
		if files, ok := value.([]string); ok {
			th.result.FailedFiles = append(th.result.FailedFiles, files...)
		}
	case "TransferCount":
		if count, ok := value.(int); ok {
			th.result.TransferCount += count
		}
	case "TotalCount":
		if count, ok := value.(int); ok {
			th.result.TotalCount += count
		}
	case "NeedScrape":
		if need, ok := value.(bool); ok {
			th.result.NeedScrape = need
		}
	case "ScrapeStatus":
		if status, ok := value.(bool); ok {
			th.result.ScrapeStatus = status
		}
	case "NeedNotify":
		if need, ok := value.(bool); ok {
			th.result.NeedNotify = need
		}
	case "NotifyStatus":
		if status, ok := value.(bool); ok {
			th.result.NotifyStatus = status
		}
	}
}

// transferDirectory 转移目录
func (th *TransHandler) transferDirectory(ctx context.Context, req *TransHandlerRequest) error {
	// 列出目录下的所有文件
	files, err := req.SourceOper.ListFiles(ctx, req.FileItem.Path, true)
	if err != nil {
		return fmt.Errorf("failed to list directory files: %w", err)
	}

	th.setResult(TotalCount, len(files))

	// 过滤媒体文件
	mediaFiles := th.filterMediaFiles(files)
	if len(mediaFiles) == 0 {
		return fmt.Errorf("no media files found in directory")
	}

	// 逐个转移文件
	for _, file := range mediaFiles {
		if err := th.transferFile(ctx, req, file); err != nil {
			th.logger.Warn("Failed to transfer file", 
				zap.String("file", file.Path), 
				zap.Error(err))
			th.setResult(FailedFiles, []string{file.Path})
			continue
		}
		th.setResult(TransferCount, 1)
	}

	if len(th.result.FailedFiles) > 0 {
		return fmt.Errorf("failed to transfer %d files", len(th.result.FailedFiles))
	}

	return nil
}

// transferFile 转移单个文件
func (th *TransHandler) transferFile(ctx context.Context, req *TransHandlerRequest, file *FileItem) error {
	// 生成目标文件名
	targetFileName := file.Name
	if req.NeedRename {
		fileName, err := th.generateTargetFileName(ctx, req, file)
		if err != nil {
			th.logger.Warn("Failed to generate target file name", 
				zap.String("file", file.Path), 
				zap.Error(err))
		} else {
			targetFileName = fileName
		}
	}

	// 构建目标文件路径
	targetFilePath := filepath.Join(req.TargetPath, targetFileName)

	// 检查目标文件是否已存在
	if exists, err := req.TargetOper.FileExists(ctx, targetFilePath); err == nil && exists {
		// 处理覆盖逻辑
		shouldOverwrite, err := th.shouldOverwrite(ctx, req, file, targetFilePath)
		if err != nil {
			return err
		}
		if !shouldOverwrite {
			th.logger.Info("Target file already exists, skipping", 
				zap.String("target", targetFilePath))
			return nil
		}
	}

	// 执行文件转移
	switch TransferType(req.TransferType) {
	case TransferTypeCopy:
		return th.copyFile(ctx, req.SourceOper, req.TargetOper, file.Path, targetFilePath, file.Size)
	case TransferTypeMove:
		return th.moveFile(ctx, req.SourceOper, req.TargetOper, file.Path, targetFilePath, file.Size)
	case TransferTypeLink:
		return th.linkFile(ctx, req.SourceOper, req.TargetOper, file.Path, targetFilePath)
	case TransferTypeSoftLink:
		return th.softLinkFile(ctx, req.SourceOper, req.TargetOper, file.Path, targetFilePath)
	case TransferTypeRclone:
		return th.rcloneFile(ctx, req.SourceOper, req.TargetOper, file.Path, targetFilePath)
	default:
		return fmt.Errorf("unsupported transfer type: %s", req.TransferType)
	}
}

// copyFile 复制文件
func (th *TransHandler) copyFile(ctx context.Context, sourceOper, targetOper storage.Storage, sourcePath, targetPath string, size int64) error {
	// 创建目标目录
	targetDir := filepath.Dir(targetPath)
	if err := targetOper.CreateDirectory(ctx, targetDir); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// 如果是同一存储，使用存储的复制方法
	if sourceOper == targetOper {
		return sourceOper.CopyFile(ctx, sourcePath, targetPath)
	}

	// 不同存储，使用流式复制
	reader, err := sourceOper.OpenFile(ctx, sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer reader.Close()

	return targetOper.SaveFile(ctx, targetPath, reader, size)
}

// moveFile 移动文件
func (th *TransHandler) moveFile(ctx context.Context, sourceOper, targetOper storage.Storage, sourcePath, targetPath string, size int64) error {
	// 如果是同一存储，使用重命名
	if sourceOper == targetOper {
		return sourceOper.RenameFile(ctx, sourcePath, targetPath)
	}

	// 不同存储，先复制再删除
	if err := th.copyFile(ctx, sourceOper, targetOper, sourcePath, targetPath, size); err != nil {
		return err
	}
	return sourceOper.DeleteFile(ctx, sourcePath)
}

// linkFile 硬链接文件
func (th *TransHandler) linkFile(ctx context.Context, sourceOper, targetOper storage.Storage, sourcePath, targetPath string) error {
	// 硬链接只支持同一存储
	if sourceOper != targetOper {
		return fmt.Errorf("hard link only supported within same storage")
	}

	return sourceOper.LinkFile(ctx, sourcePath, targetPath)
}

// softLinkFile 软链接文件
func (th *TransHandler) softLinkFile(ctx context.Context, sourceOper, targetOper storage.Storage, sourcePath, targetPath string) error {
	// 软链接只支持同一存储
	if sourceOper != targetOper {
		return fmt.Errorf("soft link only supported within same storage")
	}

	return sourceOper.SymlinkFile(ctx, sourcePath, targetPath)
}

// rcloneFile 使用rclone传输文件
func (th *TransHandler) rcloneFile(ctx context.Context, sourceOper, targetOper storage.Storage, sourcePath, targetPath string) error {
	// rclone传输实现
	// 这里需要调用rclone命令行工具
	return fmt.Errorf("rclone transfer not implemented yet")
}

// filterMediaFiles 过滤媒体文件
func (th *TransHandler) filterMediaFiles(files []*FileItem) []*FileItem {
	var mediaFiles []*FileItem
	mediaExtensions := map[string]bool{
		".mp4":  true,
		".mkv":  true,
		".avi":  true,
		".mov":  true,
		".wmv":  true,
		".flv":  true,
		".webm": true,
		".m4v":  true,
		".mp3":  true,
		".flac": true,
		".wav":  true,
		".aac":  true,
		".ogg":  true,
	}

	for _, file := range files {
		if !file.IsDir {
			ext := strings.ToLower(filepath.Ext(file.Name))
			if mediaExtensions[ext] {
				mediaFiles = append(mediaFiles, file)
			}
		}
	}

	return mediaFiles
}

// generateTargetFileName 生成目标文件名
func (th *TransHandler) generateTargetFileName(ctx context.Context, req *TransHandlerRequest, file *FileItem) (string, error) {
	if th.metaHelper == nil {
		return file.Name, nil
	}

	// 识别文件元数据
	metaBase, err := th.metaHelper.RecognizeFile(ctx, file.Path)
	if err != nil {
		th.logger.Warn("Failed to recognize file", zap.String("file", file.Path), zap.Error(err))
		return file.Name, nil
	}

	// 生成文件名
	fileName, err := th.metaHelper.GenerateFileName(ctx, req.MediaInfo, metaBase)
	if err != nil {
		return file.Name, err
	}

	// 保持原扩展名
	ext := filepath.Ext(file.Name)
	return fileName + ext, nil
}

// shouldOverwrite 判断是否应该覆盖
func (th *TransHandler) shouldOverwrite(ctx context.Context, req *TransHandlerRequest, sourceFile *FileItem, targetPath string) (bool, error) {
	targetInfo, err := req.TargetOper.GetFileInfo(ctx, targetPath)
	if err != nil {
		return false, err
	}

	switch OverwriteMode(req.OverwriteMode) {
	case OverwriteModeNever:
		return false, nil
	case OverwriteModeAlways:
		return true, nil
	case OverwriteModeSmaller:
		return sourceFile.Size < targetInfo.Size, nil
	case OverwriteModeSize:
		return sourceFile.Size != targetInfo.Size, nil
	case OverwriteModeLatest:
		return sourceFile.ModTime > targetInfo.ModTime, nil
	default:
		return false, nil
	}
}

// scrapeMetadata 刮削元数据
func (th *TransHandler) scrapeMetadata(ctx context.Context, req *TransHandlerRequest) error {
	if th.tmdbHelper == nil {
		return fmt.Errorf("tmdb helper not set")
	}

	return th.tmdbHelper.ScrapeMetadata(ctx, th.result.TargetFile, req.MediaInfo)
}

// sendNotification 发送通知
func (th *TransHandler) sendNotification(ctx context.Context, req *TransHandlerRequest) error {
	// 发送转移完成事件
	if th.eventBus != nil {
		renameData := &TransferRenameEventData{
			SourceFile: th.result.SourceFile,
			TargetFile: th.result.TargetFile,
			MediaInfo:  req.MediaInfo,
		}
		if err := th.eventBus.EmitEvent(ctx, "transfer.rename", renameData); err != nil {
			return err
		}
	}

	return nil
}

// buildSearchPath 构建搜索路径
func (th *TransHandler) buildSearchPath(mediaInfo *models.MediaInfo, targetPath string, season int) string {
	if mediaInfo.Type == "movie" {
		return filepath.Join(targetPath, utils.CleanFileName(mediaInfo.Title))
	}

	// TV Series
	seasonDir := fmt.Sprintf("Season %02d", season)
	return filepath.Join(targetPath, utils.CleanFileName(mediaInfo.Title), seasonDir)
}

// extractEpisodeNumber 从文件名提取集数
func (th *TransHandler) extractEpisodeNumber(fileName string) int {
	// 使用正则表达式提取集数
	// 这里简化实现，实际应该使用更复杂的正则表达式
	re := utils.MustCompile(`[Ss](\d+)[Ee](\d+)`)
	matches := re.FindStringSubmatch(fileName)
	if len(matches) >= 3 {
		if episode, err := utils.Atoi(matches[2]); err == nil {
			return episode
		}
	}

	// 尝试其他格式
	re = utils.MustCompile(`E(\d+)`)
	matches = re.FindStringSubmatch(fileName)
	if len(matches) >= 2 {
		if episode, err := utils.Atoi(matches[1]); err == nil {
			return episode
		}
	}

	return 0
}

// 事件数据结构

// TransferInterceptEventData 转移拦截事件数据
type TransferInterceptEventData struct {
	FileItem  *FileItem        `json:"file_item"`
	MediaInfo *models.MediaInfo `json:"media_info"`
	MetaBase  *models.MetaBase `json:"meta_base"`
}

// TransferRenameEventData 转移重命名事件数据
type TransferRenameEventData struct {
	SourceFile string           `json:"source_file"`
	TargetFile string           `json:"target_file"`
	MediaInfo  *models.MediaInfo `json:"media_info"`
}

// MediaExistsHandlerRequest 媒体存在处理器请求
type MediaExistsHandlerRequest struct {
	MediaInfo  *models.MediaInfo `json:"media_info"`
	TargetPath string           `json:"target_path"`
	TargetOper storage.Storage   `json:"target_oper"`
	Season     int              `json:"season"`
}

// MoveFileHandlerRequest 移动文件处理器请求
type MoveFileHandlerRequest struct {
	SourcePath string         `json:"source_path"`
	TargetPath string         `json:"target_path"`
	SourceOper storage.Storage `json:"source_oper"`
	TargetOper storage.Storage `json:"target_oper"`
}