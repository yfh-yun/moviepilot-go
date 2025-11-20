#!/bin/bash

# MoviePilot-Go 配置管理重构脚本
# 创建统一的配置管理结构

set -e

echo "🚀 开始配置管理重构..."

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查是否在正确的目录
if [ ! -f "go.mod" ]; then
    log_error "请在 moviepilot-go 根目录执行此脚本"
    exit 1
fi

# 创建配置目录结构
log_info "创建配置目录结构..."

mkdir -p config/{core,environments,validation,providers}
mkdir -p config/environments/{dev,test,prod}
mkdir -p config/validation/{schemas,rules}
mkdir -p config/providers/{file,env,vault}

# 移动现有配置文件
log_info "移动现有配置文件..."

# 移动核心配置
if [ -f "configs/config.yaml.sample" ]; then
    cp configs/config.yaml.sample config/core/
    log_info "核心配置已移动"
fi

if [ -f "configs/plugins.json" ]; then
    cp configs/plugins.json config/core/
    log_info "插件配置已移动"
fi

# 创建环境特定配置
log_info "创建环境特定配置..."

# 开发环境配置
cat > config/environments/dev/config.yaml << 'EOF'
# 开发环境配置
app:
  name: "MoviePilot-Go"
  version: "2.8.1"
  env: "development"
  debug: true

server:
  host: "0.0.0.0"
  port: 3001
  timeout: 30s

database:
  type: "sqlite"
  connection: "./data/moviepilot-dev.db"
  max_connections: 10
  log_level: "debug"

redis:
  host: "localhost"
  port: 6379
  db: 0
  password: ""

logging:
  level: "debug"
  format: "console"
  output: "stdout"
EOF

# 测试环境配置
cat > config/environments/test/config.yaml << 'EOF'
# 测试环境配置
app:
  name: "MoviePilot-Go"
  version: "2.8.1"
  env: "test"
  debug: true

server:
  host: "0.0.0.0"
  port: 3002
  timeout: 30s

database:
  type: "sqlite"
  connection: ":memory:"
  max_connections: 5
  log_level: "warn"

redis:
  host: "localhost"
  port: 6379
  db: 1
  password: ""

logging:
  level: "info"
  format: "json"
  output: "stdout"
EOF

# 生产环境配置
cat > config/environments/prod/config.yaml << 'EOF'
# 生产环境配置
app:
  name: "MoviePilot-Go"
  version: "2.8.1"
  env: "production"
  debug: false

server:
  host: "0.0.0.0"
  port: 3001
  timeout: 30s

database:
  type: "postgres"
  connection: "postgres://user:password@localhost:5432/moviepilot?sslmode=disable"
  max_connections: 20
  log_level: "error"

redis:
  host: "redis-cluster"
  port: 6379
  db: 0
  password: "${REDIS_PASSWORD}"

logging:
  level: "warn"
  format: "json"
  output: "/var/log/moviepilot/app.log"
EOF

# 创建配置验证规则
log_info "创建配置验证规则..."

cat > config/validation/rules/app_rules.yaml << 'EOF'
# 应用配置验证规则
app:
  required: true
  type: object
  properties:
    name:
      type: string
      minLength: 1
      maxLength: 100
    version:
      type: string
      pattern: "^\\d+\\.\\d+\\.\\d+$"
    env:
      type: string
      enum: ["development", "test", "production"]
    debug:
      type: boolean

server:
  required: true
  type: object
  properties:
    host:
      type: string
      minLength: 1
    port:
      type: integer
      minimum: 1
      maximum: 65535
    timeout:
      type: string
      pattern: "^\\d+[smh]$"
EOF

# 创建配置提供者接口
log_info "创建配置提供者..."

cat > config/providers/provider.go << 'EOF'
// Package providers 配置提供者接口
package providers

import "context"

// ConfigProvider 配置提供者接口
type ConfigProvider interface {
    // Load 加载配置
    Load(ctx context.Context, path string) (map[string]interface{}, error)
    
    // Watch 监听配置变化
    Watch(ctx context.Context, path string, callback func(map[string]interface{})) error
    
    // Save 保存配置
    Save(ctx context.Context, path string, config map[string]interface{}) error
    
    // Validate 验证配置
    Validate(ctx context.Context, config map[string]interface{}, rules map[string]interface{}) error
}

// ProviderConfig 提供者配置
type ProviderConfig struct {
    Type     string                 `yaml:"type"`
    Settings map[string]interface{} `yaml:"settings"`
}
EOF

# 创建文件配置提供者
cat > config/providers/file/file_provider.go << 'EOF'
// Package file 文件配置提供者
package file

import (
    "context"
    "os"
    "path/filepath"
    
    "gopkg.in/yaml.v3"
    "github.com/yfh-yun/moviepilot-go/config/providers"
)

// FileProvider 文件配置提供者
type FileProvider struct {
    basePath string
}

// NewFileProvider 创建文件配置提供者
func NewFileProvider(basePath string) *FileProvider {
    return &FileProvider{
        basePath: basePath,
    }
}

// Load 加载配置文件
func (p *FileProvider) Load(ctx context.Context, path string) (map[string]interface{}, error) {
    fullPath := filepath.Join(p.basePath, path)
    
    data, err := os.ReadFile(fullPath)
    if err != nil {
        return nil, err
    }
    
    var config map[string]interface{}
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, err
    }
    
    return config, nil
}

// Watch 监听文件变化
func (p *FileProvider) Watch(ctx context.Context, path string, callback func(map[string]interface{})) error {
    // TODO: 实现文件监听
    return nil
}

// Save 保存配置文件
func (p *FileProvider) Save(ctx context.Context, path string, config map[string]interface{}) error {
    fullPath := filepath.Join(p.basePath, path)
    
    // 确保目录存在
    if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
        return err
    }
    
    data, err := yaml.Marshal(config)
    if err != nil {
        return err
    }
    
    return os.WriteFile(fullPath, data, 0644)
}

// Validate 验证配置
func (p *FileProvider) Validate(ctx context.Context, config map[string]interface{}, rules map[string]interface{}) error {
    // TODO: 实现配置验证
    return nil
}
EOF

# 创建配置管理器
cat > config/manager.go << 'EOF'
// Package config 配置管理
package config

import (
    "context"
    "fmt"
    "os"
    "path/filepath"
    
    "github.com/yfh-yun/moviepilot-go/config/providers"
    "github.com/yfh-yun/moviepilot-go/config/providers/file"
)

// Manager 配置管理器
type Manager struct {
    providers map[string]providers.ConfigProvider
    env       string
}

// NewManager 创建配置管理器
func NewManager(env string) *Manager {
    return &Manager{
        providers: make(map[string]providers.ConfigProvider),
        env:       env,
    }
}

// RegisterProvider 注册配置提供者
func (m *Manager) RegisterProvider(name string, provider providers.ConfigProvider) {
    m.providers[name] = provider
}

// Load 加载配置
func (m *Manager) Load(ctx context.Context, path string) (map[string]interface{}, error) {
    // 尝试从环境特定配置加载
    envPath := filepath.Join("environments", m.env, path)
    
    if provider, exists := m.providers["file"]; exists {
        config, err := provider.Load(ctx, envPath)
        if err == nil {
            return config, nil
        }
    }
    
    // 回退到默认配置
    if provider, exists := m.providers["file"]; exists {
        return provider.Load(ctx, path)
    }
    
    return nil, fmt.Errorf("no suitable provider found")
}

// GetEnv 获取当前环境
func (m *Manager) GetEnv() string {
    if m.env == "" {
        m.env = os.Getenv("APP_ENV")
        if m.env == "" {
            m.env = "development"
        }
    }
    return m.env
}

// Init 初始化配置管理器
func Init() (*Manager, error) {
    env := os.Getenv("APP_ENV")
    if env == "" {
        env = "development"
    }
    
    manager := NewManager(env)
    
    // 注册文件提供者
    fileProvider := file.NewFileProvider("config")
    manager.RegisterProvider("file", fileProvider)
    
    return manager, nil
}
EOF

# 创建配置示例
log_info "创建配置使用示例..."

cat > examples/config_example.go << 'EOF'
// Package main 配置使用示例
package main

import (
    "context"
    "log"
    
    "github.com/yfh-yun/moviepilot-go/config"
)

func main() {
    // 初始化配置管理器
    manager, err := config.Init()
    if err != nil {
        log.Fatal(err)
    }
    
    // 加载配置
    ctx := context.Background()
    appConfig, err := manager.Load(ctx, "config.yaml")
    if err != nil {
        log.Fatal(err)
    }
    
    // 使用配置
    if app, ok := appConfig["app"].(map[string]interface{}); ok {
        log.Printf("应用名称: %v", app["name"])
        log.Printf("应用版本: %v", app["version"])
        log.Printf("运行环境: %v", app["env"])
    }
    
    if server, ok := appConfig["server"].(map[string]interface{}); ok {
        log.Printf("服务端口: %v", server["port"])
    }
}
EOF

log_info "🎉 配置管理重构完成！"
log_info ""
log_info "新的配置结构："
log_info "config/"
log_info "├── core/           # 核心配置"
log_info "├── environments/   # 环境配置"
log_info "│   ├── dev/       # 开发环境"
log_info "│   ├── test/      # 测试环境"
log_info "│   └── prod/      # 生产环境"
log_info "├── validation/     # 配置验证"
log_info "│   ├── schemas/   # 验证模式"
log_info "│   └── rules/     # 验证规则"
log_info "└── providers/     # 配置提供者"
log_info "    ├── file/      # 文件提供者"
log_info "    ├── env/       # 环境变量提供者"
log_info "    └── vault/     # Vault提供者"
log_info ""
log_info "使用示例: go run examples/config_example.go"