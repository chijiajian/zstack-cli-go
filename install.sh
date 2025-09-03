#!/bin/bash
set -e

REPO="chijiajian/zstack-cli-go"
BIN_NAME="zstack-cli-go"
INSTALL_NAME="zstack-cli"
INSTALL_DIR="/usr/local/bin"

COMPLETION_DIR_BASH="/etc/bash_completion.d"
COMPLETION_DIR_ZSH="$HOME/.zsh/completions"
COMPLETION_DIR_FISH="$HOME/.config/fish/completions"

OS=$(uname | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
esac

VERSION=${VERSION:-"v1.0.0"}
TARBALL="${BIN_NAME}_${VERSION}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${TARBALL}"

echo "Downloading $URL ..."
curl -sSL "$URL" -o "/tmp/${TARBALL}"

echo "Extracting..."
tar -xzf "/tmp/${TARBALL}" -C /tmp

# 找到真正的解压出来的二进制文件
EXTRACTED_BIN=$(find /tmp -maxdepth 1 -type f -name "${BIN_NAME}_*_${OS}_${ARCH}")
if [ -z "$EXTRACTED_BIN" ]; then
  echo "Error: extracted binary not found!"
  exit 1
fi

echo "Installing $EXTRACTED_BIN to $INSTALL_DIR/$INSTALL_NAME ..."
sudo install -m 0755 "$EXTRACTED_BIN" "$INSTALL_DIR/$INSTALL_NAME"

# 安装 bash 补全
if [ -d "$COMPLETION_DIR_BASH" ]; then
  echo "Installing bash completion..."
  sudo mkdir -p "$COMPLETION_DIR_BASH"
  sudo "$INSTALL_DIR/$INSTALL_NAME" completion bash | sudo tee "$COMPLETION_DIR_BASH/$INSTALL_NAME" >/dev/null
fi

# 安装 zsh 补全
mkdir -p "$COMPLETION_DIR_ZSH"
"$INSTALL_DIR/$INSTALL_NAME" completion zsh > "$COMPLETION_DIR_ZSH/_$INSTALL_NAME"

# 安装 fish 补全
mkdir -p "$COMPLETION_DIR_FISH"
"$INSTALL_DIR/$INSTALL_NAME" completion fish > "$COMPLETION_DIR_FISH/$INSTALL_NAME.fish"

echo "$INSTALL_NAME installation completed."
