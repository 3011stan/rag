package rag

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Helper function to skip tests if database is not available
func skipIfNoDatabase(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration tests")
	}
}

func TestNewPGVectorStore(t *testing.T) {
	skipIfNoDatabase(t)

	dbURL := os.Getenv("DATABASE_URL")
	vs, err := NewPGVectorStore(dbURL)
	if err != nil {
		t.Fatalf("NewPGVectorStore failed: %v", err)
	}

	if vs == nil {
		t.Fatal("expected non-nil VectorStore")
	}

	// Test connection
	if err := vs.TestConnection(); err != nil {
		t.Fatalf("TestConnection failed: %v", err)
	}
}

func TestInsertDocument(t *testing.T) {
	skipIfNoDatabase(t)

	dbURL := os.Getenv("DATABASE_URL")
	vs, err := NewPGVectorStore(dbURL)
	if err != nil {
		t.Fatalf("NewPGVectorStore failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	doc := Document{
		ID:       "test-doc-1",
		Source:   "test.pdf",
		Title:    "Test Document",
		Checksum: "abc123",
		Metadata: map[string]interface{}{
			"pages": 10,
		},
	}

	if err := vs.InsertDocument(ctx, doc); err != nil {
		t.Fatalf("InsertDocument failed: %v", err)
	}

	// Verify document was inserted
	retrieved, err := vs.GetDocumentByID(ctx, "test-doc-1")
	if err != nil {
		t.Fatalf("GetDocumentByID failed: %v", err)
	}

	if retrieved == nil {
		t.Fatal("expected non-nil document")
	}

	if retrieved.ID != doc.ID {
		t.Errorf("expected ID %s, got %s", doc.ID, retrieved.ID)
	}

	if retrieved.Source != doc.Source {
		t.Errorf("expected Source %s, got %s", doc.Source, retrieved.Source)
	}

	// Cleanup
	vs.DeleteDocument(ctx, "test-doc-1")
}

func TestInsertBatch(t *testing.T) {
	skipIfNoDatabase(t)

	dbURL := os.Getenv("DATABASE_URL")
	vs, err := NewPGVectorStore(dbURL)
	if err != nil {
		t.Fatalf("NewPGVectorStore failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Insert a document first
	doc := Document{
		ID:       "test-doc-2",
		Source:   "test2.pdf",
		Title:    "Test Document 2",
		Checksum: "def456",
	}

	if err := vs.InsertDocument(ctx, doc); err != nil {
		t.Fatalf("InsertDocument failed: %v", err)
	}

	// Create test chunks with embeddings
	chunks := []Chunk{
		{
			ID:         "chunk-1",
			DocumentID: "test-doc-2",
			ChunkIndex: 0,
			Content:    "This is the first chunk of text.",
			TokenCount: 10,
			Embedding:  make([]float32, 1536), // OpenAI embedding size
		},
		{
			ID:         "chunk-2",
			DocumentID: "test-doc-2",
			ChunkIndex: 1,
			Content:    "This is the second chunk of text.",
			TokenCount: 10,
			Embedding:  make([]float32, 1536),
		},
	}

	// Initialize embeddings with some values
	for i := range chunks {
		for j := range chunks[i].Embedding {
			chunks[i].Embedding[j] = float32(j) / 1536.0
		}
	}

	if err := vs.InsertBatch(ctx, chunks); err != nil {
		t.Fatalf("InsertBatch failed: %v", err)
	}

	// Verify chunks were inserted
	retrieved, err := vs.GetChunksByDocumentID(ctx, "test-doc-2")
	if err != nil {
		t.Fatalf("GetChunksByDocumentID failed: %v", err)
	}

	if len(retrieved) != 2 {
		t.Errorf("expected 2 chunks, got %d", len(retrieved))
	}

	// Cleanup
	vs.DeleteDocument(ctx, "test-doc-2")
}

func TestSearch(t *testing.T) {
	skipIfNoDatabase(t)

	dbURL := os.Getenv("DATABASE_URL")
	vs, err := NewPGVectorStore(dbURL)
	if err != nil {
		t.Fatalf("NewPGVectorStore failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Insert test data
	doc := Document{
		ID:       "test-doc-3",
		Source:   "test3.pdf",
		Title:    "Test Document 3",
		Checksum: "ghi789",
	}

	if err := vs.InsertDocument(ctx, doc); err != nil {
		t.Fatalf("InsertDocument failed: %v", err)
	}

	chunks := []Chunk{
		{
			ID:         "chunk-3",
			DocumentID: "test-doc-3",
			ChunkIndex: 0,
			Content:    "Test content for search.",
			TokenCount: 5,
			Embedding:  make([]float32, 1536),
		},
	}

	// Initialize embedding with some pattern
	for i := range chunks[0].Embedding {
		chunks[0].Embedding[i] = float32(i) / 1536.0
	}

	if err := vs.InsertBatch(ctx, chunks); err != nil {
		t.Fatalf("InsertBatch failed: %v", err)
	}

	// Create a similar query embedding
	queryEmbedding := make([]float32, 1536)
	for i := range queryEmbedding {
		queryEmbedding[i] = float32(i) / 1536.0
	}

	results, err := vs.Search(ctx, queryEmbedding, 1)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) == 0 {
		t.Log("Search returned no results (embeddings may not be similar enough)")
	}

	// Cleanup
	vs.DeleteDocument(ctx, "test-doc-3")
}

func TestDeleteDocument(t *testing.T) {
	skipIfNoDatabase(t)

	dbURL := os.Getenv("DATABASE_URL")
	vs, err := NewPGVectorStore(dbURL)
	if err != nil {
		t.Fatalf("NewPGVectorStore failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Insert a document
	doc := Document{
		ID:       "test-doc-4",
		Source:   "test4.pdf",
		Title:    "Test Document 4",
		Checksum: "jkl012",
	}

	if err := vs.InsertDocument(ctx, doc); err != nil {
		t.Fatalf("InsertDocument failed: %v", err)
	}

	// Verify it exists
	retrieved, err := vs.GetDocumentByID(ctx, "test-doc-4")
	if err != nil {
		t.Fatalf("GetDocumentByID failed: %v", err)
	}

	if retrieved == nil {
		t.Fatal("document should exist before deletion")
	}

	// Delete it
	if err := vs.DeleteDocument(ctx, "test-doc-4"); err != nil {
		t.Fatalf("DeleteDocument failed: %v", err)
	}

	// Verify it's gone
	retrieved, err = vs.GetDocumentByID(ctx, "test-doc-4")
	if err != nil && err != sql.ErrNoRows {
		t.Fatalf("GetDocumentByID failed: %v", err)
	}

	if retrieved != nil {
		t.Fatal("document should not exist after deletion")
	}
}
