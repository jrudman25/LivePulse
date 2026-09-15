package api

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// parseOriginList parses a comma-separated origin allowlist.
func parseOriginList(raw string) map[string]bool {
	origins := make(map[string]bool)
	for _, o := range strings.Split(raw, ",") {
		if trimmed := strings.TrimSpace(o); trimmed != "" {
			origins[trimmed] = true
		}
	}
	return origins
}

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Call the next handler
		next(w, r)

		// Log the request
		log.Printf("%s %s %s", r.Method, r.RequestURI, time.Since(start))
	}
}

// CORSMiddleware adds CORS headers. When CORS_ALLOWED_ORIGINS (comma-separated)
// is set, only those origins receive allow headers; otherwise it falls back to
// Allow-Origin: * for local development.
func CORSMiddleware(next http.HandlerFunc) http.HandlerFunc {
	allowed := parseOriginList(os.Getenv("CORS_ALLOWED_ORIGINS"))

	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if len(allowed) == 0 {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else if allowed[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Add("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	}
}

// RecoveryMiddleware recovers from panics
func RecoveryMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()
		next(w, r)
	}
}

// Chain combines multiple middleware
func Chain(handler http.HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}
