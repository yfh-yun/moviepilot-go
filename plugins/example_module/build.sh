#!/bin/bash

# 构建插件
echo "构建示例插件..."
go build -buildmode=plugin -o example_module.so module.go

if [ $? -eq 0 ]; then
    echo "插件构建成功: example_module.so"
else
    echo "插件构建失败"
    exit 1
fi