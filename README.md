# Sentinel - Service Monitoring System

Sentinel is a lightweight, multi-protocol service monitoring system written in Go. It monitors HTTP/HTTPS, TCP, gRPC, and Redis services, providing real-time status updates and incident management with Telegram notifications.

## Features

- **Multi-Protocol Support**: HTTP/HTTPS, TCP, gRPC, Redis
- **Real-time Monitoring**: Configurable check intervals and timeouts
- **Incident Management**: Automatic incident creation and resolution
- **Telegram Notifications**: Alert and recovery notifications
- **Web Dashboard**: Clean, responsive web interface
- **REST API**: Full API for integration with other tools
- **Persistent Storage**: Incident history using SQLite
- **Configuration**: YAML-based configuration with environment variable support

## Quick Start

### Using Docker Compose (Recommended)

1. Clone the repository:

```bash
git clone https://github.com/sxwebdev/sentinel
cd sentinel
```

2. Create environment file:

```bash
cp .env.example .env
# Edit .env with your Telegram bot credentials
```

3. Start the services:

```bash
docker-compose up -d
```

4. Access the dashboard at http://localhost:8080

### Manual Installation

1. Install Go 1.21 or later
2. Clone and build:

```bash
git clone https://github.com/sxwebdev/sentinel
cd sentinel
go mod download
go build -o sentinel ./cmd/server
```

3. Configure your services in `config.yaml`
4. Set environment variables:

```bash
export TELEGRAM_BOT_TOKEN="your_bot_token"
export TELEGRAM_CHAT_ID="your_chat_id"
```

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

telegram:
  enabled: false
  bot_token: "${TELEGRAM_BOT_TOKEN}"
  chat_id: "${TELEGRAM_CHAT_ID}"

services:
  - name: "my-api"
    protocol: "http"
    endpoint: "https://api.example.com/health"
    interval: 60s
    timeout: 5s
    retries: 2
    tags: ["api", "critical"]
    config:
      method: "GET"
      expected_status: 200
```

### Protocol-Specific Configuration

#### HTTP/HTTPS

```yaml
- name: "web-service"
  protocol: "http"
  endpoint: "https://example.com/health"
  interval: 30s
  timeout: 10s
  retries: 3
  tags: ["web", "critical"]
  config:
    method: "GET" # HTTP method
    expected_status: 200 # Expected status code
    headers: # Optional: custom headers
      User-Agent: "Sentinel Monitor"
      Authorization: "Bearer token"
```

#### TCP

```yaml
- name: "database"
  protocol: "tcp"
  endpoint: "db.example.com:5432"
  interval: 30s
  timeout: 5s
  retries: 3
  tags: ["database", "postgres"]
  config:
    send_data: "ping" # Optional: data to send
    expect_data: "pong" # Optional: expected response
```

#### gRPC

```yaml
- name: "grpc-service"
  protocol: "grpc"
  endpoint: "grpc.example.com:50051"
  interval: 30s
  timeout: 10s
  retries: 3
  tags: ["grpc", "backend"]
  config:
    check_type: "health" # "health", "reflection", "connectivity"
    service_name: "myapp.MyService" # Optional: specific service name
    tls: true # Use TLS
    insecure_tls: false # Skip TLS verification
```

#### Redis

```yaml
- name: "redis-cache"
  protocol: "redis"
  endpoint: "redis.example.com:6379"
  interval: 30s
  timeout: 5s
  retries: 3
  tags: ["cache", "redis"]
  config:
    password: "${REDIS_PASSWORD}" # Optional: Redis password
    db: 0 # Redis database number
```

### gRPC Check Types

The gRPC monitor supports three types of checks:

1. **Health Check** (`check_type: "health"`): Uses standard gRPC health service
2. **Reflection Check** (`check_type: "reflection"`): Checks gRPC reflection availability
3. **Connectivity Check** (`check_type: "connectivity"`): Simple connection test

## Telegram Setup

1. Create a Telegram bot:

   - Message @BotFather on Telegram
   - Send `/newbot` and follow instructions
   - Save the bot token

2. Get your chat ID:

   - Add the bot to your group/channel
   - Send a message to the bot
   - Visit `https://api.telegram.org/bot<TOKEN>/getUpdates`
   - Find your chat ID in the response

3. Set environment variables:

```bash
export TELEGRAM_BOT_TOKEN="1234567890:ABCdefGHIjklMNOpqrsTUVwxyz"
export TELEGRAM_CHAT_ID="-1001234567890"
```

4. Enable Telegram in config:

```yaml
telegram:
  enabled: true
  bot_token: "${TELEGRAM_BOT_TOKEN}"
  chat_id: "${TELEGRAM_CHAT_ID}"
```

## API Reference

### Get All Services

```bash
GET /api/services
```

### Get Service Details

```bash
GET /api/services/{name}
```

### Get Service Incidents

```bash
GET /api/services/{name}/incidents
```

### Get Service Statistics

```bash
GET /api/services/{name}/stats?days=30
```

### Trigger Manual Check

```bash
POST /api/services/{name}/check
```

### Get Recent Incidents

```bash
GET /api/incidents?limit=50
```

## Web Interface

- **Dashboard** (`/`): Overview of all services
- **Service Detail** (`/service/{name}`): Detailed view with incident history
- **Auto-refresh**: Dashboard refreshes every 30 seconds

## Monitoring Logic

1. **Health Checks**: Each service is checked at configured intervals
2. **Retry Logic**: Failed checks are retried with exponential backoff
3. **State Changes**: Status changes trigger incident creation/resolution
4. **Notifications**: Alerts sent only on status changes (UP ↔ DOWN)

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
│   ├── storage/         # Data persistence (SQLite)
│   ├── notifier/        # Notification system
│   ├── scheduler/       # Monitoring scheduler
│   ├── service/         # Business logic
│   └── web/             # Web interface
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
Type=simple
User=sentinel
WorkingDirectory=/opt/sentinel
ExecStart=/opt/sentinel/sentinel
Restart=always
RestartSec=5
Environment=TELEGRAM_BOT_TOKEN=your_token
Environment=TELEGRAM_CHAT_ID=your_chat_id

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl enable sentinel
sudo systemctl start sentinel
```

### Docker

```bash
# Build image
docker build -t sentinel .

# Run container
docker run -d \
  --name sentinel \
  -p 8080:8080 \
  -e TELEGRAM_BOT_TOKEN="your_token" \
  -e TELEGRAM_CHAT_ID="your_chat_id" \
  -v ./data:/root/data \
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

- GitHub Issues: https://github.com/sxwebdev/sentinel/issues
- Documentation: https://github.com/sxwebdev/sentinel/wiki
