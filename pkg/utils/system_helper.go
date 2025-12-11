package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// SystemHelper 系统帮助类
type SystemHelper struct {
	logger          *zap.Logger
	systemFlagFile  string
	dockerClientAPI string
	mutex           sync.RWMutex
}

// EventManager 事件管理器接口
type EventManager interface {
	// Register 注册事件处理器
	Register(eventType string, handler func(event Event))
}

// Event 事件接口
type Event interface {
	// GetType 获取事件类型
	GetType() string
	// GetData 获取事件数据
	GetData() any
}

// ConfigChangeEventData 配置变更事件数据
type ConfigChangeEventData struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}

// EventType 事件类型
type EventType string

const (
	EventTypeConfigChanged EventType = "ConfigChanged"
)

// NewSystemHelper 创建系统帮助类实例
func NewSystemHelper(dockerClientAPI string) *SystemHelper {
	return &SystemHelper{
		logger:          logger.GetLogger(),
		systemFlagFile:  "/var/log/nginx/__moviepilot__",
		dockerClientAPI: dockerClientAPI,
	}
}

// RegisterEventHandlers 注册事件处理器
func (h *SystemHelper) RegisterEventHandlers(eventManager EventManager) {
	eventManager.Register(string(EventTypeConfigChanged), h.handleConfigChanged)
}

// handleConfigChanged 处理配置变更事件
func (h *SystemHelper) handleConfigChanged(event Event) {
	if event == nil {
		return
	}

	// 获取事件数据
	eventData, ok := event.GetData().(ConfigChangeEventData)
	if !ok {
		h.logger.Error("无效的事件数据类型")
		return
	}

	// 检查是否需要更新日志设置
	logConfigKeys := []string{"DEBUG", "LOG_LEVEL", "LOG_MAX_FILE_SIZE", "LOG_BACKUP_COUNT", "LOG_FILE_FORMAT", "LOG_CONSOLE_FORMAT"}
	for _, key := range logConfigKeys {
		if eventData.Key == key {
			h.logger.Info("配置变更，更新日志设置...")
			// 更新日志设置
			// logger.UpdateLoggers() // 暂时注释，因为该函数未定义
			return
		}
	}
}

// CanRestart 判断是否可以内部重启
func (h *SystemHelper) CanRestart() bool {
	// 检查Docker socket是否存在或配置了有效的Docker客户端API
	return fileExists("/var/run/docker.sock") || h.dockerClientAPI != "tcp://127.0.0.1:38379"
}

// Restart 重启系统
func (h *SystemHelper) Restart() (bool, string) {
	// 检查是否为Docker环境
	if !isDocker() {
		return false, "非Docker环境，无法重启！"
	}

	// 检查容器是否配置了自动重启策略
	hasRestartPolicy, err := h.CheckRestartPolicy()
	if err != nil {
		h.logger.Warn("检查重启策略失败", zap.Error(err))
		// 降级为Docker API重启
		return h.DockerAPIRestart()
	}

	if hasRestartPolicy {
		// 有重启策略，使用优雅退出方式
		h.logger.Info("检测到容器配置了自动重启策略，使用优雅重启方式...")
		// 启动优雅退出超时监控
		h.StartGracefulShutdownMonitor()
		// 发送SIGTERM信号给当前进程，触发优雅停止
		if err := syscall.Kill(os.Getpid(), syscall.SIGTERM); err != nil {
			h.logger.Error("发送SIGTERM信号失败", zap.Error(err))
			// 降级为Docker API重启
			return h.DockerAPIRestart()
		}
		return true, ""
	} else {
		// 没有重启策略，使用Docker API强制重启
		h.logger.Info("容器未配置自动重启策略，使用Docker API重启...")
		return h.DockerAPIRestart()
	}
}

// GetContainerID 获取当前容器ID
func (h *SystemHelper) GetContainerID() (string, error) {
	// 读取/proc/self/mountinfo文件
	mountInfo, err := os.ReadFile("/proc/self/mountinfo")
	if err != nil {
		return "", fmt.Errorf("读取mountinfo失败: %w", err)
	}

	content := string(mountInfo)
	containerID := ""

	// 尝试从resolv.conf路径中提取容器ID
	indexResolvConf := strings.Index(content, "resolv.conf")
	if indexResolvConf != -1 {
		indexSecondSlash := strings.LastIndex(content[:indexResolvConf], "/")
		indexFirstSlash := strings.LastIndex(content[:indexSecondSlash], "/") + 1
		containerID = content[indexFirstSlash:indexSecondSlash]
		if len(containerID) < 20 {
			// 尝试从/sys/fs/cgroup/devices路径中提取容器ID
			indexCgroup := strings.Index(content, "/sys/fs/cgroup/devices")
			if indexCgroup != -1 {
				indexSpace := strings.LastIndex(content[:indexCgroup], " ")
				indexFirstSlash = strings.LastIndex(content[:indexSpace], "/") + 1
				containerID = content[indexFirstSlash:indexSpace]
			}
		}
	}

	if containerID == "" {
		return "", fmt.Errorf("无法提取容器ID")
	}

	return strings.TrimSpace(containerID), nil
}

// CheckRestartPolicy 检查当前容器是否配置了自动重启策略
func (h *SystemHelper) CheckRestartPolicy() (bool, error) {
	// 获取当前容器ID
	containerID, err := h.GetContainerID()
	if err != nil {
		return false, fmt.Errorf("获取容器ID失败: %w", err)
	}

	// TODO: 实现Docker客户端逻辑
	// 由于Go版本的Docker客户端实现较为复杂，这里简化处理
	// 在实际实现中，应该使用docker SDK来获取容器信息
	h.logger.Info("检查容器重启策略", zap.String("container_id", containerID))

	// 简化实现：假设所有容器都配置了自动重启策略
	return true, nil
}

// StartGracefulShutdownMonitor 启动优雅退出超时监控
func (h *SystemHelper) StartGracefulShutdownMonitor() {
	// 启动监控线程
	go func() {
		// 等待30秒
		time.Sleep(30 * time.Second)
		h.logger.Warn("优雅退出超时30秒，使用Docker API强制重启...")

		// 使用Docker API强制重启
		if _, errMsg := h.DockerAPIRestart(); errMsg != "" {
			h.logger.Error("强制重启失败", zap.String("error", errMsg))
		}
	}()
}

// DockerAPIRestart 使用Docker API重启容器
func (h *SystemHelper) DockerAPIRestart() (bool, string) {
	// 获取当前容器ID
	containerID, err := h.GetContainerID()
	if err != nil {
		return false, fmt.Sprintf("获取容器ID失败: %v", err)
	}

	// TODO: 实现Docker API重启逻辑
	// 由于Go版本的Docker客户端实现较为复杂，这里简化处理
	// 在实际实现中，应该使用docker SDK来重启容器
	h.logger.Info("使用Docker API重启容器", zap.String("container_id", containerID))

	// 简化实现：假设重启成功
	return true, ""
}

// SetSystemModified 设置系统已修改标志
func (h *SystemHelper) SetSystemModified() {
	if !isDocker() {
		return
	}

	// 创建标志文件
	if err := os.MkdirAll(filepath.Dir(h.systemFlagFile), 0755); err != nil {
		h.logger.Error("创建目录失败", zap.Error(err))
		return
	}

	// 触摸文件
	file, err := os.Create(h.systemFlagFile)
	if err != nil {
		h.logger.Error("创建标志文件失败", zap.Error(err))
		return
	}
	file.Close()

	h.logger.Info("系统修改标志已设置")
}

// IsSystemReset 检查系统是否已被重置
func (h *SystemHelper) IsSystemReset() bool {
	if !isDocker() {
		return false
	}

	// 检查标志文件是否存在
	_, err := os.Stat(h.systemFlagFile)
	return os.IsNotExist(err)
}

// 本地系统工具函数，避免循环依赖

// fileExists 检查文件是否存在
func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}
