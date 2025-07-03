package web

import (
	"embed"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/template/html/v2"
	"github.com/sxwebdev/sentinel/internal/config"
	"github.com/sxwebdev/sentinel/internal/service"
)

//go:embed views/*
var viewsFS embed.FS

// Server represents the web server
type Server struct {
	monitorService *service.MonitorService
	config         *config.Config
	app            *fiber.App
}

// NewServer creates a new web server
func NewServer(monitorService *service.MonitorService, cfg *config.Config) (*Server, error) {
	// Create template engine with embedded templates
	templateEngine := html.NewFileSystem(http.FS(viewsFS), ".html")

	// Add template functions
	templateEngine.
		AddFunc("lower", strings.ToLower).
		AddFunc("statusToString", func(status config.ServiceStatus) string {
			return status.String()
		}).
		AddFunc("statusToLower", func(status config.ServiceStatus) string {
			return strings.ToLower(status.String())
		}).
		AddFunc("formatDateTime", func(t time.Time) string {
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
	}

	// Setup routes
	server.setupRoutes()

	return server, nil
}

// setupRoutes configures all routes
func (s *Server) setupRoutes() {
	// Web UI routes
	s.app.Get("/", s.handleDashboard)
	s.app.Get("/service/:name", s.handleServiceDetail)

	// API routes
	api := s.app.Group("/api")
	api.Get("/services", s.handleAPIServices)
	api.Get("/services/:name", s.handleAPIServiceDetail)
	api.Get("/services/:name/incidents", s.handleAPIServiceIncidents)
	api.Get("/services/:name/stats", s.handleAPIServiceStats)
	api.Post("/services/:name/check", s.handleAPIServiceCheck)
	api.Post("/services/:name/resolve", s.handleAPIServiceResolve)
	api.Get("/incidents", s.handleAPIRecentIncidents)
}

// App returns the Fiber app instance
func (s *Server) App() *fiber.App {
	return s.app
}

// handleDashboard renders the main dashboard
func (s *Server) handleDashboard(c *fiber.Ctx) error {
	services, err := s.monitorService.GetAllServices()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.Render("views/dashboard", fiber.Map{
		"Services": services,
		"Title":    "Sentinel Dashboard",
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
	serviceName := c.Params("name")
	if serviceName == "" {
		return c.Status(fiber.StatusBadRequest).SendString("service name is required")
	}

	// URL decode the service name
	decodedName, err := url.QueryUnescape(serviceName)
	if err != nil {
		decodedName = serviceName // fallback to original if decoding fails
	}

	// Get service state
	state, err := s.monitorService.GetServiceState(decodedName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("service not found: " + decodedName)
	}

	// Get recent incidents
	incidents, err := s.monitorService.GetServiceIncidents(c.Context(), decodedName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	// Get service stats
	stats, err := s.monitorService.GetServiceStats(c.Context(), decodedName, time.Now().AddDate(0, 0, -30))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.Render("views/service_detail", fiber.Map{
		"Service":      state,
		"Incidents":    incidents,
		"Stats":        stats,
		"Title":        "Service: " + decodedName,
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
	services, err := s.monitorService.GetAllServices()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(services)
}

// handleAPIServiceDetail returns service details
func (s *Server) handleAPIServiceDetail(c *fiber.Ctx) error {
	serviceName := c.Params("name")
	if serviceName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "service name is required",
		})
	}

	// URL decode the service name
	decodedName, err := url.QueryUnescape(serviceName)
	if err != nil {
		decodedName = serviceName // fallback to original if decoding fails
	}

	state, err := s.monitorService.GetServiceState(decodedName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "service not found: " + decodedName,
		})
	}

	return c.JSON(state)
}

// handleAPIServiceIncidents returns service incidents
func (s *Server) handleAPIServiceIncidents(c *fiber.Ctx) error {
	serviceName := c.Params("name")
	if serviceName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "service name is required",
		})
	}

	// URL decode the service name
	decodedName, err := url.QueryUnescape(serviceName)
	if err != nil {
		decodedName = serviceName // fallback to original if decoding fails
	}

	incidents, err := s.monitorService.GetServiceIncidents(c.Context(), decodedName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(incidents)
}

// handleAPIServiceStats returns service statistics
func (s *Server) handleAPIServiceStats(c *fiber.Ctx) error {
	serviceName := c.Params("name")
	if serviceName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "service name is required",
		})
	}

	// URL decode the service name
	decodedName, err := url.QueryUnescape(serviceName)
	if err != nil {
		decodedName = serviceName // fallback to original if decoding fails
	}

	daysStr := c.Query("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil {
		days = 30
	}

	since := time.Now().AddDate(0, 0, -days)
	stats, err := s.monitorService.GetServiceStats(c.Context(), decodedName, since)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(stats)
}

// handleAPIServiceCheck triggers a manual check
func (s *Server) handleAPIServiceCheck(c *fiber.Ctx) error {
	serviceName := c.Params("name")
	if serviceName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "service name is required",
		})
	}

	// URL decode the service name
	decodedName, err := url.QueryUnescape(serviceName)
	if err != nil {
		decodedName = serviceName // fallback to original if decoding fails
	}

	err = s.monitorService.TriggerCheck(c.Context(), decodedName)
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
	serviceName := c.Params("name")
	if serviceName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "service name is required",
		})
	}

	// URL decode the service name
	decodedName, err := url.QueryUnescape(serviceName)
	if err != nil {
		decodedName = serviceName // fallback to original if decoding fails
	}

	err = s.monitorService.ForceResolveIncidents(c.Context(), decodedName)
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
