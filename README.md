# MoviePilot-Go

MoviePilot的Go语言版本实现，一个自动化媒体管理工具。

## 项目结构

```
moviepilot/
├── cmd/
│   └── main.go                          # Go 主入口
├── internal/
│   ├── config/                          # Viper 配置加载
│   ├── logger/                          # Zap 日志封装
│   ├── db/                              # GORM 数据库连接与迁移
│   ├── auth/                            # JWT 认证中间件
│   ├── api/                             # Gin 路由与控制器
│   ├── web/                             # 静态资源 + WebSocket
│   ├── scheduler/                       # robfig/cron 定时任务
│   ├── watcher/                         # fsnotify 文件监控
│   └── core/
│       ├── eventbus.go                  # 带 trace_id 的内存事件总线
│       └── plugin_engine.go             # 插件扫描、匹配、调用引擎
├── pkg/
│   ├── models/                          # GORM 模型 + DTO
│   ├── utils/                           # 工具函数（字符串、时间等）
│   └── client/
│       ├── worker_client.go             # gRPC 客户端（带重试/超时）
│       └── sandbox_client.go            # HTTP 客户端（retryablehttp）
├── python_services/
│   ├── worker/
│   │   ├── main.py                      # gRPC 服务实现
│   │   └── worker.proto                 # Protobuf 接口定义
│   └── plugin_sandbox/
│       ├── main.py                      # FastAPI 服务
│       └── plugin_loader.py             # 动态加载 plugins/*.py（支持 venv）
├── plugins/                             # 用户插件目录（运行时挂载，不打包）
│   └── example_notify/
│       ├── plugin.py                    # def on_event(event: dict) -> dict
│       ├── manifest.yaml                # name, events, permissions, version
│       └── requirements.txt             # 插件依赖（沙箱自动安装）
├── configs/
│   └── config.yaml                      # 主配置文件（数据库、路径、端口等）
├── deploy/
│   └── docker/
│       ├── go.Dockerfile                # 多阶段构建，非 root，Alpine
│       ├── python.Dockerfile            # 统一镜像，MODE=worker/sandbox
│       ├── start_python.sh              # 启动路由脚本
│       └── docker-compose.yml           # 三服务互联 + volume 挂载
```

## 技术栈

- Web框架: Gin
- ORM: GORM
- 配置管理: Viper
- 日志系统: Zap
- 定时任务: robfig/cron
- 文件监控: fsnotify
- 微服务通信: gRPC
- 插件系统: Python沙箱

## 功能特性

- 媒体信息识别和搜索
- 下载管理（支持多种下载器）
- 文件整理和重命名
- 订阅系统
- 通知系统
- 插件系统
- Web API接口

## 安装和运行

### 环境要求

- Go 1.19+
- Python 3.8+ (用于插件沙箱和工作节点)
- SQLite 3 (默认数据库)

### 安装步骤

```bash
# 克隆项目
git clone https://github.com/moviepilot/moviepilot-go.git

# 进入项目目录
cd moviepilot-go

# 安装Go依赖
go mod tidy

# 运行项目
go run cmd/moviepilot/main.go
```

## 配置

项目配置文件位于 `configs/config.yaml`，可以根据需要进行修改。

## 贡献

欢迎提交Issue和Pull Request来改进项目。

## 许可证

本项目仅供学习交流使用。