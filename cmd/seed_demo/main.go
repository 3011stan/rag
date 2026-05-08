package main

import (
	"context"
	"crypto/md5"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stan/Projects/studies/rag/internal/config"
	"github.com/stan/Projects/studies/rag/internal/rag"
	"github.com/stan/Projects/studies/rag/internal/rag/chunker"
	"github.com/stan/Projects/studies/rag/internal/rag/providers"
)

const defaultDemoDocsDir = "demo/docs"

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	docsDir := os.Getenv("DEMO_DOCS_DIR")
	if docsDir == "" {
		docsDir = defaultDemoDocsDir
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

	seeded, err := seedDirectory(ctx, docsDir, store, embedder, chk)
	if err != nil {
		log.Fatalf("failed to seed demo docs: %v", err)
	}

	fmt.Printf("Seeded %d demo document(s)\n", seeded)
}

func seedDirectory(
	ctx context.Context,
	docsDir string,
	store rag.VectorStore,
	embedder interface {
		Embed(context.Context, []string) ([][]float32, error)
	},
	chk *chunker.Chunker,
) (int, error) {
	var seeded int

	err := filepath.WalkDir(docsDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !isDemoDoc(path) {
			return nil
		}

		if err := seedFile(ctx, path, store, embedder, chk); err != nil {
			return err
		}
		seeded++
		return nil
	})
	if err != nil {
		return 0, err
	}

	return seeded, nil
}

func seedFile(
	ctx context.Context,
	path string,
	store rag.VectorStore,
	embedder interface {
		Embed(context.Context, []string) ([][]float32, error)
	},
	chk *chunker.Chunker,
) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", path, err)
	}

	text := strings.TrimSpace(string(data))
	if text == "" {
		return nil
	}

	checksum := fmt.Sprintf("%x", md5.Sum(data))
	docID := uuid.NewSHA1(uuid.NameSpaceURL, []byte(path+":"+checksum)).String()

	doc := rag.Document{
		ID:       docID,
		Source:   "demo",
		Title:    strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
		Checksum: checksum,
		Metadata: map[string]interface{}{
			"path": path,
			"type": strings.TrimPrefix(filepath.Ext(path), "."),
		},
	}

	if err := store.InsertDocument(ctx, doc); err != nil {
		return fmt.Errorf("failed to insert demo document %s: %w", path, err)
	}

	chunks, err := chk.ChunkText(doc.ID, text, doc.Metadata)
	if err != nil {
		return fmt.Errorf("failed to chunk demo document %s: %w", path, err)
	}
	if len(chunks) == 0 {
		return nil
	}

	texts := make([]string, len(chunks))
	for i, chunk := range chunks {
		texts[i] = chunk.Content
		chunks[i].ID = uuid.NewSHA1(
			uuid.NameSpaceURL,
			[]byte(fmt.Sprintf("%s:%d:%s", doc.ID, chunk.ChunkIndex, chunk.Content)),
		).String()
	}

	embeddings, err := embedder.Embed(ctx, texts)
	if err != nil {
		return fmt.Errorf("failed to embed demo document %s: %w", path, err)
	}
	if len(embeddings) != len(chunks) {
		return fmt.Errorf("embedding count mismatch for %s", path)
	}

	for i := range chunks {
		chunks[i].Embedding = embeddings[i]
	}

	if err := store.InsertBatch(ctx, chunks); err != nil {
		return fmt.Errorf("failed to insert demo chunks %s: %w", path, err)
	}

	return nil
}

func isDemoDoc(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".md", ".txt":
		return true
	default:
		return false
	}
}
