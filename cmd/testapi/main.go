package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/sxwebdev/sentinel/internal/config"
	"github.com/sxwebdev/sentinel/internal/monitor"
	"github.com/sxwebdev/sentinel/internal/monitors"
	"github.com/sxwebdev/sentinel/internal/notifier"
	"github.com/sxwebdev/sentinel/internal/receiver"
	"github.com/sxwebdev/sentinel/internal/storage"
	"github.com/sxwebdev/sentinel/internal/web"
	"github.com/sxwebdev/sentinel/pkg/dbutils"
	"github.com/tkcrm/mx/logger"
)

type TestSuite struct {
	server       *web.Server
	baseURL      string
	client       *http.Client
	stor         storage.Storage
	ctx          context.Context
	services     map[string]*web.ServiceDTO
	incidents    map[string]*web.Incident
	testServices []TestService
}

type TestService struct {
	Name     string
	Protocol storage.ServiceProtocolType
	Tags     []string
	Config   monitors.Config
	Enabled  bool
}

var testServices = []TestService{
	{
		Name:     "HTTP Test Service 1",
		Protocol: storage.ServiceProtocolTypeHTTP,
		Tags:     []string{"http", "production", "api"},
		Config: monitors.Config{
			HTTP: &monitors.HTTPConfig{
				Timeout: 5000,
				Endpoints: []monitors.EndpointConfig{
					{
						Name:           "Health Check",
						URL:            "https://httpbin.org/status/200",
						Method:         "GET",
						ExpectedStatus: 200,
					},
				},
			},
		},
		Enabled: true,
	},
	{
		Name:     "HTTP Test Service 2",
		Protocol: storage.ServiceProtocolTypeHTTP,
		Tags:     []string{"http", "staging", "web"},
		Config: monitors.Config{
			HTTP: &monitors.HTTPConfig{
				Timeout: 3000,
				Endpoints: []monitors.EndpointConfig{
					{
						Name:           "Home Page",
						URL:            "https://httpbin.org/status/404",
						Method:         "GET",
						ExpectedStatus: 404,
					},
				},
			},
		},
		Enabled: false,
	},
	{
		Name:     "TCP Test Service",
		Protocol: storage.ServiceProtocolTypeTCP,
		Tags:     []string{"tcp", "database", "production"},
		Config: monitors.Config{
			TCP: &monitors.TCPConfig{
				Endpoint: "google.com:80",
			},
		},
		Enabled: true,
	},
	{
		Name:     "gRPC Test Service",
		Protocol: storage.ServiceProtocolTypeGRPC,
		Tags:     []string{"grpc", "api", "microservice"},
		Config: monitors.Config{
			GRPC: &monitors.GRPCConfig{
				Endpoint:  "grpc.example.com:443",
				CheckType: "connectivity",
				TLS:       true,
			},
		},
		Enabled: true,
	},
	{
		Name:     "Disabled Service",
		Protocol: storage.ServiceProtocolTypeHTTP,
		Tags:     []string{"disabled", "test"},
		Config: monitors.Config{
			HTTP: &monitors.HTTPConfig{
				Timeout: 5000,
				Endpoints: []monitors.EndpointConfig{
					{
						Name:           "Test Endpoint",
						URL:            "https://httpbin.org/status/500",
						Method:         "GET",
						ExpectedStatus: 200,
					},
				},
			},
		},
		Enabled: false,
	},
}

func main() {
	if len(os.Args) < 2 || os.Args[1] != "test" {
		fmt.Println("Usage: go run main.go test")
		os.Exit(1)
	}

	// Run tests
	suite, err := setupTestSuite()
	if err != nil {
		log.Fatalf("Failed to setup test suite: %v", err)
	}
	defer suite.cleanup()

	// Run all tests
	tests := []struct {
		name string
		fn   func(*TestSuite) error
	}{
		{"TestHealthCheck", testHealthCheck},
		{"TestDashboardStats", testDashboardStats},
		{"TestCreateServices", testCreateServices},
		{"TestGetServices", testGetServices},
		{"TestServiceFilters", testServiceFilters},
		{"TestServiceDetail", testServiceDetail},
		{"TestUpdateService", testUpdateService},
		{"TestServiceStats", testServiceStats},
		{"TestServiceCheck", testServiceCheck},
		{"TestIncidents", testIncidents},
		{"TestIncidentFilters", testIncidentFilters},
		{"TestTags", testTags},
		{"TestPagination", testPagination},
		{"TestErrorHandling", testErrorHandling},
		{"TestDeleteService", testDeleteService},
	}

	// Add extended tests
	extendedTests := getExtendedTests()
	tests = append(tests, extendedTests...)

	// Add model validation tests
	modelTests := getModelValidationTests()
	tests = append(tests, modelTests...)

	failed := 0
	for _, test := range tests {
		fmt.Printf("Running %s...\n", test.name)
		if err := test.fn(suite); err != nil {
			fmt.Printf("FAIL: %s - %v\n", test.name, err)
			failed++
		} else {
			fmt.Printf("PASS: %s\n", test.name)
		}
	}

	if failed > 0 {
		fmt.Printf("\n%d test(s) failed\n", failed)
		os.Exit(1)
	} else {
		fmt.Println("\nAll tests passed!")
	}
}

func setupTestSuite() (*TestSuite, error) {
	ctx := context.Background()

	// Create temporary database file
	tmpDir, err := os.MkdirTemp("", "sentinel_test_*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")

	// Load config
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Path: dbPath,
		},
		Monitoring: config.MonitoringConfig{
			Global: config.GlobalConfig{
				DefaultInterval: 30 * time.Second,
				DefaultTimeout:  5 * time.Second,
				DefaultRetries:  3,
			},
		},
		Server: config.ServerConfig{
			Host:     "localhost",
			Port:     8899, // Use different port for testing
			BaseHost: "localhost:8899",
		},
		Timezone: "UTC",
	}

	l := logger.Default()

	// Initialize storage
	stor, err := storage.NewStorage(storage.StorageTypeSQLite, dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Initialize notifier (disabled for tests)
	var notif notifier.Notifier

	// Initialize receiver
	rc := receiver.New()
	if err := rc.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start receiver: %w", err)
	}

	// Create monitor service
	monitorService := monitor.NewMonitorService(stor, cfg, notif, rc)

	// Create web server
	webServer, err := web.NewServer(l, cfg, web.ServerInfo{}, monitorService, stor, rc)
	if err != nil {
		return nil, fmt.Errorf("failed to create web server: %w", err)
	}

	// Start server in background
	go func() {
		if err := webServer.App().Listen(fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Wait a bit for server to start
	time.Sleep(100 * time.Millisecond)

	suite := &TestSuite{
		server:       webServer,
		baseURL:      fmt.Sprintf("http://%s:%d", cfg.Server.Host, cfg.Server.Port),
		client:       &http.Client{Timeout: 10 * time.Second},
		stor:         stor,
		ctx:          ctx,
		services:     make(map[string]*web.ServiceDTO),
		incidents:    make(map[string]*web.Incident),
		testServices: testServices,
	}

	return suite, nil
}

func (s *TestSuite) cleanup() {
	if s.stor != nil {
		s.stor.Stop(context.Background())
	}
}

func (s *TestSuite) makeRequest(method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, s.baseURL+path, reqBody)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return s.client.Do(req)
}

func (s *TestSuite) decodeResponse(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		var errResp web.ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
		}
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, errResp.Error)
	}

	return json.Unmarshal(body, target)
}

// Test cases implementation

func testHealthCheck(s *TestSuite) error {
	resp, err := s.makeRequest("GET", "/", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	return nil
}

func testDashboardStats(s *TestSuite) error {
	resp, err := s.makeRequest("GET", "/api/v1/dashboard/stats", nil)
	if err != nil {
		return err
	}

	var stats web.DashboardStats
	if err := s.decodeResponse(resp, &stats); err != nil {
		return err
	}

	// Basic validation
	if stats.TotalServices < 0 {
		return fmt.Errorf("total services should be >= 0, got %d", stats.TotalServices)
	}

	return nil
}

func testCreateServices(s *TestSuite) error {
	for i, testSvc := range s.testServices {
		createReq := web.CreateUpdateServiceRequest{
			Name:      testSvc.Name,
			Protocol:  testSvc.Protocol,
			Interval:  30000, // 30s
			Timeout:   5000,  // 5s
			Retries:   3,
			Tags:      testSvc.Tags,
			Config:    testSvc.Config,
			IsEnabled: testSvc.Enabled,
		}

		resp, err := s.makeRequest("POST", "/api/v1/services", createReq)
		if err != nil {
			return fmt.Errorf("service %d: %w", i, err)
		}

		var service web.ServiceDTO
		if err := s.decodeResponse(resp, &service); err != nil {
			return fmt.Errorf("service %d: %w", i, err)
		}

		// Validate response
		if service.ID == "" {
			return fmt.Errorf("service %d: missing ID", i)
		}
		if service.Name != testSvc.Name {
			return fmt.Errorf("service %d: name mismatch", i)
		}
		if service.Protocol != testSvc.Protocol {
			return fmt.Errorf("service %d: protocol mismatch", i)
		}
		if !reflect.DeepEqual(service.Tags, testSvc.Tags) {
			return fmt.Errorf("service %d: tags mismatch", i)
		}
		if service.IsEnabled != testSvc.Enabled {
			return fmt.Errorf("service %d: enabled status mismatch", i)
		}

		s.services[service.Name] = &service
	}

	return nil
}

func testGetServices(s *TestSuite) error {
	resp, err := s.makeRequest("GET", "/api/v1/services", nil)
	if err != nil {
		return err
	}

	var result dbutils.FindResponseWithCount[web.ServiceDTO]
	if err := s.decodeResponse(resp, &result); err != nil {
		return err
	}

	if int(result.Count) != len(s.testServices) {
		return fmt.Errorf("expected %d services, got %d", len(s.testServices), result.Count)
	}

	if len(result.Items) != len(s.testServices) {
		return fmt.Errorf("expected %d items, got %d", len(s.testServices), len(result.Items))
	}

	return nil
}

func testServiceFilters(s *TestSuite) error {
	// Test filter by name
	resp, err := s.makeRequest("GET", "/api/v1/services?name=HTTP Test Service 1", nil)
	if err != nil {
		return err
	}

	var result dbutils.FindResponseWithCount[web.ServiceDTO]
	if err := s.decodeResponse(resp, &result); err != nil {
		return err
	}

	if result.Count != 1 {
		return fmt.Errorf("name filter: expected 1 service, got %d", result.Count)
	}
	if result.Items[0].Name != "HTTP Test Service 1" {
		return fmt.Errorf("name filter: wrong service returned")
	}

	// Test filter by protocol
	resp, err = s.makeRequest("GET", "/api/v1/services?protocol=http", nil)
	if err != nil {
		return err
	}

	if err := s.decodeResponse(resp, &result); err != nil {
		return err
	}

	expectedHTTPServices := 0
	for _, svc := range s.testServices {
		if svc.Protocol == storage.ServiceProtocolTypeHTTP {
			expectedHTTPServices++
		}
	}

	if int(result.Count) != expectedHTTPServices {
		return fmt.Errorf("protocol filter: expected %d HTTP services, got %d", expectedHTTPServices, result.Count)
	}

	// Test filter by tags
	resp, err = s.makeRequest("GET", "/api/v1/services?tags=production", nil)
	if err != nil {
		return err
	}

	if err := s.decodeResponse(resp, &result); err != nil {
		return err
	}

	expectedProdServices := 0
	for _, svc := range s.testServices {
		for _, tag := range svc.Tags {
			if tag == "production" {
				expectedProdServices++
				break
			}
		}
	}

	if int(result.Count) != expectedProdServices {
		return fmt.Errorf("tags filter: expected %d production services, got %d", expectedProdServices, result.Count)
	}

	// Test filter by enabled status
	resp, err = s.makeRequest("GET", "/api/v1/services?is_enabled=true", nil)
	if err != nil {
		return err
	}

	if err := s.decodeResponse(resp, &result); err != nil {
		return err
	}

	expectedEnabledServices := 0
	for _, svc := range s.testServices {
		if svc.Enabled {
			expectedEnabledServices++
		}
	}

	if int(result.Count) != expectedEnabledServices {
		return fmt.Errorf("enabled filter: expected %d enabled services, got %d", expectedEnabledServices, result.Count)
	}

	// Test ordering
	resp, err = s.makeRequest("GET", "/api/v1/services?order_by=name", nil)
	if err != nil {
		return err
	}

	if err := s.decodeResponse(resp, &result); err != nil {
		return err
	}

	// Check if results are ordered by name
	for i := 1; i < len(result.Items); i++ {
		if result.Items[i-1].Name > result.Items[i].Name {
			return fmt.Errorf("services are not ordered by name")
		}
	}

	// Test multiple filters
	resp, err = s.makeRequest("GET", "/api/v1/services?protocol=http&tags=production&is_enabled=true", nil)
	if err != nil {
		return err
	}

	if err := s.decodeResponse(resp, &result); err != nil {
		return err
	}

	// Validate each service matches all filters
	for _, item := range result.Items {
		if item.Protocol != storage.ServiceProtocolTypeHTTP {
			return fmt.Errorf("multiple filters: service %s doesn't match protocol filter", item.Name)
		}
		if !item.IsEnabled {
			return fmt.Errorf("multiple filters: service %s doesn't match enabled filter", item.Name)
		}
		hasProdTag := false
		for _, tag := range item.Tags {
			if tag == "production" {
				hasProdTag = true
				break
			}
		}
		if !hasProdTag {
			return fmt.Errorf("multiple filters: service %s doesn't have production tag", item.Name)
		}
	}

	return nil
}

func testServiceDetail(s *TestSuite) error {
	// Get first service
	var serviceID string
	for _, svc := range s.services {
		serviceID = svc.ID
		break
	}

	resp, err := s.makeRequest("GET", "/api/v1/services/"+serviceID, nil)
	if err != nil {
		return err
	}

	var service web.ServiceDTO
	if err := s.decodeResponse(resp, &service); err != nil {
		return err
	}

	if service.ID != serviceID {
		return fmt.Errorf("service detail: ID mismatch")
	}

	// Test non-existent service
	resp, err = s.makeRequest("GET", "/api/v1/services/non-existent", nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusInternalServerError {
		return fmt.Errorf("expected 500 for non-existent service, got %d", resp.StatusCode)
	}

	return nil
}

func testUpdateService(s *TestSuite) error {
	// Get first service
	var service *web.ServiceDTO
	for _, svc := range s.services {
		service = svc
		break
	}

	// Update service
	updateReq := web.CreateUpdateServiceRequest{
		Name:      service.Name + " Updated",
		Protocol:  service.Protocol,
		Interval:  60000, // 60s
		Timeout:   10000, // 10s
		Retries:   5,
		Tags:      append(service.Tags, "updated"),
		Config:    service.Config,
		IsEnabled: !service.IsEnabled,
	}

	resp, err := s.makeRequest("PUT", "/api/v1/services/"+service.ID, updateReq)
	if err != nil {
		return err
	}

	var updatedService web.ServiceDTO
	if err := s.decodeResponse(resp, &updatedService); err != nil {
		return err
	}

	// Validate updates
	if updatedService.Name != updateReq.Name {
		return fmt.Errorf("update: name not updated")
	}
	if updatedService.Interval != updateReq.Interval {
		return fmt.Errorf("update: interval not updated")
	}
	if updatedService.Timeout != updateReq.Timeout {
		return fmt.Errorf("update: timeout not updated")
	}
	if updatedService.Retries != updateReq.Retries {
		return fmt.Errorf("update: retries not updated")
	}
	if updatedService.IsEnabled != updateReq.IsEnabled {
		return fmt.Errorf("update: enabled status not updated")
	}

	return nil
}

func testServiceStats(s *TestSuite) error {
	// Get first service
	var serviceID string
	for _, svc := range s.services {
		serviceID = svc.ID
		break
	}

	// Test service stats
	resp, err := s.makeRequest("GET", "/api/v1/services/"+serviceID+"/stats", nil)
	if err != nil {
		return err
	}

	var stats web.ServiceStats
	if err := s.decodeResponse(resp, &stats); err != nil {
		return err
	}

	if stats.ServiceID != serviceID {
		return fmt.Errorf("stats: service ID mismatch")
	}

	// Test with custom days parameter
	resp, err = s.makeRequest("GET", "/api/v1/services/"+serviceID+"/stats?days=7", nil)
	if err != nil {
		return err
	}

	if err := s.decodeResponse(resp, &stats); err != nil {
		return err
	}

	return nil
}

func testServiceCheck(s *TestSuite) error {
	// Get first service
	var serviceID string
	for _, svc := range s.services {
		serviceID = svc.ID
		break
	}

	// Trigger check
	resp, err := s.makeRequest("POST", "/api/v1/services/"+serviceID+"/check", nil)
	if err != nil {
		return err
	}

	var result web.SuccessResponse
	if err := s.decodeResponse(resp, &result); err != nil {
		return err
	}

	if result.Message == "" {
		return fmt.Errorf("check: empty success message")
	}

	return nil
}

func testIncidents(s *TestSuite) error {
	// Test get all incidents
	resp, err := s.makeRequest("GET", "/api/v1/incidents", nil)
	if err != nil {
		return err
	}

	var incidents dbutils.FindResponseWithCount[web.Incident]
	if err := s.decodeResponse(resp, &incidents); err != nil {
		return err
	}

	// Get first service for service-specific incidents
	var serviceID string
	for _, svc := range s.services {
		serviceID = svc.ID
		break
	}

	// Test service incidents
	resp, err = s.makeRequest("GET", "/api/v1/services/"+serviceID+"/incidents", nil)
	if err != nil {
		return err
	}

	var serviceIncidents dbutils.FindResponseWithCount[web.Incident]
	if err := s.decodeResponse(resp, &serviceIncidents); err != nil {
		return err
	}

	// All incidents should belong to the service
	for _, incident := range serviceIncidents.Items {
		if incident.ServiceID != serviceID {
			return fmt.Errorf("incident %s doesn't belong to service %s", incident.ID, serviceID)
		}
	}

	return nil
}

func testIncidentFilters(s *TestSuite) error {
	// Get first service
	var serviceID string
	for _, svc := range s.services {
		serviceID = svc.ID
		break
	}

	// Test filter by resolved status
	resp, err := s.makeRequest("GET", "/api/v1/services/"+serviceID+"/incidents?resolved=false", nil)
	if err != nil {
		return err
	}

	var incidents dbutils.FindResponseWithCount[web.Incident]
	if err := s.decodeResponse(resp, &incidents); err != nil {
		return err
	}

	// All incidents should be unresolved
	for _, incident := range incidents.Items {
		if incident.Resolved {
			return fmt.Errorf("incident filter: found resolved incident when filtering for unresolved")
		}
	}

	// Test time-based filtering
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	resp, err = s.makeRequest("GET", "/api/v1/incidents?start_time="+url.QueryEscape(yesterday.Format(time.RFC3339)), nil)
	if err != nil {
		return err
	}

	if err := s.decodeResponse(resp, &incidents); err != nil {
		return err
	}

	return nil
}

func testTags(s *TestSuite) error {
	// Test get all tags
	resp, err := s.makeRequest("GET", "/api/v1/tags", nil)
	if err != nil {
		return err
	}

	var tags []string
	if err := s.decodeResponse(resp, &tags); err != nil {
		return err
	}

	// Collect expected tags
	expectedTags := make(map[string]bool)
	for _, svc := range s.testServices {
		for _, tag := range svc.Tags {
			expectedTags[tag] = true
		}
	}

	// Check if all expected tags are present
	tagSet := make(map[string]bool)
	for _, tag := range tags {
		tagSet[tag] = true
	}

	for expectedTag := range expectedTags {
		if !tagSet[expectedTag] {
			return fmt.Errorf("tags: missing expected tag %s", expectedTag)
		}
	}

	// Test get tags with count
	resp, err = s.makeRequest("GET", "/api/v1/tags/count", nil)
	if err != nil {
		return err
	}

	var tagsWithCount map[string]int
	if err := s.decodeResponse(resp, &tagsWithCount); err != nil {
		return err
	}

	// Validate counts
	for tag, count := range tagsWithCount {
		if count <= 0 {
			return fmt.Errorf("tags count: tag %s has invalid count %d", tag, count)
		}
	}

	return nil
}

func testPagination(s *TestSuite) error {
	// Test services pagination
	resp, err := s.makeRequest("GET", "/api/v1/services?page=1&page_size=2", nil)
	if err != nil {
		return err
	}

	var result dbutils.FindResponseWithCount[web.ServiceDTO]
	if err := s.decodeResponse(resp, &result); err != nil {
		return err
	}

	if len(result.Items) > 2 {
		return fmt.Errorf("pagination: page size not respected, got %d items", len(result.Items))
	}

	// Test second page
	resp, err = s.makeRequest("GET", "/api/v1/services?page=2&page_size=2", nil)
	if err != nil {
		return err
	}

	var result2 dbutils.FindResponseWithCount[web.ServiceDTO]
	if err := s.decodeResponse(resp, &result2); err != nil {
		return err
	}

	// Make sure pages don't overlap
	if len(result.Items) > 0 && len(result2.Items) > 0 {
		for _, item1 := range result.Items {
			for _, item2 := range result2.Items {
				if item1.ID == item2.ID {
					return fmt.Errorf("pagination: services overlap between pages")
				}
			}
		}
	}

	return nil
}

func testErrorHandling(s *TestSuite) error {
	// Test invalid service creation
	invalidService := web.CreateUpdateServiceRequest{
		Name:     "", // Empty name should fail
		Protocol: storage.ServiceProtocolTypeHTTP,
	}

	resp, err := s.makeRequest("POST", "/api/v1/services", invalidService)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusBadRequest {
		return fmt.Errorf("error handling: expected 400 for invalid service, got %d", resp.StatusCode)
	}

	// Test invalid query parameters
	resp, err = s.makeRequest("GET", "/api/v1/services?status=invalid", nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusBadRequest {
		return fmt.Errorf("error handling: expected 400 for invalid status filter, got %d", resp.StatusCode)
	}

	// Test invalid pagination
	resp, err = s.makeRequest("GET", "/api/v1/services?page=0", nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusBadRequest {
		return fmt.Errorf("error handling: expected 400 for invalid page, got %d", resp.StatusCode)
	}

	return nil
}

func testDeleteService(s *TestSuite) error {
	// Get a service to delete
	var serviceID string
	for _, svc := range s.services {
		serviceID = svc.ID
		break
	}

	// Delete service
	resp, err := s.makeRequest("DELETE", "/api/v1/services/"+serviceID, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("delete: expected 204, got %d", resp.StatusCode)
	}

	// Verify service is deleted
	resp, err = s.makeRequest("GET", "/api/v1/services/"+serviceID, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusInternalServerError {
		return fmt.Errorf("delete: service still exists after deletion")
	}

	return nil
}
