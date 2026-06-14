#!/bin/bash
set -e

echo "========================================"
echo "    Installing WebScript Server         "
echo "========================================"

if ! command -v go &> /dev/null; then
    echo "Error: 'go' is not installed."
    echo "Please install Go first. On Ubuntu/Debian: sudo apt install golang"
    exit 1
fi

if ! command -v git &> /dev/null; then
    echo "Error: 'git' is not installed."
    echo "Please install Git first. On Ubuntu/Debian: sudo apt install git"
    exit 1
fi

echo "-> Cloning repository..."
rm -rf /tmp/webscript-install
git clone https://github.com/LukasYTTT/webscript.git /tmp/webscript-install
cd /tmp/webscript-install

echo "-> Building WebScript..."
CGO_ENABLED=0 go build -trimpath -o wbs .

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

echo "========================================"
echo "  WebScript successfully installed! 🎉  "
echo "========================================"
echo "Type 'wbs' or 'webscript' to use it."
echo "Run 'sudo wbs service' to set up the background systemd service."
