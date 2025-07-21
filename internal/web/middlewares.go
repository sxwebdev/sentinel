package web

import (
	"strings"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
)

// createBasicAuthMiddleware creates and configures basic authentication middleware
func (s *Server) createBasicAuthMiddleware() fiber.Handler {
	// Create users map from config
	users := make(map[string]string)
	for _, user := range s.config.Server.Auth.Users {
		users[user.Username] = user.Password
	}

	return func(c *fiber.Ctx) error {
		// Skip auth for WebSocket upgrade requests
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}

		// Skip auth for health endpoints (optional)
		if strings.HasPrefix(c.Path(), "/health") {
			return c.Next()
		}

		// Apply basic auth
		return basicauth.New(basicauth.Config{
			Users: users,
			Realm: "Sentinel Monitoring",
			Authorizer: func(user, pass string) bool {
				// Check if user exists and password matches
				if storedPass, exists := users[user]; exists {
					return storedPass == pass
				}
				return false
			},
		})(c)
	}
}
