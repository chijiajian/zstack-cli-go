#!/bin/bash
set -e

BIN_NAME="zstack-cli"
COMPLETION_DIR_BASH="/etc/bash_completion.d"
COMPLETION_DIR_ZSH="$HOME/.zsh/completions"
COMPLETION_DIR_FISH="$HOME/.config/fish/completions"

OS=$(uname | tr '[:upper:]' '[:lower:]')

# 检查 Go module
if [ ! -f "go.mod" ]; then
  echo "go.mod not found. Please run 'go mod init' or ensure you're in the project root."
  exit 1
fi

echo "Building $BIN_NAME..."
go build -o "$BIN_NAME" main.go

case "$OS" in
  linux|darwin)
    echo "Installing $BIN_NAME to /usr/local/bin..."
    sudo install -m 0755 "$BIN_NAME" /usr/local/bin/

    # 安装 bash 补全
    if [ -d "$COMPLETION_DIR_BASH" ]; then
      echo "Installing bash completion..."
      sudo mkdir -p "$COMPLETION_DIR_BASH"
      sudo bash -c "./$BIN_NAME completion bash > $COMPLETION_DIR_BASH/$BIN_NAME"
    fi

    # 安装 zsh 补全
    mkdir -p "$COMPLETION_DIR_ZSH"
    ./$BIN_NAME completion zsh > "$COMPLETION_DIR_ZSH/_$BIN_NAME"

    # 安装 fish 补全
    mkdir -p "$COMPLETION_DIR_FISH"
    ./$BIN_NAME completion fish > "$COMPLETION_DIR_FISH/$BIN_NAME.fish"
    ;;
  msys*|cygwin*|mingw*)
    BIN_PATH="$USERPROFILE/bin/$BIN_NAME.exe"
    mkdir -p "$(dirname "$BIN_PATH")"
    install -m 0755 "$BIN_NAME.exe" "$BIN_PATH"
    echo "Installed $BIN_NAME.exe to $BIN_PATH"
    ;;
  *)
    echo "Unsupported OS: $OS"
    exit 1
    ;;
esac

echo "$BIN_NAME installation completed."
