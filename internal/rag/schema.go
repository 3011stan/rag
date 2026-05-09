package rag

import (
	"context"
	"fmt"
	"strings"
)

func (store *PGVectorStore) EnsureSchema(ctx context.Context, embeddingDimensions int) error {
	if embeddingDimensions <= 0 {
		return fmt.Errorf("embedding dimensions must be positive")
	}

	statements := []string{
		`CREATE EXTENSION IF NOT EXISTS vector`,
		`CREATE TABLE IF NOT EXISTS rag_documents (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			source TEXT,
			title TEXT,
			checksum TEXT UNIQUE,
			metadata JSONB,
			created_at timestamptz DEFAULT now()
		)`,
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS rag_chunks (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			document_id UUID REFERENCES rag_documents(id) ON DELETE CASCADE,
			chunk_index INT NOT NULL,
			content TEXT NOT NULL,
			token_count INT,
			metadata JSONB,
			embedding vector(%d),
			created_at timestamptz DEFAULT now()
		)`, embeddingDimensions),
		`CREATE INDEX IF NOT EXISTS rag_chunks_embedding_idx
			ON rag_chunks USING ivfflat (embedding vector_cosine_ops)`,
	}

	for _, statement := range statements {
		if _, err := store.DB.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("failed to apply schema statement: %w", err)
		}
	}

	return store.validateEmbeddingDimensions(ctx, embeddingDimensions)
}

func (store *PGVectorStore) validateEmbeddingDimensions(ctx context.Context, expectedDimensions int) error {
	var embeddingType string
	err := store.DB.QueryRowContext(ctx, `
		SELECT format_type(a.atttypid, a.atttypmod)
		FROM pg_attribute a
		JOIN pg_class c ON c.oid = a.attrelid
		WHERE c.relname = 'rag_chunks'
			AND a.attname = 'embedding'
			AND a.attisdropped = false
		LIMIT 1
	`).Scan(&embeddingType)
	if err != nil {
		return fmt.Errorf("failed to inspect embedding column: %w", err)
	}

	expectedType := fmt.Sprintf("vector(%d)", expectedDimensions)
	if !strings.EqualFold(embeddingType, expectedType) {
		return fmt.Errorf("embedding column has type %s, expected %s", embeddingType, expectedType)
	}

	return nil
}
