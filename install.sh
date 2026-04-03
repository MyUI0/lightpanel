#!/bin/bash
set -e

# LightPanel 一键安装脚本
# 用法: curl -L https://raw.githubusercontent.com/你的用户名/lightpanel/main/install.sh | bash -s v3.0.1

VERSION=${1:-latest}

# 检测架构
ARCH=$(uname -m)
case $ARCH in
    x86_64)
        ARCH_NAME="amd64"
        ;;
    aarch64)
        ARCH_NAME="arm64"
        ;;
    armv7l)
        ARCH_NAME="armv7"
        ;;
    *)
        echo "❌ 不支持的架构: $ARCH"
        exit 1
        ;;
esac

echo "🖥️  检测到架构: $ARCH ($ARCH_NAME)"

# 确定下载链接
REPO="https://github.com/你的用户名/lightpanel"
FILENAME="lightpanel-${VERSION}-linux-${ARCH_NAME}.tar.gz"
URL="${REPO}/releases/download/${VERSION}/${FILENAME}"

echo "📦 正在下载 LightPanel ${VERSION}..."
curl -L "$URL" -o "${FILENAME}"

echo "📂 正在解压..."
tar -xzf "${FILENAME}"
chmod +x lightpanel
rm -f "${FILENAME}"

echo ""
echo "✅ 安装完成！"
echo ""
echo "📝 使用方法:"
echo "   启动面板: ./lightpanel"
echo "   访问地址: http://127.0.0.1:31956"
echo "   默认账号: admin / admin"
echo ""
echo "⚙️  首次运行会自动创建数据目录"
echo ""