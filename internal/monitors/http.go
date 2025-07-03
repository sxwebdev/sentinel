package monitors

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"

	"github.com/sxwebdev/sentinel/internal/storage"
)

// HTTPMonitor monitors HTTP/HTTPS endpoints
type HTTPMonitor struct {
	BaseMonitor
	client         *http.Client
	method         string
	expectedStatus int
	headers        map[string]string
	body           string
	expectedText   string
}

// NewHTTPMonitor creates a new HTTP monitor
func NewHTTPMonitor(cfg storage.Service) (*HTTPMonitor, error) {
	// Extract HTTP config
	var httpConfig *storage.HTTPConfig
	if cfg.Config.HTTP != nil {
		httpConfig = cfg.Config.HTTP
	}

	monitor := &HTTPMonitor{
		BaseMonitor:    NewBaseMonitor(cfg),
		method:         "GET",
		expectedStatus: 200,
		headers:        make(map[string]string),
	}

	// Apply HTTP-specific config if available
	if httpConfig != nil {
		if httpConfig.Method != "" {
			monitor.method = httpConfig.Method
		}
		if httpConfig.ExpectedStatus != 0 {
			monitor.expectedStatus = httpConfig.ExpectedStatus
		}
		if httpConfig.Headers != nil {
			monitor.headers = httpConfig.Headers
		}
	}

	// Create HTTP client with timeout
	monitor.client = &http.Client{
		Timeout: cfg.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Allow up to 5 redirects
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	// Validate method
	validMethods := []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"}
	method := strings.ToUpper(monitor.method)
	valid := slices.Contains(validMethods, method)
	if !valid {
		return nil, fmt.Errorf("invalid HTTP method: %s", monitor.method)
	}
	monitor.method = method

	return monitor, nil
}

// Check performs the HTTP health check
func (h *HTTPMonitor) Check(ctx context.Context) error {
	// Create request
	var bodyReader io.Reader
	if h.body != "" {
		bodyReader = strings.NewReader(h.body)
	}

	req, err := http.NewRequestWithContext(ctx, h.method, h.config.Endpoint, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for key, value := range h.headers {
		req.Header.Set(key, value)
	}

	// Set default headers if not provided
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "Sentinel-Monitor/1.0")
	}
	if h.body != "" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Make request
	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != h.expectedStatus {
		return fmt.Errorf("unexpected status code: got %d, expected %d", resp.StatusCode, h.expectedStatus)
	}

	// Check response body if expected text is specified
	if h.expectedText != "" {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		if !strings.Contains(string(body), h.expectedText) {
			return fmt.Errorf("expected text not found in response: %s", h.expectedText)
		}
	}

	return nil
}
