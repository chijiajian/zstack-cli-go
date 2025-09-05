#!/bin/bash
set -e

# --- 变量定义 ---
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

# --- 主脚本 ---
echo "正在安装 $BIN_NAME version $VERSION for $OS/$ARCH..."

# 1. 架构映射
case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64 | arm64) ARCH="arm64" ;;
  armv7*) ARCH="arm" ;;
  i386 | i686) ARCH="386" ;;
  *) echo "不支持的架构: $ARCH"; exit 1 ;;
esac

# 2. 下载二进制包
DOWNLOAD_URL="https://github.com/$GITHUB_REPO/releases/download/$VERSION/${BIN_NAME}-go_${VERSION}_${OS}_${ARCH}.tar.gz"
echo "正在下载 $DOWNLOAD_URL ..."
if ! curl -L -o "$TMP_DIR/${BIN_NAME}.tar.gz" "$DOWNLOAD_URL"; then
    echo "错误: 下载失败，请检查网络或 URL。终止安装。"
    rm -rf "$TMP_DIR"
    exit 1
fi

# 3. 解压
echo "正在解压..."
if ! tar -xzf "$TMP_DIR/${BIN_NAME}.tar.gz" -C "$TMP_DIR"; then
    echo "错误: 解压失败。终止安装。"
    rm -rf "$TMP_DIR"
    exit 1
fi

# 查找二进制文件
BIN_PATH=$(find "$TMP_DIR" -maxdepth 1 -type f -name "${BIN_NAME}-go_${VERSION}_${OS}_${ARCH}" | head -n1)
if [ ! -f "$BIN_PATH" ]; then
    echo "错误: 解压后未找到二进制文件。终止安装。"
    rm -rf "$TMP_DIR"
    exit 1
fi

# 4. 安装二进制文件到 /usr/local/bin
echo "正在安装 $BIN_NAME 到 $INSTALL_DIR/$BIN_NAME ..."
sudo install -m 0755 "$BIN_PATH" "$INSTALL_DIR/$BIN_NAME"

# 5. 安装 Shell 补全
echo "正在安装 Shell 补全脚本..."

CURRENT_SHELL=$(basename "$SHELL")

case "$CURRENT_SHELL" in
  "bash")
    # Bash 补全 (优先系统级，若失败则回退到用户级)
    echo "--> 正在为 Bash 安装补全..."
    if sudo test -w "$COMPLETION_DIR_BASH" || sudo mkdir -p "$COMPLETION_DIR_BASH" >/dev/null 2>&1; then
        sudo "$INSTALL_DIR/$BIN_NAME" completion bash > "$COMPLETION_DIR_BASH/$BIN_NAME"
        echo "Bash 补全已安装到系统目录: $COMPLETION_DIR_BASH"
    else
        # 回退到用户家目录
        COMPLETION_DIR_USER="$HOME/.bash_completion.d"
        mkdir -p "$COMPLETION_DIR_USER"
        "$INSTALL_DIR/$BIN_NAME" completion bash > "$COMPLETION_DIR_USER/$BIN_NAME"
        if ! grep -qxF "source $COMPLETION_DIR_USER/$BIN_NAME" "$HOME/.bashrc"; then
            echo "source $COMPLETION_DIR_USER/$BIN_NAME" >> "$HOME/.bashrc"
        fi
        echo "Bash 补全已安装到用户目录: $COMPLETION_DIR_USER"
    fi
    ;;
  "zsh")
    # Zsh 补全
    echo "--> 正在为 Zsh 安装补全..."
    mkdir -p "$COMPLETION_DIR_ZSH"
    "$INSTALL_DIR/$BIN_NAME" completion zsh > "$COMPLETION_DIR_ZSH/_$BIN_NAME"
    if ! grep -qxF "fpath+=($COMPLETION_DIR_ZSH)" "$HOME/.zshrc"; then
        echo "fpath+=($COMPLETION_DIR_ZSH)" >> "$HOME/.zshrc"
    fi
    echo "Zsh 补全已安装到 $COMPLETION_DIR_ZSH"
    ;;
  "fish")
    # Fish 补全
    echo "--> 正在为 Fish 安装补全..."
    mkdir -p "$COMPLETION_DIR_FISH"
    "$INSTALL_DIR/$BIN_NAME" completion fish > "$COMPLETION_DIR_FISH/$BIN_NAME.fish"
    echo "Fish 补全已安装到 $COMPLETION_DIR_FISH"
    ;;
  *)
    echo "警告: 未知的 Shell 类型 ($CURRENT_SHELL)。跳过 Shell 补全安装。"
    ;;
esac

# 6. 清理临时文件
rm -rf "$TMP_DIR"

# --- 安装完成提示 ---
echo "---"
echo "恭喜！$BIN_NAME 已成功安装！"
echo "请重启你的 Shell 或执行以下命令来启用补全功能:"

case "$CURRENT_SHELL" in
  "bash")
    echo "  - Bash: source ~/.bashrc"
    ;;
  "zsh")
    echo "  - Zsh:  source ~/.zshrc"
    ;;
  "fish")
    echo "  - Fish: source ~/.config/fish/config.fish"
    ;;
esac

echo "---"