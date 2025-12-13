#!/bin/bash

# 优化后的测试运行脚本
# 解决集成测试超时问题

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 获取脚本目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

# 显示帮助信息
show_help() {
    echo "MoviePilot Go 优化测试运行脚本"
    echo ""
    echo "用法: $0 [选项] [测试类型]"
    echo ""
    echo "测试类型:"
    echo "  all          运行所有测试（默认）"
    echo "  api          运行API测试"
    echo "  actions      运行Actions测试"
    echo "  business     运行Business测试"
    echo "  integration  运行集成测试（优化超时）"
    echo "  coverage     生成覆盖率报告"
    echo "  quick        快速测试（跳过集成测试）"
    echo ""
    echo "选项:"
    echo "  -v, --verbose   详细输出"
    echo "  -t, --timeout   设置超时时间（秒）"
    echo "  -r, --race      启用竞态检测"
    echo "  -c, --clean     清理测试缓存"
    echo "  -h, --help      显示帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 quick                    # 快速测试"
    echo "  $0 integration -t 60        # 集成测试，60秒超时"
    echo "  $0 all -v -r               # 所有测试，详细输出，竞态检测"
}

# 默认参数
TEST_TYPE="all"
VERBOSE=false
TIMEOUT=30
RACE=false
CLEAN=false

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -t|--timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        -r|--race)
            RACE=true
            shift
            ;;
        -c|--clean)
            CLEAN=true
            shift
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        all|api|actions|business|integration|coverage|quick)
            TEST_TYPE="$1"
            shift
            ;;
        *)
            log_error "未知参数: $1"
            show_help
            exit 1
            ;;
    esac
done

# 清理测试缓存
if [ "$CLEAN" = true ]; then
    log_info "清理测试缓存..."
    go clean -testcache
    log_success "测试缓存已清理"
fi

# 构建测试参数
TEST_ARGS=""
if [ "$VERBOSE" = true ]; then
    TEST_ARGS="$TEST_ARGS -v"
fi

if [ "$RACE" = true ]; then
    TEST_ARGS="$TEST_ARGS -race"
fi

# 设置超时
TEST_TIMEOUT="$TIMEOUT"
if [ "$TEST_TYPE" = "integration" ]; then
    # 集成测试需要更长的超时时间
    TEST_TIMEOUT=$((TIMEOUT * 2))
    log_warning "集成测试超时设置为 ${TEST_TIMEOUT}s"
fi

# 函数：运行测试并检查结果
run_test() {
    local test_path="$1"
    local test_name="$2"
    local custom_timeout="$3"
    
    if [ -z "$custom_timeout" ]; then
        custom_timeout="$TEST_TIMEOUT"
    fi
    
    log_info "运行 $test_name 测试..."
    
    # 使用timeout命令控制超时
    if timeout "$custom_timeout"s go test $TEST_ARGS "$test_path" -timeout="${custom_timeout}s"; then
        log_success "$test_name 测试通过"
        return 0
    else
        local exit_code=$?
        if [ $exit_code -eq 124 ]; then
            log_error "$test_name 测试超时 (${custom_timeout}s)"
        else
            log_error "$test_name 测试失败 (退出码: $exit_code)"
        fi
        return $exit_code
    fi
}

# 函数：运行快速测试（跳过慢速测试）
run_quick_test() {
    log_info "运行快速测试（跳过集成测试）..."
    
    local failed=false
    
    # API测试
    if ! run_test "./tests/api/..." "API" 30; then
        failed=true
    fi
    
    # Actions测试
    if ! run_test "./tests/actions/..." "Actions" 30; then
        failed=true
    fi
    
    # Business测试
    if ! run_test "./tests/business/..." "Business" 30; then
        failed=true
    fi
    
    if [ "$failed" = true ]; then
        log_error "快速测试失败"
        exit 1
    else
        log_success "快速测试全部通过"
    fi
}

# 函数：运行集成测试（优化版）
run_integration_test() {
    log_info "运行集成测试（优化版）..."
    
    # 设置环境变量以优化测试性能
    export GOMAXPROCS=2  # 限制CPU核心数，避免过度并行
    export CGO_ENABLED=0  # 禁用CGO，加速编译
    
    # 运行集成测试
    if run_test "./tests/integration/..." "Integration" "$TEST_TIMEOUT"; then
        log_success "集成测试通过"
    else
        log_error "集成测试失败"
        exit 1
    fi
    
    # 恢复环境变量
    unset GOMAXPROCS
    unset CGO_ENABLED
}

# 函数：生成覆盖率报告
generate_coverage() {
    log_info "生成覆盖率报告..."
    
    # 创建覆盖率目录
    mkdir -p coverage
    
    # 运行测试并生成覆盖率
    go test -v -race -coverprofile=coverage/coverage.out -timeout=60s ./tests/... 2>&1
    
    if [ $? -eq 0 ]; then
        # 生成HTML报告
        go tool cover -html=coverage/coverage.out -o coverage/coverage.html
        
        # 显示覆盖率统计
        echo ""
        log_info "覆盖率统计:"
        go tool cover -func=coverage/coverage.out | tail -1
        
        log_success "覆盖率报告已生成: coverage/coverage.html"
    else
        log_error "覆盖率生成失败"
        exit 1
    fi
}

# 主执行逻辑
log_info "开始运行 MoviePilot Go 测试..."
log_info "测试类型: $TEST_TYPE"
log_info "超时设置: ${TIMEOUT}s"
log_info "详细输出: $VERBOSE"
log_info "竞态检测: $RACE"

# 检查Go环境
if ! command -v go &> /dev/null; then
    log_error "Go 未安装或不在PATH中"
    exit 1
fi

# 检查项目结构
if [ ! -f "go.mod" ]; then
    log_error "不在Go项目根目录中"
    exit 1
fi

# 下载依赖
log_info "检查依赖..."
go mod download
go mod verify

# 根据测试类型执行
case $TEST_TYPE in
    all)
        log_info "运行所有测试..."
        
        # 快速测试
        run_quick_test
        
        # 集成测试
        run_integration_test
        
        log_success "所有测试通过！"
        ;;
    api)
        run_test "./tests/api/..." "API"
        ;;
    actions)
        run_test "./tests/actions/..." "Actions"
        ;;
    business)
        run_test "./tests/business/..." "Business"
        ;;
    integration)
        run_integration_test
        ;;
    coverage)
        generate_coverage
        ;;
    quick)
        run_quick_test
        ;;
    *)
        log_error "未知测试类型: $TEST_TYPE"
        show_help
        exit 1
        ;;
esac

log_success "测试完成！"