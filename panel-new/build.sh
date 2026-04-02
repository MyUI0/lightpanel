#!/bin/bash
set -e

echo "[1/3] 初始化 Go 模块..."
go mod tidy

echo "[2/3] 编译项目..."
go build -o lightpanel -ldflags="-s -w" .

echo "[3/3] 完成!"
echo "运行: ./lightpanel"
