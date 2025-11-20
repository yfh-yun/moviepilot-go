#!/bin/bash

# MoviePilot-Go 架构重构第二阶段脚本
# 合并冗余模块，简化目录结构

set -e

echo "🚀 开始 MoviePilot-Go 第二阶段重构..."

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
BACKUP_DIR="backup_phase2_$(date +%Y%m%d_%H%M%S)"
mkdir -p "../$BACKUP_DIR"
cp -r internal "../$BACKUP_DIR/"
log_info "备份完成: ../$BACKUP_DIR"

# 阶段1: 创建business目录结构
log_info "阶段1: 创建business目录结构..."

mkdir -p internal/business/{domains,services,workflows,policies}

# 阶段2: 移动services内容到business/services
log_info "阶段2: 移动services内容..."

if [ -d "internal/services" ]; then
    cp -r internal/services/* internal/business/services/
    log_info "services内容已复制到business/services"
fi

# 阶段3: 移动modules内容到business
log_info "阶段3: 处理modules内容..."

if [ -d "internal/modules" ]; then
    # 检查modules目录内容
    if [ "$(ls -A internal/modules)" ]; then
        # 将modules内容合并到business的相应子目录
        for module_dir in internal/modules/*; do
            if [ -d "$module_dir" ]; then
                module_name=$(basename "$module_dir")
                log_info "处理module: $module_name"
                
                # 根据内容类型决定移动位置
                if [ -f "$module_dir/service.go" ] || [ -f "$module_name/${module_name}_service.go" ]; then
                    # 服务相关，移动到business/services
                    cp -r "$module_dir" "internal/business/services/"
                elif [ -f "$module_dir/domain.go" ] || [ -f "$module_name/${module_name}_domain.go" ]; then
                    # 领域相关，移动到business/domains
                    cp -r "$module_dir" "internal/business/domains/"
                else
                    # 默认移动到business/services
                    cp -r "$module_dir" "internal/business/services/"
                fi
            fi
        done
        log_info "modules内容已分发到business子目录"
    else
        log_warn "modules目录为空"
    fi
fi

# 阶段4: 移动workflows内容到business/workflows
log_info "阶段4: 移动workflows内容..."

if [ -d "internal/workflows" ]; then
    cp -r internal/workflows/* internal/business/workflows/
    log_info "workflows内容已复制到business/workflows"
fi

# 阶段5: 创建infrastructure目录结构
log_info "阶段5: 创建infrastructure目录结构..."

mkdir -p internal/infrastructure/{config,security,events,context}

# 阶段6: 移动core内容到infrastructure
log_info "阶段6: 移动core内容..."

if [ -d "internal/core" ]; then
    # 移动配置相关
    if [ -f "internal/core/config.go" ] || [ -d "internal/core/config" ]; then
        cp -r internal/core/config* internal/infrastructure/config/ 2>/dev/null || true
    fi
    
    # 移动上下文相关
    if [ -f "internal/core/context.go" ] || [ -d "internal/core/context" ]; then
        cp -r internal/core/context* internal/infrastructure/context/ 2>/dev/null || true
    fi
    
    # 移动事件相关
    if [ -f "internal/core/event.go" ] || [ -d "internal/core/event" ]; then
        cp -r internal/core/event* internal/infrastructure/events/ 2>/dev/null || true
    fi
    
    # 移动安全相关
    if [ -f "internal/core/core.go" ]; then
        # 检查是否包含安全组件
        if grep -q "security\|JWT\|Password\|APIKey" internal/core/core.go; then
            cp internal/core/core.go internal/infrastructure/security/
        fi
    fi
    
    log_info "core内容已分发到infrastructure子目录"
fi

# 阶段7: 移动foundation内容到infrastructure
log_info "阶段7: 移动foundation内容..."

if [ -d "internal/foundation" ]; then
    cp -r internal/foundation/* internal/infrastructure/
    log_info "foundation内容已复制到infrastructure"
fi

# 阶段8: 更新导入路径（简化版）
log_info "阶段8: 更新导入路径..."

# 备份原始文件，然后批量更新
find . -name "*.go" -type f -not -path "./vendor/*" -not -path "./.git/*" -not -path "./backup_*" | while read -r file; do
    log_info "处理文件: $file"
    
    # 更新services路径
    sed -i 's|github\.com/yfh-yun/moviepilot-go/internal/services/|github.com/yfh-yun/moviepilot-go/internal/business/services/|g' "$file"
    
    # 更新workflows路径
    sed -i 's|github\.com/yfh-yun/moviepilot-go/internal/workflows/|github.com/yfh-yun/moviepilot-go/internal/business/workflows/|g' "$file"
    
    # 更新core路径
    sed -i 's|github\.com/yfh-yun/moviepilot-go/internal/core/|github.com/yfh-yun/moviepilot-go/internal/infrastructure/|g' "$file"
    
    # 更新foundation路径
    sed -i 's|github\.com/yfh-yun/moviepilot-go/internal/foundation/|github.com/yfh-yun/moviepilot-go/internal/infrastructure/|g' "$file"
done

# 阶段9: 清理旧目录（可选，先保留备份）
log_info "阶段9: 保留旧目录作为备份..."

# 创建清理脚本供后续使用
cat > cleanup_old_dirs.sh << 'EOF'
#!/bin/bash
echo "清理旧目录..."
rm -rf internal/services
rm -rf internal/modules  
rm -rf internal/workflows
rm -rf internal/core
rm -rf internal/foundation
echo "清理完成"
EOF

chmod +x cleanup_old_dirs.sh
log_info "清理脚本已创建: ./cleanup_old_dirs.sh"

log_info "🎉 第二阶段重构完成！"
log_info "请检查代码并运行测试确保一切正常"
log_info "备份位置: ../$BACKUP_DIR"
log_info "如需清理旧目录，请运行: ./cleanup_old_dirs.sh"