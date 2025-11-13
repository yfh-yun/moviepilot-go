package sandbox

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"moviepilot-go/internal/config"
	"moviepilot-go/pkg/plugins"
)

// PluginService 插件服务，通过 HTTP/JSON 动态加载并执行用户插件
type PluginService struct {
	// 插件容器映射
	pluginContainers map[string]*PluginContainer
}

// PluginContainer 插件容器信息
type PluginContainer struct {
	ID          string    `json:"id"`
	PluginPath  string    `json:"plugin_path"`
	ContainerID string    `json:"container_id"`
	Status      string    `json:"status"`
	StartTime   time.Time `json:"start_time"`
}

// PluginRequest 插件请求结构
type PluginRequest struct {
	PluginPath string                 `json:"plugin_path" binding:"required"`
	Action     string                 `json:"action" binding:"required"`
	Params     map[string]interface{} `json:"params,omitempty"`
}

// PluginResponse 插件响应结构
type PluginResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// NewPluginService 创建一个新的插件服务实�?func NewPluginService() *PluginService {
	return &PluginService{
		pluginContainers: make(map[string]*PluginContainer),
	}
}

// LoadPlugin 加载插件到独立的Docker容器�?func (ps *PluginService) LoadPlugin(pluginPath string) (*PluginContainer, error) {
	// 检查插件路径是否存�?	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("plugin path does not exist: %s", pluginPath)
	}
	
	// 生成唯一ID
	containerID := uuid.New().String()
	
	// 构建Docker运行命令
	// 注意：这里假设系统已经安装了Docker，并且Docker守护进程正在运行
	cmd := exec.Command("docker", "run", "-d", 
		"--name", fmt.Sprintf("moviepilot-plugin-%s", containerID),
		"-v", fmt.Sprintf("%s:/app/plugin", pluginPath),
		"moviepilot/python-plugin-runtime:latest",
		"python", "/app/plugin/plugin.py")
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to start plugin container: %v, output: %s", err, string(output))
	}
	
	container := &PluginContainer{
		ID:          containerID,
		PluginPath:  pluginPath,
		ContainerID: string(output),
		Status:      "running",
		StartTime:   time.Now(),
	}
	
	// 保存容器信息
	ps.pluginContainers[containerID] = container
	
	return container, nil
}

// ExecutePlugin 执行插件中的特定操作
func (ps *PluginService) ExecutePlugin(containerID, action string, params map[string]interface{}) (*PluginResponse, error) {
	// 检查容器是否存�?	container, exists := ps.pluginContainers[containerID]
	if !exists {
		return nil, fmt.Errorf("plugin container not found: %s", containerID)
	}
	
	// 构建执行命令
	// 这里我们假设插件容器暴露了一个HTTP API端点来接收命�?	// 实际实现中可能需要根据具体需求调�?	
	// 构造请求数�?	requestData := map[string]interface{}{
		"action": action,
		"params": params,
	}
	
	// 序列化请求数�?	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request data: %v", err)
	}
	
	// 发送HTTP请求到插件容�?	// 注意：这里假设插件容器在本地运行并监听特定端�?	url := fmt.Sprintf("http://localhost:8080/execute") // 假设的端�?	resp, err := http.Post(url, "application/json", 
		fmt.Sprintf("%s", jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to execute plugin: %v", err)
	}
	defer resp.Body.Close()
	
	// 解析响应
	var pluginResp PluginResponse
	if err := json.NewDecoder(resp.Body).Decode(&pluginResp); err != nil {
		return nil, fmt.Errorf("failed to decode plugin response: %v", err)
	}
	
	return &pluginResp, nil
}

// UnloadPlugin 卸载插件容器
func (ps *PluginService) UnloadPlugin(containerID string) error {
	// 检查容器是否存�?	container, exists := ps.pluginContainers[containerID]
	if !exists {
		return fmt.Errorf("plugin container not found: %s", containerID)
	}
	
	// 停止并删除Docker容器
	cmd := exec.Command("docker", "rm", "-f", fmt.Sprintf("moviepilot-plugin-%s", container.ID))
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to remove plugin container: %v, output: %s", err, string(output))
	}
	
	// 从映射中删除
	delete(ps.pluginContainers, containerID)
	
	return nil
}

// ListPlugins 列出所有加载的插件
func (ps *PluginService) ListPlugins() []*PluginContainer {
	containers := make([]*PluginContainer, 0, len(ps.pluginContainers))
	for _, container := range ps.pluginContainers {
		containers = append(containers, container)
	}
	return containers
}

// RegisterRoutes 注册插件服务的HTTP路由
func (ps *PluginService) RegisterRoutes(router *gin.Engine) {
	pluginGroup := router.Group("/api/v1/plugin")
	{
		// 加载插件
		pluginGroup.POST("/load", ps.handleLoadPlugin)
		
		// 执行插件
		pluginGroup.POST("/execute", ps.handleExecutePlugin)
		
		// 卸载插件
		pluginGroup.POST("/unload", ps.handleUnloadPlugin)
		
		// 列出插件
		pluginGroup.GET("/list", ps.handleListPlugins)
	}
}

// handleLoadPlugin 处理加载插件的HTTP请求
func (ps *PluginService) handleLoadPlugin(c *gin.Context) {
	var req PluginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, PluginResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid request: %v", err),
		})
		return
	}
	
	// 获取完整插件路径
	pluginPath := filepath.Join(config.Settings.PluginPath, req.PluginPath)
	
	container, err := ps.LoadPlugin(pluginPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, PluginResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to load plugin: %v", err),
		})
		return
	}
	
	c.JSON(http.StatusOK, PluginResponse{
		Success: true,
		Data:    container,
		Message: "Plugin loaded successfully",
	})
}

// handleExecutePlugin 处理执行插件的HTTP请求
func (ps *PluginService) handleExecutePlugin(c *gin.Context) {
	var req PluginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, PluginResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid request: %v", err),
		})
		return
	}
	
	response, err := ps.ExecutePlugin(req.PluginPath, req.Action, req.Params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, PluginResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to execute plugin: %v", err),
		})
		return
	}
	
	c.JSON(http.StatusOK, response)
}

// handleUnloadPlugin 处理卸载插件的HTTP请求
func (ps *PluginService) handleUnloadPlugin(c *gin.Context) {
	var req PluginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, PluginResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid request: %v", err),
		})
		return
	}
	
	err := ps.UnloadPlugin(req.PluginPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, PluginResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to unload plugin: %v", err),
		})
		return
	}
	
	c.JSON(http.StatusOK, PluginResponse{
		Success: true,
		Message: "Plugin unloaded successfully",
	})
}

// handleListPlugins 处理列出插件的HTTP请求
func (ps *PluginService) handleListPlugins(c *gin.Context) {
	containers := ps.ListPlugins()
	
	c.JSON(http.StatusOK, PluginResponse{
		Success: true,
		Data:    containers,
	})
}

// Start 启动插件服务
func (ps *PluginService) Start() error {
	// 确保插件目录存在
	if err := os.MkdirAll(config.Settings.PluginPath, 0755); err != nil {
		return fmt.Errorf("failed to create plugin directory: %v", err)
	}
	
	return nil
}

// Stop 停止插件服务
func (ps *PluginService) Stop() error {
	// 卸载所有插件容�?	for containerID := range ps.pluginContainers {
		if err := ps.UnloadPlugin(containerID); err != nil {
			// 记录错误但继续卸载其他容�?			fmt.Printf("Failed to unload plugin container %s: %v\n", containerID, err)
		}
	}
	
	return nil
}
