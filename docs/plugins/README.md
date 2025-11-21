# 插件系统开发指南

## 🔌 插件系统概览

MoviePilot Go 采用微服务插件架构，支持通过 gRPC 与 Python 插件服务通信，提供灵活的扩展能力。

### 架构设计

```
┌─────────────────┐    gRPC     ┌─────────────────┐
│   Go Main App   │ ◄─────────► │ Python Plugins  │
│   (Port 3001)   │             │   (Port 5000)   │
│                 │             │                 │
│ ┌─────────────┐ │             │ ┌─────────────┐ │
│ │ Plugin Mgr  │ │             │ │ Plugin Mgr  │ │
│ │ gRPC Client │ │             │ │ gRPC Server │ │
│ └─────────────┘ │             │ └─────────────┘ │
└─────────────────┘             └─────────────────┘
         │                               │
         ▼                               ▼
┌─────────────────┐             ┌─────────────────┐
│   Plugin Config │             │   Plugin Files  │
│   Registry      │             │   & Metadata    │
└─────────────────┘             └─────────────────┘
```

## 🏗️ 插件类型

### 1. Site Plugins (站点插件)
负责从各种站点抓取媒体资源信息。

**功能**:
- 站点资源搜索
- 详细信息获取
- 下载链接解析
- 认证和会话管理

### 2. Indexer Plugins (索引器插件)
提供媒体资源的索引和搜索功能。

**功能**:
- 资源索引构建
- 全文搜索
- 分类筛选
- 排序和推荐

### 3. MediaServer Plugins (媒体服务器插件)
与外部媒体服务器集成。

**功能**:
- 媒体库同步
- 播放状态更新
- 元数据管理
- 用户认证

### 4. Notification Plugins (通知插件)
发送各种类型的通知。

**功能**:
- 邮件通知
- 短信通知
- 推送通知
- Webhook 集成

## 🐍 Python 插件开发

### 1. 插件基础结构

#### 插件目录结构
```
python-plugins/plugins/site/example_site/
├── __init__.py
├── plugin.json
├── main.py
├── config.py
├── client.py
├── utils.py
└── tests/
    ├── __init__.py
    └── test_main.py
```

#### 插件配置文件
```json
// plugin.json
{
    "id": "example_site",
    "name": "Example Site Plugin",
    "version": "1.0.0",
    "type": "site",
    "description": "Example site plugin for demonstration",
    "author": "MoviePilot Team",
    "homepage": "https://github.com/yfh-yun/moviepilot-go",
    "license": "MIT",
    "dependencies": [
        "requests>=2.28.0",
        "beautifulsoup4>=4.11.0"
    ],
    "config_schema": {
        "type": "object",
        "properties": {
            "base_url": {
                "type": "string",
                "title": "Base URL",
                "default": "https://example.com"
            },
            "api_key": {
                "type": "string",
                "title": "API Key",
                "format": "password"
            },
            "timeout": {
                "type": "integer",
                "title": "Request Timeout",
                "default": 30,
                "minimum": 1,
                "maximum": 300
            }
        },
        "required": ["base_url"]
    },
    "permissions": [
        "network.request",
        "file.read"
    ]
}
```

### 2. 插件接口实现

#### 基础插件类
```python
# main.py
from abc import ABC, abstractmethod
from typing import Dict, List, Any, Optional
import logging

logger = logging.getLogger(__name__)

class BasePlugin(ABC):
    """插件基类"""
    
    def __init__(self, config: Dict[str, Any]):
        self.config = config
        self.logger = logging.getLogger(self.__class__.__name__)
    
    @abstractmethod
    def initialize(self) -> bool:
        """插件初始化"""
        pass
    
    @abstractmethod
    def cleanup(self) -> bool:
        """插件清理"""
        pass
    
    def get_info(self) -> Dict[str, Any]:
        """获取插件信息"""
        return {
            "id": self.config.get("id"),
            "name": self.config.get("name"),
            "version": self.config.get("version"),
            "type": self.config.get("type")
        }

class SitePlugin(BasePlugin):
    """站点插件基类"""
    
    @abstractmethod
    async def search(self, keyword: str, **kwargs) -> List[Dict[str, Any]]:
        """搜索媒体资源"""
        pass
    
    @abstractmethod
    async def get_details(self, media_id: str) -> Optional[Dict[str, Any]]:
        """获取媒体详情"""
        pass
    
    @abstractmethod
    async def get_download_links(self, media_id: str) -> List[str]:
        """获取下载链接"""
        pass
```

#### 具体插件实现
```python
# main.py (续)
import aiohttp
import asyncio
from bs4 import BeautifulSoup

class ExampleSitePlugin(SitePlugin):
    """示例站点插件"""
    
    def __init__(self, config: Dict[str, Any]):
        super().__init__(config)
        self.base_url = config["base_url"]
        self.api_key = config.get("api_key")
        self.timeout = config.get("timeout", 30)
        self.session = None
    
    def initialize(self) -> bool:
        """初始化插件"""
        try:
            self.session = aiohttp.ClientSession(
                timeout=aiohttp.ClientTimeout(total=self.timeout),
                headers={
                    "User-Agent": "MoviePilot/1.0",
                    "Authorization": f"Bearer {self.api_key}" if self.api_key else None
                }
            )
            self.logger.info(f"Plugin {self.config['id']} initialized successfully")
            return True
        except Exception as e:
            self.logger.error(f"Failed to initialize plugin: {e}")
            return False
    
    async def cleanup(self) -> bool:
        """清理插件资源"""
        if self.session:
            await self.session.close()
        self.logger.info(f"Plugin {self.config['id']} cleaned up")
        return True
    
    async def search(self, keyword: str, **kwargs) -> List[Dict[str, Any]]:
        """搜索媒体资源"""
        try:
            url = f"{self.base_url}/api/search"
            params = {
                "q": keyword,
                "type": kwargs.get("media_type", "all"),
                "page": kwargs.get("page", 1),
                "limit": kwargs.get("limit", 20)
            }
            
            async with self.session.get(url, params=params) as response:
                if response.status == 200:
                    data = await response.json()
                    return self._parse_search_results(data)
                else:
                    self.logger.error(f"Search failed with status {response.status}")
                    return []
        
        except Exception as e:
            self.logger.error(f"Search error: {e}")
            return []
    
    async def get_details(self, media_id: str) -> Optional[Dict[str, Any]]:
        """获取媒体详情"""
        try:
            url = f"{self.base_url}/api/media/{media_id}"
            
            async with self.session.get(url) as response:
                if response.status == 200:
                    data = await response.json()
                    return self._parse_media_details(data)
                else:
                    self.logger.error(f"Get details failed with status {response.status}")
                    return None
        
        except Exception as e:
            self.logger.error(f"Get details error: {e}")
            return None
    
    async def get_download_links(self, media_id: str) -> List[str]:
        """获取下载链接"""
        try:
            url = f"{self.base_url}/api/media/{media_id}/download"
            
            async with self.session.get(url) as response:
                if response.status == 200:
                    data = await response.json()
                    return data.get("download_links", [])
                else:
                    self.logger.error(f"Get download links failed with status {response.status}")
                    return []
        
        except Exception as e:
            self.logger.error(f"Get download links error: {e}")
            return []
    
    def _parse_search_results(self, data: Dict[str, Any]) -> List[Dict[str, Any]]:
        """解析搜索结果"""
        results = []
        for item in data.get("items", []):
            results.append({
                "id": item["id"],
                "title": item["title"],
                "type": item["type"],
                "year": item.get("year"),
                "rating": item.get("rating"),
                "poster_url": item.get("poster_url"),
                "overview": item.get("overview", "")
            })
        return results
    
    def _parse_media_details(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """解析媒体详情"""
        return {
            "id": data["id"],
            "title": data["title"],
            "type": data["type"],
            "year": data.get("year"),
            "genre": data.get("genre", []),
            "rating": data.get("rating"),
            "poster_url": data.get("poster_url"),
            "backdrop_url": data.get("backdrop_url"),
            "overview": data.get("overview", ""),
            "cast": data.get("cast", []),
            "director": data.get("director", []),
            "duration": data.get("duration"),
            "release_date": data.get("release_date")
        }

# 插件入口点
def create_plugin(config: Dict[str, Any]) -> BasePlugin:
    return ExampleSitePlugin(config)
```

### 3. 插件配置管理

#### 配置类
```python
# config.py
from dataclasses import dataclass
from typing import Optional, Dict, Any
import json
import os

@dataclass
class PluginConfig:
    """插件配置类"""
    id: str
    name: str
    version: str
    type: str
    description: str = ""
    author: str = ""
    homepage: str = ""
    license: str = ""
    dependencies: List[str] = None
    config_schema: Dict[str, Any] = None
    permissions: List[str] = None
    
    @classmethod
    def from_file(cls, config_path: str) -> 'PluginConfig':
        """从文件加载配置"""
        with open(config_path, 'r', encoding='utf-8') as f:
            data = json.load(f)
        return cls(**data)
    
    def validate(self) -> bool:
        """验证配置"""
        required_fields = ['id', 'name', 'version', 'type']
        for field in required_fields:
            if not getattr(self, field):
                return False
        return True

class ConfigManager:
    """配置管理器"""
    
    def __init__(self, config_dir: str):
        self.config_dir = config_dir
        self.configs = {}
    
    def load_config(self, plugin_id: str) -> Optional[PluginConfig]:
        """加载插件配置"""
        config_path = os.path.join(self.config_dir, plugin_id, "plugin.json")
        if not os.path.exists(config_path):
            return None
        
        try:
            config = PluginConfig.from_file(config_path)
            if config.validate():
                self.configs[plugin_id] = config
                return config
            else:
                raise ValueError(f"Invalid plugin config: {plugin_id}")
        except Exception as e:
            print(f"Failed to load config for {plugin_id}: {e}")
            return None
    
    def get_config(self, plugin_id: str) -> Optional[PluginConfig]:
        """获取插件配置"""
        return self.configs.get(plugin_id)
```

## 🔗 gRPC 通信

### 1. Protocol Buffers 定义

#### plugin.proto
```protobuf
syntax = "proto3";

package plugin;

// 插件服务定义
service PluginService {
  // 获取插件列表
  rpc ListPlugins(ListPluginsRequest) returns (ListPluginsResponse);
  
  // 启用插件
  rpc EnablePlugin(EnablePluginRequest) returns (EnablePluginResponse);
  
  // 禁用插件
  rpc DisablePlugin(DisablePluginRequest) returns (DisablePluginResponse);
  
  // 配置插件
  rpc ConfigurePlugin(ConfigurePluginRequest) returns (ConfigurePluginResponse);
  
  // 执行插件方法
  rpc ExecutePlugin(ExecutePluginRequest) returns (ExecutePluginResponse);
  
  // 获取插件状态
  rpc GetPluginStatus(GetPluginStatusRequest) returns (GetPluginStatusResponse);
}

// 基础消息
message PluginInfo {
  string id = 1;
  string name = 2;
  string version = 3;
  string type = 4;
  string description = 5;
  bool enabled = 6;
  map<string, string> config = 7;
}

// 请求和响应消息
message ListPluginsRequest {
  string type = 1;  // 插件类型过滤
}

message ListPluginsResponse {
  repeated PluginInfo plugins = 1;
  string error = 2;
}

message EnablePluginRequest {
  string plugin_id = 1;
}

message EnablePluginResponse {
  bool success = 1;
  string error = 2;
}

message DisablePluginRequest {
  string plugin_id = 1;
}

message DisablePluginResponse {
  bool success = 1;
  string error = 2;
}

message ConfigurePluginRequest {
  string plugin_id = 1;
  map<string, string> config = 2;
}

message ConfigurePluginResponse {
  bool success = 1;
  string error = 2;
}

message ExecutePluginRequest {
  string plugin_id = 1;
  string method = 2;
  map<string, string> params = 3;
}

message ExecutePluginResponse {
  bool success = 1;
  string data = 2;  // JSON 格式的响应数据
  string error = 3;
}

message GetPluginStatusRequest {
  string plugin_id = 1;
}

message GetPluginStatusResponse {
  string status = 1;  // running, stopped, error
  string message = 2;
  map<string, string> metrics = 3;
}
```

### 2. Python gRPC 服务端

#### server.py
```python
# python-plugins/internal/api/server.py
import grpc
from concurrent import futures
import json
import logging
from typing import Dict, Any

import plugin_pb2
import plugin_pb2_grpc

logger = logging.getLogger(__name__)

class PluginServiceImpl(plugin_pb2_grpc.PluginServiceServicer):
    """插件服务实现"""
    
    def __init__(self, plugin_manager):
        self.plugin_manager = plugin_manager
    
    def ListPlugins(self, request, context):
        """获取插件列表"""
        try:
            plugins = self.plugin_manager.list_plugins(request.type)
            plugin_infos = []
            
            for plugin in plugins:
                info = plugin_pb2.PluginInfo(
                    id=plugin.id,
                    name=plugin.name,
                    version=plugin.version,
                    type=plugin.type,
                    description=plugin.description,
                    enabled=plugin.enabled
                )
                # 添加配置信息
                if plugin.config:
                    info.config.update(plugin.config)
                plugin_infos.append(info)
            
            return plugin_pb2.ListPluginsResponse(plugins=plugin_infos)
        
        except Exception as e:
            logger.error(f"List plugins error: {e}")
            return plugin_pb2.ListPluginsResponse(error=str(e))
    
    def EnablePlugin(self, request, context):
        """启用插件"""
        try:
            success = self.plugin_manager.enable_plugin(request.plugin_id)
            return plugin_pb2.EnablePluginResponse(success=success)
        
        except Exception as e:
            logger.error(f"Enable plugin error: {e}")
            return plugin_pb2.EnablePluginResponse(success=False, error=str(e))
    
    def DisablePlugin(self, request, context):
        """禁用插件"""
        try:
            success = self.plugin_manager.disable_plugin(request.plugin_id)
            return plugin_pb2.DisablePluginResponse(success=success)
        
        except Exception as e:
            logger.error(f"Disable plugin error: {e}")
            return plugin_pb2.DisablePluginResponse(success=False, error=str(e))
    
    def ConfigurePlugin(self, request, context):
        """配置插件"""
        try:
            config = dict(request.config)
            success = self.plugin_manager.configure_plugin(
                request.plugin_id, config
            )
            return plugin_pb2.ConfigurePluginResponse(success=success)
        
        except Exception as e:
            logger.error(f"Configure plugin error: {e}")
            return plugin_pb2.ConfigurePluginResponse(
                success=False, error=str(e)
            )
    
    def ExecutePlugin(self, request, context):
        """执行插件方法"""
        try:
            params = dict(request.params)
            result = self.plugin_manager.execute_plugin(
                request.plugin_id, request.method, params
            )
            
            if result is not None:
                data = json.dumps(result) if isinstance(result, (dict, list)) else str(result)
                return plugin_pb2.ExecutePluginResponse(
                    success=True, data=data
                )
            else:
                return plugin_pb2.ExecutePluginResponse(
                    success=False, error="No result returned"
                )
        
        except Exception as e:
            logger.error(f"Execute plugin error: {e}")
            return plugin_pb2.ExecutePluginResponse(
                success=False, error=str(e)
            )
    
    def GetPluginStatus(self, request, context):
        """获取插件状态"""
        try:
            status = self.plugin_manager.get_plugin_status(request.plugin_id)
            return plugin_pb2.GetPluginStatusResponse(
                status=status["status"],
                message=status.get("message", ""),
                metrics=status.get("metrics", {})
            )
        
        except Exception as e:
            logger.error(f"Get plugin status error: {e}")
            return plugin_pb2.GetPluginStatusResponse(
                status="error",
                message=str(e)
            )

def serve(port: int = 5000, plugin_manager=None):
    """启动 gRPC 服务"""
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    plugin_pb2_grpc.add_PluginServiceServicer_to_server(
        PluginServiceImpl(plugin_manager), server
    )
    
    server.add_insecure_port(f'[::]:{port}')
    logger.info(f"Plugin server starting on port {port}")
    server.start()
    server.wait_for_termination()
```

## 🎯 Go 插件客户端

### 1. gRPC 客户端实现

```go
// pkg/plugin/client.go
package plugin

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    
    pb "moviepilot-go/shared/proto/plugin"
)

type Client struct {
    conn   *grpc.ClientConn
    client pb.PluginServiceClient
}

func NewClient(address string) (*Client, error) {
    conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        return nil, fmt.Errorf("failed to connect to plugin server: %w", err)
    }

    client := pb.NewPluginServiceClient(conn)
    
    return &Client{
        conn:   conn,
        client: client,
    }, nil
}

func (c *Client) Close() error {
    return c.conn.Close()
}

func (c *Client) ListPlugins(ctx context.Context, pluginType string) ([]*PluginInfo, error) {
    req := &pb.ListPluginsRequest{
        Type: pluginType,
    }
    
    resp, err := c.client.ListPlugins(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("failed to list plugins: %w", err)
    }
    
    if resp.Error != "" {
        return nil, fmt.Errorf("plugin server error: %s", resp.Error)
    }
    
    var plugins []*PluginInfo
    for _, p := range resp.Plugins {
        plugins = append(plugins, &PluginInfo{
            ID:          p.Id,
            Name:        p.Name,
            Version:     p.Version,
            Type:        p.Type,
            Description: p.Description,
            Enabled:     p.Enabled,
            Config:      p.Config,
        })
    }
    
    return plugins, nil
}

func (c *Client) EnablePlugin(ctx context.Context, pluginID string) error {
    req := &pb.EnablePluginRequest{
        PluginId: pluginID,
    }
    
    resp, err := c.client.EnablePlugin(ctx, req)
    if err != nil {
        return fmt.Errorf("failed to enable plugin: %w", err)
    }
    
    if !resp.Success {
        return fmt.Errorf("enable plugin failed: %s", resp.Error)
    }
    
    return nil
}

func (c *Client) ExecutePlugin(ctx context.Context, pluginID, method string, params map[string]interface{}) (interface{}, error) {
    paramsBytes, _ := json.Marshal(params)
    paramsStr := string(paramsBytes)
    
    req := &pb.ExecutePluginRequest{
        PluginId: pluginID,
        Method:   method,
        Params:   paramsStr,
    }
    
    resp, err := c.client.ExecutePlugin(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("failed to execute plugin: %w", err)
    }
    
    if !resp.Success {
        return nil, fmt.Errorf("execute plugin failed: %s", resp.Error)
    }
    
    var result interface{}
    if resp.Data != "" {
        err = json.Unmarshal([]byte(resp.Data), &result)
        if err != nil {
            return nil, fmt.Errorf("failed to parse plugin result: %w", err)
        }
    }
    
    return result, nil
}
```

### 2. 插件管理器

```go
// pkg/plugin/manager.go
package plugin

import (
    "context"
    "sync"
    "time"
    
    "moviepilot-go/pkg/logger"
)

type Manager struct {
    client     *Client
    plugins    map[string]*PluginInfo
    mutex      sync.RWMutex
    logger     logger.Logger
}

type PluginInfo struct {
    ID          string            `json:"id"`
    Name        string            `json:"name"`
    Version     string            `json:"version"`
    Type        string            `json:"type"`
    Description string            `json:"description"`
    Enabled     bool              `json:"enabled"`
    Config      map[string]string `json:"config"`
}

func NewManager(pluginServerAddress string, logger logger.Logger) (*Manager, error) {
    client, err := NewClient(pluginServerAddress)
    if err != nil {
        return nil, err
    }
    
    manager := &Manager{
        client:  client,
        plugins: make(map[string]*PluginInfo),
        logger:  logger,
    }
    
    // 启动插件发现
    go manager.discoveryLoop()
    
    return manager, nil
}

func (m *Manager) discoveryLoop() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            m.refreshPlugins()
        }
    }
}

func (m *Manager) refreshPlugins() {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    plugins, err := m.client.ListPlugins(ctx, "")
    if err != nil {
        m.logger.Error("Failed to refresh plugins", "error", err.Error())
        return
    }
    
    m.mutex.Lock()
    defer m.mutex.Unlock()
    
    m.plugins = make(map[string]*PluginInfo)
    for _, plugin := range plugins {
        m.plugins[plugin.ID] = plugin
    }
    
    m.logger.Info("Plugins refreshed", "count", len(m.plugins))
}

func (m *Manager) GetPlugin(id string) (*PluginInfo, bool) {
    m.mutex.RLock()
    defer m.mutex.RUnlock()
    
    plugin, exists := m.plugins[id]
    return plugin, exists
}

func (m *Manager) ListPlugins(pluginType string) []*PluginInfo {
    m.mutex.RLock()
    defer m.mutex.RUnlock()
    
    var plugins []*PluginInfo
    for _, plugin := range m.plugins {
        if pluginType == "" || plugin.Type == pluginType {
            plugins = append(plugins, plugin)
        }
    }
    
    return plugins
}

func (m *Manager) EnablePlugin(id string) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    err := m.client.EnablePlugin(ctx, id)
    if err != nil {
        m.logger.Error("Failed to enable plugin", "plugin_id", id, "error", err.Error())
        return err
    }
    
    m.logger.Info("Plugin enabled", "plugin_id", id)
    m.refreshPlugins()
    
    return nil
}

func (m *Manager) ExecuteSitePluginSearch(ctx context.Context, pluginID, keyword string, params map[string]interface{}) (interface{}, error) {
    return m.client.ExecutePlugin(ctx, pluginID, "search", map[string]interface{}{
        "keyword": keyword,
        "params":  params,
    })
}
```

## 🧪 插件测试

### 1. 单元测试

```python
# tests/test_example_site.py
import pytest
import asyncio
from unittest.mock import Mock, AsyncMock

from plugins.site.example_site.main import ExampleSitePlugin

@pytest.fixture
def plugin_config():
    return {
        "id": "example_site",
        "name": "Example Site",
        "version": "1.0.0",
        "type": "site",
        "base_url": "https://api.example.com",
        "api_key": "test_key",
        "timeout": 30
    }

@pytest.fixture
async def plugin(plugin_config):
    plugin = ExampleSitePlugin(plugin_config)
    await plugin.initialize()
    yield plugin
    await plugin.cleanup()

@pytest.mark.asyncio
async def test_search(plugin):
    """测试搜索功能"""
    results = await plugin.search("test movie")
    
    assert isinstance(results, list)
    if results:
        result = results[0]
        assert "id" in result
        assert "title" in result
        assert "type" in result

@pytest.mark.asyncio
async def test_get_details(plugin):
    """测试获取详情功能"""
    details = await plugin.get_details("test_id")
    
    if details:
        assert "id" in details
        assert "title" in details
        assert "overview" in details

@pytest.mark.asyncio
async def test_get_download_links(plugin):
    """测试获取下载链接功能"""
    links = await plugin.get_download_links("test_id")
    
    assert isinstance(links, list)
```

### 2. 集成测试

```python
# tests/test_plugin_integration.py
import pytest
import grpc
from concurrent import futures

from python_plugins.internal.api.server import PluginServiceImpl
import plugin_pb2_grpc

@pytest.fixture
def grpc_service():
    plugin_manager = Mock()  # 使用模拟的插件管理器
    service = PluginServiceImpl(plugin_manager)
    return service

@pytest.fixture
def grpc_server(grpc_service):
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    plugin_pb2_grpc.add_PluginServiceServicer_to_server(grpc_service, server)
    server.add_insecure_port('[::]:5001')
    server.start()
    yield server
    server.stop(None)

@pytest.mark.asyncio
async def test_list_plugins_grpc(grpc_server):
    """测试 gRPC 插件列表接口"""
    channel = grpc.insecure_channel('localhost:5001')
    stub = plugin_pb2_grpc.PluginServiceStub(channel)
    
    request = plugin_pb2.ListPluginsRequest(type="site")
    response = stub.ListPlugins(request)
    
    assert response is not None
    # 进一步的断言...
    
    channel.close()
```

## 📦 插件部署

### 1. Docker 部署

```dockerfile
# python-plugins/Dockerfile
FROM python:3.11-slim

WORKDIR /app

# 安装系统依赖
RUN apt-get update && apt-get install -y \
    gcc \
    && rm -rf /var/lib/apt/lists/*

# 复制依赖文件
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# 复制插件代码
COPY . .

# 安装插件依赖
RUN python scripts/install_plugin_deps.py

# 暴露 gRPC 端口
EXPOSE 5000

# 启动插件服务
CMD ["python", "cmd/server/main.py"]
```

### 2. Docker Compose 配置

```yaml
# deployments/docker-compose.yml (插件部分)
services:
  plugins:
    build:
      context: ../python-plugins
      dockerfile: Dockerfile
    container_name: moviepilot-plugins
    restart: unless-stopped
    ports:
      - "5000:5000"
    environment:
      - GRPC_SERVER_HOST=0.0.0.0
      - GRPC_SERVER_PORT=5000
      - LOG_LEVEL=info
    volumes:
      - ../python-plugins/plugins:/app/plugins
      - ../python-plugins/configs:/app/configs
    networks:
      - moviepilot-network
    depends_on:
      - redis
```

## 📚 最佳实践

### 1. 插件开发规范

- **错误处理**: 使用结构化错误信息，包含足够的上下文
- **日志记录**: 使用结构化日志，包含插件 ID 和操作信息
- **配置验证**: 在初始化时验证配置参数
- **资源管理**: 正确处理资源的创建和清理
- **异步操作**: 使用 async/await 处理 I/O 操作

### 2. 性能优化

- **连接池**: 复用 HTTP 连接
- **缓存**: 缓存频繁访问的数据
- **超时控制**: 设置合理的请求超时
- **并发控制**: 限制并发请求数量

### 3. 安全考虑

- **输入验证**: 验证所有外部输入
- **敏感信息**: 安全处理 API 密钥等敏感信息
- **权限控制**: 遵循最小权限原则
- **安全通信**: 使用 HTTPS/TLS 加密通信

---

**注意**: 插件开发需要遵循项目规范和最佳实践，确保系统的稳定性和安全性。在提交插件前，请确保所有测试通过并且插件符合质量标准。