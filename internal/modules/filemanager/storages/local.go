package storages

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"unsafe"

	"moviepilot-go/internal/helper/directory"
	"moviepilot-go/internal/helper/progress"
	"moviepilot-go/internal/helper/storage"
	"moviepilot-go/internal/logger"
	"moviepilot-go/internal/schemas"
	"moviepilot-go/internal/schemas/types"
	"moviepilot-go/internal/utils/crypto"
	"moviepilot-go/internal/utils/system"
)

// LocalStorage 本地文件操作
type LocalStorage struct {
	BaseStorage
	chunkSize int64
}

// NewLocalStorage 创建本地存储实例
func NewLocalStorage() *LocalStorage {
	return &LocalStorage{
		BaseStorage: *NewBaseStorage(),
		chunkSize:   10 * 1024 * 1024, // 10MB
	}
}

// Schema 获取存储模式
func (l *LocalStorage) Schema() *StorageSchema {
	return &StorageSchema{Value: string(types.StorageSchemaLocal)}
}

// InitStorage 初始�?func (l *LocalStorage) InitStorage() {
	// 空实�?}

// Check 检查存储是否可�?func (l *LocalStorage) Check() bool {
	return true
}

// getFileItem 获取文件�?func (l *LocalStorage) getFileItem(path string) *schemas.FileItem {
	pathObj := filepath.Clean(path)
	
	// 获取文件信息
	fileInfo, err := os.Stat(pathObj)
	if err != nil {
		return nil
	}
	
	extension := filepath.Ext(pathObj)
	if extension != "" {
		extension = extension[1:] // 移除点号
	}
	
	return &schemas.FileItem{
		Storage:    string(types.StorageSchemaLocal),
		Type:       "file",
		Path:       pathObj,
		Name:       filepath.Base(pathObj),
		Basename:   strings.TrimSuffix(filepath.Base(pathObj), filepath.Ext(pathObj)),
		Extension:  &extension,
		Size:       fileInfo.Size(),
		ModifyTime: float64(fileInfo.ModTime().Unix()),
	}
}

// getDirItem 获取目录�?func (l *LocalStorage) getDirItem(path string) *schemas.FileItem {
	pathObj := filepath.Clean(path)
	
	// 获取目录信息
	fileInfo, err := os.Stat(pathObj)
	if err != nil {
		return nil
	}
	
	// 确保目录路径以分隔符结尾
	dirPath := pathObj
	if !strings.HasSuffix(dirPath, string(filepath.Separator)) {
		dirPath += string(filepath.Separator)
	}
	
	return &schemas.FileItem{
		Storage:    string(types.StorageSchemaLocal),
		Type:       "dir",
		Path:       dirPath,
		Name:       filepath.Base(pathObj),
		Basename:   strings.TrimSuffix(filepath.Base(pathObj), filepath.Ext(pathObj)),
		ModifyTime: float64(fileInfo.ModTime().Unix()),
	}
}

// List 浏览文件
func (l *LocalStorage) List(fileItem *schemas.FileItem) []*schemas.FileItem {
	// 返回结果
	var retItems []*schemas.FileItem
	path := fileItem.Path
	
	if path == "" || path == "/" {
		if runtime.GOOS == "windows" {
			// Windows系统获取分区
			partitions := system.GetWindowsDrives()
			if len(partitions) == 0 {
				partitions = []string{"C:"}
			}
			
			for _, partition := range partitions {
				partitionPath := partition + ":\\"
				retItems = append(retItems, &schemas.FileItem{
					Storage:  string(types.StorageSchemaLocal),
					Type:     "dir",
					Path:     partitionPath,
					Name:     partition + ":",
					Basename: partition,
				})
			}
			return retItems
		} else {
			path = "/"
		}
	} else {
		if runtime.GOOS == "windows" {
			// Windows系统处理路径
			path = strings.TrimPrefix(path, "/")
		} else {
			// Unix系统确保路径�?开�?			if !strings.HasPrefix(path, "/") {
				path = "/" + path
			}
		}
	}
	
	// 检查路径是否存�?	pathObj := filepath.Clean(path)
	if _, err := os.Stat(pathObj); os.IsNotExist(err) {
		logger.Warnf("【本地】目录不存在�?s", path)
		return []*schemas.FileItem{}
	}
	
	// 如果是文�?	fileInfo, _ := os.Stat(pathObj)
	if fileInfo != nil && !fileInfo.IsDir() {
		retItems = append(retItems, l.getFileItem(pathObj))
		return retItems
	}
	
	// 遍历目录
	entries, err := os.ReadDir(pathObj)
	if err != nil {
		logger.Warnf("【本地】读取目录失败：%s", err.Error())
		return []*schemas.FileItem{}
	}
	
	for _, entry := range entries {
		itemPath := filepath.Join(pathObj, entry.Name())
		if entry.IsDir() {
			retItems = append(retItems, l.getDirItem(itemPath))
		} else {
			retItems = append(retItems, l.getFileItem(itemPath))
		}
	}
	
	return retItems
}

// CreateFolder 创建目录
func (l *LocalStorage) CreateFolder(fileItem *schemas.FileItem, name string) *schemas.FileItem {
	if fileItem.Path == "" {
		return nil
	}
	
	pathObj := filepath.Join(fileItem.Path, name)
	
	// 检查目录是否存在，不存在则创建
	if _, err := os.Stat(pathObj); os.IsNotExist(err) {
		err := os.MkdirAll(pathObj, 0755)
		if err != nil {
			logger.Errorf("【本地】创建目录失败：%s", err.Error())
			return nil
		}
	}
	
	return l.getDirItem(pathObj)
}

// GetFolder 获取目录
func (l *LocalStorage) GetFolder(path string) *schemas.FileItem {
	pathObj := filepath.Clean(path)
	
	// 检查目录是否存在，不存在则创建
	if _, err := os.Stat(pathObj); os.IsNotExist(err) {
		err := os.MkdirAll(pathObj, 0755)
		if err != nil {
			logger.Errorf("【本地】创建目录失败：%s", err.Error())
			return nil
		}
	}
	
	return l.getDirItem(pathObj)
}

// GetItem 获取文件或目录，不存在返回nil
func (l *LocalStorage) GetItem(path string) *schemas.FileItem {
	pathObj := filepath.Clean(path)
	
	// 检查路径是否存�?	if _, err := os.Stat(pathObj); os.IsNotExist(err) {
		return nil
	}
	
	// 判断是文件还是目�?	fileInfo, _ := os.Stat(pathObj)
	if fileInfo != nil && fileInfo.IsDir() {
		return l.getDirItem(pathObj)
	}
	
	return l.getFileItem(pathObj)
}

// Detail 获取文件详情
func (l *LocalStorage) Detail(fileItem *schemas.FileItem) *schemas.FileItem {
	pathObj := filepath.Clean(fileItem.Path)
	
	// 检查路径是否存�?	if _, err := os.Stat(pathObj); os.IsNotExist(err) {
		return nil
	}
	
	return l.getFileItem(pathObj)
}

// Delete 删除文件
func (l *LocalStorage) Delete(fileItem *schemas.FileItem) bool {
	if fileItem.Path == "" {
		return false
	}
	
	pathObj := filepath.Clean(fileItem.Path)
	
	// 检查路径是否存�?	if _, err := os.Stat(pathObj); os.IsNotExist(err) {
		return true
	}
	
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("【本地】删除文件异常：%v", r)
		}
	}()
	
	// 判断是文件还是目�?	fileInfo, _ := os.Stat(pathObj)
	if fileInfo != nil && fileInfo.IsDir() {
		// 删除目录
		err := os.RemoveAll(pathObj)
		if err != nil {
			logger.Errorf("【本地】删除目录失败：%s", err.Error())
			return false
		}
	} else {
		// 删除文件
		err := os.Remove(pathObj)
		if err != nil {
			logger.Errorf("【本地】删除文件失败：%s", err.Error())
			return false
		}
	}
	
	return true
}

// Rename 重命名文�?func (l *LocalStorage) Rename(fileItem *schemas.FileItem, name string) bool {
	pathObj := filepath.Clean(fileItem.Path)
	
	// 检查路径是否存�?	if _, err := os.Stat(pathObj); os.IsNotExist(err) {
		return false
	}
	
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("【本地】重命名文件异常�?v", r)
		}
	}()
	
	// 重命�?	newPath := filepath.Join(filepath.Dir(pathObj), name)
	err := os.Rename(pathObj, newPath)
	if err != nil {
		logger.Errorf("【本地】重命名文件失败�?s", err.Error())
		return false
	}
	
	return true
}

// Download 下载文件
func (l *LocalStorage) Download(fileItem *schemas.FileItem, path string) string {
	return fileItem.Path
}

// copyWithProgress 分块复制文件并回调进�?func (l *LocalStorage) copyWithProgress(src string, dest string) bool {
	srcPath := filepath.Clean(src)
	destPath := filepath.Clean(dest)
	
	// 获取源文件信�?	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		logger.Errorf("【本地】获取源文件信息失败�?s", err.Error())
		return false
	}
	
	totalSize := srcInfo.Size()
	copiedSize := int64(0)
	
	// 初始化进度回�?	progressCallback := progress.NewProgressHelper(crypto.HashUtils.Md5(srcPath))
	progressCallback.Start()
	
	defer func() {
		progressCallback.Update(100, fmt.Sprintf("%s 进度�?00%%", srcPath))
		progressCallback.End()
	}()
	
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("【本地】复制文件异常：%v", r)
		}
	}()
	
	// 打开源文件和目标文件
	srcFile, err := os.Open(srcPath)
	if err != nil {
		logger.Errorf("【本地】打开源文件失败：%s", err.Error())
		return false
	}
	defer srcFile.Close()
	
	destFile, err := os.Create(destPath)
	if err != nil {
		logger.Errorf("【本地】创建目标文件失败：%s", err.Error())
		return false
	}
	defer destFile.Close()
	
	// 分块复制
	buffer := make([]byte, l.chunkSize)
	for {
		n, err := srcFile.Read(buffer)
		if n > 0 {
			_, writeErr := destFile.Write(buffer[:n])
			if writeErr != nil {
				logger.Errorf("【本地】写入目标文件失败：%s", writeErr.Error())
				return false
			}
			
			copiedSize += int64(n)
			
			// 更新进度
			if progressCallback != nil && totalSize > 0 {
				percent := float64(copiedSize*100) / float64(totalSize)
				progressCallback.Update(percent, fmt.Sprintf("%s 进度�?0.2f%%", srcPath, percent))
			}
		}
		
		if err != nil {
			if err != io.EOF {
				logger.Errorf("【本地】复制文件过程中发生错误�?s", err.Error())
				return false
			}
			break
		}
	}
	
	// 保留文件时间戳、权限等信息
	l.copyFileStats(srcPath, destPath)
	
	return true
}

// copyFileStats 复制文件属性（时间戳、权限等�?func (l *LocalStorage) copyFileStats(src string, dest string) {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return
	}
	
	// 复制权限
	err = os.Chmod(dest, srcInfo.Mode())
	if err != nil {
		logger.Debugf("【本地】复制文件权限失败：%s", err.Error())
	}
	
	// 复制时间�?	err = os.Chtimes(dest, srcInfo.ModTime(), srcInfo.ModTime())
	if err != nil {
		logger.Debugf("【本地】复制文件时间戳失败�?s", err.Error())
	}
}

// Upload 上传文件（带进度�?func (l *LocalStorage) Upload(fileItem *schemas.FileItem, path string, newName *string) *schemas.FileItem {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("【本地】上传文件异常：%v", r)
		}
	}()
	
	dirPath := filepath.Clean(fileItem.Path)
	targetName := filepath.Base(path)
	if newName != nil {
		targetName = *newName
	}
	
	targetPath := filepath.Join(dirPath, targetName)
	
	if l.copyWithProgress(path, targetPath) {
		// 上传成功后删除源文件
		err := os.Remove(path)
		if err != nil {
			logger.Warnf("【本地】删除源文件失败�?s", err.Error())
		}
		
		return l.GetItem(targetPath)
	}
	
	return nil
}

// shouldShowProgress 是否显示进度�?func (l *LocalStorage) shouldShowProgress(src string, dest string) bool {
	srcIsNetwork := system.IsNetworkFilesystem(src)
	destIsNetwork := system.IsNetworkFilesystem(dest)
	
	if srcIsNetwork && destIsNetwork && system.IsSameDisk(src, dest) {
		return true
	}
	
	return false
}

// Copy 复制文件（带进度�?func (l *LocalStorage) Copy(fileItem *schemas.FileItem, path string, newName string) bool {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("【本地】复制文件异常：%v", r)
		}
	}()
	
	src := filepath.Clean(fileItem.Path)
	dest := filepath.Join(path, newName)
	
	if l.shouldShowProgress(src, dest) {
		return l.copyWithProgress(src, dest)
	} else {
		code, message := system.Copy(src, dest)
		if code == 0 {
			return true
		} else {
			logger.Errorf("【本地】复制文件失败：%s", message)
		}
	}
	
	return false
}

// Move 移动文件（带进度�?func (l *LocalStorage) Move(fileItem *schemas.FileItem, path string, newName string) bool {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("【本地】移动文件异常：%v", r)
		}
	}()
	
	src := filepath.Clean(fileItem.Path)
	dest := filepath.Join(path, newName)
	
	// 目标和源文件相同，直接返回成功，不做任何操作
	if src == dest {
		return true
	}
	
	if l.shouldShowProgress(src, dest) {
		if l.copyWithProgress(src, dest) {
			// 复制成功删除源文�?			err := os.Remove(src)
			if err != nil {
				logger.Errorf("【本地】删除源文件失败�?s", err.Error())
				return false
			}
			return true
		}
	} else {
		code, message := system.Move(src, dest)
		if code == 0 {
			return true
		} else {
			logger.Errorf("【本地】移动文件失败：%s", message)
		}
	}
	
	return false
}

// Link 硬链接文�?func (l *LocalStorage) Link(fileItem *schemas.FileItem, targetFile string) bool {
	filePath := filepath.Clean(fileItem.Path)
	code, message := system.Link(filePath, targetFile)
	if code != 0 {
		logger.Errorf("【本地】硬链接文件失败�?s", message)
		return false
	}
	return true
}

// Softlink 软链接文�?func (l *LocalStorage) Softlink(fileItem *schemas.FileItem, targetFile string) bool {
	filePath := filepath.Clean(fileItem.Path)
	code, message := system.Softlink(filePath, targetFile)
	if code != 0 {
		logger.Errorf("【本地】软链接文件失败�?s", message)
		return false
	}
	return true
}

// Usage 存储使用情况
func (l *LocalStorage) Usage() *schemas.StorageUsage {
	directoryHelper := directory.NewDirectoryHelper()
	
	// 获取本地下载目录
	var paths []string
	localDownloadDirs := directoryHelper.GetLocalDownloadDirs()
	for _, dir := range localDownloadDirs {
		if dir.DownloadPath != "" {
			paths = append(paths, dir.DownloadPath)
		}
	}
	
	// 获取本地媒体库目�?	localLibraryDirs := directoryHelper.GetLocalLibraryDirs()
	for _, dir := range localLibraryDirs {
		if dir.LibraryPath != "" {
			paths = append(paths, dir.LibraryPath)
		}
	}
	
	totalStorage, freeStorage := system.SpaceUsage(paths)
	
	return &schemas.StorageUsage{
		Total:     totalStorage,
		Available: freeStorage,
	}
}
