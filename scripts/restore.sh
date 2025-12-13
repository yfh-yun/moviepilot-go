#!/bin/bash
# MoviePilot 数据恢复脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

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

# 配置
BACKUP_DIR="${BACKUP_DIR:-/app/backups}"
CONTAINER_NAME="${CONTAINER_NAME:-moviepilot-postgres-prod}"
DB_NAME="${DB_NAME:-moviepilot}"
DB_USER="${DB_USER:-moviepilot}"

# 检查参数
if [ $# -eq 0 ]; then
    log_error "请指定备份文件"
    echo "用法: $0 <备份文件路径>"
    echo ""
    echo "可用的备份文件:"
    ls -lh "$BACKUP_DIR"/moviepilot_*.sql.gz 2>/dev/null || echo "无备份文件"
    exit 1
fi

BACKUP_FILE="$1"

# 检查备份文件是否存在
if [ ! -f "$BACKUP_FILE" ]; then
    log_error "备份文件不存在: $BACKUP_FILE"
    exit 1
fi

log_warn "警告: 此操作将覆盖当前数据库！"
log_warn "数据库: $DB_NAME"
log_warn "备份文件: $BACKUP_FILE"
read -p "确认继续? (yes/no): " CONFIRM

if [ "$CONFIRM" != "yes" ]; then
    log_info "操作已取消"
    exit 0
fi

log_info "开始恢复数据库..."

# 停止应用服务（避免数据冲突）
log_info "停止应用服务..."
docker-compose -f deployments/docker-compose.prod.yml stop moviepilot-go python-plugins || true

# 删除现有数据库
log_info "删除现有数据库..."
docker exec "$CONTAINER_NAME" psql -U "$DB_USER" -c "DROP DATABASE IF EXISTS $DB_NAME;"

# 创建新数据库
log_info "创建新数据库..."
docker exec "$CONTAINER_NAME" psql -U "$DB_USER" -c "CREATE DATABASE $DB_NAME;"

# 恢复数据
log_info "恢复数据..."
if gunzip -c "$BACKUP_FILE" | docker exec -i "$CONTAINER_NAME" psql -U "$DB_USER" "$DB_NAME"; then
    log_info "数据库恢复成功"
else
    log_error "数据库恢复失败"
    exit 1
fi

# 重启应用服务
log_info "重启应用服务..."
docker-compose -f deployments/docker-compose.prod.yml start moviepilot-go python-plugins

log_info "恢复完成！"
