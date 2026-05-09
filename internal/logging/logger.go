package logging

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	RequestIDHeader = "X-Request-ID"
	contextKeyReqID = "request_id"
)

// Init configures the process logger.
func Init() {
	log.Logger = zerolog.New(
		zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		},
	).With().Timestamp().Caller().Logger()

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
}

// FromContext returns a request-scoped logger when a request ID is present.
func FromContext(ctx context.Context) *zerolog.Logger {
	if ctx == nil {
		return &log.Logger
	}

	reqID, ok := ctx.Value(contextKeyReqID).(string)
	if !ok {
		return &log.Logger
	}

	logger := log.With().Str("request_id", reqID).Logger()
	return &logger
}

// ContextWithRequestID stores the request ID in the context.
func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, contextKeyReqID, requestID)
}

// RequestIDFromContext returns the request ID from the context.
func RequestIDFromContext(ctx context.Context) string {
	reqID, ok := ctx.Value(contextKeyReqID).(string)
	if !ok {
		return ""
	}
	return reqID
}
