package retriever

import (
	"context"
	"fmt"

	"github.com/stan/Projects/studies/rag/internal/rag"
	"github.com/stan/Projects/studies/rag/internal/rag/embeddings"
)

type Result struct {
	Chunk rag.Chunk
	Score float64
	Rank  int
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
	if topK <= 0 {
		topK = r.defaultTopK
	}

	embedding, err := r.embedder.EmbedSingle(ctx, question)
	if err != nil {
		return nil, fmt.Errorf("failed to generate question embedding: %w", err)
	}

	chunks, err := r.store.Search(ctx, embedding, topK)
	if err != nil {
		return nil, fmt.Errorf("failed to search chunks: %w", err)
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
