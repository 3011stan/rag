package retriever

import (
	"context"
	"fmt"
	"sort"

	"github.com/stan/Projects/studies/rag/internal/rag"
	"github.com/stan/Projects/studies/rag/internal/rag/embeddings"
)

const (
	candidateMultiplier = 4
	maxCandidateK       = 40
)

type Result struct {
	Chunk rag.Chunk
	Score float64
	Rank  int
}

type Preferences struct {
	Layers        []string
	Categories    []string
	Platforms     []string
	SourceKinds   []string
	SourceQuality []string
}

type Retriever struct {
	store       rag.VectorStore
	embedder    embeddings.Provider
	defaultTopK int
}

func NewRetriever(store rag.VectorStore, embedder embeddings.Provider, defaultTopK int) *Retriever {
	return &Retriever{
		store:       store,
		embedder:    embedder,
		defaultTopK: defaultTopK,
	}
}

func (r *Retriever) Retrieve(ctx context.Context, question string, topK int) ([]Result, error) {
	return r.retrieve(ctx, question, topK, nil)
}

func (r *Retriever) RetrieveWithPreferences(ctx context.Context, question string, topK int, preferences *Preferences) ([]Result, error) {
	return r.retrieve(ctx, question, topK, preferences)
}

func (r *Retriever) retrieve(ctx context.Context, question string, topK int, preferences *Preferences) ([]Result, error) {
	if topK <= 0 {
		topK = r.defaultTopK
	}

	embedding, err := r.embedder.EmbedSingle(ctx, question)
	if err != nil {
		return nil, fmt.Errorf("failed to generate question embedding: %w", err)
	}

	candidateK := topK
	if hasPreferences(preferences) {
		candidateK = candidatePoolSize(topK)
	}

	chunks, err := r.store.Search(ctx, embedding, candidateK)
	if err != nil {
		return nil, fmt.Errorf("failed to search chunks: %w", err)
	}

	if hasPreferences(preferences) {
		sort.SliceStable(chunks, func(i, j int) bool {
			return adjustedScore(chunks[i], preferences) < adjustedScore(chunks[j], preferences)
		})
		if len(chunks) > topK {
			chunks = chunks[:topK]
		}
	}

	results := make([]Result, len(chunks))
	for i, chunk := range chunks {
		results[i] = Result{
			Chunk: chunk,
			Score: chunk.Score,
			Rank:  i + 1,
		}
	}

	return results, nil
}

func candidatePoolSize(topK int) int {
	if topK <= 0 {
		return topK
	}
	candidateK := topK * candidateMultiplier
	if candidateK > maxCandidateK {
		return maxCandidateK
	}
	return candidateK
}

func adjustedScore(chunk rag.Chunk, preferences *Preferences) float64 {
	return chunk.Score - metadataBoost(chunk.Metadata, preferences)
}

func metadataBoost(metadata map[string]interface{}, preferences *Preferences) float64 {
	if len(metadata) == 0 || !hasPreferences(preferences) {
		return 0
	}

	var boost float64
	if matchesPreference(metadata["layer"], preferences.Layers) {
		boost += 0.06
	}
	if matchesPreference(metadata["category"], preferences.Categories) {
		boost += 0.05
	}
	if matchesPreference(metadata["platform"], preferences.Platforms) {
		boost += 0.04
	}
	if matchesPreference(metadata["source_kind"], preferences.SourceKinds) {
		boost += 0.03
	}
	if matchesPreference(metadata["source_quality"], preferences.SourceQuality) {
		boost += 0.03
	}
	return boost
}

func matchesPreference(value interface{}, preferences []string) bool {
	text, ok := value.(string)
	if !ok || text == "" || len(preferences) == 0 {
		return false
	}
	for _, preference := range preferences {
		if text == preference {
			return true
		}
	}
	return false
}

func hasPreferences(preferences *Preferences) bool {
	if preferences == nil {
		return false
	}
	return len(preferences.Layers) > 0 ||
		len(preferences.Categories) > 0 ||
		len(preferences.Platforms) > 0 ||
		len(preferences.SourceKinds) > 0 ||
		len(preferences.SourceQuality) > 0
}
