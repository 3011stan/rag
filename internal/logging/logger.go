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

// Init configura o zerolog com as configurações padrão
func Init() {
	// Configurar saída console com cores e timestamps
	log.Logger = zerolog.New(
		zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		},
	).With().Timestamp().Caller().Logger()

	// Definir nível de log baseado na variável de ambiente
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

// FromContext retorna o logger do contexto ou o logger global
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

// ContextWithRequestID adiciona um request_id ao contexto
func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, contextKeyReqID, requestID)
}

// RequestIDFromContext extrai o request_id do contexto
func RequestIDFromContext(ctx context.Context) string {
	reqID, ok := ctx.Value(contextKeyReqID).(string)
	if !ok {
		return ""
	}
	return reqID
}
