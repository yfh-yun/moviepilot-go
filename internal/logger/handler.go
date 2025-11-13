package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	
	"gopkg.in/natefinch/lumberjack.v2"
)

// NonBlockingFileHandler 非阻塞文件处理器
type NonBlockingFileHandler struct {
	// 日志轮转处理器映�?	rotatingHandlers map[string]*lumberjack.Logger
	mutex            sync.RWMutex
	
	// 写入队列和相关控�?	writeQueue chan *LogEntry
	running    bool
	workers    int
	
	// 控制协程
	workerWg sync.WaitGroup
	quitChan chan struct{}
	
	// 批处理相�?	batchWriteSize int
	writeTimeout   float64
}

// nonBlockingFileHandlerInstance 非阻塞文件处理器单例
var nonBlockingFileHandlerInstance *NonBlockingFileHandler
var handlerOnce sync.Once

// GetNonBlockingFileHandler 获取非阻塞文件处理器单例
func GetNonBlockingFileHandler() *NonBlockingFileHandler {
	handlerOnce.Do(func() {
		settings := GetLogSettings()
		
		nonBlockingFileHandlerInstance = &NonBlockingFileHandler{
			rotatingHandlers: make(map[string]*lumberjack.Logger),
			writeQueue:       make(chan *LogEntry, settings.AsyncFileQueueSize),
			running:          true,
			workers:          settings.AsyncFileWorkers,
			quitChan:         make(chan struct{}),
			batchWriteSize:   settings.BatchWriteSize,
			writeTimeout:     settings.WriteTimeout,
		}
		
		// 启动后台写入工作�?		nonBlockingFileHandlerInstance.startWorkers()
	})
	
	return nonBlockingFileHandlerInstance
}

// startWorkers 启动后台写入工作�?func (h *NonBlockingFileHandler) startWorkers() {
	for i := 0; i < h.workers; i++ {
		h.workerWg.Add(1)
		go h.batchWriter()
	}
}

// getRotatingHandler 获取或创建日志轮转处理器实例
func (h *NonBlockingFileHandler) getRotatingHandler(filePath string) *lumberjack.Logger {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	// 检查是否已存在
	if handler, exists := h.rotatingHandlers[filePath]; exists {
		return handler
	}
	
	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("创建日志目录失败 %s: %v\n", dir, err)
	}
	
	settings := GetLogSettings()
	
	// 创建新的日志轮转处理�?	handler := &lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    settings.LogMaxFileSize, // MB
		MaxBackups: settings.LogBackupCount,
		MaxAge:     0, // 不根据时间轮�?		Compress:   false,
	}
	
	h.rotatingHandlers[filePath] = handler
	return handler
}

// WriteLog 写入日志
func (h *NonBlockingFileHandler) WriteLog(level LogLevel, message, filePath string) {
	entry := NewLogEntry(level, message, filePath, time.Now())
	
	// 尝试非阻塞写�?	select {
	case h.writeQueue <- entry:
		// 成功放入队列
	default:
		// 队列满时，使用同步方式直接写�?		h.writeSync(entry)
	}
}

// writeSync 同步写入日志
func (h *NonBlockingFileHandler) writeSync(entry *LogEntry) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("日志写入失败 %s: %v\n", entry.FilePath, r)
			fmt.Printf("�?s�?s - %s\n", entry.Level.Name(), entry.Timestamp.Format("2006/01/02 15:04:05"), entry.Message)
		}
	}()
	
	// 获取日志轮转处理器实�?	handler := h.getRotatingHandler(entry.FilePath)
	
	// 格式化日志消�?	settings := GetLogSettings()
	formattedMessage := fmt.Sprintf("�?s�?s - %s\n", 
		entry.Level.Name(), 
		entry.Timestamp.Format("2006/01/02 15:04:05"), 
		entry.Message)
	
	// 如果有自定义格式，使用自定义格式
	if settings.LogFileFormat != "" {
		formattedMessage = settings.LogFileFormat
		formattedMessage = strings.Replace(formattedMessage, "%(levelname)s", entry.Level.Name(), -1)
		formattedMessage = strings.Replace(formattedMessage, "%(asctime)s", entry.Timestamp.Format("2006/01/02 15:04:05"), -1)
		formattedMessage = strings.Replace(formattedMessage, "%(message)s", entry.Message, -1)
		formattedMessage = strings.Replace(formattedMessage, "%(leveltext)s", entry.Level.Name()+":", -1)
		formattedMessage = strings.Replace(formattedMessage, "%(name)s", filepath.Base(entry.FilePath), -1)
		formattedMessage += "\n"
	}
	
	// 写入日志
	if _, err := handler.Write([]byte(formattedMessage)); err != nil {
		fmt.Printf("日志写入失败 %s: %v\n", entry.FilePath, err)
		fmt.Printf("�?s�?s - %s\n", entry.Level.Name(), entry.Timestamp.Format("2006/01/02 15:04:05"), entry.Message)
	}
}

// batchWriter 后台批量写入工作�?func (h *NonBlockingFileHandler) batchWriter() {
	defer h.workerWg.Done()
	
	for {
		select {
		case <-h.quitChan:
			// 收到退出信�?			return
		default:
			// 收集一批日志条�?			batch := h.collectBatch()
			if len(batch) > 0 {
				h.writeBatch(batch)
			}
			// 短暂休眠避免过度占用CPU
			time.Sleep(10 * time.Millisecond)
		}
	}
}

// collectBatch 收集一批日志条�?func (h *NonBlockingFileHandler) collectBatch() []*LogEntry {
	var batch []*LogEntry
	endTime := time.Now().Add(time.Duration(h.writeTimeout * float64(time.Second)))
	
	for len(batch) < h.batchWriteSize && time.Now().Before(endTime) {
		select {
		case entry := <-h.writeQueue:
			batch = append(batch, entry)
		case <-time.After(time.Until(endTime)):
			// 超时
			return batch
		}
	}
	
	return batch
}

// writeBatch 批量写入日志
func (h *NonBlockingFileHandler) writeBatch(batch []*LogEntry) {
	// 按文件分�?	fileGroups := make(map[string][]*LogEntry)
	for _, entry := range batch {
		if _, exists := fileGroups[entry.FilePath]; !exists {
			fileGroups[entry.FilePath] = []*LogEntry{}
		}
		fileGroups[entry.FilePath] = append(fileGroups[entry.FilePath], entry)
	}
	
	// 批量写入每个文件
	for filePath, entries := range fileGroups {
		// 获取日志轮转处理�?		handler := h.getRotatingHandler(filePath)
		
		// 构建批量内容
		var content string
		for _, entry := range entries {
			settings := GetLogSettings()
			formattedMessage := fmt.Sprintf("�?s�?s - %s\n",
				entry.Level.Name(),
				entry.Timestamp.Format("2006/01/02 15:04:05"),
				entry.Message)
			
			// 如果有自定义格式，使用自定义格式
			if settings.LogFileFormat != "" {
				formattedMessage = settings.LogFileFormat
				formattedMessage = strings.Replace(formattedMessage, "%(levelname)s", entry.Level.Name(), -1)
				formattedMessage = strings.Replace(formattedMessage, "%(asctime)s", entry.Timestamp.Format("2006/01/02 15:04:05"), -1)
				formattedMessage = strings.Replace(formattedMessage, "%(message)s", entry.Message, -1)
				formattedMessage = strings.Replace(formattedMessage, "%(leveltext)s", entry.Level.Name()+":", -1)
				formattedMessage = strings.Replace(formattedMessage, "%(name)s", filepath.Base(entry.FilePath), -1)
				formattedMessage += "\n"
			}
			
			content += formattedMessage
		}
		
		// 批量写入
		if _, err := handler.Write([]byte(content)); err != nil {
			fmt.Printf("批量写入失败 %s: %v\n", filePath, err)
			// 回退到逐个写入
			for _, entry := range entries {
				h.writeSync(entry)
			}
		}
	}
}

// Shutdown 关闭文件处理�?func (h *NonBlockingFileHandler) Shutdown() {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	if !h.running {
		return
	}
	
	h.running = false
	
	// 发送退出信�?	close(h.quitChan)
	
	// 等待工作者完�?	h.workerWg.Wait()
	
	// 关闭所有日志轮转处理器
	for _, handler := range h.rotatingHandlers {
		handler.Close()
	}
	
	// 清理映射
	h.rotatingHandlers = make(map[string]*lumberjack.Logger)
}
