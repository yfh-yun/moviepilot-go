#!/bin/bash

# Go版本兼容性检查脚本
echo "=== MoviePilot Go版本兼容性检查 ==="

# 检查当前Go版本
CURRENT_VERSION=$(go version | awk '{print $3}')
echo "当前Go版本: $CURRENT_VERSION"

# 检查项目要求的Go版本
if [ -f "go.mod" ]; then
    REQUIRED_VERSION=$(grep "^go " go.mod | awk '{print $2}')
    echo "项目要求Go版本: $REQUIRED_VERSION"
else
    echo "未找到go.mod文件"
    exit 1
fi

# 检查.go-version文件
if [ -f ".go-version" ]; then
    DOT_VERSION=$(cat .go-version)
    echo ".go-version文件: $DOT_VERSION"
fi

# 验证版本兼容性
echo "=== 验证兼容性 ==="
if command -v go &> /dev/null; then
    echo "✅ Go已安装"
    
    # 尝试编译检查
    echo "正在检查项目编译..."
    if go build -o /dev/null ./cmd/server 2>/dev/null; then
        echo "✅ 项目编译成功"
    else
        echo "❌ 项目编译失败"
        echo "运行 'go mod tidy' 来修复依赖问题"
    fi
else
    echo "❌ Go未安装"
    exit 1
fi

echo "=== 检查完成 ==="