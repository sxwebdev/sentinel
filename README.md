# Sentinel - Service Monitoring System

Sentinel is a lightweight, multi-protocol service monitoring system written in Go. It monitors HTTP/HTTPS, TCP, gRPC, and Redis services, providing real-time status updates and incident management with multi-provider notifications.

## Features

- **Multi-Protocol Support**: HTTP/HTTPS, TCP, gRPC, Redis
- **Real-time Monitoring**: Configurable check intervals and timeouts
- **Incident Management**: Automatic incident creation and resolution
- **Multi-Provider Notifications**: Alert and recovery notifications via multiple providers (Telegram, Discord, Slack, Email, Webhooks, etc.)
- **Web Dashboard**: Clean, responsive web interface with YAML configuration
- **REST API**: Full API for integration with other tools
- **WebSocket Support**: Real-time updates via WebSocket connections
- **Persistent Storage**: Incident history using SQLite with improved concurrency handling
- **Configuration**: YAML-based configuration with environment variable support
- **Clean Logging**: Minimal logging focused on check results and notifications

## Quick Start

### Using Docker Compose (Recommended)

1. Clone the repository:

```bash
git clone https://github.com/sxwebdev/sentinel
cd sentinel
```

2. Create configuration file:

```bash
cp config.yaml.example config.yaml
# Edit config.yaml with your notification provider credentials
```

3. Start the services:

```bash
docker-compose up -d
```

4. Access the dashboard at http://localhost:8080

### Manual Installation

1. Install Go 1.24 or later
2. Clone and build:

```bash
git clone https://github.com/sxwebdev/sentinel
cd sentinel
go mod download
go build -o sentinel ./cmd/server
```

3. Configure your services in `config.yaml`
4. Set up notification providers (see Notification Setup section below)
5. Run:

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

monitoring:
  global:
    default_interval: 30s
    default_timeout: 10s
    default_retries: 3

database:
  path: "./data/db.sqlite3"

notifications:
  enabled: true
  urls:
    # Telegram
    - "telegram://token@telegram?chats=@channel-1[,chat-id-1,...]"
    # Discord (optional)
    - "discord://token@id"
    # Slack (optional)
    - "slack://[botname@]token-a/token-b/token-c"
    # Email (optional)
    - "smtp://username:password@host:port/?from=fromAddress&to=recipient1[,recipient2,...]"
```

**Note**: Services are now managed through the web interface or API, not through static configuration in `config.yaml`.

## Web Interface

### Dashboard Features

- **Service Management**: Add, edit, and delete services through the web interface
- **YAML Configuration**: Configure services using YAML format directly in the UI
- **Real-time Status**: Live status updates with WebSocket connections
- **Manual Checks**: Trigger manual health checks for any service
- **Incident Management**: View and resolve incidents
- **Service Details**: Detailed view with incident history and statistics

### Service Configuration via UI

When creating or editing services through the web interface:

1. **Protocol Selection**: Choose from HTTP/HTTPS, TCP, gRPC, or Redis
2. **YAML Configuration**: Enter protocol-specific configuration in YAML format
3. **Default Templates**: UI provides default YAML templates for each protocol
4. **Validation**: Configuration is validated before saving

Example YAML configurations for each protocol:

**HTTP/HTTPS:**

```yaml
method: "GET"
expected_status: 200
headers:
  User-Agent: "Sentinel Monitor"
  Authorization: "Bearer token"
```

**TCP:**

```yaml
send_data: "ping"
expect_data: "pong"
```

**gRPC:**

```yaml
check_type: "connectivity"
service_name: ""
tls: true
insecure_tls: false
```

**Redis:**

```yaml
password: "your_password"
db: 0
```

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

Sentinel uses [Shoutrrr](https://github.com/containrrr/shoutrrr) for notifications, which supports multiple providers:

### Telegram Setup

1. Create a Telegram bot:

   - Message @BotFather on Telegram
   - Send `/newbot` and follow instructions
   - Save the bot token

2. Get your chat ID or channel username:

   - For private chats: Add the bot to your group/channel, send a message, then visit `https://api.telegram.org/bot<TOKEN>/getUpdates`
   - For public channels: Use the channel username (e.g., `@mychannel`)

3. Configure in `config.yaml`:

```yaml
notifications:
  enabled: true
  urls:
    - "telegram://token@telegram?chats=@channel-1[,chat-id-1,...]"
```

### Discord Setup

1. Create a Discord webhook in your server settings
2. Use the webhook URL format: `discord://token@id`

### Slack Setup

1. Create a Slack app and get the tokens
2. Use the Slack URL format: `slack://[botname@]token-a/token-b/token-c`

### Email Setup

1. Configure SMTP settings
2. Use the SMTP URL format: `smtp://username:password@host:port/?from=fromAddress&to=recipient1[,recipient2,...]`

### Multiple Providers

You can configure multiple notification providers simultaneously. If one provider fails, notifications will still be sent to the others:

```yaml
notifications:
  enabled: true
  urls:
    # Telegram
    - "telegram://token@telegram?chats=@channel-1[,chat-id-1,...]"
    # Discord
    - "discord://token@id"
    # Slack
    - "slack://[botname@]token-a/token-b/token-c"
    # Email
    - "smtp://username:password@host:port/?from=fromAddress&to=recipient1[,recipient2,...]"
```

## API Reference

### Get All Services

```bash
GET /api/services
```

### Get Services Table Data

```bash
GET /api/services/table
```

Returns services with incident statistics for table display.

### Get Service Details

```bash
GET /api/services/{id}
```

### Get Service Configuration

```bash
GET /api/services/config/{id}
```

### Get Service Incidents

```bash
GET /api/services/{id}/incidents
```

### Get Service Statistics

```bash
GET /api/services/{id}/stats?days=30
```

### Trigger Manual Check

```bash
POST /api/services/{id}/check
```

### Resolve Incidents

```bash
POST /api/services/{id}/resolve
```

### Create Service

```bash
POST /api/services
Content-Type: application/json

{
  "name": "my-service",
  "protocol": "http",
  "endpoint": "https://example.com/health",
  "interval": "30s",
  "timeout": "10s",
  "retries": 3,
  "tags": ["api", "critical"],
  "config": "method: \"GET\"\nexpected_status: 200"
}
```

### Update Service

```bash
PUT /api/services/{id}
Content-Type: application/json

{
  "name": "my-service",
  "protocol": "http",
  "endpoint": "https://example.com/health",
  "interval": "30s",
  "timeout": "10s",
  "retries": 3,
  "tags": ["api", "critical"],
  "config": "method: \"GET\"\nexpected_status: 200"
}
```

### Delete Service

```bash
DELETE /api/services/{id}
```

### Get Recent Incidents

```bash
GET /api/incidents?limit=50
```

### Get Dashboard Statistics

```bash
GET /api/dashboard/stats
```

Returns overall dashboard statistics including uptime percentage and average response time.

## WebSocket API

### Connect to WebSocket

```javascript
const ws = new WebSocket("ws://localhost:8080/ws");
```

### Message Format

```json
{
  "type": "service_update",
  "services": [
    {
      "id": "service-id",
      "name": "Service Name",
      "protocol": "http",
      "endpoint": "https://example.com",
      "interval": "30s",
      "timeout": "10s",
      "retries": 3,
      "tags": ["api", "critical"],
      "config": "method: \"GET\"\nexpected_status: 200",
      "state": {
        "status": "up",
        "last_check": "2024-01-01T12:00:00Z",
        "response_time": "150ms",
        "consecutive_success": 10,
        "total_checks": 100
      },
      "active_incidents": 0,
      "total_incidents": 5
    }
  ],
  "timestamp": 1704110400
}
```

## Monitoring Logic

1. **Health Checks**: Each service is checked at configured intervals
2. **Retry Logic**: Failed checks are retried with exponential backoff
3. **State Changes**: Status changes trigger incident creation/resolution
4. **Notifications**: Alerts sent only on status changes (UP ↔ DOWN)
5. **Real-time Updates**: WebSocket broadcasts for instant UI updates
6. **Concurrent Access**: Improved SQLite handling for concurrent operations
7. **Clean Logging**: Minimal logging focused on check results and notifications

## Development

### Project Structure

```
sentinel/
├── cmd/
│   ├── server/          # Main application
│   ├── tcpserver/       # TCP server for testing
│   └── grpcserver/      # gRPC server for testing
├── internal/
│   ├── config/          # Configuration management
│   ├── monitors/        # Protocol-specific monitors
│   ├── storage/         # Data persistence (SQLite with retry logic)
│   ├── notifier/        # Notification system
│   ├── receiver/        # Event receiver for real-time updates
│   ├── scheduler/       # Monitoring scheduler
│   ├── service/         # Business logic
│   └── web/             # Web interface with YAML configuration
├── data/                # SQLite database files
├── config.yaml          # Configuration file
└── docker-compose.yml   # Docker services
```

### Building

```bash
# Build all binaries
make build

# Build specific binary
go build -o sentinel ./cmd/server
go build -o tcpserver ./cmd/tcpserver
go build -o grpcserver ./cmd/grpcserver

# Production build with optimizations
CGO_ENABLED=0 go build -ldflags="-w -s" -o sentinel ./cmd/server

# Cross-compile for Linux
GOOS=linux GOARCH=amd64 go build -o sentinel-linux ./cmd/server
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test ./internal/storage
```

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

## Logging

Sentinel uses minimal logging focused on essential information:

- **Check Results**: Success/failure of each service check
- **Notifications**: Success/failure of notification delivery
- **Errors**: Critical errors that require attention

Debug logging has been removed to keep logs clean and focused.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - see LICENSE file for details.

## Support

- GitHub Issues: https://github.com/sxwebdev/sentinel/issues
- Documentation: https://github.com/sxwebdev/sentinel/wiki
