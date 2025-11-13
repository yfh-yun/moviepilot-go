package utils

import (
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"moviepilot-go/internal/config"
	"moviepilot-go/pkg/models"
)

// DirectoryHelper 下载目录/媒体库目录帮助类
type DirectoryHelper struct {
}

// JINJA2_VAR_PATTERN Jinja2变量匹配模式
var JINJA2_VAR_PATTERN = regexp.MustCompile(`\{\{.*?\}\}`)

// NewDirectoryHelper 创建一个新�?DirectoryHelper 实例
func NewDirectoryHelper() *DirectoryHelper {
	return &DirectoryHelper{}
}

// GetDirs 获取所有下载目�?func (dh *DirectoryHelper) GetDirs() []*models.TransferDirectoryConf {
	/*
		获取所有下载目�?	*/
	settings := config.GetConfig()
	
	// 从配置中获取目录配置
	dirConfs := settings.Directories
	if dirConfs == nil {
		return []*models.TransferDirectoryConf{}
	}
	
	// 转换为TransferDirectoryConf对象列表
	var result []*models.TransferDirectoryConf
	for _, d := range dirConfs {
		conf := &models.TransferDirectoryConf{
			Name:                   d.Name,
			Priority:               d.Priority,
			Storage:                d.Storage,
			DownloadPath:           d.DownloadPath,
			MediaType:              d.MediaType,
			MediaCategory:          d.MediaCategory,
			DownloadTypeFolder:     d.DownloadTypeFolder,
			DownloadCategoryFolder: d.DownloadCategoryFolder,
			MonitorType:            d.MonitorType,
			MonitorMode:            d.MonitorMode,
			TransferType:           d.TransferType,
			OverwriteMode:          d.OverwriteMode,
			LibraryPath:            d.LibraryPath,
			LibraryStorage:         d.LibraryStorage,
			Renaming:               d.Renaming,
			Scraping:               d.Scraping,
			Notify:                 d.Notify,
			LibraryTypeFolder:      d.LibraryTypeFolder,
			LibraryCategoryFolder:  d.LibraryCategoryFolder,
		}
		result = append(result, conf)
	}
	
	return result
}

// GetDownloadDirs 获取所有下载目�?func (dh *DirectoryHelper) GetDownloadDirs() []*models.TransferDirectoryConf {
	/*
		获取所有下载目�?	*/
	var result []*models.TransferDirectoryConf
	for _, d := range dh.GetDirs() {
		if d.DownloadPath != "" {
			result = append(result, d)
		}
	}
	
	// 按优先级排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].Priority < result[j].Priority
	})
	
	return result
}

// GetLocalDownloadDirs 获取所有本地的可下载目�?func (dh *DirectoryHelper) GetLocalDownloadDirs() []*models.TransferDirectoryConf {
	/*
		获取所有本地的可下载目�?	*/
	var result []*models.TransferDirectoryConf
	for _, d := range dh.GetDownloadDirs() {
		if d.Storage == "local" {
			result = append(result, d)
		}
	}
	return result
}

// GetLibraryDirs 获取所有媒体库目录
func (dh *DirectoryHelper) GetLibraryDirs() []*models.TransferDirectoryConf {
	/*
		获取所有媒体库目录
	*/
	var result []*models.TransferDirectoryConf
	for _, d := range dh.GetDirs() {
		if d.LibraryPath != "" {
			result = append(result, d)
		}
	}
	
	// 按优先级排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].Priority < result[j].Priority
	})
	
	return result
}

// GetLocalLibraryDirs 获取所有本地的媒体库目�?func (dh *DirectoryHelper) GetLocalLibraryDirs() []*models.TransferDirectoryConf {
	/*
		获取所有本地的媒体库目�?	*/
	var result []*models.TransferDirectoryConf
	for _, d := range dh.GetLibraryDirs() {
		if d.LibraryStorage == "local" {
			result = append(result, d)
		}
	}
	return result
}

// GetDir 根据媒体信息获取下载目录、媒体库目录配置
func (dh *DirectoryHelper) GetDir(media *models.MediaInfo, includeUnsorted bool, storage string,
	srcPath string, targetStorage string, destPath string) *models.TransferDirectoryConf {
	/*
		根据媒体信息获取下载目录、媒体库目录配置
		:param media: 媒体信息
		:param include_unsorted: 包含不整理目�?		:param storage: 源存储类�?		:param target_storage: 目标存储类型
		:param src_path: 源目录，有值时直接匹配
		:param dest_path: 目标目录，有值时直接匹配
	*/
	// 处理类型
	if media == nil {
		return nil
	}
	
	// 电影/电视�?	mediaType := media.Type
	dirs := dh.GetDirs()

	// 如果存在源目录，并源目录为任一下载目录的子目录时，则进行源目录匹配，否则，允许源目录按同盘优先的逻辑匹配
	var matchingDirs []*models.TransferDirectoryConf
	if srcPath != "" {
		for _, d := range dirs {
			// 检查srcPath是否是d.DownloadPath的子目录
			rel, err := filepath.Rel(d.DownloadPath, srcPath)
			if err == nil && !strings.HasPrefix(rel, "..") {
				matchingDirs = append(matchingDirs, d)
			}
		}
	}
	
	// 根据是否有匹配的源目录，决定要考虑的目录集�?	var dirsToConsider []*models.TransferDirectoryConf
	if len(matchingDirs) > 0 {
		dirsToConsider = matchingDirs
	} else {
		dirsToConsider = dirs
	}

	// 已匹配的目录
	var matchedDirs []*models.TransferDirectoryConf
	// 按照配置顺序查找
	for _, d := range dirsToConsider {
		// 没有启用整理的目�?		if d.MonitorType == "" && !includeUnsorted {
			continue
		}
		// 源存储类型不匹配
		if storage != "" && d.Storage != storage {
			continue
		}
		// 目标存储类型不匹�?		if targetStorage != "" && d.LibraryStorage != targetStorage {
			continue
		}
		// 有目标目录时，目标目录不匹配媒体库目�?		if destPath != "" && destPath != d.LibraryPath {
			continue
		}
		// 目录类型为全部的，符合条�?		if d.MediaType == "" {
			matchedDirs = append(matchedDirs, d)
			continue
		}
		// 目录类型相等，目录类别为全部，符合条�?		if d.MediaType == mediaType && d.MediaCategory == "" {
			matchedDirs = append(matchedDirs, d)
			continue
		}
		// 目录类型相等，目录类别相等，符合条件
		if d.MediaType == mediaType && d.MediaCategory == media.Category {
			matchedDirs = append(matchedDirs, d)
			continue
		}
	}
	
	if len(matchedDirs) > 0 {
		if srcPath != "" {
			// 优先源目录同�?			systemUtils := NewSystemUtils()
			for _, matchedDir := range matchedDirs {
				matchedPath := matchedDir.DownloadPath
				if systemUtils.IsSameDisk(matchedPath, srcPath) {
					return matchedDir
				}
			}
		}
		return matchedDirs[0]
	}
	return nil
}

// GetMediaRootPath 获取重命名后的媒体文件根路径
func (dh *DirectoryHelper) GetMediaRootPath(renameFormat string, renamePath string) string {
	/*
		获取重命名后的媒体文件根路径

		:param rename_format: 重命名格�?		:param rename_path: 重命名后的路�?		:return: 媒体文件根路�?	*/
	if renameFormat == "" {
		// logger.Error("重命名格式不能为�?)
		return ""
	}
	
	// 计算重命名中的文件夹层数
	renameList := strings.Split(renameFormat, "/")
	renameFormatLevel := len(renameList) - 1
	
	// 查找标题参数所在层
	foundTitle := false
	for level, name := range renameList {
		matchs := JINJA2_VAR_PATTERN.FindAllString(name, -1)
		if len(matchs) == 0 {
			continue
		}
		// 处理特例，有的人重命名的第一层是年份、分辨率
		hasTitle := false
		for _, m := range matchs {
			if strings.Contains(m, "title") {
				hasTitle = true
				break
			}
		}
		if hasTitle {
			// 找出含标题的这一层作为媒体根目录
			renameFormatLevel -= level
			foundTitle = true
			break
		}
	}
	
	if !foundTitle {
		// 假定第一层目录是媒体根目�?		// logger.Warn(f"重命名格�?{rename_format} 缺少标题参数")
	}
	
	// 解析renamePath为路径组�?	pathComponents := strings.Split(renamePath, string(filepath.Separator))
	
	if renameFormatLevel > len(pathComponents) {
		// 通常因为路径�?结尾，被Path规范化删除了
		// logger.Error(f"路径 {rename_path} 不匹配重命名格式 {rename_format}")
		return ""
	}
	
	if renameFormatLevel <= 0 {
		// 所有媒体文件都存在一个目录内的特殊需�?		renameFormatLevel = 1
	}
	
	// 媒体根路�?	// 计算从路径组件末尾开始的索引
	startIndex := len(pathComponents) - renameFormatLevel
	if startIndex < 0 {
		startIndex = 0
	}
	
	mediaRoot := filepath.Join(pathComponents[startIndex:]...)
	return mediaRoot
}
