package monitors

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dop251/goja"
	"github.com/sxwebdev/sentinel/internal/storage"
)

// HTTPConfig represents configuration for HTTP monitoring
type HTTPConfig struct {
	Timeout   time.Duration    `json:"timeout"`
	Endpoints []EndpointConfig `json:"endpoints"`
	Condition string           `json:"condition"`
}

// EndpointConfig represents a single endpoint configuration
type EndpointConfig struct {
	Name           string            `json:"name"`
	URL            string            `json:"url"`
	Method         string            `json:"method"`
	Headers        map[string]string `json:"headers"`
	Body           string            `json:"body"`
	ExpectedStatus int               `json:"expected_status"`
	JSONPath       string            `json:"json_path"` // Path to extract value from JSON response
	Username       string            `json:"username"`  // Basic Auth username
	Password       string            `json:"password"`  // Basic Auth password
}

// EndpointResult represents result from a single endpoint
type EndpointResult struct {
	Name     string        `json:"name"`
	URL      string        `json:"url"`
	Success  bool          `json:"success"`
	Value    any           `json:"value,omitempty"`
	Error    string        `json:"error,omitempty"`
	Response string        `json:"response,omitempty"`
	Duration time.Duration `json:"duration"`
}

// HTTPMonitor monitors HTTP/HTTPS endpoints
type HTTPMonitor struct {
	BaseMonitor
	conf    HTTPConfig
	retries int
}

// NewHTTPMonitor creates a new HTTP monitor
func NewHTTPMonitor(cfg storage.Service) (*HTTPMonitor, error) {
	conf, err := GetConfig[HTTPConfig](cfg.Config, storage.ServiceProtocolTypeHTTP)
	if err != nil {
		return nil, fmt.Errorf("HTTP config not found")
	}

	monitor := &HTTPMonitor{
		BaseMonitor: NewBaseMonitor(cfg),
		conf:        conf,
		retries:     cfg.Retries,
	}

	return monitor, nil
}

// Check performs a health check on the HTTP endpoint
func (h *HTTPMonitor) Check(ctx context.Context) error {
	return h.checkEndpoints(ctx)
}

// checkEndpoints performs health checks on multiple endpoints and evaluates conditions
func (h *HTTPMonitor) checkEndpoints(ctx context.Context) error {
	config := h.conf.Endpoints
	results := make([]EndpointResult, 0, len(config))

	// Check all endpoints concurrently
	type endpointResult struct {
		result EndpointResult
		index  int
	}

	resultChan := make(chan endpointResult, len(config))

	for i, endpoint := range config {
		go func(ep EndpointConfig, idx int) {
			result := h.checkEndpoint(ctx, ep)
			resultChan <- endpointResult{result: result, index: idx}
		}(endpoint, i)
	}

	// Collect results
	for i := 0; i < len(config); i++ {
		select {
		case result := <-resultChan:
			results = append(results, result.result)
		case <-ctx.Done():
			return fmt.Errorf("context cancelled during multi-endpoint check")
		}
	}

	// Evaluate condition
	conditionMet, err := evaluateCondition(h.conf.Condition, results)
	if err != nil {
		return fmt.Errorf("failed to evaluate condition: %w", err)
	}

	if conditionMet {
		return fmt.Errorf("multi-endpoint condition triggered, results: %v", results)
	}

	return nil
}

// checkEndpoint performs a health check on a single endpoint
func (h *HTTPMonitor) checkEndpoint(ctx context.Context, endpoint EndpointConfig) EndpointResult {
	start := time.Now()

	client := &http.Client{}
	client.Timeout = h.config.Timeout

	req, err := http.NewRequestWithContext(ctx, endpoint.Method, endpoint.URL, strings.NewReader(endpoint.Body))
	if err != nil {
		return EndpointResult{
			Name:     endpoint.Name,
			URL:      endpoint.URL,
			Success:  false,
			Error:    fmt.Sprintf("failed to create request: %v", err),
			Duration: time.Since(start),
		}
	}

	// Add headers
	for key, value := range endpoint.Headers {
		req.Header.Set(key, value)
	}

	// Add Basic Auth if username and password are provided
	if endpoint.Username != "" && endpoint.Password != "" {
		auth := endpoint.Username + ":" + endpoint.Password
		encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
		req.Header.Set("Authorization", "Basic "+encodedAuth)
	}

	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		return EndpointResult{
			Name:     endpoint.Name,
			URL:      endpoint.URL,
			Success:  false,
			Error:    err.Error(),
			Duration: duration,
		}
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return EndpointResult{
			Name:     endpoint.Name,
			URL:      endpoint.URL,
			Success:  false,
			Error:    fmt.Sprintf("failed to read response body: %v", err),
			Duration: duration,
		}
	}

	// Check if status code indicates success
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return EndpointResult{
			Name:     endpoint.Name,
			URL:      endpoint.URL,
			Success:  false,
			Error:    fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)),
			Response: string(body),
			Duration: duration,
		}
	}

	if endpoint.ExpectedStatus != 0 && resp.StatusCode != endpoint.ExpectedStatus {
		return EndpointResult{
			Name:     endpoint.Name,
			URL:      endpoint.URL,
			Success:  false,
			Error:    fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)),
			Response: string(body),
			Duration: duration,
		}
	}

	// Extract value from JSON if path is specified
	var value interface{}
	if endpoint.JSONPath != "" {
		value, err = extractValueFromJSON(body, endpoint.JSONPath)
		if err != nil {
			return EndpointResult{
				Name:     endpoint.Name,
				URL:      endpoint.URL,
				Success:  false,
				Error:    fmt.Sprintf("failed to extract value from JSON: %v", err),
				Response: string(body),
				Duration: duration,
			}
		}
	}

	return EndpointResult{
		Name:     endpoint.Name,
		URL:      endpoint.URL,
		Success:  true,
		Value:    value,
		Response: string(body),
		Duration: duration,
	}
}

// extractValueFromJSON extracts value from JSON response using JSONPath-like syntax
func extractValueFromJSON(data []byte, path string) (interface{}, error) {
	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if path == "" {
		return jsonData, nil
	}

	// Simple JSONPath implementation for common cases
	parts := strings.Split(path, ".")
	current := jsonData

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			if val, exists := v[part]; exists {
				current = val
			} else {
				return nil, fmt.Errorf("path not found: %s", path)
			}
		case []interface{}:
			// Handle array indexing like "items.0.name"
			if idx, err := parseInt(part); err == nil && idx >= 0 && idx < len(v) {
				current = v[idx]
			} else {
				return nil, fmt.Errorf("invalid array index: %s", part)
			}
		default:
			return nil, fmt.Errorf("cannot traverse path %s on type %T", part, current)
		}
	}

	return current, nil
}

// parseInt safely parses string to int
func parseInt(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}

// evaluateCondition evaluates JavaScript condition with endpoint results
func evaluateCondition(condition string, results []EndpointResult) (bool, error) {
	vm := goja.New()

	// Create results object for JavaScript
	resultsObj := make(map[string]any)
	for _, result := range results {
		resultsObj[result.Name] = map[string]any{
			"success":  result.Success,
			"value":    result.Value,
			"error":    result.Error,
			"response": result.Response,
			"duration": result.Duration.Milliseconds(),
		}
	}

	// Set global variables
	vm.Set("results", resultsObj)
	vm.Set("console", map[string]any{
		"log": func(args ...any) {
			fmt.Println("JS:", fmt.Sprint(args...))
		},
	})

	// Execute condition
	value, err := vm.RunString(condition)
	if err != nil {
		return false, fmt.Errorf("failed to execute condition: %w", err)
	}

	// Convert result to boolean
	return value.ToBoolean(), nil
}
