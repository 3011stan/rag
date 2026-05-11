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
	logging.Init()

	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load configuration")
	}

	server, err := api.NewAPIServer(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize API server")
	}

	router := chi.NewRouter()

	router.Use(logging.RequestIDMiddleware)
	router.Use(logging.LoggingMiddleware)
	router.Use(server.CORSMiddleware)
	router.Use(api.SecurityHeadersMiddleware)
	router.Use(middleware.Recoverer)

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	router.Post("/rag/ingest", server.IngestHandler())
	router.Post("/rag/ask", server.AskHandler())
	router.Post("/auth/temp-token", server.TemporaryTokenHandler())
	router.Get("/documents", server.ListDocumentsHandler())
	router.Delete("/documents/{id}", server.DeleteDocumentHandler())
	router.Get("/debug/metadata", server.DebugMetadataHandler())
	router.Post("/admin/seed-demo", server.SeedDemoHandler())

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
