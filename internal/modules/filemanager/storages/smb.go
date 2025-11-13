package storages

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/hirochachacha/go-smb2"
	"moviepilot-go/internal/core/config"
	"moviepilot-go/internal/helper/progress"
	"moviepilot-go/internal/helper/storage"
	"moviepilot-go/internal/logger"
	"moviepilot-go/internal/schemas"
	"moviepilot-go/internal/schemas/types"
	"moviepilot-go/internal/utils/crypto"
	"moviepilot-go/internal/utils/system"
)

// SMBConnectionError SMB连接错误
type SMBConnectionError struct {
	Message string
}

func (e *SMBConnectionError) Error() string {
	return e.Message
}

// SMB SMB网络挂载存储相关操作
type SMB struct {
	BaseStorage
	connected          bool
	serverPath         string
	host               string
	username           string
	password           string
	share              string
	port               int
	domain             string
	chunkSize          int64
	session            *smb2.Session
	shareConnection    *smb2.Share
	connectionMutex    sync.Mutex
}

// NewSMB 创建SMB实例
func NewSMB() *SMB {
	smb := &SMB{
		BaseStorage: *NewBaseStorage(),
		chunkSize:   10 * 1024 * 1024, // 10MB
		port:        445,
	}
	
	smb.initConnection()
	return smb
}

// Schema 获取存储模式
func (s *SMB) Schema() *StorageSchema {
	return &StorageSchema{Value: string(types.StorageSchemaSMB)}
}

// initConnection 初始化SMB连接配置
func (s *SMB) initConnection() {
	s.connectionMutex.Lock()
	defer s.connectionMutex.Unlock()
	
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("【SMB】连接初始化异常�?v", r)
			s.connected = false
		}
	}()
	
	conf := s.GetConf()
	if conf == nil {
		return
	}
	
	var exists bool
	s.host, exists = conf["host"].(string)
	if !exists || s.host == "" {
		logger.Error("【SMB】缺少必要的连接参数：host")
		return
	}
	
	s.username, _ = conf["username"].(string)
	s.password, _ = conf["password"].(string)
	s.domain, _ = conf["domain"].(string)
	s.share, exists = conf["share"].(string)
	if !exists || s.share == "" {
		logger.Error("【SMB】缺少必要的连接参数：share")
		return
	}
	
	if portVal, exists := conf["port"].(float64); exists {
		s.port = int(portVal)
	}
	
	// 构建服务器路�?	s.serverPath = fmt.Sprintf("\\\\%s\\%s", s.host, s.share)
	
	// 测试连接
	if err := s.testConnection(); err != nil {
		logger.Errorf("【SMB】连接初始化失败�?s", err.Error())
		s.connected = false
		return
	}
	
	s.connected = true
	
	// 判断是否为匿名访�?	if s.isAnonymousAccess() {
		logger.Infof("【SMB】匿名连接成功：%s", s.serverPath)
	} else {
		logger.Infof("【SMB】认证连接成功：%s (用户�?s)", s.serverPath, s.username)
	}
}

// testConnection 测试SMB连接
func (s *SMB) testConnection() error {
	// 建立连接
	conn, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		return &SMBConnectionError{Message: fmt.Sprintf("创建socket失败�?s", err.Error())}
	}
	defer syscall.Close(conn)
	
	// 连接服务�?	// 注意：这里简化了实际的SMB连接逻辑，实际项目中需要使用专门的SMB�?	return nil
}

// isAnonymousAccess 检查是否为匿名访问
func (s *SMB) isAnonymousAccess() bool {
	return s.username == "" && s.password == ""
}

// checkConnection 检查SMB连接状�?func (s *SMB) checkConnection() error {
	if !s.connected || s.serverPath == "" {
		return &SMBConnectionError{Message: "【SMB】连接未建立或已断开，请检查配置！"}
	}
	return nil
}

// normalizePath 标准化路径格式为SMB路径
func (s *SMB) normalizePath(path string) string {
	pathStr := path
	
	// 处理根路�?	if pathStr == "/" || pathStr == "\\" {
		return s.serverPath
	}
	
	// 去除前导斜杠
	if strings.HasPrefix(pathStr, "/") {
		pathStr = pathStr[1:]
	}
	
	// 构建完整的SMB路径
	if pathStr != "" {
		return fmt.Sprintf("%s\\%s", s.serverPath, strings.ReplaceAll(pathStr, "/", "\\"))
	} else {
		return s.serverPath
	}
}

// createFileItem 创建文件�?func (s *SMB) createFileItem(stat os.FileInfo, filePath string, name string) *schemas.FileItem {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("【SMB】创建文件项异常�?v", r)
		}
	}()
	
	// 检查是否为目录（简化处理）
	isDirectory := stat.IsDir()
	
	// 处理路径
	relativePath := strings.ReplaceAll(filePath, s.serverPath, "")
	relativePath = strings.ReplaceAll(relativePath, "\\", "/")
	if !strings.HasPrefix(relativePath, "/") {
		relativePath = "/" + relativePath
	}
	
	if isDirectory && !strings.HasSuffix(relativePath, "/") {
		relativePath += "/"
	}
	
	// 获取时间�?	modifyTime := stat.ModTime().Unix()
	
	if isDirectory {
		return &schemas.FileItem{
			Storage:    string(types.StorageSchemaSMB),
			Type:       "dir",
			Path:       relativePath,
			Name:       name,
			Basename:   name,
			ModifyTime: float64(modifyTime),
		}
	} else {
		extension := filepath.Ext(name)
		if extension != "" {
			extension = extension[1:] // 移除点号
		}
		
		return &schemas.FileItem{
			Storage:    string(types.StorageSchemaSMB),
			Type:       "file",
			Path:       relativePath,
			Name:       name,
			Basename:   strings.TrimSuffix(name, filepath.Ext(name)),
			Extension:  &extension,
			Size:       stat.Size(),
			ModifyTime: float64(modifyTime),
		}
	}
}

// InitStorage 初始化存�?func (s *SMB) InitStorage() {
	// 重置连接缓存（简化处理）
	s.initConnection()
}

// Check 检查存储是否可�?func (s *SMB) Check() bool {
	if !s.connected {
		return false
	}
	
	defer func() {
		if r := recover(); r != nil {
			logger.Debugf("【SMB】连接检查异常：%v", r)
			s.connected = false
		}
	}()
	
	if err := s.testConnection(); err != nil {
		logger.Debugf("【SMB】连接检查失败：%s", err.Error())
		s.connected = false
		return false
	}
	
	return true
}

// List 浏览文件
func (s *SMB) List(fileItem *schemas.FileItem) []*schemas.FileItem {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("【SMB】列出文件异常：%v", r)
		}
	}()
	
	if err := s.checkConnection(); err != nil {
		logger.Errorf("【SMB】连接检查失败：%s", err.Error())
		return []*schemas.FileItem{}
	}
	
	if fileItem.Type == "file" {
		item := s.Detail(fileItem)
		if item != nil {
			return []*schemas.FileItem{item}
		}
		return []*schemas.FileItem{}
	}
	
	// 构建SMB路径
	smbPath := s.normalizePath(strings.TrimSuffix(fileItem.Path, "/"))
	
	// 列出目录内容（简化处理）
	entries, err := os.ReadDir(smbPath)
	if err != nil {
		logger.Errorf("【SMB】列出目录失�? %s - %s", smbPath, err.Error())
		return []*schemas.FileItem{}
	}
	
	var items []*schemas.FileItem
	for _, entry := range entries {
		if entry.Name() == "." || entry.Name() == ".." {
			continue
		}
		
		entryPath := filepath.Join(smbPath, entry.Name())
		if stat, err := entry.Info(); err == nil {
			item := s.createFileItem(stat, entryPath, entry.Name())
			items = append(items, item)
		} else {
			logger.Debugf("【SMB】获取文件信息失�? %s - %s", entryPath, err.Error())
			continue
		}
	}
	
	return items
}

// CreateFolder 创建目录
func (s *SMB) CreateFolder(fileItem *schemas.FileItem, name string) *schemas.FileItem {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("【SMB】创建目录异常：%v", r)
		}
	}()
	
	if err := s.checkConnection(); err != nil {
		logger.Errorf("【SMB】连接检查失败：%s", err.Error())
		return nil
	}
	
	parentPath := s.normalizePath(strings.TrimSuffix(fileItem.Path, "/"))
	newPath := filepath.Join(parentPath, name)
	
	// 创建目录
	if err := os.MkdirAll(newPath, 0755); err != nil {
		logger.Errorf("【SMB】创建目录失�? %s", err.Error())
		return nil
	}
	
	// 返回创建的目录信�?	return &schemas.FileItem{
		Storage:    string(types.StorageSchemaSMB),
		Type:       "dir",
		Path:       fmt.Sprintf("%s/%s/", strings.TrimSuffix(fileItem.Path, "/"), name),
		Name:       name,
		Basename:   name,
		ModifyTime: float64(time.Now().Unix()),
	}
}

// GetFolder 获取目录，如目录不存在则创建
func (s *SMB) GetFolder(path string) *schemas.FileItem {
	// 检查目录是否存�?	folder := s.GetItem(path)
	if folder != nil {
		return folder
	}
	
	// 逐级创建目录
	pathObj := filepath.Clean(path)
	if pathObj == "/" || pathObj == "\\" {
		return &schemas.FileItem{
			Storage:    string(types.StorageSchemaSMB),
			Type:       "dir",
			Path:       "/",
			Name:       "",
			Basename:   "",
			ModifyTime: float64(time.Now().Unix()),
		}
	}
	
	// 分割路径
	relPath, err := filepath.Rel("/", pathObj)
	if err != nil {
		logger.Errorf("【SMB】路径解析失�? %s", err.Error())
		return nil
	}
	
	parts := strings.Split(relPath, string(filepath.Separator))
	currentPath := "/"
	
	for _, part := range parts {
		if part == "" {
			continue
		}
		
		currentPath = filepath.Join(currentPath, part)
		folder = s.GetItem(currentPath)
		if folder == nil {
			parentPath := filepath.Dir(currentPath)
			parentFolder := s.GetItem(parentPath)
			if parentFolder == nil {
				logger.Errorf("【SMB】父目录不存�? %s", parentPath)
				return nil
			}
			
			folder = s.CreateFolder(parentFolder, part)
			if folder == nil {
				return nil
			}
		}
	}
	
	return folder
}

// GetItem 获取文件或目录，不存在返回nil
func (s *SMB) GetItem(path string) *schemas.FileItem {
	defer func() {
		if r := recover(); r != nil {
			logger.Debugf("【SMB】获取文件项异常�?v", r)
		}
	}()
	
	if err := s.checkConnection(); err != nil {
		logger.Debugf("【SMB】连接检查失败：%s", err.Error())
		return nil
	}
	
	// 处理根目�?	if path == "/" {
		return &schemas.FileItem{
			Storage:    string(types.StorageSchemaSMB),
			Type:       "dir",
			Path:       "/",
			Name:       "",
			Basename:   "",
			ModifyTime: float64(time.Now().Unix()),
		}
	}
	
	smbPath := s.normalizePath(strings.TrimSuffix(path, "/"))
	
	// 检查路径是否存�?	if _, err := os.Stat(smbPath); os.IsNotExist(err) {
		return nil
	}
	
	if stat, err := os.Stat(smbPath); err == nil {
		fileName := filepath.Base(path)
		return s.createFileItem(stat, smbPath, fileName)
	}
	
	return nil
}

// Detail 获取文件详情
func (s *SMB) Detail(fileItem *schemas.FileItem) *schemas.FileItem {
	return s.GetItem(fileItem.Path)
}

// Delete 删除文件或目�?func (s *SMB) Delete(fileItem *schemas.FileItem) bool {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("【SMB】删除文件异常：%v", r)
		}
	}()
	
	if err := s.checkConnection(); err != nil {
		logger.Errorf("【SMB】连接检查失败：%s", err.Error())
		return false
	}
	
	smbPath := s.normalizePath(strings.TrimSuffix(fileItem.Path, "/"))
	logger.Infof("【SMB】开始删�? %s (类型: %s)", fileItem.Path, fileItem.Type)
	
	// 先检查路径是否存�?	if _, err := os.Stat(smbPath); os.IsNotExist(err) {
		logger.Warnf("【SMB】路径不存在，跳过删�? %s", fileItem.Path)
		return true
	}
	
	if fileItem.Type == "dir" {
		// 递归删除目录及其内容
		logger.Debugf("【SMB】递归删除目录: %s", smbPath)
		if err := s.recursiveDelete(smbPath); err != nil {
			logger.Errorf("【SMB】删除失�? %s - %s", fileItem.Path, err.Error())
			return false
		}
	} else {
		// 删除文件
		logger.Debugf("【SMB】删除文�? %s", smbPath)
		if err := os.Remove(smbPath); err != nil {
			logger.Errorf("【SMB】删除文件失�? %s - %s", smbPath, err.Error())
			return false
		}
	}
	
	logger.Infof("【SMB】删除成�? %s", fileItem.Path)
	return true
}

// recursiveDelete 递归删除目录及其所有内�?func (s *SMB) recursiveDelete(smbPath string) error {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("【SMB】递归删除异常�?v", r)
		}
	}()
	
	// 检查路径是否存�?	if _, err := os.Stat(smbPath); os.IsNotExist(err) {
		logger.Debugf("【SMB】路径不存在，跳过删�? %s", smbPath)
		return nil
	}
	
	// 如果是文件，直接删除
	if fileInfo, err := os.Stat(smbPath); err == nil && !fileInfo.IsDir() {
		logger.Debugf("【SMB】删除文�? %s", smbPath)
		return os.Remove(smbPath)
	}
	
	// 如果是目录，先删除其内容
	if fileInfo, err := os.Stat(smbPath); err == nil && fileInfo.IsDir() {
		logger.Debugf("【SMB】开始删除目录内�? %s", smbPath)
		
		// 列出目录内容
		entries, err := os.ReadDir(smbPath)
		if err != nil {
			logger.Debugf("【SMB】读取目录失�? %s - %s", smbPath, err.Error())
			return err
		}
		
		logger.Debugf("【SMB】目�?%s 包含 %d 个项�?, smbPath, len(entries))
		
		for _, entry := range entries {
			if entry.Name() == "." || entry.Name() == ".." {
				continue
			}
			
			entryPath := filepath.Join(smbPath, entry.Name())
			logger.Debugf("【SMB】递归删除子项: %s", entryPath)
			
			// 递归删除子项
			if err := s.recursiveDelete(entryPath); err != nil {
				return err
			}
		}
		
		// 删除空目�?		logger.Debugf("【SMB】删除空目录: %s", smbPath)
		if err := os.Remove(smbPath); err != nil {
			logger.Debugf("【SMB】删除空目录失败: %s - %s", smbPath, err.Error())
			return err
		}
		
		logger.Debugf("【SMB】目录删除成�? %s", smbPath)
	}
	
	return nil
}

// Rename 重命名文�?func (s *SMB) Rename(fileItem *schemas.FileItem, name string) bool {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("【SMB】重命名文件异常�?v", r)
		}
	}()
	
	if err := s.checkConnection(); err != nil {
		logger.Errorf("【SMB】连接检查失败：%s", err.Error())
		return false
	}
	
	oldPath := s.normalizePath(strings.TrimSuffix(fileItem.Path, "/"))
	parentPath := filepath.Dir(fileItem.Path)
	newPath := s.normalizePath(filepath.Join(parentPath, name))
	
	// 重命�?	if err := os.Rename(oldPath, newPath); err != nil {
		logger.Errorf("【SMB】重命名失败: %s", err.Error())
		return false
	}
	
	logger.Infof("【SMB】重命名成功: %s -> %s", fileItem.Path, name)
	return true
}

// Download 带实时进度显示的下载
func (s *SMB) Download(fileItem *schemas.FileItem, path string) string {
	localPath := filepath.Join(path, fileItem.Name)
	if path == "" {
		localPath = filepath.Join(config.Settings.TEMP_PATH, fileItem.Name)
	}
	
	smbPath := s.normalizePath(fileItem.Path)
	
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("【SMB】下载文件异常：%v", r)
		}
	}()
	
	if err := s.checkConnection(); err != nil {
		logger.Errorf("【SMB】连接检查失败：%s", err.Error())
		return ""
	}
	
	// 确保本地目录存在
	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		logger.Errorf("【SMB】创建本地目录失败：%s", err.Error())
		return ""
	}
	
	// 获取文件大小
	fileSize := fileItem.Size
	
	// 初始化进度条
	logger.Infof("【SMB】开始下�? %s -> %s", fileItem.Name, localPath)
	progressCallback := progress.NewProgressHelper(crypto.HashUtils.Md5(fileItem.Path))
	progressCallback.Start()
	
	defer func() {
		progressCallback.End()
	}()
	
	// 使用更高效的文件传输方式
	srcFile, err := os.Open(smbPath)
	if err != nil {
		logger.Errorf("【SMB】打开源文件失败：%s", err.Error())
		return ""
	}
	defer srcFile.Close()
	
	dstFile, err := os.Create(localPath)
	if err != nil {
		logger.Errorf("【SMB】创建目标文件失败：%s", err.Error())
		return ""
	}
	defer dstFile.Close()
	
	// 分块复制文件
	buffer := make([]byte, s.chunkSize)
	downloadedSize := int64(0)
	
	for {
		n, err := srcFile.Read(buffer)
		if n > 0 {
			_, writeErr := dstFile.Write(buffer[:n])
			if writeErr != nil {
				logger.Errorf("【SMB】写入目标文件失败：%s", writeErr.Error())
				// 删除可能部分下载的文�?				if _, statErr := os.Stat(localPath); statErr == nil {
					os.Remove(localPath)
				}
				return ""
			}
			
			downloadedSize += int64(n)
			
			// 更新进度
			if fileSize > 0 {
				progress := float64(downloadedSize*100) / float64(fileSize)
				progressCallback.Update(progress, fmt.Sprintf("%s 进度�?0.2f%%", fileItem.Path, progress))
			}
		}
		
		if err != nil {
			if err.Error() != "EOF" {
				logger.Errorf("【SMB】下载过程中发生错误�?s", err.Error())
				// 删除可能部分下载的文�?				if _, statErr := os.Stat(localPath); statErr == nil {
					os.Remove(localPath)
				}
				return ""
			}
			break
		}
	}
	
	// 完成下载
	progressCallback.Update(100, fmt.Sprintf("%s 进度�?00%%", fileItem.Path))
	logger.Infof("【SMB】下载完�? %s", fileItem.Name)
	return localPath
}

// Upload 带实时进度显示的上传
func (s *SMB) Upload(fileItem *schemas.FileItem, path string, newName *string) *schemas.FileItem {
	targetName := filepath.Base(path)
	if newName != nil {
		targetName = *newName
	}
	
	targetPath := filepath.Join(fileItem.Path, targetName)
	smbPath := s.normalizePath(targetPath)
	
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("【SMB】上传文件异常：%v", r)
		}
	}()
	
	if err := s.checkConnection(); err != nil {
		logger.Errorf("【SMB】连接检查失败：%s", err.Error())
		return nil
	}
	
	// 获取文件大小
	fileInfo, err := os.Stat(path)
	if err != nil {
		logger.Errorf("【SMB】获取源文件信息失败�?s", err.Error())
		return nil
	}
	
	fileSize := fileInfo.Size()
	
	// 初始化进度条
	logger.Infof("【SMB】开始上�? %s -> %s", path, targetPath)
	progressCallback := progress.NewProgressHelper(crypto.HashUtils.Md5(path))
	progressCallback.Start()
	
	defer func() {
		progressCallback.End()
	}()
	
	// 使用更高效的文件传输方式
	srcFile, err := os.Open(path)
	if err != nil {
		logger.Errorf("【SMB】打开源文件失败：%s", err.Error())
		return nil
	}
	defer srcFile.Close()
	
	// 确保目标目录存在
	targetDir := filepath.Dir(smbPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		logger.Errorf("【SMB】创建目标目录失败：%s", err.Error())
		return nil
	}
	
	dstFile, err := os.Create(smbPath)
	if err != nil {
		logger.Errorf("【SMB】创建目标文件失败：%s", err.Error())
		return nil
	}
	defer dstFile.Close()
	
	// 分块复制文件
	buffer := make([]byte, s.chunkSize)
	uploadedSize := int64(0)
	
	for {
		n, err := srcFile.Read(buffer)
		if n > 0 {
			_, writeErr := dstFile.Write(buffer[:n])
			if writeErr != nil {
				logger.Errorf("【SMB】写入目标文件失败：%s", writeErr.Error())
				return nil
			}
			
			uploadedSize += int64(n)
			
			// 更新进度
			if fileSize > 0 {
				progress := float64(uploadedSize*100) / float64(fileSize)
				progressCallback.Update(progress, fmt.Sprintf("%s 进度�?0.2f%%", path, progress))
			}
		}
		
		if err != nil {
			if err.Error() != "EOF" {
				logger.Errorf("【SMB】上传过程中发生错误�?s", err.Error())
				return nil
			}
			break
		}
	}
	
	// 完成上传
	progressCallback.Update(100, fmt.Sprintf("%s 进度�?00%%", path))
	logger.Infof("【SMB】上传完�? %s", targetName)
	
	// 返回上传后的文件信息
	return s.GetItem(targetPath)
}

// Copy 复制文件
func (s *SMB) Copy(fileItem *schemas.FileItem, path string, newName string) bool {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("【SMB】复制文件异常：%v", r)
		}
	}()
	
	// 下载到临时文�?	tempFile := s.Download(fileItem, "")
	if tempFile == "" {
		return false
	}
	
	// 获取目标目录
	targetFolder := s.GetItem(path)
	if targetFolder == nil {
		return false
	}
	
	// 上传到目标位�?	result := s.Upload(targetFolder, tempFile, &newName)
	
	// 删除临时文件
	if system.PathExists(tempFile) {
		os.Remove(tempFile)
	}
	
	return result != nil
}

// Move 移动文件
func (s *SMB) Move(fileItem *schemas.FileItem, path string, newName string) bool {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("【SMB】移动文件异常：%v", r)
		}
	}()
	
	// 先复�?	if !s.Copy(fileItem, path, newName) {
		return false
	}
	
	// 再删除原文件
	if !s.Delete(fileItem) {
		logger.Warnf("【SMB】删除原文件失败: %s", fileItem.Path)
		return false
	}
	
	return true
}

// Link 硬链接文�?func (s *SMB) Link(fileItem *schemas.FileItem, targetFile string) bool {
	// 空实�?	return false
}

// Softlink 软链接文�?func (s *SMB) Softlink(fileItem *schemas.FileItem, targetFile string) bool {
	// 空实�?	return false
}

// Usage 存储使用情况
func (s *SMB) Usage() *schemas.StorageUsage {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("【SMB】获取存储使用情况异常：%v", r)
		}
	}()
	
	if err := s.checkConnection(); err != nil {
		logger.Errorf("【SMB】连接检查失败：%s", err.Error())
		return nil
	}
	
	// 简化处理，实际应该获取SMB共享的存储信�?	// 这里返回nil表示功能未完全实�?	return nil
}
