# gRPC Monitor

The gRPC monitor supports various types of gRPC service checks.

## Check Types

### 1. Health Check (`check_type: "health"`)

Standard gRPC health service check. Uses `grpc.health.v1.Health` service.

**Configuration:**

```json
{
  "name": "grpc-health",
  "protocol": "grpc",
  "endpoint": "localhost:50051",
  "interval": 30,
  "timeout": 10,
  "retries": 3,
  "config": {
    "check_type": "health",
    "service_name": "grpc.health.v1.Health",
    "tls": false
  }
}
```

**Parameters:**

- `service_name` - specific service name to check (optional)
- `tls` - use TLS connection
- `server_name` - server name for TLS (if different from endpoint)
- `insecure_tls` - skip certificate verification

### 2. Reflection Check (`check_type: "reflection"`)

Check gRPC reflection service availability. In current implementation uses simple connectivity check.

**Configuration:**

```json
{
  "name": "grpc-reflection",
  "protocol": "grpc",
  "endpoint": "localhost:50051",
  "interval": 30,
  "timeout": 10,
  "retries": 3,
  "config": {
    "check_type": "reflection",
    "tls": false
  }
}
```

### 3. Connectivity Check (`check_type: "connectivity"`)

Simple gRPC server connectivity check. Useful for services that don't implement health or reflection API.

**Configuration:**

```json
{
  "name": "grpc-connectivity",
  "protocol": "grpc",
  "endpoint": "localhost:50051",
  "interval": 30,
  "timeout": 10,
  "retries": 3,
  "config": {
    "check_type": "connectivity",
    "tls": false
  }
}
```

## Usage Examples

### Check standard health service

```json
{
  "name": "my-grpc-service",
  "protocol": "grpc",
  "endpoint": "my-service.example.com:50051",
  "interval": 30,
  "timeout": 10,
  "retries": 3,
  "config": {
    "check_type": "health",
    "tls": true,
    "server_name": "my-service.example.com"
  }
}
```

### Check specific service

```json
{
  "name": "user-service",
  "protocol": "grpc",
  "endpoint": "user-service:50051",
  "interval": 30,
  "timeout": 10,
  "retries": 3,
  "config": {
    "check_type": "health",
    "service_name": "user.UserService",
    "tls": false
  }
}
```

### Check reflection

```json
{
  "name": "grpc-reflection",
  "protocol": "grpc",
  "endpoint": "localhost:50051",
  "interval": 30,
  "timeout": 10,
  "retries": 3,
  "config": {
    "check_type": "reflection"
  }
}
```

### Simple connectivity check

```json
{
  "name": "grpc-connectivity",
  "protocol": "grpc",
  "endpoint": "localhost:50051",
  "interval": 30,
  "timeout": 10,
  "retries": 3,
  "config": {
    "check_type": "connectivity"
  }
}
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

# HTTP Multi-Endpoint Monitoring

Extension of HTTP monitoring for checking multiple endpoints with dynamic conditions.

## Features

- Monitor multiple endpoints simultaneously
- Extract values from JSON responses using JSONPath
- Dynamic conditions in JavaScript
- Compare values between endpoints
- Flexible configuration through web interface
- Basic Auth support for each endpoint

## Usage Examples

### 1. Blockchain Node Monitoring

**Scenario**: Check synchronization between primary and backup Ethereum nodes.

**Configuration**:

```json
{
  "protocol": "http",
  "endpoint": "https://main-node.example.com",
  "config": {
    "method": "GET",
    "expected_status": 200,
    "headers": {
      "Authorization": "Bearer token"
    },
    "multi_endpoint": {
      "endpoints": [
        {
          "name": "main",
          "url": "https://main-node.example.com/eth/v1/beacon/states/head",
          "method": "GET",
          "json_path": "data.slot",
          "headers": {
            "Authorization": "Bearer main-token"
          },
          "username": "admin",
          "password": "secret123"
        },
        {
          "name": "backup",
          "url": "https://backup-node.example.com/eth/v1/beacon/states/head",
          "method": "GET",
          "json_path": "data.slot",
          "headers": {
            "Authorization": "Bearer backup-token"
          },
          "username": "monitor",
          "password": "backup456"
        }
      ],
      "condition": "// Check slot difference between nodes\nif (!results.main.success || !results.backup.success) {\n  return true; // Incident if one of the nodes is unavailable\n}\n\nconst mainSlot = Number(results.main.value);\nconst backupSlot = Number(results.backup.value);\n\n// Incident if difference is more than 5 slots\nreturn Math.abs(mainSlot - backupSlot) > 5;",
      "timeout": 30
    }
  }
}
```

### 2. API Version Monitoring

**Scenario**: Check API version compatibility between production and staging environments.

**Configuration**:

```json
{
  "protocol": "http",
  "endpoint": "https://api.example.com/health",
  "config": {
    "multi_endpoint": {
      "endpoints": [
        {
          "name": "production",
          "url": "https://api.example.com/health",
          "method": "GET",
          "json_path": "version",
          "username": "prod_user",
          "password": "prod_pass"
        },
        {
          "name": "staging",
          "url": "https://staging-api.example.com/health",
          "method": "GET",
          "json_path": "version",
          "username": "stage_user",
          "password": "stage_pass"
        }
      ],
      "condition": "// Check version compatibility\nif (!results.production.success || !results.staging.success) {\n  return true;\n}\n\nconst prodVersion = results.production.value;\nconst stagingVersion = results.staging.value;\n\n// Incident if versions don't match\nreturn prodVersion !== stagingVersion;"
    }
  }
}
```

### 3. Load Balancing Monitoring

**Scenario**: Check load distribution uniformity between servers.

**Configuration**:

```json
{
  "protocol": "http",
  "endpoint": "https://loadbalancer.example.com/status",
  "config": {
    "multi_endpoint": {
      "endpoints": [
        {
          "name": "server1",
          "url": "https://server1.example.com/metrics",
          "method": "GET",
          "json_path": "load_average",
          "username": "monitor",
          "password": "metrics123"
        },
        {
          "name": "server2",
          "url": "https://server2.example.com/metrics",
          "method": "GET",
          "json_path": "load_average",
          "username": "monitor",
          "password": "metrics123"
        },
        {
          "name": "server3",
          "url": "https://server3.example.com/metrics",
          "method": "GET",
          "json_path": "load_average",
          "username": "monitor",
          "password": "metrics123"
        }
      ],
      "condition": "// Check load balance\nconst loads = [];\nlet totalLoad = 0;\n\nfor (const [name, result] of Object.entries(results)) {\n  if (!result.success) {\n    return true; // Incident if server is unavailable\n  }\n  const load = Number(result.value);\n  loads.push(load);\n  totalLoad += load;\n}\n\nconst avgLoad = totalLoad / loads.length;\n\n// Incident if difference from average is more than 20%\nfor (const load of loads) {\n  if (Math.abs(load - avgLoad) / avgLoad > 0.2) {\n    return true;\n  }\n}\n\nreturn false;"
    }
  }
}
```

## Basic Auth Support

Each endpoint in multi-endpoint configuration supports Basic Authentication:

```json
{
  "name": "secure-endpoint",
  "url": "https://api.example.com/secure",
  "method": "GET",
  "json_path": "status",
  "username": "monitor_user",
  "password": "secure_password"
}
```

The monitor will automatically add the `Authorization: Basic <base64-encoded-credentials>` header to requests.

## JavaScript API

The following variables are available in conditions:

### `results`

Object with results from all endpoints:

```javascript
{
  "endpoint_name": {
    "success": true,           // Request success
    "value": 12345,           // Extracted value (if JSONPath is specified)
    "error": "",              // Error (if any)
    "response": "...",        // Full response
    "duration": 150           // Execution time in milliseconds
  }
}
```

### `console`

For debugging, `console.log()` is available:

```javascript
console.log("Main node slot:", results.main.value);
console.log("Backup node slot:", results.backup.value);
```

## JSONPath Syntax

Simple JSONPath syntax is supported:

- `field` - access object field
- `0`, `1`, `2` - access array element
- `field.subfield` - nested fields
- `items.0.name` - combined access

**Examples**:

- `result.block_number` → `{"result": {"block_number": 12345}}`
- `data.0.id` → `{"data": [{"id": "abc"}]}`
- `status` → `{"status": "ok"}`

## Configuration via Web Interface

1. Create a new service with HTTP protocol
2. Enable "Multi-Endpoint Monitoring"
3. Add endpoints with their configuration including Basic Auth credentials
4. Write JavaScript condition
5. Save the service

## Limitations

- Maximum 10 endpoints per service
- JavaScript execution timeout: 5 seconds
- Maximum response size: 1MB
- Only HTTP/HTTPS endpoints are supported
- Basic Auth credentials are stored in plain text (consider using environment variables for production)
