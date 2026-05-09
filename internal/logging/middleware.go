package logging

import (
	"net/http"
	"time"

	"github.com/google/uuid"
)

// RequestIDMiddleware attaches a request ID to the request context and response.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		ctx := ContextWithRequestID(r.Context(), requestID)
		w.Header().Set(RequestIDHeader, requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LoggingMiddleware logs request metadata and response status.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := FromContext(r.Context())

		start := time.Now()

		statusWriter := &statusWriter{ResponseWriter: w, statusCode: http.StatusOK}

		logger.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote_addr", r.RemoteAddr).
			Str("user_agent", r.Header.Get("User-Agent")).
			Msg("incoming request")

		next.ServeHTTP(statusWriter, r)

		duration := time.Since(start)
		logger.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", statusWriter.statusCode).
			Dur("duration_ms", duration).
			Msg("request completed")
	})
}

type statusWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}
