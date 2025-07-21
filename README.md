# Sentinel - Service Monitoring System

Sentinel is a lightweight, multi-protocol service monitoring system written in Go. It monitors HTTP/HTTPS, TCP and gRPC services, providing real-time status updates and incident management with multi-provider notifications.

![Preview](https://github.com/sxwebdev/sentinel/blob/master/screenshots/dashboard.png?raw=true)

![Preview](https://github.com/sxwebdev/sentinel/blob/master/screenshots/service_detail.png?raw=true)

## Features

- **Multi-Protocol Support**: HTTP/HTTPS, TCP, gRPC
- **Real-time Monitoring**: Configurable check intervals and timeouts
- **Incident Management**: Automatic incident creation and resolution
- **Multi-Provider Notifications**: Alert and recovery notifications via multiple providers (Telegram, Discord, Slack, Email, Webhooks, etc.)
- **Web Dashboard**: Clean, responsive web interface with JSON configuration
- **REST API**: Full API for integration with other tools
- **WebSocket Support**: Real-time updates via WebSocket connections
- **Persistent Storage**: Incident history using SQLite with improved concurrency handling
- **Configuration**: YAML-based configuration with environment variable support
- **Clean Logging**: Minimal logging focused on check results and notifications

## Quick Start

### Automated Installation (Recommended for Linux)

Install Sentinel automatically on Linux systems with systemd:

```bash
# Install with default settings
curl -L https://raw.githubusercontent.com/sxwebdev/sentinel/master/scripts/install.sh | sudo bash

# Install with custom options
curl -L https://raw.githubusercontent.com/sxwebdev/sentinel/master/scripts/install.sh | sudo bash -s -- \
    --install-dir /opt/sentinel/bin \
    --config-dir /etc/sentinel \
    --data-dir /var/lib/sentinel \
    --service-name sentinel-monitor \
    --user sentinel

# Install specific version
curl -L https://raw.githubusercontent.com/sxwebdev/sentinel/master/scripts/install.sh | sudo bash -s -- \
    --version v1.0.0
```

**Installation Features:**

- Automatic binary download for your platform (Linux amd64/arm64)
- System user creation
- Systemd service configuration with security hardening
- Configuration file creation with sensible defaults
- Automatic service startup and enablement

**Post-Installation:**

- Dashboard: <http://localhost:8080>
- Configuration: `/etc/sentinel/config.yaml`
- Service: `systemctl status sentinel`
- Logs: `journalctl -u sentinel -f`

### Using Docker Compose

1. Clone the repository:

```bash
git clone https://github.com/sxwebdev/sentinel
cd sentinel
```

1. Create configuration file:

```bash
cp config.template.yaml config.yaml
# Edit config.yaml with your notification provider credentials
```

1. Start the services:

```bash
docker-compose up -d
```

1. Access the dashboard at <http://localhost:8080>

### Manual Installation

1. Install Go 1.24 or later
1. Clone and build:

```bash
git clone https://github.com/sxwebdev/sentinel
cd sentinel
go mod download
go build -o sentinel ./cmd/sentinel
```

1. Configure your services in `config.yaml`
1. Set up notification providers (see Notification Setup section below)
1. Run:

```bash
./sentinel
```

## Configuration

### Basic Configuration

Create a `config.yaml` file:

```yaml
server:
  port: 8080
  host: "0.0.0.0"
  auth:
    enabled: false # Set to true to enable basic authentication
    users:
      - username: "admin"
        password: "changeme"

monitoring:
  global:
    default_interval: 30s
    default_timeout: 10s
    default_retries: 3

database:
  path: "./data/db.sqlite"

notifications:
  enabled: true
  urls:
    # Telegram
    - "telegram://token@telegram?chats=@channel-1[,chat-id-1,...]&preview=false"
    # Discord (optional)
    - "discord://token@id"
    # Slack (optional)
    - "slack://[botname@]token-a/token-b/token-c"
    # Email (optional)
    - "smtp://username:password@host:port/?from=fromAddress&to=recipient1[,recipient2,...]"
```

**Note**: Services are now managed through the web interface or API, not through static configuration in `config.yaml`.

### Security

Sentinel supports HTTP Basic Authentication to protect your monitoring dashboard and API endpoints.

#### Enabling Authentication

Set `auth.enabled: true` in your `config.yaml` and configure users:

```yaml
server:
  auth:
    enabled: true
    users:
      - username: "admin"
        password: "secure_password_123"
```

#### Security Features

- **HTTP Basic Auth**: Industry-standard authentication method
- **WebSocket Bypass**: WebSocket connections bypass auth for real-time updates
- **API Protection**: All REST API endpoints require authentication when enabled
- **Swagger Integration**: API documentation includes authentication requirements

#### Security Best Practices

- Use strong, unique passwords for each user
- Consider using environment variables for passwords in production
- Enable authentication in production environments
- Regularly rotate passwords
- Use HTTPS in production (configure reverse proxy)

## Web Interface

### Dashboard Features

- **Service Management**: Add, edit, and delete services through the web interface
- **JSON Configuration**: Configure services using JSON format directly in the UI
- **Real-time Status**: Live status updates with WebSocket connections
- **Manual Checks**: Trigger manual health checks for any service
- **Incident Management**: View and resolve incidents
- **Service Details**: Detailed view with incident history and statistics

### HTTP Monitor Features

![Preview](https://github.com/sxwebdev/sentinel/blob/master/screenshots/create_http_service.png?raw=true)

The HTTP monitor provides advanced capabilities for complex monitoring scenarios:

- **Multi-Endpoint Monitoring**: Monitor multiple endpoints within a single service and compare their responses
- **JSON Path Extraction**: Extract specific values from JSON responses using JSONPath syntax
- **JavaScript Conditions**: Set custom alert conditions using JavaScript to analyze responses from multiple endpoints
- **Basic Authentication**: Support for HTTP Basic Auth on per-endpoint basis
- **Custom Headers**: Configure custom HTTP headers for each endpoint

### TCP Monitor Features

- **Simple Connectivity**: Basic TCP port connectivity checks
- **Send/Expect Protocol**: Send specific data and validate expected responses
- **Custom Protocol Support**: Monitor any TCP-based protocol (Redis, MySQL, custom protocols)

### gRPC Check Types

The gRPC monitor supports three types of checks:

1. **Health Check** (`check_type: "health"`): Uses standard gRPC health service (`grpc.health.v1.Health`)
2. **Reflection Check** (`check_type: "reflection"`): Checks gRPC reflection availability
3. **Connectivity Check** (`check_type: "connectivity"`): Simple connection test

**Available check types:**

- `health` - Standard gRPC health service check
- `reflection` - gRPC reflection service check
- `connectivity` - Basic connectivity test

## Notification Setup

Sentinel uses [Shoutrrr](https://github.com/containrrr/shoutrrr) for notifications, which supports multiple providers

You can configure multiple notification providers simultaneously. If one provider fails, notifications will still be sent to the others:

```yaml
notifications:
  enabled: true
  urls:
    # Telegram
    - "telegram://token@telegram?chats=@channel-1[,chat-id-1,...]&preview=false"
    # Discord
    - "discord://token@id"
    # Slack
    - "slack://[botname@]token-a/token-b/token-c"
    # Email
    - "smtp://username:password@host:port/?from=fromAddress&to=recipient1[,recipient2,...]"
```

## Upgrading

### Automatic Upgrade Script

Sentinel includes an upgrade script that automatically downloads and installs the latest release with zero-downtime deployment:

```bash
# Download and run the upgrade script
curl -L -o ./upgrade.sh https://raw.githubusercontent.com/sxwebdev/sentinel/master/scripts/upgrade.sh
chmod +x ./upgrade.sh
sudo ./upgrade.sh sentinel /usr/local/bin/sentinel

# Or if you have the repository cloned
sudo ./scripts/upgrade.sh sentinel /usr/local/bin/sentinel
```

### Upgrade Features

- 🔄 **Automatic Download**: Fetches the latest release for your platform (Linux/macOS, amd64/arm64)
- 🛡️ **Safe Deployment**: Creates backup and rollback on failure
- 🔧 **Service Management**: Handles systemd service stop/start automatically
- ✅ **Verification**: Tests new binary before deployment
- 📊 **Status Check**: Shows service status after upgrade

### Custom Binary Path

```bash
# For custom installation paths
sudo ./scripts/upgrade.sh sentinel-monitoring /opt/sentinel/bin/sentinel
```

### Manual Upgrade

1. Download the latest release for your platform from [GitHub Releases](https://github.com/sxwebdev/sentinel/releases)
2. Stop the service: `sudo systemctl stop sentinel`
3. Replace the binary: `sudo cp sentinel /usr/local/bin/sentinel`
4. Start the service: `sudo systemctl start sentinel`
5. Check status: `systemctl status sentinel`

### Health Monitoring

After installation or upgrade, monitor service health:

```bash
# Check service status
systemctl status sentinel

# View recent logs
journalctl -u sentinel --since "10 minutes ago"

# Monitor in real-time
journalctl -u sentinel -f

# Check web interface
curl -f http://localhost:8080/health || echo "Service not responding"
```

## API Documentation

Sentinel provides a comprehensive REST API and WebSocket support for programmatic access to all monitoring features.

**Interactive API Documentation**: <http://localhost:8080/api/v1/swagger/index.html>

The Swagger UI provides:

- Complete endpoint documentation with examples
- Request/response schemas
- Authentication setup
- Interactive testing interface
- WebSocket API documentation

## Monitoring Logic

1. **Health Checks**: Each service is checked at configured intervals
2. **Retry Logic**: Failed checks are retried with exponential backoff
3. **State Changes**: Status changes trigger incident creation/resolution
4. **Notifications**: Alerts sent only on status changes (UP ↔ DOWN)
5. **Real-time Updates**: WebSocket broadcasts for instant UI updates

## Development

## Deployment

### Docker Compose (Recommended)

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

### Systemd Service

Create `/etc/systemd/system/sentinel.service`:

```ini
[Unit]
Description=Sentinel Service Monitor
After=network.target

[Service]
User=sentinel
WorkingDirectory=/opt/sentinel
ExecStart=/opt/sentinel/sentinel
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable sentinel
sudo systemctl start sentinel
```

### Docker

```bash
# Build image
docker build -t sentinel .

# Or pull
docker pull sxwebdev/sentinel:latest

# Run container
docker run -d \
  --name sentinel \
  -p 8080:8080 \
  -v ./data:/root/data \
  -v config.yaml:/root/config.yaml
  sentinel
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - see LICENSE file for details.

## Support

- GitHub Issues: <https://github.com/sxwebdev/sentinel/issues>
- Documentation: <https://github.com/sxwebdev/sentinel/wiki>
