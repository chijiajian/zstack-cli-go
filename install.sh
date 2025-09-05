#!/bin/bash
set -e

# --- Variables ---
BIN_NAME="zstack-cli"
GITHUB_REPO="chijiajian/zstack-cli-go"
VERSION=${1:-$(curl -s "https://api.github.com/repos/$GITHUB_REPO/releases/latest" | grep -Po '"tag_name": "\K.*?(?=")')}
OS=$(uname | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
TMP_DIR=$(mktemp -d)
INSTALL_DIR="/usr/local/bin"

# Completion directories
COMPLETION_DIR_BASH="/etc/bash_completion.d"
COMPLETION_DIR_ZSH="$HOME/.zsh/completions"
COMPLETION_DIR_FISH="$HOME/.config/fish/completions"

# --- Main Script ---
echo "Installing $BIN_NAME version $VERSION for $OS/$ARCH..."

# 1. Map architecture
case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64 | arm64) ARCH="arm64" ;;
  armv7*) ARCH="arm" ;;
  i386 | i686) ARCH="386" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# 2. Download the binary package
DOWNLOAD_URL="https://github.com/$GITHUB_REPO/releases/download/$VERSION/${BIN_NAME}-go_${VERSION}_${OS}_${ARCH}.tar.gz"
echo "Downloading $DOWNLOAD_URL..."
if ! curl -L -o "$TMP_DIR/${BIN_NAME}.tar.gz" "$DOWNLOAD_URL"; then
    echo "Error: Download failed. Check your network or the URL. Aborting installation."
    rm -rf "$TMP_DIR"
    exit 1
fi

# 3. Extract
echo "Extracting..."
if ! tar -xzf "$TMP_DIR/${BIN_NAME}.tar.gz" -C "$TMP_DIR"; then
    echo "Error: Extraction failed. Aborting installation."
    rm -rf "$TMP_DIR"
    exit 1
fi

# Find the binary
BIN_PATH=$(find "$TMP_DIR" -maxdepth 1 -type f -name "${BIN_NAME}-go_${VERSION}_${OS}_${ARCH}" | head -n1)
if [ ! -f "$BIN_PATH" ]; then
    echo "Error: Binary not found after extraction. Aborting installation."
    rm -rf "$TMP_DIR"
    exit 1
fi

# 4. Install the binary to /usr/local/bin
echo "Installing $BIN_NAME to $INSTALL_DIR/$BIN_NAME..."
sudo install -m 0755 "$BIN_PATH" "$INSTALL_DIR/$BIN_NAME"

# 5. Install shell completion
echo "Installing shell completion scripts..."

CURRENT_SHELL=$(basename "$SHELL")

case "$CURRENT_SHELL" in
  "bash")
    # Bash completion (system-wide first, then fallback to user home)
    echo "--> Installing Bash completion..."
    if sudo test -w "$COMPLETION_DIR_BASH" || sudo mkdir -p "$COMPLETION_DIR_BASH" >/dev/null 2>&1; then
        sudo "$INSTALL_DIR/$BIN_NAME" completion bash > "$COMPLETION_DIR_BASH/$BIN_NAME"
        echo "Bash completion installed to system directory: $COMPLETION_DIR_BASH"
    else
        # Fallback to user home directory
        COMPLETION_DIR_USER="$HOME/.bash_completion.d"
        mkdir -p "$COMPLETION_DIR_USER"
        "$INSTALL_DIR/$BIN_NAME" completion bash > "$COMPLETION_DIR_USER/$BIN_NAME"
        if ! grep -qxF "source $COMPLETION_DIR_USER/$BIN_NAME" "$HOME/.bashrc"; then
            echo "source $COMPLETION_DIR_USER/$BIN_NAME" >> "$HOME/.bashrc"
        fi
        echo "Bash completion installed to user directory: $COMPLETION_DIR_USER"
    fi
    ;;
  "zsh")
    # Zsh completion
    echo "--> Installing Zsh completion..."
    mkdir -p "$COMPLETION_DIR_ZSH"
    "$INSTALL_DIR/$BIN_NAME" completion zsh > "$COMPLETION_DIR_ZSH/_$BIN_NAME"
    if ! grep -qxF "fpath+=($COMPLETION_DIR_ZSH)" "$HOME/.zshrc"; then
        echo "fpath+=($COMPLETION_DIR_ZSH)" >> "$HOME/.zshrc"
    fi
    echo "Zsh completion installed to $COMPLETION_DIR_ZSH"
    ;;
  "fish")
    # Fish completion
    echo "--> Installing Fish completion..."
    mkdir -p "$COMPLETION_DIR_FISH"
    "$INSTALL_DIR/$BIN_NAME" completion fish > "$COMPLETION_DIR_FISH/$BIN_NAME.fish"
    echo "Fish completion installed to $COMPLETION_DIR_FISH"
    ;;
  *)
    echo "Warning: Unknown shell type ($CURRENT_SHELL). Skipping shell completion installation."
    ;;
esac

# 6. Clean up temporary files
rm -rf "$TMP_DIR"

# --- Installation Complete Message ---
echo "---"
echo "Congratulations! $BIN_NAME has been successfully installed!"
echo "Please restart your shell or run the following command to enable completion:"

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