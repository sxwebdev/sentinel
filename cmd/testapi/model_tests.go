package main

import (
	"fmt"
	"reflect"

	"github.com/sxwebdev/sentinel/internal/monitors"
	"github.com/sxwebdev/sentinel/internal/storage"
	"github.com/sxwebdev/sentinel/internal/web"
	"github.com/sxwebdev/sentinel/pkg/dbutils"
)

// Model validation tests

func testModelsValidation(s *TestSuite) error {
	// Test HTTP config validation
	httpConfig := monitors.Config{
		HTTP: &monitors.HTTPConfig{
			Timeout: 5000,
			Endpoints: []monitors.EndpointConfig{
				{
					Name:           "Valid Endpoint",
					URL:            "https://example.com/api",
					Method:         "GET",
					ExpectedStatus: 200,
					Headers: map[string]string{
						"Authorization": "Bearer token",
						"Content-Type":  "application/json",
					},
					Body:     `{"test": true}`,
					JSONPath: "$.status",
					Username: "user",
					Password: "pass",
				},
			},
			Condition: "all",
		},
	}

	if err := httpConfig.Validate(storage.ServiceProtocolTypeHTTP); err != nil {
		return fmt.Errorf("valid HTTP config should not fail validation: %w", err)
	}

	// Test invalid HTTP config - missing endpoints
	invalidHTTPConfig := monitors.Config{
		HTTP: &monitors.HTTPConfig{
			Timeout:   5000,
			Endpoints: []monitors.EndpointConfig{}, // Empty endpoints
		},
	}

	if err := invalidHTTPConfig.Validate(storage.ServiceProtocolTypeHTTP); err == nil {
		return fmt.Errorf("HTTP config with empty endpoints should fail validation")
	}

	// Test invalid HTTP config - invalid endpoint
	invalidEndpointConfig := monitors.Config{
		HTTP: &monitors.HTTPConfig{
			Timeout: 5000,
			Endpoints: []monitors.EndpointConfig{
				{
					Name:           "", // Empty name
					URL:            "invalid-url",
					Method:         "INVALID",
					ExpectedStatus: 999, // Invalid status
				},
			},
		},
	}

	if err := invalidEndpointConfig.Validate(storage.ServiceProtocolTypeHTTP); err == nil {
		return fmt.Errorf("HTTP config with invalid endpoint should fail validation")
	}

	// Test TCP config validation
	tcpConfig := monitors.Config{
		TCP: &monitors.TCPConfig{
			Endpoint:   "example.com:80",
			SendData:   "test data",
			ExpectData: "expected response",
		},
	}

	if err := tcpConfig.Validate(storage.ServiceProtocolTypeTCP); err != nil {
		return fmt.Errorf("valid TCP config should not fail validation: %w", err)
	}

	// Test invalid TCP config - missing endpoint
	invalidTCPConfig := monitors.Config{
		TCP: &monitors.TCPConfig{
			Endpoint: "", // Empty endpoint
		},
	}

	if err := invalidTCPConfig.Validate(storage.ServiceProtocolTypeTCP); err == nil {
		return fmt.Errorf("TCP config with empty endpoint should fail validation")
	}

	// Test gRPC config validation
	grpcConfig := monitors.Config{
		GRPC: &monitors.GRPCConfig{
			Endpoint:    "grpc.example.com:443",
			CheckType:   "health",
			ServiceName: "ExampleService",
			TLS:         true,
			InsecureTLS: false,
		},
	}

	if err := grpcConfig.Validate(storage.ServiceProtocolTypeGRPC); err != nil {
		return fmt.Errorf("valid gRPC config should not fail validation: %w", err)
	}

	// Test invalid gRPC config - missing endpoint
	invalidGRPCConfig := monitors.Config{
		GRPC: &monitors.GRPCConfig{
			Endpoint:  "", // Empty endpoint
			CheckType: "invalid-type",
		},
	}

	if err := invalidGRPCConfig.Validate(storage.ServiceProtocolTypeGRPC); err == nil {
		return fmt.Errorf("gRPC config with empty endpoint should fail validation")
	}

	// Test config with wrong protocol
	httpConfigForTCP := monitors.Config{
		HTTP: &monitors.HTTPConfig{
			Timeout: 5000,
			Endpoints: []monitors.EndpointConfig{
				{
					Name:           "Test",
					URL:            "https://example.com",
					Method:         "GET",
					ExpectedStatus: 200,
				},
			},
		},
	}

	if err := httpConfigForTCP.Validate(storage.ServiceProtocolTypeTCP); err == nil {
		return fmt.Errorf("HTTP config should fail validation for TCP protocol")
	}

	return nil
}

func testServiceDTOFields(s *TestSuite) error {
	// Create a test service to validate DTO conversion
	testService := web.CreateUpdateServiceRequest{
		Name:     "DTO Test Service",
		Protocol: storage.ServiceProtocolTypeHTTP,
		Interval: 30000,
		Timeout:  5000,
		Retries:  3,
		Tags:     []string{"dto", "test", "validation"},
		Config: monitors.Config{
			HTTP: &monitors.HTTPConfig{
				Timeout: 5000,
				Endpoints: []monitors.EndpointConfig{
					{
						Name:           "Test Endpoint",
						URL:            "https://httpbin.org/status/200",
						Method:         "GET",
						ExpectedStatus: 200,
					},
				},
			},
		},
		IsEnabled: true,
	}

	// Create service
	resp, err := s.makeRequest("POST", "/api/v1/services", testService)
	if err != nil {
		return fmt.Errorf("create DTO test service: %w", err)
	}

	var service web.ServiceDTO
	if err := s.decodeResponse(resp, &service); err != nil {
		return fmt.Errorf("decode DTO test service: %w", err)
	}

	// Validate all DTO fields
	if service.ID == "" {
		return fmt.Errorf("service ID should not be empty")
	}
	if service.Name != testService.Name {
		return fmt.Errorf("service name mismatch")
	}
	if service.Protocol != testService.Protocol {
		return fmt.Errorf("service protocol mismatch")
	}
	if service.Interval != testService.Interval {
		return fmt.Errorf("service interval mismatch")
	}
	if service.Timeout != testService.Timeout {
		return fmt.Errorf("service timeout mismatch")
	}
	if service.Retries != testService.Retries {
		return fmt.Errorf("service retries mismatch")
	}
	if !reflect.DeepEqual(service.Tags, testService.Tags) {
		return fmt.Errorf("service tags mismatch")
	}
	if service.IsEnabled != testService.IsEnabled {
		return fmt.Errorf("service enabled status mismatch")
	}

	// Check config conversion
	if service.Config.HTTP == nil {
		return fmt.Errorf("HTTP config should not be nil")
	}
	if service.Config.HTTP.Timeout != testService.Config.HTTP.Timeout {
		return fmt.Errorf("HTTP config timeout mismatch")
	}
	if len(service.Config.HTTP.Endpoints) != len(testService.Config.HTTP.Endpoints) {
		return fmt.Errorf("HTTP endpoints count mismatch")
	}

	// Check endpoint details
	endpoint := service.Config.HTTP.Endpoints[0]
	originalEndpoint := testService.Config.HTTP.Endpoints[0]
	if endpoint.Name != originalEndpoint.Name {
		return fmt.Errorf("endpoint name mismatch")
	}
	if endpoint.URL != originalEndpoint.URL {
		return fmt.Errorf("endpoint URL mismatch")
	}
	if endpoint.Method != originalEndpoint.Method {
		return fmt.Errorf("endpoint method mismatch")
	}
	if endpoint.ExpectedStatus != originalEndpoint.ExpectedStatus {
		return fmt.Errorf("endpoint expected status mismatch")
	}

	// Check state fields are present (even if default values)
	if service.ActiveIncidents < 0 {
		return fmt.Errorf("active incidents should be >= 0")
	}
	if service.TotalIncidents < 0 {
		return fmt.Errorf("total incidents should be >= 0")
	}
	if service.ConsecutiveFails < 0 {
		return fmt.Errorf("consecutive fails should be >= 0")
	}
	if service.ConsecutiveSuccess < 0 {
		return fmt.Errorf("consecutive success should be >= 0")
	}
	if service.TotalChecks < 0 {
		return fmt.Errorf("total checks should be >= 0")
	}

	// Status should be one of the valid values
	validStatuses := []storage.ServiceStatus{
		storage.StatusUnknown,
		storage.StatusUp,
		storage.StatusDown,
		storage.StatusMaintenance,
	}
	isValidStatus := false
	for _, validStatus := range validStatuses {
		if service.Status == validStatus {
			isValidStatus = true
			break
		}
	}
	if !isValidStatus {
		return fmt.Errorf("invalid service status: %s", service.Status)
	}

	// Clean up
	if _, err := s.makeRequest("DELETE", "/api/v1/services/"+service.ID, nil); err != nil {
		return fmt.Errorf("cleanup DTO test service: %w", err)
	}

	return nil
}

func testIncidentFields(s *TestSuite) error {
	// Get incidents to validate structure
	resp, err := s.makeRequest("GET", "/api/v1/incidents?page_size=1", nil)
	if err != nil {
		return fmt.Errorf("get incidents for validation: %w", err)
	}

	var incidents dbutils.FindResponseWithCount[web.Incident]
	if err := s.decodeResponse(resp, &incidents); err != nil {
		return fmt.Errorf("decode incidents for validation: %w", err)
	}

	if len(incidents.Items) > 0 {
		incident := incidents.Items[0]

		// Validate incident fields
		if incident.ID == "" {
			return fmt.Errorf("incident ID should not be empty")
		}
		if incident.ServiceID == "" {
			return fmt.Errorf("incident service ID should not be empty")
		}
		if incident.ServiceName == "" {
			return fmt.Errorf("incident service name should not be empty")
		}
		if incident.Status == "" {
			return fmt.Errorf("incident status should not be empty")
		}
		if incident.Message == "" {
			return fmt.Errorf("incident message should not be empty")
		}

		// StartedAt should be a valid time
		if incident.StartedAt.IsZero() {
			return fmt.Errorf("incident started at should not be zero")
		}

		// If resolved, ResolvedAt should not be nil
		if incident.Resolved && incident.ResolvedAt == nil {
			return fmt.Errorf("resolved incident should have resolved at time")
		}

		// If not resolved, ResolvedAt should be nil
		if !incident.Resolved && incident.ResolvedAt != nil {
			return fmt.Errorf("unresolved incident should not have resolved at time")
		}

		// Duration should be a valid string
		if incident.Duration == "" {
			return fmt.Errorf("incident duration should not be empty")
		}
	}

	return nil
}

func testResponseModels(s *TestSuite) error {
	// Test ErrorResponse
	resp, err := s.makeRequest("GET", "/api/v1/services/non-existent", nil)
	if err != nil {
		return fmt.Errorf("test error response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errorResp web.ErrorResponse
		if err := s.decodeResponse(resp, &errorResp); err == nil {
			if errorResp.Error == "" {
				return fmt.Errorf("error response should have non-empty error message")
			}
		}
	}

	// Test SuccessResponse by triggering a check
	// Create a test service first to ensure we have a valid service ID
	createReq := web.CreateUpdateServiceRequest{
		Name:     "Test Service for Response Models",
		Protocol: storage.ServiceProtocolTypeHTTP,
		Interval: 60000,
		Timeout:  10000,
		Retries:  3,
		Tags:     []string{"test", "response-models"},
		Config: monitors.Config{
			HTTP: &monitors.HTTPConfig{
				Timeout: 30000,
				Endpoints: []monitors.EndpointConfig{
					{
						Name:           "test",
						URL:            "https://httpbin.org/status/200",
						Method:         "GET",
						ExpectedStatus: 200,
					},
				},
			},
		},
		IsEnabled: true,
	}

	resp, err = s.makeRequest("POST", "/api/v1/services", createReq)
	if err != nil {
		return fmt.Errorf("create test service: %w", err)
	}

	var createResponse web.ServiceDTO
	if err := s.decodeResponse(resp, &createResponse); err != nil {
		return fmt.Errorf("decode create service response: %w", err)
	}

	serviceID := createResponse.ID

	// Now test the service check endpoint
	resp, err = s.makeRequest("POST", "/api/v1/services/"+serviceID+"/check", nil)
	if err != nil {
		return fmt.Errorf("test success response: %w", err)
	}

	var successResp web.SuccessResponse
	if err := s.decodeResponse(resp, &successResp); err != nil {
		return fmt.Errorf("decode success response: %w", err)
	}

	if successResp.Message == "" {
		return fmt.Errorf("success response should have non-empty message")
	}

	// Clean up: delete the test service
	_, err = s.makeRequest("DELETE", "/api/v1/services/"+serviceID, nil)
	if err != nil {
		return fmt.Errorf("cleanup test service: %w", err)
	}

	// Test DashboardStats
	resp, err = s.makeRequest("GET", "/api/v1/dashboard/stats", nil)
	if err != nil {
		return fmt.Errorf("test dashboard stats: %w", err)
	}

	var stats web.DashboardStats
	if err := s.decodeResponse(resp, &stats); err != nil {
		return fmt.Errorf("decode dashboard stats: %w", err)
	}

	// Validate stats fields
	if stats.TotalServices < 0 {
		return fmt.Errorf("total services should be >= 0")
	}
	if stats.ServicesUp < 0 {
		return fmt.Errorf("services up should be >= 0")
	}
	if stats.ServicesDown < 0 {
		return fmt.Errorf("services down should be >= 0")
	}
	if stats.ServicesUnknown < 0 {
		return fmt.Errorf("services unknown should be >= 0")
	}
	if stats.ActiveIncidents < 0 {
		return fmt.Errorf("active incidents should be >= 0")
	}
	if stats.UptimePercentage < 0 || stats.UptimePercentage > 100 {
		return fmt.Errorf("uptime percentage should be between 0 and 100")
	}
	if stats.TotalChecks < 0 {
		return fmt.Errorf("total checks should be >= 0")
	}
	if stats.ChecksPerMinute < 0 {
		return fmt.Errorf("checks per minute should be >= 0")
	}

	// Protocols map should contain valid protocols
	for protocol, count := range stats.Protocols {
		validProtocols := []storage.ServiceProtocolType{
			storage.ServiceProtocolTypeHTTP,
			storage.ServiceProtocolTypeTCP,
			storage.ServiceProtocolTypeGRPC,
		}
		isValid := false
		for _, validProtocol := range validProtocols {
			if protocol == validProtocol {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("invalid protocol in stats: %s", protocol)
		}
		if count < 0 {
			return fmt.Errorf("protocol count should be >= 0")
		}
	}

	return nil
}

func testServiceStatsModel(s *TestSuite) error {
	// Create a dedicated service for stats model testing
	testService := web.CreateUpdateServiceRequest{
		Name:     "Stats Model Test Service",
		Protocol: storage.ServiceProtocolTypeHTTP,
		Interval: 30000,
		Timeout:  5000,
		Retries:  3,
		Tags:     []string{"stats-model-test"},
		Config: monitors.Config{
			HTTP: &monitors.HTTPConfig{
				Timeout: 5000,
				Endpoints: []monitors.EndpointConfig{
					{
						Name:           "Test Endpoint",
						URL:            "https://httpbin.org/status/200",
						Method:         "GET",
						ExpectedStatus: 200,
					},
				},
			},
		},
		IsEnabled: true,
	}

	// Create the test service
	resp, err := s.makeRequest("POST", "/api/v1/services", testService)
	if err != nil {
		return fmt.Errorf("create stats model test service: %w", err)
	}

	var createdService web.ServiceDTO
	if err := s.decodeResponse(resp, &createdService); err != nil {
		return fmt.Errorf("decode stats model test service: %w", err)
	}

	serviceID := createdService.ID

	// Ensure cleanup
	defer func() {
		s.makeRequest("DELETE", "/api/v1/services/"+serviceID, nil)
	}()

	resp, err = s.makeRequest("GET", "/api/v1/services/"+serviceID+"/stats", nil)
	if err != nil {
		return fmt.Errorf("get service stats: %w", err)
	}

	var stats web.ServiceStats
	if err := s.decodeResponse(resp, &stats); err != nil {
		return fmt.Errorf("decode service stats: %w", err)
	}

	// Validate stats fields
	if stats.ServiceID != serviceID {
		return fmt.Errorf("service stats service ID mismatch")
	}
	if stats.TotalIncidents < 0 {
		return fmt.Errorf("total incidents should be >= 0")
	}
	if stats.UptimePercentage < 0 || stats.UptimePercentage > 100 {
		return fmt.Errorf("uptime percentage should be between 0 and 100")
	}
	if stats.TotalDowntime < 0 {
		return fmt.Errorf("total downtime should be >= 0")
	}
	if stats.Period <= 0 {
		return fmt.Errorf("period should be > 0")
	}
	if stats.AvgResponseTime < 0 {
		return fmt.Errorf("avg response time should be >= 0")
	}

	return nil
}

func testPaginationResponseModel(s *TestSuite) error {
	// Test services pagination response
	resp, err := s.makeRequest("GET", "/api/v1/services?page=1&page_size=3", nil)
	if err != nil {
		return fmt.Errorf("get paginated services: %w", err)
	}

	var result dbutils.FindResponseWithCount[web.ServiceDTO]
	if err := s.decodeResponse(resp, &result); err != nil {
		return fmt.Errorf("decode paginated services: %w", err)
	}

	// Validate pagination structure
	// Count is uint32, so it's always >= 0
	if len(result.Items) > 3 {
		return fmt.Errorf("items should not exceed page size")
	}

	// Test incidents pagination response
	resp, err = s.makeRequest("GET", "/api/v1/incidents?page=1&page_size=5", nil)
	if err != nil {
		return fmt.Errorf("get paginated incidents: %w", err)
	}

	var incidentResult dbutils.FindResponseWithCount[web.Incident]
	if err := s.decodeResponse(resp, &incidentResult); err != nil {
		return fmt.Errorf("decode paginated incidents: %w", err)
	}

	// Validate pagination structure
	// Count is uint32, so it's always >= 0
	if len(incidentResult.Items) > 5 {
		return fmt.Errorf("incident items should not exceed page size")
	}

	return nil
}

// Helper function to add model validation tests
func getModelValidationTests() []struct {
	name string
	fn   func(*TestSuite) error
} {
	return []struct {
		name string
		fn   func(*TestSuite) error
	}{
		{"TestModelsValidation", testModelsValidation},
		{"TestServiceDTOFields", testServiceDTOFields},
		{"TestIncidentFields", testIncidentFields},
		{"TestResponseModels", testResponseModels},
		{"TestServiceStatsModel", testServiceStatsModel},
		{"TestPaginationResponseModel", testPaginationResponseModel},
	}
}
