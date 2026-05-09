package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/stan/Projects/studies/rag/internal/config"
	demoseed "github.com/stan/Projects/studies/rag/internal/demo/seed"
	"github.com/stan/Projects/studies/rag/internal/rag"
	"github.com/stan/Projects/studies/rag/internal/rag/chunker"
	"github.com/stan/Projects/studies/rag/internal/rag/providers"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	docsDir := os.Getenv("DEMO_DOCS_DIR")
	if docsDir == "" {
		docsDir = demoseed.DefaultDocsDir
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	store, err := rag.NewPGVectorStore(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to create vector store: %v", err)
	}
	defer store.DB.Close()

	if err := store.EnsureSchema(ctx, cfg.EmbeddingDimensions); err != nil {
		log.Fatalf("failed to ensure schema: %v", err)
	}

	embedder, err := providers.NewEmbeddingProvider(cfg)
	if err != nil {
		log.Fatalf("failed to create embedding provider: %v", err)
	}

	chk, err := chunker.NewChunker(chunker.ChunkerConfig{
		ChunkTokens:   cfg.ChunkTokens,
		OverlapTokens: cfg.OverlapTokens,
	})
	if err != nil {
		log.Fatalf("failed to create chunker: %v", err)
	}

	seeded, err := demoseed.Directory(ctx, docsDir, store, embedder, chk)
	if err != nil {
		log.Fatalf("failed to seed demo docs: %v", err)
	}

	fmt.Printf("Seeded %d demo document(s)\n", seeded)
}
