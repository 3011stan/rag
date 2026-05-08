package embeddings

import (
	"context"
	"os"
	"testing"
)

func TestNewOpenAIProvider(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping embeddings tests")
	}

	provider := NewOpenAIProvider(apiKey)
	if provider == nil {
		t.Fatal("expected non-nil provider")
	}
}

func TestEmbed_EmptyList(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	provider := NewOpenAIProvider(apiKey)
	ctx := context.Background()

	embeddings, err := provider.Embed(ctx, []string{})
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}

	if len(embeddings) != 0 {
		t.Errorf("expected 0 embeddings for empty list, got %d", len(embeddings))
	}
}

func TestEmbed_SingleText(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	provider := NewOpenAIProvider(apiKey)
	ctx := context.Background()

	texts := []string{"Hello, world!"}
	embeddings, err := provider.Embed(ctx, texts)

	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}

	if len(embeddings) != 1 {
		t.Errorf("expected 1 embedding, got %d", len(embeddings))
	}

	// Check embedding is valid (non-empty, correct size for OpenAI small)
	if len(embeddings[0]) == 0 {
		t.Fatal("embedding is empty")
	}

	// OpenAI embeddings are typically 1536-dimensional for text-embedding-3-small
	if len(embeddings[0]) < 100 {
		t.Errorf("embedding seems too small: %d dimensions", len(embeddings[0]))
	}
}

func TestEmbed_MultipleTexts(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	provider := NewOpenAIProvider(apiKey)
	ctx := context.Background()

	texts := []string{
		"First text",
		"Second text",
		"Third text",
	}

	embeddings, err := provider.Embed(ctx, texts)

	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}

	if len(embeddings) != 3 {
		t.Errorf("expected 3 embeddings, got %d", len(embeddings))
	}

	// Check all embeddings have the same dimensionality
	firstDim := len(embeddings[0])
	for i, emb := range embeddings {
		if len(emb) != firstDim {
			t.Errorf("embedding %d has different dimensions: %d vs %d", i, len(emb), firstDim)
		}
	}
}

func TestEmbed_BatchProcessing(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	provider := NewOpenAIProvider(apiKey)
	ctx := context.Background()

	// Create many texts to test batching (default batch size is 64)
	texts := make([]string, 100)
	for i := 0; i < 100; i++ {
		texts[i] = "Text number " + string(rune(i))
	}

	embeddings, err := provider.Embed(ctx, texts)

	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}

	if len(embeddings) != 100 {
		t.Errorf("expected 100 embeddings, got %d", len(embeddings))
	}
}
