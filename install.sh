#!/usr/bin/env bash
set -euo pipefail

# Downloads only published artifacts; does not publish or configure a service.
VERSION=${1:-latest}
REPO="MyUI0/lightpanel"
INSTALL_DIR=${LIGHTPANEL_INSTALL_DIR:-$PWD}
if [[ "$(uname -s)" != "Linux" ]]; then echo "仅支持 Linux" >&2; exit 1; fi
case "$(uname -m)" in
  x86_64) ARCH_NAME=amd64 ;;
  aarch64|arm64) ARCH_NAME=arm64 ;;
  *) echo "不支持此架构" >&2; exit 1 ;;
esac
if [[ "$VERSION" == latest ]]; then
  resolved=$(curl -fsSL --proto '=https' --proto-redir '=https' -o /dev/null -w '%{url_effective}' "https://github.com/$REPO/releases/latest")
  VERSION=${resolved##*/}
fi
if [[ ! "$VERSION" =~ ^v[0-9][A-Za-z0-9._-]*$ ]]; then echo "版本号无效" >&2; exit 1; fi
PKG="lightpanel-$VERSION-linux-$ARCH_NAME"
URL="https://github.com/$REPO/releases/download/$VERSION"
temp_dir=$(mktemp -d)
trap 'rm -rf -- "$temp_dir"' EXIT
curl -fsSL --proto '=https' --proto-redir '=https' "$URL/$PKG.tar.gz" -o "$temp_dir/$PKG.tar.gz"
curl -fsSL --proto '=https' --proto-redir '=https' "$URL/$PKG.tar.gz.sha256" -o "$temp_dir/$PKG.tar.gz.sha256"
expected=$(awk 'NR==1{print $1}' "$temp_dir/$PKG.tar.gz.sha256")
if [[ ! "$expected" =~ ^[a-fA-F0-9]{64}$ ]]; then echo "校验文件无效" >&2; exit 1; fi
actual=$(sha256sum "$temp_dir/$PKG.tar.gz"); actual=${actual%% *}
if [[ "$expected" != "$actual" ]]; then echo "校验失败" >&2; exit 1; fi
tar -xOzf "$temp_dir/$PKG.tar.gz" "$PKG/lightpanel" > "$temp_dir/lightpanel"
mkdir -p -- "$INSTALL_DIR"
install -m 0755 "$temp_dir/lightpanel" "$INSTALL_DIR/lightpanel.new"
mv -f -- "$INSTALL_DIR/lightpanel.new" "$INSTALL_DIR/lightpanel"
echo "已安装到 $INSTALL_DIR/lightpanel"
echo "请停止旧面板和受管应用后升级；默认前台运行，仅监听 127.0.0.1:31956。"
echo "初次运行后，请从 data/config/initial-password.txt 读取初始密码。"
