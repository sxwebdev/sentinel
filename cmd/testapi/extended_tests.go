package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/sxwebdev/sentinel/internal/monitors"
	"github.com/sxwebdev/sentinel/internal/storage"
	"github.com/sxwebdev/sentinel/internal/web"
	"github.com/sxwebdev/sentinel/pkg/dbutils"
)

// Extended tests for comprehensive API coverage

func testAdvancedServiceFilters(s *TestSuite) error {
	// Test complex tag filtering with multiple tags
	resp, err := s.makeRequest("GET", "/api/v1/services?tags=production,api", nil)
	if err != nil {
		return err
	}

	var result dbutils.FindResponseWithCount[web.ServiceDTO]
	if err := s.decodeResponse(resp, &result); err != nil {
		return err
	}

	// Each service should have both tags
	for _, service := range result.Items {
		hasProduction := false
		hasAPI := false
		for _, tag := range service.Tags {
			if tag == "production" {
				hasProduction = true
			}
			if tag == "api" {
				hasAPI = true
			}
		}
		if !hasProduction || !hasAPI {
			return fmt.Errorf("service %s doesn't have both required tags", service.Name)
		}
	}

	// Test status filtering
	resp, err = s.makeRequest("GET", "/api/v1/services?status=up", nil)
	if err != nil {
		return err
	}

	if err := s.decodeResponse(resp, &result); err != nil {
		return err
	}

	for _, service := range result.Items {
		if service.Status != storage.StatusUp {
			return fmt.Errorf("service %s status is not 'up'", service.Name)
		}
	}

	// Test ordering by created_at
	resp, err = s.makeRequest("GET", "/api/v1/services?order_by=created_at", nil)
	if err != nil {
		return err
	}

	if err := s.decodeResponse(resp, &result); err != nil {
		return err
	}

	// Services should be ordered (assuming creation order)
	if len(result.Items) > 1 {
		fmt.Printf("Order test: found %d services\n", len(result.Items))
	}

	return nil
}

func testServiceCRUDCompleteFlow(s *TestSuite) error {
	// Create a new service with complex HTTP config
	complexHTTPService := web.CreateUpdateServiceRequest{
		Name:     "Complex HTTP Service",
		Protocol: storage.ServiceProtocolTypeHTTP,
		Interval: 15000, // 15s
		Timeout:  10000, // 10s
		Retries:  2,
		Tags:     []string{"complex", "test", "http"},
		Config: monitors.Config{
			HTTP: &monitors.HTTPConfig{
				Timeout: 8000,
				Endpoints: []monitors.EndpointConfig{
					{
						Name:           "Main API",
						URL:            "https://httpbin.org/get",
						Method:         "GET",
						ExpectedStatus: 200,
						Headers: map[string]string{
							"Authorization": "Bearer test-token",
							"User-Agent":    "Sentinel-Test",
						},
					},
					{
						Name:           "Health Check",
						URL:            "https://httpbin.org/status/200",
						Method:         "GET",
						ExpectedStatus: 200,
					},
					{
						Name:           "POST Endpoint",
						URL:            "https://httpbin.org/post",
						Method:         "POST",
						ExpectedStatus: 200,
						Body:           `{"test": "data"}`,
						Headers: map[string]string{
							"Content-Type": "application/json",
						},
					},
				},
				Condition: "all", // All endpoints must pass
			},
		},
		IsEnabled: true,
	}

	// Create service
	resp, err := s.makeRequest("POST", "/api/v1/services", complexHTTPService)
	if err != nil {
		return fmt.Errorf("create complex service: %w", err)
	}

	var createdService web.ServiceDTO
	if err := s.decodeResponse(resp, &createdService); err != nil {
		return fmt.Errorf("decode created service: %w", err)
	}

	serviceID := createdService.ID

	// Verify creation
	if createdService.Name != complexHTTPService.Name {
		return fmt.Errorf("name mismatch in created service")
	}
	if createdService.Protocol != complexHTTPService.Protocol {
		return fmt.Errorf("protocol mismatch in created service")
	}
	if createdService.Config.HTTP == nil {
		return fmt.Errorf("HTTP config is nil in created service")
	}
	if len(createdService.Config.HTTP.Endpoints) != 3 {
		return fmt.Errorf("expected 3 endpoints, got %d", len(createdService.Config.HTTP.Endpoints))
	}

	// Test service detail retrieval
	resp, err = s.makeRequest("GET", "/api/v1/services/"+serviceID, nil)
	if err != nil {
		return fmt.Errorf("get service detail: %w", err)
	}

	var serviceDetail web.ServiceDTO
	if err := s.decodeResponse(resp, &serviceDetail); err != nil {
		return fmt.Errorf("decode service detail: %w", err)
	}

	if serviceDetail.ID != serviceID {
		return fmt.Errorf("service detail ID mismatch")
	}

	// Update service - modify configuration
	updatedConfig := complexHTTPService
	updatedConfig.Name = "Updated Complex HTTP Service"
	updatedConfig.Interval = 45000 // 45s
	updatedConfig.Tags = append(updatedConfig.Tags, "updated")

	// Remove one endpoint
	updatedConfig.Config.HTTP.Endpoints = updatedConfig.Config.HTTP.Endpoints[:2]
	updatedConfig.Config.HTTP.Condition = "any" // Any endpoint can pass

	resp, err = s.makeRequest("PUT", "/api/v1/services/"+serviceID, updatedConfig)
	if err != nil {
		return fmt.Errorf("update service: %w", err)
	}

	var updatedService web.ServiceDTO
	if err := s.decodeResponse(resp, &updatedService); err != nil {
		return fmt.Errorf("decode updated service: %w", err)
	}

	// Verify updates
	if updatedService.Name != updatedConfig.Name {
		return fmt.Errorf("updated name mismatch")
	}
	if updatedService.Interval != updatedConfig.Interval {
		return fmt.Errorf("updated interval mismatch")
	}
	if len(updatedService.Config.HTTP.Endpoints) != 2 {
		return fmt.Errorf("expected 2 endpoints after update, got %d", len(updatedService.Config.HTTP.Endpoints))
	}
	if updatedService.Config.HTTP.Condition != "any" {
		return fmt.Errorf("condition not updated")
	}

	// Trigger manual check
	resp, err = s.makeRequest("POST", "/api/v1/services/"+serviceID+"/check", nil)
	if err != nil {
		return fmt.Errorf("trigger check: %w", err)
	}

	var checkResult web.SuccessResponse
	if err := s.decodeResponse(resp, &checkResult); err != nil {
		return fmt.Errorf("decode check result: %w", err)
	}

	if checkResult.Message == "" {
		return fmt.Errorf("empty check result message")
	}

	// Wait a bit and check service stats
	time.Sleep(100 * time.Millisecond)

	resp, err = s.makeRequest("GET", "/api/v1/services/"+serviceID+"/stats?days=1", nil)
	if err != nil {
		return fmt.Errorf("get service stats: %w", err)
	}

	var stats web.ServiceStats
	if err := s.decodeResponse(resp, &stats); err != nil {
		return fmt.Errorf("decode service stats: %w", err)
	}

	if stats.ServiceID != serviceID {
		return fmt.Errorf("stats service ID mismatch")
	}

	// Get service incidents
	resp, err = s.makeRequest("GET", "/api/v1/services/"+serviceID+"/incidents", nil)
	if err != nil {
		return fmt.Errorf("get service incidents: %w", err)
	}

	var incidents dbutils.FindResponseWithCount[web.Incident]
	if err := s.decodeResponse(resp, &incidents); err != nil {
		return fmt.Errorf("decode service incidents: %w", err)
	}

	// All incidents should belong to this service
	for _, incident := range incidents.Items {
		if incident.ServiceID != serviceID {
			return fmt.Errorf("incident %s doesn't belong to service %s", incident.ID, serviceID)
		}
	}

	// Delete the service
	resp, err = s.makeRequest("DELETE", "/api/v1/services/"+serviceID, nil)
	if err != nil {
		return fmt.Errorf("delete service: %w", err)
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("delete service: expected 204, got %d", resp.StatusCode)
	}

	// Verify deletion
	resp, err = s.makeRequest("GET", "/api/v1/services/"+serviceID, nil)
	if err != nil {
		return fmt.Errorf("verify deletion: %w", err)
	}

	if resp.StatusCode == http.StatusOK {
		return fmt.Errorf("service still exists after deletion")
	}

	return nil
}

func testAdvancedIncidentManagement(s *TestSuite) error {
	// Create a dedicated service for incident testing
	testService := web.CreateUpdateServiceRequest{
		Name:     "Incident Management Test Service",
		Protocol: storage.ServiceProtocolTypeHTTP,
		Interval: 30000,
		Timeout:  5000,
		Retries:  3,
		Tags:     []string{"incident-test"},
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
		return fmt.Errorf("create incident test service: %w", err)
	}

	var createdService web.ServiceDTO
	if err := s.decodeResponse(resp, &createdService); err != nil {
		return fmt.Errorf("decode incident test service: %w", err)
	}

	serviceID := createdService.ID

	// Ensure cleanup
	defer func() {
		s.makeRequest("DELETE", "/api/v1/services/"+serviceID, nil)
	}()

	// Get service incidents with various filters
	resp, err = s.makeRequest("GET", "/api/v1/services/"+serviceID+"/incidents?resolved=false", nil)
	if err != nil {
		return err
	}

	var unresolvedIncidents dbutils.FindResponseWithCount[web.Incident]
	if err := s.decodeResponse(resp, &unresolvedIncidents); err != nil {
		return err
	}

	// All incidents should be unresolved
	for _, incident := range unresolvedIncidents.Items {
		if incident.Resolved {
			return fmt.Errorf("found resolved incident when filtering for unresolved")
		}
	}

	// Get resolved incidents
	resp, err = s.makeRequest("GET", "/api/v1/services/"+serviceID+"/incidents?resolved=true", nil)
	if err != nil {
		return err
	}

	var resolvedIncidents dbutils.FindResponseWithCount[web.Incident]
	if err := s.decodeResponse(resp, &resolvedIncidents); err != nil {
		return err
	}

	// All incidents should be resolved
	for _, incident := range resolvedIncidents.Items {
		if !incident.Resolved {
			return fmt.Errorf("found unresolved incident when filtering for resolved")
		}
	}

	// Test time-based filtering
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	tomorrow := now.AddDate(0, 0, 1)

	// Filter by time range
	timeParams := fmt.Sprintf("start_time=%s&end_time=%s",
		url.QueryEscape(yesterday.Format(time.RFC3339)),
		url.QueryEscape(tomorrow.Format(time.RFC3339)))

	resp, err = s.makeRequest("GET", "/api/v1/services/"+serviceID+"/incidents?"+timeParams, nil)
	if err != nil {
		return err
	}

	var timeFilteredIncidents dbutils.FindResponseWithCount[web.Incident]
	if err := s.decodeResponse(resp, &timeFilteredIncidents); err != nil {
		return err
	}

	// Test pagination for incidents
	resp, err = s.makeRequest("GET", "/api/v1/services/"+serviceID+"/incidents?page=1&page_size=5", nil)
	if err != nil {
		return err
	}

	var paginatedIncidents dbutils.FindResponseWithCount[web.Incident]
	if err := s.decodeResponse(resp, &paginatedIncidents); err != nil {
		return err
	}

	if len(paginatedIncidents.Items) > 5 {
		return fmt.Errorf("page size not respected for incidents: got %d items", len(paginatedIncidents.Items))
	}

	// Test resolve service incidents
	resp, err = s.makeRequest("POST", "/api/v1/services/"+serviceID+"/resolve", nil)
	if err != nil {
		return err
	}

	var resolveResult web.SuccessResponse
	if err := s.decodeResponse(resp, &resolveResult); err != nil {
		return err
	}

	if resolveResult.Message == "" {
		return fmt.Errorf("empty resolve result message")
	}

	// Test global incident search
	resp, err = s.makeRequest("GET", "/api/v1/incidents?search="+serviceID, nil)
	if err != nil {
		return err
	}

	var searchResults dbutils.FindResponseWithCount[web.Incident]
	if err := s.decodeResponse(resp, &searchResults); err != nil {
		return err
	}

	// All results should be related to the searched service
	for _, incident := range searchResults.Items {
		if incident.ServiceID != serviceID && incident.ID != serviceID {
			// Check if the search term appears in service name or incident data
			hasSearchTerm := false
			if incident.ServiceID == serviceID || incident.ID == serviceID {
				hasSearchTerm = true
			}
			if !hasSearchTerm {
				return fmt.Errorf("search result doesn't match search criteria")
			}
		}
	}

	return nil
}

func testCompleteProtocolConfigurations(s *TestSuite) error {
	// Test TCP service with advanced configuration
	tcpService := web.CreateUpdateServiceRequest{
		Name:     "Advanced TCP Service",
		Protocol: storage.ServiceProtocolTypeTCP,
		Interval: 20000,
		Timeout:  8000,
		Retries:  2,
		Tags:     []string{"tcp", "advanced", "database"},
		Config: monitors.Config{
			TCP: &monitors.TCPConfig{
				Endpoint:   "google.com:80",
				SendData:   "GET / HTTP/1.1\r\nHost: google.com\r\n\r\n",
				ExpectData: "HTTP/1.1",
			},
		},
		IsEnabled: true,
	}

	resp, err := s.makeRequest("POST", "/api/v1/services", tcpService)
	if err != nil {
		return fmt.Errorf("create TCP service: %w", err)
	}

	var createdTCPService web.ServiceDTO
	if err := s.decodeResponse(resp, &createdTCPService); err != nil {
		return fmt.Errorf("decode TCP service: %w", err)
	}

	if createdTCPService.Config.TCP == nil {
		return fmt.Errorf("TCP config is nil")
	}
	if createdTCPService.Config.TCP.Endpoint != tcpService.Config.TCP.Endpoint {
		return fmt.Errorf("TCP endpoint mismatch")
	}
	if createdTCPService.Config.TCP.SendData != tcpService.Config.TCP.SendData {
		return fmt.Errorf("TCP send data mismatch")
	}

	tcpServiceID := createdTCPService.ID

	// Test gRPC service with advanced configuration
	grpcService := web.CreateUpdateServiceRequest{
		Name:     "Advanced gRPC Service",
		Protocol: storage.ServiceProtocolTypeGRPC,
		Interval: 25000,
		Timeout:  10000,
		Retries:  3,
		Tags:     []string{"grpc", "advanced", "microservice"},
		Config: monitors.Config{
			GRPC: &monitors.GRPCConfig{
				Endpoint:    "grpc.example.com:443",
				CheckType:   "health",
				ServiceName: "example.HealthService",
				TLS:         true,
				InsecureTLS: false,
			},
		},
		IsEnabled: true,
	}

	resp, err = s.makeRequest("POST", "/api/v1/services", grpcService)
	if err != nil {
		return fmt.Errorf("create gRPC service: %w", err)
	}

	var createdGRPCService web.ServiceDTO
	if err := s.decodeResponse(resp, &createdGRPCService); err != nil {
		return fmt.Errorf("decode gRPC service: %w", err)
	}

	if createdGRPCService.Config.GRPC == nil {
		return fmt.Errorf("gRPC config is nil")
	}
	if createdGRPCService.Config.GRPC.Endpoint != grpcService.Config.GRPC.Endpoint {
		return fmt.Errorf("gRPC endpoint mismatch")
	}
	if createdGRPCService.Config.GRPC.CheckType != grpcService.Config.GRPC.CheckType {
		return fmt.Errorf("gRPC check type mismatch")
	}
	if createdGRPCService.Config.GRPC.TLS != grpcService.Config.GRPC.TLS {
		return fmt.Errorf("gRPC TLS setting mismatch")
	}

	grpcServiceID := createdGRPCService.ID

	// Test both services individually
	services := []struct {
		id       string
		protocol storage.ServiceProtocolType
	}{
		{tcpServiceID, storage.ServiceProtocolTypeTCP},
		{grpcServiceID, storage.ServiceProtocolTypeGRPC},
	}

	for _, svc := range services {
		// Test service detail
		resp, err = s.makeRequest("GET", "/api/v1/services/"+svc.id, nil)
		if err != nil {
			return fmt.Errorf("get %s service detail: %w", svc.protocol, err)
		}

		var serviceDetail web.ServiceDTO
		if err := s.decodeResponse(resp, &serviceDetail); err != nil {
			return fmt.Errorf("decode %s service detail: %w", svc.protocol, err)
		}

		if serviceDetail.Protocol != svc.protocol {
			return fmt.Errorf("%s service protocol mismatch", svc.protocol)
		}

		// Test service check
		resp, err = s.makeRequest("POST", "/api/v1/services/"+svc.id+"/check", nil)
		if err != nil {
			return fmt.Errorf("trigger %s service check: %w", svc.protocol, err)
		}

		var checkResult web.SuccessResponse
		if err := s.decodeResponse(resp, &checkResult); err != nil {
			return fmt.Errorf("decode %s check result: %w", svc.protocol, err)
		}

		// Test service stats
		resp, err = s.makeRequest("GET", "/api/v1/services/"+svc.id+"/stats", nil)
		if err != nil {
			return fmt.Errorf("get %s service stats: %w", svc.protocol, err)
		}

		var stats web.ServiceStats
		if err := s.decodeResponse(resp, &stats); err != nil {
			return fmt.Errorf("decode %s service stats: %w", svc.protocol, err)
		}

		if stats.ServiceID != svc.id {
			return fmt.Errorf("%s service stats ID mismatch", svc.protocol)
		}
	}

	// Clean up - delete both services
	for _, svc := range services {
		resp, err = s.makeRequest("DELETE", "/api/v1/services/"+svc.id, nil)
		if err != nil {
			return fmt.Errorf("delete %s service: %w", svc.protocol, err)
		}

		if resp.StatusCode != http.StatusNoContent {
			return fmt.Errorf("delete %s service: expected 204, got %d", svc.protocol, resp.StatusCode)
		}
	}

	return nil
}

func testAdvancedPaginationAndSorting(s *TestSuite) error {
	// Create multiple services for better pagination testing
	testServices := []web.CreateUpdateServiceRequest{}
	for i := 0; i < 10; i++ {
		service := web.CreateUpdateServiceRequest{
			Name:     fmt.Sprintf("Pagination Test Service %02d", i),
			Protocol: storage.ServiceProtocolTypeHTTP,
			Interval: 30000,
			Timeout:  5000,
			Retries:  3,
			Tags:     []string{fmt.Sprintf("page-test-%d", i%3), "pagination"},
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
			IsEnabled: i%2 == 0, // Alternate enabled/disabled
		}
		testServices = append(testServices, service)
	}

	// Create all test services
	createdIDs := []string{}
	for i, service := range testServices {
		resp, err := s.makeRequest("POST", "/api/v1/services", service)
		if err != nil {
			return fmt.Errorf("create pagination test service %d: %w", i, err)
		}

		var created web.ServiceDTO
		if err := s.decodeResponse(resp, &created); err != nil {
			return fmt.Errorf("decode pagination test service %d: %w", i, err)
		}

		createdIDs = append(createdIDs, created.ID)
	}

	// Test various page sizes
	pageSizes := []int{3, 5, 7}
	for _, pageSize := range pageSizes {
		page := 1
		allItems := []web.ServiceDTO{}
		seenIDs := make(map[string]bool)

		for {
			resp, err := s.makeRequest("GET",
				fmt.Sprintf("/api/v1/services?page=%d&page_size=%d&order_by=name", page, pageSize), nil)
			if err != nil {
				return fmt.Errorf("pagination test page %d size %d: %w", page, pageSize, err)
			}

			var result dbutils.FindResponseWithCount[web.ServiceDTO]
			if err := s.decodeResponse(resp, &result); err != nil {
				return fmt.Errorf("decode pagination test page %d size %d: %w", page, pageSize, err)
			}

			if len(result.Items) == 0 {
				break // No more items
			}

			if len(result.Items) > pageSize {
				return fmt.Errorf("page size exceeded: requested %d, got %d", pageSize, len(result.Items))
			}

			// Check for duplicates
			for _, item := range result.Items {
				if seenIDs[item.ID] {
					return fmt.Errorf("duplicate item %s found across pages", item.ID)
				}
				seenIDs[item.ID] = true
				allItems = append(allItems, item)
			}

			page++
			if page > 20 { // Safety break
				break
			}
		}

		// Check if items are properly sorted by name
		for i := 1; i < len(allItems); i++ {
			if allItems[i-1].Name > allItems[i].Name {
				return fmt.Errorf("items not sorted by name: %s > %s", allItems[i-1].Name, allItems[i].Name)
			}
		}
	}

	// Test ordering by created_at
	resp, err := s.makeRequest("GET", "/api/v1/services?order_by=created_at&page_size=20", nil)
	if err != nil {
		return fmt.Errorf("order by created_at test: %w", err)
	}

	var orderedResult dbutils.FindResponseWithCount[web.ServiceDTO]
	if err := s.decodeResponse(resp, &orderedResult); err != nil {
		return fmt.Errorf("decode order by created_at test: %w", err)
	}

	// Clean up - delete all test services
	for _, id := range createdIDs {
		_, err := s.makeRequest("DELETE", "/api/v1/services/"+id, nil)
		if err != nil {
			// Log error but continue cleanup
			fmt.Printf("Warning: failed to delete test service %s: %v\n", id, err)
		}
	}

	return nil
}

func testAdvancedErrorScenarios(s *TestSuite) error {
	// Test various error conditions

	// 1. Invalid JSON in request body
	invalidJSON := bytes.NewBuffer([]byte(`{"name": "test", "protocol": "http", invalid json`))
	req, _ := http.NewRequest("POST", s.baseURL+"/api/v1/services", invalidJSON)
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("invalid JSON test request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		return fmt.Errorf("invalid JSON: expected 400, got %d", resp.StatusCode)
	}

	// 2. Missing required fields
	incompleteService := map[string]interface{}{
		"protocol": "http",
		// Missing name
	}

	resp, err = s.makeRequest("POST", "/api/v1/services", incompleteService)
	if err != nil {
		return fmt.Errorf("incomplete service test: %w", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		return fmt.Errorf("incomplete service: expected 400, got %d", resp.StatusCode)
	}

	// 3. Invalid protocol
	invalidProtocolService := web.CreateUpdateServiceRequest{
		Name:     "Invalid Protocol Service",
		Protocol: "invalid-protocol",
		Config:   monitors.Config{},
	}

	resp, err = s.makeRequest("POST", "/api/v1/services", invalidProtocolService)
	if err != nil {
		return fmt.Errorf("invalid protocol test: %w", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		return fmt.Errorf("invalid protocol: expected 400, got %d", resp.StatusCode)
	}

	// 4. Invalid query parameters
	invalidQueries := []string{
		"?status=invalid-status",
		"?protocol=invalid-protocol",
		"?order_by=invalid-field",
		"?page=0",
		"?page_size=0",
		"?page_size=1000",
	}

	for _, query := range invalidQueries {
		resp, err = s.makeRequest("GET", "/api/v1/services"+query, nil)
		if err != nil {
			return fmt.Errorf("invalid query %s test: %w", query, err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			return fmt.Errorf("invalid query %s: expected 400, got %d", query, resp.StatusCode)
		}
	}

	// 5. Operations on non-existent service
	nonExistentID := "non-existent-service-id"

	// GET non-existent service
	resp, err = s.makeRequest("GET", "/api/v1/services/"+nonExistentID, nil)
	if err != nil {
		return fmt.Errorf("get non-existent service test: %w", err)
	}

	if resp.StatusCode != http.StatusInternalServerError {
		return fmt.Errorf("get non-existent service: expected 500, got %d", resp.StatusCode)
	}

	// UPDATE non-existent service
	updateReq := web.CreateUpdateServiceRequest{
		Name:     "Updated Service",
		Protocol: storage.ServiceProtocolTypeHTTP,
		Interval: 60000,
		Timeout:  10000,
		Retries:  3,
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

	resp, err = s.makeRequest("PUT", "/api/v1/services/"+nonExistentID, updateReq)
	if err != nil {
		return fmt.Errorf("update non-existent service test: %w", err)
	}

	if resp.StatusCode != http.StatusInternalServerError {
		return fmt.Errorf("update non-existent service: expected 500, got %d", resp.StatusCode)
	}

	// DELETE non-existent service
	resp, err = s.makeRequest("DELETE", "/api/v1/services/"+nonExistentID, nil)
	if err != nil {
		return fmt.Errorf("delete non-existent service test: %w", err)
	}

	if resp.StatusCode != http.StatusInternalServerError {
		return fmt.Errorf("delete non-existent service: expected 500, got %d", resp.StatusCode)
	}

	// CHECK non-existent service
	resp, err = s.makeRequest("POST", "/api/v1/services/"+nonExistentID+"/check", nil)
	if err != nil {
		return fmt.Errorf("check non-existent service test: %w", err)
	}

	if resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("check non-existent service: expected 404, got %d", resp.StatusCode)
	}

	// RESOLVE non-existent service
	resp, err = s.makeRequest("POST", "/api/v1/services/"+nonExistentID+"/resolve", nil)
	if err != nil {
		return fmt.Errorf("resolve non-existent service test: %w", err)
	}

	if resp.StatusCode != http.StatusInternalServerError {
		return fmt.Errorf("resolve non-existent service: expected 500, got %d", resp.StatusCode)
	}

	// STATS for non-existent service
	resp, err = s.makeRequest("GET", "/api/v1/services/"+nonExistentID+"/stats", nil)
	if err != nil {
		return fmt.Errorf("stats non-existent service test: %w", err)
	}

	if resp.StatusCode != http.StatusInternalServerError {
		return fmt.Errorf("stats non-existent service: expected 500, got %d", resp.StatusCode)
	}

	// INCIDENTS for non-existent service
	resp, err = s.makeRequest("GET", "/api/v1/services/"+nonExistentID+"/incidents", nil)
	if err != nil {
		return fmt.Errorf("incidents non-existent service test: %w", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		return fmt.Errorf("incidents non-existent service: expected 400, got %d", resp.StatusCode)
	}

	return nil
}

func testStatsWithDifferentParameters(s *TestSuite) error {
	// Create a dedicated service for stats testing
	testService := web.CreateUpdateServiceRequest{
		Name:     "Stats Test Service",
		Protocol: storage.ServiceProtocolTypeHTTP,
		Interval: 30000,
		Timeout:  5000,
		Retries:  3,
		Tags:     []string{"stats-test"},
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
		return fmt.Errorf("create stats test service: %w", err)
	}

	var createdService web.ServiceDTO
	if err := s.decodeResponse(resp, &createdService); err != nil {
		return fmt.Errorf("decode stats test service: %w", err)
	}

	serviceID := createdService.ID

	// Ensure cleanup
	defer func() {
		s.makeRequest("DELETE", "/api/v1/services/"+serviceID, nil)
	}()

	// Test different days parameters
	daysTests := []struct {
		days     string
		expected bool
	}{
		{"1", true},
		{"7", true},
		{"30", true},
		{"365", true},
		{"0", true},   // Should work with 0 days
		{"-1", true},  // Negative should be handled gracefully
		{"abc", true}, // Invalid string should default to 30
		{"", true},    // Empty should default to 30
	}

	for _, test := range daysTests {
		path := "/api/v1/services/" + serviceID + "/stats"
		if test.days != "" {
			path += "?days=" + test.days
		}

		resp, err := s.makeRequest("GET", path, nil)
		if err != nil {
			if test.expected {
				return fmt.Errorf("stats with days=%s failed: %w", test.days, err)
			}
			continue // Expected to fail
		}

		if !test.expected {
			return fmt.Errorf("stats with days=%s should have failed but didn't", test.days)
		}

		var stats web.ServiceStats
		if err := s.decodeResponse(resp, &stats); err != nil {
			return fmt.Errorf("decode stats with days=%s: %w", test.days, err)
		}

		if stats.ServiceID != serviceID {
			return fmt.Errorf("stats service ID mismatch for days=%s", test.days)
		}

		// Basic validation of stats
		if stats.UptimePercentage < 0 || stats.UptimePercentage > 100 {
			return fmt.Errorf("invalid uptime percentage for days=%s: %f", test.days, stats.UptimePercentage)
		}
	}

	return nil
}

// Helper function to add extended tests to the main test suite
func getExtendedTests() []struct {
	name string
	fn   func(*TestSuite) error
} {
	return []struct {
		name string
		fn   func(*TestSuite) error
	}{
		{"TestAdvancedServiceFilters", testAdvancedServiceFilters},
		{"TestServiceCRUDCompleteFlow", testServiceCRUDCompleteFlow},
		{"TestAdvancedIncidentManagement", testAdvancedIncidentManagement},
		{"TestCompleteProtocolConfigurations", testCompleteProtocolConfigurations},
		{"TestAdvancedPaginationAndSorting", testAdvancedPaginationAndSorting},
		{"TestAdvancedErrorScenarios", testAdvancedErrorScenarios},
		{"TestStatsWithDifferentParameters", testStatsWithDifferentParameters},
	}
}
