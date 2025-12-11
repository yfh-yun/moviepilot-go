#!/bin/bash
# MoviePilot 数据备份脚本

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

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 配置
BACKUP_DIR="${BACKUP_DIR:-/app/backups}"
CONTAINER_NAME="${CONTAINER_NAME:-moviepilot-postgres-prod}"
DB_NAME="${DB_NAME:-moviepilot}"
DB_USER="${DB_USER:-moviepilot}"
RETENTION_DAYS="${RETENTION_DAYS:-7}"

# 创建备份目录
mkdir -p "$BACKUP_DIR"

# 生成备份文件名
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/moviepilot_${TIMESTAMP}.sql.gz"

log_info "开始备份数据库..."
log_info "容器: $CONTAINER_NAME"
log_info "数据库: $DB_NAME"
log_info "备份文件: $BACKUP_FILE"

# 执行备份
if docker exec "$CONTAINER_NAME" pg_dump -U "$DB_USER" "$DB_NAME" | gzip > "$BACKUP_FILE"; then
    log_info "数据库备份成功: $BACKUP_FILE"
    
    # 显示备份文件大小
    BACKUP_SIZE=$(du -h "$BACKUP_FILE" | cut -f1)
    log_info "备份文件大小: $BACKUP_SIZE"
else
    log_error "数据库备份失败"
    exit 1
fi

# 清理旧备份
log_info "清理 ${RETENTION_DAYS} 天前的备份..."
find "$BACKUP_DIR" -name "moviepilot_*.sql.gz" -mtime +${RETENTION_DAYS} -delete

# 显示当前备份列表
log_info "当前备份列表:"
ls -lh "$BACKUP_DIR"/moviepilot_*.sql.gz 2>/dev/null || log_info "无备份文件"

log_info "备份完成！"
