#!/bin/bash
set -e

echo "========================================"
echo "    Installing WebScript Server         "
echo "========================================"

if ! command -v git &> /dev/null; then
    echo "Error: 'git' is not installed."
    echo "Please install Git first. (e.g. sudo apt install git)"
    exit 1
fi

echo "-> Cloning repository..."
rm -rf /tmp/webscript-install
git clone https://github.com/LukasYTTT/webscript.git /tmp/webscript-install
cd /tmp/webscript-install

GO_BIN="go"
INSTALL_GO=false

if ! command -v go &> /dev/null; then
    echo "-> Go is missing. Downloading a temporary Go compiler..."
    INSTALL_GO=true
else
    # Check if the installed Go version is older than 1.21
    GO_VER=$(go version | awk '{print $3}' | sed 's/go//')
    if [ "$(printf '%s\n' "1.21" "$GO_VER" | sort -V | head -n1)" != "1.21" ]; then
        echo "-> Your Go version ($GO_VER) is too old. Downloading a temporary Go compiler..."
        INSTALL_GO=true
    fi
fi

if [ "$INSTALL_GO" = true ]; then
    ARCH=$(uname -m)
    case $ARCH in
        x86_64) GO_ARCH="amd64" ;;
        aarch64) GO_ARCH="arm64" ;;
        armv7l) GO_ARCH="armv6l" ;;
        *) echo "Unsupported architecture: $ARCH. Please install Go manually."; exit 1 ;;
    esac
    
    GO_TAR="go1.22.4.linux-${GO_ARCH}.tar.gz"
    echo "   Downloading $GO_TAR ..."
    curl -sSL "https://go.dev/dl/$GO_TAR" -o /tmp/wbs-go.tar.gz
    echo "   Extracting Go..."
    rm -rf /tmp/wbs-go
    mkdir -p /tmp/wbs-go
    tar -C /tmp/wbs-go -xzf /tmp/wbs-go.tar.gz
    GO_BIN="/tmp/wbs-go/go/bin/go"
fi

echo "-> Building WebScript..."
CGO_ENABLED=0 $GO_BIN build -trimpath -o wbs .

echo "-> Installing to /usr/local/bin..."
# Use sudo only if we are not root
if [ "$EUID" -ne 0 ]; then
    sudo install -Dm755 wbs /usr/local/bin/wbs
    sudo ln -sf /usr/local/bin/wbs /usr/local/bin/webscript
else
    install -Dm755 wbs /usr/local/bin/wbs
    ln -sf /usr/local/bin/wbs /usr/local/bin/webscript
fi

echo "-> Cleaning up..."
cd /
rm -rf /tmp/webscript-install
if [ "$INSTALL_GO" = true ]; then
    rm -rf /tmp/wbs-go /tmp/wbs-go.tar.gz
fi

echo "========================================"
echo "  WebScript successfully installed! 🎉  "
echo "========================================"
echo "Type 'wbs' or 'webscript' to use it."
echo "Run 'sudo wbs service' to set up the background systemd service."
