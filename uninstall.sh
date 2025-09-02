#!/bin/bash
set -e

BIN_NAME="zstack-cli"
COMPLETION_DIR_BASH="/etc/bash_completion.d"
COMPLETION_DIR_ZSH="$HOME/.zsh/completions"
COMPLETION_DIR_FISH="$HOME/.config/fish/completions"

OS=$(uname | tr '[:upper:]' '[:lower:]')

case "$OS" in
  linux|darwin)
    BIN_PATH="/usr/local/bin/$BIN_NAME"
    if [ -f "$BIN_PATH" ]; then
      sudo rm -f "$BIN_PATH"
      echo "Removed $BIN_PATH"
    fi

    # 删除 bash 补全
    if [ -f "$COMPLETION_DIR_BASH/$BIN_NAME" ]; then
      sudo rm -f "$COMPLETION_DIR_BASH/$BIN_NAME"
      echo "Removed bash completion"
    fi

    # 删除 zsh 补全
    if [ -f "$COMPLETION_DIR_ZSH/_$BIN_NAME" ]; then
      rm -f "$COMPLETION_DIR_ZSH/_$BIN_NAME"
      echo "Removed zsh completion"
    fi

    # 删除 fish 补全
    if [ -f "$COMPLETION_DIR_FISH/$BIN_NAME.fish" ]; then
      rm -f "$COMPLETION_DIR_FISH/$BIN_NAME.fish"
      echo "Removed fish completion"
    fi
    ;;
  msys*|cygwin*|mingw*)
    BIN_PATH="$USERPROFILE/bin/$BIN_NAME.exe"
    if [ -f "$BIN_PATH" ]; then
      rm -f "$BIN_PATH"
      echo "Removed $BIN_PATH"
    fi
    ;;
  *)
    echo "Unsupported OS: $OS"
    exit 1
    ;;
esac

echo "$BIN_NAME uninstallation completed."
