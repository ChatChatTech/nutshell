#!/bin/sh
# nutshell installer — auto-detects OS/arch, downloads pre-built binary
# Usage: curl -fsSL https://raw.githubusercontent.com/ChatChatTech/nutshell/main/install.sh | sh
set -e

REPO="ChatChatTech/nutshell"
INSTALL_DIR="/usr/local/bin"
BINARY="nutshell"

# Detect OS
OS="$(uname -s)"
case "$OS" in
  Linux*)  OS_TAG="linux" ;;
  Darwin*) OS_TAG="darwin" ;;
  MINGW*|MSYS*|CYGWIN*) OS_TAG="windows" ;;
  *) echo "Error: unsupported OS: $OS" >&2; exit 1 ;;
esac

# Detect architecture
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64)  ARCH_TAG="amd64" ;;
  aarch64|arm64)  ARCH_TAG="arm64" ;;
  *) echo "Error: unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

# Build asset name
if [ "$OS_TAG" = "windows" ]; then
  ASSET="${BINARY}-${OS_TAG}-${ARCH_TAG}.exe"
else
  ASSET="${BINARY}-${OS_TAG}-${ARCH_TAG}"
fi

# Get latest release tag from GitHub API
echo "Detecting latest nutshell release..."
TAG=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')
if [ -z "$TAG" ]; then
  echo "Error: could not determine latest release" >&2
  exit 1
fi

URL="https://github.com/${REPO}/releases/download/${TAG}/${ASSET}"
echo "Downloading nutshell ${TAG} for ${OS_TAG}/${ARCH_TAG}..."
echo "  ${URL}"

# Download to temp file
TMP="$(mktemp)"
trap 'rm -f "$TMP"' EXIT
curl -fSL --progress-bar -o "$TMP" "$URL"

# Install
chmod +x "$TMP"
if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP" "${INSTALL_DIR}/${BINARY}"
else
  echo "Installing to ${INSTALL_DIR} (requires sudo)..."
  sudo mv "$TMP" "${INSTALL_DIR}/${BINARY}"
fi

echo ""
echo "nutshell ${TAG} installed to ${INSTALL_DIR}/${BINARY}"
echo "Run 'nutshell version' to verify."
