package retriever

import (
	"context"
	"math"
	"testing"

	"github.com/stan/Projects/studies/rag/internal/rag"
)

type fakeStore struct {
	chunks     []rag.Chunk
	searchedK  int
	embedding  []float32
	searchCall int
}

func (f *fakeStore) TestConnection() error { return nil }
func (f *fakeStore) InsertBatch(context.Context, []rag.Chunk) error {
	return nil
}
func (f *fakeStore) Search(_ context.Context, embedding []float32, topK int) ([]rag.Chunk, error) {
	f.searchCall++
	f.searchedK = topK
	f.embedding = embedding
	if topK > len(f.chunks) {
		topK = len(f.chunks)
	}
	return append([]rag.Chunk(nil), f.chunks[:topK]...), nil
}
func (f *fakeStore) SearchWithFilters(context.Context, []float32, int, map[string]interface{}) ([]rag.Chunk, error) {
	return nil, nil
}
func (f *fakeStore) GetChunksByDocumentID(context.Context, string) ([]rag.Chunk, error) {
	return nil, nil
}
func (f *fakeStore) InsertDocument(context.Context, rag.Document) error { return nil }
func (f *fakeStore) GetDocumentByID(context.Context, string) (*rag.Document, error) {
	return nil, nil
}
func (f *fakeStore) ListDocuments(context.Context) ([]rag.DocumentSummary, error) {
	return nil, nil
}
func (f *fakeStore) DeleteDocument(context.Context, string) error { return nil }

type fakeEmbedder struct{}

func (fakeEmbedder) Embed(context.Context, []string) ([][]float32, error) {
	return [][]float32{{0.1, 0.2}}, nil
}
func (fakeEmbedder) EmbedSingle(context.Context, string) ([]float32, error) {
	return []float32{0.1, 0.2}, nil
}

func TestRetrieveWithPreferencesKeepsSemanticOrderWithoutPreferences(t *testing.T) {
	store := &fakeStore{
		chunks: []rag.Chunk{
			{ID: "a", Score: 0.10},
			{ID: "b", Score: 0.20, Metadata: map[string]interface{}{"layer": "foundations"}},
		},
	}
	ret := NewRetriever(store, fakeEmbedder{}, 5)

	results, err := ret.RetrieveWithPreferences(context.Background(), "question", 2, nil)
	if err != nil {
		t.Fatalf("expected retrieval to succeed: %v", err)
	}
	if store.searchedK != 2 {
		t.Fatalf("expected search topK 2 without preferences, got %d", store.searchedK)
	}
	if results[0].Chunk.ID != "a" || results[1].Chunk.ID != "b" {
		t.Fatalf("expected semantic order to be preserved, got %s, %s", results[0].Chunk.ID, results[1].Chunk.ID)
	}
}

func TestRetrieveWithPreferencesBoostsMetadataMatches(t *testing.T) {
	store := &fakeStore{
		chunks: []rag.Chunk{
			{ID: "semantic-best", Score: 0.10},
			{ID: "metadata-match", Score: 0.15, Metadata: map[string]interface{}{"layer": "foundations"}},
		},
	}
	ret := NewRetriever(store, fakeEmbedder{}, 5)

	results, err := ret.RetrieveWithPreferences(context.Background(), "question", 2, &Preferences{
		Layers: []string{"foundations"},
	})
	if err != nil {
		t.Fatalf("expected retrieval to succeed: %v", err)
	}
	if results[0].Chunk.ID != "metadata-match" {
		t.Fatalf("expected metadata match to be boosted first, got %s", results[0].Chunk.ID)
	}
	if results[0].Score != 0.15 {
		t.Fatalf("expected original semantic score to be preserved, got %v", results[0].Score)
	}
}

func TestRetrieveWithPreferencesUsesCandidatePoolAndRespectsTopK(t *testing.T) {
	store := &fakeStore{
		chunks: []rag.Chunk{
			{ID: "a", Score: 0.10},
			{ID: "b", Score: 0.11},
			{ID: "c", Score: 0.12},
			{ID: "d", Score: 0.13, Metadata: map[string]interface{}{"category": "storytelling"}},
			{ID: "e", Score: 0.14},
			{ID: "f", Score: 0.15},
			{ID: "g", Score: 0.16},
			{ID: "h", Score: 0.17},
		},
	}
	ret := NewRetriever(store, fakeEmbedder{}, 5)

	results, err := ret.RetrieveWithPreferences(context.Background(), "question", 2, &Preferences{
		Categories: []string{"storytelling"},
	})
	if err != nil {
		t.Fatalf("expected retrieval to succeed: %v", err)
	}
	if store.searchedK != 8 {
		t.Fatalf("expected candidate pool of 8, got %d", store.searchedK)
	}
	if len(results) != 2 {
		t.Fatalf("expected final topK 2, got %d", len(results))
	}
	if results[0].Chunk.ID != "d" {
		t.Fatalf("expected boosted candidate from wider pool first, got %s", results[0].Chunk.ID)
	}
}

func TestCandidatePoolSizeIsCapped(t *testing.T) {
	if got := candidatePoolSize(20); got != maxCandidateK {
		t.Fatalf("expected candidate pool cap %d, got %d", maxCandidateK, got)
	}
}

func TestRetrieveCapsTopKBeforeCandidatePool(t *testing.T) {
	chunks := make([]rag.Chunk, 60)
	for i := range chunks {
		chunks[i] = rag.Chunk{ID: "chunk", Score: float64(i) / 100}
	}
	store := &fakeStore{chunks: chunks}
	ret := NewRetriever(store, fakeEmbedder{}, 5)

	results, err := ret.RetrieveWithPreferences(context.Background(), "question", 50, &Preferences{
		Layers: []string{"foundations"},
	})
	if err != nil {
		t.Fatalf("expected retrieval to succeed: %v", err)
	}
	if store.searchedK != maxCandidateK {
		t.Fatalf("expected candidate pool cap %d, got %d", maxCandidateK, store.searchedK)
	}
	if len(results) != maxTopK {
		t.Fatalf("expected normalized topK %d, got %d", maxTopK, len(results))
	}
}

func TestMetadataBoostAddsMultipleMatches(t *testing.T) {
	boost := metadataBoost(map[string]interface{}{
		"layer":          "foundations",
		"category":       "storytelling",
		"platform":       "reels",
		"source_kind":    "note",
		"source_quality": "high",
	}, &Preferences{
		Layers:        []string{"foundations"},
		Categories:    []string{"storytelling"},
		Platforms:     []string{"reels"},
		SourceKinds:   []string{"note"},
		SourceQuality: []string{"high"},
	})

	if math.Abs(boost-0.21) > 0.00001 {
		t.Fatalf("expected total boost 0.21, got %v", boost)
	}
}
