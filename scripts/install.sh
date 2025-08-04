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
DEFAULT_INSTALL_DIR="/usr/local/bin"
DEFAULT_CONFIG_DIR="/etc/sentinel"
DEFAULT_DATA_DIR="/var/lib/sentinel"
DEFAULT_SERVICE_NAME="sentinel"
DEFAULT_USER="sentinel"
TEMP_DIR="/tmp/sentinel-install-$$"

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
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -d, --install-dir DIR     Installation directory (default: $DEFAULT_INSTALL_DIR)"
    echo "  -c, --config-dir DIR      Configuration directory (default: $DEFAULT_CONFIG_DIR)"
    echo "  -D, --data-dir DIR        Data directory (default: $DEFAULT_DATA_DIR)"
    echo "  -s, --service-name NAME   Systemd service name (default: $DEFAULT_SERVICE_NAME)"
    echo "  -u, --user USER           Service user (default: $DEFAULT_USER)"
    echo "  -v, --version VERSION     Install specific version (default: latest)"
    echo "  -h, --help                Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                                    # Install with defaults"
    echo "  $0 -d /opt/sentinel/bin              # Custom install directory"
    echo "  $0 -u myuser -s sentinel-monitor     # Custom user and service name"
    echo "  $0 -v v1.0.0                        # Install specific version"
    echo ""
    echo "This script will:"
    echo "  1. Download the latest Sentinel release for your platform"
    echo "  2. Create system user and directories"
    echo "  3. Install the binary"
    echo "  4. Create configuration file"
    echo "  5. Create and enable systemd service"
    echo "  6. Start the service"
}

# Parse command line arguments
INSTALL_DIR="$DEFAULT_INSTALL_DIR"
CONFIG_DIR="$DEFAULT_CONFIG_DIR"
DATA_DIR="$DEFAULT_DATA_DIR"
SERVICE_NAME="$DEFAULT_SERVICE_NAME"
SERVICE_USER="$DEFAULT_USER"
VERSION=""

while [[ $# -gt 0 ]]; do
    case $1 in
        -d|--install-dir)
            INSTALL_DIR="$2"
            shift 2
            ;;
        -c|--config-dir)
            CONFIG_DIR="$2"
            shift 2
            ;;
        -D|--data-dir)
            DATA_DIR="$2"
            shift 2
            ;;
        -s|--service-name)
            SERVICE_NAME="$2"
            shift 2
            ;;
        -u|--user)
            SERVICE_USER="$2"
            shift 2
            ;;
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

BINARY_PATH="$INSTALL_DIR/$TARGET_BINARY"
CONFIG_FILE="$CONFIG_DIR/config.yaml"
SERVICE_FILE="/etc/systemd/system/$SERVICE_NAME.service"

print_info "Starting Sentinel installation..."
print_info "Install directory: $INSTALL_DIR"
print_info "Config directory: $CONFIG_DIR"
print_info "Data directory: $DATA_DIR"
print_info "Service name: $SERVICE_NAME"
print_info "Service user: $SERVICE_USER"
print_info "Binary path: $BINARY_PATH"

# Check if running as root
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
    print_error "macOS is not supported for systemd installation"
    print_info "For macOS, please install manually and use launchd instead"
    exit 1
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

# Check if service already exists
if systemctl list-unit-files | grep -q "^$SERVICE_NAME.service"; then
    print_warning "Service '$SERVICE_NAME' already exists"
    read -p "Do you want to continue and overwrite? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "Installation cancelled"
        exit 0
    fi
    
    # Stop existing service
    print_info "Stopping existing service..."
    systemctl stop "$SERVICE_NAME" || true
fi

# Fetch release information from GitHub API
if [ -n "$VERSION" ]; then
    print_info "Fetching release information for version $VERSION..."
    API_URL="https://api.github.com/repos/${GITHUB_REPO}/releases/tags/$VERSION"
else
    print_info "Fetching latest release information..."
    API_URL="https://api.github.com/repos/${GITHUB_REPO}/releases/latest"
fi

RELEASE_INFO=$(curl --silent "$API_URL")

if [ -z "$RELEASE_INFO" ]; then
    print_error "Failed to fetch release information from GitHub"
    exit 1
fi

# Get the version
RELEASE_VERSION=$(echo "$RELEASE_INFO" | jq -r '.tag_name')
if [ -z "$RELEASE_VERSION" ] || [ "$RELEASE_VERSION" = "null" ]; then
    print_error "Failed to parse version from GitHub API"
    exit 1
fi

print_info "Installing version: $RELEASE_VERSION"

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
print_info "Verifying binary..."
BINARY_VERSION=$("$EXTRACTED_BINARY" version 2>/dev/null || echo "unknown")
print_info "Binary version: $BINARY_VERSION"

# Create system user if it doesn't exist
if ! id "$SERVICE_USER" >/dev/null 2>&1; then
    print_info "Creating system user '$SERVICE_USER'..."
    useradd --system --no-create-home --shell /bin/false "$SERVICE_USER"
    print_success "User '$SERVICE_USER' created"
else
    print_info "User '$SERVICE_USER' already exists"
fi

# Create directories
print_info "Creating directories..."
mkdir -p "$INSTALL_DIR"
mkdir -p "$CONFIG_DIR"
mkdir -p "$DATA_DIR"

# Set ownership and permissions
chown root:root "$INSTALL_DIR"
chown "$SERVICE_USER:$SERVICE_USER" "$CONFIG_DIR"
chown "$SERVICE_USER:$SERVICE_USER" "$DATA_DIR"
chmod 755 "$INSTALL_DIR"
chmod 755 "$CONFIG_DIR"
chmod 755 "$DATA_DIR"

print_success "Directories created and configured"

# Install the binary
print_info "Installing binary to $BINARY_PATH..."
cp "$EXTRACTED_BINARY" "$BINARY_PATH"
chmod +x "$BINARY_PATH"
chown root:root "$BINARY_PATH"

print_success "Binary installed successfully"

# Create configuration file if it doesn't exist
if [ ! -f "$CONFIG_FILE" ]; then
    print_info "Creating configuration file at $CONFIG_FILE..."
    cat > "$CONFIG_FILE" << EOF
server:
  port: 8080
  host: "0.0.0.0"
  base_host: "localhost:8080"
  auth:
    enabled: false  # Set to true to enable basic authentication
    users:
      - username: "admin"
        password: "changeme"
      - username: "viewer"
        password: "readonly"

monitoring:
  global:
    default_interval: 30s
    default_timeout: 10s
    default_retries: 3

timezone: UTC

database:
  path: "$DATA_DIR/db.sqlite"

notifications:
  enabled: false
  urls:
    # Telegram
    # - "telegram://token@telegram?chats=@channel-1[,chat-id-1,...]&preview=false"
EOF

    chown "$SERVICE_USER:$SERVICE_USER" "$CONFIG_FILE"
    chmod 644 "$CONFIG_FILE"
    print_success "Configuration file created"
else
    print_info "Configuration file already exists at $CONFIG_FILE"
fi

# Create systemd service file
print_info "Creating systemd service file at $SERVICE_FILE..."
cat > "$SERVICE_FILE" << EOF
[Unit]
Description=Sentinel Monitoring Service
Documentation=https://github.com/sxwebdev/sentinel
After=network.target
Wants=network.target

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_USER
ExecStart=$BINARY_PATH start --config $CONFIG_FILE
WorkingDirectory=$CONFIG_DIR

# Restart policy
Restart=always
RestartSec=5
StartLimitInterval=60s
StartLimitBurst=3

# Logging
SyslogIdentifier=$SERVICE_NAME

[Install]
WantedBy=multi-user.target
EOF

chmod 644 "$SERVICE_FILE"
print_success "Systemd service file created"

# Reload systemd and enable service
print_info "Reloading systemd daemon..."
systemctl daemon-reload

print_info "Enabling service '$SERVICE_NAME'..."
systemctl enable "$SERVICE_NAME"

# Start the service
print_info "Starting service '$SERVICE_NAME'..."
if ! systemctl start "$SERVICE_NAME"; then
    print_error "Failed to start service '$SERVICE_NAME'"
    print_info "Check logs with: journalctl -u $SERVICE_NAME -f"
    rm -rf "$TEMP_DIR"
    exit 1
fi

# Wait a moment for service to initialize
sleep 3

# Check service status
print_info "Checking service status..."
if systemctl is-active --quiet "$SERVICE_NAME"; then
    print_success "Service is running successfully"
    
    # Show service status
    echo ""
    print_info "Service status:"
    systemctl status "$SERVICE_NAME" --no-pager --lines=10
    
    echo ""
    print_info "Service logs (last 5 lines):"
    journalctl -u "$SERVICE_NAME" --no-pager --lines=5
    
else
    print_error "Service failed to start properly"
    print_info "Check logs with: journalctl -u $SERVICE_NAME -f"
    rm -rf "$TEMP_DIR"
    exit 1
fi

# Clean up temporary directory
print_info "Cleaning up temporary files..."
rm -rf "$TEMP_DIR"

# Final information
echo ""
print_success "Sentinel installation completed successfully!"
echo ""
print_info "Installation details:"
echo "  Version: $BINARY_VERSION"
echo "  Binary: $BINARY_PATH"
echo "  Config: $CONFIG_FILE"
echo "  Data: $DATA_DIR"
echo "  Service: $SERVICE_NAME"
echo "  User: $SERVICE_USER"
echo ""
print_info "Useful commands:"
echo "  View status:    systemctl status $SERVICE_NAME"
echo "  View logs:      journalctl -u $SERVICE_NAME -f"
echo "  Stop service:   sudo systemctl stop $SERVICE_NAME"
echo "  Start service:  sudo systemctl start $SERVICE_NAME"
echo "  Restart:        sudo systemctl restart $SERVICE_NAME"
echo "  Disable:        sudo systemctl disable $SERVICE_NAME"
echo ""
print_info "Web interface:"
echo "  URL: http://localhost:8080"
echo ""
print_info "Configuration:"
echo "  Edit: $CONFIG_FILE"
echo "  After changes: sudo systemctl restart $SERVICE_NAME"
echo ""
print_info "Upgrade:"
echo "  Run: curl -L https://raw.githubusercontent.com/sxwebdev/sentinel/master/scripts/upgrade.sh | sudo bash -s $SERVICE_NAME"

print_success "Happy monitoring! 🚀"
