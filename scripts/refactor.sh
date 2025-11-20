#!/bin/bash

# MoviePilot-Go 架构重构脚本
# 用于统一命名规范和简化目录结构

set -e

echo "🚀 开始 MoviePilot-Go 架构重构..."

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

# 创建备份
log_info "创建当前状态备份..."
BACKUP_DIR="backup_$(date +%Y%m%d_%H%M%S)"
mkdir -p "../$BACKUP_DIR"
cp -r internal "../$BACKUP_DIR/"
cp -r pkg "../$BACKUP_DIR/"
log_info "备份完成: ../$BACKUP_DIR"

# 阶段1: 统一命名规范
log_info "阶段1: 统一命名规范..."

# 重命名目录
declare -A RENAME_MAP=(
    ["internal/api"]="internal/apis"
    ["internal/service"]="internal/services"
    ["internal/model"]="internal/models"
    ["internal/repository"]="internal/repositories"
    ["internal/actions"]="internal/workflows"
    ["internal/scheduler"]="internal/schedulers"
)

for old_path in "${!RENAME_MAP[@]}"; do
    new_path="${RENAME_MAP[$old_path]}"
    if [ -d "$old_path" ]; then
        log_info "重命名: $old_path → $new_path"
        mv "$old_path" "$new_path"
    else
        log_warn "目录不存在: $old_path"
    fi
done

# 重命名API子目录
if [ -d "internal/apis/handlers/middleware" ]; then
    log_info "重命名: internal/apis/handlers/middleware → internal/apis/handlers/middlewares"
    mv "internal/apis/handlers/middleware" "internal/apis/handlers/middlewares"
fi

if [ -d "internal/apis/handlers/validator" ]; then
    log_info "重命名: internal/apis/handlers/validator → internal/apis/handlers/validators"
    mv "internal/apis/handlers/validator" "internal/apis/handlers/validators"
fi

log_info "阶段1完成: 命名规范统一"

# 阶段2: 更新导入路径
log_info "阶段2: 更新导入路径..."

# 查找所有Go文件并更新导入路径
find . -name "*.go" -type f -not -path "./vendor/*" -not -path "./.git/*" | while read -r file; do
    log_info "处理文件: $file"
    
    # 更新导入路径
    sed -i 's|github\.com/yfh-yun/moviepilot-go/internal/api/|github.com/yfh-yun/moviepilot-go/internal/apis/|g' "$file"
    sed -i 's|github\.com/yfh-yun/moviepilot-go/internal/service/|github.com/yfh-yun/moviepilot-go/internal/services/|g' "$file"
    sed -i 's|github\.com/yfh-yun/moviepilot-go/internal/model/|github.com/yfh-yun/moviepilot-go/internal/models/|g' "$file"
    sed -i 's|github\.com/yfh-yun/moviepilot-go/internal/repository/|github.com/yfh-yun/moviepilot-go/internal/repositories/|g' "$file"
    sed -i 's|github\.com/yfh-yun/moviepilot-go/internal/actions/|github.com/yfh-yun/moviepilot-go/internal/workflows/|g' "$file"
    sed -i 's|github\.com/yfh-yun/moviepilot-go/internal/scheduler/|github.com/yfh-yun/moviepilot-go/internal/schedulers/|g' "$file"
    
    # 更新middleware和validator的导入
    sed -i 's|github\.com/yfh-yun/moviepilot-go/internal/api/middleware/|github.com/yfh-yun/moviepilot-go/internal/api/middlewares/|g' "$file"
    sed -i 's|github\.com/yfh-yun/moviepilot-go/internal/api/validator/|github.com/yfh-yun/moviepilot-go/internal/api/validators/|g' "$file"
done

log_info "阶段2完成: 导入路径更新"

# 阶段3: 更新包声明
log_info "阶段3: 更新包声明..."

find internal -name "*.go" -type f | while read -r file; do
    # 更新package声明
    case "$file" in
        */apis/*)
            sed -i 's/^package api$/package apis/' "$file"
            ;;
        */services/*)
            sed -i 's/^package service$/package services/' "$file"
            ;;
        */models/*)
            sed -i 's/^package model$/package models/' "$file"
            ;;
        */repositories/*)
            sed -i 's/^package repository$/package repositories/' "$file"
            ;;
        */workflows/*)
            sed -i 's/^package actions$/package workflows/' "$file"
            ;;
        */schedulers/*)
            sed -i 's/^package scheduler$/package schedulers/' "$file"
            ;;
    esac
done

log_info "阶段3完成: 包声明更新"

# 阶段4: 验证编译
log_info "阶段4: 验证编译..."

if go mod tidy; then
    log_info "go mod tidy 成功"
else
    log_error "go mod tidy 失败"
    exit 1
fi

if go build ./cmd/server; then
    log_info "编译验证成功"
else
    log_error "编译验证失败"
    log_error "请检查导入路径和包声明"
    exit 1
fi

log_info "🎉 重构完成！"
log_info "请检查代码并运行测试确保一切正常"
log_info "备份位置: ../$BACKUP_DIR"