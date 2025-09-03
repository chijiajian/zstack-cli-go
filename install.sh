#!/bin/bash
set -e

# -------------------- 配置 --------------------
BIN_NAME="zstack-cli"
GITHUB_REPO="chijiajian/zstack-cli-go"
VERSION=${1:-$(curl -s "https://api.github.com/repos/$GITHUB_REPO/releases/latest" | grep -Po '"tag_name": "\K.*?(?=")')}
OS=$(uname | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
TMP_DIR=$(mktemp -d)
INSTALL_DIR="/usr/local/bin"

# 补全目录
COMPLETION_DIR_BASH="/etc/bash_completion.d"
COMPLETION_DIR_ZSH="$HOME/.zsh/completions"
COMPLETION_DIR_FISH="$HOME/.config/fish/completions"

echo "Installing $BIN_NAME version $VERSION for $OS/$ARCH..."

# -------------------- 构建下载 URL --------------------
case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64 | arm64) ARCH="arm64" ;;
  armv7*) ARCH="arm" ;;
  i386 | i686) ARCH="386" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

DOWNLOAD_URL="https://github.com/$GITHUB_REPO/releases/download/$VERSION/${BIN_NAME}-go_${VERSION}_${OS}_${ARCH}.tar.gz"

# -------------------- 下载 --------------------
echo "Downloading $DOWNLOAD_URL ..."
curl -L -o "$TMP_DIR/${BIN_NAME}.tar.gz" "$DOWNLOAD_URL"

# -------------------- 解压 --------------------
echo "Extracting..."
tar -xzf "$TMP_DIR/${BIN_NAME}.tar.gz" -C "$TMP_DIR"

# 假设解压出来的文件就是二进制
BIN_PATH=$(find "$TMP_DIR" -maxdepth 1 -type f -name "${BIN_NAME}-go_${VERSION}_${OS}_${ARCH}" | head -n1)
if [ ! -f "$BIN_PATH" ]; then
    echo "Binary not found after extraction!"
    exit 1
fi

# -------------------- 安装 --------------------
echo "Installing $BIN_NAME to $INSTALL_DIR/$BIN_NAME ..."
sudo install -m 0755 "$BIN_PATH" "$INSTALL_DIR/$BIN_NAME"

# -------------------- 安装补全 --------------------
# Bash
if [ -d "$COMPLETION_DIR_BASH" ]; then
    sudo mkdir -p "$COMPLETION_DIR_BASH"
    sudo "$INSTALL_DIR/$BIN_NAME" completion bash > "$COMPLETION_DIR_BASH/$BIN_NAME"
    # 永久生效
    grep -qxF "source $COMPLETION_DIR_BASH/$BIN_NAME" "$HOME/.bashrc" || \
        echo "source $COMPLETION_DIR_BASH/$BIN_NAME" >> "$HOME/.bashrc"
    echo "Bash completion installed and will be active in new sessions."
    # 立即生效
    source "$COMPLETION_DIR_BASH/$BIN_NAME" || true
fi

# Zsh
mkdir -p "$COMPLETION_DIR_ZSH"
"$INSTALL_DIR/$BIN_NAME" completion zsh > "$COMPLETION_DIR_ZSH/_$BIN_NAME"
grep -qxF "fpath+=($COMPLETION_DIR_ZSH)" "$HOME/.zshrc" || \
    echo "fpath+=($COMPLETION_DIR_ZSH)" >> "$HOME/.zshrc"
echo "Zsh completion installed."

# Fish
mkdir -p "$COMPLETION_DIR_FISH"
"$INSTALL_DIR/$BIN_NAME" completion fish > "$COMPLETION_DIR_FISH/$BIN_NAME.fish"
echo "Fish completion installed."

# -------------------- 清理 --------------------
rm -rf "$TMP_DIR"

echo "$BIN_NAME installation completed!"
echo "Restart your shell or source your ~/.bashrc / ~/.zshrc to enable completions."
