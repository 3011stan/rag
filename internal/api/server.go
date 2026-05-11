package api

import (
	"context"
	"fmt"
	"time"

	"github.com/stan/Projects/studies/rag/internal/ai"
	"github.com/stan/Projects/studies/rag/internal/config"
	"github.com/stan/Projects/studies/rag/internal/rag"
	"github.com/stan/Projects/studies/rag/internal/rag/chunker"
	"github.com/stan/Projects/studies/rag/internal/rag/embeddings"
	"github.com/stan/Projects/studies/rag/internal/rag/loader"
	ragpipeline "github.com/stan/Projects/studies/rag/internal/rag/pipeline"
	"github.com/stan/Projects/studies/rag/internal/rag/retriever"
)

const (
	contentTypeJSON   = "application/json"
	contentTypeHeader = "Content-Type"
	maxQuestionLength = 5000
	maxAskBodySize    = 16 << 10
	maxTopK           = 20
	ingestTimeout     = 5 * time.Minute
	askTimeout        = 2 * time.Minute
	seedTimeout       = 10 * time.Minute
	authTimeout       = 10 * time.Second
	documentsTimeout  = 30 * time.Second
)

type APIServer struct {
	cfg      *config.Config
	vs       rag.VectorStore
	embedder embeddings.Provider
	chunker  *chunker.Chunker
	pipeline *ragpipeline.Pipeline
}

func NewAPIServer(cfg *config.Config) (*APIServer, error) {
	applyDefaults(cfg)

	vs, err := rag.NewPGVectorStore(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create vector store: %w", err)
	}
	if err := vs.EnsureSchema(context.Background(), cfg.EmbeddingDimensions); err != nil {
		return nil, fmt.Errorf("failed to ensure vector store schema: %w", err)
	}

	embedder, err := ai.NewEmbeddingProvider(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding provider: %w", err)
	}
	answerer, err := ai.NewAnsweringService(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create QA service: %w", err)
	}

	chunker, err := chunker.NewChunker(chunker.ChunkerConfig{
		ChunkTokens:   cfg.ChunkTokens,
		OverlapTokens: cfg.OverlapTokens,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create chunker: %w", err)
	}

	retriever := retriever.NewRetriever(vs, embedder, cfg.TopK)
	pipeline := ragpipeline.New(vs, embedder, chunker, loader.DefaultRegistry(), retriever, answerer)

	return &APIServer{
		cfg:      cfg,
		vs:       vs,
		embedder: embedder,
		chunker:  chunker,
		pipeline: pipeline,
	}, nil
}

func applyDefaults(cfg *config.Config) {
	if cfg.Env == "" {
		cfg.Env = "development"
		cfg.PublicUploadEnabled = true
	}
	if cfg.MaxUploadBytes <= 0 {
		cfg.MaxUploadBytes = 10 << 20
	}
}
