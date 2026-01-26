#!/usr/bin/env bash
set -euo pipefail

REPO="madhermit/omr"
INSTALL_DIR="${HOME}/.local/bin"

# Detect platform
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Get latest release
LATEST=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)
if [[ -z "$LATEST" ]]; then
  echo "Could not determine latest version"
  exit 1
fi

URL="https://github.com/${REPO}/releases/download/${LATEST}/omr-${OS}-${ARCH}"

echo "Installing omr ${LATEST} (${OS}/${ARCH})..."

mkdir -p "$INSTALL_DIR"
curl -fsSL "$URL" -o "${INSTALL_DIR}/omr"
chmod +x "${INSTALL_DIR}/omr"

echo "Installed to ${INSTALL_DIR}/omr"

# Check if in PATH
if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
  echo ""
  echo "Add to your shell profile:"
  echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
fi
