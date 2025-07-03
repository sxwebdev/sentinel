# gRPC Monitor

The gRPC monitor supports various types of gRPC service checks.

## Check Types

### 1. Health Check (`check_type: "health"`)

Standard gRPC health service check. Uses `grpc.health.v1.Health` service.

**Configuration:**

```yaml
- name: "grpc-health"
  protocol: "grpc"
  endpoint: "localhost:50051"
  interval: 30s
  timeout: 10s
  retries: 3
  config:
    check_type: "health"
    service_name: "grpc.health.v1.Health" # Optional
    tls: false
```

**Parameters:**

- `service_name` - specific service name to check (optional)
- `tls` - use TLS connection
- `server_name` - server name for TLS (if different from endpoint)
- `insecure_tls` - skip certificate verification

### 2. Reflection Check (`check_type: "reflection"`)

Check gRPC reflection service availability. In current implementation uses simple connectivity check.

**Configuration:**

```yaml
- name: "grpc-reflection"
  protocol: "grpc"
  endpoint: "localhost:50051"
  interval: 30s
  timeout: 10s
  retries: 3
  config:
    check_type: "reflection"
    tls: false
```

### 3. Connectivity Check (`check_type: "connectivity"`)

Simple gRPC server connectivity check. Useful for services that don't implement health or reflection API.

**Configuration:**

```yaml
- name: "grpc-connectivity"
  protocol: "grpc"
  endpoint: "localhost:50051"
  interval: 30s
  timeout: 10s
  retries: 3
  config:
    check_type: "connectivity"
    tls: false
```

## Usage Examples

### Check standard health service

```yaml
- name: "my-grpc-service"
  protocol: "grpc"
  endpoint: "my-service.example.com:50051"
  interval: 30s
  timeout: 10s
  retries: 3
  config:
    check_type: "health"
    tls: true
    server_name: "my-service.example.com"
```

### Check specific service

```yaml
- name: "user-service"
  protocol: "grpc"
  endpoint: "user-service:50051"
  interval: 30s
  timeout: 10s
  retries: 3
  config:
    check_type: "health"
    service_name: "user.UserService"
    tls: false
```

### Check reflection

```yaml
- name: "grpc-reflection"
  protocol: "grpc"
  endpoint: "localhost:50051"
  interval: 30s
  timeout: 10s
  retries: 3
  config:
    check_type: "reflection"
```

### Simple connectivity check

```yaml
- name: "grpc-connectivity"
  protocol: "grpc"
  endpoint: "localhost:50051"
  interval: 30s
  timeout: 10s
  retries: 3
  config:
    check_type: "connectivity"
```

## Features

1. **Non-blocking connection**: gRPC monitor doesn't block when creating connection
2. **Automatic reconnection**: gRPC client automatically reconnects on connection loss
3. **Timeouts**: All checks are performed with configured timeout
4. **TLS support**: Full TLS connection support with certificate configuration
5. **Modern API**: Uses `grpc.NewClient` instead of deprecated `grpc.Dial`

## Connection States

The monitor checks the following gRPC connection states:

- `Ready` - connection is ready for use
- `Idle` - connection is in idle mode
- `Connecting` - connection is being established
- `TransientFailure` - temporary connection failure
- `Shutdown` - connection is closed

## Limitations

- Reflection check in current implementation is simplified and uses connectivity check
- Full reflection API support requires more complex implementation
- No support for custom gRPC methods (requires dynamic calls)
