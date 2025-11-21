// Package filemanager 文件整理模块
// 提供文件转移、重命名、整理等功能
package filemanager

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"moviepilot-go/pkg/logger"
	"moviepilot-go/internal/models"
	"moviepilot-go/internal/models"

	"go.uber.org/zap"
)

// ModuleType 模块类型
type ModuleType string

const (
	ModuleTypeOther ModuleType = "Other"
)

// OtherModulesType 其他模块类型
type OtherModulesType string

const (
	OtherModulesTypeFileManager OtherModulesType = "FileManager"
)

// TransferType 转移类型
type TransferType string

const (
	TransferTypeCopy    TransferType = "copy"
	TransferTypeMove    TransferType = "move"
	TransferTypeLink    TransferType = "link"
	TransferTypeSoftLink TransferType = "softlink"
	TransferTypeRclone  TransferType = "rclone"
)

// OverwriteMode 覆盖模式
type OverwriteMode string

const (
	OverwriteModeNever    OverwriteMode = "never"
	OverwriteModeSmaller  OverwriteMode = "smaller"
	OverwriteModeSize     OverwriteMode = "size"
	OverwriteModeAlways   OverwriteMode = "always"
	OverwriteModeLatest   OverwriteMode = "latest"
)

// FileManagerModule 文件整理模块
type FileManagerModule struct {
	logger          *zap.Logger
	storageManager  *storage.Manager
	directoryHelper *DirectoryHelper
	messageHelper   *MessageHelper
	transHandler    *TransHandler
	storageSchemas  []string
	supportStorages []string
	mutex           sync.RWMutex
}

// NewFileManagerModule 创建文件整理模块
func NewFileManagerModule(storageManager *storage.Manager) *FileManagerModule {
	return &FileManagerModule{
		logger:          logger.Logger,
		storageManager:  storageManager,
		directoryHelper: NewDirectoryHelper(),
		messageHelper:   NewMessageHelper(),
		transHandler:    NewTransHandler(),
		storageSchemas:  []string{},
		supportStorages: []string{},
	}
}

// InitModule 初始化模块
func (fm *FileManagerModule) InitModule(ctx context.Context) error {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	// 加载支持的存储类型
	fm.storageSchemas = fm.storageManager.GetSupportedSchemas()
	fm.supportStorages = fm.storageSchemas

	fm.logger.Info("File manager module initialized",
		zap.Strings("supported_storages", fm.supportStorages))

	return nil
}

// GetName 获取模块名称
func (fm *FileManagerModule) GetName() string {
	return "文件整理"
}

// GetType 获取模块类型
func (fm *FileManagerModule) GetType() ModuleType {
	return ModuleTypeOther
}

// GetSubType 获取模块子类型
func (fm *FileManagerModule) GetSubType() OtherModulesType {
	return OtherModulesTypeFileManager
}

// GetPriority 获取模块优先级
func (fm *FileManagerModule) GetPriority() int {
	return 4
}

// Stop 停止模块
func (fm *FileManagerModule) Stop(ctx context.Context) error {
	fm.logger.Info("File manager module stopped")
	return nil
}

// Test 测试模块连接性
func (fm *FileManagerModule) Test(ctx context.Context) error {
	// 检查目录配置
	dirs, err := fm.directoryHelper.GetDirs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get directories: %w", err)
	}

	if len(dirs) == 0 {
		return fmt.Errorf("未设置任何目录")
	}

	for _, dir := range dirs {
		// 检查下载目录
		if dir.DownloadPath == "" {
			return fmt.Errorf("%s 的下载目录未设置", dir.Name)
		}

		if dir.Storage == "local" {
			if !fm.directoryHelper.Exists(dir.DownloadPath) {
				return fmt.Errorf("%s 的下载目录 %s 不存在", dir.Name, dir.DownloadPath)
			}
		}

		// 检查媒体库目录
		if dir.LibraryPath == "" {
			return fmt.Errorf("%s 的媒体库目录未设置", dir.Name)
		}

		if dir.LibraryStorage == "local" {
			if !fm.directoryHelper.Exists(dir.LibraryPath) {
				return fmt.Errorf("%s 的媒体库目录 %s 不存在", dir.Name, dir.LibraryPath)
			}
		}

		// 检查硬链接条件
		if dir.TransferType == string(TransferTypeLink) &&
			dir.Storage == "local" &&
			dir.LibraryStorage == "local" &&
			!fm.directoryHelper.IsSameDisk(dir.DownloadPath, dir.LibraryPath) {
			return fmt.Errorf("%s 的下载目录 %s 与媒体库目录 %s 不在同一磁盘，无法硬链接",
				dir.Name, dir.DownloadPath, dir.LibraryPath)
		}

		// 检查存储支持
		storageOper := fm.getStorageOper(dir.Storage)
		if storageOper == nil {
			return fmt.Errorf("%s 的存储类型 %s 不支持", dir.Name, dir.Storage)
		}

		if err := storageOper.Check(ctx); err != nil {
			return fmt.Errorf("%s 存储检查失败: %w", dir.Name, err)
		}
	}

	return nil
}

// TransferMedia 转移媒体文件
func (fm *FileManagerModule) TransferMedia(ctx context.Context, req *TransferRequest) (*TransferInfo, error) {
	// 获取源存储操作对象
	sourceOper := fm.getStorageOper(req.SourceStorage)
	if sourceOper == nil {
		return nil, fmt.Errorf("不支持的源存储类型: %s", req.SourceStorage)
	}

	// 获取目标存储操作对象
	targetOper := fm.getStorageOper(req.TargetStorage)
	if targetOper == nil {
		return nil, fmt.Errorf("不支持的目标存储类型: %s", req.TargetStorage)
	}

	// 执行转移
	return fm.transHandler.TransferMedia(ctx, &TransHandlerRequest{
		FileItem:        req.FileItem,
		InMeta:          req.InMeta,
		MediaInfo:       req.MediaInfo,
		TargetStorage:   req.TargetStorage,
		TargetPath:      req.TargetPath,
		TransferType:    req.TransferType,
		SourceOper:      sourceOper,
		TargetOper:      targetOper,
		NeedScrape:      req.NeedScrape,
		NeedRename:      req.NeedRename,
		NeedNotify:      req.NeedNotify,
		OverwriteMode:   req.OverwriteMode,
		EpisodesInfo:    req.EpisodesInfo,
	})
}

// GetStorageOper 获取存储操作对象
func (fm *FileManagerModule) getStorageOper(storageType string) storage.Storage {
	return fm.storageManager.GetStorage(storageType)
}

// GetSupportStorages 获取支持的存储类型
func (fm *FileManagerModule) GetSupportStorages() []string {
	fm.mutex.RLock()
	defer fm.mutex.RUnlock()

	result := make([]string, len(fm.supportStorages))
	copy(result, fm.supportStorages)
	return result
}

// MediaExists 媒体是否存在
func (fm *FileManagerModule) MediaExists(ctx context.Context, req *MediaExistsRequest) (*ExistMediaInfo, error) {
	targetOper := fm.getStorageOper(req.Storage)
	if targetOper == nil {
		return nil, fmt.Errorf("不支持的存储类型: %s", req.Storage)
	}

	return fm.transHandler.MediaExists(ctx, &MediaExistsHandlerRequest{
		MediaInfo:  req.MediaInfo,
		TargetPath: req.TargetPath,
		TargetOper: targetOper,
		Season:     req.Season,
	})
}

// GetStorageUsage 获取存储使用情况
func (fm *FileManagerModule) GetStorageUsage(ctx context.Context, storageType string) (*StorageUsage, error) {
	storageOper := fm.getStorageOper(storageType)
	if storageOper == nil {
		return nil, fmt.Errorf("不支持的存储类型: %s", storageType)
	}

	return storageOper.GetUsage(ctx)
}

// ListFiles 列出文件
func (fm *FileManagerModule) ListFiles(ctx context.Context, req *ListFilesRequest) ([]*FileItem, error) {
	storageOper := fm.getStorageOper(req.Storage)
	if storageOper == nil {
		return nil, fmt.Errorf("不支持的存储类型: %s", req.Storage)
	}

	return storageOper.ListFiles(ctx, req.Path, req.Recursive)
}

// DeleteFile 删除文件
func (fm *FileManagerModule) DeleteFile(ctx context.Context, req *DeleteFileRequest) error {
	storageOper := fm.getStorageOper(req.Storage)
	if storageOper == nil {
		return fmt.Errorf("不支持的存储类型: %s", req.Storage)
	}

	return storageOper.DeleteFile(ctx, req.Path)
}

// MoveFile 移动文件
func (fm *FileManagerModule) MoveFile(ctx context.Context, req *MoveFileRequest) error {
	sourceOper := fm.getStorageOper(req.SourceStorage)
	if sourceOper == nil {
		return fmt.Errorf("不支持的源存储类型: %s", req.SourceStorage)
	}

	targetOper := fm.getStorageOper(req.TargetStorage)
	if targetOper == nil {
		return fmt.Errorf("不支持的目标存储类型: %s", req.TargetStorage)
	}

	return fm.transHandler.MoveFile(ctx, &MoveFileHandlerRequest{
		SourcePath:   req.SourcePath,
		TargetPath:   req.TargetPath,
		SourceOper:   sourceOper,
		TargetOper:   targetOper,
	})
}

// CreateDirectory 创建目录
func (fm *FileManagerModule) CreateDirectory(ctx context.Context, req *CreateDirectoryRequest) error {
	storageOper := fm.getStorageOper(req.Storage)
	if storageOper == nil {
		return fmt.Errorf("不支持的存储类型: %s", req.Storage)
	}

	return storageOper.CreateDirectory(ctx, req.Path)
}

// GetFileHash 获取文件哈希
func (fm *FileManagerModule) GetFileHash(ctx context.Context, req *GetFileHashRequest) (string, error) {
	storageOper := fm.getStorageOper(req.Storage)
	if storageOper == nil {
		return "", fmt.Errorf("不支持的存储类型: %s", req.Storage)
	}

	return storageOper.GetFileHash(ctx, req.Path, req.Algorithm)
}

// 请求和响应结构体定义

// TransferRequest 转移请求
type TransferRequest struct {
	FileItem      *FileItem                `json:"file_item"`
	InMeta        *models.MetaBase          `json:"in_meta"`
	MediaInfo     *models.MediaInfo         `json:"media_info"`
	TargetStorage string                   `json:"target_storage"`
	TargetPath    string                   `json:"target_path"`
	TransferType  string                   `json:"transfer_type"`
	SourceStorage string                   `json:"source_storage"`
	NeedScrape    bool                     `json:"need_scrape"`
	NeedRename    bool                     `json:"need_rename"`
	NeedNotify    bool                     `json:"need_notify"`
	OverwriteMode string                   `json:"overwrite_mode"`
	EpisodesInfo  []*models.TmdbEpisode     `json:"episodes_info"`
}

// TransferInfo 转移信息
type TransferInfo struct {
	Success        bool     `json:"success"`
	Message        string   `json:"message"`
	SourceFile     string   `json:"source_file"`
	TargetFile     string   `json:"target_file"`
	FailedFiles    []string `json:"failed_files"`
	TransferCount  int      `json:"transfer_count"`
	TotalCount     int      `json:"total_count"`
	NeedScrape     bool     `json:"need_scrape"`
	ScrapeStatus   bool     `json:"scrape_status"`
	NeedNotify     bool     `json:"need_notify"`
	NotifyStatus   bool     `json:"notify_status"`
}

// MediaExistsRequest 媒体存在检查请求
type MediaExistsRequest struct {
	MediaInfo  *models.MediaInfo `json:"media_info"`
	TargetPath string           `json:"target_path"`
	Storage    string           `json:"storage"`
	Season     int              `json:"season"`
}

// ExistMediaInfo 存在的媒体信息
type ExistMediaInfo struct {
	Exists   bool     `json:"exists"`
	Path     string   `json:"path"`
	Files    []string `json:"files"`
	Season   int      `json:"season"`
	Episodes []int    `json:"episodes"`
}

// StorageUsage 存储使用情况
type StorageUsage struct {
	Total     uint64 `json:"total"`
	Used      uint64 `json:"used"`
	Free      uint64 `json:"free"`
	Available bool   `json:"available"`
}

// ListFilesRequest 列出文件请求
type ListFilesRequest struct {
	Storage   string `json:"storage"`
	Path      string `json:"path"`
	Recursive bool   `json:"recursive"`
}

// FileItem 文件项
type FileItem struct {
	Path     string `json:"path"`
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	IsDir    bool   `json:"is_dir"`
	ModTime  int64  `json:"mod_time"`
	Hash     string `json:"hash"`
	FileType string `json:"file_type"`
}

// DeleteFileRequest 删除文件请求
type DeleteFileRequest struct {
	Storage string `json:"storage"`
	Path    string `json:"path"`
}

// MoveFileRequest 移动文件请求
type MoveFileRequest struct {
	SourceStorage string `json:"source_storage"`
	TargetStorage string `json:"target_storage"`
	SourcePath   string `json:"source_path"`
	TargetPath   string `json:"target_path"`
}

// CreateDirectoryRequest 创建目录请求
type CreateDirectoryRequest struct {
	Storage string `json:"storage"`
	Path    string `json:"path"`
}

// GetFileHashRequest 获取文件哈希请求
type GetFileHashRequest struct {
	Storage  string `json:"storage"`
	Path     string `json:"path"`
	Algorithm string `json:"algorithm"`
}