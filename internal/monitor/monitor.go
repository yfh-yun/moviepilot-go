package monitor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	
	"moviepilot-go/internal/logger"
)

// Monitor 目录监控处理结构
type Monitor struct {
	// 退出事�?	event *sync.WaitGroup
	
	// 监控服务
	observers []*fsnotify.Watcher
	
	// 定时服务
	scheduler *cron.Cron
	
	// 存储过照间隔（分钟）
	snapshotInterval int
	
	// TTL缓存�?0秒钟有效
	cache map[string]time.Time
	cacheMutex sync.RWMutex
	
	// 快照文件缓存
	snapshotCache map[string]*SnapshotData
	snapshotMutex sync.RWMutex
	
	// 监控的文件扩展名
	allExts []string
	
	// 日志记录�?	logger *zap.Logger
	
	// 是否运行�?	running bool
	
	// 事件处理函数
	eventHandler EventHandler
}

// SnapshotData 快照数据结构
type SnapshotData struct {
	Timestamp  float64                `json:"timestamp"`
	FileCount  int                    `json:"file_count"`
	Snapshot   map[string]interface{} `json:"snapshot"`
}

// NewMonitor 创建新的目录监控实例
func NewMonitor() *Monitor {
	// 获取日志记录�?	logManager := logger.GetLoggerManager()
	zapLogger := logManager.GetLogger("monitor")
	
	monitor := &Monitor{
		event:            &sync.WaitGroup{},
		observers:        make([]*fsnotify.Watcher, 0),
		scheduler:        cron.New(),
		snapshotInterval: 5,
		cache:            make(map[string]time.Time),
		snapshotCache:    make(map[string]*SnapshotData),
		allExts:          []string{".mp4", ".mkv", ".ts", ".iso", ".rmvb", ".avi", ".mov", ".mpeg", ".mpg", ".wmv", ".3gp", ".asf", ".m4v", ".flv", ".m2ts", ".strm", ".tp", ".f4v"},
		logger:           zapLogger,
		running:          true,
	}
	
	// 启动目录监控和文件整�?	monitor.Init()
	
	return monitor
}

// Init 启动监控
func (m *Monitor) Init() {
	// 停止现有任务
	m.Stop()
	
	// 读取目录配置 (简化实现，实际应从配置文件读取)
	monitorDirs := m.getMonitorDirs()
	if len(monitorDirs) == 0 {
		m.logger.Info("未找到任何目录监控配�?)
		return
	}
	
	// 按下载目录去�?	uniqueDirs := m.getUniqueDirs(monitorDirs)
	
	m.logger.Info(fmt.Sprintf("找到 %d 个目录监控配�?, len(uniqueDirs)))
	
	// 启动定时服务进程
	m.scheduler.Start()
	
	monStorages := make(map[string][]string)
	
	for _, monPath := range uniqueDirs {
		// 检查媒体库目录是不是下载目录的子目�?(简化实�?
		// targetPath := filepath.Join(monPath, "library")
		
		// 启动监控
		m.logger.Info(fmt.Sprintf("正在启动本地目录监控: %s", monPath))
		m.logger.Info("*** 重要提示：目录监控只处理新增和修改的文件，不会处理监控启动前已存在的文件 ***")
		
		// 本地目录监控
		m.startLocalMonitor(monPath)
		
		// 远程存储监控 (简化实�?
		storage := "local"
		if storage != "local" {
			if _, exists := monStorages[storage]; !exists {
				monStorages[storage] = make([]string, 0)
			}
			monStorages[storage] = append(monStorages[storage], monPath)
		}
	}
	
	// 启动远程存储监控
	for storage, paths := range monStorages {
		// 远程目录监控 - 使用智能间隔
		// 先尝试加载已有快照获取文件数�?		snapshotData := m.LoadSnapshot(storage)
		fileCount := 0
		if snapshotData != nil {
			fileCount = snapshotData.FileCount
		}
		interval := m.AdjustMonitorInterval(fileCount)
		
		for _, path := range paths {
			m.logger.Info(fmt.Sprintf("正在启动远程目录监控: %s [%s]", path, storage))
		}
		
		m.logger.Info("*** 重要提示：远程目录监控只处理新增和修改的文件，不会处理监控启动前已存在的文件 ***")
		m.logger.Info(fmt.Sprintf("预估文件数量: %d, 监控间隔: %d分钟", fileCount, interval))
		
		// 添加定时任务
		m.scheduler.AddFunc(fmt.Sprintf("@every %dm", interval), func() {
			m.PollingObserver(storage, paths)
		})
		
		m.logger.Info(fmt.Sprintf("�?远程目录监控已启�? [间隔: %d分钟]", interval))
	}
	
	// 启动定时服务
	if len(m.scheduler.Entries()) > 0 {
		m.logger.Info("定时监控服务已启�?)
	}
	
	// 输出监控总结
	m.logger.Info(fmt.Sprintf("目录监控启动完成: 本地监控 %d �?, len(uniqueDirs)))
}

// getMonitorDirs 获取监控目录列表 (简化实�?
func (m *Monitor) getMonitorDirs() []string {
	// 实际应该从配置文件读取监控目�?	// 这里应该读取配置，示例实�?
	// config := config.GetConfig()
	// monitorDirs := config.GetMonitorDirs()
	// return monitorDirs
	
	// 临时实现，应该从配置中读�?	return []string{filepath.Join(".", "downloads")}
}

// getUniqueDirs 按下载目录去�?func (m *Monitor) getUniqueDirs(monitorDirs []string) map[string]string {
	// 按下载目录去�?(简化实�?
	uniqueDirs := make(map[string]string)
	for _, dir := range monitorDirs {
		uniqueDirs[fmt.Sprintf("local_%s", dir)] = dir
	}
	return uniqueDirs
}

// startLocalMonitor 启动本地目录监控
func (m *Monitor) startLocalMonitor(monPath string) {
	// 创建文件监控�?	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		m.logger.Error(fmt.Sprintf("创建文件监控器失�? %s", err.Error()))
		return
	}
	
	// 添加监控路径
	err = watcher.Add(monPath)
	if err != nil {
		m.logger.Error(fmt.Sprintf("添加监控路径失败: %s", err.Error()))
		
		// 检查错误类型并提供相应建议
		if strings.Contains(err.Error(), "permission") {
			m.logger.Error("权限错误，请检�?MoviePilot 是否有足够的权限访问监控目录")
		} else if strings.Contains(err.Error(), "inotify") {
			m.logger.Error("inotify 相关错误，这通常是由于系统监控数量限制导致的")
			m.logger.Error("解决方案:")
			tips := m.GetSystemOptimizationTips()
			for _, tip := range tips {
				m.logger.Error(tip)
			}
			m.logger.Error("执行上述命令后重�?MoviePilot")
		} else {
			m.logger.Error("建议尝试使用兼容模式进行监控")
		}
		
		return
	}
	
	m.observers = append(m.observers, watcher)
	
	// 启动监控协程
	go m.watchDirectory(watcher, monPath)
	
	// 统计文件数量并给出提�?	fileCount := m.CountDirectoryFiles(monPath, 10000)
	m.logger.Info(fmt.Sprintf("监控目录 %s 包含�?%d 个文�?, monPath, fileCount))
	
	// 检查系统限�?	limits := m.CheckSystemLimits()
	usePolling, reason := m.ShouldUsePolling(monPath, "normal", fileCount, limits)
	m.logger.Info(fmt.Sprintf("监控模式决策: %s", reason))
	
	modeName := "快速模�?
	if usePolling {
		modeName = "兼容模式(轮询)"
	}
	
	m.logger.Info(fmt.Sprintf("�?本地目录监控已启�? %s [%s]", monPath, modeName))
}

// watchDirectory 监控目录变化
func (m *Monitor) watchDirectory(watcher *fsnotify.Watcher, monPath string) {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			
			// 处理文件事件
			if event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Write == fsnotify.Write {
				// 检查是否为目录
				fi, err := os.Stat(event.Name)
				if err != nil {
					continue
				}
				
				if !fi.IsDir() {
					// 文件发生变化
					m.logger.Debug(fmt.Sprintf("检测到文件变化: %s [创建/修改]", event.Name))
					// 整理文件
					m.HandleEvent(event, "创建/修改", event.Name, fi.Size())
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			m.logger.Error(fmt.Sprintf("监控错误: %s", err.Error()))
		}
	}
}

// EventHandler 事件处理函数类型
type EventHandler func(event fsnotify.Event, text string, eventPath string, fileSize int64)

// SetEventHandler 设置事件处理函数
func (m *Monitor) SetEventHandler(handler EventHandler) {
	m.eventHandler = handler
}

// eventHandler 事件处理函数
var eventHandler EventHandler

// HandleEvent 处理文件变化事件
func (m *Monitor) HandleEvent(event fsnotify.Event, text string, eventPath string, fileSize int64) {
	if eventHandler != nil {
		// 使用自定义事件处理函�?		eventHandler(event, text, eventPath, fileSize)
	} else {
		// 使用默认处理函数
		m.handleFile("local", eventPath, fileSize)
	}
}

// isBluraySub 判断是否蓝光原盘目录内的子目录或文件
func (m *Monitor) isBluraySub(path string) bool {
	// 判断是否蓝光原盘目录内的子目录或文件
	match, _ := regexp.MatchString(`BDMV[/\\]STREAM`, path)
	return match
}

// getBlurayDir 获取蓝光原盘BDMV目录的上级目�?func (m *Monitor) getBlurayDir(path string) string {
	// 获取蓝光原盘BDMV目录的上级目�?	p := path
	for {
		dir := filepath.Dir(p)
		if dir == p {
			// 到达根目�?			break
		}
		
		if filepath.Base(dir) == "BDMV" {
			return filepath.Dir(dir)
		}
		
		p = dir
	}
	return ""
}

// handleFile 整理一个文�?func (m *Monitor) handleFile(storage string, eventPath string, fileSize int64) {
	// 全程加锁
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()
	
	// 蓝光原盘文件处理
	if m.isBluraySub(eventPath) {
		eventPath = m.getBlurayDir(eventPath)
		if eventPath == "" {
			return
		}
	}
	
	// TTL缓存控重
	if timestamp, exists := m.cache[eventPath]; exists {
		// 检查是否在10秒内
		if time.Since(timestamp) < 10*time.Second {
			m.logger.Debug(fmt.Sprintf("文件 %s 在缓存中，跳过处�?, eventPath))
			return
		}
	}
	
	// 更新缓存
	m.cache[eventPath] = time.Now()
	
	// 处理文件 (简化实�?
	m.logger.Info(fmt.Sprintf("开始整理文�? %s", eventPath))
	
	// 实际应该调用文件转移服务处理文件
	// 这里简化实现，仅记录日�?	m.logger.Info(fmt.Sprintf("文件整理完成: %s", eventPath))
}

// Stop 退出监�?func (m *Monitor) Stop() {
	m.running = false
	
	// 停止所有监控器
	for _, observer := range m.observers {
		observer.Close()
	}
	m.observers = make([]*fsnotify.Watcher, 0)
	
	// 停止定时任务
	if m.scheduler != nil {
		m.scheduler.Stop()
	}
	
	m.logger.Info("目录监控服务已停�?)
}

// SaveSnapshot 保存快照到文件缓�?func (m *Monitor) SaveSnapshot(storage string, snapshot map[string]interface{}, fileCount int, lastSnapshotTime *float64) {
	m.snapshotMutex.Lock()
	defer m.snapshotMutex.Unlock()
	
	var snapshotTime float64
	// 获取快照中的最大时间戳
	for _, item := range snapshot {
		if itemMap, ok := item.(map[string]interface{}); ok {
			if modifyTime, exists := itemMap["modify_time"]; exists {
				if t, ok := modifyTime.(float64); ok && t > snapshotTime {
					snapshotTime = t
				}
			}
		}
	}
	
	if snapshotTime == 0 && lastSnapshotTime != nil {
		snapshotTime = *lastSnapshotTime
	} else if snapshotTime == 0 {
		snapshotTime = float64(time.Now().Unix())
	}
	
	snapshotData := &SnapshotData{
		Timestamp: snapshotTime,
		FileCount: fileCount,
		Snapshot:  snapshot,
	}
	
	cacheKey := fmt.Sprintf("%s_snapshot", storage)
	m.snapshotCache[cacheKey] = snapshotData
	
	// 持久化存储快照数�?	m.persistSnapshot(storage, snapshotData)
	
	m.logger.Debug(fmt.Sprintf("快照已保存到缓存: %s", storage))
}

// persistSnapshot 持久化快照数�?func (m *Monitor) persistSnapshot(storage string, snapshotData *SnapshotData) {
	// 创建快照目录
	snapshotDir := filepath.Join(".", "data", "snapshots")
	if err := os.MkdirAll(snapshotDir, 0755); err != nil {
		m.logger.Warn(fmt.Sprintf("创建快照目录失败: %s", err.Error()))
		return
	}
	
	// 保存快照文件
	snapshotFile := filepath.Join(snapshotDir, fmt.Sprintf("%s.json", storage))
	
	// 序列化数�?	data, err := json.MarshalIndent(snapshotData, "", "  ")
	if err != nil {
		m.logger.Warn(fmt.Sprintf("序列化快照数据失�? %s", err.Error()))
		return
	}
	
	// 写入文件
	if err := os.WriteFile(snapshotFile, data, 0644); err != nil {
		m.logger.Warn(fmt.Sprintf("保存快照文件失败: %s", err.Error()))
		return
	}
	
	m.logger.Debug(fmt.Sprintf("快照已持久化: %s", snapshotFile))
}

// LoadSnapshot 从文件缓存加载快�?func (m *Monitor) LoadSnapshot(storage string) *SnapshotData {
	m.snapshotMutex.RLock()
	defer m.snapshotMutex.RUnlock()
	
	cacheKey := fmt.Sprintf("%s_snapshot", storage)
	if snapshotData, exists := m.snapshotCache[cacheKey]; exists {
		m.logger.Debug(fmt.Sprintf("成功从缓存加载快�? %s, 包含 %d 个文�?, storage, len(snapshotData.Snapshot)))
		return snapshotData
	}
	
	// 从持久化存储中加�?	snapshotData := m.loadPersistedSnapshot(storage)
	if snapshotData != nil {
		// 放入缓存
		m.snapshotMutex.RUnlock()
		m.snapshotMutex.Lock()
		m.snapshotCache[cacheKey] = snapshotData
		m.snapshotMutex.Unlock()
		m.snapshotMutex.RLock()
		
		m.logger.Debug(fmt.Sprintf("成功从文件加载快�? %s, 包含 %d 个文�?, storage, len(snapshotData.Snapshot)))
		return snapshotData
	}
	
	m.logger.Debug(fmt.Sprintf("快照文件不存�? %s", storage))
	return nil
}

// loadPersistedSnapshot 从持久化存储加载快照
func (m *Monitor) loadPersistedSnapshot(storage string) *SnapshotData {
	snapshotFile := filepath.Join(".", "data", "snapshots", fmt.Sprintf("%s.json", storage))
	
	// 检查文件是否存�?	if _, err := os.Stat(snapshotFile); os.IsNotExist(err) {
		return nil
	}
	
	// 读取文件
	data, err := os.ReadFile(snapshotFile)
	if err != nil {
		m.logger.Warn(fmt.Sprintf("读取快照文件失败: %s", err.Error()))
		return nil
	}
	
	// 反序列化数据
	var snapshotData SnapshotData
	if err := json.Unmarshal(data, &snapshotData); err != nil {
		m.logger.Warn(fmt.Sprintf("反序列化快照数据失败: %s", err.Error()))
		return nil
	}
	
	return &snapshotData
}

// ResetSnapshot 重置快照，强制下次扫描时重新建立基准
func (m *Monitor) ResetSnapshot(storage string) bool {
	m.snapshotMutex.Lock()
	defer m.snapshotMutex.Unlock()
	
	cacheKey := fmt.Sprintf("%s_snapshot", storage)
	if _, exists := m.snapshotCache[cacheKey]; exists {
		delete(m.snapshotCache, cacheKey)
		m.logger.Info(fmt.Sprintf("快照已重�? %s", storage))
		return true
	}
	
	m.logger.Debug(fmt.Sprintf("快照文件不存在，无需重置: %s", storage))
	return true
}

// ForceFullScan 强制全量扫描并处理所有文件（包括已存在的文件�?func (m *Monitor) ForceFullScan(storage string, monPath string) bool {
	m.logger.Info(fmt.Sprintf("开始强制全量扫�? %s:%s", storage, monPath))
	
	// 生成快照 (简化实�?
	newSnapshot := m.snapshotStorage(storage, monPath, 0) // 全量扫描，不使用增量
	
	if newSnapshot == nil {
		m.logger.Warn(fmt.Sprintf("获取 %s:%s 快照失败", storage, monPath))
		return false
	}
	
	fileCount := len(newSnapshot)
	m.logger.Info(fmt.Sprintf("%s:%s 全量扫描完成，发�?%d 个文�?, storage, monPath, fileCount))
	
	// 处理所有文�?	processedCount := 0
	for filePath := range newSnapshot {
		m.logger.Info(fmt.Sprintf("处理文件�?s", filePath))
		
		// 获取文件大小
		var fileSize int64
		if fileInfo, ok := newSnapshot[filePath].(map[string]interface{}); ok {
			if size, exists := fileInfo["size"]; exists {
				if sizeFloat, ok := size.(float64); ok {
					fileSize = int64(sizeFloat)
				}
			}
		}
		
		m.handleFile(storage, filePath, fileSize)
		processedCount++
	}
	
	m.logger.Info(fmt.Sprintf("%s:%s 全量扫描完成，共处理 %d/%d 个文�?, storage, monPath, processedCount, fileCount))
	
	// 保存快照
	m.SaveSnapshot(storage, newSnapshot, fileCount, nil)
	
	return true
}

// snapshotStorage 生成存储快照 (简化实�?
func (m *Monitor) snapshotStorage(storage string, path string, lastSnapshotTime float64) map[string]interface{} {
	snapshot := make(map[string]interface{})
	
	// 遍历目录获取文件信息
	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// 跳过目录
		if info.IsDir() {
			return nil
		}
		
		// 检查文件扩展名
		ext := strings.ToLower(filepath.Ext(filePath))
		validExt := false
		for _, validExtItem := range m.allExts {
			if ext == validExtItem {
				validExt = true
				break
			}
		}
		
		if !validExt {
			return nil
		}
		
		// 添加到快�?		snapshot[filePath] = map[string]interface{}{
			"size":       info.Size(),
			"modify_time": float64(info.ModTime().Unix()),
		}
		
		return nil
	})
	
	if err != nil {
		m.logger.Error(fmt.Sprintf("生成快照失败: %s", err.Error()))
		return nil
	}
	
	return snapshot
}

// AdjustMonitorInterval 根据文件数量动态调整监控间�?func (m *Monitor) AdjustMonitorInterval(fileCount int) int {
	if fileCount < 100 {
		return 5 // 5分钟
	} else if fileCount < 500 {
		return 10 // 10分钟
	} else if fileCount < 1000 {
		return 15 // 15分钟
	} else {
		return 30 // 30分钟
	}
}

// CompareSnapshots 比对快照，找出变化的文件（只处理新增和修改，不处理删除）
func (m *Monitor) CompareSnapshots(oldSnapshot, newSnapshot map[string]interface{}) map[string][]string {
	changes := map[string][]string{
		"added":    make([]string, 0),
		"modified": make([]string, 0),
	}
	
	oldFiles := make(map[string]bool)
	newFiles := make(map[string]bool)
	
	// 构建文件集合
	for file := range oldSnapshot {
		oldFiles[file] = true
	}
	
	for file := range newSnapshot {
		newFiles[file] = true
	}
	
	// 新增文件
	for file := range newFiles {
		if !oldFiles[file] {
			changes["added"] = append(changes["added"], file)
		}
	}
	
	// 修改文件（大小或时间变化�?	for file := range oldFiles {
		if newFiles[file] {
			oldInfo := oldSnapshot[file]
			newInfo := newSnapshot[file]
			
			// 检查文件大小变�?			var oldSize, newSize float64
			var oldTime, newTime float64
			
			if oldInfoMap, ok := oldInfo.(map[string]interface{}); ok {
				if size, exists := oldInfoMap["size"]; exists {
					if sizeFloat, ok := size.(float64); ok {
						oldSize = sizeFloat
					}
				}
				if modifyTime, exists := oldInfoMap["modify_time"]; exists {
					if timeFloat, ok := modifyTime.(float64); ok {
						oldTime = timeFloat
					}
				}
			}
			
			if newInfoMap, ok := newInfo.(map[string]interface{}); ok {
				if size, exists := newInfoMap["size"]; exists {
					if sizeFloat, ok := size.(float64); ok {
						newSize = sizeFloat
					}
				}
				if modifyTime, exists := newInfoMap["modify_time"]; exists {
					if timeFloat, ok := modifyTime.(float64); ok {
						newTime = timeFloat
					}
				}
			}
			
			if oldSize != newSize || (oldTime > 0 && newTime > 0 && oldTime != newTime) {
				changes["modified"] = append(changes["modified"], file)
			}
		}
	}
	
	return changes
}

// CountDirectoryFiles 统计目录下的文件数量（用于检测是否超过系统限制）
func (m *Monitor) CountDirectoryFiles(directory string, maxCheck int) int {
	count := 0
	
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() {
			count++
		}
		
		// 限制检查数�?		if count > maxCheck {
			return filepath.SkipDir
		}
		
		return nil
	})
	
	if err != nil {
		m.logger.Debug(fmt.Sprintf("统计目录文件数量失败: %s", err.Error()))
		return 0
	}
	
	return count
}

// CheckSystemLimits 检查系统限�?func (m *Monitor) CheckSystemLimits() map[string]interface{} {
	limits := map[string]interface{}{
		"max_user_watches":     0,
		"max_user_instances":   0,
		"current_watches":      0,
		"warnings":             make([]string, 0),
	}
	
	system := runtime.GOOS
	if system == "linux" {
		// 检�?inotify 限制
		maxWatches, err := readIntFromFile("/proc/sys/fs/inotify/max_user_watches")
		if err != nil {
			m.logger.Debug(fmt.Sprintf("读取 inotify 限制失败: %s", err.Error()))
			limits["max_user_watches"] = 8192 // 默认�?		} else {
			limits["max_user_watches"] = maxWatches
		}
		
		maxInstances, err := readIntFromFile("/proc/sys/fs/inotify/max_user_instances")
		if err != nil {
			m.logger.Debug(fmt.Sprintf("读取 inotify 实例限制失败: %s", err.Error()))
		} else {
			limits["max_user_instances"] = maxInstances
		}
		
		// 检查当前使用的watches
		currentWatches, err := countInotifyInstances()
		if err != nil {
			m.logger.Debug(fmt.Sprintf("检查当�?inotify 使用失败: %s", err.Error()))
		} else {
			limits["current_watches"] = currentWatches
		}
	}
	
	return limits
}

// readIntFromFile 从文件中读取整数�?func readIntFromFile(filename string) (int, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return 0, err
	}
	
	var value int
	_, err = fmt.Sscanf(string(data), "%d", &value)
	if err != nil {
		return 0, err
	}
	
	return value, nil
}

// countInotifyInstances 统计当前 inotify 实例数量
func countInotifyInstances() (int, error) {
	// 这是一个简化的实现，实际应该统�?/proc/*/fd 中指�?inotify 的文件描述符
	// 由于在Go中实现较为复杂，这里返回一个估计�?	return 0, nil
}

// ShouldUsePolling 判断是否应该使用轮询模式
func (m *Monitor) ShouldUsePolling(directory string, monitorMode string, fileCount int, limits map[string]interface{}) (bool, string) {
	if monitorMode == "compatibility" {
		return true, "用户配置为兼容模�?
	}
	
	// 检查网络文件系�?(简化实�?
	// if SystemUtils.is_network_filesystem(directory) {
	//     return true, "检测到网络文件系统，建议使用兼容模�?
	// }
	
	if maxWatches, exists := limits["max_user_watches"]; exists {
		if maxWatchesFloat, ok := maxWatches.(float64); ok {
			if float64(fileCount) > maxWatchesFloat*0.8 {
				return true, fmt.Sprintf("目录文件数量(%d)接近系统限制(%.0f)", fileCount, maxWatchesFloat)
			}
		} else if maxWatchesInt, ok := maxWatches.(int); ok {
			if float64(fileCount) > float64(maxWatchesInt)*0.8 {
				return true, fmt.Sprintf("目录文件数量(%d)接近系统限制(%d)", fileCount, maxWatchesInt)
			}
		}
	}
	
	return false, "使用快速模�?
}

// GetSystemOptimizationTips 获取系统优化建议
func (m *Monitor) GetSystemOptimizationTips() []string {
	tips := make([]string, 0)
	system := runtime.GOOS
	
	if system == "linux" {
		tips = append(tips, []string{
			"增加 inotify 监控数量限制:",
			"echo fs.inotify.max_user_watches=524288 | sudo tee -a /etc/sysctl.conf",
			"echo fs.inotify.max_user_instances=524288 | sudo tee -a /etc/sysctl.conf",
			"sudo sysctl -p",
			"",
			"如果在Docker中运行，请在宿主机上执行以上命令",
		}...)
	} else if system == "darwin" {
		tips = append(tips, []string{
			"macOS 系统优化建议:",
			"sudo sysctl kern.maxfiles=65536",
			"sudo sysctl kern.maxfilesperproc=32768",
			"ulimit -n 32768",
		}...)
	} else if system == "windows" {
		tips = append(tips, []string{
			"Windows 系统优化建议:",
			"1. 关闭不必要的实时保护软件对监控目录的扫描",
			"2. 将监控目录添加到Windows Defender排除列表",
			"3. 确保有足够的可用内存",
		}...)
	}
	
	return tips
}

// PollingObserver 轮询监控（改进版�?func (m *Monitor) PollingObserver(storage string, monPaths []string) {
	m.snapshotMutex.Lock()
	defer m.snapshotMutex.Unlock()
	
	// 加载上次快照数据
	oldSnapshotData := m.LoadSnapshot(storage)
	var oldSnapshot map[string]interface{}
	var lastSnapshotTime float64
	
	if oldSnapshotData != nil {
		oldSnapshot = oldSnapshotData.Snapshot
		lastSnapshotTime = oldSnapshotData.Timestamp
	}
	
	// 判断是否为首次快照：检查快照文件是否存在且有效
	isFirstSnapshot := oldSnapshotData == nil
	newSnapshot := make(map[string]interface{})
	
	for _, monPath := range monPaths {
		m.logger.Debug(fmt.Sprintf("开始对 %s:%s 进行快照...", storage, monPath))
		
		// 生成新快照（增量模式�?		snapshot := m.snapshotStorage(storage, monPath, lastSnapshotTime)
		
		if snapshot == nil {
			m.logger.Warn(fmt.Sprintf("获取 %s:%s 快照失败", storage, monPath))
			continue
		}
		
		// 合并快照
		for k, v := range snapshot {
			newSnapshot[k] = v
		}
		
		fileCount := len(snapshot)
		m.logger.Info(fmt.Sprintf("%s:%s 快照完成，发�?%d 个文�?, storage, monPath, fileCount))
	}
	
	fileCount := len(newSnapshot)
	
	if !isFirstSnapshot {
		// 比较快照找出变化
		changes := m.CompareSnapshots(oldSnapshot, newSnapshot)
		
		// 处理新增文件
		for _, newFile := range changes["added"] {
			m.logger.Info(fmt.Sprintf("发现新增文件�?s", newFile))
			
			fileInfo, _ := newSnapshot[newFile].(map[string]interface{})
			var fileSize int64
			if size, exists := fileInfo["size"]; exists {
				if sizeFloat, ok := size.(float64); ok {
					fileSize = int64(sizeFloat)
				}
			}
			
			m.handleFile(storage, newFile, fileSize)
		}
		
		// 处理修改文件
		for _, modifiedFile := range changes["modified"] {
			m.logger.Info(fmt.Sprintf("发现修改文件�?s", modifiedFile))
			
			fileInfo, _ := newSnapshot[modifiedFile].(map[string]interface{})
			var fileSize int64
			if size, exists := fileInfo["size"]; exists {
				if sizeFloat, ok := size.(float64); ok {
					fileSize = int64(sizeFloat)
				}
			}
			
			m.handleFile(storage, modifiedFile, fileSize)
		}
		
		if len(changes["added"]) > 0 || len(changes["modified"]) > 0 {
			m.logger.Info(fmt.Sprintf("%s 发现 %d 个新增文件，%d 个修改文�?, storage, len(changes["added"]), len(changes["modified"])))
		} else {
			m.logger.Debug(fmt.Sprintf("%s 无文件变�?, storage))
		}
	} else {
		m.logger.Info(fmt.Sprintf("%s 首次快照完成，共 %d 个文�?, storage, fileCount))
		m.logger.Info("*** 首次快照仅建立基准，不会处理现有文件。后续监控将处理新增和修改的文件 ***")
	}
	
	// 保存新快�?	m.SaveSnapshot(storage, newSnapshot, fileCount, &lastSnapshotTime)
	
	// 动态调整监控间�?	newInterval := m.AdjustMonitorInterval(fileCount)
	// 实际应用中需要根据定时任务ID调整间隔，这里简化实�?	m.logger.Info(fmt.Sprintf("%s 监控间隔�?%d 分钟", storage, newInterval))
}
