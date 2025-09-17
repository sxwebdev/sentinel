package handlerutils

import (
	"net/http"

	connectcors "connectrpc.com/cors"
	"github.com/rs/cors"
)

func WithCORS(connectHandler http.Handler) http.Handler {
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"}, // Allow all origins
		AllowedMethods: connectcors.AllowedMethods(),
		AllowedHeaders: connectcors.AllowedHeaders(),
		ExposedHeaders: connectcors.ExposedHeaders(),
		MaxAge:         7200, // 2 hours in seconds,
	})
	return c.Handler(connectHandler)
}
