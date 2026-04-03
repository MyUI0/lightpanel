#!/bin/bash
set -e

echo "=== LightPanel 一键编译脚本 ==="

# 检查 Go
if ! command -v go &> /dev/null; then
    echo "[错误] 未检测到 Go，请先安装:"
    echo "  wget https://go.dev/dl/go1.22.5.linux-amd64.tar.gz"
    echo "  rm -rf /usr/local/go && tar -C /usr/local -xzf go1.22.5.linux-amd64.tar.gz"
    echo '  export PATH=$PATH:/usr/local/go/bin'
    exit 1
fi

echo "[OK] Go $(go version)"

echo "[1/4] 清理缓存..."
go clean -cache -modcache || true

echo "[2/4] 下载依赖..."
go mod tidy

echo "[3/4] 编译..."
go build -ldflags="-s -w" -o lightpanel .
chmod +x lightpanel

echo "[4/4] 验证..."
grep -q "X-Content-Type" handlers/app_create.go && echo "[OK] 源码最新" || echo "[WARN] 源码可能过期，请重新上传"

echo ""
echo "运行: ./lightpanel"
echo "访问: http://127.0.0.1:31956"
echo "账号: admin / admin"
