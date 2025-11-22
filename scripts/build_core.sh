#!/bin/bash

echo "验证核心组件构建..."

echo "1. 检查主程序构建..."
if go build ./cmd/server > /dev/null 2>&1; then
    echo "✓ 主程序构建成功"
    SUCCESS_COUNT=1
else
    echo "✗ 主程序构建失败"
    echo "  构建失败详情:"
    go build ./cmd/server 2>&1 | head -10
    SUCCESS_COUNT=0
fi

echo ""
echo "构建结果: $SUCCESS_COUNT/1"

if [ $SUCCESS_COUNT -eq 1 ]; then
    echo "✓ 核心组件构建成功"
    exit 0
else
    echo "✗ 核心组件构建失败"
    exit 1
fi