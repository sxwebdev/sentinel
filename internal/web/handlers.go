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
	goHTML "html"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
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
	"github.com/sxwebdev/sentinel/internal/utils"
	"github.com/sxwebdev/sentinel/pkg/dbutils"
	_ "github.com/sxwebdev/sentinel/pkg/dbutils"
)

//go:embed views/*
var viewsFS embed.FS

// Server represents the web server
type Server struct {
	monitorService *service.MonitorService
	storage        storage.Storage
	receiver       *receiver.Receiver
	config         *config.Config
	app            *fiber.App
	wsConnections  map[*websocket.Conn]bool
	wsMutex        sync.Mutex
	validator      *validator.Validate
}

// NewServer creates a new web server
func NewServer(
	cfg *config.Config,
	monitorService *service.MonitorService,
	storage storage.Storage,
	receiver *receiver.Receiver,
) (*Server, error) {
	// Create template engine with embedded templates
	templateEngine := html.NewFileSystem(http.FS(viewsFS), ".html")

	// Add template functions
	templateEngine.
		AddFunc("lower", strings.ToLower).
		AddFunc("statusToString", func(status string) string {
			return string(status)
		}).
		AddFunc("statusToLower", func(status string) string {
			return strings.ToLower(status)
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
		Views:   templateEngine,
		AppName: "Sentinel",
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
		storage:        storage,
		receiver:       receiver,
		config:         cfg,
		app:            app,
		wsConnections:  make(map[*websocket.Conn]bool),
		validator:      validator.New(),
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
	fmt.Println("Web server: stopping, closing all WebSocket connections")

	s.wsMutex.Lock()
	// Закрываем все WebSocket соединения
	for conn := range s.wsConnections {
		conn.Close()
	}
	// Очищаем map
	s.wsConnections = make(map[*websocket.Conn]bool)
	s.wsMutex.Unlock()

	fmt.Println("Web server: stopped")
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

	// Dashboard API
	api.Get("/dashboard/stats", s.handleAPIDashboardStats)

	// Service management API
	api.Get("/services", s.handleFindServices)
	api.Post("/services", s.handleAPICreateService)
	api.Put("/services/:id", s.handleAPIUpdateService)
	api.Delete("/services/:id", s.handleAPIDeleteService)
	api.Post("/services/:id/check", s.handleAPIServiceCheck)
	api.Post("/services/:id/resolve", s.handleAPIServiceResolve)

	// Service detail API
	api.Get("/services/:id", s.handleAPIServiceDetail)
	api.Get("/services/:id/stats", s.handleAPIServiceStats)

	// Incident management API
	api.Get("/incidents", s.handleFindIncidents)
	api.Get("/services/:id/incidents", s.handleAPIServiceIncidents)
	api.Delete("/services/:id/incidents/:incidentId", s.handleAPIDeleteIncident)

	// Tags API
	api.Get("/tags", s.handleGetAllTags)
	api.Get("/tags/count", s.handleGetAllTagsWithCount)

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

	return c.Render("views/service_detail", fiber.Map{
		"Service":      targetService,
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

// handleFindServices handles GET /api/v1/services
//
//	@Summary		Get all services
//	@Description	Returns a list of all services with their current states
//	@Tags			services
//	@Accept			json
//	@Produce		json
//	@Param			name		query		string										false	"Filter by service name"
//	@Param			tags		query		[]string									false	"Filter by service tags"
//	@Param			status		query		string										false	"Filter by service status"	ENUM("up", "down")
//	@Param			is_enabled	query		bool										false	"Filter by enabled status"
//	@Param			protocol	query		string										false	"Filter by protocol"	ENUM("http", "tcp", "grpc")
//	@Param			order_by	query		string										false	"Order by field"		ENUM("name", "created_at")
//	@Param			page		query		uint32										false	"Page number (for pagination)"
//	@Param			page_size	query		uint32										false	"Number of items per page (default 20)"
//	@Success		200			{array}		dbutils.FindResponseWithCount[ServiceDTO]	"List of services with states"
//	@Failure		500			{object}	ErrorResponse								"Internal server error"
//	@Router			/services [get]
func (s *Server) handleFindServices(c *fiber.Ctx) error {
	ctx := c.Context()

	// Parse query parameters
	params := struct {
		Name      string   `json:"name" query:"name"`
		Tags      []string `json:"tags" query:"tags"`
		Status    string   `json:"status" query:"status" validate:"omitempty,oneof=up down"`
		IsEnabled *bool    `json:"is_enabled" query:"is_enabled"`
		Protocol  string   `json:"protocol" query:"protocol" validate:"omitempty,oneof=http tcp grpc"`
		OrderBy   string   `json:"order_by" query:"order_by" validate:"omitempty,oneof=name created_at"`
		Page      *uint32  `json:"page" query:"page"`
		PageSize  *uint32  `json:"page_size" query:"page_size"`
	}{}
	if err := c.QueryParser(&params); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "Invalid query parameters",
		})
	}

	// Validate query parameters
	if err := s.validator.Struct(params); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "Invalid query parameters",
		})
	}

	services, err := s.monitorService.FindServices(ctx, storage.FindServicesParams{
		Name:      params.Name,
		Tags:      params.Tags,
		Status:    params.Status,
		IsEnabled: params.IsEnabled,
		Protocol:  params.Protocol,
		OrderBy:   params.OrderBy,
		Page:      params.Page,
		PageSize:  params.PageSize,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	result := dbutils.FindResponseWithCount[ServiceDTO]{
		Items: make([]ServiceDTO, 0, len(services.Items)),
		Count: services.Count,
	}

	// Get services with their states
	for _, service := range services.Items {
		svc, err := convertServiceToDTO(service)
		if err != nil {
			return err
		}

		result.Items = append(result.Items, svc)
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
//	@Success		200	{object}	ServiceDTO		"Service details with state"
//	@Failure		400	{object}	ErrorResponse	"Bad request"
//	@Failure		404	{object}	ErrorResponse	"Service not found"
//	@Router			/services/{id} [get]
func (s *Server) handleAPIServiceDetail(c *fiber.Ctx) error {
	serviceID := c.Params("id")
	if serviceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "service ID is required",
		})
	}

	targetService, err := s.monitorService.GetServiceByID(c.Context(), serviceID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	// Get service with state
	svcDTO, err := convertServiceToDTO(targetService)
	if err != nil {
		return err
	}

	if svcDTO.LastError != nil && *svcDTO.LastError != "" {
		svcDTO.LastError = utils.Pointer(goHTML.EscapeString(*svcDTO.LastError))
	}

	return c.JSON(svcDTO)
}

// handleAPIServiceIncidents returns service incidents
//
//	@Summary		Get service incidents
//	@Description	Returns a list of incidents for a specific service
//	@Tags			incidents
//	@Accept			json
//	@Produce		json
//	@Param			id			path		string									true	"Service ID"
//	@Param			incident_id	query		string									false	"Filter by incident ID"
//	@Param			resolved	query		bool									false	"Filter by resolved status"
//	@Param			start_time	query		time.Time								false	"Filter by start time (RFC3339 format)"
//	@Param			end_time	query		time.Time								false	"Filter by end time (RFC3339 format)"
//	@Param			page		query		uint32									false	"Page number (for pagination)"
//	@Param			page_size	query		uint32									false	"Number of items per page (default 20)"
//	@Success		200			{array}		dbutils.FindResponseWithCount[Incident]	"List of incidents"
//	@Failure		400			{object}	ErrorResponse							"Bad request"
//	@Failure		500			{object}	ErrorResponse							"Internal server error"
//	@Router			/services/{id}/incidents [get]
func (s *Server) handleAPIServiceIncidents(c *fiber.Ctx) error {
	serviceID := c.Params("id")
	if serviceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "service ID is required",
		})
	}

	params := struct {
		IncidentID string     `query:"incident_id"`
		Resolved   *bool      `query:"resolved"`
		StartTime  *time.Time `query:"start_time"`
		EndTime    *time.Time `query:"end_time"`
		Page       *uint32    `query:"page" validate:"omitempty,gte=1"`
		PageSize   *uint32    `query:"page_size" validate:"omitempty,gte=1,lte=100"`
	}{}

	if err := c.QueryParser(&params); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "Invalid query parameters",
		})
	}

	// Validate query parameters
	if err := s.validator.Struct(params); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "Invalid query parameters",
		})
	}

	incidents, err := s.storage.FindIncidents(c.Context(), storage.FindIncidentsParams{
		ID:        params.IncidentID,
		ServiceID: serviceID,
		Resolved:  params.Resolved,
		StartTime: params.StartTime,
		EndTime:   params.EndTime,
		Page:      params.Page,
		PageSize:  params.PageSize,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	for _, incident := range incidents.Items {
		incident.Error = goHTML.EscapeString(incident.Error)
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
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "service ID is required",
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
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: err.Error(),
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
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "service ID is required",
		})
	}

	// First check if service exists
	_, err := s.monitorService.GetServiceByID(c.Context(), serviceID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Error: "service not found: " + serviceID,
		})
	}

	err = s.monitorService.TriggerCheck(c.Context(), serviceID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	return c.JSON(SuccessResponse{
		Message: "check triggered successfully",
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
//	@Failure		500	{object}	ErrorResponse	"Internal server error"
//	@Router			/services/{id}/resolve [post]
func (s *Server) handleAPIServiceResolve(c *fiber.Ctx) error {
	serviceID := c.Params("id")
	if serviceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "service ID is required",
		})
	}

	err := s.monitorService.ForceResolveIncidents(c.Context(), serviceID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	return c.JSON(SuccessResponse{
		Message: "incident resolved successfully",
	})
}

// handleFindIncidents returns recent incidents
//
//	@Summary		Get recent incidents
//	@Description	Returns a list of recent incidents across all services
//	@Tags			incidents
//	@Accept			json
//	@Produce		json
//	@Param			search		query		string									false	"Filter by service ID or incident ID"
//	@Param			resolved	query		bool									false	"Filter by resolved status"
//	@Param			start_time	query		time.Time								false	"Start time for filtering (RFC3339 format)"
//	@Param			end_time	query		time.Time								false	"End time for filtering (RFC3339 format)"
//	@Param			page		query		uint32									false	"Page number (default 1)"
//	@Param			page_size	query		uint32									false	"Number of items per page (default 100)"
//	@Success		200			{array}		dbutils.FindResponseWithCount[Incident]	"List of incidents"
//	@Failure		500			{object}	ErrorResponse							"Internal server error"
//	@Router			/incidents [get]
func (s *Server) handleFindIncidents(c *fiber.Ctx) error {
	params := struct {
		Search    string     `query:"search"`
		Resolved  *bool      `query:"resolved"`
		StartTime *time.Time `query:"start_time"`
		EndTime   *time.Time `query:"end_time"`
		Page      *uint32    `query:"page" validate:"omitempty,gte=1"`
		PageSize  *uint32    `query:"page_size" validate:"omitempty,gte=1,lte=100"`
	}{}

	if err := c.QueryParser(&params); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "Invalid query parameters",
		})
	}

	// Validate query parameters
	if err := s.validator.Struct(params); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "Invalid query parameters",
		})
	}

	incidents, err := s.storage.FindIncidents(c.Context(), storage.FindIncidentsParams{
		Search:    params.Search,
		Resolved:  params.Resolved,
		StartTime: params.StartTime,
		EndTime:   params.EndTime,
		Page:      params.Page,
		PageSize:  params.PageSize,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	for _, incident := range incidents.Items {
		incident.Error = goHTML.EscapeString(incident.Error)
	}

	return c.JSON(incidents)
}

// handleAPIDeleteIncident deletes a specific incident
//
//	@Summary		Delete incident
//	@Description	Deletes a specific incident for a service
//	@Tags			incidents
//	@Accept			json
//	@Produce		json
//	@Param			id			path	string	true	"Service ID"
//	@Param			incidentId	path	string	true	"Incident ID"
//	@Success		204			"Incident deleted"
//	@Failure		400			{object}	ErrorResponse	"Bad request"
//	@Failure		404			{object}	ErrorResponse	"Incident not found"
//	@Failure		500			{object}	ErrorResponse	"Internal server error"
//	@Router			/services/{id}/incidents/{incidentId} [delete]
func (s *Server) handleAPIDeleteIncident(c *fiber.Ctx) error {
	serviceID := c.Params("id")
	incidentID := c.Params("incidentId")

	if serviceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "service ID is required",
		})
	}

	if incidentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "incident ID is required",
		})
	}

	// Check if service exists
	_, err := s.monitorService.GetServiceByID(c.Context(), serviceID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Error: "service not found: " + serviceID,
		})
	}

	// Delete incident
	err = s.monitorService.DeleteIncident(c.Context(), serviceID, incidentID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// handleAPICreateService creates a new service
//
//	@Summary		Create new service
//	@Description	Creates a new service for monitoring
//	@Tags			services
//	@Accept			json
//	@Produce		json
//	@Param			service	body		CreateUpdateServiceRequest	true	"Service configuration"
//	@Success		201		{object}	ServiceDTO					"Service created"
//	@Failure		400		{object}	ErrorResponse				"Bad request"
//	@Failure		500		{object}	ErrorResponse				"Internal server error"
//	@Router			/services [post]
func (s *Server) handleAPICreateService(c *fiber.Ctx) error {
	var serviceDTO CreateUpdateServiceRequest
	if err := c.BodyParser(&serviceDTO); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "Invalid request body: " + err.Error(),
		})
	}

	// Validate required fields
	if serviceDTO.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "Service name is required",
		})
	}
	if serviceDTO.Protocol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "Protocol is required",
		})
	}

	// Convert to storage.Service
	createParams := storage.CreateUpdateServiceRequest{
		Name:      serviceDTO.Name,
		Protocol:  serviceDTO.Protocol,
		Interval:  time.Millisecond * time.Duration(serviceDTO.Interval),
		Timeout:   time.Millisecond * time.Duration(serviceDTO.Timeout),
		Retries:   serviceDTO.Retries,
		Tags:      serviceDTO.Tags,
		IsEnabled: serviceDTO.IsEnabled,
	}

	// Set default values
	if createParams.Interval == 0 {
		createParams.Interval = s.config.Monitoring.Global.DefaultInterval
	}
	if createParams.Timeout == 0 {
		createParams.Timeout = s.config.Monitoring.Global.DefaultTimeout
	}
	if createParams.Retries == 0 {
		createParams.Retries = s.config.Monitoring.Global.DefaultRetries
	}

	// Convert flat config to proper MonitorConfig structure
	if err := serviceDTO.Config.Validate(serviceDTO.Protocol); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "Invalid config: " + err.Error(),
		})
	}

	createParams.Config = serviceDTO.Config.ConvertToMap()

	// Add service
	svc, err := s.monitorService.CreateService(c.Context(), createParams)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	res, err := convertServiceToDTO(svc)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(res)
}

// handleAPIUpdateService updates an existing service
//
//	@Summary		Update service
//	@Description	Updates an existing service
//	@Tags			services
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string						true	"Service ID"
//	@Param			service	body		CreateUpdateServiceRequest	true	"New service configuration"
//	@Success		200		{object}	ServiceDTO					"Service updated"
//	@Failure		400		{object}	ErrorResponse				"Bad request"
//	@Failure		404		{object}	ErrorResponse				"Service not found"
//	@Failure		500		{object}	ErrorResponse				"Internal server error"
//	@Router			/services/{id} [put]
func (s *Server) handleAPIUpdateService(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "Service ID is required",
		})
	}

	var serviceDTO CreateUpdateServiceRequest
	if err := c.BodyParser(&serviceDTO); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "Invalid request body: " + err.Error(),
		})
	}

	// Debug: log the received data
	fmt.Printf("Update service request: %+v\n", serviceDTO)

	// Convert to storage.Service
	updateParams := storage.CreateUpdateServiceRequest{
		Name:      serviceDTO.Name,
		Protocol:  serviceDTO.Protocol,
		Interval:  time.Millisecond * time.Duration(serviceDTO.Interval),
		Timeout:   time.Millisecond * time.Duration(serviceDTO.Timeout),
		Retries:   serviceDTO.Retries,
		Tags:      serviceDTO.Tags,
		IsEnabled: serviceDTO.IsEnabled,
	}

	// Convert flat config to proper MonitorConfig structure
	if err := serviceDTO.Config.Validate(serviceDTO.Protocol); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "Invalid config: " + err.Error(),
		})
	}

	updateParams.Config = serviceDTO.Config.ConvertToMap()

	// Validate required fields
	if updateParams.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "Service name is required",
		})
	}
	if updateParams.Protocol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "Protocol is required",
		})
	}

	// Set default values if not provided
	if updateParams.Interval == 0 {
		updateParams.Interval = s.config.Monitoring.Global.DefaultInterval
	}
	if updateParams.Timeout == 0 {
		updateParams.Timeout = s.config.Monitoring.Global.DefaultTimeout
	}
	if updateParams.Retries == 0 {
		updateParams.Retries = s.config.Monitoring.Global.DefaultRetries
	}

	// Update service
	svc, err := s.monitorService.UpdateService(c.Context(), id, updateParams)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	res, err := convertServiceToDTO(svc)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	return c.JSON(res)
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
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "Service ID is required",
		})
	}

	if err := s.monitorService.DeleteService(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
