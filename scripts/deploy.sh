#!/bin/bash
# MoviePilot Docker部署脚本

set -e

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

# 检查命令是否存在
check_command() {
    if ! command -v $1 &> /dev/null; then
        log_error "$1 未安装，请先安装"
        exit 1
    fi
}

# 检查必要的命令
log_info "检查必要的命令..."
check_command docker
check_command docker-compose

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DEPLOY_DIR="$PROJECT_ROOT/deployments"

# 切换到项目根目录
cd "$PROJECT_ROOT"

# 解析参数
ENV="prod"
ACTION="up"
BUILD=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -e|--env)
            ENV="$2"
            shift 2
            ;;
        -a|--action)
            ACTION="$2"
            shift 2
            ;;
        -b|--build)
            BUILD=true
            shift
            ;;
        -h|--help)
            echo "用法: $0 [选项]"
            echo ""
            echo "选项:"
            echo "  -e, --env ENV        环境 (dev/prod, 默认: prod)"
            echo "  -a, --action ACTION  操作 (up/down/restart/logs, 默认: up)"
            echo "  -b, --build          构建镜像"
            echo "  -h, --help           显示帮助信息"
            exit 0
            ;;
        *)
            log_error "未知参数: $1"
            exit 1
            ;;
    esac
done

# 设置compose文件
if [ "$ENV" = "dev" ]; then
    COMPOSE_FILE="$DEPLOY_DIR/docker-compose.dev.yml"
    log_info "使用开发环境配置"
elif [ "$ENV" = "prod" ]; then
    COMPOSE_FILE="$DEPLOY_DIR/docker-compose.prod.yml"
    log_info "使用生产环境配置"
else
    COMPOSE_FILE="$DEPLOY_DIR/docker-compose.yml"
    log_info "使用默认配置"
fi

# 检查配置文件是否存在
if [ ! -f "$COMPOSE_FILE" ]; then
    log_error "配置文件不存在: $COMPOSE_FILE"
    exit 1
fi

# 检查.env文件
if [ "$ENV" = "prod" ] && [ ! -f "$DEPLOY_DIR/.env" ]; then
    log_warn ".env文件不存在，将从.env.example复制"
    if [ -f "$DEPLOY_DIR/.env.example" ]; then
        cp "$DEPLOY_DIR/.env.example" "$DEPLOY_DIR/.env"
        log_warn "请编辑 $DEPLOY_DIR/.env 文件，配置必要的环境变量"
        exit 1
    else
        log_error ".env.example文件不存在"
        exit 1
    fi
fi

# 创建必要的目录
log_info "创建必要的目录..."
mkdir -p "$PROJECT_ROOT/logs"
mkdir -p "$PROJECT_ROOT/data"
mkdir -p "$PROJECT_ROOT/configs"

# 执行操作
case $ACTION in
    up)
        log_info "启动服务..."
        if [ "$BUILD" = true ]; then
            docker-compose -f "$COMPOSE_FILE" up -d --build
        else
            docker-compose -f "$COMPOSE_FILE" up -d
        fi
        log_info "服务启动成功！"
        log_info "查看日志: docker-compose -f $COMPOSE_FILE logs -f"
        ;;
    down)
        log_info "停止服务..."
        docker-compose -f "$COMPOSE_FILE" down
        log_info "服务已停止"
        ;;
    restart)
        log_info "重启服务..."
        docker-compose -f "$COMPOSE_FILE" restart
        log_info "服务已重启"
        ;;
    logs)
        log_info "查看日志..."
        docker-compose -f "$COMPOSE_FILE" logs -f
        ;;
    build)
        log_info "构建镜像..."
        docker-compose -f "$COMPOSE_FILE" build
        log_info "镜像构建完成"
        ;;
    ps)
        log_info "查看服务状态..."
        docker-compose -f "$COMPOSE_FILE" ps
        ;;
    *)
        log_error "未知操作: $ACTION"
        log_info "支持的操作: up, down, restart, logs, build, ps"
        exit 1
        ;;
esac

# 显示服务访问地址
if [ "$ACTION" = "up" ]; then
    echo ""
    log_info "服务访问地址:"
    echo "  - MoviePilot API: http://localhost:3001"
    echo "  - Prometheus: http://localhost:9090"
    echo "  - Grafana: http://localhost:3000 (admin/admin123)"
    if [ "$ENV" = "dev" ]; then
        echo "  - pgAdmin: http://localhost:8080 (admin@moviepilot.com/admin123)"
        echo "  - Redis Commander: http://localhost:8081 (admin/admin123)"
    fi
    echo ""
fi
