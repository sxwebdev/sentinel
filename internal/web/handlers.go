package web

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/template/html/v2"
	"github.com/sxwebdev/sentinel/internal/config"
	"github.com/sxwebdev/sentinel/internal/service"
	"github.com/sxwebdev/sentinel/internal/storage"
	"gopkg.in/yaml.v3"
)

//go:embed views/*
var viewsFS embed.FS

// FlatServiceConfig represents a service with flat config structure
type FlatServiceConfig struct {
	ID       string                `json:"id"`
	Name     string                `json:"name"`
	Protocol string                `json:"protocol"`
	Endpoint string                `json:"endpoint"`
	Interval time.Duration         `json:"interval"`
	Timeout  time.Duration         `json:"timeout"`
	Retries  int                   `json:"retries"`
	Tags     []string              `json:"tags"`
	Config   string                `json:"config"` // YAML string
	State    *storage.ServiceState `json:"state,omitempty"`
}

// ServiceTableDTO represents a service with incident statistics for table display
type ServiceTableDTO struct {
	ID              string                `json:"id"`
	Name            string                `json:"name"`
	Protocol        string                `json:"protocol"`
	Endpoint        string                `json:"endpoint"`
	Interval        time.Duration         `json:"interval"`
	Timeout         time.Duration         `json:"timeout"`
	Retries         int                   `json:"retries"`
	Tags            []string              `json:"tags"`
	Config          string                `json:"config"` // YAML string
	State           *storage.ServiceState `json:"state,omitempty"`
	ActiveIncidents int                   `json:"active_incidents"`
	TotalIncidents  int                   `json:"total_incidents"`
}

// Server represents the web server
type Server struct {
	monitorService *service.MonitorService
	config         *config.Config
	app            *fiber.App
	wsConnections  map[*websocket.Conn]bool
	wsMutex        sync.Mutex
}

// NewServer creates a new web server
func NewServer(monitorService *service.MonitorService, cfg *config.Config) (*Server, error) {
	// Create template engine with embedded templates
	templateEngine := html.NewFileSystem(http.FS(viewsFS), ".html")

	// Add template functions
	templateEngine.
		AddFunc("lower", strings.ToLower).
		AddFunc("statusToString", func(status storage.ServiceStatus) string {
			return status.String()
		}).
		AddFunc("statusToLower", func(status storage.ServiceStatus) string {
			return strings.ToLower(status.String())
		}).
		AddFunc("formatDateTime", func(t time.Time) string {
			return t.Format("2006-01-02 15:04:05")
		}).
		AddFunc("formatDateTimePtr", func(t *time.Time) string {
			if t == nil {
				return "Never"
			}
			return t.Format("2006-01-02 15:04:05")
		}).
		AddFunc("urlquery", url.QueryEscape)

	if err := templateEngine.Load(); err != nil {
		return nil, err
	}

	// Create Fiber app
	app := fiber.New(fiber.Config{
		Views: templateEngine,
	})

	app.Use(cors.New())

	// Serve static files from embed
	app.Get("/static/*", func(c *fiber.Ctx) error {
		filePath := c.Params("*")
		if filePath == "" {
			return c.Status(404).SendString("File not found")
		}

		content, err := viewsFS.ReadFile("views/static/" + filePath)
		if err != nil {
			return c.Status(404).SendString("File not found")
		}

		// Set content type based on file extension
		if strings.HasSuffix(filePath, ".css") {
			c.Set("Content-Type", "text/css")
		} else if strings.HasSuffix(filePath, ".js") {
			c.Set("Content-Type", "application/javascript")
		}

		return c.Send(content)
	})

	server := &Server{
		monitorService: monitorService,
		config:         cfg,
		app:            app,
		wsConnections:  make(map[*websocket.Conn]bool),
	}

	// Setup routes
	server.setupRoutes()

	return server, nil
}

// setupRoutes configures all routes
func (s *Server) setupRoutes() {
	// Web UI routes
	s.app.Get("/", s.handleDashboard)
	s.app.Get("/service/:id", s.handleServiceDetail)

	// API routes
	api := s.app.Group("/api")
	api.Get("/services", s.handleAPIServices)
	api.Get("/services/table", s.handleAPIServicesTable)
	api.Get("/services/:id", s.handleAPIServiceDetail)
	api.Get("/services/:id/incidents", s.handleAPIServiceIncidents)
	api.Get("/services/:id/stats", s.handleAPIServiceStats)
	api.Post("/services/:id/check", s.handleAPIServiceCheck)
	api.Post("/services/:id/resolve", s.handleAPIServiceResolve)
	api.Get("/incidents", s.handleAPIRecentIncidents)
	api.Get("/dashboard/stats", s.handleAPIDashboardStats)

	// Service management API
	api.Post("/services", s.handleAPICreateService)
	api.Put("/services/:id", s.handleAPIUpdateService)
	api.Delete("/services/:id", s.handleAPIDeleteService)
	api.Get("/services/config/:id", s.handleAPIGetServiceConfig)

	// WebSocket endpoint
	s.app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	s.app.Get("/ws", websocket.New(s.handleWebSocket))
}

// App returns the Fiber app instance
func (s *Server) App() *fiber.App {
	return s.app
}

// handleDashboard renders the main dashboard
func (s *Server) handleDashboard(c *fiber.Ctx) error {
	return c.Render("views/dashboard", fiber.Map{
		"Title": "Sentinel Dashboard",
		"Actions": []fiber.Map{
			{
				"Text":  "Refresh",
				"Class": "btn-secondary",
			},
		},
	})
}

// handleServiceDetail renders service detail page
func (s *Server) handleServiceDetail(c *fiber.Ctx) error {
	serviceID := c.Params("id")
	if serviceID == "" {
		return c.Status(fiber.StatusBadRequest).SendString("service ID is required")
	}

	// Get service by ID
	targetService, err := s.monitorService.GetServiceByID(c.Context(), serviceID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("service not found: " + serviceID)
	}

	// Get recent incidents
	incidents, err := s.monitorService.GetServiceIncidents(c.Context(), targetService.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	// Get service stats
	stats, err := s.monitorService.GetServiceStats(c.Context(), targetService.ID, time.Now().AddDate(0, 0, -30))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.Render("views/service_detail", fiber.Map{
		"Service":      targetService,
		"State":        targetService.State,
		"Incidents":    incidents,
		"Stats":        stats,
		"Title":        "Service: " + targetService.Name,
		"BackLink":     "/",
		"BackLinkText": "Back to Dashboard",
		"Actions": []fiber.Map{
			{
				"Text":  "Trigger Check",
				"Class": "",
			},
			{
				"Text":  "Resolve Incidents",
				"Class": "btn-secondary",
			},
		},
	})
}

// handleAPIServices returns all services
func (s *Server) handleAPIServices(c *fiber.Ctx) error {
	// Get all services with their states
	services, err := s.monitorService.GetAllServiceConfigs(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(services)
}

// handleAPIServicesTable returns services with incident statistics for table display
func (s *Server) handleAPIServicesTable(c *fiber.Ctx) error {
	// Get all services with their states
	services, err := s.monitorService.GetAllServiceConfigs(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Get incident statistics for all services
	incidentStats, err := s.monitorService.GetAllServicesIncidentStats(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Create a map for quick lookup of incident stats by service ID
	statsMap := make(map[string]*storage.ServiceIncidentStats)
	for _, stats := range incidentStats {
		statsMap[stats.ServiceID] = stats
	}

	// Convert services to DTO with incident statistics
	serviceDTOs := make([]ServiceTableDTO, len(services))
	for i, service := range services {
		// Convert config to YAML string
		configYAML, err := s.convertConfigToYAML(service.Config)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to convert service config: " + err.Error(),
			})
		}

		dto := ServiceTableDTO{
			ID:              service.ID,
			Name:            service.Name,
			Protocol:        service.Protocol,
			Endpoint:        service.Endpoint,
			Interval:        service.Interval,
			Timeout:         service.Timeout,
			Retries:         service.Retries,
			Tags:            service.Tags,
			Config:          configYAML,
			State:           service.State,
			ActiveIncidents: 0,
			TotalIncidents:  0,
		}

		// Add incident statistics if available
		if stats, exists := statsMap[service.ID]; exists {
			dto.ActiveIncidents = stats.ActiveIncidents
			dto.TotalIncidents = stats.TotalIncidents
		}

		serviceDTOs[i] = dto
	}

	return c.JSON(serviceDTOs)
}

// handleAPIServiceDetail returns service details
func (s *Server) handleAPIServiceDetail(c *fiber.Ctx) error {
	serviceID := c.Params("id")
	if serviceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "service ID is required",
		})
	}

	targetService, err := s.monitorService.GetServiceByID(c.Context(), serviceID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "service not found: " + serviceID,
		})
	}

	return c.JSON(targetService)
}

// handleAPIServiceIncidents returns service incidents
func (s *Server) handleAPIServiceIncidents(c *fiber.Ctx) error {
	serviceID := c.Params("id")
	if serviceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "service ID is required",
		})
	}

	incidents, err := s.monitorService.GetServiceIncidents(c.Context(), serviceID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(incidents)
}

// handleAPIServiceStats returns service statistics
func (s *Server) handleAPIServiceStats(c *fiber.Ctx) error {
	serviceID := c.Params("id")
	if serviceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "service ID is required",
		})
	}

	daysStr := c.Query("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil {
		days = 30
	}

	since := time.Now().AddDate(0, 0, -days)
	stats, err := s.monitorService.GetServiceStats(c.Context(), serviceID, since)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(stats)
}

// handleAPIServiceCheck triggers a manual check
func (s *Server) handleAPIServiceCheck(c *fiber.Ctx) error {
	serviceID := c.Params("id")
	if serviceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "service ID is required",
		})
	}

	// First check if service exists
	_, err := s.monitorService.GetServiceByID(c.Context(), serviceID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "service not found: " + serviceID,
		})
	}

	err = s.monitorService.TriggerCheck(c.Context(), serviceID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "check triggered successfully",
	})
}

// handleAPIServiceResolve resolves a service incident
func (s *Server) handleAPIServiceResolve(c *fiber.Ctx) error {
	serviceID := c.Params("id")
	if serviceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "service ID is required",
		})
	}

	service, err := s.monitorService.GetServiceByID(c.Context(), serviceID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "service not found: " + serviceID,
		})
	}

	err = s.monitorService.ForceResolveIncidents(c.Context(), serviceID, service.Name)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "incident resolved successfully",
	})
}

// handleAPIRecentIncidents returns recent incidents
func (s *Server) handleAPIRecentIncidents(c *fiber.Ctx) error {
	limitStr := c.Query("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}

	incidents, err := s.monitorService.GetRecentIncidents(c.Context(), limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(incidents)
}

// handleAPIDashboardStats returns dashboard statistics
func (s *Server) handleAPIDashboardStats(c *fiber.Ctx) error {
	// Get all services
	services, err := s.monitorService.GetAllServiceConfigs(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Get recent incidents
	recentIncidents, err := s.monitorService.GetRecentIncidents(c.Context(), 100)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Calculate statistics
	stats := fiber.Map{
		"total_services":    len(services),
		"services_up":       0,
		"services_down":     0,
		"services_unknown":  0,
		"protocols":         make(map[string]int),
		"recent_incidents":  len(recentIncidents),
		"active_incidents":  0,
		"avg_response_time": 0,
		"total_checks":      0,
		"uptime_percentage": 0,
		"last_check_time":   nil,
		"checks_per_minute": 0,
	}

	// Calculate service status distribution and protocol distribution
	totalResponseTime := time.Duration(0)
	servicesWithResponseTime := 0
	totalChecks := 0
	upServices := 0
	var lastCheckTime *time.Time

	for _, service := range services {
		// Count by status
		if service.State != nil {
			switch service.State.Status {
			case storage.StatusUp:
				stats["services_up"] = stats["services_up"].(int) + 1
				upServices++
			case storage.StatusDown:
				stats["services_down"] = stats["services_down"].(int) + 1
			case storage.StatusUnknown:
				stats["services_unknown"] = stats["services_unknown"].(int) + 1
			}

			// Sum response times (only from services that have response time data)
			if service.State.ResponseTime > 0 {
				totalResponseTime += service.State.ResponseTime
				servicesWithResponseTime++
			}
			totalChecks += service.State.TotalChecks

			// Track last check time
			if service.State.LastCheck != nil {
				if lastCheckTime == nil || service.State.LastCheck.After(*lastCheckTime) {
					lastCheckTime = service.State.LastCheck
				}
			}
		}

		// Count by protocol
		protocol := service.Protocol
		if protocol == "" {
			protocol = "unknown"
		}
		stats["protocols"].(map[string]int)[protocol]++
	}

	// Calculate averages
	if upServices > 0 {
		stats["uptime_percentage"] = float64(upServices) / float64(len(services)) * 100
	}
	if servicesWithResponseTime > 0 {
		stats["avg_response_time"] = totalResponseTime.Milliseconds() / int64(servicesWithResponseTime)
	}
	stats["total_checks"] = totalChecks

	// Count active incidents
	activeIncidents := 0
	for _, incident := range recentIncidents {
		if !incident.Resolved {
			activeIncidents++
		}
	}
	stats["active_incidents"] = activeIncidents

	// Set last check time
	stats["last_check_time"] = lastCheckTime

	// Calculate checks per minute (estimate based on intervals)
	checksPerMinute := 0
	for _, service := range services {
		if service.Interval > 0 {
			checksPerMinute += int(time.Minute / service.Interval)
		}
	}
	stats["checks_per_minute"] = checksPerMinute

	return c.JSON(stats)
}

// handleAPICreateService creates a new service
func (s *Server) handleAPICreateService(c *fiber.Ctx) error {
	var flatService FlatServiceConfig
	if err := c.BodyParser(&flatService); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body: " + err.Error(),
		})
	}

	// Validate required fields
	if flatService.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Service name is required",
		})
	}
	if flatService.Protocol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Protocol is required",
		})
	}
	if flatService.Endpoint == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Endpoint is required",
		})
	}

	// Convert to storage.Service
	service := storage.Service{
		ID:       flatService.ID,
		Name:     flatService.Name,
		Protocol: flatService.Protocol,
		Endpoint: flatService.Endpoint,
		Interval: flatService.Interval,
		Timeout:  flatService.Timeout,
		Retries:  flatService.Retries,
		Tags:     flatService.Tags,
		State:    flatService.State,
	}

	// Set default values
	if service.Interval == 0 {
		service.Interval = s.config.Monitoring.Global.DefaultInterval
	}
	if service.Timeout == 0 {
		service.Timeout = s.config.Monitoring.Global.DefaultTimeout
	}
	if service.Retries == 0 {
		service.Retries = s.config.Monitoring.Global.DefaultRetries
	}

	// Convert flat config to proper MonitorConfig structure
	config, err := s.convertFlatConfigToMonitorConfig(flatService.Protocol, flatService.Config)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid config: " + err.Error(),
		})
	}
	service.Config = config

	// Add service
	if err := s.monitorService.AddService(c.Context(), &service); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create service: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(service)
}

// handleAPIUpdateService updates an existing service
func (s *Server) handleAPIUpdateService(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Service ID is required",
		})
	}

	var flatService FlatServiceConfig
	if err := c.BodyParser(&flatService); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body: " + err.Error(),
		})
	}

	// Convert to storage.Service
	service := storage.Service{
		ID:       id,
		Name:     flatService.Name,
		Protocol: flatService.Protocol,
		Endpoint: flatService.Endpoint,
		Interval: flatService.Interval,
		Timeout:  flatService.Timeout,
		Retries:  flatService.Retries,
		Tags:     flatService.Tags,
		State:    flatService.State,
	}

	// Convert flat config to proper MonitorConfig structure
	config, err := s.convertFlatConfigToMonitorConfig(flatService.Protocol, flatService.Config)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid config: " + err.Error(),
		})
	}
	service.Config = config

	// Validate required fields
	if service.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Service name is required",
		})
	}
	if service.Protocol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Protocol is required",
		})
	}
	if service.Endpoint == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Endpoint is required",
		})
	}

	// Set default values if not provided
	if service.Interval == 0 {
		service.Interval = s.config.Monitoring.Global.DefaultInterval
	}
	if service.Timeout == 0 {
		service.Timeout = s.config.Monitoring.Global.DefaultTimeout
	}
	if service.Retries == 0 {
		service.Retries = s.config.Monitoring.Global.DefaultRetries
	}

	// Update service
	if err := s.monitorService.UpdateService(c.Context(), &service); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update service: " + err.Error(),
		})
	}

	return c.JSON(service)
}

// handleAPIDeleteService deletes a service
func (s *Server) handleAPIDeleteService(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Service ID is required",
		})
	}

	if err := s.monitorService.DeleteService(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete service: " + err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// handleAPIGetServiceConfig gets service configuration by ID
func (s *Server) handleAPIGetServiceConfig(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Service ID is required",
		})
	}

	service, err := s.monitorService.GetServiceByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Service not found: " + err.Error(),
		})
	}

	return c.JSON(service)
}

// convertFlatConfigToMonitorConfig converts YAML config string to proper MonitorConfig structure
func (s *Server) convertFlatConfigToMonitorConfig(protocol string, yamlConfig string) (storage.MonitorConfig, error) {
	if yamlConfig == "" {
		// Return default config based on protocol
		return s.getDefaultConfig(protocol), nil
	}

	// Parse YAML into map
	var configMap map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlConfig), &configMap); err != nil {
		return storage.MonitorConfig{}, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	// Validate and convert based on protocol
	switch protocol {
	case "http", "https":
		return s.parseHTTPConfig(configMap)
	case "tcp":
		return s.parseTCPConfig(configMap)
	case "grpc":
		return s.parseGRPCConfig(configMap)
	case "redis":
		return s.parseRedisConfig(configMap)
	default:
		return storage.MonitorConfig{}, fmt.Errorf("unsupported protocol: %s", protocol)
	}
}

// getDefaultConfig returns default config for a protocol
func (s *Server) getDefaultConfig(protocol string) storage.MonitorConfig {
	switch protocol {
	case "http", "https":
		return storage.MonitorConfig{
			HTTP: &storage.HTTPConfig{
				Method:         "GET",
				ExpectedStatus: 200,
				Headers:        make(map[string]string),
			},
		}
	case "tcp":
		return storage.MonitorConfig{
			TCP: &storage.TCPConfig{
				SendData:   "",
				ExpectData: "",
			},
		}
	case "grpc":
		return storage.MonitorConfig{
			GRPC: &storage.GRPCConfig{
				CheckType:   "health",
				ServiceName: "",
				TLS:         false,
				InsecureTLS: false,
			},
		}
	case "redis":
		return storage.MonitorConfig{
			Redis: &storage.RedisConfig{
				Password: "",
				DB:       0,
			},
		}
	default:
		return storage.MonitorConfig{}
	}
}

// parseHTTPConfig parses and validates HTTP config
func (s *Server) parseHTTPConfig(configMap map[string]interface{}) (storage.MonitorConfig, error) {
	httpConfig := &storage.HTTPConfig{
		Method:         "GET",
		ExpectedStatus: 200,
		Headers:        make(map[string]string),
	}

	// Extract and validate method
	if method, ok := configMap["method"].(string); ok {
		method = strings.ToUpper(method)
		if method != "GET" && method != "POST" && method != "PUT" && method != "DELETE" && method != "HEAD" && method != "OPTIONS" {
			return storage.MonitorConfig{}, fmt.Errorf("invalid HTTP method: %s", method)
		}
		httpConfig.Method = method
	}

	// Extract and validate expected status
	if expectedStatus, ok := configMap["expected_status"].(int); ok {
		if expectedStatus < 100 || expectedStatus > 599 {
			return storage.MonitorConfig{}, fmt.Errorf("invalid HTTP status code: %d", expectedStatus)
		}
		httpConfig.ExpectedStatus = expectedStatus
	} else if expectedStatus, ok := configMap["expected_status"].(float64); ok {
		status := int(expectedStatus)
		if status < 100 || status > 599 {
			return storage.MonitorConfig{}, fmt.Errorf("invalid HTTP status code: %d", status)
		}
		httpConfig.ExpectedStatus = status
	}

	// Extract headers
	if headers, ok := configMap["headers"].(map[string]interface{}); ok {
		for key, value := range headers {
			if strValue, ok := value.(string); ok {
				httpConfig.Headers[key] = strValue
			} else {
				return storage.MonitorConfig{}, fmt.Errorf("invalid header value for %s: must be string", key)
			}
		}
	}

	// Check for unknown fields
	allowedFields := map[string]bool{"method": true, "expected_status": true, "headers": true}
	for field := range configMap {
		if !allowedFields[field] {
			return storage.MonitorConfig{}, fmt.Errorf("unknown field in HTTP config: %s", field)
		}
	}

	return storage.MonitorConfig{HTTP: httpConfig}, nil
}

// parseTCPConfig parses and validates TCP config
func (s *Server) parseTCPConfig(configMap map[string]interface{}) (storage.MonitorConfig, error) {
	tcpConfig := &storage.TCPConfig{
		SendData:   "",
		ExpectData: "",
	}

	// Extract send_data
	if sendData, ok := configMap["send_data"].(string); ok {
		tcpConfig.SendData = sendData
	}

	// Extract expect_data
	if expectData, ok := configMap["expect_data"].(string); ok {
		tcpConfig.ExpectData = expectData
	}

	// Check for unknown fields
	allowedFields := map[string]bool{"send_data": true, "expect_data": true}
	for field := range configMap {
		if !allowedFields[field] {
			return storage.MonitorConfig{}, fmt.Errorf("unknown field in TCP config: %s", field)
		}
	}

	return storage.MonitorConfig{TCP: tcpConfig}, nil
}

// parseGRPCConfig parses and validates gRPC config
func (s *Server) parseGRPCConfig(configMap map[string]interface{}) (storage.MonitorConfig, error) {
	grpcConfig := &storage.GRPCConfig{
		CheckType:   "health",
		ServiceName: "",
		TLS:         false,
		InsecureTLS: false,
	}

	// Extract check_type
	if checkType, ok := configMap["check_type"].(string); ok {
		if checkType != "health" && checkType != "reflection" && checkType != "connectivity" {
			return storage.MonitorConfig{}, fmt.Errorf("invalid gRPC check type: %s", checkType)
		}
		grpcConfig.CheckType = checkType
	}

	// Extract service_name
	if serviceName, ok := configMap["service_name"].(string); ok {
		grpcConfig.ServiceName = serviceName
	}

	// Extract TLS settings
	if tls, ok := configMap["tls"].(bool); ok {
		grpcConfig.TLS = tls
	}
	if insecureTLS, ok := configMap["insecure_tls"].(bool); ok {
		grpcConfig.InsecureTLS = insecureTLS
	}

	// Check for unknown fields
	allowedFields := map[string]bool{"check_type": true, "service_name": true, "tls": true, "insecure_tls": true}
	for field := range configMap {
		if !allowedFields[field] {
			return storage.MonitorConfig{}, fmt.Errorf("unknown field in gRPC config: %s", field)
		}
	}

	return storage.MonitorConfig{GRPC: grpcConfig}, nil
}

// parseRedisConfig parses and validates Redis config
func (s *Server) parseRedisConfig(configMap map[string]interface{}) (storage.MonitorConfig, error) {
	redisConfig := &storage.RedisConfig{
		Password: "",
		DB:       0,
	}

	// Extract password
	if password, ok := configMap["password"].(string); ok {
		redisConfig.Password = password
	}

	// Extract DB number
	if db, ok := configMap["db"].(int); ok {
		if db < 0 || db > 15 {
			return storage.MonitorConfig{}, fmt.Errorf("invalid Redis DB number: %d (must be 0-15)", db)
		}
		redisConfig.DB = db
	} else if db, ok := configMap["db"].(float64); ok {
		dbNum := int(db)
		if dbNum < 0 || dbNum > 15 {
			return storage.MonitorConfig{}, fmt.Errorf("invalid Redis DB number: %d (must be 0-15)", dbNum)
		}
		redisConfig.DB = dbNum
	}

	// Check for unknown fields
	allowedFields := map[string]bool{"password": true, "db": true}
	for field := range configMap {
		if !allowedFields[field] {
			return storage.MonitorConfig{}, fmt.Errorf("unknown field in Redis config: %s", field)
		}
	}

	return storage.MonitorConfig{Redis: redisConfig}, nil
}

// convertConfigToYAML converts MonitorConfig to YAML string
func (s *Server) convertConfigToYAML(config storage.MonitorConfig) (string, error) {
	data, err := yaml.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("failed to marshal config to YAML: %w", err)
	}
	return string(data), nil
}

// handleWebSocket handles WebSocket connections
func (s *Server) handleWebSocket(c *websocket.Conn) {
	// Add connection to the map
	s.wsMutex.Lock()
	s.wsConnections[c] = true
	s.wsMutex.Unlock()

	// Remove connection when it closes
	defer func() {
		s.wsMutex.Lock()
		delete(s.wsConnections, c)
		s.wsMutex.Unlock()
		c.Close()
	}()

	// Send initial data
	if err := s.sendServiceUpdate(c); err != nil {
		return
	}

	// Keep connection alive and handle messages
	for {
		_, _, err := c.ReadMessage()
		if err != nil {
			break
		}
	}
}

// sendServiceUpdate sends service updates to a specific WebSocket connection
func (s *Server) sendServiceUpdate(conn *websocket.Conn) error {
	// Get all services with incident statistics (without locking)
	services, err := s.monitorService.GetAllServiceConfigs(context.Background())
	if err != nil {
		return err
	}

	incidentStats, err := s.monitorService.GetAllServicesIncidentStats(context.Background())
	if err != nil {
		return err
	}

	// Create a map for quick lookup of incident stats by service ID
	statsMap := make(map[string]*storage.ServiceIncidentStats)
	for _, stats := range incidentStats {
		statsMap[stats.ServiceID] = stats
	}

	// Convert services to DTO with incident statistics
	serviceDTOs := make([]ServiceTableDTO, len(services))
	for i, service := range services {
		// Convert config to YAML string
		configYAML, err := s.convertConfigToYAML(service.Config)
		if err != nil {
			continue
		}

		dto := ServiceTableDTO{
			ID:              service.ID,
			Name:            service.Name,
			Protocol:        service.Protocol,
			Endpoint:        service.Endpoint,
			Interval:        service.Interval,
			Timeout:         service.Timeout,
			Retries:         service.Retries,
			Tags:            service.Tags,
			Config:          configYAML,
			State:           service.State,
			ActiveIncidents: 0,
			TotalIncidents:  0,
		}

		// Add incident statistics if available
		if stats, exists := statsMap[service.ID]; exists {
			dto.ActiveIncidents = stats.ActiveIncidents
			dto.TotalIncidents = stats.TotalIncidents
		}

		serviceDTOs[i] = dto
	}

	// Send update message
	update := fiber.Map{
		"type":      "service_update",
		"services":  serviceDTOs,
		"timestamp": time.Now().Unix(),
	}

	// Lock only for WebSocket write
	s.wsMutex.Lock()
	defer s.wsMutex.Unlock()

	return conn.WriteJSON(update)
}

// BroadcastServiceUpdate sends service updates to all connected WebSocket clients
func (s *Server) BroadcastServiceUpdate() {
	// Get updated service data (without locking)
	services, err := s.monitorService.GetAllServiceConfigs(context.Background())
	if err != nil {
		return
	}

	incidentStats, err := s.monitorService.GetAllServicesIncidentStats(context.Background())
	if err != nil {
		return
	}

	// Create a map for quick lookup of incident stats by service ID
	statsMap := make(map[string]*storage.ServiceIncidentStats)
	for _, stats := range incidentStats {
		statsMap[stats.ServiceID] = stats
	}

	// Convert services to DTO with incident statistics
	serviceDTOs := make([]ServiceTableDTO, len(services))
	for i, service := range services {
		// Convert config to YAML string
		configYAML, err := s.convertConfigToYAML(service.Config)
		if err != nil {
			continue
		}

		dto := ServiceTableDTO{
			ID:              service.ID,
			Name:            service.Name,
			Protocol:        service.Protocol,
			Endpoint:        service.Endpoint,
			Interval:        service.Interval,
			Timeout:         service.Timeout,
			Retries:         service.Retries,
			Tags:            service.Tags,
			Config:          configYAML,
			State:           service.State,
			ActiveIncidents: 0,
			TotalIncidents:  0,
		}

		// Add incident statistics if available
		if stats, exists := statsMap[service.ID]; exists {
			dto.ActiveIncidents = stats.ActiveIncidents
			dto.TotalIncidents = stats.TotalIncidents
		}

		serviceDTOs[i] = dto
	}

	// Prepare update message
	update := fiber.Map{
		"type":      "service_update",
		"services":  serviceDTOs,
		"timestamp": time.Now().Unix(),
	}

	// Lock only for WebSocket operations
	s.wsMutex.Lock()
	defer s.wsMutex.Unlock()

	// Send to all connections
	for conn := range s.wsConnections {
		if err := conn.WriteJSON(update); err != nil {
			// Remove failed connection
			delete(s.wsConnections, conn)
			conn.Close()
		}
	}
}
