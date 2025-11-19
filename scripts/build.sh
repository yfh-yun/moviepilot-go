#!/bin/bash

# MoviePilot Go 构建脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 项目信息
APP_NAME="moviepilot-go"
VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
BUILD_TIME=$(date +%Y-%m-%dT%H:%M:%S%z)
GO_VERSION=$(go version | awk '{print $3}')

# 构建标志
LDFLAGS="-ldflags \"-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GoVersion=${GO_VERSION}\""

# 输出目录
OUTPUT_DIR="bin"
mkdir -p ${OUTPUT_DIR}

echo -e "${GREEN}开始构建 ${APP_NAME}...${NC}"
echo -e "${YELLOW}版本: ${VERSION}${NC}"
echo -e "${YELLOW}构建时间: ${BUILD_TIME}${NC}"
echo -e "${YELLOW}Go版本: ${GO_VERSION}${NC}"

# 清理之前的构建
echo -e "${YELLOW}清理之前的构建...${NC}"
rm -f ${OUTPUT_DIR}/${APP_NAME}

# 运行测试
echo -e "${YELLOW}运行测试...${NC}"
go test -v ./...

# 代码检查
echo -e "${YELLOW}代码检查...${NC}"
go vet ./...
if command -v golangci-lint &> /dev/null; then
    golangci-lint run
fi

# 构建应用
echo -e "${YELLOW}构建应用...${NC}"
CGO_ENABLED=0 GOOS=linux go build ${LDFLAGS} -o ${OUTPUT_DIR}/${APP_NAME} cmd/server/main.go

# 检查构建结果
if [ -f "${OUTPUT_DIR}/${APP_NAME}" ]; then
    echo -e "${GREEN}构建成功!${NC}"
    echo -e "${GREEN}可执行文件: ${OUTPUT_DIR}/${APP_NAME}${NC}"
    ls -lh ${OUTPUT_DIR}/${APP_NAME}
else
    echo -e "${RED}构建失败!${NC}"
    exit 1
fi

# 生成校验和
echo -e "${YELLOW}生成校验和...${NC}"
cd ${OUTPUT_DIR}
sha256sum ${APP_NAME} > ${APP_NAME}.sha256
cd ..

echo -e "${GREEN}构建完成!${NC}"