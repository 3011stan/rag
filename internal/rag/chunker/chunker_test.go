package chunker

import (
	"strings"
	"testing"
)

func TestNewChunker(t *testing.T) {
	config := ChunkerConfig{
		ChunkTokens:   800,
		OverlapTokens: 100,
	}

	chunker, err := NewChunker(config)
	if err != nil {
		t.Fatalf("NewChunker failed: %v", err)
	}

	if chunker == nil {
		t.Fatal("chunker is nil")
	}
}

func TestChunkText_EmptyText(t *testing.T) {
	config := ChunkerConfig{
		ChunkTokens:   800,
		OverlapTokens: 100,
	}

	chunker, err := NewChunker(config)
	if err != nil {
		t.Fatalf("NewChunker failed: %v", err)
	}

	chunks, err := chunker.ChunkText("doc-1", "", nil)
	if err != nil {
		t.Fatalf("ChunkText failed: %v", err)
	}

	if len(chunks) != 0 {
		t.Errorf("expected 0 chunks for empty text, got %d", len(chunks))
	}
}

func TestChunkText_SmallText(t *testing.T) {
	config := ChunkerConfig{
		ChunkTokens:   800,
		OverlapTokens: 100,
	}

	chunker, err := NewChunker(config)
	if err != nil {
		t.Fatalf("NewChunker failed: %v", err)
	}

	text := "This is a small test text. It should fit in one chunk."
	chunks, err := chunker.ChunkText("doc-1", text, nil)
	if err != nil {
		t.Fatalf("ChunkText failed: %v", err)
	}

	if len(chunks) != 1 {
		t.Errorf("expected 1 chunk for small text, got %d", len(chunks))
	}

	if chunks[0].DocumentID != "doc-1" {
		t.Errorf("expected document_id 'doc-1', got '%s'", chunks[0].DocumentID)
	}

	if len(chunks[0].Content) == 0 {
		t.Fatal("chunk content is empty")
	}
}

func TestChunkText_LargeText(t *testing.T) {
	config := ChunkerConfig{
		ChunkTokens:   100,
		OverlapTokens: 10,
	}

	chunker, err := NewChunker(config)
	if err != nil {
		t.Fatalf("NewChunker failed: %v", err)
	}

	// Create a large text with multiple words
	text := ""
	for i := 0; i < 500; i++ {
		text += "word "
	}

	chunks, err := chunker.ChunkText("doc-2", text, nil)
	if err != nil {
		t.Fatalf("ChunkText failed: %v", err)
	}

	if len(chunks) == 0 {
		t.Fatal("expected at least one chunk for large text")
	}

	// All chunks should have the same document ID
	for i, chunk := range chunks {
		if chunk.DocumentID != "doc-2" {
			t.Errorf("chunk %d has wrong document_id: %s", i, chunk.DocumentID)
		}

		if len(chunk.Content) == 0 {
			t.Errorf("chunk %d has empty content", i)
		}

		if chunk.ChunkIndex != i {
			t.Errorf("chunk %d has wrong index: %d", i, chunk.ChunkIndex)
		}
	}
}

func TestChunkText_WithMetadata(t *testing.T) {
	config := ChunkerConfig{
		ChunkTokens:   800,
		OverlapTokens: 100,
	}

	chunker, err := NewChunker(config)
	if err != nil {
		t.Fatalf("NewChunker failed: %v", err)
	}

	text := "This is a test. " + strings.Repeat("word ", 100)
	metadata := map[string]interface{}{
		"source": "test.pdf",
		"page":   "1",
	}

	chunks, err := chunker.ChunkText("doc-3", text, metadata)
	if err != nil {
		t.Fatalf("ChunkText failed: %v", err)
	}

	if len(chunks) == 0 {
		t.Fatal("expected chunks")
	}

	// Check that metadata is preserved
	if chunks[0].Metadata["source"] != "test.pdf" {
		t.Errorf("expected source 'test.pdf', got '%s'", chunks[0].Metadata["source"])
	}
}
