package main

import (
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"github.com/stan/Projects/studies/rag/internal/api"
	"github.com/stan/Projects/studies/rag/internal/config"
	"github.com/stan/Projects/studies/rag/internal/logging"
)

func main() {
	// Inicializar logging
	logging.Init()

	// Carregar variáveis de ambiente do arquivo .env
	_ = godotenv.Load()

	// Carregar configuração
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load configuration")
	}

	// Criar API Server com todas as dependências
	server, err := api.NewAPIServer(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize API server")
	}

	// Criar roteador
	router := chi.NewRouter()

	// Middleware
	router.Use(logging.RequestIDMiddleware)
	router.Use(logging.LoggingMiddleware)
	router.Use(api.SecurityHeadersMiddleware)
	router.Use(middleware.Recoverer)

	// Health check
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Registrar endpoints com handlers completos
	router.Post("/rag/ingest", server.IngestHandler())
	router.Post("/rag/ask", server.AskHandler())
	router.Post("/admin/seed-demo", server.SeedDemoHandler())

	// Iniciar servidor
	log.Info().
		Str("port", cfg.Port).
		Msg("starting API server")

	httpServer := &http.Server{
		Addr:              cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      2 * time.Minute,
		IdleTimeout:       60 * time.Second,
	}

	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatal().Err(err).Msg("failed to start server")
		os.Exit(1)
	}
}
