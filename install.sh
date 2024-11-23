#!/usr/bin/env bash

set -e

# Configuration
BINARY_NAME="scrapdf"
GITHUB_REPO="ppicom/scrapedf"
INSTALL_DIR="/usr/local/bin"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print step description
step() {
    echo -e "${BLUE}==>${NC} $1"
}

# Print error and exit
error() {
    echo -e "${RED}Error:${NC} $1" >&2
    exit 1
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Detect architecture and OS
detect_platform() {
    local ARCH OS
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$ARCH" in
        x86_64|amd64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *) error "Unsupported architecture: $ARCH" ;;
    esac

    case "$OS" in
        linux) OS="linux" ;;
        darwin) OS="darwin" ;;
        *) error "Unsupported operating system: $OS" ;;
    esac

    echo "${OS}_${ARCH}"
}

# Get the latest release version
get_latest_version() {
    if ! command_exists curl; then
        error "curl is required but not installed"
    fi

    local latest_version
    latest_version=$(curl -s "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" | 
        grep '"tag_name":' | 
        sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$latest_version" ]; then
        error "Failed to get latest version"
    fi
    
    echo "$latest_version"
}

# Main installation
main() {
    # Check if running with sudo when needed
    if [ ! -w "$INSTALL_DIR" ]; then
        error "Installation requires sudo access. Please run: sudo curl ... | sudo bash"
    }

    step "Detecting platform"
    PLATFORM=$(detect_platform)
    echo "Detected: $PLATFORM"

    step "Getting latest version"
    VERSION=$(get_latest_version)
    echo "Latest version: $VERSION"

    DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/download/${VERSION}/${BINARY_NAME}_${VERSION}_${PLATFORM}.tar.gz"
    
    step "Downloading binary"
    TMP_DIR=$(mktemp -d)
    curl -sL "$DOWNLOAD_URL" | tar xz -C "$TMP_DIR"
    
    step "Installing binary"
    mv "$TMP_DIR/${BINARY_NAME}" "$INSTALL_DIR/"
    chmod +x "$INSTALL_DIR/${BINARY_NAME}"
    rm -rf "$TMP_DIR"

    step "Verifying installation"
    if command_exists $BINARY_NAME; then
        echo -e "${GREEN}Successfully installed ${BINARY_NAME}${NC}"
        echo "Run '${BINARY_NAME} --help' to get started"
    else
        error "Installation failed"
    fi
}

main "$@" 