package plugin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// PythonPluginManager Python插件管理器
type PythonPluginManager struct {
	httpClient *http.Client
	baseURL    string
	plugins    map[string]*PythonPluginFields
	logger     Logger
}

// NewPythonPluginManager 创建Python插件管理器
func NewPythonPluginManager(host string, port int, timeout int) (*PythonPluginManager, error) {
	baseURL := fmt.Sprintf("http://%s:%d", host, port)
	
	return &PythonPluginManager{
		httpClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
		baseURL: baseURL,
		plugins: make(map[string]*PythonPluginFields),
		logger:  NewLogger(),
	}, nil
}

// LoadPlugin 加载Python插件
func (ppm *PythonPluginManager) LoadPlugin(pluginPath string) (Plugin, error) {
	// 解析插件信息（这里简化处理，实际应该从plugin.json读取）
	pluginID := fmt.Sprintf("python_%d", time.Now().UnixNano())
	
	// 创建Python插件实例
	pythonPlugin := &PythonPluginFields{
		id:          pluginID,
		name:        "Python Plugin",
		version:     "1.0.0",
		pluginType:  PluginTypeScript,
		description: "Python script plugin",
		config:      make(map[string]interface{}),
		state:       StateLoaded,
		pythonURL:   ppm.baseURL,
	}

	// 存储插件
	ppm.plugins[pluginID] = pythonPlugin

	ppm.logger.Info("Python plugin loaded", "id", pluginID, "path", pluginPath)

	return pythonPlugin, nil
}

// UnloadPlugin 卸载Python插件
func (ppm *PythonPluginManager) UnloadPlugin(pluginID string) error {
	wrapper, exists := ppm.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	// 停止插件
	if wrapper.state == StateRunning {
		if err := wrapper.Stop(); err != nil {
			ppm.logger.Error("Failed to stop plugin before unloading", "id", pluginID, "error", err)
		}
	}

	// 从内存中移除
	delete(ppm.plugins, pluginID)

	ppm.logger.Info("Python plugin unloaded", "id", pluginID)

	return nil
}

// GetPlugin 获取插件
func (ppm *PythonPluginManager) GetPlugin(pluginID string) (Plugin, error) {
	wrapper, exists := ppm.plugins[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", pluginID)
	}

	return wrapper, nil
}

// ListPlugins 列出所有插件
func (ppm *PythonPluginManager) ListPlugins() []Plugin {
	var plugins []Plugin
	for _, wrapper := range ppm.plugins {
		plugins = append(plugins, wrapper)
	}

	return plugins
}

// PythonPluginFields Python插件内部字段
type PythonPluginFields struct {
	id          string
	name        string
	version     string
	pluginType  PluginType
	description string
	config      map[string]interface{}
	state       PluginState
	pythonURL   string
}

// 实现Plugin接口的方法
func (pp *PythonPluginFields) ID() string {
	return pp.id
}

func (pp *PythonPluginFields) Name() string {
	return pp.name
}

func (pp *PythonPluginFields) Version() string {
	return pp.version
}

func (pp *PythonPluginFields) Type() PluginType {
	return pp.pluginType
}

func (pp *PythonPluginFields) Description() string {
	return pp.description
}

func (pp *PythonPluginFields) Initialize(config map[string]interface{}) error {
	if pp.state != StateLoaded {
		return fmt.Errorf("plugin not in loaded state")
	}

	pp.config = config
	
	// 调用Python插件的初始化方法
	url := fmt.Sprintf("%s/plugin/%s/initialize", pp.pythonURL, pp.id)
	
	request := map[string]interface{}{
		"config": config,
	}
	
	if err := pp.callPythonAPI(url, request); err != nil {
		pp.state = StateError
		return fmt.Errorf("failed to initialize python plugin: %w", err)
	}

	pp.state = StateInitialized
	return nil
}

func (pp *PythonPluginFields) Start() error {
	if pp.state != StateInitialized {
		return fmt.Errorf("plugin not initialized")
	}

	// 调用Python插件的启动方法
	url := fmt.Sprintf("%s/plugin/%s/start", pp.pythonURL, pp.id)
	
	if err := pp.callPythonAPI(url, nil); err != nil {
		pp.state = StateError
		return fmt.Errorf("failed to start python plugin: %w", err)
	}

	pp.state = StateRunning
	return nil
}

func (pp *PythonPluginFields) Stop() error {
	if pp.state != StateRunning {
		return fmt.Errorf("plugin not running")
	}

	// 调用Python插件的停止方法
	url := fmt.Sprintf("%s/plugin/%s/stop", pp.pythonURL, pp.id)
	
	if err := pp.callPythonAPI(url, nil); err != nil {
		return err
	}

	pp.state = StateStopped
	return nil
}

func (pp *PythonPluginFields) Destroy() error {
	// 调用Python插件的销毁方法
	url := fmt.Sprintf("%s/plugin/%s/destroy", pp.pythonURL, pp.id)
	
	if err := pp.callPythonAPI(url, nil); err != nil {
		return err
	}

	pp.state = StateUnloaded
	return nil
}

func (pp *PythonPluginFields) GetState() PluginState {
	return pp.state
}

func (pp *PythonPluginFields) HandleEvent(event Event) error {
	// 调用Python插件的事件处理方法
	url := fmt.Sprintf("%s/plugin/%s/event", pp.pythonURL, pp.id)
	
	return pp.callPythonAPI(url, event)
}

func (pp *PythonPluginFields) GetConfigForm() *ConfigForm {
	// 从Python插件获取配置表单
	url := fmt.Sprintf("%s/plugin/%s/config_form", pp.pythonURL, pp.id)
	
	var configForm ConfigForm
	if err := pp.callPythonAPIWithResponse(url, nil, &configForm); err != nil {
		return nil
	}
	
	return &configForm
}

func (pp *PythonPluginFields) GetAPIRoutes() []APIRoute {
	// 从Python插件获取API路由
	url := fmt.Sprintf("%s/plugin/%s/api_routes", pp.pythonURL, pp.id)
	
	var routes []APIRoute
	if err := pp.callPythonAPIWithResponse(url, nil, &routes); err != nil {
		return nil
	}
	
	return routes
}

func (pp *PythonPluginFields) GetCommands() []Command {
	// 从Python插件获取命令
	url := fmt.Sprintf("%s/plugin/%s/commands", pp.pythonURL, pp.id)
	
	var commands []Command
	if err := pp.callPythonAPIWithResponse(url, nil, &commands); err != nil {
		return nil
	}
	
	return commands
}

func (pp *PythonPluginFields) GetServices() []Service {
	// 从Python插件获取服务
	url := fmt.Sprintf("%s/plugin/%s/services", pp.pythonURL, pp.id)
	
	var services []Service
	if err := pp.callPythonAPIWithResponse(url, nil, &services); err != nil {
		return nil
	}
	
	return services
}

// callPythonAPI 调用Python API
func (pp *PythonPluginFields) callPythonAPI(url string, data interface{}) error {
	var requestBody *bytes.Buffer
	
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return err
		}
		requestBody = bytes.NewBuffer(jsonData)
	} else {
		requestBody = bytes.NewBuffer([]byte("{}"))
	}

	resp, err := http.Post(url, "application/json", requestBody)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("python API call failed: status %d", resp.StatusCode)
	}

	return nil
}

// callPythonAPIWithResponse 调用Python API并获取响应
func (pp *PythonPluginFields) callPythonAPIWithResponse(url string, data interface{}, response interface{}) error {
	var requestBody *bytes.Buffer
	
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return err
		}
		requestBody = bytes.NewBuffer(jsonData)
	} else {
		requestBody = bytes.NewBuffer([]byte("{}"))
	}

	resp, err := http.Post(url, "application/json", requestBody)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("python API call failed: status %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(response)
}