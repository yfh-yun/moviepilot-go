package filemanager

import (
	"path/filepath"
	
	"moviepilot-go/internal/core/config"
	"moviepilot-go/internal/core/context"
	"moviepilot-go/internal/core/meta"
	"moviepilot-go/internal/helper/directory"
	"moviepilot-go/internal/helper/module"
	"moviepilot-go/internal/logger"
	"moviepilot-go/internal/modules"
	"moviepilot-go/internal/modules/filemanager/storages"
	"moviepilot-go/internal/schemas"
	"moviepilot-go/internal/schemas/types"
	"moviepilot-go/internal/utils/system"
)

// FileManagerModule 文件整理模块
type FileManagerModule struct {
	directoryHelper   *directory.DirectoryHelper
	messageHelper     *message.MessageHelper
	storageSchemas    []storages.StorageBase
	supportStorages   []string
}

// NewFileManagerModule 创建文件管理器模块实�?func NewFileManagerModule() *FileManagerModule {
	return &FileManagerModule{
		directoryHelper: directory.NewDirectoryHelper(),
		messageHelper:   message.NewMessageHelper(),
	}
}

// InitModule 初始化模�?func (f *FileManagerModule) InitModule() {
	// 加载模块
	// 注意：Go 中的模块加载机制�?Python 不同，这里需要根据实际实现进行调�?	f.storageSchemas = module.Load("moviepilot-go/internal/modules/filemanager/storages",
		func(obj interface{}) bool {
			// 检查对象是否有 schema 字段
			// 实际实现需要根据具体结构进行调�?			return true
		})
	
	// 获取存储类型
	for _, storage := range f.storageSchemas {
		if storage.Schema() != nil {
			f.supportStorages = append(f.supportStorages, string(storage.Schema().Value))
		}
	}
}

// GetName 获取模块名称
func (f *FileManagerModule) GetName() string {
	return "文件整理"
}

// GetType 获取模块类型
func (f *FileManagerModule) GetType() types.ModuleType {
	return types.ModuleTypeOther
}

// GetSubtype 获取模块子类�?func (f *FileManagerModule) GetSubtype() types.OtherModulesType {
	return types.OtherModulesTypeFileManager
}

// GetPriority 获取模块优先�?func (f *FileManagerModule) GetPriority() int {
	// 获取模块优先级，数字越小优先级越高，只有同一接口下优先级才生�?	return 4
}

// Stop 停止模块
func (f *FileManagerModule) Stop() {
	// 空实�?}

// Test 测试模块连接�?func (f *FileManagerModule) Test() (bool, string) {
	// 测试模块连接�?	// 检查目�?	dirs := f.directoryHelper.GetDirs()
	if len(dirs) == 0 {
		return false, "未设置任何目�?
	}
	
	for _, d := range dirs {
		// 下载目录
		downloadPath := d.DownloadPath
		if downloadPath == "" {
			return false, d.Name + " 的下载目录未设置"
		}
		
		if d.Storage == "local" && !system.PathExists(downloadPath) {
			return false, d.Name + " 的下载目�?" + downloadPath + " 不存�?
		}
		
		// 媒体库目�?		libraryPath := d.LibraryPath
		if libraryPath == "" {
			return false, d.Name + " 的媒体库目录未设�?
		}
		
		if d.LibraryStorage == "local" && !system.PathExists(libraryPath) {
			return false, d.Name + " 的媒体库目录 " + libraryPath + " 不存�?
		}
		
		// 硬链�?		if d.TransferType == "link" &&
			d.Storage == "local" &&
			d.LibraryStorage == "local" &&
			!system.IsSameDisk(downloadPath, libraryPath) {
			return false, d.Name + " 的下载目�?" + downloadPath + " 与媒体库目录 " + libraryPath + " 不在同一磁盘，无法硬链接"
		}
		
		// 存储
		storageOper := f.getStorageOper(d.Storage)
		if storageOper == nil {
			return false, d.Name + " 的存储类�?" + d.Storage + " 不支�?
		}
		
		if !storageOper.Check() {
			return false, d.Name + " 的存储测试不通过"
		}
		
		if d.TransferType != "" && !storageOper.SupportTranstype()[d.TransferType] {
			return false, d.Name + " 的存储不支持 " + d.TransferType + " 整理方式"
		}
	}
	
	return true, ""
}

// getStorageOper 获取存储操作对象
func (f *FileManagerModule) getStorageOper(storage string, funcName ...string) storages.StorageBase {
	for _, storageSchema := range f.storageSchemas {
		if storageSchema.Schema() != nil && 
			string(storageSchema.Schema().Value) == storage &&
			(len(funcName) == 0 || hasFunc(storageSchema, funcName[0])) {
			return storageSchema
		}
	}
	return nil
}

// hasFunc 检查对象是否有指定方法
func hasFunc(obj interface{}, funcName string) bool {
	// 实现检查对象是否有指定方法的逻辑
	// �?Go 中可以通过反射实现
	return true
}

// SupportTranstype 支持的整理方�?func (f *FileManagerModule) SupportTranstype(storage string) map[string]string {
	if !contains(f.supportStorages, storage) {
		return nil
	}
	
	storageOper := f.getStorageOper(storage)
	if storageOper == nil {
		logger.Errorf("不支�?%s 的整理方式获�?, storage)
		return nil
	}
	
	return storageOper.SupportTranstype()
}

// contains 检查字符串是否在切片中
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// RecommendName 获取重命名后的名�?func (f *FileManagerModule) RecommendName(meta *meta.MetaBase, mediainfo *context.MediaInfo) string {
	handler := NewTransHandler()
	
	// 重命名格�?	renameFormat := config.Settings.RENAME_FORMAT(mediainfo.Type)
	
	// 获取集信�?	var episodesInfo []schemas.TmdbEpisode
	if mediainfo.Type == types.MediaTypeTV {
		// 判断注意season�?的情�?		seasonNum := mediainfo.Season
		if seasonNum == nil && meta.SeasonSeq != "" {
			// 判断是否为数�?			if isNumeric(meta.SeasonSeq) {
				num := parseInt(meta.SeasonSeq)
				seasonNum = &num
			}
		}
		
		// 默认�?
		if seasonNum == nil {
			num := 1
			seasonNum = &num
		}
		
		// 这里需要调�?TmdbChain().tmdb_episodes 方法
		// episodesInfo = TmdbChain().tmdb_episodes(
		//     tmdbid=mediainfo.tmdb_id,
		//     season=seasonNum,
		//     episode_group=mediainfo.episode_group,
		// )
	}
	
	// 获取重命名后的名�?	path := handler.getRenamePath(
		renameFormat,
		handler.getNamingDict(meta, mediainfo, filepath.Ext(meta.Title), episodesInfo),
		"",
	)
	
	return path
}

// Transfer 文件整理
func (f *FileManagerModule) Transfer(
	fileitem *schemas.FileItem,
	meta *meta.MetaBase,
	mediainfo *context.MediaInfo,
	targetDirectory *schemas.TransferDirectoryConf,
	targetStorage string,
	targetPath string,
	transferType string,
	scrape *bool,
	libraryTypeFolder *bool,
	libraryCategoryFolder *bool,
	episodesInfo []schemas.TmdbEpisode,
	sourceOper storages.StorageBase,
	targetOper storages.StorageBase,
) *schemas.TransferInfo {
	
	handler := NewTransHandler()
	
	// 检查目录路�?	if fileitem.Storage == "local" && !system.PathExists(fileitem.Path) {
		return &schemas.TransferInfo{
			Success: false,
			FileItem: fileitem,
			Message: fileitem.Path + " 不存�?,
		}
	}
	
	// 目标路径不能是文�?	if targetPath != "" && system.IsFile(targetPath) {
		logger.Errorf("整理目标路径 %s 是一个文�?, targetPath)
		return &schemas.TransferInfo{
			Success: false,
			FileItem: fileitem,
			Message: targetPath + " 不是有效目录",
		}
	}
	
	// 获取目标路径
	var needScrape bool
	needRename := true
	needNotify := false
	overwriteMode := "never"
	
	if targetDirectory != nil {
		// 目标媒体库目录未设置
		if targetDirectory.LibraryPath == "" {
			logger.Errorf("%s %s 未找到有效的媒体库目录，无法整理文件，源路径�?s", 
				mediainfo.Type, mediainfo.TitleYear, fileitem.Path)
			return &schemas.TransferInfo{
				Success: false,
				FileItem: fileitem,
				Message: "未找到有效的媒体库目�?,
			}
		}
		
		// 整理方式
		if transferType == "" {
			transferType = targetDirectory.TransferType
		}
		
		// 目标存储
		if targetStorage == "" {
			targetStorage = targetDirectory.LibraryStorage
		}
		
		// 是否需要重命名
		needRename = targetDirectory.Renaming
		
		// 是否需要通知
		needNotify = targetDirectory.Notify
		
		// 覆盖模式
		overwriteMode = targetDirectory.OverwriteMode
		
		// 是否需要刮�?		if scrape != nil {
			needScrape = *scrape
		} else {
			needScrape = targetDirectory.Scraping
		}
		
		// 拼装媒体库一、二级子目录
		targetPath = handler.getDestDir(mediainfo, targetDirectory, libraryTypeFolder, libraryCategoryFolder)
	} else if targetPath != "" {
		if scrape != nil {
			needScrape = *scrape
		} else {
			needScrape = false
		}
		needRename = true
		needNotify = false
		overwriteMode = "never"
		// 手动整理的场景，有自定义目标路径
		targetPath = handler.getDestPath(mediainfo, targetPath, libraryTypeFolder != nil && *libraryTypeFolder, libraryCategoryFolder != nil && *libraryCategoryFolder)
	} else {
		// 未找到有效的媒体库目�?		logger.Errorf("%s %s 未找到有效的媒体库目录，无法整理文件，源路径�?s",
			mediainfo.Type, mediainfo.TitleYear, fileitem.Path)
		return &schemas.TransferInfo{
			Success: false,
			FileItem: fileitem,
			Message: "未找到有效的媒体库目�?,
		}
	}
	
	// 整理方式
	if transferType == "" {
		logger.Errorf("未设置整理方�?)
		return &schemas.TransferInfo{
			Success: false,
			FileItem: fileitem,
			Message: "未设置整理方�?,
		}
	}
	
	// 源操作对�?	if sourceOper == nil {
		sourceOper = f.getStorageOper(fileitem.Storage)
	}
	
	if sourceOper == nil {
		return &schemas.TransferInfo{
			Success: false,
			Message: "不支持的存储类型�? + fileitem.Storage,
			FileItem: fileitem,
			FailList: []string{fileitem.Path},
			TransferType: transferType,
			NeedNotify: needNotify,
		}
	}
	
	// 目的操作对象
	if targetOper == nil {
		if targetStorage == "" {
			targetStorage = fileitem.Storage
		}
		targetOper = f.getStorageOper(targetStorage)
	}
	
	if targetOper == nil {
		return &schemas.TransferInfo{
			Success: false,
			Message: "不支持的存储类型�? + targetStorage,
			FileItem: fileitem,
			FailList: []string{fileitem.Path},
			TransferType: transferType,
			NeedNotify: needNotify,
		}
	}
	
	// 整理
	logger.Infof("获取整理目标路径：�?s�?s", targetStorage, targetPath)
	
	// 实际调用 handler.TransferMedia
	return handler.TransferMedia(
		fileitem,
		meta,
		mediainfo,
		targetStorage,
		targetPath,
		transferType,
		sourceOper,
		targetOper,
		needScrape,
		needRename,
		needNotify,
		overwriteMode,
		episodesInfo,
	)
}

// GetInstance 获取文件管理器模块实�?func GetInstance() modules.Module {
	return NewFileManagerModule()
}

// 辅助函数
func isNumeric(s string) bool {
	_, err := parseInt(s)
	return err == nil
}

func parseInt(s string) (int, error) {
	// 实现字符串转整数的逻辑
	return 0, nil
}
