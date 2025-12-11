#!/bin/bash
# MoviePilot 健康检查脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 日志函数
log_info() {
    echo -e "${GREEN}[✓]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[!]${NC} $1"
}

log_error() {
    echo -e "${RED}[✗]${NC} $1"
}

# 配置
GO_SERVICE_URL="${GO_SERVICE_URL:-http://localhost:3001}"
PYTHON_SERVICE_URL="${PYTHON_SERVICE_URL:-http://localhost:5000}"
PROMETHEUS_URL="${PROMETHEUS_URL:-http://localhost:9090}"
GRAFANA_URL="${GRAFANA_URL:-http://localhost:3000}"

echo "========================================"
echo "MoviePilot 健康检查"
echo "========================================"
echo ""

# 检查Docker服务
echo "检查Docker服务..."
if docker ps &> /dev/null; then
    log_info "Docker服务运行正常"
else
    log_error "Docker服务未运行"
    exit 1
fi
echo ""

# 检查容器状态
echo "检查容器状态..."
CONTAINERS=(
    "moviepilot-go"
    "moviepilot-python-plugins"
    "moviepilot-postgres"
    "moviepilot-redis"
)

ALL_RUNNING=true
for container in "${CONTAINERS[@]}"; do
    if docker ps --format '{{.Names}}' | grep -q "^${container}"; then
        STATUS=$(docker inspect --format='{{.State.Health.Status}}' "$container" 2>/dev/null || echo "unknown")
        if [ "$STATUS" = "healthy" ] || [ "$STATUS" = "unknown" ]; then
            log_info "$container: 运行中"
        else
            log_warn "$container: 不健康 ($STATUS)"
            ALL_RUNNING=false
        fi
    else
        log_error "$container: 未运行"
        ALL_RUNNING=false
    fi
done
echo ""

# 检查Go服务
echo "检查Go服务..."
if curl -sf "$GO_SERVICE_URL/health" > /dev/null 2>&1; then
    log_info "Go服务响应正常"
else
    log_error "Go服务无响应"
    ALL_RUNNING=false
fi
echo ""

# 检查Python插件服务
echo "检查Python插件服务..."
if curl -sf "$PYTHON_SERVICE_URL/health" > /dev/null 2>&1; then
    log_info "Python插件服务响应正常"
else
    log_warn "Python插件服务无响应（可能正在启动）"
fi
echo ""

# 检查Prometheus
echo "检查Prometheus..."
if curl -sf "$PROMETHEUS_URL/-/healthy" > /dev/null 2>&1; then
    log_info "Prometheus运行正常"
else
    log_warn "Prometheus无响应"
fi
echo ""

# 检查Grafana
echo "检查Grafana..."
if curl -sf "$GRAFANA_URL/api/health" > /dev/null 2>&1; then
    log_info "Grafana运行正常"
else
    log_warn "Grafana无响应"
fi
echo ""

# 检查磁盘空间
echo "检查磁盘空间..."
DISK_USAGE=$(df -h . | awk 'NR==2 {print $5}' | sed 's/%//')
if [ "$DISK_USAGE" -lt 80 ]; then
    log_info "磁盘空间充足 (已使用: ${DISK_USAGE}%)"
elif [ "$DISK_USAGE" -lt 90 ]; then
    log_warn "磁盘空间不足 (已使用: ${DISK_USAGE}%)"
else
    log_error "磁盘空间严重不足 (已使用: ${DISK_USAGE}%)"
    ALL_RUNNING=false
fi
echo ""

# 总结
echo "========================================"
if [ "$ALL_RUNNING" = true ]; then
    log_info "所有服务运行正常"
    exit 0
else
    log_error "部分服务存在问题"
    exit 1
fi
