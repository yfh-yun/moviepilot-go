package storages

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"moviepilot-go/internal/core/config"
	"moviepilot-go/internal/helper/progress"
	"moviepilot-go/internal/helper/storage"
	"moviepilot-go/internal/logger"
	"moviepilot-go/internal/schemas"
	"moviepilot-go/internal/schemas/types"
	"moviepilot-go/internal/utils/crypto"
	"moviepilot-go/internal/utils/stringutils"
	"moviepilot-go/internal/utils/system"
)

// Rclone rclone相关操作
type Rclone struct {
	BaseStorage
}

// NewRclone 创建rclone实例
func NewRclone() *Rclone {
	rclone := &Rclone{
		BaseStorage: *NewBaseStorage(),
	}
	
	// 设置快照检查目录修改时间标�?	rclone.snapshotCheckFolderModtime = config.Settings.RCLONE_SNAPSHOT_CHECK_FOLDER_MODTIME
	
	return rclone
}

// Schema 获取存储模式
func (r *Rclone) Schema() *StorageSchema {
	return &StorageSchema{Value: string(types.StorageSchemaRclone)}
}

// InitStorage 初始�?func (r *Rclone) InitStorage() {
	// 空实�?}

// SetConfig 设置配置
func (r *Rclone) SetConfig(conf map[string]interface{}) {
	// 调用父类方法
	r.BaseStorage.SetConfig(conf)
	
	filepathVal, exists := conf["filepath"].(string)
	if !exists || filepathVal == "" {
		logger.Warn("【rclone】保存配置失败：未设置配置文件路�?)
		return
	}
	
	logger.Infof("【rclone】配置写入文件：%s", filepathVal)
	
	path := filepath.Clean(filepathVal)
	dir := filepath.Dir(path)
	
	// 创建目录
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			logger.Errorf("【rclone】创建配置目录失败：%s", err.Error())
			return
		}
	}
	
	// 写入配置内容
	content, contentExists := conf["content"].(string)
	if contentExists && content != "" {
		err := os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			logger.Errorf("【rclone】写入配置文件失败：%s", err.Error())
		}
	}
}

// getHiddenShell 获取隐藏shell
func (r *Rclone) getHiddenShell() *syscall.SysProcAttr {
	if runtime.GOOS == "windows" {
		// Windows系统隐藏窗口
		return &syscall.SysProcAttr{
			HideWindow: true,
		}
	}
	// 其他系统返回nil
	return nil
}

// parseRcloneProgress 解析rclone进度输出
func (r *Rclone) parseRcloneProgress(line string) *float64 {
	if line == "" {
		return nil
	}
	
	line = strings.TrimSpace(line)
	
	// 检查是否包含百分比
	if !strings.Contains(line, "%") {
		return nil
	}
	
	defer func() {
		if recover() != nil {
			// 忽略解析错误
		}
	}()
	
	tryParse := func(str string) *float64 {
		// 尝试解析浮点�?		if val, err := strconv.ParseFloat(str, 64); err == nil {
			return &val
		}
		return nil
	}
	
	// 尝试多种进度输出格式
	if strings.Contains(line, "ETA") {
		// 格式: "Transferred: 1.234M / 5.678M, 22%, 1.234MB/s, ETA 2m3s"
		parts := strings.Split(line, "%")
		if len(parts) > 0 {
			percentPart := strings.TrimSpace(parts[0])
			percentStrs := strings.Fields(percentPart)
			if len(percentStrs) > 0 {
				percentStr := percentStrs[len(percentStrs)-1]
				return tryParse(percentStr)
			}
		}
	} else if strings.Contains(line, "Transferred:") && strings.Contains(line, "100%") {
		// 传输完成
		val := 100.0
		return &val
	} else {
		// 其他包含百分比的格式
		parts := strings.Fields(line)
		for _, part := range parts {
			if strings.Contains(part, "%") {
				percentStr := strings.ReplaceAll(part, "%", "")
				return tryParse(percentStr)
			}
		}
	}
	
	return nil
}

// getRcloneItem 获取rclone文件�?func (r *Rclone) getRcloneItem(item map[string]interface{}, parent string) *schemas.FileItem {
	if item == nil {
		return &schemas.FileItem{}
	}
	
	isDir, _ := item["IsDir"].(bool)
	name, _ := item["Name"].(string)
	modTime, _ := item["ModTime"].(string)
	
	if isDir {
		return &schemas.FileItem{
			Storage:    string(types.StorageSchemaRclone),
			Type:       "dir",
			Path:       filepath.Join(parent, name) + "/",
			Name:       name,
			Basename:   name,
			ModifyTime: stringutils.StrToTimestamp(modTime),
		}
	} else {
		size, _ := item["Size"].(float64)
		extension := filepath.Ext(name)
		if extension != "" {
			extension = extension[1:] // 移除点号
		}
		
		return &schemas.FileItem{
			Storage:    string(types.StorageSchemaRclone),
			Type:       "file",
			Path:       filepath.Join(parent, name),
			Name:       name,
			Basename:   strings.TrimSuffix(name, filepath.Ext(name)),
			Extension:  &extension,
			Size:       int64(size),
			ModifyTime: stringutils.StrToTimestamp(modTime),
		}
	}
}

// Check 检查存储是否可�?func (r *Rclone) Check() bool {
	defer func() {
		if recover() != nil {
			logger.Error("【rclone】存储检查异�?)
		}
	}()
	
	cmd := exec.Command("rclone", "lsf", "MP:")
	cmd.SysProcAttr = r.getHiddenShell()
	
	err := cmd.Run()
	if err != nil {
		logger.Errorf("【rclone】存储检查失败：%s", err.Error())
		return false
	}
	
	return true
}

// List 浏览文件
func (r *Rclone) List(fileItem *schemas.FileItem) []*schemas.FileItem {
	if fileItem.Type == "file" {
		return []*schemas.FileItem{fileItem}
	}
	
	defer func() {
		if recover() != nil {
			logger.Error("【rclone】浏览文件异�?)
		}
	}()
	
	cmd := exec.Command("rclone", "lsjson", fmt.Sprintf("MP:%s", fileItem.Path))
	cmd.SysProcAttr = r.getHiddenShell()
	
	output, err := cmd.Output()
	if err != nil {
		logger.Errorf("【rclone】浏览文件失败：%s", err.Error())
		return []*schemas.FileItem{}
	}
	
	var items []map[string]interface{}
	if err := json.Unmarshal(output, &items); err != nil {
		logger.Errorf("【rclone】解析文件列表失败：%s", err.Error())
		return []*schemas.FileItem{}
	}
	
	var result []*schemas.FileItem
	for _, item := range items {
		result = append(result, r.getRcloneItem(item, fileItem.Path))
	}
	
	return result
}

// CreateFolder 创建目录
func (r *Rclone) CreateFolder(fileItem *schemas.FileItem, name string) *schemas.FileItem {
	defer func() {
		if recover() != nil {
			logger.Error("【rclone】创建目录异�?)
		}
	}()
	
	path := filepath.Join(fileItem.Path, name)
	cmd := exec.Command("rclone", "mkdir", fmt.Sprintf("MP:%s", path))
	cmd.SysProcAttr = r.getHiddenShell()
	
	err := cmd.Run()
	if err != nil {
		logger.Errorf("【rclone】创建目录失败：%s", err.Error())
		return nil
	}
	
	return r.GetItem(path)
}

// findDir 查找下级目录中匹配名称的目录
func (r *Rclone) findDir(fileItem *schemas.FileItem, name string) *schemas.FileItem {
	// 查找下级目录中匹配名称的目录
	for _, subFolder := range r.List(fileItem) {
		if subFolder.Type != "dir" {
			continue
		}
		if subFolder.Name == name {
			return subFolder
		}
	}
	return nil
}

// GetFolder 根据文件路程获取目录，不存在则创�?func (r *Rclone) GetFolder(path string) *schemas.FileItem {
	// 是否已存�?	folder := r.GetItem(path)
	if folder != nil {
		return folder
	}
	
	// 逐级查找和创建目�?	fileItem := &schemas.FileItem{
		Storage: string(types.StorageSchemaRclone),
		Path:    "/",
	}
	
	// 分割路径
	relPath, err := filepath.Rel("/", path)
	if err != nil {
		logger.Warnf("【rclone】路径解析失�? %s", err.Error())
		return nil
	}
	
	parts := strings.Split(relPath, string(filepath.Separator))
	for _, part := range parts {
		if part == "" {
			continue
		}
		
		dirFile := r.findDir(fileItem, part)
		if dirFile != nil {
			fileItem = dirFile
		} else {
			dirFile = r.CreateFolder(fileItem, part)
			if dirFile == nil {
				logger.Warnf("【rclone】创建目�?%s%s 失败�?, fileItem.Path, part)
				return nil
			}
			fileItem = dirFile
		}
	}
	
	return fileItem
}

// GetItem 获取文件或目录，不存在返回nil
func (r *Rclone) GetItem(path string) *schemas.FileItem {
	defer func() {
		if recover() != nil {
			logger.Debugf("【rclone】获取文件项异常")
		}
	}()
	
	parent := filepath.Dir(path)
	name := filepath.Base(path)
	
	cmd := exec.Command("rclone", "lsjson", fmt.Sprintf("MP:%s", parent))
	cmd.SysProcAttr = r.getHiddenShell()
	
	output, err := cmd.Output()
	if err != nil {
		logger.Debugf("【rclone】获取文件项失败�?s", err.Error())
		return nil
	}
	
	var items []map[string]interface{}
	if err := json.Unmarshal(output, &items); err != nil {
		logger.Debugf("【rclone】解析文件项失败�?s", err.Error())
		return nil
	}
	
	for _, item := range items {
		itemName, _ := item["Name"].(string)
		if itemName == name {
			return r.getRcloneItem(item, parent+"/")
		}
	}
	
	return nil
}

// Delete 删除文件
func (r *Rclone) Delete(fileItem *schemas.FileItem) bool {
	defer func() {
		if recover() != nil {
			logger.Error("【rclone】删除文件异�?)
		}
	}()
	
	cmd := exec.Command("rclone", "deletefile", fmt.Sprintf("MP:%s", fileItem.Path))
	cmd.SysProcAttr = r.getHiddenShell()
	
	err := cmd.Run()
	if err != nil {
		logger.Errorf("【rclone】删除文件失败：%s", err.Error())
		return false
	}
	
	return true
}

// Rename 重命名文�?func (r *Rclone) Rename(fileItem *schemas.FileItem, name string) bool {
	defer func() {
		if recover() != nil {
			logger.Error("【rclone】重命名文件异常")
		}
	}()
	
	parent := filepath.Dir(fileItem.Path)
	newPath := filepath.Join(parent, name)
	
	cmd := exec.Command("rclone", "moveto", fmt.Sprintf("MP:%s", fileItem.Path), fmt.Sprintf("MP:%s", newPath))
	cmd.SysProcAttr = r.getHiddenShell()
	
	err := cmd.Run()
	if err != nil {
		logger.Errorf("【rclone】重命名文件失败�?s", err.Error())
		return false
	}
	
	return true
}

// Download 带实时进度显示的下载
func (r *Rclone) Download(fileItem *schemas.FileItem, path string) string {
	localPath := filepath.Join(path, fileItem.Name)
	if path == "" {
		localPath = filepath.Join(config.Settings.TEMP_PATH, fileItem.Name)
	}
	
	// 初始化进度条
	logger.Infof("【rclone】开始下�? %s -> %s", fileItem.Name, localPath)
	progressCallback := progress.NewProgressHelper(crypto.HashUtils.Md5(fileItem.Path))
	progressCallback.Start()
	
	defer func() {
		progressCallback.End()
		if recover() != nil {
			logger.Errorf("【rclone】下载文件异�?)
		}
	}()
	
	// 使用rclone的进度显示功�?	cmd := exec.Command("rclone", "copyto", "--progress", "--stats", "1s", fmt.Sprintf("MP:%s", fileItem.Path), localPath)
	cmd.SysProcAttr = r.getHiddenShell()
	
	// 获取命令的stdout管道
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logger.Errorf("【rclone】创建管道失败：%s", err.Error())
		return ""
	}
	
	// 合并stderr到stdout
	cmd.Stderr = cmd.Stdout
	
	// 启动命令
	if err := cmd.Start(); err != nil {
		logger.Errorf("【rclone】启动下载命令失败：%s", err.Error())
		return ""
	}
	
	// 监控进度输出
	lastProgress := 0.0
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			// 解析rclone的进度输�?			progress := r.parseRcloneProgress(line)
			if progress != nil && *progress > lastProgress {
				progressCallback.Update(*progress, fmt.Sprintf("%s 进度�?0.2f%%", fileItem.Path, *progress))
				lastProgress = *progress
				if *progress >= 100 {
					break
				}
			}
		}
	}
	
	// 等待进程完成
	if err := cmd.Wait(); err != nil {
		logger.Errorf("【rclone】下载失�? %s - %s", fileItem.Name, err.Error())
		// 删除可能部分下载的文�?		if _, statErr := os.Stat(localPath); statErr == nil {
			os.Remove(localPath)
		}
		return ""
	}
	
	logger.Infof("【rclone】下载完�? %s", fileItem.Name)
	return localPath
}

// Upload 带实时进度显示的上传
func (r *Rclone) Upload(fileItem *schemas.FileItem, path string, newName *string) *schemas.FileItem {
	targetName := filepath.Base(path)
	if newName != nil {
		targetName = *newName
	}
	
	newPath := filepath.Join(fileItem.Path, targetName)
	
	// 初始化进度条
	logger.Infof("【rclone】开始上�? %s -> %s", path, newPath)
	progressCallback := progress.NewProgressHelper(crypto.HashUtils.Md5(path))
	progressCallback.Start()
	
	defer func() {
		progressCallback.End()
		if recover() != nil {
			logger.Errorf("【rclone】上传文件异�?)
		}
	}()
	
	// 使用rclone的进度显示功�?	cmd := exec.Command("rclone", "copyto", "--progress", "--stats", "1s", path, fmt.Sprintf("MP:%s", newPath))
	cmd.SysProcAttr = r.getHiddenShell()
	
	// 获取命令的stdout管道
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logger.Errorf("【rclone】创建管道失败：%s", err.Error())
		return nil
	}
	
	// 合并stderr到stdout
	cmd.Stderr = cmd.Stdout
	
	// 启动命令
	if err := cmd.Start(); err != nil {
		logger.Errorf("【rclone】启动上传命令失败：%s", err.Error())
		return nil
	}
	
	// 监控进度输出
	lastProgress := 0.0
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			// 解析rclone的进度输�?			progress := r.parseRcloneProgress(line)
			if progress != nil && *progress > lastProgress {
				progressCallback.Update(*progress, fmt.Sprintf("%s 进度�?0.2f%%", path, *progress))
				lastProgress = *progress
				if *progress >= 100 {
					break
				}
			}
		}
	}
	
	// 等待进程完成
	if err := cmd.Wait(); err != nil {
		logger.Errorf("【rclone】上传失�? %s - %s", targetName, err.Error())
		return nil
	}
	
	logger.Infof("【rclone】上传完�? %s", targetName)
	return r.GetItem(newPath)
}

// Detail 获取文件详情
func (r *Rclone) Detail(fileItem *schemas.FileItem) *schemas.FileItem {
	defer func() {
		if recover() != nil {
			logger.Error("【rclone】获取文件详情异�?)
		}
	}()
	
	cmd := exec.Command("rclone", "lsjson", fmt.Sprintf("MP:%s", fileItem.Path))
	cmd.SysProcAttr = r.getHiddenShell()
	
	output, err := cmd.Output()
	if err != nil {
		logger.Errorf("【rclone】获取文件详情失败：%s", err.Error())
		return nil
	}
	
	var items []map[string]interface{}
	if err := json.Unmarshal(output, &items); err != nil {
		logger.Errorf("【rclone】解析文件详情失败：%s", err.Error())
		return nil
	}
	
	if len(items) > 0 {
		return r.getRcloneItem(items[0], "")
	}
	
	return nil
}

// Move 移动文件
func (r *Rclone) Move(fileItem *schemas.FileItem, path string, newName string) bool {
	targetPath := filepath.Join(path, newName)
	
	// 初始化进度条
	logger.Infof("【rclone】开始移�? %s -> %s", fileItem.Path, targetPath)
	progressCallback := progress.NewProgressHelper(crypto.HashUtils.Md5(fileItem.Path))
	progressCallback.Start()
	
	defer func() {
		progressCallback.End()
		if recover() != nil {
			logger.Errorf("【rclone】移动文件异�?)
		}
	}()
	
	// 使用rclone的进度显示功�?	cmd := exec.Command("rclone", "moveto", "--progress", "--stats", "1s", fmt.Sprintf("MP:%s", fileItem.Path), fmt.Sprintf("MP:%s", targetPath))
	cmd.SysProcAttr = r.getHiddenShell()
	
	// 获取命令的stdout管道
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logger.Errorf("【rclone】创建管道失败：%s", err.Error())
		return false
	}
	
	// 合并stderr到stdout
	cmd.Stderr = cmd.Stdout
	
	// 启动命令
	if err := cmd.Start(); err != nil {
		logger.Errorf("【rclone】启动移动命令失败：%s", err.Error())
		return false
	}
	
	// 监控进度输出
	lastProgress := 0.0
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			// 解析rclone的进度输�?			progress := r.parseRcloneProgress(line)
			if progress != nil && *progress > lastProgress {
				progressCallback.Update(*progress, fmt.Sprintf("%s 进度�?0.2f%%", fileItem.Path, *progress))
				lastProgress = *progress
				if *progress >= 100 {
					break
				}
			}
		}
	}
	
	// 等待进程完成
	if err := cmd.Wait(); err != nil {
		logger.Errorf("【rclone】移动失�? %s - %s", fileItem.Name, err.Error())
		return false
	}
	
	logger.Infof("【rclone】移动完�? %s", fileItem.Name)
	return true
}

// Copy 复制文件
func (r *Rclone) Copy(fileItem *schemas.FileItem, path string, newName string) bool {
	targetPath := filepath.Join(path, newName)
	
	// 初始化进度条
	logger.Infof("【rclone】开始复�? %s -> %s", fileItem.Path, targetPath)
	progressCallback := progress.NewProgressHelper(crypto.HashUtils.Md5(fileItem.Path))
	progressCallback.Start()
	
	defer func() {
		progressCallback.End()
		if recover() != nil {
			logger.Errorf("【rclone】复制文件异�?)
		}
	}()
	
	// 使用rclone的进度显示功�?	cmd := exec.Command("rclone", "copyto", "--progress", "--stats", "1s", fmt.Sprintf("MP:%s", fileItem.Path), fmt.Sprintf("MP:%s", targetPath))
	cmd.SysProcAttr = r.getHiddenShell()
	
	// 获取命令的stdout管道
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logger.Errorf("【rclone】创建管道失败：%s", err.Error())
		return false
	}
	
	// 合并stderr到stdout
	cmd.Stderr = cmd.Stdout
	
	// 启动命令
	if err := cmd.Start(); err != nil {
		logger.Errorf("【rclone】启动复制命令失败：%s", err.Error())
		return false
	}
	
	// 监控进度输出
	lastProgress := 0.0
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			// 解析rclone的进度输�?			progress := r.parseRcloneProgress(line)
			if progress != nil && *progress > lastProgress {
				progressCallback.Update(*progress, fmt.Sprintf("%s 进度�?0.2f%%", fileItem.Path, *progress))
				lastProgress = *progress
				if *progress >= 100 {
					break
				}
			}
		}
	}
	
	// 等待进程完成
	if err := cmd.Wait(); err != nil {
		logger.Errorf("【rclone】复制失�? %s - %s", fileItem.Name, err.Error())
		return false
	}
	
	logger.Infof("【rclone】复制完�? %s", fileItem.Name)
	return true
}

// Link 硬链接文�?func (r *Rclone) Link(fileItem *schemas.FileItem, targetFile string) bool {
	// 空实�?	return false
}

// Softlink 软链接文�?func (r *Rclone) Softlink(fileItem *schemas.FileItem, targetFile string) bool {
	// 空实�?	return false
}

// Usage 存储使用情况
func (r *Rclone) Usage() *schemas.StorageUsage {
	conf := r.GetConfig()
	if conf == nil {
		return nil
	}
	
	filePath, exists := conf.Config["filepath"].(string)
	if !exists || filePath == "" || !system.PathExists(filePath) {
		return nil
	}
	
	// 读取rclone文件，检查是否有[MP]节点配置
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}
	
	lines := strings.Split(string(content), "\n")
	if len(lines) == 0 {
		return nil
	}
	
	hasMP := false
	for _, line := range lines {
		if strings.Contains(strings.TrimSpace(line), "[MP]") {
			hasMP = true
			break
		}
	}
	
	if !hasMP {
		return nil
	}
	
	defer func() {
		if recover() != nil {
			logger.Error("【rclone】获取存储使用情况异�?)
		}
	}()
	
	cmd := exec.Command("rclone", "about", "MP:/", "--json")
	cmd.SysProcAttr = r.getHiddenShell()
	
	output, err := cmd.Output()
	if err != nil {
		logger.Errorf("【rclone】获取存储使用情况失败：%s", err.Error())
		return nil
	}
	
	var items map[string]interface{}
	if err := json.Unmarshal(output, &items); err != nil {
		logger.Errorf("【rclone】解析存储使用情况失败：%s", err.Error())
		return nil
	}
	
	total, _ := items["total"].(float64)
	free, _ := items["free"].(float64)
	
	return &schemas.StorageUsage{
		Total:     int64(total),
		Available: int64(free),
	}
}
