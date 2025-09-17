#!/bin/bash
# SSH Secret Keeper Installation Script
#
# This script automatically detects your operating system and architecture,
# downloads the appropriate binary from GitHub releases, and installs it.
#
# Usage:
#   curl -sSL https://github.com/rafaelvzago/ssh-secret-keeper/raw/main/install.sh | bash
#
# Options:
#   --version VERSION    Install specific version (default: latest)
#   --install-dir DIR    Installation directory (default: /usr/local/bin)
#   --user              Install to user directory (~/.local/bin)
#   --help              Show help message

set -e  # Exit on any error

# Configuration
REPO="rafaelvzago/ssh-secret-keeper"
BINARY_NAME="sshsk"
GITHUB_API="https://api.github.com/repos"
GITHUB_RELEASES="https://github.com/rafaelvzago/ssh-secret-keeper/releases"
DEFAULT_INSTALL_DIR="/usr/local/bin"
USER_INSTALL_DIR="$HOME/.local/bin"
TEMP_DIR="/tmp/sshsk-install-$$"

# Global variables
VERSION=""
INSTALL_DIR="$DEFAULT_INSTALL_DIR"
USER_INSTALL=false
OS=""
ARCH=""
PLATFORM=""
FINAL_INSTALL_DIR=""

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Logging functions
info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
    cleanup
    exit 1
}

debug() {
    if [ "${DEBUG:-}" = "1" ]; then
        echo -e "${BOLD}[DEBUG]${NC} $1" >&2
    fi
}

# Cleanup function
cleanup() {
    if [ -d "$TEMP_DIR" ]; then
        debug "Cleaning up temporary directory: $TEMP_DIR"
        rm -rf "$TEMP_DIR"
    fi
}

# Set up cleanup trap
trap cleanup EXIT

# Show help message
show_help() {
    cat << EOF
SSH Secret Keeper Installation Script

This script automatically detects your system and installs the appropriate
SSH Secret Keeper binary from GitHub releases.

USAGE:
    curl -sSL https://github.com/rafaelvzago/ssh-secret-keeper/raw/main/install.sh | bash

    # Or download and run locally:
    wget https://github.com/rafaelvzago/ssh-secret-keeper/raw/main/install.sh
    chmod +x install.sh
    ./install.sh [OPTIONS]

OPTIONS:
    --version VERSION    Install specific version (e.g., --version 1.2.0)
                        Default: latest release

    --install-dir DIR    Custom installation directory
                        Default: /usr/local/bin (system-wide)

    --user              Install to user directory (~/.local/bin)
                        Useful when you don't have sudo access

    --help              Show this help message

EXAMPLES:
    # Install latest version (system-wide, requires sudo)
    curl -sSL https://github.com/rafaelvzago/ssh-secret-keeper/raw/main/install.sh | bash

    # Install to user directory (no sudo required)
    curl -sSL https://github.com/rafaelvzago/ssh-secret-keeper/raw/main/install.sh | bash -s -- --user

    # Install specific version
    curl -sSL https://github.com/rafaelvzago/ssh-secret-keeper/raw/main/install.sh | bash -s -- --version 1.2.0

    # Install to custom directory
    curl -sSL https://github.com/rafaelvzago/ssh-secret-keeper/raw/main/install.sh | bash -s -- --install-dir /opt/bin

SUPPORTED PLATFORMS:
    - Linux (amd64, arm64)
    - macOS (amd64, arm64)

REQUIREMENTS:
    - curl or wget
    - tar
    - One of: sha256sum, shasum (for checksum verification)

For more information, visit: https://github.com/rafaelvzago/ssh-secret-keeper
EOF
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Compare version numbers
version_ge() {
    [ "$(printf '%s\n' "$2" "$1" | sort -V | head -n1)" = "$2" ]
}

# Check prerequisites
check_prerequisites() {
    debug "Checking prerequisites..."

    # Check for download tools
    if ! command_exists curl && ! command_exists wget; then
        error "Neither curl nor wget found. Please install one of them."
    fi

    # Check for tar
    if ! command_exists tar; then
        error "tar command not found. Please install tar."
    fi

    # Check for checksum tools (warn if missing)
    if ! command_exists sha256sum && ! command_exists shasum; then
        warn "No checksum utility found (sha256sum or shasum). Checksum verification will be skipped."
    fi

    debug "Prerequisites check passed"
}

# Detect platform (OS and architecture)
detect_platform() {
    debug "Detecting platform..."

    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    debug "Raw OS: $OS, Raw ARCH: $ARCH"

    case "$OS" in
        linux*)
            OS="linux"
            ;;
        darwin*)
            OS="darwin"
            ;;
        *)
            error "Unsupported operating system: $OS. Supported: linux, darwin"
            ;;
    esac

    case "$ARCH" in
        x86_64)
            ARCH="amd64"
            ;;
        aarch64)
            ARCH="arm64"
            ;;
        arm64)
            ARCH="arm64"
            ;;
        *)
            error "Unsupported architecture: $ARCH. Supported: amd64, arm64"
            ;;
    esac

    PLATFORM="${OS}-${ARCH}"
    info "Detected platform: $PLATFORM"
}

# Get latest version from GitHub API
get_latest_version() {
    debug "Fetching latest version from GitHub API..."

    local api_url="$GITHUB_API/$REPO/releases/latest"
    debug "API URL: $api_url"

    if command_exists curl; then
        VERSION=$(curl -s "$api_url" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/' | sed 's/^v//')
    elif command_exists wget; then
        VERSION=$(wget -qO- "$api_url" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/' | sed 's/^v//')
    else
        error "Neither curl nor wget available for API call"
    fi

    if [ -z "$VERSION" ]; then
        error "Failed to get latest version from GitHub API. Please check your internet connection."
    fi

    debug "Fetched version: $VERSION"
    info "Latest version: v$VERSION"
}

# Handle existing installation
handle_existing_installation() {
    if command_exists "$BINARY_NAME"; then
        local existing_path
        local existing_version

        existing_path=$(which "$BINARY_NAME")
        existing_version=$($BINARY_NAME --version 2>/dev/null | grep -o 'v[0-9][0-9.]*' | head -1 | sed 's/^v//' || echo "unknown")

        info "Found existing installation:"
        info "  Version: $existing_version"
        info "  Location: $existing_path"

        if [ "$existing_version" = "$VERSION" ]; then
            success "Already up to date! (v$VERSION)"
            info "To reinstall, remove the existing binary first:"
            info "  sudo rm $existing_path"
            exit 0
        else
            info "Updating from v$existing_version to v$VERSION"
        fi
    else
        debug "No existing installation found"
    fi
}

# Verify checksum
verify_checksum() {
    local archive_name="$1"

    debug "Verifying checksum for $archive_name..."

    if [ ! -f "checksums.txt" ]; then
        warn "Checksums file not found, skipping verification"
        return 0
    fi

    local expected actual

    if command_exists sha256sum; then
        expected=$(grep "$archive_name" checksums.txt | cut -d' ' -f1)
        actual=$(sha256sum "$archive_name" | cut -d' ' -f1)
    elif command_exists shasum; then
        expected=$(grep "$archive_name" checksums.txt | cut -d' ' -f1)
        actual=$(shasum -a 256 "$archive_name" | cut -d' ' -f1)
    else
        warn "No checksum utility found, skipping verification"
        return 0
    fi

    debug "Expected checksum: $expected"
    debug "Actual checksum: $actual"

    if [ -z "$expected" ]; then
        warn "Checksum not found in checksums.txt for $archive_name"
        return 0
    fi

    if [ "$expected" != "$actual" ]; then
        error "Checksum verification failed for $archive_name"
    fi

    success "Checksum verification passed"
}

# Download file with progress
download_file() {
    local url="$1"
    local output="$2"

    debug "Downloading: $url -> $output"

    if command_exists curl; then
        if curl -sSL --fail "$url" -o "$output"; then
            debug "Download successful with curl"
            return 0
        else
            debug "Download failed with curl"
            return 1
        fi
    elif command_exists wget; then
        if wget -q "$url" -O "$output"; then
            debug "Download successful with wget"
            return 0
        else
            debug "Download failed with wget"
            return 1
        fi
    else
        error "No download tool available"
    fi
}

# Download and extract binary
download_and_install() {
    local archive_name="ssh-secret-keeper-${VERSION}-${PLATFORM}.tar.gz"
    local download_url="$GITHUB_RELEASES/download/v${VERSION}/$archive_name"
    local checksums_url="$GITHUB_RELEASES/download/v${VERSION}/checksums.txt"

    info "Downloading SSH Secret Keeper v$VERSION for $PLATFORM..."
    debug "Archive: $archive_name"
    debug "Download URL: $download_url"

    # Create temporary directory
    debug "Creating temporary directory: $TEMP_DIR"
    mkdir -p "$TEMP_DIR"
    cd "$TEMP_DIR"

    # Download archive
    info "Downloading binary archive..."
    if ! download_file "$download_url" "$archive_name"; then
        error "Failed to download $archive_name. Please check if the release exists for your platform."
    fi

    # Download checksums
    info "Downloading checksums..."
    if ! download_file "$checksums_url" "checksums.txt"; then
        warn "Failed to download checksums.txt, skipping verification"
    else
        verify_checksum "$archive_name"
    fi

    # Extract archive
    info "Extracting archive..."
    if ! tar -xzf "$archive_name"; then
        error "Failed to extract $archive_name"
    fi

    # Verify binary exists
    if [ ! -f "$BINARY_NAME" ]; then
        error "Binary $BINARY_NAME not found in archive"
    fi

    # Make binary executable
    chmod +x "$BINARY_NAME"

    success "Binary downloaded and extracted successfully"
}

# Determine installation directory
determine_install_dir() {
    if [ "$USER_INSTALL" = true ]; then
        FINAL_INSTALL_DIR="$USER_INSTALL_DIR"
        debug "Using user installation directory: $FINAL_INSTALL_DIR"
    else
        FINAL_INSTALL_DIR="$INSTALL_DIR"
        debug "Using system installation directory: $FINAL_INSTALL_DIR"
    fi
}

# Install binary to system directory
install_system_wide() {
    local target_dir="$1"

    debug "Attempting system-wide installation to: $target_dir"

    # Check if directory is writable
    if [ -w "$target_dir" ] || [ "$EUID" -eq 0 ]; then
        debug "Directory is writable, installing directly"
        cp "$BINARY_NAME" "$target_dir/"
        chmod +x "$target_dir/$BINARY_NAME"
        FINAL_INSTALL_DIR="$target_dir"
        return 0
    fi

    # Try with sudo
    if command_exists sudo; then
        info "Installing to $target_dir (requires sudo)..."
        if sudo cp "$BINARY_NAME" "$target_dir/" && sudo chmod +x "$target_dir/$BINARY_NAME"; then
            FINAL_INSTALL_DIR="$target_dir"
            return 0
        else
            debug "Sudo installation failed"
            return 1
        fi
    else
        debug "Sudo not available"
        return 1
    fi
}

# Install binary to user directory
install_user_local() {
    local target_dir="$USER_INSTALL_DIR"

    debug "Installing to user directory: $target_dir"

    # Create directory if it doesn't exist
    mkdir -p "$target_dir"

    # Copy binary
    cp "$BINARY_NAME" "$target_dir/"
    chmod +x "$target_dir/$BINARY_NAME"

    FINAL_INSTALL_DIR="$target_dir"

    # Check if directory is in PATH
    if ! echo "$PATH" | grep -q "$target_dir"; then
        warn "Installation directory not in PATH"
        info "Add the following to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
        echo ""
        echo "    export PATH=\"\$PATH:$target_dir\""
        echo ""
        info "Then reload your shell or run: source ~/.bashrc"
    fi
}

# Install binary
install_binary() {
    cd "$TEMP_DIR"

    determine_install_dir

    if [ "$USER_INSTALL" = true ]; then
        info "Installing to user directory..."
        install_user_local
    else
        info "Installing to system directory..."
        if ! install_system_wide "$FINAL_INSTALL_DIR"; then
            warn "System installation failed, falling back to user installation"
            install_user_local
        fi
    fi

    success "Binary installed to: $FINAL_INSTALL_DIR"
}

# Build from source (fallback)
build_from_source() {
    warn "Attempting to build from source as fallback..."

    # Check if Go is available
    if ! command_exists go; then
        error "Go not found. Please install Go 1.21+ or use a supported platform for pre-built binaries."
    fi

    # Check Go version
    local go_version
    go_version=$(go version | grep -o 'go[0-9]\+\.[0-9]\+' | sed 's/go//')
    if ! version_ge "$go_version" "1.21"; then
        error "Go 1.21+ required, found Go $go_version"
    fi

    # Check if git is available
    if ! command_exists git; then
        error "Git not found. Required for building from source."
    fi

    info "Building SSH Secret Keeper from source..."
    info "Go version: $go_version"

    # Clone repository
    local source_dir="$TEMP_DIR/source"
    info "Cloning repository..."
    if ! git clone --depth 1 "https://github.com/$REPO.git" "$source_dir"; then
        error "Failed to clone repository"
    fi

    cd "$source_dir"

    # Build using make
    info "Building binary..."
    if command_exists make; then
        if ! make build; then
            error "Build failed using make"
        fi

        # Copy binary from bin directory
        if [ -f "bin/$BINARY_NAME" ]; then
            cp "bin/$BINARY_NAME" "$TEMP_DIR/"
        else
            error "Binary not found after build"
        fi
    else
        # Fallback to direct go build
        warn "Make not found, using direct go build"
        if ! go build -o "$TEMP_DIR/$BINARY_NAME" cmd/main.go; then
            error "Direct go build failed"
        fi
    fi

    cd "$TEMP_DIR"
    chmod +x "$BINARY_NAME"

    success "Built from source successfully"
}

# Test installation
test_installation() {
    debug "Testing installation..."

    if command_exists "$BINARY_NAME"; then
        local installed_version
        installed_version=$($BINARY_NAME --version 2>/dev/null | head -1 || echo "Version check failed")
        success "Installation test passed!"
        info "Installed version: $installed_version"
    else
        warn "Binary not found in PATH. You may need to:"
        info "1. Restart your shell/terminal"
        info "2. Add $FINAL_INSTALL_DIR to your PATH"
        info "3. Run: export PATH=\"\$PATH:$FINAL_INSTALL_DIR\""
    fi
}

# Post-installation information
show_post_install_info() {
    echo ""
    success "SSH Secret Keeper installation completed!"
    echo ""
    info "Installation details:"
    info "  Version: v$VERSION"
    info "  Location: $FINAL_INSTALL_DIR/$BINARY_NAME"
    info "  Platform: $PLATFORM"
    echo ""
    info "Quick start guide:"
    echo "  1. Set Vault address:    export VAULT_ADDR='https://your-vault:8200'"
    echo "  2. Set Vault token:      export VAULT_TOKEN='your-token'"
    echo "  3. Initialize:           $BINARY_NAME init"
    echo "  4. Analyze SSH dir:      $BINARY_NAME analyze"
    echo "  5. Create backup:        $BINARY_NAME backup"
    echo ""
    info "For detailed documentation, visit:"
    info "  https://github.com/$REPO#readme"
    echo ""
    info "Need help? Check the documentation or create an issue:"
    info "  https://github.com/$REPO/issues"
}

# Parse command line arguments
parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --version)
                if [ -z "$2" ] || [[ "$2" =~ ^-- ]]; then
                    error "Version argument required. Usage: --version 1.2.0"
                fi
                VERSION="$2"
                debug "Version specified: $VERSION"
                shift 2
                ;;
            --install-dir)
                if [ -z "$2" ] || [[ "$2" =~ ^-- ]]; then
                    error "Install directory argument required. Usage: --install-dir /opt/bin"
                fi
                INSTALL_DIR="$2"
                debug "Install directory specified: $INSTALL_DIR"
                shift 2
                ;;
            --user)
                USER_INSTALL=true
                debug "User installation requested"
                shift
                ;;
            --help)
                show_help
                exit 0
                ;;
            --debug)
                DEBUG=1
                debug "Debug mode enabled"
                shift
                ;;
            *)
                error "Unknown option: $1. Use --help for usage information."
                ;;
        esac
    done
}

# Main installation function
main() {
    # Show banner
    echo ""
    echo -e "${BOLD}SSH Secret Keeper Installation Script${NC}"
    echo "======================================="
    echo ""

    # Parse arguments
    parse_arguments "$@"

    # Check prerequisites
    check_prerequisites

    # Detect platform
    detect_platform

    # Get version to install
    if [ -z "$VERSION" ]; then
        get_latest_version
    else
        info "Installing specified version: v$VERSION"
    fi

    # Handle existing installation
    handle_existing_installation

    # Try to download and install pre-built binary
    if download_and_install; then
        install_binary
    else
        # Fallback to building from source
        warn "Pre-built binary not available for $PLATFORM"
        build_from_source
        install_binary
    fi

    # Test installation
    test_installation

    # Show post-installation information
    show_post_install_info
}

# Run main function with all arguments
main "$@"
