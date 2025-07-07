// Package web provides HTTP handlers for the Sentinel monitoring system
//
//	@title			Sentinel Monitoring API
//	@version		1.0
//	@description	API for service monitoring and incident management
//	@termsOfService	http://swagger.io/terms/
//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io
//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html
//	@BasePath		/api/v1
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
	swagger "github.com/swaggo/fiber-swagger"
	"github.com/sxwebdev/sentinel/docs/docsv1"
	"github.com/sxwebdev/sentinel/internal/config"
	"github.com/sxwebdev/sentinel/internal/receiver"
	"github.com/sxwebdev/sentinel/internal/service"
	"github.com/sxwebdev/sentinel/internal/storage"
)

//go:embed views/*
var viewsFS embed.FS

// ServiceDTO represents a service with optional incident statistics
//
//	@Description	Service with optional incident statistics
type ServiceDTO struct {
	ID              string                `json:"id" example:"service-1"`
	Name            string                `json:"name" example:"Web Server"`
	Protocol        string                `json:"protocol" example:"http"`
	Endpoint        string                `json:"endpoint" example:"https://example.com"`
	Interval        time.Duration         `json:"interval" swaggertype:"primitive,integer" example:"30000000000"`
	Timeout         time.Duration         `json:"timeout" swaggertype:"primitive,integer" example:"5000000000"`
	Retries         int                   `json:"retries" example:"3"`
	Tags            []string              `json:"tags" example:"web,production"`
	Config          map[string]any        `json:"config"` // JSON object
	State           *storage.ServiceState `json:"state,omitempty"`
	IsEnabled       bool                  `json:"is_enabled" example:"true"`
	ActiveIncidents int                   `json:"active_incidents,omitempty" example:"2"`
	TotalIncidents  int                   `json:"total_incidents,omitempty" example:"10"`
}

// Server represents the web server
type Server struct {
	monitorService *service.MonitorService
	receiver       *receiver.Receiver
	config         *config.Config
	app            *fiber.App
	wsConnections  map[*websocket.Conn]bool
	wsMutex        sync.Mutex
}

// NewServer creates a new web server
func NewServer(
	cfg *config.Config,
	monitorService *service.MonitorService,
	receiver *receiver.Receiver,
) (*Server, error) {
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
		config:         cfg,
		monitorService: monitorService,
		receiver:       receiver,
		app:            app,
		wsConnections:  make(map[*websocket.Conn]bool),
	}

	// Set Swagger host from config
	docsv1.SwaggerInfo.Host = cfg.Server.BaseHost
	docsv1.SwaggerInfo.BasePath = "/api/v1"
	docsv1.SwaggerInfo.Schemes = []string{"http", "https"}

	// Setup routes
	server.setupRoutes()

	return server, nil
}

// Start starts the web server
func (s *Server) Start(ctx context.Context) error {
	errChan := make(chan error, 1)
	go func() {
		errChan <- s.subscribeEvents(ctx)
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
	}

	return nil
}

func (s *Server) Stop(_ context.Context) error {
	return nil
}

// setupRoutes configures all routes
func (s *Server) setupRoutes() {
	// Web UI routes
	s.app.Get("/", s.handleDashboard)
	s.app.Get("/service/:id", s.handleServiceDetail)

	// API routes
	api := s.app.Group("/api/v1")
	// Swagger UI
	api.Get("/swagger/*", swagger.WrapHandler)
	api.Get("/services", s.handleAPIServices)
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

// handleAPIServices returns all services with incident statistics
//
//	@Summary		Get all services
//	@Description	Returns a list of all configured services with incident statistics
//	@Tags			services
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}		ServiceDTO		"List of services with statistics"
//	@Failure		500	{object}	ErrorResponse	"Internal server error"
//	@Router			/services [get]
func (s *Server) handleAPIServices(c *fiber.Ctx) error {
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
	var result []ServiceDTO
	for _, service := range services {
		dto := ServiceDTO{
			ID:              service.ID,
			Name:            service.Name,
			Protocol:        service.Protocol,
			Endpoint:        service.Endpoint,
			Interval:        service.Interval,
			Timeout:         service.Timeout,
			Retries:         service.Retries,
			Tags:            service.Tags,
			Config:          s.convertConfigToMap(service.Config),
			State:           service.State,
			IsEnabled:       service.IsEnabled,
			ActiveIncidents: 0,
			TotalIncidents:  0,
		}

		// Add incident statistics if available
		if stats, exists := statsMap[service.ID]; exists {
			dto.ActiveIncidents = stats.ActiveIncidents
			dto.TotalIncidents = stats.TotalIncidents
		}

		result = append(result, dto)
	}

	return c.JSON(result)
}

// handleAPIServiceDetail returns service details
//
//	@Summary		Get service details
//	@Description	Returns detailed information about a specific service
//	@Tags			services
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string			true	"Service ID"
//	@Success		200	{object}	ServiceDTO		"Service details"
//	@Failure		400	{object}	ErrorResponse	"Bad request"
//	@Failure		404	{object}	ErrorResponse	"Service not found"
//	@Router			/services/{id} [get]
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
//
//	@Summary		Get service incidents
//	@Description	Returns a list of incidents for a specific service
//	@Tags			incidents
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string			true	"Service ID"
//	@Success		200	{array}		Incident		"List of incidents"
//	@Failure		400	{object}	ErrorResponse	"Bad request"
//	@Failure		500	{object}	ErrorResponse	"Internal server error"
//	@Router			/services/{id}/incidents [get]
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
//
//	@Summary		Get service statistics
//	@Description	Returns service statistics for the specified period
//	@Tags			statistics
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string			true	"Service ID"
//	@Param			days	query		int				false	"Number of days (default 30)"
//	@Success		200		{object}	ServiceStats	"Service statistics"
//	@Failure		400		{object}	ErrorResponse	"Bad request"
//	@Failure		500		{object}	ErrorResponse	"Internal server error"
//	@Router			/services/{id}/stats [get]
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
//
//	@Summary		Trigger service check
//	@Description	Triggers a manual check of service status
//	@Tags			services
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string			true	"Service ID"
//	@Success		200	{object}	SuccessResponse	"Check triggered successfully"
//	@Failure		400	{object}	ErrorResponse	"Bad request"
//	@Failure		404	{object}	ErrorResponse	"Service not found"
//	@Failure		500	{object}	ErrorResponse	"Internal server error"
//	@Router			/services/{id}/check [post]
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
//
//	@Summary		Resolve service incidents
//	@Description	Forcefully resolves all active incidents for a service
//	@Tags			incidents
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string			true	"Service ID"
//	@Success		200	{object}	SuccessResponse	"Incidents resolved successfully"
//	@Failure		400	{object}	ErrorResponse	"Bad request"
//	@Failure		404	{object}	ErrorResponse	"Service not found"
//	@Failure		500	{object}	ErrorResponse	"Internal server error"
//	@Router			/services/{id}/resolve [post]
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
//
//	@Summary		Get recent incidents
//	@Description	Returns a list of recent incidents across all services
//	@Tags			incidents
//	@Accept			json
//	@Produce		json
//	@Param			limit	query		int				false	"Number of incidents (default 50)"
//	@Success		200		{array}		Incident		"List of incidents"
//	@Failure		500		{object}	ErrorResponse	"Internal server error"
//	@Router			/incidents [get]
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
//
//	@Summary		Get dashboard statistics
//	@Description	Returns overall statistics for the dashboard
//	@Tags			statistics
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	DashboardStats	"Dashboard statistics"
//	@Failure		500	{object}	ErrorResponse	"Internal server error"
//	@Router			/dashboard/stats [get]
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
//
//	@Summary		Create new service
//	@Description	Creates a new service for monitoring
//	@Tags			services
//	@Accept			json
//	@Produce		json
//	@Param			service	body		ServiceDTO		true	"Service configuration"
//	@Success		201		{object}	storage.Service	"Service created"
//	@Failure		400		{object}	ErrorResponse	"Bad request"
//	@Failure		500		{object}	ErrorResponse	"Internal server error"
//	@Router			/services [post]
func (s *Server) handleAPICreateService(c *fiber.Ctx) error {
	var serviceDTO ServiceDTO
	if err := c.BodyParser(&serviceDTO); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body: " + err.Error(),
		})
	}

	// Validate required fields
	if serviceDTO.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Service name is required",
		})
	}
	if serviceDTO.Protocol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Protocol is required",
		})
	}
	if serviceDTO.Endpoint == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Endpoint is required",
		})
	}

	// Convert to storage.Service
	service := storage.Service{
		ID:        serviceDTO.ID,
		Name:      serviceDTO.Name,
		Protocol:  serviceDTO.Protocol,
		Endpoint:  serviceDTO.Endpoint,
		Interval:  serviceDTO.Interval,
		Timeout:   serviceDTO.Timeout,
		Retries:   serviceDTO.Retries,
		Tags:      serviceDTO.Tags,
		IsEnabled: serviceDTO.IsEnabled,
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
	config, err := s.convertFlatConfigToMonitorConfig(serviceDTO.Protocol, serviceDTO.Config)
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
//
//	@Summary		Update service
//	@Description	Updates an existing service
//	@Tags			services
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string			true	"Service ID"
//	@Param			service	body		ServiceDTO		true	"New service configuration"
//	@Success		200		{object}	storage.Service	"Service updated"
//	@Failure		400		{object}	ErrorResponse	"Bad request"
//	@Failure		404		{object}	ErrorResponse	"Service not found"
//	@Failure		500		{object}	ErrorResponse	"Internal server error"
//	@Router			/services/{id} [put]
func (s *Server) handleAPIUpdateService(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Service ID is required",
		})
	}

	var serviceDTO ServiceDTO
	if err := c.BodyParser(&serviceDTO); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body: " + err.Error(),
		})
	}

	// Debug: log the received data
	fmt.Printf("Update service request: %+v\n", serviceDTO)

	// Get existing service to preserve state
	existingService, err := s.monitorService.GetServiceByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Service not found: " + err.Error(),
		})
	}

	// Convert to storage.Service, preserving existing state
	service := storage.Service{
		ID:        id,
		Name:      serviceDTO.Name,
		Protocol:  serviceDTO.Protocol,
		Endpoint:  serviceDTO.Endpoint,
		Interval:  serviceDTO.Interval,
		Timeout:   serviceDTO.Timeout,
		Retries:   serviceDTO.Retries,
		Tags:      serviceDTO.Tags,
		State:     existingService.State,
		IsEnabled: serviceDTO.IsEnabled,
	}

	// Convert flat config to proper MonitorConfig structure
	config, err := s.convertFlatConfigToMonitorConfig(serviceDTO.Protocol, serviceDTO.Config)
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
//
//	@Summary		Delete service
//	@Description	Deletes a service from the monitoring system
//	@Tags			services
//	@Accept			json
//	@Produce		json
//	@Param			id	path	string	true	"Service ID"
//	@Success		204	"Service deleted"
//	@Failure		400	{object}	ErrorResponse	"Bad request"
//	@Failure		500	{object}	ErrorResponse	"Internal server error"
//	@Router			/services/{id} [delete]
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
//
//	@Summary		Get service configuration
//	@Description	Returns the complete service configuration by ID
//	@Tags			services
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string			true	"Service ID"
//	@Success		200	{object}	storage.Service	"Service configuration"
//	@Failure		400	{object}	ErrorResponse	"Bad request"
//	@Failure		404	{object}	ErrorResponse	"Service not found"
//	@Router			/services/config/{id} [get]
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

// convertFlatConfigToMonitorConfig converts JSON config object to proper MonitorConfig structure
func (s *Server) convertFlatConfigToMonitorConfig(protocol string, configObj map[string]any) (storage.MonitorConfig, error) {
	if configObj == nil {
		// Return default config based on protocol
		return s.getDefaultConfig(protocol), nil
	}

	// Convert interface{} to map[string]interface{}

	// Validate and convert based on protocol
	switch protocol {
	case "http", "https":
		return s.parseHTTPConfig(configObj)
	case "tcp":
		return s.parseTCPConfig(configObj)
	case "grpc":
		return s.parseGRPCConfig(configObj)
	case "redis":
		return s.parseRedisConfig(configObj)
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
				CheckType:   "connectivity",
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

	// Extract multi-endpoint configuration if present
	if multiEndpoint, ok := configMap["multi_endpoint"].(map[string]interface{}); ok {
		httpConfig.ExtendedConfig = multiEndpoint
	}

	// Extract extended_config if present (for backward compatibility)
	if extendedConfig, ok := configMap["extended_config"].(map[string]interface{}); ok {
		httpConfig.ExtendedConfig = extendedConfig
	}

	// Check for unknown fields
	allowedFields := map[string]bool{"method": true, "expected_status": true, "headers": true, "multi_endpoint": true, "extended_config": true}
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

	// Quick lookup map for incident stats by service ID
	statsMap := make(map[string]*storage.ServiceIncidentStats)
	for _, stats := range incidentStats {
		statsMap[stats.ServiceID] = stats
	}

	// Build DTOs
	serviceDTOs := make([]ServiceDTO, len(services))
	for i, service := range services {
		dto := ServiceDTO{
			ID:              service.ID,
			Name:            service.Name,
			Protocol:        service.Protocol,
			Endpoint:        service.Endpoint,
			Interval:        service.Interval,
			Timeout:         service.Timeout,
			Retries:         service.Retries,
			Tags:            service.Tags,
			Config:          s.convertConfigToMap(service.Config),
			State:           service.State,
			IsEnabled:       service.IsEnabled,
			ActiveIncidents: 0,
			TotalIncidents:  0,
		}
		if stats, exists := statsMap[service.ID]; exists {
			dto.ActiveIncidents = stats.ActiveIncidents
			dto.TotalIncidents = stats.TotalIncidents
		}
		serviceDTOs[i] = dto
	}

	update := fiber.Map{
		"type":      "service_update",
		"services":  serviceDTOs,
		"timestamp": time.Now().Unix(),
	}

	s.wsMutex.Lock()
	defer s.wsMutex.Unlock()

	return conn.WriteJSON(update)
}

// BroadcastServiceUpdate sends service updates to all connected WebSocket clients
func (s *Server) broadcastServiceUpdate() {
	services, err := s.monitorService.GetAllServiceConfigs(context.Background())
	if err != nil {
		return
	}

	incidentStats, err := s.monitorService.GetAllServicesIncidentStats(context.Background())
	if err != nil {
		return
	}

	statsMap := make(map[string]*storage.ServiceIncidentStats)
	for _, stats := range incidentStats {
		statsMap[stats.ServiceID] = stats
	}

	serviceDTOs := make([]ServiceDTO, len(services))
	for i, service := range services {
		dto := ServiceDTO{
			ID:              service.ID,
			Name:            service.Name,
			Protocol:        service.Protocol,
			Endpoint:        service.Endpoint,
			Interval:        service.Interval,
			Timeout:         service.Timeout,
			Retries:         service.Retries,
			Tags:            service.Tags,
			Config:          s.convertConfigToMap(service.Config),
			State:           service.State,
			IsEnabled:       service.IsEnabled,
			ActiveIncidents: 0,
			TotalIncidents:  0,
		}
		if stats, exists := statsMap[service.ID]; exists {
			dto.ActiveIncidents = stats.ActiveIncidents
			dto.TotalIncidents = stats.TotalIncidents
		}
		serviceDTOs[i] = dto
	}

	update := fiber.Map{
		"type":      "service_update",
		"services":  serviceDTOs,
		"timestamp": time.Now().Unix(),
	}

	s.wsMutex.Lock()
	defer s.wsMutex.Unlock()

	for conn := range s.wsConnections {
		if err := conn.WriteJSON(update); err != nil {
			delete(s.wsConnections, conn)
			conn.Close()
		}
	}
}

func (s *Server) subscribeEvents(ctx context.Context) error {
	broker := s.receiver.ServiceUpdated()
	sub := broker.Subscribe()
	defer broker.Unsubscribe(sub)

	errChan := make(chan error, 1)
	go func() {
		for ctx.Err() == nil {
			select {
			case <-sub:
				s.broadcastServiceUpdate()

			case <-ctx.Done():
				return
			}
		}
	}()

	select {
	case <-ctx.Done():
	case err := <-errChan:
		return err
	}

	return nil
}

// convertConfigToMap converts storage.MonitorConfig to map[string]any
func (s *Server) convertConfigToMap(cfg storage.MonitorConfig) map[string]any {
	result := make(map[string]any)

	if cfg.HTTP != nil {
		result["method"] = cfg.HTTP.Method
		result["expected_status"] = cfg.HTTP.ExpectedStatus
		result["headers"] = cfg.HTTP.Headers
		if cfg.HTTP.ExtendedConfig != nil {
			result["multi_endpoint"] = cfg.HTTP.ExtendedConfig
		}
	}

	if cfg.TCP != nil {
		result["send_data"] = cfg.TCP.SendData
		result["expect_data"] = cfg.TCP.ExpectData
	}

	if cfg.GRPC != nil {
		result["check_type"] = cfg.GRPC.CheckType
		result["service_name"] = cfg.GRPC.ServiceName
		result["tls"] = cfg.GRPC.TLS
		result["insecure_tls"] = cfg.GRPC.InsecureTLS
	}

	if cfg.Redis != nil {
		result["password"] = cfg.Redis.Password
		result["db"] = cfg.Redis.DB
	}

	return result
}
