# Sentinel API Integration Tests

This package contains comprehensive integration tests for the Sentinel monitoring system API.

## What is tested

### Core API functions:

- ✅ Dashboard and statistics
- ✅ Service CRUD operations (Create, Read, Update, Delete)
- ✅ Service filtering by various parameters
- ✅ Result pagination
- ✅ Incident management
- ✅ Tag operations
- ✅ Service statistics
- ✅ Manual service checks
- ✅ Error handling

### Monitoring protocols:

- ✅ HTTP/HTTPS monitoring with support for:
  - Multiple endpoints
  - Various HTTP methods (GET, POST, PUT, DELETE)
  - Headers and request body
  - Basic authentication
  - JSON Path validation
  - Execution conditions (all/any)
- ✅ TCP monitoring with data sending and receiving
- ✅ gRPC monitoring with various check types

### Filters and query parameters:

- ✅ Filtering by service name
- ✅ Filtering by tags (single and multiple)
- ✅ Filtering by status (up/down/unknown)
- ✅ Filtering by enabled status (enabled/disabled)
- ✅ Filtering by protocol (http/tcp/grpc)
- ✅ Sorting by name and creation date
- ✅ Pagination with configurable page size
- ✅ Incident filtering by time
- ✅ Incident filtering by resolution status

### Data model validation:

- ✅ Configuration structures for all protocols
- ✅ DTO conversions
- ✅ API response models
- ✅ Error models
- ✅ Pagination structures

### Error handling:

- ✅ Invalid JSON data
- ✅ Missing required fields
- ✅ Invalid parameter values
- ✅ Operations on non-existent resources
- ✅ Configuration validation errors

## Test structure

### Files:

- `main.go` - Main file with basic tests and test environment setup
- `extended_tests.go` - Extended tests for complex scenarios
- `model_tests.go` - Data model validation tests

### Test cases:

#### Basic tests (main.go):

1. **TestHealthCheck** - main page availability check
2. **TestDashboardStats** - dashboard statistics API test
3. **TestCreateServices** - creating test services of all types
4. **TestGetServices** - getting service list
5. **TestServiceFilters** - testing service filters
6. **TestServiceDetail** - getting detailed service information
7. **TestUpdateService** - service update
8. **TestServiceStats** - getting service statistics
9. **TestServiceCheck** - manual service check trigger
10. **TestIncidents** - incident operations
11. **TestIncidentFilters** - incident filtering
12. **TestTags** - tag operations
13. **TestPagination** - basic pagination testing
14. **TestErrorHandling** - error handling
15. **TestDeleteService** - service deletion

#### Extended tests (extended_tests.go):

1. **TestAdvancedServiceFilters** - complex filter combinations
2. **TestServiceCRUDCompleteFlow** - complete service lifecycle
3. **TestAdvancedIncidentManagement** - advanced incident management
4. **TestCompleteProtocolConfigurations** - testing all protocols
5. **TestAdvancedPaginationAndSorting** - advanced pagination and sorting
6. **TestAdvancedErrorScenarios** - complex error scenarios
7. **TestStatsWithDifferentParameters** - statistics with various parameters

#### Model tests (model_tests.go):

1. **TestModelsValidation** - monitoring configuration validation
2. **TestServiceDTOFields** - service DTO field verification
3. **TestIncidentFields** - incident field verification
4. **TestResponseModels** - response model verification
5. **TestServiceStatsModel** - service statistics model verification
6. **TestPaginationResponseModel** - paginated response model verification

## Running tests

### Prerequisites:

- Go 1.21+
- Sentinel project dependencies must be installed

### Run command:

```bash
cd cmd/testapi
go run *.go test
```

### Expected output:

```
Running TestHealthCheck...
PASS: TestHealthCheck
Running TestDashboardStats...
PASS: TestDashboardStats
...
Running TestPaginationResponseModel...
PASS: TestPaginationResponseModel

All tests passed!
```

### In case of errors:

```
Running TestCreateServices...
FAIL: TestCreateServices - service 0: HTTP 400: Service name is required
...

2 test(s) failed
```

## Test data

Tests create the following test services:

1. **HTTP Test Service 1** (enabled)

   - Protocol: HTTP
   - Endpoint: https://httpbin.org/status/200
   - Tags: [http, production, api]

2. **HTTP Test Service 2** (disabled)

   - Protocol: HTTP
   - Endpoint: https://httpbin.org/status/404
   - Tags: [http, staging, web]

3. **TCP Test Service** (enabled)

   - Protocol: TCP
   - Endpoint: google.com:80
   - Tags: [tcp, database, production]

4. **gRPC Test Service** (enabled)

   - Protocol: gRPC
   - Endpoint: grpc.example.com:443
   - Tags: [grpc, api, microservice]

5. **Disabled Service** (disabled)
   - Protocol: HTTP
   - Endpoint: https://httpbin.org/status/500
   - Tags: [disabled, test]

## API endpoint coverage

### Dashboard

- `GET /` - main page
- `GET /api/v1/dashboard/stats` - dashboard statistics

### Services

- `GET /api/v1/services` - service list with filters
- `POST /api/v1/services` - service creation
- `GET /api/v1/services/{id}` - service details
- `PUT /api/v1/services/{id}` - service update
- `DELETE /api/v1/services/{id}` - service deletion
- `POST /api/v1/services/{id}/check` - manual check
- `POST /api/v1/services/{id}/resolve` - incident resolution
- `GET /api/v1/services/{id}/stats` - service statistics

### Incidents

- `GET /api/v1/incidents` - all incidents list
- `GET /api/v1/services/{id}/incidents` - service incidents
- `DELETE /api/v1/services/{id}/incidents/{incidentId}` - incident deletion

### Tags

- `GET /api/v1/tags` - tags list
- `GET /api/v1/tags/count` - tags with usage count

## Testing features

### Test isolation

- Each test run uses a temporary SQLite database
- Test server runs on port 8899
- All resources are cleaned up after test completion

### External dependencies

- Tests use httpbin.org for HTTP checks
- TCP tests use google.com:80
- gRPC tests use mock endpoints

### Concurrency

- Tests run sequentially for predictability
- Each test can create and delete its own resources

## Test configuration

Tests use the following configuration:

- **Database**: temporary SQLite
- **Server**: localhost:8899
- **Monitoring interval**: 30 seconds (default)
- **Timeout**: 5 seconds (default)
- **Retries**: 3 (default)
- **Timezone**: UTC
- **Notifications**: disabled

## Debugging

For debugging, you can add additional logging to test code or use Go debugger. Tests output detailed error messages indicating the specific problem location.

## Extending tests

To add new tests:

1. Create a new function in the appropriate file
2. Add it to the test array in the main() function
3. Ensure the test returns an error on failure
4. Add resource cleanup if necessary
