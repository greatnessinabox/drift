#!/bin/sh
# drift installer — detects OS/arch and downloads the right binary
# Usage: curl -fsSL https://raw.githubusercontent.com/greatnessinabox/drift/main/install.sh | sh

set -e

REPO="greatnessinabox/drift"
BINARY="drift"
INSTALL_DIR="/usr/local/bin"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
NC='\033[0m'

info()  { printf "${CYAN}==>${NC} %s\n" "$1"; }
ok()    { printf "${GREEN}==>${NC} %s\n" "$1"; }
warn()  { printf "${YELLOW}==>${NC} %s\n" "$1"; }
err()   { printf "${RED}error:${NC} %s\n" "$1" >&2; exit 1; }

# Detect OS
detect_os() {
    case "$(uname -s)" in
        Linux*)  echo "linux" ;;
        Darwin*) echo "darwin" ;;
        MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
        *) err "Unsupported OS: $(uname -s). Install manually from https://github.com/$REPO/releases" ;;
    esac
}

# Detect architecture
detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)   echo "amd64" ;;
        aarch64|arm64)  echo "arm64" ;;
        *) err "Unsupported architecture: $(uname -m). Install manually from https://github.com/$REPO/releases" ;;
    esac
}

# Get latest release tag from GitHub API
get_latest_version() {
    if command -v curl > /dev/null 2>&1; then
        curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/'
    elif command -v wget > /dev/null 2>&1; then
        wget -qO- "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/'
    else
        err "curl or wget is required"
    fi
}

# Download a file
download() {
    if command -v curl > /dev/null 2>&1; then
        curl -fsSL -o "$2" "$1"
    elif command -v wget > /dev/null 2>&1; then
        wget -qO "$2" "$1"
    fi
}

main() {
    info "Installing drift — real-time codebase health dashboard"
    echo ""

    OS=$(detect_os)
    ARCH=$(detect_arch)
    info "Detected platform: ${OS}/${ARCH}"

    VERSION=$(get_latest_version)
    if [ -z "$VERSION" ]; then
        err "Could not determine latest version. Check https://github.com/$REPO/releases"
    fi
    info "Latest version: ${VERSION}"

    # Build download URL
    if [ "$OS" = "windows" ]; then
        ARCHIVE="${BINARY}_${OS}_${ARCH}.zip"
    else
        ARCHIVE="${BINARY}_${OS}_${ARCH}.tar.gz"
    fi
    URL="https://github.com/$REPO/releases/download/${VERSION}/${ARCHIVE}"

    # Download to temp directory
    TMP_DIR=$(mktemp -d)
    trap 'rm -rf "$TMP_DIR"' EXIT

    info "Downloading ${ARCHIVE}..."
    download "$URL" "$TMP_DIR/$ARCHIVE"

    # Extract
    info "Extracting..."
    if [ "$OS" = "windows" ]; then
        unzip -q "$TMP_DIR/$ARCHIVE" -d "$TMP_DIR"
    else
        tar -xzf "$TMP_DIR/$ARCHIVE" -C "$TMP_DIR"
    fi

    # Install binary
    if [ -w "$INSTALL_DIR" ]; then
        mv "$TMP_DIR/$BINARY" "$INSTALL_DIR/$BINARY"
    elif command -v sudo > /dev/null 2>&1; then
        warn "Installing to $INSTALL_DIR (requires sudo)"
        sudo mv "$TMP_DIR/$BINARY" "$INSTALL_DIR/$BINARY"
    else
        # Fallback to user-local directory
        INSTALL_DIR="$HOME/.local/bin"
        mkdir -p "$INSTALL_DIR"
        mv "$TMP_DIR/$BINARY" "$INSTALL_DIR/$BINARY"
        warn "Installed to $INSTALL_DIR — make sure it's in your PATH"
    fi

    chmod +x "$INSTALL_DIR/$BINARY"

    # Verify
    if "$INSTALL_DIR/$BINARY" --version > /dev/null 2>&1; then
        echo ""
        ok "drift installed successfully! (${VERSION})"
        echo ""
        echo "  Get started:"
        echo "    cd your-project"
        echo "    drift          # launch TUI dashboard"
        echo "    drift fix      # AI-powered refactoring"
        echo "    drift check    # quick health check"
        echo ""
        echo "  Docs: https://github.com/$REPO"
    else
        ok "drift binary placed at $INSTALL_DIR/$BINARY"
        warn "Could not verify — you may need to restart your shell or add $INSTALL_DIR to PATH"
    fi
}

main
