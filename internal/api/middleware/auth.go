package middleware

import (
	"net/http"

	"github.com/arkishshah/go-infra-provisioner/pkg/logger"
)

// Logging middleware to log HTTP requests
func Logging(logger *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info("Request:", r.Method, r.URL.Path)
			next.ServeHTTP(w, r)
		})
	}
}

// Auth middleware for authentication (if needed later)
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Add authentication logic here
		// For now, just pass through
		next.ServeHTTP(w, r)
	})
}

// CORS middleware if needed
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
