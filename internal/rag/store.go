package rag

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type VectorStore interface {
	TestConnection() error
	InsertBatch(ctx context.Context, chunks []Chunk) error
	Search(ctx context.Context, embedding []float32, topK int) ([]Chunk, error)
	SearchWithFilters(ctx context.Context, embedding []float32, topK int, filters map[string]interface{}) ([]Chunk, error)
	GetChunksByDocumentID(ctx context.Context, documentID string) ([]Chunk, error)
	InsertDocument(ctx context.Context, doc Document) error
	GetDocumentByID(ctx context.Context, documentID string) (*Document, error)
	DeleteDocument(ctx context.Context, documentID string) error
}

type PGVectorStore struct {
	DB *sql.DB
}

func NewPGVectorStore(dsn string) (*PGVectorStore, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}
	return &PGVectorStore{DB: db}, nil
}

func (store *PGVectorStore) TestConnection() error {
	if err := store.DB.Ping(); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	return nil
}
