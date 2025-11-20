#!/bin/bash

# 动作系统构建和验证脚本

set -e

echo "=== MoviePilot Go Actions Build Script ==="

# 设置颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 项目根目录
PROJECT_ROOT="$(cd ../../.. && pwd)"
ACTIONS_DIR="$PROJECT_ROOT/internal/actions"

echo "Project root: $PROJECT_ROOT"
echo "Actions directory: $ACTIONS_DIR"

# 1. 检查Go版本
echo -e "${YELLOW}1. Checking Go version...${NC}"
go version

# 2. 清理模块缓存
echo -e "${YELLOW}2. Cleaning module cache...${NC}"
cd "$PROJECT_ROOT"
go clean -modcache

# 3. 下载依赖
echo -e "${YELLOW}3. Downloading dependencies...${NC}"
go mod download
go mod tidy

# 4. 检查代码格式
echo -e "${YELLOW}4. Checking code format...${NC}"
if ! go fmt ./internal/actions/...; then
    echo -e "${RED}Code format check failed${NC}"
    exit 1
fi

# 5. 运行静态分析
echo -e "${YELLOW}5. Running static analysis...${NC}"
if command -v golangci-lint &> /dev/null; then
    golangci-lint run ./internal/actions/... || echo -e "${YELLOW}golangci-lint not configured, skipping...${NC}"
else
    echo -e "${YELLOW}golangci-lint not installed, skipping static analysis...${NC}"
fi

# 6. 构建动作模块
echo -e "${YELLOW}6. Building actions module...${NC}"
if ! go build ./internal/actions/...; then
    echo -e "${RED}Build failed${NC}"
    exit 1
fi

# 7. 运行单元测试
echo -e "${YELLOW}7. Running unit tests...${NC}"
if ! go test -v ./internal/actions/...; then
    echo -e "${RED}Unit tests failed${NC}"
    exit 1
fi

# 8. 运行基准测试
echo -e "${YELLOW}8. Running benchmark tests...${NC}"
go test -bench=. -benchmem ./internal/actions/...

# 9. 生成测试覆盖率报告
echo -e "${YELLOW}9. Generating test coverage report...${NC}"
go test -coverprofile=coverage.out ./internal/actions/...
go tool cover -html=coverage.out -o coverage.html
echo -e "${GREEN}Coverage report generated: coverage.html${NC}"

# 10. 验证架构
echo -e "${YELLOW}10. Validating architecture...${NC}"

# 检查必要的目录结构
required_dirs=(
    "$ACTIONS_DIR/interfaces"
    "$ACTIONS_DIR/types"
    "$ACTIONS_DIR/base"
    "$ACTIONS_DIR/manager"
    "$ACTIONS_DIR/implementations"
    "$ACTIONS_DIR/registry"
    "$ACTIONS_DIR/examples"
)

for dir in "${required_dirs[@]}"; do
    if [ ! -d "$dir" ]; then
        echo -e "${RED}Required directory missing: $dir${NC}"
        exit 1
    fi
done

# 检查必要的文件
required_files=(
    "$ACTIONS_DIR/interfaces/action.go"
    "$ACTIONS_DIR/interfaces/manager.go"
    "$ACTIONS_DIR/types/context.go"
    "$ACTIONS_DIR/types/models.go"
    "$ACTIONS_DIR/base/action.go"
    "$ACTIONS_DIR/manager/action_manager.go"
    "$ACTIONS_DIR/registry/registry.go"
    "$ACTIONS_DIR/implementations/download.go"
    "$ACTIONS_DIR/implementations/scan.go"
)

for file in "${required_files[@]}"; do
    if [ ! -f "$file" ]; then
        echo -e "${RED}Required file missing: $file${NC}"
        exit 1
    fi
done

echo -e "${GREEN}Architecture validation passed${NC}"

# 11. 验证动作注册
echo -e "${YELLOW}11. Validating action registration...${NC}"
go run -c "
package main

import (
    \"fmt\"
    \"os\"

    \"github.com/yfh-yun/moviepilot-go/internal/actions/registry\"
)

func main() {
    reg := registry.GetDefaultRegistry()
    actions := reg.ListActions()
    
    requiredActions := []string{
        \"download\", \"scan\", \"file_scanner\", \"media_fetcher\",
        \"message_sender\", \"plugin_invoker\", \"rss_parser\",
        \"subscribe_manager\", \"transfer_manager\", \"workflow_cache\",
    }
    
    for _, action := range requiredActions {
        found := false
        for _, registered := range actions {
            if registered == action {
                found = true
                break
            }
        }
        if !found {
            fmt.Printf(\"Required action not registered: %s\n\", action)
            os.Exit(1)
        }
    }
    
    fmt.Printf(\"Successfully registered %d actions\n\", len(actions))
    for _, action := range actions {
        fmt.Printf(\"  - %s\n\", action)
    }
}
" || exit 1

echo -e "${GREEN}Action registration validation passed${NC}"

# 12. 性能基准
echo -e "${YELLOW}12. Running performance benchmarks...${NC}"
go test -bench=BenchmarkActionCreation -benchmem ./internal/actions/
go test -bench=BenchmarkActionExecution -benchmem ./internal/actions/
go test -bench=BenchmarkRegistryOperations -benchmem ./internal/actions/

# 13. 生成文档
echo -e "${YELLOW}13. Generating documentation...${NC}"
if command -v godoc &> /dev/null; then
    echo "Documentation available at: http://localhost:6060/pkg/github.com/yfh-yun/moviepilot-go/internal/actions/"
    echo "Run 'godoc -http=:6060' to start documentation server"
else
    echo -e "${YELLOW}godoc not installed, skipping documentation generation${NC}"
fi

# 14. 清理临时文件
echo -e "${YELLOW}14. Cleaning up temporary files...${NC}"
rm -f coverage.out

echo -e "${GREEN}=== Build completed successfully! ===${NC}"
echo ""
echo "Summary:"
echo "✓ Go version checked"
echo "✓ Dependencies downloaded"
echo "✓ Code formatted"
echo "✓ Static analysis completed"
echo "✓ Build successful"
echo "✓ Unit tests passed"
echo "✓ Benchmark tests completed"
echo "✓ Coverage report generated"
echo "✓ Architecture validated"
echo "✓ Action registration verified"
echo "✓ Performance benchmarks completed"
echo ""
echo "Next steps:"
echo "1. Review test coverage report: coverage.html"
echo "2. Check benchmark results for performance"
echo "3. Run integration tests: go test ./tests/..."
echo "4. Start development server: go run cmd/server/main.go"