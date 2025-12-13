#!/bin/bash

# MoviePilot Go 测试运行脚本
# 用于快速运行各种测试套件

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 显示帮助信息
show_help() {
    echo "MoviePilot Go 测试运行脚本"
    echo ""
    echo "用法: $0 [选项] [测试类型]"
    echo ""
    echo "测试类型:"
    echo "  api         - 运行 API 测试"
    echo "  actions     - 运行 Actions 测试"
    echo "  business    - 运行 Business Services 测试"
    echo "  integration - 运行集成测试"
    echo "  all         - 运行所有测试"
    echo "  coverage    - 运行测试并生成覆盖率报告"
    echo ""
    echo "选项:"
    echo "  -v, --verbose   - 详细输出"
    echo "  -r, --race     - 启用竞态检测"
    echo "  -t, --timeout  - 设置测试超时时间 (默认: 30s)"
    echo "  -h, --help     - 显示帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 api                    # 运行 API 测试"
    echo "  $0 -v all                 # 详细模式运行所有测试"
    echo "  $0 -r -t 60s integration  # 启用竞态检测，60秒超时运行集成测试"
    echo "  $0 coverage               # 生成覆盖率报告"
}

# 默认参数
VERBOSE=""
RACE=""
TIMEOUT="30s"
TEST_TYPE=""

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--verbose)
            VERBOSE="-v"
            shift
            ;;
        -r|--race)
            RACE="-race"
            shift
            ;;
        -t|--timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        api|actions|business|integration|all|coverage)
            TEST_TYPE="$1"
            shift
            ;;
        *)
            print_error "未知参数: $1"
            show_help
            exit 1
            ;;
    esac
done

# 检查测试类型
if [[ -z "$TEST_TYPE" ]]; then
    print_error "请指定测试类型"
    show_help
    exit 1
fi

# 进入项目根目录
cd "$(dirname "$0")/.."
PROJECT_ROOT=$(pwd)

print_info "项目根目录: $PROJECT_ROOT"
print_info "测试类型: $TEST_TYPE"
print_info "测试超时: $TIMEOUT"
if [[ -n "$RACE" ]]; then
    print_info "竞态检测: 启用"
fi
if [[ -n "$VERBOSE" ]]; then
    print_info "详细输出: 启用"
fi

# 构建 go test 命令
GO_TEST="go test $RACE $VERBOSE -timeout $TIMEOUT"

# 运行测试
run_tests() {
    local test_path="$1"
    local test_name="$2"
    
    print_info "运行 $test_name 测试..."
    
    if eval "$GO_TEST $test_path"; then
        print_success "$test_name 测试通过"
    else
        print_error "$test_name 测试失败"
        return 1
    fi
}

# 根据测试类型执行相应的测试
case $TEST_TYPE in
    api)
        run_tests "./tests/api/..." "API"
        ;;
    actions)
        run_tests "./tests/actions/..." "Actions"
        ;;
    business)
        run_tests "./tests/business/..." "Business Services"
        ;;
    integration)
        run_tests "./tests/integration/..." "Integration"
        ;;
    all)
        print_info "运行所有测试套件..."
        
        # 按顺序运行各个测试套件
        if run_tests "./tests/api/..." "API" && \
           run_tests "./tests/actions/..." "Actions" && \
           run_tests "./tests/business/..." "Business Services" && \
           run_tests "./tests/integration/..." "Integration"; then
            print_success "所有测试通过!"
        else
            print_error "部分测试失败"
            exit 1
        fi
        ;;
    coverage)
        print_info "生成测试覆盖率报告..."
        
        # 创建覆盖率目录
        mkdir -p coverage
        
        # 运行测试并生成覆盖率
        COVERAGE_CMD="go test $RACE -coverprofile=coverage/coverage.out -covermode=atomic ./..."
        
        if eval "$COVERAGE_CMD"; then
            # 生成 HTML 报告
            go tool cover -html=coverage/coverage.out -o coverage/coverage.html
            
            # 显示覆盖率统计
            echo ""
            print_info "覆盖率统计:"
            go tool cover -func=coverage/coverage.out | tail -1
            
            print_success "覆盖率报告已生成: coverage/coverage.html"
        else
            print_error "覆盖率测试失败"
            exit 1
        fi
        ;;
    *)
        print_error "未知的测试类型: $TEST_TYPE"
        show_help
        exit 1
        ;;
esac

print_success "测试完成!"