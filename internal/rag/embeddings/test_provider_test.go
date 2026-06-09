package embeddings

import (
	"context"
	"testing"
)

func TestTestProviderIsDeterministic(t *testing.T) {
	provider := NewTestProvider(16)

	first, err := provider.EmbedSingle(context.Background(), "technical concept")
	if err != nil {
		t.Fatalf("expected first embedding: %v", err)
	}
	second, err := provider.EmbedSingle(context.Background(), "technical concept")
	if err != nil {
		t.Fatalf("expected second embedding: %v", err)
	}

	if len(first) != 16 {
		t.Fatalf("expected 16 dimensions, got %d", len(first))
	}
	for i := range first {
		if first[i] != second[i] {
			t.Fatalf("expected deterministic embedding at index %d", i)
		}
	}
}

func TestTestProviderDistinguishesDifferentTexts(t *testing.T) {
	provider := NewTestProvider(32)

	first, err := provider.EmbedSingle(context.Background(), "technical foundations")
	if err != nil {
		t.Fatalf("expected first embedding: %v", err)
	}
	second, err := provider.EmbedSingle(context.Background(), "podcast storytelling")
	if err != nil {
		t.Fatalf("expected second embedding: %v", err)
	}

	var different bool
	for i := range first {
		if first[i] != second[i] {
			different = true
			break
		}
	}
	if !different {
		t.Fatal("expected different texts to produce different embeddings")
	}
}
