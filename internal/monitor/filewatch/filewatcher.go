package filewatch

import (
	"context"
	"errors"
	"sync"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// FileWatcher 文件监控器接口
type FileWatcher interface {
	Watch(path string, handler EventHandler) error
	Stop() error
}

// fileWatcher 文件监控器实现
type fileWatcher struct {
	watcher  *fsnotify.Watcher
	handlers map[string]EventHandler
	mu       sync.RWMutex
	logger   *zap.Logger
	ctx      context.Context
	cancel   context.CancelFunc
	running  bool
}

// NewFileWatcher 创建文件监控器
func NewFileWatcher() (FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	fw := &fileWatcher{
		watcher:  watcher,
		handlers: make(map[string]EventHandler),
		logger:   logger.GetLogger(),
		ctx:      ctx,
		cancel:   cancel,
		running:  false,
	}

	// 启动事件处理协程
	go fw.run()

	return fw, nil
}

// Watch 监控指定路径
func (fw *fileWatcher) Watch(path string, handler EventHandler) error {
	if err := fw.watcher.Add(path); err != nil {
		return err
	}

	fw.mu.Lock()
	fw.handlers[path] = handler
	fw.mu.Unlock()

	fw.logger.Info("started watching path", zap.String("path", path))
	return nil
}

// Stop 停止监控
func (fw *fileWatcher) Stop() error {
	if !fw.running {
		return errors.New("watcher already stopped")
	}

	fw.cancel()
	return fw.watcher.Close()
}

// run 事件处理循环
func (fw *fileWatcher) run() {
	fw.running = true
	defer func() {
		fw.running = false
	}()

	for {
		select {
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}
			fw.handleEvent(event)
		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			fw.logger.Error("file watcher error", zap.Error(err))
		case <-fw.ctx.Done():
			return
		}
	}
}

// handleEvent 处理文件系统事件
func (fw *fileWatcher) handleEvent(event fsnotify.Event) {
	// 转换fsnotify事件为自定义Event
	var op Operation
	switch {
	case event.Op&fsnotify.Create != 0:
		op = Create
	case event.Op&fsnotify.Write != 0:
		op = Write
	case event.Op&fsnotify.Remove != 0:
		op = Remove
	case event.Op&fsnotify.Rename != 0:
		op = Rename
	default:
		return
	}

	customEvent := Event{
		Path: event.Name,
		Op:   op,
	}

	// 调用对应的事件处理器
	fw.mu.RLock()
	handler, exists := fw.handlers[event.Name]
	fw.mu.RUnlock()

	if exists {
		handler(customEvent)
	}

	fw.logger.Debug("file system event",
		zap.String("path", event.Name),
		zap.String("operation", op.String()))
}
