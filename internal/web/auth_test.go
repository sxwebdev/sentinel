package web

import (
	"encoding/base64"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/sxwebdev/sentinel/internal/config"
)

func TestBasicAuth(t *testing.T) {
	// Create test config with auth enabled
	cfg := &config.Config{
		Server: config.ServerConfig{
			Auth: config.AuthConfig{
				Enabled: true,
				Users: []config.UserAuth{
					{Username: "admin", Password: "secret"},
					{Username: "viewer", Password: "readonly"},
				},
			},
		},
	}

	app := fiber.New()

	// Create server instance
	server := &Server{
		config: cfg,
		app:    app,
	}

	// Apply auth middleware
	app.Use(server.createBasicAuthMiddleware())

	// Add test route
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	tests := []struct {
		name           string
		username       string
		password       string
		expectedStatus int
	}{
		{
			name:           "Valid admin credentials",
			username:       "admin",
			password:       "secret",
			expectedStatus: 200,
		},
		{
			name:           "Valid viewer credentials",
			username:       "viewer",
			password:       "readonly",
			expectedStatus: 200,
		},
		{
			name:           "Invalid username",
			username:       "invalid",
			password:       "secret",
			expectedStatus: 401,
		},
		{
			name:           "Invalid password",
			username:       "admin",
			password:       "wrong",
			expectedStatus: 401,
		},
		{
			name:           "No credentials",
			username:       "",
			password:       "",
			expectedStatus: 401,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)

			if tt.username != "" && tt.password != "" {
				credentials := base64.StdEncoding.EncodeToString([]byte(tt.username + ":" + tt.password))
				req.Header.Set("Authorization", "Basic "+credentials)
			}

			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestWebSocketBypass(t *testing.T) {
	// Create test config with auth enabled
	cfg := &config.Config{
		Server: config.ServerConfig{
			Auth: config.AuthConfig{
				Enabled: true,
				Users: []config.UserAuth{
					{Username: "admin", Password: "secret"},
				},
			},
		},
	}

	app := fiber.New()

	// Create server instance
	server := &Server{
		config: cfg,
		app:    app,
	}

	// Apply auth middleware
	app.Use(server.createBasicAuthMiddleware())

	// Add test WebSocket route
	app.Get("/ws", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "websocket endpoint"})
	})

	// Test WebSocket upgrade request (should bypass auth)
	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
	req.Header.Set("Sec-WebSocket-Version", "13")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}
