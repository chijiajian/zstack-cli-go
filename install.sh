#!/bin/bash
set -e

BIN_NAME="zstack-cli"
REPO="chijiajian/zstack-cli-go"
VERSION=${VERSION:-"v1.0.0"}

COMPLETION_DIR_BASH="/etc/bash_completion.d"
COMPLETION_DIR_ZSH="$HOME/.zsh/completions"
COMPLETION_DIR_FISH="$HOME/.config/fish/completions"

OS=$(uname | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  armv7l) ARCH="arm" ;;
  i386|i686) ARCH="386" ;;
esac

# Release 
TARBALL="${BIN_NAME}-go_${VERSION#v}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${TARBALL}"

echo "Downloading $URL ..."
curl -sSL "$URL" -o "$TARBALL"

echo "Extracting..."
tar -xzf "$TARBALL"


BIN_FILE="${BIN_NAME}-go_v${VERSION#v}"

case "$OS" in
  linux|darwin)
    echo "Installing $BIN_FILE to /usr/local/bin/${BIN_NAME} ..."
    sudo install -m 0755 "$BIN_FILE" "/usr/local/bin/${BIN_NAME}"


    if [ -d "$COMPLETION_DIR_BASH" ]; then
      echo "Installing bash completion..."
      sudo mkdir -p "$COMPLETION_DIR_BASH"
      sudo "$BIN_NAME" completion bash > "$COMPLETION_DIR_BASH/$BIN_NAME"
    fi


    mkdir -p "$COMPLETION_DIR_ZSH"
    "$BIN_NAME" completion zsh > "$COMPLETION_DIR_ZSH/_$BIN_NAME"


    mkdir -p "$COMPLETION_DIR_FISH"
    "$BIN_NAME" completion fish > "$COMPLETION_DIR_FISH/$BIN_NAME.fish"
    ;;
  msys*|cygwin*|mingw*)
    BIN_PATH="$USERPROFILE/bin/${BIN_NAME}.exe"
    mkdir -p "$(dirname "$BIN_PATH")"
    install -m 0755 "${BIN_FILE}.exe" "$BIN_PATH"
    echo "Installed ${BIN_NAME}.exe to $BIN_PATH"
    ;;
  *)
    echo "Unsupported OS: $OS"
    exit 1
    ;;
esac

echo "✅ $BIN_NAME installation completed (version $VERSION)"
