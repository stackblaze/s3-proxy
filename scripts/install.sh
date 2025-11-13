#!/bin/sh
set -e

# S3 Proxy Installer
# Downloads and installs the appropriate binary for your platform

GITHUB_REPO="stackblaze/s3-proxy"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
BINARY_NAME="s3-proxy"

# Detect OS and architecture
detect_platform() {
    OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
    ARCH="$(uname -m)"
    
    case "$ARCH" in
        x86_64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            echo "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac
    
    case "$OS" in
        linux|darwin)
            ;;
        *)
            echo "Unsupported OS: $OS"
            exit 1
            ;;
    esac
    
    # Windows detection
    if [ "$OS" = "mingw" ] || [ "$OS" = "msys" ] || [ "$OS" = "cygwin" ]; then
        OS="windows"
        BINARY_NAME="${BINARY_NAME}.exe"
    fi
}

# Get latest release version
get_latest_version() {
    if command -v curl >/dev/null 2>&1; then
        VERSION=$(curl -sSfL "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    elif command -v wget >/dev/null 2>&1; then
        VERSION=$(wget -qO- "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    else
        echo "Error: curl or wget is required"
        exit 1
    fi
    
    if [ -z "$VERSION" ]; then
        echo "Error: Could not determine latest version"
        exit 1
    fi
}

# Download and install binary
install_binary() {
    VERSION="${1:-latest}"
    URL="https://github.com/${GITHUB_REPO}/releases/download/${VERSION}/s3-proxy-${OS}-${ARCH}"
    
    if [ "$OS" = "windows" ]; then
        URL="${URL}.exe"
    fi
    
    echo "Downloading s3-proxy ${VERSION} for ${OS}/${ARCH}..."
    echo "  From: ${URL}"
    
    # Create temp file
    TMP_FILE=$(mktemp)
    trap "rm -f $TMP_FILE" EXIT
    
    # Download
    if command -v curl >/dev/null 2>&1; then
        curl -sSfL -o "$TMP_FILE" "$URL"
    elif command -v wget >/dev/null 2>&1; then
        wget -qO "$TMP_FILE" "$URL"
    else
        echo "Error: curl or wget is required"
        exit 1
    fi
    
    # Check if download was successful
    if [ ! -s "$TMP_FILE" ]; then
        echo "Error: Download failed or file is empty"
        exit 1
    fi
    
    # Make executable (if not Windows)
    if [ "$OS" != "windows" ]; then
        chmod +x "$TMP_FILE"
    fi
    
    # Install
    echo "Installing to ${INSTALL_DIR}/${BINARY_NAME}..."
    mkdir -p "$INSTALL_DIR"
    
    if [ "$OS" = "windows" ]; then
        cp "$TMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
    else
        sudo cp "$TMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}" 2>/dev/null || cp "$TMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
        sudo chmod +x "${INSTALL_DIR}/${BINARY_NAME}" 2>/dev/null || chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    fi
    
    # Verify installation
    if command -v "$BINARY_NAME" >/dev/null 2>&1 || [ -x "${INSTALL_DIR}/${BINARY_NAME}" ]; then
        echo ""
        echo "✅ s3-proxy installed successfully!"
        echo ""
        echo "Version: $VERSION"
        echo "Location: ${INSTALL_DIR}/${BINARY_NAME}"
        echo ""
        echo "Run: ${BINARY_NAME} --help"
    else
        echo "⚠️  Installation completed, but binary not found in PATH"
        echo "   Add ${INSTALL_DIR} to your PATH, or run: ${INSTALL_DIR}/${BINARY_NAME}"
    fi
}

# Main
main() {
    echo "S3 Proxy Installer"
    echo "=================="
    echo ""
    
    detect_platform
    echo "Detected platform: ${OS}/${ARCH}"
    echo ""
    
    get_latest_version
    echo "Latest version: ${VERSION}"
    echo ""
    
    install_binary "$VERSION"
}

main "$@"

