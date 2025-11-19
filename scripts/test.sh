#!/bin/bash

# MoviePilot Go 测试脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 测试类型
TEST_TYPE=${1:-"all"}
COVERAGE=${COVERAGE:-"true"}

echo -e "${GREEN}开始运行测试...${NC}"
echo -e "${YELLOW}测试类型: ${TEST_TYPE}${NC}"
echo -e "${YELLOW}生成覆盖率: ${COVERAGE}${NC}"

# 清理测试缓存
echo -e "${YELLOW}清理测试缓存...${NC}"
go clean -testcache

# 创建覆盖率目录
mkdir -p coverage

case $TEST_TYPE in
    "unit")
        echo -e "${YELLOW}运行单元测试...${NC}"
        if [ "$COVERAGE" = "true" ]; then
            go test -v -coverprofile=coverage/unit.out ./internal/... ./pkg/...
            go tool cover -html=coverage/unit.out -o coverage/unit.html
            echo -e "${GREEN}单元测试覆盖率报告: coverage/unit.html${NC}"
        else
            go test -v ./internal/... ./pkg/...
        fi
        ;;
    "integration")
        echo -e "${YELLOW}运行集成测试...${NC}"
        if [ "$COVERAGE" = "true" ]; then
            go test -v -coverprofile=coverage/integration.out ./tests/integration/...
            go tool cover -html=coverage/integration.out -o coverage/integration.html
            echo -e "${GREEN}集成测试覆盖率报告: coverage/integration.html${NC}"
        else
            go test -v ./tests/integration/...
        fi
        ;;
    "e2e")
        echo -e "${YELLOW}运行端到端测试...${NC}"
        go test -v ./tests/e2e/...
        ;;
    "performance")
        echo -e "${YELLOW}运行性能测试...${NC}"
        go test -bench=. -benchmem -run=^$ ./internal/... ./pkg/... > coverage/benchmark.txt 2>&1
        echo -e "${GREEN}性能测试报告: coverage/benchmark.txt${NC}"
        ;;
    "all")
        echo -e "${YELLOW}运行所有测试...${NC}"
        if [ "$COVERAGE" = "true" ]; then
            go test -v -coverprofile=coverage/all.out ./...
            go tool cover -html=coverage/all.out -o coverage/all.html
            echo -e "${GREEN}总体覆盖率报告: coverage/all.html${NC}"
            
            # 生成覆盖率统计
            go test -cover ./... | grep -E "ok|FAIL" | tee coverage/coverage.txt
        else
            go test -v ./...
        fi
        ;;
    *)
        echo -e "${RED}未知的测试类型: ${TEST_TYPE}${NC}"
        echo "可用的测试类型: unit, integration, e2e, performance, all"
        exit 1
        ;;
esac

# 显示测试结果摘要
if [ "$COVERAGE" = "true" ]; then
    echo -e "${YELLOW}测试覆盖率摘要:${NC}"
    if [ -f "coverage/all.out" ]; then
        go tool cover -func=coverage/all.out | tail -1
    fi
fi

echo -e "${GREEN}测试完成!${NC}"