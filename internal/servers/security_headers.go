package servers

import "net/http"

// withSecurityHeaders adds security-related HTTP headers to all responses.
func withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()

		// Prevent MIME type sniffing
		h.Set("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking
		h.Set("X-Frame-Options", "DENY")

		// Enable XSS filter in older browsers
		h.Set("X-XSS-Protection", "1; mode=block")

		// Prevent referrer leakage
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy — restrict resource loading to same origin.
		// 'unsafe-inline' is required for Vite-bundled styles.
		h.Set("Content-Security-Policy",
			"default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'",
		)

		next.ServeHTTP(w, r)
	})
}
