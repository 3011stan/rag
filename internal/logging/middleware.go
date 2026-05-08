package logging

import (
	"net/http"
	"time"

	"github.com/google/uuid"
)

// RequestIDMiddleware adiciona um request_id único a cada request
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Adicionar request_id ao contexto
		ctx := ContextWithRequestID(r.Context(), requestID)

		// Adicionar header na resposta
		w.Header().Set(RequestIDHeader, requestID)

		// Chamar próximo handler com contexto atualizado
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LoggingMiddleware faz logging de requests e responses
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := FromContext(r.Context())

		start := time.Now()

		// Wrapper para capturar status code
		statusWriter := &statusWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Log do request
		logger.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote_addr", r.RemoteAddr).
			Str("user_agent", r.Header.Get("User-Agent")).
			Msg("incoming request")

		// Executar handler
		next.ServeHTTP(statusWriter, r)

		// Log da resposta
		duration := time.Since(start)
		logger.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", statusWriter.statusCode).
			Dur("duration_ms", duration).
			Msg("request completed")
	})
}

// statusWriter wrapper para capturar o status code da resposta
type statusWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}
