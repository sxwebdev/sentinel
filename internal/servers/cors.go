package servers

import (
	"net/http"

	connectcors "connectrpc.com/cors"
	"github.com/rs/cors"
)

func withCORS(connectHandler http.Handler) http.Handler {
	exposedHeaders := connectcors.ExposedHeaders()
	exposedHeaders = append(exposedHeaders, "Rpc-Error-Code")

	allowedHeaders := connectcors.AllowedHeaders()
	allowedHeaders = append(allowedHeaders, "Authorization", "Origin", "Access-Control-Allow-Origin", "Accept", "Options", "X-Project-ID")

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"}, // Allow all origins
		AllowedMethods: connectcors.AllowedMethods(),
		AllowedHeaders: allowedHeaders,
		ExposedHeaders: exposedHeaders,
		MaxAge:         7200, // 2 hours in seconds,
	})

	return c.Handler(connectHandler)
}
