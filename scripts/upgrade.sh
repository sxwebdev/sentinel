#!/usr/bin/env bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
GITHUB_REPO="sxwebdev/sentinel"
TARGET_BINARY="sentinel"
TEMP_DIR="/tmp/sentinel-upgrade-$$"

# Print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Show usage
usage() {
    echo "Usage: $0 <systemd-service-name> [binary-path]"
    echo ""
    echo "Arguments:"
    echo "  systemd-service-name  Name of the systemd service to restart"
    echo "  binary-path          Path to install binary (default: /usr/local/bin/sentinel)"
    echo ""
    echo "Examples:"
    echo "  $0 sentinel"
    echo "  $0 sentinel ./"
    echo "  $0 sentinel-monitoring /opt/sentinel/bin/sentinel"
    echo ""
    echo "This script will:"
    echo "  1. Download the latest Sentinel release for your platform"
    echo "  2. Stop the specified systemd service"
    echo "  3. Replace the binary with the new version"
    echo "  4. Start the service"
    echo "  5. Show service status"
}

# Check if required arguments are provided
if [ $# -lt 1 ] || [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
    usage
    exit 0
fi

SERVICE_NAME="$1"
BINARY_PATH="${2:-/usr/local/bin/sentinel}"
INSTALL_DIR=$(dirname "$BINARY_PATH")

print_info "Starting Sentinel upgrade process..."
print_info "Service: $SERVICE_NAME"
print_info "Binary path: $BINARY_PATH"

# Check if running as root or with sudo
if [ "$EUID" -ne 0 ]; then
    print_error "This script requires root privileges. Please run with sudo."
    exit 1
fi

# Ensure required tools are available
for tool in curl jq tar systemctl; do
    if ! command -v "$tool" &>/dev/null; then
        print_error "$tool is not installed. Please install $tool and try again."
        exit 1
    fi
done

# Determine the operating system
OS=$(uname)
if [ "$OS" = "Darwin" ]; then
    PLATFORM="darwin"
elif [ "$OS" = "Linux" ]; then
    PLATFORM="linux"
else
    print_error "Unsupported OS: $OS"
    exit 1
fi

# Determine the architecture
ARCH=$(uname -m)
case "$ARCH" in
    x86_64)
        ARCH="amd64"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ;;
    *)
        print_error "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

print_info "Detected platform: $PLATFORM, architecture: $ARCH"

# Check if service exists
if ! systemctl list-unit-files | grep -q "^$SERVICE_NAME.service"; then
    print_error "Service '$SERVICE_NAME' not found in systemd"
    print_warning "Available services containing 'sentinel':"
    systemctl list-unit-files | grep -i sentinel || true
    exit 1
fi

# Check if binary exists
if [ ! -f "$BINARY_PATH" ]; then
    print_error "Binary not found at $BINARY_PATH"
    exit 1
fi

# Get current version if possible
CURRENT_VERSION=""
if [ -x "$BINARY_PATH" ]; then
    CURRENT_VERSION=$("$BINARY_PATH" --version 2>/dev/null || echo "unknown")
    print_info "Current version: $CURRENT_VERSION"
fi

# Fetch the latest release information from GitHub API
print_info "Fetching latest release information..."
API_URL="https://api.github.com/repos/${GITHUB_REPO}/releases/latest"
RELEASE_INFO=$(curl --silent "$API_URL")

if [ -z "$RELEASE_INFO" ]; then
    print_error "Failed to fetch release information from GitHub"
    exit 1
fi

# Get the latest version
LATEST_VERSION=$(echo "$RELEASE_INFO" | jq -r '.tag_name')
if [ -z "$LATEST_VERSION" ] || [ "$LATEST_VERSION" = "null" ]; then
    print_error "Failed to parse latest version from GitHub API"
    exit 1
fi

print_info "Latest version: $LATEST_VERSION"

# Check if update is needed
if [ "$CURRENT_VERSION" = "$LATEST_VERSION" ]; then
    print_warning "Already running the latest version ($LATEST_VERSION)"
    read -p "Do you want to continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "Upgrade cancelled"
        exit 0
    fi
fi

# Find the asset matching the platform and architecture
ASSET_NAME=$(echo "$RELEASE_INFO" | jq -r --arg platform "$PLATFORM" --arg arch "$ARCH" '
    .assets[] | select(.name | test($platform) and test($arch)) | .name
')

if [ -z "$ASSET_NAME" ]; then
    print_error "No archive found for platform $PLATFORM and architecture $ARCH"
    print_info "Available assets:"
    echo "$RELEASE_INFO" | jq -r '.assets[].name'
    exit 1
fi

print_info "Found archive: $ASSET_NAME"

# Get the download URL for the asset
ASSET_URL=$(echo "$RELEASE_INFO" | jq -r --arg asset "$ASSET_NAME" '
    .assets[] | select(.name == $asset) | .browser_download_url
')

if [ -z "$ASSET_URL" ]; then
    print_error "Failed to obtain download URL"
    exit 1
fi

# Create temporary directory
mkdir -p "$TEMP_DIR"
cd "$TEMP_DIR"

print_info "Downloading binary from $ASSET_URL"
if ! curl -L --silent --show-error -o "$ASSET_NAME" "$ASSET_URL"; then
    print_error "Failed to download archive"
    rm -rf "$TEMP_DIR"
    exit 1
fi

# Extract the archive
print_info "Extracting archive..."
if ! tar -xzf "$ASSET_NAME"; then
    print_error "Failed to extract archive"
    rm -rf "$TEMP_DIR"
    exit 1
fi

# Remove the downloaded archive
rm "$ASSET_NAME"

# Find the extracted binary file recursively in subdirectories
EXTRACTED_BINARY=$(find . -type f -name "$TARGET_BINARY" ! -name "*.tar.gz" | head -n 1)

if [ -z "$EXTRACTED_BINARY" ]; then
    print_error "No binary file found after extraction"
    print_info "Contents of extracted archive:"
    find . -type f
    rm -rf "$TEMP_DIR"
    exit 1
fi

print_info "Found extracted binary: $EXTRACTED_BINARY"

# Ensure the extracted file is executable
if [ ! -x "$EXTRACTED_BINARY" ]; then
    print_error "Extracted file is not executable"
    rm -rf "$TEMP_DIR"
    exit 1
fi

# Verify the new binary works
print_info "Verifying new binary..."
NEW_VERSION=$("$EXTRACTED_BINARY" --version 2>/dev/null || echo "unknown")
print_info "New binary version: $NEW_VERSION"

# Stop the service
print_info "Stopping service '$SERVICE_NAME'..."
if ! systemctl stop "$SERVICE_NAME"; then
    print_error "Failed to stop service '$SERVICE_NAME'"
    rm -rf "$TEMP_DIR"
    exit 1
fi

print_success "Service stopped successfully"

# Backup current binary
BACKUP_PATH="${BINARY_PATH}.backup.$(date +%Y%m%d-%H%M%S)"
print_info "Creating backup at $BACKUP_PATH"
if ! cp "$BINARY_PATH" "$BACKUP_PATH"; then
    print_error "Failed to create backup"
    print_warning "Attempting to start service with old binary..."
    systemctl start "$SERVICE_NAME" || true
    rm -rf "$TEMP_DIR"
    exit 1
fi

# Replace the binary
print_info "Replacing binary at $BINARY_PATH..."
if ! cp "$EXTRACTED_BINARY" "$BINARY_PATH"; then
    print_error "Failed to replace binary"
    print_warning "Restoring from backup..."
    cp "$BACKUP_PATH" "$BINARY_PATH"
    systemctl start "$SERVICE_NAME" || true
    rm -rf "$TEMP_DIR"
    exit 1
fi

# Ensure the binary is executable
chmod +x "$BINARY_PATH"

# Start the service
print_info "Starting service '$SERVICE_NAME'..."
if ! systemctl start "$SERVICE_NAME"; then
    print_error "Failed to start service with new binary"
    print_warning "Restoring from backup..."
    cp "$BACKUP_PATH" "$BINARY_PATH"
    systemctl start "$SERVICE_NAME" || true
    rm -rf "$TEMP_DIR"
    exit 1
fi

# Wait a moment for service to initialize
sleep 2

# Check service status
print_info "Checking service status..."
if systemctl is-active --quiet "$SERVICE_NAME"; then
    print_success "Service is running successfully"
    
    # Show service status
    echo ""
    print_info "Service status:"
    systemctl status "$SERVICE_NAME" --no-pager --lines=5
    
    # Clean up backup if everything is working
    print_info "Removing backup file..."
    rm -f "$BACKUP_PATH"
    
    print_success "Upgrade completed successfully!"
    print_info "Upgraded from $CURRENT_VERSION to $NEW_VERSION"
else
    print_error "Service failed to start properly"
    print_warning "Restoring from backup..."
    systemctl stop "$SERVICE_NAME" || true
    cp "$BACKUP_PATH" "$BINARY_PATH"
    systemctl start "$SERVICE_NAME" || true
    rm -rf "$TEMP_DIR"
    exit 1
fi

# Clean up temporary directory
print_info "Cleaning up temporary files..."
rm -rf "$TEMP_DIR"

print_success "All done! Sentinel has been upgraded to $NEW_VERSION"
