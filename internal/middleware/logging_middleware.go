package middleware

import (
	"net/http"
	"time"

	"go-banking-service/internal/logger"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		duration := time.Since(start)
		logger.Log.WithFields(map[string]any{
			"method":   r.Method,
			"path":     r.URL.Path,
			"duration": duration.String(),
			"remote":   r.RemoteAddr,
		}).Info("Request processed")
	})
}
