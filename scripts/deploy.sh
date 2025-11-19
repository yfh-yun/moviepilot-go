#!/bin/bash

# MoviePilot Go 部署脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 配置
APP_NAME="moviepilot-go"
DOCKER_REGISTRY=${DOCKER_REGISTRY:-"localhost:5000"}
VERSION=${VERSION:-"latest"}
ENVIRONMENT=${ENVIRONMENT:-"production"}

echo -e "${GREEN}开始部署 ${APP_NAME}...${NC}"
echo -e "${YELLOW}环境: ${ENVIRONMENT}${NC}"
echo -e "${YELLOW}版本: ${VERSION}${NC}"

# 构建Docker镜像
echo -e "${YELLOW}构建Docker镜像...${NC}"
docker build -t ${APP_NAME}:${VERSION} .
docker tag ${APP_NAME}:${VERSION} ${APP_NAME}:latest

# 推送到镜像仓库（如果配置了）
if [ "$DOCKER_REGISTRY" != "localhost:5000" ]; then
    echo -e "${YELLOW}推送到镜像仓库...${NC}"
    docker tag ${APP_NAME}:${VERSION} ${DOCKER_REGISTRY}/${APP_NAME}:${VERSION}
    docker tag ${APP_NAME}:latest ${DOCKER_REGISTRY}/${APP_NAME}:latest
    docker push ${DOCKER_REGISTRY}/${APP_NAME}:${VERSION}
    docker push ${DOCKER_REGISTRY}/${APP_NAME}:latest
fi

# 部署到不同环境
case $ENVIRONMENT in
    "dev")
        echo -e "${YELLOW}部署到开发环境...${NC}"
        docker-compose -f deployments/docker-compose.dev.yml down
        docker-compose -f deployments/docker-compose.dev.yml up -d
        ;;
    "prod")
        echo -e "${YELLOW}部署到生产环境...${NC}"
        docker-compose -f deployments/docker-compose.prod.yml down
        VERSION=${VERSION} docker-compose -f deployments/docker-compose.prod.yml up -d
        ;;
    *)
        echo -e "${YELLOW}使用默认配置部署...${NC}"
        docker-compose down
        docker-compose up -d
        ;;
esac

# 等待服务启动
echo -e "${YELLOW}等待服务启动...${NC}"
sleep 10

# 健康检查
echo -e "${YELLOW}执行健康检查...${NC}"
if curl -f http://localhost:3001/health > /dev/null 2>&1; then
    echo -e "${GREEN}部署成功! 服务运行正常${NC}"
else
    echo -e "${RED}部署失败! 服务健康检查失败${NC}"
    exit 1
fi

# 显示服务状态
echo -e "${YELLOW}服务状态:${NC}"
docker-compose ps

echo -e "${GREEN}部署完成!${NC}"